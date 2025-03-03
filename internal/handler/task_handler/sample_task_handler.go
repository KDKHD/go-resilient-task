package taskhandler

import (
	"time"

	concurrencypolicy "github.com/KDKHD/go-resilient-task/internal/handler/concurrency_policy"
	processingpolicy "github.com/KDKHD/go-resilient-task/internal/handler/processing_policy"
	retrypolicy "github.com/KDKHD/go-resilient-task/internal/handler/retry_policy"
	taskprocessor "github.com/KDKHD/go-resilient-task/internal/handler/task_processor"
	taskmodel "github.com/KDKHD/go-resilient-task/internal/model/task"
	"go.uber.org/zap"
)

type TaskHandler struct {
	taskProcessor        taskprocessor.ITaskProcessor
	logger               *zap.Logger
	handles              func(task taskmodel.IBaseTask) bool
	concurrencypolicy    concurrencypolicy.ITaskConcurrencyPolicy
	taskProcessingPolicy processingpolicy.ITaskProcessingPolicy
}

type TaskProcessor struct {
	logger  *zap.Logger
	process func(task taskmodel.ITask)
}

type TaskRetryPolicy struct {
	logger *zap.Logger
}

func NewTaskProcessor(logger *zap.Logger, process func(task taskmodel.ITask)) *TaskProcessor {
	return &TaskProcessor{
		logger:  logger,
		process: process,
	}
}

func NewTaskRetryPolicy(logger *zap.Logger) *TaskRetryPolicy {
	return &TaskRetryPolicy{
		logger: logger,
	}
}

func NewTaskHandler(taskProcessor taskprocessor.ITaskProcessor, logger *zap.Logger, handles func(task taskmodel.IBaseTask) bool) *TaskHandler {
	return &TaskHandler{
		taskProcessor:        taskProcessor,
		logger:               logger,
		handles:              handles,
		concurrencypolicy:    concurrencypolicy.NewSimpleTaskConcurrencyPolicy(1000, logger),
		taskProcessingPolicy: processingpolicy.NewSimpleTaskProcessingPolicy(time.Minute * 20),
	}
}

func (handler TaskHandler) GetProcessingPolicy(task taskmodel.IBaseTask) processingpolicy.ITaskProcessingPolicy {
	return handler.taskProcessingPolicy
}

// GetProcessor returns the processor for the given task

func (handler TaskHandler) GetProcessor(task taskmodel.IBaseTask) taskprocessor.ITaskProcessor {
	return handler.taskProcessor
}

// GetRetryPolicy returns the retry policy for the given task
func (handler TaskHandler) GetRetryPolicy(task taskmodel.IBaseTask) retrypolicy.ITaskRetryPolicy {
	// 5s 20s 1m20s 5m20s 20m
	return retrypolicy.NewExponentialRetryPolicy(
		retrypolicy.WithDelay(5*time.Second),
		retrypolicy.WithMultiplier(4),
		retrypolicy.WithMaxCount(3),
		retrypolicy.WithMaxDelay(20*time.Minute),
	)
}

func (handler TaskHandler) GetConcurrencyPolicy() concurrencypolicy.ITaskConcurrencyPolicy {
	return handler.concurrencypolicy
}

// Handles returns true if the handler can handle the given task

func (handler TaskHandler) Handles(task taskmodel.IBaseTask) bool {
	return handler.handles(task)
}

/* func (handler TaskHandler) GetProcessingPolicy(task taskmodel.IBaseTask) ITaskProcessingPolicy {
	return NewSimpleTaskProcessingPolicy()
} */

// GetRetryTime returns the time when the task should be retried

func (retryPolicy TaskRetryPolicy) GetRetryTime(task taskmodel.ITask) time.Time {
	return time.Now().UTC()
}

// Process processes the task

func (processor TaskProcessor) Process(task taskmodel.ITask) (taskprocessor.ProcessResult, error) {

	processor.process(task)

	processor.logger.Debug("Processing task start", zap.String("data", task.GetData()))
	// Generate a random number between 1 and 10

	time.Sleep(5 * time.Second)

	processor.logger.Debug("Processing task finish", zap.String("data", task.GetData()))

	//return taskprocessor.ProcessResult{}, errors.New("error")

	return taskprocessor.ProcessResult{
		ResultCode: taskprocessor.DONE,
	}, nil
}
