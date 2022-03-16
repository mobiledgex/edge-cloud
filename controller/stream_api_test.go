package main

import (
	"context"
	"fmt"
	"testing"

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

	for ii := 1; ii <= 1000; ii++ {
		iterMsg := fmt.Sprintf("Iter[%d]", ii)

		sendObj, _, err := streamObjApi.startStream(ctx, cctx, streamKey, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, iterMsg)

		_, _, err = streamObjApi.startStream(ctx, cctx, streamKey, testutil.NewCudStreamoutAppInst(ctx))
		require.NotNil(t, err, iterMsg)
		require.Contains(t, err.Error(), "action is already in progress")

		err = streamObjApi.stopStream(ctx, cctx, streamKey, sendObj, nil)
		require.Nil(t, err, iterMsg)
	}
}
