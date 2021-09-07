// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: stream.proto

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
func StreamObjHideTags(in *edgeproto.StreamObj) {
	if cli.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cli.HideTags, ",") {
		tags[tag] = struct{}{}
	}
	if _, found := tags["nocmp"]; found {
		in.Status.TaskNumber = 0
	}
	if _, found := tags["nocmp"]; found {
		in.Status.TaskName = ""
	}
}

var StreamObjApiCmd edgeproto.StreamObjApiClient

var StreamAppInstCmd = &cli.Command{
	Use:          "StreamAppInst",
	RequiredArgs: strings.Join(AppInstKeyRequiredArgs, " "),
	OptionalArgs: strings.Join(AppInstKeyOptionalArgs, " "),
	AliasArgs:    strings.Join(AppInstKeyAliasArgs, " "),
	SpecialArgs:  &AppInstKeySpecialArgs,
	Comments:     AppInstKeyComments,
	ReqData:      &edgeproto.AppInstKey{},
	ReplyData:    &edgeproto.Result{},
	Run:          runStreamAppInst,
}

func runStreamAppInst(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.AppInstKey)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return StreamAppInst(c, obj)
}

func StreamAppInst(c *cli.Command, in *edgeproto.AppInstKey) error {
	if StreamObjApiCmd == nil {
		return fmt.Errorf("StreamObjApi client not initialized")
	}
	ctx := context.Background()
	stream, err := StreamObjApiCmd.StreamAppInst(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("StreamAppInst failed: %s", errstr)
	}

	objs := make([]*edgeproto.Result, 0)
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
			return fmt.Errorf("StreamAppInst recv failed: %s", errstr)
		}
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
func StreamAppInsts(c *cli.Command, data []edgeproto.AppInstKey, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("StreamAppInst %v\n", data[ii])
		myerr := StreamAppInst(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var StreamClusterInstCmd = &cli.Command{
	Use:          "StreamClusterInst",
	RequiredArgs: strings.Join(ClusterInstKeyRequiredArgs, " "),
	OptionalArgs: strings.Join(ClusterInstKeyOptionalArgs, " "),
	AliasArgs:    strings.Join(ClusterInstKeyAliasArgs, " "),
	SpecialArgs:  &ClusterInstKeySpecialArgs,
	Comments:     ClusterInstKeyComments,
	ReqData:      &edgeproto.ClusterInstKey{},
	ReplyData:    &edgeproto.Result{},
	Run:          runStreamClusterInst,
}

func runStreamClusterInst(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.ClusterInstKey)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return StreamClusterInst(c, obj)
}

func StreamClusterInst(c *cli.Command, in *edgeproto.ClusterInstKey) error {
	if StreamObjApiCmd == nil {
		return fmt.Errorf("StreamObjApi client not initialized")
	}
	ctx := context.Background()
	stream, err := StreamObjApiCmd.StreamClusterInst(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("StreamClusterInst failed: %s", errstr)
	}

	objs := make([]*edgeproto.Result, 0)
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
			return fmt.Errorf("StreamClusterInst recv failed: %s", errstr)
		}
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
func StreamClusterInsts(c *cli.Command, data []edgeproto.ClusterInstKey, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("StreamClusterInst %v\n", data[ii])
		myerr := StreamClusterInst(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var StreamCloudletCmd = &cli.Command{
	Use:          "StreamCloudlet",
	RequiredArgs: strings.Join(CloudletKeyRequiredArgs, " "),
	OptionalArgs: strings.Join(CloudletKeyOptionalArgs, " "),
	AliasArgs:    strings.Join(CloudletKeyAliasArgs, " "),
	SpecialArgs:  &CloudletKeySpecialArgs,
	Comments:     CloudletKeyComments,
	ReqData:      &edgeproto.CloudletKey{},
	ReplyData:    &edgeproto.Result{},
	Run:          runStreamCloudlet,
}

func runStreamCloudlet(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.CloudletKey)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return StreamCloudlet(c, obj)
}

func StreamCloudlet(c *cli.Command, in *edgeproto.CloudletKey) error {
	if StreamObjApiCmd == nil {
		return fmt.Errorf("StreamObjApi client not initialized")
	}
	ctx := context.Background()
	stream, err := StreamObjApiCmd.StreamCloudlet(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("StreamCloudlet failed: %s", errstr)
	}

	objs := make([]*edgeproto.Result, 0)
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
			return fmt.Errorf("StreamCloudlet recv failed: %s", errstr)
		}
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
func StreamCloudlets(c *cli.Command, data []edgeproto.CloudletKey, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("StreamCloudlet %v\n", data[ii])
		myerr := StreamCloudlet(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var StreamGPUDriverCmd = &cli.Command{
	Use:          "StreamGPUDriver",
	RequiredArgs: strings.Join(GPUDriverKeyRequiredArgs, " "),
	OptionalArgs: strings.Join(GPUDriverKeyOptionalArgs, " "),
	AliasArgs:    strings.Join(GPUDriverKeyAliasArgs, " "),
	SpecialArgs:  &GPUDriverKeySpecialArgs,
	Comments:     GPUDriverKeyComments,
	ReqData:      &edgeproto.GPUDriverKey{},
	ReplyData:    &edgeproto.Result{},
	Run:          runStreamGPUDriver,
}

func runStreamGPUDriver(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.GPUDriverKey)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return StreamGPUDriver(c, obj)
}

func StreamGPUDriver(c *cli.Command, in *edgeproto.GPUDriverKey) error {
	if StreamObjApiCmd == nil {
		return fmt.Errorf("StreamObjApi client not initialized")
	}
	ctx := context.Background()
	stream, err := StreamObjApiCmd.StreamGPUDriver(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("StreamGPUDriver failed: %s", errstr)
	}

	objs := make([]*edgeproto.Result, 0)
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
			return fmt.Errorf("StreamGPUDriver recv failed: %s", errstr)
		}
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
func StreamGPUDrivers(c *cli.Command, data []edgeproto.GPUDriverKey, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("StreamGPUDriver %v\n", data[ii])
		myerr := StreamGPUDriver(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var StreamLocalMsgsCmd = &cli.Command{
	Use:          "StreamLocalMsgs",
	RequiredArgs: strings.Join(AppInstKeyRequiredArgs, " "),
	OptionalArgs: strings.Join(AppInstKeyOptionalArgs, " "),
	AliasArgs:    strings.Join(AppInstKeyAliasArgs, " "),
	SpecialArgs:  &AppInstKeySpecialArgs,
	Comments:     AppInstKeyComments,
	ReqData:      &edgeproto.AppInstKey{},
	ReplyData:    &edgeproto.Result{},
	Run:          runStreamLocalMsgs,
}

func runStreamLocalMsgs(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.AppInstKey)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return StreamLocalMsgs(c, obj)
}

func StreamLocalMsgs(c *cli.Command, in *edgeproto.AppInstKey) error {
	if StreamObjApiCmd == nil {
		return fmt.Errorf("StreamObjApi client not initialized")
	}
	ctx := context.Background()
	stream, err := StreamObjApiCmd.StreamLocalMsgs(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("StreamLocalMsgs failed: %s", errstr)
	}

	objs := make([]*edgeproto.Result, 0)
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
			return fmt.Errorf("StreamLocalMsgs recv failed: %s", errstr)
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
func StreamLocalMsgss(c *cli.Command, data []edgeproto.AppInstKey, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("StreamLocalMsgs %v\n", data[ii])
		myerr := StreamLocalMsgs(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var StreamObjApiCmds = []*cobra.Command{
	StreamAppInstCmd.GenCmd(),
	StreamClusterInstCmd.GenCmd(),
	StreamCloudletCmd.GenCmd(),
	StreamGPUDriverCmd.GenCmd(),
	StreamLocalMsgsCmd.GenCmd(),
}

var StreamObjRequiredArgs = []string{
	"key.appkey.organization",
	"key.appkey.name",
	"key.appkey.version",
	"key.clusterinstkey.clusterkey.name",
	"key.clusterinstkey.cloudletkey.organization",
	"key.clusterinstkey.cloudletkey.name",
	"key.clusterinstkey.organization",
}
var StreamObjOptionalArgs = []string{
	"status.tasknumber",
	"status.maxtasks",
	"status.taskname",
	"status.stepname",
	"status.msgcount",
	"status.msgs",
}
var StreamObjAliasArgs = []string{}
var StreamObjComments = map[string]string{
	"key.appkey.organization":                     "App developer organization",
	"key.appkey.name":                             "App name",
	"key.appkey.version":                          "App version",
	"key.clusterinstkey.clusterkey.name":          "Cluster name",
	"key.clusterinstkey.cloudletkey.organization": "Organization of the cloudlet site",
	"key.clusterinstkey.cloudletkey.name":         "Name of the cloudlet",
	"key.clusterinstkey.organization":             "Name of Developer organization that this cluster belongs to",
}
var StreamObjSpecialArgs = map[string]string{
	"status.msgs": "StringArray",
}
