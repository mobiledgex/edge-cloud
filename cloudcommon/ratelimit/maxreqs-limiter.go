package ratelimit

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
)

// MaxReqs algorithms rate limiter
// Imposes a maximum number of requests in a time frame (either fixed or rolling) for a user, organization, or ip
// This is useful if we want to provide billing tiers for API usage (particularly free vs. paid tiers)
type MaxReqsLimiter struct {
	util.Mutex
	limiter Limiter
}

func NewMaxReqsLimiter(settings *MaxReqsRateLimitSettings) *MaxReqsLimiter {
	maxReqsLimiter := &MaxReqsLimiter{}
	switch settings.MaxReqsAlgorithm {
	case edgeproto.MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM:
		maxReqsLimiter.limiter = NewFixedWindowLimiter(int(settings.MaxRequestsPerSecond), int(settings.MaxRequestsPerMinute), int(settings.MaxRequestsPerHour))
	case edgeproto.MaxReqsRateLimitAlgorithm_ROLLING_WINDOW_ALGORITHM:
		fallthrough
	default:
		return nil
	}
	return maxReqsLimiter
}

func (f *MaxReqsLimiter) Limit(ctx context.Context) (bool, error) {
	f.Lock()
	defer f.Unlock()
	if f.limiter != nil {
		return f.limiter.Limit(ctx)
	}
	return false, nil
}

type MaxReqsRateLimitSettings struct {
	MaxReqsAlgorithm     edgeproto.MaxReqsRateLimitAlgorithm
	MaxRequestsPerSecond int
	MaxRequestsPerMinute int
	MaxRequestsPerHour   int
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
