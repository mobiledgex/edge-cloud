package ratelimit

import "sync"

// Flow rate limiting algorithms
// Limiting the number of requests/ flow of requests through an endpoint (ie. 1 request every second).
// This is useful for preventing DDos attacks and will allow functions that require a lot of work to handle a reasonable number of requests per time frame.
type FlowLimiter struct {
	limiter Limiter
	mux     sync.Mutex
}

type FlowLimiterConfig struct {
	FlowAlgorithm     FlowRateLimitingAlgorithm
	RequestsPerSecond int
	BurstSize         int
}

type FlowRateLimitingAlgorithm int

const (
	NoFlowAlgorithm FlowRateLimitingAlgorithm = iota
	TokenBucketAlgorithm
	LeakyBucketAlgorithm
)

func NewFlowLimiter(config *FlowLimiterConfig) *FlowLimiter {
	flowLimiter := &FlowLimiter{}
	switch config.FlowAlgorithm {
	case TokenBucketAlgorithm:
		flowLimiter.limiter = NewTokenBucketLimiter(config.RequestsPerSecond, config.BurstSize)
	case LeakyBucketAlgorithm:
		flowLimiter.limiter = NewLeakyBucketLimiter(config.RequestsPerSecond)
	case NoFlowAlgorithm:
		// log
		fallthrough
	default:
		// log
		return nil
	}
	return flowLimiter
}

func (f *FlowLimiter) Limit(ctx Context) (bool, error) {
	f.mux.Lock()
	defer f.mux.Unlock()
	if f.limiter != nil {
		return f.limiter.Limit(ctx)
	}
	return false, nil
}

// TODO: Add to settings
var DefaultReqsPerSecondPerApi = 100
var DefaultTokenBucketSize = 10 // equivalent to burst size

var DefaultDmeApiFlowLimiterConfig = &FlowLimiterConfig{
	FlowAlgorithm:     TokenBucketAlgorithm,
	RequestsPerSecond: DefaultReqsPerSecondPerApi,
	BurstSize:         DefaultTokenBucketSize,
}
