// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: developer.proto

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
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
var DeveloperApiCmd edgeproto.DeveloperApiClient

var CreateDeveloperCmd = &cli.Command{
	Use:          "CreateDeveloper",
	RequiredArgs: strings.Join(DeveloperRequiredArgs, " "),
	OptionalArgs: strings.Join(DeveloperOptionalArgs, " "),
	AliasArgs:    strings.Join(DeveloperAliasArgs, " "),
	SpecialArgs:  &DeveloperSpecialArgs,
	Comments:     DeveloperComments,
	ReqData:      &edgeproto.Developer{},
	ReplyData:    &edgeproto.Result{},
	Run:          runCreateDeveloper,
}

func runCreateDeveloper(c *cli.Command, args []string) error {
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj := c.ReqData.(*edgeproto.Developer)
	return CreateDeveloper(c, obj)
}

func CreateDeveloper(c *cli.Command, in *edgeproto.Developer) error {
	if DeveloperApiCmd == nil {
		return fmt.Errorf("DeveloperApi client not initialized")
	}
	ctx := context.Background()
	obj, err := DeveloperApiCmd.CreateDeveloper(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateDeveloper failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func CreateDevelopers(c *cli.Command, data []edgeproto.Developer, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("CreateDeveloper %v\n", data[ii])
		myerr := CreateDeveloper(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var DeleteDeveloperCmd = &cli.Command{
	Use:          "DeleteDeveloper",
	RequiredArgs: strings.Join(DeveloperRequiredArgs, " "),
	OptionalArgs: strings.Join(DeveloperOptionalArgs, " "),
	AliasArgs:    strings.Join(DeveloperAliasArgs, " "),
	SpecialArgs:  &DeveloperSpecialArgs,
	Comments:     DeveloperComments,
	ReqData:      &edgeproto.Developer{},
	ReplyData:    &edgeproto.Result{},
	Run:          runDeleteDeveloper,
}

func runDeleteDeveloper(c *cli.Command, args []string) error {
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj := c.ReqData.(*edgeproto.Developer)
	return DeleteDeveloper(c, obj)
}

func DeleteDeveloper(c *cli.Command, in *edgeproto.Developer) error {
	if DeveloperApiCmd == nil {
		return fmt.Errorf("DeveloperApi client not initialized")
	}
	ctx := context.Background()
	obj, err := DeveloperApiCmd.DeleteDeveloper(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("DeleteDeveloper failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func DeleteDevelopers(c *cli.Command, data []edgeproto.Developer, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("DeleteDeveloper %v\n", data[ii])
		myerr := DeleteDeveloper(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var UpdateDeveloperCmd = &cli.Command{
	Use:          "UpdateDeveloper",
	RequiredArgs: strings.Join(DeveloperRequiredArgs, " "),
	OptionalArgs: strings.Join(DeveloperOptionalArgs, " "),
	AliasArgs:    strings.Join(DeveloperAliasArgs, " "),
	SpecialArgs:  &DeveloperSpecialArgs,
	Comments:     DeveloperComments,
	ReqData:      &edgeproto.Developer{},
	ReplyData:    &edgeproto.Result{},
	Run:          runUpdateDeveloper,
}

func runUpdateDeveloper(c *cli.Command, args []string) error {
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj := c.ReqData.(*edgeproto.Developer)
	return UpdateDeveloper(c, obj)
}

func UpdateDeveloper(c *cli.Command, in *edgeproto.Developer) error {
	if DeveloperApiCmd == nil {
		return fmt.Errorf("DeveloperApi client not initialized")
	}
	ctx := context.Background()
	obj, err := DeveloperApiCmd.UpdateDeveloper(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("UpdateDeveloper failed: %s", errstr)
	}
	c.WriteOutput(obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func UpdateDevelopers(c *cli.Command, data []edgeproto.Developer, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("UpdateDeveloper %v\n", data[ii])
		myerr := UpdateDeveloper(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ShowDeveloperCmd = &cli.Command{
	Use:          "ShowDeveloper",
	OptionalArgs: strings.Join(append(DeveloperRequiredArgs, DeveloperOptionalArgs...), " "),
	AliasArgs:    strings.Join(DeveloperAliasArgs, " "),
	SpecialArgs:  &DeveloperSpecialArgs,
	Comments:     DeveloperComments,
	ReqData:      &edgeproto.Developer{},
	ReplyData:    &edgeproto.Developer{},
	Run:          runShowDeveloper,
}

func runShowDeveloper(c *cli.Command, args []string) error {
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj := c.ReqData.(*edgeproto.Developer)
	return ShowDeveloper(c, obj)
}

func ShowDeveloper(c *cli.Command, in *edgeproto.Developer) error {
	if DeveloperApiCmd == nil {
		return fmt.Errorf("DeveloperApi client not initialized")
	}
	ctx := context.Background()
	stream, err := DeveloperApiCmd.ShowDeveloper(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowDeveloper failed: %s", errstr)
	}
	objs := make([]*edgeproto.Developer, 0)
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ShowDeveloper recv failed: %s", err.Error())
		}
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(objs, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func ShowDevelopers(c *cli.Command, data []edgeproto.Developer, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowDeveloper %v\n", data[ii])
		myerr := ShowDeveloper(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var DeveloperApiCmds = []*cobra.Command{
	CreateDeveloperCmd.GenCmd(),
	DeleteDeveloperCmd.GenCmd(),
	UpdateDeveloperCmd.GenCmd(),
	ShowDeveloperCmd.GenCmd(),
}

var DeveloperKeyRequiredArgs = []string{}
var DeveloperKeyOptionalArgs = []string{
	"name",
}
var DeveloperKeyAliasArgs = []string{}
var DeveloperKeyComments = map[string]string{
	"name": "Organization or Company Name that a Developer is part of",
}
var DeveloperKeySpecialArgs = map[string]string{}
var DeveloperRequiredArgs = []string{
	"name",
}
var DeveloperOptionalArgs = []string{}
var DeveloperAliasArgs = []string{
	"name=key.name",
}
var DeveloperComments = map[string]string{
	"name": "Organization or Company Name that a Developer is part of",
}
var DeveloperSpecialArgs = map[string]string{}
