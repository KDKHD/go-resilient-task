package taskproperties

import "time"

type ITaskResumerProperties interface {
	GetBatchSize() int
	GetConcurrency() int
	GetPollingInterval() time.Duration
}

type TaskResumerProperties struct {
	batchSize       int
	concurrency     int
	pollingInterval time.Duration
}

type TaskResumerPropertiesConfigOption func(*TaskResumerProperties)

func WithBatchSize(batchSize int) TaskResumerPropertiesConfigOption {
	return func(config *TaskResumerProperties) {
		config.batchSize = batchSize
	}
}

func WithConcurrency(concurrency int) TaskResumerPropertiesConfigOption {
	return func(config *TaskResumerProperties) {
		config.concurrency = concurrency
	}
}

func WithPollingInterval(pollingInterval time.Duration) TaskResumerPropertiesConfigOption {
	return func(config *TaskResumerProperties) {
		config.pollingInterval = pollingInterval
	}
}

func NewTaskResumerProperties(options ...TaskResumerPropertiesConfigOption) *TaskResumerProperties {
	config := &TaskResumerProperties{
		batchSize:       1000,
		concurrency:     1,
		pollingInterval: 2 * time.Second,
	}
	for _, option := range options {
		option(config)
	}

	return config
}

func (trp TaskResumerProperties) GetBatchSize() int {
	return trp.batchSize
}

func (trp TaskResumerProperties) GetConcurrency() int {
	return trp.concurrency
}

func (trp TaskResumerProperties) GetPollingInterval() time.Duration {
	return trp.pollingInterval
}
