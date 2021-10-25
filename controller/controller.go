// Main process

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	baselog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	influxq "github.com/mobiledgex/edge-cloud/controller/influxq_client"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/mobiledgex/edge-cloud/vmspec"
	yaml "github.com/mobiledgex/yaml/v2"
	"google.golang.org/grpc"
)

// Command line options
var rootDir = flag.String("r", "", "root directory; set for testing")
var localEtcd = flag.Bool("localEtcd", false, "set to start local etcd for testing")
var initLocalEtcd = flag.Bool("initLocalEtcd", false, "set to init local etcd database")
var region = flag.String("region", "local", "region name")
var etcdUrls = flag.String("etcdUrls", "http://127.0.0.1:2380", "etcd client listener URLs")
var apiAddr = flag.String("apiAddr", "127.0.0.1:55001", "API listener address")

// external API Addr is registered with etcd so other controllers can connect
// directly to this controller.
var externalApiAddr = flag.String("externalApiAddr", "", "External API listener address if behind proxy/LB. Defaults to apiAddr")
var httpAddr = flag.String("httpAddr", "127.0.0.1:8091", "HTTP listener address")
var notifyAddr = flag.String("notifyAddr", "127.0.0.1:50001", "Notify listener address")
var notifyRootAddrs = flag.String("notifyRootAddrs", "", "Comma separated list of notifyroots")
var notifyParentAddrs = flag.String("notifyParentAddrs", "", "Comma separated list of notify parents")
var accessApiAddr = flag.String("accessApiAddr", "127.0.0.1:41001", "listener address for external services with access key")
var publicAddr = flag.String("publicAddr", "127.0.0.1", "Public facing address/hostname of controller")
var edgeTurnAddr = flag.String("edgeTurnAddr", "127.0.0.1:6080", "Address to EdgeTurn Server")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var shortTimeouts = flag.Bool("shortTimeouts", false, "set timeouts short for simulated cloudlet testing")
var influxAddr = flag.String("influxAddr", "http://127.0.0.1:8086", "InfluxDB listener address")
var registryFQDN = flag.String("registryFQDN", "", "default docker image registry FQDN")
var artifactoryFQDN = flag.String("artifactoryFQDN", "", "default VM image registry (artifactory) FQDN")
var cloudletRegistryPath = flag.String("cloudletRegistryPath", "", "edge-cloud image registry path for deploying cloudlet services")
var cloudletVMImagePath = flag.String("cloudletVMImagePath", "", "VM image for deploying cloudlet services")
var chefServerPath = flag.String("chefServerPath", "", "Path to chef server organization")
var versionTag = flag.String("versionTag", "", "edge-cloud image tag indicating controller version")
var skipVersionCheck = flag.Bool("skipVersionCheck", false, "Skip etcd version hash verification")
var autoUpgrade = flag.Bool("autoUpgrade", false, "Automatically upgrade etcd database to the current version")
var testMode = flag.Bool("testMode", false, "Run controller in test mode")
var commercialCerts = flag.Bool("commercialCerts", false, "Have CRM grab certs from LetsEncrypt. If false then CRM will generate its onwn self-signed cert")
var checkpointInterval = flag.String("checkpointInterval", "MONTH", "Interval at which to checkpoint cluster usage")
var appDNSRoot = flag.String("appDNSRoot", "mobiledgex.net", "App domain name root")
var requireNotifyAccessKey = flag.Bool("requireNotifyAccessKey", false, "Require AccessKey authentication on notify API")

var ControllerId = ""
var InfluxDBName = cloudcommon.DeveloperMetricsDbName

func GetRootDir() string {
	return *rootDir
}

var ErrCtrlAlreadyInProgress = errors.New("Change already in progress")
var ErrCtrlUpgradeRequired = errors.New("data mode upgrade required")

var sigChan chan os.Signal
var services Services
var vaultConfig *vault.Config
var nodeMgr node.NodeMgr

type Services struct {
	etcdLocal                 *process.Etcd
	sync                      *Sync
	influxQ                   *influxq.InfluxQ
	events                    *influxq.InfluxQ
	edgeEventsInfluxQ         *influxq.InfluxQ
	cloudletResourcesInfluxQ  *influxq.InfluxQ
	downsampledMetricsInfluxQ *influxq.InfluxQ
	notifyServerMgr           bool
	grpcServer                *grpc.Server
	httpServer                *http.Server
	notifyClient              *notify.Client
	accessKeyGrpcServer       node.AccessKeyGrpcServer
	listeners                 []net.Listener
	publicCertManager         *node.PublicCertManager
	stopInitCC                chan bool
}

