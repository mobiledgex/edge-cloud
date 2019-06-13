// Main process

package main

import (
	"context"
	ctls "crypto/tls"
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
	influxq "github.com/mobiledgex/edge-cloud/controller/influxq_client"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"
)

// Command line options
var rootDir = flag.String("r", "", "root directory; set for testing")
var localEtcd = flag.Bool("localEtcd", false, "set to start local etcd for testing")
var initLocalEtcd = flag.Bool("initLocalEtcd", false, "set to init local etcd database")
var region = flag.Uint("region", 1, "Region")
var etcdUrls = flag.String("etcdUrls", "http://127.0.0.1:2380", "etcd client listener URLs")
var apiAddr = flag.String("apiAddr", "127.0.0.1:55001", "API listener address")

// external API Addr is registered with etcd so other controllers can connect
// directly to this controller.
var externalApiAddr = flag.String("externalApiAddr", "", "External API listener address if behind proxy/LB. Defaults to apiAddr")
var httpAddr = flag.String("httpAddr", "127.0.0.1:8091", "HTTP listener address")
var notifyAddr = flag.String("notifyAddr", "127.0.0.1:50001", "Notify listener address")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var tlsCertFile = flag.String("tls", "", "server tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory")
var shortTimeouts = flag.Bool("shortTimeouts", false, "set CRM timeouts short for simulated cloudlet testing")
var influxAddr = flag.String("influxAddr", "http://127.0.0.1:8086", "InfluxDB listener address")
var skipVersionCheck = flag.Bool("skipVersionCheck", false, "Skip etcd version hash verification")
var autoUpgrade = flag.Bool("autoUpgrade", false, "Automatically upgrade etcd database to the current version")
var ControllerId = ""
var InfluxDBName = "metrics"

func GetRootDir() string {
	return *rootDir
}

var ErrCtrlAlreadyInProgress = errors.New("Change already in progress")
var ErrCtrlUpgradeRequired = errors.New("data mode upgrade required")

var sigChan chan os.Signal
var services Services

type Services struct {
	etcdLocal       *process.Etcd
	sync            *Sync
	influxQ         *influxq.InfluxQ
	notifyServerMgr bool
	grpcServer      *grpc.Server
	httpServer      *http.Server
}

