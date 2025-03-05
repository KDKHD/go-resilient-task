package taskproperties

import "time"

type ITaskProperties interface {
	GetTaskStuckTimeout() time.Duration
	GetTaskResumer() ITaskResumerProperties
	GetKafka() IKafkaProperties
}

type TaskProperties struct {
	taskStuckTimeout time.Duration
	taskResumer      ITaskResumerProperties
	kafka            IKafkaProperties
}

type TaskPropertiesConfigOption func(*TaskProperties)

func WithTaskStuckTimeout(taskStuckTimeout time.Duration) TaskPropertiesConfigOption {
	return func(config *TaskProperties) {
		config.taskStuckTimeout = taskStuckTimeout
	}
}

func WithTaskResumer(taskResumer ITaskResumerProperties) TaskPropertiesConfigOption {
	return func(config *TaskProperties) {
		config.taskResumer = taskResumer
	}
}

func WithKafka(kafka IKafkaProperties) TaskPropertiesConfigOption {
	return func(config *TaskProperties) {
		config.kafka = kafka
	}
}

func NewTaskProperties(options ...TaskPropertiesConfigOption) *TaskProperties {
	config := &TaskProperties{
		taskStuckTimeout: time.Minute * 5,
		taskResumer:      NewTaskResumerProperties(),
		kafka:            NewKafkaProperties(),
	}
	for _, option := range options {
		option(config)
	}

	return config
}

func (tp TaskProperties) GetTaskStuckTimeout() time.Duration {
	return tp.taskStuckTimeout
}

func (tp TaskProperties) GetTaskResumer() ITaskResumerProperties {
	return tp.taskResumer
}

func (tp TaskProperties) GetKafka() IKafkaProperties {
	return tp.kafka
}
