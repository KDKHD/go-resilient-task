package handlerregistry

import (
	"errors"
	"time"

	processingpolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/processing_policy"
	taskhandler "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/task_handler"
	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/model/task"
	"go.uber.org/zap"
)

type ITaskHandlerRegistry interface {
	GetTaskHandler(task taskmodel.IBaseTask) (taskhandler.ITaskHandler, error)
	GetTaskHandlers() []taskhandler.ITaskHandler
	SetTaskHandlers(handlers []taskhandler.ITaskHandler)
	GetExpectedProcessingMoment(task taskmodel.IBaseTask) (time.Time, error)
}

type TaskHandlerRegistry struct {
	handlers []taskhandler.ITaskHandler
	logger   *zap.Logger
}

func NewTaskHandlerRegistry(logger *zap.Logger) *TaskHandlerRegistry {
	return &TaskHandlerRegistry{logger: logger}
}

func (registry *TaskHandlerRegistry) SetTaskHandlers(handlers []taskhandler.ITaskHandler) {
	registry.handlers = handlers
}

func (registry TaskHandlerRegistry) GetTaskHandler(task taskmodel.IBaseTask) (taskhandler.ITaskHandler, error) {
	var foundHandlers []taskhandler.ITaskHandler
	for _, handler := range registry.handlers {
		if handler.Handles(task) {
			foundHandlers = append(foundHandlers, handler)
		}
	}
	if len(foundHandlers) == 0 {
		return nil, errors.New("No handler found")
	}
	if len(foundHandlers) > 1 {
		return nil, errors.New("More than one handler found")
	}

	return foundHandlers[0], nil
}

func (thr TaskHandlerRegistry) GetTaskHandlers() []taskhandler.ITaskHandler {
	return thr.handlers
}

func (thr TaskHandlerRegistry) getTaskProcessingPolicy(task taskmodel.IBaseTask) processingpolicy.ITaskProcessingPolicy {
	handler, err := thr.GetTaskHandler(task)
	if err != nil {
		thr.logger.Error("Failed to get task handler", zap.Error(err))
		return nil
	}

	return handler.GetProcessingPolicy(task)
}

func (thr TaskHandlerRegistry) GetExpectedProcessingMoment(task taskmodel.IBaseTask) (time.Time, error) {
	taskProcessingPolicy := thr.getTaskProcessingPolicy(task)

	if taskProcessingPolicy == nil {
		return time.Now().Add(1 * time.Minute), nil
	}

	expectedQueueTime, exists := taskProcessingPolicy.GetExpectedQueueTime(task)

	if !exists {
		return time.Now().Add(1 * time.Minute), nil // TODO: Read from config file
	}

	return time.Now().UTC().Add(expectedQueueTime), nil
}
