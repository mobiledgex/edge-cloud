// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: refs.proto

package gencmd

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
import "strconv"
import "github.com/spf13/cobra"
import "context"
import "os"
import "io"
import "text/tabwriter"
import "github.com/spf13/pflag"
import "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/cmdsup"
import "google.golang.org/grpc/status"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/protocmd"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
var CloudletRefsApiCmd edgeproto.CloudletRefsApiClient
var ClusterRefsApiCmd edgeproto.ClusterRefsApiClient
var CloudletRefsIn edgeproto.CloudletRefs
var CloudletRefsFlagSet = pflag.NewFlagSet("CloudletRefs", pflag.ExitOnError)
var CloudletRefsNoConfigFlagSet = pflag.NewFlagSet("CloudletRefsNoConfig", pflag.ExitOnError)
var CloudletRefsInRootLbPortsValue string
var ClusterRefsIn edgeproto.ClusterRefs
var ClusterRefsFlagSet = pflag.NewFlagSet("ClusterRefs", pflag.ExitOnError)
var ClusterRefsNoConfigFlagSet = pflag.NewFlagSet("ClusterRefsNoConfig", pflag.ExitOnError)
var port_protoStrings = []string{
	"Proto_Unknown",
	"Proto_TCP",
	"Proto_UDP",
	"Proto_TCP_UDP",
}

func CloudletRefsSlicer(in *edgeproto.CloudletRefs) []string {
	s := make([]string, 0, 8)
	s = append(s, in.Key.OperatorKey.Name)
	s = append(s, in.Key.Name)
	if in.Clusters == nil {
		in.Clusters = make([]edgeproto.ClusterKey, 1)
	}
	s = append(s, in.Clusters[0].Name)
	s = append(s, strconv.FormatUint(uint64(in.UsedRam), 10))
	s = append(s, strconv.FormatUint(uint64(in.UsedVcores), 10))
	s = append(s, strconv.FormatUint(uint64(in.UsedDisk), 10))
	s = append(s, strconv.FormatUint(uint64(in.UsedDynamicIps), 10))
	s = append(s, in.UsedStaticIps)
	return s
}

func CloudletRefsHeaderSlicer() []string {
	s := make([]string, 0, 8)
	s = append(s, "Key-OperatorKey-Name")
	s = append(s, "Key-Name")
	s = append(s, "Clusters-Name")
	s = append(s, "UsedRam")
	s = append(s, "UsedVcores")
	s = append(s, "UsedDisk")
	s = append(s, "UsedDynamicIps")
	s = append(s, "UsedStaticIps")
	return s
}

