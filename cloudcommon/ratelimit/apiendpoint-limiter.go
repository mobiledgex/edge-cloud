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
 * Can limit all requests coming into an endpoint, per IP, and/or per User
 */
type apiEndpointLimiter struct {
	// Name of API endpoint
	apiName string
	// All Rate limit settings for an api endpoint
	apiEndpointRateLimitSettings *apiEndpointRateLimitSettings
	// Limiter for all requests that come into the api endpoint
	limitAllRequests *CompositeLimiter
	// Maps of ip or user to CompositeLimiters
	limitsPerIp   map[string]*CompositeLimiter
	limitsPerUser map[string]*CompositeLimiter
	// Maximum number of Ips and/or Users allowed in hashmaps
	maxNumIps   int
	maxNumUsers int
}

// Rate Limit Settings for an API endpoint
type apiEndpointRateLimitSettings struct {
	AllRequestsRateLimitSettings *edgeproto.RateLimitSettings
	PerIpRateLimitSettings       *edgeproto.RateLimitSettings
	PerUserRateLimitSettings     *edgeproto.RateLimitSettings
}

// Create an ApiEndpointLimiter
func newApiEndpointLimiter(apiName string, apiEndpointRateLimitSettings *apiEndpointRateLimitSettings, maxNumIps int, maxNumUsers int) *apiEndpointLimiter {
	a := &apiEndpointLimiter{}
	a.apiName = apiName
	a.apiEndpointRateLimitSettings = apiEndpointRateLimitSettings
	limiters := getLimitersFromRateLimitSettings(apiEndpointRateLimitSettings.AllRequestsRateLimitSettings)
	a.limitAllRequests = NewCompositeLimiter(limiters...)
	a.limitsPerIp = make(map[string]*CompositeLimiter)
	a.limitsPerUser = make(map[string]*CompositeLimiter)
	a.maxNumIps = maxNumIps
	a.maxNumUsers = maxNumUsers
	return a
}

// Implements the Limiter interface
func (a *apiEndpointLimiter) Limit(ctx context.Context, info *CallerInfo) error {
	// Check for LimiterInfo which provides essential information about the api, ip, user, and org
	if info == nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "nil CallerInfo - skipping rate limit")
		return fmt.Errorf("nil CallerInfo - skipping rate limit")
	}

	if a.doesLimitByIp() && info.Ip != "" {
		// limit per ip
		limiter, ok := a.limitsPerIp[info.Ip]
		if !ok {
			a.removeExcessIps()
			// add ip
			limiters := getLimitersFromRateLimitSettings(a.apiEndpointRateLimitSettings.PerIpRateLimitSettings)
			limiter = NewCompositeLimiter(limiters...)
			a.limitsPerIp[info.Ip] = limiter
		}
		err := limiter.Limit(ctx, info)
		if err != nil {
			return fmt.Errorf("Client exceeded api rate limit per ip. %s", err)
		}
	}
	if a.doesLimitByUser() && info.User != "" {
		// limit per user
		limiter, ok := a.limitsPerUser[info.User]
		if !ok {
			a.removeExcessUsers()
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
		return a.limitAllRequests.Limit(ctx, info)
	}
	return nil
}

func (a *apiEndpointLimiter) Type() string {
	return "apiEndpointLimiter"
}

