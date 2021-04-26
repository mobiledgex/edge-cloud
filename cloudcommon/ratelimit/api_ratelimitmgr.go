package ratelimit

import "fmt"

// TODO: MUTEX
// TODO: Allow unlimited for admin
// TODO: rate limit before auth or vice versa?? (ans: before)

type ApiRateLimitManager struct {
	limitsPerApi map[string]*rateLimitPerApi
}

func NewApiRateLimitManager() *ApiRateLimitManager {
	r := &ApiRateLimitManager{}
	r.limitsPerApi = make(map[string]*rateLimitPerApi)
	return r
}

func (r *ApiRateLimitManager) AddRateLimitPerApi(api string, flowLimiter *FlowLimiter, perIpConfig *FullLimiterConfig, perUserConfig *FullLimiterConfig, perOrgConfig *FullLimiterConfig) {
	r.limitsPerApi[api] = newRateLimitPerApi(flowLimiter, perIpConfig, perUserConfig, perOrgConfig)
}

// TODO: Consolidate dict lookups into helper function
func (r *ApiRateLimitManager) Limit(ctx Context) (bool, error) {
	api := ctx.Api
	rateLimitPerApi, ok := r.limitsPerApi[api]
	if !ok {
		// log or return error
		return false, nil
	}
	if r.doesLimitByIp(api) && ctx.Ip != "" {
		// limit per ip
		limiter, ok := rateLimitPerApi.limitsPerIp[ctx.Ip]
		if !ok {
			// add ip
			limiter = NewFullLimiter(rateLimitPerApi.perUserConfig)
			rateLimitPerApi.limitsPerIp[ctx.Ip] = limiter
		}
		limit, err := limiter.Limit(ctx)
		if limit {
			return limit, fmt.Errorf("client exceeded api rate limit per ip. %s", err)
		}
	}
	if r.doesLimitByUser(api) && ctx.User != "" {
		// limit per user
		limiter, ok := rateLimitPerApi.limitsPerUser[ctx.User]
		if !ok {
			// add user
			limiter = NewFullLimiter(rateLimitPerApi.perUserConfig)
			rateLimitPerApi.limitsPerUser[ctx.User] = limiter
		}
		limit, err := limiter.Limit(ctx)
		if limit {
			return limit, fmt.Errorf("user \"%s\" exceeded api rate limit per user. %s", ctx.User, err)
		}
	}
	if r.doesLimitByOrg(api) && ctx.Org != "" {
		// limit per org
		limiter, ok := rateLimitPerApi.limitsPerOrg[ctx.Org]
		if !ok {
			// add org
			limiter = NewFullLimiter(rateLimitPerApi.perOrgConfig)
			rateLimitPerApi.limitsPerOrg[ctx.Org] = limiter
		}
		limit, err := limiter.Limit(ctx)
		if limit {
			return limit, fmt.Errorf("org \"%s\" exceeded api rate limit per org. %s", ctx.Org, err)
		}
	}
	return rateLimitPerApi.flowLimiter.Limit(ctx)
}

func (r *ApiRateLimitManager) doesLimitByUser(api string) bool {
	rateLimitPerApi, ok := r.limitsPerApi[api]
	if !ok {
		// log or return error
		return false
	}
	if rateLimitPerApi.perUserConfig == nil {
		return false
	}
	return true
}

func (r *ApiRateLimitManager) doesLimitByOrg(api string) bool {
	rateLimitPerApi, ok := r.limitsPerApi[api]
	if !ok {
		// log or return error
		return false
	}
	if rateLimitPerApi.perOrgConfig == nil {
		return false
	}
	return true
}

func (r *ApiRateLimitManager) doesLimitByIp(api string) bool {
	rateLimitPerApi, ok := r.limitsPerApi[api]
	if !ok {
		// log or return error
		return false
	}
	if rateLimitPerApi.perUserConfig == nil {
		return false
	}
	return true
}

type rateLimitPerApi struct {
	// Flow limiter for API endpoint
	flowLimiter *FlowLimiter
	// FullLimiterConfigs per ip, user, and org
	perIpConfig   *FullLimiterConfig
	perUserConfig *FullLimiterConfig
	perOrgConfig  *FullLimiterConfig
	// Maps of ip, user, or org to FullLimiters
	limitsPerIp   map[string]*FullLimiter
	limitsPerUser map[string]*FullLimiter
	limitsPerOrg  map[string]*FullLimiter
}

func newRateLimitPerApi(flowLimiter *FlowLimiter, perIpConfig *FullLimiterConfig, perUserConfig *FullLimiterConfig, perOrgConfig *FullLimiterConfig) *rateLimitPerApi {
	r := &rateLimitPerApi{}
	r.flowLimiter = flowLimiter
	r.perIpConfig = perIpConfig
	r.perUserConfig = perUserConfig
	r.perOrgConfig = perOrgConfig
	r.limitsPerUser = make(map[string]*FullLimiter)
	r.limitsPerOrg = make(map[string]*FullLimiter)
	r.limitsPerIp = make(map[string]*FullLimiter)
	return r
}
