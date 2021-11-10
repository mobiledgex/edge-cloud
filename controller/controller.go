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
	allApis                   *AllApis
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
	allApis := NewAllApis(sync)
	services.allApis = allApis
	sync.Start()
	services.sync = sync
	// requireNotifyAccessKey allows for backwards compatibility when
	// set to false, because it allows CRMs to connect to notify without
	// an access key (as long as pki internal cert is verified).
	allApis.cloudletApi.accessKeyServer.SetRequireTlsAccessKey(*requireNotifyAccessKey)

	allApis.Start(ctx)

	initDebug(ctx, &nodeMgr, allApis)

	err = allApis.settingsApi.initDefaults(ctx)
	if err != nil {
		return fmt.Errorf("Failed to init settings, %v", err)
	}
	// cleanup thread must start after settings are loaded
	go allApis.clusterInstApi.cleanupThread()

	err = allApis.flowRateLimitSettingsApi.initDefaultRateLimitSettings(ctx)
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
	downsampledMetricsInfluxQ.InitRetentionPolicy(allApis.settingsApi.Get().InfluxDbDownsampledMetricsRetention.TimeDuration())
	err = downsampledMetricsInfluxQ.Start(*influxAddr)
	if err != nil {
		return fmt.Errorf("Failed to start influx queue address %s, %v",
			*influxAddr, err)
	}
	services.downsampledMetricsInfluxQ = downsampledMetricsInfluxQ

	// metrics influx
	influxQ := influxq.NewInfluxQ(InfluxDBName, influxAuth.User, influxAuth.Pass)
	influxQ.InitRetentionPolicy(allApis.settingsApi.Get().InfluxDbMetricsRetention.TimeDuration())
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
	edgeEventsInfluxQ.InitRetentionPolicy(allApis.settingsApi.Get().InfluxDbEdgeEventsMetricsRetention.TimeDuration())
	err = edgeEventsInfluxQ.Start(*influxAddr)
	if err != nil {
		return fmt.Errorf("Failed to start influx queue address %s, %v",
			*influxAddr, err)
	}
	services.edgeEventsInfluxQ = edgeEventsInfluxQ

	// cloudlet resources influx
	cloudletResourcesInfluxQ := influxq.NewInfluxQ(cloudcommon.CloudletResourceUsageDbName, influxAuth.User, influxAuth.Pass)
	cloudletResourcesInfluxQ.InitRetentionPolicy(allApis.settingsApi.Get().InfluxDbCloudletUsageMetricsRetention.TimeDuration())
	err = cloudletResourcesInfluxQ.Start(*influxAddr)
	if err != nil {
		return fmt.Errorf("Failed to start influx queue address %s, %v",
			*influxAddr, err)
	}
	services.cloudletResourcesInfluxQ = cloudletResourcesInfluxQ

	// create continuous queries for edgeevents metrics
	services.stopInitCC = make(chan bool)
	go initContinuousQueries(allApis)

	InitNotify(influxQ, edgeEventsInfluxQ, allApis.appInstClientApi, allApis)
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
		notifyClient.RegisterSendAlertCache(&allApis.alertApi.cache)
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
			allApis.cloudletApi.accessKeyServer.UnaryTlsAccessKey,
		))
	notifyStreamInterceptor := grpc.StreamInterceptor(
		grpc_middleware.ChainStreamServer(
			cloudcommon.AuditStreamInterceptor,
			allApis.cloudletApi.accessKeyServer.StreamTlsAccessKey,
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
	err = services.accessKeyGrpcServer.Start(*accessApiAddr, allApis.cloudletApi.accessKeyServer, accessServerTlsConfig, func(accessServer *grpc.Server) {
		edgeproto.RegisterCloudletAccessApiServer(accessServer, allApis.cloudletApi)
		edgeproto.RegisterCloudletAccessKeyApiServer(accessServer, allApis.cloudletApi)
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
	edgeproto.RegisterAppApiServer(server, allApis.appApi)
	edgeproto.RegisterResTagTableApiServer(server, allApis.resTagTableApi)
	edgeproto.RegisterOperatorCodeApiServer(server, allApis.operatorCodeApi)
	edgeproto.RegisterFlavorApiServer(server, allApis.flavorApi)
	edgeproto.RegisterClusterInstApiServer(server, allApis.clusterInstApi)
	edgeproto.RegisterCloudletApiServer(server, allApis.cloudletApi)
	edgeproto.RegisterAppInstApiServer(server, allApis.appInstApi)
	edgeproto.RegisterCloudletInfoApiServer(server, allApis.cloudletInfoApi)
	edgeproto.RegisterVMPoolApiServer(server, allApis.vmPoolApi)
	edgeproto.RegisterCloudletRefsApiServer(server, allApis.cloudletRefsApi)
	edgeproto.RegisterClusterRefsApiServer(server, allApis.clusterRefsApi)
	edgeproto.RegisterAppInstRefsApiServer(server, allApis.appInstRefsApi)
	edgeproto.RegisterStreamObjApiServer(server, allApis.streamObjApi)
	edgeproto.RegisterControllerApiServer(server, allApis.controllerApi)
	edgeproto.RegisterNodeApiServer(server, &nodeApi)
	edgeproto.RegisterExecApiServer(server, allApis.execApi)
	edgeproto.RegisterCloudletPoolApiServer(server, allApis.cloudletPoolApi)
	edgeproto.RegisterAlertApiServer(server, allApis.alertApi)
	edgeproto.RegisterAutoScalePolicyApiServer(server, allApis.autoScalePolicyApi)
	edgeproto.RegisterAutoProvPolicyApiServer(server, allApis.autoProvPolicyApi)
	edgeproto.RegisterTrustPolicyApiServer(server, allApis.trustPolicyApi)
	edgeproto.RegisterTrustPolicyExceptionApiServer(server, allApis.trustPolicyExceptionApi)
	edgeproto.RegisterSettingsApiServer(server, allApis.settingsApi)
	rateLimitSettingsApi := RateLimitSettingsApi{
		FlowRateLimitSettingsApi:    allApis.flowRateLimitSettingsApi,
		MaxReqsRateLimitSettingsApi: allApis.maxReqsRateLimitSettingsApi,
	}
	edgeproto.RegisterRateLimitSettingsApiServer(server, &rateLimitSettingsApi)
	edgeproto.RegisterAppInstClientApiServer(server, allApis.appInstClientApi)
	edgeproto.RegisterDebugApiServer(server, &debugApi)
	edgeproto.RegisterDeviceApiServer(server, allApis.deviceApi)
	edgeproto.RegisterOrganizationApiServer(server, allApis.organizationApi)
	edgeproto.RegisterAppInstLatencyApiServer(server, allApis.appInstLatencyApi)
	edgeproto.RegisterGPUDriverApiServer(server, allApis.gpuDriverApi)
	edgeproto.RegisterAlertPolicyApiServer(server, allApis.alertPolicyApi)
	edgeproto.RegisterNetworkApiServer(server, allApis.networkApi)

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
	go allApis.clusterInstApi.runCheckpoints(ctx)

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
	if services.allApis != nil {
		services.allApis.Stop()
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

type AllApis struct {
	appApi                      *AppApi
	operatorCodeApi             *OperatorCodeApi
	cloudletApi                 *CloudletApi
	appInstApi                  *AppInstApi
	flavorApi                   *FlavorApi
	streamObjApi                *StreamObjApi
	clusterInstApi              *ClusterInstApi
	cloudletInfoApi             *CloudletInfoApi
	vmPoolApi                   *VMPoolApi
	vmPoolInfoApi               *VMPoolInfoApi
	appInstInfoApi              *AppInstInfoApi
	clusterInstInfoApi          *ClusterInstInfoApi
	cloudletRefsApi             *CloudletRefsApi
	clusterRefsApi              *ClusterRefsApi
	appInstRefsApi              *AppInstRefsApi
	controllerApi               *ControllerApi
	cloudletPoolApi             *CloudletPoolApi
	execApi                     *ExecApi
	alertApi                    *AlertApi
	autoScalePolicyApi          *AutoScalePolicyApi
	autoProvPolicyApi           *AutoProvPolicyApi
	autoProvInfoApi             *AutoProvInfoApi
	resTagTableApi              *ResTagTableApi
	trustPolicyApi              *TrustPolicyApi
	trustPolicyExceptionApi     *TrustPolicyExceptionApi
	settingsApi                 *SettingsApi
	flowRateLimitSettingsApi    *FlowRateLimitSettingsApi
	maxReqsRateLimitSettingsApi *MaxReqsRateLimitSettingsApi
	appInstClientKeyApi         *AppInstClientKeyApi
	appInstClientApi            *AppInstClientApi
	deviceApi                   *DeviceApi
	organizationApi             *OrganizationApi
	appInstLatencyApi           *AppInstLatencyApi
	gpuDriverApi                *GPUDriverApi
	alertPolicyApi              *AlertPolicyApi
	networkApi                  *NetworkApi
	syncLeaseData               *SyncLeaseData
}

func NewAllApis(sync *Sync) *AllApis {
	all := &AllApis{}
	all.appApi = NewAppApi(sync, all)
	all.operatorCodeApi = NewOperatorCodeApi(sync, all)
	all.cloudletApi = NewCloudletApi(sync, all)
	all.appInstApi = NewAppInstApi(sync, all)
	all.flavorApi = NewFlavorApi(sync, all)
	all.streamObjApi = NewStreamObjApi(sync, all)
	all.clusterInstApi = NewClusterInstApi(sync, all)
	all.cloudletInfoApi = NewCloudletInfoApi(sync, all)
	all.vmPoolApi = NewVMPoolApi(sync, all)
	all.vmPoolInfoApi = NewVMPoolInfoApi(sync, all)
	all.appInstInfoApi = NewAppInstInfoApi(sync, all)
	all.clusterInstInfoApi = NewClusterInstInfoApi(sync, all)
	all.cloudletRefsApi = NewCloudletRefsApi(sync, all)
	all.clusterRefsApi = NewClusterRefsApi(sync, all)
	all.appInstRefsApi = NewAppInstRefsApi(sync, all)
	all.controllerApi = NewControllerApi(sync, all)
	all.cloudletPoolApi = NewCloudletPoolApi(sync, all)
	all.execApi = NewExecApi(all)
	all.alertApi = NewAlertApi(sync, all)
	all.autoScalePolicyApi = NewAutoScalePolicyApi(sync, all)
	all.autoProvPolicyApi = NewAutoProvPolicyApi(sync, all)
	all.autoProvInfoApi = NewAutoProvInfoApi(sync, all)
	all.resTagTableApi = NewResTagTableApi(sync, all)
	all.trustPolicyApi = NewTrustPolicyApi(sync, all)
	all.trustPolicyExceptionApi = NewTrustPolicyExceptionApi(sync, all)
	all.settingsApi = NewSettingsApi(sync, all)
	all.flowRateLimitSettingsApi = NewFlowRateLimitSettingsApi(sync, all)
	all.maxReqsRateLimitSettingsApi = NewMaxReqsRateLimitSettingsApi(sync, all)
	all.appInstClientKeyApi = NewAppInstClientKeyApi(sync, all)
	all.appInstClientApi = NewAppInstClientApi(all)
	all.deviceApi = NewDeviceApi(sync, all)
	all.organizationApi = NewOrganizationApi(sync, all)
	all.appInstLatencyApi = NewAppInstLatencyApi(sync, all)
	all.gpuDriverApi = NewGPUDriverApi(sync, all)
	all.alertPolicyApi = NewAlertPolicyApi(sync, all)
	all.networkApi = NewNetworkApi(sync, all)
	all.syncLeaseData = NewSyncLeaseData(sync, all)
	return all
}

func (s *AllApis) Start(ctx context.Context) {
	s.syncLeaseData.Start(ctx)
}

func (s *AllApis) Stop() {
	if s.syncLeaseData.stop != nil {
		s.syncLeaseData.Stop()
	}
}

func InitNotify(metricsInflux *influxq.InfluxQ, edgeEventsInflux *influxq.InfluxQ, clientQ notify.RecvAppInstClientHandler, allApis *AllApis) {
	notify.ServerMgrOne.RegisterSendSettingsCache(&allApis.settingsApi.cache)
	notify.ServerMgrOne.RegisterSendFlowRateLimitSettingsCache(&allApis.flowRateLimitSettingsApi.cache)
	notify.ServerMgrOne.RegisterSendMaxReqsRateLimitSettingsCache(&allApis.maxReqsRateLimitSettingsApi.cache)
	notify.ServerMgrOne.RegisterSendOperatorCodeCache(&allApis.operatorCodeApi.cache)
	notify.ServerMgrOne.RegisterSendFlavorCache(&allApis.flavorApi.cache)
	notify.ServerMgrOne.RegisterSendGPUDriverCache(&allApis.gpuDriverApi.cache)
	notify.ServerMgrOne.RegisterSendVMPoolCache(&allApis.vmPoolApi.cache)
	notify.ServerMgrOne.RegisterSendResTagTableCache(&allApis.resTagTableApi.cache)
	notify.ServerMgrOne.RegisterSendTrustPolicyCache(&allApis.trustPolicyApi.cache)
	notify.ServerMgrOne.RegisterSendCloudletCache(allApis.cloudletApi.cache)
	// Be careful on dependencies.
	// CloudletPools must be sent after cloudlets, because they reference cloudlets.
	notify.ServerMgrOne.RegisterSendCloudletPoolCache(allApis.cloudletPoolApi.cache)

	notify.ServerMgrOne.RegisterSendCloudletInfoCache(&allApis.cloudletInfoApi.cache)
	notify.ServerMgrOne.RegisterSendAutoScalePolicyCache(&allApis.autoScalePolicyApi.cache)
	notify.ServerMgrOne.RegisterSendAutoProvPolicyCache(&allApis.autoProvPolicyApi.cache)
	notify.ServerMgrOne.RegisterSendNetworkCache(&allApis.networkApi.cache)
	notify.ServerMgrOne.RegisterSendClusterInstCache(&allApis.clusterInstApi.cache)
	notify.ServerMgrOne.RegisterSendAppCache(&allApis.appApi.cache)
	notify.ServerMgrOne.RegisterSendAppInstCache(&allApis.appInstApi.cache)
	notify.ServerMgrOne.RegisterSendAppInstRefsCache(&allApis.appInstRefsApi.cache)
	notify.ServerMgrOne.RegisterSendAlertCache(&allApis.alertApi.cache)
	notify.ServerMgrOne.RegisterSendAppInstClientKeyCache(&allApis.appInstClientKeyApi.cache)
	notify.ServerMgrOne.RegisterSendAlertPolicyCache(&allApis.alertPolicyApi.cache)
	// TrustPolicyExceptions depend on App and Cloudlet so must be sent after them.
	notify.ServerMgrOne.RegisterSendTrustPolicyExceptionCache(&allApis.trustPolicyExceptionApi.cache)
	notify.ServerMgrOne.RegisterSend(execRequestSendMany)

	nodeMgr.RegisterServer(&notify.ServerMgrOne)
	notify.ServerMgrOne.RegisterRecv(notify.NewCloudletInfoRecvMany(allApis.cloudletInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewAppInstInfoRecvMany(allApis.appInstInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewVMPoolInfoRecvMany(allApis.vmPoolInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewClusterInstInfoRecvMany(allApis.clusterInstInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewExecRequestRecvMany(allApis.execApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewAlertRecvMany(allApis.alertApi))
	allApis.autoProvPolicyApi.SetInfluxQ(metricsInflux)
	notify.ServerMgrOne.RegisterRecv(notify.NewAutoProvCountsRecvMany(allApis.autoProvPolicyApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewAppInstClientRecvMany(clientQ))
	notify.ServerMgrOne.RegisterRecv(notify.NewDeviceRecvMany(allApis.deviceApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewAutoProvInfoRecvMany(allApis.autoProvInfoApi))
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

func initDebug(ctx context.Context, nodeMgr *node.NodeMgr, allApis *AllApis) {
	nodeMgr.Debug.AddDebugFunc(ToggleFlavorMatchVerbose,
		func(ctx context.Context, req *edgeproto.DebugRequest) string {
			return vmspec.ToggleFlavorMatchVerbose()
		})
	nodeMgr.Debug.AddDebugFunc(ShowControllers, allApis.controllerApi.showControllers)
}

func (s *ControllerApi) showControllers(ctx context.Context, req *edgeproto.DebugRequest) string {
	objs := []edgeproto.Controller{}
	s.cache.Show(&edgeproto.Controller{}, func(obj *edgeproto.Controller) error {
		objs = append(objs, *obj)
		return nil
	})
	out, err := yaml.Marshal(objs)
	if err != nil {
		return fmt.Sprintf("Failed to marshal objs, %v", err)
	}
	return string(out)
}

func initContinuousQueries(allApis *AllApis) {
	done := false
	for !done {
		if services.stopInitCC == nil {
			break
		}
		span := log.StartSpan(log.DebugLevelInfo, "initContinuousQueries")
		ctx := log.ContextWithSpan(context.Background(), span)

		// create continuous queries for edgeevents metrics
		var err error
		for _, collectioninterval := range allApis.settingsApi.Get().EdgeEventsMetricsContinuousQueriesCollectionIntervals {
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
