package ratelimit

// Handles all the rate limiting for an API
// Can limit all requests coming into an enpoint, per IP, per User, and/or per Organization
type ApiEndpointLimiter struct {
	// Rate limit settings for an api endpoint
	apiEndpointRateLimitSettings *ApiEndpointRateLimitSettings
	// FullLimiter for api endpoint
	limits *FullLimiter
	// Maps of ip, user, or org to FullLimiters
	limitsPerIp   map[string]*FullLimiter
	limitsPerUser map[string]*FullLimiter
	limitsPerOrg  map[string]*FullLimiter
}

func NewApiEndpointLimiter(apiEndpointRateLimitSettings *ApiEndpointRateLimitSettings) *ApiEndpointLimiter {
	r := &ApiEndpointLimiter{}
	r.apiEndpointRateLimitSettings = apiEndpointRateLimitSettings
	r.limits = NewFullLimiter(apiEndpointRateLimitSettings.FullEndpointRateLimitSettings)
	r.limitsPerUser = make(map[string]*FullLimiter)
	r.limitsPerOrg = make(map[string]*FullLimiter)
	r.limitsPerIp = make(map[string]*FullLimiter)
	return r
}

// Rate Limit Settings for an API endpoint
// SHOULD ALL OF THESE SETTINGS BE IN EDGEPROTO??
/*type ApiEndpointRateLimitSettings struct {
	FullEndpointRateLimitSettings *edgeproto.RateLimitSettings
	PerIpRateLimitSettings        *edgeproto.RateLimitSettings
	PerUserRateLimitSettings      *edgeproto.RateLimitSettings
	PerOrgRateLimitSettings       *edgeproto.RateLimitSettings
}*/
