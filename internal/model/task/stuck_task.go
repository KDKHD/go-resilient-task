package taskmodel

import "github.com/google/uuid"

type IStuckTask interface {
	IBaseTask
	GetStatus() TaskStatus
}

type StuckTask struct {
	BaseTask
	Status TaskStatus `json:"status"`
}

func NewStuckTask(id uuid.UUID, version int, taskType string, priority int, status TaskStatus) *StuckTask {
	return &StuckTask{
		BaseTask: BaseTask{
			Id:       id,
			Version:  version,
			Type:     taskType,
			Priority: priority,
		},
		Status: status,
	}
}

func (t StuckTask) GetStatus() TaskStatus {
	return t.Status
}

func (t *StuckTask) SetStatus(status TaskStatus) IBaseTask {
	t.Status = status
	return t
}
