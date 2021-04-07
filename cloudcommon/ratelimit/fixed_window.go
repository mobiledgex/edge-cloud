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

type FixedWindowLimiterConfig struct {
	reqsPerMinute int
	reqsPerHour   int
	reqsPerDay    int
	reqsPerMonth  int
}

// TODO: Default reqsPer[Interval], Handle unlimited reqs (-1??)
func NewFixedWindowLimiter(config *FixedWindowLimiterConfig) *FixedWindowLimiter {
	f := &FixedWindowLimiter{}
	f.limiterPerMinute = NewIntervalLimiter(config.reqsPerMinute, time.Minute)
	f.limiterPerHour = NewIntervalLimiter(config.reqsPerHour, time.Hour)
	f.limiterPerDay = NewIntervalLimiter(config.reqsPerDay, 24*time.Hour)
	f.limiterPerMonth = NewIntervalLimiter(config.reqsPerMonth, 30*24*time.Hour)
	return f
}

// TODO: limit by largest timer interval first??
func (f *FixedWindowLimiter) Limit(ctx context.Context) (bool, error) {
	limit, err := f.limiterPerMinute.Limit(ctx)
	if limit {
		// log
		return true, fmt.Errorf("reached limit per minute. %s", err.Error())
	}
	limit, err = f.limiterPerHour.Limit(ctx)
	if limit {
		// log
		return true, fmt.Errorf("reached limit per hour. %s", err.Error())
	}
	limit, err = f.limiterPerDay.Limit(ctx)
	if limit {
		// log
		return true, fmt.Errorf("reached limit per day. %s", err.Error())
	}
	limit, err = f.limiterPerMonth.Limit(ctx)
	if limit {
		// log
		return true, fmt.Errorf("reached limit per month. %s", err.Error())
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

func (i *IntervalLimiter) Limit(ctx context.Context) (bool, error) {
	// Check start of interval
	if time.Since(i.intervalStartTime) > i.interval {
		i.intervalStartTime = time.Now().Truncate(i.interval)
		i.currentNumberOfRequests = 0
	}
	if i.currentNumberOfRequests >= i.requestLimit {
		return true, fmt.Errorf("limit is %d", i.requestLimit)
	} else {
		i.currentNumberOfRequests++
		return false, nil
	}
}

// TODO: Implement Sliding Window rate limiting algorithm
type SlidingWindowLimiter struct {
}
