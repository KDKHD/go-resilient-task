package concurrencypolicy

import (
	"time"
)

type BookSpaceResponse struct {
	hasRoom      bool
	tryAgainTime time.Time
}

func NewBookSpaceResponse(hasRoom bool, tryAgainTime time.Time) *BookSpaceResponse {
	return &BookSpaceResponse{hasRoom: hasRoom, tryAgainTime: tryAgainTime}
}

func (bsr BookSpaceResponse) HasRoom() bool {
	return bsr.hasRoom
}

func (bsr BookSpaceResponse) GetTryAgainTime() time.Time {
	return bsr.tryAgainTime
}

type ITaskConcurrencyPolicy interface {
	BookSpace() *BookSpaceResponse
	FreeSpace()
}
