package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

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
    organization: DMUUS
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
	notify.WaitFor(&controllerData.FlavorCache, 3)
	// Note for ClusterInsts and AppInsts, only those that match
	// myCloudlet Key will be sent.
	log.SpanLog(ctx, log.DebugLevelApi, "wait for instances")
	notify.WaitFor(&controllerData.ClusterInstCache, 2)
	notify.WaitFor(&controllerData.AppInstCache, 2)

	// TODO: check that the above changes triggered cloudlet cluster/app creates
	// for now just check stats
	log.SpanLog(ctx, log.DebugLevelApi, "check counts")
	require.Equal(t, 3, len(controllerData.FlavorCache.Objs))
	require.Equal(t, 2, len(controllerData.ClusterInstCache.Objs))
	require.Equal(t, 2, len(controllerData.AppInstCache.Objs))

	// delete
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

	// TODO: check that deletes triggered cloudlet cluster/app deletes.
	require.Equal(t, 0, len(controllerData.FlavorCache.Objs))
	require.Equal(t, 0, len(controllerData.ClusterInstCache.Objs))
	require.Equal(t, 0, len(controllerData.AppInstCache.Objs))

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
