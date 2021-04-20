package ratelimit

import "fmt"

// TODO: MUTEX
// TODO: Allow unlimited for admin
// TODO: rate limit before auth or vice versa??
// TODO: Add time until next request available

type ApiRateLimitManager struct {
	limitsPerApi map[string]*rateLimitPerApi
}

func NewApiRateLimitManager() *ApiRateLimitManager {
	r := &ApiRateLimitManager{}
	r.limitsPerApi = make(map[string]*rateLimitPerApi)
	return r
}

func (r *ApiRateLimitManager) AddRateLimitPerApi(api string, perUserMaxReqs *ApiRateLimitMaxReqs, perOrgMaxReqs *ApiRateLimitMaxReqs, flowLimiter Limiter) {
	r.limitsPerApi[api] = newRateLimitPerApi(perUserMaxReqs, perOrgMaxReqs, flowLimiter)
}

// TODO: Consolidate dict lookups into helper function
func (r *ApiRateLimitManager) Limit(ctx Context) (bool, error) {
	api := ctx.Api
	rateLimitPerApi, ok := r.limitsPerApi[api]
	if !ok {
		// log or return error
		return false, nil
	}
	if r.doesLimitByUser(api) && ctx.User != "" {
		// limit per user
		limiter, ok := rateLimitPerApi.limitsPerUser[ctx.User]
		if !ok {
			// add user
			limiter = NewFixedWindowLimiter(rateLimitPerApi.perUserMaxReqs)
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
			limiter = NewFixedWindowLimiter(rateLimitPerApi.perOrgMaxReqs)
			rateLimitPerApi.limitsPerOrg[ctx.Org] = limiter
		}
		limit, err := limiter.Limit(ctx)
		if limit {
			return limit, fmt.Errorf("org \"%s\" exceeded api rate limit per org. %s", ctx.Org, err)
		}
	}
	if r.doesLimitByIp(api) && ctx.Ip != "" {
		// limit per ip
		limiter, ok := rateLimitPerApi.limitsPerIp[ctx.Ip]
		if !ok {
			// add ip
			limiter = NewFixedWindowLimiter(rateLimitPerApi.perUserMaxReqs)
			rateLimitPerApi.limitsPerIp[ctx.Ip] = limiter
		}
		limit, err := limiter.Limit(ctx)
		if limit {
			return limit, fmt.Errorf("client exceeded api rate limit per ip. %s", err)
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
	if rateLimitPerApi.perUserMaxReqs == nil {
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
	if rateLimitPerApi.perOrgMaxReqs == nil {
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
	if rateLimitPerApi.perUserMaxReqs == nil {
		return false
	}
	return true
}

type rateLimitPerApi struct {
	perUserMaxReqs *ApiRateLimitMaxReqs
	perOrgMaxReqs  *ApiRateLimitMaxReqs
	limitsPerUser  map[string]*FixedWindowLimiter
	limitsPerOrg   map[string]*FixedWindowLimiter
	limitsPerIp    map[string]*FixedWindowLimiter
	flowLimiter    Limiter
}

func newRateLimitPerApi(perUserMaxReqs *ApiRateLimitMaxReqs, perOrgMaxReqs *ApiRateLimitMaxReqs, flowLimiter Limiter) *rateLimitPerApi {
	r := &rateLimitPerApi{}
	r.perUserMaxReqs = perUserMaxReqs
	r.perOrgMaxReqs = perOrgMaxReqs
	r.flowLimiter = flowLimiter
	r.limitsPerUser = make(map[string]*FixedWindowLimiter)
	r.limitsPerOrg = make(map[string]*FixedWindowLimiter)
	r.limitsPerIp = make(map[string]*FixedWindowLimiter)
	return r
}

// per user and per org and per ip
// map of APIs to map of users and orgs
// each user and org is mapped to FixedWindowLimiter

// Rate limit interceptor that uses this Manager

// TODO: differentiate between RBAC (devs, operators, admin)

// TODO: add capability to do tiered api rates (connect with billing????)