func main() {
	nodeMgr.InitFlags()
	flag.Parse()

	services.listeners = make([]net.Listener, 0)
	err := startServices()
	if err != nil {
		stopServices()
		log.FatalLog(err.Error())
	}
	defer stopServices()

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}

func validateFields(ctx context.Context) error {
	if *cloudletRegistryPath != "" {
		if *versionTag == "" {
			return fmt.Errorf("Version tag is required")
		}
		parts := strings.Split(*cloudletRegistryPath, "/")
		if len(parts) < 2 || !strings.Contains(parts[0], ".") {
			return fmt.Errorf("Cloudlet registry path should be full registry URL: <domain-name>/<registry-path>")
		}
		urlObj, err := util.ImagePathParse(*cloudletRegistryPath)
		if err != nil {
			return fmt.Errorf("Invalid cloudlet registry path: %v", err)
		}
		out := strings.Split(urlObj.Path, ":")
		if len(out) == 2 {
			return fmt.Errorf("Cloudlet registry path should not have image tag")
		} else if len(out) != 1 {
			return fmt.Errorf("Invalid registry path")
		}
		platform_registry_path := *cloudletRegistryPath + ":" + strings.TrimSpace(string(*versionTag))
		authApi := &cloudcommon.VaultRegistryAuthApi{
			VaultConfig: vaultConfig,
		}
		err = cloudcommon.ValidateDockerRegistryPath(ctx, platform_registry_path, authApi)
		if err != nil {
			return err
		}
	}
	return nil
}

