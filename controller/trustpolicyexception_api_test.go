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
	testutil.InternalCloudletCreate(t, apis.cloudletApi, testutil.CloudletData())
	insertCloudletInfo(ctx, apis, testutil.CloudletInfoData)
	testutil.InternalAutoProvPolicyCreate(t, apis.autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalAutoScalePolicyCreate(t, apis.autoScalePolicyApi, testutil.AutoScalePolicyData)
	testutil.InternalAppCreate(t, apis.appApi, testutil.AppData)
	testutil.InternalClusterInstCreate(t, apis.clusterInstApi, testutil.ClusterInstData)
	testutil.InternalAppInstCreate(t, apis.appInstApi, testutil.AppInstData)
	testutil.InternalCloudletPoolTest(t, "cud", apis.cloudletPoolApi, testutil.CloudletPoolData)

	app := edgeproto.App{
		Key: edgeproto.AppKey{
			Organization: "org",
			Name:         "someapp",
			Version:      "1.0.1",
		},
		ImageType:          edgeproto.ImageType_IMAGE_TYPE_DOCKER,
		AccessPorts:        "tcp:445,udp:1212",
		Deployment:         "docker", // avoid trying to parse k8s manifest
		DeploymentManifest: "some manifest",
		DefaultFlavor:      testutil.FlavorData[2].Key,
	}
	_, err := apis.appApi.CreateApp(ctx, &app)
	require.Nil(t, err, "Create app with deployment manifest")
	checkApp := edgeproto.App{}
	found := apis.appApi.Get(&app.Key, &checkApp)
	require.True(t, found, "found app")
	require.Equal(t, "", checkApp.ImagePath, "image path empty")
	_, err = apis.appApi.DeleteApp(ctx, &app)
	require.Nil(t, err)

	// CUD for Trust Policy Exception
	testutil.InternalTrustPolicyExceptionTest(t, "cud", apis.trustPolicyExceptionApi, testutil.TrustPolicyExceptionData)

	_, err = apis.trustPolicyExceptionApi.CreateTrustPolicyException(ctx, &testutil.TrustPolicyExceptionData[0])
	require.NotNil(t, err)
	require.Contains(t, err.Error(), " already exists")

	tpeData := edgeproto.TrustPolicyException{
		Key: edgeproto.TrustPolicyExceptionKey{
			AppKey: app.Key,
			CloudletPoolKey: edgeproto.CloudletPoolKey{
				Organization: "Verizon",
				Name:         "test-and-dev",
			},
			Name: "someapp-tpe",
		},
		State: edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_APPROVAL_REQUESTED,
		OutboundSecurityRules: []edgeproto.SecurityRule{
			edgeproto.SecurityRule{
				Protocol:     "tcp",
				RemoteCidr:   "10.1.0.0/16",
				PortRangeMin: 201,
				PortRangeMax: 210,
			},
		},
	}
	_, err = apis.trustPolicyExceptionApi.CreateTrustPolicyException(ctx, &tpeData)
	require.NotNil(t, err)

	_, err = apis.trustPolicyExceptionApi.DeleteTrustPolicyException(ctx, &tpeData)

	// error cases for Trust Policy Exception
	expectCreatePolicyExceptionError(t, ctx, apis, &testutil.TrustPolicyExceptionErrorData[0], "cannot be higher than max")
	expectCreatePolicyExceptionError(t, ctx, apis, &testutil.TrustPolicyExceptionErrorData[1], "invalid CIDR")
	expectCreatePolicyExceptionError(t, ctx, apis, &testutil.TrustPolicyExceptionErrorData[2], "Invalid min port")
	expectCreatePolicyExceptionError(t, ctx, apis, &testutil.TrustPolicyExceptionErrorData[3], "App does not exist")
	expectCreatePolicyExceptionError(t, ctx, apis, &testutil.TrustPolicyExceptionErrorData[4], "CloudletPoolKey does not exist")

	dummy.Stop()
}

func expectCreatePolicyExceptionError(t *testing.T, ctx context.Context, apis *AllApis, in *edgeproto.TrustPolicyException, msg string) {
	_, err := apis.trustPolicyExceptionApi.CreateTrustPolicyException(ctx, in)
	require.NotNil(t, err, "create %v", in)
	require.Contains(t, err.Error(), msg, "error %v contains %s", err, msg)
}
