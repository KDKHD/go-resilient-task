//go:generate stringer -type=ProcessResultCode

package taskprocessor

import taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/model/task"

type ProcessResultCode int

const (
	DONE ProcessResultCode = iota
	DONE_AND_DELETE
	COMMIT_AND_RETRY
)

type ProcessResult struct {
	ResultCode ProcessResultCode
}

type ITaskProcessor interface {
	Process(task taskmodel.ITask) (ProcessResult, error)
}

func (pr ProcessResult) GetResultCode() ProcessResultCode {
	return pr.ResultCode
}
