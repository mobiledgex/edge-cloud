package apis

// interacts with the controller APIs for use by the e2e test tool

import (
	"context"
	"fmt"
	"log"
	"os"
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
			fmt.Fprintf(os.Stderr, "Error in unmarshal for file %s", file)
			os.Exit(1)
		}
	}
}

func runShowCommands(ctrl *util.ControllerProcess, outputDir string) bool {
	errFound := false
	var showCmds = []string{
		"flavors: ShowFlavor",
		"clusters: ShowCluster",
		"clusterinsts: ShowClusterInst",
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

func runFlavorApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	opAPI := edgeproto.NewFlavorApiClient(conn)
	var err error = nil
	for _, o := range appdata.Flavors {
		log.Printf("API %v for flavor: %v", mode, o.Key.Name)
		switch mode {
		case "create":
			_, err = opAPI.CreateFlavor(ctx, &o)
		case "update":
			_, err = opAPI.UpdateFlavor(ctx, &o)
		case "delete":
			_, err = opAPI.DeleteFlavor(ctx, &o)
		}
		if err != nil {
			return err
		}
	}
	return nil
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
		if err != nil {
			return err
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
		if err != nil {
			return err
		}
	}
	return nil
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
		if err != nil {
			return err
		}
	}
	return nil
}

func runClusterApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	var err error = nil
	clusterAPI := edgeproto.NewClusterApiClient(conn)
	for _, a := range appdata.Clusters {
		log.Printf("API %v for cluster: %v", mode, a.Key.Name)
		switch mode {
		case "create":
			_, err = clusterAPI.CreateCluster(ctx, &a)
		case "update":
			_, err = clusterAPI.UpdateCluster(ctx, &a)
		case "delete":
			_, err = clusterAPI.DeleteCluster(ctx, &a)
		}
		if err != nil {
			return err
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
			_, err = appAPI.UpdateApp(ctx, &a)
		case "delete":
			_, err = appAPI.DeleteApp(ctx, &a)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func runClusterInstApi(conn *grpc.ClientConn, ctx context.Context, appdata *edgeproto.ApplicationData, mode string) error {
	var err error = nil
	clusterinAPI := edgeproto.NewClusterInstApiClient(conn)
	for _, a := range appdata.ClusterInsts {
		log.Printf("API %v for clusterinst: %v", mode, a.Key.ClusterKey.Name)
		switch mode {
		case "create":
			_, err = clusterinAPI.CreateClusterInst(ctx, &a)
		case "update":
			_, err = clusterinAPI.UpdateClusterInst(ctx, &a)
		case "delete":
			_, err = clusterinAPI.DeleteClusterInst(ctx, &a)
		}
		if err != nil {
			return err
		}
	}

	return nil
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
		if err != nil {
			return err
		}
	}

	return nil
}

func RunControllerAPI(api string, ctrlname string, apiFile string, outputDir string) bool {
	log.Printf("Applying data via APIs\n")
	apiConnectTimeout := 5 * time.Second
	apiTimeout := 120 * time.Second

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
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)

		var err error
		switch api {
		case "delete":
			//run in reverse order to delete child keys
			err = runAppinstApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in appinst API %v\n", err)
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
			err = runClusterApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in cluster API %v\n", err)
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
		case "update":
			err = runFlavorApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in operator API %v\n", err)
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
			err = runClusterApi(ctrlapi, ctx, &appData, api)
			if err != nil {
				log.Printf("Error in cluster API %v\n", err)
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

func RunControllerInfoAPI(api, ctrlname, apiFile, outputDir string) bool {
	log.Printf("Showing info structs via APIs\n")

	ctrl := util.GetController(ctrlname)
	errFound := false

	var showCmds = []string{
		"clusterInstInfos: ShowClusterInstInfo",
		"appInstInfos: ShowAppInstInfo",
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
