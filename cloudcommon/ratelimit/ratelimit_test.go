package ratelimit

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLeakyBucketLimiter(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	// Spawn a couple "clients" that each wait for LeakyBucket algorithm
	// test that the LeakyBucket will allow all of the requests through in the correct amount of time
	numRequestsPerSecond := 0.5
	numClients := 4
	leakyBucket := NewLeakyBucketLimiter(numRequestsPerSecond)
	before := time.Now()
	done := make(chan bool, numClients)
	for i := 0; i < numClients; i++ {
		go func() {
			leakyBucket.Limit(ctx, nil)
			done <- true
		}()
	}
	for i := 0; i < numClients; i++ {
		<-done
	}
	expectedTime := time.Duration((float64(numClients) - 1.0) / numRequestsPerSecond)
	require.True(t, time.Since(before) > expectedTime*time.Second)
}

func TestTokenBucketLimiter(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	// test that TokenBucket rejects requests that come in too quickly
	tokenBucket := NewTokenBucketLimiter(1, 1)
	err := tokenBucket.Limit(ctx, nil)
	require.Nil(t, err)
	err = tokenBucket.Limit(ctx, nil)
	require.NotNil(t, err)
	require.True(t, strings.Contains(err.Error(), "Exceeded rate"))

	// test that TokenBucket allows requests coming in at the same time that don't exceed burst size
	numClients := 3
	tokenBucket = NewTokenBucketLimiter(float64(numClients), numClients)
	start := make(chan struct{})
	done := make(chan error, numClients)
	for i := 0; i < numClients; i++ {
		go func() {
			<-start
			err = tokenBucket.Limit(ctx, nil)
			done <- err
		}()
	}
	// after 1 second, the token bucket will be full
	time.Sleep(time.Second)
	close(start)
	for i := 0; i < numClients; i++ {
		err = <-done
		require.Nil(t, err)
	}
}

func TestIntervalLimiter(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	// test fixed number of requests that come in serially
	numRequestsPerSecond := 3
	intervalLimiter := NewIntervalLimiter(numRequestsPerSecond, time.Duration(time.Second))
	for i := 0; i < numRequestsPerSecond; i++ {
		err := intervalLimiter.Limit(ctx, nil)
		assert.Nil(t, err)
	}
	err := intervalLimiter.Limit(ctx, nil)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "exceeded limit"))

	// test fixed number of requests that come in concurrently
	intervalLimiter = NewIntervalLimiter(numRequestsPerSecond, time.Duration(time.Second))
	done := make(chan error, numRequestsPerSecond+1)
	for i := 0; i < numRequestsPerSecond+1; i++ {
		go func() {
			err := intervalLimiter.Limit(ctx, nil)
			done <- err
		}()
	}
	for i := 0; i < numRequestsPerSecond; i++ {
		err := <-done
		assert.Nil(t, err)
	}
	err = <-done
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "exceeded limit"))
}

func TestCompositeLimiter(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	// Create CompositeLimiter that is composed of two IntervalLimiters
	intervalLimiter1 := NewIntervalLimiter(1, time.Duration(time.Second))
	intervalLimiter2 := NewIntervalLimiter(2, time.Duration(time.Minute))
	compositeLimiter := NewCompositeLimiter(intervalLimiter1, intervalLimiter2)

	// test composite limiter serially
	err := compositeLimiter.Limit(ctx, nil)
	assert.Nil(t, err)
	time.Sleep(time.Second)
	err = compositeLimiter.Limit(ctx, nil)
	assert.Nil(t, err)
	time.Sleep(time.Second)
	err = compositeLimiter.Limit(ctx, nil)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "exceeded limit"))

	// test composite limiter concurrently
	numRequests := 5
	intervalLimiter1 = NewIntervalLimiter(numRequests, time.Duration(time.Second))
	intervalLimiter2 = NewIntervalLimiter(numRequests, time.Duration(time.Minute))
	compositeLimiter = NewCompositeLimiter(intervalLimiter1, intervalLimiter2)
	done := make(chan error, numRequests+1)
	for i := 0; i < numRequests+1; i++ {
		go func() {
			time.Sleep(time.Duration(rand.Intn(numRequests-1)) * time.Second)
			err := compositeLimiter.Limit(ctx, nil)
			done <- err
		}()
	}
	for i := 0; i < numRequests; i++ {
		err := <-done
		assert.Nil(t, err)
	}
	err = <-done
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "exceeded limit"))
}

