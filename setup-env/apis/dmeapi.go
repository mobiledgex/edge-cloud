package apis

// interacts with the DME APIs for use by the e2e test tool

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"
	"time"

	dmeproto "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/protoc-gen-cmd/yaml"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
)

type dmeApiRequest struct {
	Rcreq           dmeproto.RegisterClientRequest  `yaml:"registerclientrequest"`
	Fcreq           dmeproto.FindCloudletRequest    `yaml:"findcloudletrequest"`
	Vlreq           dmeproto.VerifyLocationRequest  `yaml:"verifylocationrequest"`
	Glreq           dmeproto.GetLocationRequest     `yaml:"getlocationrequest"`
	Dlreq           dmeproto.DynamicLocGroupRequest `yaml:"dynamiclocgrouprequest"`
	Aireq           dmeproto.AppInstListRequest     `yaml:"appinstlistrequest"`
	TokenServerPath string                          `yaml:"token-server-path"`
}

type registration struct {
	Req   dmeproto.RegisterClientRequest `yaml:"registerclientrequest"`
	Reply dmeproto.RegisterClientReply   `yaml:"registerclientreply"`
}

var apiRequest dmeApiRequest

func readMERFile(merfile string) {
	err := util.ReadYamlFile(merfile, &apiRequest, "", true)
	if err != nil {
		if !util.IsYamlOk(err, "mer") {
			fmt.Fprintf(os.Stderr, "Error in unmarshal for file %s", merfile)
			os.Exit(1)
		}
	}
}

func readMatchEngineStatus(filename string, mes *registration) {
	util.ReadYamlFile(filename, &mes, "", false)
}

