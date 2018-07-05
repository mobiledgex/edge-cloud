package apis

// interacts with the controller APIs for use by the e2e test tool

import (
	"context"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
	"google.golang.org/grpc"
)

var appData edgeproto.ApplicationData

func readAppDataFile(file string) {
	err := util.ReadYamlFile(file, &appData, "", true)
	if err != nil {
		if !util.IsYamlOk(err, "appdata") {
			log.Fatal("One or more fatal unmarshal errors, exiting")
		}
	}
}

func runShowCommands(ctrl *util.ControllerProcess, outputDir string) bool {
	errFound := false
	var showCmds = []string{
		"operators: ShowOperator",
		"developers: ShowDeveloper",
		"cloudlets: ShowCloudlet",
		"apps: ShowApp",
		"appinstances: ShowAppInst",
	}
	for i, c := range showCmds {
		label := strings.Split(c, " ")[0]
		cmdstr := strings.Split(c, " ")[1]
		cmd := exec.Command("edgectl", "--addr", ctrl.ApiAddr, "controller", cmdstr)
		log.Printf("generating output for %s\n", label)
		out, _ := cmd.CombinedOutput()
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

func runOperatorApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	opAPI := edgeproto.NewOperatorApiClient(conn)
	var err error = nil
	for _, o := range appdata.Operators {
		log.Printf("API %v for operator: %v", mode, o.Key.Name)
		switch mode {
		case "create":
			_, err = opAPI.CreateOperator(ctx, &o)
		case "update":
			_, err = opAPI.UpdateOperator(ctx, &o)
		case "delete":
			_, err = opAPI.DeleteOperator(ctx, &o)
		}
	}
	return err
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
	}
	return err
}

func runCloudletApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	var err error = nil
	clAPI := edgeproto.NewCloudletApiClient(conn)
	for _, c := range appdata.Cloudlets {
		log.Printf("API %v for cloudlet: %v", mode, c.Key.Name)
		switch mode {
		case "create":
			_, err = clAPI.CreateCloudlet(ctx, &c)
		case "update":
			_, err = clAPI.UpdateCloudlet(ctx, &c)
		case "delete":
			_, err = clAPI.DeleteCloudlet(ctx, &c)
		}
	}
	return err
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
			_, err = appAPI.UpdateApp(ctx, &a)
		case "delete":
			_, err = appAPI.DeleteApp(ctx, &a)
		}
	}
	return err
}

func runAppinstApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	var err error = nil
	appinAPI := edgeproto.NewAppInstApiClient(conn)
	for _, a := range appdata.AppInstances {
		log.Printf("API %v for appinstance: %v", mode, a.Key.AppKey.Name)
		switch mode {
		case "create":
			_, err = appinAPI.CreateAppInst(ctx, &a)
		case "update":
			_, err = appinAPI.UpdateAppInst(ctx, &a)
		case "delete":
			_, err = appinAPI.DeleteAppInst(ctx, &a)
		}
	}
	return err
}

func RunControllerAPI(api string, ctrlname string, apiFile string, outputDir string) bool {
	log.Printf("Applying data via APIs\n")
	apiConnectTimeout := 5 * time.Second

	ctrl := util.GetController(ctrlname)

	if api == "show" {
		//handled separately since it uses edgectl not direct api connection
		return runShowCommands(ctrl, outputDir)
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
		ctx, cancel := context.WithTimeout(context.Background(), apiConnectTimeout)

		switch api {
		case "delete":
			//run in reverse order to delete child keys
			runAppinstApi(ctrlapi, ctx, &appData, api)
			runAppApi(ctrlapi, ctx, &appData, api)
			runCloudletApi(ctrlapi, ctx, &appData, api)
			runDeveloperApi(ctrlapi, ctx, &appData, api)
			runOperatorApi(ctrlapi, ctx, &appData, api)
		case "create":
			fallthrough
		case "update":
			runOperatorApi(ctrlapi, ctx, &appData, api)
			runDeveloperApi(ctrlapi, ctx, &appData, api)
			runCloudletApi(ctrlapi, ctx, &appData, api)
			runAppApi(ctrlapi, ctx, &appData, api)
			runAppinstApi(ctrlapi, ctx, &appData, api)
		default:
			log.Printf("Error: unsupported controller API %s\n", api)
			rc = false
		}
		cancel()
	}
	ctrlapi.Close()
	return rc
}
