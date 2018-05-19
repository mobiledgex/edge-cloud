package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestController(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelEtcd)
	// these vars are defined in main()
	mainStarted = make(chan struct{})
	enable := true
	localEtcd = &enable
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
		return
	}
	defer conn.Close()

	devApi := edgeproto.NewDeveloperApiClient(conn)
	appApi := edgeproto.NewAppApiClient(conn)
	operApi := edgeproto.NewOperatorApiClient(conn)
	cloudletApi := edgeproto.NewCloudletApiClient(conn)
	appInstApi := edgeproto.NewAppInstApiClient(conn)

	testutil.ClientDeveloperCudTest(t, devApi, testutil.DevData)
	testutil.ClientAppCudTest(t, appApi, testutil.AppData)
	testutil.ClientOperatorCudTest(t, operApi, testutil.OperatorData)
	testutil.ClientCloudletCudTest(t, cloudletApi, testutil.CloudletData)
	testutil.ClientAppInstCudTest(t, appInstApi, testutil.AppInstData)

	util.InfoLog("done")
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
