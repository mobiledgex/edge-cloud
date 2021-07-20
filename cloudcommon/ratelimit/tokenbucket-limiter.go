package ratelimit

import (
	"context"
	"fmt"

	"golang.org/x/time/rate"
)

/*
 * The time/rate package limiter that uses Allow() implements the token bucket algorithm (which is equivalent to the leaky bucket algorithm as a meter. To use leaky bucket as a queue, use LeakyBucketLimiter)
 * A bucket is filled with tokens at tokensPerSecond and the bucket has a maximum size of bucketSize (bucketSize also acts as a burst size (ie. the number of requests that come at the same time that can be fulfilled)
 * A token is taken out on each request
 * Requests that come in when there are no tokens in the bucket are rejected
 * Useful for throttling requests (eg. grpc interceptor)
 * FlowRateLimitAlgorithm
 */
type TokenBucketLimiter struct {
	limiter         *rate.Limiter
	tokensPerSecond float64
	bucketSize      int
}

func NewTokenBucketLimiter(tokensPerSecond float64, bucketSize int) *TokenBucketLimiter {
	t := &TokenBucketLimiter{}
	t.tokensPerSecond = tokensPerSecond
	t.bucketSize = bucketSize
	t.limiter = rate.NewLimiter(rate.Limit(tokensPerSecond), bucketSize)
	return t
}

func (t *TokenBucketLimiter) Limit(ctx context.Context, info *CallerInfo) error {
	tokenAvailable := t.limiter.Allow()
	if !tokenAvailable {
		return fmt.Errorf("Exceeded rate of %f requests per second.", t.tokensPerSecond)
	} else {
		return nil
	}
}

func (t *TokenBucketLimiter) Type() string {
	return "TokenBucketLimiter"
}
