// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app_inst.proto

package gencmd

import distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
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
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
func AppInstHideTags(in *edgeproto.AppInst) {
	if cli.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cli.HideTags, ",") {
		tags[tag] = struct{}{}
	}
	if _, found := tags["nocmp"]; found {
		in.Uri = ""
	}
	for i0 := 0; i0 < len(in.MappedPorts); i0++ {
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
		in.RuntimeInfo.ContainerIds = nil
	}
	if _, found := tags["timestamp"]; found {
		in.CreatedAt = distributed_match_engine.Timestamp{}
	}
	if _, found := tags["nocmp"]; found {
		in.ForceUpdate = false
	}
	if _, found := tags["nocmp"]; found {
		in.UpdateMultiple = false
	}
	for i0 := 0; i0 < len(in.Configs); i0++ {
	}
}

func AppInstRuntimeHideTags(in *edgeproto.AppInstRuntime) {
	if cli.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cli.HideTags, ",") {
		tags[tag] = struct{}{}
	}
	if _, found := tags["nocmp"]; found {
		in.ContainerIds = nil
	}
}

func AppInstInfoHideTags(in *edgeproto.AppInstInfo) {
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
		in.RuntimeInfo.ContainerIds = nil
	}
}

var AppInstApiCmd edgeproto.AppInstApiClient

var CreateAppInstCmd = &cli.Command{
	Use:          "CreateAppInst",
	RequiredArgs: strings.Join(CreateAppInstRequiredArgs, " "),
	OptionalArgs: strings.Join(CreateAppInstOptionalArgs, " "),
	AliasArgs:    strings.Join(AppInstAliasArgs, " "),
	SpecialArgs:  &AppInstSpecialArgs,
	Comments:     AppInstComments,
	ReqData:      &edgeproto.AppInst{},
	ReplyData:    &edgeproto.Result{},
	Run:          runCreateAppInst,
}

func runCreateAppInst(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.AppInst)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return CreateAppInst(c, obj)
}