func RunDmeAPI(api string, procname string, apiFile string, outputDir string) bool {
	if apiFile == "" {
		log.Println("Error: Cannot run DME APIs without API file")
		return false
	}
	log.Printf("RunDmeAPI for api %s, %s\n", api, apiFile)
	apiConnectTimeout := 5 * time.Second

	readMERFile(apiFile)

	dme := util.GetDme(procname)
	conn, err := dme.DmeLocal.ConnectAPI(apiConnectTimeout)
	if err != nil {
		log.Printf("Error: unable to connect to dme addr %v\n", dme.ApiAddr)
		return false
	}
	defer conn.Close()
	client := dmeproto.NewMatch_Engine_ApiClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	//generic struct so we can do the marshal in one place even though return types are different
	var dmereply interface{}
	var dmeerror error

	sessionCookie := ""
	var registerStatus registration
	if api != "register" {
		//read the results from the last register so we can get the cookie.
		//if the current app is different, re-register
		readMatchEngineStatus(outputDir+"/register.yml", &registerStatus)
		if registerStatus.Req.DevName != apiRequest.Rcreq.DevName ||
			registerStatus.Req.AppName != apiRequest.Rcreq.AppName ||
			registerStatus.Req.AppVers != apiRequest.Rcreq.AppVers {
			log.Printf("Re-registering for api %s\n", api)
			ok := RunDmeAPI("register", procname, apiFile, outputDir)
			if !ok {
				return false
			}
			readMatchEngineStatus(outputDir+"/register.yml", &registerStatus)
		}
		sessionCookie = registerStatus.Reply.SessionCookie
		log.Printf("Using session cookie: %s\n", sessionCookie)
	}

	switch api {
	case "findcloudlet":
		apiRequest.Fcreq.SessionCookie = sessionCookie
		fc, err := client.FindCloudlet(ctx, &apiRequest.Fcreq)
		sort.Slice(fc.Ports, func(i, j int) bool {
			return fc.Ports[i].InternalPort < fc.Ports[j].InternalPort
		})
		dmereply = fc
		dmeerror = err
	case "register":
		reply := new(dmeproto.RegisterClientReply)
		reply, dmeerror = client.RegisterClient(ctx, &apiRequest.Rcreq)
		dmereply = &registration{
			Req:   apiRequest.Rcreq,
			Reply: *reply,
		}
	case "verifylocation":
		tokSrvUrl := registerStatus.Reply.TokenServerURI
		log.Printf("found token server url from register response %s\n", tokSrvUrl)

		if tokSrvUrl == "" {
			log.Printf("no token service URL in setup")
			return false
		}
		//override the token server path if specified in the request.  This is used
		//for testcases like expired token
		if apiRequest.TokenServerPath != "" {
			//remove the original path and replace with the one in the test
			u, err := url.Parse(tokSrvUrl)
			if err != nil {
				log.Printf("unable to parse tokserv url %s -- %v\n", tokSrvUrl, err)
				return false
			}
			u.Path = apiRequest.TokenServerPath
			tokSrvUrl = u.String()
		}
		token := GetTokenFromTokSrv(tokSrvUrl)
		if token == "" {
			return false
		}
		apiRequest.Vlreq.SessionCookie = sessionCookie
		apiRequest.Vlreq.VerifyLocToken = token
		dmereply, dmeerror = client.VerifyLocation(ctx, &apiRequest.Vlreq)
	case "getappinstlist":
		// unlike the other responses, this is a slice of multiple entries which needs
		// to be sorted to allow a consistent yaml compare
		apiRequest.Aireq.SessionCookie = sessionCookie
		log.Printf("DME REQUEST: %+v\n", apiRequest.Aireq)
		mel, err := client.GetAppInstList(ctx, &apiRequest.Aireq)
		if err == nil {
			sort.Slice((*mel).Cloudlets, func(i, j int) bool {
				return (*mel).Cloudlets[i].CloudletName < (*mel).Cloudlets[j].CloudletName
			})
			//appinstances within the cloudlet must be sorted too
			for _, c := range (*mel).Cloudlets {
				sort.Slice(c.Appinstances, func(i, j int) bool {
					return c.Appinstances[i].Appname < c.Appinstances[j].Appname
				})
			}

		}
		dmereply = mel
		dmeerror = err
	default:
		log.Printf("Unsupported dme api %s\n", api)
		return false
	}
	if dmeerror != nil {
		log.Printf("Error in dme api %s -- %v\n", api, dmeerror)
		return false
	}

	log.Printf("DME REPLY %s\n", dmereply)
	out, ymlerror := yaml.Marshal(dmereply)
	if ymlerror != nil {
		fmt.Printf("Error: Unable to marshal %s reply: %v\n", api, ymlerror)
		return false
	}

	util.PrintToFile(api+".yml", outputDir, string(out), true)
	return true
}

