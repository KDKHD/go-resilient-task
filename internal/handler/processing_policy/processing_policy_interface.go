package processingpolicy

import (
	"time"

	taskmodel "github.com/KDKHD/go-resilient-task/internal/model/task"
)

type StuckTaskResolutionStrategy int

const (
	RETRY StuckTaskResolutionStrategy = iota
	MARK_AS_ERROR
	MARK_AS_FAILED
	IGNORE
	NOOP
)

type ITaskProcessingPolicy interface {
	GetStuckTaskResolutionStrategy(task taskmodel.IStuckTask) StuckTaskResolutionStrategy
	GetExpectedQueueTime(task taskmodel.IBaseTask) (time.Duration, bool)
}
