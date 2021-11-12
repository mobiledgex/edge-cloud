package main

import (
	"context"
	"testing"

	"github.com/coreos/etcd/clientv3/concurrency"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func testStreamObjExists(t *testing.T, ctx context.Context, apis *AllApis, streamKey *edgeproto.AppInstKey, exists bool) {
	err := apis.streamObjApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !apis.streamObjApi.store.STMGet(stm, streamKey, &edgeproto.StreamObj{}) {
			return streamKey.NotFoundError()
		}
		return nil
	})
	if exists {
		require.Nil(t, err, "stream obj exists")
		return
	} else {
		require.NotNil(t, err, "stream obj error")
		require.Equal(t, err.Error(), streamKey.NotFoundError().Error(), "stream obj doesnot exist")
	}
}

func TestStreamObjApi(t *testing.T) {
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

	reduceInfoTimeouts(t, ctx, apis)

	// create supporting data
	testutil.InternalFlavorCreate(t, apis.flavorApi, testutil.FlavorData)
	testutil.InternalAutoProvPolicyCreate(t, apis.autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalAutoScalePolicyCreate(t, apis.autoScalePolicyApi, testutil.AutoScalePolicyData)
	testutil.InternalAppCreate(t, apis.appApi, testutil.AppData)
	testutil.InternalGPUDriverCreate(t, apis.gpuDriverApi, testutil.GPUDriverData)

	// ensure that streamObj is cleaned up after removal of parent object
	exists := true
	cloudlet := testutil.CloudletData()[0]
	clStreamKey := edgeproto.GetStreamKeyFromCloudletKey(&cloudlet.Key)
	err := apis.cloudletApi.CreateCloudlet(&cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err, "create cloudlet")
	testStreamObjExists(t, ctx, apis, &clStreamKey, exists)
	clInfo := testutil.CloudletInfoData[0]
	clInfo.State = dme.CloudletState_CLOUDLET_STATE_READY
	apis.cloudletInfoApi.Update(ctx, &clInfo, 0)

	clusterInst := testutil.ClusterInstData[0]
	clusterStreamKey := edgeproto.AppInstKey{ClusterInstKey: *clusterInst.Key.Virtual("")}
	err = apis.clusterInstApi.CreateClusterInst(&clusterInst, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "create clusterinst")
	testStreamObjExists(t, ctx, apis, &clusterStreamKey, exists)

	appInst := testutil.AppInstData[0]
	err = apis.appInstApi.CreateAppInst(&appInst, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "create appinst")
	testStreamObjExists(t, ctx, apis, &appInst.Key, exists)

	err = apis.appInstApi.DeleteAppInst(&appInst, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "delete appinst")
	testStreamObjExists(t, ctx, apis, &appInst.Key, !exists)

	err = apis.clusterInstApi.DeleteClusterInst(&clusterInst, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "delete clusterinst")
	testStreamObjExists(t, ctx, apis, &clusterStreamKey, !exists)

	err = apis.cloudletApi.DeleteCloudlet(&cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err, "delete cloudlet")
	testStreamObjExists(t, ctx, apis, &clStreamKey, !exists)
}
