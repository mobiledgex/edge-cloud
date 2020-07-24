package main

import (
	"context"
	"testing"

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

	dummy.Stop()
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
