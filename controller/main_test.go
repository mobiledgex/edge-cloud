package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

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
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelNotify | log.DebugLevelApi)

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
	crmNotify.SetClientCb(crmClient)
	go crmClient.Start()
	defer crmClient.Stop()
	dmeNotify := notify.NewDummyHandler()
	dmeClient := notify.NewDMEClient(notifyAddrs, dmeNotify)
	dmeNotify.SetClientCb(dmeClient)
	go dmeClient.Start()
	defer dmeClient.Stop()

	devApi := edgeproto.NewDeveloperApiClient(conn)
	appApi := edgeproto.NewAppApiClient(conn)
	operApi := edgeproto.NewOperatorApiClient(conn)
	cloudletApi := edgeproto.NewCloudletApiClient(conn)
	appInstApi := edgeproto.NewAppInstApiClient(conn)
	flavorApi := edgeproto.NewFlavorApiClient(conn)
	clusterFlavorApi := edgeproto.NewClusterFlavorApiClient(conn)
	clusterApi := edgeproto.NewClusterApiClient(conn)
	clusterInstApi := edgeproto.NewClusterInstApiClient(conn)
	appInstInfoClient := edgeproto.NewAppInstInfoApiClient(conn)
	cloudletInfoClient := edgeproto.NewCloudletInfoApiClient(conn)

	crmClient.WaitForConnect(1)
	dmeClient.WaitForConnect(1)

	testutil.ClientDeveloperTest(t, "cud", devApi, testutil.DevData)
	testutil.ClientFlavorTest(t, "cud", flavorApi, testutil.FlavorData)
	testutil.ClientClusterFlavorTest(t, "cud", clusterFlavorApi, testutil.ClusterFlavorData)
	testutil.ClientClusterTest(t, "cud", clusterApi, testutil.ClusterData)
	testutil.ClientAppTest(t, "cud", appApi, testutil.AppData)
	testutil.ClientOperatorTest(t, "cud", operApi, testutil.OperatorData)
	testutil.ClientCloudletTest(t, "cud", cloudletApi, testutil.CloudletData)
	insertCloudletInfo(testutil.CloudletInfoData)
	testutil.ClientClusterInstTest(t, "cud", clusterInstApi, testutil.ClusterInstData)
	testutil.ClientAppInstTest(t, "cud", appInstApi, testutil.AppInstData)

	dmeNotify.WaitForAppInsts(5)
	crmNotify.WaitForFlavors(3)
	crmNotify.WaitForClusterFlavors(3)

	assert.Equal(t, 5, len(dmeNotify.AppInstCache.Objs), "num appinsts")
	assert.Equal(t, 3, len(crmNotify.FlavorCache.Objs), "num flavors")
	assert.Equal(t, 3, len(crmNotify.ClusterFlavorCache.Objs), "num cluster flavors")

	ClientAppInstCachedFieldsTest(t, appApi, cloudletApi, appInstApi)

	// test info to persistent db
	for ii := range testutil.AppInstInfoData {
		dmeNotify.AppInstInfoCache.Update(&testutil.AppInstInfoData[ii], 0)
	}
	for ii := range testutil.CloudletInfoData {
		crmNotify.CloudletInfoCache.Update(&testutil.CloudletInfoData[ii], 0)
	}

	WaitForAppInstInfo(len(testutil.AppInstInfoData))
	WaitForCloudletInfo(len(testutil.CloudletInfoData))
	assert.Equal(t, len(testutil.AppInstInfoData), len(appInstInfoApi.cache.Objs))
	assert.Equal(t, len(testutil.CloudletInfoData), len(cloudletInfoApi.cache.Objs))
	assert.Equal(t, len(dmeNotify.AppInstInfoCache.Objs), len(appInstInfoApi.cache.Objs))
	assert.Equal(t, len(crmNotify.CloudletInfoCache.Objs), len(cloudletInfoApi.cache.Objs))

	// test show api for info structs
	// XXX These checks won't work until we move notifyId out of struct
	// and into meta data in cache (TODO)
	if false {
		CheckAppInstInfo(t, appInstInfoClient, testutil.AppInstInfoData)
		CheckCloudletInfo(t, cloudletInfoClient, testutil.CloudletInfoData)
	}

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
		_, err = appInstApi.DeleteAppInst(ctx, &inst)
		assert.Nil(t, err)
	}
	for _, inst := range testutil.ClusterInstData {
		_, err = clusterInstApi.DeleteClusterInst(ctx, &inst)
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
	// deleting appinsts/cloudlets should also delete associated info
	assert.Equal(t, 0, len(appInstInfoApi.cache.Objs))
	assert.Equal(t, 4, len(cloudletInfoApi.cache.Objs))

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
	flavorApi := edgeproto.NewFlavorApiClient(conn)
	clusterFlavorApi := edgeproto.NewClusterFlavorApiClient(conn)

	yamlData := `
operators:
- key:
    name: TMUS
cloudlets:
- key:
    operatorkey:
      name: TMUS
    name: cloud2
  ipsupport: IpSupportDynamic
  numdynamicips: 100
flavors:
- key:
    name: m1.small
  ram: 1024
  vcpus: 1
  disk: 1
clusterflavors:
- key:
    name: x1.small
  nodeflavor:
    name: m1.small
  masterflavor:
    name: m1.small
  numnodes: 2
  maxnodes: 2
  masternodes: 1
developers:
- key:
    name: AcmeAppCo
apps:
- key:
    developerkey:
      name: AcmeAppCo
    name: someApplication
    version: 1.0
  defaultflavor:
    name: m1.small
  imagetype: ImageTypeDocker
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

cloudletinfos:
- key:
    operatorkey:
      name: TMUS
    name: cloud2
  state: CloudletStateReady
  osmaxram: 65536
  osmaxvcores: 16
  osmaxvolgb: 500
  rootlbfqdn: mexlb.cloud2.tmus.mobiledgex.net
`
	data := edgeproto.ApplicationData{}
	err = yaml.Unmarshal([]byte(yamlData), &data)
	require.Nil(t, err, "unmarshal data")

	ctx := context.TODO()
	_, err = devApi.CreateDeveloper(ctx, &data.Developers[0])
	assert.Nil(t, err, "create dev")
	_, err = flavorApi.CreateFlavor(ctx, &data.Flavors[0])
	assert.Nil(t, err, "create flavor")
	_, err = clusterFlavorApi.CreateClusterFlavor(ctx, &data.ClusterFlavors[0])
	assert.Nil(t, err, "create cluster flavor")
	_, err = appApi.CreateApp(ctx, &data.Applications[0])
	assert.Nil(t, err, "create app")
	_, err = operApi.CreateOperator(ctx, &data.Operators[0])
	assert.Nil(t, err, "create operator")
	_, err = cloudletApi.CreateCloudlet(ctx, &data.Cloudlets[0])
	assert.Nil(t, err, "create cloudlet")
	insertCloudletInfo(data.CloudletInfos)

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

func WaitForAppInstInfo(count int) {
	for i := 0; i < 10; i++ {
		if len(appInstInfoApi.cache.Objs) == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func WaitForCloudletInfo(count int) {
	for i := 0; i < 10; i++ {
		if len(cloudletInfoApi.cache.Objs) == count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func CheckCloudletInfo(t *testing.T, client edgeproto.CloudletInfoApiClient, data []edgeproto.CloudletInfo) {
	api := testutil.NewClientCloudletInfoApi(client)
	ctx := context.TODO()

	show := testutil.ShowCloudletInfo{}
	show.Init()
	filterNone := edgeproto.CloudletInfo{}
	err := api.ShowCloudletInfo(ctx, &filterNone, &show)
	assert.Nil(t, err, "show cloudlet info")
	for _, obj := range data {
		show.AssertFound(t, &obj)
	}
	assert.Equal(t, len(data), len(show.Data), "show count")
}

func CheckAppInstInfo(t *testing.T, client edgeproto.AppInstInfoApiClient, data []edgeproto.AppInstInfo) {
	api := testutil.NewClientAppInstInfoApi(client)
	ctx := context.TODO()

	show := testutil.ShowAppInstInfo{}
	show.Init()
	filterNone := edgeproto.AppInstInfo{}
	err := api.ShowAppInstInfo(ctx, &filterNone, &show)
	assert.Nil(t, err, "show appinst info")
	for _, obj := range data {
		show.AssertFound(t, &obj)
	}
	assert.Equal(t, len(data), len(show.Data), "show count")
}
