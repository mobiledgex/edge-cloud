package apis

// interacts with the DME APIs for use by the e2e test tool

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dmeproto "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	edgeutil "github.com/mobiledgex/edge-cloud/util"
	yaml "github.com/mobiledgex/yaml/v2"
	"google.golang.org/grpc"
)

type dmeApiRequest struct {
	Rcreq            dmeproto.RegisterClientRequest       `yaml:"registerclientrequest"`
	Fcreq            dmeproto.FindCloudletRequest         `yaml:"findcloudletrequest"`
	Pfcreq           dmeproto.PlatformFindCloudletRequest `yaml:"platformfindcloudletrequest"`
	Vlreq            dmeproto.VerifyLocationRequest       `yaml:"verifylocationrequest"`
	Glreq            dmeproto.GetLocationRequest          `yaml:"getlocationrequest"`
	Dlreq            dmeproto.DynamicLocGroupRequest      `yaml:"dynamiclocgrouprequest"`
	Aireq            dmeproto.AppInstListRequest          `yaml:"appinstlistrequest"`
	Fqreq            dmeproto.FqdnListRequest             `yaml:"fqdnlistrequest"`
	Qosreq           dmeproto.QosPositionRequest          `yaml:"qospositionrequest"`
	AppOFqreq        dmeproto.AppOfficialFqdnRequest      `yaml:"appofficialfqdnrequest"`
	Eereq            dmeproto.ClientEdgeEvent             `yaml:"clientedgeevent"`
	TokenServerPath  string                               `yaml:"token-server-path"`
	ErrorExpected    string                               `yaml:"error-expected"`
	Repeat           int                                  `yaml:"repeat"`
	CountPerInterval int                                  `yaml:"countperinterval"`
	RunAtIntervalSec float64                              `yaml:"runatintervalsec"`
	RunAtOffsetSec   float64                              `yaml:"runatoffsetsec"`
}

type registration struct {
	Req   dmeproto.RegisterClientRequest `yaml:"registerclientrequest"`
	Reply dmeproto.RegisterClientReply   `yaml:"registerclientreply"`
	At    time.Time                      `yaml:"at"`
}

type RegisterReplyWithError struct {
	dmeproto.RegisterClientReply
}

type findcloudlet struct {
	Req   dmeproto.FindCloudletRequest `yaml:"findcloudletrequest"`
	Reply dmeproto.FindCloudletReply   `yaml:"findcloudletreply"`
	At    time.Time                    `yaml:"at"`
}

var apiRequests []*dmeApiRequest
var singleRequest bool

// REST client implementation of MatchEngineApiClient interface
type dmeRestClient struct {
	client *http.Client
	addr   string
}

func NewdmeRestClient(client *http.Client, httpAddr string) dmeproto.MatchEngineApiClient {
	return &dmeRestClient{client, httpAddr}
}

func (c *dmeRestClient) RegisterClient(ctx context.Context, in *dmeproto.RegisterClientRequest, opts ...grpc.CallOption) (*dmeproto.RegisterClientReply, error) {
	out := new(dmeproto.RegisterClientReply)
	err := util.CallRESTPost("https://"+c.addr+"/v1/registerclient",
		c.client, in, out)
	if err != nil {
		log.Printf("Register rest API failed\n")
		return nil, err
	}
	return out, nil
}
func (c *dmeRestClient) FindCloudlet(ctx context.Context, in *dmeproto.FindCloudletRequest, opts ...grpc.CallOption) (*dmeproto.FindCloudletReply, error) {
	out := new(dmeproto.FindCloudletReply)
	err := util.CallRESTPost("https://"+c.addr+"/v1/findcloudlet",
		c.client, in, out)
	if err != nil {
		log.Printf("findcloudlet rest API failed\n")
		return nil, err
	}
	return out, nil
}

