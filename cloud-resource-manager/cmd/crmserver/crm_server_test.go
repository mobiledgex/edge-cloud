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

cloudletvmpools:
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
	for ii := range data.CloudletVmPools {
		ctrlHandler.CloudletVMPoolCache.Update(ctx, &data.CloudletVmPools[ii], 0)
	}
	notify.WaitFor(&controllerData.FlavorCache, 3)
	// Note for ClusterInsts and AppInsts, only those that match
	// myCloudlet Key will be sent.
	log.SpanLog(ctx, log.DebugLevelApi, "wait for instances")
	notify.WaitFor(&controllerData.ClusterInstCache, 2)
	notify.WaitFor(&controllerData.AppInstCache, 2)
	notify.WaitFor(&controllerData.CloudletVMPoolCache, 1)

	// TODO: check that the above changes triggered cloudlet cluster/app creates
	// for now just check stats
	log.SpanLog(ctx, log.DebugLevelApi, "check counts")
	require.Equal(t, 3, len(controllerData.FlavorCache.Objs))
	require.Equal(t, 2, len(controllerData.ClusterInstCache.Objs))
	require.Equal(t, 2, len(controllerData.AppInstCache.Objs))
	require.Equal(t, 1, len(controllerData.CloudletVMPoolCache.Objs))

	testCloudletVMPoolInfo(t, ctx, ctrlHandler, data.CloudletVmPools)

	// delete
	for ii := range data.CloudletVmPools {
		ctrlHandler.CloudletVMPoolCache.Delete(ctx, &data.CloudletVmPools[ii], 0)
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
	notify.WaitFor(&controllerData.CloudletVMPoolCache, 0)

	// TODO: check that deletes triggered cloudlet cluster/app deletes.
	require.Equal(t, 0, len(controllerData.FlavorCache.Objs))
	require.Equal(t, 0, len(controllerData.ClusterInstCache.Objs))
	require.Equal(t, 0, len(controllerData.AppInstCache.Objs))
	require.Equal(t, 0, len(controllerData.CloudletVMPoolCache.Objs))

	// closing the signal channel triggers main to exit
	close(sigChan)
	// wait until main is done so it can clean up properly
	<-mainDone
	ctrlMgr.Stop()
}

