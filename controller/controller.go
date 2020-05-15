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
var publicAddr = flag.String("publicAddr", "127.0.0.1", "Public facing address/hostname of controller")
var edgeTurnAddr = flag.String("edgeTurnAddr", "127.0.0.1:6080", "Address to EdgeTurn Server")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var shortTimeouts = flag.Bool("shortTimeouts", false, "set timeouts short for simulated cloudlet testing")
var influxAddr = flag.String("influxAddr", "http://127.0.0.1:8086", "InfluxDB listener address")
var registryFQDN = flag.String("registryFQDN", "", "default docker image registry FQDN")
var artifactoryFQDN = flag.String("artifactoryFQDN", "", "default VM image registry (artifactory) FQDN")
var cloudletRegistryPath = flag.String("cloudletRegistryPath", "", "edge-cloud image registry path for deploying cloudlet services")
var cloudletVMImagePath = flag.String("cloudletVMImagePath", "", "VM image for deploying cloudlet services")
var versionTag = flag.String("versionTag", "", "edge-cloud image tag indicating controller version")
var skipVersionCheck = flag.Bool("skipVersionCheck", false, "Skip etcd version hash verification")
var autoUpgrade = flag.Bool("autoUpgrade", false, "Automatically upgrade etcd database to the current version")
var testMode = flag.Bool("testMode", false, "Run controller in test mode")
var commercialCerts = flag.Bool("commercialCerts", false, "Have CRM grab certs from LetsEncrypt. If false then CRM will generate its onwn self-signed cert")
var checkpointInterval = flag.Duration("checkpointInterval", time.Hour*24*30, "Interval at which to checkpoint cluster usage")

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
	etcdLocal       *process.Etcd
	sync            *Sync
	influxQ         *influxq.InfluxQ
	events          *influxq.InfluxQ
	notifyServerMgr bool
	grpcServer      *grpc.Server
	httpServer      *http.Server
	notifyClient    *notify.Client
}

func main() {
	nodeMgr.InitFlags()
	flag.Parse()

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
		err = cloudcommon.ValidateDockerRegistryPath(ctx, platform_registry_path, vaultConfig)
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
	log.InitTracer(nodeMgr.TlsCertFile)
	span := log.StartSpan(log.DebugLevelInfo, "main")
	span.SetTag("level", "init")
	defer span.Finish()
	ctx := log.ContextWithSpan(context.Background(), span)

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "nohostname"
	}
	ControllerId = hostname + "@" + *externalApiAddr

	log.SpanLog(ctx, log.DebugLevelInfo, "Start up", "rootDir", *rootDir, "apiAddr", *apiAddr, "externalApiAddr", *externalApiAddr)
	// region number for etcd is a deprecated concept since we decided
	// etcd is per-region.
	objstore.InitRegion(uint32(1))
	if !util.ValidRegion(*region) {
		return fmt.Errorf("invalid region name")
	}

	err = nodeMgr.Init(ctx, node.NodeTypeController, node.WithName(ControllerId), node.WithContainerVersion(*versionTag), node.WithRegion(*region))
	if err != nil {
		return err
	}
	vaultConfig = nodeMgr.VaultConfig

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

	sync := InitSync(objStore)
	InitApis(sync)
	sync.Start()
	services.sync = sync

	// register controller must be called before starting Notify protocol
	// to set up controllerAliveLease.
	err = controllerApi.registerController(ctx)
	if err != nil {
		return fmt.Errorf("Failed to register controller, %v", err)
	}
	err = settingsApi.initDefaults(ctx)
	if err != nil {
		return fmt.Errorf("Failed to init settings, %v", err)
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
	// metrics influx
	influxQ := influxq.NewInfluxQ(InfluxDBName, influxAuth.User, influxAuth.Pass)
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
		notifyClient := notify.NewClient(addrs, dialOption)
		notifyClient.Start()
		services.notifyClient = notifyClient
		nodeMgr.RegisterClient(notifyClient)
	}
	notifyServerTls, err := nodeMgr.InternalPki.GetServerTlsConfig(ctx,
		nodeMgr.CommonName(),
		node.CertIssuerRegional,
		[]node.MatchCA{
			node.SameRegionalMatchCA(),
			node.SameRegionalCloudletMatchCA(),
		})
	if err != nil {
		return err
	}
	InitNotify(influxQ, &appInstClientApi)
	notify.ServerMgrOne.Start(*notifyAddr, notifyServerTls)
	services.notifyServerMgr = true

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
		grpc.UnaryInterceptor(cloudcommon.AuditUnaryInterceptor),
		grpc.StreamInterceptor(cloudcommon.AuditStreamInterceptor))
	edgeproto.RegisterAppApiServer(server, &appApi)
	edgeproto.RegisterResTagTableApiServer(server, &resTagTableApi)
	edgeproto.RegisterOperatorCodeApiServer(server, &operatorCodeApi)
	edgeproto.RegisterFlavorApiServer(server, &flavorApi)
	edgeproto.RegisterClusterInstApiServer(server, &clusterInstApi)
	edgeproto.RegisterCloudletApiServer(server, &cloudletApi)
	edgeproto.RegisterAppInstApiServer(server, &appInstApi)
	edgeproto.RegisterCloudletInfoApiServer(server, &cloudletInfoApi)
	edgeproto.RegisterCloudletRefsApiServer(server, &cloudletRefsApi)
	edgeproto.RegisterControllerApiServer(server, &controllerApi)
	edgeproto.RegisterNodeApiServer(server, &nodeApi)
	edgeproto.RegisterExecApiServer(server, &execApi)
	edgeproto.RegisterCloudletPoolApiServer(server, &cloudletPoolApi)
	edgeproto.RegisterCloudletPoolMemberApiServer(server, &cloudletPoolMemberApi)
	edgeproto.RegisterCloudletPoolShowApiServer(server, &cloudletPoolMemberApi)
	edgeproto.RegisterAlertApiServer(server, &alertApi)
	edgeproto.RegisterAutoScalePolicyApiServer(server, &autoScalePolicyApi)
	edgeproto.RegisterAutoProvPolicyApiServer(server, &autoProvPolicyApi)
	edgeproto.RegisterPrivacyPolicyApiServer(server, &privacyPolicyApi)
	edgeproto.RegisterSettingsApiServer(server, &settingsApi)
	edgeproto.RegisterAppInstClientApiServer(server, &appInstClientApi)
	edgeproto.RegisterDebugApiServer(server, &debugApi)
	edgeproto.RegisterDeviceApiServer(server, &deviceApi)
	edgeproto.RegisterOrganizationApiServer(server, &organizationApi)

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
			edgeproto.RegisterFlavorApiHandler,
			edgeproto.RegisterClusterInstApiHandler,
			edgeproto.RegisterControllerApiHandler,
			edgeproto.RegisterNodeApiHandler,
			edgeproto.RegisterCloudletPoolApiHandler,
			edgeproto.RegisterCloudletPoolMemberApiHandler,
			edgeproto.RegisterCloudletPoolShowApiHandler,
			edgeproto.RegisterAlertApiHandler,
			edgeproto.RegisterAutoScalePolicyApiHandler,
			edgeproto.RegisterAutoProvPolicyApiHandler,
			edgeproto.RegisterResTagTableApiHandler,
			edgeproto.RegisterPrivacyPolicyApiHandler,
			edgeproto.RegisterSettingsApiHandler,
			edgeproto.RegisterAppInstClientApiHandler,
			edgeproto.RegisterDebugApiHandler,
			edgeproto.RegisterDeviceApiHandler,
			edgeproto.RegisterOrganizationApiHandler,
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
	go runClusterCheckpoints(ctx)

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
	if services.notifyServerMgr {
		notify.ServerMgrOne.Stop()
	}
	if services.notifyClient != nil {
		services.notifyClient.Stop()
	}
	if services.influxQ != nil {
		services.influxQ.Stop()
	}
	if services.events != nil {
		services.events.Stop()
	}
	if services.sync != nil {
		services.sync.Done()
	}
	if services.etcdLocal != nil {
		services.etcdLocal.StopLocal()
	}
	log.FinishTracer()
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
	InitClusterInstApi(sync)
	InitCloudletInfoApi(sync)
	InitAppInstInfoApi(sync)
	InitClusterInstInfoApi(sync)
	InitCloudletRefsApi(sync)
	InitControllerApi(sync)
	InitCloudletPoolApi(sync)
	InitCloudletPoolMemberApi(sync)
	InitExecApi()
	InitAlertApi(sync)
	InitAutoScalePolicyApi(sync)
	InitAutoProvPolicyApi(sync)
	InitResTagTableApi(sync)
	InitPrivacyPolicyApi(sync)
	InitSettingsApi(sync)
	InitAppInstClientKeyApi(sync)
	InitAppInstClientApi()
	InitDeviceApi(sync)
	InitOrganizationApi(sync)
}

