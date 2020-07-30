package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/testutil/testservices"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

// Data in test_data.go is meant to go through controller, which will
// fill in certain fields (like copying Flavor from App to AppInst)
// This test data is post-copying of that data, and limited to just
// testing CRM.
var yamlData = `
cloudlets:
- key:
    organization: TMUS
    name: cloud2

flavors:
- key:
    name: x1.tiny
  ram: 1024
  vcpus: 1
  disk: 1
- key:
    name: x1.small
  ram: 2048
  vcpus: 2
  disk: 2
- key:
    name: x1.medium
  ram: 4096
  vcpus: 4
  disk: 4

clusterinsts:
- key:
    clusterkey:
      name: pokemon_cluster
    cloudletkey:
      organization: TMUS
      name: cloud2
  flavor:
    name: x1.tiny
  nodes: 3
  liveness: LivenessStatic
- key:
    clusterkey:
      name: 1000realities_cluster
    cloudletkey:
      organization: TMUS
      name: cloud2
  flavor:
    name: x1.small
  nodes: 3
  liveness: LivenessDynamic

appinstances:
- key:
    appkey:
      organization: Niantic
      name: Pokemon Go
      version: 1.0.0
    clusterinstkey:
      clusterkey:
        name: pokemon_cluster
      cloudletkey:
        organization: TMUS
        name: cloud2
      developer: Niantic
  cloudletloc:
    latitude: 31
    longitude: -91
  liveness: LivenessStatic
  flavor:
    name: x1.tiny
- key:
    appkey:
      organization: 1000realities
      name: VRmax
      version: 1.0.0
    clusterinstkey:
      clusterkey:
        name: 1000realities_cluster
      cloudletkey:
        organization: TMUS
        name: cloud2
      developer: 1000realities
  cloudletloc:
    latitude: 31
    longitude: -91
  liveness: LivenessDynamic
  flavor:
    name: x1.small

vmpools:
- key:
    organization: TMUS
    name: vmpool1
  vms:
  - name: vm1
    netinfo:
      externalip: 192.168.1.101
      internalip: 192.168.100.101
  - name: vm2
    netinfo:
      externalip: 192.168.1.102
      internalip: 192.168.100.102
  - name: vm3
    netinfo:
      internalip: 192.168.100.103
  - name: vm4
    netinfo:
      internalip: 192.168.100.104
`

func startMain(t *testing.T) (chan struct{}, error) {
	mainStarted = make(chan struct{})
	mainDone := make(chan struct{})
	*platformName = "PLATFORM_TYPE_FAKE"
	go func() {
		main()
		close(mainDone)
	}()
	// wait until main is ready
	select {
	case <-mainStarted:
	case <-mainDone:
		return nil, fmt.Errorf("main unexpectedly quit")
	}
	return mainDone, nil
}

