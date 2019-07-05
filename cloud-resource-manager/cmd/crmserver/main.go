package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"plugin"
	"strings"
	"syscall"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/dind"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/fake"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var bindAddress = flag.String("apiAddr", "0.0.0.0:55099", "Address to bind")
var controllerAddress = flag.String("controller", "127.0.0.1:55001", "Address of controller API")
var vaultAddr = flag.String("vaultAddr", "", "Address to vault")
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var notifySrvAddr = flag.String("notifySrvAddr", "127.0.0.1:51001", "Address for the CRM notify listener to run on")
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"DMUUS\"},\"name\":\"tmocloud1\"}'")
var physicalName = flag.String("physicalName", "", "Physical infrastructure cloudlet name, defaults to cloudlet name in cloudletKey")
var debugLevels = flag.String("d", "", fmt.Sprintf("Comma separated list of %v", log.DebugLevelStrings))
var tlsCertFile = flag.String("tls", "", "server tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory")
var hostname = flag.String("hostname", "", "Unique hostname within Cloudlet")
var platformName = flag.String("platform", "", "Platform type of Cloudlet")
var solib = flag.String("plugin", "", "plugin file")
var region = flag.String("region", "local", "region name")

// myCloudlet is the information for the cloudlet in which the CRM is instantiated.
// The key for myCloudlet is provided as a configuration - either command line or
// from a file. The rest of the data is extraced from Openstack.
var myCloudlet edgeproto.CloudletInfo //XXX this effectively makes one CRM per cloudlet
var myNode edgeproto.Node

var sigChan chan os.Signal
var mainStarted chan struct{}
var controllerData *crmutil.ControllerData
var notifyClient *notify.Client
var platform pf.Platform

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	standalone := false
	cloudcommon.ParseMyCloudletKey(standalone, cloudletKeyStr, &myCloudlet.Key)
	cloudcommon.SetNodeKey(hostname, edgeproto.NodeType_NODE_CRM, &myCloudlet.Key, &myNode.Key)
	if *platformName == "" {
		// see if env var was set
		*platformName = os.Getenv("PLATFORM")
	}
	if *platformName == "" {
		// if not specified, platform is derived from operator name
		*platformName = myCloudlet.Key.OperatorKey.Name
	}
	if *physicalName == "" {
		*physicalName = myCloudlet.Key.Name
	}
	log.DebugLog(log.DebugLevelMexos, "Using cloudletKey", "key", myCloudlet.Key, "platform", *platformName, "physicalName", physicalName)

	// Load platform implementation.
	var err error
	platform, err = getPlatform(*platformName)
	if err != nil {
		log.FatalLog("Failed to get platform %s, %s", *platformName, err.Error())
	}

	// Start Communictions.
	listener, err := net.Listen("tcp", *bindAddress)
	if err != nil {
		log.FatalLog("Failed to bind", "addr", *bindAddress, "err", err)
	}
	controllerData = crmutil.NewControllerData(platform)

	creds, err := tls.GetTLSServerCreds(*tlsCertFile, true)
	if err != nil {
		log.FatalLog("get TLS Credentials", "error", err)
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	go func() {
		log.DebugLog(log.DebugLevelMexos, "starting to init platform")
		if err := initPlatform(&myCloudlet, *physicalName, *vaultAddr, &controllerData.ClusterInstInfoCache); err != nil {
			log.FatalLog("failed to init platform", "err", err)
		}

		log.DebugLog(log.DebugLevelMexos, "gathering cloudlet info")
		controllerData.GatherCloudletInfo(&myCloudlet)

		log.DebugLog(log.DebugLevelMexos, "sending cloudlet info cache update")
		// trigger send of info upstream to controller
		controllerData.CloudletInfoCache.Update(&myCloudlet, 0)
		controllerData.NodeCache.Update(&myNode, 0)
		log.DebugLog(log.DebugLevelMexos, "sent cloudletinfocache update")
	}()

	//ctl notify
	addrs := strings.Split(*notifyAddrs, ",")
	notifyClient = notify.NewClient(addrs, *tlsCertFile)
	notifyClient.SetFilterByCloudletKey()
	InitNotify(notifyClient, controllerData)
	notifyClient.Start()
	defer notifyClient.Stop()
	reflection.Register(grpcServer)

	//setup crm notify listener (for shepherd)
	var notifyServer notify.ServerMgr
	notifyServer.Init()
	initCrmNotify(&notifyServer)
	notifyServer.Start(*notifySrvAddr, *tlsCertFile)
	defer notifyServer.Stop()

	go func() {
		if err = grpcServer.Serve(listener); err != nil {
			log.FatalLog("Failed to serve grpc", "err", err)
		}
	}()
	defer grpcServer.Stop()

	log.InfoLog("Server started", "addr", *bindAddress)
	dialOption, err := tls.GetTLSClientDialOption(*controllerAddress, *tlsCertFile)
	if err != nil {
		log.FatalLog("Failed get TLS options", "error", err)
		os.Exit(1)
	}
	conn, err := grpc.Dial(*controllerAddress, dialOption)
	if err != nil {
		log.FatalLog("Failed to connect to controller",
			"addr", *controllerAddress, "err", err)
		os.Exit(1)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.FatalLog("Failed to close connection", "error", err)
			os.Exit(1)
		}
	}()

	if mainStarted != nil {
		// for unit testing to detect when main is ready
		close(mainStarted)
	}

	<-sigChan
	os.Exit(0)
}

