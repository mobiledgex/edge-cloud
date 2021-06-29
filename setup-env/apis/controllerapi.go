package apis

// interacts with the controller APIs for use by the e2e test tool

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgectl/wrapper"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	"github.com/mobiledgex/edge-cloud/testutil"
	uutil "github.com/mobiledgex/edge-cloud/util"
)

var appData edgeproto.AllData
var appDataMap map[string]interface{}

var apiConnectTimeout = 5 * time.Second
var apiTimeout = 30 * time.Minute

type runCommandData struct {
	Request        edgeproto.ExecRequest
	ExpectedOutput string
}

func readAppDataFile(file string, vars map[string]string) {
	vars = uutil.AddMaps(util.DeploymentReplacementVars, vars)
	err := util.ReadYamlFile(file, &appData, util.WithVars(vars), util.ValidateReplacedVars())
	if err != nil {
		if !util.IsYamlOk(err, "appdata") {
			fmt.Fprintf(os.Stderr, "Error in unmarshal for file %s", file)
			os.Exit(1)
		}
	}
}

func readAppDataFileGeneric(file string, vars map[string]string) {
	vars = uutil.AddMaps(util.DeploymentReplacementVars, vars)
	err := util.ReadYamlFile(file, &appDataMap, util.WithVars(vars), util.ValidateReplacedVars())
	if err != nil {
		if !util.IsYamlOk(err, "appdata") {
			fmt.Fprintf(os.Stderr, "Error in unmarshal for file %s", file)
			os.Exit(1)
		}
	}
}

func RunControllerAPI(api string, ctrlname string, apiFile string, apiFileVars map[string]string, outputDir string, mods []string, retry *bool) bool {
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
		tlsFile := ctrl.GetTlsFile()
		if tlsFile != "" {
			args = append(args, "--tls", tlsFile)
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
		util.PrintToYamlFile("show-commands.yml", outputDir, output, true)
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
		util.PrintToYamlFile("show-commands.yml", outputDir, &output, true)
	} else if strings.HasPrefix(api, "debug") {
		runDebug(run, api, apiFile, apiFileVars, outputDir)
	} else if api == "deviceshow" {
		output := &edgeproto.DeviceData{}
		run.Mode = "show"
		run.DeviceApi(nil, nil, &output.Devices)
		util.PrintToYamlFile("show-commands.yml", outputDir, &output, true)
	} else if strings.HasPrefix(api, "organization") {
		runOrg(run, api, apiFile, apiFileVars, outputDir)
	} else {
		if apiFile == "" {
			log.Println("Error: Cannot run controller APIs without API file")
			return false
		}

		readAppDataFile(apiFile, apiFileVars)
		readAppDataFileGeneric(apiFile, apiFileVars)

		switch api {
		case "delete":
			fallthrough
		case "remove":
			//run in reverse order to delete child keys
			output := &testutil.AllDataOut{}
			testutil.RunAllDataReverseApis(run, &appData, appDataMap, output, testutil.NoApiCallback)
			util.PrintToYamlFile("api-output.yml", outputDir, output, true)
		case "create":
			fallthrough
		case "add":
			fallthrough
		case "refresh":
			fallthrough
		case "update":
			output := &testutil.AllDataOut{}
			testutil.RunAllDataApis(run, &appData, appDataMap, output, testutil.NoApiCallback)
			util.PrintToYamlFile("api-output.yml", outputDir, output, true)
		case "stream":
			output := &testutil.AllDataStreamOut{}
			testutil.RunAllDataStreamApis(run, &appData, output)
			util.PrintToYamlFile("show-commands.yml", outputDir, output, true)
		case "showfiltered":
			output := &edgeproto.AllData{}
			testutil.RunAllDataShowApis(run, &appData, output)
			util.PrintToYamlFile("show-commands.yml", outputDir, output, true)
			*retry = true
		default:
			log.Printf("Error: unsupported controller API %s\n", api)
			rc = false
		}
	}
	run.CheckErrs(api, tag)
	return rc
}

