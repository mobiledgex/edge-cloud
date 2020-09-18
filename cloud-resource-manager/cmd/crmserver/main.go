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
)

var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var notifySrvAddr = flag.String("notifySrvAddr", "127.0.0.1:51001", "Address for the CRM notify listener to run on")
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"DMUUS\"},\"name\":\"tmocloud1\"}'")
var physicalName = flag.String("physicalName", "", "Physical infrastructure cloudlet name, defaults to cloudlet name in cloudletKey")
var debugLevels = flag.String("d", "", fmt.Sprintf("Comma separated list of %v", log.DebugLevelStrings))
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
var commercialCerts = flag.Bool("commercialCerts", false, "Get TLS certs from LetsEncrypt. If false CRM will generate its own self-signed certs")
var appDNSRoot = flag.String("appDNSRoot", "mobiledgex.net", "App domain name root")
var chefServerPath = flag.String("chefServerPath", "", "Chef server path")
var deploymentTag = flag.String("deploymentTag", "", "Tag to indicate type of deployment setup. Ex: production, staging, etc")
var upgrade = flag.Bool("upgrade", false, "Flag to initiate upgrade run as part of crm bringup")

// myCloudletInfo is the information for the cloudlet in which the CRM is instantiated.
// The key for myCloudletInfo is provided as a configuration - either command line or
// from a file. The rest of the data is extraced from Openstack.
var myCloudletInfo edgeproto.CloudletInfo //XXX this effectively makes one CRM per cloudlet
var nodeMgr node.NodeMgr

var sigChan chan os.Signal
var mainStarted chan struct{}
var controllerData *crmutil.ControllerData
var notifyClient *notify.Client
var platform pf.Platform

const ControllerTimeout = 1 * time.Minute

