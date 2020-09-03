package main

import (
	"context"
	ctls "crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	baselog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"plugin"
	"strings"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	op "github.com/mobiledgex/edge-cloud/d-match-engine/operator"
	operator "github.com/mobiledgex/edge-cloud/d-match-engine/operator"
	"github.com/mobiledgex/edge-cloud/d-match-engine/operator/defaultoperator"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	log "github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/segmentio/ksuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
)

// Command line options
var rootDir = flag.String("r", "", "root directory for testing")
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var apiAddr = flag.String("apiAddr", "localhost:50051", "API listener address")
var httpAddr = flag.String("httpAddr", "127.0.0.1:38001", "HTTP listener address")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var locVerUrl = flag.String("locverurl", "", "location verification REST API URL to connect to")
var tokSrvUrl = flag.String("toksrvurl", "", "token service URL to provide to client on register")
var qosPosUrl = flag.String("qosposurl", "", "QOS Position KPI URL to connect to")
var tlsApiCertFile = flag.String("tlsApiCertFile", "", "Public-CA signed TLS cert file for serving DME APIs")
var tlsApiKeyFile = flag.String("tlsApiKeyFile", "", "Public-CA signed TLS key file for serving DME APIs")
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"TMUS\"},\"name\":\"tmocloud1\"}'")
var scaleID = flag.String("scaleID", "", "ID to distinguish multiple DMEs in the same cloudlet. Defaults to hostname if unspecified.")
var statsInterval = flag.Int("statsInterval", 1, "interval in seconds between sending stats")
var statsShards = flag.Uint("statsShards", 10, "number of shards (locks) in memory for parallel stat collection")
var cookieExpiration = flag.Duration("cookieExpiration", time.Hour*24, "Cookie expiration time")
var region = flag.String("region", "local", "region name")
var solib = flag.String("plugin", "", "plugin file")
var eesolib = flag.String("eeplugin", "", "plugin file")
var testMode = flag.Bool("testMode", false, "Run controller in test mode")
var monitorUuidType = flag.String("monitorUuidType", "MobiledgeXMonitorProbe", "AppInstClient UUID Type used for monitoring purposes")
var cloudletDme = flag.Bool("cloudletDme", false, "this is a cloudlet DME deployed on cloudlet infrastructure and uses the crm access key")

// TODO: carrier arg is redundant with Organization in myCloudletKey, and
// should be replaced by it, but requires dealing with carrier-specific
// verify location API behavior and e2e test setups.
var carrier = flag.String("carrier", "standalone", "carrier name for API connection, or standalone for no external APIs")

var operatorApiGw op.OperatorApiGw

// server is used to implement helloworld.GreeterServer.
type server struct{}

// myCloudlet is the information for the cloudlet in which the DME is instantiated.
// The key for myCloudlet is provided as a configuration - either command line or
// from a file.
var myCloudletKey edgeproto.CloudletKey
var nodeMgr node.NodeMgr

var sigChan chan os.Signal

func validateLocation(loc *dme.Loc) error {
	if loc == nil || (loc.Latitude == 0 && loc.Longitude == 0) {
		return grpc.Errorf(codes.InvalidArgument, "Missing GpsLocation")
	}
	if !util.IsLatitudeValid(loc.Latitude) || !util.IsLongitudeValid(loc.Longitude) {
		return grpc.Errorf(codes.InvalidArgument, "Invalid GpsLocation")
	}
	return nil
}

func (s *server) FindCloudlet(ctx context.Context, req *dme.FindCloudletRequest) (*dme.FindCloudletReply, error) {
	reply := new(dme.FindCloudletReply)
	var appkey edgeproto.AppKey
	ckey, ok := dmecommon.CookieFromContext(ctx)
	if !ok {
		return reply, grpc.Errorf(codes.InvalidArgument, "No valid session cookie")
	}
	appkey.Organization = ckey.OrgName
	appkey.Name = ckey.AppName
	appkey.Version = ckey.AppVers

	err := validateLocation(req.GpsLocation)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelDmereq, "Invalid FindCloudlet request, invalid location", "loc", req.GpsLocation, "err", err)
		return reply, err
	}
	err = dmecommon.FindCloudlet(ctx, &appkey, req.CarrierName, req.GpsLocation, reply, cookieExpiration)
	log.SpanLog(ctx, log.DebugLevelDmereq, "FindCloudlet returns", "reply", reply, "error", err)
	return reply, err
}

