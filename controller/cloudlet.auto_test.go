// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

package main

import (
	"context"
	fmt "fmt"
	"github.com/coreos/etcd/clientv3/concurrency"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/objstore"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
	math "math"
	"testing"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

// Caller must write by hand the test data generator.
// Each Ref object should only have a single reference to the key,
// in order to properly test each reference (i.e. don't have a single
// object that has multiple references).
type GPUDriverDeleteDataGen interface {
	GetGPUDriverTestObj() (*edgeproto.GPUDriver, *testSupportData)
	GetCloudletGpuConfigDriverRef(key *edgeproto.GPUDriverKey) (*edgeproto.Cloudlet, *testSupportData)
}

// GPUDriverDeleteStore wraps around the usual
// store to instrument checks and inject data while
// the delete api code is running.
type GPUDriverDeleteStore struct {
	edgeproto.GPUDriverStore
	t                   *testing.T
	allApis             *AllApis
	putDeletePrepare    bool
	putDeletePrepareCb  func()
	putDeletePrepareSTM concurrency.STM
}

func (s *GPUDriverDeleteStore) Put(ctx context.Context, m *edgeproto.GPUDriver, wait func(int64), ops ...objstore.KVOp) (*edgeproto.Result, error) {
	if wait != nil {
		s.putDeletePrepare = m.DeletePrepare
	}
	res, err := s.GPUDriverStore.Put(ctx, m, wait, ops...)
	if s.putDeletePrepare && s.putDeletePrepareCb != nil {
		s.putDeletePrepareCb()
	}
	return res, err
}

func (s *GPUDriverDeleteStore) STMPut(stm concurrency.STM, obj *edgeproto.GPUDriver, ops ...objstore.KVOp) {
	// there's an assumption that this is run within an ApplySTMWait,
	// where we wait for the caches to be updated with the transaction.
	if obj.DeletePrepare {
		s.putDeletePrepare = true
		s.putDeletePrepareSTM = stm
	} else {
		s.putDeletePrepare = false
		s.putDeletePrepareSTM = nil
	}
	s.GPUDriverStore.STMPut(stm, obj, ops...)
	if s.putDeletePrepare && s.putDeletePrepareCb != nil {
		s.putDeletePrepareCb()
	}
}

func (s *GPUDriverDeleteStore) Delete(ctx context.Context, m *edgeproto.GPUDriver, wait func(int64)) (*edgeproto.Result, error) {
	require.True(s.t, s.putDeletePrepare, "DeletePrepare must be comitted to database with a sync.Wait before deleting")
	return s.GPUDriverStore.Delete(ctx, m, wait)
}

func (s *GPUDriverDeleteStore) STMDel(stm concurrency.STM, key *edgeproto.GPUDriverKey) {
	require.True(s.t, s.putDeletePrepare, "DeletePrepare must be comitted to database with a sync.Wait before deleting")
	s.GPUDriverStore.STMDel(stm, key)
}

func (s *GPUDriverDeleteStore) requireUndoDeletePrepare(ctx context.Context, obj *edgeproto.GPUDriver) {
	buf := edgeproto.GPUDriver{}
	found := s.Get(ctx, obj.GetKey(), &buf)
	require.True(s.t, found, "expected test object to be found")
	require.False(s.t, buf.DeletePrepare, "undo delete prepare field")
}

func deleteGPUDriverChecks(t *testing.T, ctx context.Context, all *AllApis, dataGen GPUDriverDeleteDataGen) {
	var err error
	// override store so we can inject data and check data
	api := all.gpuDriverApi
	origStore := api.store
	deleteStore := &GPUDriverDeleteStore{
		GPUDriverStore: origStore,
		t:              t,
		allApis:        all,
	}
	api.store = deleteStore
	defer func() {
		api.store = origStore
	}()

	// inject testObj directly, bypassing create checks/deps
	testObj, supportData := dataGen.GetGPUDriverTestObj()
	supportData.put(t, ctx, all)
	defer supportData.delete(t, ctx, all)
	origStore.Put(ctx, testObj, api.sync.syncWait)

	// Positive test, delete should succeed without any references.
	// The overrided store checks that delete prepare was set on the
	// object in the database before actually doing the delete.
	testObj, _ = dataGen.GetGPUDriverTestObj()
	err = api.DeleteGPUDriver(testObj, testutil.NewCudStreamoutGPUDriver(ctx))
	require.Nil(t, err, "delete must succeed with no refs")

	// Negative test, inject testObj with prepare delete already set.
	testObj, _ = dataGen.GetGPUDriverTestObj()
	testObj.DeletePrepare = true
	origStore.Put(ctx, testObj, api.sync.syncWait)
	// delete should fail with already being deleted
	testObj, _ = dataGen.GetGPUDriverTestObj()
	err = api.DeleteGPUDriver(testObj, testutil.NewCudStreamoutGPUDriver(ctx))
	require.NotNil(t, err, "delete must fail if already being deleted")
	require.Contains(t, err.Error(), "already being deleted")

	// inject testObj for ref tests
	testObj, _ = dataGen.GetGPUDriverTestObj()
	origStore.Put(ctx, testObj, api.sync.syncWait)

	{
		// Negative test, Cloudlet refers to GPUDriver.
		// The cb will inject refBy obj after delete prepare has been set.
		refBy, supportData := dataGen.GetCloudletGpuConfigDriverRef(testObj.GetKey())
		supportData.put(t, ctx, all)
		deleteStore.putDeletePrepareCb = func() {
			all.cloudletApi.store.Put(ctx, refBy, all.cloudletApi.sync.syncWait)
		}
		testObj, _ = dataGen.GetGPUDriverTestObj()
		err = api.DeleteGPUDriver(testObj, testutil.NewCudStreamoutGPUDriver(ctx))
		require.NotNil(t, err, "must fail delete with ref from Cloudlet")
		require.Contains(t, err.Error(), "in use")
		// check that delete prepare was reset
		deleteStore.requireUndoDeletePrepare(ctx, testObj)
		// remove Cloudlet obj
		_, err = all.cloudletApi.store.Delete(ctx, refBy, all.cloudletApi.sync.syncWait)
		require.Nil(t, err, "cleanup ref from Cloudlet must succeed")
		deleteStore.putDeletePrepareCb = nil
		supportData.delete(t, ctx, all)
	}

	// clean up testObj
	testObj, _ = dataGen.GetGPUDriverTestObj()
	err = api.DeleteGPUDriver(testObj, testutil.NewCudStreamoutGPUDriver(ctx))
	require.Nil(t, err, "cleanup must succeed")
}

// Caller must write by hand the test data generator.
// Each Ref object should only have a single reference to the key,
// in order to properly test each reference (i.e. don't have a single
// object that has multiple references).
type CloudletDeleteDataGen interface {
	GetCloudletTestObj() (*edgeproto.Cloudlet, *testSupportData)
	GetAutoProvPolicyCloudletsRef(key *edgeproto.CloudletKey) (*edgeproto.AutoProvPolicy, *testSupportData)
	GetCloudletPoolCloudletsRef(key *edgeproto.CloudletKey) (*edgeproto.CloudletPool, *testSupportData)
	GetNetworkKeyCloudletKeyRef(key *edgeproto.CloudletKey) (*edgeproto.Network, *testSupportData)
	GetCloudletClusterInstClusterInstsRef(key *edgeproto.CloudletKey) (*edgeproto.CloudletRefs, *testSupportData)
}

// CloudletDeleteStore wraps around the usual
// store to instrument checks and inject data while
// the delete api code is running.
type CloudletDeleteStore struct {
	edgeproto.CloudletStore
	t                   *testing.T
	allApis             *AllApis
	putDeletePrepare    bool
	putDeletePrepareCb  func()
	putDeletePrepareSTM concurrency.STM
}

func (s *CloudletDeleteStore) Put(ctx context.Context, m *edgeproto.Cloudlet, wait func(int64), ops ...objstore.KVOp) (*edgeproto.Result, error) {
	if wait != nil {
		s.putDeletePrepare = m.DeletePrepare
	}
	res, err := s.CloudletStore.Put(ctx, m, wait, ops...)
	if s.putDeletePrepare && s.putDeletePrepareCb != nil {
		s.putDeletePrepareCb()
	}
	return res, err
}

func (s *CloudletDeleteStore) STMPut(stm concurrency.STM, obj *edgeproto.Cloudlet, ops ...objstore.KVOp) {
	// there's an assumption that this is run within an ApplySTMWait,
	// where we wait for the caches to be updated with the transaction.
	if obj.DeletePrepare {
		s.putDeletePrepare = true
		s.putDeletePrepareSTM = stm
	} else {
		s.putDeletePrepare = false
		s.putDeletePrepareSTM = nil
	}
	s.CloudletStore.STMPut(stm, obj, ops...)
	if s.putDeletePrepare && s.putDeletePrepareCb != nil {
		s.putDeletePrepareCb()
	}
}

func (s *CloudletDeleteStore) Delete(ctx context.Context, m *edgeproto.Cloudlet, wait func(int64)) (*edgeproto.Result, error) {
	require.True(s.t, s.putDeletePrepare, "DeletePrepare must be comitted to database with a sync.Wait before deleting")
	return s.CloudletStore.Delete(ctx, m, wait)
}

func (s *CloudletDeleteStore) STMDel(stm concurrency.STM, key *edgeproto.CloudletKey) {
	require.True(s.t, s.putDeletePrepare, "DeletePrepare must be comitted to database with a sync.Wait before deleting")
	s.CloudletStore.STMDel(stm, key)
}

func (s *CloudletDeleteStore) requireUndoDeletePrepare(ctx context.Context, obj *edgeproto.Cloudlet) {
	buf := edgeproto.Cloudlet{}
	found := s.Get(ctx, obj.GetKey(), &buf)
	require.True(s.t, found, "expected test object to be found")
	require.False(s.t, buf.DeletePrepare, "undo delete prepare field")
}

func deleteCloudletChecks(t *testing.T, ctx context.Context, all *AllApis, dataGen CloudletDeleteDataGen) {
	var err error
	// override store so we can inject data and check data
	api := all.cloudletApi
	origStore := api.store
	deleteStore := &CloudletDeleteStore{
		CloudletStore: origStore,
		t:             t,
		allApis:       all,
	}
	api.store = deleteStore
	origcloudletRefsApiStore := all.cloudletRefsApi.store
	cloudletRefsApiStore := &CloudletRefsDeleteStore{
		CloudletRefsStore: origcloudletRefsApiStore,
	}
	all.cloudletRefsApi.store = cloudletRefsApiStore
	defer func() {
		api.store = origStore
		all.cloudletRefsApi.store = origcloudletRefsApiStore
	}()

	// inject testObj directly, bypassing create checks/deps
	testObj, supportData := dataGen.GetCloudletTestObj()
	supportData.put(t, ctx, all)
	defer supportData.delete(t, ctx, all)
	origStore.Put(ctx, testObj, api.sync.syncWait)

	// Positive test, delete should succeed without any references.
	// The overrided store checks that delete prepare was set on the
	// object in the database before actually doing the delete.
	// This call back checks that any refs lookups are done in the
	// same stm as the delete prepare is set.
	deleteStore.putDeletePrepareCb = func() {
		// make sure ref objects reads happen in same stm
		// as delete prepare is set
		require.NotNil(t, deleteStore.putDeletePrepareSTM, "must set delete prepare in STM")
		require.NotNil(t, cloudletRefsApiStore.getRefSTM, "must check for refs from CloudletRefs in STM")
		require.Equal(t, deleteStore.putDeletePrepareSTM, cloudletRefsApiStore.getRefSTM, "delete prepare and ref check for CloudletRefs must be done in the same STM")
	}
	testObj, _ = dataGen.GetCloudletTestObj()
	err = api.DeleteCloudlet(testObj, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err, "delete must succeed with no refs")
	deleteStore.putDeletePrepareCb = nil

	// Negative test, inject testObj with prepare delete already set.
	testObj, _ = dataGen.GetCloudletTestObj()
	testObj.DeletePrepare = true
	origStore.Put(ctx, testObj, api.sync.syncWait)
	// delete should fail with already being deleted
	testObj, _ = dataGen.GetCloudletTestObj()
	err = api.DeleteCloudlet(testObj, testutil.NewCudStreamoutCloudlet(ctx))
	require.NotNil(t, err, "delete must fail if already being deleted")
	require.Contains(t, err.Error(), "already being deleted")

	// inject testObj for ref tests
	testObj, _ = dataGen.GetCloudletTestObj()
	origStore.Put(ctx, testObj, api.sync.syncWait)

	{
		// Negative test, AutoProvPolicy refers to Cloudlet.
		// The cb will inject refBy obj after delete prepare has been set.
		refBy, supportData := dataGen.GetAutoProvPolicyCloudletsRef(testObj.GetKey())
		supportData.put(t, ctx, all)
		deleteStore.putDeletePrepareCb = func() {
			all.autoProvPolicyApi.store.Put(ctx, refBy, all.autoProvPolicyApi.sync.syncWait)
		}
		testObj, _ = dataGen.GetCloudletTestObj()
		err = api.DeleteCloudlet(testObj, testutil.NewCudStreamoutCloudlet(ctx))
		require.NotNil(t, err, "must fail delete with ref from AutoProvPolicy")
		require.Contains(t, err.Error(), "in use")
		// check that delete prepare was reset
		deleteStore.requireUndoDeletePrepare(ctx, testObj)
		// remove AutoProvPolicy obj
		_, err = all.autoProvPolicyApi.store.Delete(ctx, refBy, all.autoProvPolicyApi.sync.syncWait)
		require.Nil(t, err, "cleanup ref from AutoProvPolicy must succeed")
		deleteStore.putDeletePrepareCb = nil
		supportData.delete(t, ctx, all)
	}
	{
		// Negative test, CloudletPool refers to Cloudlet.
		// The cb will inject refBy obj after delete prepare has been set.
		refBy, supportData := dataGen.GetCloudletPoolCloudletsRef(testObj.GetKey())
		supportData.put(t, ctx, all)
		deleteStore.putDeletePrepareCb = func() {
			all.cloudletPoolApi.store.Put(ctx, refBy, all.cloudletPoolApi.sync.syncWait)
		}
		testObj, _ = dataGen.GetCloudletTestObj()
		err = api.DeleteCloudlet(testObj, testutil.NewCudStreamoutCloudlet(ctx))
		require.NotNil(t, err, "must fail delete with ref from CloudletPool")
		require.Contains(t, err.Error(), "in use")
		// check that delete prepare was reset
		deleteStore.requireUndoDeletePrepare(ctx, testObj)
		// remove CloudletPool obj
		_, err = all.cloudletPoolApi.store.Delete(ctx, refBy, all.cloudletPoolApi.sync.syncWait)
		require.Nil(t, err, "cleanup ref from CloudletPool must succeed")
		deleteStore.putDeletePrepareCb = nil
		supportData.delete(t, ctx, all)
	}
	{
		// Negative test, Network refers to Cloudlet.
		// The cb will inject refBy obj after delete prepare has been set.
		refBy, supportData := dataGen.GetNetworkKeyCloudletKeyRef(testObj.GetKey())
		supportData.put(t, ctx, all)
		deleteStore.putDeletePrepareCb = func() {
			all.networkApi.store.Put(ctx, refBy, all.networkApi.sync.syncWait)
		}
		testObj, _ = dataGen.GetCloudletTestObj()
		err = api.DeleteCloudlet(testObj, testutil.NewCudStreamoutCloudlet(ctx))
		require.NotNil(t, err, "must fail delete with ref from Network")
		require.Contains(t, err.Error(), "in use")
		// check that delete prepare was reset
		deleteStore.requireUndoDeletePrepare(ctx, testObj)
		// remove Network obj
		_, err = all.networkApi.store.Delete(ctx, refBy, all.networkApi.sync.syncWait)
		require.Nil(t, err, "cleanup ref from Network must succeed")
		deleteStore.putDeletePrepareCb = nil
		supportData.delete(t, ctx, all)
	}
	{
		// Negative test, CloudletRefs refers to Cloudlet via refs object.
		// Inject the refs object to trigger an "in use" error.
		refBy, supportData := dataGen.GetCloudletClusterInstClusterInstsRef(testObj.GetKey())
		supportData.put(t, ctx, all)
		_, err = all.cloudletRefsApi.store.Put(ctx, refBy, all.cloudletRefsApi.sync.syncWait)
		require.Nil(t, err)
		testObj, _ = dataGen.GetCloudletTestObj()
		err = api.DeleteCloudlet(testObj, testutil.NewCudStreamoutCloudlet(ctx))
		require.NotNil(t, err, "delete with ref from CloudletRefs must fail")
		require.Contains(t, err.Error(), "in use")
		// check that delete prepare was reset
		deleteStore.requireUndoDeletePrepare(ctx, testObj)
		// remove CloudletRefs obj
		_, err = all.cloudletRefsApi.store.Delete(ctx, refBy, all.cloudletRefsApi.sync.syncWait)
		require.Nil(t, err, "cleanup ref from CloudletRefs must succeed")
		supportData.delete(t, ctx, all)
	}

	// clean up testObj
	testObj, _ = dataGen.GetCloudletTestObj()
	err = api.DeleteCloudlet(testObj, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err, "cleanup must succeed")
}
