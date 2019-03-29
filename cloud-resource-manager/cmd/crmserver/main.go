package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/azure"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/gcp"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/mexdind"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/openstack"
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
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"TMUS\"},\"name\":\"tmocloud1\"}'")
var standalone = flag.Bool("standalone", false, "Standalone mode. CRM does not interact with controller. Cloudlet/AppInsts can be created directly on CRM using controller API")
var fakecloudlet = flag.Bool("fakecloudlet", false, "Fake cloudlet mode.  A fake cloudlet is reported to the controller")
var debugLevels = flag.String("d", "", fmt.Sprintf("Comma separated list of %v", log.DebugLevelStrings))
var tlsCertFile = flag.String("tls", "", "server tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory")
var hostname = flag.String("hostname", "", "Unique hostname within Cloudlet")
var platformName = flag.String("platform", "", "Platform type of Cloudlet")

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

	cloudcommon.ParseMyCloudletKey(*standalone, cloudletKeyStr, &myCloudlet.Key)
	cloudcommon.SetNodeKey(hostname, edgeproto.NodeType_NodeCRM, &myCloudlet.Key, &myNode.Key)
	if *standalone {
		*fakecloudlet = true
	}
	if *fakecloudlet {
		*platformName = "fakecloudlet"
	}
	if *platformName == "" {
		// see if env var was set
		*platformName = os.Getenv("PLATFORM")
	}
	if *platformName == "" {
		// if not specified, platform is derived from operator name
		*platformName = myCloudlet.Key.OperatorKey.Name
	}
	log.DebugLog(log.DebugLevelMexos, "Using cloudletKey", "key", myCloudlet.Key, "platform", *platformName)

	// Get platform implementation.
	// this is a switch for now, but eventually will load plugin modules
	// based on platform name so we can disentangle edge-cloud from
	// edge-cloud-infra.
	switch *platformName {
	case "fakecloudlet":
		platform = &fake.Platform{}
	case "dind":
		platform = &dind.Platform{}
	case "mexdind":
		platform = &mexdind.Platform{}
	case "gcp":
		platform = &gcp.Platform{}
	case "azure":
		platform = &azure.Platform{}
	default:
		platform = &openstack.Platform{}
	}

	listener, err := net.Listen("tcp", *bindAddress)
	if err != nil {
		log.FatalLog("Failed to bind", "addr", *bindAddress, "err", err)
	}
	controllerData = crmutil.NewControllerData(platform)

	creds, err := tls.GetTLSServerCreds(*tlsCertFile)
	if err != nil {
		log.FatalLog("get TLS Credentials", "error", err)
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	platChan := make(chan string)

	go func() {
		log.DebugLog(log.DebugLevelMexos, "starting to init platform")
		if err := initPlatform(&myCloudlet); err != nil {
			log.FatalLog("failed to init platform", "err", err)
		}
		log.DebugLog(log.DebugLevelMexos, "send status on plat channel")
		platChan <- "ready"
	}()
	// GatherInsts should be called before the notify client is started,
	// so that the initial send to the controller has the current state.
	controllerData.GatherInsts()

	if *standalone {
		// In standalone mode, use "touch allownoconfig" for edgectl
		// to set no-config fields like "flavor" on CreateClusterInst.
		log.InfoLog("Running in Standalone mode")
		saServer := standaloneServer{data: controllerData}
		edgeproto.RegisterCloudletApiServer(grpcServer, &saServer)
		edgeproto.RegisterFlavorApiServer(grpcServer, &saServer)
		edgeproto.RegisterClusterInstApiServer(grpcServer, &saServer)
		edgeproto.RegisterAppInstApiServer(grpcServer, &saServer)
		edgeproto.RegisterAppInstInfoApiServer(grpcServer, &saServer)
		edgeproto.RegisterClusterInstInfoApiServer(grpcServer, &saServer)
		edgeproto.RegisterCloudletInfoApiServer(grpcServer, &saServer)
	} else {
		addrs := strings.Split(*notifyAddrs, ",")
		notifyClient = notify.NewClient(addrs, *tlsCertFile)
		notifyClient.SetFilterByCloudletKey()
		InitNotify(notifyClient, controllerData)
		notifyClient.Start()
		defer notifyClient.Stop()
	}
	reflection.Register(grpcServer)

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
	log.DebugLog(log.DebugLevelMexos, "gather cloudlet info")

	go func() {
		log.DebugLog(log.DebugLevelMexos, "wait for status on plat channel")
		platStat := <-platChan
		log.DebugLog(log.DebugLevelMexos, "got status on plat channel", "status", platStat)
		controllerData.GatherCloudletInfo(&myCloudlet)
		log.DebugLog(log.DebugLevelMexos, "sending cloudlet info cache update")
		// trigger send of info upstream to controller
		controllerData.NodeCache.Update(&myNode, 0)
		log.DebugLog(log.DebugLevelMexos, "sent cloudletinfocache update")
	}()
	if mainStarted != nil {
		// for unit testing to detect when main is ready
		close(mainStarted)
	}

	<-sigChan
	os.Exit(0)
}

//initializePlatform *Must be called as a seperate goroutine.*
func initPlatform(cloudlet *edgeproto.CloudletInfo) error {
	loc := util.DNSSanitize(cloudlet.Key.Name) //XXX  key.name => loc
	oper := util.DNSSanitize(cloudlet.Key.OperatorKey.Name)
	//if err := mexos.FillManifestValues(mf, "platform"); err != nil {
	//	return err
	//}
	log.DebugLog(log.DebugLevelMexos, "init platform", "location(cloudlet.key.name)", loc, "operator", oper)
	err := platform.Init(&cloudlet.Key)
	return err
}
