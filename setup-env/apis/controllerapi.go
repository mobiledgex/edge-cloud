package apis

// interacts with the controller APIs for use by the e2e test tool

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	"github.com/mobiledgex/edge-cloud/testutil"
	yaml "github.com/mobiledgex/yaml/v2"
	"google.golang.org/grpc"
)

var appData edgeproto.ApplicationData
var appDataMap edgeproto.ApplicationDataMap

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

func runShow(ctrl *process.Controller, showCmds []string, outputDir string) bool {
	errFound := false
	for i, c := range showCmds {
		label := strings.Split(c, " ")[0]
		cmdstr := strings.Split(c, " ")[1]
		var cmdargs = []string{cmdstr}
		log.Printf("generating output for %s\n", label)
		out, _ := util.ControllerCLI(ctrl, cmdargs...)
		truncate := false
		//truncate the file for the first command output, afterwards append
		if i == 0 {
			truncate = true
		}
		//edgectl returns exitcode 0 even if it cannot connect, so look for the error
		if strings.Contains(string(out), cmdstr+" failed") {
			log.Printf("Found failure in show output\n")
			errFound = true
		}
		// contents under label needs to be indented for non-lists
		buf := bytes.Buffer{}
		for _, line := range strings.Split(string(out), "\n") {
			if line != "" {
				buf.WriteString("  ")
				buf.WriteString(line)
			}
			buf.WriteString("\n")
		}
		util.PrintToFile("show-commands.yml", outputDir, label+"\n"+buf.String()+"\n", truncate)
	}
	return !errFound
}

func runShowCommands(ctrl *process.Controller, outputDir string) bool {
	var showCmds = []string{
		"settings: ShowSettings",
		"flavors: ShowFlavor",
		"clusterinsts: ShowClusterInst",
		"operatorcodes: ShowOperatorCode",
		"cloudlets: ShowCloudlet",
		"apps: ShowApp",
		"appinstances: ShowAppInst",
		"autoscalepolicies: ShowAutoScalePolicy",
		"autoprovpolicies: ShowAutoProvPolicy",
		"privacypolicies: ShowPrivacyPolicy",
	}
	// Some objects are generated asynchronously in response to
	// other objects being created. For example, Prometheus metric
	// AppInst is created after a cluster create. Because its run
	// asynchronously, it may or may not be there before the show
	// command. So if show fails, we retry a few times to see
	// these objects show up a little later.
	tries := 10
	for ii := 0; ii < tries; ii++ {
		if ii != 0 {
			time.Sleep(100 * time.Millisecond)
		}
		ret := runShow(ctrl, showCmds, outputDir)
		if ret {
			return true
		}
	}
	return false
}

func runNodeShow(ctrl *process.Controller, outputDir string) bool {
	var showCmds = []string{
		"nodes: ShowNode",
	}
	return runShow(ctrl, showCmds, outputDir)
}

