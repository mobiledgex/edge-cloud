package main

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	influxq "github.com/mobiledgex/edge-cloud/controller/influxq_client"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestClusterInstApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi | log.DebugLevelNotify)
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
	responder := NewDummyInfoResponder(&appInstApi.cache, &clusterInstApi.cache,
		&appInstInfoApi, &clusterInstInfoApi)

	reduceInfoTimeouts(t, ctx)
	InfluxUsageUnitTestSetup(t)

	// cannot create insts without cluster/cloudlet
	for _, obj := range testutil.ClusterInstData {
		err := clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
		require.NotNil(t, err, "Create ClusterInst without cloudlet")
	}

	// create support data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)
	insertCloudletInfo(ctx, testutil.CloudletInfoData)
	testutil.InternalAutoProvPolicyCreate(t, &autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalAutoScalePolicyCreate(t, &autoScalePolicyApi, testutil.AutoScalePolicyData)

	// Set responder to fail. This should clean up the object after
	// the fake crm returns a failure. If it doesn't, the next test to
	// create all the cluster insts will fail.
	responder.SetSimulateClusterCreateFailure(true)
	for _, obj := range testutil.ClusterInstData {
		err := clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
		require.NotNil(t, err, "Create ClusterInst responder failures")
		// make sure error matches responder
		require.Equal(t, "Encountered failures: crm create ClusterInst failed", err.Error())
	}
	responder.SetSimulateClusterCreateFailure(false)
	require.Equal(t, 0, len(clusterInstApi.cache.Objs))

	testutil.InternalClusterInstTest(t, "cud", &clusterInstApi, testutil.ClusterInstData)
	// after cluster insts create, check that cloudlet refs data is correct.
	testutil.InternalCloudletRefsTest(t, "show", &cloudletRefsApi, testutil.CloudletRefsData)

	commonApi := testutil.NewInternalClusterInstApi(&clusterInstApi)

	// Set responder to fail delete.
	responder.SetSimulateClusterDeleteFailure(true)
	obj := testutil.ClusterInstData[0]
	err := clusterInstApi.DeleteClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.NotNil(t, err, "Delete ClusterInst responder failure")
	responder.SetSimulateClusterDeleteFailure(false)
	checkClusterInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_READY)

	// check override of error DELETE_ERROR
	err = forceClusterInstState(ctx, &obj, edgeproto.TrackedState_DELETE_ERROR)
	require.Nil(t, err, "force state")
	checkClusterInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_DELETE_ERROR)
	err = clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "create overrides delete error")
	checkClusterInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_READY)

	// check override of error CREATE_ERROR
	err = forceClusterInstState(ctx, &obj, edgeproto.TrackedState_CREATE_ERROR)
	require.Nil(t, err, "force state")
	checkClusterInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_CREATE_ERROR)
	err = clusterInstApi.DeleteClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "delete overrides create error")
	checkClusterInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_NOT_PRESENT)

	// test update of autoscale policy
	obj = testutil.ClusterInstData[0]
	obj.Key.Organization = testutil.AutoScalePolicyData[1].Key.Organization
	err = clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "create ClusterInst")
	check := edgeproto.ClusterInst{}
	found := clusterInstApi.cache.Get(&obj.Key, &check)
	require.True(t, found)
	require.Equal(t, 2, int(check.NumNodes))

	obj.AutoScalePolicy = testutil.AutoScalePolicyData[1].Key.Name
	obj.Fields = []string{edgeproto.ClusterInstFieldAutoScalePolicy}
	err = clusterInstApi.UpdateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err)
	check = edgeproto.ClusterInst{}
	found = clusterInstApi.cache.Get(&obj.Key, &check)
	require.True(t, found)
	require.Equal(t, testutil.AutoScalePolicyData[1].Key.Name, check.AutoScalePolicy)
	require.Equal(t, 4, int(check.NumNodes))

	// override CRM error
	responder.SetSimulateClusterCreateFailure(true)
	responder.SetSimulateClusterDeleteFailure(true)
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "override crm error")
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = clusterInstApi.DeleteClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "override crm error")

	// ignore CRM
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
	err = clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "ignore crm")
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
	err = clusterInstApi.DeleteClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "ignore crm")

	// inavailability of matching node flavor
	obj = testutil.ClusterInstData[0]
	obj.Flavor = testutil.FlavorData[0].Key
	err = clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.NotNil(t, err, "flavor not available")

	responder.SetSimulateClusterCreateFailure(false)
	responder.SetSimulateClusterDeleteFailure(false)

	testReservableClusterInst(t, ctx, commonApi)
	testClusterInstOverrideTransientDelete(t, ctx, commonApi, responder)

	dummy.Stop()
}

