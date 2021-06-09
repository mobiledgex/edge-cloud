package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

/*
 * Handles all the rate limiting for an API
 * Can limit all requests coming into an enpoint, per IP, and/or per User
 */
type apiEndpointLimiter struct {
	// Rate limit settings for an api endpoint
	apiEndpointRateLimitSettings *apiEndpointRateLimitSettings
	// FullLimiter for api endpoint
	limits *CompositeLimiter
	// Maps of ip or user to FullLimiters
	limitsPerIp    map[string]*CompositeLimiter
	limitsPerUser  map[string]*CompositeLimiter
	maxNumLimiters int64
}

// Rate Limit Settings for an API endpoint
type apiEndpointRateLimitSettings struct {
	AllRequestsRateLimitSettings *edgeproto.RateLimitSettings
	PerIpRateLimitSettings       *edgeproto.RateLimitSettings
	PerUserRateLimitSettings     *edgeproto.RateLimitSettings
}

func NewApiEndpointLimiter(apiEndpointRateLimitSettings *apiEndpointRateLimitSettings) *apiEndpointLimiter {
	if apiEndpointRateLimitSettings == nil {
		return nil
	}
	a := &apiEndpointLimiter{}
	a.apiEndpointRateLimitSettings = apiEndpointRateLimitSettings
	limiters := getLimitersFromRateLimitSettings(apiEndpointRateLimitSettings.AllRequestsRateLimitSettings)
	a.limits = NewCompositeLimiter(limiters...)
	a.limitsPerIp = make(map[string]*CompositeLimiter)
	a.limitsPerUser = make(map[string]*CompositeLimiter)
	return a
}

func (a *apiEndpointLimiter) UpdateRateLimitSettings(rateLimitSettings *edgeproto.RateLimitSettings) {
	if rateLimitSettings == nil {
		return
	}
	key := rateLimitSettings.Key
	switch key.RateLimitTarget {
	case edgeproto.RateLimitTarget_ALL_REQUESTS:
		a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings = rateLimitSettings
		limiters := getLimitersFromRateLimitSettings(a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings)
		a.limits = NewCompositeLimiter(limiters...)
	case edgeproto.RateLimitTarget_PER_IP:
		a.apiEndpointRateLimitSettings.PerIpRateLimitSettings = rateLimitSettings
		a.limitsPerIp = make(map[string]*CompositeLimiter)
	case edgeproto.RateLimitTarget_PER_USER:
		a.apiEndpointRateLimitSettings.PerUserRateLimitSettings = rateLimitSettings
		a.limitsPerUser = make(map[string]*CompositeLimiter)
	default:
		// log
	}
}

func (a *apiEndpointLimiter) RemoveRateLimitSettings(target edgeproto.RateLimitTarget) {
	switch target {
	case edgeproto.RateLimitTarget_ALL_REQUESTS:
		a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings = nil
		a.limits = nil
	case edgeproto.RateLimitTarget_PER_IP:
		a.apiEndpointRateLimitSettings.PerIpRateLimitSettings = nil
		a.limitsPerIp = make(map[string]*CompositeLimiter)
	case edgeproto.RateLimitTarget_PER_USER:
		a.apiEndpointRateLimitSettings.PerUserRateLimitSettings = nil
		a.limitsPerUser = make(map[string]*CompositeLimiter)
	default:
		// log
	}
}

// Implements the Limiter interface
func (a *apiEndpointLimiter) Limit(ctx context.Context, info *CallerInfo) error {
	// Check for LimiterInfo which provides essential information about the api, ip, user, and org
	if info == nil {
		log.DebugLog(log.DebugLevelInfo, "nil CallerInfo - skipping rate limit")
		return fmt.Errorf("nil CallerInfo - skipping rate limit")
	}

	if a.doesLimitByIp() && info.Ip != "" {
		// limit per ip
		limiter, ok := a.limitsPerIp[info.Ip]
		if !ok {
			// add ip
			limiters := getLimitersFromRateLimitSettings(a.apiEndpointRateLimitSettings.PerIpRateLimitSettings)
			limiter = NewCompositeLimiter(limiters...)
			a.limitsPerIp[info.Ip] = limiter
		}
		err := limiter.Limit(ctx, info)
		if err != nil {
			return fmt.Errorf("client exceeded api rate limit per ip. %s", err)
		}
	}
	if a.doesLimitByUser() && info.User != "" {
		// limit per user
		limiter, ok := a.limitsPerUser[info.User]
		if !ok {
			// add user
			limiters := getLimitersFromRateLimitSettings(a.apiEndpointRateLimitSettings.PerUserRateLimitSettings)
			limiter = NewCompositeLimiter(limiters...)
			a.limitsPerUser[info.User] = limiter
		}
		err := limiter.Limit(ctx, info)
		if err != nil {
			return fmt.Errorf("user \"%s\" exceeded api rate limit per user. %s", info.User, err)
		}
	}
	if a.doesLimitByAllRequests() {
		// limit for the entire endpoint
		return a.limits.Limit(ctx, info)
	}
	return nil
}

func (a *apiEndpointLimiter) Type() string {
	return "apiEndpointLimiter"
}

// Helper function that checks if the mgr should rate limit per user
func (a *apiEndpointLimiter) doesLimitByUser() bool {
	return a.apiEndpointRateLimitSettings.PerUserRateLimitSettings != nil
}

// Helper function that checks if the mgr should rate limit per ip
func (a *apiEndpointLimiter) doesLimitByIp() bool {
	return a.apiEndpointRateLimitSettings.PerIpRateLimitSettings != nil

}

// Helper function that checks if the mgr should rate limit the endpoint on all requests
func (a *apiEndpointLimiter) doesLimitByAllRequests() bool {
	return a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings != nil
}

// Helper function that creates slice of Limiters to be passed into NewCompositeLimiter
func getLimitersFromRateLimitSettings(settings *edgeproto.RateLimitSettings) []Limiter {
	limiters := make([]Limiter, 0)

	for _, fsettings := range settings.FlowSettings {
		// Generate Flow Limiters
		switch fsettings.FlowAlgorithm {
		case edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM:
			limiters = append(limiters, NewTokenBucketLimiter(fsettings.ReqsPerSecond, int(fsettings.BurstSize)))
		case edgeproto.FlowRateLimitAlgorithm_LEAKY_BUCKET_ALGORITHM:
			limiters = append(limiters, NewLeakyBucketLimiter(fsettings.ReqsPerSecond))
		default:
			// log
		}
	}

	for _, msettings := range settings.MaxReqsSettings {
		// Generate MaxReqs Limiters
		switch msettings.MaxReqsAlgorithm {
		case edgeproto.MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM:
			limiters = append(limiters, NewIntervalLimiter(int(msettings.MaxRequests), time.Duration(msettings.Interval)))
		default:
			// log
		}
	}

	return limiters
}
