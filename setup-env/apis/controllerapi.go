package apis

// interacts with the controller APIs for use by the e2e test tool

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgectl/wrapper"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	"github.com/mobiledgex/edge-cloud/testutil"
	yaml "github.com/mobiledgex/yaml/v2"
)

var appData edgeproto.AllData
var appDataMap map[string]interface{}

var apiConnectTimeout = 5 * time.Second
var apiTimeout = 30 * time.Minute

type runCommandData struct {
	Request        edgeproto.ExecRequest
	ExpectedOutput string
}

func readAppDataFile(file string) {
	err := util.ReadYamlFile(file, &appData, util.WithVars(util.DeploymentReplacementVars), util.ValidateReplacedVars())
	if err != nil {
		if !util.IsYamlOk(err, "appdata") {
			fmt.Fprintf(os.Stderr, "Error in unmarshal for file %s", file)
			os.Exit(1)
		}
	}
}

func readAppDataFileGeneric(file string) {
	err := util.ReadYamlFile(file, &appDataMap, util.WithVars(util.DeploymentReplacementVars), util.ValidateReplacedVars())
	if err != nil {
		if !util.IsYamlOk(err, "appdata") {
			fmt.Fprintf(os.Stderr, "Error in unmarshal for file %s", file)
			os.Exit(1)
		}
	}
}

func RunControllerAPI(api string, ctrlname string, apiFile string, outputDir string, mods []string, retry *bool) bool {
	runCLI := false
	for _, mod := range mods {
		if mod == "cli" {
			runCLI = true
		}
	}

	tag := ""
	apiParams := strings.Split(api, "-")
	if len(apiParams) > 1 {
		api = apiParams[0]
		tag = apiParams[1]
	}

	ctrl := util.GetController(ctrlname)
	var client testutil.Client
	if runCLI {
		args := []string{"--output-stream=false", "--silence-usage"}
		if ctrl.TLS.ClientCert != "" {
			args = append(args, "--tls", ctrl.TLS.ClientCert)
		}
		if ctrl.ApiAddr != "" {
			args = append(args, "--addr", ctrl.ApiAddr)
		}
		client = &testutil.CliClient{
			BaseArgs: args,
			RunOps: []wrapper.RunOp{
				wrapper.WithDebug(),
			},
		}
	} else {
		log.Printf("Connecting to controller %v at address %v", ctrl.Name, ctrl.ApiAddr)
		conn, err := ctrl.ConnectAPI(apiConnectTimeout)
		if err != nil {
			log.Printf("Error connecting to controller api: %v\n", ctrl.ApiAddr)
			return false
		}
		client = &testutil.ApiClient{
			Conn: conn,
		}
		defer conn.Close()
	}

	log.Printf("Applying %s via APIs for mods %v, apiFile %s\n", api, mods, apiFile)
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	rc := true
	run := testutil.NewRun(client, ctx, api, &rc)

	if api == "show" {
		filter := &edgeproto.AllData{}
		output := &edgeproto.AllData{}
		testutil.RunAllDataShowApis(run, filter, output)
		writeOutput(output, "show-commands.yml", outputDir, &rc)
		// Some objects are generated asynchronously in response to
		// other objects being created. For example, Prometheus metric
		// AppInst is created after a cluster create. Because its run
		// asynchronously, it may or may not be there before the show
		// command. Tell caller that if compareyaml fails, retry this
		// show action.
		*retry = true
	} else if api == "nodeshow" {
		filter := &edgeproto.NodeData{}
		output := &edgeproto.NodeData{}
		run.Mode = "show"
		testutil.RunNodeDataShowApis(run, filter, output)
		writeOutput(&output, "show-commands.yml", outputDir, &rc)
	} else if strings.HasPrefix(api, "debug") {
		runDebug(run, api, apiFile, outputDir)
	} else {
		if apiFile == "" {
			log.Println("Error: Cannot run controller APIs without API file")
			return false
		}

		readAppDataFile(apiFile)
		readAppDataFileGeneric(apiFile)

		switch api {
		case "delete":
			fallthrough
		case "remove":
			//run in reverse order to delete child keys
			output := &testutil.AllDataOut{}
			testutil.RunAllDataReverseApis(run, &appData, appDataMap, output)
			writeOutput(output, "api-output.yml", outputDir, &rc)
		case "create":
			fallthrough
		case "add":
			fallthrough
		case "refresh":
			fallthrough
		case "update":
			output := &testutil.AllDataOut{}
			testutil.RunAllDataApis(run, &appData, appDataMap, output)
			writeOutput(output, "api-output.yml", outputDir, &rc)
		default:
			log.Printf("Error: unsupported controller API %s\n", api)
			rc = false
		}
	}
	if tag != "expecterr" {
		// no errors expected
		failOnErrs(api, run)
	}
	return rc
}