func CreateAppInst(c *cli.Command, in *edgeproto.AppInst) error {
	if AppInstApiCmd == nil {
		return fmt.Errorf("AppInstApi client not initialized")
	}
	ctx := context.Background()
	stream, err := AppInstApiCmd.CreateAppInst(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateAppInst failed: %s", errstr)
	}
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("CreateAppInst recv failed: %s", err.Error())
		}
		c.WriteOutput(obj, cli.OutputFormat)
	}
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func CreateAppInsts(c *cli.Command, data []edgeproto.AppInst, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("CreateAppInst %v\n", data[ii])
		myerr := CreateAppInst(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var DeleteAppInstCmd = &cli.Command{
	Use:          "DeleteAppInst",
	RequiredArgs: strings.Join(AppInstRequiredArgs, " "),
	OptionalArgs: strings.Join(AppInstOptionalArgs, " "),
	AliasArgs:    strings.Join(AppInstAliasArgs, " "),
	SpecialArgs:  &AppInstSpecialArgs,
	Comments:     AppInstComments,
	ReqData:      &edgeproto.AppInst{},
	ReplyData:    &edgeproto.Result{},
	Run:          runDeleteAppInst,
}

func runDeleteAppInst(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.AppInst)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return DeleteAppInst(c, obj)
}

func DeleteAppInst(c *cli.Command, in *edgeproto.AppInst) error {
	if AppInstApiCmd == nil {
		return fmt.Errorf("AppInstApi client not initialized")
	}
	ctx := context.Background()
	stream, err := AppInstApiCmd.DeleteAppInst(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("DeleteAppInst failed: %s", errstr)
	}
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("DeleteAppInst recv failed: %s", err.Error())
		}
		c.WriteOutput(obj, cli.OutputFormat)
	}
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func DeleteAppInsts(c *cli.Command, data []edgeproto.AppInst, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("DeleteAppInst %v\n", data[ii])
		myerr := DeleteAppInst(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var UpdateAppInstCmd = &cli.Command{
	Use:          "UpdateAppInst",
	RequiredArgs: strings.Join(UpdateAppInstRequiredArgs, " "),
	OptionalArgs: strings.Join(UpdateAppInstOptionalArgs, " "),
	AliasArgs:    strings.Join(AppInstAliasArgs, " "),
	SpecialArgs:  &AppInstSpecialArgs,
	Comments:     AppInstComments,
	ReqData:      &edgeproto.AppInst{},
	ReplyData:    &edgeproto.Result{},
	Run:          runUpdateAppInst,
}

func runUpdateAppInst(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.AppInst)
	jsonMap, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj.Fields = cli.GetSpecifiedFields(jsonMap, c.ReqData, cli.JsonNamespace)
	return UpdateAppInst(c, obj)
}

func UpdateAppInst(c *cli.Command, in *edgeproto.AppInst) error {
	if AppInstApiCmd == nil {
		return fmt.Errorf("AppInstApi client not initialized")
	}
	ctx := context.Background()
	stream, err := AppInstApiCmd.UpdateAppInst(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("UpdateAppInst failed: %s", errstr)
	}
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("UpdateAppInst recv failed: %s", err.Error())
		}
		c.WriteOutput(obj, cli.OutputFormat)
	}
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func UpdateAppInsts(c *cli.Command, data []edgeproto.AppInst, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("UpdateAppInst %v\n", data[ii])
		myerr := UpdateAppInst(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ShowAppInstCmd = &cli.Command{
	Use:          "ShowAppInst",
	OptionalArgs: strings.Join(append(AppInstRequiredArgs, AppInstOptionalArgs...), " "),
	AliasArgs:    strings.Join(AppInstAliasArgs, " "),
	SpecialArgs:  &AppInstSpecialArgs,
	Comments:     AppInstComments,
	ReqData:      &edgeproto.AppInst{},
	ReplyData:    &edgeproto.AppInst{},
	Run:          runShowAppInst,
}

func runShowAppInst(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.AppInst)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowAppInst(c, obj)
}

func ShowAppInst(c *cli.Command, in *edgeproto.AppInst) error {
	if AppInstApiCmd == nil {
		return fmt.Errorf("AppInstApi client not initialized")
	}
	ctx := context.Background()
	stream, err := AppInstApiCmd.ShowAppInst(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowAppInst failed: %s", errstr)
	}
	objs := make([]*edgeproto.AppInst, 0)
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ShowAppInst recv failed: %s", err.Error())
		}
		AppInstHideTags(obj)
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(objs, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func ShowAppInsts(c *cli.Command, data []edgeproto.AppInst, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowAppInst %v\n", data[ii])
		myerr := ShowAppInst(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AppInstApiCmds = []*cobra.Command{
	CreateAppInstCmd.GenCmd(),
	DeleteAppInstCmd.GenCmd(),
	UpdateAppInstCmd.GenCmd(),
	ShowAppInstCmd.GenCmd(),
}

var AppInstInfoApiCmd edgeproto.AppInstInfoApiClient

var ShowAppInstInfoCmd = &cli.Command{
	Use:          "ShowAppInstInfo",
	OptionalArgs: strings.Join(append(AppInstInfoRequiredArgs, AppInstInfoOptionalArgs...), " "),
	AliasArgs:    strings.Join(AppInstInfoAliasArgs, " "),
	SpecialArgs:  &AppInstInfoSpecialArgs,
	Comments:     AppInstInfoComments,
	ReqData:      &edgeproto.AppInstInfo{},
	ReplyData:    &edgeproto.AppInstInfo{},
	Run:          runShowAppInstInfo,
}

func runShowAppInstInfo(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.AppInstInfo)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowAppInstInfo(c, obj)
}

func ShowAppInstInfo(c *cli.Command, in *edgeproto.AppInstInfo) error {
	if AppInstInfoApiCmd == nil {
		return fmt.Errorf("AppInstInfoApi client not initialized")
	}
	ctx := context.Background()
	stream, err := AppInstInfoApiCmd.ShowAppInstInfo(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowAppInstInfo failed: %s", errstr)
	}
	objs := make([]*edgeproto.AppInstInfo, 0)
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ShowAppInstInfo recv failed: %s", err.Error())
		}
		AppInstInfoHideTags(obj)
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(objs, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func ShowAppInstInfos(c *cli.Command, data []edgeproto.AppInstInfo, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowAppInstInfo %v\n", data[ii])
		myerr := ShowAppInstInfo(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AppInstInfoApiCmds = []*cobra.Command{
	ShowAppInstInfoCmd.GenCmd(),
}

var AppInstMetricsApiCmd edgeproto.AppInstMetricsApiClient

var ShowAppInstMetricsCmd = &cli.Command{
	Use:          "ShowAppInstMetrics",
	OptionalArgs: strings.Join(append(AppInstMetricsRequiredArgs, AppInstMetricsOptionalArgs...), " "),
	AliasArgs:    strings.Join(AppInstMetricsAliasArgs, " "),
	SpecialArgs:  &AppInstMetricsSpecialArgs,
	Comments:     AppInstMetricsComments,
	ReqData:      &edgeproto.AppInstMetrics{},
	ReplyData:    &edgeproto.AppInstMetrics{},
	Run:          runShowAppInstMetrics,
}

func runShowAppInstMetrics(c *cli.Command, args []string) error {
	obj := c.ReqData.(*edgeproto.AppInstMetrics)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return ShowAppInstMetrics(c, obj)
}

func ShowAppInstMetrics(c *cli.Command, in *edgeproto.AppInstMetrics) error {
	if AppInstMetricsApiCmd == nil {
		return fmt.Errorf("AppInstMetricsApi client not initialized")
	}
	ctx := context.Background()
	stream, err := AppInstMetricsApiCmd.ShowAppInstMetrics(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("ShowAppInstMetrics failed: %s", errstr)
	}
	objs := make([]*edgeproto.AppInstMetrics, 0)
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ShowAppInstMetrics recv failed: %s", err.Error())
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
func ShowAppInstMetricss(c *cli.Command, data []edgeproto.AppInstMetrics, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowAppInstMetrics %v\n", data[ii])
		myerr := ShowAppInstMetrics(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AppInstMetricsApiCmds = []*cobra.Command{
	ShowAppInstMetricsCmd.GenCmd(),
}

var AppInstKeyRequiredArgs = []string{}
var AppInstKeyOptionalArgs = []string{
	"appkey.developerkey.name",
	"appkey.name",
	"appkey.version",
	"clusterinstkey.clusterkey.name",
	"clusterinstkey.cloudletkey.operatorkey.name",
	"clusterinstkey.cloudletkey.name",
	"clusterinstkey.developer",
}
var AppInstKeyAliasArgs = []string{}
var AppInstKeyComments = map[string]string{
	"appkey.developerkey.name":                    "Organization or Company Name that a Developer is part of",
	"appkey.name":                                 "App name",
	"appkey.version":                              "App version",
	"clusterinstkey.clusterkey.name":              "Cluster name",
	"clusterinstkey.cloudletkey.operatorkey.name": "Company or Organization name of the operator",
	"clusterinstkey.cloudletkey.name":             "Name of the cloudlet",
	"clusterinstkey.developer":                    "Name of Developer that this cluster belongs to",
}
var AppInstKeySpecialArgs = map[string]string{}
var AppInstRequiredArgs = []string{
	"developer",
	"appname",
	"appvers",
	"cluster",
	"operator",
	"cloudlet",
}
var AppInstOptionalArgs = []string{
	"clusterdeveloper",
	"flavor",
	"state",
	"crmoverride",
	"autoclusteripaccess",
	"forceupdate",
	"updatemultiple",
	"configs.kind",
	"configs.config",
}
var AppInstAliasArgs = []string{
	"developer=key.appkey.developerkey.name",
	"appname=key.appkey.name",
	"appvers=key.appkey.version",
	"cluster=key.clusterinstkey.clusterkey.name",
	"operator=key.clusterinstkey.cloudletkey.operatorkey.name",
	"cloudlet=key.clusterinstkey.cloudletkey.name",
	"clusterdeveloper=key.clusterinstkey.developer",
	"flavor=flavor.name",
}
var AppInstComments = map[string]string{
	"developer":                      "Organization or Company Name that a Developer is part of",
	"appname":                        "App name",
	"appvers":                        "App version",
	"cluster":                        "Cluster name",
	"operator":                       "Company or Organization name of the operator",
	"cloudlet":                       "Name of the cloudlet",
	"clusterdeveloper":               "Name of Developer that this cluster belongs to",
	"cloudletloc.latitude":           "latitude in WGS 84 coordinates",
	"cloudletloc.longitude":          "longitude in WGS 84 coordinates",
	"cloudletloc.horizontalaccuracy": "horizontal accuracy (radius in meters)",
	"cloudletloc.verticalaccuracy":   "vertical accuracy (meters)",
	"cloudletloc.altitude":           "On android only lat and long are guaranteed to be supplied altitude in meters",
	"cloudletloc.course":             "course (IOS) / bearing (Android) (degrees east relative to true north)",
	"cloudletloc.speed":              "speed (IOS) / velocity (Android) (meters/sec)",
	"uri":                            "Base FQDN (not really URI) for the App. See Service FQDN for endpoint access.",
	"liveness":                       "Liveness of instance (see Liveness), one of LivenessUnknown, LivenessStatic, LivenessDynamic",
	"mappedports.proto":              "TCP (L4), UDP (L4), or HTTP (L7) protocol, one of LProtoUnknown, LProtoTcp, LProtoUdp, LProtoHttp",
	"mappedports.internalport":       "Container port",
	"mappedports.publicport":         "Public facing port for TCP/UDP (may be mapped on shared LB reverse proxy)",
	"mappedports.pathprefix":         "Public facing path for HTTP L7 access.",
	"mappedports.fqdnprefix":         "FQDN prefix to append to base FQDN in FindCloudlet response. May be empty.",
	"mappedports.endport":            "A non-zero end port indicates a port range from internal port to end port, inclusive.",
	"flavor":                         "Flavor name",
	"state":                          "Current state of the AppInst on the Cloudlet, one of TrackedStateUnknown, NotPresent, CreateRequested, Creating, CreateError, Ready, UpdateRequested, Updating, UpdateError, DeleteRequested, Deleting, DeleteError, DeletePrepare",
	"errors":                         "Any errors trying to create, update, or delete the AppInst on the Cloudlet",
	"crmoverride":                    "Override actions to CRM, one of NoOverride, IgnoreCrmErrors, IgnoreCrm, IgnoreTransientState, IgnoreCrmAndTransientState",
	"runtimeinfo.containerids":       "List of container names",
	"autoclusteripaccess":            "IpAccess for auto-clusters. Ignored otherwise., one of IpAccessUnknown, IpAccessDedicated, IpAccessDedicatedOrShared, IpAccessShared",
	"revision":                       "Revision increments each time the App is updated.  Updating the App Instance will sync the revision with that of the App",
	"forceupdate":                    "Force Appinst update when UpdateAppInst is done if revision matches",
	"updatemultiple":                 "Allow multiple instances to be updated at once",
	"configs.kind":                   "kind (type) of config, i.e. k8s-manifest, helm-values, deploygen-config",
	"configs.config":                 "config file contents or URI reference",
}
var AppInstSpecialArgs = map[string]string{
	"errors":                   "StringArray",
	"runtimeinfo.containerids": "StringArray",
}
var AppInstRuntimeRequiredArgs = []string{}
var AppInstRuntimeOptionalArgs = []string{
	"containerids",
}
var AppInstRuntimeAliasArgs = []string{}
var AppInstRuntimeComments = map[string]string{
	"containerids": "List of container names",
}
var AppInstRuntimeSpecialArgs = map[string]string{
	"containerids": "StringArray",
}
var AppInstInfoRequiredArgs = []string{
	"key.appkey.developerkey.name",
	"key.appkey.name",
	"key.appkey.version",
	"key.clusterinstkey.clusterkey.name",
	"key.clusterinstkey.cloudletkey.operatorkey.name",
	"key.clusterinstkey.cloudletkey.name",
	"key.clusterinstkey.developer",
}
var AppInstInfoOptionalArgs = []string{
	"notifyid",
	"state",
	"errors",
	"runtimeinfo.containerids",
	"status.tasknumber",
	"status.maxtasks",
	"status.taskname",
	"status.stepname",
}
var AppInstInfoAliasArgs = []string{}
var AppInstInfoComments = map[string]string{
	"key.appkey.developerkey.name":                    "Organization or Company Name that a Developer is part of",
	"key.appkey.name":                                 "App name",
	"key.appkey.version":                              "App version",
	"key.clusterinstkey.clusterkey.name":              "Cluster name",
	"key.clusterinstkey.cloudletkey.operatorkey.name": "Company or Organization name of the operator",
	"key.clusterinstkey.cloudletkey.name":             "Name of the cloudlet",
	"key.clusterinstkey.developer":                    "Name of Developer that this cluster belongs to",
	"notifyid":                                        "Id of client assigned by server (internal use only)",
	"state":                                           "Current state of the AppInst on the Cloudlet, one of TrackedStateUnknown, NotPresent, CreateRequested, Creating, CreateError, Ready, UpdateRequested, Updating, UpdateError, DeleteRequested, Deleting, DeleteError, DeletePrepare",
	"errors":                                          "Any errors trying to create, update, or delete the AppInst on the Cloudlet",
	"runtimeinfo.containerids":                        "List of container names",
}
var AppInstInfoSpecialArgs = map[string]string{
	"errors":                   "StringArray",
	"runtimeinfo.containerids": "StringArray",
}
var AppInstMetricsRequiredArgs = []string{}
var AppInstMetricsOptionalArgs = []string{
	"something",
}
var AppInstMetricsAliasArgs = []string{}
var AppInstMetricsComments = map[string]string{
	"something": "what goes here? Note that metrics for grpc calls can be done by a prometheus interceptor in grpc, so adding call metrics here may be redundant unless theyre needed for billing.",
}
var AppInstMetricsSpecialArgs = map[string]string{}
var CreateAppInstRequiredArgs = []string{
	"developer",
	"appname",
	"appvers",
	"cluster",
	"operator",
	"cloudlet",
}
var CreateAppInstOptionalArgs = []string{
	"clusterdeveloper",
	"flavor",
	"state",
	"crmoverride",
	"autoclusteripaccess",
	"forceupdate",
	"configs.kind",
	"configs.config",
}
var UpdateAppInstRequiredArgs = []string{
	"developer",
	"appname",
	"appvers",
}
var UpdateAppInstOptionalArgs = []string{
	"cluster",
	"operator",
	"cloudlet",
	"clusterdeveloper",
	"crmoverride",
	"forceupdate",
	"updatemultiple",
	"configs.kind",
	"configs.config",
}