func (s *server) PlatformFindCloudlet(ctx context.Context, req *dme.PlatformFindCloudletRequest) (*dme.FindCloudletReply, error) {
	reply := new(dme.FindCloudletReply)
	ckey, ok := dmecommon.CookieFromContext(ctx)
	if !ok {
		return reply, grpc.Errorf(codes.InvalidArgument, "No valid session cookie")
	}
	if !cloudcommon.IsPlatformApp(ckey.OrgName, ckey.AppName) {
		log.SpanLog(ctx, log.DebugLevelDmereq, "PlatformFindCloudlet API Not allowed for developer app", "org", ckey.OrgName, "name", ckey.AppName)
		return reply, grpc.Errorf(codes.PermissionDenied, "API Not allowed for developer: %s app: %s", ckey.OrgName, ckey.AppName)
	}
	if req.ClientToken == "" {
		log.SpanLog(ctx, log.DebugLevelDmereq, "Invalid PlatformFindCloudlet request", "Error", "Missing ClientToken")
		return reply, grpc.Errorf(codes.InvalidArgument, "Missing ClientToken")
	}
	tokdata, err := dmecommon.GetClientDataFromToken(req.ClientToken)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelDmereq, "Invalid PlatformFindCloudletRequest request, unable to get data from token", "token", req.ClientToken, "err", err)
		return reply, grpc.Errorf(codes.InvalidArgument, "Invalid ClientToken")
	}

	if tokdata.AppKey.Organization == "" {
		log.SpanLog(ctx, log.DebugLevelDmereq, "OrgName in token cannot be empty")
		return reply, grpc.Errorf(codes.InvalidArgument, "OrgName in token cannot be empty")
	}
	if tokdata.AppKey.Name == "" {
		log.SpanLog(ctx, log.DebugLevelDmereq, "AppName in token cannot be empty")
		return reply, grpc.Errorf(codes.InvalidArgument, "AppName in token cannot be empty")
	}
	if tokdata.AppKey.Version == "" {
		log.SpanLog(ctx, log.DebugLevelDmereq, "AppVers in token cannot be empty")
		return reply, grpc.Errorf(codes.InvalidArgument, "AppVers in token cannot be empty")
	}
	if !dmecommon.AppExists(tokdata.AppKey.Organization, tokdata.AppKey.Name, tokdata.AppKey.Version) {
		log.SpanLog(ctx, log.DebugLevelDmereq, "Requested app does not exist", "requestedAppKey", tokdata.AppKey)
		return reply, grpc.Errorf(codes.InvalidArgument, "Requested app does not exist")
	}
	err = validateLocation(&tokdata.Location)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelDmereq, "Invalid PlatformFindCloudletRequest request, invalid location", "loc", tokdata.Location, "err", err)
		return reply, grpc.Errorf(codes.InvalidArgument, "Invalid ClientToken")
	}
	err = dmecommon.FindCloudlet(ctx, &tokdata.AppKey, req.CarrierName, &tokdata.Location, reply, cookieExpiration)
	log.SpanLog(ctx, log.DebugLevelDmereq, "PlatformFindCloudletRequest returns", "reply", reply, "error", err)
	return reply, err
}

func (s *server) GetFqdnList(ctx context.Context, req *dme.FqdnListRequest) (*dme.FqdnListReply, error) {
	log.SpanLog(ctx, log.DebugLevelDmereq, "GetFqdnList", "req", req)
	flist := new(dme.FqdnListReply)

	ckey, ok := dmecommon.CookieFromContext(ctx)
	if !ok {
		return nil, grpc.Errorf(codes.InvalidArgument, "No valid session cookie")
	}
	// normal applications are not allowed to access this, only special platform developer/app combos
	if !cloudcommon.IsPlatformApp(ckey.OrgName, ckey.AppName) {
		return nil, grpc.Errorf(codes.PermissionDenied, "API Not allowed for developer: %s app: %s", ckey.OrgName, ckey.AppName)
	}

	dmecommon.GetFqdnList(req, flist)
	log.SpanLog(ctx, log.DebugLevelDmereq, "GetFqdnList returns", "status", flist.Status)
	return flist, nil
}