func RunCommandAPI(api string, ctrlname string, apiFile string, apiFileVars map[string]string, outputDir string) bool {
	log.Printf("Exec %s using %s\n", api, apiFile)

	ctrl := util.GetController(ctrlname)

	data := runCommandData{}
	if apiFile == "" {
		log.Printf("Error: Cannot exec %s API without API file\n", api)
		return false
	}
	err := util.ReadYamlFile(apiFile, &data, util.WithVars(apiFileVars))
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
	if api == "runconsole" {
		args = append(args, "RunConsole")
	}
	if api == "accesscloudlet" {
		args = append(args, "AccessCloudlet")
	}
	args = append(args, "app-org="+req.AppInstKey.AppKey.Organization)
	args = append(args, "appname="+req.AppInstKey.AppKey.Name)
	args = append(args, "appvers="+req.AppInstKey.AppKey.Version)
	args = append(args, "cloudlet="+req.AppInstKey.ClusterInstKey.CloudletKey.Name)
	args = append(args, "cloudlet-org="+req.AppInstKey.ClusterInstKey.CloudletKey.Organization)
	args = append(args, "cluster="+req.AppInstKey.ClusterInstKey.ClusterKey.Name)
	args = append(args, "cluster-org="+req.AppInstKey.ClusterInstKey.Organization)
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
	if api == "accesscloudlet" && req.Cmd != nil {
		args = append(args, "command=\""+req.Cmd.Command+"\"")
		if req.Cmd.CloudletMgmtNode != nil {
			args = append(args, "node-type=\""+req.Cmd.CloudletMgmtNode.Type+"\"")
			args = append(args, "node-name=\""+req.Cmd.CloudletMgmtNode.Name+"\"")
		}
	}
	out, err := util.ControllerCLI(ctrl, args...)
	if err != nil {
		log.Printf("Error running exec %s API %v\n", api, err)
		return false
	}
	log.Printf("Exec %s output: %s\n", api, string(out))
	actual := strings.TrimSpace(string(out))
	if api != "runconsole" {
		if actual != data.ExpectedOutput {
			log.Printf("Did not get expected output: %s\n", data.ExpectedOutput)
			return false
		}
	} else {
		content, err := util.ReadConsoleURL(actual, nil)
		if err != nil {
			log.Printf("Error fetching contents from %s for %s API %v\n", actual, api, err)
			return false
		}
		actualContent := strings.TrimSpace(content)
		if actualContent != data.ExpectedOutput {
			log.Printf("Did not get expected output from console URL %s: \"%s\" (expected) \"%s\" (actual)\n", actual, data.ExpectedOutput, actualContent)
			return false
		}
	}

	return true
}

func StartCrmsLocal(ctx context.Context, physicalName string, apiFile string, apiFileVars map[string]string, outputDir string) error {
	if apiFile == "" {
		log.Println("Error: Cannot run RunCommand API without API file")
		return fmt.Errorf("Error: Cannot run controller APIs without API file")
	}
	readAppDataFile(apiFile, apiFileVars)

	ctrl := util.GetController("")

	for _, c := range appData.Cloudlets {
		if c.NotifySrvAddr == "" {
			c.NotifySrvAddr = "127.0.0.1:51001"
		}

		if c.PhysicalName == "" {
			c.PhysicalName = c.Key.Name
		}
		c.ContainerVersion = ctrl.VersionTag

		pfConfig := edgeproto.PlatformConfig{}
		region := ctrl.Region
		if region == "" {
			region = "local"
		}
		pfConfig.EnvVar = make(map[string]string)

		// Defaults
		pfConfig.PlatformTag = ""
		pfConfig.TlsCertFile = ctrl.TLS.ServerCert
		pfConfig.TlsKeyFile = ctrl.TLS.ServerKey
		pfConfig.TlsCaFile = ctrl.TLS.CACert
		pfConfig.UseVaultPki = ctrl.UseVaultPki
		pfConfig.ContainerRegistryPath = "registry.mobiledgex.net:5000/mobiledgex/edge-cloud"
		pfConfig.TestMode = true
		pfConfig.NotifyCtrlAddrs = ctrl.NotifyAddr
		pfConfig.DeploymentTag = ctrl.DeploymentTag
		pfConfig.AccessApiAddr = ctrl.AccessApiAddr
		for k, v := range ctrl.Common.EnvVars {
			pfConfig.EnvVar[k] = v
		}

		if err := cloudcommon.StartCRMService(ctx, &c, &pfConfig); err != nil {
			return err
		}
	}
	return nil
}

// Walk through all the secified cloudlets and stop CRM procecess for them
func StopCrmsLocal(ctx context.Context, physicalName string, apiFile string, apiFileVars map[string]string) error {
	if apiFile == "" {
		log.Println("Error: Cannot run RunCommand API without API file")
		return fmt.Errorf("Error: Cannot run controller APIs without API file")
	}
	readAppDataFile(apiFile, apiFileVars)

	for _, c := range appData.Cloudlets {
		if err := cloudcommon.StopCRMService(ctx, &c); err != nil {
			return err
		}
	}
	return nil
}

func runDebug(run *testutil.Run, api, apiFile string, apiFileVars map[string]string, outputDir string) {
	data := edgeproto.DebugData{}

	if apiFile == "" {
		log.Println("Error: Cannot run Debug API without API file")
		*run.Rc = false
		return
	}
	err := util.ReadYamlFile(apiFile, &data, util.WithVars(apiFileVars))
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
	testutil.RunDebugDataApis(run, &data, make(map[string]interface{}), &output, testutil.NoApiCallback)
	util.PrintToYamlFile("api-output.yml", outputDir, &output, true)
}

func runOrg(run *testutil.Run, api, apiFile string, apiFileVars map[string]string, outputDir string) {
	data := edgeproto.OrganizationData{}

	if apiFile == "" {
		log.Println("Error: Cannot run Org API without API file")
		*run.Rc = false
		return
	}
	err := util.ReadYamlFile(apiFile, &data, util.WithVars(apiFileVars))
	if err != nil {
		log.Printf("Error in unmarshal for file %s, %v\n", apiFile, err)
		os.Exit(1)
	}

	output := testutil.OrganizationDataOut{}
	testutil.RunOrganizationDataApis(run, &data, make(map[string]interface{}), &output, testutil.NoApiCallback)
	util.PrintToYamlFile("api-output.yml", outputDir, &output, true)
}
