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

func TestVMPoolApi(t *testing.T) {
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

	testutil.InternalVMPoolTest(t, "cud", &vmPoolApi, testutil.VMPoolData)

	testAddRemoveVM(t, ctx)
	testVMPoolAction(t, ctx)

	dummy.Stop()
}

func waitForAction(key *edgeproto.CloudletKey, action edgeproto.VMAction) error {
	var lastAction edgeproto.VMAction
	for i := 0; i < 100; i++ {
		pool := edgeproto.VMPool{}
		if vmPoolApi.cache.Get(key, &pool) {
			if pool.Action == action {
				return nil
			}
			lastAction = pool.Action
		}
		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("Unable to get desired Cloudlet VM Pool action, actual action %s, desired action %s", lastAction, action)
}

func testAddRemoveVM(t *testing.T, ctx context.Context) {
	// test adding cloudlet vm to the pool
	cm1 := edgeproto.VMPoolMember{}
	cm1.Key = testutil.VMPoolData[1].Key
	cm1.Vm = edgeproto.VM{
		Name: "vmX",
		NetInfo: edgeproto.VMNetInfo{
			ExternalIp: "192.168.1.111",
			InternalIp: "192.168.100.111",
		},
	}

	_, err := vmPoolApi.AddVMPoolMember(ctx, &cm1)
	require.Nil(t, err, "add cloudlet vm to cloudlet vm pool")

	vmPool := edgeproto.VMPool{}
	found := vmPoolApi.cache.Get(&cm1.Key, &vmPool)
	require.True(t, found, "get cloudlet vm pool %v", cm1.Key)
	require.Equal(t, 4, len(vmPool.Vms))

	// test adding another cloudlet vm to the pool
	cm2 := edgeproto.VMPoolMember{}
	cm2.Key = testutil.VMPoolData[1].Key
	cm2.Vm = edgeproto.VM{
		Name: "vmY",
		NetInfo: edgeproto.VMNetInfo{
			ExternalIp: "192.168.1.121",
			InternalIp: "192.168.100.121",
		},
	}

	_, err = vmPoolApi.AddVMPoolMember(ctx, &cm2)
	require.Nil(t, err, "add cloudlet vm to cloudlet vm pool")

	found = vmPoolApi.cache.Get(&cm2.Key, &vmPool)
	require.True(t, found, "get cloudlet vm pool %v", cm2.Key)
	require.Equal(t, 5, len(vmPool.Vms))

	// remove cloudlet vm from pool
	_, err = vmPoolApi.RemoveVMPoolMember(ctx, &cm1)
	require.Nil(t, err, "remove cloudlet vm from pool")
	found = vmPoolApi.cache.Get(&cm1.Key, &vmPool)
	require.True(t, found, "get cloudlet vm pool %v", cm1.Key)
	require.Equal(t, 4, len(vmPool.Vms))

	// remove cloudlet vm from pool
	_, err = vmPoolApi.RemoveVMPoolMember(ctx, &cm2)
	require.Nil(t, err, "remove cloudlet vm from pool")
	found = vmPoolApi.cache.Get(&cm2.Key, &vmPool)
	require.True(t, found, "get cloudlet vm pool %v", cm2.Key)
	require.Equal(t, 3, len(vmPool.Vms))

	// try to add cloudlet vm to non-existent pool
	cm1.Key.Name = "SomeNonExistentCloudlet"
	_, err = vmPoolApi.AddVMPoolMember(ctx, &cm1)
	require.NotNil(t, err)
}

var Pass = true
var Fail = false

func verifyVMAction(t *testing.T, ctx context.Context, info *edgeproto.VMPoolInfo, inUseCount int, success bool) {
	key := &info.Key
	vmPoolInfoApi.Update(ctx, info, 0)
	err := waitForAction(key, info.Action)
	require.Nil(t, err, "cloudlet vm pool action transtions")

	info.Action = edgeproto.VMAction_VM_ACTION_DONE
	vmPoolInfoApi.Update(ctx, info, 0)
	err = waitForAction(key, edgeproto.VMAction_VM_ACTION_DONE)
	require.Nil(t, err, "cloudlet vm pool action transtions")

	vmPool := edgeproto.VMPool{}
	found := vmPoolApi.cache.Get(key, &vmPool)
	require.True(t, found, "get cloudlet vm pool %v", vmPool.Key)
	if success {
		require.Empty(t, vmPool.Error)
	} else {
		require.NotEmpty(t, vmPool.Error, "cloudlet vm pool allocation failed")
	}

	inuseVms := []edgeproto.VM{}
	for _, vm := range vmPool.Vms {
		if vm.GroupName == info.GroupName && vm.State == edgeproto.VMState_VM_IN_USE {
			inuseVms = append(inuseVms, vm)
		}
	}
	require.Equal(t, len(inuseVms), inUseCount)
}

func testVMPoolAction(t *testing.T, ctx context.Context) {
	vmPool := testutil.VMPoolData[1]
	info := edgeproto.VMPoolInfo{}

	group1 := "testvmpoolVMs1"
	group2 := "testvmpoolVMs2"

	// Allocate VMs
	info.GroupName = group1
	info.Key = vmPool.Key
	info.VmSpecs = []edgeproto.VMSpec{
		edgeproto.VMSpec{
			InternalName:    "vm1.testcluster.testorg.mobiledgex.net",
			ExternalNetwork: true,
		},
		edgeproto.VMSpec{
			InternalName:    "vm2.testcluster.testorg.mobiledgex.net",
			InternalNetwork: true,
		},
	}
	info.Action = edgeproto.VMAction_VM_ACTION_ALLOCATE
	verifyVMAction(t, ctx, &info, 2, Pass)

	// Allocate some more VMs but by different group
	// Below also tests that previous VM allocation didn't use VM with external
	// connectivity for a spec with just requesting internal network access
	info.GroupName = group2
	info.VmSpecs = []edgeproto.VMSpec{
		edgeproto.VMSpec{
			InternalName:    "vm2.testcluster.testorg.mobiledgex.net",
			ExternalNetwork: true,
		},
	}
	info.Action = edgeproto.VMAction_VM_ACTION_ALLOCATE
	verifyVMAction(t, ctx, &info, 1, Pass)

	// Allocate some more VMs, but it should fail as there aren't enough free VMs
	info.GroupName = group1
	info.Action = edgeproto.VMAction_VM_ACTION_ALLOCATE
	verifyVMAction(t, ctx, &info, 2, Fail)

	// Release 1 VM from the pool
	info.GroupName = group1
	info.Action = edgeproto.VMAction_VM_ACTION_RELEASE
	info.VmSpecs = []edgeproto.VMSpec{
		edgeproto.VMSpec{
			InternalName: "vm2.testcluster.testorg.mobiledgex.net",
		},
	}
	verifyVMAction(t, ctx, &info, 1, Pass)

	// A VM is free, but allocation should fail as no VM with external network
	// connectivity is free
	info.GroupName = group1
	info.VmSpecs = []edgeproto.VMSpec{
		edgeproto.VMSpec{
			InternalName:    "vm2.testcluster.testorg.mobiledgex.net",
			ExternalNetwork: true,
		},
	}
	info.Action = edgeproto.VMAction_VM_ACTION_ALLOCATE
	verifyVMAction(t, ctx, &info, 1, Fail)

	// Release 2nd VM from the pool
	info.GroupName = group1
	info.Action = edgeproto.VMAction_VM_ACTION_RELEASE
	info.VmSpecs = []edgeproto.VMSpec{
		edgeproto.VMSpec{
			InternalName: "vm1.testcluster.testorg.mobiledgex.net",
		},
	}
	verifyVMAction(t, ctx, &info, 0, Pass)

	// Add VMs to the pool
	cm1 := edgeproto.VMPoolMember{}
	cm1.Key = info.Key
	cm1.Vm = edgeproto.VM{
		Name: "vmX",
		NetInfo: edgeproto.VMNetInfo{
			InternalIp: "192.168.100.111",
		},
	}
	_, err := vmPoolApi.AddVMPoolMember(ctx, &cm1)
	require.Nil(t, err, "add cloudlet vm to cloudlet vm pool")

	// Allocate VMs, it should succeed now
	info.GroupName = group2
	info.Action = edgeproto.VMAction_VM_ACTION_ALLOCATE
	info.VmSpecs = []edgeproto.VMSpec{
		edgeproto.VMSpec{
			InternalName:    "vm1.testcluster.testorg.mobiledgex.net",
			InternalNetwork: true,
		},
		edgeproto.VMSpec{
			InternalName:    "vm2.testcluster.testorg.mobiledgex.net",
			InternalNetwork: true,
		},
		// Below also tests that previous VM allocation didn't use VM with external
		// connectivity for a spec with just requesting internal network access
		edgeproto.VMSpec{
			InternalName:    "vm3.testcluster.testorg.mobiledgex.net",
			ExternalNetwork: true,
		},
	}
	verifyVMAction(t, ctx, &info, 4, Pass)

	// Release 1 VM from the pool
	info.GroupName = group2
	info.Action = edgeproto.VMAction_VM_ACTION_RELEASE
	info.VmSpecs = []edgeproto.VMSpec{
		edgeproto.VMSpec{
			InternalName: "vm2.testcluster.testorg.mobiledgex.net",
		},
	}
	verifyVMAction(t, ctx, &info, 2, Pass)

	// Release all VMs
	info.GroupName = group2
	info.Action = edgeproto.VMAction_VM_ACTION_RELEASE
	info.VmSpecs = []edgeproto.VMSpec{}
	verifyVMAction(t, ctx, &info, 0, Pass)

	// Release all VMs
	info.GroupName = group1
	info.Action = edgeproto.VMAction_VM_ACTION_RELEASE
	info.VmSpecs = []edgeproto.VMSpec{}
	verifyVMAction(t, ctx, &info, 0, Pass)
}
