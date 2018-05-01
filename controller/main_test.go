package main

import (
	"context"
	"io"
	"testing"

	"github.com/mobiledgex/edge-cloud/proto"
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

	ctx := context.Background()
	devApi := proto.NewDeveloperApiClient(conn)
	_, err = devApi.CreateDeveloper(ctx, &Dev1)
	assert.Nil(t, err, "create dev1")
	_, err = devApi.CreateDeveloper(ctx, &Dev2)
	assert.Nil(t, err, "create dev2")
	_, err = devApi.CreateDeveloper(ctx, &Dev3)
	assert.Nil(t, err, "create dev3")
	_, err = devApi.CreateDeveloper(ctx, &Dev1)
	assert.NotNil(t, err, "create dev1 again")

	devFilter := proto.Developer{}
	devMap := make(map[proto.DeveloperKey]proto.Developer)
	showDevStream, err := devApi.ShowDeveloper(ctx, &devFilter)
	assert.Nil(t, err, "show")
	if err == nil {
		for {
			showDev, err := showDevStream.Recv()
			if err == io.EOF {
				break
			}
			util.InfoLog("Show dev", "key", showDev.Key.Name)
			assert.Nil(t, err, "show")
			if err != nil {
				break
			}
			devMap[*showDev.Key] = *showDev
		}
		_, found := devMap[*Dev1.Key]
		assert.True(t, found, "Show Dev1")
		_, found = devMap[*Dev2.Key]
		assert.True(t, found, "Show Dev2")
		_, found = devMap[*Dev3.Key]
		assert.True(t, found, "Show Dev3")
		_, found = devMap[*Dev4.Key]
		assert.False(t, found, "Show Dev4)")
	}
	util.InfoLog("done")
	// closing the signal channel triggers main to exit
	close(sigChan)
	// wait until main is done so it can clean up properly
	<-mainDone
}
