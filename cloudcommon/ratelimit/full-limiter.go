package ratelimit

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// FullLimiter
// Includes a MaxReqsLimiter and a FlowLimiter
type FullLimiter struct {
	flowLimiter    *FlowLimiter
	maxReqsLimiter *MaxReqsLimiter
}

func NewFullLimiter(settings *edgeproto.RateLimitSettings) *FullLimiter {
	if settings == nil {
		return nil
	}
	fullLimiter := &FullLimiter{}
	// Get flowLimiter
	flowRateLimitSettings := &FlowRateLimitSettings{
		FlowAlgorithm: settings.FlowAlgorithm,
		ReqsPerSecond: settings.ReqsPerSecond,
		BurstSize:     int(settings.BurstSize),
	}
	fullLimiter.flowLimiter = NewFlowLimiter(flowRateLimitSettings)
	// Get maxReqsLimiter
	maxReqsRateLimitSettings := &MaxReqsRateLimitSettings{
		MaxReqsAlgorithm:     settings.MaxReqsAlgorithm,
		MaxRequestsPerSecond: int(settings.MaxRequestsPerSecond),
		MaxRequestsPerMinute: int(settings.MaxRequestsPerMinute),
		MaxRequestsPerHour:   int(settings.MaxRequestsPerHour),
	}
	fullLimiter.maxReqsLimiter = NewMaxReqsLimiter(maxReqsRateLimitSettings)
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
