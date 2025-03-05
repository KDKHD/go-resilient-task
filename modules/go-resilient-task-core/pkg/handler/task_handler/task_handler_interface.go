package taskhandler

import (
	concurrencypolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/concurrency_policy"
	processingpolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/processing_policy"
	retrypolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/retry_policy"
	taskprocessor "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_processor"
	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/model/task"
)

type ITaskHandler interface {
	GetProcessor(taskmodel.IBaseTask) taskprocessor.ITaskProcessor
	GetRetryPolicy(taskmodel.IBaseTask) retrypolicy.ITaskRetryPolicy
	Handles(taskmodel.IBaseTask) bool
	GetConcurrencyPolicy() concurrencypolicy.ITaskConcurrencyPolicy
	GetProcessingPolicy(taskmodel.IBaseTask) processingpolicy.ITaskProcessingPolicy
}