func getPlatform(plat string) (pf.Platform, error) {
	// Building plugins is slow, so directly importable
	// platforms are not built as plugins.
	if plat == "dind" {
		return &dind.Platform{}, nil
	} else if plat == "fakecloudlet" {
		return &fake.Platform{}, nil
	}

	// Load platform from plugin
	if *solib == "" {
		*solib = os.Getenv("GOPATH") + "/plugins/platforms.so"
	}
	log.DebugLog(log.DebugLevelMexos, "Loading plugin", "plugin", *solib)
	plug, err := plugin.Open(*solib)
	if err != nil {
		log.FatalLog("failed to load plugin", "plugin", *solib, "error", err)
	}
	sym, err := plug.Lookup("GetPlatform")
	if err != nil {
		log.FatalLog("plugin does not have GetPlatform symbol", "plugin", *solib)
	}
	getPlatFunc, ok := sym.(func(plat string) (pf.Platform, error))
	if !ok {
		log.FatalLog("plugin GetPlatform symbol does not implement func(plat string) (platform.Platform, error)", "plugin", *solib)
	}
	log.DebugLog(log.DebugLevelMexos, "Creating platform")
	return getPlatFunc(*platformName)
}

//initializePlatform *Must be called as a seperate goroutine.*
func initPlatform(cloudlet *edgeproto.CloudletInfo, physicalName, vaultAddr string, clusterInstCache *edgeproto.ClusterInstInfoCache) error {
	loc := util.DNSSanitize(cloudlet.Key.Name) //XXX  key.name => loc
	oper := util.DNSSanitize(cloudlet.Key.OperatorKey.Name)

	pc := pf.PlatformConfig{
		CloudletKey:  &cloudlet.Key,
		PhysicalName: physicalName,
		VaultAddr:    vaultAddr}
	log.DebugLog(log.DebugLevelMexos, "init platform", "location(cloudlet.key.name)", loc, "operator", oper)
	err := platform.Init(&pc)
	return err
}

//shepherd only needs these two for now, will need to be able to recieve metrics later as well
func initCrmNotify(notifyServer *notify.ServerMgr) {
	notifyServer.RegisterSendClusterInstCache(&controllerData.ClusterInstCache)
	notifyServer.RegisterSendAppInstCache(&controllerData.AppInstCache)
}
