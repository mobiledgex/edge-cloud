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
// settingsPerApiEndpointType maps the apiEndpointType (eg. ControllerCreate) to a list of APIs (eg. CreateApp, CreateCloudlet, etc.). This is used to update the rate limit settings if the settings api is updated
type RateLimitManager struct {
	util.Mutex
	limitsPerApi               map[string]*ApiEndpointLimiter
	settingsPerApiEndpointType map[edgeproto.ApiEndpointType][]string
}

// Create a RateLimitManager
func NewRateLimitManager() *RateLimitManager {
	r := &RateLimitManager{}
	r.limitsPerApi = make(map[string]*ApiEndpointLimiter)
	r.settingsPerApiEndpointType = make(map[edgeproto.ApiEndpointType][]string)
	return r
}

// Initialize an ApiEndpointLimiter struct for the specified API. This is called for each API during the initialization of the node (ie. in dme-main.go or controller.go)
func (r *RateLimitManager) AddApiEndpointLimiter(api string, apiEndpointRateLimitSettings *ApiEndpointRateLimitSettings, apiEndpointType edgeproto.ApiEndpointType) {
	if apiEndpointRateLimitSettings != nil {
		// Add apiEndpointType to map of settingsPerApi
		apis, ok := r.settingsPerApiEndpointType[apiEndpointType]
		if ok && apis != nil {
			r.settingsPerApiEndpointType[apiEndpointType] = append(apis, api)
		} else {
			r.settingsPerApiEndpointType[apiEndpointType] = []string{api}
		}

		r.limitsPerApi[api] = NewApiEndpointLimiter(apiEndpointRateLimitSettings)
	}
}

// Update the rate limit settings for all the APIs that use the rate limit settings associated with the specified apiEndpointType
func (r *RateLimitManager) UpdateRateLimitSettings(apiEndpointRateLimitSettings *ApiEndpointRateLimitSettings, apiEndpointType edgeproto.ApiEndpointType) {
	r.Lock()
	defer r.Unlock()
	apis, ok := r.settingsPerApiEndpointType[apiEndpointType]
	if ok && apis != nil {
		for _, api := range apis {
			r.AddApiEndpointLimiter(api, apiEndpointRateLimitSettings, apiEndpointType)
		}
	}
}

// Remove rate limit settings for all the APIs that use the rate limit settings associated with the specified apiEndpointType
func (r *RateLimitManager) RemoveRateLimitSettings(apiEndpointType edgeproto.ApiEndpointType) {
	r.Lock()
	defer r.Unlock()
	apis, ok := r.settingsPerApiEndpointType[apiEndpointType]
	if ok && apis != nil {
		for _, api := range apis {
			// Remove the api from limitsPerApi map
			delete(r.limitsPerApi, api)
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
		return false, fmt.Errorf("Unable to get LimiterInfo from context. Skipping rate limit")
	}
	// Check that api exists
	api := li.Api
	ApiEndpointLimiter, ok := r.limitsPerApi[api]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfo, "Unable to find api in RateLimitManager", "api", api)
		return false, nil
	}

	if r.doesLimitByIp(api) && li.Ip != "" {
		// limit per ip
		limiter, ok := ApiEndpointLimiter.limitsPerIp[li.Ip]
		if !ok {
			// add ip
			limiter = NewFullLimiter(ApiEndpointLimiter.apiEndpointRateLimitSettings.PerIpRateLimitSettings)
			ApiEndpointLimiter.limitsPerIp[li.Ip] = limiter
		}
		limit, err := limiter.Limit(ctx)
		if limit {
			return limit, fmt.Errorf("client exceeded api rate limit per ip. %s", err)
		}
	}
	if r.doesLimitByUser(api) && li.User != "" {
		// limit per user
		limiter, ok := ApiEndpointLimiter.limitsPerUser[li.User]
		if !ok {
			// add user
			limiter = NewFullLimiter(ApiEndpointLimiter.apiEndpointRateLimitSettings.PerUserRateLimitSettings)
			ApiEndpointLimiter.limitsPerUser[li.User] = limiter
		}
		limit, err := limiter.Limit(ctx)
		if limit {
			return limit, fmt.Errorf("user \"%s\" exceeded api rate limit per user. %s", li.User, err)
		}
	}
	if r.doesLimitByOrg(api) && li.Org != "" {
		// limit per org
		limiter, ok := ApiEndpointLimiter.limitsPerOrg[li.Org]
		if !ok {
			// add org
			limiter = NewFullLimiter(ApiEndpointLimiter.apiEndpointRateLimitSettings.PerOrgRateLimitSettings)
			ApiEndpointLimiter.limitsPerOrg[li.Org] = limiter
		}
		limit, err := limiter.Limit(ctx)
		if limit {
			return limit, fmt.Errorf("org \"%s\" exceeded api rate limit per org. %s", li.Org, err)
		}
	}
	if r.doesLimitByEndpoint(api) {
		// limit for the entire endpoint
		return ApiEndpointLimiter.limits.Limit(ctx)
	}
	return false, nil
}

// Helper function that checks if the mgr should rate limit per user
func (r *RateLimitManager) doesLimitByUser(api string) bool {
	ApiEndpointLimiter, ok := r.limitsPerApi[api]
	if !ok {
		return false
	}
	if ApiEndpointLimiter.apiEndpointRateLimitSettings.PerUserRateLimitSettings == nil {
		return false
	}
	return true
}

// Helper function that checks if the mgr should rate limit per org
func (r *RateLimitManager) doesLimitByOrg(api string) bool {
	ApiEndpointLimiter, ok := r.limitsPerApi[api]
	if !ok {
		return false
	}
	if ApiEndpointLimiter.apiEndpointRateLimitSettings.PerOrgRateLimitSettings == nil {
		return false
	}
	return true
}

// Helper function that checks if the mgr should rate limit per ip
func (r *RateLimitManager) doesLimitByIp(api string) bool {
	ApiEndpointLimiter, ok := r.limitsPerApi[api]
	if !ok {
		return false
	}
	if ApiEndpointLimiter.apiEndpointRateLimitSettings.PerIpRateLimitSettings == nil {
		return false
	}
	return true
}

// Helper function that checks if the mgr should rate limit the endpoint
func (r *RateLimitManager) doesLimitByEndpoint(api string) bool {
	ApiEndpointLimiter, ok := r.limitsPerApi[api]
	if !ok {
		return false
	}
	if ApiEndpointLimiter.apiEndpointRateLimitSettings.FullEndpointRateLimitSettings == nil {
		return false
	}
	return true
}