func startServices() error {
	var err error

	log.SetDebugLevelStrs(*debugLevels)

	if *externalApiAddr == "" {
		*externalApiAddr, err = util.GetExternalApiAddr(*apiAddr)
		if err != nil {
			return err
		}
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "nohostname"
	}
	ControllerId = hostname + "@" + *externalApiAddr

	// region number for etcd is a deprecated concept since we decided
	// etcd is per-region.
	objstore.InitRegion(uint32(1))
	if !util.ValidRegion(*region) {
		return fmt.Errorf("invalid region name")
	}

	ctx, span, err := nodeMgr.Init(node.NodeTypeController, node.CertIssuerRegional, node.WithName(ControllerId), node.WithContainerVersion(*versionTag), node.WithRegion(*region))
	if err != nil {
		return err
	}
	initDebug(ctx, &nodeMgr)
	defer span.Finish()
	vaultConfig = nodeMgr.VaultConfig

	log.SpanLog(ctx, log.DebugLevelInfo, "Start up", "rootDir", *rootDir, "apiAddr", *apiAddr, "externalApiAddr", *externalApiAddr)
	err = validateFields(ctx)
	if err != nil {
		return err
	}

	if *localEtcd {
		opts := []process.StartOp{}
		if *initLocalEtcd {
			opts = append(opts, process.WithCleanStartup())
		}
		etcdLocal, err := StartLocalEtcdServer(opts...)
		if err != nil {
			return fmt.Errorf("starting local etcd server failed: %v", err)
		}
		services.etcdLocal = etcdLocal
		etcdUrls = &etcdLocal.ClientAddrs
	}
	objStore, err := GetEtcdClientBasic(*etcdUrls)
	if err != nil {
		return fmt.Errorf("Failed to initialize Object Store, %v", err)
	}
	err = objStore.CheckConnected(50, 20*time.Millisecond)
	if err != nil {
		return fmt.Errorf("Failed to connect to etcd servers, %v", err)
	}

	// We might need to upgrade the stored objects
	if !*skipVersionCheck {
		// First off - check version of the objectStore we are running
		version, err := checkVersion(ctx, objStore)
		if err != nil && strings.Contains(err.Error(), ErrCtrlUpgradeRequired.Error()) && *autoUpgrade {
			err = edgeproto.UpgradeToLatest(version, objStore)
			if err != nil {
				return fmt.Errorf("Failed to ugprade data model: %v", err)
			}
		} else if err != nil {
			return fmt.Errorf("Running version doesn't match the version of etcd, %v", err)
		}
	}
	lis, err := net.Listen("tcp", *apiAddr)
	if err != nil {
		return fmt.Errorf("Failed to listen on address %s, %v", *apiAddr, err)
	}
	services.listeners = append(services.listeners, lis)

	sync := InitSync(objStore)
	InitApis(sync)
	sync.Start()
	services.sync = sync
	// requireNotifyAccessKey allows for backwards compatibility when
	// set to false, because it allows CRMs to connect to notify without
	// an access key (as long as pki internal cert is verified).
	cloudletApi.accessKeyServer.SetRequireTlsAccessKey(*requireNotifyAccessKey)

	InitSyncLeaseData(sync)
	syncLeaseData.Start(ctx)

	err = settingsApi.initDefaults(ctx)
	if err != nil {
		return fmt.Errorf("Failed to init settings, %v", err)
	}
	// cleanup thread must start after settings are loaded
	go clusterInstApi.cleanupThread()

	err = rateLimitSettingsApi.initDefaultRateLimitSettings(ctx)
	if err != nil {
		return fmt.Errorf("Failed to init default rate limit settings, %v", err)
	}

	// get influxDB credentials from vault
	influxAuth := &cloudcommon.InfluxCreds{}
	influxAuth, err = cloudcommon.GetInfluxDataAuth(vaultConfig, *region)
	// Default to empty credentials if in test mode
	if *testMode && err != nil {
		influxAuth = &cloudcommon.InfluxCreds{}
	} else if err != nil {
		return fmt.Errorf("Failed to get influxDB auth, %v", err)
	}

	// downsampled metrics influx
	downsampledMetricsInfluxQ := influxq.NewInfluxQ(cloudcommon.DownsampledMetricsDbName, influxAuth.User, influxAuth.Pass)
	downsampledMetricsInfluxQ.InitRetentionPolicy(settingsApi.Get().InfluxDbDownsampledMetricsRetention.TimeDuration())
	err = downsampledMetricsInfluxQ.Start(*influxAddr)
	if err != nil {
		return fmt.Errorf("Failed to start influx queue address %s, %v",
			*influxAddr, err)
	}
	services.downsampledMetricsInfluxQ = downsampledMetricsInfluxQ

	// metrics influx
	influxQ := influxq.NewInfluxQ(InfluxDBName, influxAuth.User, influxAuth.Pass)
	influxQ.InitRetentionPolicy(settingsApi.Get().InfluxDbMetricsRetention.TimeDuration())
	err = influxQ.Start(*influxAddr)
	if err != nil {
		return fmt.Errorf("Failed to start influx queue address %s, %v",
			*influxAddr, err)
	}
	services.influxQ = influxQ

	// events influx
	events := influxq.NewInfluxQ(cloudcommon.EventsDbName, influxAuth.User, influxAuth.Pass)
	err = events.Start(*influxAddr)
	if err != nil {
		return fmt.Errorf("Failed to start influx queue address %s, %v",
			*influxAddr, err)
	}
	services.events = events

	// persistent stats influx
	edgeEventsInfluxQ := influxq.NewInfluxQ(cloudcommon.EdgeEventsMetricsDbName, influxAuth.User, influxAuth.Pass)
	edgeEventsInfluxQ.InitRetentionPolicy(settingsApi.Get().InfluxDbEdgeEventsMetricsRetention.TimeDuration())
	err = edgeEventsInfluxQ.Start(*influxAddr)
	if err != nil {
		return fmt.Errorf("Failed to start influx queue address %s, %v",
			*influxAddr, err)
	}
	services.edgeEventsInfluxQ = edgeEventsInfluxQ

	// cloudlet resources influx
	cloudletResourcesInfluxQ := influxq.NewInfluxQ(cloudcommon.CloudletResourceUsageDbName, influxAuth.User, influxAuth.Pass)
	cloudletResourcesInfluxQ.InitRetentionPolicy(settingsApi.Get().InfluxDbCloudletUsageMetricsRetention.TimeDuration())
	err = cloudletResourcesInfluxQ.Start(*influxAddr)
	if err != nil {
		return fmt.Errorf("Failed to start influx queue address %s, %v",
			*influxAddr, err)
	}
	services.cloudletResourcesInfluxQ = cloudletResourcesInfluxQ

	// create continuous queries for edgeevents metrics
	services.stopInitCC = make(chan bool)
	go initContinuousQueries()

	InitNotify(influxQ, edgeEventsInfluxQ, &appInstClientApi)
	if *notifyParentAddrs != "" {
		addrs := strings.Split(*notifyParentAddrs, ",")
		tlsConfig, err := nodeMgr.InternalPki.GetClientTlsConfig(ctx,
			nodeMgr.CommonName(),
			node.CertIssuerRegional,
			[]node.MatchCA{node.GlobalMatchCA()})
		if err != nil {
			return err
		}
		dialOption := tls.GetGrpcDialOption(tlsConfig)
		notifyClient := notify.NewClient(nodeMgr.Name(), addrs, dialOption)
		notifyClient.RegisterSendAlertCache(&alertApi.cache)
		nodeMgr.RegisterClient(notifyClient)
		notifyClient.Start()
		services.notifyClient = notifyClient
	}
	notifyServerTls, err := nodeMgr.InternalPki.GetServerTlsConfig(ctx,
		nodeMgr.CommonName(),
		node.CertIssuerRegional,
		[]node.MatchCA{
			node.SameRegionalMatchCA(),
			node.SameRegionalCloudletMatchCA(),
			node.GlobalMatchCA(),
		})
	if err != nil {
		return err
	}
	notifyUnaryInterceptor := grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(
			cloudcommon.AuditUnaryInterceptor,
			cloudletApi.accessKeyServer.UnaryTlsAccessKey,
		))
	notifyStreamInterceptor := grpc.StreamInterceptor(
		grpc_middleware.ChainStreamServer(
			cloudcommon.AuditStreamInterceptor,
			cloudletApi.accessKeyServer.StreamTlsAccessKey,
		))
	notify.ServerMgrOne.Start(nodeMgr.Name(), *notifyAddr, notifyServerTls,
		notify.ServerUnaryInterceptor(notifyUnaryInterceptor),
		notify.ServerStreamInterceptor(notifyStreamInterceptor),
	)
	services.notifyServerMgr = true

	// VaultPublicCertClient implements GetPublicCertApi
	// Allows controller to get public certs from vault
	var getPublicCertApi cloudcommon.GetPublicCertApi
	if tls.IsTestTls() || *testMode {
		getPublicCertApi = &cloudcommon.TestPublicCertApi{}
	} else if nodeMgr.InternalPki.UseVaultPki {
		getPublicCertApi = &cloudcommon.VaultPublicCertApi{
			VaultConfig: vaultConfig,
		}
	}
	publicCertManager, err := node.NewPublicCertManager(*publicAddr, getPublicCertApi, "", "")
	if err != nil {
		span.Finish()
		log.FatalLog("unable to get public cert manager", "err", err)
	}
	services.publicCertManager = publicCertManager
	accessServerTlsConfig, err := services.publicCertManager.GetServerTlsConfig(ctx)
	if err != nil {
		return err
	}
	services.publicCertManager.StartRefresh()
	// Start access server
	err = services.accessKeyGrpcServer.Start(*accessApiAddr, cloudletApi.accessKeyServer, accessServerTlsConfig, func(accessServer *grpc.Server) {
		edgeproto.RegisterCloudletAccessApiServer(accessServer, &cloudletApi)
		edgeproto.RegisterCloudletAccessKeyApiServer(accessServer, &cloudletApi)
	})
	if err != nil {
		return err
	}

	// External API (for clients or MC).
	apiTlsConfig, err := nodeMgr.InternalPki.GetServerTlsConfig(ctx,
		nodeMgr.CommonName(),
		node.CertIssuerRegional,
		[]node.MatchCA{
			node.GlobalMatchCA(),
			node.SameRegionalMatchCA(),
		})
	if err != nil {
		return err
	}

	server := grpc.NewServer(cloudcommon.GrpcCreds(apiTlsConfig),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(cloudcommon.AuditUnaryInterceptor)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(cloudcommon.AuditStreamInterceptor)))
	edgeproto.RegisterAppApiServer(server, &appApi)
	edgeproto.RegisterResTagTableApiServer(server, &resTagTableApi)
	edgeproto.RegisterOperatorCodeApiServer(server, &operatorCodeApi)
	edgeproto.RegisterFlavorApiServer(server, &flavorApi)
	edgeproto.RegisterClusterInstApiServer(server, &clusterInstApi)
	edgeproto.RegisterCloudletApiServer(server, &cloudletApi)
	edgeproto.RegisterAppInstApiServer(server, &appInstApi)
	edgeproto.RegisterCloudletInfoApiServer(server, &cloudletInfoApi)
	edgeproto.RegisterVMPoolApiServer(server, &vmPoolApi)
	edgeproto.RegisterCloudletRefsApiServer(server, &cloudletRefsApi)
	edgeproto.RegisterAppInstRefsApiServer(server, &appInstRefsApi)
	edgeproto.RegisterStreamObjApiServer(server, &streamObjApi)
	edgeproto.RegisterControllerApiServer(server, &controllerApi)
	edgeproto.RegisterNodeApiServer(server, &nodeApi)
	edgeproto.RegisterExecApiServer(server, &execApi)
	edgeproto.RegisterCloudletPoolApiServer(server, &cloudletPoolApi)
	edgeproto.RegisterAlertApiServer(server, &alertApi)
	edgeproto.RegisterAutoScalePolicyApiServer(server, &autoScalePolicyApi)
	edgeproto.RegisterAutoProvPolicyApiServer(server, &autoProvPolicyApi)
	edgeproto.RegisterTrustPolicyApiServer(server, &trustPolicyApi)
	edgeproto.RegisterTrustPolicyExceptionApiServer(server, &trustPolicyExceptionApi)
	edgeproto.RegisterSettingsApiServer(server, &settingsApi)
	edgeproto.RegisterRateLimitSettingsApiServer(server, &rateLimitSettingsApi)
	edgeproto.RegisterAppInstClientApiServer(server, &appInstClientApi)
	edgeproto.RegisterDebugApiServer(server, &debugApi)
	edgeproto.RegisterDeviceApiServer(server, &deviceApi)
	edgeproto.RegisterOrganizationApiServer(server, &organizationApi)
	edgeproto.RegisterAppInstLatencyApiServer(server, &appInstLatencyApi)
	edgeproto.RegisterGPUDriverApiServer(server, &gpuDriverApi)
	edgeproto.RegisterAlertPolicyApiServer(server, &userAlertApi)
	edgeproto.RegisterNetworkApiServer(server, &networkApi)

	go func() {
		// Serve will block until interrupted and Stop is called
		if err := server.Serve(lis); err != nil {
			log.FatalLog("Failed to serve", "error", err)
		}
	}()
	services.grpcServer = server

	// REST gateway
	mux := http.NewServeMux()
	gwcfg := &cloudcommon.GrpcGWConfig{
		ApiAddr: *apiAddr,
		ApiHandles: []func(context.Context, *gwruntime.ServeMux, *grpc.ClientConn) error{
			edgeproto.RegisterAppApiHandler,
			edgeproto.RegisterAppInstApiHandler,
			edgeproto.RegisterOperatorCodeApiHandler,
			edgeproto.RegisterCloudletApiHandler,
			edgeproto.RegisterCloudletInfoApiHandler,
			edgeproto.RegisterVMPoolApiHandler,
			edgeproto.RegisterGPUDriverApiHandler,
			edgeproto.RegisterFlavorApiHandler,
			edgeproto.RegisterClusterInstApiHandler,
			edgeproto.RegisterControllerApiHandler,
			edgeproto.RegisterNodeApiHandler,
			edgeproto.RegisterCloudletPoolApiHandler,
			edgeproto.RegisterAlertApiHandler,
			edgeproto.RegisterAutoScalePolicyApiHandler,
			edgeproto.RegisterAutoProvPolicyApiHandler,
			edgeproto.RegisterResTagTableApiHandler,
			edgeproto.RegisterTrustPolicyApiHandler,
			edgeproto.RegisterTrustPolicyExceptionApiHandler,
			edgeproto.RegisterSettingsApiHandler,
			edgeproto.RegisterRateLimitSettingsApiHandler,
			edgeproto.RegisterAppInstClientApiHandler,
			edgeproto.RegisterDebugApiHandler,
			edgeproto.RegisterDeviceApiHandler,
			edgeproto.RegisterOrganizationApiHandler,
			edgeproto.RegisterAlertPolicyApiHandler,
		},
	}
	gw, err := cloudcommon.GrpcGateway(gwcfg)
	if err != nil {
		return fmt.Errorf("Failed to create grpc gateway, %v", err)
	}
	mux.Handle("/", gw)
	// Suppress contant stream of TLS error logs due to LB health check. There is discussion in the community
	//to get rid of some of these logs, but as of now this a the way around it.   We could miss other logs here but
	// the excessive error logs are drowning out everthing else.
	var nullLogger baselog.Logger
	nullLogger.SetOutput(ioutil.Discard)

	httpServer := &http.Server{
		Addr:      *httpAddr,
		Handler:   mux,
		TLSConfig: apiTlsConfig,
		ErrorLog:  &nullLogger,
	}
	go func() {
		var err error
		if httpServer.TLSConfig == nil {
			err = httpServer.ListenAndServe()
		} else {
			err = httpServer.ListenAndServeTLS("", "")
		}
		if err != nil && err != http.ErrServerClosed {
			log.FatalLog("Failed to server grpc gateway", "err", err)
		}
	}()
	services.httpServer = httpServer

	// start the checkpointer
	err = checkInterval()
	if err != nil {
		return err
	}
	go runCheckpoints(ctx)

	// setup cleanup timer to remove expired stream messages
	go streamObjs.SetupCleanupTimer()

	log.SpanLog(ctx, log.DebugLevelInfo, "Ready")
	return nil
}