func (s *server) GetAppInstList(ctx context.Context, req *dme.AppInstListRequest) (*dme.AppInstListReply, error) {
	ckey, ok := dmecommon.CookieFromContext(ctx)
	if !ok {
		return nil, grpc.Errorf(codes.InvalidArgument, "No valid session cookie")
	}

	log.SpanLog(ctx, log.DebugLevelDmereq, "GetAppInstList", "carrier", req.CarrierName, "ckey", ckey)

	if req.GpsLocation == nil {
		log.SpanLog(ctx, log.DebugLevelDmereq, "Invalid GetAppInstList request", "Error", "Missing GpsLocation")
		return nil, grpc.Errorf(codes.InvalidArgument, "Missing GPS location")
	}
	alist := new(dme.AppInstListReply)
	dmecommon.GetAppInstList(ctx, ckey, req, alist)
	log.SpanLog(ctx, log.DebugLevelDmereq, "GetAppInstList returns", "status", alist.Status)
	return alist, nil
}

func (s *server) GetAppOfficialFqdn(ctx context.Context, req *dme.AppOfficialFqdnRequest) (*dme.AppOfficialFqdnReply, error) {
	ckey, ok := dmecommon.CookieFromContext(ctx)
	reply := new(dme.AppOfficialFqdnReply)
	if !ok {
		return nil, grpc.Errorf(codes.InvalidArgument, "No valid session cookie")
	}
	log.SpanLog(ctx, log.DebugLevelDmereq, "GetAppOfficialFqdn", "ckey", ckey, "loc", req.GpsLocation)
	err := validateLocation(req.GpsLocation)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelDmereq, "Invalid GetAppOfficialFqdn request, invalid location", "loc", req.GpsLocation, "err", err)
		return reply, err
	}
	dmecommon.GetAppOfficialFqdn(ctx, ckey, req, reply)
	log.SpanLog(ctx, log.DebugLevelDmereq, "GetAppOfficialFqdn returns", "status", reply.Status)
	return reply, nil
}

func (s *server) VerifyLocation(ctx context.Context,
	req *dme.VerifyLocationRequest) (*dme.VerifyLocationReply, error) {

	reply := new(dme.VerifyLocationReply)

	reply.GpsLocationStatus = dme.VerifyLocationReply_LOC_UNKNOWN
	reply.GpsLocationAccuracyKm = -1

	log.SpanLog(ctx, log.DebugLevelDmereq, "Received Verify Location",
		"VerifyLocToken", req.VerifyLocToken,
		"GpsLocation", req.GpsLocation)

	if req.GpsLocation == nil || (req.GpsLocation.Latitude == 0 && req.GpsLocation.Longitude == 0) {
		log.SpanLog(ctx, log.DebugLevelDmereq, "Invalid VerifyLocation request", "Error", "Missing GpsLocation")
		return reply, grpc.Errorf(codes.InvalidArgument, "Missing GPS location")
	}

	if !util.IsLatitudeValid(req.GpsLocation.Latitude) || !util.IsLongitudeValid(req.GpsLocation.Longitude) {
		log.SpanLog(ctx, log.DebugLevelDmereq, "Invalid VerifyLocation GpsLocation", "lat", req.GpsLocation.Latitude, "long", req.GpsLocation.Longitude)
		return reply, grpc.Errorf(codes.InvalidArgument, "Invalid GpsLocation")
	}
	err := operatorApiGw.VerifyLocation(req, reply)
	return reply, err

}

func (s *server) GetLocation(ctx context.Context,
	req *dme.GetLocationRequest) (*dme.GetLocationReply, error) {
	reply := new(dme.GetLocationReply)
	err := operatorApiGw.GetLocation(req, reply)
	return reply, err
}

