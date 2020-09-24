package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestPrivacyPolicyApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testinit()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	testutil.InternalPrivacyPolicyTest(t, "cud", &privacyPolicyApi, testutil.PrivacyPolicyData)
	// error cases
	expectCreatePolicyError(t, ctx, &testutil.PrivacyPolicyErrorData[0], "cannot be higher than max")
	expectCreatePolicyError(t, ctx, &testutil.PrivacyPolicyErrorData[1], "invalid CIDR")
	expectCreatePolicyError(t, ctx, &testutil.PrivacyPolicyErrorData[2], "Invalid min port range")

	dummy.Stop()
}

func expectCreatePolicyError(t *testing.T, ctx context.Context, in *edgeproto.PrivacyPolicy, msg string) {
	_, err := privacyPolicyApi.CreatePrivacyPolicy(ctx, in)
	require.NotNil(t, err, "create %v", in)
	require.Contains(t, err.Error(), msg, "error %v contains %s", err, msg)
}
