package concurrencypolicy

import (
	"slices"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type SimpleTaskConcurrencyPolicy struct {
	maxConcurrency   int32
	inProgressCnt    *atomic.Int32
	maxInProgressCnt int
	logger           *zap.Logger
}

func NewSimpleTaskConcurrencyPolicy(maxConcurrency int32, logger *zap.Logger) *SimpleTaskConcurrencyPolicy {
	return &SimpleTaskConcurrencyPolicy{
		maxConcurrency: maxConcurrency,
		inProgressCnt:  &atomic.Int32{},
		logger:         logger,
	}
}

func (stcp *SimpleTaskConcurrencyPolicy) GetMaxConcurrency() int32 {
	return stcp.maxConcurrency
}

func (stcp *SimpleTaskConcurrencyPolicy) GetInProgressCnt() *atomic.Int32 {
	return stcp.inProgressCnt
}

func (stcp *SimpleTaskConcurrencyPolicy) BookSpace() *BookSpaceResponse {
	cnt := stcp.inProgressCnt.Add(1)
	stcp.logger.Debug("BookSpace", zap.Int32("inProgressCnt", cnt))
	if cnt > stcp.maxConcurrency {
		stcp.inProgressCnt.Add(-1)
		return NewBookSpaceResponse(false, time.Time{})
	}

	stcp.maxInProgressCnt = slices.Max([]int{stcp.maxInProgressCnt, int(cnt)})
	return NewBookSpaceResponse(true, time.Time{})
}

func (stcp *SimpleTaskConcurrencyPolicy) FreeSpace() {
	if stcp.inProgressCnt.Add(-1) < 0 {
		stcp.logger.Error("Negative inProgressCnt", zap.Int32("inProgressCnt", stcp.inProgressCnt.Load()))
	}
}
