package triggering

import (
	"time"

	taskmodel "github.com/KDKHD/go-resilient-task/internal/model/task"
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

func (t *TaskTriggering) GetTask() taskmodel.IBaseTask {
	return t.task
}

func (t *TaskTriggering) GetTriggeredAt() time.Time {
	return t.triggeredAt
}
