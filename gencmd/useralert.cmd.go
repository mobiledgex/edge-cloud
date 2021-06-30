// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: useralert.proto

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
var UserAlertApiCmd edgeproto.UserAlertApiClient

var CreateUserAlertCmd = &cli.Command{
	Use:          "CreateUserAlert",
	RequiredArgs: strings.Join(CreateUserAlertRequiredArgs, " "),
	OptionalArgs: strings.Join(CreateUserAlertOptionalArgs, " "),
	AliasArgs:    strings.Join(UserAlertAliasArgs, " "),
	SpecialArgs:  &UserAlertSpecialArgs,
	Comments:     UserAlertComments,
	ReqData:      &edgeproto.UserAlert{},
	ReplyData:    &edgeproto.Result{},
	Run:          runCreateUserAlert,
}

func runCreateUserAlert(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.UserAlert)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return CreateUserAlert(c, obj)
}

func CreateUserAlert(c *cli.Command, in *edgeproto.UserAlert) error {
	if UserAlertApiCmd == nil {
		return fmt.Errorf("UserAlertApi client not initialized")
	}
	ctx := context.Background()
	obj, err := UserAlertApiCmd.CreateUserAlert(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateUserAlert failed: %s", errstr)
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func CreateUserAlerts(c *cli.Command, data []edgeproto.UserAlert, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("CreateUserAlert %v\n", data[ii])
		myerr := CreateUserAlert(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var DeleteUserAlertCmd = &cli.Command{
	Use:          "DeleteUserAlert",
	RequiredArgs: strings.Join(UserAlertRequiredArgs, " "),
	OptionalArgs: strings.Join(UserAlertOptionalArgs, " "),
	AliasArgs:    strings.Join(UserAlertAliasArgs, " "),
	SpecialArgs:  &UserAlertSpecialArgs,
	Comments:     UserAlertComments,
	ReqData:      &edgeproto.UserAlert{},
	ReplyData:    &edgeproto.Result{},
	Run:          runDeleteUserAlert,
}

func runDeleteUserAlert(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.UserAlert)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return DeleteUserAlert(c, obj)
}

func DeleteUserAlert(c *cli.Command, in *edgeproto.UserAlert) error {
	if UserAlertApiCmd == nil {
		return fmt.Errorf("UserAlertApi client not initialized")
	}
	ctx := context.Background()
	obj, err := UserAlertApiCmd.DeleteUserAlert(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("DeleteUserAlert failed: %s", errstr)
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func DeleteUserAlerts(c *cli.Command, data []edgeproto.UserAlert, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("DeleteUserAlert %v\n", data[ii])
		myerr := DeleteUserAlert(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var UpdateUserAlertCmd = &cli.Command{
	Use:          "UpdateUserAlert",
	RequiredArgs: strings.Join(UserAlertRequiredArgs, " "),
	OptionalArgs: strings.Join(UserAlertOptionalArgs, " "),
	AliasArgs:    strings.Join(UserAlertAliasArgs, " "),
	SpecialArgs:  &UserAlertSpecialArgs,
	Comments:     UserAlertComments,
	ReqData:      &edgeproto.UserAlert{},
	ReplyData:    &edgeproto.Result{},
	Run:          runUpdateUserAlert,
}

func runUpdateUserAlert(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.UserAlert)
	jsonMap, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj.Fields = cli.GetSpecifiedFields(jsonMap, c.ReqData, cli.JsonNamespace)
	return UpdateUserAlert(c, obj)
}

func UpdateUserAlert(c *cli.Command, in *edgeproto.UserAlert) error {
	if UserAlertApiCmd == nil {
		return fmt.Errorf("UserAlertApi client not initialized")
	}
	ctx := context.Background()
	obj, err := UserAlertApiCmd.UpdateUserAlert(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("UpdateUserAlert failed: %s", errstr)
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func UpdateUserAlerts(c *cli.Command, data []edgeproto.UserAlert, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("UpdateUserAlert %v\n", data[ii])
		myerr := UpdateUserAlert(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ShowUserAlertCmd = &cli.Command{
	Use:          "ShowUserAlert",
	OptionalArgs: strings.Join(append(UserAlertRequiredArgs, UserAlertOptionalArgs...), " "),
	AliasArgs:    strings.Join(UserAlertAliasArgs, " "),
	SpecialArgs:  &UserAlertSpecialArgs,
	Comments:     UserAlertComments,
	ReqData:      &edgeproto.UserAlert{},
	ReplyData:    &edgeproto.UserAlert{},
	Run:          runShowUserAlert,
}

func runShowUserAlert(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.UserAlert)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowUserAlert(c, obj)
}

func ShowUserAlert(c *cli.Command, in *edgeproto.UserAlert) error {
	if UserAlertApiCmd == nil {
		return fmt.Errorf("UserAlertApi client not initialized")
	}
	ctx := context.Background()
	stream, err := UserAlertApiCmd.ShowUserAlert(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowUserAlert failed: %s", errstr)
	}

	objs := make([]*edgeproto.UserAlert, 0)
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
			return fmt.Errorf("ShowUserAlert recv failed: %s", errstr)
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
func ShowUserAlerts(c *cli.Command, data []edgeproto.UserAlert, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowUserAlert %v\n", data[ii])
		myerr := ShowUserAlert(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var UserAlertApiCmds = []*cobra.Command{
	CreateUserAlertCmd.GenCmd(),
	DeleteUserAlertCmd.GenCmd(),
	UpdateUserAlertCmd.GenCmd(),
	ShowUserAlertCmd.GenCmd(),
}

var UserAlertKeyRequiredArgs = []string{}
var UserAlertKeyOptionalArgs = []string{
	"organization",
	"name",
}
var UserAlertKeyAliasArgs = []string{}
var UserAlertKeyComments = map[string]string{
	"organization": "Name of the organization for the app that this alert can be applied to",
	"name":         "Alert name",
}
var UserAlertKeySpecialArgs = map[string]string{}
var UserAlertRequiredArgs = []string{
	"alert-org",
	"name",
}
var UserAlertOptionalArgs = []string{
	"cpu-utilization",
	"mem-usage",
	"disk-usage",
	"active-connections",
	"severity",
	"trigger-time",
	"labels",
	"annotations",
}
var UserAlertAliasArgs = []string{
	"alert-org=key.organization",
	"name=key.name",
	"cpu-utilization=cpulimit",
	"mem-usage=memlimit",
	"disk-usage=disklimit",
	"active-connections=activeconnlimit",
	"trigger-time=triggertime",
}
var UserAlertComments = map[string]string{
	"alert-org":          "Name of the organization for the app that this alert can be applied to",
	"name":               "Alert name",
	"cpu-utilization":    "CPU",
	"mem-usage":          "Mem",
	"disk-usage":         "Disk",
	"active-connections": "Active Connections",
	"severity":           "Alert Severity",
	"trigger-time":       "Trigger threshold interval",
	"labels":             "Additional Labels",
	"annotations":        "Additional Annotations for extra information about the alert",
}
var UserAlertSpecialArgs = map[string]string{
	"annotations": "StringToString",
	"fields":      "StringArray",
	"labels":      "StringToString",
}
var CreateUserAlertRequiredArgs = []string{
	"alert-org",
	"name",
	"severity",
}
var CreateUserAlertOptionalArgs = []string{
	"cpu-utilization",
	"mem-usage",
	"disk-usage",
	"active-connections",
	"trigger-time",
	"labels",
	"annotations",
}
