// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: clusterinst.proto

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
func ClusterInstHideTags(in *edgeproto.ClusterInst) {
	if cli.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cli.HideTags, ",") {
		tags[tag] = struct{}{}
	}
	if _, found := tags["nocmp"]; found {
		in.State = 0
	}
	if _, found := tags["nocmp"]; found {
		in.Errors = nil
	}
	if _, found := tags["nocmp"]; found {
		in.CrmOverride = 0
	}
	if _, found := tags["nocmp"]; found {
		in.AllocatedIp = ""
	}
	if _, found := tags["nocmp"]; found {
		in.NodeFlavor = ""
	}
	if _, found := tags["nocmp"]; found {
		in.ExternalVolumeSize = 0
	}
}

func ClusterInstInfoHideTags(in *edgeproto.ClusterInstInfo) {
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

var ClusterInstApiCmd edgeproto.ClusterInstApiClient

var CreateClusterInstCmd = &cli.Command{
	Use:          "CreateClusterInst",
	RequiredArgs: strings.Join(ClusterInstRequiredArgs, " "),
	OptionalArgs: strings.Join(ClusterInstOptionalArgs, " "),
	AliasArgs:    strings.Join(ClusterInstAliasArgs, " "),
	SpecialArgs:  &ClusterInstSpecialArgs,
	Comments:     ClusterInstComments,
	ReqData:      &edgeproto.ClusterInst{},
	ReplyData:    &edgeproto.Result{},
	Run:          runCreateClusterInst,
}

func runCreateClusterInst(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.ClusterInst)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return CreateClusterInst(c, obj)
}

