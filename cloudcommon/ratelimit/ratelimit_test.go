package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
)

func TestLeakyBucket(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	rateLimitCtx := Context{Context: ctx}
	leakyBucket := NewLeakyBucketLimiter(0.5)
	before := time.Now()
	done := make(chan bool, 4)
	go func() {
		leakyBucket.Limit(rateLimitCtx)
		done <- true
	}()
	go func() {
		leakyBucket.Limit(rateLimitCtx)
		done <- true
	}()
	go func() {
		leakyBucket.Limit(rateLimitCtx)
		done <- true
	}()
	go func() {
		leakyBucket.Limit(rateLimitCtx)
		done <- true
	}()
	<-done
	<-done
	<-done
	<-done
	require.True(t, time.Since(before) > 6*time.Second)
}
