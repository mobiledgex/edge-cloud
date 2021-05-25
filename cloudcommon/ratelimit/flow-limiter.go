package ratelimit

import (
	"context"
	"sync"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// Flow rate limiting algorithms
// Limiting the number of requests/ flow of requests through an endpoint (ie. 1 request every second).
// This is useful for preventing DDos attacks and will allow functions that require a lot of work to handle a reasonable number of requests per time frame.
type FlowLimiter struct {
	limiter Limiter
	mux     sync.Mutex
}

func NewFlowLimiter(settings *FlowRateLimitSettings) *FlowLimiter {
	flowLimiter := &FlowLimiter{}
	switch settings.FlowAlgorithm {
	case edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM:
		flowLimiter.limiter = NewTokenBucketLimiter(settings.ReqsPerSecond, int(settings.BurstSize))
	case edgeproto.FlowRateLimitAlgorithm_LEAKY_BUCKET_ALGORITHM:
		flowLimiter.limiter = NewLeakyBucketLimiter(settings.ReqsPerSecond)
	default:
		return nil
	}
	return flowLimiter
}

func (f *FlowLimiter) Limit(ctx context.Context) (bool, error) {
	f.mux.Lock()
	defer f.mux.Unlock()
	if f.limiter != nil {
		return f.limiter.Limit(ctx)
	}
	return false, nil
}

type FlowRateLimitSettings struct {
	FlowAlgorithm edgeproto.FlowRateLimitAlgorithm
	ReqsPerSecond float64
	BurstSize     int
}
