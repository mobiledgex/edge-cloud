package ratelimit

import "sync"

// MaxReqs algorithms rate limiter
// Imposes a maximum number of requests in a time frame (either fixed or rolling) for a user, organization, or ip
// This is useful if we want to provide billing tiers for API usage (particularly free vs. paid tiers)
type MaxReqsLimiter struct {
	limiter Limiter
	mux     sync.Mutex
}

type MaxReqsLimiterConfig struct {
	MaxReqsAlgorithm MaxReqsRateLimitingAlgorithm
	MaxReqsPerSecond int
	MaxReqsPerMinute int
	MaxReqsPerHour   int
}

type MaxReqsRateLimitingAlgorithm int

const (
	NoMaxReqsAlgorithm MaxReqsRateLimitingAlgorithm = iota
	RollingWindowAlgorithm
	FixedWindowAlgorithm
)

func NewMaxReqsLimiter(config *MaxReqsLimiterConfig) *MaxReqsLimiter {
	maxReqsLimiter := &MaxReqsLimiter{}
	switch config.MaxReqsAlgorithm {
	case FixedWindowAlgorithm:
		maxReqsLimiter.limiter = NewFixedWindowLimiter(config.MaxReqsPerSecond, config.MaxReqsPerMinute, config.MaxReqsPerHour)
	case RollingWindowAlgorithm:
		// log
		fallthrough
	case NoMaxReqsAlgorithm:
		// log
		fallthrough
	default:
		// log
		return nil
	}
	return maxReqsLimiter
}

func (f *MaxReqsLimiter) Limit(ctx Context) (bool, error) {
	f.mux.Lock()
	defer f.mux.Unlock()
	if f.limiter != nil {
		return f.limiter.Limit(ctx)
	}
	return false, nil
}

/*type ApiRateLimitMaxReqs struct {
	maxReqsPerMinutePerConsumer int
	maxReqsPerHourPerConsumer   int
	maxReqsPerDayPerConsumer    int
}

func (a *ApiRateLimitMaxReqs) Scale(scaleFactor int) {
	a.maxReqsPerDayPerConsumer *= scaleFactor
	a.maxReqsPerHourPerConsumer *= scaleFactor
	a.maxReqsPerMinutePerConsumer *= scaleFactor
}

// TODO: Get rid of perMonth limit?

// ApiTier is used as a multiplier for each rate
// For example, a tier2 DeveloperCreateMaxReqs would allow 5*10 maxReqsPerMinute, 100*10 maxReqsPerHour, 1000*10 maxReqsPerDay, and 10000*10 maxReqsPerMonth
type ApiTier int

const (
	tier1 ApiTier = 1
	tier2 ApiTier = 10
	tier3 ApiTier = 100
)

// TODO: GROUP rates BY INDIVIDUAL RPCs or SERVICES??? (answer: lets do services)

var DefaultIpApiRateLimitMaxReqs = &ApiRateLimitMaxReqs{
	maxReqsPerMinutePerConsumer: 5,
	maxReqsPerHourPerConsumer:   100,
	maxReqsPerDayPerConsumer:    1000,
}

var DefaultUserApiRateLimitMaxReqs = &ApiRateLimitMaxReqs{
	maxReqsPerMinutePerConsumer: 5,
	maxReqsPerHourPerConsumer:   100,
	maxReqsPerDayPerConsumer:    1000,
}

var DefaultOrgApiRateLimitMaxReqs = &ApiRateLimitMaxReqs{
	maxReqsPerMinutePerConsumer: 50,
	maxReqsPerHourPerConsumer:   1000,
	maxReqsPerDayPerConsumer:    10000,
}

var DefaultDmeApiRateLimitMaxReqs = &ApiRateLimitMaxReqs{
	maxReqsPerMinutePerConsumer: 100,
	maxReqsPerHourPerConsumer:   1000,
	maxReqsPerDayPerConsumer:    10000,
}

var TestDmeApiRateLimitMaxReqs = &ApiRateLimitMaxReqs{
	maxReqsPerMinutePerConsumer: 5,
	maxReqsPerHourPerConsumer:   10,
	maxReqsPerDayPerConsumer:    100,
}*/
