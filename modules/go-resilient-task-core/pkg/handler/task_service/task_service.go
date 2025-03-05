package taskservice

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/dao"
	taskexecutiontrigger "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/triggering/task_execution_trigger"
	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/model/task"
	taskproperties "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/model/task_properties"
	u "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AddTaskRequest struct {
	Type               string
	SubType            string
	Data               []byte
	TaskId             uuid.UUID
	UniqueKey          string
	RunAfterTime       time.Time
	WarnWhenTaskExists bool
	ExpectedQueueTime  time.Duration
}

type Result int

const (
	OK Result = iota
	ALREADY_EXISTS
)

type AddTaskResponse struct {
	TaskId uuid.UUID
	Result Result
}

type ITasksService interface {
	AddTask(request AddTaskRequest) (AddTaskResponse, error)
}

type TasksService struct {
	taskDao                 dao.ITaskDao
	tasksExecutionTriggerer taskexecutiontrigger.ITasksExecutionTriggerer
	logger                  *zap.Logger
	taskProperties          taskproperties.ITaskProperties
}

func NewTasksService(taskDao dao.ITaskDao, tasksExecutionTriggerer taskexecutiontrigger.ITasksExecutionTriggerer, logger *zap.Logger, taskProperties taskproperties.ITaskProperties) ITasksService {
	return TasksService{taskDao: taskDao, tasksExecutionTriggerer: tasksExecutionTriggerer, logger: logger, taskProperties: taskProperties}
}

func (ts TasksService) AddTask(request AddTaskRequest) (AddTaskResponse, error) {

	now := time.Now().UTC()

	status := u.If(request.RunAfterTime.After(now), taskmodel.WAITING, taskmodel.SUBMITTED)

	if strings.TrimSpace(request.Type) == "" {
		slog.Error("Type is required")
		return AddTaskResponse{}, errors.New("type is required")
	}

	maxStuckTime := u.If(request.ExpectedQueueTime == 0, now.Add(ts.taskProperties.GetTaskStuckTimeout()), now.Add(request.ExpectedQueueTime))

	priority := 1 // Multiple priorities are not supported yet

	insertTaskRequest := dao.InsertTaskRequest{
		Data:         request.Data,
		Key:          request.UniqueKey,
		RunAfterTime: request.RunAfterTime,
		SubType:      request.SubType,
		Type:         request.Type,
		TaskId:       request.TaskId,
		MaxStuckTime: maxStuckTime,
		Status:       status,
		Priority:     priority,
	}

	insertTaskResponse, error := ts.taskDao.InsertTask(insertTaskRequest)

	if error != nil {
		ts.logger.Error("Failed to insert task", zap.Error(error))
		return AddTaskResponse{}, error
	}

	if !insertTaskResponse.Inserted {
		ts.logger.Warn("Task already exists")

		return AddTaskResponse{TaskId: insertTaskResponse.TaskId, Result: ALREADY_EXISTS}, nil
	}

	if status == taskmodel.SUBMITTED {

		ts.triggerTask(&taskmodel.BaseTask{Id: insertTaskResponse.TaskId, Type: request.Type, Priority: priority})
	}

	ts.logger.Debug("Task added", zap.String("TaskId", insertTaskResponse.TaskId.String()))
	return AddTaskResponse{TaskId: insertTaskResponse.TaskId, Result: OK}, nil
}

func (ts *TasksService) triggerTask(baseTask taskmodel.IBaseTask) {
	ts.tasksExecutionTriggerer.Trigger(baseTask)
}
