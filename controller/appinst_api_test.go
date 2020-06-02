package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/require"
)

func TestAppInstApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer("")
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

	// cannote create instances without apps and cloudlets
	for _, obj := range testutil.AppInstData {
		err := appInstApi.CreateAppInst(&obj, &testutil.CudStreamoutAppInst{})
		require.NotNil(t, err, "Create app inst without apps/cloudlets")
	}

	// create supporting data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)
	insertCloudletInfo(ctx, testutil.CloudletInfoData)
	testutil.InternalAutoProvPolicyCreate(t, &autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalAutoScalePolicyCreate(t, &autoScalePolicyApi, testutil.AutoScalePolicyData)
	testutil.InternalAppCreate(t, &appApi, testutil.AppData)
	testutil.InternalClusterInstCreate(t, &clusterInstApi, testutil.ClusterInstData)
	testutil.InternalCloudletRefsTest(t, "show", &cloudletRefsApi, testutil.CloudletRefsData)
	clusterInstCnt := len(clusterInstApi.cache.Objs)

	// Set responder to fail. This should clean up the object after
	// the fake crm returns a failure. If it doesn't, the next test to
	// create all the app insts will fail.
	responder.SetSimulateAppCreateFailure(true)
	// clean up on failure may find ports inconsistent
	RequireAppInstPortConsistency = false
	for _, obj := range testutil.AppInstData {
		if testutil.IsAutoClusterAutoDeleteApp(&obj.Key) {
			continue
		}
		err := appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.NotNil(t, err, "Create app inst responder failures")
		// make sure error matches responder
		// if app-inst triggers auto-cluster, the error will be for a cluster
		if strings.Contains(err.Error(), "cluster inst") {
			require.Equal(t, "Encountered failures: crm create cluster inst failed", err.Error())
		} else {
			require.Equal(t, "Encountered failures: crm create app inst failed", err.Error())
		}
	}
	responder.SetSimulateAppCreateFailure(false)
	RequireAppInstPortConsistency = true
	require.Equal(t, 0, len(appInstApi.cache.Objs))
	require.Equal(t, clusterInstCnt, len(clusterInstApi.cache.Objs))

	testutil.InternalAppInstTest(t, "cud", &appInstApi, testutil.AppInstData)
	InternalAppInstCachedFieldsTest(t, ctx)
	// check cluster insts created (includes explicit and auto)
	testutil.InternalClusterInstTest(t, "show", &clusterInstApi,
		append(testutil.ClusterInstData, testutil.ClusterInstAutoData...))

	// after app insts create, check that cloudlet refs data is correct.
	// Note this refs data is a second set after app insts were created.
	testutil.InternalCloudletRefsTest(t, "show", &cloudletRefsApi, testutil.CloudletRefsWithAppInstsData)
	testutil.InternalAppInstRefsTest(t, "show", &appInstRefsApi, testutil.AppInstRefsData)

	commonApi := testutil.NewInternalAppInstApi(&appInstApi)

	// Set responder to fail delete.
	responder.SetSimulateAppDeleteFailure(true)
	obj := testutil.AppInstData[0]
	err := appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.NotNil(t, err, "Delete AppInst responder failure")
	responder.SetSimulateAppDeleteFailure(false)
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_READY)
	testutil.InternalAppInstRefsTest(t, "show", &appInstRefsApi, testutil.AppInstRefsData)

	obj = testutil.AppInstData[0]
	// check override of error DELETE_ERROR
	err = forceAppInstState(ctx, &obj, edgeproto.TrackedState_DELETE_ERROR)
	require.Nil(t, err, "force state")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_DELETE_ERROR)
	err = appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "create overrides delete error")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_READY)
	testutil.InternalAppInstRefsTest(t, "show", &appInstRefsApi, testutil.AppInstRefsData)

	// check override of error CREATE_ERROR
	err = forceAppInstState(ctx, &obj, edgeproto.TrackedState_CREATE_ERROR)
	require.Nil(t, err, "force state")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_CREATE_ERROR)
	err = appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "delete overrides create error")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_NOT_PRESENT)
	// create copy of refs without the deleted AppInst
	appInstRefsDeleted := append([]edgeproto.AppInstRefs{}, testutil.AppInstRefsData...)
	appInstRefsDeleted[0].Insts = make(map[string]uint32)
	for k, v := range testutil.AppInstRefsData[0].Insts {
		if k == obj.Key.GetKeyString() {
			continue
		}
		appInstRefsDeleted[0].Insts[k] = v
	}
	testutil.InternalAppInstRefsTest(t, "show", &appInstRefsApi, appInstRefsDeleted)

	// check override of error UPDATE_ERROR
	err = appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "create appinst")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_READY)
	err = forceAppInstState(ctx, &obj, edgeproto.TrackedState_UPDATE_ERROR)
	require.Nil(t, err, "force state")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_UPDATE_ERROR)
	err = appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "delete overrides create error")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_NOT_PRESENT)
	testutil.InternalAppInstRefsTest(t, "show", &appInstRefsApi, appInstRefsDeleted)

	// override CRM error
	responder.SetSimulateAppCreateFailure(true)
	responder.SetSimulateAppDeleteFailure(true)
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "override crm error")
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "override crm error")

	// ignore CRM
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
	err = appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "ignore crm")
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
	err = appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "ignore crm")
	responder.SetSimulateAppCreateFailure(false)
	responder.SetSimulateAppDeleteFailure(false)

	// ignore CRM and transient state on delete of AppInst
	responder.SetSimulateAppDeleteFailure(true)
	for val, stateName := range edgeproto.TrackedState_name {
		state := edgeproto.TrackedState(val)
		if !edgeproto.IsTransientState(state) {
			continue
		}
		obj = testutil.AppInstData[0]
		err = appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "create AppInst")
		err = forceAppInstState(ctx, &obj, state)
		require.Nil(t, err, "force state")
		checkAppInstState(t, ctx, commonApi, &obj, state)
		obj = testutil.AppInstData[0]
		obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_AND_TRANSIENT_STATE
		err = appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "override crm and transient state %s", stateName)
	}
	responder.SetSimulateAppDeleteFailure(false)

	testAppInstOverrideTransientDelete(t, ctx, commonApi, responder)

	// Test Fqdn prefix
	for _, data := range appInstApi.cache.Objs {
		obj := data.Obj
		app_name := util.K8SSanitize(obj.Key.AppKey.Name)
		if app_name == "helmapp" {
			continue
		}
		for _, port := range obj.MappedPorts {
			lproto, err := edgeproto.LProtoStr(port.Proto)
			if err != nil {
				continue
			}
			if lproto == "http" {
				continue
			}
			test_prefix := fmt.Sprintf("%s-%s.", app_name, lproto)
			require.Equal(t, test_prefix, port.FqdnPrefix, "check port fqdn prefix")
		}
	}
	testAppFlavorRequest(t, ctx, commonApi, responder)

	// delete all AppInsts and Apps and check that refs are empty
	for _, obj := range testutil.AppInstData {
		if testutil.IsAutoClusterAutoDeleteApp(&obj.Key) {
			continue
		}
		err := appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		if err != nil && err.Error() == obj.Key.NotFoundError().Error() {
			continue
		}
		require.Nil(t, err, "Delete app inst failed")
	}
	for _, obj := range testutil.AppData {
		_, err := appApi.DeleteApp(ctx, &obj)
		if err != nil && err.Error() == obj.Key.NotFoundError().Error() {
			continue
		}
		require.Nil(t, err, "Delete app failed")
	}
	testutil.InternalAppInstRefsTest(t, "show", &appInstRefsApi, []edgeproto.AppInstRefs{})

	dummy.Stop()
}

