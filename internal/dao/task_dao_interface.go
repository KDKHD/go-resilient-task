package dao

import (
	"time"

	taskmodel "github.com/KDKHD/go-resilient-task/internal/model/task"
	"github.com/google/uuid"
)

type GetStuckTasksResponse struct {
	StuckTasks []taskmodel.IStuckTask
	HasMore    bool
}

type InsertTaskRequest struct {
	Type         string
	SubType      string
	Data         []byte
	TaskId       uuid.UUID
	Key          string
	RunAfterTime time.Time
	Status       taskmodel.TaskStatus
	MaxStuckTime time.Time
	Priority     int
}

type InsertTaskResponse struct {
	TaskId   uuid.UUID
	Inserted bool
}

type ITaskDao interface {
	InsertTask(request InsertTaskRequest) (InsertTaskResponse, error)
	SetStatus(taskId uuid.UUID, taskStatus taskmodel.TaskStatus, version int) (bool, error)
	GrabForProcessing(baseTask taskmodel.IBaseTask, maxProcessingEndTime time.Time) (taskmodel.ITask, error)
	DeleteTask(taskId uuid.UUID, version int) (bool, error)
	GetTask(taskId uuid.UUID) (taskmodel.ITask, error)
	GetStuckTasks(maxTasks int, status taskmodel.TaskStatus) (GetStuckTasksResponse, error)
	MarkAsSubmitted(taskId uuid.UUID, version int, maxStuckTime time.Time) (bool, error)
	SetToBeRetried(taskId uuid.UUID, shouldRetry bool, retryTime time.Time, version int, resetTriesCount bool) (bool, error)
}

func NewInsertTaskRequest(taskType string, subType string, data []byte, taskId uuid.UUID, key string, runAfterTime time.Time, status taskmodel.TaskStatus, maxStuckTime time.Time, priority int) InsertTaskRequest {
	return InsertTaskRequest{
		Type:         taskType,
		SubType:      subType,
		Data:         data,
		TaskId:       taskId,
		Key:          key,
		RunAfterTime: runAfterTime,
		Status:       status,
		MaxStuckTime: maxStuckTime,
		Priority:     priority,
	}
}

func NewInsertTaskResponse(taskId uuid.UUID, inserted bool) InsertTaskResponse {
	return InsertTaskResponse{
		TaskId:   taskId,
		Inserted: inserted,
	}
}
