// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: exec.proto

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
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
func RunVMConsoleHideTags(in *edgeproto.RunVMConsole) {
	if cli.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cli.HideTags, ",") {
		tags[tag] = struct{}{}
	}
	if _, found := tags["nocmp"]; found {
		in.Url = ""
	}
}

func ExecRequestHideTags(in *edgeproto.ExecRequest) {
	if cli.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cli.HideTags, ",") {
		tags[tag] = struct{}{}
	}
	if _, found := tags["nocmp"]; found {
		in.Offer = ""
	}
	if _, found := tags["nocmp"]; found {
		in.Answer = ""
	}
	if _, found := tags["nocmp"]; found {
		in.Console.Url = ""
	}
}

var ExecApiCmd edgeproto.ExecApiClient

var RunCommandCmd = &cli.Command{
	Use:          "RunCommand",
	RequiredArgs: strings.Join(RunCommandRequiredArgs, " "),
	OptionalArgs: strings.Join(RunCommandOptionalArgs, " "),
	AliasArgs:    strings.Join(ExecRequestAliasArgs, " "),
	SpecialArgs:  &ExecRequestSpecialArgs,
	Comments:     ExecRequestComments,
	ReqData:      &edgeproto.ExecRequest{},
	ReplyData:    &edgeproto.ExecRequest{},
	Run:          runRunCommand,
}

func runRunCommand(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.ExecRequest)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return RunCommand(c, obj)
}

