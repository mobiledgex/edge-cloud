package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	pfutils "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/utils"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/util"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

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
var testMode = flag.Bool("testMode", false, "Run CRM in test mode")
var parentSpan = flag.String("span", "", "Use parent span for logging")
var containerVersion = flag.String("containerVersion", "", "edge-cloud container version")
var vmImageVersion = flag.String("vmImageVersion", "", "CRM VM baseimage version")
var packageVersion = flag.String("packageVersion", "", "CRM VM baseimage debian package version")
var cloudletVMImagePath = flag.String("cloudletVMImagePath", "", "Image path where CRM VM baseimages are present")
var cleanupMode = flag.Bool("cleanupMode", false, "cleanup previous versions of CRM if present")
var commercialCerts = flag.Bool("commercialCerts", false, "Get TLS certs from LetsEncrypt. If false CRM will generate its own self-signed certs")

// myCloudletInfo is the information for the cloudlet in which the CRM is instantiated.
// The key for myCloudletInfo is provided as a configuration - either command line or
// from a file. The rest of the data is extraced from Openstack.
var myCloudletInfo edgeproto.CloudletInfo //XXX this effectively makes one CRM per cloudlet
var nodeMgr *node.NodeMgr

var sigChan chan os.Signal
var mainStarted chan struct{}
var controllerData *crmutil.ControllerData
var notifyClient *notify.Client
var platform pf.Platform

