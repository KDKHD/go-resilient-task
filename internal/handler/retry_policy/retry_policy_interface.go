package retrypolicy

import (
	"time"

	taskmodel "github.com/KDKHD/go-resilient-task/internal/model/task"
)

type ITaskRetryPolicy interface {
	GetRetryTime(taskmodel.ITask) (bool, time.Time)
	ResetTriesCountOnSuccess(taskmodel.ITask) bool
}
