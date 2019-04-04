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
	"time"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	etcd "github.com/mobiledgex/edge-cloud/controller/etcd_client"
	influxq "github.com/mobiledgex/edge-cloud/controller/influxq_client"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"
)

// Command line options
var rootDir = flag.String("r", "", "root directory; set for testing")
var localEtcd = flag.Bool("localEtcd", false, "set to start local etcd for testing")
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
var influxAddr = flag.String("influxAddr", "127.0.0.1:8086", "InfluxDB listener address")
var ControllerId = ""
var InfluxDBName = "metrics"

func GetRootDir() string {
	return *rootDir
}

var ErrCtrlAlreadyInProgress = errors.New("Change already in progress")

var sigChan chan os.Signal

// testing hook
var mainStarted chan struct{}

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)

	if *externalApiAddr == "" {
		*externalApiAddr = *apiAddr
	}

	log.InfoLog("Start up", "rootDir", *rootDir, "apiAddr", *apiAddr, "externalApiAddr", *externalApiAddr)
	objstore.InitRegion(uint32(*region))

	if *localEtcd {
		etcdServer, err := StartLocalEtcdServer()
		if err != nil {
			log.FatalLog("No clientIP and clientPort specified, starting local etcd server failed: %s", err)
		}
		etcdUrls = &etcdServer.Config.ClientUrls
		defer etcdServer.Stop()
	}
	objStore, err := etcd.GetEtcdClientBasic(*etcdUrls)
	if err != nil {
		log.FatalLog("Failed to initialize Object Store", "err", err)
	}
	err = objStore.CheckConnected(50, 20*time.Millisecond)
	if err != nil {
		log.FatalLog("Failed to connect to etcd servers", "err", err)
	}

	lis, err := net.Listen("tcp", *apiAddr)
	if err != nil {
		log.FatalLog("Failed to listen on address", "address", *apiAddr,
			"error", err)
	}

	sync := InitSync(objStore)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	err = controllerApi.registerController()
	if err != nil {
		log.FatalLog("Failed to register controller", "err", err)
	}

	influxQ := influxq.NewInfluxQ(InfluxDBName)
	err = influxQ.Start(*influxAddr)
	if err != nil {
		log.FatalLog("Failed to start influx queue",
			"address", *influxAddr, "err", err)
	}
	defer influxQ.Stop()

	InitNotify(influxQ)
	notify.ServerMgrOne.Start(*notifyAddr, *tlsCertFile)
	defer notify.ServerMgrOne.Stop()

	creds, err := tls.GetTLSServerCreds(*tlsCertFile)
	if err != nil {
		log.FatalLog("get TLS Credentials", "error", err)
	}
	server := grpc.NewServer(grpc.Creds(creds),
		grpc.UnaryInterceptor(AuditUnaryInterceptor),
		grpc.StreamInterceptor(AuditStreamInterceptor))
	edgeproto.RegisterDeveloperApiServer(server, &developerApi)
	edgeproto.RegisterAppApiServer(server, &appApi)
	edgeproto.RegisterOperatorApiServer(server, &operatorApi)
	edgeproto.RegisterFlavorApiServer(server, &flavorApi)
	edgeproto.RegisterClusterFlavorApiServer(server, &clusterFlavorApi)
	edgeproto.RegisterClusterApiServer(server, &clusterApi)
	edgeproto.RegisterClusterInstApiServer(server, &clusterInstApi)
	edgeproto.RegisterCloudletApiServer(server, &cloudletApi)
	edgeproto.RegisterAppInstApiServer(server, &appInstApi)
	edgeproto.RegisterCloudletInfoApiServer(server, &cloudletInfoApi)
	edgeproto.RegisterCloudletRefsApiServer(server, &cloudletRefsApi)
	edgeproto.RegisterControllerApiServer(server, &controllerApi)
	edgeproto.RegisterNodeApiServer(server, &nodeApi)
	log.RegisterDebugApiServer(server, &log.Api{})

	go func() {
		// Serve will block until interrupted and Stop is called
		if err := server.Serve(lis); err != nil {
			log.FatalLog("Failed to serve", "error", err)
		}
	}()
	defer server.Stop()

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
			edgeproto.RegisterClusterFlavorApiHandler,
			edgeproto.RegisterClusterApiHandler,
			edgeproto.RegisterClusterInstApiHandler,
			edgeproto.RegisterControllerApiHandler,
			edgeproto.RegisterNodeApiHandler,
		},
	}
	gw, err := cloudcommon.GrpcGateway(gwcfg)
	if err != nil {
		log.FatalLog("Failed to create grpc gateway", "error", err)
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
	defer httpServer.Shutdown(context.Background())

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	if mainStarted != nil {
		close(mainStarted)
	}
	log.InfoLog("Ready")

	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}

func InitApis(sync *Sync) {
	InitDeveloperApi(sync)
	InitAppApi(sync)
	InitOperatorApi(sync)
	InitCloudletApi(sync)
	InitAppInstApi(sync)
	InitFlavorApi(sync)
	InitClusterFlavorApi(sync)
	InitClusterApi(sync)
	InitClusterInstApi(sync)
	InitCloudletInfoApi(sync)
	InitAppInstInfoApi(sync)
	InitClusterInstInfoApi(sync)
	InitCloudletRefsApi(sync)
	InitControllerApi(sync)
	InitNodeApi(sync)
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "nohostname"
	}
	ControllerId = hostname + "@" + *externalApiAddr
}

func InitNotify(influxQ *influxq.InfluxQ) {
	notify.ServerMgrOne.RegisterSendFlavorCache(&flavorApi.cache)
	notify.ServerMgrOne.RegisterSendClusterFlavorCache(&clusterFlavorApi.cache)
	notify.ServerMgrOne.RegisterSendCloudletCache(&cloudletApi.cache)
	notify.ServerMgrOne.RegisterSendClusterCache(&clusterApi.cache)
	notify.ServerMgrOne.RegisterSendClusterInstCache(&clusterInstApi.cache)
	notify.ServerMgrOne.RegisterSendAppCache(&appApi.cache)
	notify.ServerMgrOne.RegisterSendAppInstCache(&appInstApi.cache)

	notify.ServerMgrOne.RegisterRecv(notify.NewCloudletInfoRecvMany(&cloudletInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewAppInstInfoRecvMany(&appInstInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewClusterInstInfoRecvMany(&clusterInstInfoApi))
	notify.ServerMgrOne.RegisterRecv(notify.NewMetricRecvMany(influxQ))
	notify.ServerMgrOne.RegisterRecv(notify.NewNodeRecvMany(&nodeApi))
}