// Update the FlowRateLimitSettings for the corresponding RateLimitTarget (eg. update PerIp FlowRateLimitSettings)
func (a *apiEndpointLimiter) updateFlowRateLimitSettings(flowRateLimitSettings *edgeproto.FlowRateLimitSettings) {
	if flowRateLimitSettings == nil {
		return
	}
	// Update correct flow settings RateLimitSettings and initialize new limiters for specified RateLimitTarget
	switch flowRateLimitSettings.Key.RateLimitKey.RateLimitTarget {
	case edgeproto.RateLimitTarget_ALL_REQUESTS:
		if a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings == nil {
			a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings = &edgeproto.RateLimitSettings{
				Key: flowRateLimitSettings.Key.RateLimitKey,
			}
		}
		a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings.UpdateFlowSettings(flowRateLimitSettings)
		limiters := getLimitersFromRateLimitSettings(a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings)
		a.limitAllRequests = NewCompositeLimiter(limiters...)
	case edgeproto.RateLimitTarget_PER_IP:
		if a.apiEndpointRateLimitSettings.PerIpRateLimitSettings == nil {
			a.apiEndpointRateLimitSettings.PerIpRateLimitSettings = &edgeproto.RateLimitSettings{
				Key: flowRateLimitSettings.Key.RateLimitKey,
			}
		}
		a.apiEndpointRateLimitSettings.PerIpRateLimitSettings.UpdateFlowSettings(flowRateLimitSettings)
		a.limitsPerIp = make(map[string]*CompositeLimiter)
	case edgeproto.RateLimitTarget_PER_USER:
		if a.apiEndpointRateLimitSettings.PerUserRateLimitSettings == nil {
			a.apiEndpointRateLimitSettings.PerUserRateLimitSettings = &edgeproto.RateLimitSettings{
				Key: flowRateLimitSettings.Key.RateLimitKey,
			}
		}
		a.apiEndpointRateLimitSettings.PerUserRateLimitSettings.UpdateFlowSettings(flowRateLimitSettings)
		a.limitsPerUser = make(map[string]*CompositeLimiter)
	}
}

// Remove the FlowRateLimitSettings for the corresponding RateLimitTarget (eg. remove PerIp RateLimitSettings)
func (a *apiEndpointLimiter) removeFlowRateLimitSettings(target edgeproto.RateLimitTarget, name string) {
	// Remove FlowRateLimitSettings and limiters for specified RateLimitTarget
	switch target {
	case edgeproto.RateLimitTarget_ALL_REQUESTS:
		a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings.RemoveFlowSettings(name)
		limiters := getLimitersFromRateLimitSettings(a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings)
		a.limitAllRequests = NewCompositeLimiter(limiters...)
	case edgeproto.RateLimitTarget_PER_IP:
		a.apiEndpointRateLimitSettings.PerIpRateLimitSettings.RemoveFlowSettings(name)
		a.limitsPerIp = make(map[string]*CompositeLimiter)
	case edgeproto.RateLimitTarget_PER_USER:
		a.apiEndpointRateLimitSettings.PerUserRateLimitSettings.RemoveFlowSettings(name)
		a.limitsPerUser = make(map[string]*CompositeLimiter)
	}
}

// Update the MaxReqsRateLimitSettings for the corresponding RateLimitTarget (eg. update PerIp MaxReqsRateLimitSettings)
func (a *apiEndpointLimiter) updateMaxReqsRateLimitSettings(maxReqsRateLimitSettings *edgeproto.MaxReqsRateLimitSettings) {
	if maxReqsRateLimitSettings == nil {
		return
	}
	// Update max reqs settings of correct RateLimitSettings and initialize new limiters for specified RateLimitTarget
	switch maxReqsRateLimitSettings.Key.RateLimitKey.RateLimitTarget {
	case edgeproto.RateLimitTarget_ALL_REQUESTS:
		if a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings == nil {
			a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings = &edgeproto.RateLimitSettings{
				Key: maxReqsRateLimitSettings.Key.RateLimitKey,
			}
		}
		a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings.UpdateMaxReqsSettings(maxReqsRateLimitSettings)
		limiters := getLimitersFromRateLimitSettings(a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings)
		a.limitAllRequests = NewCompositeLimiter(limiters...)
	case edgeproto.RateLimitTarget_PER_IP:
		if a.apiEndpointRateLimitSettings.PerIpRateLimitSettings == nil {
			a.apiEndpointRateLimitSettings.PerIpRateLimitSettings = &edgeproto.RateLimitSettings{
				Key: maxReqsRateLimitSettings.Key.RateLimitKey,
			}
		}
		a.apiEndpointRateLimitSettings.PerIpRateLimitSettings.UpdateMaxReqsSettings(maxReqsRateLimitSettings)
		a.limitsPerIp = make(map[string]*CompositeLimiter)
	case edgeproto.RateLimitTarget_PER_USER:
		if a.apiEndpointRateLimitSettings.PerUserRateLimitSettings == nil {
			a.apiEndpointRateLimitSettings.PerUserRateLimitSettings = &edgeproto.RateLimitSettings{
				Key: maxReqsRateLimitSettings.Key.RateLimitKey,
			}
		}
		a.apiEndpointRateLimitSettings.PerUserRateLimitSettings.UpdateMaxReqsSettings(maxReqsRateLimitSettings)
		a.limitsPerUser = make(map[string]*CompositeLimiter)
	}
}

