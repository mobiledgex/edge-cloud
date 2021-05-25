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

func TestLeakyBucket(t *testing.T) {
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
			leakyBucket.Limit(ctx)
			done <- true
		}()
	}
	for i := 0; i < numClients; i++ {
		<-done
	}
	expectedTime := time.Duration((float64(numClients) - 1.0) / numRequestsPerSecond)
	require.True(t, time.Since(before) > expectedTime*time.Second)
}

func TestTokenBucket(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	// test that TokenBucket rejects requests that come in too quickly
	tokenBucket := NewTokenBucketLimiter(1, 1)
	limit, err := tokenBucket.Limit(ctx)
	require.False(t, limit)
	require.Nil(t, err)
	limit, err = tokenBucket.Limit(ctx)
	require.True(t, limit)
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
			limit, err = tokenBucket.Limit(ctx)
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

func TestFixedWindow(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	// test fixed number of requests that come in serially
	numRequestsPerSecond := 3
	fixedWindow := NewFixedWindowLimiter(numRequestsPerSecond, 0, 0)
	for i := 0; i < numRequestsPerSecond; i++ {
		limit, err := fixedWindow.Limit(ctx)
		assert.False(t, limit)
		assert.Nil(t, err)
	}
	limit, err := fixedWindow.Limit(ctx)
	assert.True(t, limit)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "reached limit per second"))

	// test fixed number of requests that come in concurrently
	fixedWindow = NewFixedWindowLimiter(numRequestsPerSecond, 0, 0)
	done := make(chan error, numRequestsPerSecond+1)
	for i := 0; i < numRequestsPerSecond+1; i++ {
		go func() {
			_, err := fixedWindow.Limit(ctx)
			done <- err
		}()
	}
	for i := 0; i < numRequestsPerSecond; i++ {
		err := <-done
		assert.Nil(t, err)
	}
	err = <-done
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "reached limit per second"))
}

func TestApiRateLimitMgr(t *testing.T) {
	// init ratelimitmgr
	mgr := NewApiRateLimitManager()
	// init apis
	api1 := "api1"
	api2 := "api2"
	api3 := "api3"
	api4 := "api4"
	api5 := "api5"
	api6 := "api6"
	api7 := "api7"
	apis := []string{api1, api2, api3, api4, api5, api6, api7}
	// init rate limit settings (similar to controller and dme settings)
	createLimitSettings := edgeproto.ApiEndpointRateLimitSettings{
		EndpointRateLimitSettings: &edgeproto.RateLimitSettings{
			FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
				FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
				ReqsPerSecond: 100,
				BurstSize:     100,
			},
		},
		EndpointPerIpRateLimitSettings: &edgeproto.RateLimitSettings{
			FlowRateLimitSettings: &edgeproto.FlowRateLimitSettings{
				FlowAlgorithm: edgeproto.FlowRateLimitAlgorithm_TOKEN_BUCKET_ALGORITHM,
				ReqsPerSecond: 100,
				BurstSize:     100,
			},
		},
	}
	deleteLimitSettings := createLimitSettings
	showLimitSettings := createLimitSettings
	settings := []edgeproto.ApiEndpointRateLimitSettings{createLimitSettings, deleteLimitSettings, showLimitSettings}
	settingsNames := []string{"create", "delete", "show"}

	// Add apis and their rate limit settings to mgr
	for i, api := range apis {
		mgr.AddRateLimitPerApi(api, &settings[i%len(settings)], settingsNames[i%len(settingsNames)])
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
			rateLimitInfo := &LimiterInfo{
				Api: apis[rand.Intn(len(apis))],
				Ip:  fmt.Sprintf("client%d", idx),
			}
			ctx = NewLimiterInfoContext(ctx, rateLimitInfo)
			limit, err := mgr.Limit(ctx)
			require.Nil(t, err)
			require.False(t, limit)
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
			rateLimitInfo := &LimiterInfo{
				Api: apis[0],
				Ip:  fmt.Sprintf("client%d", idx),
			}
			ctx = NewLimiterInfoContext(ctx, rateLimitInfo)
			limit, e := mgr.Limit(ctx)
			if limit && e != nil {
				err = e
			}
		}(i)
		// Update settings midway through
		if i == numClients/2 {
			newCreateLimitSettings := &edgeproto.ApiEndpointRateLimitSettings{
				EndpointRateLimitSettings: &edgeproto.RateLimitSettings{
					MaxReqsRateLimitSettings: &edgeproto.MaxReqsRateLimitSettings{
						MaxReqsAlgorithm:     edgeproto.MaxReqsRateLimitAlgorithm_FIXED_WINDOW_ALGORITHM,
						MaxRequestsPerSecond: 1,
					},
				},
			}
			mgr.UpdateRateLimitSettings(newCreateLimitSettings, settingsNames[0])
		}
	}
	wg.Wait()
	require.NotNil(t, err)
	require.True(t, strings.Contains(err.Error(), "reached limit per second"))
}
