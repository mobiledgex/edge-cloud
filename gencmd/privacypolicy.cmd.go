// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: privacypolicy.proto

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
var PrivacyPolicyApiCmd edgeproto.PrivacyPolicyApiClient

var CreatePrivacyPolicyCmd = &cli.Command{
	Use:          "CreatePrivacyPolicy",
	RequiredArgs: strings.Join(PrivacyPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(PrivacyPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(PrivacyPolicyAliasArgs, " "),
	SpecialArgs:  &PrivacyPolicySpecialArgs,
	Comments:     PrivacyPolicyComments,
	ReqData:      &edgeproto.PrivacyPolicy{},
	ReplyData:    &edgeproto.Result{},
	Run:          runCreatePrivacyPolicy,
}

func runCreatePrivacyPolicy(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.PrivacyPolicy)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return CreatePrivacyPolicy(c, obj)
}

func CreatePrivacyPolicy(c *cli.Command, in *edgeproto.PrivacyPolicy) error {
	if PrivacyPolicyApiCmd == nil {
		return fmt.Errorf("PrivacyPolicyApi client not initialized")
	}
	ctx := context.Background()
	obj, err := PrivacyPolicyApiCmd.CreatePrivacyPolicy(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreatePrivacyPolicy failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func CreatePrivacyPolicys(c *cli.Command, data []edgeproto.PrivacyPolicy, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("CreatePrivacyPolicy %v\n", data[ii])
		myerr := CreatePrivacyPolicy(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var DeletePrivacyPolicyCmd = &cli.Command{
	Use:          "DeletePrivacyPolicy",
	RequiredArgs: strings.Join(PrivacyPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(PrivacyPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(PrivacyPolicyAliasArgs, " "),
	SpecialArgs:  &PrivacyPolicySpecialArgs,
	Comments:     PrivacyPolicyComments,
	ReqData:      &edgeproto.PrivacyPolicy{},
	ReplyData:    &edgeproto.Result{},
	Run:          runDeletePrivacyPolicy,
}

func runDeletePrivacyPolicy(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.PrivacyPolicy)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return DeletePrivacyPolicy(c, obj)
}

func DeletePrivacyPolicy(c *cli.Command, in *edgeproto.PrivacyPolicy) error {
	if PrivacyPolicyApiCmd == nil {
		return fmt.Errorf("PrivacyPolicyApi client not initialized")
	}
	ctx := context.Background()
	obj, err := PrivacyPolicyApiCmd.DeletePrivacyPolicy(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("DeletePrivacyPolicy failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func DeletePrivacyPolicys(c *cli.Command, data []edgeproto.PrivacyPolicy, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("DeletePrivacyPolicy %v\n", data[ii])
		myerr := DeletePrivacyPolicy(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var UpdatePrivacyPolicyCmd = &cli.Command{
	Use:          "UpdatePrivacyPolicy",
	RequiredArgs: strings.Join(PrivacyPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(PrivacyPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(PrivacyPolicyAliasArgs, " "),
	SpecialArgs:  &PrivacyPolicySpecialArgs,
	Comments:     PrivacyPolicyComments,
	ReqData:      &edgeproto.PrivacyPolicy{},
	ReplyData:    &edgeproto.Result{},
	Run:          runUpdatePrivacyPolicy,
}

func runUpdatePrivacyPolicy(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.PrivacyPolicy)
	jsonMap, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj.Fields = cli.GetSpecifiedFields(jsonMap, c.ReqData, cli.JsonNamespace)
	return UpdatePrivacyPolicy(c, obj)
}

func UpdatePrivacyPolicy(c *cli.Command, in *edgeproto.PrivacyPolicy) error {
	if PrivacyPolicyApiCmd == nil {
		return fmt.Errorf("PrivacyPolicyApi client not initialized")
	}
	ctx := context.Background()
	obj, err := PrivacyPolicyApiCmd.UpdatePrivacyPolicy(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("UpdatePrivacyPolicy failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func UpdatePrivacyPolicys(c *cli.Command, data []edgeproto.PrivacyPolicy, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("UpdatePrivacyPolicy %v\n", data[ii])
		myerr := UpdatePrivacyPolicy(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ShowPrivacyPolicyCmd = &cli.Command{
	Use:          "ShowPrivacyPolicy",
	OptionalArgs: strings.Join(append(PrivacyPolicyRequiredArgs, PrivacyPolicyOptionalArgs...), " "),
	AliasArgs:    strings.Join(PrivacyPolicyAliasArgs, " "),
	SpecialArgs:  &PrivacyPolicySpecialArgs,
	Comments:     PrivacyPolicyComments,
	ReqData:      &edgeproto.PrivacyPolicy{},
	ReplyData:    &edgeproto.PrivacyPolicy{},
	Run:          runShowPrivacyPolicy,
}

func runShowPrivacyPolicy(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.PrivacyPolicy)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowPrivacyPolicy(c, obj)
}

func ShowPrivacyPolicy(c *cli.Command, in *edgeproto.PrivacyPolicy) error {
	if PrivacyPolicyApiCmd == nil {
		return fmt.Errorf("PrivacyPolicyApi client not initialized")
	}
	ctx := context.Background()
	stream, err := PrivacyPolicyApiCmd.ShowPrivacyPolicy(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowPrivacyPolicy failed: %s", errstr)
	}

	objs := make([]*edgeproto.PrivacyPolicy, 0)
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
			return fmt.Errorf("ShowPrivacyPolicy recv failed: %s", errstr)
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
func ShowPrivacyPolicys(c *cli.Command, data []edgeproto.PrivacyPolicy, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowPrivacyPolicy %v\n", data[ii])
		myerr := ShowPrivacyPolicy(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var PrivacyPolicyApiCmds = []*cobra.Command{
	CreatePrivacyPolicyCmd.GenCmd(),
	DeletePrivacyPolicyCmd.GenCmd(),
	UpdatePrivacyPolicyCmd.GenCmd(),
	ShowPrivacyPolicyCmd.GenCmd(),
}

var OutboundSecurityRuleRequiredArgs = []string{}
var OutboundSecurityRuleOptionalArgs = []string{
	"protocol",
	"portrangemin",
	"portrangemax",
	"remotecidr",
}
var OutboundSecurityRuleAliasArgs = []string{}
var OutboundSecurityRuleComments = map[string]string{
	"protocol":     "tcp, udp, icmp",
	"portrangemin": "TCP or UDP port range start",
	"portrangemax": "TCP or UDP port range end",
	"remotecidr":   "remote CIDR X.X.X.X/X",
}
var OutboundSecurityRuleSpecialArgs = map[string]string{}
var PrivacyPolicyRequiredArgs = []string{
	"cluster-org",
	"name",
}
var PrivacyPolicyOptionalArgs = []string{
	"outboundsecurityrules[#].protocol",
	"outboundsecurityrules[#].portrangemin",
	"outboundsecurityrules[#].portrangemax",
	"outboundsecurityrules[#].remotecidr",
}
var PrivacyPolicyAliasArgs = []string{
	"cluster-org=key.organization",
	"name=key.name",
}
var PrivacyPolicyComments = map[string]string{
	"fields":                                "Fields are used for the Update API to specify which fields to apply",
	"cluster-org":                           "Name of the organization for the cluster that this policy will apply to",
	"name":                                  "Policy name",
	"outboundsecurityrules[#].protocol":     "tcp, udp, icmp",
	"outboundsecurityrules[#].portrangemin": "TCP or UDP port range start",
	"outboundsecurityrules[#].portrangemax": "TCP or UDP port range end",
	"outboundsecurityrules[#].remotecidr":   "remote CIDR X.X.X.X/X",
}
var PrivacyPolicySpecialArgs = map[string]string{
	"fields": "StringArray",
}