func (c *dmeRestClient) PlatformFindCloudlet(ctx context.Context, in *dmeproto.PlatformFindCloudletRequest, opts ...grpc.CallOption) (*dmeproto.FindCloudletReply, error) {
	out := new(dmeproto.FindCloudletReply)
	err := util.CallRESTPost("https://"+c.addr+"/v1/platformfindcloudlet",
		c.client, in, out)
	if err != nil {
		log.Printf("findcloudlet rest API failed\n")
		return nil, err
	}
	return out, nil
}

func (c *dmeRestClient) VerifyLocation(ctx context.Context, in *dmeproto.VerifyLocationRequest, opts ...grpc.CallOption) (*dmeproto.VerifyLocationReply, error) {
	out := new(dmeproto.VerifyLocationReply)
	err := util.CallRESTPost("https://"+c.addr+"/v1/verifylocation",
		c.client, in, out)
	if err != nil {
		log.Printf("verifylocation rest API failed\n")
		return nil, err
	}
	return out, nil
}

func (c *dmeRestClient) GetLocation(ctx context.Context, in *dmeproto.GetLocationRequest, opts ...grpc.CallOption) (*dmeproto.GetLocationReply, error) {
	out := new(dmeproto.GetLocationReply)
	err := util.CallRESTPost("https://"+c.addr+"/v1/getlocation",
		c.client, in, out)
	if err != nil {
		log.Printf("getlocation rest API failed\n")
		return nil, err
	}
	return out, nil
}

func (c *dmeRestClient) GetQosPositionKpi(ctx context.Context, in *dmeproto.QosPositionRequest, opts ...grpc.CallOption) (dmeproto.MatchEngineApi_GetQosPositionKpiClient, error) {
	return nil, fmt.Errorf("GetQosPositionKpi not supported yet in E2E via REST")
}

func (c *dmeRestClient) AddUserToGroup(ctx context.Context, in *dmeproto.DynamicLocGroupRequest, opts ...grpc.CallOption) (*dmeproto.DynamicLocGroupReply, error) {
	out := new(dmeproto.DynamicLocGroupReply)
	err := util.CallRESTPost("https://"+c.addr+"/v1/addusertogroup",
		c.client, in, out)
	if err != nil {
		log.Printf("addusertogroup rest API failed\n")
		return nil, err
	}
	return out, nil
}

func (c *dmeRestClient) GetFqdnList(ctx context.Context, in *dmeproto.FqdnListRequest, opts ...grpc.CallOption) (*dmeproto.FqdnListReply, error) {
	out := new(dmeproto.FqdnListReply)
	err := util.CallRESTPost("https://"+c.addr+"/v1/getfqdnlist",
		c.client, in, out)
	if err != nil {
		log.Printf("getfqdnlist rest API failed\n")
		return nil, err
	}
	return out, nil
}

func (c *dmeRestClient) GetAppInstList(ctx context.Context, in *dmeproto.AppInstListRequest, opts ...grpc.CallOption) (*dmeproto.AppInstListReply, error) {
	out := new(dmeproto.AppInstListReply)
	err := util.CallRESTPost("https://"+c.addr+"/v1/getappinstlist",
		c.client, in, out)
	if err != nil {
		log.Printf("getappinstlist rest API failed\n")
		return nil, err
	}
	return out, nil
}

func (c *dmeRestClient) GetAppOfficialFqdn(ctx context.Context, in *dmeproto.AppOfficialFqdnRequest, opts ...grpc.CallOption) (*dmeproto.AppOfficialFqdnReply, error) {
	out := new(dmeproto.AppOfficialFqdnReply)
	err := util.CallRESTPost("https://"+c.addr+"/v1/getappofficialfqdn",
		c.client, in, out)
	if err != nil {
		log.Printf("getappofficialfqdn rest API failed\n")
		return nil, err
	}
	return out, nil
}

