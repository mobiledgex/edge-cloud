// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: alert.proto

/*
Package gencmd is a generated protocol buffer package.

It is generated from these files:
	alert.proto
	app.proto
	app_inst.proto
	autoscalepolicy.proto
	cloudlet.proto
	cloudletpool.proto
	cluster.proto
	clusterinst.proto
	common.proto
	controller.proto
	developer.proto
	exec.proto
	flavor.proto
	metric.proto
	node.proto
	notice.proto
	operator.proto
	refs.proto
	restagtable.proto
	result.proto
	version.proto

It has these top-level messages:
	Alert
	AppKey
	ConfigFile
	App
	AppInstKey
	AppInst
	AppInstRuntime
	AppInstInfo
	AppInstMetrics
	PolicyKey
	AutoScalePolicy
	CloudletKey
	OperationTimeLimits
	CloudletInfraCommon
	AzureProperties
	GcpProperties
	OpenStackProperties
	CloudletInfraProperties
	PlatformConfig
	Cloudlet
	FlavorInfo
	CloudletInfo
	CloudletMetrics
	CloudletPoolKey
	CloudletPool
	CloudletPoolMember
	ClusterKey
	ClusterInstKey
	ClusterInst
	ClusterInstInfo
	StatusInfo
	ControllerKey
	Controller
	DeveloperKey
	Developer
	ExecRequest
	FlavorKey
	Flavor
	MetricTag
	MetricVal
	Metric
	NodeKey
	Node
	Notice
	OperatorKey
	Operator
	CloudletRefs
	ClusterRefs
	ResTagTableKey
	ResTagTable
	Result
*/
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
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
func AlertHideTags(in *edgeproto.Alert) {
	if cli.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cli.HideTags, ",") {
		tags[tag] = struct{}{}
	}
	if _, found := tags["nocmp"]; found {
		in.NotifyId = 0
	}
	if _, found := tags["nocmp"]; found {
		in.Controller = ""
	}
}

var AlertApiCmd edgeproto.AlertApiClient

var ShowAlertCmd = &cli.Command{
	Use:          "ShowAlert",
	OptionalArgs: strings.Join(append(AlertRequiredArgs, AlertOptionalArgs...), " "),
	AliasArgs:    strings.Join(AlertAliasArgs, " "),
	SpecialArgs:  &AlertSpecialArgs,
	Comments:     AlertComments,
	ReqData:      &edgeproto.Alert{},
	ReplyData:    &edgeproto.Alert{},
	Run:          runShowAlert,
}

func runShowAlert(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.Alert)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowAlert(c, obj)
}

func ShowAlert(c *cli.Command, in *edgeproto.Alert) error {
	if AlertApiCmd == nil {
		return fmt.Errorf("AlertApi client not initialized")
	}
	ctx := context.Background()
	stream, err := AlertApiCmd.ShowAlert(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowAlert failed: %s", errstr)
	}
	objs := make([]*edgeproto.Alert, 0)
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ShowAlert recv failed: %s", err.Error())
		}
		AlertHideTags(obj)
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(objs, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func ShowAlerts(c *cli.Command, data []edgeproto.Alert, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowAlert %v\n", data[ii])
		myerr := ShowAlert(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AlertApiCmds = []*cobra.Command{
	ShowAlertCmd.GenCmd(),
}

var AlertRequiredArgs = []string{}
var AlertOptionalArgs = []string{
	"labels",
	"annotations",
	"state",
	"activeat.seconds",
	"activeat.nanos",
	"value",
	"notifyid",
	"controller",
}
var AlertAliasArgs = []string{}
var AlertComments = map[string]string{
	"labels":      "Labels uniquely define the alert",
	"annotations": "Annotations are extra information about the alert",
	"state":       "State of the alert",
	"value":       "Any value associated with alert",
	"notifyid":    "Id of client assigned by server (internal use only)",
	"controller":  "Connected controller unique id",
}
var AlertSpecialArgs = map[string]string{
	"annotations": "StringToString",
	"labels":      "StringToString",
}
var LabelsEntryRequiredArgs = []string{}
var LabelsEntryOptionalArgs = []string{
	"key",
	"value",
}
var LabelsEntryAliasArgs = []string{}
var LabelsEntryComments = map[string]string{}
var LabelsEntrySpecialArgs = map[string]string{}
var AnnotationsEntryRequiredArgs = []string{}
var AnnotationsEntryOptionalArgs = []string{
	"key",
	"value",
}
var AnnotationsEntryAliasArgs = []string{}
var AnnotationsEntryComments = map[string]string{}
var AnnotationsEntrySpecialArgs = map[string]string{}
