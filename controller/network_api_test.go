package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestNetworkApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testinit()
	// defer testfinish() TODO uncomment after PR 1493 merged
	dummy := dummyEtcd{}
	dummy.Start()
	defer dummy.Stop()
	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	testutil.InternalNetworkTest(t, "cud", &networkApi, testutil.NetworkData)
	// error cases
	expectCreateNetworkError(t, ctx, &testutil.NetworkErrorData[0], "Invalid route destination cidr")
	expectCreateNetworkError(t, ctx, &testutil.NetworkErrorData[1], "Invalid next hop")
	expectCreateNetworkError(t, ctx, &testutil.NetworkErrorData[2], "Invalid connection type")

}

func expectCreateNetworkError(t *testing.T, ctx context.Context, in *edgeproto.Network, msg string) {
	err := networkApi.CreateNetwork(in, testutil.NewCudStreamoutNetwork(ctx))
	require.NotNil(t, err, "create %v", in)
	require.Contains(t, err.Error(), msg, "error %v contains %s", err, msg)
}