func (c *dmeRestClient) StreamEdgeEvent(ctx context.Context, opts ...grpc.CallOption) (dmeproto.MatchEngineApi_StreamEdgeEventClient, error) {
	return nil, fmt.Errorf("StreamEdgeEvent not supported yet in E2E via REST")
}

func readDMEApiFile(apifile string, apiFileVars map[string]string) {
	err := util.ReadYamlFile(apifile, &apiRequests, util.WithVars(apiFileVars), util.ValidateReplacedVars())
	if err != nil && !util.IsYamlOk(err, "dmeapi") {
		// old yaml files are not arrayed dmeApiRequests
		apiRequest := dmeApiRequest{}
		apiRequests = append(apiRequests, &apiRequest)
		err = util.ReadYamlFile(apifile, &apiRequest, util.ValidateReplacedVars())
		singleRequest = true
	}
	if err != nil {
		if !util.IsYamlOk(err, "dmeapi") {
			fmt.Fprintf(os.Stderr, "Error in unmarshal for file %s", apifile)
			os.Exit(1)
		}
	}
}

func readMatchEngineStatus(filename string, mes *registration) {
	util.ReadYamlFile(filename, &mes)
}

func RunDmeAPI(api string, procname string, apiFile string, apiFileVars map[string]string, apiType string, outputDir string) bool {
	if apiFile == "" {
		log.Println("Error: Cannot run DME APIs without API file")
		return false
	}
	log.Printf("RunDmeAPI for api %s, %s, %s\n", api, apiFile, apiType)
	apiConnectTimeout := 5 * time.Second

	readDMEApiFile(apiFile, apiFileVars)

	dme := util.GetDme(procname)
	var client dmeproto.MatchEngineApiClient

	if apiType == "rest" {
		httpClient, err := dme.GetRestClient(apiConnectTimeout)
		if err != nil {
			log.Printf("Error: unable to connect to dme addr %v\n", dme.HttpAddr)
			return false
		}
		client = NewdmeRestClient(httpClient, dme.HttpAddr)
	} else {
		conn, err := dme.ConnectAPI(apiConnectTimeout)
		if err != nil {
			log.Printf("Error: unable to connect to dme addr %v\n", dme.ApiAddr)
			return false
		}
		defer conn.Close()
		client = dmeproto.NewMatchEngineApiClient(conn)

	}

	rc := true
	replies := make([]interface{}, 0)

	for ii, apiRequest := range apiRequests {
		if apiRequest.Repeat == 0 {
			apiRequest.Repeat = 1
		}
		if apiRequest.CountPerInterval == 0 {
			apiRequest.CountPerInterval = 1
		}
		numSecs := 1.0 + apiRequest.RunAtIntervalSec + float64(apiRequest.Repeat)
		timeout := time.Duration(float64(time.Second) * numSecs)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		log.Printf("RunDmeAPIiter[%d]\n", ii)
		ok, reply := runDmeAPIiter(ctx, api, apiFile, outputDir, apiRequest, client)
		if !ok {
			rc = false
			continue
		}
		replies = append(replies, reply)
	}
	if !rc {
		return false
	}

	var out []byte
	var ymlerror error
	if singleRequest && len(replies) == 1 {
		out, ymlerror = yaml.Marshal(replies[0])
	} else {
		out, ymlerror = yaml.Marshal(replies)
	}
	if ymlerror != nil {
		fmt.Printf("Error: Unable to marshal %s reply: %v\n", api, ymlerror)
		return false
	}
	util.PrintToFile(api+".yml", outputDir, string(out), true)
	return true
}

