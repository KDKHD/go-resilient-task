package retrypolicy

import (
	"time"

	taskmodel "github.com/KDKHD/go-resilient-task/internal/model/task"
)

type NoRetryPolicy struct {
}

func NewNoRetryPolicy() *NoRetryPolicy {
	return &NoRetryPolicy{}
}

func (p NoRetryPolicy) GetRetryTime(task taskmodel.ITask) (bool, time.Time) {
	return false, time.Time{}
}

func (p NoRetryPolicy) ResetTriesCountOnSuccess(task taskmodel.ITask) bool {
	return false
}
