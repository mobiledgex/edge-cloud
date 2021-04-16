package ratelimit

import (
	"context"
	"fmt"
	"time"
)

// Useful for rate limiting per user (eg. API limits)
type FixedWindowLimiter struct {
	limiterPerMinute *IntervalLimiter
	limiterPerHour   *IntervalLimiter
	limiterPerDay    *IntervalLimiter
	limiterPerMonth  *IntervalLimiter
}

// TODO: Default reqsPer[Interval], Handle unlimited reqs (-1??)
func NewFixedWindowLimiter(maxReqs *ApiRateLimitMaxReqs) *FixedWindowLimiter {
	f := &FixedWindowLimiter{}
	f.limiterPerMinute = NewIntervalLimiter(maxReqs.maxReqsPerMinutePerConsumer, time.Minute)
	f.limiterPerHour = NewIntervalLimiter(maxReqs.maxReqsPerHourPerConsumer, time.Hour)
	f.limiterPerDay = NewIntervalLimiter(maxReqs.maxReqsPerDayPerConsumer, 24*time.Hour)
	f.limiterPerMonth = NewIntervalLimiter(maxReqs.maxReqsPerMonthPerConsumer, 30*24*time.Hour)
	return f
}

// TODO: limit by largest timer interval first??
func (f *FixedWindowLimiter) Limit(ctx context.Context) (bool, error) {
	limit, err := f.limiterPerMinute.Limit(ctx)
	if limit {
		// log
		return true, fmt.Errorf("reached limit per minute - %s", err.Error())
	}
	limit, err = f.limiterPerHour.Limit(ctx)
	if limit {
		// log
		return true, fmt.Errorf("reached limit per hour - %s", err.Error())
	}
	limit, err = f.limiterPerDay.Limit(ctx)
	if limit {
		// log
		return true, fmt.Errorf("reached limit per day - %s", err.Error())
	}
	limit, err = f.limiterPerMonth.Limit(ctx)
	if limit {
		// log
		return true, fmt.Errorf("reached limit per month - %s", err.Error())
	}
	return false, nil
}

// Limits requests based on requestLimit set for the specified interval
// For example, if the interval is 1 hour and requestLimit is 100, the Limit function will reject requests once the 100 requests is reached, but will reset the count when an hour has passed.
type IntervalLimiter struct {
	requestLimit            int
	currentNumberOfRequests int
	interval                time.Duration
	intervalStartTime       time.Time
}

func NewIntervalLimiter(reqLimit int, interval time.Duration) *IntervalLimiter {
	return &IntervalLimiter{
		requestLimit:            reqLimit,
		currentNumberOfRequests: 0,
		interval:                interval,
		intervalStartTime:       time.Now().Truncate(interval),
	}
}

// TODO: Limit only successful requests???
// TODO: Charge once surpass api limit???
func (i *IntervalLimiter) Limit(ctx context.Context) (bool, error) {
	// Check start of interval
	if time.Since(i.intervalStartTime) > i.interval {
		i.intervalStartTime = time.Now().Truncate(i.interval)
		i.currentNumberOfRequests = 0
	}
	if i.currentNumberOfRequests >= i.requestLimit && i.requestLimit != 0 {
		return true, fmt.Errorf("exceeded limit of %d", i.requestLimit)
	} else {
		i.currentNumberOfRequests++
		return false, nil
	}
}

// TODO: Implement Sliding Window rate limiting algorithm
type SlidingWindowLimiter struct {
}