func RunCommandAPI(api string, ctrlname string, apiFile string, outputDir string) bool {
	log.Printf("Exec %s using %s\n", api, apiFile)

	ctrl := util.GetController(ctrlname)

	data := runCommandData{}
	if apiFile == "" {
		log.Printf("Error: Cannot exec %s API without API file\n", api)
		return false
	}
	err := util.ReadYamlFile(apiFile, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in unmarshal for file %s, %v\n", apiFile, err)
		return false
	}
	req := &data.Request

	args := []string{}
	if api == "runcommand" {
		args = append(args, "RunCommand")
	}
	if api == "showlogs" {
		args = append(args, "ShowLogs")
	}
	args = append(args, "developer="+req.AppInstKey.AppKey.DeveloperKey.Name)
	args = append(args, "appname="+req.AppInstKey.AppKey.Name)
	args = append(args, "appvers="+req.AppInstKey.AppKey.Version)
	args = append(args, "cloudlet="+req.AppInstKey.ClusterInstKey.CloudletKey.Name)
	args = append(args, "operator="+req.AppInstKey.ClusterInstKey.CloudletKey.OperatorKey.Name)
	args = append(args, "cluster="+req.AppInstKey.ClusterInstKey.ClusterKey.Name)
	args = append(args, "clusterdeveloper="+req.AppInstKey.ClusterInstKey.Developer)
	if api == "runcommand" && req.Cmd != nil {
		args = append(args, "command=\""+req.Cmd.Command+"\"")
	}
	if api == "showlogs" && req.Log != nil {
		if req.Log.Since != "" {
			args = append(args, "since=\""+req.Log.Since+"\"")
		}
		if req.Log.Tail != 0 {
			args = append(args, fmt.Sprintf("tail=%d", req.Log.Tail))
		}
		if req.Log.Timestamps {
			args = append(args, "timestamps=true")
		}
		if req.Log.Follow {
			args = append(args, "follow=true")
		}
	}
	out, err := util.ControllerCLI(ctrl, args...)
	if err != nil {
		log.Printf("Error running exec %s API %v\n", api, err)
		return false
	}
	log.Printf("Exec %s output: %s\n", api, string(out))
	actual := strings.TrimSpace(string(out))
	if actual != data.ExpectedOutput {
		log.Printf("Did not get expected output: %s\n", data.ExpectedOutput)
		return false
	}

	return true
}

func StartCrmsLocal(ctx context.Context, physicalName string, apiFile string, outputDir string) error {
	if apiFile == "" {
		log.Println("Error: Cannot run RunCommand API without API file")
		return fmt.Errorf("Error: Cannot run controller APIs without API file")
	}
	readAppDataFile(apiFile)

	ctrl := util.GetController("")

	for _, c := range appData.Cloudlets {
		if c.NotifySrvAddr == "" {
			c.NotifySrvAddr = "127.0.0.1:51001"
		}

		if c.PhysicalName == "" {
			c.PhysicalName = c.Key.Name
		}

		pfConfig := edgeproto.PlatformConfig{}

		rolesfile := outputDir + "/roles.yaml"
		dat, err := ioutil.ReadFile(rolesfile)
		if err != nil {
			return err
		}
		roles := process.VaultRoles{}
		err = yaml.Unmarshal(dat, &roles)
		if err != nil {
			return err
		}
		pfConfig.EnvVar = make(map[string]string)
		pfConfig.EnvVar["VAULT_ROLE_ID"] = roles.CRMRoleID
		pfConfig.EnvVar["VAULT_SECRET_ID"] = roles.CRMSecretID

		// Defaults
		pfConfig.PlatformTag = ""
		pfConfig.TlsCertFile = ctrl.TLS.ServerCert
		pfConfig.VaultAddr = "http://127.0.0.1:8200"
		pfConfig.ContainerRegistryPath = "registry.mobiledgex.net:5000/mobiledgex/edge-cloud"
		pfConfig.TestMode = true
		pfConfig.NotifyCtrlAddrs = ctrl.NotifyAddr

		if err := cloudcommon.StartCRMService(ctx, &c, &pfConfig); err != nil {
			return err
		}
	}
	return nil
}

// Walk through all the secified cloudlets and stop CRM procecess for them
func StopCrmsLocal(ctx context.Context, physicalName string, apiFile string) error {
	if apiFile == "" {
		log.Println("Error: Cannot run RunCommand API without API file")
		return fmt.Errorf("Error: Cannot run controller APIs without API file")
	}
	readAppDataFile(apiFile)

	for _, c := range appData.Cloudlets {
		if err := cloudcommon.StopCRMService(ctx, &c); err != nil {
			return err
		}
	}
	return nil
}

func runDebug(run *testutil.Run, api, apiFile, outputDir string) {
	data := edgeproto.DebugData{}

	if apiFile == "" {
		log.Println("Error: Cannot run Debug API without API file")
		*run.Rc = false
		return
	}
	err := util.ReadYamlFile(apiFile, &data)
	if err != nil {
		log.Printf("Error in unmarshal for file %s, %v\n", apiFile, err)
		os.Exit(1)
	}

	output := testutil.DebugDataOut{}
	switch api {
	case "debugenable":
		run.Mode = "enabledebuglevels"
	case "debugdisable":
		run.Mode = "disabledebuglevels"
	case "debugshow":
		run.Mode = "show"
	case "debugrun":
		run.Mode = "rundebug"
	default:
		log.Printf("Invalid debug api %s\n", api)
		*run.Rc = false
		return
	}
	testutil.RunDebugDataApis(run, &data, make(map[string]interface{}), &output)
	writeOutput(&output, "api-output.yml", outputDir, run.Rc)
}

func writeOutput(output interface{}, fileName, outputDir string, rc *bool) {
	ymlOut, err := yaml.Marshal(output)
	if err != nil {
		log.Printf("Failed to marshal debug output, %v\n", err)
		*rc = false
	} else {
		util.PrintToFile(fileName, outputDir, string(ymlOut), true)
	}
}

func failOnErrs(api string, run *testutil.Run) {
	for _, err := range run.Errs {
		log.Printf("\"%s\" run %s failed: %s\n", api, err.Desc, err.Msg)
		*run.Rc = false
	}
}