func main() {
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

func startServices() error {
	log.SetDebugLevelStrs(*debugLevels)

	if *externalApiAddr == "" {
		*externalApiAddr = *apiAddr
	}

	log.InfoLog("Start up", "rootDir", *rootDir, "apiAddr", *apiAddr, "externalApiAddr", *externalApiAddr)
	objstore.InitRegion(uint32(*region))

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
		version, err := checkVersion(objStore)
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
	err = controllerApi.registerController()
	if err != nil {
		return fmt.Errorf("Failed to register controller, %v", err)
	}

	influxQ := influxq.NewInfluxQ(InfluxDBName)
	clientCert := strings.Replace(*tlsCertFile, "server", "client", 1)
	err = influxQ.Start(*influxAddr, clientCert)
	if err != nil {
		return fmt.Errorf("Failed to start influx queue address %s, %v",
			*influxAddr, err)
	}
	services.influxQ = influxQ

	InitNotify(influxQ)
	notify.ServerMgrOne.Start(*notifyAddr, *tlsCertFile)
	services.notifyServerMgr = true

	creds, err := tls.GetTLSServerCreds(*tlsCertFile)
	if err != nil {
		return fmt.Errorf("get TLS Credentials failed, %v", err)
	}
	server := grpc.NewServer(grpc.Creds(creds),
		grpc.UnaryInterceptor(AuditUnaryInterceptor),
		grpc.StreamInterceptor(AuditStreamInterceptor))
	edgeproto.RegisterDeveloperApiServer(server, &developerApi)
	edgeproto.RegisterAppApiServer(server, &appApi)
	edgeproto.RegisterOperatorApiServer(server, &operatorApi)
	edgeproto.RegisterFlavorApiServer(server, &flavorApi)
	edgeproto.RegisterClusterInstApiServer(server, &clusterInstApi)
	edgeproto.RegisterCloudletApiServer(server, &cloudletApi)
	edgeproto.RegisterAppInstApiServer(server, &appInstApi)
	edgeproto.RegisterCloudletInfoApiServer(server, &cloudletInfoApi)
	edgeproto.RegisterCloudletRefsApiServer(server, &cloudletRefsApi)
	edgeproto.RegisterControllerApiServer(server, &controllerApi)
	edgeproto.RegisterNodeApiServer(server, &nodeApi)
	edgeproto.RegisterExecApiServer(server, &execApi)
	log.RegisterDebugApiServer(server, &log.Api{})

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
		ApiAddr:     *apiAddr,
		TlsCertFile: *tlsCertFile,
		ApiHandles: []func(context.Context, *gwruntime.ServeMux, *grpc.ClientConn) error{
			edgeproto.RegisterDeveloperApiHandler,
			edgeproto.RegisterAppApiHandler,
			edgeproto.RegisterAppInstApiHandler,
			edgeproto.RegisterOperatorApiHandler,
			edgeproto.RegisterCloudletApiHandler,
			edgeproto.RegisterCloudletInfoApiHandler,
			edgeproto.RegisterFlavorApiHandler,
			edgeproto.RegisterClusterInstApiHandler,
			edgeproto.RegisterControllerApiHandler,
			edgeproto.RegisterNodeApiHandler,
		},
	}
	gw, err := cloudcommon.GrpcGateway(gwcfg)
	if err != nil {
		return fmt.Errorf("Failed to create grpc gateway, %v", err)
	}
	mux.Handle("/", gw)
	tlscfg := &ctls.Config{
		MinVersion:               ctls.VersionTLS12,
		CurvePreferences:         []ctls.CurveID{ctls.CurveP521, ctls.CurveP384, ctls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			ctls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			ctls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			ctls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			ctls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			ctls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	// Suppress contant stream of TLS error logs due to LB health check. There is discussion in the community
	//to get rid of some of these logs, but as of now this a the way around it.   We could miss other logs here but
	// the excessive error logs are drowning out everthing else.
	var nullLogger baselog.Logger
	nullLogger.SetOutput(ioutil.Discard)

	httpServer := &http.Server{
		Addr:      *httpAddr,
		Handler:   mux,
		TLSConfig: tlscfg,
		ErrorLog:  &nullLogger,
	}
	go cloudcommon.GrpcGatewayServe(gwcfg, httpServer)
	services.httpServer = httpServer

	log.InfoLog("Ready")
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
	if services.influxQ != nil {
		services.influxQ.Stop()
	}
	if services.sync != nil {
		services.sync.Done()
	}
	if services.etcdLocal != nil {
		services.etcdLocal.StopLocal()
	}
}

// Helper function to verify the compatibility of etcd version and
// current data model version
func checkVersion(objStore objstore.KVStore) (string, error) {
	key := objstore.DbKeyPrefixString("Version")
	val, _, _, err := objStore.Get(key)
	if err != nil {
		if !strings.Contains(err.Error(), objstore.ErrKVStoreKeyNotFound.Error()) {
			return "", err
		}
	}
	verHash := string(val)
	// If this is the first upgrade, just write the latest hash into etcd
	if verHash == "" {
		log.InfoLog("Could not find a previous version", "curr hash", edgeproto.GetDataModelVersion())
		key := objstore.DbKeyPrefixString("Version")
		_, err = objStore.Put(key, edgeproto.GetDataModelVersion())
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
	InitDeveloperApi(sync)
	InitAppApi(sync)
	InitOperatorApi(sync)
	InitCloudletApi(sync)
	InitAppInstApi(sync)
	InitFlavorApi(sync)
	InitClusterInstApi(sync)
	InitCloudletInfoApi(sync)
	InitAppInstInfoApi(sync)
	InitClusterInstInfoApi(sync)
	InitCloudletRefsApi(sync)
	InitControllerApi(sync)
	InitNodeApi(sync)
	InitExecApi()
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "nohostname"
	}
	ControllerId = hostname + "@" + *externalApiAddr
}

func InitNotify(influxQ *influxq.InfluxQ) {
	notify.ServerMgrOne.RegisterSendFlavorCache(&flavorApi.cache)
	notify.ServerMgrOne.RegisterSendCloudletCache(&cloudletApi.cache)
	notify.ServerMgrOne.RegisterSendClusterInstCache(&clusterInstApi.cache)
	notify.ServerMgrOne.RegisterSendAppCache(&appApi.cache)
	notify.ServerMgrOne.RegisterSendAppInstCache(&appInstApi.cache)
	notify.ServerMgrOne.RegisterSend(execRequestSendMany)

	notify.ServerMgrOne.RegisterRecv(notify.NewCloudletInfoRecvMany(&cloudletInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewAppInstInfoRecvMany(&appInstInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewClusterInstInfoRecvMany(&clusterInstInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewMetricRecvMany(influxQ))
	notify.ServerMgrOne.RegisterRecv(notify.NewNodeRecvMany(&nodeApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewExecRequestRecvMany(&execApi))
}
