package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/stretchr/testify/assert"
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
    operatorkey:
      name: TMUS
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
      operatorkey:
        name: TMUS
      name: cloud2
  flavor:
    name: x1.tiny
  nodes: 3
  liveness: LivenessStatic
- key:
    clusterkey:
      name: 1000realities_cluster
    cloudletkey:
      operatorkey:
        name: TMUS
      name: cloud2
  flavor:
    name: x1.small
  nodes: 3
  liveness: LivenessDynamic

appinstances:
- key:
    appkey:
      developerkey:
        name: Niantic
      name: Pokemon Go
      version: 1.0.0
    cloudletkey:
      operatorkey:
        name: TMUS
      name: cloud2
    id: 99
  clusterinstkey:
    clusterkey:
      name: pokemon_cluster
    cloudletkey:
      operatorkey:
        name: TMUS
      name: cloud2
  liveness: LivenessStatic
  imagepath: some/path
  imagetype: ImageTypeDocker
  mapppedports: 10000-10010
  mappepdpath: /serverpath
  configmap: /configmap
  flavor:
    name: x1.tiny
- key:
    appkey:
      developerkey:
        name: 1000realities
      name: VRmax
      version: 1.0.0
    cloudletkey:
      operatorkey:
        name: TMUS
      name: cloud2
    id: 100
  clusterinstkey:
    clusterkey:
      name: 1000realities_cluster
    cloudletkey:
      operatorkey:
        name: TMUS
      name: cloud2
  liveness: LivenessDynamic
  imagepath: some/other/path
  imagetype: ImageTypeDocker
  mappepdpath: /serverpath
  configmap: /configmap
  flavor:
    name: x1.small
`

func startMain(t *testing.T) (chan struct{}, error) {
	mainStarted = make(chan struct{})
	mainDone := make(chan struct{})
	*platformName = "fakecloudlet"
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
	log.SetDebugLevel(log.DebugLevelNotify | log.DebugLevelMexos)

	notifyAddr := "127.0.0.1:61245"

	data := edgeproto.ApplicationData{}
	err = yaml.Unmarshal([]byte(yamlData), &data)
	require.Nil(t, err, "unmarshal yaml data")

	bytes, _ := json.Marshal(&data.Cloudlets[0].Key)

	os.Args = append(os.Args, "-cloudletKey")
	os.Args = append(os.Args, string(bytes))
	os.Args = append(os.Args, "-notifyCtlAddrs")
	os.Args = append(os.Args, notifyAddr)
	mainDone, err := startMain(t)
	if err != nil {
		close(sigChan)
		require.Nil(t, err, "start main")
		return
	}

	// CRM is driven by controller
	ctrlHandler := notify.NewDummyHandler()
	ctrlMgr := notify.ServerMgr{}
	ctrlHandler.RegisterServer(&ctrlMgr)
	ctrlMgr.Start(notifyAddr, "")

	notifyClient.WaitForConnect(1)
	stats := notify.Stats{}
	notifyClient.GetStats(&stats)
	assert.Equal(t, uint64(1), stats.Connects)

	// Add data to controller
	for ii := range data.Flavors {
		ctrlHandler.FlavorCache.Update(&data.Flavors[ii], 0)
	}
	for ii := range data.ClusterInsts {
		ctrlHandler.ClusterInstCache.Update(&data.ClusterInsts[ii], 0)
	}
	for ii := range data.AppInstances {
		ctrlHandler.AppInstCache.Update(&data.AppInstances[ii], 0)
	}
	notify.WaitFor(&controllerData.FlavorCache, 3)
	// Note for ClusterInsts and AppInsts, only those that match
	// myCloudlet Key will be sent.
	notify.WaitFor(&controllerData.ClusterInstCache, 2)
	notify.WaitFor(&controllerData.AppInstCache, 2)

	// TODO: check that the above changes triggered cloudlet cluster/app creates
	// for now just check stats
	assert.Equal(t, 3, len(controllerData.FlavorCache.Objs))
	assert.Equal(t, 2, len(controllerData.ClusterInstCache.Objs))
	assert.Equal(t, 2, len(controllerData.AppInstCache.Objs))

	// delete
	for ii := range data.AppInstances {
		ctrlHandler.AppInstCache.Delete(&data.AppInstances[ii], 0)
	}
	for ii := range data.ClusterInsts {
		ctrlHandler.ClusterInstCache.Delete(&data.ClusterInsts[ii], 0)
	}
	for ii := range data.Flavors {
		ctrlHandler.FlavorCache.Delete(&data.Flavors[ii], 0)
	}
	notify.WaitFor(&controllerData.FlavorCache, 0)
	notify.WaitFor(&controllerData.ClusterInstCache, 0)
	notify.WaitFor(&controllerData.AppInstCache, 0)

	// TODO: check that deletes triggered cloudlet cluster/app deletes.
	assert.Equal(t, 0, len(controllerData.FlavorCache.Objs))
	assert.Equal(t, 0, len(controllerData.ClusterInstCache.Objs))
	assert.Equal(t, 0, len(controllerData.AppInstCache.Objs))

	// closing the signal channel triggers main to exit
	close(sigChan)
	// wait until main is done so it can clean up properly
	<-mainDone
}
