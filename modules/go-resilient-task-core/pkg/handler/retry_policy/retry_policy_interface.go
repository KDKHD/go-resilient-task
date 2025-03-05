package retrypolicy

import (
	"time"

	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/model/task"
)

type ITaskRetryPolicy interface {
	GetRetryTime(taskmodel.ITask) (bool, time.Time)
	ResetTriesCountOnSuccess(taskmodel.ITask) bool
}