func RunCommand(c *cli.Command, in *edgeproto.ExecRequest) error {
	if ExecApiCmd == nil {
		return fmt.Errorf("ExecApi client not initialized")
	}
	ctx := context.Background()
	obj, err := ExecApiCmd.RunCommand(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("RunCommand failed: %s", errstr)
	}
	ExecRequestHideTags(obj)
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func RunCommands(c *cli.Command, data []edgeproto.ExecRequest, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("RunCommand %v\n", data[ii])
		myerr := RunCommand(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var RunConsoleCmd = &cli.Command{
	Use:          "RunConsole",
	RequiredArgs: strings.Join(RunConsoleRequiredArgs, " "),
	OptionalArgs: strings.Join(RunConsoleOptionalArgs, " "),
	AliasArgs:    strings.Join(ExecRequestAliasArgs, " "),
	SpecialArgs:  &ExecRequestSpecialArgs,
	Comments:     ExecRequestComments,
	ReqData:      &edgeproto.ExecRequest{},
	ReplyData:    &edgeproto.ExecRequest{},
	Run:          runRunConsole,
}

func runRunConsole(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.ExecRequest)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return RunConsole(c, obj)
}

func RunConsole(c *cli.Command, in *edgeproto.ExecRequest) error {
	if ExecApiCmd == nil {
		return fmt.Errorf("ExecApi client not initialized")
	}
	ctx := context.Background()
	obj, err := ExecApiCmd.RunConsole(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("RunConsole failed: %s", errstr)
	}
	ExecRequestHideTags(obj)
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func RunConsoles(c *cli.Command, data []edgeproto.ExecRequest, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("RunConsole %v\n", data[ii])
		myerr := RunConsole(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ShowLogsCmd = &cli.Command{
	Use:          "ShowLogs",
	RequiredArgs: strings.Join(ShowLogsRequiredArgs, " "),
	OptionalArgs: strings.Join(ShowLogsOptionalArgs, " "),
	AliasArgs:    strings.Join(ExecRequestAliasArgs, " "),
	SpecialArgs:  &ExecRequestSpecialArgs,
	Comments:     ExecRequestComments,
	ReqData:      &edgeproto.ExecRequest{},
	ReplyData:    &edgeproto.ExecRequest{},
	Run:          runShowLogs,
}

func runShowLogs(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.ExecRequest)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowLogs(c, obj)
}

func ShowLogs(c *cli.Command, in *edgeproto.ExecRequest) error {
	if ExecApiCmd == nil {
		return fmt.Errorf("ExecApi client not initialized")
	}
	ctx := context.Background()
	obj, err := ExecApiCmd.ShowLogs(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowLogs failed: %s", errstr)
	}
	ExecRequestHideTags(obj)
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func ShowLogss(c *cli.Command, data []edgeproto.ExecRequest, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowLogs %v\n", data[ii])
		myerr := ShowLogs(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var SendLocalRequestCmd = &cli.Command{
	Use:          "SendLocalRequest",
	RequiredArgs: strings.Join(ExecRequestRequiredArgs, " "),
	OptionalArgs: strings.Join(ExecRequestOptionalArgs, " "),
	AliasArgs:    strings.Join(ExecRequestAliasArgs, " "),
	SpecialArgs:  &ExecRequestSpecialArgs,
	Comments:     ExecRequestComments,
	ReqData:      &edgeproto.ExecRequest{},
	ReplyData:    &edgeproto.ExecRequest{},
	Run:          runSendLocalRequest,
}

func runSendLocalRequest(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.ExecRequest)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return SendLocalRequest(c, obj)
}

func SendLocalRequest(c *cli.Command, in *edgeproto.ExecRequest) error {
	if ExecApiCmd == nil {
		return fmt.Errorf("ExecApi client not initialized")
	}
	ctx := context.Background()
	obj, err := ExecApiCmd.SendLocalRequest(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("SendLocalRequest failed: %s", errstr)
	}
	ExecRequestHideTags(obj)
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func SendLocalRequests(c *cli.Command, data []edgeproto.ExecRequest, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("SendLocalRequest %v\n", data[ii])
		myerr := SendLocalRequest(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ExecApiCmds = []*cobra.Command{
	RunCommandCmd.GenCmd(),
	RunConsoleCmd.GenCmd(),
	ShowLogsCmd.GenCmd(),
	SendLocalRequestCmd.GenCmd(),
}

var RunCmdRequiredArgs = []string{}
var RunCmdOptionalArgs = []string{
	"command",
}
var RunCmdAliasArgs = []string{}
var RunCmdComments = map[string]string{
	"command": "Command or Shell",
}
var RunCmdSpecialArgs = map[string]string{}
var RunVMConsoleRequiredArgs = []string{}
var RunVMConsoleOptionalArgs = []string{
	"url",
}
var RunVMConsoleAliasArgs = []string{}
var RunVMConsoleComments = map[string]string{
	"url": "VM Console URL",
}
var RunVMConsoleSpecialArgs = map[string]string{}
var ShowLogRequiredArgs = []string{}
var ShowLogOptionalArgs = []string{
	"since",
	"tail",
	"timestamps",
	"follow",
}
var ShowLogAliasArgs = []string{}
var ShowLogComments = map[string]string{
	"since":      "Show logs since either a duration ago (5s, 2m, 3h) or a timestamp (RFC3339)",
	"tail":       "Show only a recent number of lines",
	"timestamps": "Show timestamps",
	"follow":     "Stream data",
}
var ShowLogSpecialArgs = map[string]string{}
var ExecRequestRequiredArgs = []string{
	"organization",
	"appname",
	"appvers",
	"cluster",
	"operatororg",
	"cloudlet",
	"clusterdevorg",
}
var ExecRequestOptionalArgs = []string{
	"containerid",
	"command",
	"since",
	"tail",
	"timestamps",
	"follow",
}
var ExecRequestAliasArgs = []string{
	"organization=appinstkey.appkey.organization",
	"appname=appinstkey.appkey.name",
	"appvers=appinstkey.appkey.version",
	"cluster=appinstkey.clusterinstkey.clusterkey.name",
	"operatororg=appinstkey.clusterinstkey.cloudletkey.organization",
	"cloudlet=appinstkey.clusterinstkey.cloudletkey.name",
	"clusterdevorg=appinstkey.clusterinstkey.organization",
	"command=cmd.command",
	"since=log.since",
	"tail=log.tail",
	"timestamps=log.timestamps",
	"follow=log.follow",
}
var ExecRequestComments = map[string]string{
	"organization":  "Developer Organization",
	"appname":       "App name",
	"appvers":       "App version",
	"cluster":       "Cluster name",
	"operatororg":   "Operator of the cloudlet site",
	"cloudlet":      "Name of the cloudlet",
	"clusterdevorg": "Name of Developer organization that this cluster belongs to",
	"containerid":   "ContainerId is the name or ID of the target container, if applicable",
	"offer":         "WebRTC Offer",
	"answer":        "WebRTC Answer",
	"err":           "Any error message",
	"command":       "Command or Shell",
	"since":         "Show logs since either a duration ago (5s, 2m, 3h) or a timestamp (RFC3339)",
	"tail":          "Show only a recent number of lines",
	"timestamps":    "Show timestamps",
	"follow":        "Stream data",
	"console.url":   "VM Console URL",
	"timeout":       "Timeout",
}
var ExecRequestSpecialArgs = map[string]string{}
var RunCommandRequiredArgs = []string{
	"organization",
	"appname",
	"appvers",
	"cluster",
	"operatororg",
	"cloudlet",
	"clusterdevorg",
	"command",
}
var RunCommandOptionalArgs = []string{
	"containerid",
}
var RunConsoleRequiredArgs = []string{
	"organization",
	"appname",
	"appvers",
	"cluster",
	"operatororg",
	"cloudlet",
	"clusterdevorg",
}
var RunConsoleOptionalArgs = []string{
	"containerid",
}
var ShowLogsRequiredArgs = []string{
	"organization",
	"appname",
	"appvers",
	"cluster",
	"operatororg",
	"cloudlet",
	"clusterdevorg",
}
var ShowLogsOptionalArgs = []string{
	"containerid",
	"since",
	"tail",
	"timestamps",
	"follow",
}