func TestCRM(t *testing.T) {
	var err error
	log.SetDebugLevel(log.DebugLevelApi | log.DebugLevelNotify | log.DebugLevelInfra)
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	notifyAddr := "127.0.0.1:61245"

	data := edgeproto.AllData{}
	err = yaml.Unmarshal([]byte(yamlData), &data)
	require.Nil(t, err, "unmarshal yaml data")

	bytes, _ := json.Marshal(&data.Cloudlets[0].Key)

	// CRM is driven by controller
	ctrlHandler := notify.NewDummyHandler()
	ctrlMgr := notify.ServerMgr{}
	ctrlHandler.RegisterServer(&ctrlMgr)
	// set Cloudlet state to CRM_INIT to allow crm notify exchange to proceed
	cdata := data.Cloudlets[0]
	cdata.State = edgeproto.TrackedState_CRM_INITOK
	ctrlHandler.CloudletCache.Update(ctx, &cdata, 0)
	ctrlMgr.Start("ctrl", notifyAddr, nil)

	os.Args = append(os.Args, "-cloudletKey")
	os.Args = append(os.Args, string(bytes))
	os.Args = append(os.Args, "-notifyAddrs")
	os.Args = append(os.Args, notifyAddr)
	mainDone, err := startMain(t)
	if err != nil {
		close(sigChan)
		require.Nil(t, err, "start main")
		return
	}

	notifyClient.WaitForConnect(1)
	stats := notify.Stats{}
	notifyClient.GetStats(&stats)
	require.Equal(t, uint64(1), stats.Connects)

	// Add data to controller
	for ii := range data.Flavors {
		ctrlHandler.FlavorCache.Update(ctx, &data.Flavors[ii], 0)
	}
	for ii := range data.ClusterInsts {
		ctrlHandler.ClusterInstCache.Update(ctx, &data.ClusterInsts[ii], 0)
	}
	for ii := range data.AppInstances {
		ctrlHandler.AppInstCache.Update(ctx, &data.AppInstances[ii], 0)
	}
	for ii := range data.VmPools {
		ctrlHandler.VMPoolCache.Update(ctx, &data.VmPools[ii], 0)
	}
	notify.WaitFor(&controllerData.FlavorCache, 3)
	// Note for ClusterInsts and AppInsts, only those that match
	// myCloudlet Key will be sent.
	log.SpanLog(ctx, log.DebugLevelApi, "wait for instances")
	notify.WaitFor(&controllerData.ClusterInstCache, 2)
	notify.WaitFor(&controllerData.AppInstCache, 2)
	notify.WaitFor(&controllerData.VMPoolCache, 1)

	// TODO: check that the above changes triggered cloudlet cluster/app creates
	// for now just check stats
	log.SpanLog(ctx, log.DebugLevelApi, "check counts")
	require.Equal(t, 3, len(controllerData.FlavorCache.Objs))
	require.Equal(t, 2, len(controllerData.ClusterInstCache.Objs))
	require.Equal(t, 2, len(controllerData.AppInstCache.Objs))
	require.Equal(t, 1, len(controllerData.VMPoolCache.Objs))

	testVMPoolUpdates(t, ctx, &data.VmPools[0], ctrlHandler)

	// delete
	for ii := range data.VmPools {
		ctrlHandler.VMPoolCache.Delete(ctx, &data.VmPools[ii], 0)
	}
	for ii := range data.AppInstances {
		ctrlHandler.AppInstCache.Delete(ctx, &data.AppInstances[ii], 0)
	}
	for ii := range data.ClusterInsts {
		ctrlHandler.ClusterInstCache.Delete(ctx, &data.ClusterInsts[ii], 0)
	}
	for ii := range data.Flavors {
		ctrlHandler.FlavorCache.Delete(ctx, &data.Flavors[ii], 0)
	}
	notify.WaitFor(&controllerData.FlavorCache, 0)
	notify.WaitFor(&controllerData.ClusterInstCache, 0)
	notify.WaitFor(&controllerData.AppInstCache, 0)
	notify.WaitFor(&controllerData.VMPoolCache, 0)

	// TODO: check that deletes triggered cloudlet cluster/app deletes.
	require.Equal(t, 0, len(controllerData.FlavorCache.Objs))
	require.Equal(t, 0, len(controllerData.ClusterInstCache.Objs))
	require.Equal(t, 0, len(controllerData.AppInstCache.Objs))
	require.Equal(t, 0, len(controllerData.VMPoolCache.Objs))

	// closing the signal channel triggers main to exit
	close(sigChan)
	// wait until main is done so it can clean up properly
	<-mainDone
	ctrlMgr.Stop()
}

func TestNotifyOrder(t *testing.T) {
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	err := nodeMgr.Init(ctx, node.NodeTypeCRM)
	require.Nil(t, err)
	controllerData = crmutil.NewControllerData(nil, &nodeMgr)
	mgr := notify.ServerMgr{}
	initSrvNotify(&mgr)
	testservices.CheckNotifySendOrder(t, mgr.GetSendOrder())
}

func WaitForInfoState(key *edgeproto.VMPoolKey, state edgeproto.TrackedState) (*edgeproto.VMPoolInfo, error) {
	lastState := edgeproto.TrackedState_TRACKED_STATE_UNKNOWN
	info := edgeproto.VMPoolInfo{}
	for i := 0; i < 100; i++ {
		if controllerData.VMPoolInfoCache.Get(key, &info) {
			if info.State == state {
				return &info, nil
			}
			lastState = info.State
		}
		time.Sleep(10 * time.Millisecond)
	}

	return &info, fmt.Errorf("Unable to get desired vmPool state, actual state %s, desired state %s", lastState, state)
}