func appInstCachedFieldsTest(t *testing.T, ctx context.Context, cAppApi *testutil.AppCommonApi, cCloudletApi *testutil.CloudletCommonApi, cAppInstApi *testutil.AppInstCommonApi) {
	// test assumes test data has already been loaded

	// update app and check that app insts are updated
	updater := edgeproto.App{}
	updater.Key = testutil.AppData[0].Key
	newPath := "resources: a new config"
	updater.AndroidPackageName = newPath
	updater.Fields = make([]string, 0)
	updater.Fields = append(updater.Fields, edgeproto.AppFieldAndroidPackageName)
	_, err := cAppApi.UpdateApp(ctx, &updater)
	require.Nil(t, err, "Update app")

	show := testutil.ShowAppInst{}
	show.Init()
	filter := edgeproto.AppInst{}
	filter.Key.AppKey = testutil.AppData[0].Key
	err = cAppInstApi.ShowAppInst(ctx, &filter, &show)
	require.Nil(t, err, "show app inst data")
	require.True(t, len(show.Data) > 0, "number of matching app insts")

	// update cloudlet and check that app insts are updated
	updater2 := edgeproto.Cloudlet{}
	updater2.Key = testutil.CloudletData[0].Key
	newLat := 52.84583
	updater2.Location.Latitude = newLat
	updater2.Fields = make([]string, 0)
	updater2.Fields = append(updater2.Fields, edgeproto.CloudletFieldLocationLatitude)
	_, err = cCloudletApi.UpdateCloudlet(ctx, &updater2)
	require.Nil(t, err, "Update cloudlet")

	show.Init()
	filter = edgeproto.AppInst{}
	filter.Key.ClusterInstKey.CloudletKey = testutil.CloudletData[0].Key
	err = cAppInstApi.ShowAppInst(ctx, &filter, &show)
	require.Nil(t, err, "show app inst data")
	for _, inst := range show.Data {
		require.Equal(t, newLat, inst.CloudletLoc.Latitude, "check app inst latitude")
	}
	require.True(t, len(show.Data) > 0, "number of matching app insts")
}

