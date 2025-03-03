package taskmodel

import (
	"github.com/google/uuid"
)

type ITask interface {
	IBaseTask
	GetSubType() string
	GetData() string
	GetStatus() TaskStatus
	GetProcessingTriesCount() int
	IncrementVersion() IBaseTask
	SetProcessingTriesCount(count int) IBaseTask
}

type Task struct {
	Id                   uuid.UUID  `json:"id"`
	Type                 string     `json:"type"`
	SubType              string     `json:"type"`
	Data                 []byte     `json:"data"`
	Status               TaskStatus `json:"status"`
	Version              int        `json:"version"`
	ProcessingTriesCount int        `json:"processingTriesCount"`
	Priority             int        `json:"priority"`
}

func (t Task) GetId() uuid.UUID {
	return t.Id
}

func (t Task) GetVersion() int {
	return t.Version
}

func (t Task) GetType() string {
	return t.Type
}

func (t Task) GetSubType() string {
	return t.SubType
}

func (t Task) GetPriority() int {
	return t.Priority
}

func (t Task) GetData() string {
	return string(t.Data[:])
}

func (t Task) GetStatus() TaskStatus {
	return t.Status
}

func (t Task) GetProcessingTriesCount() int {
	return t.ProcessingTriesCount
}

func (t *Task) IncrementVersion() IBaseTask {
	t.Version++
	return t
}

func (t *Task) SetProcessingTriesCount(count int) IBaseTask {
	t.ProcessingTriesCount = count
	return t
}
