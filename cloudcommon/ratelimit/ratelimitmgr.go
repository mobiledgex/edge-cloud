package ratelimit

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

// RateLimitManager manages all the rate limits per API for a node (DME, Controller, and MC)
// limitsPerApi maps an API to a ApiEndpointLimiter struct, which will handle all of the rate limiting for the endpoint and per ip, user, and/or org
// apisPerRateLimitSettingsKey maps the RateLimitSettingsKey to a list of APIs (eg. CreateApp, CreateCloudlet, etc.). This is used to update the rate limit settings if the rate limit settings api is updated
type RateLimitManager struct {
	util.Mutex
	limitsPerApi                map[string]*apiEndpointLimiter
	apisPerRateLimitSettingsKey map[edgeproto.RateLimitSettingsKey][]string
}

// Create a RateLimitManager
func NewRateLimitManager() *RateLimitManager {
	r := &RateLimitManager{}
	r.limitsPerApi = make(map[string]*apiEndpointLimiter)
	r.apisPerRateLimitSettingsKey = make(map[edgeproto.RateLimitSettingsKey][]string)
	return r
}

// Initialize an ApiEndpointLimiter struct for the specified API. This is called for each API during the initialization of the node (ie. in dme-main.go or controller.go)
func (r *RateLimitManager) AddApiEndpointLimiter(api string, allRequestsRateLimitSettings *edgeproto.RateLimitSettings, perIpRateLimitSettings *edgeproto.RateLimitSettings, perUserRateLimitSettings *edgeproto.RateLimitSettings, perOrgRateLimitSettings *edgeproto.RateLimitSettings) {
	r.Lock()
	defer r.Unlock()
	// Add apiEndpointType apisPerRateLimitSettingsKey map
	r.addRateLimitSettingsKey(api, allRequestsRateLimitSettings)
	r.addRateLimitSettingsKey(api, perIpRateLimitSettings)
	r.addRateLimitSettingsKey(api, perUserRateLimitSettings)
	r.addRateLimitSettingsKey(api, perOrgRateLimitSettings)

	apiEndpointRateLimitSettings := &apiEndpointRateLimitSettings{
		AllRequestsRateLimitSettings: allRequestsRateLimitSettings,
		PerIpRateLimitSettings:       perIpRateLimitSettings,
		PerUserRateLimitSettings:     perUserRateLimitSettings,
		PerOrgRateLimitSettings:      perOrgRateLimitSettings,
	}
	r.limitsPerApi[api] = NewApiEndpointLimiter(apiEndpointRateLimitSettings)
}

// Update the rate limit settings for all the APIs that use the rate limit settings associated with the specified RateLimitSettingsKey
func (r *RateLimitManager) UpdateRateLimitSettings(rateLimitSettings *edgeproto.RateLimitSettings) {
	r.Lock()
	defer r.Unlock()
	key := rateLimitSettings.Key
	apis, ok := r.apisPerRateLimitSettingsKey[key]
	if ok && apis != nil {
		for _, api := range apis {
			limiter, ok := r.limitsPerApi[api]
			if !ok || limiter == nil {
				limiter = NewApiEndpointLimiter(&apiEndpointRateLimitSettings{})
				r.addRateLimitSettingsKey(api, rateLimitSettings)
			}
			limiter.UpdateRateLimitSettings(rateLimitSettings)
			r.limitsPerApi[api] = limiter
		}
	}
}

// Implements the Limiter interface
func (r *RateLimitManager) Limit(ctx context.Context) (bool, error) {
	r.Lock()
	defer r.Unlock()
	// Check for LimiterInfo which provides essential information about the api, ip, user, and org
	li, ok := LimiterInfoFromContext(ctx)
	if !ok || li == nil {
		log.DebugLog(log.DebugLevelInfo, "Unable to find LimiterInfo from context")
		return false, fmt.Errorf("Unable to get LimiterInfo from context. Skipping rate limit")
	}
	// Check that api exists
	api := li.Api
	limiter, ok := r.limitsPerApi[api]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfo, "Unable to find api in ApiEndpointLimiter", "api", api)
		return false, nil
	}
	// Check that ApiEndpointLimiter is non nil (ie. does rate limiting)
	if limiter == nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "No rate limiting on api", "api", api)
		return false, nil
	}
	return limiter.Limit(ctx)
}

// Helper function that adds the RateLimitSettingsKey to the apisPerRateLimitSettingsKey map (must lock before calling)
func (r *RateLimitManager) addRateLimitSettingsKey(api string, settings *edgeproto.RateLimitSettings) {
	if settings == nil {
		return
	}
	key := settings.Key
	apis, ok := r.apisPerRateLimitSettingsKey[key]
	if ok && apis != nil {
		r.apisPerRateLimitSettingsKey[key] = append(apis, api)
	} else {
		r.apisPerRateLimitSettingsKey[key] = []string{api}
	}
}
