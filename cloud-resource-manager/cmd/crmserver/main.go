package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/accessapi"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	pfutils "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/utils"
	proxycerts "github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy/certs"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/util"
	opentracing "github.com/opentracing/opentracing-go"
)

var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var notifySrvAddr = flag.String("notifySrvAddr", "127.0.0.1:51001", "Address for the CRM notify listener to run on")
var notifyServer notify.ServerMgr
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"TMUS\"},\"name\":\"tmocloud1\"}'")
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
var upgrade = flag.Bool("upgrade", false, "Flag to initiate upgrade run as part of crm bringup")
var cacheDir = flag.String("cacheDir", "/tmp/", "Cache used by CRM to store frequently accessed data")

// myCloudletInfo is the information for the cloudlet in which the CRM is instantiated.
// The key for myCloudletInfo is provided as a configuration - either command line or
// from a file. The rest of the data is extraced from Openstack.
var myCloudletInfo edgeproto.CloudletInfo //XXX this effectively makes one CRM per cloudlet
var nodeMgr node.NodeMgr

var sigChan chan os.Signal
var services Services
var mainStarted chan struct{}
var controllerData *crmutil.ControllerData
var notifyClient *notify.Client
var platform pf.Platform

const ControllerTimeout = 1 * time.Minute

const (
	envMexBuild       = "MEX_BUILD"
	envMexBuildTag    = "MEX_BUILD_TAG"
	envMexBuildFlavor = "MEX_BUILD_FLAVOR"
)

type Services struct {
	notifyClientRunning          bool
	notifyServerRunning          bool
	infraResourcesRefreshRunning bool
	flavorRefreshRunning         bool
}