func (s *server) RegisterClient(ctx context.Context,
	req *dme.RegisterClientRequest) (*dme.RegisterClientReply, error) {

	mstatus := new(dme.RegisterClientReply)

	log.SpanLog(ctx, log.DebugLevelDmereq, "RegisterClient received", "request", req)

	if req.OrgName == "" {
		log.SpanLog(ctx, log.DebugLevelDmereq, "OrgName cannot be empty")
		mstatus.Status = dme.ReplyStatus_RS_FAIL
		return mstatus, grpc.Errorf(codes.InvalidArgument, "OrgName cannot be empty")
	}
	if req.AppName == "" {
		log.SpanLog(ctx, log.DebugLevelDmereq, "AppName cannot be empty")
		mstatus.Status = dme.ReplyStatus_RS_FAIL
		return mstatus, grpc.Errorf(codes.InvalidArgument, "AppName cannot be empty")
	}
	if req.AppVers == "" {
		log.SpanLog(ctx, log.DebugLevelDmereq, "AppVers cannot be empty")
		mstatus.Status = dme.ReplyStatus_RS_FAIL
		return mstatus, grpc.Errorf(codes.InvalidArgument, "AppVers cannot be empty")
	}
	authkey, err := dmecommon.GetAuthPublicKey(req.OrgName, req.AppName, req.AppVers)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelDmereq, "fail to get public key", "err", err)
		mstatus.Status = dme.ReplyStatus_RS_FAIL
		return mstatus, err
	}

	//the token is currently optional, but once the SDK is enhanced to send one, it should
	// be a mandatory parameter.  For now, only validate the token if we receive one
	if req.AuthToken == "" {
		if authkey != "" {
			// we provisioned a key, and one was not provided.
			log.SpanLog(ctx, log.DebugLevelDmereq, "App has key, no token received")
			mstatus.Status = dme.ReplyStatus_RS_FAIL
			return mstatus, grpc.Errorf(codes.InvalidArgument, "No authtoken received")
		}
		// for now we will allow a tokenless register to pass if the app does not have one
		log.SpanLog(ctx, log.DebugLevelDmereq, "Allowing register without token")

	} else {
		if authkey == "" {
			log.SpanLog(ctx, log.DebugLevelDmereq, "No authkey provisioned to validate token")
			mstatus.Status = dme.ReplyStatus_RS_FAIL
			return mstatus, grpc.Errorf(codes.Unauthenticated, "No authkey found to validate token")
		}
		err := dmecommon.VerifyAuthToken(ctx, req.AuthToken, authkey, req.OrgName, req.AppName, req.AppVers)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelDmereq, "Failed to verify token", "err", err)
			mstatus.Status = dme.ReplyStatus_RS_FAIL
			return mstatus, grpc.Errorf(codes.Unauthenticated, "failed to verify token - %s", err.Error())
		}
	}

	key := dmecommon.CookieKey{
		OrgName: req.OrgName,
		AppName: req.AppName,
		AppVers: req.AppVers,
	}

	// Both UniqueId and UniqueIdType should be set, or none
	if (req.UniqueId != "" && req.UniqueIdType == "") ||
		(req.UniqueIdType != "" && req.UniqueId == "") {
		mstatus.Status = dme.ReplyStatus_RS_FAIL
		return mstatus, grpc.Errorf(codes.InvalidArgument,
			"Both, or none of UniqueId and UniqueIdType should be set")
	}

	// If we get UUID from the Request, use that, otherwise generate a new one
	if req.UniqueIdType != "" && req.UniqueId != "" {
		key.UniqueId = req.UniqueId
		key.UniqueIdType = req.UniqueIdType
	} else {
		// Generate KSUID
		uid := ksuid.New()
		key.UniqueId = uid.String()
		if cloudcommon.IsPlatformApp(req.OrgName, req.AppName) {
			key.UniqueIdType = req.OrgName + ":" + req.AppName
		} else {
			key.UniqueIdType = "dme-ksuid"
		}
	}

	cookie, err := dmecommon.GenerateCookie(&key, ctx, cookieExpiration)
	if err != nil {
		return mstatus, grpc.Errorf(codes.Internal, err.Error())
	}

	// Set UUID in the reply if none were sent in the request
	// We only respond back with a UUID if it's generated by the DME
	if req.UniqueIdType == "" || req.UniqueId == "" {
		mstatus.UniqueIdType = key.UniqueIdType
		mstatus.UniqueId = key.UniqueId
	}
	mstatus.SessionCookie = cookie
	mstatus.TokenServerUri = *tokSrvUrl
	mstatus.Status = dme.ReplyStatus_RS_SUCCESS
	return mstatus, nil
}

func (s *server) AddUserToGroup(ctx context.Context,
	req *dme.DynamicLocGroupRequest) (*dme.DynamicLocGroupReply, error) {

	mreq := new(dme.DynamicLocGroupReply)
	mreq.Status = dme.ReplyStatus_RS_SUCCESS

	return mreq, nil
}