func InternalAppInstCachedFieldsTest(t *testing.T, ctx context.Context) {
	cAppApi := testutil.NewInternalAppApi(&appApi)
	cCloudletApi := testutil.NewInternalCloudletApi(&cloudletApi)
	cAppInstApi := testutil.NewInternalAppInstApi(&appInstApi)
	appInstCachedFieldsTest(t, ctx, cAppApi, cCloudletApi, cAppInstApi)
}

func ClientAppInstCachedFieldsTest(t *testing.T, ctx context.Context, appApi edgeproto.AppApiClient, cloudletApi edgeproto.CloudletApiClient, appInstApi edgeproto.AppInstApiClient) {
	cAppApi := testutil.NewClientAppApi(appApi)
	cCloudletApi := testutil.NewClientCloudletApi(cloudletApi)
	cAppInstApi := testutil.NewClientAppInstApi(appInstApi)
	appInstCachedFieldsTest(t, ctx, cAppApi, cCloudletApi, cAppInstApi)
}

func TestAutoClusterInst(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer("")
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

	reduceInfoTimeouts(t, ctx)
	InfluxUsageUnitTestSetup(t)

	// create supporting data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)
	insertCloudletInfo(ctx, testutil.CloudletInfoData)
	testutil.InternalAutoProvPolicyCreate(t, &autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalAppCreate(t, &appApi, testutil.AppData)

	// since cluster inst does not exist, it will be auto-created
	copy := testutil.AppInstData[0]
	copy.Key.ClusterInstKey.ClusterKey.Name = ClusterAutoPrefix
	err := appInstApi.CreateAppInst(&copy, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "create app inst")
	clusterInst := edgeproto.ClusterInst{}
	found := clusterInstApi.Get(&copy.Key.ClusterInstKey, &clusterInst)
	require.True(t, found, "get auto-clusterinst")
	require.True(t, clusterInst.Auto, "clusterinst is auto")
	// delete appinst should also delete clusterinst
	err = appInstApi.DeleteAppInst(&copy, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "delete app inst")
	found = clusterInstApi.Get(&copy.Key.ClusterInstKey, &clusterInst)
	require.False(t, found, "get auto-clusterinst")
	// Autocluster AppInst with AutoDelete delete option should fail
	autoDeleteAppInst := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:         testutil.AppData[9].Key,
			ClusterInstKey: testutil.ClusterInstData[0].Key,
		},
	}
	autoDeleteAppInst.Key.ClusterInstKey.ClusterKey.Name = ClusterAutoPrefix
	err = appInstApi.CreateAppInst(&autoDeleteAppInst, testutil.NewCudStreamoutAppInst(ctx))
	require.NotNil(t, err, "create autodelete apppInst")
	require.Contains(t, err.Error(), "requires an existing ClusterInst")
	dummy.Stop()
}

func checkAppInstState(t *testing.T, ctx context.Context, api *testutil.AppInstCommonApi, in *edgeproto.AppInst, state edgeproto.TrackedState) {
	out := edgeproto.AppInst{}
	found := testutil.GetAppInst(t, ctx, api, &in.Key, &out)
	if state == edgeproto.TrackedState_NOT_PRESENT {
		require.False(t, found, "get app inst")
	} else {
		require.True(t, found, "get app inst")
		require.Equal(t, state, out.State, "app inst state")
	}
}

