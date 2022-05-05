// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ratelimit

import (
	"context"
	"fmt"

	"github.com/edgexr/edge-cloud/log"
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
