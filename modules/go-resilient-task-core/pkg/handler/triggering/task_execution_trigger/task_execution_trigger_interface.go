package taskexecutiontrigger

import taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/model/task"

type ITasksExecutionTriggerer interface {
	Trigger(taskmodel.IBaseTask) error
	StartTasksProcessing() error
	StopTasksProcessing() error
}