func waitForAction(key *edgeproto.CloudletKey, action edgeproto.CloudletVMAction) (*edgeproto.CloudletVMPoolInfo, error) {
	info := edgeproto.CloudletVMPoolInfo{}
	var lastAction edgeproto.CloudletVMAction
	for i := 0; i < 100; i++ {
		if controllerData.CloudletVMPoolInfoCache.Get(key, &info) {
			if info.Action == action {
				return &info, nil
			}
			lastAction = info.Action
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil, fmt.Errorf("Unable to get desired Cloudlet VM Pool action, actual action %s, desired action %s", lastAction, action)
}

func testCloudletVMPoolInfo(t *testing.T, ctx context.Context, ctrlHandler *notify.DummyHandler, cloudletVmPools []edgeproto.CloudletVMPool) {
	vmPool := cloudletVmPools[0]

	// Allocate VMs from the pool
	info := edgeproto.CloudletVMPoolInfo{}
	info.Key = vmPool.Key
	info.User = "cluster:testcluster,cluster-org:testorg"
	info.VmSpecs = []edgeproto.CloudletVMSpec{
		edgeproto.CloudletVMSpec{
			ExternalNetwork: true,
		},
		edgeproto.CloudletVMSpec{
			InternalNetwork: true,
		},
	}
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_ALLOCATE
	controllerData.CloudletVMPoolInfoCache.Update(ctx, &info, 0)

	edgeproto.AllocateCloudletVMsFromPool(ctx, &info, &vmPool)
	ctrlHandler.CloudletVMPoolCache.Update(ctx, &vmPool, 0)

	// wait for cloudletvmpoolinfo action to get changed to done
	infoFound, err := waitForAction(&vmPool.Key, edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE)
	require.Nil(t, err, "get valid cloudlet vm pool info for %v, desired action %s", info.Key, info.Action)
	require.Empty(t, infoFound.Error, "cloudlet vm pool allocation failed")
	require.Equal(t, 2, len(infoFound.CloudletVms), "get desired number of vms for %v", info.Key)

	vmPool.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE
	ctrlHandler.CloudletVMPoolCache.Update(ctx, &vmPool, 0)

	// Allocate some more VMs from the pool
	info.VmSpecs = []edgeproto.CloudletVMSpec{
		edgeproto.CloudletVMSpec{
			InternalNetwork: true,
		},
	}
	info.CloudletVms = []edgeproto.CloudletVM{}
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_ALLOCATE
	controllerData.CloudletVMPoolInfoCache.Update(ctx, &info, 0)

	edgeproto.AllocateCloudletVMsFromPool(ctx, &info, &vmPool)
	ctrlHandler.CloudletVMPoolCache.Update(ctx, &vmPool, 0)

	// wait for cloudletvmpoolinfo action to get changed to done
	infoFound, err = waitForAction(&vmPool.Key, edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE)
	require.Nil(t, err)
	require.Empty(t, infoFound.Error, "cloudlet vm pool allocation failed")
	require.Equal(t, 3, len(infoFound.CloudletVms), "get desired number of vms for %v", info.Key)

	vmPool.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE
	ctrlHandler.CloudletVMPoolCache.Update(ctx, &vmPool, 0)

	// Allocate some more VMs from the pool, should fail
	info.VmSpecs = []edgeproto.CloudletVMSpec{
		edgeproto.CloudletVMSpec{
			ExternalNetwork: true,
		},
		edgeproto.CloudletVMSpec{
			ExternalNetwork: true,
			InternalNetwork: true,
		},
	}
	info.CloudletVms = []edgeproto.CloudletVM{}
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_ALLOCATE
	controllerData.CloudletVMPoolInfoCache.Update(ctx, &info, 0)

	edgeproto.AllocateCloudletVMsFromPool(ctx, &info, &vmPool)
	ctrlHandler.CloudletVMPoolCache.Update(ctx, &vmPool, 0)

	// wait for cloudletvmpoolinfo action to get changed to done
	infoFound, err = waitForAction(&vmPool.Key, edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE)
	require.Nil(t, err)
	require.NotEmpty(t, infoFound.Error, "cloudlet vm pool allocation failed")
	require.Equal(t, 3, len(infoFound.CloudletVms), "get desired number of vms for %v", info.Key)

	vmPool.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE
	ctrlHandler.CloudletVMPoolCache.Update(ctx, &vmPool, 0)

	// Release VMs from the pool
	info.VmSpecs = []edgeproto.CloudletVMSpec{}
	for _, vm := range info.CloudletVms {
		info.VmSpecs = append(info.VmSpecs, edgeproto.CloudletVMSpec{
			Name: vm.Name,
		})
	}
	info.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_RELEASE
	controllerData.CloudletVMPoolInfoCache.Update(ctx, &info, 0)

	edgeproto.ReleaseCloudletVMsFromPool(ctx, &info, &vmPool)
	ctrlHandler.CloudletVMPoolCache.Update(ctx, &vmPool, 0)

	// wait for cloudletvmpoolinfo action to get changed to done
	infoFound, err = waitForAction(&vmPool.Key, edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE)
	require.Nil(t, err)
	require.Empty(t, infoFound.Error, "cloudlet vm pool allocation failed")
	require.Equal(t, 0, len(infoFound.CloudletVms), "get desired number of vms for %v", info.Key)

	vmPool.Action = edgeproto.CloudletVMAction_CLOUDLET_VM_ACTION_DONE
	ctrlHandler.CloudletVMPoolCache.Update(ctx, &vmPool, 0)
}
