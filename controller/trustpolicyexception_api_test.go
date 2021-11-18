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
	apis := NewAllApis(sync)
	sync.Start()
	defer sync.Done()

	NewDummyInfoResponder(&apis.appInstApi.cache, &apis.clusterInstApi.cache,
		apis.appInstInfoApi, apis.clusterInstInfoApi)

	// create supporting data
	testutil.InternalFlavorCreate(t, apis.flavorApi, testutil.FlavorData)
	testutil.InternalGPUDriverCreate(t, apis.gpuDriverApi, testutil.GPUDriverData)
	testutil.InternalResTagTableCreate(t, apis.resTagTableApi, testutil.ResTagTableData)
	testutil.InternalCloudletCreate(t, apis.cloudletApi, testutil.CloudletData())
	insertCloudletInfo(ctx, apis, testutil.CloudletInfoData)
	testutil.InternalAutoProvPolicyCreate(t, apis.autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalAutoScalePolicyCreate(t, apis.autoScalePolicyApi, testutil.AutoScalePolicyData)
	testutil.InternalAppCreate(t, apis.appApi, testutil.AppData)
	testutil.InternalClusterInstCreate(t, apis.clusterInstApi, testutil.ClusterInstData)
	testutil.InternalAppInstCreate(t, apis.appInstApi, testutil.AppInstData)
	testutil.InternalCloudletPoolTest(t, "cud", apis.cloudletPoolApi, testutil.CloudletPoolData)

	// CUD for Trust Policy Exception
	testutil.InternalTrustPolicyExceptionTest(t, "cud", apis.trustPolicyExceptionApi, testutil.TrustPolicyExceptionData)

	// error cases for Trust Policy Exception
	expectCreatePolicyExceptionError(t, ctx, apis, &testutil.TrustPolicyExceptionErrorData[0], "cannot be higher than max")
	expectCreatePolicyExceptionError(t, ctx, apis, &testutil.TrustPolicyExceptionErrorData[1], "invalid CIDR")
	expectCreatePolicyExceptionError(t, ctx, apis, &testutil.TrustPolicyExceptionErrorData[2], "Invalid min port")

	dummy.Stop()
}

func expectCreatePolicyExceptionError(t *testing.T, ctx context.Context, apis *AllApis, in *edgeproto.TrustPolicyException, msg string) {
	_, err := apis.trustPolicyExceptionApi.CreateTrustPolicyException(ctx, in)
	require.NotNil(t, err, "create %v", in)
	require.Contains(t, err.Error(), msg, "error %v contains %s", err, msg)
}
