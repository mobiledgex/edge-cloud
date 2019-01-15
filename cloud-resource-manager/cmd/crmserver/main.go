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

var base = flag.String("base", fmt.Sprintf("scp://%s/files-repo/mobiledgex", cloudcommon.Registry), "base")
var bindAddress = flag.String("apiAddr", "0.0.0.0:55099", "Address to bind")
var controllerAddress = flag.String("controller", "127.0.0.1:55001", "Address of controller API")
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"TMUS\"},\"name\":\"tmocloud1\"}'")
var standalone = flag.Bool("standalone", false, "Standalone mode. CRM does not interact with controller. Cloudlet/AppInsts can be created directly on CRM using controller API")
var fakecloudlet = flag.Bool("fakecloudlet", false, "Fake cloudlet mode.  A fake cloudlet is reported to the controller")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var tlsCertFile = flag.String("tls", "", "server tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory")
var hostname = flag.String("hostname", "", "unique hostname within Cloudlet")

// myCloudlet is the information for the cloudlet in which the CRM is instantiated.
// The key for myCloudlet is provided as a configuration - either command line or
// from a file. The rest of the data is extraced from Openstack.
var myCloudlet edgeproto.CloudletInfo //XXX this effectively makes one CRM per cloudlet
var myNode edgeproto.Node
var isDIND bool

var sigChan chan os.Signal
var mainStarted chan struct{}
var notifyHandler *notify.DefaultHandler
var controllerData *crmutil.ControllerData
var notifyClient *notify.Client

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	cloudcommon.ParseMyCloudletKey(*standalone, cloudletKeyStr, &myCloudlet.Key)
	cloudcommon.SetNodeKey(hostname, edgeproto.NodeType_NodeCRM, &myCloudlet.Key, &myNode.Key)
	log.DebugLog(log.DebugLevelMexos, "Using cloudletKey", "key", myCloudlet.Key)
	rootLBName := cloudcommon.GetRootLBFQDN(&myCloudlet.Key)
	log.DebugLog(log.DebugLevelMexos, "rootlb name", "rootLBName", rootLBName)

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

	platChan := make(chan string)
	if *fakecloudlet {
		log.DebugLog(log.DebugLevelMexos, "running in fake cloudlet mode")
	} else {
		go func() {
			log.DebugLog(log.DebugLevelMexos, "starting to init platform")
			if err := initPlatform(&myCloudlet); err != nil {
				log.DebugLog(log.DebugLevelMexos, "error, cannot initialize platform", "error", err)
				return
			}
			if controllerData.CRMRootLB == nil {
				log.DebugLog(log.DebugLevelMexos, "error, failed to init platform, crmRootLB is null")
				return
			}
			log.DebugLog(log.DebugLevelMexos, "send status on plat channel")
			platChan <- "ready"
		}()
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
		controllerData.NodeCache.SetNotifyCb(notifyClient.UpdateNode)
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
	log.DebugLog(log.DebugLevelMexos, "check isdind", "isDIND", isDIND)

	if *standalone || *fakecloudlet {
		// set fake cloudlet info
		myCloudlet.OsMaxRam = 500
		myCloudlet.OsMaxVcores = 50
		myCloudlet.OsMaxVolGb = 5000
		myCloudlet.State = edgeproto.CloudletState_CloudletStateReady
		log.DebugLog(log.DebugLevelMexos, "sending fake cloudlet info cache update")
		// trigger send of info upstream to controller
		controllerData.CloudletInfoCache.Update(&myCloudlet, 0)
		controllerData.NodeCache.Update(&myNode, 0)
		log.DebugLog(log.DebugLevelMexos, "sent fake cloudletinfocache update")
	} else {
		go func() {
			log.DebugLog(log.DebugLevelMexos, "wait for status on plat channel")
			platStat := <-platChan
			if isDIND {
				myCloudlet.State = edgeproto.CloudletState_CloudletStateReady

			} else {
				log.DebugLog(log.DebugLevelMexos, "got status on plat channel", "status", platStat)
				// gather cloudlet info from openstack, etc.
				if controllerData.CRMRootLB == nil {
					log.DebugLog(log.DebugLevelMexos, "rootlb is nil in controllerdata")
					return
				}
				crmutil.GatherCloudletInfo(controllerData.CRMRootLB, &myCloudlet)
			}
			log.DebugLog(log.DebugLevelMexos, "sending cloudlet info cache update")
			// trigger send of info upstream to controller
			controllerData.CloudletInfoCache.Update(&myCloudlet, 0)
			controllerData.NodeCache.Update(&myNode, 0)
			log.DebugLog(log.DebugLevelMexos, "sent cloudletinfocache update")
		}()
	}
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
	mf := &mexos.Manifest{Base: *base}
	uri := fmt.Sprintf("kustomize/infrastructure/output/%s.%s.yaml", loc, oper)
	log.DebugLog(log.DebugLevelMexos, "init platform, creating new rootLB", "base", *base, "location(cloudlet.key.name)", loc, "operator", oper, "uri", uri)
	if err := mexos.GetVaultEnv(mf, uri); err != nil {
		return err
	}
	if err := mexos.FillManifestValues(mf, "platform", *base); err != nil {
		return err
	}
	if err := mexos.CheckManifest(mf); err != nil {
		return err
	}
	crmRootLB, cerr := mexos.NewRootLBManifest(mf)
	if cerr != nil {
		return cerr
	}
	if crmRootLB == nil {
		return fmt.Errorf("rootLB is not initialized")
	}
	log.DebugLog(log.DebugLevelMexos, "created rootLB", "rootlb", crmRootLB.Name)
	controllerData.CRMRootLB = crmRootLB
	if err := mexos.MEXInit(mf); err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "calling init platform with key", "cloudletkeystr", *cloudletKeyStr)
	if err := mexos.MEXPlatformInitCloudletKey(controllerData.CRMRootLB, *cloudletKeyStr); err != nil {
		return err
	}
	isDIND = mexos.IsLocalDIND(mf)
	log.DebugLog(log.DebugLevelMexos, "IS DIND SET", "isDIND", isDIND)

	log.DebugLog(log.DebugLevelMexos, "ok, init platform with cloudlet key")
	return nil
}
