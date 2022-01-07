package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/fake"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
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
    organization: TMUS
    name: cloud2
  vmpool: vmpool1
  gpuconfig:
    driver:
      name: gpudriver1
      organization: TMUS
  restagmap:
    gpu:
      name: gpudriver1
      organization: TMUS

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
  ipaccess: Dedicated
- key:
    clusterkey:
      name: 1000realities_cluster_22
    cloudletkey:
      organization: TMUS
      name: cloud2
  flavor:
    name: x1.small
  nodes: 3
  liveness: LivenessDynamic
  ipaccess: Dedicated



`

var primaryStarted chan struct{}
var secondaryStarted chan struct{}

func startMainPrimary(t *testing.T) (chan struct{}, error) {
	primaryStarted = make(chan struct{})
	primaryDone := make(chan struct{})
	go func() {
		main()
		close(primaryDone)
	}()
	// wait until main is ready
	select {
	case <-primaryStarted:
	case <-primaryDone:
		return nil, fmt.Errorf("primary unexpectedly quit")
	}
	return primaryDone, nil
}

func TestCRMHA(t *testing.T) {
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
func copyCrmVMPool() *edgeproto.VMPool {
	vmPool := edgeproto.VMPool{}
	vmPool.DeepCopyIn(&controllerData.VMPool)
	return &vmPool
}
