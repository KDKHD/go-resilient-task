package taskmodel

import "github.com/google/uuid"

type IBaseTask interface {
	GetId() uuid.UUID
	GetVersion() int
	GetType() string
	GetPriority() int
	IncrementVersion() IBaseTask
}

type BaseTask struct {
	Id       uuid.UUID `json:"id"`
	Type     string    `json:"type"`
	Version  int       `json:"version"`
	Priority int       `json:"priority"`
}

func NewBaseTask(id uuid.UUID, version int, taskType string, priority int) *BaseTask {
	return &BaseTask{
		Id:       id,
		Version:  version,
		Type:     taskType,
		Priority: priority,
	}
}

func (bt BaseTask) GetId() uuid.UUID {
	return bt.Id
}

func (bt BaseTask) GetVersion() int {
	return bt.Version
}

func (bt BaseTask) GetType() string {
	return bt.Type
}

func (bt BaseTask) GetPriority() int {
	return bt.Priority
}

func (bt *BaseTask) IncrementVersion() IBaseTask {
	bt.Version++
	return bt
}
