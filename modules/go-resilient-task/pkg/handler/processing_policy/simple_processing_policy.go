package processingpolicy

import (
	"time"

	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/model/task"
)

type SinpleTaskProcessingPolicy struct {
	stuckTaskResolutionStrategy StuckTaskResolutionStrategy
	expectedQueueTime           time.Duration
}

func NewSimpleTaskProcessingPolicy(expectedQueueTime time.Duration) *SinpleTaskProcessingPolicy {
	return &SinpleTaskProcessingPolicy{stuckTaskResolutionStrategy: MARK_AS_ERROR, expectedQueueTime: expectedQueueTime}
}

func (p SinpleTaskProcessingPolicy) GetStuckTaskResolutionStrategy(task taskmodel.IStuckTask) StuckTaskResolutionStrategy {
	return p.stuckTaskResolutionStrategy
}

func (p SinpleTaskProcessingPolicy) GetExpectedQueueTime(task taskmodel.IBaseTask) (time.Duration, bool) {
	return p.expectedQueueTime, true
}
