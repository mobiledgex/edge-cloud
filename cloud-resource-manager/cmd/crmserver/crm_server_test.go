package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/fake"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/testutil/testservices"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	yaml "gopkg.in/yaml.v2"
)

// Data in test_data.go is meant to go through controller, which will
// fill in certain fields (like copying Flavor from App to AppInst)
// This test data is post-copying of that data, and limited to just
// testing CRM.
var yamlData = `
cloudlets:
- key:
    organization: DMUUS
    name: cloud2
  vmpool: vmpool1
  gpuconfig:
    driver:
      name: gpudriver1
      organization: DMUUS
  restagmap:
    gpu:
      name: gpudriver1
      organization: DMUUS

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
      name: pillimo_cluster
    cloudletkey:
      organization: DMUUS
      name: cloud2
  flavor:
    name: x1.tiny
  nodes: 3
  liveness: LivenessStatic
- key:
    clusterkey:
      name: Untomt_cluster
    cloudletkey:
      organization: DMUUS
      name: cloud2
  flavor:
    name: x1.small
  nodes: 3
  liveness: LivenessDynamic
  ipaccess: Dedicated
- key:
    clusterkey:
      name: Untomt_cluster_22
    cloudletkey:
      organization: DMUUS
      name: cloud2
  flavor:
    name: x1.small
  nodes: 3
  liveness: LivenessDynamic
  ipaccess: Dedicated

appinstances:
- key:
    appkey:
      organization: Atlantic
      name: Pillimo Go
      version: 1.0.0
    clusterinstkey:
      clusterkey:
        name: pillimo_cluster
      cloudletkey:
        organization: DMUUS
        name: cloud2
      developer: Atlantic
  cloudletloc:
    latitude: 31
    longitude: -91
  liveness: LivenessStatic
  flavor:
    name: x1.tiny
- key:
    appkey:
      organization: Untomt
      name: VRmax
      version: 1.0.0
    clusterinstkey:
      clusterkey:
        name: Untomt_cluster
      cloudletkey:
        organization: DMUUS
        name: cloud2
      developer: Untomt
  cloudletloc:
    latitude: 31
    longitude: -91
  liveness: LivenessDynamic
  flavor:
    name: x1.small
- key:
    appkey:
      organization: Untomt
      name: VRmax
      version: 1.0.0
    clusterinstkey:
      clusterkey:
        name: Untomt_cluster_22
      cloudletkey:
        organization: DMUUS
        name: cloud2
      developer: Untomt
  cloudletloc:
    latitude: 310
    longitude: -910
  liveness: LivenessDynamic
  flavor:
    name: x1.medium

vmpools:
- key:
    organization: DMUUS
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
- key:
    organization: DMUUS
    name: vmpool2
  vms:
  - name: vm1
    netinfo:
      externalip: 192.168.2.101
      internalip: 192.168.200.101

gpudrivers:
- key:
    organization: DMUUS
    name: gpudriver1
  type: GpuTypePassthrough
- key:
    organization: DMUUS
    name: gpudriver2
  type: GpuTypeVgpu
- key:
    name: gpudriver2
  type: GpuTypeVgpu

cloudletpools:
- key:
    organization: DMUUS
    name: cloud2-pool
  cloudlets:
  - cloud2

networks:
- key:
    cloudletkey:
       organization: dmuus
       cloudlet: cloud2
    name: net1
  routes:
  - destinationcidr: 10.200.10.0/24
    nexthopip: 10.200.10.1
  - destinationcidr: 10.140.20.0/24
    nexthopip: 10.140.20.1
  connectiontype: ConnectToClusterNodes

trustpolicyexceptions:
- key:
    appkey:
        organization: Untomt
        name: VRmax
        version: 1.0.0
    cloudletpoolkey:
        organization: DMUUS
        name: cloud2-pool
    name: tpe1
    requiredoutboundconnections:
    - protocol: tcp
      remotecidr: "1.1.1.1/32"
      portrangemin: 1
      portrangemax: 111
- key:
    appkey:
        organization: Untomt
        name: VRmax
        version: 1.0.0
    cloudletpoolkey:
        organization: DMUUS
        name: cloud2-pool
    name: tpe2
    requiredoutboundconnections:
    - protocol: udp
      remotecidr: "2.2.2.2/24"
      portrangemin: 22
      portrangemax: 222
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
	log.InitTracer(nil)
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
	// handle access API
	keyServer := node.NewAccessKeyServer(&ctrlHandler.CloudletCache, "")
	accessKeyGrpcServer := node.AccessKeyGrpcServer{}
	basicUpgradeHandler := node.BasicUpgradeHandler{
		KeyServer: keyServer,
	}
	getPublicCertApi := &cloudcommon.TestPublicCertApi{}
	publicCertManager, err := node.NewPublicCertManager("localhost", getPublicCertApi, "", "")
	require.Nil(t, err)
	tlsConfig, err := publicCertManager.GetServerTlsConfig(ctx)
	require.Nil(t, err)
	accessKeyGrpcServer.Start("127.0.0.1:0", keyServer, tlsConfig, func(server *grpc.Server) {
		edgeproto.RegisterCloudletAccessKeyApiServer(server, &basicUpgradeHandler)
	})
	defer accessKeyGrpcServer.Stop()
	// setup access key
	accessKey, err := node.GenerateAccessKey()
	require.Nil(t, err)
	accessKeyFile := "/tmp/accesskey_crm_unittest"
	err = ioutil.WriteFile(accessKeyFile, []byte(accessKey.PrivatePEM), 0600)
	require.Nil(t, err)

	// update VMPool object before creating cloudlet
	for ii := range data.VmPools {
		ctrlHandler.VMPoolCache.Update(ctx, &data.VmPools[ii], 0)
	}

	// update GPUDriver object before creating cloudlet
	for ii := range data.GpuDrivers {
		ctrlHandler.GPUDriverCache.Update(ctx, &data.GpuDrivers[ii], 0)
	}

	// set Cloudlet state to CRM_INIT to allow crm notify exchange to proceed
	cdata := data.Cloudlets[0]
	cdata.State = edgeproto.TrackedState_CRM_INITOK
	cdata.CrmAccessPublicKey = accessKey.PublicPEM
	ctrlHandler.CloudletCache.Update(ctx, &cdata, 0)
	ctrlMgr.Start("ctrl", notifyAddr, nil)

	os.Args = append(os.Args, "-cloudletKey")
	os.Args = append(os.Args, string(bytes))
	os.Args = append(os.Args, "-notifyAddrs")
	os.Args = append(os.Args, notifyAddr)
	os.Args = append(os.Args, "--accessApiAddr", accessKeyGrpcServer.ApiAddr())
	os.Args = append(os.Args, "--accessKeyFile", accessKeyFile)
	os.Args = append(os.Args, "--HARole", string(process.HARolePrimary))
	nodeMgr.AccessKeyClient.TestSkipTlsVerify = true
	defer nodeMgr.Finish()
	mainDone, err := startMain(t)
	if err != nil {
		close(sigChan)
		require.Nil(t, err, "start main")
		return
	}
	defer func() {
		// closing the signal channel triggers main to exit
		close(sigChan)
		// wait until main is done so it can clean up properly
		<-mainDone
		ctrlMgr.Stop()
	}()

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
		data.AppInstances[ii].State = edgeproto.TrackedState_READY
		ctrlHandler.AppInstCache.Update(ctx, &data.AppInstances[ii], 0)
	}
	for ii := range data.CloudletPools {
		ctrlHandler.CloudletPoolCache.Update(ctx, &data.CloudletPools[ii], 0)
	}

	for ii := range data.Networks {
		ctrlHandler.NetworkCache.Update(ctx, &data.Networks[ii], 0)
	}

	for ii := range data.TrustPolicyExceptions {
		ctrlHandler.TrustPolicyExceptionCache.Update(ctx, &data.TrustPolicyExceptions[ii], 0)
	}

	require.Nil(t, notify.WaitFor(&controllerData.FlavorCache, 3))
	// Note for ClusterInsts and AppInsts, only those that match
	// myCloudlet Key will be sent.
	log.SpanLog(ctx, log.DebugLevelApi, "wait for instances")
	require.Nil(t, notify.WaitFor(&controllerData.ClusterInstCache, 3))
	require.Nil(t, notify.WaitFor(&controllerData.AppInstCache, 3))
	// ensure that only vmpool object associated with cloudlet is received
	require.Nil(t, notify.WaitFor(&controllerData.VMPoolCache, 1))

	require.Nil(t, notify.WaitFor(controllerData.CloudletPoolCache, 1))

	// ensure that only gpudriver object associated with cloudlet is received
	require.Nil(t, notify.WaitFor(&controllerData.GPUDriverCache, 1))
	require.Nil(t, notify.WaitFor(&controllerData.NetworkCache, 1))

	require.Nil(t, notify.WaitFor(&controllerData.TrustPolicyExceptionCache, 2))

	// TODO: check that the above changes triggered cloudlet cluster/app creates
	// for now just check stats
	log.SpanLog(ctx, log.DebugLevelApi, "check counts")
	require.Equal(t, 3, len(controllerData.FlavorCache.Objs))
	require.Equal(t, 3, len(controllerData.ClusterInstCache.Objs))
	require.Equal(t, 3, len(controllerData.AppInstCache.Objs))
	require.Equal(t, 1, len(controllerData.VMPoolCache.Objs))
	require.Equal(t, 1, len(controllerData.GPUDriverCache.Objs))
	require.Equal(t, 1, len(controllerData.CloudletPoolCache.Objs))
	require.Equal(t, 1, len(controllerData.NetworkCache.Objs))
	require.Equal(t, 2, len(controllerData.TrustPolicyExceptionCache.Objs))

	testVMPoolUpdates(t, ctx, &data.VmPools[0], ctrlHandler)

	fakePlatform, ok := platform.(*fake.Platform)
	require.True(t, ok)

	testTrustPolicyExceptionUpdates(t, ctx, ctrlHandler, &data, fakePlatform)

	// delete
	for ii := range data.VmPools {
		ctrlHandler.VMPoolCache.Delete(ctx, &data.VmPools[ii], 0)
	}
	for ii := range data.GpuDrivers {
		ctrlHandler.GPUDriverCache.Delete(ctx, &data.GpuDrivers[ii], 0)
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
	for ii := range data.CloudletPools {
		ctrlHandler.CloudletPoolCache.Delete(ctx, &data.CloudletPools[ii], 0)
	}

	for ii := range data.Networks {
		ctrlHandler.NetworkCache.Delete(ctx, &data.Networks[ii], 0)
	}

	require.Nil(t, notify.WaitFor(&controllerData.FlavorCache, 0))
	require.Nil(t, notify.WaitFor(&controllerData.ClusterInstCache, 0))
	require.Nil(t, notify.WaitFor(&controllerData.AppInstCache, 0))
	require.Nil(t, notify.WaitFor(&controllerData.VMPoolCache, 0))
	require.Nil(t, notify.WaitFor(&controllerData.GPUDriverCache, 0))
	require.Nil(t, notify.WaitFor(controllerData.CloudletPoolCache, 0))
	require.Nil(t, notify.WaitFor(&controllerData.NetworkCache, 0))

	// TODO: check that deletes triggered cloudlet cluster/app deletes.
	require.Equal(t, 0, len(controllerData.FlavorCache.Objs))
	require.Equal(t, 0, len(controllerData.ClusterInstCache.Objs))
	require.Equal(t, 0, len(controllerData.AppInstCache.Objs))
	require.Equal(t, 0, len(controllerData.VMPoolCache.Objs))
	require.Equal(t, 0, len(controllerData.GPUDriverCache.Objs))
	require.Equal(t, 0, len(controllerData.CloudletPoolCache.Objs))
	require.Equal(t, 0, len(controllerData.NetworkCache.Objs))
	require.Equal(t, 0, len(controllerData.TrustPolicyExceptionCache.Objs))

}

func TestNotifyOrder(t *testing.T) {
	_, _, err := nodeMgr.Init(node.NodeTypeCRM, node.NoTlsClientIssuer)
	require.Nil(t, err)
	defer nodeMgr.Finish()
	controllerData = crmutil.NewControllerData(nil, &edgeproto.CloudletKey{}, &nodeMgr, &highAvailabilityManager)
	mgr := notify.ServerMgr{}
	crmutil.InitSrvNotify(&mgr, &nodeMgr, controllerData)
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
		if vm.State == edgeproto.VMState_VM_UPDATE || vm.State == edgeproto.VMState_VM_FORCE_FREE {
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
			if updatedVM.State == edgeproto.VMState_VM_FORCE_FREE {
				require.Equal(t, edgeproto.VMState_VM_FREE, vm.State, "vm is forcefully freed")
				require.Empty(t, vm.InternalName)
				require.Empty(t, vm.GroupName)
			}
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

	// Add another VM which will fail verification
	vmPoolAddMember := copyCrmVMPool()
	vmPoolAddMember.Vms = append(vmPoolAddMember.Vms, edgeproto.VM{
		Name: "vmFailVerification",
		NetInfo: edgeproto.VMNetInfo{
			InternalIp: "192.168.100.106",
		},
		State: edgeproto.VMState_VM_ADD,
	})
	vmPoolAddMember.State = edgeproto.TrackedState_UPDATE_REQUESTED
	verifyVMPoolUpdate(t, ctx, vmPoolAddMember, ctrlHandler, Fail)
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

	// Forcefully free a VM which is busy
	vmPoolUpdate7 := copyCrmVMPool()
	vmPoolUpdate7.Vms[0].State = edgeproto.VMState_VM_FORCE_FREE
	for ii, _ := range vmPoolUpdate7.Vms {
		if vmPoolUpdate7.Vms[ii].State != edgeproto.VMState_VM_FORCE_FREE {
			vmPoolUpdate7.Vms[ii].State = edgeproto.VMState_VM_UPDATE
		}
	}
	vmPoolUpdate7.State = edgeproto.TrackedState_UPDATE_REQUESTED
	verifyVMPoolUpdate(t, ctx, vmPoolUpdate7, ctrlHandler, Pass)
	require.Equal(t, 5, len(controllerData.VMPool.Vms), "matches crm global vmpool")
}

func WaitForTrustPolicyExceptionCacheState(key *edgeproto.TrustPolicyExceptionKey, state edgeproto.TrustPolicyExceptionState) (*edgeproto.TrustPolicyException, error) {
	lastState := edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_UNKNOWN
	tpe := edgeproto.TrustPolicyException{}
	for i := 0; i < 100; i++ {
		if controllerData.TrustPolicyExceptionCache.Get(key, &tpe) {
			if tpe.State == state {
				return &tpe, nil
			}
			lastState = tpe.State
		}
		time.Sleep(10 * time.Millisecond)
	}

	return &tpe, fmt.Errorf("Unable to get desired tpe state, actual state %s, desired state %s", lastState, state)
}

// Single TPE tests
func testTrustPolicyExceptionUpdates1(t *testing.T, ctx context.Context, tpe *edgeproto.TrustPolicyException, clusterInst *edgeproto.ClusterInst, ctrlHandler *notify.DummyHandler, data *edgeproto.AllData, fakePlatform *fake.Platform) {

	log.SpanLog(ctx, log.DebugLevelApi, "############ Begin testTrustPolicyExceptionUpdates1 ##############")

	log.SpanLog(ctx, log.DebugLevelApi, "Input:", "tpe", tpe, "clusterInst", clusterInst)

	// test that adding a TPE with State Approval Requested does not add it to existing AppInst
	log.SpanLog(ctx, log.DebugLevelApi, "############ UT1.1", "fakePlatform", fakePlatform)

	tpe.State = edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_APPROVAL_REQUESTED
	ctrlHandler.TrustPolicyExceptionCache.Update(ctx, tpe, 0)
	state := tpe.State
	_, err := WaitForTrustPolicyExceptionCacheState(&tpe.Key, state)
	require.Nil(t, err)

	found := fakePlatform.HasTrustPolicyException(ctx, &tpe.Key, clusterInst)
	require.False(t, found, "tpe not found")

	// test that Approved TPE is not programmed on any clusters
	count := fakePlatform.TrustPolicyExceptionCount(ctx)
	require.Equal(t, 0, count)

	// test that a TPE with State ACTIVE, for an existing AppInst, adds that TPEs to that ClusterInst
	log.SpanLog(ctx, log.DebugLevelApi, "############ UT1.2")
	tpe.State = edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE
	ctrlHandler.TrustPolicyExceptionCache.Update(ctx, tpe, 0)
	state = tpe.State
	_, err = WaitForTrustPolicyExceptionCacheState(&tpe.Key, state)
	require.Nil(t, err)

	found = fakePlatform.HasTrustPolicyException(ctx, &tpe.Key, clusterInst)
	require.True(t, found, "tpe found")

	// test that Active TPE is programmed on both clusters of cloudletpool
	count = fakePlatform.TrustPolicyExceptionCount(ctx)
	require.Equal(t, 2, count)

	// test that updating a TPE from ApprovalRequested/Active->Rejected does not add it or removes it from existing AppInsts
	log.SpanLog(ctx, log.DebugLevelApi, "############ UT1.3")

	tpe.State = edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_REJECTED
	ctrlHandler.TrustPolicyExceptionCache.Update(ctx, tpe, 0)
	state = tpe.State
	_, err = WaitForTrustPolicyExceptionCacheState(&tpe.Key, state)
	require.Nil(t, err)

	found = fakePlatform.HasTrustPolicyException(ctx, &tpe.Key, clusterInst)
	require.False(t, found, "tpe not found")

	count = fakePlatform.TrustPolicyExceptionCount(ctx)
	require.Equal(t, 0, count)

	// test that updating a TPE from ApprovalRequested/Rejected->Active adds it to existing AppInsts
	log.SpanLog(ctx, log.DebugLevelApi, "############ UT1.4")
	tpe.State = edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE
	ctrlHandler.TrustPolicyExceptionCache.Update(ctx, tpe, 0)
	state = tpe.State
	_, err = WaitForTrustPolicyExceptionCacheState(&tpe.Key, state)
	require.Nil(t, err)

	found = fakePlatform.HasTrustPolicyException(ctx, &tpe.Key, clusterInst)
	require.True(t, found, "tpe found")

	// test that Active TPE is programmed on both clusters of cloudletpool
	count = fakePlatform.TrustPolicyExceptionCount(ctx)
	require.Equal(t, 2, count)

	log.SpanLog(ctx, log.DebugLevelApi, "############# end testTrustPolicyExceptionUpdates1 #############")
}

// Mutliple TPE tests
func testTrustPolicyExceptionUpdates2(t *testing.T, ctx context.Context, tpe *edgeproto.TrustPolicyException, clusterInst *edgeproto.ClusterInst, ctrlHandler *notify.DummyHandler, data *edgeproto.AllData, fakePlatform *fake.Platform) {

	log.SpanLog(ctx, log.DebugLevelApi, "############ Begin testTrustPolicyExceptionUpdates2 ##############")

	log.SpanLog(ctx, log.DebugLevelApi, "Input:", "tpe", tpe, "clusterInst", clusterInst)

	// test that adding a TPE with State Approval Requested does not add it to existing AppInst
	log.SpanLog(ctx, log.DebugLevelApi, "############ UT2.1")

	tpe.State = edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_APPROVAL_REQUESTED
	ctrlHandler.TrustPolicyExceptionCache.Update(ctx, tpe, 0)
	state := tpe.State
	_, err := WaitForTrustPolicyExceptionCacheState(&tpe.Key, state)
	require.Nil(t, err)

	found := fakePlatform.HasTrustPolicyException(ctx, &tpe.Key, clusterInst)
	require.False(t, found, "tpe not found")

	// test that the new Approved TPE is not programmed on any clusters, count should still be the old count
	count := fakePlatform.TrustPolicyExceptionCount(ctx)
	require.Equal(t, 2, count)

	// test that a TPE with State ACTIVE, for an existing AppInst, adds that TPEs to that ClusterInst
	log.SpanLog(ctx, log.DebugLevelApi, "############ UT2.2")
	tpe.State = edgeproto.TrustPolicyExceptionState_TRUST_POLICY_EXCEPTION_STATE_ACTIVE
	ctrlHandler.TrustPolicyExceptionCache.Update(ctx, tpe, 0)
	state = tpe.State
	_, err = WaitForTrustPolicyExceptionCacheState(&tpe.Key, state)
	require.Nil(t, err)

	found = fakePlatform.HasTrustPolicyException(ctx, &tpe.Key, clusterInst)
	require.True(t, found, "tpe found")

	// test that Multiple TPEs are configured per app, on multiple clusters. Total count increases to 4
	count = fakePlatform.TrustPolicyExceptionCount(ctx)
	require.Equal(t, 4, count)

	log.SpanLog(ctx, log.DebugLevelApi, "############# end testTrustPolicyExceptionUpdates2 #############")
}

// CloudletPool related changes and TPE tests
func testTrustPolicyExceptionUpdates3(t *testing.T, ctx context.Context, ctrlHandler *notify.DummyHandler, data *edgeproto.AllData, fakePlatform *fake.Platform) {

	log.SpanLog(ctx, log.DebugLevelApi, "############ begin CloudletPoolCache.Update")
	count := fakePlatform.TrustPolicyExceptionCount(ctx)
	require.Equal(t, 4, count)

	var CloudletsSaved [][]string

	for ii := range data.CloudletPools {
		cloudletPool := &data.CloudletPools[ii]
		CloudletsSaved = append(CloudletsSaved, cloudletPool.Cloudlets)
		cloudletPool.Cloudlets = nil
		ctrlHandler.CloudletPoolCache.Update(ctx, cloudletPool, 0)
	}

	// We've only removed a cloudlet, but cloudletPool cache count remains the same
	// hence notify wont sleep, poll for count test
	for ii := 0; ii < 20; ii++ {
		count = fakePlatform.TrustPolicyExceptionCount(ctx)
		if count == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	require.Equal(t, 0, count)

	log.SpanLog(ctx, log.DebugLevelApi, "############ restore CloudletPoolCache.Update")

	for ii := range data.CloudletPools {
		cloudletPool := &data.CloudletPools[ii]
		cloudletPool.Cloudlets = CloudletsSaved[ii]
		ctrlHandler.CloudletPoolCache.Update(ctx, cloudletPool, 0)
	}

	// We've only removed a cloudlet, but cloudletPool cache count remains the same
	// hence notify wont sleep, poll for count test
	for ii := 0; ii < 20; ii++ {
		count = fakePlatform.TrustPolicyExceptionCount(ctx)
		if count == 4 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	require.Equal(t, 4, count)

	log.SpanLog(ctx, log.DebugLevelApi, "############ end CloudletPoolCache.Update")

	// test that deleting a TPE removes it from existing AppInsts/ClusterInsts

	for ii := range data.TrustPolicyExceptions {
		ctrlHandler.TrustPolicyExceptionCache.Delete(ctx, &data.TrustPolicyExceptions[ii], 0)
	}

	require.Nil(t, notify.WaitFor(&controllerData.TrustPolicyExceptionCache, 0))

	found := fakePlatform.HasTrustPolicyException(ctx, &data.TrustPolicyExceptions[0].Key, &data.ClusterInsts[1])
	require.False(t, found, "tpe not found")

	count = fakePlatform.TrustPolicyExceptionCount(ctx)
	require.Equal(t, 0, count)

}

func testTrustPolicyExceptionUpdates(t *testing.T, ctx context.Context, ctrlHandler *notify.DummyHandler, data *edgeproto.AllData, fakePlatform *fake.Platform) {
	tpe := &data.TrustPolicyExceptions[0]
	clusterInst := data.ClusterInsts[1]

	testTrustPolicyExceptionUpdates1(t, ctx, tpe, &clusterInst, ctrlHandler, data, fakePlatform)

	tpe = &data.TrustPolicyExceptions[1]
	testTrustPolicyExceptionUpdates2(t, ctx, tpe, &clusterInst, ctrlHandler, data, fakePlatform)

	testTrustPolicyExceptionUpdates3(t, ctx, ctrlHandler, data, fakePlatform)

}
