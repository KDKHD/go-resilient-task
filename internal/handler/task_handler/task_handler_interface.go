package taskhandler

import (
	concurrencypolicy "github.com/KDKHD/go-resilient-task/internal/handler/concurrency_policy"
	processingpolicy "github.com/KDKHD/go-resilient-task/internal/handler/processing_policy"
	retrypolicy "github.com/KDKHD/go-resilient-task/internal/handler/retry_policy"
	processor "github.com/KDKHD/go-resilient-task/internal/handler/task_processor"
	taskmodel "github.com/KDKHD/go-resilient-task/internal/model/task"
)

type ITaskHandler interface {
	GetProcessor(taskmodel.IBaseTask) processor.ITaskProcessor
	GetRetryPolicy(taskmodel.IBaseTask) retrypolicy.ITaskRetryPolicy
	Handles(taskmodel.IBaseTask) bool
	GetConcurrencyPolicy() concurrencypolicy.ITaskConcurrencyPolicy
	GetProcessingPolicy(taskmodel.IBaseTask) processingpolicy.ITaskProcessingPolicy
}
