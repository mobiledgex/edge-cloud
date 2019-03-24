package main

import (
	"context"
	"strings"
	"testing"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAppInstApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	objstore.InitRegion(1)
	reduceInfoTimeouts()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()
	responder := NewDummyInfoResponder(&appInstApi.cache, &clusterInstApi.cache,
		&appInstInfoApi, &clusterInstInfoApi)

	// cannote create instances without apps and cloudlets
	for _, obj := range testutil.AppInstData {
		err := appInstApi.CreateAppInst(&obj, &testutil.CudStreamoutAppInst{})
		assert.NotNil(t, err, "Create app inst without apps/cloudlets")
	}

	// create supporting data
	testutil.InternalDeveloperCreate(t, &developerApi, testutil.DevData)
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalClusterFlavorCreate(t, &clusterFlavorApi, testutil.ClusterFlavorData)
	testutil.InternalOperatorCreate(t, &operatorApi, testutil.OperatorData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)
	insertCloudletInfo(testutil.CloudletInfoData)
	testutil.InternalClusterCreate(t, &clusterApi, testutil.ClusterData)
	testutil.InternalAppCreate(t, &appApi, testutil.AppData)
	testutil.InternalClusterInstCreate(t, &clusterInstApi, testutil.ClusterInstData)
	testutil.InternalCloudletRefsTest(t, "show", &cloudletRefsApi, testutil.CloudletRefsData)
	clusterInstCnt := len(clusterInstApi.cache.Objs)

	// Set responder to fail. This should clean up the object after
	// the fake crm returns a failure. If it doesn't, the next test to
	// create all the app insts will fail.
	responder.SetSimulateAppCreateFailure(true)
	for _, obj := range testutil.AppInstData {
		err := appInstApi.CreateAppInst(&obj, &testutil.CudStreamoutAppInst{})
		assert.NotNil(t, err, "Create app inst responder failures")
		// make sure error matches responder
		// if app-inst triggers auto-cluster, the error will be for a cluster
		if strings.Contains(err.Error(), "cluster inst") {
			assert.Equal(t, "Encountered failures: [crm create cluster inst failed]", err.Error())
		} else {
			assert.Equal(t, "Encountered failures: [crm create app inst failed]", err.Error())
		}
	}
	responder.SetSimulateAppCreateFailure(false)
	assert.Equal(t, 0, len(appInstApi.cache.Objs))
	assert.Equal(t, clusterInstCnt, len(clusterInstApi.cache.Objs))

	testutil.InternalAppInstTest(t, "cud", &appInstApi, testutil.AppInstData)
	InternalAppInstCachedFieldsTest(t)
	// check cluster insts created (includes explicit and auto)
	testutil.InternalClusterInstTest(t, "show", &clusterInstApi,
		append(testutil.ClusterInstData, testutil.ClusterInstAutoData...))

	// after app insts create, check that cloudlet refs data is correct.
	// Note this refs data is a second set after app insts were created.
	testutil.InternalCloudletRefsTest(t, "show", &cloudletRefsApi, testutil.CloudletRefsWithAppInstsData)

	commonApi := testutil.NewInternalAppInstApi(&appInstApi)

	// Set responder to fail delete.
	responder.SetSimulateAppDeleteFailure(true)
	obj := testutil.AppInstData[0]
	err := appInstApi.DeleteAppInst(&obj, &testutil.CudStreamoutAppInst{})
	assert.NotNil(t, err, "Delete AppInst responder failure")
	responder.SetSimulateAppDeleteFailure(false)
	checkAppInstState(t, commonApi, &obj, edgeproto.TrackedState_Ready)

	obj = testutil.AppInstData[0]
	// check override of error DeleteError
	err = forceAppInstState(&obj, edgeproto.TrackedState_DeleteError)
	assert.Nil(t, err, "force state")
	checkAppInstState(t, commonApi, &obj, edgeproto.TrackedState_DeleteError)
	err = appInstApi.CreateAppInst(&obj, &testutil.CudStreamoutAppInst{})
	assert.Nil(t, err, "create overrides delete error")
	checkAppInstState(t, commonApi, &obj, edgeproto.TrackedState_Ready)

	// check override of error CreateError
	err = forceAppInstState(&obj, edgeproto.TrackedState_CreateError)
	assert.Nil(t, err, "force state")
	checkAppInstState(t, commonApi, &obj, edgeproto.TrackedState_CreateError)
	err = appInstApi.DeleteAppInst(&obj, &testutil.CudStreamoutAppInst{})
	assert.Nil(t, err, "delete overrides create error")
	checkAppInstState(t, commonApi, &obj, edgeproto.TrackedState_NotPresent)

	// override CRM error
	responder.SetSimulateAppCreateFailure(true)
	responder.SetSimulateAppDeleteFailure(true)
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IgnoreCRMErrors
	err = appInstApi.CreateAppInst(&obj, &testutil.CudStreamoutAppInst{})
	assert.Nil(t, err, "override crm error")
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IgnoreCRMErrors
	err = appInstApi.DeleteAppInst(&obj, &testutil.CudStreamoutAppInst{})
	assert.Nil(t, err, "override crm error")

	// ignore CRM
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IgnoreCRM
	err = appInstApi.CreateAppInst(&obj, &testutil.CudStreamoutAppInst{})
	assert.Nil(t, err, "ignore crm")
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IgnoreCRM
	err = appInstApi.DeleteAppInst(&obj, &testutil.CudStreamoutAppInst{})
	assert.Nil(t, err, "ignore crm")
	responder.SetSimulateAppCreateFailure(false)
	responder.SetSimulateAppDeleteFailure(false)

	dummy.Stop()
}

