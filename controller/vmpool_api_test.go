// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testSvcs := testinit(ctx, t)
	defer testfinish(testSvcs)

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	apis := NewAllApis(sync)
	sync.Start()
	defer sync.Done()

	testutil.InternalVMPoolTest(t, "cud", apis.vmPoolApi, testutil.VMPoolData)

	testAddRemoveVM(t, ctx, apis)
	testUpdateVMPool(t, ctx, apis)

	dummy.Stop()
}

func testAddRemoveVM(t *testing.T, ctx context.Context, apis *AllApis) {
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

	_, err := apis.vmPoolApi.AddVMPoolMember(ctx, &cm1)
	require.Nil(t, err, "add cloudlet vm to cloudlet vm pool")

	vmPool := edgeproto.VMPool{}
	found := apis.vmPoolApi.cache.Get(&cm1.Key, &vmPool)
	require.True(t, found, "get cloudlet vm pool %v", cm1.Key)
	require.Equal(t, 4, len(vmPool.Vms))

	// test adding another cloudlet vm to the pool
	cm2 := edgeproto.VMPoolMember{}
	cm2.Key = testutil.VMPoolData[1].Key
	cm2.Vm = edgeproto.VM{
		Name: "vmY",
		NetInfo: edgeproto.VMNetInfo{
			ExternalIp: "0.0.0.0",
			InternalIp: "192.168.100.121",
		},
	}
	_, err = apis.vmPoolApi.AddVMPoolMember(ctx, &cm2)
	require.NotNil(t, err, "invalid external ip")

	cm2.Vm = edgeproto.VM{
		Name: "vmY",
		NetInfo: edgeproto.VMNetInfo{
			ExternalIp: "192.168.1.121",
			InternalIp: "127.0.0.1",
		},
	}
	_, err = apis.vmPoolApi.AddVMPoolMember(ctx, &cm2)
	require.NotNil(t, err, "invalid internal ip")

	cm2.Vm = edgeproto.VM{
		Name: "vmY",
		NetInfo: edgeproto.VMNetInfo{
			ExternalIp: "192.168.1.121",
			InternalIp: "192.168.100.121",
		},
	}
	_, err = apis.vmPoolApi.AddVMPoolMember(ctx, &cm2)
	require.Nil(t, err, "add cloudlet vm to cloudlet vm pool")

	found = apis.vmPoolApi.cache.Get(&cm2.Key, &vmPool)
	require.True(t, found, "get cloudlet vm pool %v", cm2.Key)
	require.Equal(t, 5, len(vmPool.Vms))

	// add/update VM with same external/internal IP as another VM to the pool
	cm3 := edgeproto.VMPoolMember{}
	cm3.Key = testutil.VMPoolData[1].Key
	cm3.Vm = edgeproto.VM{
		Name: "vmZ",
		NetInfo: edgeproto.VMNetInfo{
			ExternalIp: "192.168.1.121",
			InternalIp: "192.168.100.129",
		},
	}

	_, err = apis.vmPoolApi.AddVMPoolMember(ctx, &cm3)
	require.NotNil(t, err, "add cloudlet vm to cloudlet vm pool should fail as same externalIP exists")

	cm3.Vm = edgeproto.VM{
		Name: "vmZ",
		NetInfo: edgeproto.VMNetInfo{
			InternalIp: "192.168.100.121",
		},
	}

	_, err = apis.vmPoolApi.AddVMPoolMember(ctx, &cm3)
	require.NotNil(t, err, "add cloudlet vm to cloudlet vm pool should fail as same internalIP exists")

	updateCM := edgeproto.VMPool{}
	updateCM.Key = testutil.VMPoolData[1].Key
	updateCM.Vms = []edgeproto.VM{
		edgeproto.VM{
			Name: "vmX",
			NetInfo: edgeproto.VMNetInfo{
				ExternalIp: "192.168.1.111",
				InternalIp: "192.168.100.111",
			},
		},
		edgeproto.VM{
			Name: "vmY",
			NetInfo: edgeproto.VMNetInfo{
				InternalIp: "192.168.100.111",
			},
		},
	}
	updateCM.Fields = []string{
		edgeproto.VMPoolFieldVmsNetInfoExternalIp,
		edgeproto.VMPoolFieldVmsNetInfoInternalIp,
	}
	_, err = apis.vmPoolApi.UpdateVMPool(ctx, &updateCM)
	require.NotNil(t, err, "update cloudlet vm should fail as same internalIP exists")

	found = apis.vmPoolApi.cache.Get(&cm3.Key, &vmPool)
	require.True(t, found, "get cloudlet vm pool %v", cm3.Key)
	require.Equal(t, 5, len(vmPool.Vms))

	// remove cloudlet vm from pool
	_, err = apis.vmPoolApi.RemoveVMPoolMember(ctx, &cm1)
	require.Nil(t, err, "remove cloudlet vm from pool")
	found = apis.vmPoolApi.cache.Get(&cm1.Key, &vmPool)
	require.True(t, found, "get cloudlet vm pool %v", cm1.Key)
	require.Equal(t, 4, len(vmPool.Vms))

	// remove cloudlet vm from pool
	_, err = apis.vmPoolApi.RemoveVMPoolMember(ctx, &cm2)
	require.Nil(t, err, "remove cloudlet vm from pool")
	found = apis.vmPoolApi.cache.Get(&cm2.Key, &vmPool)
	require.True(t, found, "get cloudlet vm pool %v", cm2.Key)
	require.Equal(t, 3, len(vmPool.Vms))

	// try to add cloudlet vm to non-existent pool
	cm1.Key.Name = "SomeNonExistentCloudlet"
	_, err = apis.vmPoolApi.AddVMPoolMember(ctx, &cm1)
	require.NotNil(t, err)
}

