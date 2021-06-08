package ratelimit

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/log"
	"golang.org/x/time/rate"
)

/*
 * The time/rate package limiter that uses Wait() with maxBurstSize == 1 implements the leaky bucket algorithm as a queue (to use leaky bucket as a meter, use TokenBucketLimiter)
 * Requests are never rejected, just queued up and then "leaked" out of the bucket at a set rate (reqsPerSecond)
 * Useful for throttling requests (eg. grpc interceptor)
 * FlowRateLimitAlgorithm
 */
type LeakyBucketLimiter struct {
	limiter       *rate.Limiter
	reqsPerSecond float64
}

func NewLeakyBucketLimiter(reqsPerSecond float64) *LeakyBucketLimiter {
	l := &LeakyBucketLimiter{}
	l.reqsPerSecond = reqsPerSecond
	l.limiter = rate.NewLimiter(rate.Limit(reqsPerSecond), 1)
	return l
}

func (l *LeakyBucketLimiter) Limit(ctx context.Context, info *CallerInfo) error {
	err := l.limiter.Wait(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Error during leakybucket rate limiting", "error", err)
		return fmt.Errorf("error during leakybucker rate limiting: %s", err)
	}
	return nil
}

func (l *LeakyBucketLimiter) Type() string {
	return "LeakyBucketLimiter"
}
