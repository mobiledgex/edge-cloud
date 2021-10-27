// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: trustpolicyexception.proto

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
var TrustPolicyExceptionApiCmd edgeproto.TrustPolicyExceptionApiClient

var CreateTrustPolicyExceptionCmd = &cli.Command{
	Use:          "CreateTrustPolicyException",
	RequiredArgs: strings.Join(TrustPolicyExceptionRequiredArgs, " "),
	OptionalArgs: strings.Join(TrustPolicyExceptionOptionalArgs, " "),
	AliasArgs:    strings.Join(TrustPolicyExceptionAliasArgs, " "),
	SpecialArgs:  &TrustPolicyExceptionSpecialArgs,
	Comments:     TrustPolicyExceptionComments,
	ReqData:      &edgeproto.TrustPolicyException{},
	ReplyData:    &edgeproto.Result{},
	Run:          runCreateTrustPolicyException,
}

func runCreateTrustPolicyException(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.TrustPolicyException)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return CreateTrustPolicyException(c, obj)
}

func CreateTrustPolicyException(c *cli.Command, in *edgeproto.TrustPolicyException) error {
	if TrustPolicyExceptionApiCmd == nil {
		return fmt.Errorf("TrustPolicyExceptionApi client not initialized")
	}
	ctx := context.Background()
	obj, err := TrustPolicyExceptionApiCmd.CreateTrustPolicyException(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateTrustPolicyException failed: %s", errstr)
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func CreateTrustPolicyExceptions(c *cli.Command, data []edgeproto.TrustPolicyException, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("CreateTrustPolicyException %v\n", data[ii])
		myerr := CreateTrustPolicyException(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var UpdateTrustPolicyExceptionCmd = &cli.Command{
	Use:          "UpdateTrustPolicyException",
	RequiredArgs: strings.Join(UpdateTrustPolicyExceptionRequiredArgs, " "),
	OptionalArgs: strings.Join(UpdateTrustPolicyExceptionOptionalArgs, " "),
	AliasArgs:    strings.Join(TrustPolicyExceptionAliasArgs, " "),
	SpecialArgs:  &TrustPolicyExceptionSpecialArgs,
	Comments:     TrustPolicyExceptionComments,
	ReqData:      &edgeproto.TrustPolicyException{},
	ReplyData:    &edgeproto.Result{},
	Run:          runUpdateTrustPolicyException,
}

func runUpdateTrustPolicyException(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.TrustPolicyException)
	jsonMap, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj.Fields = cli.GetSpecifiedFields(jsonMap, c.ReqData)
	return UpdateTrustPolicyException(c, obj)
}

func UpdateTrustPolicyException(c *cli.Command, in *edgeproto.TrustPolicyException) error {
	if TrustPolicyExceptionApiCmd == nil {
		return fmt.Errorf("TrustPolicyExceptionApi client not initialized")
	}
	ctx := context.Background()
	obj, err := TrustPolicyExceptionApiCmd.UpdateTrustPolicyException(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("UpdateTrustPolicyException failed: %s", errstr)
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func UpdateTrustPolicyExceptions(c *cli.Command, data []edgeproto.TrustPolicyException, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("UpdateTrustPolicyException %v\n", data[ii])
		myerr := UpdateTrustPolicyException(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var DeleteTrustPolicyExceptionCmd = &cli.Command{
	Use:          "DeleteTrustPolicyException",
	RequiredArgs: strings.Join(TrustPolicyExceptionRequiredArgs, " "),
	OptionalArgs: strings.Join(TrustPolicyExceptionOptionalArgs, " "),
	AliasArgs:    strings.Join(TrustPolicyExceptionAliasArgs, " "),
	SpecialArgs:  &TrustPolicyExceptionSpecialArgs,
	Comments:     TrustPolicyExceptionComments,
	ReqData:      &edgeproto.TrustPolicyException{},
	ReplyData:    &edgeproto.Result{},
	Run:          runDeleteTrustPolicyException,
}

func runDeleteTrustPolicyException(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.TrustPolicyException)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return DeleteTrustPolicyException(c, obj)
}

func DeleteTrustPolicyException(c *cli.Command, in *edgeproto.TrustPolicyException) error {
	if TrustPolicyExceptionApiCmd == nil {
		return fmt.Errorf("TrustPolicyExceptionApi client not initialized")
	}
	ctx := context.Background()
	obj, err := TrustPolicyExceptionApiCmd.DeleteTrustPolicyException(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("DeleteTrustPolicyException failed: %s", errstr)
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func DeleteTrustPolicyExceptions(c *cli.Command, data []edgeproto.TrustPolicyException, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("DeleteTrustPolicyException %v\n", data[ii])
		myerr := DeleteTrustPolicyException(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ShowTrustPolicyExceptionCmd = &cli.Command{
	Use:          "ShowTrustPolicyException",
	OptionalArgs: strings.Join(append(TrustPolicyExceptionRequiredArgs, TrustPolicyExceptionOptionalArgs...), " "),
	AliasArgs:    strings.Join(TrustPolicyExceptionAliasArgs, " "),
	SpecialArgs:  &TrustPolicyExceptionSpecialArgs,
	Comments:     TrustPolicyExceptionComments,
	ReqData:      &edgeproto.TrustPolicyException{},
	ReplyData:    &edgeproto.TrustPolicyException{},
	Run:          runShowTrustPolicyException,
}

func runShowTrustPolicyException(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.TrustPolicyException)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowTrustPolicyException(c, obj)
}

func ShowTrustPolicyException(c *cli.Command, in *edgeproto.TrustPolicyException) error {
	if TrustPolicyExceptionApiCmd == nil {
		return fmt.Errorf("TrustPolicyExceptionApi client not initialized")
	}
	ctx := context.Background()
	stream, err := TrustPolicyExceptionApiCmd.ShowTrustPolicyException(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowTrustPolicyException failed: %s", errstr)
	}

	objs := make([]*edgeproto.TrustPolicyException, 0)
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
			return fmt.Errorf("ShowTrustPolicyException recv failed: %s", errstr)
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
func ShowTrustPolicyExceptions(c *cli.Command, data []edgeproto.TrustPolicyException, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowTrustPolicyException %v\n", data[ii])
		myerr := ShowTrustPolicyException(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var TrustPolicyExceptionApiCmds = []*cobra.Command{
	CreateTrustPolicyExceptionCmd.GenCmd(),
	UpdateTrustPolicyExceptionCmd.GenCmd(),
	DeleteTrustPolicyExceptionCmd.GenCmd(),
	ShowTrustPolicyExceptionCmd.GenCmd(),
}

var TrustPolicyExceptionKeyRequiredArgs = []string{}
var TrustPolicyExceptionKeyOptionalArgs = []string{
	"appkey.organization",
	"appkey.name",
	"appkey.version",
	"cloudletpoolkey.organization",
	"cloudletpoolkey.name",
	"name",
}
var TrustPolicyExceptionKeyAliasArgs = []string{}
var TrustPolicyExceptionKeyComments = map[string]string{
	"appkey.organization":          "App developer organization",
	"appkey.name":                  "App name",
	"appkey.version":               "App version",
	"cloudletpoolkey.organization": "Name of the organization this pool belongs to",
	"cloudletpoolkey.name":         "CloudletPool Name",
	"name":                         "TrustPolicyExceptionKey name",
}
var TrustPolicyExceptionKeySpecialArgs = map[string]string{}
var TrustPolicyExceptionRequiredArgs = []string{
	"app-org",
	"app-name",
	"app-ver",
	"cloudletpool-org",
	"cloudletpool-name",
	"name",
}
var TrustPolicyExceptionOptionalArgs = []string{
	"state",
	"outboundsecurityrules:empty",
	"outboundsecurityrules:#.protocol",
	"outboundsecurityrules:#.portrangemin",
	"outboundsecurityrules:#.portrangemax",
	"outboundsecurityrules:#.remotecidr",
}
var TrustPolicyExceptionAliasArgs = []string{
	"app-org=key.appkey.organization",
	"app-name=key.appkey.name",
	"app-ver=key.appkey.version",
	"cloudletpool-org=key.cloudletpoolkey.organization",
	"cloudletpool-name=key.cloudletpoolkey.name",
	"name=key.name",
}
var TrustPolicyExceptionComments = map[string]string{
	"fields":                               "Fields are used for the Update API to specify which fields to apply",
	"app-org":                              "App developer organization",
	"app-name":                             "App name",
	"app-ver":                              "App version",
	"cloudletpool-org":                     "Name of the organization this pool belongs to",
	"cloudletpool-name":                    "CloudletPool Name",
	"name":                                 "TrustPolicyExceptionKey name",
	"state":                                "State of the exception within the approval process, one of Unknown, ApprovalRequested, Active, Rejected",
	"outboundsecurityrules:empty":          "List of outbound security rules for whitelisting traffic, specify outboundsecurityrules:empty=true to clear",
	"outboundsecurityrules:#.protocol":     "tcp, udp, icmp",
	"outboundsecurityrules:#.portrangemin": "TCP or UDP port range start",
	"outboundsecurityrules:#.portrangemax": "TCP or UDP port range end",
	"outboundsecurityrules:#.remotecidr":   "remote CIDR X.X.X.X/X",
}
var TrustPolicyExceptionSpecialArgs = map[string]string{
	"fields": "StringArray",
}
var UpdateTrustPolicyExceptionRequiredArgs = []string{
	"app-org",
	"app-name",
	"app-ver",
	"cloudletpool-org",
	"cloudletpool-name",
	"name",
	"state",
}
var UpdateTrustPolicyExceptionOptionalArgs = []string{
	"outboundsecurityrules:empty",
	"outboundsecurityrules:#.protocol",
	"outboundsecurityrules:#.portrangemin",
	"outboundsecurityrules:#.portrangemax",
	"outboundsecurityrules:#.remotecidr",
}
