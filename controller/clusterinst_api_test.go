package main

import (
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestClusterInstApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi | log.DebugLevelNotify)
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

	// cannot create insts without cluster/cloudlet
	for _, obj := range testutil.ClusterInstData {
		err := clusterInstApi.CreateClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
		require.NotNil(t, err, "Create ClusterInst without cloudlet")
	}

	// create support data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalOperatorCreate(t, &operatorApi, testutil.OperatorData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)
	insertCloudletInfo(testutil.CloudletInfoData)

	// Set responder to fail. This should clean up the object after
	// the fake crm returns a failure. If it doesn't, the next test to
	// create all the cluster insts will fail.
	responder.SetSimulateClusterCreateFailure(true)
	for _, obj := range testutil.ClusterInstData {
		err := clusterInstApi.CreateClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
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
	err := clusterInstApi.DeleteClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	require.NotNil(t, err, "Delete ClusterInst responder failure")
	responder.SetSimulateClusterDeleteFailure(false)
	checkClusterInstState(t, commonApi, &obj, edgeproto.TrackedState_READY)

	// check override of error DELETE_ERROR
	err = forceClusterInstState(&obj, edgeproto.TrackedState_DELETE_ERROR)
	require.Nil(t, err, "force state")
	checkClusterInstState(t, commonApi, &obj, edgeproto.TrackedState_DELETE_ERROR)
	err = clusterInstApi.CreateClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	require.Nil(t, err, "create overrides delete error")
	checkClusterInstState(t, commonApi, &obj, edgeproto.TrackedState_READY)

	// check override of error CREATE_ERROR
	err = forceClusterInstState(&obj, edgeproto.TrackedState_CREATE_ERROR)
	require.Nil(t, err, "force state")
	checkClusterInstState(t, commonApi, &obj, edgeproto.TrackedState_CREATE_ERROR)
	err = clusterInstApi.DeleteClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	require.Nil(t, err, "delete overrides create error")
	checkClusterInstState(t, commonApi, &obj, edgeproto.TrackedState_NOT_PRESENT)

	// override CRM error
	responder.SetSimulateClusterCreateFailure(true)
	responder.SetSimulateClusterDeleteFailure(true)
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = clusterInstApi.CreateClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	require.Nil(t, err, "override crm error")
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = clusterInstApi.DeleteClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	require.Nil(t, err, "override crm error")

	// ignore CRM
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
	err = clusterInstApi.CreateClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	require.Nil(t, err, "ignore crm")
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
	err = clusterInstApi.DeleteClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	require.Nil(t, err, "ignore crm")

	// inavailability of matching node flavor
	obj = testutil.ClusterInstData[0]
	obj.Flavor = testutil.FlavorData[0].Key
	err = clusterInstApi.CreateClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	require.NotNil(t, err, "flavor not available")

	responder.SetSimulateClusterCreateFailure(false)
	responder.SetSimulateClusterDeleteFailure(false)

	dummy.Stop()
}

func reduceInfoTimeouts() {
	CreateClusterInstTimeout = 1 * time.Second
	UpdateClusterInstTimeout = 1 * time.Second
	DeleteClusterInstTimeout = 1 * time.Second

	CreateAppInstTimeout = 1 * time.Second
	UpdateAppInstTimeout = 1 * time.Second
	DeleteAppInstTimeout = 1 * time.Second
}

func checkClusterInstState(t *testing.T, api *testutil.ClusterInstCommonApi, in *edgeproto.ClusterInst, state edgeproto.TrackedState) {
	out := edgeproto.ClusterInst{}
	found := testutil.GetClusterInst(t, api, &in.Key, &out)
	if state == edgeproto.TrackedState_NOT_PRESENT {
		require.False(t, found, "get cluster inst")
	} else {
		require.True(t, found, "get cluster inst")
		require.Equal(t, state, out.State, "cluster inst state")
	}
}

func forceClusterInstState(in *edgeproto.ClusterInst, state edgeproto.TrackedState) error {
	err := clusterInstApi.sync.ApplySTMWait(func(stm concurrency.STM) error {
		obj := edgeproto.ClusterInst{}
		if !clusterInstApi.store.STMGet(stm, &in.Key, &obj) {
			return objstore.ErrKVStoreKeyNotFound
		}
		obj.State = state
		clusterInstApi.store.STMPut(stm, &obj)
		return nil
	})
	return err
}
