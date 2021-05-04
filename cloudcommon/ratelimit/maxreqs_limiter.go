package ratelimit

import (
	"sync"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// MaxReqs algorithms rate limiter
// Imposes a maximum number of requests in a time frame (either fixed or rolling) for a user, organization, or ip
// This is useful if we want to provide billing tiers for API usage (particularly free vs. paid tiers)
type MaxReqsLimiter struct {
	limiter Limiter
	mux     sync.Mutex
}

func NewMaxReqsLimiter(settings *edgeproto.MaxReqsRateLimitSettings) *MaxReqsLimiter {
	maxReqsLimiter := &MaxReqsLimiter{}
	switch settings.MaxReqsAlgorithm {
	case edgeproto.MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM:
		maxReqsLimiter.limiter = NewFixedWindowLimiter(int(settings.MaxRequestsPerSecond), int(settings.MaxRequestsPerMinute), int(settings.MaxRequestsPerHour))
	case edgeproto.MaxReqsRateLimitAlgorithm_ROLLING_WINDOW_ALGORITHM:
		// log
		fallthrough
	case edgeproto.MaxReqsRateLimitAlgorithm_NO_MAX_REQS_ALGORITHM:
		// log
		fallthrough
	default:
		// log
		return nil
	}
	return maxReqsLimiter
}

func (f *MaxReqsLimiter) Limit(ctx Context) (bool, error) {
	f.mux.Lock()
	defer f.mux.Unlock()
	if f.limiter != nil {
		return f.limiter.Limit(ctx)
	}
	return false, nil
}

/*
// ApiTier is used as a multiplier for each rate
// For example, a tier2 DeveloperCreateMaxReqs would allow 5*10 maxReqsPerMinute, 100*10 maxReqsPerHour, 1000*10 maxReqsPerDay, and 10000*10 maxReqsPerMonth
type ApiTier int

const (
	tier1 ApiTier = 1
	tier2 ApiTier = 10
	tier3 ApiTier = 100
)

var DefaultPerApiMaxReqsRateLimitSettings = &edgeproto.MaxReqsRateLimitSettings{
	MaxReqsAlgorithm:     edgeproto.MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM,
	MaxRequestsPerSecond: 100,
	MaxRequestsPerMinute: 500,
	MaxRequestsPerHour:   1000,
}

var DefaultPerApiPerIpMaxReqsRateLimitSettings = &edgeproto.MaxReqsRateLimitSettings{
	MaxReqsAlgorithm:     edgeproto.MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM,
	MaxRequestsPerSecond: 1,
	MaxRequestsPerMinute: 10,
	MaxRequestsPerHour:   100,
}

var DefaultPerApiPerUserMaxReqsRateLimitSettings = &edgeproto.MaxReqsRateLimitSettings{
	MaxReqsAlgorithm:     edgeproto.MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM,
	MaxRequestsPerSecond: 1,
	MaxRequestsPerMinute: 10,
	MaxRequestsPerHour:   100,
}

var DefaultPerApiPerOrgMaxReqsRateLimitSettings = &edgeproto.MaxReqsRateLimitSettings{
	MaxReqsAlgorithm:     edgeproto.MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM,
	MaxRequestsPerSecond: 50,
	MaxRequestsPerMinute: 100,
	MaxRequestsPerHour:   500,
}*/
