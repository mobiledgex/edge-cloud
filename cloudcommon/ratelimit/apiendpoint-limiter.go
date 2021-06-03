package ratelimit

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// Handles all the rate limiting for an API
// Can limit all requests coming into an enpoint, per IP, per User, and/or per Organization
type apiEndpointLimiter struct {
	// Rate limit settings for an api endpoint
	apiEndpointRateLimitSettings *apiEndpointRateLimitSettings
	// FullLimiter for api endpoint
	limits *FullLimiter
	// Maps of ip, user, or org to FullLimiters
	limitsPerIp   map[string]*FullLimiter
	limitsPerUser map[string]*FullLimiter
	limitsPerOrg  map[string]*FullLimiter
}

// Rate Limit Settings for an API endpoint
type apiEndpointRateLimitSettings struct {
	AllRequestsRateLimitSettings *edgeproto.RateLimitSettings
	PerIpRateLimitSettings       *edgeproto.RateLimitSettings
	PerUserRateLimitSettings     *edgeproto.RateLimitSettings
	PerOrgRateLimitSettings      *edgeproto.RateLimitSettings
}

func NewApiEndpointLimiter(apiEndpointRateLimitSettings *apiEndpointRateLimitSettings) *apiEndpointLimiter {
	if apiEndpointRateLimitSettings == nil {
		return nil
	}
	a := &apiEndpointLimiter{}
	a.apiEndpointRateLimitSettings = apiEndpointRateLimitSettings
	a.limits = NewFullLimiter(apiEndpointRateLimitSettings.AllRequestsRateLimitSettings)
	a.limitsPerIp = make(map[string]*FullLimiter)
	a.limitsPerUser = make(map[string]*FullLimiter)
	a.limitsPerOrg = make(map[string]*FullLimiter)
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
	case edgeproto.RateLimitTarget_PER_ORG:
		a.apiEndpointRateLimitSettings.PerOrgRateLimitSettings = rateLimitSettings
		a.limitsPerOrg = make(map[string]*FullLimiter)
	default:
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
	case edgeproto.RateLimitTarget_PER_ORG:
		a.apiEndpointRateLimitSettings.PerOrgRateLimitSettings = nil
		a.limitsPerOrg = make(map[string]*FullLimiter)
	default:
	}
}

// Implements the Limiter interface
func (a *apiEndpointLimiter) Limit(ctx context.Context) (bool, error) {
	// Check for LimiterInfo which provides essential information about the api, ip, user, and org
	li, ok := LimiterInfoFromContext(ctx)
	if !ok || li == nil {
		log.DebugLog(log.DebugLevelInfo, "Unable to get LimiterInfo from context. Skipping rate limit")
		return false, fmt.Errorf("Unable to get LimiterInfo from context. Skipping rate limit")
	}

	if a.doesLimitByIp() && li.Ip != "" {
		// limit per ip
		limiter, ok := a.limitsPerIp[li.Ip]
		if !ok {
			// add ip
			limiter = NewFullLimiter(a.apiEndpointRateLimitSettings.PerIpRateLimitSettings)
			a.limitsPerIp[li.Ip] = limiter
		}
		limit, err := limiter.Limit(ctx)
		if limit {
			return limit, fmt.Errorf("client exceeded api rate limit per ip. %s", err)
		}
	}
	if a.doesLimitByUser() && li.User != "" {
		// limit per user
		limiter, ok := a.limitsPerUser[li.User]
		if !ok {
			// add user
			limiter = NewFullLimiter(a.apiEndpointRateLimitSettings.PerUserRateLimitSettings)
			a.limitsPerUser[li.User] = limiter
		}
		limit, err := limiter.Limit(ctx)
		if limit {
			return limit, fmt.Errorf("user \"%s\" exceeded api rate limit per usea. %s", li.User, err)
		}
	}
	if a.doesLimitByOrg() && li.Org != "" {
		// limit per org
		limiter, ok := a.limitsPerOrg[li.Org]
		if !ok {
			// add org
			limiter = NewFullLimiter(a.apiEndpointRateLimitSettings.PerOrgRateLimitSettings)
			a.limitsPerOrg[li.Org] = limiter
		}
		limit, err := limiter.Limit(ctx)
		if limit {
			return limit, fmt.Errorf("org \"%s\" exceeded api rate limit per org. %s", li.Org, err)
		}
	}
	if a.doesLimitByAllRequests() {
		// limit for the entire endpoint
		return a.limits.Limit(ctx)
	}
	return false, nil
}

// Helper function that checks if the mgr should rate limit per user
func (a *apiEndpointLimiter) doesLimitByUser() bool {
	return a.apiEndpointRateLimitSettings.PerUserRateLimitSettings != nil
}

// Helper function that checks if the mgr should rate limit per org
func (a *apiEndpointLimiter) doesLimitByOrg() bool {
	return a.apiEndpointRateLimitSettings.PerOrgRateLimitSettings != nil
}

// Helper function that checks if the mgr should rate limit per ip
func (a *apiEndpointLimiter) doesLimitByIp() bool {
	return a.apiEndpointRateLimitSettings.PerIpRateLimitSettings != nil

}

// Helper function that checks if the mgr should rate limit the endpoint on all requests
func (a *apiEndpointLimiter) doesLimitByAllRequests() bool {
	return a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings != nil
}