func TestApiRateLimitMgr(t *testing.T) {
	// init ratelimitmgr
	mgr := NewRateLimitManager(false, 100, 100)
	// init apis
	api1 := "api1"
	api2 := "api2"
	api3 := "api3"
	global := edgeproto.GlobalApiName
	apis := []string{api1, api2, api3, global}
	// init rate limit settings (similar dme settings)
	settings1 := &edgeproto.RateLimitSettings{
		Key: edgeproto.RateLimitSettingsKey{
			ApiName:         api1,
			ApiEndpointType: edgeproto.ApiEndpointType_DME,
			RateLimitTarget: edgeproto.RateLimitTarget_ALL_REQUESTS,
		},
		FlowSettings: map[string]*edgeproto.FlowSettings{
			"api1allreqs1": &edgeproto.FlowSettings{
				FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
				ReqsPerSecond: 100,
				BurstSize:     100,
			},
		},
	}
	settings2 := &edgeproto.RateLimitSettings{
		Key: edgeproto.RateLimitSettingsKey{
			ApiName:         api2,
			ApiEndpointType: edgeproto.ApiEndpointType_DME,
			RateLimitTarget: edgeproto.RateLimitTarget_ALL_REQUESTS,
		},
		FlowSettings: map[string]*edgeproto.FlowSettings{
			"api2allreqs1": &edgeproto.FlowSettings{
				FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
				ReqsPerSecond: 200,
				BurstSize:     100,
			},
		},
	}
	settings3 := &edgeproto.RateLimitSettings{
		Key: edgeproto.RateLimitSettingsKey{
			ApiName:         api3,
			ApiEndpointType: edgeproto.ApiEndpointType_DME,
			RateLimitTarget: edgeproto.RateLimitTarget_ALL_REQUESTS,
		},
		FlowSettings: map[string]*edgeproto.FlowSettings{
			"api3allreqs1": &edgeproto.FlowSettings{
				FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
				ReqsPerSecond: 300,
				BurstSize:     100,
			},
		},
	}
	settingsGlobal := &edgeproto.RateLimitSettings{
		Key: edgeproto.RateLimitSettingsKey{
			ApiName:         global,
			ApiEndpointType: edgeproto.ApiEndpointType_DME,
			RateLimitTarget: edgeproto.RateLimitTarget_ALL_REQUESTS,
		},
		FlowSettings: map[string]*edgeproto.FlowSettings{
			"globalallreqs1": &edgeproto.FlowSettings{
				FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
				ReqsPerSecond: 1000,
				BurstSize:     100,
			},
		},
	}

	allsettings := []*edgeproto.RateLimitSettings{settings1, settings2, settings3, settingsGlobal}

	// Add apis and their rate limit settings to mgr
	for _, settings := range allsettings {
		mgr.CreateApiEndpointLimiter(settings.Key.ApiName, settings, nil, nil)
	}

	// Spawn fake clients that "call" the apis (all should pass)
	numClients := 100
	var wg sync.WaitGroup
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			ctx := context.Background()
			callerInfo := &CallerInfo{
				Api: apis[rand.Intn(len(apis))],
				Ip:  fmt.Sprintf("client%d", idx),
			}
			err := mgr.Limit(ctx, callerInfo)
			require.Nil(t, err)
		}(i)
	}
	wg.Wait()

	// Update Settings while clients are "calling" apis (some should fail)
	var err error
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ctx := context.Background()
			callerInfo := &CallerInfo{
				Api: apis[0],
				Ip:  fmt.Sprintf("client%d", idx),
			}
			e := mgr.Limit(ctx, callerInfo)
			if e != nil {
				err = e
			}
		}(i)
		// Update settings midway through
		if i == numClients/2 {
			newCreateLimitSettings := &edgeproto.MaxReqsRateLimitSettings{
				Key: edgeproto.MaxReqsRateLimitSettingsKey{
					MaxReqsSettingsName: "api1allreqs1",
					RateLimitKey: edgeproto.RateLimitSettingsKey{
						ApiName:         apis[0],
						ApiEndpointType: edgeproto.ApiEndpointType_DME,
						RateLimitTarget: edgeproto.RateLimitTarget_ALL_REQUESTS,
					},
				},
				Settings: edgeproto.MaxReqsSettings{
					MaxReqsAlgorithm: edgeproto.MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM,
					MaxRequests:      1,
					Interval:         edgeproto.Duration(time.Second),
				},
			}
			mgr.UpdateMaxReqsRateLimitSettings(newCreateLimitSettings)
		}
	}
	wg.Wait()
	require.NotNil(t, err)
	require.True(t, strings.Contains(err.Error(), "exceeded limit"))
}
