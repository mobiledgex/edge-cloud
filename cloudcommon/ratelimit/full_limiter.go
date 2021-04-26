package ratelimit

// FullLimiter
// Includes limiter on MaxReqs and Flow
type FullLimiter struct {
	flowLimiter    *FlowLimiter
	maxReqsLimiter *MaxReqsLimiter
}

type FullLimiterConfig struct {
	MaxReqsAlgorithm  MaxReqsRateLimitingAlgorithm
	FlowAlgorithm     FlowRateLimitingAlgorithm
	MaxReqsPerSecond  int
	MaxReqsPerMinute  int
	MaxReqsPerHour    int
	RequestsPerSecond int
	BurstSize         int
}

func NewFullLimiter(config *FullLimiterConfig) *FullLimiter {
	fullLimiter := &FullLimiter{}
	flowConfig := &FlowLimiterConfig{
		FlowAlgorithm:     config.FlowAlgorithm,
		RequestsPerSecond: config.RequestsPerSecond,
		BurstSize:         config.BurstSize,
	}
	fullLimiter.flowLimiter = NewFlowLimiter(flowConfig)
	maxReqsConfig := &MaxReqsLimiterConfig{
		MaxReqsAlgorithm: config.MaxReqsAlgorithm,
		MaxReqsPerSecond: config.MaxReqsPerSecond,
		MaxReqsPerMinute: config.MaxReqsPerMinute,
		MaxReqsPerHour:   config.MaxReqsPerHour,
	}
	fullLimiter.maxReqsLimiter = NewMaxReqsLimiter(maxReqsConfig)
	return fullLimiter
}

func (f *FullLimiter) Limit(ctx Context) (bool, error) {
	if f.flowLimiter != nil {
		return f.flowLimiter.Limit(ctx)
	}
	if f.maxReqsLimiter != nil {
		return f.maxReqsLimiter.Limit(ctx)
	}
	return false, nil
}

var DefaultDmeApiPerIpFullLimiterConfig = &FullLimiterConfig{
	MaxReqsAlgorithm:  NoMaxReqsAlgorithm,
	FlowAlgorithm:     TokenBucketAlgorithm,
	RequestsPerSecond: 5,
	BurstSize:         1,
}

var DefaultDmeApiPerUserFullLimiterConfig = &FullLimiterConfig{
	MaxReqsAlgorithm:  NoMaxReqsAlgorithm,
	FlowAlgorithm:     TokenBucketAlgorithm,
	RequestsPerSecond: 5,
	BurstSize:         1,
}

var DefaultDmeApiPerOrgFullLimiterConfig = &FullLimiterConfig{
	MaxReqsAlgorithm:  NoMaxReqsAlgorithm,
	FlowAlgorithm:     TokenBucketAlgorithm,
	RequestsPerSecond: 50,
	BurstSize:         5,
}
