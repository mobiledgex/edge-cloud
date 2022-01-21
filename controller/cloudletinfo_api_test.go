package main

import (
	"context"
	"testing"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestCloudletInfo(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testSvcs := testinit(ctx, t)
	defer testfinish(testSvcs)

	dummy := dummyEtcd{}
	dummy.Start()
	defer dummy.Stop()

	sync := InitSync(&dummy)
	apis := NewAllApis(sync)
	sync.Start()
	defer sync.Done()

	// create supporting data
	testutil.InternalFlavorCreate(t, apis.flavorApi, testutil.FlavorData)
	testutil.InternalGPUDriverCreate(t, apis.gpuDriverApi, testutil.GPUDriverData)
	testutil.InternalResTagTableCreate(t, apis.resTagTableApi, testutil.ResTagTableData)
	testutil.InternalCloudletCreate(t, apis.cloudletApi, testutil.CloudletData())
	insertCloudletInfo(ctx, apis, testutil.CloudletInfoData)

	testutil.InternalCloudletInfoTest(t, "show", apis.cloudletInfoApi, testutil.CloudletInfoData)
	evictCloudletInfo(ctx, apis, testutil.CloudletInfoData)

	// test revision changes to cloudletinfo object on update
	testCloudletInfoRevs(t, ctx, &dummy, apis, testutil.CloudletInfoData)
}

func insertCloudletInfo(ctx context.Context, apis *AllApis, data []edgeproto.CloudletInfo) {
	for ii := range data {
		in := &data[ii]
		in.State = dme.CloudletState_CLOUDLET_STATE_READY
		apis.cloudletInfoApi.Update(ctx, in, 0)
	}
}

func evictCloudletInfo(ctx context.Context, apis *AllApis, data []edgeproto.CloudletInfo) {
	for ii := range data {
		in := &data[ii]
		apis.cloudletInfoApi.Delete(ctx, in, 0)
	}
}

func testCloudletInfoRevs(t *testing.T, ctx context.Context, dummy *dummyEtcd, apis *AllApis, data []edgeproto.CloudletInfo) {
	testData := &data[0]
	apis.cloudletInfoApi.Update(ctx, testData, 0)
	keyStr := objstore.DbKeyString("CloudletInfo", testData.GetKey())
	_, _, rev0, err := dummy.Get(keyStr)
	require.Nil(t, err)

	// updating state should update the object in etcd, revs should change
	testData.State = dme.CloudletState_CLOUDLET_STATE_OFFLINE
	apis.cloudletInfoApi.Update(ctx, testData, 0)
	_, _, rev1, err := dummy.Get(keyStr)
	require.Nil(t, err)
	require.Greater(t, rev1, rev0)

	// updating the same object should not affect revs
	apis.cloudletInfoApi.Update(ctx, testData, 0)
	_, _, rev2, err := dummy.Get(keyStr)
	require.Nil(t, err)
	require.Equal(t, rev1, rev2)

	// updating redis only field should not affect revs
	testData.Status = edgeproto.StatusInfo{MsgCount: 1}
	apis.cloudletInfoApi.Update(ctx, testData, 0)
	_, _, rev3, err := dummy.Get(keyStr)
	require.Nil(t, err)
	require.Equal(t, rev2, rev3)

	// updating maintenance state should change revs
	testData.MaintenanceState = dme.MaintenanceState_MAINTENANCE_START
	apis.cloudletInfoApi.Update(ctx, testData, 0)
	_, _, rev4, err := dummy.Get(keyStr)
	require.Nil(t, err)
	require.Greater(t, rev4, rev3)
}
