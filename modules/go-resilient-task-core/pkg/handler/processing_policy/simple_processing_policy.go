package processingpolicy

import (
	"time"

	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/model/task"
)

type SimpleTaskProcessingPolicy struct {
	stuckTaskResolutionStrategy StuckTaskResolutionStrategy
	expectedQueueTime           time.Duration
}

func NewSimpleTaskProcessingPolicy(expectedQueueTime time.Duration) *SimpleTaskProcessingPolicy {
	return &SimpleTaskProcessingPolicy{stuckTaskResolutionStrategy: MARK_AS_ERROR, expectedQueueTime: expectedQueueTime}
}

func (stpp SimpleTaskProcessingPolicy) GetStuckTaskResolutionStrategy(task taskmodel.IStuckTask) StuckTaskResolutionStrategy {
	return stpp.stuckTaskResolutionStrategy
}

func (stpp SimpleTaskProcessingPolicy) GetExpectedQueueTime(task taskmodel.IBaseTask) (time.Duration, bool) {
	return stpp.expectedQueueTime, true
}
