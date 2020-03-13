// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: autoscalepolicy.proto

package gencmd

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
import "github.com/spf13/cobra"
import "context"
import "io"
import "github.com/mobiledgex/edge-cloud/cli"
import "google.golang.org/grpc/status"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
var AutoScalePolicyApiCmd edgeproto.AutoScalePolicyApiClient

var CreateAutoScalePolicyCmd = &cli.Command{
	Use:          "CreateAutoScalePolicy",
	RequiredArgs: strings.Join(CreateAutoScalePolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(CreateAutoScalePolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(AutoScalePolicyAliasArgs, " "),
	SpecialArgs:  &AutoScalePolicySpecialArgs,
	Comments:     AutoScalePolicyComments,
	ReqData:      &edgeproto.AutoScalePolicy{},
	ReplyData:    &edgeproto.Result{},
	Run:          runCreateAutoScalePolicy,
}

func runCreateAutoScalePolicy(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.AutoScalePolicy)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return CreateAutoScalePolicy(c, obj)
}

func CreateAutoScalePolicy(c *cli.Command, in *edgeproto.AutoScalePolicy) error {
	if AutoScalePolicyApiCmd == nil {
		return fmt.Errorf("AutoScalePolicyApi client not initialized")
	}
	ctx := context.Background()
	obj, err := AutoScalePolicyApiCmd.CreateAutoScalePolicy(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateAutoScalePolicy failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func CreateAutoScalePolicys(c *cli.Command, data []edgeproto.AutoScalePolicy, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("CreateAutoScalePolicy %v\n", data[ii])
		myerr := CreateAutoScalePolicy(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var DeleteAutoScalePolicyCmd = &cli.Command{
	Use:          "DeleteAutoScalePolicy",
	RequiredArgs: strings.Join(AutoScalePolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(AutoScalePolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(AutoScalePolicyAliasArgs, " "),
	SpecialArgs:  &AutoScalePolicySpecialArgs,
	Comments:     AutoScalePolicyComments,
	ReqData:      &edgeproto.AutoScalePolicy{},
	ReplyData:    &edgeproto.Result{},
	Run:          runDeleteAutoScalePolicy,
}

func runDeleteAutoScalePolicy(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.AutoScalePolicy)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return DeleteAutoScalePolicy(c, obj)
}

func DeleteAutoScalePolicy(c *cli.Command, in *edgeproto.AutoScalePolicy) error {
	if AutoScalePolicyApiCmd == nil {
		return fmt.Errorf("AutoScalePolicyApi client not initialized")
	}
	ctx := context.Background()
	obj, err := AutoScalePolicyApiCmd.DeleteAutoScalePolicy(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("DeleteAutoScalePolicy failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func DeleteAutoScalePolicys(c *cli.Command, data []edgeproto.AutoScalePolicy, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("DeleteAutoScalePolicy %v\n", data[ii])
		myerr := DeleteAutoScalePolicy(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var UpdateAutoScalePolicyCmd = &cli.Command{
	Use:          "UpdateAutoScalePolicy",
	RequiredArgs: strings.Join(AutoScalePolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(AutoScalePolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(AutoScalePolicyAliasArgs, " "),
	SpecialArgs:  &AutoScalePolicySpecialArgs,
	Comments:     AutoScalePolicyComments,
	ReqData:      &edgeproto.AutoScalePolicy{},
	ReplyData:    &edgeproto.Result{},
	Run:          runUpdateAutoScalePolicy,
}

func runUpdateAutoScalePolicy(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.AutoScalePolicy)
	jsonMap, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj.Fields = cli.GetSpecifiedFields(jsonMap, c.ReqData, cli.JsonNamespace)
	return UpdateAutoScalePolicy(c, obj)
}

func UpdateAutoScalePolicy(c *cli.Command, in *edgeproto.AutoScalePolicy) error {
	if AutoScalePolicyApiCmd == nil {
		return fmt.Errorf("AutoScalePolicyApi client not initialized")
	}
	ctx := context.Background()
	obj, err := AutoScalePolicyApiCmd.UpdateAutoScalePolicy(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("UpdateAutoScalePolicy failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func UpdateAutoScalePolicys(c *cli.Command, data []edgeproto.AutoScalePolicy, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("UpdateAutoScalePolicy %v\n", data[ii])
		myerr := UpdateAutoScalePolicy(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ShowAutoScalePolicyCmd = &cli.Command{
	Use:          "ShowAutoScalePolicy",
	OptionalArgs: strings.Join(append(AutoScalePolicyRequiredArgs, AutoScalePolicyOptionalArgs...), " "),
	AliasArgs:    strings.Join(AutoScalePolicyAliasArgs, " "),
	SpecialArgs:  &AutoScalePolicySpecialArgs,
	Comments:     AutoScalePolicyComments,
	ReqData:      &edgeproto.AutoScalePolicy{},
	ReplyData:    &edgeproto.AutoScalePolicy{},
	Run:          runShowAutoScalePolicy,
}

func runShowAutoScalePolicy(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.AutoScalePolicy)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowAutoScalePolicy(c, obj)
}

func ShowAutoScalePolicy(c *cli.Command, in *edgeproto.AutoScalePolicy) error {
	if AutoScalePolicyApiCmd == nil {
		return fmt.Errorf("AutoScalePolicyApi client not initialized")
	}
	ctx := context.Background()
	stream, err := AutoScalePolicyApiCmd.ShowAutoScalePolicy(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowAutoScalePolicy failed: %s", errstr)
	}

	objs := make([]*edgeproto.AutoScalePolicy, 0)
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
			return fmt.Errorf("ShowAutoScalePolicy recv failed: %s", errstr)
		}
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(objs, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func ShowAutoScalePolicys(c *cli.Command, data []edgeproto.AutoScalePolicy, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowAutoScalePolicy %v\n", data[ii])
		myerr := ShowAutoScalePolicy(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AutoScalePolicyApiCmds = []*cobra.Command{
	CreateAutoScalePolicyCmd.GenCmd(),
	DeleteAutoScalePolicyCmd.GenCmd(),
	UpdateAutoScalePolicyCmd.GenCmd(),
	ShowAutoScalePolicyCmd.GenCmd(),
}

var PolicyKeyRequiredArgs = []string{}
var PolicyKeyOptionalArgs = []string{
	"organization",
	"name",
}
var PolicyKeyAliasArgs = []string{}
var PolicyKeyComments = map[string]string{
	"organization": "Name of the organization for the cluster that this policy will apply to",
	"name":         "Policy name",
}
var PolicyKeySpecialArgs = map[string]string{}
var AutoScalePolicyRequiredArgs = []string{
	"cluster-org",
	"name",
}
var AutoScalePolicyOptionalArgs = []string{
	"minnodes",
	"maxnodes",
	"scaleupcputhresh",
	"scaledowncputhresh",
	"triggertimesec",
}
var AutoScalePolicyAliasArgs = []string{
	"cluster-org=key.organization",
	"name=key.name",
}
var AutoScalePolicyComments = map[string]string{
	"fields":             "Fields are used for the Update API to specify which fields to apply",
	"cluster-org":        "Name of the organization for the cluster that this policy will apply to",
	"name":               "Policy name",
	"minnodes":           "Minimum number of cluster nodes",
	"maxnodes":           "Maximum number of cluster nodes",
	"scaleupcputhresh":   "Scale up cpu threshold (percentage 1 to 100)",
	"scaledowncputhresh": "Scale down cpu threshold (percentage 1 to 100)",
	"triggertimesec":     "Trigger time defines how long trigger threshold must be satified in seconds before acting upon it.",
}
var AutoScalePolicySpecialArgs = map[string]string{
	"fields": "StringArray",
}
var CreateAutoScalePolicyRequiredArgs = []string{
	"cluster-org",
	"name",
	"minnodes",
	"maxnodes",
}
var CreateAutoScalePolicyOptionalArgs = []string{
	"scaleupcputhresh",
	"scaledowncputhresh",
	"triggertimesec",
}