const ControllerTimeout = 1 * time.Minute

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	log.InitTracer(*tlsCertFile)
	defer log.FinishTracer()

	var span opentracing.Span
	if *parentSpan != "" {
		span = log.NewSpanFromString(log.DebugLevelInfo, *parentSpan, "main")
	} else {
		span = log.StartSpan(log.DebugLevelInfo, "main")
	}
	ctx := log.ContextWithSpan(context.Background(), span)

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	standalone := false
	cloudcommon.ParseMyCloudletKey(standalone, cloudletKeyStr, &myCloudletInfo.Key)
	nodeMgr = node.Init(ctx, node.NodeTypeCRM, node.WithName(*hostname), node.WithCloudletKey(&myCloudletInfo.Key), node.WithNoUpdateMyNode())
	if *platformName == "" {
		// see if env var was set
		*platformName = os.Getenv("PLATFORM")
	}
	if *platformName == "" {
		// if not specified, platform is derived from operator name
		*platformName = myCloudletInfo.Key.Organization
	}
	if *physicalName == "" {
		*physicalName = myCloudletInfo.Key.Name
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "Using cloudletKey", "key", myCloudletInfo.Key, "platform", *platformName, "physicalName", physicalName)

	// Load platform implementation.
	var err error
	platform, err = pfutils.GetPlatform(ctx, *platformName)
	if err != nil {
		span.Finish()
		log.FatalLog(err.Error())
	}

	controllerData = crmutil.NewControllerData(platform)

	creds, err := tls.GetTLSServerCreds(*tlsCertFile, true)
	if err != nil {
		span.Finish()
		log.FatalLog("get TLS Credentials", "error", err)
	}

	updateCloudletStatus := func(updateType edgeproto.CacheUpdateType, value string) {
		switch updateType {
		case edgeproto.UpdateTask:
			myCloudletInfo.Status.SetTask(value)
		case edgeproto.UpdateStep:
			myCloudletInfo.Status.SetStep(value)
		}
		controllerData.CloudletInfoCache.Update(ctx, &myCloudletInfo, 0)
	}

	//ctl notify
	addrs := strings.Split(*notifyAddrs, ",")
	notifyClient = notify.NewClient(addrs, *tlsCertFile)
	notifyClient.SetFilterByCloudletKey()
	InitClientNotify(notifyClient, controllerData)
	notifyClient.Start()
	defer notifyClient.Stop()

	grpcServer := grpc.NewServer(grpc.Creds(creds))
	reflection.Register(grpcServer)

	go func() {
		cspan := log.StartSpan(log.DebugLevelInfo, "cloudlet init thread", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
		log.SpanLog(ctx, log.DebugLevelInfo, "starting to init platform")

		cloudletContainerVersion := ""
		if *containerVersion == "" {
			cloudletContainerVersion, err = cloudcommon.GetDockerBaseImageVersion()
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "unable to fetch docker image version", "err", err)
			}
		} else {
			cloudletContainerVersion = *containerVersion
		}

		myCloudletInfo.State = edgeproto.CloudletState_CLOUDLET_STATE_INIT
		myCloudletInfo.ContainerVersion = cloudletContainerVersion
		controllerData.CloudletInfoCache.Update(ctx, &myCloudletInfo, 0)

		var cloudlet edgeproto.Cloudlet
		log.SpanLog(ctx, log.DebugLevelInfo, "wait for cloudlet cache", "key", myCloudletInfo.Key)
		// Wait for cloudlet cache from controller
		// This ensures that crm is able to communicate to controller via Notify Channel
		select {
		case <-controllerData.ControllerWait:
			if !controllerData.CloudletCache.Get(&myCloudletInfo.Key, &cloudlet) {
				log.FatalLog("failed to fetch cloudlet cache from controller")
			}
		case <-time.After(ControllerTimeout):
			log.FatalLog("Timed out waiting for cloudlet cache from controller")
		}
		log.SpanLog(ctx, log.DebugLevelInfo, "fetched cloudlet cache from controller", "cloudlet", cloudlet)

		updateCloudletStatus(edgeproto.UpdateTask, "Initializing platform")
		if err := initPlatform(ctx, &cloudlet, &myCloudletInfo, *physicalName, *vaultAddr, &controllerData.ClusterInstInfoCache, updateCloudletStatus); err != nil {
			myCloudletInfo.Errors = append(myCloudletInfo.Errors, fmt.Sprintf("Failed to init platform: %v", err))
			myCloudletInfo.State = edgeproto.CloudletState_CLOUDLET_STATE_ERRORS
		} else {
			log.SpanLog(ctx, log.DebugLevelInfo, "gathering cloudlet info")
			updateCloudletStatus(edgeproto.UpdateTask, "Gathering Cloudlet Info")
			err = controllerData.GatherCloudletInfo(ctx, &myCloudletInfo)

			if err != nil {
				myCloudletInfo.Errors = append(myCloudletInfo.Errors, err.Error())
				myCloudletInfo.State = edgeproto.CloudletState_CLOUDLET_STATE_ERRORS
			} else {
				myCloudletInfo.Errors = nil
				if *cleanupMode {
					controllerData.CleanupOldCloudlet(ctx, &cloudlet, updateCloudletStatus)
				}
				myCloudletInfo.State = edgeproto.CloudletState_CLOUDLET_STATE_READY
				log.SpanLog(ctx, log.DebugLevelMexos, "cloudlet state", "state", myCloudletInfo.State, "myCloudletInfo", myCloudletInfo)
			}
		}

		log.SpanLog(ctx, log.DebugLevelInfo, "sending cloudlet info cache update")
		// trigger send of info upstream to controller
		controllerData.CloudletInfoCache.Update(ctx, &myCloudletInfo, 0)

		nodeMgr.MyNode.ContainerVersion = cloudletContainerVersion
		nodeMgr.UpdateMyNode(ctx)
		log.SpanLog(ctx, log.DebugLevelInfo, "sent cloudletinfocache update")
		cspan.Finish()

		// setup rootlb certs
		tlsSpan := log.StartSpan(log.DebugLevelInfo, "tls certs thread", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
		commonName := cloudcommon.GetRootLBFQDN(&myCloudletInfo.Key)
		dedicatedCommonName := "*." + commonName // wildcard so dont have to generate certs every time a dedicated cluster is started
		rootlb, err := platform.GetPlatformClient(ctx, &edgeproto.ClusterInst{IpAccess: edgeproto.IpAccess_IP_ACCESS_SHARED})
		if err == nil {
			proxy.GetRootLbCerts(ctx, commonName, dedicatedCommonName, *vaultAddr, rootlb, *commercialCerts)
		}
		tlsSpan.Finish()
	}()

	// setup crm notify listener (for shepherd)
	var notifyServer notify.ServerMgr
	initSrvNotify(&notifyServer)
	notifyServer.Start(*notifySrvAddr, *tlsCertFile)
	defer notifyServer.Stop()

	span.Finish()

	if mainStarted != nil {
		// for unit testing to detect when main is ready
		close(mainStarted)
	}

	<-sigChan
	os.Exit(0)
}

//initializePlatform *Must be called as a seperate goroutine.*
func initPlatform(ctx context.Context, cloudlet *edgeproto.Cloudlet, cloudletInfo *edgeproto.CloudletInfo, physicalName, vaultAddr string, clusterInstCache *edgeproto.ClusterInstInfoCache, updateCallback edgeproto.CacheUpdateCallback) error {
	loc := util.DNSSanitize(cloudletInfo.Key.Name) //XXX  key.name => loc
	oper := util.DNSSanitize(cloudletInfo.Key.Organization)

	pc := pf.PlatformConfig{
		CloudletKey:         &cloudletInfo.Key,
		PhysicalName:        physicalName,
		VaultAddr:           vaultAddr,
		Region:              *region,
		TestMode:            *testMode,
		CloudletVMImagePath: *cloudletVMImagePath,
		VMImageVersion:      *vmImageVersion,
		PackageVersion:      *packageVersion,
		EnvVars:             cloudlet.EnvVar,
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "init platform", "location(cloudlet.key.name)", loc, "operator", oper, "Platform", pc)
	err := platform.Init(ctx, &pc, updateCallback)
	return err
}
