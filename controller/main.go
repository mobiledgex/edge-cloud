// Main process

package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/util"
	"google.golang.org/grpc"
)

// Command line options
var rootDir = flag.String("r", "", "root directory; set for testing")
var localEtcd = flag.Bool("localEtcd", false, "set to start local etcd for testing")
var region = flag.Uint("region", 1, "Region")
var etcdUrls = flag.String("etcdUrls", "http://127.0.0.1:2380", "etcd client listener URLs")
var apiIP = flag.String("apiIP", "127.0.0.1", "API listener IP")
var apiPort = flag.Uint("apiPort", 55001, "API listener port")
var httpPort = flag.Uint("httpPort", 8091, "HTTP listener port")
var matcherAddrs = flag.String("matcherAddrs", "", "comma separated list of distributed matching engine addresses")
var cloudletMgrAddrs = flag.String("cloudletMgrAddrs", "", "comma separated list of cloudlet manager addresses")
var debugLevels = flag.String("d", "", "comma separated list of debug levels")

func GetRootDir() string {
	return *rootDir
}

var sigChan chan os.Signal

// testing hook
var mainStarted chan struct{}

func main() {
	flag.Parse()
	util.SetDebugLevelStrs(*debugLevels)

	util.InfoLog("Start up", "rootDir", *rootDir, "apiPort", *apiPort)
	InitRegion(uint32(*region))

	if *localEtcd {
		etcdServer, err := StartLocalEtcdServer()
		if err != nil {
			util.FatalLog("No clientIP and clientPort specified, starting local etcd server failed: %s", err)
		}
		etcdUrls = &etcdServer.Config.ClientUrls
		defer etcdServer.Stop()
	}
	objStore, err := GetEtcdClientBasic(*etcdUrls)
	if err != nil {
		util.FatalLog("Failed to initialize Object Store")
	}
	err = objStore.CheckConnected(50, 20*time.Millisecond)
	if err != nil {
		util.FatalLog("Failed to connect to etcd servers")
	}

	address := fmt.Sprintf("%s:%d", *apiIP, *apiPort)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		util.FatalLog("Failed to listen on address", "address", address,
			"error", err)
	}

	developerApi := InitDeveloperApi(objStore)
	if developerApi == nil {
		util.FatalLog("Failed to initialize developer API")
	}
	appApi := InitAppApi(objStore, developerApi)
	if appApi == nil {
		util.FatalLog("Failed to initialize app API")
	}
	operatorApi := InitOperatorApi(objStore)
	if operatorApi == nil {
		util.FatalLog("Failed to initialize operator API")
	}
	cloudletApi := InitCloudletApi(objStore, operatorApi)
	if cloudletApi == nil {
		util.FatalLog("Failed to initialize cloudlet API")
	}
	appInstApi := InitAppInstApi(objStore, appApi, cloudletApi)
	if appInstApi == nil {
		util.FatalLog("Failed to initialize app inst API")
	}
	notify.InitNotifySenders(NewControllerNotifier(appInstApi, cloudletApi))
	notify.RegisterMatcherAddrs(*matcherAddrs)
	notify.RegisterCloudletAddrs(*cloudletMgrAddrs)

	server := grpc.NewServer()
	edgeproto.RegisterDeveloperApiServer(server, developerApi)
	edgeproto.RegisterAppApiServer(server, appApi)
	edgeproto.RegisterOperatorApiServer(server, operatorApi)
	edgeproto.RegisterCloudletApiServer(server, cloudletApi)
	edgeproto.RegisterAppInstApiServer(server, appInstApi)

	go func() {
		// Serve will block until interrupted and Stop is called
		if err := server.Serve(lis); err != nil {
			util.FatalLog("Failed to serve", "error", err)
		}
	}()
	defer server.Stop()

	// REST gateway
	mux := http.NewServeMux()
	gw, err := grpcGateway(address)
	if err != nil {
		util.FatalLog("Failed to create grpc gateway", "error", err)
	}
	mux.Handle("/", gw)
	httpAddress := fmt.Sprintf("%s:%d", *apiIP, *httpPort)
	httpServer := &http.Server{
		Addr:    httpAddress,
		Handler: mux,
	}
	go func() {
		// Serve REST gateway
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			util.FatalLog("Failed to serve HTTP", "error", err)
		}
	}()
	defer httpServer.Shutdown(context.Background())

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	if mainStarted != nil {
		close(mainStarted)
	}
	util.InfoLog("Ready")

	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}
