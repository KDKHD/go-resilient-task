package taskproperties

import "time"

type ITaskProperties interface {
	GetTaskStuckTimeout() time.Duration
}

type TaskProperties struct {
	taskStuckTimeout time.Duration
}

type TaskPropertiesConfigOption func(*TaskProperties)

func WithTaskStuckTimeout(taskStuckTimeout time.Duration) TaskPropertiesConfigOption {
	return func(config *TaskProperties) {
		config.taskStuckTimeout = taskStuckTimeout
	}
}

func NewTaskProperties(options ...TaskPropertiesConfigOption) *TaskProperties {
	config := &TaskProperties{}
	for _, option := range options {
		option(config)
	}

	if config.taskStuckTimeout == 0 {
		panic("TaskStuckTimeout is required")
	}

	return config
}

func (tp TaskProperties) GetTaskStuckTimeout() time.Duration {
	return tp.taskStuckTimeout
}