func runDmeAPIiter(ctx context.Context, api, apiFile, outputDir string, apiRequest *dmeApiRequest, client dmeproto.MatchEngineApiClient) (bool, interface{}) {
	//generic struct so we can do the marshal in one place even though return types are different
	var dmereply interface{}
	var dmeerror error

	sessionCookie := ""
	eeCookie := ""
	var registerStatus registration
	if api != "register" {
		//read the results from the last register so we can get the cookie.
		//if the current app is different, re-register
		readMatchEngineStatus(outputDir+"/register.yml", &registerStatus)
		if apiRequest.Rcreq.AppName != "" &&
			(registerStatus.Req.OrgName != apiRequest.Rcreq.OrgName ||
				registerStatus.Req.AppName != apiRequest.Rcreq.AppName ||
				registerStatus.Req.AppVers != apiRequest.Rcreq.AppVers ||
				time.Since(registerStatus.At) > time.Hour) {
			log.Printf("Re-registering for api %s - cached registerStatus: %+v, current Rcreq: %+v\n", api, registerStatus, apiRequest.Rcreq)
			ok, reply := runDmeAPIiter(ctx, "register", apiFile, outputDir, apiRequest, client)
			if !ok {
				return false, nil
			}
			out, ymlerror := yaml.Marshal(reply)
			if ymlerror != nil {
				fmt.Printf("Error: Unable to marshal %s reply: %v\n", api, ymlerror)
				return false, nil
			}
			util.PrintToFile("register.yml", outputDir, string(out), true)
			readMatchEngineStatus(outputDir+"/register.yml", &registerStatus)
		}
		sessionCookie = registerStatus.Reply.SessionCookie
		log.Printf("Using session cookie: %s\n", sessionCookie)

		// If StreamEdgeEvent, we need the edgeeventscookie from FindCloudletReply as well
		if api == "edgeeventinit" || api == "edgeeventlatency" || api == "edgeeventnewcloudlet" {
			var findCloudlet findcloudlet
			err := util.ReadYamlFile(outputDir+"/edgeeventfindcloudlet.yml", &findCloudlet)
			if err != nil || findCloudlet.Req.CarrierName != apiRequest.Fcreq.CarrierName ||
				findCloudlet.Req.GpsLocation.Latitude != apiRequest.Fcreq.GpsLocation.Latitude ||
				findCloudlet.Req.GpsLocation.Longitude != apiRequest.Fcreq.GpsLocation.Longitude ||
				time.Since(findCloudlet.At) > 10*time.Minute {
				log.Printf("Redoing findcloudlet for api %s - cached findCloudlet %+v, current Fcreq: %+v\n", api, findCloudlet, apiRequest.Fcreq)
				ctx = context.WithValue(ctx, "edgeevents", true)
				ok, reply := runDmeAPIiter(ctx, "findcloudlet", apiFile, outputDir, apiRequest, client)
				if !ok {
					return false, nil
				}
				out, ymlerror := yaml.Marshal(reply)
				if ymlerror != nil {
					fmt.Printf("Error: Unable to marshal %s reply: %v\n", api, ymlerror)
					return false, nil
				}
				util.PrintToFile("edgeeventfindcloudlet.yml", outputDir, string(out), true)
				util.ReadYamlFile(outputDir+"/edgeeventfindcloudlet.yml", &findCloudlet)
			}
			eeCookie = findCloudlet.Reply.EdgeEventsCookie
			log.Printf("Using eeCookie: %s\n", eeCookie)
		}
	}

	switch api {
	case "platformfindcloudlet":
		log.Printf("reading AppOfficialFqdn response to get token for platformfindcloudlet")
		var fqdnreply dmeproto.AppOfficialFqdnReply
		err := util.ReadYamlFile(outputDir+"/getappofficialfqdn.yml", &fqdnreply)
		if err != nil {
			log.Printf("error reading AppOfficialFqdn response - %v", err)
			return false, nil
		}
		apiRequest.Pfcreq.SessionCookie = sessionCookie
		apiRequest.Pfcreq.ClientToken = fqdnreply.ClientToken
		log.Printf("platformfindcloudlet using client token: %s\n", apiRequest.Pfcreq.ClientToken)
		fallthrough
	case "findcloudlet":
		apiRequest.Fcreq.SessionCookie = sessionCookie
		for ii := 0; ii < apiRequest.Repeat; ii++ {
			if apiRequest.RunAtIntervalSec != 0 {
				dur := edgeutil.GetWaitTime(time.Now(), apiRequest.RunAtIntervalSec, apiRequest.RunAtOffsetSec)
				time.Sleep(dur)
			}
			if apiRequest.Repeat != 1 {
				log.Printf("repeat interval %d of %d\n", ii+1, apiRequest.Repeat)
			}
			for jj := 0; jj < apiRequest.CountPerInterval; jj++ {
				if apiRequest.CountPerInterval != 1 {
					log.Printf("repeat interval %d count %d of %d\n", ii+1, jj+1, apiRequest.CountPerInterval)
				}
				var reply *dmeproto.FindCloudletReply
				var err error
				if api == "platformfindcloudlet" {
					log.Printf("platformfindcloudlet %v\n", apiRequest.Pfcreq)
					reply, err = client.PlatformFindCloudlet(ctx, &apiRequest.Pfcreq)
				} else {
					log.Printf("fcreq %v\n", apiRequest.Fcreq)
					reply, err = client.FindCloudlet(ctx, &apiRequest.Fcreq)
				}
				if reply != nil {
					sort.Slice(reply.Ports, func(i, j int) bool {
						return reply.Ports[i].InternalPort < reply.Ports[j].InternalPort
					})
				}
				dmereply = reply
				dmeerror = err
				if err == nil {
					_, ok := ctx.Value("edgeevents").(bool)
					if ok {
						dmereply = &findcloudlet{
							Req:   apiRequest.Fcreq,
							Reply: *reply,
							At:    time.Now(),
						}
					}
				}
			}
		}
	case "register":
		var expirySeconds int64 = 600
		if strings.Contains(apiRequest.Rcreq.AuthToken, "GENTOKEN:") {
			goPath := os.Getenv("GOPATH")
			datadir := goPath + "/" + "src/github.com/mobiledgex/edge-cloud/setup-env/e2e-tests/data"
			privKeyFile := datadir + "/" + strings.Split(apiRequest.Rcreq.AuthToken, ":")[1]
			expTime := time.Now().Add(time.Duration(expirySeconds) * time.Second).Unix()
			token, err := dmecommon.GenerateAuthToken(privKeyFile, apiRequest.Rcreq.OrgName,
				apiRequest.Rcreq.AppName, apiRequest.Rcreq.AppVers, expTime)
			if err == nil {
				log.Printf("Got AuthToken: %s\n", token)
				apiRequest.Rcreq.AuthToken = token
			} else {
				log.Printf("Error getting AuthToken: %v\n", err)
				return false, nil
			}
		}
		reply := new(dmeproto.RegisterClientReply)
		reply, dmeerror = client.RegisterClient(ctx, &apiRequest.Rcreq)
		if dmeerror == nil {
			dmereply = &registration{
				Req:   apiRequest.Rcreq,
				Reply: *reply,
				At:    time.Now(),
			}
		}
	case "verifylocation":
		tokSrvUrl := registerStatus.Reply.TokenServerUri
		log.Printf("found token server url from register response %s\n", tokSrvUrl)
		token := ""
		if tokSrvUrl == "" {
			// this is OK for the simulated case
			log.Printf("notice: no token service URL in register response")
		} else {
			//override the token server path if specified in the request.  This is used
			//for testcases like expired token
			if apiRequest.TokenServerPath != "" {
				//remove the original path and replace with the one in the test
				u, err := url.Parse(tokSrvUrl)
				if err != nil {
					log.Printf("unable to parse tokserv url %s -- %v\n", tokSrvUrl, err)
					return false, nil
				}
				u.Path = apiRequest.TokenServerPath
				tokSrvUrl = u.String()
			}
			token = GetTokenFromTokSrv(tokSrvUrl)
			if token == "" {
				log.Printf("fail to get token from token server")
				return false, nil
			}
		}
		apiRequest.Vlreq.SessionCookie = sessionCookie
		apiRequest.Vlreq.VerifyLocToken = token
		dmereply, dmeerror = client.VerifyLocation(ctx, &apiRequest.Vlreq)
	case "getappinstlist":
		// unlike the other responses, this is a slice of multiple entries which needs
		// to be sorted to allow a consistent yaml compare
		apiRequest.Aireq.SessionCookie = sessionCookie
		log.Printf("aiRequest: %+v\n", apiRequest.Aireq)
		mel, err := client.GetAppInstList(ctx, &apiRequest.Aireq)
		if err == nil {
			sort.Slice((*mel).Cloudlets, func(i, j int) bool {
				return (*mel).Cloudlets[i].CloudletName < (*mel).Cloudlets[j].CloudletName
			})
			//appinstances within the cloudlet must be sorted too
			for _, c := range (*mel).Cloudlets {
				sort.Slice(c.Appinstances, func(i, j int) bool {
					return c.Appinstances[i].AppName < c.Appinstances[j].AppName
				})
			}

		}
		dmereply = mel
		dmeerror = err
	case "getfqdnlist":
		apiRequest.Fqreq.SessionCookie = sessionCookie
		log.Printf("fqdnRequest: %+v\n", apiRequest.Fqreq)
		resp, err := client.GetFqdnList(ctx, &apiRequest.Fqreq)
		if err == nil {
			sort.Slice((*resp).AppFqdns, func(i, j int) bool {
				return (*resp).AppFqdns[i].Fqdns[0] < (*resp).AppFqdns[j].Fqdns[0]
			})
		}
		dmereply = resp
		dmeerror = err

	case "getappofficialfqdn":
		apiRequest.AppOFqreq.SessionCookie = sessionCookie
		log.Printf("AppOfficialFqdnRequest: %+v\n", apiRequest.AppOFqreq)
		resp, err := client.GetAppOfficialFqdn(ctx, &apiRequest.AppOFqreq)
		dmereply = resp
		dmeerror = err

	case "getqospositionkpi":
		apiRequest.Qosreq.SessionCookie = sessionCookie
		log.Printf("getqospositionkpi request: %+v\n", apiRequest.Qosreq)
		resp, err := client.GetQosPositionKpi(ctx, &apiRequest.Qosreq)
		if err == nil {
			dmereply, err = resp.Recv()
		}
		dmeerror = err
	case "edgeeventinit":
		apiRequest.Eereq.SessionCookie = sessionCookie
		apiRequest.Eereq.EdgeEventsCookie = eeCookie
		log.Printf("StreamEdgeEvent request: %+v\n", apiRequest.Eereq)
		resp, err := client.StreamEdgeEvent(ctx)
		if err == nil {
			// Send init request
			err = resp.Send(&apiRequest.Eereq)
			// Receive init confirmation
			dmereply, err = resp.Recv()
			if err != nil {
				dmeerror = err
				break
			}
			// Terminate persistent connection
			terminateEvent := new(dmeproto.ClientEdgeEvent)
			terminateEvent.EventType = dmeproto.ClientEdgeEvent_EVENT_TERMINATE_CONNECTION
			err = resp.Send(terminateEvent)
		}
		dmeerror = err
	case "edgeeventlatency":
		apiRequest.Eereq.SessionCookie = sessionCookie
		apiRequest.Eereq.EdgeEventsCookie = eeCookie
		log.Printf("StreamEdgeEvent request: %+v\n", apiRequest.Eereq)
		resp, err := client.StreamEdgeEvent(ctx)
		if err == nil {
			// Send init request
			err = resp.Send(&apiRequest.Eereq)
			// Receive init confirmation
			_, err = resp.Recv()
			if err != nil {
				dmeerror = err
				break
			}
			// Send dummy latency samples as Latency Event
			latencyEvent := new(dmeproto.ClientEdgeEvent)
			latencyEvent.EventType = dmeproto.ClientEdgeEvent_EVENT_LATENCY_SAMPLES
			latencyEvent.GpsLocation = &dmeproto.Loc{
				Latitude:  31.00,
				Longitude: -91.00,
			}
			latencyEvent.DeviceInfoDynamic = apiRequest.Eereq.DeviceInfoDynamic
			samples := make([]*dmeproto.Sample, 0)
			// Create dummy samples
			list := []float64{1.12, 2.354, 3.85, 4.23, 5.33}
			for i, val := range list {
				s := &dmeproto.Sample{
					Value: val,
					Timestamp: &dmeproto.Timestamp{
						Seconds: int64(i),
						Nanos:   12345,
					},
				}
				samples = append(samples, s)
			}
			latencyEvent.Samples = samples
			err = resp.Send(latencyEvent)
			// Receive processed latency samples
			dmereply, err = resp.Recv()
			if err != nil {
				dmeerror = err
				break
			}
			// Terminate persistent connection
			terminateEvent := new(dmeproto.ClientEdgeEvent)
			terminateEvent.EventType = dmeproto.ClientEdgeEvent_EVENT_TERMINATE_CONNECTION
			err = resp.Send(terminateEvent)
		}
		dmeerror = err
	case "edgeeventnewcloudlet":
		apiRequest.Eereq.SessionCookie = sessionCookie
		apiRequest.Eereq.EdgeEventsCookie = eeCookie
		log.Printf("StreamEdgeEvent request: %+v\n", apiRequest.Eereq)
		resp, err := client.StreamEdgeEvent(ctx)
		if err == nil {
			// Send init request
			err = resp.Send(&apiRequest.Eereq)
			// Receive init confirmation
			_, err = resp.Recv()
			if err != nil {
				dmeerror = err
				break
			}
			// Send dummy latency samples as Latency Event
			gpsUpdateEvent := new(dmeproto.ClientEdgeEvent)
			gpsUpdateEvent.EventType = dmeproto.ClientEdgeEvent_EVENT_LOCATION_UPDATE
			gpsUpdateEvent.GpsLocation = &dmeproto.Loc{
				Latitude:  35.00,
				Longitude: -95.00,
			}
			gpsUpdateEvent.DeviceInfoDynamic = apiRequest.Eereq.DeviceInfoDynamic
			err = resp.Send(gpsUpdateEvent)
			// Receive processed latency samples
			dmereply, err = resp.Recv()
			if err != nil {
				dmeerror = err
				break
			}
			// Terminate persistent connection
			terminateEvent := new(dmeproto.ClientEdgeEvent)
			terminateEvent.EventType = dmeproto.ClientEdgeEvent_EVENT_TERMINATE_CONNECTION
			err = resp.Send(terminateEvent)
		}
		dmeerror = err
	default:
		log.Printf("Unsupported dme api %s\n", api)
		return false, nil
	}
	if dmeerror == nil {
		// if the test is looking for an error, it needs to be there
		if apiRequest.ErrorExpected != "" {
			log.Printf("Missing error in DME API: %s", apiRequest.ErrorExpected)
			return false, nil
		}
	} else {
		// see if the error was expected
		if apiRequest.ErrorExpected != "" {
			if strings.Contains(dmeerror.Error(), apiRequest.ErrorExpected) {
				log.Printf("found expected error string in api response: %s", apiRequest.ErrorExpected)
			} else {
				log.Printf("Mismatched error in DME API: %s Expected: %s", dmeerror.Error(), apiRequest.ErrorExpected)
				return false, nil
			}
		} else {
			log.Printf("Unexpected error in DME API: %s -- %v\n", api, dmeerror)
			return false, nil
		}
	}
	log.Printf("DME REPLY %v\n", dmereply)
	return true, dmereply
}
