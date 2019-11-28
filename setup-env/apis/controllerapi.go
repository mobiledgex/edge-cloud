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
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	"github.com/mobiledgex/edge-cloud/testutil"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

var appData edgeproto.ApplicationData

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

const (
	HideCmp bool = false
	ShowCmp bool = true
)

func runShow(ctrl *process.Controller, showCmds []string, outputDir string, cmp bool) bool {
	errFound := false
	for i, c := range showCmds {
		label := strings.Split(c, " ")[0]
		cmdstr := strings.Split(c, " ")[1]
		var cmdargs = []string{cmdstr}
		if cmp == HideCmp {
			cmdargs = append(cmdargs, "--hidetags")
			cmdargs = append(cmdargs, "nocmp,timestamp")
		} else {
			cmdargs = append(cmdargs, "--hidetags")
			cmdargs = append(cmdargs, "timestamp")
		}
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
		util.PrintToFile("show-commands.yml", outputDir, label+"\n"+string(out)+"\n", truncate)
	}
	return !errFound
}

func runShowCommands(ctrl *process.Controller, outputDir string, cmp bool) bool {
	var showCmds = []string{
		"flavors: ShowFlavor",
		"clusterinsts: ShowClusterInst",
		"operators: ShowOperator",
		"developers: ShowDeveloper",
		"cloudlets: ShowCloudlet",
		"apps: ShowApp",
		"appinstances: ShowAppInst",
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
		ret := runShow(ctrl, showCmds, outputDir, cmp)
		if ret {
			return true
		}
	}
	return false
}

func runNodeShow(ctrl *process.Controller, outputDir string, cmp bool) bool {
	var showCmds = []string{
		"nodes: ShowNode",
	}
	return runShow(ctrl, showCmds, outputDir, cmp)
}

//based on the api some errors will be converted to no error
func ignoreExpectedErrors(api string, key objstore.ObjKey, err error) error {
	if err == nil {
		return err
	}
	if api == "delete" {
		if strings.Contains(err.Error(), key.NotFoundError().Error()) {
			log.Printf("ignoring error on delete : %v\n", err)
			return nil
		}
	} else if api == "create" {
		if strings.Contains(err.Error(), key.ExistsError().Error()) {
			log.Printf("ignoring error on create : %v\n", err)
			return nil
		}
	}
	return err
}

func runFlavorApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	opAPI := edgeproto.NewFlavorApiClient(conn)
	var err error = nil
	for _, f := range appdata.Flavors {
		log.Printf("API %v for flavor: %v", mode, f.Key)
		switch mode {
		case "create":
			_, err = opAPI.CreateFlavor(ctx, &f)
		case "update":
			f.SetUpdateFields()
			_, err = opAPI.UpdateFlavor(ctx, &f)
		case "delete":
			_, err = opAPI.DeleteFlavor(ctx, &f)
		}
		err = ignoreExpectedErrors(mode, &f.Key, err)
		if err != nil {
			return fmt.Errorf("API %s failed for %v -- err %v", mode, f.Key, err)
		}
	}
	return nil
}

func runOperatorApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	opAPI := edgeproto.NewOperatorApiClient(conn)
	var err error = nil
	for _, o := range appdata.Operators {
		log.Printf("API %v for operator: %v", mode, o.Key)
		switch mode {
		case "create":
			_, err = opAPI.CreateOperator(ctx, &o)
		case "update":
			_, err = opAPI.UpdateOperator(ctx, &o)
		case "delete":
			_, err = opAPI.DeleteOperator(ctx, &o)
		}
		err = ignoreExpectedErrors(mode, &o.Key, err)
		if err != nil {
			return fmt.Errorf("API %s failed for %v -- err %v", mode, o.Key, err)
		}
	}
	return nil
}

func runDeveloperApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	var err error = nil
	devApi := edgeproto.NewDeveloperApiClient(conn)
	for _, d := range appdata.Developers {
		log.Printf("API %v for developer: %v", mode, d.Key.Name)
		switch mode {
		case "create":
			_, err = devApi.CreateDeveloper(ctx, &d)
		case "update":
			_, err = devApi.UpdateDeveloper(ctx, &d)
		case "delete":
			_, err = devApi.DeleteDeveloper(ctx, &d)
		}
		err = ignoreExpectedErrors(mode, &d.Key, err)
		if err != nil {
			return fmt.Errorf("API %s failed for %v -- err %v", mode, d.Key, err)
		}
	}
	return nil
}

func runCloudletApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	var err error = nil
	clAPI := edgeproto.NewCloudletApiClient(conn)
	for _, c := range appdata.Cloudlets {
		log.Printf("API %v for cloudlet: %v data %+v", mode, c.Key.Name, c)
		var stream testutil.CloudletStream
		switch mode {
		case "create":
			stream, err = clAPI.CreateCloudlet(ctx, &c)
		case "update":
			c.SetUpdateFields()
			stream, err = clAPI.UpdateCloudlet(ctx, &c)
		case "delete":
			stream, err = clAPI.DeleteCloudlet(ctx, &c)
		}
		err = testutil.CloudletReadResultStream(stream, err)
		err = ignoreExpectedErrors(mode, &c.Key, err)
		if err != nil {
			return fmt.Errorf("API %s failed for %v -- err %v", mode, c.Key, err)
		}
	}
	return nil
}

func runCloudletInfoApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	var err error = nil
	clAPI := edgeproto.NewCloudletInfoApiClient(conn)
	for _, c := range appdata.CloudletInfos {
		log.Printf("API %v for cloudletinfos: %v", mode, c.Key)
		switch mode {
		case "create":
			_, err = clAPI.InjectCloudletInfo(ctx, &c)
		case "update":
			// no-op
		case "delete":
			_, err = clAPI.EvictCloudletInfo(ctx, &c)
		}
		err = ignoreExpectedErrors(mode, &c.Key, err)
		if err != nil {
			return fmt.Errorf("API %s failed for %v -- err %v", mode, c.Key, err)
		}
	}
	return nil
}

func runAppApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	var err error = nil
	appAPI := edgeproto.NewAppApiClient(conn)
	for _, a := range appdata.Applications {
		log.Printf("API %v for app: %v", mode, a.Key.Name)
		switch mode {
		case "create":
			_, err = appAPI.CreateApp(ctx, &a)
		case "update":
			a.SetUpdateFields()
			_, err = appAPI.UpdateApp(ctx, &a)
		case "delete":
			_, err = appAPI.DeleteApp(ctx, &a)
		}
		err = ignoreExpectedErrors(mode, &a.Key, err)
		if err != nil {
			return fmt.Errorf("API %s failed for %v -- err %v", mode, a.Key, err)
		}
	}
	return nil
}

func runClusterInstApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	var err error = nil
	clusterinAPI := edgeproto.NewClusterInstApiClient(conn)
	for _, c := range appdata.ClusterInsts {
		log.Printf("API %v for clusterinst: %v", mode, c.Key)
		var stream testutil.ClusterInstStream
		switch mode {
		case "create":
			stream, err = clusterinAPI.CreateClusterInst(ctx, &c)
		case "update":
			c.SetUpdateFields()
			stream, err = clusterinAPI.UpdateClusterInst(ctx, &c)
		case "delete":
			stream, err = clusterinAPI.DeleteClusterInst(ctx, &c)
		}
		err = testutil.ClusterInstReadResultStream(stream, err)
		err = ignoreExpectedErrors(mode, &c.Key, err)
		if err != nil {
			return fmt.Errorf("API %s failed for %v -- err %v", mode, c.Key, err)
		}
	}

	return nil
}

func runAppinstApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	var err error = nil
	appinAPI := edgeproto.NewAppInstApiClient(conn)
	for _, a := range appdata.AppInstances {
		log.Printf("API %v for appinstance: %v", mode, a.Key)
		var stream testutil.AppInstStream
		switch mode {
		case "create":
			stream, err = appinAPI.CreateAppInst(ctx, &a)
		case "update":
			a.SetUpdateFields()
			stream, err = appinAPI.UpdateAppInst(ctx, &a)
		case "refresh":
			stream, err = appinAPI.RefreshAppInst(ctx, &a)
		case "delete":
			stream, err = appinAPI.DeleteAppInst(ctx, &a)
		}
		err = testutil.AppInstReadResultStream(stream, err)
		err = ignoreExpectedErrors(mode, &a.Key, err)
		if err != nil {
			return fmt.Errorf("API %s failed for %v -- err %v", mode, a.Key, err)
		}
	}

	return nil
}