func stopServices() {
	if services.httpServer != nil {
		services.httpServer.Shutdown(context.Background())
	}
	if services.grpcServer != nil {
		services.grpcServer.Stop()
	}
	if services.publicCertManager != nil {
		services.publicCertManager.StopRefresh()
	}
	services.accessKeyGrpcServer.Stop()
	if services.notifyServerMgr {
		notify.ServerMgrOne.Stop()
	}
	if services.notifyClient != nil {
		services.notifyClient.Stop()
	}
	if services.stopInitCC != nil {
		close(services.stopInitCC)
	}
	if services.influxQ != nil {
		services.influxQ.Stop()
	}
	if services.events != nil {
		services.events.Stop()
	}
	if services.edgeEventsInfluxQ != nil {
		services.edgeEventsInfluxQ.Stop()
	}
	if services.cloudletResourcesInfluxQ != nil {
		services.cloudletResourcesInfluxQ.Stop()
	}
	if services.downsampledMetricsInfluxQ != nil {
		services.downsampledMetricsInfluxQ.Stop()
	}
	if syncLeaseData.stop != nil {
		syncLeaseData.Stop()
	}
	if services.sync != nil {
		services.sync.Done()
	}
	if services.etcdLocal != nil {
		services.etcdLocal.StopLocal()
	}
	for _, lis := range services.listeners {
		lis.Close()
	}
	nodeMgr.Finish()
	services = Services{}
}

