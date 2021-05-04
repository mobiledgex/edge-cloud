package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// Useful for rate limiting per user (eg. API limits)
// MaxReqsRateLimitingAlgorithm
type FixedWindowLimiter struct {
	limiterPerSecond *IntervalLimiter
	limiterPerMinute *IntervalLimiter
	limiterPerHour   *IntervalLimiter
}

// TODO: Default reqsPer[Interval], Handle unlimited reqs (-1??)
func NewFixedWindowLimiter(maxReqsPerSecondPerConsumer int, maxReqsPerMinutePerConsumer int, maxReqsPerHourPerConsumer int) *FixedWindowLimiter {
	f := &FixedWindowLimiter{}
	if maxReqsPerSecondPerConsumer > 0 {
		f.limiterPerSecond = NewIntervalLimiter(maxReqsPerSecondPerConsumer, time.Second)
	}
	if maxReqsPerMinutePerConsumer > 0 {
		f.limiterPerMinute = NewIntervalLimiter(maxReqsPerMinutePerConsumer, time.Minute)
	}
	if maxReqsPerHourPerConsumer > 0 {
		f.limiterPerHour = NewIntervalLimiter(maxReqsPerHourPerConsumer, time.Hour)
	}
	return f
}

// TODO: limit by largest timer interval first??
func (f *FixedWindowLimiter) Limit(ctx Context) (bool, error) {
	if f.limiterPerSecond != nil {
		limit, err := f.limiterPerSecond.Limit(ctx)
		if limit {
			// log
			return true, fmt.Errorf("reached limit per second - %s", err.Error())
		}
	}
	if f.limiterPerMinute != nil {
		limit, err := f.limiterPerMinute.Limit(ctx)
		if limit {
			// log
			return true, fmt.Errorf("reached limit per minute - %s", err.Error())
		}
	}
	if f.limiterPerHour != nil {
		limit, err := f.limiterPerHour.Limit(ctx)
		if limit {
			// log
			return true, fmt.Errorf("reached limit per hour - %s", err.Error())
		}
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
	mux                     sync.Mutex
}

func NewIntervalLimiter(reqLimit int, interval time.Duration) *IntervalLimiter {
	return &IntervalLimiter{
		requestLimit:            reqLimit,
		currentNumberOfRequests: 0,
		interval:                interval,
		intervalStartTime:       time.Now().Truncate(interval),
	}
}

// TODO: Charge once surpass api limit???
func (i *IntervalLimiter) Limit(ctx Context) (bool, error) {
	i.mux.Lock()
	defer i.mux.Unlock()
	// Check start of interval
	if time.Since(i.intervalStartTime) > i.interval {
		i.intervalStartTime = time.Now().Truncate(i.interval)
		i.currentNumberOfRequests = 0
	}
	if i.currentNumberOfRequests >= i.requestLimit && i.requestLimit != 0 {
		waitTime := i.interval - (time.Now().Sub(i.intervalStartTime))
		return true, fmt.Errorf("exceeded limit of %d, retry again in %v", i.requestLimit, waitTime)
	} else {
		i.currentNumberOfRequests++
		return false, nil
	}
}

// TODO: Implement Sliding Window rate limiting algorithm
// MaxReqsRateLimitingAlgorithm
type SlidingWindowLimiter struct {
}
