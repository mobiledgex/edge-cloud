package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
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
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"DMUUS\"},\"name\":\"tmocloud1\"}'")
var fakecloudlet = flag.Bool("fakecloudlet", false, "Fake cloudlet mode.  A fake cloudlet is reported to the controller")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var tlsCertFile = flag.String("tls", "", "server tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory")
var hostname = flag.String("hostname", "", "unique hostname within Cloudlet")

// myCloudlet is the information for the cloudlet in which the CRM is instantiated.
// The key for myCloudlet is provided as a configuration - either command line or
// from a file. The rest of the data is extraced from Openstack.
var myCloudlet edgeproto.CloudletInfo //XXX this effectively makes one CRM per cloudlet
var myNode edgeproto.Node

var sigChan chan os.Signal
var mainStarted chan struct{}
var controllerData *crmutil.ControllerData
var notifyClient *notify.Client

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	standalone := false
	cloudcommon.ParseMyCloudletKey(standalone, cloudletKeyStr, &myCloudlet.Key)
	cloudcommon.SetNodeKey(hostname, edgeproto.NodeType_NodeCRM, &myCloudlet.Key, &myNode.Key)
	log.DebugLog(log.DebugLevelMexos, "Using cloudletKey", "key", myCloudlet.Key)

	listener, err := net.Listen("tcp", *bindAddress)
	if err != nil {
		log.FatalLog("Failed to bind", "addr", *bindAddress, "err", err)
	}
	controllerData = crmutil.NewControllerData()

	creds, err := tls.GetTLSServerCreds(*tlsCertFile)
	if err != nil {
		log.FatalLog("get TLS Credentials", "error", err)
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	// this is to be done even for fake cloudlet
	if err := mexos.InitializeCloudletInfra(*fakecloudlet); err != nil {
		log.DebugLog(log.DebugLevelMexos, "error, cannot initialize cloudlet infra", "error", err)
		return
	}

	log.DebugLog(log.DebugLevelMexos, "gather cloudlet info")

	if *fakecloudlet {
		// set fake cloudlet info
		myCloudlet.OsMaxRam = 500
		myCloudlet.OsMaxVcores = 50
		myCloudlet.OsMaxVolGb = 5000
		myCloudlet.Flavors = []*edgeproto.FlavorInfo{
			&edgeproto.FlavorInfo{
				Name:  "flavor1",
				Vcpus: uint64(10),
				Ram:   uint64(101024),
				Disk:  uint64(500),
			},
		}
		myCloudlet.State = edgeproto.CloudletState_CloudletStateReady
		log.DebugLog(log.DebugLevelMexos, "sending fake cloudlet info cache update")
		// trigger send of info upstream to controller
		controllerData.CloudletInfoCache.Update(&myCloudlet, 0)
		controllerData.NodeCache.Update(&myNode, 0)
		log.DebugLog(log.DebugLevelMexos, "sent fake cloudletinfocache update")
		log.DebugLog(log.DebugLevelMexos, "running in fake cloudlet mode")
	} else {
		go func() {
			// gather cloudlet info from openstack, etc.
			crmutil.GatherCloudletInfo(&myCloudlet)
			if len(myCloudlet.Errors) > 0 {
				log.DebugLog(log.DebugLevelMexos, "GatherCloudletInfo", "Error", myCloudlet.Errors[0])
				return
			}

			/*
			 * FIXME:For now choose first flavor as platform flavor
			 * This will soon change when we support one rootlb per cluster for dedicated IP type
			 */
			pFlavor := myCloudlet.Flavors[0]
			log.DebugLog(log.DebugLevelMexos, "platform flavor", "flavor", pFlavor)

			log.DebugLog(log.DebugLevelMexos, "starting to init platform")
			if err := initPlatform(&myCloudlet, pFlavor.Name); err != nil {
				log.DebugLog(log.DebugLevelMexos, "error, cannot initialize platform", "error", err)
				return
			}
			if controllerData.CRMRootLB == nil {
				log.DebugLog(log.DebugLevelMexos, "error, failed to init platform, crmRootLB is null")
				return
			}

			log.DebugLog(log.DebugLevelMexos, "sending cloudlet info cache update")
			// trigger send of info upstream to controller
			controllerData.CloudletInfoCache.Update(&myCloudlet, 0)
			controllerData.NodeCache.Update(&myNode, 0)
			log.DebugLog(log.DebugLevelMexos, "sent cloudletinfocache update")
		}()
	}

	// GatherInsts should be called before the notify client is started,
	// so that the initial send to the controller has the current state.
	controllerData.GatherInsts()

	addrs := strings.Split(*notifyAddrs, ",")
	notifyClient = notify.NewClient(addrs, *tlsCertFile)
	notifyClient.SetFilterByCloudletKey()
	InitNotify(notifyClient, controllerData)
	notifyClient.Start()
	defer notifyClient.Stop()
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

	if mainStarted != nil {
		// for unit testing to detect when main is ready
		close(mainStarted)
	}

	<-sigChan
	os.Exit(0)
}

//initializePlatform *Must be called as a seperate goroutine.*
func initPlatform(cloudlet *edgeproto.CloudletInfo, platformFlavor string) error {
	loc := util.DNSSanitize(cloudlet.Key.Name) //XXX  key.name => loc
	oper := util.DNSSanitize(cloudlet.Key.OperatorKey.Name)
	//if err := mexos.FillManifestValues(mf, "platform"); err != nil {
	//	return err
	//}

	rootLBName := cloudcommon.GetRootLBFQDN(&myCloudlet.Key)

	log.DebugLog(log.DebugLevelMexos, "init platform, creating new rootLB", "location(cloudlet.key.name)", loc, "operator", oper, "rootLBName", rootLBName)

	crmRootLB, cerr := mexos.NewRootLB(rootLBName)
	if cerr != nil {
		return cerr
	}
	if crmRootLB == nil {
		return fmt.Errorf("rootLB is not initialized")
	}
	log.DebugLog(log.DebugLevelMexos, "created rootLB", "rootlb", crmRootLB.Name)
	controllerData.CRMRootLB = crmRootLB

	log.DebugLog(log.DebugLevelMexos, "calling RunMEXAgentCloudletKey", "cloudletkeystr", *cloudletKeyStr)
	if err := mexos.RunMEXAgentCloudletKey(controllerData.CRMRootLB.Name, *cloudletKeyStr, platformFlavor); err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "ok, RunMEXAgentCloudletKey with cloudlet key")
	return nil
}
