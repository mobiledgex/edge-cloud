package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestTrustPolicyExceptionApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testinit()
	defer testfinish()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	//testutil.InternalTrustPolicyExceptionTest(t, "cud", &trustPolicyExceptionApi, testutil.TrustPolicyExceptionData)
	// error cases
	expectCreatePolicyExceptionError(t, ctx, &testutil.TrustPolicyExceptionErrorData[0], "cannot be higher than max")
	expectCreatePolicyExceptionError(t, ctx, &testutil.TrustPolicyExceptionErrorData[1], "invalid CIDR")
	expectCreatePolicyExceptionError(t, ctx, &testutil.TrustPolicyExceptionErrorData[2], "Invalid min port")

	dummy.Stop()
}

func expectCreatePolicyExceptionError(t *testing.T, ctx context.Context, in *edgeproto.TrustPolicyException, msg string) {
	_, err := trustPolicyExceptionApi.CreateTrustPolicyException(ctx, in)
	require.NotNil(t, err, "create %v", in)
	require.Contains(t, err.Error(), msg, "error %v contains %s", err, msg)
}