func CreateClusterInst(c *cli.Command, in *edgeproto.ClusterInst) error {
	if ClusterInstApiCmd == nil {
		return fmt.Errorf("ClusterInstApi client not initialized")
	}
	ctx := context.Background()
	stream, err := ClusterInstApiCmd.CreateClusterInst(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateClusterInst failed: %s", errstr)
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
			return fmt.Errorf("CreateClusterInst recv failed: %s", errstr)
		}
		c.WriteOutput(obj, cli.OutputFormat)
	}
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func CreateClusterInsts(c *cli.Command, data []edgeproto.ClusterInst, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("CreateClusterInst %v\n", data[ii])
		myerr := CreateClusterInst(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var DeleteClusterInstCmd = &cli.Command{
	Use:          "DeleteClusterInst",
	RequiredArgs: strings.Join(ClusterInstRequiredArgs, " "),
	OptionalArgs: strings.Join(ClusterInstOptionalArgs, " "),
	AliasArgs:    strings.Join(ClusterInstAliasArgs, " "),
	SpecialArgs:  &ClusterInstSpecialArgs,
	Comments:     ClusterInstComments,
	ReqData:      &edgeproto.ClusterInst{},
	ReplyData:    &edgeproto.Result{},
	Run:          runDeleteClusterInst,
}

func runDeleteClusterInst(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.ClusterInst)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return DeleteClusterInst(c, obj)
}

func DeleteClusterInst(c *cli.Command, in *edgeproto.ClusterInst) error {
	if ClusterInstApiCmd == nil {
		return fmt.Errorf("ClusterInstApi client not initialized")
	}
	ctx := context.Background()
	stream, err := ClusterInstApiCmd.DeleteClusterInst(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("DeleteClusterInst failed: %s", errstr)
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
			return fmt.Errorf("DeleteClusterInst recv failed: %s", errstr)
		}
		c.WriteOutput(obj, cli.OutputFormat)
	}
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func DeleteClusterInsts(c *cli.Command, data []edgeproto.ClusterInst, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("DeleteClusterInst %v\n", data[ii])
		myerr := DeleteClusterInst(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var UpdateClusterInstCmd = &cli.Command{
	Use:          "UpdateClusterInst",
	RequiredArgs: strings.Join(ClusterInstRequiredArgs, " "),
	OptionalArgs: strings.Join(ClusterInstOptionalArgs, " "),
	AliasArgs:    strings.Join(ClusterInstAliasArgs, " "),
	SpecialArgs:  &ClusterInstSpecialArgs,
	Comments:     ClusterInstComments,
	ReqData:      &edgeproto.ClusterInst{},
	ReplyData:    &edgeproto.Result{},
	Run:          runUpdateClusterInst,
}

func runUpdateClusterInst(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.ClusterInst)
	jsonMap, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj.Fields = cli.GetSpecifiedFields(jsonMap, c.ReqData, cli.JsonNamespace)
	return UpdateClusterInst(c, obj)
}

func UpdateClusterInst(c *cli.Command, in *edgeproto.ClusterInst) error {
	if ClusterInstApiCmd == nil {
		return fmt.Errorf("ClusterInstApi client not initialized")
	}
	ctx := context.Background()
	stream, err := ClusterInstApiCmd.UpdateClusterInst(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("UpdateClusterInst failed: %s", errstr)
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
			return fmt.Errorf("UpdateClusterInst recv failed: %s", errstr)
		}
		c.WriteOutput(obj, cli.OutputFormat)
	}
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func UpdateClusterInsts(c *cli.Command, data []edgeproto.ClusterInst, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("UpdateClusterInst %v\n", data[ii])
		myerr := UpdateClusterInst(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ShowClusterInstCmd = &cli.Command{
	Use:          "ShowClusterInst",
	OptionalArgs: strings.Join(append(ClusterInstRequiredArgs, ClusterInstOptionalArgs...), " "),
	AliasArgs:    strings.Join(ClusterInstAliasArgs, " "),
	SpecialArgs:  &ClusterInstSpecialArgs,
	Comments:     ClusterInstComments,
	ReqData:      &edgeproto.ClusterInst{},
	ReplyData:    &edgeproto.ClusterInst{},
	Run:          runShowClusterInst,
}

func runShowClusterInst(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.ClusterInst)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowClusterInst(c, obj)
}

func ShowClusterInst(c *cli.Command, in *edgeproto.ClusterInst) error {
	if ClusterInstApiCmd == nil {
		return fmt.Errorf("ClusterInstApi client not initialized")
	}
	ctx := context.Background()
	stream, err := ClusterInstApiCmd.ShowClusterInst(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowClusterInst failed: %s", errstr)
	}
	objs := make([]*edgeproto.ClusterInst, 0)
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
			return fmt.Errorf("ShowClusterInst recv failed: %s", errstr)
		}
		ClusterInstHideTags(obj)
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(objs, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func ShowClusterInsts(c *cli.Command, data []edgeproto.ClusterInst, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowClusterInst %v\n", data[ii])
		myerr := ShowClusterInst(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ClusterInstApiCmds = []*cobra.Command{
	CreateClusterInstCmd.GenCmd(),
	DeleteClusterInstCmd.GenCmd(),
	UpdateClusterInstCmd.GenCmd(),
	ShowClusterInstCmd.GenCmd(),
}

var ClusterInstInfoApiCmd edgeproto.ClusterInstInfoApiClient

var ShowClusterInstInfoCmd = &cli.Command{
	Use:          "ShowClusterInstInfo",
	OptionalArgs: strings.Join(append(ClusterInstInfoRequiredArgs, ClusterInstInfoOptionalArgs...), " "),
	AliasArgs:    strings.Join(ClusterInstInfoAliasArgs, " "),
	SpecialArgs:  &ClusterInstInfoSpecialArgs,
	Comments:     ClusterInstInfoComments,
	ReqData:      &edgeproto.ClusterInstInfo{},
	ReplyData:    &edgeproto.ClusterInstInfo{},
	Run:          runShowClusterInstInfo,
}

func runShowClusterInstInfo(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.ClusterInstInfo)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowClusterInstInfo(c, obj)
}

func ShowClusterInstInfo(c *cli.Command, in *edgeproto.ClusterInstInfo) error {
	if ClusterInstInfoApiCmd == nil {
		return fmt.Errorf("ClusterInstInfoApi client not initialized")
	}
	ctx := context.Background()
	stream, err := ClusterInstInfoApiCmd.ShowClusterInstInfo(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowClusterInstInfo failed: %s", errstr)
	}
	objs := make([]*edgeproto.ClusterInstInfo, 0)
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
			return fmt.Errorf("ShowClusterInstInfo recv failed: %s", errstr)
		}
		ClusterInstInfoHideTags(obj)
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(objs, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func ShowClusterInstInfos(c *cli.Command, data []edgeproto.ClusterInstInfo, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowClusterInstInfo %v\n", data[ii])
		myerr := ShowClusterInstInfo(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ClusterInstInfoApiCmds = []*cobra.Command{
	ShowClusterInstInfoCmd.GenCmd(),
}

var ClusterInstKeyRequiredArgs = []string{}
var ClusterInstKeyOptionalArgs = []string{
	"clusterkey.name",
	"cloudletkey.operatorkey.name",
	"cloudletkey.name",
	"developer",
}
var ClusterInstKeyAliasArgs = []string{}
var ClusterInstKeyComments = map[string]string{
	"clusterkey.name":              "Cluster name",
	"cloudletkey.operatorkey.name": "Company or Organization name of the operator",
	"cloudletkey.name":             "Name of the cloudlet",
	"developer":                    "Name of Developer that this cluster belongs to",
}
var ClusterInstKeySpecialArgs = map[string]string{}
var ClusterInstRequiredArgs = []string{
	"cluster",
	"operator",
	"cloudlet",
	"developer",
}
var ClusterInstOptionalArgs = []string{
	"flavor",
	"state",
	"errors",
	"crmoverride",
	"ipaccess",
	"deployment",
	"nummasters",
	"numnodes",
	"autoscalepolicy",
	"availabilityzone",
	"imagename",
	"reservable",
}
var ClusterInstAliasArgs = []string{
	"cluster=key.clusterkey.name",
	"operator=key.cloudletkey.operatorkey.name",
	"cloudlet=key.cloudletkey.name",
	"developer=key.developer",
	"flavor=flavor.name",
}
var ClusterInstComments = map[string]string{
	"cluster":            "Cluster name",
	"operator":           "Company or Organization name of the operator",
	"cloudlet":           "Name of the cloudlet",
	"developer":          "Name of Developer that this cluster belongs to",
	"flavor":             "Flavor name",
	"liveness":           "Liveness of instance (see Liveness), one of LivenessUnknown, LivenessStatic, LivenessDynamic",
	"auto":               "Auto is set to true when automatically created by back-end (internal use only)",
	"state":              "State of the cluster instance, one of TrackedStateUnknown, NotPresent, CreateRequested, Creating, CreateError, Ready, UpdateRequested, Updating, UpdateError, DeleteRequested, Deleting, DeleteError, DeletePrepare, CrmInitok, HealthcheckFailed",
	"errors":             "Any errors trying to create, update, or delete the ClusterInst on the Cloudlet.",
	"crmoverride":        "Override actions to CRM, one of NoOverride, IgnoreCrmErrors, IgnoreCrm, IgnoreTransientState, IgnoreCrmAndTransientState",
	"ipaccess":           "IP access type (RootLB Type), one of IpAccessUnknown, IpAccessDedicated, IpAccessDedicatedOrShared, IpAccessShared",
	"allocatedip":        "Allocated IP for dedicated access",
	"nodeflavor":         "Cloudlet specific node flavor",
	"deployment":         "Deployment type (kubernetes or docker)",
	"nummasters":         "Number of k8s masters (In case of docker deployment, this field is not required)",
	"numnodes":           "Number of k8s nodes (In case of docker deployment, this field is not required)",
	"externalvolumesize": "Size of external volume to be attached to nodes",
	"autoscalepolicy":    "Auto scale policy name",
	"availabilityzone":   "Optional Resource AZ if any",
	"imagename":          "Optional resource specific image to launch",
	"reservable":         "If ClusterInst is reservable",
	"reservedby":         "For reservable MobiledgeX ClusterInsts, the current developer tenant",
}
var ClusterInstSpecialArgs = map[string]string{
	"errors": "StringArray",
}
var ClusterInstInfoRequiredArgs = []string{
	"key.clusterkey.name",
	"key.cloudletkey.operatorkey.name",
	"key.cloudletkey.name",
	"key.developer",
}
var ClusterInstInfoOptionalArgs = []string{
	"notifyid",
	"state",
	"errors",
	"status.tasknumber",
	"status.maxtasks",
	"status.taskname",
	"status.stepname",
}
var ClusterInstInfoAliasArgs = []string{}
var ClusterInstInfoComments = map[string]string{
	"key.clusterkey.name":              "Cluster name",
	"key.cloudletkey.operatorkey.name": "Company or Organization name of the operator",
	"key.cloudletkey.name":             "Name of the cloudlet",
	"key.developer":                    "Name of Developer that this cluster belongs to",
	"notifyid":                         "Id of client assigned by server (internal use only)",
	"state":                            "State of the cluster instance, one of TrackedStateUnknown, NotPresent, CreateRequested, Creating, CreateError, Ready, UpdateRequested, Updating, UpdateError, DeleteRequested, Deleting, DeleteError, DeletePrepare, CrmInitok, HealthcheckFailed",
	"errors":                           "Any errors trying to create, update, or delete the ClusterInst on the Cloudlet.",
}
var ClusterInstInfoSpecialArgs = map[string]string{
	"errors": "StringArray",
}
