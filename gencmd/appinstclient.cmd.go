// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: appinstclient.proto

package gencmd

import (
	"context"
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud/cli"
	_ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
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
func AppInstClientHideTags(in *edgeproto.AppInstClient) {
	if cli.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cli.HideTags, ",") {
		tags[tag] = struct{}{}
	}
	if _, found := tags["nocmp"]; found {
		in.Location = distributed_match_engine.Loc{}
	}
	if _, found := tags["nocmp"]; found {
		in.NotifyId = 0
	}
}

var AppInstClientApiCmd edgeproto.AppInstClientApiClient

var ShowAppInstClientCmd = &cli.Command{
	Use:          "ShowAppInstClient",
	RequiredArgs: strings.Join(AppInstClientKeyRequiredArgs, " "),
	OptionalArgs: strings.Join(AppInstClientKeyOptionalArgs, " "),
	AliasArgs:    strings.Join(AppInstClientKeyAliasArgs, " "),
	SpecialArgs:  &AppInstClientKeySpecialArgs,
	Comments:     AppInstClientKeyComments,
	ReqData:      &edgeproto.AppInstClientKey{},
	ReplyData:    &edgeproto.AppInstClient{},
	Run:          runShowAppInstClient,
}

func runShowAppInstClient(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
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

	objs := make([]*edgeproto.AppInstClient, 0)
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
		if cli.OutputStream {
			c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
			continue
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

var StreamAppInstClientsLocalCmd = &cli.Command{
	Use:          "StreamAppInstClientsLocal",
	RequiredArgs: strings.Join(AppInstClientKeyRequiredArgs, " "),
	OptionalArgs: strings.Join(AppInstClientKeyOptionalArgs, " "),
	AliasArgs:    strings.Join(AppInstClientKeyAliasArgs, " "),
	SpecialArgs:  &AppInstClientKeySpecialArgs,
	Comments:     AppInstClientKeyComments,
	ReqData:      &edgeproto.AppInstClientKey{},
	ReplyData:    &edgeproto.AppInstClient{},
	Run:          runStreamAppInstClientsLocal,
}

func runStreamAppInstClientsLocal(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.AppInstClientKey)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return StreamAppInstClientsLocal(c, obj)
}

func StreamAppInstClientsLocal(c *cli.Command, in *edgeproto.AppInstClientKey) error {
	if AppInstClientApiCmd == nil {
		return fmt.Errorf("AppInstClientApi client not initialized")
	}
	ctx := context.Background()
	stream, err := AppInstClientApiCmd.StreamAppInstClientsLocal(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("StreamAppInstClientsLocal failed: %s", errstr)
	}

	objs := make([]*edgeproto.AppInstClient, 0)
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
			return fmt.Errorf("StreamAppInstClientsLocal recv failed: %s", errstr)
		}
		AppInstClientHideTags(obj)
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), objs, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func StreamAppInstClientsLocals(c *cli.Command, data []edgeproto.AppInstClientKey, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("StreamAppInstClientsLocal %v\n", data[ii])
		myerr := StreamAppInstClientsLocal(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AppInstClientApiCmds = []*cobra.Command{
	ShowAppInstClientCmd.GenCmd(),
	StreamAppInstClientsLocalCmd.GenCmd(),
}

var AppInstClientKeyRequiredArgs = []string{
	"apporg",
}
var AppInstClientKeyOptionalArgs = []string{
	"appname",
	"appvers",
	"cluster",
	"cloudletorg",
	"cloudlet",
	"federatedorg",
	"clusterorg",
	"uniqueid",
	"uniqueidtype",
}
var AppInstClientKeyAliasArgs = []string{
	"apporg=appinstkey.appkey.organization",
	"appname=appinstkey.appkey.name",
	"appvers=appinstkey.appkey.version",
	"cluster=appinstkey.clusterinstkey.clusterkey.name",
	"cloudletorg=appinstkey.clusterinstkey.cloudletkey.organization",
	"cloudlet=appinstkey.clusterinstkey.cloudletkey.name",
	"federatedorg=appinstkey.clusterinstkey.cloudletkey.federatedorganization",
	"clusterorg=appinstkey.clusterinstkey.organization",
	"uniqueid=uniqueid",
	"uniqueidtype=uniqueidtype",
}
var AppInstClientKeyComments = map[string]string{
	"apporg":       "App developer organization",
	"appname":      "App name",
	"appvers":      "App version",
	"cluster":      "Cluster name",
	"cloudletorg":  "Organization of the cloudlet site",
	"cloudlet":     "Name of the cloudlet",
	"federatedorg": "Federated operator organization who shared this cloudlet",
	"clusterorg":   "Name of Developer organization that this cluster belongs to",
	"uniqueid":     "AppInstClient Unique Id",
	"uniqueidtype": "AppInstClient Unique Id Type",
}
var AppInstClientKeySpecialArgs = map[string]string{}
var AppInstClientRequiredArgs = []string{}
var AppInstClientOptionalArgs = []string{
	"clientkey.appinstkey.appkey.organization",
	"clientkey.appinstkey.appkey.name",
	"clientkey.appinstkey.appkey.version",
	"clientkey.appinstkey.clusterinstkey.clusterkey.name",
	"clientkey.appinstkey.clusterinstkey.cloudletkey.organization",
	"clientkey.appinstkey.clusterinstkey.cloudletkey.name",
	"clientkey.appinstkey.clusterinstkey.cloudletkey.federatedorganization",
	"clientkey.appinstkey.clusterinstkey.organization",
	"clientkey.uniqueid",
	"clientkey.uniqueidtype",
	"location.latitude",
	"location.longitude",
	"location.horizontalaccuracy",
	"location.verticalaccuracy",
	"location.altitude",
	"location.course",
	"location.speed",
	"location.timestamp",
	"notifyid",
}
var AppInstClientAliasArgs = []string{}
var AppInstClientComments = map[string]string{
	"fields": "Fields are used for the Update API to specify which fields to apply",
	"clientkey.appinstkey.appkey.organization":                              "App developer organization",
	"clientkey.appinstkey.appkey.name":                                      "App name",
	"clientkey.appinstkey.appkey.version":                                   "App version",
	"clientkey.appinstkey.clusterinstkey.clusterkey.name":                   "Cluster name",
	"clientkey.appinstkey.clusterinstkey.cloudletkey.organization":          "Organization of the cloudlet site",
	"clientkey.appinstkey.clusterinstkey.cloudletkey.name":                  "Name of the cloudlet",
	"clientkey.appinstkey.clusterinstkey.cloudletkey.federatedorganization": "Federated operator organization who shared this cloudlet",
	"clientkey.appinstkey.clusterinstkey.organization":                      "Name of Developer organization that this cluster belongs to",
	"clientkey.uniqueid":                                                    "AppInstClient Unique Id",
	"clientkey.uniqueidtype":                                                "AppInstClient Unique Id Type",
	"location.latitude":                                                     "Latitude in WGS 84 coordinates",
	"location.longitude":                                                    "Longitude in WGS 84 coordinates",
	"location.horizontalaccuracy":                                           "Horizontal accuracy (radius in meters)",
	"location.verticalaccuracy":                                             "Vertical accuracy (meters)",
	"location.altitude":                                                     "On android only lat and long are guaranteed to be supplied Altitude in meters",
	"location.course":                                                       "Course (IOS) / bearing (Android) (degrees east relative to true north)",
	"location.speed":                                                        "Speed (IOS) / velocity (Android) (meters/sec)",
	"location.timestamp":                                                    "Timestamp",
	"notifyid":                                                              "Id of client assigned by server (internal use only)",
}
var AppInstClientSpecialArgs = map[string]string{
	"fields": "StringArray",
}
