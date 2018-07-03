package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	yaml "gopkg.in/yaml.v2"
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
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelNotify)

	os.Args = append(os.Args, "-localEtcd")

	conn, mainDone, err := startMain(t)
	if err != nil {
		close(sigChan)
		return
	}
	defer conn.Close()

	// test notify clients
	notifyAddrs := []string{*notifyAddr}
	crmNotify := notify.NewDummyHandler()
	crmClient := notify.NewCRMClient(notifyAddrs, crmNotify)
	go crmClient.Start()
	defer crmClient.Stop()
	dmeNotify := notify.NewDummyHandler()
	dmeClient := notify.NewDMEClient(notifyAddrs, dmeNotify)
	go dmeClient.Start()
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

	assert.Equal(t, 5, len(dmeNotify.AppInstCache.Objs), "num appinsts")
	assert.Equal(t, 4, len(crmNotify.CloudletCache.Objs), "num cloudlets")

	ClientAppInstCachedFieldsTest(t, appApi, cloudletApi, appInstApi)

	// test that delete checks disallow deletes of dependent objects
	ctx := context.TODO()
	_, err = devApi.DeleteDeveloper(ctx, &testutil.DevData[0])
	assert.NotNil(t, err)
	_, err = operApi.DeleteOperator(ctx, &testutil.OperatorData[0])
	assert.NotNil(t, err)
	_, err = cloudletApi.DeleteCloudlet(ctx, &testutil.CloudletData[0])
	assert.NotNil(t, err)
	_, err = appApi.DeleteApp(ctx, &testutil.AppData[0])
	assert.NotNil(t, err)
	// test that delete works after removing dependencies
	for _, inst := range testutil.AppInstData {
		if inst.Liveness == edgeproto.AppInst_DYNAMIC {
			// skip dynamic, they are not counted as dependencies
			continue
		}
		_, err = appInstApi.DeleteAppInst(ctx, &inst)
		assert.Nil(t, err)
	}
	for _, obj := range testutil.AppData {
		_, err = appApi.DeleteApp(ctx, &obj)
		assert.Nil(t, err)
	}
	for _, obj := range testutil.CloudletData {
		_, err = cloudletApi.DeleteCloudlet(ctx, &obj)
		assert.Nil(t, err)
	}
	for _, obj := range testutil.DevData {
		_, err = devApi.DeleteDeveloper(ctx, &obj)
		assert.Nil(t, err)
	}
	for _, obj := range testutil.OperatorData {
		_, err = operApi.DeleteOperator(ctx, &obj)
		assert.Nil(t, err)
	}
	// make sure dynamic app insts were deleted along with Apps
	dmeNotify.WaitForAppInsts(0)
	assert.Equal(t, 0, len(dmeNotify.AppInstCache.Objs), "num appinsts")

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

func TestEdgeCloudBug26(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelNotify)

	os.Args = append(os.Args, "-localEtcd")

	conn, mainDone, err := startMain(t)
	if err != nil {
		close(sigChan)
		return
	}
	defer conn.Close()

	devApi := edgeproto.NewDeveloperApiClient(conn)
	appApi := edgeproto.NewAppApiClient(conn)
	operApi := edgeproto.NewOperatorApiClient(conn)
	cloudletApi := edgeproto.NewCloudletApiClient(conn)
	appInstApi := edgeproto.NewAppInstApiClient(conn)

	yamlData := `
operators:
- key:
    name: TMUS
cloudlets:
- key:
    operatorkey:
      name: TMUS
    name: cloud2
developers:
- key:
    name: AcmeAppCo
apps:
- key:
    developerkey:
      name: AcmeAppCo
    name: someApplication
    version: 1.0
appinstances:
- key:
    appkey:
      developerkey:
        name: AcmeAppCo
      name: someApplication
      version: 1.0
    cloudletkey:
      operatorkey:
        name: TMUS
      name: cloud2
    id: 99
  liveness: 1
  port: 8080
  ip: [10,100,10,4]
`
	data := edgeproto.ApplicationData{}
	err = yaml.Unmarshal([]byte(yamlData), &data)
	require.Nil(t, err, "unmarshal data")

	ctx := context.TODO()
	_, err = devApi.CreateDeveloper(ctx, &data.Developers[0])
	assert.Nil(t, err, "create dev")
	_, err = appApi.CreateApp(ctx, &data.Applications[0])
	assert.Nil(t, err, "create app")
	_, err = operApi.CreateOperator(ctx, &data.Operators[0])
	assert.Nil(t, err, "create operator")
	_, err = cloudletApi.CreateCloudlet(ctx, &data.Cloudlets[0])
	assert.Nil(t, err, "create cloudlet")

	show := testutil.ShowApp{}
	show.Init()
	filterNone := edgeproto.App{}
	stream, err := appApi.ShowApp(ctx, &filterNone)
	show.ReadStream(stream, err)
	assert.Nil(t, err, "show data")
	assert.Equal(t, 1, len(show.Data), "show app count")

	_, err = appInstApi.CreateAppInst(ctx, &data.AppInstances[0])
	assert.Nil(t, err, "create app inst")

	show.Init()
	stream, err = appApi.ShowApp(ctx, &filterNone)
	show.ReadStream(stream, err)
	assert.Nil(t, err, "show data")
	assert.Equal(t, 1, len(show.Data), "show app count after creating app inst")

	// closing the signal channel triggers main to exit
	close(sigChan)
	// wait until main is done so it can clean up properly
	<-mainDone
}
