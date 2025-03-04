package taskhandler

import (
	"time"

	concurrencypolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/concurrency_policy"
	processingpolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/processing_policy"
	retrypolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/retry_policy"
	taskprocessor "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/task_processor"
	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/model/task"
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
	logger *zap.Logger
}

type TaskRetryPolicy struct {
	logger *zap.Logger
}

func NewTaskProcessor(logger *zap.Logger) *TaskProcessor {
	return &TaskProcessor{
		logger: logger,
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

func (th TaskHandler) GetProcessingPolicy(task taskmodel.IBaseTask) processingpolicy.ITaskProcessingPolicy {
	return th.taskProcessingPolicy
}

func (th TaskHandler) GetProcessor(task taskmodel.IBaseTask) taskprocessor.ITaskProcessor {
	return th.taskProcessor
}

func (th TaskHandler) GetRetryPolicy(task taskmodel.IBaseTask) retrypolicy.ITaskRetryPolicy {
	// 5s 20s 1m20s 5m20s 20m
	return retrypolicy.NewExponentialRetryPolicy(
		retrypolicy.WithDelay(5*time.Second),
		retrypolicy.WithMultiplier(4),
		retrypolicy.WithMaxCount(3),
		retrypolicy.WithMaxDelay(20*time.Minute),
	)
}

func (th TaskHandler) GetConcurrencyPolicy() concurrencypolicy.ITaskConcurrencyPolicy {
	return th.concurrencypolicy
}

func (th TaskHandler) Handles(task taskmodel.IBaseTask) bool {
	return th.handles(task)
}

func (trp TaskRetryPolicy) GetRetryTime(task taskmodel.ITask) time.Time {
	return time.Now().UTC()
}

func (tp TaskProcessor) Process(task taskmodel.ITask) (taskprocessor.ProcessResult, error) {

	tp.logger.Debug("Processing task start", zap.String("data", task.GetData()))

	time.Sleep(5 * time.Second)

	tp.logger.Debug("Processing task finish", zap.String("data", task.GetData()))

	//return taskprocessor.ProcessResult{}, errors.New("error")

	return taskprocessor.ProcessResult{
		ResultCode: taskprocessor.DONE,
	}, nil
}
