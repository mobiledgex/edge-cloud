package ratelimit

type ApiRateLimitMaxReqs struct {
	maxReqsPerMinutePerConsumer int
	maxReqsPerHourPerConsumer   int
	maxReqsPerDayPerConsumer    int
	maxReqsPerMonthPerConsumer  int
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

var DefaultReqsPerSecond = 100

var DefaultTokenBucketSize = 10

// TODO: GROUP rates BY INDIVIDUAL RPCs or SERVICES??? (answer: lets do services)

var DefaultDeveloperApiRateLimitMaxReqs = &ApiRateLimitMaxReqs{
	maxReqsPerMinutePerConsumer: 5,
	maxReqsPerHourPerConsumer:   100,
	maxReqsPerDayPerConsumer:    1000,
	maxReqsPerMonthPerConsumer:  10000,
}

var DefaultOperatorApiRateLimitMaxReqs = &ApiRateLimitMaxReqs{
	maxReqsPerMinutePerConsumer: 10,
	maxReqsPerHourPerConsumer:   200,
	maxReqsPerDayPerConsumer:    2000,
	maxReqsPerMonthPerConsumer:  20000,
}

var DefaultDeveloperOrgApiRateLimitMaxReqs = &ApiRateLimitMaxReqs{
	maxReqsPerMinutePerConsumer: 50,
	maxReqsPerHourPerConsumer:   1000,
	maxReqsPerDayPerConsumer:    10000,
	maxReqsPerMonthPerConsumer:  100000,
}

var DefaultOperatorOrgApiRateLimitMaxReqs = &ApiRateLimitMaxReqs{
	maxReqsPerMinutePerConsumer: 100,
	maxReqsPerHourPerConsumer:   2000,
	maxReqsPerDayPerConsumer:    20000,
	maxReqsPerMonthPerConsumer:  200000,
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
}
