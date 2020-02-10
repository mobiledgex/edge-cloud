// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: settings.proto

package gencmd

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
import "github.com/spf13/cobra"
import "context"
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
var SettingsApiCmd edgeproto.SettingsApiClient

var UpdateSettingsCmd = &cli.Command{
	Use:          "UpdateSettings",
	RequiredArgs: strings.Join(SettingsRequiredArgs, " "),
	OptionalArgs: strings.Join(SettingsOptionalArgs, " "),
	AliasArgs:    strings.Join(SettingsAliasArgs, " "),
	SpecialArgs:  &SettingsSpecialArgs,
	Comments:     SettingsComments,
	ReqData:      &edgeproto.Settings{},
	ReplyData:    &edgeproto.Result{},
	Run:          runUpdateSettings,
}

func runUpdateSettings(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.Settings)
	jsonMap, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj.Fields = cli.GetSpecifiedFields(jsonMap, c.ReqData, cli.JsonNamespace)
	return UpdateSettings(c, obj)
}

func UpdateSettings(c *cli.Command, in *edgeproto.Settings) error {
	if SettingsApiCmd == nil {
		return fmt.Errorf("SettingsApi client not initialized")
	}
	ctx := context.Background()
	obj, err := SettingsApiCmd.UpdateSettings(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("UpdateSettings failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func UpdateSettingsBatch(c *cli.Command, data *edgeproto.Settings, err *error) {
	if *err != nil || data == nil {
		return
	}
	fmt.Printf("UpdateSettings %v\n", data)
	myerr := UpdateSettings(c, data)
	if myerr != nil {
		*err = myerr
	}
}

var ResetSettingsCmd = &cli.Command{
	Use:          "ResetSettings",
	RequiredArgs: strings.Join(SettingsRequiredArgs, " "),
	OptionalArgs: strings.Join(SettingsOptionalArgs, " "),
	AliasArgs:    strings.Join(SettingsAliasArgs, " "),
	SpecialArgs:  &SettingsSpecialArgs,
	Comments:     SettingsComments,
	ReqData:      &edgeproto.Settings{},
	ReplyData:    &edgeproto.Result{},
	Run:          runResetSettings,
}

func runResetSettings(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.Settings)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ResetSettings(c, obj)
}

func ResetSettings(c *cli.Command, in *edgeproto.Settings) error {
	if SettingsApiCmd == nil {
		return fmt.Errorf("SettingsApi client not initialized")
	}
	ctx := context.Background()
	obj, err := SettingsApiCmd.ResetSettings(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ResetSettings failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func ResetSettingsBatch(c *cli.Command, data *edgeproto.Settings, err *error) {
	if *err != nil || data == nil {
		return
	}
	fmt.Printf("ResetSettings %v\n", data)
	myerr := ResetSettings(c, data)
	if myerr != nil {
		*err = myerr
	}
}

var ShowSettingsCmd = &cli.Command{
	Use:          "ShowSettings",
	OptionalArgs: strings.Join(append(SettingsRequiredArgs, SettingsOptionalArgs...), " "),
	AliasArgs:    strings.Join(SettingsAliasArgs, " "),
	SpecialArgs:  &SettingsSpecialArgs,
	Comments:     SettingsComments,
	ReqData:      &edgeproto.Settings{},
	ReplyData:    &edgeproto.Settings{},
	Run:          runShowSettings,
}

func runShowSettings(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.Settings)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowSettings(c, obj)
}

func ShowSettings(c *cli.Command, in *edgeproto.Settings) error {
	if SettingsApiCmd == nil {
		return fmt.Errorf("SettingsApi client not initialized")
	}
	ctx := context.Background()
	obj, err := SettingsApiCmd.ShowSettings(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowSettings failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func ShowSettingsBatch(c *cli.Command, data *edgeproto.Settings, err *error) {
	if *err != nil || data == nil {
		return
	}
	fmt.Printf("ShowSettings %v\n", data)
	myerr := ShowSettings(c, data)
	if myerr != nil {
		*err = myerr
	}
}

var SettingsApiCmds = []*cobra.Command{
	UpdateSettingsCmd.GenCmd(),
	ResetSettingsCmd.GenCmd(),
	ShowSettingsCmd.GenCmd(),
}

var SettingsRequiredArgs = []string{}
var SettingsOptionalArgs = []string{
	"shepherdmetricscollectioninterval",
	"shepherdhealthcheckretries",
	"shepherdhealthcheckinterval",
	"autodeployintervalsec",
	"autodeployoffsetsec",
	"autodeploymaxintervals",
	"createappinsttimeout",
	"updateappinsttimeout",
	"deleteappinsttimeout",
	"createclusterinsttimeout",
	"updateclusterinsttimeout",
	"deleteclusterinsttimeout",
	"masternodeflavor",
	"loadbalancermaxportrange",
}
var SettingsAliasArgs = []string{}
var SettingsComments = map[string]string{
	"shepherdmetricscollectioninterval": "Shepherd metrics collection interval for k8s and docker appInstances (duration)",
	"shepherdhealthcheckretries":        "Number of times Shepherd Health Check fails before we mark appInst down",
	"shepherdhealthcheckinterval":       "Health Checking probing frequency (duration)",
	"autodeployintervalsec":             "Auto Provisioning Stats push and analysis interval (seconds)",
	"autodeployoffsetsec":               "Auto Provisioning analysis offset from interval (seconds)",
	"autodeploymaxintervals":            "Auto Provisioning Policy max allowed intervals",
	"createappinsttimeout":              "Create AppInst timeout (duration)",
	"updateappinsttimeout":              "Update AppInst timeout (duration)",
	"deleteappinsttimeout":              "Delete AppInst timeout (duration)",
	"createclusterinsttimeout":          "Create ClusterInst timeout (duration)",
	"updateclusterinsttimeout":          "Update ClusterInst timeout (duration)",
	"deleteclusterinsttimeout":          "Delete ClusterInst timeout (duration)",
	"masternodeflavor":                  "Default flavor for k8s master VM and > 0  workers",
	"loadbalancermaxportrange":          "Max IP Port range when using a load balancer",
}
var SettingsSpecialArgs = map[string]string{}
