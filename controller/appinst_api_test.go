package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAppInstApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	objstore.InitRegion(1)

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	// cannote create instances without apps and cloudlets
	ctx := context.TODO()
	for _, obj := range testutil.AppInstData {
		_, err := appInstApi.CreateAppInst(ctx, &obj)
		assert.NotNil(t, err, "Create app inst without apps/cloudlets")
	}

	// create supporting data
	for _, obj := range testutil.DevData {
		_, err := developerApi.CreateDeveloper(ctx, &obj)
		assert.Nil(t, err, "Create developer")
	}
	for _, obj := range testutil.FlavorData {
		_, err := flavorApi.CreateFlavor(ctx, &obj)
		assert.Nil(t, err, "Create flavor")
	}
	for _, obj := range testutil.ClusterData {
		_, err := clusterApi.CreateCluster(ctx, &obj)
		assert.Nil(t, err, "Create cluster")
	}
	for _, obj := range testutil.AppData {
		_, err := appApi.CreateApp(ctx, &obj)
		assert.Nil(t, err, "Create app %v", obj)
	}
	for _, obj := range testutil.OperatorData {
		_, err := operatorApi.CreateOperator(ctx, &obj)
		assert.Nil(t, err, "Create operator")
	}
	for _, obj := range testutil.CloudletData {
		_, err := cloudletApi.CreateCloudlet(ctx, &obj)
		assert.Nil(t, err, "Create cloudlet")
	}
	for _, obj := range testutil.ClusterInstData {
		_, err := clusterInstApi.CreateClusterInst(ctx, &obj)
		assert.Nil(t, err, "Create clusterInst")
	}

	testutil.InternalAppInstCudTest(t, &appInstApi, testutil.AppInstData)
	InternalAppInstCachedFieldsTest(t)

	dummy.Stop()
}

func appInstCachedFieldsTest(t *testing.T, cAppApi *testutil.AppCommonApi, cCloudletApi *testutil.CloudletCommonApi, cAppInstApi *testutil.AppInstCommonApi) {
	// test assumes test data has already been loaded
	ctx := context.TODO()

	// update app and check that app insts are updated
	updater := edgeproto.App{}
	updater.Key = testutil.AppData[0].Key
	newPath := "a new config"
	updater.ConfigMap = newPath
	updater.Fields = make([]string, 0)
	updater.Fields = append(updater.Fields, edgeproto.AppFieldConfigMap)
	_, err := cAppApi.UpdateApp(ctx, &updater)
	assert.Nil(t, err, "Update app")

	show := testutil.ShowAppInst{}
	show.Init()
	filter := edgeproto.AppInst{}
	filter.Key.AppKey = testutil.AppData[0].Key
	err = cAppInstApi.ShowAppInst(ctx, &filter, &show)
	assert.Nil(t, err, "show app inst data")
	for _, inst := range show.Data {
		assert.Equal(t, newPath, inst.ConfigMap, "check app inst")
	}
	assert.True(t, len(show.Data) > 0, "number of matching app insts")

	// update cloudlet and check that app insts are updated
	updater2 := edgeproto.Cloudlet{}
	updater2.Key = testutil.CloudletData[0].Key
	newLat := 152.84583
	updater2.Location.Lat = newLat
	updater2.Fields = make([]string, 0)
	updater2.Fields = append(updater2.Fields, edgeproto.CloudletFieldLocationLat)
	_, err = cCloudletApi.UpdateCloudlet(ctx, &updater2)
	assert.Nil(t, err, "Update cloudlet")

	show.Init()
	filter = edgeproto.AppInst{}
	filter.Key.CloudletKey = testutil.CloudletData[0].Key
	err = cAppInstApi.ShowAppInst(ctx, &filter, &show)
	assert.Nil(t, err, "show app inst data")
	for _, inst := range show.Data {
		assert.Equal(t, newLat, inst.CloudletLoc.Lat, "check app inst latitude")
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

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	// create supporting data
	ctx := context.TODO()
	for _, obj := range testutil.DevData {
		_, err := developerApi.CreateDeveloper(ctx, &obj)
		assert.Nil(t, err, "Create developer")
	}
	for _, obj := range testutil.FlavorData {
		_, err := flavorApi.CreateFlavor(ctx, &obj)
		assert.Nil(t, err, "Create flavor")
	}
	for _, obj := range testutil.ClusterData {
		_, err := clusterApi.CreateCluster(ctx, &obj)
		assert.Nil(t, err, "Create cluster")
	}
	for _, obj := range testutil.AppData {
		_, err := appApi.CreateApp(ctx, &obj)
		assert.Nil(t, err, "Create app")
	}
	for _, obj := range testutil.OperatorData {
		_, err := operatorApi.CreateOperator(ctx, &obj)
		assert.Nil(t, err, "Create operator")
	}
	for _, obj := range testutil.CloudletData {
		_, err := cloudletApi.CreateCloudlet(ctx, &obj)
		assert.Nil(t, err, "Create cloudlet")
	}

	// since cluster inst does not exist, it will be auto-created
	_, err := appInstApi.CreateAppInst(ctx, &testutil.AppInstData[0])
	assert.Nil(t, err, "create app inst")
	clusterInst := edgeproto.ClusterInst{}
	found := clusterInstApi.Get(&testutil.AppInstData[0].ClusterInstKey, &clusterInst)
	assert.True(t, found, "get auto-clusterinst")
	assert.True(t, clusterInst.Auto, "clusterinst is auto")
	// delete appinst should also delete clusterinst
	_, err = appInstApi.DeleteAppInst(ctx, &testutil.AppInstData[0])
	assert.Nil(t, err, "delete app inst")
	found = clusterInstApi.Get(&testutil.AppInstData[0].ClusterInstKey, &clusterInst)
	assert.False(t, found, "get auto-clusterinst")

	dummy.Stop()
}