// Remove the MaxReqsRateLimitSettings for the corresponding RateLimitTarget (eg. remove PerIp RateLimitSettings)
func (a *apiEndpointLimiter) removeMaxReqsRateLimitSettings(target edgeproto.RateLimitTarget, name string) {
	// Remove MaxReqsRateLimitSettings and limiters for specified RateLimitTarget
	switch target {
	case edgeproto.RateLimitTarget_ALL_REQUESTS:
		a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings.RemoveMaxReqsSettings(name)
		limiters := getLimitersFromRateLimitSettings(a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings)
		a.limitAllRequests = NewCompositeLimiter(limiters...)
	case edgeproto.RateLimitTarget_PER_IP:
		a.apiEndpointRateLimitSettings.PerIpRateLimitSettings.RemoveMaxReqsSettings(name)
		a.limitsPerIp = make(map[string]*CompositeLimiter)
	case edgeproto.RateLimitTarget_PER_USER:
		a.apiEndpointRateLimitSettings.PerUserRateLimitSettings.RemoveMaxReqsSettings(name)
		a.limitsPerUser = make(map[string]*CompositeLimiter)
	}
}

// Prune the FlowRateLimitSettings that are not in the keys map
func (a *apiEndpointLimiter) pruneFlowRateLimitSettings(keys map[edgeproto.FlowRateLimitSettingsKey]struct{}) {
	if a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings != nil {
		for name, _ := range a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings.FlowSettings {
			key := edgeproto.FlowRateLimitSettingsKey{
				FlowSettingsName: name,
				RateLimitKey:     a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings.Key,
			}
			if _, ok := keys[key]; !ok {
				a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings.RemoveFlowSettings(name)
				limiters := getLimitersFromRateLimitSettings(a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings)
				a.limitAllRequests = NewCompositeLimiter(limiters...)
			}
		}
	}

	if a.apiEndpointRateLimitSettings.PerIpRateLimitSettings != nil {
		for name, _ := range a.apiEndpointRateLimitSettings.PerIpRateLimitSettings.FlowSettings {
			key := edgeproto.FlowRateLimitSettingsKey{
				FlowSettingsName: name,
				RateLimitKey:     a.apiEndpointRateLimitSettings.PerIpRateLimitSettings.Key,
			}
			if _, ok := keys[key]; !ok {
				a.apiEndpointRateLimitSettings.PerIpRateLimitSettings.RemoveFlowSettings(name)
				a.limitsPerIp = make(map[string]*CompositeLimiter)
			}
		}
	}

	if a.apiEndpointRateLimitSettings.PerUserRateLimitSettings != nil {
		for name, _ := range a.apiEndpointRateLimitSettings.PerUserRateLimitSettings.FlowSettings {
			key := edgeproto.FlowRateLimitSettingsKey{
				FlowSettingsName: name,
				RateLimitKey:     a.apiEndpointRateLimitSettings.PerUserRateLimitSettings.Key,
			}
			if _, ok := keys[key]; !ok {
				a.apiEndpointRateLimitSettings.PerUserRateLimitSettings.RemoveFlowSettings(name)
				a.limitsPerIp = make(map[string]*CompositeLimiter)
			}
		}
	}
}

