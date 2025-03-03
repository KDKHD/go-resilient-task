package taskexecutiontrigger

import taskmodel "github.com/KDKHD/go-resilient-task/internal/model/task"

type ITasksExecutionTriggerer interface {
	Trigger(taskmodel.IBaseTask) error
	StartTasksProcessing() error
	StopTasksProcessing() error
}
