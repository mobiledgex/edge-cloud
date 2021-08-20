package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"testing"
	"time"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/testutil/testservices"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	yaml "gopkg.in/yaml.v2"
)

func getGrpcClient(t *testing.T) (*grpc.ClientConn, error) {
	// grpc client
	return grpc.Dial("127.0.0.1:55001", grpc.WithInsecure())
}

func TestController(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelNotify | log.DebugLevelApi | log.DebugLevelUpgrade)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	flag.Parse() // set defaults
	*localEtcd = true
	*initLocalEtcd = true
	testinit()
	leaseTimeoutSec = 3
	syncLeaseDataRetry = 0

	// start influxd if not already running
	addr := "127.0.0.1:8086"
	_, err := exec.Command("sh", "-c", "pgrep -x influxd").Output()
	if err != nil {
		p := process.Influx{}
		p.Common.Name = "influx-test"
		p.HttpAddr = addr
		p.DataDir = "/var/tmp/.influxdb"
		// start influx
		err = p.StartLocal("/var/tmp/influxdb.log",
			process.WithCleanStartup())
		require.Nil(t, err, "start InfluxDB server")
		defer p.StopLocal()
	}

	err = startServices()
	require.Nil(t, err, "start")
	defer stopServices()

	reduceInfoTimeouts(t, ctx)

	testservices.CheckNotifySendOrder(t, notify.ServerMgrOne.GetSendOrder())

	conn, err := getGrpcClient(t)
	require.Nil(t, err, "grpc client")
	defer conn.Close()

	// test notify clients
	notifyAddrs := []string{*notifyAddr}
	crmNotify := notify.NewDummyHandler()
	crmClient := notify.NewClient("crm", notifyAddrs, nil)
	crmNotify.RegisterCRMClient(crmClient)
	NewDummyInfoResponder(&crmNotify.AppInstCache, &crmNotify.ClusterInstCache,
		&crmNotify.AppInstInfoCache, &crmNotify.ClusterInstInfoCache)
	for ii, _ := range testutil.CloudletInfoData {
		crmNotify.CloudletInfoCache.Update(ctx, &testutil.CloudletInfoData[ii], 0)
	}
	go crmClient.Start()
	defer crmClient.Stop()
	dmeNotify := notify.NewDummyHandler()
	dmeClient := notify.NewClient("dme", notifyAddrs, nil)
	dmeNotify.RegisterDMEClient(dmeClient)
	go dmeClient.Start()
	defer dmeClient.Stop()

	appClient := edgeproto.NewAppApiClient(conn)
	gpuDriverClient := edgeproto.NewGPUDriverApiClient(conn)
	cloudletClient := edgeproto.NewCloudletApiClient(conn)
	appInstClient := edgeproto.NewAppInstApiClient(conn)
	flavorClient := edgeproto.NewFlavorApiClient(conn)
	clusterInstClient := edgeproto.NewClusterInstApiClient(conn)
	cloudletInfoClient := edgeproto.NewCloudletInfoApiClient(conn)
	autoScalePolicyClient := edgeproto.NewAutoScalePolicyApiClient(conn)
	autoProvPolicyClient := edgeproto.NewAutoProvPolicyApiClient(conn)

	crmClient.WaitForConnect(1)
	dmeClient.WaitForConnect(1)
	for ii, _ := range testutil.CloudletInfoData {
		err := cloudletInfoApi.cache.WaitForState(ctx, &testutil.CloudletInfoData[ii].Key, dme.CloudletState_CLOUDLET_STATE_READY, time.Second)
		require.Nil(t, err)
	}

	testutil.ClientFlavorTest(t, "cud", flavorClient, testutil.FlavorData)
	testutil.ClientAutoProvPolicyTest(t, "cud", autoProvPolicyClient, testutil.AutoProvPolicyData)
	testutil.ClientAutoScalePolicyTest(t, "cud", autoScalePolicyClient, testutil.AutoScalePolicyData)
	testutil.ClientAppTest(t, "cud", appClient, testutil.AppData)
	testutil.ClientGPUDriverTest(t, "cud", gpuDriverClient, testutil.GPUDriverData)
	testutil.ClientCloudletTest(t, "cud", cloudletClient, testutil.CloudletData)
	testutil.ClientClusterInstTest(t, "cud", clusterInstClient, testutil.ClusterInstData)
	testutil.ClientAppInstTest(t, "cud", appInstClient, testutil.AppInstData, testutil.WithCreatedAppInstTestData(testutil.CreatedAppInstData()))

	require.Nil(t, dmeNotify.WaitForAppInsts(len(testutil.AppInstData)))
	require.Nil(t, crmNotify.WaitForFlavors(len(testutil.FlavorData)))

	require.Equal(t, len(testutil.AppInstData), len(dmeNotify.AppInstCache.Objs), "num appinsts")
	require.Equal(t, len(testutil.FlavorData), len(crmNotify.FlavorCache.Objs), "num flavors")
	require.Equal(t, len(testutil.ClusterInstData)+len(testutil.ClusterInstAutoData), len(crmNotify.ClusterInstInfoCache.Objs), "crm cluster inst infos")
	require.Equal(t, len(testutil.AppInstData), len(crmNotify.AppInstInfoCache.Objs), "crm cluster inst infos")

	ClientAppInstCachedFieldsTest(t, ctx, appClient, cloudletClient, appInstClient)

	WaitForCloudletInfo(len(testutil.CloudletInfoData))
	require.Equal(t, len(testutil.CloudletInfoData), len(cloudletInfoApi.cache.Objs))
	require.Equal(t, len(crmNotify.CloudletInfoCache.Objs), len(cloudletInfoApi.cache.Objs))

	// test show api for info structs
	// XXX These checks won't work until we move notifyId out of struct
	// and into meta data in cache (TODO)
	if false {
		CheckCloudletInfo(t, cloudletInfoClient, testutil.CloudletInfoData)
	}
	testKeepAliveRecovery(t, ctx)

	// test that delete checks disallow deletes of dependent objects
	stream, err := cloudletClient.DeleteCloudlet(ctx, &testutil.CloudletData[0])
	err = testutil.CloudletReadResultStream(stream, err)
	require.NotNil(t, err)
	_, err = appClient.DeleteApp(ctx, &testutil.AppData[0])
	require.NotNil(t, err)
	_, err = autoScalePolicyClient.DeleteAutoScalePolicy(ctx, &testutil.AutoScalePolicyData[0])
	require.NotNil(t, err)
	// test that delete works after removing dependencies
	for _, inst := range testutil.AppInstData {
		if testutil.IsAutoClusterAutoDeleteApp(&inst.Key) {
			continue
		}
		stream, err := appInstClient.DeleteAppInst(ctx, &inst)
		err = testutil.AppInstReadResultStream(stream, err)
		require.Nil(t, err)
	}
	for _, inst := range testutil.ClusterInstData {
		stream, err := clusterInstClient.DeleteClusterInst(ctx, &inst)
		err = testutil.ClusterInstReadResultStream(stream, err)
		require.Nil(t, err)
	}
	// cleanup unused reservable auto clusters
	_, err = clusterInstClient.DeleteIdleReservableClusterInsts(ctx, &edgeproto.IdleReservableClusterInsts{})
	require.Nil(t, err)
	for _, obj := range testutil.AppData {
		_, err = appClient.DeleteApp(ctx, &obj)
		require.Nil(t, err)
	}
	for _, obj := range testutil.AutoScalePolicyData {
		_, err = autoScalePolicyClient.DeleteAutoScalePolicy(ctx, &obj)
		require.Nil(t, err)
	}
	for ii, _ := range testutil.CloudletInfoData {
		obj := testutil.CloudletInfoData[ii]
		obj.State = dme.CloudletState_CLOUDLET_STATE_OFFLINE
		crmNotify.CloudletInfoCache.Update(ctx, &obj, 0)
	}
	for _, obj := range testutil.CloudletData {
		stream, err := cloudletClient.DeleteCloudlet(ctx, &obj)
		err = testutil.CloudletReadResultStream(stream, err)
		require.Nil(t, err)
	}

	// make sure dynamic app insts were deleted along with Apps
	require.Nil(t, dmeNotify.WaitForAppInsts(0))
	require.Equal(t, 0, len(dmeNotify.AppInstCache.Objs), "num appinsts")
	// deleting appinsts/cloudlets should also delete associated info
	require.Equal(t, 0, len(cloudletInfoApi.cache.Objs))
	require.Equal(t, 0, len(clusterInstApi.cache.Objs))
	require.Equal(t, 0, len(appInstApi.cache.Objs))
}