func RunControllerAPI(api string, ctrlname string, apiFile string, outputDir string, mods []string) bool {
	runCLI := false
	for _, mod := range mods {
		if mod == "cli" {
			runCLI = true
		}
	}
	if strings.HasPrefix(api, "debug") {
		// no cli support for now
		runCLI = false
	}

	if runCLI {
		return RunControllerCLI(api, ctrlname, apiFile, outputDir, mods)
	}

	log.Printf("Applying data via APIs for %s\n", apiFile)

	ctrl := util.GetController(ctrlname)

	if api == "show" {
		//handled separately since it uses edgectl not direct api connection
		return runShowCommands(ctrl, outputDir)
	} else if api == "nodeshow" {
		return runNodeShow(ctrl, outputDir)
	} else if strings.HasPrefix(api, "debug") {
		return runDebug(ctrl, api, apiFile, outputDir)
	}

	if apiFile == "" {
		log.Println("Error: Cannot run controller APIs without API file")
		return false
	}

	readAppDataFile(apiFile)
	readAppDataFileGeneric(apiFile)

	log.Printf("Connecting to controller %v at address %v", ctrl.Name, ctrl.ApiAddr)
	ctrlapi, err := ctrl.ConnectAPI(apiConnectTimeout)

	rc := true

	if err != nil {
		log.Printf("Error connecting to controller api: %v\n", ctrl.ApiAddr)
		return false
	} else {
		log.Printf("Connected to controller %v success", ctrl.Name)
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)

		var err error
		switch api {
		case "delete":
			fallthrough
		case "remove":
			//run in reverse order to delete child keys
			err = testutil.RunAppInstApi(ctrlapi, ctx, &appData.AppInstances, appDataMap["appinstances"], api)
			if err != nil {
				log.Printf("Error in appinst API %v \n", err)
				rc = false
			}
			err = testutil.RunClusterInstApi(ctrlapi, ctx, &appData.ClusterInsts, appDataMap["clusterinsts"], api)
			if err != nil {
				log.Printf("Error in clusterinst API %v\n", err)
				rc = false
			}
			err = testutil.RunAppApi(ctrlapi, ctx, &appData.Applications, appDataMap["apps"], api)
			if err != nil {
				log.Printf("Error in app API %v\n", err)
				rc = false
			}
			err = testutil.RunAutoScalePolicyApi(ctrlapi, ctx, &appData.AutoScalePolicies, appDataMap["autoscalepolicies"], api)
			if err != nil {
				log.Printf("Error in auto scale policy API %v\n", err)
				rc = false
			}
			err = testutil.RunAutoProvPolicyApi_AutoProvPolicyCloudlet(ctrlapi, ctx, &appData.AutoProvPolicyCloudlets, appDataMap["autoprovpolicycloudlets"], api)
			if err != nil {
				log.Printf("Error in auto prov policy cloudlet API %v\n", err)
				rc = false
			}
			err = testutil.RunAutoProvPolicyApi(ctrlapi, ctx, &appData.AutoProvPolicies, appDataMap["autoprovpolicies"], api)
			if err != nil {
				log.Printf("Error in auto prov policy API %v\n", err)
				rc = false
			}
			err = testutil.RunPrivacyPolicyApi(ctrlapi, ctx, &appData.PrivacyPolicies, appDataMap["privacypolicies"], api)
			if err != nil {
				log.Printf("Error in privacy policy API %v\n", err)
				rc = false
			}
			err = testutil.RunCloudletInfoApi(ctrlapi, ctx, &appData.CloudletInfos, appDataMap["cloudletinfos"], api)
			if err != nil {
				log.Printf("Error in cloudletInfo API %v\n", err)
				rc = false
			}
			err = testutil.RunCloudletApi(ctrlapi, ctx, &appData.Cloudlets, appDataMap["cloudlets"], api)
			if err != nil {
				log.Printf("Error in cloudlet API %v\n", err)
				rc = false
			}
			err = testutil.RunOperatorCodeApi(ctrlapi, ctx, &appData.OperatorCodes, appDataMap["operatorcodes"], api)
			if err != nil {
				log.Printf("Error in operator code API %v\n", err)
				rc = false
			}
			err = testutil.RunSettingsApi(ctrlapi, ctx, appData.Settings, appDataMap["settings"], "reset")
			if err != nil {
				log.Printf("Error in settings API %v\n", err)
				rc = false
			}
			err = testutil.RunFlavorApi(ctrlapi, ctx, &appData.Flavors, appDataMap["flavors"], api)
			if err != nil {
				log.Printf("Error in flavor API %v\n", err)
				rc = false
			}
		case "create":
			fallthrough
		case "add":
			fallthrough
		case "refresh":
			fallthrough
		case "update":
			err = testutil.RunFlavorApi(ctrlapi, ctx, &appData.Flavors, appDataMap["flavors"], api)
			if err != nil {
				log.Printf("Error in flavor API %v\n", err)
				rc = false
			}
			err = testutil.RunSettingsApi(ctrlapi, ctx, appData.Settings, appDataMap["settings"], "update")
			if err != nil {
				log.Printf("Error in settigs API %v\n", err)
				rc = false
			}
			err = testutil.RunOperatorCodeApi(ctrlapi, ctx, &appData.OperatorCodes, appDataMap["operatorcodes"], api)
			if err != nil {
				log.Printf("Error in operator code API %v\n", err)
				rc = false
			}
			err = testutil.RunCloudletApi(ctrlapi, ctx, &appData.Cloudlets, appDataMap["cloudlets"], api)
			if err != nil {
				log.Printf("Error in cloudlet API %v\n", err)
				rc = false
			}
			err = testutil.RunCloudletInfoApi(ctrlapi, ctx, &appData.CloudletInfos, appDataMap["cloudletinfos"], api)
			if err != nil {
				log.Printf("Error in cloudletInfo API %v\n", err)
				rc = false
			}
			err = testutil.RunAutoProvPolicyApi(ctrlapi, ctx, &appData.AutoProvPolicies, appDataMap["autoprovpolicies"], api)
			if err != nil {
				log.Printf(" %v\n", err)
				rc = false
			}
			err = testutil.RunAutoProvPolicyApi_AutoProvPolicyCloudlet(ctrlapi, ctx, &appData.AutoProvPolicyCloudlets, appDataMap["autoprovpolicycloudlets"], api)
			if err != nil {
				log.Printf("Error in auto prov policy cloudlet API %v\n", err)
				rc = false
			}
			err = testutil.RunAutoScalePolicyApi(ctrlapi, ctx, &appData.AutoScalePolicies, appDataMap["autoscalepolicies"], api)
			if err != nil {
				log.Printf("Error in auto scale policy API %v\n", err)
				rc = false
			}
			err = testutil.RunPrivacyPolicyApi(ctrlapi, ctx, &appData.PrivacyPolicies, appDataMap["privacypolicies"], api)
			if err != nil {
				log.Printf("Error in privacy policy API %v\n", err)
				rc = false
			}
			err = testutil.RunAppApi(ctrlapi, ctx, &appData.Applications, appDataMap["apps"], api)
			if err != nil {
				log.Printf("Error in app API %v\n", err)
				rc = false
			}
			err = testutil.RunClusterInstApi(ctrlapi, ctx, &appData.ClusterInsts, appDataMap["clusterinsts"], api)
			if err != nil {
				log.Printf("Error in clusterinst API %v\n", err)
				rc = false
			}
			err = testutil.RunAppInstApi(ctrlapi, ctx, &appData.AppInstances, appDataMap["appinstances"], api)
			if err != nil {
				log.Printf("Error in appinst API %v\n", err)
				rc = false
			}
		default:
			log.Printf("Error: unsupported controller API %s\n", api)
			rc = false
		}
		cancel()
	}
	ctrlapi.Close()
	return rc
}