func RunDmeRest(api string, procname string, apiFile string, outputDir string) bool {
	if apiFile == "" {
		log.Println("Error: Cannot run DME RESTs without API file")
		return false
	}
	log.Printf("RunDmeRest for api %s file %s\n", api, apiFile)
	restConnectTimeout := 5 * time.Second

	readMERFile(apiFile)

	dme := util.GetDme(procname)
	client, err := dme.DmeLocal.GetRestClient(restConnectTimeout)
	if err != nil {
		log.Printf("Error: unable to connect to dme addr %v\n", dme.HttpAddr)
		return false
	}

	//generic struct so we can do the marshal in one place even though return types are different
	var dmereply interface{}
	var dmeerror error

	sessionCookie := ""
	var registerStatus registration
	if api != "register" {
		//read the results from the last register so we can get the cookie.
		//if the current app is different, re-register
		readMatchEngineStatus(outputDir+"/register.yml", &registerStatus)
		if registerStatus.Req.DevName != apiRequest.Rcreq.DevName ||
			registerStatus.Req.AppName != apiRequest.Rcreq.AppName ||
			registerStatus.Req.AppVers != apiRequest.Rcreq.AppVers {
			log.Printf("Re-registering for api %s\n", api)
			ok := RunDmeRest("register", procname, apiFile, outputDir)
			if !ok {
				return false
			}
			readMatchEngineStatus(outputDir+"/register.yml", &registerStatus)
		}
		sessionCookie = registerStatus.Reply.SessionCookie
		log.Printf("Using session cookie: %s\n", sessionCookie)
	}

	switch api {
	case "findcloudlet":
		apiRequest.Fcreq.SessionCookie = sessionCookie
		reply := new(dmeproto.FindCloudletReply)
		err = util.CallRESTPost("https://"+dme.DmeLocal.HttpAddr+"/v1/"+api,
			client, &apiRequest.Fcreq, reply)
		if err != nil {
			log.Printf("findcoudlet rest API failed\n")
			return false
		}
		dmereply = reply
	case "register":
		reply := new(dmeproto.RegisterClientReply)
		err = util.CallRESTPost("https://"+dme.DmeLocal.HttpAddr+"/v1/registerclient",
			client, &apiRequest.Rcreq, reply)
		if err != nil {
			log.Printf("Register rest API failed\n")
			return false
		}
		dmereply = &registration{
			Req:   apiRequest.Rcreq,
			Reply: *reply,
		}
	case "verifylocation":
		tokSrvUrl := registerStatus.Reply.TokenServerURI
		log.Printf("found token server url from register response %s\n", tokSrvUrl)

		if tokSrvUrl == "" {
			log.Printf("no token service URL in setup")
			return false
		}
		//override the token server path if specified in the request.  This is used
		//for testcases like expired token
		if apiRequest.TokenServerPath != "" {
			//remove the original path and replace with the one in the test
			u, err := url.Parse(tokSrvUrl)
			if err != nil {
				log.Printf("unable to parse tokserv url %s -- %v\n", tokSrvUrl, err)
				return false
			}
			u.Path = apiRequest.TokenServerPath
			tokSrvUrl = u.String()
		}
		token := GetTokenFromTokSrv(tokSrvUrl)
		if token == "" {
			return false
		}
		apiRequest.Vlreq.SessionCookie = sessionCookie
		apiRequest.Vlreq.VerifyLocToken = token
		apiRequest.Fcreq.SessionCookie = sessionCookie

		reply := new(dmeproto.VerifyLocationReply)
		err = util.CallRESTPost("https://"+dme.DmeLocal.HttpAddr+"/v1/"+api,
			client, &apiRequest.Vlreq, reply)
		if err != nil {
			log.Printf("verifylocation rest API failed\n")
			return false
		}
		dmereply = reply
	case "getappinstlist":
		// unlike the other responses, this is a slice of multiple entries which needs
		// to be sorted to allow a consistent yaml compare
		apiRequest.Aireq.SessionCookie = sessionCookie
		log.Printf("DME REQUEST: %+v\n", apiRequest.Aireq)
		mel := new(dmeproto.AppInstListReply)
		err = util.CallRESTPost("https://"+dme.DmeLocal.HttpAddr+"/v1/"+api,
			client, &apiRequest.Aireq, mel)
		if err != nil {
			log.Printf("verifylocation rest API failed\n")
			return false
		}
		if err == nil {
			sort.Slice((*mel).Cloudlets, func(i, j int) bool {
				return (*mel).Cloudlets[i].CloudletName < (*mel).Cloudlets[j].CloudletName
			})
			//appinstances within the cloudlet must be sorted too
			for _, c := range (*mel).Cloudlets {
				sort.Slice(c.Appinstances, func(i, j int) bool {
					return c.Appinstances[i].Appname < c.Appinstances[j].Appname
				})
			}
		}
		dmereply = mel
		dmeerror = err
	default:
		log.Printf("Unsupported dme api %s\n", api)
		return false
	}
	if dmeerror != nil {
		log.Printf("Error in dme api %s -- %v\n", api, dmeerror)
		return false
	}

	log.Printf("DME REPLY %s\n", dmereply)
	out, ymlerror := yaml.Marshal(dmereply)
	if ymlerror != nil {
		fmt.Printf("Error: Unable to marshal %s reply: %v\n", api, ymlerror)
		return false
	}

	util.PrintToFile(api+".yml", outputDir, string(out), true)
	return true
}
