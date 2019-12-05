package main

import (
	"context"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestClusterInstApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi | log.DebugLevelNotify)
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testinit()
	reduceInfoTimeouts()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()
	responder := NewDummyInfoResponder(&appInstApi.cache, &clusterInstApi.cache,
		&appInstInfoApi, &clusterInstInfoApi)

	// cannot create insts without cluster/cloudlet
	for _, obj := range testutil.ClusterInstData {
		err := clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
		require.NotNil(t, err, "Create ClusterInst without cloudlet")
	}

	// create support data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalOperatorCreate(t, &operatorApi, testutil.OperatorData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)
	insertCloudletInfo(ctx, testutil.CloudletInfoData)
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

	dummy.Stop()
}

func reduceInfoTimeouts() {
	cloudcommon.CreateClusterInstTimeout = 1 * time.Second
	cloudcommon.UpdateClusterInstTimeout = 1 * time.Second
	cloudcommon.DeleteClusterInstTimeout = 1 * time.Second

	cloudcommon.CreateAppInstTimeout = 1 * time.Second
	cloudcommon.UpdateAppInstTimeout = 1 * time.Second
	cloudcommon.DeleteAppInstTimeout = 1 * time.Second
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
	checkReservedBy(t, ctx, api, &cinst.Key, appinst.Key.AppKey.DeveloperKey.Name)

	// Cannot create another AppInst on it from different developer
	appinst2 := edgeproto.AppInst{}
	appinst2.Key.AppKey = testutil.AppData[10].Key
	appinst2.Key.ClusterInstKey = cinst.Key
	require.NotEqual(t, appinst.Key.AppKey.DeveloperKey.Name, appinst2.Key.AppKey.DeveloperKey.Name)
	err = appInstApi.CreateAppInst(&appinst2, streamOut)
	require.NotNil(t, err, "create AppInst on already reserved ClusterInst")
	// Cannot create another AppInst on it from the same developer
	appinst3 := edgeproto.AppInst{}
	appinst3.Key.AppKey = testutil.AppData[1].Key
	appinst3.Key.ClusterInstKey = cinst.Key
	require.Equal(t, appinst.Key.AppKey.DeveloperKey.Name, appinst3.Key.AppKey.DeveloperKey.Name)
	err = appInstApi.CreateAppInst(&appinst3, streamOut)
	require.NotNil(t, err, "create AppInst on already reserved ClusterInst")

	// Make sure above changes have not affected ReservedBy setting
	checkReservedBy(t, ctx, api, &cinst.Key, appinst.Key.AppKey.DeveloperKey.Name)

	// Deleting AppInst should removed ReservedBy
	err = appInstApi.DeleteAppInst(&appinst, streamOut)
	require.Nil(t, err, "delete AppInst")
	checkReservedBy(t, ctx, api, &cinst.Key, "")

	// Can now create AppInst from different developer
	err = appInstApi.CreateAppInst(&appinst2, streamOut)
	require.Nil(t, err, "create AppInst on reservable ClusterInst")
	checkReservedBy(t, ctx, api, &cinst.Key, appinst2.Key.AppKey.DeveloperKey.Name)

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
	require.Equal(t, cloudcommon.DeveloperMobiledgeX, cinst.Key.Developer)
}
