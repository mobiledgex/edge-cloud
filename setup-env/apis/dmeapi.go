package apis

// interacts with the DME APIs for use by the e2e test tool

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	url "net/url"

	dmeproto "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/protoc-gen-cmd/yaml"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	"google.golang.org/grpc"
)

type dmeApiRequest struct {
	MatchEngineRequest dmeproto.Match_Engine_Request `yaml:"match-engine-request"`
	TokenServerPath    string                        `yaml:"token-server-path"`
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

func readMatchEngineStatus(filename string, mes *dmeproto.Match_Engine_Status) {
	util.ReadYamlFile(filename, &mes, "", false)
}

func RunDmeAPI(api string, procname string, apiFile string, outputDir string) bool {
	if apiFile == "" {
		log.Println("Error: Cannot run DME APIs without API file")
		return false
	}
	log.Printf("RunDmeAPI for api %s\n", api)

	readMERFile(apiFile)

	dme := util.GetDme(procname)
	conn, err := grpc.Dial(dme.ApiAddr, grpc.WithInsecure())

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

	var registerStatus dmeproto.Match_Engine_Status
	if api != "register" {
		//read the results from the last register so we can get the cookie
		readMatchEngineStatus(outputDir+"/register.yml", &registerStatus)
		apiRequest.MatchEngineRequest.SessionCookie = registerStatus.SessionCookie
		log.Printf("Got session cookie from previous register: %s\n", apiRequest.MatchEngineRequest.SessionCookie)
	}

	switch api {
	case "findcloudlet":
		dmereply, dmeerror = client.FindCloudlet(ctx, &apiRequest.MatchEngineRequest)
	case "register":
		dmereply, dmeerror = client.RegisterClient(ctx, &apiRequest.MatchEngineRequest)
	case "verifylocation":
		tokSrvUrl := registerStatus.TokenServerURI
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
		apiRequest.MatchEngineRequest.VerifyLocToken = token
		dmereply, dmeerror = client.VerifyLocation(ctx, &apiRequest.MatchEngineRequest)
	case "getcloudlets":
		// unlike the other responses, this is a slice of multiple entries which needs
		// to be sorted to allow a consistent yaml compare
		log.Printf("DME REQUEST: %+v\n", apiRequest.MatchEngineRequest)
		mel, err := client.GetCloudlets(ctx, &apiRequest.MatchEngineRequest)
		sort.Slice((*mel).Cloudlets, func(i, j int) bool {
			return (*mel).Cloudlets[i].CloudletName < (*mel).Cloudlets[j].CloudletName
		})
		//appinstances within the cloudlet must be sorted too
		for _, c := range (*mel).Cloudlets {
			sort.Slice(c.Appinstances, func(i, j int) bool {
				return c.Appinstances[i].Appname < c.Appinstances[j].Appname
			})
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
