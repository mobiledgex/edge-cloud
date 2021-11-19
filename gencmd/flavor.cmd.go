// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: flavor.proto

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
func FlavorHideTags(in *edgeproto.Flavor) {
	if cli.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cli.HideTags, ",") {
		tags[tag] = struct{}{}
	}
	if _, found := tags["nocmp"]; found {
		in.DeletePrepare = false
	}
}

var FlavorApiCmd edgeproto.FlavorApiClient

var CreateFlavorCmd = &cli.Command{
	Use:          "CreateFlavor",
	RequiredArgs: strings.Join(CreateFlavorRequiredArgs, " "),
	OptionalArgs: strings.Join(CreateFlavorOptionalArgs, " "),
	AliasArgs:    strings.Join(FlavorAliasArgs, " "),
	SpecialArgs:  &FlavorSpecialArgs,
	Comments:     FlavorComments,
	ReqData:      &edgeproto.Flavor{},
	ReplyData:    &edgeproto.Result{},
	Run:          runCreateFlavor,
}

func runCreateFlavor(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.Flavor)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return CreateFlavor(c, obj)
}

func CreateFlavor(c *cli.Command, in *edgeproto.Flavor) error {
	if FlavorApiCmd == nil {
		return fmt.Errorf("FlavorApi client not initialized")
	}
	ctx := context.Background()
	obj, err := FlavorApiCmd.CreateFlavor(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateFlavor failed: %s", errstr)
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func CreateFlavors(c *cli.Command, data []edgeproto.Flavor, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("CreateFlavor %v\n", data[ii])
		myerr := CreateFlavor(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var DeleteFlavorCmd = &cli.Command{
	Use:          "DeleteFlavor",
	RequiredArgs: strings.Join(FlavorRequiredArgs, " "),
	OptionalArgs: strings.Join(FlavorOptionalArgs, " "),
	AliasArgs:    strings.Join(FlavorAliasArgs, " "),
	SpecialArgs:  &FlavorSpecialArgs,
	Comments:     FlavorComments,
	ReqData:      &edgeproto.Flavor{},
	ReplyData:    &edgeproto.Result{},
	Run:          runDeleteFlavor,
}

func runDeleteFlavor(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.Flavor)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return DeleteFlavor(c, obj)
}

func DeleteFlavor(c *cli.Command, in *edgeproto.Flavor) error {
	if FlavorApiCmd == nil {
		return fmt.Errorf("FlavorApi client not initialized")
	}
	ctx := context.Background()
	obj, err := FlavorApiCmd.DeleteFlavor(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("DeleteFlavor failed: %s", errstr)
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func DeleteFlavors(c *cli.Command, data []edgeproto.Flavor, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("DeleteFlavor %v\n", data[ii])
		myerr := DeleteFlavor(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var UpdateFlavorCmd = &cli.Command{
	Use:          "UpdateFlavor",
	RequiredArgs: strings.Join(FlavorRequiredArgs, " "),
	OptionalArgs: strings.Join(FlavorOptionalArgs, " "),
	AliasArgs:    strings.Join(FlavorAliasArgs, " "),
	SpecialArgs:  &FlavorSpecialArgs,
	Comments:     FlavorComments,
	ReqData:      &edgeproto.Flavor{},
	ReplyData:    &edgeproto.Result{},
	Run:          runUpdateFlavor,
}

func runUpdateFlavor(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.Flavor)
	jsonMap, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj.Fields = cli.GetSpecifiedFields(jsonMap, c.ReqData)
	return UpdateFlavor(c, obj)
}

func UpdateFlavor(c *cli.Command, in *edgeproto.Flavor) error {
	if FlavorApiCmd == nil {
		return fmt.Errorf("FlavorApi client not initialized")
	}
	ctx := context.Background()
	obj, err := FlavorApiCmd.UpdateFlavor(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("UpdateFlavor failed: %s", errstr)
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func UpdateFlavors(c *cli.Command, data []edgeproto.Flavor, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("UpdateFlavor %v\n", data[ii])
		myerr := UpdateFlavor(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ShowFlavorCmd = &cli.Command{
	Use:          "ShowFlavor",
	OptionalArgs: strings.Join(append(FlavorRequiredArgs, FlavorOptionalArgs...), " "),
	AliasArgs:    strings.Join(FlavorAliasArgs, " "),
	SpecialArgs:  &FlavorSpecialArgs,
	Comments:     FlavorComments,
	ReqData:      &edgeproto.Flavor{},
	ReplyData:    &edgeproto.Flavor{},
	Run:          runShowFlavor,
}

func runShowFlavor(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.Flavor)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowFlavor(c, obj)
}

func ShowFlavor(c *cli.Command, in *edgeproto.Flavor) error {
	if FlavorApiCmd == nil {
		return fmt.Errorf("FlavorApi client not initialized")
	}
	ctx := context.Background()
	stream, err := FlavorApiCmd.ShowFlavor(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowFlavor failed: %s", errstr)
	}

	objs := make([]*edgeproto.Flavor, 0)
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
			return fmt.Errorf("ShowFlavor recv failed: %s", errstr)
		}
		FlavorHideTags(obj)
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), objs, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func ShowFlavors(c *cli.Command, data []edgeproto.Flavor, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowFlavor %v\n", data[ii])
		myerr := ShowFlavor(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AddFlavorResCmd = &cli.Command{
	Use:          "AddFlavorRes",
	RequiredArgs: strings.Join(FlavorRequiredArgs, " "),
	OptionalArgs: strings.Join(FlavorOptionalArgs, " "),
	AliasArgs:    strings.Join(FlavorAliasArgs, " "),
	SpecialArgs:  &FlavorSpecialArgs,
	Comments:     FlavorComments,
	ReqData:      &edgeproto.Flavor{},
	ReplyData:    &edgeproto.Result{},
	Run:          runAddFlavorRes,
}

func runAddFlavorRes(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.Flavor)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return AddFlavorRes(c, obj)
}

func AddFlavorRes(c *cli.Command, in *edgeproto.Flavor) error {
	if FlavorApiCmd == nil {
		return fmt.Errorf("FlavorApi client not initialized")
	}
	ctx := context.Background()
	obj, err := FlavorApiCmd.AddFlavorRes(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("AddFlavorRes failed: %s", errstr)
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func AddFlavorRess(c *cli.Command, data []edgeproto.Flavor, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("AddFlavorRes %v\n", data[ii])
		myerr := AddFlavorRes(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var RemoveFlavorResCmd = &cli.Command{
	Use:          "RemoveFlavorRes",
	RequiredArgs: strings.Join(FlavorRequiredArgs, " "),
	OptionalArgs: strings.Join(FlavorOptionalArgs, " "),
	AliasArgs:    strings.Join(FlavorAliasArgs, " "),
	SpecialArgs:  &FlavorSpecialArgs,
	Comments:     FlavorComments,
	ReqData:      &edgeproto.Flavor{},
	ReplyData:    &edgeproto.Result{},
	Run:          runRemoveFlavorRes,
}

func runRemoveFlavorRes(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.Flavor)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return RemoveFlavorRes(c, obj)
}

func RemoveFlavorRes(c *cli.Command, in *edgeproto.Flavor) error {
	if FlavorApiCmd == nil {
		return fmt.Errorf("FlavorApi client not initialized")
	}
	ctx := context.Background()
	obj, err := FlavorApiCmd.RemoveFlavorRes(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("RemoveFlavorRes failed: %s", errstr)
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func RemoveFlavorRess(c *cli.Command, data []edgeproto.Flavor, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("RemoveFlavorRes %v\n", data[ii])
		myerr := RemoveFlavorRes(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var FlavorApiCmds = []*cobra.Command{
	CreateFlavorCmd.GenCmd(),
	DeleteFlavorCmd.GenCmd(),
	UpdateFlavorCmd.GenCmd(),
	ShowFlavorCmd.GenCmd(),
	AddFlavorResCmd.GenCmd(),
	RemoveFlavorResCmd.GenCmd(),
}

var FlavorKeyRequiredArgs = []string{}
var FlavorKeyOptionalArgs = []string{
	"name",
}
var FlavorKeyAliasArgs = []string{}
var FlavorKeyComments = map[string]string{
	"name": "Flavor name",
}
var FlavorKeySpecialArgs = map[string]string{}
var FlavorRequiredArgs = []string{
	"name",
}
var FlavorOptionalArgs = []string{
	"ram",
	"vcpus",
	"disk",
	"optresmap",
}
var FlavorAliasArgs = []string{
	"name=key.name",
}
var FlavorComments = map[string]string{
	"fields":        "Fields are used for the Update API to specify which fields to apply",
	"name":          "Flavor name",
	"ram":           "RAM in megabytes",
	"vcpus":         "Number of virtual CPUs",
	"disk":          "Amount of disk space in gigabytes",
	"optresmap":     "Optional Resources request, key = gpu form: $resource=$kind:[$alias]$count ex: optresmap=gpu=vgpu:nvidia-63:1, specify optresmap:empty=true to clear",
	"deleteprepare": "Preparing to be deleted",
}
var FlavorSpecialArgs = map[string]string{
	"fields":    "StringArray",
	"optresmap": "StringToString",
}
var CreateFlavorRequiredArgs = []string{
	"name",
	"ram",
	"vcpus",
	"disk",
}
var CreateFlavorOptionalArgs = []string{
	"optresmap",
}
