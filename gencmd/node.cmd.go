// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: node.proto

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
import _ "github.com/gogo/protobuf/gogoproto"
import _ "github.com/mobiledgex/edge-cloud/protogen"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
func NodeKeyHideTags(in *edgeproto.NodeKey) {
	if cli.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cli.HideTags, ",") {
		tags[tag] = struct{}{}
	}
	if _, found := tags["nocmp"]; found {
		in.Name = ""
	}
}

func NodeHideTags(in *edgeproto.Node) {
	if cli.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cli.HideTags, ",") {
		tags[tag] = struct{}{}
	}
	if _, found := tags["nocmp"]; found {
		in.Key.Name = ""
	}
	if _, found := tags["nocmp"]; found {
		in.NotifyId = 0
	}
	if _, found := tags["nocmp"]; found {
		in.BuildMaster = ""
	}
	if _, found := tags["nocmp"]; found {
		in.BuildHead = ""
	}
	if _, found := tags["nocmp"]; found {
		in.BuildAuthor = ""
	}
	if _, found := tags["nocmp"]; found {
		in.Hostname = ""
	}
}

var NodeApiCmd edgeproto.NodeApiClient

var ShowNodeCmd = &cli.Command{
	Use:          "ShowNode",
	OptionalArgs: strings.Join(append(NodeRequiredArgs, NodeOptionalArgs...), " "),
	AliasArgs:    strings.Join(NodeAliasArgs, " "),
	SpecialArgs:  &NodeSpecialArgs,
	Comments:     NodeComments,
	ReqData:      &edgeproto.Node{},
	ReplyData:    &edgeproto.Node{},
	Run:          runShowNode,
}

func runShowNode(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.Node)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowNode(c, obj)
}

func ShowNode(c *cli.Command, in *edgeproto.Node) error {
	if NodeApiCmd == nil {
		return fmt.Errorf("NodeApi client not initialized")
	}
	ctx := context.Background()
	stream, err := NodeApiCmd.ShowNode(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowNode failed: %s", errstr)
	}
	objs := make([]*edgeproto.Node, 0)
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
			return fmt.Errorf("ShowNode recv failed: %s", errstr)
		}
		NodeHideTags(obj)
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(objs, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func ShowNodes(c *cli.Command, data []edgeproto.Node, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowNode %v\n", data[ii])
		myerr := ShowNode(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var NodeApiCmds = []*cobra.Command{
	ShowNodeCmd.GenCmd(),
}

var NodeKeyRequiredArgs = []string{}
var NodeKeyOptionalArgs = []string{
	"name",
	"type",
	"cloudletkey.organization",
	"cloudletkey.name",
	"region",
}
var NodeKeyAliasArgs = []string{}
var NodeKeyComments = map[string]string{
	"name":                     "Name or hostname of node",
	"type":                     "Node type",
	"cloudletkey.organization": "Organization of the cloudlet site",
	"cloudletkey.name":         "Name of the cloudlet",
	"region":                   "Region the node is in",
}
var NodeKeySpecialArgs = map[string]string{}
var NodeRequiredArgs = []string{
	"name",
	"type",
	"organization",
	"cloudlet",
	"region",
}
var NodeOptionalArgs = []string{
	"notifyid",
	"buildmaster",
	"buildhead",
	"buildauthor",
	"hostname",
	"containerversion",
}
var NodeAliasArgs = []string{
	"name=key.name",
	"type=key.type",
	"organization=key.cloudletkey.organization",
	"cloudlet=key.cloudletkey.name",
	"region=key.region",
}
var NodeComments = map[string]string{
	"name":             "Name or hostname of node",
	"type":             "Node type",
	"organization":     "Organization of the cloudlet site",
	"cloudlet":         "Name of the cloudlet",
	"region":           "Region the node is in",
	"notifyid":         "Id of client assigned by server (internal use only)",
	"buildmaster":      "Build Master Version",
	"buildhead":        "Build Head Version",
	"buildauthor":      "Build Author",
	"hostname":         "Hostname",
	"containerversion": "Docker edge-cloud container version which node instance use",
}
var NodeSpecialArgs = map[string]string{}
