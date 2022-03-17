package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestStreamObjs(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testSvcs := testinit(ctx, t)
	defer testfinish(testSvcs)

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
}