func testUpdateVMPool(t *testing.T, ctx context.Context, apis *AllApis) {
	dummyResponder := DummyInfoResponder{
		VMPoolCache:    &apis.vmPoolApi.cache,
		RecvVMPoolInfo: apis.vmPoolInfoApi,
	}
	dummyResponder.InitDummyInfoResponder()
	reduceInfoTimeouts(t, ctx, apis)

	// create support data
	testutil.InternalFlavorCreate(t, apis.flavorApi, testutil.FlavorData)

	cl := testutil.CloudletData()[1]
	vmp := testutil.VMPoolData[0]
	cl.VmPool = vmp.Key.Name
	cl.PlatformType = edgeproto.PlatformType_PLATFORM_TYPE_FAKE_VM_POOL
	err := apis.cloudletApi.CreateCloudlet(&cl, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)

	// test adding vm to the pool
	cm1 := edgeproto.VMPoolMember{}
	cm1.Key = vmp.Key
	cm1.Vm = edgeproto.VM{
		Name: "vmX",
		NetInfo: edgeproto.VMNetInfo{
			ExternalIp: "192.168.1.111",
			InternalIp: "192.168.100.111",
		},
	}
	_, err = apis.vmPoolApi.AddVMPoolMember(ctx, &cm1)
	require.Nil(t, err, "add vm to vm pool")

	// simulate vmpool responder failure
	dummyResponder.SetSimulateVMPoolUpdateFailure(true)
	// test adding another vm to the pool, should fail
	cm2 := edgeproto.VMPoolMember{}
	cm2.Key = vmp.Key
	cm2.Vm = edgeproto.VM{
		Name: "vmY",
		NetInfo: edgeproto.VMNetInfo{
			ExternalIp: "192.168.1.121",
			InternalIp: "192.168.100.121",
		},
	}
	_, err = apis.vmPoolApi.AddVMPoolMember(ctx, &cm2)
	require.NotNil(t, err, "crm failure")
	require.Contains(t, err.Error(), "crm update VMPool failed")
	dummyResponder.SetSimulateVMPoolUpdateFailure(false)

	// clean up
	_, err = apis.vmPoolApi.RemoveVMPoolMember(ctx, &cm1)
	require.Nil(t, err, "remove vm from pool")
	err = apis.cloudletApi.DeleteCloudlet(&cl, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)
	testutil.InternalFlavorDelete(t, apis.flavorApi, testutil.FlavorData)
}