func (s *server) GetQosPositionKpi(req *dme.QosPositionRequest, getQosSvr dme.MatchEngineApi_GetQosPositionKpiServer) error {
	log.SpanLog(getQosSvr.Context(), log.DebugLevelDmereq, "GetQosPositionKpi", "request", req)
	return operatorApiGw.GetQOSPositionKPI(req, getQosSvr)
}

func (s *server) SendEdgeEvent(sendEdgeEventSvr dme.MatchEngineApi_SendEdgeEventServer) error {
	// ctx := sendEdgeEventSvr.Context()
	ctx, cancel := context.WithCancel(sendEdgeEventSvr.Context())
	log.SpanLog(ctx, log.DebugLevelDmereq, "SendEdgeEvent")

	return dmecommon.SendEdgeEvent(ctx, &sendEdgeEventSvr, cancel)
}

func initEdgeEventsPlugin(ctx context.Context) error {
	// Load Edge Events Plugin
	*eesolib = os.Getenv("GOPATH") + "/plugins/edgeevents.so"
	log.SpanLog(ctx, log.DebugLevelDmereq, "Loading plugin", "plugin", *eesolib)
	plug, err := plugin.Open(*eesolib)
	if err != nil {
		log.WarnLog("failed to load plugin", "plugin", *eesolib, "error", err)
		return err
	}

	// Load AddClient in Plugin
	sym, err := plug.Lookup("AddClient")
	if err != nil {
		log.FatalLog("plugin does not have AddClient symbol", "plugin", *eesolib)
		return err
	}
	addClientFunc, ok := sym.(func(ctx context.Context, appInstKey *edgeproto.AppInstKey, peer *peer.Peer, mgr *dmecommon.EdgeEventPersistentMgr))
	if !ok {
		log.FatalLog("plugin AddClient symbol does not implement func(ctx context.Context, appInstKey *edgeproto.AppInstKey, peer *peer.Peer, mgr *EdgeEventPersistentMgr)", "plugin", *eesolib)
	}
	dmecommon.AddClient = addClientFunc

	// Load RemoveClient in Plugin
	sym, err = plug.Lookup("RemoveClient")
	if err != nil {
		log.FatalLog("plugin does not have RemoveClient symbol", "plugin", *eesolib)
		return err
	}
	removeClientFunc, ok := sym.(func(ctx context.Context, appInstKey *edgeproto.AppInstKey, peer *peer.Peer))
	if !ok {
		log.FatalLog("plugin AddClient symbol does not implement func(ctx context.Context, appInstKey *edgeproto.AppInstKey, peer *peer.Peer)", "plugin", *eesolib)
	}
	dmecommon.RemoveClient = removeClientFunc

	// Load MeasureLatency in Plugin
	sym, err = plug.Lookup("MeasureLatency")
	if err != nil {
		log.FatalLog("plugin does not have MeasureLatency symbol", "plugin", *eesolib)
		return err
	}
	measureLatencyFunc, ok := sym.(func(ctx context.Context, appInst *dmecommon.DmeAppInst, appInstKey *edgeproto.AppInstKey))
	if !ok {
		log.FatalLog("plugin MeasureLatency symbol does not implement func(ctx context.Context, appInst *DmeAppInst, appInstKey *edgeproto.AppInstKey)", "plugin", *eesolib)
	}
	dmecommon.MeasureLatency = measureLatencyFunc

	// Load UpdateAppInstState in Plugin
	sym, err = plug.Lookup("UpdateAppInstState")
	if err != nil {
		log.FatalLog("plugin does not have UpdateAppInstState symbol", "plugin", *eesolib)
		return err
	}
	updateAppInstStateFunc, ok := sym.(func(ctx context.Context, appInst *dmecommon.DmeAppInst, appInstKey *edgeproto.AppInstKey))
	if !ok {
		log.FatalLog("plugin UpdateAppInstState symbol does not implement func(ctx context.Context, appInst *DmeAppInst, appInstKey *edgeproto.AppInstKey)", "plugin", *eesolib)
	}
	dmecommon.UpdateAppInstState = updateAppInstStateFunc

	// Load UpdateClients in Plugin
	sym, err = plug.Lookup("UpdateClients")
	if err != nil {
		log.FatalLog("plugin does not have UpdateClients symbol", "plugin", *eesolib)
		return err
	}
	updateClientsFunc, ok := sym.(func(ctx context.Context, serverEdgeEvent *dme.ServerEdgeEvent, mgr *dmecommon.EdgeEventPersistentMgr))
	if !ok {
		log.FatalLog("plugin UpdateClients symbol does not implement func(ctx context.Context, serverEdgeEvent *dme.ServerEdgeEvent, mgr *EdgeEventPersistentMgr)", "plugin", *eesolib)
	}
	dmecommon.UpdateClients = updateClientsFunc

	return nil
}