// Helper function to verify the compatibility of etcd version and
// current data model version
func checkVersion(ctx context.Context, objStore objstore.KVStore) (string, error) {
	key := objstore.DbKeyPrefixString("Version")
	val, _, _, err := objStore.Get(key)
	if err != nil {
		if !strings.Contains(err.Error(), objstore.NotFoundError(key).Error()) {
			return "", err
		}
	}
	verHash := string(val)
	// If this is the first upgrade, just write the latest hash into etcd
	if verHash == "" {
		log.InfoLog("Could not find a previous version", "curr hash", edgeproto.GetDataModelVersion())
		key := objstore.DbKeyPrefixString("Version")
		_, err = objStore.Put(ctx, key, edgeproto.GetDataModelVersion())
		if err != nil {
			return "", err
		}
		return edgeproto.GetDataModelVersion(), nil
	}
	if edgeproto.GetDataModelVersion() != verHash {
		return verHash, ErrCtrlUpgradeRequired
	}
	return verHash, nil
}

func InitApis(sync *Sync) {
	InitAppApi(sync)
	InitOperatorCodeApi(sync)
	InitCloudletApi(sync)
	InitAppInstApi(sync)
	InitFlavorApi(sync)
	InitStreamObjApi(sync)
	InitClusterInstApi(sync)
	InitCloudletInfoApi(sync)
	InitVMPoolApi(sync)
	InitVMPoolInfoApi(sync)
	InitAppInstInfoApi(sync)
	InitClusterInstInfoApi(sync)
	InitCloudletRefsApi(sync)
	InitAppInstRefsApi(sync)
	InitControllerApi(sync)
	InitCloudletPoolApi(sync)
	InitExecApi()
	InitAlertApi(sync)
	InitAutoScalePolicyApi(sync)
	InitAutoProvPolicyApi(sync)
	InitAutoProvInfoApi(sync)
	InitResTagTableApi(sync)
	InitTrustPolicyApi(sync)
	InitTrustPolicyExceptionApi(sync)
	InitSettingsApi(sync)
	InitRateLimitSettingsApi(sync)
	InitAppInstClientKeyApi(sync)
	InitAppInstClientApi()
	InitDeviceApi(sync)
	InitOrganizationApi(sync)
	InitAppInstLatencyApi(sync)
	InitGPUDriverApi(sync)
	InitAlertPolicyApi(sync)
	InitNetworkApi(sync)
}

