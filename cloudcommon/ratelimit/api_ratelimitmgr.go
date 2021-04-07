package ratelimit

import "context"

type ApiRateLimitManager struct {
	limitsPerApi map[string]*rateLimitPerApi
}

func NewRateLimitManager() *ApiRateLimitManager {
	r := &ApiRateLimitManager{}
	r.limitsPerApi = make(map[string]*rateLimitPerApi)
	return r
}

func (r *ApiRateLimitManager) AddRateLimitPerApi(api string) {
	r.limitsPerApi[api] = newRateLimitPerApi()
}

func (r *ApiRateLimitManager) AddRateLimitPerUser(api string, reqsPerMinute int, reqsPerHour int, reqsPerDay int, reqsPerMonth int) {
	rateLimitPerApi, ok := r.limitsPerApi[api]
	if !ok {
		// log or return error
	}
	rateLimitPerApi.perUserFixedWindowLimiterConfig = &FixedWindowLimiterConfig{
		reqsPerMinute: reqsPerMinute,
		reqsPerHour:   reqsPerHour,
		reqsPerDay:    reqsPerDay,
		reqsPerMonth:  reqsPerMonth,
	}
}

func (r *ApiRateLimitManager) AddRateLimitPerOrg(api string, reqsPerMinute int, reqsPerHour int, reqsPerDay int, reqsPerMonth int) {
	rateLimitPerApi, ok := r.limitsPerApi[api]
	if !ok {
		// log or return error
	}
	rateLimitPerApi.perOrgFixedWindowLimiterConfig = &FixedWindowLimiterConfig{
		reqsPerMinute: reqsPerMinute,
		reqsPerHour:   reqsPerHour,
		reqsPerDay:    reqsPerDay,
		reqsPerMonth:  reqsPerMonth,
	}
}

func (r *ApiRateLimitManager) AddRateLimitPerIp(api string, reqsPerMinute int, reqsPerHour int, reqsPerDay int, reqsPerMonth int) {
	rateLimitPerApi, ok := r.limitsPerApi[api]
	if !ok {
		// log or return error
	}
	rateLimitPerApi.perIpFixedWindowLimiterConfig = &FixedWindowLimiterConfig{
		reqsPerMinute: reqsPerMinute,
		reqsPerHour:   reqsPerHour,
		reqsPerDay:    reqsPerDay,
		reqsPerMonth:  reqsPerMonth,
	}
}

// TODO: Consolidate dict lookups into helper function
func (r *ApiRateLimitManager) Limit(ctx context.Context, api string, user string, org string, ip string) (bool, error) {
	rateLimitPerApi, ok := r.limitsPerApi[api]
	if !ok {
		// log or return error
		return false, nil
	}
	if r.LimitByUser(api) && user != "" {
		// limit per user
		limiter, ok := rateLimitPerApi.limitsPerUser[user]
		if !ok {
			// add user
			limiter = NewFixedWindowLimiter(rateLimitPerApi.perUserFixedWindowLimiterConfig)
			rateLimitPerApi.limitsPerUser[user] = limiter
		}
		return limiter.Limit(ctx)
	}
	if r.LimitByOrg(api) && org != "" {
		// limit per org
		limiter, ok := rateLimitPerApi.limitsPerOrg[org]
		if !ok {
			// add org
			limiter = NewFixedWindowLimiter(rateLimitPerApi.perOrgFixedWindowLimiterConfig)
			rateLimitPerApi.limitsPerOrg[org] = limiter
		}
		return limiter.Limit(ctx)
	}
	if r.LimitByIp(api) && ip != "" {
		// limit per ip
		limiter, ok := rateLimitPerApi.limitsPerIp[ip]
		if !ok {
			// add ip
			limiter = NewFixedWindowLimiter(rateLimitPerApi.perIpFixedWindowLimiterConfig)
			rateLimitPerApi.limitsPerIp[ip] = limiter
		}
		return limiter.Limit(ctx)
	}
	return false, nil
}

func (r *ApiRateLimitManager) LimitByUser(api string) bool {
	rateLimitPerApi, ok := r.limitsPerApi[api]
	if !ok {
		// log or return error
		return false
	}
	if rateLimitPerApi.perUserFixedWindowLimiterConfig == nil {
		return false
	}
	return true
}

func (r *ApiRateLimitManager) LimitByOrg(api string) bool {
	rateLimitPerApi, ok := r.limitsPerApi[api]
	if !ok {
		// log or return error
		return false
	}
	if rateLimitPerApi.perOrgFixedWindowLimiterConfig == nil {
		return false
	}
	return true
}

func (r *ApiRateLimitManager) LimitByIp(api string) bool {
	rateLimitPerApi, ok := r.limitsPerApi[api]
	if !ok {
		// log or return error
		return false
	}
	if rateLimitPerApi.perIpFixedWindowLimiterConfig == nil {
		return false
	}
	return true
}

type rateLimitPerApi struct {
	perUserFixedWindowLimiterConfig *FixedWindowLimiterConfig
	limitsPerUser                   map[string]*FixedWindowLimiter
	perOrgFixedWindowLimiterConfig  *FixedWindowLimiterConfig
	limitsPerOrg                    map[string]*FixedWindowLimiter
	perIpFixedWindowLimiterConfig   *FixedWindowLimiterConfig
	limitsPerIp                     map[string]*FixedWindowLimiter
}

func newRateLimitPerApi() *rateLimitPerApi {
	r := &rateLimitPerApi{}
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
