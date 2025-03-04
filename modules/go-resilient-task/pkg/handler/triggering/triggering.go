package triggering

import (
	"time"

	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/model/task"
)

type TaskTriggering struct {
	task        taskmodel.IBaseTask
	triggeredAt time.Time
}

func NewTaskTriggering(task taskmodel.IBaseTask, triggeredAt time.Time) *TaskTriggering {
	return &TaskTriggering{
		task:        task,
		triggeredAt: triggeredAt,
	}
}

func (tt *TaskTriggering) GetTask() taskmodel.IBaseTask {
	return tt.task
}

func (tt *TaskTriggering) GetTriggeredAt() time.Time {
	return tt.triggeredAt
}