func main() {
	nodeMgr.InitFlags()
	flag.Parse()

	if strings.Contains(*debugLevels, "mexos") {
		log.WarnLog("mexos log level is obsolete, please use infra")
		*debugLevels = strings.ReplaceAll(*debugLevels, "mexos", "infra")
	}
	log.SetDebugLevelStrs(*debugLevels)

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	standalone := false
	cloudcommon.ParseMyCloudletKey(standalone, cloudletKeyStr, &myCloudletInfo.Key)
	ctx, span, err := nodeMgr.Init(node.NodeTypeCRM, node.CertIssuerRegionalCloudlet, node.WithName(*hostname), node.WithCloudletKey(&myCloudletInfo.Key), node.WithNoUpdateMyNode(), node.WithRegion(*region), node.WithParentSpan(*parentSpan))
	if err != nil {
		log.FatalLog(err.Error())
	}
	defer nodeMgr.Finish()
	log.SetTags(span, myCloudletInfo.Key.GetTags())
	crmutil.InitDebug(&nodeMgr)

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
	platform, err = pfutils.GetPlatform(ctx, *platformName)
	if err != nil {
		span.Finish()
		log.FatalLog(err.Error())
	}

	controllerData = crmutil.NewControllerData(platform, &nodeMgr)

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
	notifyClientTls, err := nodeMgr.InternalPki.GetClientTlsConfig(ctx,
		nodeMgr.CommonName(),
		node.CertIssuerRegionalCloudlet,
		[]node.MatchCA{node.SameRegionalMatchCA()})
	if err != nil {
		log.FatalLog(err.Error())
	}
	notifyServerTls, err := nodeMgr.InternalPki.GetServerTlsConfig(ctx,
		nodeMgr.CommonName(),
		node.CertIssuerRegionalCloudlet,
		[]node.MatchCA{node.SameRegionalCloudletMatchCA()})
	if err != nil {
		log.FatalLog(err.Error())
	}
	dialOption := tls.GetGrpcDialOption(notifyClientTls)
	notifyClient = notify.NewClient(nodeMgr.Name(), addrs, dialOption)
	notifyClient.SetFilterByCloudletKey()
	InitClientNotify(notifyClient, controllerData)
	notifyClient.Start()
	defer notifyClient.Stop()

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

		caches := pf.Caches{
			FlavorCache:        &controllerData.FlavorCache,
			PrivacyPolicyCache: &controllerData.PrivacyPolicyCache,
			ClusterInstCache:   &controllerData.ClusterInstCache,
			AppCache:           &controllerData.AppCache,
			AppInstCache:       &controllerData.AppInstCache,
			ResTagTableCache:   &controllerData.ResTagTableCache,
			CloudletCache:      &controllerData.CloudletCache,
			VMPoolCache:        &controllerData.VMPoolCache,
			VMPoolInfoCache:    &controllerData.VMPoolInfoCache,
		}

		if cloudlet.PlatformType == edgeproto.PlatformType_PLATFORM_TYPE_VM_POOL {
			if cloudlet.VmPool == "" {
				log.FatalLog("Cloudlet is missing VM pool name")
			}
			vmPoolKey := edgeproto.VMPoolKey{
				Name:         cloudlet.VmPool,
				Organization: myCloudletInfo.Key.Organization,
			}
			var vmPool edgeproto.VMPool
			if !controllerData.VMPoolCache.Get(&vmPoolKey, &vmPool) {
				log.FatalLog("failed to fetch vm pool cache from controller")
			}
			controllerData.VMPool = vmPool
			caches.VMPool = &controllerData.VMPool
			caches.VMPoolMux = &controllerData.VMPoolMux
			// Update VMPool Info, this is to notify shepherd about VMPool
			controllerData.UpdateVMPoolInfo(ctx, edgeproto.TrackedState_READY, "")
		}

		updateCloudletStatus(edgeproto.UpdateTask, "Initializing platform")
		if err = initPlatform(ctx, &cloudlet, &myCloudletInfo, *physicalName, nodeMgr.VaultAddr, &caches, updateCloudletStatus); err != nil {
			myCloudletInfo.Errors = append(myCloudletInfo.Errors, fmt.Sprintf("Failed to init platform: %v", err))
			myCloudletInfo.State = edgeproto.CloudletState_CLOUDLET_STATE_ERRORS
		} else {
			log.SpanLog(ctx, log.DebugLevelInfo, "gathering cloudlet info")
			updateCloudletStatus(edgeproto.UpdateTask, "Gathering Cloudlet Info")
			err = controllerData.GatherCloudletInfo(ctx, &myCloudletInfo)
			log.SpanLog(ctx, log.DebugLevelInfra, "GatherCloudletInfo done", "state", myCloudletInfo.State)

			if err != nil {
				myCloudletInfo.Errors = append(myCloudletInfo.Errors, err.Error())
				myCloudletInfo.State = edgeproto.CloudletState_CLOUDLET_STATE_ERRORS
			} else {
				myCloudletInfo.State = edgeproto.CloudletState_CLOUDLET_STATE_NEED_SYNC
				log.SpanLog(ctx, log.DebugLevelInfra, "cloudlet needs sync data", "state", myCloudletInfo.State, "myCloudletInfo", myCloudletInfo)
				controllerData.ControllerSyncInProgress = true
				controllerData.CloudletInfoCache.Update(ctx, &myCloudletInfo, 0)

				// Wait for CRM to receive cluster and appinst data from notify
				select {
				case <-controllerData.ControllerSyncDone:
					if !controllerData.CloudletCache.Get(&myCloudletInfo.Key, &cloudlet) {
						log.FatalLog("failed to get sync data from controller")
					}
				case <-time.After(ControllerTimeout):
					log.FatalLog("Timed out waiting for sync data from controller")
				}
				log.SpanLog(ctx, log.DebugLevelInfra, "controller sync data received")
				myCloudletInfo.ControllerCacheReceived = true
				controllerData.CloudletInfoCache.Update(ctx, &myCloudletInfo, 0)
				err := platform.SyncControllerCache(ctx, &caches, myCloudletInfo.State)
				if err != nil {
					log.FatalLog("Platform sync fail", "err", err)
				}
				myCloudletInfo.Errors = nil
				myCloudletInfo.State = edgeproto.CloudletState_CLOUDLET_STATE_READY
				log.SpanLog(ctx, log.DebugLevelInfra, "cloudlet state", "state", myCloudletInfo.State, "myCloudletInfo", myCloudletInfo)
			}
		}

		log.SpanLog(ctx, log.DebugLevelInfo, "sending cloudlet info cache update")
		// trigger send of info upstream to controller
		controllerData.CloudletInfoCache.Update(ctx, &myCloudletInfo, 0)

		nodeMgr.MyNode.ContainerVersion = cloudletContainerVersion
		nodeMgr.UpdateMyNode(ctx)
		log.SpanLog(ctx, log.DebugLevelInfo, "sent cloudletinfocache update")
		cspan.Finish()

		if err != nil {
			// die so CRM can restart and try again
			log.FatalLog("Platform init fail", "err", err)
		}

		// setup rootlb certs
		tlsSpan := log.StartSpan(log.DebugLevelInfo, "tls certs thread", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
		commonName := cloudcommon.GetRootLBFQDN(&myCloudletInfo.Key, *appDNSRoot)
		dedicatedCommonName := "*." + commonName // wildcard so dont have to generate certs every time a dedicated cluster is started
		rootlb, err := platform.GetClusterPlatformClient(
			ctx,
			&edgeproto.ClusterInst{
				IpAccess: edgeproto.IpAccess_IP_ACCESS_SHARED,
			},
			cloudcommon.ClientTypeRootLB,
		)
		if err == nil {
			proxy.GetRootLbCerts(ctx, commonName, dedicatedCommonName, nodeMgr.VaultAddr, rootlb, *commercialCerts)
		}
		tlsSpan.Finish()
	}()

	// setup crm notify listener (for shepherd)
	var notifyServer notify.ServerMgr
	initSrvNotify(&notifyServer)
	notifyServer.Start(nodeMgr.Name(), *notifySrvAddr, notifyServerTls)
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
func initPlatform(ctx context.Context, cloudlet *edgeproto.Cloudlet, cloudletInfo *edgeproto.CloudletInfo, physicalName, vaultAddr string, caches *pf.Caches, updateCallback edgeproto.CacheUpdateCallback) error {
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
		NodeMgr:             &nodeMgr,
		AppDNSRoot:          *appDNSRoot,
		ChefServerPath:      *chefServerPath,
		DeploymentTag:       *deploymentTag,
		Upgrade:             *upgrade,
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "init platform", "location(cloudlet.key.name)", loc, "operator", oper, "Platform type", platform.GetType())
	err := platform.Init(ctx, &pc, caches, updateCallback)
	return err
}
