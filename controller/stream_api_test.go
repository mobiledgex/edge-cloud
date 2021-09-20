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

func testStreamObjExists(t *testing.T, ctx context.Context, streamKey *edgeproto.AppInstKey, exists bool) {
	err := streamObjApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !streamObjApi.store.STMGet(stm, streamKey, &edgeproto.StreamObj{}) {
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
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	NewDummyInfoResponder(&appInstApi.cache, &clusterInstApi.cache,
		&appInstInfoApi, &clusterInstInfoApi)

	reduceInfoTimeouts(t, ctx)

	// create supporting data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalAutoProvPolicyCreate(t, &autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalAutoScalePolicyCreate(t, &autoScalePolicyApi, testutil.AutoScalePolicyData)
	testutil.InternalAppCreate(t, &appApi, testutil.AppData)
	testutil.InternalGPUDriverCreate(t, &gpuDriverApi, testutil.GPUDriverData)

	// ensure that streamObj is cleaned up after removal of parent object
	exists := true
	cloudlet := testutil.CloudletData()[0]
	clStreamKey := edgeproto.GetStreamKeyFromCloudletKey(&cloudlet.Key)
	err := cloudletApi.CreateCloudlet(&cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err, "create cloudlet")
	testStreamObjExists(t, ctx, &clStreamKey, exists)
	clInfo := testutil.CloudletInfoData[0]
	clInfo.State = dme.CloudletState_CLOUDLET_STATE_READY
	cloudletInfoApi.Update(ctx, &clInfo, 0)

	clusterInst := testutil.ClusterInstData[0]
	clusterStreamKey := edgeproto.AppInstKey{ClusterInstKey: *clusterInst.Key.Virtual("")}
	err = clusterInstApi.CreateClusterInst(&clusterInst, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "create clusterinst")
	testStreamObjExists(t, ctx, &clusterStreamKey, exists)

	appInst := testutil.AppInstData[0]
	err = appInstApi.CreateAppInst(&appInst, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "create appinst")
	testStreamObjExists(t, ctx, &appInst.Key, exists)

	err = appInstApi.DeleteAppInst(&appInst, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "delete appinst")
	testStreamObjExists(t, ctx, &appInst.Key, !exists)

	err = clusterInstApi.DeleteClusterInst(&clusterInst, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "delete clusterinst")
	testStreamObjExists(t, ctx, &clusterStreamKey, !exists)

	err = cloudletApi.DeleteCloudlet(&cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err, "delete cloudlet")
	testStreamObjExists(t, ctx, &clStreamKey, !exists)
}