func RunControllerCLI(api string, ctrlname string, apiFile string, outputDir string, mods []string) bool {
	log.Printf("Applying data via CLI for %s\n", apiFile)

	ctrl := util.GetController(ctrlname)

	if api == "show" {
		return runShowCommands(ctrl, outputDir)
	}
	if api == "nodeshow" {
		return runNodeShow(ctrl, outputDir)
	}

	if apiFile == "" {
		log.Println("Error: Cannot run controller APIs without API file")
		return false
	}

	log.Printf("Using controller %v at address %v", ctrl.Name, ctrl.ApiAddr)
	switch api {
	case "create":
		out, err := util.ControllerCLI(ctrl, "Create", "--datafile", apiFile)
		log.Println(string(out))
		if err != nil {
			log.Printf("Error running Create CLI %v\n", err)
			return false
		}
	case "delete":
		out, err := util.ControllerCLI(ctrl, "Delete", "--datafile", apiFile)
		log.Println(string(out))
		if err != nil {
			log.Printf("Error running Delete CLI %v\n", err)
			return false
		}
	default:
		log.Printf("Error: unsupported controller CLI %s\n", api)
		return false
	}
	return true
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
	args = append(args, "organization="+req.AppInstKey.AppKey.Organization)
	args = append(args, "appname="+req.AppInstKey.AppKey.Name)
	args = append(args, "appvers="+req.AppInstKey.AppKey.Version)
	args = append(args, "cloudlet="+req.AppInstKey.ClusterInstKey.CloudletKey.Name)
	args = append(args, "operatororg="+req.AppInstKey.ClusterInstKey.CloudletKey.Organization)
	args = append(args, "cluster="+req.AppInstKey.ClusterInstKey.ClusterKey.Name)
	args = append(args, "clusterdevorg="+req.AppInstKey.ClusterInstKey.Organization)
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

func runDebug(ctrl *process.Controller, api, apiFile, outputDir string) bool {
	data := util.DebugData{}
	log.Printf("Applying debug via APIs for %s\n", apiFile)

	if apiFile == "" {
		log.Println("Error: Cannot run Debug API without API file")
		return false
	}
	err := util.ReadYamlFile(apiFile, &data)
	if err != nil {
		log.Printf("Error in unmarshal for file %s, %v\n", apiFile, err)
		os.Exit(1)
	}

	conn, err := ctrl.ConnectAPI(apiConnectTimeout)
	if err != nil {
		log.Printf("Error connecting to controller api: %v, %v\n", ctrl.ApiAddr, err)
		return false
	}
	defer conn.Close()

	debugApi := edgeproto.NewDebugApiClient(conn)
	log.Printf("Connected to controller %v success", ctrl.Name)
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	rc := true
	output := util.DebugOutput{}

	switch api {
	case "debugenable":
		runDebugApi(ctx, debugApi, "EnableDebugLevels", data.Requests, &output.Replies, &rc)
	case "debugdisable":
		runDebugApi(ctx, debugApi, "DisableDebugLevels", data.Requests, &output.Replies, &rc)
	case "debugshow":
		runDebugApi(ctx, debugApi, "ShowDebugLevels", data.Requests, &output.Replies, &rc)
	case "debugrun":
		runDebugApi(ctx, debugApi, "RunDebug", data.Requests, &output.Replies, &rc)
	}

	ymlOut, err := yaml.Marshal(output)
	if err != nil {
		log.Printf("Failed to marshal debug output, %v\n", err)
		rc = false
	} else {
		util.PrintToFile("api-output.yml", outputDir, string(ymlOut), true)
	}
	return rc
}

type debugReplyStream interface {
	Recv() (*edgeproto.DebugReply, error)
	grpc.ClientStream
}

func runDebugApi(ctx context.Context, client edgeproto.DebugApiClient, api string, reqs []edgeproto.DebugRequest, out *[][]edgeproto.DebugReply, rc *bool) {
	for _, req := range reqs {
		var stream debugReplyStream
		var err error
		switch api {
		case "EnableDebugLevels":
			stream, err = client.EnableDebugLevels(ctx, &req)
		case "DisableDebugLevels":
			stream, err = client.DisableDebugLevels(ctx, &req)
		case "ShowDebugLevels":
			stream, err = client.ShowDebugLevels(ctx, &req)
		case "RunDebug":
			stream, err = client.RunDebug(ctx, &req)
		default:
			log.Printf("Unknown debug API %s\n", api)
			*rc = false
		}
		if err != nil {
			log.Printf("debug api %s for request %v failed %v", api, req, err)
			*rc = false
			continue
		}
		replies := []edgeproto.DebugReply{}
		for {
			reply, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("debug api %s stream out for request %v failed %v", api, req, err)
				*rc = false
				break
			}
			replies = append(replies, *reply)
		}
		if len(replies) > 0 {
			if *out == nil {
				*out = make([][]edgeproto.DebugReply, 0)
			}
			*out = append(*out, replies)
		}
	}
}
