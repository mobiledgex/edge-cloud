// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: org.proto

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
	math "math"
	"strings"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
var OrganizationApiCmd edgeproto.OrganizationApiClient

var OrganizationInUseCmd = &cli.Command{
	Use:          "OrganizationInUse",
	RequiredArgs: strings.Join(OrganizationRequiredArgs, " "),
	OptionalArgs: strings.Join(OrganizationOptionalArgs, " "),
	AliasArgs:    strings.Join(OrganizationAliasArgs, " "),
	SpecialArgs:  &OrganizationSpecialArgs,
	Comments:     OrganizationComments,
	ReqData:      &edgeproto.Organization{},
	ReplyData:    &edgeproto.Result{},
	Run:          runOrganizationInUse,
}

func runOrganizationInUse(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.Organization)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return OrganizationInUse(c, obj)
}

func OrganizationInUse(c *cli.Command, in *edgeproto.Organization) error {
	if OrganizationApiCmd == nil {
		return fmt.Errorf("OrganizationApi client not initialized")
	}
	ctx := context.Background()
	obj, err := OrganizationApiCmd.OrganizationInUse(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("OrganizationInUse failed: %s", errstr)
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func OrganizationInUses(c *cli.Command, data []edgeproto.Organization, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("OrganizationInUse %v\n", data[ii])
		myerr := OrganizationInUse(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var OrganizationApiCmds = []*cobra.Command{
	OrganizationInUseCmd.GenCmd(),
}

var OrganizationRequiredArgs = []string{}
var OrganizationOptionalArgs = []string{
	"name",
}
var OrganizationAliasArgs = []string{}
var OrganizationComments = map[string]string{
	"name": "Organization name",
}
var OrganizationSpecialArgs = map[string]string{}
var OrganizationDataRequiredArgs = []string{}
var OrganizationDataOptionalArgs = []string{
	"orgs:#.name",
}
var OrganizationDataAliasArgs = []string{}
var OrganizationDataComments = map[string]string{
	"orgs:#.name": "Organization name",
}
var OrganizationDataSpecialArgs = map[string]string{}
