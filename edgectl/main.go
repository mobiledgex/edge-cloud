package main

import (
	"fmt"
	"os"

	"github.com/mobiledgex/edge-cloud/cli"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/gencmd"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/tls"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var addr string
var tlsCertFile string
var conn *grpc.ClientConn

var rootCmd = &cobra.Command{
	Use:                "edgectl",
	PersistentPreRunE:  connect,
	PersistentPostRunE: close,
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

var completionCmd = &cobra.Command{
	Use:   "completion-script",
	Short: "Generates bash completion script",
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := "edgectl-completion.bash"
		outfile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("Unable to open file %s in current dir for write: %s", filename, err.Error())
		}
		rootCmd.GenBashCompletion(outfile)
		fmt.Printf("Wrote file %s in current dir. Move it to /usr/local/etc/bash_completion.d/ (OSX) or /etc/bash_completion.d/ (linux)\n", filename)
		fmt.Println("On OSX, make sure bash completion is installed:")
		fmt.Println("  brew install bash-completion")
		outfile.Close()
		return nil
	},
}

func connect(cmd *cobra.Command, args []string) error {
	var err error

	dialOption, err := tls.GetTLSClientDialOption(addr, tlsCertFile, false)
	if err != nil {
		return err
	}
	conn, err = grpc.Dial(addr, dialOption)
	if err != nil {
		return fmt.Errorf("Connect to server %s failed: %s", addr, err.Error())
	}
	gencmd.DeveloperApiCmd = edgeproto.NewDeveloperApiClient(conn)
	gencmd.AppApiCmd = edgeproto.NewAppApiClient(conn)
	gencmd.OperatorApiCmd = edgeproto.NewOperatorApiClient(conn)
	gencmd.FlavorApiCmd = edgeproto.NewFlavorApiClient(conn)
	gencmd.AutoScalePolicyApiCmd = edgeproto.NewAutoScalePolicyApiClient(conn)
	gencmd.AutoProvPolicyApiCmd = edgeproto.NewAutoProvPolicyApiClient(conn)
	gencmd.PrivacyPolicyApiCmd = edgeproto.NewPrivacyPolicyApiClient(conn)
	gencmd.ClusterInstApiCmd = edgeproto.NewClusterInstApiClient(conn)
	gencmd.CloudletApiCmd = edgeproto.NewCloudletApiClient(conn)
	gencmd.AppInstApiCmd = edgeproto.NewAppInstApiClient(conn)
	gencmd.CloudletInfoApiCmd = edgeproto.NewCloudletInfoApiClient(conn)
	gencmd.AppInstInfoApiCmd = edgeproto.NewAppInstInfoApiClient(conn)
	gencmd.ClusterInstInfoApiCmd = edgeproto.NewClusterInstInfoApiClient(conn)
	gencmd.MatchEngineApiCmd = dme.NewMatchEngineApiClient(conn)
	gencmd.DebugApiCmd = log.NewDebugApiClient(conn)
	gencmd.ControllerApiCmd = edgeproto.NewControllerApiClient(conn)
	gencmd.NodeApiCmd = edgeproto.NewNodeApiClient(conn)
	gencmd.CloudletPoolApiCmd = edgeproto.NewCloudletPoolApiClient(conn)
	gencmd.CloudletPoolMemberApiCmd = edgeproto.NewCloudletPoolMemberApiClient(conn)
	gencmd.AlertApiCmd = edgeproto.NewAlertApiClient(conn)
	gencmd.ResTagTableApiCmd = edgeproto.NewResTagTableApiClient(conn)
	execApiCmd = edgeproto.NewExecApiClient(conn)
	return nil
}

func close(cmd *cobra.Command, args []string) error {
	conn.Close()
	return nil
}

func main() {
	cobra.EnableCommandSorting = false

	rootCmd.AddCommand(controllerCmd)
	rootCmd.AddCommand(dmeCmd)
	rootCmd.AddCommand(crmCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.PersistentFlags().StringVar(&addr, "addr", "127.0.0.1:55001", "address to connect to")
	rootCmd.PersistentFlags().StringVar(&tlsCertFile, "tls", "", "tls cert file")
	cli.AddInputFlags(rootCmd.PersistentFlags())
	cli.AddOutputFlags(rootCmd.PersistentFlags())
	cli.AddDebugFlag(rootCmd.PersistentFlags())
	cli.AddHideTagsFormatFlag(rootCmd.PersistentFlags())

	controllerCmd.AddCommand(gencmd.DeveloperApiCmds...)
	controllerCmd.AddCommand(gencmd.AppApiCmds...)
	controllerCmd.AddCommand(gencmd.OperatorApiCmds...)
	controllerCmd.AddCommand(gencmd.FlavorApiCmds...)
	controllerCmd.AddCommand(gencmd.AutoScalePolicyApiCmds...)
	controllerCmd.AddCommand(gencmd.AutoProvPolicyApiCmds...)
	controllerCmd.AddCommand(gencmd.PrivacyPolicyApiCmds...)
	controllerCmd.AddCommand(gencmd.ClusterInstApiCmds...)
	controllerCmd.AddCommand(gencmd.CloudletApiCmds...)
	controllerCmd.AddCommand(gencmd.AppInstApiCmds...)
	controllerCmd.AddCommand(gencmd.DebugApiCmds...)
	controllerCmd.AddCommand(gencmd.CloudletInfoApiCmds...)
	controllerCmd.AddCommand(gencmd.ControllerApiCmds...)
	controllerCmd.AddCommand(gencmd.NodeApiCmds...)
	controllerCmd.AddCommand(gencmd.CloudletPoolApiCmds...)
	controllerCmd.AddCommand(gencmd.CloudletPoolMemberApiCmds...)
	controllerCmd.AddCommand(gencmd.AlertApiCmds...)
	controllerCmd.AddCommand(gencmd.ResTagTableApiCmds...)

	controllerCmd.AddCommand(createCmd.GenCmd())
	controllerCmd.AddCommand(deleteCmd.GenCmd())
	gencmd.RunCommandCmd.Run = runExecRequest
	controllerCmd.AddCommand(gencmd.RunCommandCmd.GenCmd())

	dmeCmd.AddCommand(gencmd.MatchEngineApiCmds...)
	dmeCmd.AddCommand(gencmd.DebugApiCmds...)

	crmCmd.AddCommand(gencmd.DebugApiCmds...)
	crmCmd.AddCommand(gencmd.ClusterInstInfoApiCmds...)
	crmCmd.AddCommand(gencmd.AppInstInfoApiCmds...)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
