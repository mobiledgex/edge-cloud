package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"

	//"github.com/mobiledgex/edge-cloud-infra/openstack-prov/oscliapi"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/testutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	yaml "gopkg.in/yaml.v2"
)

var bindAddress = flag.String("apiAddr", "0.0.0.0:55099", "Address to bind")
var controllerAddress = flag.String("controller", "127.0.0.1:55001", "Address of controller API")
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"TMUS\"},\"name\":\"tmocloud1\"}'")
var standalone = flag.Bool("standalone", false, "Standalone mode. CRM does not interact with controller. Cloudlet/AppInsts can be created directly on CRM using controller API")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))

// myCloudlet is the information for the cloudlet in which the CRM is instantiated.
// The key for myCloudlet is provided as a configuration - either command line or
// from a file. The rest of the data is extraced from Openstack.
var myCloudlet edgeproto.CloudletInfo

var sigChan chan os.Signal
var mainStarted chan struct{}
var notifyHandler *notify.DefaultHandler
var controllerData *crmutil.ControllerData
var notifyClient *notify.Client

var OSEnvValid = false

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	parseCloudletKey()

	OSEnvValid = ValidateOSEnv()

	listener, err := net.Listen("tcp", *bindAddress)
	if err != nil {
		log.FatalLog("Failed to bind", "addr", *bindAddress, "err", err)
	}

	controllerData = crmutil.NewControllerData()

	srv, err := crmutil.NewCloudResourceManagerServer(controllerData)

	grpcServer := grpc.NewServer()
	edgeproto.RegisterCloudResourceManagerServer(grpcServer, srv)

	if OSEnvValid {
		go func() {
			rootLB := crmutil.GetRootLBName()

			err = crmutil.RunMEXAgent(rootLB, false)
			if err != nil {
				log.FatalLog("Error running MEX Agent", "error", err)
				os.Exit(1)
			}
		}()
	}

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
		edgeproto.RegisterCloudletInfoApiServer(grpcServer, &saServer)
	} else {
		notifyHandler = NewNotifyHandler(controllerData)
		addrs := strings.Split(*notifyAddrs, ",")
		notifyClient = notify.NewCRMClient(addrs, notifyHandler)
		notifyClient.Start()
		defer notifyClient.Stop()
	}

	reflection.Register(grpcServer)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.FatalLog("Failed to serve grpc", "err", err)
		}
	}()
	defer grpcServer.Stop()

	log.InfoLog("Server started", "addr", *bindAddress)

	conn, err := grpc.Dial(*controllerAddress, grpc.WithInsecure())
	if err != nil {
		log.FatalLog("Failed to connect to controller",
			"addr", *controllerAddress, "err", err)
	}
	defer conn.Close()

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
	// trigger send of info upstream to controller
	controllerData.CloudletInfoCache.Update(&myCloudlet, 0)

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	if mainStarted != nil {
		// for unit testing to detect when main is ready
		close(mainStarted)
	}

	<-sigChan
	os.Exit(0)
}

func parseCloudletKey() {
	if *standalone && *cloudletKeyStr == "" {
		// Use fake cloudlet
		myCloudlet.Key = testutil.CloudletData[0].Key
		bytes, _ := json.Marshal(&myCloudlet.Key)
		*cloudletKeyStr = string(bytes)
		log.InfoLog("Using cloudletKey", "key", *cloudletKeyStr)
		return
	}

	if *cloudletKeyStr == "" {
		log.FatalLog("cloudletKey not specified")
	}

	err := json.Unmarshal([]byte(*cloudletKeyStr), &myCloudlet.Key)
	if err != nil {
		err = yaml.Unmarshal([]byte(*cloudletKeyStr), &myCloudlet.Key)
	}
	if err != nil {
		log.FatalLog("Failed to parse cloudletKey", "err", err)
	}

	err = myCloudlet.Key.Validate()
	if err != nil {
		log.FatalLog("Invalid cloudletKey", "err", err)
	}
}

func ValidateOSEnv() bool {
	osUser := os.Getenv("OS_USERNAME")
	osPass := os.Getenv("OS_PASSWORD")
	osTenant := os.Getenv("OS_TENANT")
	osAuthURL := os.Getenv("OS_AUTH_URL")
	osRegion := os.Getenv("OS_REGION_NAME")
	osCACert := os.Getenv("OC_CACERT")

	if osUser != "" && osPass != "" && osTenant != "" &&
		osAuthURL != "" && osRegion != "" && osCACert != "" {
		return crmutil.ValidateMEXOSEnv(true)
	}
	return false
}
