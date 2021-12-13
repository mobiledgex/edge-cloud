// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: network.proto

package gencmd

import (
	"context"
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud/cli"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/status"
	"io"
	math "math"
	"strings"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
var NetworkApiCmd edgeproto.NetworkApiClient

var CreateNetworkCmd = &cli.Command{
	Use:          "CreateNetwork",
	RequiredArgs: strings.Join(NetworkRequiredArgs, " "),
	OptionalArgs: strings.Join(NetworkOptionalArgs, " "),
	AliasArgs:    strings.Join(NetworkAliasArgs, " "),
	SpecialArgs:  &NetworkSpecialArgs,
	Comments:     NetworkComments,
	ReqData:      &edgeproto.Network{},
	ReplyData:    &edgeproto.Result{},
	Run:          runCreateNetwork,
}

func runCreateNetwork(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.Network)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return CreateNetwork(c, obj)
}

func CreateNetwork(c *cli.Command, in *edgeproto.Network) error {
	if NetworkApiCmd == nil {
		return fmt.Errorf("NetworkApi client not initialized")
	}
	ctx := context.Background()
	stream, err := NetworkApiCmd.CreateNetwork(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateNetwork failed: %s", errstr)
	}

	objs := make([]*edgeproto.Result, 0)
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			errstr := err.Error()
			st, ok := status.FromError(err)
			if ok {
				errstr = st.Message()
			}
			return fmt.Errorf("CreateNetwork recv failed: %s", errstr)
		}
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), objs, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func CreateNetworks(c *cli.Command, data []edgeproto.Network, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("CreateNetwork %v\n", data[ii])
		myerr := CreateNetwork(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var DeleteNetworkCmd = &cli.Command{
	Use:          "DeleteNetwork",
	RequiredArgs: strings.Join(NetworkRequiredArgs, " "),
	OptionalArgs: strings.Join(NetworkOptionalArgs, " "),
	AliasArgs:    strings.Join(NetworkAliasArgs, " "),
	SpecialArgs:  &NetworkSpecialArgs,
	Comments:     NetworkComments,
	ReqData:      &edgeproto.Network{},
	ReplyData:    &edgeproto.Result{},
	Run:          runDeleteNetwork,
}

func runDeleteNetwork(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.Network)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return DeleteNetwork(c, obj)
}

func DeleteNetwork(c *cli.Command, in *edgeproto.Network) error {
	if NetworkApiCmd == nil {
		return fmt.Errorf("NetworkApi client not initialized")
	}
	ctx := context.Background()
	stream, err := NetworkApiCmd.DeleteNetwork(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("DeleteNetwork failed: %s", errstr)
	}

	objs := make([]*edgeproto.Result, 0)
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			errstr := err.Error()
			st, ok := status.FromError(err)
			if ok {
				errstr = st.Message()
			}
			return fmt.Errorf("DeleteNetwork recv failed: %s", errstr)
		}
		if cli.OutputStream {
			c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
			continue
		}
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), objs, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func DeleteNetworks(c *cli.Command, data []edgeproto.Network, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("DeleteNetwork %v\n", data[ii])
		myerr := DeleteNetwork(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var UpdateNetworkCmd = &cli.Command{
	Use:          "UpdateNetwork",
	RequiredArgs: strings.Join(NetworkRequiredArgs, " "),
	OptionalArgs: strings.Join(NetworkOptionalArgs, " "),
	AliasArgs:    strings.Join(NetworkAliasArgs, " "),
	SpecialArgs:  &NetworkSpecialArgs,
	Comments:     NetworkComments,
	ReqData:      &edgeproto.Network{},
	ReplyData:    &edgeproto.Result{},
	Run:          runUpdateNetwork,
}

func runUpdateNetwork(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.Network)
	jsonMap, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj.Fields = cli.GetSpecifiedFields(jsonMap, c.ReqData)
	return UpdateNetwork(c, obj)
}

func UpdateNetwork(c *cli.Command, in *edgeproto.Network) error {
	if NetworkApiCmd == nil {
		return fmt.Errorf("NetworkApi client not initialized")
	}
	ctx := context.Background()
	stream, err := NetworkApiCmd.UpdateNetwork(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("UpdateNetwork failed: %s", errstr)
	}

	objs := make([]*edgeproto.Result, 0)
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			errstr := err.Error()
			st, ok := status.FromError(err)
			if ok {
				errstr = st.Message()
			}
			return fmt.Errorf("UpdateNetwork recv failed: %s", errstr)
		}
		if cli.OutputStream {
			c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
			continue
		}
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), objs, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func UpdateNetworks(c *cli.Command, data []edgeproto.Network, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("UpdateNetwork %v\n", data[ii])
		myerr := UpdateNetwork(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ShowNetworkCmd = &cli.Command{
	Use:          "ShowNetwork",
	OptionalArgs: strings.Join(append(NetworkRequiredArgs, NetworkOptionalArgs...), " "),
	AliasArgs:    strings.Join(NetworkAliasArgs, " "),
	SpecialArgs:  &NetworkSpecialArgs,
	Comments:     NetworkComments,
	ReqData:      &edgeproto.Network{},
	ReplyData:    &edgeproto.Network{},
	Run:          runShowNetwork,
}

func runShowNetwork(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.Network)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowNetwork(c, obj)
}

func ShowNetwork(c *cli.Command, in *edgeproto.Network) error {
	if NetworkApiCmd == nil {
		return fmt.Errorf("NetworkApi client not initialized")
	}
	ctx := context.Background()
	stream, err := NetworkApiCmd.ShowNetwork(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowNetwork failed: %s", errstr)
	}

	objs := make([]*edgeproto.Network, 0)
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			errstr := err.Error()
			st, ok := status.FromError(err)
			if ok {
				errstr = st.Message()
			}
			return fmt.Errorf("ShowNetwork recv failed: %s", errstr)
		}
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), objs, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func ShowNetworks(c *cli.Command, data []edgeproto.Network, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowNetwork %v\n", data[ii])
		myerr := ShowNetwork(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var NetworkApiCmds = []*cobra.Command{
	CreateNetworkCmd.GenCmd(),
	DeleteNetworkCmd.GenCmd(),
	UpdateNetworkCmd.GenCmd(),
	ShowNetworkCmd.GenCmd(),
}

var RouteRequiredArgs = []string{}
var RouteOptionalArgs = []string{
	"destinationcidr",
	"nexthopip",
}
var RouteAliasArgs = []string{}
var RouteComments = map[string]string{
	"destinationcidr": "Destination CIDR",
	"nexthopip":       "Next hop IP",
}
var RouteSpecialArgs = map[string]string{}
var NetworkKeyRequiredArgs = []string{}
var NetworkKeyOptionalArgs = []string{
	"cloudletkey.organization",
	"cloudletkey.name",
	"cloudletkey.federatedorganization",
	"name",
}
var NetworkKeyAliasArgs = []string{}
var NetworkKeyComments = map[string]string{
	"cloudletkey.organization":          "Organization of the cloudlet site",
	"cloudletkey.name":                  "Name of the cloudlet",
	"cloudletkey.federatedorganization": "Federated operator organization who shared this cloudlet",
	"name":                              "Network Name",
}
var NetworkKeySpecialArgs = map[string]string{}
var NetworkRequiredArgs = []string{
	"cloudlet-org",
	"key.cloudletkey.name",
	"name",
}
var NetworkOptionalArgs = []string{
	"federated-org",
	"routes:empty",
	"routes:#.destinationcidr",
	"routes:#.nexthopip",
	"connectiontype",
}
var NetworkAliasArgs = []string{
	"cloudlet-org=key.cloudletkey.organization",
	"federated-org=key.cloudletkey.federatedorganization",
	"name=key.name",
}
var NetworkComments = map[string]string{
	"fields":                   "Fields are used for the Update API to specify which fields to apply",
	"cloudlet-org":             "Organization of the cloudlet site",
	"key.cloudletkey.name":     "Name of the cloudlet",
	"federated-org":            "Federated operator organization who shared this cloudlet",
	"name":                     "Network Name",
	"routes:empty":             "List of routes, specify routes:empty=true to clear",
	"routes:#.destinationcidr": "Destination CIDR",
	"routes:#.nexthopip":       "Next hop IP",
	"connectiontype":           "Network connection type, one of Undefined, ConnectToLoadBalancer, ConnectToClusterNodes, ConnectToAll",
	"deleteprepare":            "Preparing to be deleted",
}
var NetworkSpecialArgs = map[string]string{
	"fields": "StringArray",
}
