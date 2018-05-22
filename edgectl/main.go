package main

import (
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/gencmd"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var addr string
var conn *grpc.ClientConn

var rootCmd = &cobra.Command{
	Use:               "edgectl",
	PersistentPreRun:  connect,
	PersistentPostRun: close,
}
var controllerCmd = &cobra.Command{
	Use: "controller",
}
var dmeCmd = &cobra.Command{
	Use: "dme",
}
var crmCmd = &cobra.Command{
	Use: "crm",
}

func connect(cmd *cobra.Command, args []string) {
	var err error
	conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		util.FatalLog("Connect to server failed", "addr", addr, "err", err)
	}
	gencmd.DeveloperApiCmd = edgeproto.NewDeveloperApiClient(conn)
	gencmd.AppApiCmd = edgeproto.NewAppApiClient(conn)
	gencmd.OperatorApiCmd = edgeproto.NewOperatorApiClient(conn)
	gencmd.CloudletApiCmd = edgeproto.NewCloudletApiClient(conn)
	gencmd.AppInstApiCmd = edgeproto.NewAppInstApiClient(conn)
	gencmd.Match_Engine_ApiCmd = dme.NewMatch_Engine_ApiClient(conn)
	gencmd.CloudResourceManagerCmd = edgeproto.NewCloudResourceManagerClient(conn)
}

func close(cmd *cobra.Command, args []string) {
	conn.Close()
}

func main() {
	rootCmd.AddCommand(controllerCmd)
	rootCmd.AddCommand(dmeCmd)
	rootCmd.AddCommand(crmCmd)
	rootCmd.PersistentFlags().StringVar(&addr, "addr", "127.0.0.1:55001", "address to connect to")

	controllerCmd.AddCommand(gencmd.CreateDeveloperCmd)
	controllerCmd.AddCommand(gencmd.UpdateDeveloperCmd)
	controllerCmd.AddCommand(gencmd.DeleteDeveloperCmd)
	controllerCmd.AddCommand(gencmd.ShowDeveloperCmd)

	controllerCmd.AddCommand(gencmd.CreateAppCmd)
	controllerCmd.AddCommand(gencmd.UpdateAppCmd)
	controllerCmd.AddCommand(gencmd.DeleteAppCmd)
	controllerCmd.AddCommand(gencmd.ShowAppCmd)

	controllerCmd.AddCommand(gencmd.CreateOperatorCmd)
	controllerCmd.AddCommand(gencmd.UpdateOperatorCmd)
	controllerCmd.AddCommand(gencmd.DeleteOperatorCmd)
	controllerCmd.AddCommand(gencmd.ShowOperatorCmd)

	controllerCmd.AddCommand(gencmd.CreateCloudletCmd)
	controllerCmd.AddCommand(gencmd.UpdateCloudletCmd)
	controllerCmd.AddCommand(gencmd.DeleteCloudletCmd)
	controllerCmd.AddCommand(gencmd.ShowCloudletCmd)

	controllerCmd.AddCommand(gencmd.CreateAppInstCmd)
	controllerCmd.AddCommand(gencmd.UpdateAppInstCmd)
	controllerCmd.AddCommand(gencmd.DeleteAppInstCmd)
	controllerCmd.AddCommand(gencmd.ShowAppInstCmd)

	dmeCmd.AddCommand(gencmd.FindCloudletCmd)
	dmeCmd.AddCommand(gencmd.VerifyLocationCmd)

	crmCmd.AddCommand(gencmd.ListCloudResourceCmd)
	crmCmd.AddCommand(gencmd.AddCloudResourceCmd)
	crmCmd.AddCommand(gencmd.DeleteCloudResourceCmd)
	crmCmd.AddCommand(gencmd.DeployApplicationCmd)
	crmCmd.AddCommand(gencmd.DeleteApplicationCmd)

	rootCmd.Execute()
}
