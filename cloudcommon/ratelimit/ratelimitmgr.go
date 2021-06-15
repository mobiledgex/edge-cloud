package ratelimit

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

/*
 * RateLimitManager manages all the rate limits per API for a node (DME, Controller, and MC)
 * limitsPerApi maps an API to a ApiEndpointLimiter struct, which will handle all of the rate limiting for the endpoint and per ip, user, and/or org
 * apisPerRateLimitSettingsKey maps the RateLimitSettingsKey to a list of APIs (eg. CreateApp, CreateCloudlet, etc.). This is used to update the rate limit settings if the rate limit settings api is updated
 */
type RateLimitManager struct {
	util.Mutex
	limitsPerApi     map[string]*apiEndpointLimiter
	disableRateLimit bool
	maxNumLimiters   int
}

// Create a RateLimitManager
func NewRateLimitManager(disableRateLimit bool, maxNumLimiters int) *RateLimitManager {
	r := &RateLimitManager{}
	r.limitsPerApi = make(map[string]*apiEndpointLimiter)
	r.disableRateLimit = disableRateLimit
	r.maxNumLimiters = maxNumLimiters
	return r
}

// Initialize an ApiEndpointLimiter struct for the specified API. This is called for each API during the initialization of the node (ie. in dme-main.go)
func (r *RateLimitManager) CreateApiEndpointLimiter(allRequestsRateLimitSettings *edgeproto.RateLimitSettings, perIpRateLimitSettings *edgeproto.RateLimitSettings, perUserRateLimitSettings *edgeproto.RateLimitSettings) {
	r.Lock()
	defer r.Unlock()
	// TODO: Validate same apiname and correct target
	api := allRequestsRateLimitSettings.Key.ApiName
	apiEndpointRateLimitSettings := &apiEndpointRateLimitSettings{
		AllRequestsRateLimitSettings: allRequestsRateLimitSettings,
		PerIpRateLimitSettings:       perIpRateLimitSettings,
		PerUserRateLimitSettings:     perUserRateLimitSettings,
	}
	r.limitsPerApi[api] = newApiEndpointLimiter(api, apiEndpointRateLimitSettings)
}

// Update the rate limit settings for all the APIs that use the rate limit settings associated with the specified RateLimitSettingsKey
func (r *RateLimitManager) UpdateRateLimitSettings(rateLimitSettings *edgeproto.RateLimitSettings) {
	r.Lock()
	defer r.Unlock()
	api := rateLimitSettings.Key.ApiName
	limiter, ok := r.limitsPerApi[api]
	if !ok || limiter == nil {
		limiter = newApiEndpointLimiter(api, &apiEndpointRateLimitSettings{})
	}
	limiter.updateRateLimitSettings(rateLimitSettings)
	r.limitsPerApi[api] = limiter
}

func (r *RateLimitManager) RemoveRateLimitSettings(key edgeproto.RateLimitSettingsKey) {
	r.Lock()
	defer r.Unlock()
	api := key.ApiName
	limiter, ok := r.limitsPerApi[api]
	if !ok || limiter == nil {
		return
	}
	limiter.removeRateLimitSettings(key.RateLimitTarget)
}

func (r *RateLimitManager) UpdateDisableRateLimit(disable bool) {
	r.disableRateLimit = disable
}

func (r *RateLimitManager) UpdateMaxNumRateLimiters(max int) {
	r.maxNumLimiters = max
}

// Implements the Limiter interface
func (r *RateLimitManager) Limit(ctx context.Context, info *CallerInfo) error {
	r.Lock()
	defer r.Unlock()
	// Skip rest of function if rate limiting is not enabled
	if r.disableRateLimit {
		return nil
	}
	// Check for CallerInfo which provides essential information about the api, ip, user, and org
	if info == nil {
		log.DebugLog(log.DebugLevelInfo, "nil CallerInfo")
		return fmt.Errorf("nil CallerInfo - skipping rate limit")
	}
	// Check that api exists
	api := info.Api
	limiter, ok := r.limitsPerApi[api]
	if !ok {
		// If the api does not exist, check for global fallback limiter
		limiter, ok = r.limitsPerApi[edgeproto.GlobalApiName]
		if !ok {
			log.SpanLog(ctx, log.DebugLevelInfo, "Unable to find limiter for api or global limiter in ApiEndpointLimiter", "api", api)
			return nil
		}
	}
	// Check that ApiEndpointLimiter is non nil (ie. does rate limiting)
	if limiter == nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "No rate limiting on api", "api", api)
		return nil
	}
	return limiter.Limit(ctx, info)
}

func (r *RateLimitManager) Type() string {
	return "RateLimitManager"
}
