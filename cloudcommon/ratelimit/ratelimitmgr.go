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
	limitsPerApi              map[string]*apiEndpointLimiter
	disableRateLimit          bool
	maxNumPerIpRateLimiters   int
	maxNumPerUserRateLimiters int
}

// Create a RateLimitManager
func NewRateLimitManager(disableRateLimit bool, maxNumPerIpRateLimiters int, maxNumPerUserRateLimiters int) *RateLimitManager {
	r := &RateLimitManager{}
	r.limitsPerApi = make(map[string]*apiEndpointLimiter)
	r.disableRateLimit = disableRateLimit
	r.maxNumPerIpRateLimiters = maxNumPerIpRateLimiters
	r.maxNumPerUserRateLimiters = maxNumPerUserRateLimiters
	return r
}

// Initialize an ApiEndpointLimiter struct for the specified API. This is called for each API during the initialization of the node (ie. in dme-main.go)
func (r *RateLimitManager) CreateApiEndpointLimiter(allRequestsRateLimitSettings *edgeproto.RateLimitSettings, perIpRateLimitSettings *edgeproto.RateLimitSettings, perUserRateLimitSettings *edgeproto.RateLimitSettings) {
	r.Lock()
	defer r.Unlock()
	// Create ApiEndpointRateLimitSettings which includes limit settings for allrequests, perip, and/or peruser
	api := allRequestsRateLimitSettings.Key.ApiName
	apiEndpointRateLimitSettings := &apiEndpointRateLimitSettings{
		AllRequestsRateLimitSettings: allRequestsRateLimitSettings,
		PerIpRateLimitSettings:       perIpRateLimitSettings,
		PerUserRateLimitSettings:     perUserRateLimitSettings,
	}
	// Map API to ApiEndpointLimiter for easy lookup
	r.limitsPerApi[api] = newApiEndpointLimiter(api, apiEndpointRateLimitSettings, r.maxNumPerIpRateLimiters, r.maxNumPerIpRateLimiters)
}

// Update the rate limit settings for API that use the rate limit settings associated with the specified RateLimitSettingsKey
func (r *RateLimitManager) UpdateRateLimitSettings(rateLimitSettings *edgeproto.RateLimitSettings) {
	r.Lock()
	defer r.Unlock()
	// Look up ApiEndpointLimiter for specified API (create one if not found)
	api := rateLimitSettings.Key.ApiName
	limiter, ok := r.limitsPerApi[api]
	if !ok || limiter == nil {
		limiter = newApiEndpointLimiter(api, &apiEndpointRateLimitSettings{}, r.maxNumPerIpRateLimiters, r.maxNumPerIpRateLimiters)
	}
	// Update ApiEndpointLimiter with new RateLimitSettings
	limiter.updateApiEndpointLimiterSettings(rateLimitSettings)
	r.limitsPerApi[api] = limiter
}

/*
 * Remove the rate limit settings for API associated with the specified RateLimitSettingsKey
 * For example, this might remove the PerIp rate limiting for VerifyLocation
 */
func (r *RateLimitManager) RemoveRateLimitSettings(key edgeproto.RateLimitSettingsKey) {
	r.Lock()
	defer r.Unlock()
	api := key.ApiName
	limiter, ok := r.limitsPerApi[api]
	if !ok || limiter == nil {
		return
	}
	limiter.removeApiEndpointLimiterSettings(key.RateLimitTarget)
}

// Get RateLimitSettings associated with key. If none, return nil
func (r *RateLimitManager) GetRateLimitSettings(key edgeproto.RateLimitSettingsKey) *edgeproto.RateLimitSettings {
	r.Lock()
	defer r.Unlock()
	api := key.ApiName
	limiter, ok := r.limitsPerApi[api]
	if !ok || limiter == nil {
		return nil
	}
	return limiter.getApiEndpointLimiterSettings(key.RateLimitTarget)
}

// Update DisableRateLimit when settings are updated
func (r *RateLimitManager) UpdateDisableRateLimit(disable bool) {
	r.Lock()
	defer r.Unlock()
	r.disableRateLimit = disable
}

// Update MaxNumPerIpRateLimiters when settings are updated
func (r *RateLimitManager) UpdateMaxNumPerIpRateLimiters(max int) {
	r.Lock()
	defer r.Unlock()
	r.maxNumPerIpRateLimiters = max
	for _, limiter := range r.limitsPerApi {
		limiter.updateMaxNumIps(max)
	}
}

// Update MaxNumPerUserRateLimiters when settings are updated
func (r *RateLimitManager) UpdateMaxNumPerUserRateLimiters(max int) {
	r.Lock()
	defer r.Unlock()
	r.maxNumPerUserRateLimiters = max
	for _, limiter := range r.limitsPerApi {
		limiter.updateMaxNumUsers(max)
	}
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
