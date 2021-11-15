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
	defer testfinish()
	dummy := dummyEtcd{}
	dummy.Start()
	defer dummy.Stop()
	sync := InitSync(&dummy)
	apis := NewAllApis(sync)
	sync.Start()
	defer sync.Done()

	testutil.InternalNetworkTest(t, "cud", apis.networkApi, testutil.NetworkData)
	// error cases
	expectCreateNetworkError(t, ctx, apis, &testutil.NetworkErrorData[0], "Invalid route destination cidr")
	expectCreateNetworkError(t, ctx, apis, &testutil.NetworkErrorData[1], "Invalid next hop")
	expectCreateNetworkError(t, ctx, apis, &testutil.NetworkErrorData[2], "Invalid connection type")

}

func expectCreateNetworkError(t *testing.T, ctx context.Context, apis *AllApis, in *edgeproto.Network, msg string) {
	err := apis.networkApi.CreateNetwork(in, testutil.NewCudStreamoutNetwork(ctx))
	require.NotNil(t, err, "create %v", in)
	require.Contains(t, err.Error(), msg, "error %v contains %s", err, msg)
}
