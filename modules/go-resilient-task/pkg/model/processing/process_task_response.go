package processingmodel

import "time"

type ProcessTaskResponseResult int

const (
	OK_processTaskResponseResult ProcessTaskResponseResult = iota
	NO_SPACE_processTaskResponseResult
	ERROR_processTaskResponseResult
)

type ProcessTaskResponseCode int

const (
	NO_HANDLER_processTaskResponseCode ProcessTaskResponseCode = iota
	NO_POLICY_processTaskResponseCode
	NO_CONCURRENCY_POLICY_processTaskResponseCode
	UNKNOWN_ERROR_processTaskResponseCode
	HAPPY_FLOW_processTaskResponseCode
	NOT_ALLOWED_ON_NODE_processTaskResponseCode
)

type ProcessTaskResponse struct {
	result       ProcessTaskResponseResult
	code         ProcessTaskResponseCode
	tryAgainTime time.Time
}

func NewProcessTaskResponse(result ProcessTaskResponseResult, code ProcessTaskResponseCode, tryAgainTime time.Time) *ProcessTaskResponse {
	return &ProcessTaskResponse{
		result:       result,
		code:         code,
		tryAgainTime: tryAgainTime,
	}
}

func (r ProcessTaskResponse) GetResult() ProcessTaskResponseResult {
	return r.result
}

func (r ProcessTaskResponse) GetCode() ProcessTaskResponseCode {
	return r.code
}

func (r ProcessTaskResponse) GetTryAgainTime() time.Time {
	return r.tryAgainTime
}