func initOperator(ctx context.Context, operatorName string) (op.OperatorApiGw, error) {
	if operatorName == "" || operatorName == "standalone" {
		return &defaultoperator.OperatorApiGw{}, nil
	}
	if *solib == "" {
		*solib = os.Getenv("GOPATH") + "/plugins/platforms.so"
	}
	log.SpanLog(ctx, log.DebugLevelDmereq, "Loading plugin", "plugin", *solib)
	plug, err := plugin.Open(*solib)
	if err != nil {
		log.WarnLog("failed to load plugin", "plugin", *solib, "error", err)
		return nil, err
	}
	sym, err := plug.Lookup("GetOperatorApiGw")
	if err != nil {
		log.FatalLog("plugin does not have GetOperatorApiGw symbol", "plugin", *solib)
	}
	getOperatorFunc, ok := sym.(func(ctx context.Context, operatorName string) (op.OperatorApiGw, error))
	if !ok {
		log.FatalLog("plugin GetOperatorApiGw symbol does not implement func(ctx context.Context, opername string) (op.OperatorApiGw, error)", "plugin", *solib)
	}
	return getOperatorFunc(ctx, operatorName)
}

// allowCORS allows Cross Origin Resoruce Sharing from any origin.
func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				// preflight headers
				headers := []string{"Content-Type", "Accept"}
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
				methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func main() {
	nodeMgr.InitFlags()
	nodeMgr.AccessKeyClient.InitFlags()
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)

	var myCertIssuer string
	if *cloudletDme {
		// DME running on Cloudlet infra, requires access key
		myCertIssuer = node.CertIssuerRegionalCloudlet
	} else {
		// Regional DME not associated with any Cloudlet
		myCertIssuer = node.CertIssuerRegional
	}
	cloudcommon.ParseMyCloudletKey(false, cloudletKeyStr, &myCloudletKey)
	ctx, span, err := nodeMgr.Init(node.NodeTypeDME, myCertIssuer, node.WithName(*scaleID), node.WithCloudletKey(&myCloudletKey), node.WithRegion(*region))
	if err != nil {
		log.FatalLog("Failed init node", "err", err)
	}
	defer nodeMgr.Finish()
	operatorApiGw, err = initOperator(ctx, *carrier)
	if err != nil {
		span.Finish()
		log.FatalLog("Failed init plugin", "operator", *carrier, "err", err)
	}
	var servers = operator.OperatorApiGwServers{VaultAddr: nodeMgr.VaultAddr, QosPosUrl: *qosPosUrl, LocVerUrl: *locVerUrl, TokSrvUrl: *tokSrvUrl}
	err = operatorApiGw.Init(*carrier, &servers)
	if err != nil {
		span.Finish()
		log.FatalLog("Unable to init API GW", "err", err)

	}
	log.SpanLog(ctx, log.DebugLevelInfo, "plugin init done", "operatorApiGw", operatorApiGw)

	err = initEdgeEventsPlugin(ctx)
	if err != nil {
		span.Finish()
		log.FatalLog("Failed init edge events plugin")
	}

	err = dmecommon.InitVault(nodeMgr.VaultAddr, *region)
	if err != nil {
		span.Finish()
		log.FatalLog("Failed to init vault", "err", err)
	}
	if *testMode {
		// init JWK so Vault not required
		dmecommon.Jwks.Keys[0] = &vault.JWK{
			Secret: "secret",
		}
	}

	dmecommon.SetupMatchEngine()
	grpcOpts := make([]grpc.ServerOption, 0)

	notifyClientTls, err := nodeMgr.InternalPki.GetClientTlsConfig(ctx,
		nodeMgr.CommonName(),
		myCertIssuer,
		[]node.MatchCA{node.SameRegionalMatchCA()})
	if err != nil {
		log.FatalLog("Failed to get notify client tls config", "err", err)
	}
	notifyClientUnaryOp := notify.ClientUnaryInterceptors()
	notifyClientStreamOp := notify.ClientStreamInterceptors()
	if nodeMgr.AccessKeyClient.IsEnabled() {
		notifyClientUnaryOp = notify.ClientUnaryInterceptors(
			nodeMgr.AccessKeyClient.UnaryAddAccessKey)
		notifyClientStreamOp = notify.ClientStreamInterceptors(
			nodeMgr.AccessKeyClient.StreamAddAccessKey)
	}
	notifyClient := initNotifyClient(ctx, *notifyAddrs,
		tls.GetGrpcDialOption(notifyClientTls),
		notifyClientUnaryOp,
		notifyClientStreamOp,
	)
	sendMetric := notify.NewMetricSend()
	notifyClient.RegisterSend(sendMetric)
	sendAutoProvCounts := notify.NewAutoProvCountsSend()
	notifyClient.RegisterSend(sendAutoProvCounts)
	nodeMgr.RegisterClient(notifyClient)

	// Start autProvStats before we recieve Settings Update
	dmecommon.Settings = *edgeproto.GetDefaultSettings()
	autoProvStats := dmecommon.InitAutoProvStats(dmecommon.Settings.AutoDeployIntervalSec, 0, *statsShards, &nodeMgr.MyNode.Key, sendAutoProvCounts.Update)
	autoProvStats.Start()
	defer autoProvStats.Stop()

	notifyClient.Start()
	defer notifyClient.Stop()

	interval := time.Duration(*statsInterval) * time.Second
	stats := NewDmeStats(interval, *statsShards, sendMetric.Update)
	stats.Start()
	defer stats.Stop()

	InitAppInstClients()

	grpcOpts = append(grpcOpts,
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(dmecommon.UnaryAuthInterceptor, stats.UnaryStatsInterceptor)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(dmecommon.GetStreamInterceptor())))

	lis, err := net.Listen("tcp", *apiAddr)
	if err != nil {
		span.Finish()
		log.FatalLog("Failed to listen", "addr", *apiAddr, "err", err)
	}

	creds, err := tls.ServerAuthServerCreds(*tlsApiCertFile, *tlsApiKeyFile)
	if err != nil {
		span.Finish()
		log.FatalLog("get TLS Credentials", "error", err)
	}
	grpcOpts = append(grpcOpts, grpc.Creds(creds))
	s := grpc.NewServer(grpcOpts...)

	dme.RegisterMatchEngineApiServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			span.Finish()
			log.FatalLog("Failed to server", "err", err)
		}
	}()
	defer s.Stop()

	// REST service
	mux := http.NewServeMux()
	gwcfg := &cloudcommon.GrpcGWConfig{
		ApiAddr:     *apiAddr,
		TlsCertFile: *tlsApiCertFile,
		ApiHandles: []func(context.Context, *gwruntime.ServeMux, *grpc.ClientConn) error{
			dme.RegisterMatchEngineApiHandler,
		},
	}
	gw, err := cloudcommon.GrpcGateway(gwcfg)
	if err != nil {
		span.Finish()
		log.FatalLog("Failed to start grpc Gateway", "err", err)
	}
	mux.Handle("/", gw)
	tlscfg := &ctls.Config{
		MinVersion:               ctls.VersionTLS12,
		CurvePreferences:         []ctls.CurveID{ctls.CurveP521, ctls.CurveP384, ctls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			ctls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			ctls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			ctls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			ctls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			ctls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	// Suppress contant stream of TLS error logs due to LB health check. There is discussion in the community
	//to get rid of some of these logs, but as of now this a the way around it.   We could miss other logs here but
	// the excessive error logs are drowning out everthing else.
	var nullLogger baselog.Logger
	nullLogger.SetOutput(ioutil.Discard)

	httpServer := &http.Server{
		Addr:      *httpAddr,
		Handler:   allowCORS(mux),
		TLSConfig: tlscfg,
		ErrorLog:  &nullLogger,
	}

	go cloudcommon.GrpcGatewayServe(gwcfg, httpServer)
	defer httpServer.Shutdown(context.Background())

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	log.SpanLog(ctx, log.DebugLevelInfo, "Ready")
	span.Finish()

	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}
