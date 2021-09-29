package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud/util"
)

/*
 * Limits requests based on requestLimit set for the specified interval
 * For example, if the interval is 1 hour and requestLimit is 100, the Limit function will reject requests once the 100 requests is reached, but will reset the count when an hour has passed.
 */
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
		intervalStartTime:       time.Now(),
	}
}

// TODO: Charge once surpass api limit
func (i *IntervalLimiter) Limit(ctx context.Context, info *CallerInfo) error {
	i.Lock()
	defer i.Unlock()
	// Check start of interval
	if time.Since(i.intervalStartTime) > i.interval {
		i.intervalStartTime = time.Now()
		i.currentNumberOfRequests = 0
	}
	if i.currentNumberOfRequests >= i.requestLimit && i.requestLimit != 0 {
		waitTime := i.interval - (time.Now().Sub(i.intervalStartTime))
		return fmt.Errorf("Exceeded limit of %d, retry again in %v", i.requestLimit, waitTime)
	} else {
		i.currentNumberOfRequests++
		return nil
	}
}

func (i *IntervalLimiter) Type() string {
	return "IntervalLimiter"
}
