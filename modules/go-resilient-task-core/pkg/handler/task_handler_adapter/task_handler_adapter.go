package taskhandleradapter

import (
	concurrencypolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/concurrency_policy"
	processingpolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/processing_policy"
	retrypolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/retry_policy"
	taskprocessor "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_processor"
	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/model/task"
	u "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/utils"
)

type TaskHandlerAdapter struct {
	handlesPredicate  u.Predicate[taskmodel.IBaseTask]
	processor         taskprocessor.ITaskProcessor
	retryPolicy       retrypolicy.ITaskRetryPolicy
	concurrencyPolicy concurrencypolicy.ITaskConcurrencyPolicy
	processingPolicy  processingpolicy.ITaskProcessingPolicy
}

type TaskHandlerAdapterBuilder struct {
	handlesPredicate  u.Predicate[taskmodel.IBaseTask]
	processor         taskprocessor.ITaskProcessor
	retryPolicy       retrypolicy.ITaskRetryPolicy
	concurrencyPolicy concurrencypolicy.ITaskConcurrencyPolicy
	processingPolicy  processingpolicy.ITaskProcessingPolicy
}

func NewTaskHandlerAdapterBuilder(predicate u.Predicate[taskmodel.IBaseTask], processor taskprocessor.ITaskProcessor) *TaskHandlerAdapterBuilder {
	return &TaskHandlerAdapterBuilder{
		handlesPredicate: predicate,
		processor:        processor,
	}
}

func (thab *TaskHandlerAdapterBuilder) WithRetryPolicy(retryPolicy retrypolicy.ITaskRetryPolicy) *TaskHandlerAdapterBuilder {
	thab.retryPolicy = retryPolicy
	return thab
}

func (thab *TaskHandlerAdapterBuilder) WithConcurrencyPolicy(concurrencyPolicy concurrencypolicy.ITaskConcurrencyPolicy) *TaskHandlerAdapterBuilder {
	thab.concurrencyPolicy = concurrencyPolicy
	return thab
}

func (thab *TaskHandlerAdapterBuilder) WithProcessingPolicy(processingPolicy processingpolicy.ITaskProcessingPolicy) *TaskHandlerAdapterBuilder {
	thab.processingPolicy = processingPolicy
	return thab
}

func (thab *TaskHandlerAdapterBuilder) Build() *TaskHandlerAdapter {
	return &TaskHandlerAdapter{
		handlesPredicate:  thab.handlesPredicate,
		processor:         thab.processor,
		retryPolicy:       thab.retryPolicy,
		concurrencyPolicy: thab.concurrencyPolicy,
		processingPolicy:  thab.processingPolicy,
	}
}

func (tha TaskHandlerAdapter) Handles(task taskmodel.IBaseTask) bool {
	return tha.handlesPredicate(task)
}

func (tha TaskHandlerAdapter) GetProcessor(task taskmodel.IBaseTask) taskprocessor.ITaskProcessor {
	return tha.processor
}

func (tha TaskHandlerAdapter) GetRetryPolicy(task taskmodel.IBaseTask) retrypolicy.ITaskRetryPolicy {
	return tha.retryPolicy
}

func (tha TaskHandlerAdapter) GetConcurrencyPolicy() concurrencypolicy.ITaskConcurrencyPolicy {
	return tha.concurrencyPolicy
}

func (tha TaskHandlerAdapter) GetProcessingPolicy(task taskmodel.IBaseTask) processingpolicy.ITaskProcessingPolicy {
	return tha.processingPolicy
}