func main() {

	nodeMgr.InitFlags()
	nodeMgr.AccessKeyClient.InitFlags()
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

	if strings.Contains(*debugLevels, "mexos") {
		log.WarnLog("mexos log level is obsolete, please use infra")
		*debugLevels = strings.ReplaceAll(*debugLevels, "mexos", "infra")
	}
	log.SetDebugLevelStrs(*debugLevels)

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	standalone := false
	cloudcommon.ParseMyCloudletKey(standalone, cloudletKeyStr, &myCloudletInfo.Key)
	myCloudletInfo.CompatibilityVersion = cloudcommon.GetCRMCompatibilityVersion()
	ctx, span, err := nodeMgr.Init(node.NodeTypeCRM, node.CertIssuerRegionalCloudlet, node.WithName(*hostname), node.WithCloudletKey(&myCloudletInfo.Key), node.WithNoUpdateMyNode(), node.WithRegion(*region), node.WithParentSpan(*parentSpan))
	if err != nil {
		return err
	}
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
	platform, err = pfutils.GetPlatform(ctx, *platformName, nodeMgr.UpdateNodeProps)
	if err != nil {
		span.Finish()
		return err
	}

	if !nodeMgr.AccessKeyClient.IsEnabled() {
		span.Finish()
		return err
	}
	controllerData = crmutil.NewControllerData(platform, &myCloudletInfo.Key, &nodeMgr)

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
		return err
	}
	notifyServerTls, err := nodeMgr.InternalPki.GetServerTlsConfig(ctx,
		nodeMgr.CommonName(),
		node.CertIssuerRegionalCloudlet,
		[]node.MatchCA{node.SameRegionalCloudletMatchCA()})
	if err != nil {
		return err
	}
	dialOption := tls.GetGrpcDialOption(notifyClientTls)
	notifyClient = notify.NewClient(nodeMgr.Name(), addrs, dialOption,
		notify.ClientUnaryInterceptors(nodeMgr.AccessKeyClient.UnaryAddAccessKey),
		notify.ClientStreamInterceptors(nodeMgr.AccessKeyClient.StreamAddAccessKey),
	)
	notifyClient.SetFilterByCloudletKey()
	InitClientNotify(notifyClient, controllerData)
	notifyClient.Start()
	services.notifyClientRunning = true

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

		myCloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_INIT
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
			FlavorCache:           &controllerData.FlavorCache,
			TrustPolicyCache:      &controllerData.TrustPolicyCache,
			ClusterInstCache:      &controllerData.ClusterInstCache,
			AppCache:              &controllerData.AppCache,
			AppInstCache:          &controllerData.AppInstCache,
			ResTagTableCache:      &controllerData.ResTagTableCache,
			CloudletCache:         controllerData.CloudletCache,
			CloudletInternalCache: &controllerData.CloudletInternalCache,
			VMPoolCache:           &controllerData.VMPoolCache,
			VMPoolInfoCache:       &controllerData.VMPoolInfoCache,
			GPUDriverCache:        &controllerData.GPUDriverCache,
		}

		features := platform.GetFeatures()
		if features.IsVMPool {
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
		if err = initPlatform(ctx, &cloudlet, &myCloudletInfo, *physicalName, &caches, nodeMgr.AccessApiClient, updateCloudletStatus); err != nil {
			myCloudletInfo.Errors = append(myCloudletInfo.Errors, fmt.Sprintf("Failed to init platform: %v", err))
			myCloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_ERRORS
		} else {
			log.SpanLog(ctx, log.DebugLevelInfo, "gathering cloudlet info")
			updateCloudletStatus(edgeproto.UpdateTask, "Gathering Cloudlet Info")
			err = controllerData.GatherCloudletInfo(ctx, &myCloudletInfo)
			log.SpanLog(ctx, log.DebugLevelInfra, "GatherCloudletInfo done", "state", myCloudletInfo.State)

			if err != nil {
				myCloudletInfo.Errors = append(myCloudletInfo.Errors, err.Error())
				myCloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_ERRORS
			} else {
				myCloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_NEED_SYNC
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

				// Update AppInst runtime info in case it has changed
				controllerData.RefreshAppInstRuntime(ctx)

				resources := controllerData.CaptureResourcesSnapshot(ctx, &cloudlet.Key)
				if resources != nil {
					resMap := make(map[string]edgeproto.InfraResource)
					for _, resInfo := range resources.Info {
						resMap[resInfo.Name] = resInfo
					}
					err = cloudcommon.ValidateCloudletResourceQuotas(ctx, resMap, cloudlet.ResourceQuotas)
					if err != nil {
						log.SpanLog(ctx, log.DebugLevelInfra, "Failed to validate cloudlet resource quota", "cloudlet", cloudlet.Key, "err", err)
						err = nil
					}
					myCloudletInfo.ResourcesSnapshot = *resources
				}
				myCloudletInfo.Errors = nil
				myCloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_READY
				if cloudlet.TrustPolicy == "" {
					myCloudletInfo.TrustPolicyState = edgeproto.TrackedState_NOT_PRESENT
				} else {
					myCloudletInfo.TrustPolicyState = edgeproto.TrackedState_READY
				}
				log.SpanLog(ctx, log.DebugLevelInfra, "cloudlet state", "state", myCloudletInfo.State, "myCloudletInfo", myCloudletInfo)
			}
		}

		log.SpanLog(ctx, log.DebugLevelInfo, "sending cloudlet info cache update")
		// trigger send of info upstream to controller
		controllerData.CloudletInfoCache.Update(ctx, &myCloudletInfo, 0)

		nodeMgr.MyNode.ContainerVersion = cloudletContainerVersion
		getMexReleaseInfo(ctx)
		nodeMgr.UpdateMyNode(ctx)
		log.SpanLog(ctx, log.DebugLevelInfo, "sent cloudletinfocache update")
		cspan.Finish()

		if err != nil {
			// die so CRM can restart and try again
			log.FatalLog("Platform init fail", "err", err)
		}

		// setup rootlb certs
		tlsSpan := log.StartSpan(log.DebugLevelInfo, "tls certs thread", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
		wildcardName := cloudcommon.GetRootLBFQDNWildcard(&myCloudletInfo.Key, *appDNSRoot)
		rootlb, err := platform.GetClusterPlatformClient(
			ctx,
			&edgeproto.ClusterInst{
				IpAccess: edgeproto.IpAccess_IP_ACCESS_SHARED,
			},
			cloudcommon.ClientTypeRootLB,
		)
		if err == nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Get rootLB certs", "key", myCloudletInfo.Key)
			proxycerts.Init(ctx, platform, accessapi.NewControllerClient(nodeMgr.AccessApiClient))
			pfType := pf.GetType(cloudlet.PlatformType.String())
			proxycerts.GetRootLbCerts(ctx, &myCloudletInfo.Key, wildcardName, &nodeMgr, pfType, rootlb, *commercialCerts)
			// setup debug func to trigger refresh of rootlb certs
			nodeMgr.Debug.AddDebugFunc(crmutil.RefreshRootLBCerts, func(ctx context.Context, req *edgeproto.DebugRequest) string {
				proxycerts.TriggerRootLBCertsRefresh()
				return "triggered refresh of rootlb certs"
			})
		}
		tlsSpan.Finish()
	}()

	// setup crm notify listener (for shepherd)
	initSrvNotify(&notifyServer)
	notifyServer.Start(nodeMgr.Name(), *notifySrvAddr, notifyServerTls)
	services.notifyServerRunning = true

	log.SpanLog(ctx, log.DebugLevelInfra, "Starting Cloudlet resource refresh thread", "cloudlet", myCloudletInfo.Key)
	controllerData.StartInfraResourceRefreshThread(&myCloudletInfo)
	services.infraResourcesRefreshRunning = true

	features := platform.GetFeatures()
	if features.SupportsNativeFlavors {
		log.SpanLog(ctx, log.DebugLevelInfra, "Starting flavor refresh thread", "cloudlet", myCloudletInfo.Key, "platform", platformName, "test mode", features.IsFake)
		controllerData.StartFlavorUpdateThread(&myCloudletInfo, features.IsFake)
		services.flavorRefreshRunning = true
	}
	span.Finish()
	return nil
}

