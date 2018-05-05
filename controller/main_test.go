package main

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/proto"
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

	devApi := proto.NewDeveloperApiClient(conn)
	appApi := proto.NewAppApiClient(conn)
	operApi := proto.NewOperatorApiClient(conn)
	cloudletApi := proto.NewCloudletApiClient(conn)

	testutil.ClientDeveloperCudTest(t, devApi, DevData)
	testutil.ClientAppCudTest(t, appApi, AppData)
	testutil.ClientOperatorCudTest(t, operApi, OperatorData)
	testutil.ClientCloudletCudTest(t, cloudletApi, CloudletData)

	util.InfoLog("done")
	// closing the signal channel triggers main to exit
	close(sigChan)
	// wait until main is done so it can clean up properly
	<-mainDone
}
