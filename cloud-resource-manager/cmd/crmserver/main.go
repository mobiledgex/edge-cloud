package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	pfutils "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/utils"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/mobiledgex/edge-cloud/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var controllerAddress = flag.String("controller", "127.0.0.1:55001", "Address of controller API")
var vaultAddr = flag.String("vaultAddr", "", "Address to vault")
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var notifySrvAddr = flag.String("notifySrvAddr", "127.0.0.1:51001", "Address for the CRM notify listener to run on")
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"TMUS\"},\"name\":\"tmocloud1\"}'")
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
	log.InitTracer()
	defer log.FinishTracer()
	span := log.StartSpan(log.DebugLevelInfo, "main")
	ctx := log.ContextWithSpan(context.Background(), span)

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
	log.SpanLog(ctx, log.DebugLevelInfo, "Using cloudletKey", "key", myCloudlet.Key, "platform", *platformName, "physicalName", physicalName)

	// Load platform implementation.
	var err error
	platform, err = pfutils.GetPlatform(ctx, *platformName)
	if err != nil {
		log.FatalLog(err.Error())
	}

	controllerData = crmutil.NewControllerData(platform)

	creds, err := tls.GetTLSServerCreds(*tlsCertFile, true)
	if err != nil {
		span.Finish()
		log.FatalLog("get TLS Credentials", "error", err)
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	go func() {
		log.SpanLog(ctx, log.DebugLevelInfo, "starting to init platform")
		if err := initPlatform(&myCloudlet, *physicalName, *vaultAddr, &controllerData.ClusterInstInfoCache); err != nil {
			span.Finish()
			log.FatalLog("failed to init platform", "err", err)
		}

		log.SpanLog(ctx, log.DebugLevelInfo, "gathering cloudlet info")
		controllerData.GatherCloudletInfo(&myCloudlet)

		log.SpanLog(ctx, log.DebugLevelInfo, "sending cloudlet info cache update")
		// trigger send of info upstream to controller
		controllerData.CloudletInfoCache.Update(ctx, &myCloudlet, 0)

		myNode.BuildMaster = version.BuildMaster
		myNode.BuildHead = version.BuildHead
		myNode.BuildAuthor = version.BuildAuthor
		myNode.Hostname = cloudcommon.Hostname()
		controllerData.NodeCache.Update(ctx, &myNode, 0)
		log.SpanLog(ctx, log.DebugLevelInfo, "sent cloudletinfocache update")
	}()

	//ctl notify
	addrs := strings.Split(*notifyAddrs, ",")
	notifyClient = notify.NewClient(addrs, *tlsCertFile)
	notifyClient.SetFilterByCloudletKey()
	InitClientNotify(notifyClient, controllerData)
	notifyClient.Start()
	defer notifyClient.Stop()
	reflection.Register(grpcServer)

	//setup crm notify listener (for shepherd)
	var notifyServer notify.ServerMgr
	notifyServer.Init()
	initSrvNotify(&notifyServer)
	notifyServer.Start(*notifySrvAddr, *tlsCertFile)
	defer notifyServer.Stop()

	dialOption, err := tls.GetTLSClientDialOption(*controllerAddress, *tlsCertFile, false)
	if err != nil {
		span.Finish()
		log.FatalLog("Failed get TLS options", "error", err)
		os.Exit(1)
	}
	conn, err := grpc.Dial(*controllerAddress, dialOption)
	if err != nil {
		span.Finish()
		log.FatalLog("Failed to connect to controller",
			"addr", *controllerAddress, "err", err)
		os.Exit(1)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			span.Finish()
			log.FatalLog("Failed to close connection", "error", err)
			os.Exit(1)
		}
	}()
	span.Finish()

	if mainStarted != nil {
		// for unit testing to detect when main is ready
		close(mainStarted)
	}

	<-sigChan
	os.Exit(0)
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