func CloudletRefsWriteOutputArray(objs []*edgeproto.CloudletRefs) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(CloudletRefsHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(CloudletRefsSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func CloudletRefsWriteOutputOne(obj *edgeproto.CloudletRefs) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(CloudletRefsHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(CloudletRefsSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}
func ClusterRefsSlicer(in *edgeproto.ClusterRefs) []string {
	s := make([]string, 0, 5)
	s = append(s, in.Key.ClusterKey.Name)
	s = append(s, in.Key.CloudletKey.OperatorKey.Name)
	s = append(s, in.Key.CloudletKey.Name)
	s = append(s, in.Key.Developer)
	if in.Apps == nil {
		in.Apps = make([]edgeproto.AppKey, 1)
	}
	s = append(s, in.Apps[0].DeveloperKey.Name)
	s = append(s, in.Apps[0].Name)
	s = append(s, in.Apps[0].Version)
	s = append(s, strconv.FormatUint(uint64(in.UsedRam), 10))
	s = append(s, strconv.FormatUint(uint64(in.UsedVcores), 10))
	s = append(s, strconv.FormatUint(uint64(in.UsedDisk), 10))
	return s
}

func ClusterRefsHeaderSlicer() []string {
	s := make([]string, 0, 5)
	s = append(s, "Key-ClusterKey-Name")
	s = append(s, "Key-CloudletKey-OperatorKey-Name")
	s = append(s, "Key-CloudletKey-Name")
	s = append(s, "Key-Developer")
	s = append(s, "Apps-DeveloperKey-Name")
	s = append(s, "Apps-Name")
	s = append(s, "Apps-Version")
	s = append(s, "UsedRam")
	s = append(s, "UsedVcores")
	s = append(s, "UsedDisk")
	return s
}

func ClusterRefsWriteOutputArray(objs []*edgeproto.ClusterRefs) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(ClusterRefsHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(ClusterRefsSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func ClusterRefsWriteOutputOne(obj *edgeproto.ClusterRefs) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(ClusterRefsHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(ClusterRefsSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}

var ShowCloudletRefsCmd = &cobra.Command{
	Use: "ShowCloudletRefs",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if we got this far, usage has been met.
		cmd.SilenceUsage = true
		err := parseCloudletRefsEnums()
		if err != nil {
			return fmt.Errorf("ShowCloudletRefs failed: %s", err.Error())
		}
		return ShowCloudletRefs(&CloudletRefsIn)
	},
}

func ShowCloudletRefs(in *edgeproto.CloudletRefs) error {
	if CloudletRefsApiCmd == nil {
		return fmt.Errorf("CloudletRefsApi client not initialized")
	}
	ctx := context.Background()
	stream, err := CloudletRefsApiCmd.ShowCloudletRefs(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowCloudletRefs failed: %s", errstr)
	}
	objs := make([]*edgeproto.CloudletRefs, 0)
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ShowCloudletRefs recv failed: %s", err.Error())
		}
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	CloudletRefsWriteOutputArray(objs)
	return nil
}

func ShowCloudletRefss(data []edgeproto.CloudletRefs, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowCloudletRefs %v\n", data[ii])
		myerr := ShowCloudletRefs(&data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var CloudletRefsApiCmds = []*cobra.Command{
	ShowCloudletRefsCmd,
}

var ShowClusterRefsCmd = &cobra.Command{
	Use: "ShowClusterRefs",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if we got this far, usage has been met.
		cmd.SilenceUsage = true
		return ShowClusterRefs(&ClusterRefsIn)
	},
}

func ShowClusterRefs(in *edgeproto.ClusterRefs) error {
	if ClusterRefsApiCmd == nil {
		return fmt.Errorf("ClusterRefsApi client not initialized")
	}
	ctx := context.Background()
	stream, err := ClusterRefsApiCmd.ShowClusterRefs(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowClusterRefs failed: %s", errstr)
	}
	objs := make([]*edgeproto.ClusterRefs, 0)
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ShowClusterRefs recv failed: %s", err.Error())
		}
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	ClusterRefsWriteOutputArray(objs)
	return nil
}

func ShowClusterRefss(data []edgeproto.ClusterRefs, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowClusterRefs %v\n", data[ii])
		myerr := ShowClusterRefs(&data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ClusterRefsApiCmds = []*cobra.Command{
	ShowClusterRefsCmd,
}

func init() {
	CloudletRefsFlagSet.StringVar(&CloudletRefsIn.Key.OperatorKey.Name, "key-operatorkey-name", "", "Key.OperatorKey.Name")
	CloudletRefsFlagSet.StringVar(&CloudletRefsIn.Key.Name, "key-name", "", "Key.Name")
	CloudletRefsFlagSet.Uint64Var(&CloudletRefsIn.UsedRam, "usedram", 0, "UsedRam")
	CloudletRefsFlagSet.Uint64Var(&CloudletRefsIn.UsedVcores, "usedvcores", 0, "UsedVcores")
	CloudletRefsFlagSet.Uint64Var(&CloudletRefsIn.UsedDisk, "useddisk", 0, "UsedDisk")
	CloudletRefsFlagSet.Int32Var(&CloudletRefsIn.UsedDynamicIps, "useddynamicips", 0, "UsedDynamicIps")
	CloudletRefsFlagSet.StringVar(&CloudletRefsIn.UsedStaticIps, "usedstaticips", "", "UsedStaticIps")
	ClusterRefsFlagSet.StringVar(&ClusterRefsIn.Key.ClusterKey.Name, "key-clusterkey-name", "", "Key.ClusterKey.Name")
	ClusterRefsFlagSet.StringVar(&ClusterRefsIn.Key.CloudletKey.OperatorKey.Name, "key-cloudletkey-operatorkey-name", "", "Key.CloudletKey.OperatorKey.Name")
	ClusterRefsFlagSet.StringVar(&ClusterRefsIn.Key.CloudletKey.Name, "key-cloudletkey-name", "", "Key.CloudletKey.Name")
	ClusterRefsFlagSet.StringVar(&ClusterRefsIn.Key.Developer, "key-developer", "", "Key.Developer")
	ClusterRefsFlagSet.Uint64Var(&ClusterRefsIn.UsedRam, "usedram", 0, "UsedRam")
	ClusterRefsFlagSet.Uint64Var(&ClusterRefsIn.UsedVcores, "usedvcores", 0, "UsedVcores")
	ClusterRefsFlagSet.Uint64Var(&ClusterRefsIn.UsedDisk, "useddisk", 0, "UsedDisk")
	ShowCloudletRefsCmd.Flags().AddFlagSet(CloudletRefsFlagSet)
	ShowClusterRefsCmd.Flags().AddFlagSet(ClusterRefsFlagSet)
}

func CloudletRefsApiAllowNoConfig() {
	ShowCloudletRefsCmd.Flags().AddFlagSet(CloudletRefsNoConfigFlagSet)
}

func ClusterRefsApiAllowNoConfig() {
	ShowClusterRefsCmd.Flags().AddFlagSet(ClusterRefsNoConfigFlagSet)
}

func parseCloudletRefsEnums() error {
	return nil
}