func WaitForVMPoolState(key *edgeproto.VMPoolKey, state edgeproto.TrackedState) (*edgeproto.VMPool, error) {
	lastState := edgeproto.TrackedState_TRACKED_STATE_UNKNOWN
	pool := edgeproto.VMPool{}
	for i := 0; i < 100; i++ {
		if controllerData.VMPoolCache.Get(key, &pool) {
			if pool.State == state {
				return &pool, nil
			}
			lastState = pool.State
		}
		time.Sleep(10 * time.Millisecond)
	}

	return &pool, fmt.Errorf("Unable to get desired vmPool state, actual state %s, desired state %s", lastState, state)
}

var (
	Pass = true
	Fail = false
)

func verifyVMPoolUpdate(t *testing.T, ctx context.Context, vmPool *edgeproto.VMPool, ctrlHandler *notify.DummyHandler, pass bool) {
	key := &vmPool.Key

	ctrlHandler.VMPoolCache.Update(ctx, vmPool, 0)

	state := edgeproto.TrackedState_READY
	if !pass {
		state = edgeproto.TrackedState_UPDATE_ERROR
	}
	info, err := WaitForInfoState(key, state)
	require.Nil(t, err)

	// clear vmpoolinfo for next run
	controllerData.VMPoolInfoCache.Delete(ctx, info, 0)

	vmPool.State = edgeproto.TrackedState_READY
	ctrlHandler.VMPoolCache.Update(ctx, vmPool, 0)
	WaitForVMPoolState(key, edgeproto.TrackedState_READY)

	if !pass {
		return
	}

	count := 0
	addVMs := make(map[string]struct{})
	deleteVMs := make(map[string]struct{})
	updateVMs := make(map[string]edgeproto.VM)
	for _, vm := range vmPool.Vms {
		if vm.State == edgeproto.VMState_VM_UPDATE {
			updateVMs[vm.Name] = vm
			count++
		} else {
			if vm.State == edgeproto.VMState_VM_ADD {
				addVMs[vm.Name] = struct{}{}
			} else if vm.State == edgeproto.VMState_VM_REMOVE {
				deleteVMs[vm.Name] = struct{}{}
				continue
			}
			count++
		}
	}
	require.Equal(t, count, len(info.Vms), "info has expected vms")

	added := false
	for _, vm := range info.Vms {
		if _, ok := addVMs[vm.Name]; ok {
			require.Equal(t, edgeproto.VMState_VM_FREE, vm.State, "vm state matches")
			added = true
		}
		if _, ok := deleteVMs[vm.Name]; ok {
			require.False(t, ok, "vm should be removed")
		}
		if len(updateVMs) > 0 {
			updatedVM, ok := updateVMs[vm.Name]
			require.True(t, ok, "vm should be updated")
			require.Equal(t, updatedVM.NetInfo.InternalIp, vm.NetInfo.InternalIp, "vm internal ip matches")
			require.Equal(t, updatedVM.NetInfo.ExternalIp, vm.NetInfo.ExternalIp, "vm external ip matches")
		}

	}
	if len(addVMs) > 0 {
		require.True(t, added, "vm found")
	}
}

func copyCrmVMPool() *edgeproto.VMPool {
	vmPool := edgeproto.VMPool{}
	vmPool.DeepCopyIn(&controllerData.VMPool)
	return &vmPool
}

