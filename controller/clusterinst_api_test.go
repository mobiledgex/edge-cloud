package main

import (
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
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
		assert.NotNil(t, err, "Create cluster inst without cloudlet")
	}

	// create support data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalOperatorCreate(t, &operatorApi, testutil.OperatorData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)
	insertCloudletInfo(testutil.CloudletInfoData)
	testutil.InternalClusterCreate(t, &clusterApi, testutil.ClusterData)

	// Set responder to fail. This should clean up the object after
	// the fake crm returns a failure. If it doesn't, the next test to
	// create all the cluster insts will fail.
	responder.SetSimulateClusterCreateFailure(true)
	for _, obj := range testutil.ClusterInstData {
		err := clusterInstApi.CreateClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
		assert.NotNil(t, err, "Create cluster inst responder failures")
		// make sure error matches responder
		assert.Equal(t, "Encountered failures: crm create cluster inst failed", err.Error())
	}
	responder.SetSimulateClusterCreateFailure(false)
	assert.Equal(t, 0, len(clusterInstApi.cache.Objs))

	testutil.InternalClusterInstTest(t, "cud", &clusterInstApi, testutil.ClusterInstData)
	// after cluster insts create, check that cloudlet refs data is correct.
	testutil.InternalCloudletRefsTest(t, "show", &cloudletRefsApi, testutil.CloudletRefsData)

	commonApi := testutil.NewInternalClusterInstApi(&clusterInstApi)

	// Set responder to fail delete.
	responder.SetSimulateClusterDeleteFailure(true)
	obj := testutil.ClusterInstData[0]
	err := clusterInstApi.DeleteClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	assert.NotNil(t, err, "Delete ClusterInst responder failure")
	responder.SetSimulateClusterDeleteFailure(false)
	checkClusterInstState(t, commonApi, &obj, edgeproto.TrackedState_Ready)

	// check override of error DeleteError
	err = forceClusterInstState(&obj, edgeproto.TrackedState_DeleteError)
	assert.Nil(t, err, "force state")
	checkClusterInstState(t, commonApi, &obj, edgeproto.TrackedState_DeleteError)
	err = clusterInstApi.CreateClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	assert.Nil(t, err, "create overrides delete error")
	checkClusterInstState(t, commonApi, &obj, edgeproto.TrackedState_Ready)

	// check override of error CreateError
	err = forceClusterInstState(&obj, edgeproto.TrackedState_CreateError)
	assert.Nil(t, err, "force state")
	checkClusterInstState(t, commonApi, &obj, edgeproto.TrackedState_CreateError)
	err = clusterInstApi.DeleteClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	assert.Nil(t, err, "delete overrides create error")
	checkClusterInstState(t, commonApi, &obj, edgeproto.TrackedState_NotPresent)

	// override CRM error
	responder.SetSimulateClusterCreateFailure(true)
	responder.SetSimulateClusterDeleteFailure(true)
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IgnoreCRMErrors
	err = clusterInstApi.CreateClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	assert.Nil(t, err, "override crm error")
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IgnoreCRMErrors
	err = clusterInstApi.DeleteClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	assert.Nil(t, err, "override crm error")

	// ignore CRM
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IgnoreCRM
	err = clusterInstApi.CreateClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	assert.Nil(t, err, "ignore crm")
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IgnoreCRM
	err = clusterInstApi.DeleteClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	assert.Nil(t, err, "ignore crm")

	// inavailability of matching node flavor
	obj = testutil.ClusterInstData[0]
	obj.Flavor = testutil.FlavorData[0].Key
	err = clusterInstApi.CreateClusterInst(&obj, &testutil.CudStreamoutClusterInst{})
	assert.NotNil(t, err, "flavor not available")

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
	if state == edgeproto.TrackedState_NotPresent {
		assert.False(t, found, "get cluster inst")
	} else {
		assert.True(t, found, "get cluster inst")
		assert.Equal(t, state, out.State, "cluster inst state")
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
