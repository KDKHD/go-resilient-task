package taskservice

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/dao"
	taskexecutiontrigger "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/triggering/task_execution_trigger"
	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/model/task"
	u "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/utils"
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
	Priority           *int
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
}

func NewTasksService(taskDao dao.ITaskDao, tasksExecutionTriggerer taskexecutiontrigger.ITasksExecutionTriggerer, logger *zap.Logger) TasksService {
	return TasksService{taskDao: taskDao, tasksExecutionTriggerer: tasksExecutionTriggerer, logger: logger}
}

func (service TasksService) AddTask(request AddTaskRequest) (AddTaskResponse, error) {

	now := time.Now().UTC()

	status := u.If(request.RunAfterTime.After(now), taskmodel.WAITING, taskmodel.SUBMITTED)

	if strings.TrimSpace(request.Type) == "" {
		slog.Error("Type is required")
		return AddTaskResponse{}, errors.New("type is required")
	}

	maxStuckTime := now.Add(request.ExpectedQueueTime)

	priority := 1

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

	insertTaskResponse, error := service.taskDao.InsertTask(insertTaskRequest)

	if error != nil {
		service.logger.Error("Failed to insert task", zap.Error(error))
		return AddTaskResponse{}, error
	}

	if !insertTaskResponse.Inserted {
		service.logger.Warn("Task already exists")

		return AddTaskResponse{TaskId: insertTaskResponse.TaskId, Result: ALREADY_EXISTS}, nil
	}

	if status == taskmodel.SUBMITTED {

		service.triggerTask(&taskmodel.BaseTask{Id: insertTaskResponse.TaskId, Type: request.Type, Priority: priority})
	}

	service.logger.Debug("Task added", zap.String("TaskId", insertTaskResponse.TaskId.String()))
	return AddTaskResponse{TaskId: insertTaskResponse.TaskId, Result: OK}, nil
}

func (service *TasksService) triggerTask(baseTask taskmodel.IBaseTask) {
	service.tasksExecutionTriggerer.Trigger(baseTask)
}