func forceAppInstState(ctx context.Context, in *edgeproto.AppInst, state edgeproto.TrackedState) error {
	err := appInstApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		obj := edgeproto.AppInst{}
		if !appInstApi.store.STMGet(stm, &in.Key, &obj) {
			return in.Key.NotFoundError()
		}
		obj.State = state
		appInstApi.store.STMPut(stm, &obj)
		return nil
	})
	return err
}

func testAppFlavorRequest(t *testing.T, ctx context.Context, api *testutil.AppInstCommonApi, responder *DummyInfoResponder) {
	// Non-nomial test, request an optional resource from a cloudlet that offers none.
	var testflavor = edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "x1.large-mex",
		},
		Ram:       8192,
		Vcpus:     8,
		Disk:      40,
		OptResMap: map[string]string{"gpu": "gpu:1"},
	}
	_, err := flavorApi.CreateFlavor(ctx, &testflavor)
	require.Nil(t, err, "CreateFlavor")
	nonNomApp := testutil.AppInstData[0]
	nonNomApp.Flavor = testflavor.Key
	err = appInstApi.CreateAppInst(&nonNomApp, testutil.NewCudStreamoutAppInst(ctx))
	require.NotNil(t, err, "non-nom-app-create")
	require.Equal(t, "Optional resource requested by x1.large-mex, cloudlet San Jose Site supports none", err.Error())
}

// Test that Crm Override for Delete App overrides any failures
// on both side-car auto-apps and an underlying auto-cluster.
func testAppInstOverrideTransientDelete(t *testing.T, ctx context.Context, api *testutil.AppInstCommonApi, responder *DummyInfoResponder) {
	// autocluster
	ac := edgeproto.ClusterInst{
		Key: edgeproto.ClusterInstKey{
			ClusterKey: edgeproto.ClusterKey{
				Name: util.K8SSanitize(ClusterAutoPrefix + "override-clust"),
			},
			CloudletKey:  testutil.CloudletData[1].Key,
			Organization: testutil.AppData[0].Key.Organization,
		},
	}
	// autocluster app
	ai := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:         testutil.AppData[0].Key,
			ClusterInstKey: ac.Key,
		},
	}
	// autoapp
	require.Equal(t, edgeproto.DeleteType_AUTO_DELETE, testutil.AppData[9].DelOpt)
	aiauto := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:         testutil.AppData[9].Key, // auto-delete app
			ClusterInstKey: ac.Key,
		},
	}
	var err error
	var obj edgeproto.AppInst
	var clust edgeproto.ClusterInst
	clustApi := testutil.NewInternalClusterInstApi(&clusterInstApi)

	responder.SetSimulateAppDeleteFailure(true)
	responder.SetSimulateClusterDeleteFailure(true)
	for val, stateName := range edgeproto.TrackedState_name {
		state := edgeproto.TrackedState(val)
		if !edgeproto.IsTransientState(state) {
			continue
		}
		// create app (also creates clusterinst)
		obj = ai
		err = appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "create AppInst")
		err = forceAppInstState(ctx, &obj, state)
		require.Nil(t, err, "force state")
		checkAppInstState(t, ctx, api, &obj, state)

		// create auto app
		obj = aiauto
		err = appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "create AppInst")
		err = forceAppInstState(ctx, &obj, state)
		require.Nil(t, err, "force state")
		checkAppInstState(t, ctx, api, &obj, state)

		clust = ac
		err = forceClusterInstState(ctx, &clust, state)
		require.Nil(t, err, "force state")
		checkClusterInstState(t, ctx, clustApi, &clust, state)

		// delete app (should delete auto cluster and auto app)
		obj = ai
		obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_AND_TRANSIENT_STATE
		err = appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "override crm and transient state %s", stateName)
		// make sure autocluster got deleted (means apps also were deleted)
		showData := testutil.ShowClusterInst{}
		showData.Init()
		showData.Ctx = ctx
		err = clustApi.ShowClusterInst(ctx, &edgeproto.ClusterInst{}, &showData)
		require.Nil(t, err, "show ClusterInst")
		showData.AssertNotFound(t, &clust)
	}

	responder.SetSimulateAppDeleteFailure(false)
	responder.SetSimulateClusterDeleteFailure(false)

}
