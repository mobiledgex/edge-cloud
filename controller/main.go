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

	pb "github.com/mobiledgex/edge-cloud/proto"
	"github.com/mobiledgex/edge-cloud/util"
	"google.golang.org/grpc"
)

// Command line options
var rootDir = flag.String("r", "", "root directory; set for testing")
var localEtcd = flag.Bool("localEtcd", false, "set to start local etcd for testing")
var region = flag.Uint("region", 1, "Region")
var clientIP = flag.String("clientIP", "127.0.0.1", "client listener port")
var clientPort = flag.Uint("clientPort", 2380, "client listener port")
var apiPort = flag.Uint("apiPort", 55001, "API listener port")
var httpPort = flag.Uint("httpPort", 8091, "HTTP listener port")

func GetRootDir() string {
	return *rootDir
}

var sigChan chan os.Signal

// testing hook
var mainStarted chan struct{}

func main() {
	flag.Parse()

	util.InfoLog("Start up", "rootDir", *rootDir, "apiPort", *apiPort)
	InitRegion(uint32(*region))

	if *localEtcd {
		etcdServer, err := StartLocalEtcdServer()
		if err != nil {
			util.FatalLog("No clientIP and clientPort specified, starting local etcd server failed: %s", err)
		}
		clientIP = &etcdServer.Config.ClientIP
		clientPort = &etcdServer.Config.ClientPort
		defer etcdServer.Stop()
	}
	objStore, err := GetEtcdClientBasic(*clientIP, *clientPort)
	if err != nil {
		util.FatalLog("Failed to initialize Object Store")
	}

	address := fmt.Sprintf("%s:%d", *clientIP, *apiPort)
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
	InitNotifySenders(appInstApi, cloudletApi)

	server := grpc.NewServer()
	pb.RegisterDeveloperApiServer(server, developerApi)
	pb.RegisterAppApiServer(server, appApi)
	pb.RegisterOperatorApiServer(server, operatorApi)
	pb.RegisterCloudletApiServer(server, cloudletApi)

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
	httpAddress := fmt.Sprintf("%s:%d", *clientIP, *httpPort)
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

	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}
