package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
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
    name: cloud2
  cloudletvms:
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

	testVMPoolInfo(t, ctx, ctrlHandler, data.VmPools)

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

func waitForAction(key *edgeproto.CloudletKey, action edgeproto.CloudletVMAction) (*edgeproto.VMPoolInfo, error) {
	info := edgeproto.VMPoolInfo{}
	var lastAction edgeproto.CloudletVMAction
	for i := 0; i < 100; i++ {
		if controllerData.VMPoolInfoCache.Get(key, &info) {
			if info.Action == action {
				return &info, nil
			}
			lastAction = info.Action
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil, fmt.Errorf("Unable to get desired Cloudlet VM Pool action, actual action %s, desired action %s", lastAction, action)
}

var Pass = true
var Fail = false

func verifyCloudletVMAction(t *testing.T, ctx context.Context, info *edgeproto.VMPoolInfo, vmPool *edgeproto.VMPool, ctrlHandler *notify.DummyHandler, vmCount int, success bool) {
	controllerData.VMPoolInfoCache.Update(ctx, info, 0)

	if info.Action == edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_ALLOCATE {
		edgeproto.AllocateCloudletVMsFromPool(ctx, info, vmPool)
	} else {
		edgeproto.ReleaseCloudletVMsFromPool(ctx, info, vmPool)
	}
	ctrlHandler.VMPoolCache.Update(ctx, vmPool, 0)

	// wait for vmpoolinfo action to get changed to done
	infoFound, err := waitForAction(&info.Key, edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE)
	require.Nil(t, err)
	if success {
		require.Empty(t, infoFound.Error)
	} else {
		require.NotEmpty(t, infoFound.Error)
	}
	require.Equal(t, vmCount, len(infoFound.CloudletVms), "get desired number of vms for %v", info.Key)

	vmPool.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE
	ctrlHandler.VMPoolCache.Update(ctx, vmPool, 0)
}

func testVMPoolInfo(t *testing.T, ctx context.Context, ctrlHandler *notify.DummyHandler, vmPools []edgeproto.VMPool) {
	vmPool := vmPools[0]
	info := edgeproto.VMPoolInfo{}

	// Allocate VMs from the pool
	info.Key = vmPool.Key
	info.User = "testvmpoolvms1"
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
	verifyCloudletVMAction(t, ctx, &info, &vmPool, ctrlHandler, 2, Pass)

	// Allocate some more VMs from the pool by different user
	info.User = "testvmpoolvms2"
	info.VmSpecs = []edgeproto.CloudletVMSpec{
		edgeproto.CloudletVMSpec{
			InternalName:    "vm3.testcluster.testorg.mobiledgex.net",
			ExternalNetwork: true,
			InternalNetwork: true,
		},
	}
	info.CloudletVms = []edgeproto.CloudletVM{}
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_ALLOCATE
	verifyCloudletVMAction(t, ctx, &info, &vmPool, ctrlHandler, 1, Pass)

	// Allocate some more VMs from the pool, should fail
	info.User = "testvmpoolvms1"
	info.VmSpecs = []edgeproto.CloudletVMSpec{
		edgeproto.CloudletVMSpec{
			InternalName:    "vm4.testcluster.testorg.mobiledgex.net",
			InternalNetwork: true,
		},
		edgeproto.CloudletVMSpec{
			InternalName:    "vm5.testcluster.testorg.mobiledgex.net",
			InternalNetwork: true,
		},
	}
	info.CloudletVms = []edgeproto.CloudletVM{}
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_ALLOCATE
	verifyCloudletVMAction(t, ctx, &info, &vmPool, ctrlHandler, 0, Fail)

	// Release 1 VM from the pool
	info.VmSpecs = []edgeproto.CloudletVMSpec{
		edgeproto.CloudletVMSpec{
			InternalName: "vm2.testcluster.testorg.mobiledgex.net",
		},
	}
	info.User = "testvmpoolvms1"
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_RELEASE
	verifyCloudletVMAction(t, ctx, &info, &vmPool, ctrlHandler, 0, Pass)

	// Retry: Allocate some more VMs from the pool, should fail
	info.VmSpecs = []edgeproto.CloudletVMSpec{
		edgeproto.CloudletVMSpec{
			InternalName:    "vm4.testcluster.testorg.mobiledgex.net",
			InternalNetwork: true,
		},
		edgeproto.CloudletVMSpec{
			InternalName:    "vm5.testcluster.testorg.mobiledgex.net",
			InternalNetwork: true,
		},
	}
	info.CloudletVms = []edgeproto.CloudletVM{}
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_ALLOCATE
	verifyCloudletVMAction(t, ctx, &info, &vmPool, ctrlHandler, 2, Pass)

	// Release VMs from the pool
	info.VmSpecs = []edgeproto.CloudletVMSpec{}
	info.User = "testvmpoolvms1"
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_RELEASE
	verifyCloudletVMAction(t, ctx, &info, &vmPool, ctrlHandler, 0, Pass)

	// Release VMs from the pool
	info.VmSpecs = []edgeproto.CloudletVMSpec{}
	info.User = "testvmpoolvms2"
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_RELEASE
	verifyCloudletVMAction(t, ctx, &info, &vmPool, ctrlHandler, 0, Pass)
}
