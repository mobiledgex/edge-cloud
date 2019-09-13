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
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/protocmd"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
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
}

var ExecApiCmd edgeproto.ExecApiClient

var RunCommandCmd = &cli.Command{
	Use:          "RunCommand",
	RequiredArgs: strings.Join(ExecRequestRequiredArgs, " "),
	OptionalArgs: strings.Join(ExecRequestOptionalArgs, " "),
	AliasArgs:    strings.Join(ExecRequestAliasArgs, " "),
	SpecialArgs:  &ExecRequestSpecialArgs,
	Comments:     ExecRequestComments,
	ReqData:      &edgeproto.ExecRequest{},
	ReplyData:    &edgeproto.ExecRequest{},
	Run:          runRunCommand,
}

func runRunCommand(c *cli.Command, args []string) error {
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj := c.ReqData.(*edgeproto.ExecRequest)
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
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj := c.ReqData.(*edgeproto.ExecRequest)
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
	SendLocalRequestCmd.GenCmd(),
}

var ExecRequestRequiredArgs = []string{}
var ExecRequestOptionalArgs = []string{
	"developer",
	"appname",
	"appvers",
	"cluster",
	"operator",
	"cloudlet",
	"clusterdeveloper",
	"command",
	"containerid",
}
var ExecRequestAliasArgs = []string{
	"developer=appinstkey.appkey.developerkey.name",
	"appname=appinstkey.appkey.name",
	"appvers=appinstkey.appkey.version",
	"cluster=appinstkey.clusterinstkey.clusterkey.name",
	"operator=appinstkey.clusterinstkey.cloudletkey.operatorkey.name",
	"cloudlet=appinstkey.clusterinstkey.cloudletkey.name",
	"clusterdeveloper=appinstkey.clusterinstkey.developer",
}
var ExecRequestComments = map[string]string{
	"developer":        "Organization or Company Name that a Developer is part of",
	"appname":          "App name",
	"appvers":          "App version",
	"cluster":          "Cluster name",
	"operator":         "Company or Organization name of the operator",
	"cloudlet":         "Name of the cloudlet",
	"clusterdeveloper": "Name of Developer that this cluster belongs to",
	"command":          "Command or Shell",
	"containerid":      "ContainerID is the name of the target container, if applicable",
	"offer":            "WebRTC Offer",
	"answer":           "WebRTC Answer",
	"err":              "Any error message",
}
var ExecRequestSpecialArgs = map[string]string{}