func reduceInfoTimeouts(t *testing.T, ctx context.Context) {
	settingsApi.initDefaults(ctx)

	settings, err := settingsApi.ShowSettings(ctx, &edgeproto.Settings{})
	require.Nil(t, err)

	settings.CreateClusterInstTimeout = edgeproto.Duration(1 * time.Second)
	settings.UpdateClusterInstTimeout = edgeproto.Duration(1 * time.Second)
	settings.DeleteClusterInstTimeout = edgeproto.Duration(1 * time.Second)
	settings.CreateAppInstTimeout = edgeproto.Duration(1 * time.Second)
	settings.UpdateAppInstTimeout = edgeproto.Duration(1 * time.Second)
	settings.DeleteAppInstTimeout = edgeproto.Duration(1 * time.Second)
	settings.CloudletMaintenanceTimeout = edgeproto.Duration(2 * time.Second)

	settings.Fields = []string{
		edgeproto.SettingsFieldCreateAppInstTimeout,
		edgeproto.SettingsFieldUpdateAppInstTimeout,
		edgeproto.SettingsFieldDeleteAppInstTimeout,
		edgeproto.SettingsFieldCreateClusterInstTimeout,
		edgeproto.SettingsFieldUpdateClusterInstTimeout,
		edgeproto.SettingsFieldDeleteClusterInstTimeout,
		edgeproto.SettingsFieldCloudletMaintenanceTimeout,
	}
	_, err = settingsApi.UpdateSettings(ctx, settings)
	require.Nil(t, err)

	updated, err := settingsApi.ShowSettings(ctx, &edgeproto.Settings{})
	updated.Fields = []string{}
	settings.Fields = []string{}
	require.Equal(t, settings, updated)
}

func checkClusterInstState(t *testing.T, ctx context.Context, api *testutil.ClusterInstCommonApi, in *edgeproto.ClusterInst, state edgeproto.TrackedState) {
	out := edgeproto.ClusterInst{}
	found := testutil.GetClusterInst(t, ctx, api, &in.Key, &out)
	if state == edgeproto.TrackedState_NOT_PRESENT {
		require.False(t, found, "get cluster inst")
	} else {
		require.True(t, found, "get cluster inst")
		require.Equal(t, state, out.State, "cluster inst state")
	}
}

func forceClusterInstState(ctx context.Context, in *edgeproto.ClusterInst, state edgeproto.TrackedState) error {
	err := clusterInstApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		obj := edgeproto.ClusterInst{}
		if !clusterInstApi.store.STMGet(stm, &in.Key, &obj) {
			return in.Key.NotFoundError()
		}
		obj.State = state
		clusterInstApi.store.STMPut(stm, &obj)
		return nil
	})
	return err
}

func testReservableClusterInst(t *testing.T, ctx context.Context, api *testutil.ClusterInstCommonApi) {
	cinst := testutil.ClusterInstData[7]
	checkReservedBy(t, ctx, api, &cinst.Key, "")

	// create test app
	for _, app := range testutil.AppData {
		_, err := appApi.CreateApp(ctx, &app)
		require.Nil(t, err, "create App")
	}

	// Should be able to create a developer AppInst on the ClusterInst
	streamOut := testutil.NewCudStreamoutAppInst(ctx)
	appinst := edgeproto.AppInst{}
	appinst.Key.AppKey = testutil.AppData[0].Key
	appinst.Key.ClusterInstKey = cinst.Key
	err := appInstApi.CreateAppInst(&appinst, streamOut)
	require.Nil(t, err, "create AppInst")
	checkReservedBy(t, ctx, api, &cinst.Key, appinst.Key.AppKey.Organization)

	// Cannot create another AppInst on it from different developer
	appinst2 := edgeproto.AppInst{}
	appinst2.Key.AppKey = testutil.AppData[10].Key
	appinst2.Key.ClusterInstKey = cinst.Key
	require.NotEqual(t, appinst.Key.AppKey.Organization, appinst2.Key.AppKey.Organization)
	err = appInstApi.CreateAppInst(&appinst2, streamOut)
	require.NotNil(t, err, "create AppInst on already reserved ClusterInst")
	// Cannot create another AppInst on it from the same developer
	appinst3 := edgeproto.AppInst{}
	appinst3.Key.AppKey = testutil.AppData[1].Key
	appinst3.Key.ClusterInstKey = cinst.Key
	require.Equal(t, appinst.Key.AppKey.Organization, appinst3.Key.AppKey.Organization)
	err = appInstApi.CreateAppInst(&appinst3, streamOut)
	require.NotNil(t, err, "create AppInst on already reserved ClusterInst")

	// Make sure above changes have not affected ReservedBy setting
	checkReservedBy(t, ctx, api, &cinst.Key, appinst.Key.AppKey.Organization)

	// Deleting AppInst should removed ReservedBy
	err = appInstApi.DeleteAppInst(&appinst, streamOut)
	require.Nil(t, err, "delete AppInst")
	checkReservedBy(t, ctx, api, &cinst.Key, "")

	// Can now create AppInst from different developer
	err = appInstApi.CreateAppInst(&appinst2, streamOut)
	require.Nil(t, err, "create AppInst on reservable ClusterInst")
	checkReservedBy(t, ctx, api, &cinst.Key, appinst2.Key.AppKey.Organization)

	// Delete AppInst
	err = appInstApi.DeleteAppInst(&appinst2, streamOut)
	require.Nil(t, err, "delete AppInst on reservable ClusterInst")
	checkReservedBy(t, ctx, api, &cinst.Key, "")
	// Delete App
	for _, app := range testutil.AppData {
		_, err = appApi.DeleteApp(ctx, &app)
		require.Nil(t, err, "delete App")
	}
	checkReservedBy(t, ctx, api, &cinst.Key, "")
}

