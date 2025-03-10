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

func (st StuckTask) GetStatus() TaskStatus {
	return st.Status
}

func (st *StuckTask) SetStatus(status TaskStatus) IBaseTask {
	st.Status = status
	return st
}
