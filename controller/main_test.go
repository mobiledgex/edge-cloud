package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func startMain(t *testing.T) (*grpc.ClientConn, chan struct{}, error) {
	// these vars are defined in main()
	mainStarted = make(chan struct{})
	// channel to wait for main to finish
	mainDone := make(chan struct{})
	go func() {
		main()
		close(mainDone)
	}()
	// wait unil main is ready
	<-mainStarted
	assert.True(t, true, "Main Started")

	// grpc client
	conn, err := grpc.Dial("127.0.0.1:55001", grpc.WithInsecure())
	assert.Nil(t, err, "grpc Dial")
	if err != nil {
		return nil, nil, err
	}
	return conn, mainDone, nil
}

func TestController(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelEtcd | util.DebugLevelNotify)

	os.Args = append(os.Args, "-localEtcd")

	conn, mainDone, err := startMain(t)
	if err != nil {
		close(sigChan)
		return
	}
	defer conn.Close()

	// test notify clients
	notifyAddrs := []string{*notifyAddr}
	crmNotify := notify.NewDummyClientHandler()
	crmClient := notify.NewCRMClient(notifyAddrs, crmNotify)
	go crmClient.Run()
	defer crmClient.Stop()
	dmeNotify := notify.NewDummyClientHandler()
	dmeClient := notify.NewDMEClient(notifyAddrs, dmeNotify)
	go dmeClient.Run()
	defer dmeClient.Stop()

	devApi := edgeproto.NewDeveloperApiClient(conn)
	appApi := edgeproto.NewAppApiClient(conn)
	operApi := edgeproto.NewOperatorApiClient(conn)
	cloudletApi := edgeproto.NewCloudletApiClient(conn)
	appInstApi := edgeproto.NewAppInstApiClient(conn)

	crmClient.WaitForConnect(1)
	dmeClient.WaitForConnect(1)

	testutil.ClientDeveloperCudTest(t, devApi, testutil.DevData)
	testutil.ClientAppCudTest(t, appApi, testutil.AppData)
	testutil.ClientOperatorCudTest(t, operApi, testutil.OperatorData)
	testutil.ClientCloudletCudTest(t, cloudletApi, testutil.CloudletData)
	testutil.ClientAppInstCudTest(t, appInstApi, testutil.AppInstData)

	dmeNotify.WaitForAppInsts(5)
	crmNotify.WaitForCloudlets(4)

	assert.Equal(t, 5, len(dmeNotify.AppInsts), "num appinsts")
	assert.Equal(t, 4, len(crmNotify.Cloudlets), "num cloudlets")

	// closing the signal channel triggers main to exit
	close(sigChan)
	// wait until main is done so it can clean up properly
	<-mainDone
}

func TestDataGen(t *testing.T) {
	out, err := os.Create("data_test.json")
	if err != nil {
		assert.Nil(t, err, "open file")
		return
	}
	for _, obj := range testutil.DevData {
		val, err := json.Marshal(&obj)
		assert.Nil(t, err, "marshal %s", obj.Key.GetKeyString())
		out.Write(val)
		out.WriteString("\n")
	}
	out.Close()
}