// Prune the MaxReqsRateLimitSettings that are not in the keys map
func (a *apiEndpointLimiter) pruneMaxReqsRateLimitSettings(keys map[edgeproto.MaxReqsRateLimitSettingsKey]struct{}) {
	if a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings != nil {
		for name, _ := range a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings.MaxReqsSettings {
			key := edgeproto.MaxReqsRateLimitSettingsKey{
				MaxReqsSettingsName: name,
				RateLimitKey:        a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings.Key,
			}
			if _, ok := keys[key]; !ok {
				a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings.RemoveMaxReqsSettings(name)
				limiters := getLimitersFromRateLimitSettings(a.apiEndpointRateLimitSettings.AllRequestsRateLimitSettings)
				a.limitAllRequests = NewCompositeLimiter(limiters...)
			}
		}
	}

	if a.apiEndpointRateLimitSettings.PerIpRateLimitSettings != nil {
		for name, _ := range a.apiEndpointRateLimitSettings.PerIpRateLimitSettings.MaxReqsSettings {
			key := edgeproto.MaxReqsRateLimitSettingsKey{
				MaxReqsSettingsName: name,
				RateLimitKey:        a.apiEndpointRateLimitSettings.PerIpRateLimitSettings.Key,
			}
			if _, ok := keys[key]; !ok {
				a.apiEndpointRateLimitSettings.PerIpRateLimitSettings.RemoveMaxReqsSettings(name)
				a.limitsPerIp = make(map[string]*CompositeLimiter)
			}
		}
	}

	if a.apiEndpointRateLimitSettings.PerUserRateLimitSettings != nil {
		for name, _ := range a.apiEndpointRateLimitSettings.PerUserRateLimitSettings.MaxReqsSettings {
			key := edgeproto.MaxReqsRateLimitSettingsKey{
				MaxReqsSettingsName: name,
				RateLimitKey:        a.apiEndpointRateLimitSettings.PerUserRateLimitSettings.Key,
			}
			if _, ok := keys[key]; !ok {
				a.apiEndpointRateLimitSettings.PerUserRateLimitSettings.RemoveMaxReqsSettings(name)
				a.limitsPerIp = make(map[string]*CompositeLimiter)
			}
		}
	}
}

// TODO: Remove Ips and Users by oldest modified or inserted
// TODO: Remove Users that have been deleted

// Remove ips until size of limitsPerIp hashmap is less than maxNumIps
func (a *apiEndpointLimiter) removeExcessIps() {
	for ip, _ := range a.limitsPerIp {
		if len(a.limitsPerIp) < a.maxNumIps {
			break
		}
		delete(a.limitsPerIp, ip)
	}
}

// Remove user until size of limitsPerUser hashmap is less than maxNumUsers
func (a *apiEndpointLimiter) removeExcessUsers() {
	for user, _ := range a.limitsPerUser {
		if len(a.limitsPerUser) < a.maxNumUsers {
			break
		}
		delete(a.limitsPerUser, user)
	}
}

func (a *apiEndpointLimiter) updateMaxNumIps(max int) {
	a.maxNumIps = max
}

func (a *apiEndpointLimiter) updateMaxNumUsers(max int) {
	a.maxNumUsers = max
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
	if settings == nil {
		return limiters
	}

	// Generate Flow Limiters
	for _, fsettings := range settings.FlowSettings {
		switch fsettings.FlowAlgorithm {
		case edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM:
			limiters = append(limiters, NewTokenBucketLimiter(fsettings.ReqsPerSecond, int(fsettings.BurstSize)))
		case edgeproto.FlowRateLimitAlgorithm_LEAKY_BUCKET_ALGORITHM:
			limiters = append(limiters, NewLeakyBucketLimiter(fsettings.ReqsPerSecond))
		default:
		}
	}

	// Generate MaxReqs Limiters
	for _, msettings := range settings.MaxReqsSettings {
		switch msettings.MaxReqsAlgorithm {
		case edgeproto.MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM:
			limiters = append(limiters, NewIntervalLimiter(int(msettings.MaxRequests), time.Duration(msettings.Interval)))
		default:
		}
	}
	return limiters
}
