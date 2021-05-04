package ratelimit

import "github.com/mobiledgex/edge-cloud/edgeproto"

// FullLimiter
// Includes a MaxReqsLimiter and a FlowLimiter
type FullLimiter struct {
	flowLimiter    *FlowLimiter
	maxReqsLimiter *MaxReqsLimiter
}

func NewFullLimiter(settings *edgeproto.RateLimitSettings) *FullLimiter {
	fullLimiter := &FullLimiter{}
	fullLimiter.flowLimiter = NewFlowLimiter(settings.FlowRateLimitSettings)
	fullLimiter.maxReqsLimiter = NewMaxReqsLimiter(settings.MaxReqsRateLimitSettings)
	return fullLimiter
}

func (f *FullLimiter) Limit(ctx Context) (bool, error) {
	if f.flowLimiter != nil {
		return f.flowLimiter.Limit(ctx)
	}
	// TODO: add check from previous instead of returning
	if f.maxReqsLimiter != nil {
		return f.maxReqsLimiter.Limit(ctx)
	}
	return false, nil
}