func RunControllerAPI(api string, ctrlname string, apiFile string, outputDir string, mods []string) bool {
	runCLI := false
	for _, mod := range mods {
		if mod == "cli" {
			runCLI = true
		}
	}
	if runCLI {
		return RunControllerCLI(api, ctrlname, apiFile, outputDir, mods)
	}

	log.Printf("Applying data via APIs for %s\n", apiFile)
	apiConnectTimeout := 5 * time.Second
	apiTimeout := 30 * time.Minute

	ctrl := util.GetController(ctrlname)

	if api == "show" {
		//handled separately since it uses edgectl not direct api connection
		return runShowCommands(ctrl, outputDir, HideCmp)
	}
	if api == "showcmp" {
		return runShowCommands(ctrl, outputDir, ShowCmp)
	}
	if api == "nodeshow" {
		return runNodeShow(ctrl, outputDir, HideCmp)
	}

	if apiFile == "" {
		log.Println("Error: Cannot run controller APIs without API file")
		return false
	}

	readAppDataFile(apiFile)

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
			//run in reverse order to delete child keys
			err = runAppinstApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in appinst API %v \n", err)
				rc = false
			}
			err = runClusterInstApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in clusterinst API %v\n", err)
				rc = false
			}
			err = runAppApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in app API %v\n", err)
				rc = false
			}
			err = runCloudletInfoApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in cloudletInfo API %v\n", err)
				rc = false
			}
			err = runCloudletApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in cloudlet API %v\n", err)
				rc = false
			}
			err = runDeveloperApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in developer API %v\n", err)
				rc = false
			}
			err = runOperatorApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in operator API %v\n", err)
				rc = false
			}
			err = runFlavorApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in flavor API %v\n", err)
				rc = false
			}
		case "create":
			fallthrough
		case "refresh":
			fallthrough
		case "update":
			err = runFlavorApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in flavor API %v\n", err)
				rc = false
			}
			err = runOperatorApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in operator API %v\n", err)
				rc = false
			}
			err = runDeveloperApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in developer API %v\n", err)
				rc = false
			}
			err = runCloudletApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in cloudlet API %v\n", err)
				rc = false
			}
			err = runCloudletInfoApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in cloudletInfo API %v\n", err)
				rc = false
			}
			err = runAppApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in app API %v\n", err)
				rc = false
			}
			err = runClusterInstApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in clusterinst API %v\n", err)
				rc = false
			}
			err = runAppinstApi(ctrlapi, ctx, &appData, api)
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
		return runShowCommands(ctrl, outputDir, HideCmp)
	}
	if api == "showcmp" {
		return runShowCommands(ctrl, outputDir, ShowCmp)
	}
	if api == "nodeshow" {
		return runNodeShow(ctrl, outputDir, HideCmp)
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
	log.Printf("RunCommand using %s\n", apiFile)

	ctrl := util.GetController(ctrlname)

	data := runCommandData{}
	if apiFile == "" {
		log.Println("Error: Cannot run RunCommand API without API file")
		return false
	}
	err := util.ReadYamlFile(apiFile, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in unmarshal for file %s, %v\n", apiFile, err)
		return false
	}
	req := &data.Request

	args := []string{"RunCommand"}
	args = append(args, "developer="+req.AppInstKey.AppKey.DeveloperKey.Name)
	args = append(args, "appname="+req.AppInstKey.AppKey.Name)
	args = append(args, "appvers="+req.AppInstKey.AppKey.Version)
	args = append(args, "cloudlet="+req.AppInstKey.ClusterInstKey.CloudletKey.Name)
	args = append(args, "operator="+req.AppInstKey.ClusterInstKey.CloudletKey.OperatorKey.Name)
	args = append(args, "cluster="+req.AppInstKey.ClusterInstKey.ClusterKey.Name)
	args = append(args, "clusterdeveloper="+req.AppInstKey.ClusterInstKey.Developer)
	args = append(args, "command=\""+req.Command+"\"")
	out, err := util.ControllerCLI(ctrl, args...)
	if err != nil {
		log.Printf("Error running RunCommand API %v\n", err)
		return false
	}
	log.Printf("RunCommand output: %s\n", string(out))
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
		pfConfig.RegistryPath = "registry.mobiledgex.net:5000/mobiledgex/edge-cloud"
		pfConfig.ImagePath = ""
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