func InitNotify(metricsInflux *influxq.InfluxQ, edgeEventsInflux *influxq.InfluxQ, clientQ notify.RecvAppInstClientHandler) {
	notify.ServerMgrOne.RegisterSendSettingsCache(&settingsApi.cache)
	notify.ServerMgrOne.RegisterSendFlowRateLimitSettingsCache(&rateLimitSettingsApi.flowcache)
	notify.ServerMgrOne.RegisterSendMaxReqsRateLimitSettingsCache(&rateLimitSettingsApi.maxreqscache)
	notify.ServerMgrOne.RegisterSendOperatorCodeCache(&operatorCodeApi.cache)
	notify.ServerMgrOne.RegisterSendFlavorCache(&flavorApi.cache)
	notify.ServerMgrOne.RegisterSendGPUDriverCache(&gpuDriverApi.cache)
	notify.ServerMgrOne.RegisterSendVMPoolCache(&vmPoolApi.cache)
	notify.ServerMgrOne.RegisterSendResTagTableCache(&resTagTableApi.cache)
	notify.ServerMgrOne.RegisterSendTrustPolicyCache(&trustPolicyApi.cache)
	notify.ServerMgrOne.RegisterSendCloudletCache(cloudletApi.cache)
	// Be careful on dependencies.
	// CloudletPools must be sent after cloudlets, because they reference cloudlets.
	notify.ServerMgrOne.RegisterSendCloudletPoolCache(cloudletPoolApi.cache)

	notify.ServerMgrOne.RegisterSendCloudletInfoCache(&cloudletInfoApi.cache)
	notify.ServerMgrOne.RegisterSendAutoScalePolicyCache(&autoScalePolicyApi.cache)
	notify.ServerMgrOne.RegisterSendAutoProvPolicyCache(&autoProvPolicyApi.cache)
	notify.ServerMgrOne.RegisterSendNetworkCache(&networkApi.cache)
	notify.ServerMgrOne.RegisterSendClusterInstCache(&clusterInstApi.cache)
	notify.ServerMgrOne.RegisterSendAppCache(&appApi.cache)
	notify.ServerMgrOne.RegisterSendAppInstCache(&appInstApi.cache)
	notify.ServerMgrOne.RegisterSendAppInstRefsCache(&appInstRefsApi.cache)
	notify.ServerMgrOne.RegisterSendAlertCache(&alertApi.cache)
	notify.ServerMgrOne.RegisterSendAppInstClientKeyCache(&appInstClientKeyApi.cache)
	notify.ServerMgrOne.RegisterSendAlertPolicyCache(&userAlertApi.cache)
	// TrustPolicyExceptions depend on App and Cloudlet so must be sent after them.
	notify.ServerMgrOne.RegisterSendTrustPolicyExceptionCache(&trustPolicyExceptionApi.cache)
	notify.ServerMgrOne.RegisterSend(execRequestSendMany)

	nodeMgr.RegisterServer(&notify.ServerMgrOne)
	notify.ServerMgrOne.RegisterRecv(notify.NewCloudletInfoRecvMany(&cloudletInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewAppInstInfoRecvMany(&appInstInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewVMPoolInfoRecvMany(&vmPoolInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewClusterInstInfoRecvMany(&clusterInstInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewExecRequestRecvMany(&execApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewAlertRecvMany(&alertApi))
	autoProvPolicyApi.SetInfluxQ(metricsInflux)
	notify.ServerMgrOne.RegisterRecv(notify.NewAutoProvCountsRecvMany(&autoProvPolicyApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewAppInstClientRecvMany(clientQ))
	notify.ServerMgrOne.RegisterRecv(notify.NewDeviceRecvMany(&deviceApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewAutoProvInfoRecvMany(&autoProvInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewMetricRecvMany(NewControllerMetricsReceiver(metricsInflux, edgeEventsInflux)))
}

type ControllerMetricsReceiver struct {
	metricsInflux    *influxq.InfluxQ
	edgeEventsInflux *influxq.InfluxQ
}

func NewControllerMetricsReceiver(metricsInflux *influxq.InfluxQ, edgeEventsInflux *influxq.InfluxQ) *ControllerMetricsReceiver {
	c := new(ControllerMetricsReceiver)
	c.metricsInflux = metricsInflux
	c.edgeEventsInflux = edgeEventsInflux
	return c
}

// Send metric to correct influxdb
func (c *ControllerMetricsReceiver) RecvMetric(ctx context.Context, metric *edgeproto.Metric) {
	if _, ok := cloudcommon.EdgeEventsMetrics[metric.Name]; ok {
		c.edgeEventsInflux.AddMetric(metric)
	} else {
		c.metricsInflux.AddMetric(metric)
	}
}

const (
	ToggleFlavorMatchVerbose = "toggle-flavormatch-verbose"
	ShowControllers          = "show-controllers"
)

func initDebug(ctx context.Context, nodeMgr *node.NodeMgr) {
	nodeMgr.Debug.AddDebugFunc(ToggleFlavorMatchVerbose,
		func(ctx context.Context, req *edgeproto.DebugRequest) string {
			return vmspec.ToggleFlavorMatchVerbose()
		})
	nodeMgr.Debug.AddDebugFunc(ShowControllers, showControllers)
}

func showControllers(ctx context.Context, req *edgeproto.DebugRequest) string {
	objs := []edgeproto.Controller{}
	controllerApi.cache.Show(&edgeproto.Controller{}, func(obj *edgeproto.Controller) error {
		objs = append(objs, *obj)
		return nil
	})
	out, err := yaml.Marshal(objs)
	if err != nil {
		return fmt.Sprintf("Failed to marshal objs, %v", err)
	}
	return string(out)
}

func initContinuousQueries() {
	done := false
	for !done {
		if services.stopInitCC == nil {
			break
		}
		span := log.StartSpan(log.DebugLevelInfo, "initContinuousQueries")
		ctx := log.ContextWithSpan(context.Background(), span)

		// create continuous queries for edgeevents metrics
		var err error
		for _, collectioninterval := range settingsApi.Get().EdgeEventsMetricsContinuousQueriesCollectionIntervals {
			interval := time.Duration(collectioninterval.Interval)
			retention := time.Duration(collectioninterval.Retention)
			latencyCqSettings := influxq.CreateLatencyContinuousQuerySettings(interval, retention)
			err = influxq.CreateContinuousQuery(services.edgeEventsInfluxQ, services.downsampledMetricsInfluxQ, latencyCqSettings)
			if err != nil && strings.Contains(err.Error(), "already exists") {
				err = nil
			}
			if err != nil {
				break
			}
			deviceCqSettings := influxq.CreateDeviceInfoContinuousQuerySettings(interval, retention)
			err = influxq.CreateContinuousQuery(services.edgeEventsInfluxQ, services.downsampledMetricsInfluxQ, deviceCqSettings)
			if err != nil && strings.Contains(err.Error(), "already exists") {
				err = nil
			}
			if err != nil {
				break
			}
		}
		if err == nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "initContinuousQueries done")
			span.Finish()
			break
		}
		log.SpanLog(ctx, log.DebugLevelInfo, "initContinuousQueries", "err", err)
		span.Finish()
		select {
		case <-time.After(influxq.InfluxQReconnectDelay):
		case <-services.stopInitCC:
			done = true
		}
	}
}