func stopServices() {
	if services.flavorRefreshRunning {
		controllerData.FinishFlavorRefreshThread()
		services.flavorRefreshRunning = false
	}
	if services.infraResourcesRefreshRunning {
		controllerData.FinishInfraResourceRefreshThread()
		services.infraResourcesRefreshRunning = false
	}
	if notifyClient != nil && services.notifyClientRunning {
		notifyClient.Stop()
		services.notifyClientRunning = false
	}
	if services.notifyServerRunning {
		notifyServer.Stop()
		services.notifyServerRunning = false
	}
	nodeMgr.Finish()
}

//initializePlatform *Must be called as a seperate goroutine.*
func initPlatform(ctx context.Context, cloudlet *edgeproto.Cloudlet, cloudletInfo *edgeproto.CloudletInfo, physicalName string, caches *pf.Caches, accessClient edgeproto.CloudletAccessApiClient, updateCallback edgeproto.CacheUpdateCallback) error {
	loc := util.DNSSanitize(cloudletInfo.Key.Name) //XXX  key.name => loc
	oper := util.DNSSanitize(cloudletInfo.Key.Organization)
	accessApi := accessapi.NewControllerClient(accessClient)

	pc := pf.PlatformConfig{
		CloudletKey:         &cloudletInfo.Key,
		PhysicalName:        physicalName,
		Region:              *region,
		TestMode:            *testMode,
		CloudletVMImagePath: *cloudletVMImagePath,
		VMImageVersion:      *vmImageVersion,
		PackageVersion:      *packageVersion,
		EnvVars:             cloudlet.EnvVar,
		NodeMgr:             &nodeMgr,
		AppDNSRoot:          *appDNSRoot,
		DeploymentTag:       nodeMgr.DeploymentTag,
		Upgrade:             *upgrade,
		AccessApi:           accessApi,
		TrustPolicy:         cloudlet.TrustPolicy,
		CacheDir:            *cacheDir,
	}
	if cloudlet.GpuConfig.Driver.Name != "" {
		pc.GPUConfig = &cloudlet.GpuConfig
	}
	pfType := pf.GetType(cloudlet.PlatformType.String())
	log.SpanLog(ctx, log.DebugLevelInfra, "init platform", "location(cloudlet.key.name)", loc, "operator", oper, "Platform type", pfType)
	err := platform.Init(ctx, &pc, caches, updateCallback)
	return err
}

// Read file "/etc/mex-release" from original base vm image and parse certain env variables.
func readMexReleaseFileVars(ctx context.Context) (map[string]string, error) {
	filePath := "/etc/mex-release"
	m := make(map[string]string)

	file, err := os.Open(filePath)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Opening file /etc/mex-release failed", "err", err)
		return m, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		env := scanner.Text()
		envPair := strings.SplitN(env, "=", 2)
		if len(envPair) != 2 {
			continue
		}
		key := envPair[0]
		value := envPair[1]
		if key == envMexBuild || key == envMexBuildTag || key == envMexBuildFlavor {
			m[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Scanner failed on file /etc/mex-release", "err", err)
	}
	return m, nil
}

func getMexReleaseInfo(ctx context.Context) {
	m, err := readMexReleaseFileVars(ctx)
	if err != nil {
		return
	}
	k := envMexBuild
	v, ok := m[k]
	if ok {
		nodeMgr.MyNode.Properties[k] = v
	}
	k = envMexBuildTag
	v, ok = m[k]
	if ok {
		nodeMgr.MyNode.Properties[k] = v
	}
	k = envMexBuildFlavor
	v, ok = m[k]
	if ok {
		nodeMgr.MyNode.Properties[k] = v
	}
}
