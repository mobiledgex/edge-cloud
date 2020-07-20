package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestCloudletVMPoolApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi | log.DebugLevelNotify)
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

	testutil.InternalCloudletVMPoolTest(t, "cud", &cloudletVMPoolApi, testutil.CloudletVMPoolData)

	testAddRemoveCloudletVM(t, ctx)
	testCloudletVMPoolAction(t, ctx)

	dummy.Stop()
}

func waitForAction(key *edgeproto.CloudletKey, action edgeproto.CloudletVMAction) error {
	var lastAction edgeproto.CloudletVMAction
	for i := 0; i < 100; i++ {
		pool := edgeproto.CloudletVMPool{}
		if cloudletVMPoolApi.cache.Get(key, &pool) {
			if pool.Action == action {
				return nil
			}
			lastAction = pool.Action
		}
		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("Unable to get desired Cloudlet VM Pool action, actual action %s, desired action %s", lastAction, action)
}

func testAddRemoveCloudletVM(t *testing.T, ctx context.Context) {
	// test adding cloudlet vm to the pool
	cm1 := edgeproto.CloudletVMPoolMember{}
	cm1.Key = testutil.CloudletVMPoolData[1].Key
	cm1.CloudletVm = edgeproto.CloudletVM{
		Name: "vmX",
		NetInfo: edgeproto.CloudletVMNetInfo{
			ExternalIp: "192.168.1.111",
			InternalIp: "192.168.100.111",
		},
	}

	_, err := cloudletVMPoolApi.AddCloudletVMPoolMember(ctx, &cm1)
	require.Nil(t, err, "add cloudlet vm to cloudlet vm pool")

	vmPool := edgeproto.CloudletVMPool{}
	found := cloudletVMPoolApi.cache.Get(&cm1.Key, &vmPool)
	require.True(t, found, "get cloudlet vm pool %v", cm1.Key)
	require.Equal(t, 4, len(vmPool.CloudletVms))

	// test adding another cloudlet vm to the pool
	cm2 := edgeproto.CloudletVMPoolMember{}
	cm2.Key = testutil.CloudletVMPoolData[1].Key
	cm2.CloudletVm = edgeproto.CloudletVM{
		Name: "vmY",
		NetInfo: edgeproto.CloudletVMNetInfo{
			ExternalIp: "192.168.1.121",
			InternalIp: "192.168.100.121",
		},
	}

	_, err = cloudletVMPoolApi.AddCloudletVMPoolMember(ctx, &cm2)
	require.Nil(t, err, "add cloudlet vm to cloudlet vm pool")

	found = cloudletVMPoolApi.cache.Get(&cm2.Key, &vmPool)
	require.True(t, found, "get cloudlet vm pool %v", cm2.Key)
	require.Equal(t, 5, len(vmPool.CloudletVms))

	// remove cloudlet vm from pool
	_, err = cloudletVMPoolApi.RemoveCloudletVMPoolMember(ctx, &cm1)
	require.Nil(t, err, "remove cloudlet vm from pool")
	found = cloudletVMPoolApi.cache.Get(&cm1.Key, &vmPool)
	require.True(t, found, "get cloudlet vm pool %v", cm1.Key)
	require.Equal(t, 4, len(vmPool.CloudletVms))

	// remove cloudlet vm from pool
	_, err = cloudletVMPoolApi.RemoveCloudletVMPoolMember(ctx, &cm2)
	require.Nil(t, err, "remove cloudlet vm from pool")
	found = cloudletVMPoolApi.cache.Get(&cm2.Key, &vmPool)
	require.True(t, found, "get cloudlet vm pool %v", cm2.Key)
	require.Equal(t, 3, len(vmPool.CloudletVms))

	// try to add cloudlet vm to non-existent pool
	cm1.Key.Name = "SomeNonExistentCloudlet"
	_, err = cloudletVMPoolApi.AddCloudletVMPoolMember(ctx, &cm1)
	require.NotNil(t, err)
}

var Pass = true
var Fail = false

func verifyCloudletVMAction(t *testing.T, ctx context.Context, info *edgeproto.CloudletVMPoolInfo, inUseCount int, success bool) {
	key := &info.Key
	cloudletVMPoolInfoApi.Update(ctx, info, 0)
	err := waitForAction(key, info.Action)
	require.Nil(t, err, "cloudlet vm pool action transtions")

	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE
	cloudletVMPoolInfoApi.Update(ctx, info, 0)
	err = waitForAction(key, edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE)
	require.Nil(t, err, "cloudlet vm pool action transtions")

	vmPool := edgeproto.CloudletVMPool{}
	found := cloudletVMPoolApi.cache.Get(key, &vmPool)
	require.True(t, found, "get cloudlet vm pool %v", vmPool.Key)
	if success {
		require.Empty(t, vmPool.Error)
	} else {
		require.NotEmpty(t, vmPool.Error, "cloudlet vm pool allocation failed")
	}

	inuseVms := []edgeproto.CloudletVM{}
	for _, cloudletVm := range vmPool.CloudletVms {
		if cloudletVm.State == edgeproto.CloudletVMState_CLOUDLET_VM_IN_USE {
			inuseVms = append(inuseVms, cloudletVm)
		}
	}
	require.Equal(t, len(inuseVms), inUseCount)
}

func testCloudletVMPoolAction(t *testing.T, ctx context.Context) {
	vmPool := testutil.CloudletVMPoolData[1]
	info := edgeproto.CloudletVMPoolInfo{}

	// Allocate VMs
	info.User = "testcloudletvmpoolVMs1"
	info.Key = vmPool.Key
	info.VmSpecs = []edgeproto.CloudletVMSpec{
		edgeproto.CloudletVMSpec{
			InternalName:    "vm1.testcluster.testorg.mobiledgex.net",
			ExternalNetwork: true,
		},
		edgeproto.CloudletVMSpec{
			InternalName:    "vm2.testcluster.testorg.mobiledgex.net",
			InternalNetwork: true,
		},
	}
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_ALLOCATE
	verifyCloudletVMAction(t, ctx, &info, 2, Pass)

	// Allocate some more VMs but by different user
	info.User = "testcloudletvmpoolVMs2"
	info.VmSpecs = []edgeproto.CloudletVMSpec{
		edgeproto.CloudletVMSpec{
			InternalName:    "vm2.testcluster.testorg.mobiledgex.net",
			InternalNetwork: true,
		},
	}
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_ALLOCATE
	verifyCloudletVMAction(t, ctx, &info, 3, Pass)

	// Allocate some more VMs, but it should fail as there aren't enough free VMs
	info.User = "testcloudletvmpoolVMs1"
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_ALLOCATE
	verifyCloudletVMAction(t, ctx, &info, 3, Fail)

	// Release VMs
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_RELEASE
	info.VmSpecs = []edgeproto.CloudletVMSpec{}
	info.User = "testcloudletvmpoolVMs1"
	verifyCloudletVMAction(t, ctx, &info, 1, Pass)

	// Add VMs to the pool
	cm1 := edgeproto.CloudletVMPoolMember{}
	cm1.Key = info.Key
	cm1.CloudletVm = edgeproto.CloudletVM{
		Name: "vmX",
		NetInfo: edgeproto.CloudletVMNetInfo{
			ExternalIp: "192.168.1.111",
			InternalIp: "192.168.100.111",
		},
	}
	_, err := cloudletVMPoolApi.AddCloudletVMPoolMember(ctx, &cm1)
	require.Nil(t, err, "add cloudlet vm to cloudlet vm pool")

	// Allocate VMs, it should succeed now
	info.User = "testcloudletvmpoolVMs2"
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_ALLOCATE
	info.VmSpecs = []edgeproto.CloudletVMSpec{
		edgeproto.CloudletVMSpec{
			InternalName:    "vm1.testcluster.testorg.mobiledgex.net",
			ExternalNetwork: true,
		},
		edgeproto.CloudletVMSpec{
			InternalName:    "vm2.testcluster.testorg.mobiledgex.net",
			InternalNetwork: true,
		},
		edgeproto.CloudletVMSpec{
			InternalName:    "vm3.testcluster.testorg.mobiledgex.net",
			InternalNetwork: true,
		},
	}
	verifyCloudletVMAction(t, ctx, &info, 4, Pass)

	// Release VMs
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_RELEASE
	info.VmSpecs = []edgeproto.CloudletVMSpec{}
	info.User = "testcloudletvmpoolVMs2"
	verifyCloudletVMAction(t, ctx, &info, 0, Pass)
}