func InitNotify(influxQ *influxq.InfluxQ, clientQ notify.RecvAppInstClientHandler) {
	notify.ServerMgrOne.RegisterSendSettingsCache(&settingsApi.cache)
	notify.ServerMgrOne.RegisterSendOperatorCodeCache(&operatorCodeApi.cache)
	notify.ServerMgrOne.RegisterSendFlavorCache(&flavorApi.cache)
	notify.ServerMgrOne.RegisterSendCloudletCache(&cloudletApi.cache)
	notify.ServerMgrOne.RegisterSendCloudletInfoCache(&cloudletInfoApi.cache)
	notify.ServerMgrOne.RegisterSendAutoScalePolicyCache(&autoScalePolicyApi.cache)
	notify.ServerMgrOne.RegisterSendAutoProvPolicyCache(&autoProvPolicyApi.cache)
	notify.ServerMgrOne.RegisterSendPrivacyPolicyCache(&privacyPolicyApi.cache)
	notify.ServerMgrOne.RegisterSendClusterInstCache(&clusterInstApi.cache)
	notify.ServerMgrOne.RegisterSendAppCache(&appApi.cache)
	notify.ServerMgrOne.RegisterSendAppInstCache(&appInstApi.cache)
	notify.ServerMgrOne.RegisterSendAlertCache(&alertApi.cache)
	notify.ServerMgrOne.RegisterSendAppInstClientKeyCache(&appInstClientKeyApi.cache)

	notify.ServerMgrOne.RegisterSend(execRequestSendMany)

	nodeMgr.RegisterServer(&notify.ServerMgrOne)
	notify.ServerMgrOne.RegisterRecv(notify.NewCloudletInfoRecvMany(&cloudletInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewAppInstInfoRecvMany(&appInstInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewClusterInstInfoRecvMany(&clusterInstInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewMetricRecvMany(influxQ))
	notify.ServerMgrOne.RegisterRecv(notify.NewExecRequestRecvMany(&execApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewAlertRecvMany(&alertApi))
	autoProvPolicyApi.SetInfluxQ(influxQ)
	notify.ServerMgrOne.RegisterRecv(notify.NewAutoProvCountsRecvMany(&autoProvPolicyApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewAppInstClientRecvMany(clientQ))
	notify.ServerMgrOne.RegisterRecv(notify.NewDeviceRecvMany(&deviceApi))
}