func TestDataGen(t *testing.T) {
	out, err := os.Create("data_test.json")
	if err != nil {
		require.Nil(t, err, "open file")
		return
	}
	for _, obj := range testutil.DevData {
		val, err := json.Marshal(&obj)
		require.Nil(t, err, "marshal %s", obj)
		out.Write(val)
		out.WriteString("\n")
	}
	out.Close()
}

func TestEdgeCloudBug26(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelNotify)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	flag.Parse()
	testinit()
	*localEtcd = true
	*initLocalEtcd = true

	err := startServices()
	require.Nil(t, err, "start")
	defer stopServices()

	reduceInfoTimeouts(t, ctx)

	conn, err := getGrpcClient(t)
	require.Nil(t, err, "grcp client")
	defer conn.Close()

	appClient := edgeproto.NewAppApiClient(conn)
	cloudletClient := edgeproto.NewCloudletApiClient(conn)
	appInstClient := edgeproto.NewAppInstApiClient(conn)
	flavorClient := edgeproto.NewFlavorApiClient(conn)

	yamlData := `
cloudlets:
- key:
    operatorkey:
      name: DMUUS
    name: cloud2
  ipsupport: IpSupportDynamic
  numdynamicips: 100

flavors:
- key:
    name: m1.small
  ram: 1024
  vcpus: 1
  disk: 1
apps:
- key:
    organization: AcmeAppCo
    name: someApplication
    version: 1.0
  defaultflavor:
    name: m1.small
  imagetype: ImageTypeDocker
  accessports: "tcp:80,tcp:443,udp:10002"
  ipaccess: IpAccessShared
appinstances:
- key:
    appkey:
      organization: AcmeAppCo
      name: someApplication
      version: 1.0
    cloudletkey:
      organization: DMUUS
      name: cloud2
    id: 99
  liveness: 1
  port: 8080
  ip: [10,100,10,4]
cloudletinfos:
- key:
    organization: DMUUS
    name: cloud2
  state: CloudletStateReady
  osmaxram: 65536
  osmaxvcores: 16
  osmaxvolgb: 500
  rootlbfqdn: mexlb.cloud2.dmuus.mobiledgex.net
`
	data := edgeproto.AllData{}
	err = yaml.Unmarshal([]byte(yamlData), &data)
	require.Nil(t, err, "unmarshal data")

	_, err = flavorClient.CreateFlavor(ctx, &data.Flavors[0])
	require.Nil(t, err, "create flavor")
	_, err = appClient.CreateApp(ctx, &data.Apps[0])
	require.Nil(t, err, "create app")
	_, err = cloudletClient.CreateCloudlet(ctx, &data.Cloudlets[0])
	require.Nil(t, err, "create cloudlet")
	insertCloudletInfo(ctx, data.CloudletInfos)

	show := testutil.ShowApp{}
	show.Init()
	filterNone := edgeproto.App{}
	stream, err := appClient.ShowApp(ctx, &filterNone)
	show.ReadStream(stream, err)
	require.Nil(t, err, "show data")
	require.Equal(t, 1, len(show.Data), "show app count")

	_, err = appInstClient.CreateAppInst(ctx, &data.AppInstances[0])
	require.Nil(t, err, "create app inst")

	show.Init()
	stream, err = appClient.ShowApp(ctx, &filterNone)
	show.ReadStream(stream, err)
	require.Nil(t, err, "show data")
	require.Equal(t, 1, len(show.Data), "show app count after creating app inst")
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
	require.Nil(t, err, "show cloudlet info")
	for _, obj := range data {
		show.AssertFound(t, &obj)
	}
	require.Equal(t, len(data), len(show.Data), "show count")
}

