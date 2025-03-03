package retrypolicy

import (
	"math"
	"time"

	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/model/task"
)

type ExponentialRetryPolicy struct {
	triesDelta int
	maxCount   int
	maxDelay   time.Duration
	muliplier  int
	delay      time.Duration
}

type ExponentialRetryPolicyConfig struct {
	triesDelta int
	maxCount   int
	maxDelay   time.Duration
	muliplier  int
	delay      time.Duration
}

// Functional option type
type Option func(*ExponentialRetryPolicyConfig)

func WithTriesDelta(triesDelta int) Option {
	return func(c *ExponentialRetryPolicyConfig) {
		c.triesDelta = triesDelta
	}
}

func WithMaxCount(maxCount int) Option {
	return func(c *ExponentialRetryPolicyConfig) {
		c.maxCount = maxCount
	}
}

func WithMaxDelay(maxDelay time.Duration) Option {
	return func(c *ExponentialRetryPolicyConfig) {
		c.maxDelay = maxDelay
	}
}

func WithMultiplier(muliplier int) Option {
	return func(c *ExponentialRetryPolicyConfig) {
		c.muliplier = muliplier
	}
}

func WithDelay(delay time.Duration) Option {
	return func(c *ExponentialRetryPolicyConfig) {
		c.delay = delay
	}
}

func NewExponentialRetryPolicy(opts ...Option) *ExponentialRetryPolicy {
	config := ExponentialRetryPolicyConfig{
		triesDelta: 0,
		maxCount:   1,
		maxDelay:   time.Hour * 24,
		muliplier:  2,
		delay:      time.Duration(0),
	}

	for _, opt := range opts {
		opt(&config)
	}

	return &ExponentialRetryPolicy{

		triesDelta: config.triesDelta,
		maxCount:   config.maxCount,
		maxDelay:   config.maxDelay,
		muliplier:  config.muliplier,
		delay:      config.delay,
	}
}

func (p ExponentialRetryPolicy) GetRetryTime(task taskmodel.ITask) (bool, time.Time) {
	triesCount := task.GetProcessingTriesCount() + p.triesDelta
	if triesCount > p.maxCount {
		return false, time.Time{}
	}

	var addedTime time.Duration
	if triesCount > math.MaxInt {
		addedTime = p.maxDelay
	} else {
		addedTime = p.delay * time.Duration(math.Pow(float64(p.muliplier), float64(triesCount-1)))
		if addedTime > p.maxDelay {
			addedTime = p.maxDelay
		}
	}

	return true, time.Now().Add(addedTime).UTC()
}

func (p ExponentialRetryPolicy) ResetTriesCountOnSuccess(task taskmodel.ITask) bool {
	return false
}
