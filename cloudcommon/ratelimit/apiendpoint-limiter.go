package ratelimit

import (
	"context"
	"fmt"

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
	limitsPerIp   map[string]*CompositeLimiter
	limitsPerUser map[string]*CompositeLimiter
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
	a.limits = NewCompositeLimiter(getLimitersFromRateLimitSettings(apiEndpointRateLimitSettings.AllRequestsRateLimitSettings))
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
		a.limits = NewFullLimiter(rateLimitSettings)
	case edgeproto.RateLimitTarget_PER_IP:
		a.apiEndpointRateLimitSettings.PerIpRateLimitSettings = rateLimitSettings
		a.limitsPerIp = make(map[string]*FullLimiter)
	case edgeproto.RateLimitTarget_PER_USER:
		a.apiEndpointRateLimitSettings.PerUserRateLimitSettings = rateLimitSettings
		a.limitsPerUser = make(map[string]*FullLimiter)
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
		a.limitsPerIp = make(map[string]*FullLimiter)
	case edgeproto.RateLimitTarget_PER_USER:
		a.apiEndpointRateLimitSettings.PerUserRateLimitSettings = nil
		a.limitsPerUser = make(map[string]*FullLimiter)
	default:
		// log
	}
}

// Implements the Limiter interface
func (a *apiEndpointLimiter) Limit(ctx context.Context, info *CallerInfo) error {
	// Check for LimiterInfo which provides essential information about the api, ip, user, and org
	if info == nil {
		log.DebugLog(log.DebugLevelInfo, "nil CallerInfo - skipping rate limit")
		return false, fmt.Errorf("nil CallerInfo - skipping rate limit")
	}

	if a.doesLimitByIp() && info.Ip != "" {
		// limit per ip
		limiter, ok := a.limitsPerIp[info.Ip]
		if !ok {
			// add ip
			limiter = NewCompositeLimiter(getLimitersFromRateLimitSettings(a.apiEndpointRateLimitSettings.PerIpRateLimitSettings))
			a.limitsPerIp[info.Ip] = limiter
		}
		limit, err := limiter.Limit(ctx, info)
		if limit {
			return limit, fmt.Errorf("client exceeded api rate limit per ip. %s", err)
		}
	}
	if a.doesLimitByUser() && info.User != "" {
		// limit per user
		limiter, ok := a.limitsPerUser[info.User]
		if !ok {
			// add user
			limiter = NewCompositeLimiter(getLimitersFromRateLimitSettings(a.apiEndpointRateLimitSettings.PerUserRateLimitSettings))
			a.limitsPerUser[info.User] = limiter
		}
		limit, err := limiter.Limit(ctx, info)
		if limit {
			return limit, fmt.Errorf("user \"%s\" exceeded api rate limit per user. %s", info.User, err)
		}
	}
	if a.doesLimitByAllRequests() {
		// limit for the entire endpoint
		return a.limits.Limit(ctx, info)
	}
	return false, nil
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
	limiters := make([]Limiter)

	// Generate Flow Limiters
	switch settings.FlowAlgorithm {
	case edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM:
		limiters.append(limiters, NewTokenBucketLimiter(settings.ReqsPerSecond, settings.BurstSize))
	case edgeproto.FlowRateLimitAlgorithm_LEAKY_BUCKET_ALGORITHM:
		limiters.append(limiters, NewLeakyBucketLimiter(settings.ReqsPerSecond))
	default:
		// log
	}

	// Generate MaxReqs Limiters
	switch settings.MaxReqsAlgorithm {
	case edgeproto.MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM:
		for _, maxReqsSettings := settings.MaxReqsSettings {
			limiters.append(limiters, NewIntervalLimiter(maxReqsSettings.MaxRequests, maxReqsSettings.Interval))
		}
	default:
		// log
	}

	return limiters
}
