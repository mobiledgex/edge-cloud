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
	reduceInfoTimeouts()
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
	crmClient := notify.NewClient(notifyAddrs, "")
	crmNotify.RegisterCRMClient(crmClient)
	NewDummyInfoResponder(&crmNotify.AppInstCache, &crmNotify.ClusterInstCache,
		&crmNotify.AppInstInfoCache, &crmNotify.ClusterInstInfoCache)
	for ii, _ := range testutil.CloudletInfoData {
		crmNotify.CloudletInfoCache.Update(&testutil.CloudletInfoData[ii], 0)
	}
	go crmClient.Start()
	defer crmClient.Stop()
	dmeNotify := notify.NewDummyHandler()
	dmeClient := notify.NewClient(notifyAddrs, "")
	dmeNotify.RegisterDMEClient(dmeClient)
	go dmeClient.Start()
	defer dmeClient.Stop()

	devClient := edgeproto.NewDeveloperApiClient(conn)
	appClient := edgeproto.NewAppApiClient(conn)
	operClient := edgeproto.NewOperatorApiClient(conn)
	cloudletClient := edgeproto.NewCloudletApiClient(conn)
	appInstClient := edgeproto.NewAppInstApiClient(conn)
	flavorClient := edgeproto.NewFlavorApiClient(conn)
	clusterClient := edgeproto.NewClusterApiClient(conn)
	clusterInstClient := edgeproto.NewClusterInstApiClient(conn)
	cloudletInfoClient := edgeproto.NewCloudletInfoApiClient(conn)

	crmClient.WaitForConnect(1)
	dmeClient.WaitForConnect(1)

	testutil.ClientDeveloperTest(t, "cud", devClient, testutil.DevData)
	testutil.ClientFlavorTest(t, "cud", flavorClient, testutil.FlavorData)
	testutil.ClientClusterTest(t, "cud", clusterClient, testutil.ClusterData)
	testutil.ClientAppTest(t, "cud", appClient, testutil.AppData)
	testutil.ClientOperatorTest(t, "cud", operClient, testutil.OperatorData)
	testutil.ClientCloudletTest(t, "cud", cloudletClient, testutil.CloudletData)
	testutil.ClientClusterInstTest(t, "cud", clusterInstClient, testutil.ClusterInstData)
	testutil.ClientAppInstTest(t, "cud", appInstClient, testutil.AppInstData)

	dmeNotify.WaitForAppInsts(6)
	crmNotify.WaitForFlavors(3)

	assert.Equal(t, 6, len(dmeNotify.AppInstCache.Objs), "num appinsts")
	assert.Equal(t, 4, len(crmNotify.FlavorCache.Objs), "num flavors")
	assert.Equal(t, 9, len(crmNotify.ClusterInstInfoCache.Objs), "crm cluster inst infos")
	assert.Equal(t, 6, len(crmNotify.AppInstInfoCache.Objs), "crm cluster inst infos")

	ClientAppInstCachedFieldsTest(t, appClient, cloudletClient, appInstClient)

	WaitForCloudletInfo(len(testutil.CloudletInfoData))
	assert.Equal(t, len(testutil.CloudletInfoData), len(cloudletInfoApi.cache.Objs))
	assert.Equal(t, len(crmNotify.CloudletInfoCache.Objs), len(cloudletInfoApi.cache.Objs))

	// test show api for info structs
	// XXX These checks won't work until we move notifyId out of struct
	// and into meta data in cache (TODO)
	if false {
		CheckCloudletInfo(t, cloudletInfoClient, testutil.CloudletInfoData)
	}

	// test that delete checks disallow deletes of dependent objects
	ctx := context.TODO()
	_, err = devClient.DeleteDeveloper(ctx, &testutil.DevData[0])
	assert.NotNil(t, err)
	_, err = operClient.DeleteOperator(ctx, &testutil.OperatorData[0])
	assert.NotNil(t, err)
	stream, err := cloudletClient.DeleteCloudlet(ctx, &testutil.CloudletData[0])
	err = testutil.CloudletReadResultStream(stream, err)
	assert.NotNil(t, err)
	_, err = appClient.DeleteApp(ctx, &testutil.AppData[0])
	assert.NotNil(t, err)
	// test that delete works after removing dependencies
	for _, inst := range testutil.AppInstData {
		stream, err := appInstClient.DeleteAppInst(ctx, &inst)
		err = testutil.AppInstReadResultStream(stream, err)
		assert.Nil(t, err)
	}
	for _, inst := range testutil.ClusterInstData {
		stream, err := clusterInstClient.DeleteClusterInst(ctx, &inst)
		err = testutil.ClusterInstReadResultStream(stream, err)
		assert.Nil(t, err)
	}
	for _, obj := range testutil.AppData {
		_, err = appClient.DeleteApp(ctx, &obj)
		assert.Nil(t, err)
	}
	for _, obj := range testutil.CloudletData {
		_, err = cloudletClient.DeleteCloudlet(ctx, &obj)
		assert.Nil(t, err)
	}
	for _, obj := range testutil.DevData {
		_, err = devClient.DeleteDeveloper(ctx, &obj)
		assert.Nil(t, err)
	}
	for _, obj := range testutil.OperatorData {
		_, err = operClient.DeleteOperator(ctx, &obj)
		assert.Nil(t, err)
	}
	// make sure dynamic app insts were deleted along with Apps
	dmeNotify.WaitForAppInsts(0)
	assert.Equal(t, 0, len(dmeNotify.AppInstCache.Objs), "num appinsts")
	// deleting appinsts/cloudlets should also delete associated info
	assert.Equal(t, 4, len(cloudletInfoApi.cache.Objs))
	assert.Equal(t, 0, len(clusterInstApi.cache.Objs))
	assert.Equal(t, 0, len(appInstApi.cache.Objs))

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

	devClient := edgeproto.NewDeveloperApiClient(conn)
	appClient := edgeproto.NewAppApiClient(conn)
	operClient := edgeproto.NewOperatorApiClient(conn)
	cloudletClient := edgeproto.NewCloudletApiClient(conn)
	appInstClient := edgeproto.NewAppInstApiClient(conn)
	flavorClient := edgeproto.NewFlavorApiClient(conn)

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
  accessports: "tcp:80,http:443,udp:10002"
  ipaccess: IpAccessShared
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
	_, err = devClient.CreateDeveloper(ctx, &data.Developers[0])
	assert.Nil(t, err, "create dev")
	_, err = flavorClient.CreateFlavor(ctx, &data.Flavors[0])
	assert.Nil(t, err, "create flavor")
	_, err = appClient.CreateApp(ctx, &data.Applications[0])
	assert.Nil(t, err, "create app")
	_, err = operClient.CreateOperator(ctx, &data.Operators[0])
	assert.Nil(t, err, "create operator")
	_, err = cloudletClient.CreateCloudlet(ctx, &data.Cloudlets[0])
	assert.Nil(t, err, "create cloudlet")
	insertCloudletInfo(data.CloudletInfos)

	show := testutil.ShowApp{}
	show.Init()
	filterNone := edgeproto.App{}
	stream, err := appClient.ShowApp(ctx, &filterNone)
	show.ReadStream(stream, err)
	assert.Nil(t, err, "show data")
	assert.Equal(t, 1, len(show.Data), "show app count")

	_, err = appInstClient.CreateAppInst(ctx, &data.AppInstances[0])
	assert.Nil(t, err, "create app inst")

	show.Init()
	stream, err = appClient.ShowApp(ctx, &filterNone)
	show.ReadStream(stream, err)
	assert.Nil(t, err, "show data")
	assert.Equal(t, 1, len(show.Data), "show app count after creating app inst")

	// closing the signal channel triggers main to exit
	close(sigChan)
	// wait until main is done so it can clean up properly
	<-mainDone
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
