package ratelimit

import (
	"context"
)

// FullLimiter
// Includes a MaxReqsLimiter and a FlowLimiter
type FullLimiter struct {
	flowLimiter    *FlowLimiter
	maxReqsLimiter *MaxReqsLimiter
}

func NewFullLimiter(settings *RateLimitSettings) *FullLimiter {
	fullLimiter := &FullLimiter{}
	if settings == nil {
		return fullLimiter
	}
	// Get flowLimiter
	flowRateLimitSettings := &FlowRateLimitSettings{
		FlowAlgorithm: settings.FlowAlgorithm,
		ReqsPerSecond: settings.ReqsPerSecond,
		BurstSize:     settings.BurstSize,
	}
	fullLimiter.flowLimiter = NewFlowLimiter(flowRateAlgorithm)
	// Get maxReqsLimiter
	maxReqsRateLimitSettings := &MaxReqsRateLimitSettings{
		MaxReqsAlgorithm:     settings.MaxReqsAlgorithm,
		MaxRequestsPerSecond: settings.MaxRequestsPerSecond,
		MaxRequestsPerMinute: settings.MaxRequestsPerMinute,
		MaxRequestsPerHour:   settings.MaxRequestsPerHour,
	}
	fullLimiter.maxReqsLimiter = NewMaxReqsLimiter(maxRequsAlgorithm)
	return fullLimiter
}

func (f *FullLimiter) Limit(ctx context.Context) (bool, error) {
	if f.flowLimiter != nil {
		limit, err := f.flowLimiter.Limit(ctx)
		if limit {
			return limit, err
		}
	}
	if f.maxReqsLimiter != nil {
		return f.maxReqsLimiter.Limit(ctx)
	}
	return false, nil
}
