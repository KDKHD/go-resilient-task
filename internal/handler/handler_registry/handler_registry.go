package handlerregistry

import (
	"errors"
	"time"

	processingpolicy "github.com/KDKHD/go-resilient-task/internal/handler/processing_policy"
	taskhandler "github.com/KDKHD/go-resilient-task/internal/handler/task_handler"
	taskmodel "github.com/KDKHD/go-resilient-task/internal/model/task"
	"go.uber.org/zap"
)

type ITaskHandlerRegistry interface {
	GetTaskHandler(task taskmodel.IBaseTask) (taskhandler.ITaskHandler, error)
	GetTaskHandlers() []taskhandler.ITaskHandler
	GetExpectedProcessingMoment(task taskmodel.IBaseTask) (time.Time, error)
}

type TaskHandlerRegistry struct {
	handlers []taskhandler.ITaskHandler
	logger   *zap.Logger
}

type TaskHandlerRegistryConfig struct {
	handlers []taskhandler.ITaskHandler
	logger   *zap.Logger
}

func WithLogger(logger *zap.Logger) func(*TaskHandlerRegistryConfig) {
	return func(config *TaskHandlerRegistryConfig) {
		config.logger = logger
	}
}

func WithHandler(handler taskhandler.ITaskHandler) func(*TaskHandlerRegistryConfig) {
	return func(config *TaskHandlerRegistryConfig) {
		config.handlers = append(config.handlers, handler)
	}
}

func NewTaskHandlerRegistry(configs ...func(*TaskHandlerRegistryConfig)) *TaskHandlerRegistry {
	registryConfig := &TaskHandlerRegistryConfig{}
	for _, config := range configs {
		config(registryConfig)
	}
	return &TaskHandlerRegistry{handlers: registryConfig.handlers, logger: registryConfig.logger}
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

func (registry TaskHandlerRegistry) GetTaskHandlers() []taskhandler.ITaskHandler {
	return registry.handlers
}

func (registry TaskHandlerRegistry) getTaskProcessingPolicy(task taskmodel.IBaseTask) processingpolicy.ITaskProcessingPolicy {
	handler, err := registry.GetTaskHandler(task)
	if err != nil {
		registry.logger.Error("Failed to get task handler", zap.Error(err))
		return nil
	}

	return handler.GetProcessingPolicy(task)
}

func (registry TaskHandlerRegistry) GetExpectedProcessingMoment(task taskmodel.IBaseTask) (time.Time, error) {
	taskProcessingPolicy := registry.getTaskProcessingPolicy(task)

	if taskProcessingPolicy == nil {
		return time.Now().Add(1 * time.Minute), nil
	}

	expectedQueueTime, exists := taskProcessingPolicy.GetExpectedQueueTime(task)

	if !exists {
		return time.Now().Add(1 * time.Minute), nil // TODO: Read from config file
	}

	return time.Now().UTC().Add(expectedQueueTime), nil
}