func testVMPoolUpdates(t *testing.T, ctx context.Context, vmPool *edgeproto.VMPool, ctrlHandler *notify.DummyHandler) {
	controllerData.VMPool = *vmPool

	// Add new VM
	vmPoolUpdate1 := copyCrmVMPool()
	vmPoolUpdate1.Vms = append(vmPoolUpdate1.Vms, edgeproto.VM{
		Name: "vm5",
		NetInfo: edgeproto.VMNetInfo{
			InternalIp: "192.168.100.105",
		},
		State: edgeproto.VMState_VM_ADD,
	})
	vmPoolUpdate1.Vms = append(vmPoolUpdate1.Vms, edgeproto.VM{
		Name: "vm6",
		NetInfo: edgeproto.VMNetInfo{
			InternalIp: "192.168.100.106",
		},
		State: edgeproto.VMState_VM_ADD,
	})
	vmPoolUpdate1.State = edgeproto.TrackedState_UPDATE_REQUESTED
	verifyVMPoolUpdate(t, ctx, vmPoolUpdate1, ctrlHandler, Pass)
	require.Equal(t, 6, len(controllerData.VMPool.Vms), "matches crm global vmpool")

	// Remove VM
	vmPoolUpdate2 := copyCrmVMPool()
	vmPoolUpdate2.Vms[5].State = edgeproto.VMState_VM_REMOVE
	vmPoolUpdate2.State = edgeproto.TrackedState_UPDATE_REQUESTED
	verifyVMPoolUpdate(t, ctx, vmPoolUpdate2, ctrlHandler, Pass)
	require.Equal(t, 5, len(controllerData.VMPool.Vms), "matches crm global vmpool")

	// Update VM
	vmPoolUpdate3 := copyCrmVMPool()
	controllerData.VMPool.Vms[0].State = edgeproto.VMState_VM_IN_USE
	controllerData.VMPool.Vms[0].InternalName = "testname"
	controllerData.VMPool.Vms[0].GroupName = "testgroup"
	vmPoolUpdate3.Vms[3].NetInfo.ExternalIp = "1.1.1.1"
	// As part of Update, remove a VM as well
	vmPoolUpdate3.Vms = vmPoolUpdate3.Vms[:len(vmPoolUpdate3.Vms)-1]
	// As part of Update, add a VM as well
	vmPoolUpdate3.Vms = append(vmPoolUpdate3.Vms, edgeproto.VM{
		Name: "vm6",
		NetInfo: edgeproto.VMNetInfo{
			InternalIp: "192.168.100.106",
		},
	})
	for ii, _ := range vmPoolUpdate3.Vms {
		vmPoolUpdate3.Vms[ii].State = edgeproto.VMState_VM_UPDATE
	}
	vmPoolUpdate3.State = edgeproto.TrackedState_UPDATE_REQUESTED
	verifyVMPoolUpdate(t, ctx, vmPoolUpdate3, ctrlHandler, Pass)
	require.Equal(t, 5, len(controllerData.VMPool.Vms), "matches crm global vmpool")
	// Make sure existing VM details didnt change as part of update
	require.Equal(t, edgeproto.VMState_VM_IN_USE, controllerData.VMPool.Vms[0].State, "matches old entry")
	require.Equal(t, "testname", controllerData.VMPool.Vms[0].InternalName, "matches old entry")
	require.Equal(t, "testgroup", controllerData.VMPool.Vms[0].GroupName, "matches old entry")

	// Add already existing VM, should fail
	vmPoolUpdate4 := copyCrmVMPool()
	vmPoolUpdate4.Vms = append(vmPoolUpdate4.Vms, edgeproto.VM{
		Name: "vm6",
		NetInfo: edgeproto.VMNetInfo{
			InternalIp: "192.168.100.106",
		},
		State: edgeproto.VMState_VM_ADD,
	})
	vmPoolUpdate4.State = edgeproto.TrackedState_UPDATE_REQUESTED
	verifyVMPoolUpdate(t, ctx, vmPoolUpdate4, ctrlHandler, Fail)
	require.Equal(t, 5, len(controllerData.VMPool.Vms), "matches crm global vmpool")

	// Remove a VM which is busy
	vmPoolUpdate5 := copyCrmVMPool()
	vmPoolUpdate5.Vms[0].State = edgeproto.VMState_VM_REMOVE
	vmPoolUpdate5.State = edgeproto.TrackedState_UPDATE_REQUESTED
	verifyVMPoolUpdate(t, ctx, vmPoolUpdate5, ctrlHandler, Fail)
	require.Equal(t, 5, len(controllerData.VMPool.Vms), "matches crm global vmpool")

	// Update a VM which is busy
	vmPoolUpdate6 := copyCrmVMPool()
	vmPoolUpdate6.Vms[0].NetInfo.ExternalIp = "1.1.1.1"
	vmPoolUpdate6.Vms = append(vmPoolUpdate6.Vms, edgeproto.VM{
		Name: "vm6",
		NetInfo: edgeproto.VMNetInfo{
			InternalIp: "192.168.100.105",
		},
	})
	for ii, _ := range vmPoolUpdate6.Vms {
		vmPoolUpdate6.Vms[ii].State = edgeproto.VMState_VM_UPDATE
	}
	vmPoolUpdate6.State = edgeproto.TrackedState_UPDATE_REQUESTED
	verifyVMPoolUpdate(t, ctx, vmPoolUpdate6, ctrlHandler, Fail)
	require.Equal(t, 5, len(controllerData.VMPool.Vms), "matches crm global vmpool")
}
