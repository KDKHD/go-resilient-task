package retrypolicy

import (
	"time"

	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/model/task"
)

type NoRetryPolicy struct {
}

func NewNoRetryPolicy() *NoRetryPolicy {
	return &NoRetryPolicy{}
}

func (nrp NoRetryPolicy) GetRetryTime(task taskmodel.ITask) (bool, time.Time) {
	return false, time.Time{}
}

func (nrp NoRetryPolicy) ResetTriesCountOnSuccess(task taskmodel.ITask) bool {
	return false
}
