package dao

import (
	"database/sql"
	"hash/fnv"
	"time"

	taskmodel "github.com/KDKHD/go-resilient-task/internal/model/task"
	"github.com/KDKHD/go-resilient-task/internal/postgres"
	u "github.com/KDKHD/go-resilient-task/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type DataFormat int

const (
	JSON DataFormat = iota
)

var (
	insertTaskStmt = `
		INSERT INTO task
		(
			id,
			type,
			sub_type,
			status,
			next_event_time,
			state_time,
			time_created,
			time_updated,
			processing_tries_count,
			version,
			priority
		)
		VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	insertUniqueTaskKeyStmt = `
		INSERT INTO unique_task_key
		(
			task_id,
			key_hash,
			key
		)
		VALUES
		($1, $2, $3)
	`
	setStatusStmt = `
		UPDATE task SET status = $1, next_event_time = $2, state_time = $3, time_updated = $4, version = $5 WHERE id = $6 AND version = $7
	`
	insertTaskDataStmt = `
		INSERT INTO task_data (task_id, data_format, data) VALUES ($1, $2, $3)
	`

	grabForProcessingWithStatusStmt = `
		UPDATE task set status = $1, processing_start_time = $2, next_event_time = $3, processing_tries_count = processing_tries_count + 1, state_time = $4, time_updated = $5, version = $6 where id = $7 and version = $8 and status = $9
	`

	getTaskStmt1 = `
		SELECT id, version, type, status, priority, sub_type, processing_tries_count, d.data_format, d.data from task t left join task_data d on t.id = d.task_id where t.id = $1
	`

	getStuckTasksSql = `
		select id,version,type,priority,status from task where status = $1 and next_event_time<$2 order by next_event_time DESC limit $3
	`

	deleteTaskStmt = `
		DELETE FROM task WHERE id = $1 AND version = $2
	`

	setToBeRetriedStmt = `
		UPDATE task SET status = $1, next_event_time = $2, state_time = $3, time_updated = $4, version = $5 WHERE id = $6 AND version = $7
	`

	setToBeRetried1Stmt = `
		UPDATE task SET status = $1, next_event_time = $2, processing_tries_count = $3, state_time = $4, time_updated = $5, version = $6 WHERE id = $7 AND version = $8
	`

	getTaskVersionStmt = `
		SELECT version FROM task WHERE id = $1
	`
)

type PostgresTaskDao struct {
	db                              *sql.DB
	insertTaskStmt                  *sql.Stmt
	insertUniqueTaskKeyStmt         *sql.Stmt
	setStatusStmt                   *sql.Stmt
	insertTaskDataStmt              *sql.Stmt
	grabForProcessingWithStatusStmt *sql.Stmt
	getTaskStmt1                    *sql.Stmt
	getStuckTasksSql                *sql.Stmt
	deleteTaskStmt                  *sql.Stmt
	setToBeRetriedStmt              *sql.Stmt
	setToBeRetried1Stmt             *sql.Stmt
	getTaskVersionStmt              *sql.Stmt
	logger                          *zap.Logger
}

func NewPostgresTaskDao(db *sql.DB, logger *zap.Logger) (*PostgresTaskDao, error) {
	insertTaskStmt, err := db.Prepare(insertTaskStmt)
	if err != nil {
		return &PostgresTaskDao{}, err
	}

	insertUniqueTaskKeyStmt, err := db.Prepare(insertUniqueTaskKeyStmt)
	if err != nil {
		return &PostgresTaskDao{}, err
	}

	setStatusStmt, err := db.Prepare(setStatusStmt)
	if err != nil {
		return &PostgresTaskDao{}, err
	}

	insertTaskDataStmt, err := db.Prepare(insertTaskDataStmt)
	if err != nil {
		return &PostgresTaskDao{}, err
	}

	grabForProcessingWithStatusStmt, err := db.Prepare(grabForProcessingWithStatusStmt)
	if err != nil {
		return &PostgresTaskDao{}, err
	}

	getTaskStmt1, err := db.Prepare(getTaskStmt1)
	if err != nil {
		return &PostgresTaskDao{}, err
	}

	getStuckTasksSql, err := db.Prepare(getStuckTasksSql)
	if err != nil {
		return &PostgresTaskDao{}, err
	}

	deleteTaskStmt, err := db.Prepare(deleteTaskStmt)
	if err != nil {
		return &PostgresTaskDao{}, err
	}

	setToBeRetriedStmt, err := db.Prepare(setToBeRetriedStmt)
	if err != nil {
		return &PostgresTaskDao{}, err
	}

	setToBeRetried1Stmt, err := db.Prepare(setToBeRetried1Stmt)
	if err != nil {
		return &PostgresTaskDao{}, err
	}

	getTaskVersionStmt, err := db.Prepare(getTaskVersionStmt)
	if err != nil {
		return &PostgresTaskDao{}, err
	}

	return &PostgresTaskDao{
		db:                              db,
		insertTaskStmt:                  insertTaskStmt,
		insertUniqueTaskKeyStmt:         insertUniqueTaskKeyStmt,
		setStatusStmt:                   setStatusStmt,
		insertTaskDataStmt:              insertTaskDataStmt,
		grabForProcessingWithStatusStmt: grabForProcessingWithStatusStmt,
		getTaskStmt1:                    getTaskStmt1,
		getStuckTasksSql:                getStuckTasksSql,
		deleteTaskStmt:                  deleteTaskStmt,
		setToBeRetriedStmt:              setToBeRetriedStmt,
		setToBeRetried1Stmt:             setToBeRetried1Stmt,
		getTaskVersionStmt:              getTaskVersionStmt,
		logger:                          logger,
	}, nil
}

func stringHashCode(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func (repository PostgresTaskDao) InsertTask(request InsertTaskRequest) (InsertTaskResponse, error) {

	now := time.Now().UTC()

	nextEventTime := request.MaxStuckTime

	uuidProvided := request.TaskId != uuid.Nil
	key := request.Key
	keyProvided := key != ""

	taskId := u.If(uuidProvided, request.TaskId, uuid.New())

	repository.logger.Debug("Inserting task", zap.String("taskId", taskId.String()), zap.String("type", request.Type), zap.String("subType", request.SubType), zap.Time("nextEventTime", nextEventTime), zap.Time("now", now), zap.Int("priority", request.Priority))

	// Start a new transaction
	tx, err := repository.db.Begin()
	if err != nil {
		return InsertTaskResponse{}, err
	}
	defer tx.Rollback()

	if keyProvided {

		keyHash := stringHashCode(key)

		insertUniqueTaskKeyStmt := tx.Stmt(repository.insertUniqueTaskKeyStmt)

		defer insertUniqueTaskKeyStmt.Close()

		result, err := insertUniqueTaskKeyStmt.Exec(
			taskId,  // task_id
			keyHash, // key_hash
			key,     // key
		)

		data, ok := err.(*pgconn.PgError)

		if ok && data.Code != postgres.PostgresUniqueViolation {
			repository.logger.Error("Task insertion did not succeed. The warning code is unknown: " + data.Code)
			return InsertTaskResponse{}, err
		} else if ok && data.Code == postgres.PostgresUniqueViolation {
			return InsertTaskResponse{TaskId: taskId, Inserted: false}, nil
		}

		insertedCount, err := result.RowsAffected()

		if err != nil {
			return InsertTaskResponse{}, err
		}

		if insertedCount == 0 {

			return InsertTaskResponse{TaskId: taskId, Inserted: false}, nil
		}

	}

	insertTaskStmt := tx.Stmt(repository.insertTaskStmt)

	result, err := insertTaskStmt.Exec(
		taskId,                  // id
		request.Type,            // type
		request.SubType,         // sub_type
		request.Status.String(), // status
		nextEventTime,           // next_event_time
		now,                     // state_time
		now,                     // time_created
		now,                     // time_updated
		0,                       // process_tries_count
		0,                       // version
		request.Priority,        // priority
	)

	data, ok := err.(*pgconn.PgError)

	if ok && data.Code != postgres.PostgresUniqueViolation {
		repository.logger.Error("Task insertion did not succeed. The warning code is unknown: " + data.Code)
		return InsertTaskResponse{}, err
	} else if ok && data.Code == postgres.PostgresUniqueViolation {
		return InsertTaskResponse{TaskId: taskId, Inserted: false}, nil
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return InsertTaskResponse{}, err
	}

	if rowsAffected == 0 {

		return InsertTaskResponse{TaskId: taskId, Inserted: false}, nil
	}

	if request.Data != nil {
		insertTaskDataStmt, err := tx.Prepare(insertTaskDataStmt)
		if err != nil {
			return InsertTaskResponse{}, err
		}
		_, err = insertTaskDataStmt.Exec(taskId, JSON, request.Data)
		if err != nil {
			return InsertTaskResponse{}, err

		}
	}

	// Commit the transaction if everything is successful
	if err := tx.Commit(); err != nil {
		return InsertTaskResponse{}, err
	}

	return InsertTaskResponse{TaskId: request.TaskId, Inserted: true}, nil
}

func (repository PostgresTaskDao) SetStatus(taskId uuid.UUID, taskStatus taskmodel.TaskStatus, version int) (bool, error) {
	// set timezone,
	now := time.Now().UTC()

	result, err := repository.setStatusStmt.Exec(
		taskStatus.String(), // status
		now,                 // next_event_time
		now,                 // state_time
		now,                 // time_updated
		version+1,           // version
		taskId,              // task_id
		version,             // version
	)

	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (repository PostgresTaskDao) GrabForProcessing(baseTask taskmodel.IBaseTask, maxProcessingEndTime time.Time) (taskmodel.ITask, error) {

	now := time.Now().UTC()

	result, err := repository.grabForProcessingWithStatusStmt.Exec(
		taskmodel.PROCESSING.String(), // status
		now,                           // processing_start_time
		maxProcessingEndTime,          // next_event_time
		now,                           // state_time
		now,                           // time_updated
		baseTask.GetVersion()+1,       // new version
		baseTask.GetId(),              // task_id
		baseTask.GetVersion(),         // version
		taskmodel.SUBMITTED.String(),  // status
	)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		repository.logger.Error("No task was grabbed for processing")
		return nil, nil
	}

	task, err := repository.GetTask(baseTask.GetId())

	if err != nil {
		return nil, err
	}

	return task, nil
}

func (repository PostgresTaskDao) GetStuckTasks(maxTasks int, status taskmodel.TaskStatus) (GetStuckTasksResponse, error) {
	repository.logger.Debug("Getting stuck tasks", zap.Int("maxTasks", maxTasks), zap.String("status", status.String()), zap.Time("now", time.Now().UTC()))
	rows, err := repository.getStuckTasksSql.Query(status.String(), time.Now().UTC(), maxTasks)

	if err != nil {
		return GetStuckTasksResponse{}, err
	}

	defer rows.Close()

	var tasks []taskmodel.IStuckTask

	for rows.Next() {
		var task taskmodel.StuckTask
		var statusString string

		err := rows.Scan(
			&task.Id,
			&task.Version,
			&task.Type,
			&task.Priority,
			&statusString,
		)

		if err != nil {
			return GetStuckTasksResponse{}, err
		}

		taskStatus, err := taskmodel.TaskStatusFromString(statusString)
		if err != nil {
			return GetStuckTasksResponse{}, err
		}

		task.Status = taskStatus

		tasks = append(tasks, &task)
	}

	if len(tasks) == 0 {
		return GetStuckTasksResponse{HasMore: false}, nil
	}

	return GetStuckTasksResponse{
		StuckTasks: tasks,
		HasMore:    true,
	}, nil

}

func (repository PostgresTaskDao) GetTask(taskId uuid.UUID) (taskmodel.ITask, error) {

	rows, err := repository.getTaskStmt1.Query(taskId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var task taskmodel.Task
	var dataFormat sql.NullInt32
	var data sql.NullString
	var statusString string

	if rows.Next() {
		err := rows.Scan(
			&task.Id,
			&task.Version,
			&task.Type,
			&statusString,
			&task.Priority,
			&task.SubType,
			&task.ProcessingTriesCount,
			&dataFormat,
			&data,
		)
		if err != nil {
			return nil, err
		}

		if data.Valid {
			task.Data = []byte(data.String)
		}

		taskStatus, err := taskmodel.TaskStatusFromString(statusString)
		if err != nil {
			return nil, err
		}

		task.Status = taskStatus

	} else {
		return nil, sql.ErrNoRows
	}

	return &task, nil

}

func (repository PostgresTaskDao) MarkAsSubmitted(taskId uuid.UUID, version int, maxStuckTime time.Time) (bool, error) {
	// set timezone,
	maxStuckTimeLoc := maxStuckTime.UTC()
	now := time.Now().UTC()

	result, err := repository.setStatusStmt.Exec(
		taskmodel.SUBMITTED.String(), // status
		maxStuckTimeLoc,              // next_event_time
		now,                          // state_time
		now,                          // time_updated
		version+1,                    // version
		taskId,                       // task_id
		version,                      // version
	)

	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (repository PostgresTaskDao) DeleteTask(taskId uuid.UUID, version int) (bool, error) {
	result, err := repository.deleteTaskStmt.Exec(taskId, version)

	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (repository PostgresTaskDao) SetToBeRetried(taskId uuid.UUID, shouldRetry bool, retryTime time.Time, version int, resetTriesCount bool) (bool, error) {
	now := time.Now().UTC()

	var result sql.Result
	var err error
	if resetTriesCount {
		result, err = repository.setToBeRetried1Stmt.Exec(
			taskmodel.WAITING.String(), // status
			u.If(shouldRetry, sql.NullTime{Valid: true, Time: retryTime}, sql.NullTime{Valid: false}), // next_event_time
			0,         // process_tries_count
			now,       // state_time
			now,       // time_updated
			version+1, // version
			taskId,    // task_id
			version,   // version
		)
	} else {
		result, err = repository.setToBeRetriedStmt.Exec(
			taskmodel.WAITING.String(), // status
			retryTime,                  // next_event_time
			now,                        // state_time
			now,                        // time_updated
			version+1,                  // version
			taskId,                     // task_id
			version,                    // version
		)
	}

	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil

}

func (repository PostgresTaskDao) GetTaskVersion(taskId uuid.UUID) (int, error) {
	rows, err := repository.getTaskVersionStmt.Query(taskId)

	if err != nil {
		return 0, err
	}

	defer rows.Close()

	var version int

	if rows.Next() {
		err := rows.Scan(&version)
		if err != nil {
			return 0, err
		}
	} else {
		return 0, sql.ErrNoRows
	}

	return version, nil
}
