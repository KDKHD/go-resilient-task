package processingmodel

type ProcessingResponseResult int

const (
	OK ProcessingResponseResult = iota
	FULL
)

type ProcessingResponse struct {
	result ProcessingResponseResult
}

func NewProcessingResponse(result ProcessingResponseResult) *ProcessingResponse {
	return &ProcessingResponse{
		result: result,
	}
}
