package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/rediscache"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestStreamObjs(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	// Test with dummy server
	testSvcs := testinit(ctx, t)
	testStreamObjsWithServer(t, ctx)
	testfinish(testSvcs)

	// Test with local server
	testSvcs = testinit(ctx, t, WithLocalRedis())
	testStreamObjsWithServer(t, ctx)
	testfinish(testSvcs)
}

func testStreamObjsWithServer(t *testing.T, ctx context.Context) {
	cctx := DefCallContext()
	streamKey := "testkey"
	streamObjApi := StreamObjApi{}

	for ii := 1; ii <= 500; ii++ {
		iterMsg := fmt.Sprintf("Iter[%d]", ii)

		sendObj, _, err := streamObjApi.startStream(ctx, cctx, streamKey, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, iterMsg)

		_, _, err = streamObjApi.startStream(ctx, cctx, streamKey, testutil.NewCudStreamoutAppInst(ctx))
		require.NotNil(t, err, iterMsg)
		require.Contains(t, err.Error(), "action is already in progress")

		err = streamObjApi.stopStream(ctx, cctx, streamKey, sendObj, nil, NoCleanupStream)
		require.Nil(t, err, iterMsg)

		// add message to stream to indicate that stream is being used by someother thread
		addMsgToRedisStream(ctx, streamKey, map[string]interface{}{
			StreamMsgTypeMessage: "Some message",
		})

		_, _, err = streamObjApi.startStream(ctx, cctx, streamKey, testutil.NewCudStreamoutAppInst(ctx))
		require.NotNil(t, err, iterMsg)
		require.Contains(t, err.Error(), "action is already in progress")

		// With CRM override, startStream should reset the stream
		cctx.Override = edgeproto.CRMOverride_IGNORE_TRANSIENT_STATE
		sendObj, _, err = streamObjApi.startStream(ctx, cctx, streamKey, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, iterMsg)
		cctx.Override = edgeproto.CRMOverride_NO_OVERRIDE

		err = streamObjApi.stopStream(ctx, cctx, streamKey, sendObj, nil, CleanupStream)
		require.Nil(t, err, iterMsg)
	}

	// Ensure that stream is cleaned up
	out, err := redisClient.Exists(streamKey).Result()
	require.Nil(t, err, "check if stream exists")
	require.Equal(t, int64(0), out, "stream should not exist")

	// Test for race issues
	// ====================
	// * Start multiple threads performing [StartStream + StopStream], if a stream is
	//   already in progress, then retry.
	// * Ensure that as part of stream, [SOM & EOM/Error] exists for all the threads in-order
	wg := sync.WaitGroup{}
	numThreads := 20
	for ii := 0; ii < numThreads; ii++ {
		wg.Add(1)
		go func(iter int) {
			defer wg.Done()
			var sendObj *streamSend
			var err error
			for jj := 0; jj < numThreads; jj++ {
				sendObj, _, err = streamObjApi.startStream(ctx, cctx, streamKey, testutil.NewCudStreamoutAppInst(ctx), WithNoResetStream())
				if err != nil {
					require.Contains(t, err.Error(), "action is already in progress")
					// retry, it must succeed in at least `numThreads` iterations
					time.Sleep(100 * time.Millisecond)
					continue
				}
				break
			}
			var objErr error
			// Alternatively introduce errors so that we can test for those as well
			if iter%2 == 0 {
				objErr = fmt.Errorf("Some error")
			}
			err = streamObjApi.stopStream(ctx, cctx, streamKey, sendObj, objErr, NoCleanupStream)
			require.Nil(t, err, "stop stream")
		}(ii)
	}
	wg.Wait()

	out, err = redisClient.Exists(streamKey).Result()
	require.Nil(t, err, "check if stream exists")
	require.Equal(t, int64(1), out, "stream should exist")

	streamMsgs, err := redisClient.XRange(streamKey, rediscache.RedisSmallestId, rediscache.RedisGreatestId).Result()
	require.Nil(t, err, "get stream messages")
	// [SOM + EOM/Error] per thread
	require.Equal(t, 2*numThreads, len(streamMsgs), "check if correct number of stream messages exists")

	start := true
	for _, sMsg := range streamMsgs {
		for k, _ := range sMsg.Values {
			if start {
				require.Equal(t, StreamMsgTypeSOM, k, "Start of message")
				start = false
			} else {
				if k != StreamMsgTypeEOM {
					require.Equal(t, StreamMsgTypeError, k, "Error message")
				} else {
					require.Equal(t, StreamMsgTypeEOM, k, "End of message")
				}
				start = true
			}
		}
	}

	// Cleanup stream
	keysRem, err := redisClient.Del(streamKey).Result()
	require.Nil(t, err, "delete stream")
	require.Equal(t, int64(1), keysRem, "stream deleted")
}
