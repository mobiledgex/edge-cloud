package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	//"github.com/mobiledgex/edge-cloud-infra/openstack-prov/oscliapi"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var bindAddress = flag.String("apiAddr", "0.0.0.0:55099", "Address to bind")
var controllerAddress = flag.String("controller", "127.0.0.1:55001", "Address of controller API")
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"TMUS\"},\"name\":\"tmocloud1\"}'")
var standalone = flag.Bool("standalone", false, "Standalone mode. CRM does not interact with controller. Cloudlet/AppInsts can be created directly on CRM using controller API")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var tlsCertFile = flag.String("tls", "", "server tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory")

// myCloudlet is the information for the cloudlet in which the CRM is instantiated.
// The key for myCloudlet is provided as a configuration - either command line or
// from a file. The rest of the data is extraced from Openstack.
var myCloudlet edgeproto.CloudletInfo

var sigChan chan os.Signal
var mainStarted chan struct{}
var notifyHandler *notify.DefaultHandler
var controllerData *crmutil.ControllerData
var notifyClient *notify.Client

//OSEnvValid is used to signal validity of Openstack  platform environment
var OSEnvValid = false

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	cloudcommon.ParseMyCloudletKey(*standalone, cloudletKeyStr, &myCloudlet.Key)
	log.DebugLog(log.DebugLevelMexos, "Using cloudletKey", "key", myCloudlet.Key)
	rootLBName := cloudcommon.GetRootLBFQDN(&myCloudlet.Key)

	OSEnvValid = ValidateOSEnv()

	listener, err := net.Listen("tcp", *bindAddress)
	if err != nil {
		log.FatalLog("Failed to bind", "addr", *bindAddress, "err", err)
	}

	controllerData = crmutil.NewControllerData()

	srv, err := crmutil.NewCloudResourceManagerServer(controllerData)
	creds, err := tls.GetTLSServerCreds(*tlsCertFile)
	if err != nil {
		log.FatalLog("get TLS Credentials", "error", err)
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	edgeproto.RegisterCloudResourceManagerServer(grpcServer, srv)

	if OSEnvValid {
		log.DebugLog(log.DebugLevelMexos, "OS env valid")
		crmutil.MEXInit()
		go func() {
			log.DebugLog(log.DebugLevelMexos, "creating new rootLB", "rootlb", rootLBName)
			crmRootLB, cerr := crmutil.NewRootLB(rootLBName)
			if cerr != nil {
				log.DebugLog(log.DebugLevelMexos, "Can't get crm mex rootlb", "error", cerr)
				return
			}
			log.DebugLog(log.DebugLevelMexos, "created rootLB", "rootlb", rootLBName, "crmrootlb", crmRootLB)
			controllerData.CRMRootLB = crmRootLB
			log.DebugLog(log.DebugLevelMexos, "init platform with key", "cloudletkeystr", *cloudletKeyStr)
			err = crmutil.MEXPlatformInitCloudletKey(controllerData.CRMRootLB, *cloudletKeyStr)
			if err != nil {
				log.DebugLog(log.DebugLevelMexos, "Error running MEX Agent", "error", err)
			}
			log.DebugLog(log.DebugLevelMexos, "init platform with cloudlet key ok")
			//XXX we initialize platform when crmserver starts. But when do we clean up the platform?
		}()
	} else {
		log.DebugLog(log.DebugLevelMexos, "OS env invalid")
	}

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
		notifyHandler = NewNotifyHandler(controllerData)
		addrs := strings.Split(*notifyAddrs, ",")
		notifyClient = notify.NewCRMClient(addrs, *tlsCertFile, notifyHandler)
		// set callbacks to trigger send of infos
		controllerData.AppInstInfoCache.SetNotifyCb(notifyClient.UpdateAppInstInfo)
		controllerData.ClusterInstInfoCache.SetNotifyCb(notifyClient.UpdateClusterInstInfo)
		controllerData.CloudletInfoCache.SetNotifyCb(notifyClient.UpdateCloudletInfo)
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

	// gather cloudlet info
	if *standalone {
		// set fake cloudlet info
		myCloudlet.OsMaxRam = 500
		myCloudlet.OsMaxVcores = 50
		myCloudlet.OsMaxVolGb = 5000
	} else {
		// gather cloudlet info from openstack, etc.
		crmutil.GatherCloudletInfo(&myCloudlet)
	}
	log.DebugLog(log.DebugLevelMexos, "sending cloudlet info cache update")
	// trigger send of info upstream to controller
	controllerData.CloudletInfoCache.Update(&myCloudlet, 0)

	log.DebugLog(log.DebugLevelMexos, "sent cloudletinfocache update")
	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	if mainStarted != nil {
		// for unit testing to detect when main is ready
		close(mainStarted)
	}

	<-sigChan
	os.Exit(0)
}

//ValidateOSEnv makes sure environment is set up correctly for opensource
func ValidateOSEnv() bool {
	osUser := os.Getenv("OS_USERNAME")
	osPass := os.Getenv("OS_PASSWORD")
	osTenant := os.Getenv("OS_TENANT_NAME")
	osAuthURL := os.Getenv("OS_AUTH_URL")
	osRegion := os.Getenv("OS_REGION_NAME")
	osCACert := os.Getenv("OS_CACERT")

	if osUser != "" && osPass != "" && osTenant != "" && osAuthURL != "" && osRegion != "" && osCACert != "" {
		log.DebugLog(log.DebugLevelMexos, "Valid environment")
		return crmutil.ValidateMEXOSEnv(true)
	}
	log.DebugLog(log.DebugLevelMexos, "Invalid environment, you may need to set OS_USERNAME, OS_PASSWORD, OS_TENANT_NAME, OS_AUTH_URL, OS_REGION_NAME, OS_CACERT")
	//log.DebugLog(log.DebugLevelMexos, "Set", "osUser", osUser, "osPass", osPass, "osTenant", osTenant, "osAuthURL", osAuthURL, "osRegion", osRegion, "osCACert", osCACert)
	return false
}
