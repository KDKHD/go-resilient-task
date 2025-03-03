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

func (t BaseTask) GetId() uuid.UUID {
	return t.Id
}

func (t BaseTask) GetVersion() int {
	return t.Version
}

func (t BaseTask) GetType() string {
	return t.Type
}

func (t BaseTask) GetPriority() int {
	return t.Priority
}

func (t *BaseTask) IncrementVersion() IBaseTask {
	t.Version++
	return t
}
