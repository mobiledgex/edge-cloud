package apis

// interacts with the DME APIs for use by the e2e test tool

import (
	"context"
	"fmt"
	"log"
	"time"

	dmeproto "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/protoc-gen-cmd/yaml"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	"google.golang.org/grpc"
)

var matchEngineRequest dmeproto.Match_Engine_Request

func readMERFile(merfile string) {
	err := util.ReadYamlFile(merfile, &matchEngineRequest, "", true)
	if err != nil {
		if !util.IsYamlOk(err, "mer") {
			log.Fatal("One or more fatal unmarshal errors, exiting")
		}
	}
}

func RunDmeAPI(api string, procname string, apiFile string, outputDir string) bool {
	if apiFile == "" {
		log.Println("Error: Cannot run DME APIs without API file")
		return false
	}

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

	switch api {
	case "findcloudlet":
		dmereply, dmeerror = client.FindCloudlet(ctx, &matchEngineRequest)
	case "verifylocation":
		dmereply, dmeerror = client.VerifyLocation(ctx, &matchEngineRequest)
	default:
		log.Printf("Unsupported dme api %s\n", api)
		return false
	}
	if dmeerror != nil {
		log.Printf("Error in find api %s -- %v\n", api, dmeerror)
		return false
	}
	out, ymlerror := yaml.Marshal(dmereply)
	if ymlerror != nil {
		fmt.Printf("Error: Unable to marshal %s reply: %v\n", api, ymlerror)
		return false
	}

	util.PrintToFile(api+".yml", outputDir, string(out), true)
	return true
}
