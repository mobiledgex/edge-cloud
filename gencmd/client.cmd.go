// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: client.proto

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
func AppInstClientHideTags(in *edgeproto.AppInstClient) {
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
}

var AppInstClientApiCmd edgeproto.AppInstClientApiClient

var ShowAppInstClientCmd = &cli.Command{
	Use:          "ShowAppInstClient",
	OptionalArgs: strings.Join(append(AppInstClientKeyRequiredArgs, AppInstClientKeyOptionalArgs...), " "),
	AliasArgs:    strings.Join(AppInstClientKeyAliasArgs, " "),
	SpecialArgs:  &AppInstClientKeySpecialArgs,
	Comments:     AppInstClientKeyComments,
	ReqData:      &edgeproto.AppInstClientKey{},
	ReplyData:    &edgeproto.AppInstClient{},
	Run:          runShowAppInstClient,
}

func runShowAppInstClient(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.AppInstClientKey)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowAppInstClient(c, obj)
}

func ShowAppInstClient(c *cli.Command, in *edgeproto.AppInstClientKey) error {
	if AppInstClientApiCmd == nil {
		return fmt.Errorf("AppInstClientApi client not initialized")
	}
	ctx := context.Background()
	stream, err := AppInstClientApiCmd.ShowAppInstClient(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowAppInstClient failed: %s", errstr)
	}
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
			return fmt.Errorf("ShowAppInstClient recv failed: %s", errstr)
		}
		AppInstClientHideTags(obj)
		c.WriteOutput(obj, cli.OutputFormat)
	}
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func ShowAppInstClients(c *cli.Command, data []edgeproto.AppInstClientKey, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowAppInstClient %v\n", data[ii])
		myerr := ShowAppInstClient(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AppInstClientApiCmds = []*cobra.Command{
	ShowAppInstClientCmd.GenCmd(),
}

var AppInstClientKeyRequiredArgs = []string{
	"developer",
	"appname",
	"appvers",
	"key.clusterinstkey.clusterkey.name",
	"operator",
	"cloudlet",
	"key.clusterinstkey.developer",
}
var AppInstClientKeyOptionalArgs = []string{
	"uuid",
}
var AppInstClientKeyAliasArgs = []string{
	"developer=key.appkey.developerkey.name",
	"appname=key.appkey.name",
	"appvers=key.appkey.version",
	"operator=key.clusterinstkey.cloudletkey.operatorkey.name",
	"cloudlet=key.clusterinstkey.cloudletkey.name",
}
var AppInstClientKeyComments = map[string]string{
	"developer":                          "Organization or Company Name that a Developer is part of",
	"appname":                            "App name",
	"appvers":                            "App version",
	"key.clusterinstkey.clusterkey.name": "Cluster name",
	"operator":                           "Company or Organization name of the operator",
	"cloudlet":                           "Name of the cloudlet",
	"key.clusterinstkey.developer":       "Name of Developer that this cluster belongs to",
	"uuid":                               "App name",
}
var AppInstClientKeySpecialArgs = map[string]string{}
var AppInstClientRequiredArgs = []string{}
var AppInstClientOptionalArgs = []string{
	"clientkey.key.appkey.developerkey.name",
	"clientkey.key.appkey.name",
	"clientkey.key.appkey.version",
	"clientkey.key.clusterinstkey.clusterkey.name",
	"clientkey.key.clusterinstkey.cloudletkey.operatorkey.name",
	"clientkey.key.clusterinstkey.cloudletkey.name",
	"clientkey.key.clusterinstkey.developer",
	"clientkey.uuid",
	"location.latitude",
	"location.longitude",
	"location.horizontalaccuracy",
	"location.verticalaccuracy",
	"location.altitude",
	"location.course",
	"location.speed",
	"location.timestamp.seconds",
	"location.timestamp.nanos",
	"notifyid",
	"status",
}
var AppInstClientAliasArgs = []string{}
var AppInstClientComments = map[string]string{
	"clientkey.key.appkey.developerkey.name":                    "Organization or Company Name that a Developer is part of",
	"clientkey.key.appkey.name":                                 "App name",
	"clientkey.key.appkey.version":                              "App version",
	"clientkey.key.clusterinstkey.clusterkey.name":              "Cluster name",
	"clientkey.key.clusterinstkey.cloudletkey.operatorkey.name": "Company or Organization name of the operator",
	"clientkey.key.clusterinstkey.cloudletkey.name":             "Name of the cloudlet",
	"clientkey.key.clusterinstkey.developer":                    "Name of Developer that this cluster belongs to",
	"clientkey.uuid":                                            "App name",
	"location.latitude":                                         "latitude in WGS 84 coordinates",
	"location.longitude":                                        "longitude in WGS 84 coordinates",
	"location.horizontalaccuracy":                               "horizontal accuracy (radius in meters)",
	"location.verticalaccuracy":                                 "vertical accuracy (meters)",
	"location.altitude":                                         "On android only lat and long are guaranteed to be supplied altitude in meters",
	"location.course":                                           "course (IOS) / bearing (Android) (degrees east relative to true north)",
	"location.speed":                                            "speed (IOS) / velocity (Android) (meters/sec)",
	"notifyid":                                                  "Id of client assigned by server (internal use only)",
	"status":                                                    "Status return, one of FindUnknown, FindFound, FindNotfound",
}
var AppInstClientSpecialArgs = map[string]string{}