func appInstCachedFieldsTest(t *testing.T, cAppApi *testutil.AppCommonApi, cCloudletApi *testutil.CloudletCommonApi, cAppInstApi *testutil.AppInstCommonApi) {
	// test assumes test data has already been loaded
	ctx := context.TODO()

	// update app and check that app insts are updated
	updater := edgeproto.App{}
	updater.Key = testutil.AppData[0].Key
	newPath := "resources: a new config"
	updater.Config = newPath
	updater.Fields = make([]string, 0)
	updater.Fields = append(updater.Fields, edgeproto.AppFieldConfig)
	_, err := cAppApi.UpdateApp(ctx, &updater)
	assert.Nil(t, err, "Update app")

	show := testutil.ShowAppInst{}
	show.Init()
	filter := edgeproto.AppInst{}
	filter.Key.AppKey = testutil.AppData[0].Key
	err = cAppInstApi.ShowAppInst(ctx, &filter, &show)
	assert.Nil(t, err, "show app inst data")
	assert.True(t, len(show.Data) > 0, "number of matching app insts")

	// update cloudlet and check that app insts are updated
	updater2 := edgeproto.Cloudlet{}
	updater2.Key = testutil.CloudletData[0].Key
	newLat := 52.84583
	updater2.Location.Latitude = newLat
	updater2.Fields = make([]string, 0)
	updater2.Fields = append(updater2.Fields, edgeproto.CloudletFieldLocationLatitude)
	_, err = cCloudletApi.UpdateCloudlet(ctx, &updater2)
	assert.Nil(t, err, "Update cloudlet")

	show.Init()
	filter = edgeproto.AppInst{}
	filter.Key.CloudletKey = testutil.CloudletData[0].Key
	err = cAppInstApi.ShowAppInst(ctx, &filter, &show)
	assert.Nil(t, err, "show app inst data")
	for _, inst := range show.Data {
		assert.Equal(t, newLat, inst.CloudletLoc.Latitude, "check app inst latitude")
	}
	assert.True(t, len(show.Data) > 0, "number of matching app insts")
}

func InternalAppInstCachedFieldsTest(t *testing.T) {
	cAppApi := testutil.NewInternalAppApi(&appApi)
	cCloudletApi := testutil.NewInternalCloudletApi(&cloudletApi)
	cAppInstApi := testutil.NewInternalAppInstApi(&appInstApi)
	appInstCachedFieldsTest(t, cAppApi, cCloudletApi, cAppInstApi)
}

func ClientAppInstCachedFieldsTest(t *testing.T, appApi edgeproto.AppApiClient, cloudletApi edgeproto.CloudletApiClient, appInstApi edgeproto.AppInstApiClient) {
	cAppApi := testutil.NewClientAppApi(appApi)
	cCloudletApi := testutil.NewClientCloudletApi(cloudletApi)
	cAppInstApi := testutil.NewClientAppInstApi(appInstApi)
	appInstCachedFieldsTest(t, cAppApi, cCloudletApi, cAppInstApi)
}

func TestAutoClusterInst(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	objstore.InitRegion(1)
	reduceInfoTimeouts()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()
	NewDummyInfoResponder(&appInstApi.cache, &clusterInstApi.cache,
		&appInstInfoApi, &clusterInstInfoApi)

	// create supporting data
	testutil.InternalDeveloperCreate(t, &developerApi, testutil.DevData)
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalClusterFlavorCreate(t, &clusterFlavorApi, testutil.ClusterFlavorData)
	testutil.InternalOperatorCreate(t, &operatorApi, testutil.OperatorData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)
	insertCloudletInfo(testutil.CloudletInfoData)
	testutil.InternalClusterCreate(t, &clusterApi, testutil.ClusterData)
	testutil.InternalAppCreate(t, &appApi, testutil.AppData)

	// since cluster inst does not exist, it will be auto-created
	copy := testutil.AppInstData[0]
	copy.ClusterInstKey.ClusterKey.Name = ClusterAutoPrefix
	err := appInstApi.CreateAppInst(&copy, &testutil.CudStreamoutAppInst{})
	assert.Nil(t, err, "create app inst")
	clusterInst := edgeproto.ClusterInst{}
	found := clusterInstApi.Get(&copy.ClusterInstKey, &clusterInst)
	assert.True(t, found, "get auto-clusterinst")
	assert.True(t, clusterInst.Auto, "clusterinst is auto")
	// delete appinst should also delete clusterinst
	err = appInstApi.DeleteAppInst(&copy, &testutil.CudStreamoutAppInst{})
	assert.Nil(t, err, "delete app inst")
	found = clusterInstApi.Get(&copy.ClusterInstKey, &clusterInst)
	assert.False(t, found, "get auto-clusterinst")

	dummy.Stop()
}

func checkAppInstState(t *testing.T, api *testutil.AppInstCommonApi, in *edgeproto.AppInst, state edgeproto.TrackedState) {
	out := edgeproto.AppInst{}
	found := testutil.GetAppInst(t, api, &in.Key, &out)
	if state == edgeproto.TrackedState_NotPresent {
		assert.False(t, found, "get app inst")
	} else {
		assert.True(t, found, "get app inst")
		assert.Equal(t, state, out.State, "app inst state")
	}
}

func forceAppInstState(in *edgeproto.AppInst, state edgeproto.TrackedState) error {
	err := appInstApi.sync.ApplySTMWait(func(stm concurrency.STM) error {
		obj := edgeproto.AppInst{}
		if !appInstApi.store.STMGet(stm, &in.Key, &obj) {
			return objstore.ErrKVStoreKeyNotFound
		}
		obj.State = state
		appInstApi.store.STMPut(stm, &obj)
		return nil
	})
	return err
}
