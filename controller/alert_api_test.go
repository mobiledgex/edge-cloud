package main

import (
	"context"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	influxq "github.com/mobiledgex/edge-cloud/controller/influxq_client"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/stretchr/testify/require"
)

func TestAlertApi(t *testing.T) {
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

	for _, alert := range testutil.AlertData {
		alertApi.Update(ctx, &alert, 0)
	}
	testutil.InternalAlertTest(t, "show", &alertApi, testutil.AlertData)

	dummy.Stop()
}

func TestAppInstDownAlert(t *testing.T) {
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
	NewDummyInfoResponder(&appInstApi.cache, &clusterInstApi.cache,
		&appInstInfoApi, &clusterInstInfoApi)

	// create supporting data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)
	insertCloudletInfo(ctx, testutil.CloudletInfoData)
	testutil.InternalAutoProvPolicyCreate(t, &autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalAutoScalePolicyCreate(t, &autoScalePolicyApi, testutil.AutoScalePolicyData)
	testutil.InternalAppCreate(t, &appApi, testutil.AppData)
	testutil.InternalClusterInstCreate(t, &clusterInstApi, testutil.ClusterInstData)
	testutil.InternalAppInstCreate(t, &appInstApi, testutil.AppInstData)
	// Create a reservable clusterInst
	cinst := testutil.ClusterInstData[7]
	streamOut := testutil.NewCudStreamoutAppInst(ctx)
	appinst := edgeproto.AppInst{}
	appinst.Key.AppKey = testutil.AppData[0].Key
	appinst.Key.ClusterInstKey = *cinst.Key.Virtual("")
	err := appInstApi.CreateAppInst(&appinst, streamOut)
	require.Nil(t, err, "create AppInst")
	// Inject AppInst info check that all appInsts are Healthy
	for ii, _ := range testutil.AppInstInfoData {
		in := &testutil.AppInstInfoData[ii]
		appInstInfoApi.Update(ctx, in, 0)
	}
	for _, val := range appInstApi.cache.Objs {
		require.Equal(t, dme.HealthCheck_HEALTH_CHECK_OK, val.Obj.HealthCheck)
	}
	// Trigger Alerts
	for _, alert := range testutil.AlertData {
		alertApi.Update(ctx, &alert, 0)
	}
	// Check reservable cluster

	found := appInstApi.Get(&appinst.Key, &appinst)
	require.True(t, found)
	require.Equal(t, dme.HealthCheck_HEALTH_CHECK_FAIL_ROOTLB_OFFLINE, appinst.HealthCheck)
	// check other appInstances
	for ii, testData := range testutil.AppInstData {
		found = appInstApi.Get(&testData.Key, &appinst)
		require.True(t, found)
		if ii == 0 {
			require.Equal(t, dme.HealthCheck_HEALTH_CHECK_FAIL_SERVER_FAIL, appinst.HealthCheck)
		} else {
			require.Equal(t, dme.HealthCheck_HEALTH_CHECK_OK, appinst.HealthCheck)
		}
	}

	dummy.Stop()
}

// Set up globals for API unit tests
func testinit() {
	objstore.InitRegion(1)
	tMode := true
	testMode = &tMode
	dockerRegistry := "docker.mobiledgex.net"
	registryFQDN = &dockerRegistry
	vaultConfig, _ = vault.BestConfig("")
	services.events = influxq.NewInfluxQ("events", "user", "pass")
	cleanupCloudletInfoTimeout = 100 * time.Millisecond
	RequireAppInstPortConsistency = true
	cplookup := &node.CloudletPoolCache{}
	cplookup.Init()
	nodeMgr.CloudletPoolLookup = cplookup
}