func testKeepAliveRecovery(t *testing.T, ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelInfo, "testKeepAliveRecovery")
	// already some alerts from non-crm sources
	numPrevAlerts := len(alertApi.cache.Objs)
	require.Equal(t, 0, len(alertApi.sourceCache.Objs))

	// add some alerts from crm, will go into source cache
	for _, alert := range testutil.AlertData {
		alertApi.Update(ctx, &alert, 0)
	}
	numCrmAlerts := len(testutil.AlertData)
	totalAlerts := numPrevAlerts + numCrmAlerts
	WaitForAlerts(t, totalAlerts)
	require.Equal(t, numCrmAlerts, len(alertApi.sourceCache.Objs))
	leaseID := ControllerAliveLease()

	log.SpanLog(ctx, log.DebugLevelInfo, "grab sync lease data lock")
	// grab syncLeaseData lock to prevent it restoring any
	// source data, so that we can verify data was flushed from etcd
	syncLeaseData.mux.Lock()

	log.SpanLog(ctx, log.DebugLevelInfo, "revoke lease")
	// revoke lease to simulate timed out keepalives
	err := syncLeaseData.sync.store.Revoke(ctx, syncLeaseData.leaseID)
	require.Nil(t, err)
	// wait for lease timeout (so etcd KeepAlive function returns)
	time.Sleep(time.Duration(leaseTimeoutSec) * time.Second)

	// data should have been flushed from etcd
	WaitForAlerts(t, numPrevAlerts)
	// data should still be in source cache
	require.Equal(t, numCrmAlerts, len(alertApi.sourceCache.Objs))

	log.SpanLog(ctx, log.DebugLevelInfo, "release sync lease data lock")
	// release sync lock
	syncLeaseData.mux.Unlock()
	// alerts should be re-sync'd
	WaitForAlerts(t, totalAlerts)
	// make sure there is a new lease ID
	require.NotEqual(t, leaseID, syncLeaseData.leaseID)

	log.SpanLog(ctx, log.DebugLevelInfo, "delete alerts")
	// delete alerts
	for _, alert := range testutil.AlertData {
		alertApi.Delete(ctx, &alert, 0)
	}
	WaitForAlerts(t, numPrevAlerts)
	require.Equal(t, 0, len(alertApi.sourceCache.Objs))
}

func WaitForAlerts(t *testing.T, count int) {
	for i := 0; i < 10; i++ {
		if len(alertApi.cache.Objs) == count {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	require.Equal(t, count, len(alertApi.cache.Objs), "timed out waiting for alerts")
}
