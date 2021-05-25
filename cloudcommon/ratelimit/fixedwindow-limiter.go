package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud/util"
)

// Allows a certain number of requests to pass in a fixed time periods (perSecond, perMinute, and/or perHour)
// If the limit is reached, the client must wait until the interval is over
// Useful for rate limiting per user (eg. API limits)
// MaxReqsRateLimitingAlgorithm
type FixedWindowLimiter struct {
	limiterPerSecond *IntervalLimiter
	limiterPerMinute *IntervalLimiter
	limiterPerHour   *IntervalLimiter
}

// Create a FixedWindowLimiter
func NewFixedWindowLimiter(maxReqsPerSecond int, maxReqsPerMinute int, maxReqsPerHour int) *FixedWindowLimiter {
	f := &FixedWindowLimiter{}
	// If valid maxReqsPerSecond, create an IntervalLimiter that accepts maxReqsPerSecond
	if maxReqsPerSecond > 0 {
		f.limiterPerSecond = NewIntervalLimiter(maxReqsPerSecond, time.Second)
	}
	// If valid maxReqsPerMinute, create an IntervalLimiter that accepts maxReqsPerMinute
	if maxReqsPerMinute > 0 {
		f.limiterPerMinute = NewIntervalLimiter(maxReqsPerMinute, time.Minute)
	}
	// If valid maxReqsPerHour, create an IntervalLimiter that accepts maxReqsPerHour
	if maxReqsPerHour > 0 {
		f.limiterPerHour = NewIntervalLimiter(maxReqsPerHour, time.Hour)
	}
	return f
}

// Implements Limiter interface
func (f *FixedWindowLimiter) Limit(ctx context.Context) (bool, error) {
	if f.limiterPerSecond != nil {
		limit, err := f.limiterPerSecond.Limit(ctx)
		if limit {
			return true, fmt.Errorf("reached limit per second - %s", err.Error())
		}
	}
	if f.limiterPerMinute != nil {
		limit, err := f.limiterPerMinute.Limit(ctx)
		if limit {
			return true, fmt.Errorf("reached limit per minute - %s", err.Error())
		}
	}
	if f.limiterPerHour != nil {
		limit, err := f.limiterPerHour.Limit(ctx)
		if limit {
			return true, fmt.Errorf("reached limit per hour - %s", err.Error())
		}
	}
	return false, nil
}

// Limits requests based on requestLimit set for the specified interval
// For example, if the interval is 1 hour and requestLimit is 100, the Limit function will reject requests once the 100 requests is reached, but will reset the count when an hour has passed.
type IntervalLimiter struct {
	util.Mutex
	requestLimit            int
	currentNumberOfRequests int
	interval                time.Duration
	intervalStartTime       time.Time
}

// Creates IntervalLimiter
func NewIntervalLimiter(reqLimit int, interval time.Duration) *IntervalLimiter {
	return &IntervalLimiter{
		requestLimit:            reqLimit,
		currentNumberOfRequests: 0,
		interval:                interval,
		intervalStartTime:       time.Now().Truncate(interval),
	}
}

// TODO: Charge once surpass api limit
func (i *IntervalLimiter) Limit(ctx context.Context) (bool, error) {
	i.Lock()
	defer i.Unlock()
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
