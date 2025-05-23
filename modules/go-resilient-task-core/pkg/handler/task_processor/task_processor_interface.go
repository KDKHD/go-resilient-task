//go:generate stringer -type=ProcessResultCode

package taskprocessor

import taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/model/task"

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

type TaskProcessorFunc func(task taskmodel.ITask) (ProcessResult, error)

func (f TaskProcessorFunc) Process(task taskmodel.ITask) (ProcessResult, error) {
	return f(task)
}

func (pr ProcessResult) GetResultCode() ProcessResultCode {
	return pr.ResultCode
}
