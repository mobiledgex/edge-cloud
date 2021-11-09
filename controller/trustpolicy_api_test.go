package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestTrustPolicyApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testinit()
	defer testfinish()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	apis := NewAllApis(sync)
	sync.Start()
	defer sync.Done()

	testutil.InternalTrustPolicyTest(t, "cud", apis.trustPolicyApi, testutil.TrustPolicyData)
	// error cases
	expectCreatePolicyError(t, ctx, apis, &testutil.TrustPolicyErrorData[0], "cannot be higher than max")
	expectCreatePolicyError(t, ctx, apis, &testutil.TrustPolicyErrorData[1], "invalid CIDR")
	expectCreatePolicyError(t, ctx, apis, &testutil.TrustPolicyErrorData[2], "Invalid min port: 0")

	dummy.Stop()
}

func expectCreatePolicyError(t *testing.T, ctx context.Context, apis *AllApis, in *edgeproto.TrustPolicy, msg string) {
	err := apis.trustPolicyApi.CreateTrustPolicy(in, testutil.NewCudStreamoutTrustPolicy(ctx))
	require.NotNil(t, err, "create %v", in)
	require.Contains(t, err.Error(), msg, "error %v contains %s", err, msg)
}
