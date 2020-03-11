// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: autoprovpolicy.proto

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
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/types"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
var AutoProvPolicyApiCmd edgeproto.AutoProvPolicyApiClient

var CreateAutoProvPolicyCmd = &cli.Command{
	Use:          "CreateAutoProvPolicy",
	RequiredArgs: strings.Join(CreateAutoProvPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(CreateAutoProvPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(AutoProvPolicyAliasArgs, " "),
	SpecialArgs:  &AutoProvPolicySpecialArgs,
	Comments:     AutoProvPolicyComments,
	ReqData:      &edgeproto.AutoProvPolicy{},
	ReplyData:    &edgeproto.Result{},
	Run:          runCreateAutoProvPolicy,
}

func runCreateAutoProvPolicy(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.AutoProvPolicy)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return CreateAutoProvPolicy(c, obj)
}

func CreateAutoProvPolicy(c *cli.Command, in *edgeproto.AutoProvPolicy) error {
	if AutoProvPolicyApiCmd == nil {
		return fmt.Errorf("AutoProvPolicyApi client not initialized")
	}
	ctx := context.Background()
	obj, err := AutoProvPolicyApiCmd.CreateAutoProvPolicy(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateAutoProvPolicy failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func CreateAutoProvPolicys(c *cli.Command, data []edgeproto.AutoProvPolicy, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("CreateAutoProvPolicy %v\n", data[ii])
		myerr := CreateAutoProvPolicy(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var DeleteAutoProvPolicyCmd = &cli.Command{
	Use:          "DeleteAutoProvPolicy",
	RequiredArgs: strings.Join(AutoProvPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(AutoProvPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(AutoProvPolicyAliasArgs, " "),
	SpecialArgs:  &AutoProvPolicySpecialArgs,
	Comments:     AutoProvPolicyComments,
	ReqData:      &edgeproto.AutoProvPolicy{},
	ReplyData:    &edgeproto.Result{},
	Run:          runDeleteAutoProvPolicy,
}

func runDeleteAutoProvPolicy(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.AutoProvPolicy)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return DeleteAutoProvPolicy(c, obj)
}

func DeleteAutoProvPolicy(c *cli.Command, in *edgeproto.AutoProvPolicy) error {
	if AutoProvPolicyApiCmd == nil {
		return fmt.Errorf("AutoProvPolicyApi client not initialized")
	}
	ctx := context.Background()
	obj, err := AutoProvPolicyApiCmd.DeleteAutoProvPolicy(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("DeleteAutoProvPolicy failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func DeleteAutoProvPolicys(c *cli.Command, data []edgeproto.AutoProvPolicy, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("DeleteAutoProvPolicy %v\n", data[ii])
		myerr := DeleteAutoProvPolicy(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var UpdateAutoProvPolicyCmd = &cli.Command{
	Use:          "UpdateAutoProvPolicy",
	RequiredArgs: strings.Join(AutoProvPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(AutoProvPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(AutoProvPolicyAliasArgs, " "),
	SpecialArgs:  &AutoProvPolicySpecialArgs,
	Comments:     AutoProvPolicyComments,
	ReqData:      &edgeproto.AutoProvPolicy{},
	ReplyData:    &edgeproto.Result{},
	Run:          runUpdateAutoProvPolicy,
}

func runUpdateAutoProvPolicy(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.AutoProvPolicy)
	jsonMap, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj.Fields = cli.GetSpecifiedFields(jsonMap, c.ReqData, cli.JsonNamespace)
	return UpdateAutoProvPolicy(c, obj)
}

func UpdateAutoProvPolicy(c *cli.Command, in *edgeproto.AutoProvPolicy) error {
	if AutoProvPolicyApiCmd == nil {
		return fmt.Errorf("AutoProvPolicyApi client not initialized")
	}
	ctx := context.Background()
	obj, err := AutoProvPolicyApiCmd.UpdateAutoProvPolicy(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("UpdateAutoProvPolicy failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func UpdateAutoProvPolicys(c *cli.Command, data []edgeproto.AutoProvPolicy, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("UpdateAutoProvPolicy %v\n", data[ii])
		myerr := UpdateAutoProvPolicy(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ShowAutoProvPolicyCmd = &cli.Command{
	Use:          "ShowAutoProvPolicy",
	OptionalArgs: strings.Join(append(AutoProvPolicyRequiredArgs, AutoProvPolicyOptionalArgs...), " "),
	AliasArgs:    strings.Join(AutoProvPolicyAliasArgs, " "),
	SpecialArgs:  &AutoProvPolicySpecialArgs,
	Comments:     AutoProvPolicyComments,
	ReqData:      &edgeproto.AutoProvPolicy{},
	ReplyData:    &edgeproto.AutoProvPolicy{},
	Run:          runShowAutoProvPolicy,
}

func runShowAutoProvPolicy(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.AutoProvPolicy)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowAutoProvPolicy(c, obj)
}

func ShowAutoProvPolicy(c *cli.Command, in *edgeproto.AutoProvPolicy) error {
	if AutoProvPolicyApiCmd == nil {
		return fmt.Errorf("AutoProvPolicyApi client not initialized")
	}
	ctx := context.Background()
	stream, err := AutoProvPolicyApiCmd.ShowAutoProvPolicy(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowAutoProvPolicy failed: %s", errstr)
	}
	objs := make([]*edgeproto.AutoProvPolicy, 0)
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
			return fmt.Errorf("ShowAutoProvPolicy recv failed: %s", errstr)
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
func ShowAutoProvPolicys(c *cli.Command, data []edgeproto.AutoProvPolicy, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowAutoProvPolicy %v\n", data[ii])
		myerr := ShowAutoProvPolicy(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AddAutoProvPolicyCloudletCmd = &cli.Command{
	Use:          "AddAutoProvPolicyCloudlet",
	RequiredArgs: strings.Join(AutoProvPolicyCloudletRequiredArgs, " "),
	OptionalArgs: strings.Join(AutoProvPolicyCloudletOptionalArgs, " "),
	AliasArgs:    strings.Join(AutoProvPolicyCloudletAliasArgs, " "),
	SpecialArgs:  &AutoProvPolicyCloudletSpecialArgs,
	Comments:     AutoProvPolicyCloudletComments,
	ReqData:      &edgeproto.AutoProvPolicyCloudlet{},
	ReplyData:    &edgeproto.Result{},
	Run:          runAddAutoProvPolicyCloudlet,
}

func runAddAutoProvPolicyCloudlet(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.AutoProvPolicyCloudlet)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return AddAutoProvPolicyCloudlet(c, obj)
}

func AddAutoProvPolicyCloudlet(c *cli.Command, in *edgeproto.AutoProvPolicyCloudlet) error {
	if AutoProvPolicyApiCmd == nil {
		return fmt.Errorf("AutoProvPolicyApi client not initialized")
	}
	ctx := context.Background()
	obj, err := AutoProvPolicyApiCmd.AddAutoProvPolicyCloudlet(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("AddAutoProvPolicyCloudlet failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func AddAutoProvPolicyCloudlets(c *cli.Command, data []edgeproto.AutoProvPolicyCloudlet, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("AddAutoProvPolicyCloudlet %v\n", data[ii])
		myerr := AddAutoProvPolicyCloudlet(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var RemoveAutoProvPolicyCloudletCmd = &cli.Command{
	Use:          "RemoveAutoProvPolicyCloudlet",
	RequiredArgs: strings.Join(AutoProvPolicyCloudletRequiredArgs, " "),
	OptionalArgs: strings.Join(AutoProvPolicyCloudletOptionalArgs, " "),
	AliasArgs:    strings.Join(AutoProvPolicyCloudletAliasArgs, " "),
	SpecialArgs:  &AutoProvPolicyCloudletSpecialArgs,
	Comments:     AutoProvPolicyCloudletComments,
	ReqData:      &edgeproto.AutoProvPolicyCloudlet{},
	ReplyData:    &edgeproto.Result{},
	Run:          runRemoveAutoProvPolicyCloudlet,
}

func runRemoveAutoProvPolicyCloudlet(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.AutoProvPolicyCloudlet)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return RemoveAutoProvPolicyCloudlet(c, obj)
}

func RemoveAutoProvPolicyCloudlet(c *cli.Command, in *edgeproto.AutoProvPolicyCloudlet) error {
	if AutoProvPolicyApiCmd == nil {
		return fmt.Errorf("AutoProvPolicyApi client not initialized")
	}
	ctx := context.Background()
	obj, err := AutoProvPolicyApiCmd.RemoveAutoProvPolicyCloudlet(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("RemoveAutoProvPolicyCloudlet failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func RemoveAutoProvPolicyCloudlets(c *cli.Command, data []edgeproto.AutoProvPolicyCloudlet, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("RemoveAutoProvPolicyCloudlet %v\n", data[ii])
		myerr := RemoveAutoProvPolicyCloudlet(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AutoProvPolicyApiCmds = []*cobra.Command{
	CreateAutoProvPolicyCmd.GenCmd(),
	DeleteAutoProvPolicyCmd.GenCmd(),
	UpdateAutoProvPolicyCmd.GenCmd(),
	ShowAutoProvPolicyCmd.GenCmd(),
	AddAutoProvPolicyCloudletCmd.GenCmd(),
	RemoveAutoProvPolicyCloudletCmd.GenCmd(),
}

var AutoProvPolicyRequiredArgs = []string{
	"organization",
	"name",
}
var AutoProvPolicyOptionalArgs = []string{
	"deployclientcount",
	"deployintervalcount",
}
var AutoProvPolicyAliasArgs = []string{
	"organization=key.organization",
	"name=key.name",
}
var AutoProvPolicyComments = map[string]string{
	"organization":                     "Name of the organization that this policy belongs to",
	"name":                             "Policy name",
	"deployclientcount":                "Minimum number of clients within the auto deploy interval to trigger deployment",
	"deployintervalcount":              "Number of intervals to check before triggering deployment",
	"cloudlets.key.organization":       "Organization of the cloudlet site",
	"cloudlets.key.name":               "Name of the cloudlet",
	"cloudlets.loc.latitude":           "latitude in WGS 84 coordinates",
	"cloudlets.loc.longitude":          "longitude in WGS 84 coordinates",
	"cloudlets.loc.horizontalaccuracy": "horizontal accuracy (radius in meters)",
	"cloudlets.loc.verticalaccuracy":   "vertical accuracy (meters)",
	"cloudlets.loc.altitude":           "On android only lat and long are guaranteed to be supplied altitude in meters",
	"cloudlets.loc.course":             "course (IOS) / bearing (Android) (degrees east relative to true north)",
	"cloudlets.loc.speed":              "speed (IOS) / velocity (Android) (meters/sec)",
}
var AutoProvPolicySpecialArgs = map[string]string{}
var AutoProvCloudletRequiredArgs = []string{
	"key.organization",
	"key.name",
}
var AutoProvCloudletOptionalArgs = []string{
	"loc.latitude",
	"loc.longitude",
	"loc.horizontalaccuracy",
	"loc.verticalaccuracy",
	"loc.altitude",
	"loc.course",
	"loc.speed",
	"loc.timestamp.seconds",
	"loc.timestamp.nanos",
}
var AutoProvCloudletAliasArgs = []string{}
var AutoProvCloudletComments = map[string]string{
	"key.organization":       "Organization of the cloudlet site",
	"key.name":               "Name of the cloudlet",
	"loc.latitude":           "latitude in WGS 84 coordinates",
	"loc.longitude":          "longitude in WGS 84 coordinates",
	"loc.horizontalaccuracy": "horizontal accuracy (radius in meters)",
	"loc.verticalaccuracy":   "vertical accuracy (meters)",
	"loc.altitude":           "On android only lat and long are guaranteed to be supplied altitude in meters",
	"loc.course":             "course (IOS) / bearing (Android) (degrees east relative to true north)",
	"loc.speed":              "speed (IOS) / velocity (Android) (meters/sec)",
}
var AutoProvCloudletSpecialArgs = map[string]string{}
var AutoProvCountRequiredArgs = []string{}
var AutoProvCountOptionalArgs = []string{
	"appkey.organization",
	"appkey.name",
	"appkey.version",
	"cloudletkey.organization",
	"cloudletkey.name",
	"count",
	"processnow",
	"deploynowkey.clusterkey.name",
	"deploynowkey.cloudletkey.organization",
	"deploynowkey.cloudletkey.name",
	"deploynowkey.organization",
}
var AutoProvCountAliasArgs = []string{}
var AutoProvCountComments = map[string]string{
	"appkey.organization":                   "Developer Organization",
	"appkey.name":                           "App name",
	"appkey.version":                        "App version",
	"cloudletkey.organization":              "Organization of the cloudlet site",
	"cloudletkey.name":                      "Name of the cloudlet",
	"count":                                 "FindCloudlet client count",
	"processnow":                            "Process count immediately",
	"deploynowkey.clusterkey.name":          "Cluster name",
	"deploynowkey.cloudletkey.organization": "Organization of the cloudlet site",
	"deploynowkey.cloudletkey.name":         "Name of the cloudlet",
	"deploynowkey.organization":             "Name of Developer organization that this cluster belongs to",
}
var AutoProvCountSpecialArgs = map[string]string{}
var AutoProvCountsRequiredArgs = []string{}
var AutoProvCountsOptionalArgs = []string{
	"dmenodename",
	"timestamp.seconds",
	"timestamp.nanos",
	"counts.appkey.organization",
	"counts.appkey.name",
	"counts.appkey.version",
	"counts.cloudletkey.organization",
	"counts.cloudletkey.name",
	"counts.count",
	"counts.processnow",
	"counts.deploynowkey.clusterkey.name",
	"counts.deploynowkey.cloudletkey.organization",
	"counts.deploynowkey.cloudletkey.name",
	"counts.deploynowkey.organization",
}
var AutoProvCountsAliasArgs = []string{}
var AutoProvCountsComments = map[string]string{
	"dmenodename":                                  "DME node name",
	"timestamp.seconds":                            "Represents seconds of UTC time since Unix epoch 1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z inclusive.",
	"timestamp.nanos":                              "Non-negative fractions of a second at nanosecond resolution. Negative second values with fractions must still have non-negative nanos values that count forward in time. Must be from 0 to 999,999,999 inclusive.",
	"counts.appkey.organization":                   "Developer Organization",
	"counts.appkey.name":                           "App name",
	"counts.appkey.version":                        "App version",
	"counts.cloudletkey.organization":              "Organization of the cloudlet site",
	"counts.cloudletkey.name":                      "Name of the cloudlet",
	"counts.count":                                 "FindCloudlet client count",
	"counts.processnow":                            "Process count immediately",
	"counts.deploynowkey.clusterkey.name":          "Cluster name",
	"counts.deploynowkey.cloudletkey.organization": "Organization of the cloudlet site",
	"counts.deploynowkey.cloudletkey.name":         "Name of the cloudlet",
	"counts.deploynowkey.organization":             "Name of Developer organization that this cluster belongs to",
}
var AutoProvCountsSpecialArgs = map[string]string{}
var AutoProvPolicyCloudletRequiredArgs = []string{
	"organization",
	"name",
}
var AutoProvPolicyCloudletOptionalArgs = []string{
	"cloudlet.org",
	"cloudlet",
}
var AutoProvPolicyCloudletAliasArgs = []string{
	"organization=key.organization",
	"name=key.name",
	"cloudlet.org=cloudletkey.organization",
	"cloudlet=cloudletkey.name",
}
var AutoProvPolicyCloudletComments = map[string]string{
	"organization": "Name of the organization that this policy belongs to",
	"name":         "Policy name",
	"cloudlet.org": "Organization of the cloudlet site",
	"cloudlet":     "Name of the cloudlet",
}
var AutoProvPolicyCloudletSpecialArgs = map[string]string{}
var CreateAutoProvPolicyRequiredArgs = []string{
	"organization",
	"name",
}
var CreateAutoProvPolicyOptionalArgs = []string{
	"deployclientcount",
	"deployintervalcount",
}