func checkReservedBy(t *testing.T, ctx context.Context, api *testutil.ClusterInstCommonApi, key *edgeproto.ClusterInstKey, expected string) {
	cinst := edgeproto.ClusterInst{}
	found := testutil.GetClusterInst(t, ctx, api, key, &cinst)
	require.True(t, found, "get ClusterInst")
	require.True(t, cinst.Reservable)
	require.Equal(t, expected, cinst.ReservedBy)
	require.Equal(t, cloudcommon.OrganizationMobiledgeX, cinst.Key.Organization)
}

// Test that Crm Override for Delete ClusterInst overrides any failures
// on side-car auto-apps.
func testClusterInstOverrideTransientDelete(t *testing.T, ctx context.Context, api *testutil.ClusterInstCommonApi, responder *DummyInfoResponder) {
	clust := testutil.ClusterInstData[0]
	clust.Key.ClusterKey.Name = "crmoverride"

	// autoapp
	app := testutil.AppData[9] // auto-delete app
	require.Equal(t, edgeproto.DeleteType_AUTO_DELETE, app.DelOpt)
	_, err := appApi.CreateApp(ctx, &app)
	require.Nil(t, err, "create App")

	aiauto := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:         app.Key,
			ClusterInstKey: clust.Key,
		},
	}

	var obj edgeproto.ClusterInst
	var ai edgeproto.AppInst
	appCommon := testutil.NewInternalAppInstApi(&appInstApi)

	responder.SetSimulateClusterDeleteFailure(true)
	responder.SetSimulateAppDeleteFailure(true)
	for val, stateName := range edgeproto.TrackedState_name {
		state := edgeproto.TrackedState(val)
		if !edgeproto.IsTransientState(state) {
			continue
		}
		// create cluster
		obj = clust
		err = clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
		require.Nil(t, err, "create ClusterInst")
		// create autoapp

		ai = aiauto
		err = appInstApi.CreateAppInst(&ai, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "create auto AppInst")
		// force bad states
		err = forceAppInstState(ctx, &ai, state)
		require.Nil(t, err, "force state")
		checkAppInstState(t, ctx, appCommon, &ai, state)
		err = forceClusterInstState(ctx, &obj, state)
		require.Nil(t, err, "force state")
		checkClusterInstState(t, ctx, api, &obj, state)
		// delete cluster
		obj = clust
		obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_AND_TRANSIENT_STATE
		err = clusterInstApi.DeleteClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
		require.Nil(t, err, "override crm and transient state %s", stateName)
	}
	responder.SetSimulateClusterDeleteFailure(false)
	responder.SetSimulateAppDeleteFailure(false)

	_, err = appApi.DeleteApp(ctx, &app)
	require.Nil(t, err, "delete App")
}

func InfluxUsageUnitTestSetup(t *testing.T) {
	addr := "http://127.0.0.1:8086"

	// start influxd if not already running
	_, err := exec.Command("sh", "-c", "pgrep -x influxd").Output()
	if err != nil {
		p := process.Influx{}
		p.Common.Name = "influx-test"
		p.HttpAddr = addr
		p.DataDir = "/var/tmp/.influxdb"
		// start influx
		err = p.StartLocal("/var/tmp/influxdb.log",
			process.WithCleanStartup())
		require.Nil(t, err, "start InfluxDB server")
		defer p.StopLocal()
	}

	q := influxq.NewInfluxQ(cloudcommon.EventsDbName, "", "")
	err = q.Start(addr)
	require.Nil(t, err, "new influx q")
	defer q.Stop()
	services.events = q
}
