// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: appinst.proto

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
		in.Revision = ""
	}
	if _, found := tags["nocmp"]; found {
		in.ForceUpdate = false
	}
	if _, found := tags["nocmp"]; found {
		in.UpdateMultiple = false
	}
	for i0 := 0; i0 < len(in.Configs); i0++ {
	}
	if _, found := tags["nocmp"]; found {
		in.HealthCheck = 0
	}
	if _, found := tags["nocmp"]; found {
		in.PowerState = 0
	}
	if _, found := tags["nocmp"]; found {
		in.ExternalVolumeSize = 0
	}
	if _, found := tags["nocmp"]; found {
		in.AvailabilityZone = ""
	}
	if _, found := tags["nocmp"]; found {
		in.VmFlavor = ""
	}
	if _, found := tags["nocmp"]; found {
		in.OptRes = ""
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
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
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
			return fmt.Errorf("CreateAppInst recv failed: %s", errstr)
		}
		if cli.OutputStream {
			c.WriteOutput(obj, cli.OutputFormat)
			continue
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
	RequiredArgs: strings.Join(DeleteAppInstRequiredArgs, " "),
	OptionalArgs: strings.Join(DeleteAppInstOptionalArgs, " "),
	AliasArgs:    strings.Join(AppInstAliasArgs, " "),
	SpecialArgs:  &AppInstSpecialArgs,
	Comments:     AppInstComments,
	ReqData:      &edgeproto.AppInst{},
	ReplyData:    &edgeproto.Result{},
	Run:          runDeleteAppInst,
}

func runDeleteAppInst(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
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
			return fmt.Errorf("DeleteAppInst recv failed: %s", errstr)
		}
		if cli.OutputStream {
			c.WriteOutput(obj, cli.OutputFormat)
			continue
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

var RefreshAppInstCmd = &cli.Command{
	Use:          "RefreshAppInst",
	RequiredArgs: strings.Join(RefreshAppInstRequiredArgs, " "),
	OptionalArgs: strings.Join(RefreshAppInstOptionalArgs, " "),
	AliasArgs:    strings.Join(AppInstAliasArgs, " "),
	SpecialArgs:  &AppInstSpecialArgs,
	Comments:     AppInstComments,
	ReqData:      &edgeproto.AppInst{},
	ReplyData:    &edgeproto.Result{},
	Run:          runRefreshAppInst,
}

func runRefreshAppInst(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.AppInst)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return RefreshAppInst(c, obj)
}

func RefreshAppInst(c *cli.Command, in *edgeproto.AppInst) error {
	if AppInstApiCmd == nil {
		return fmt.Errorf("AppInstApi client not initialized")
	}
	ctx := context.Background()
	stream, err := AppInstApiCmd.RefreshAppInst(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("RefreshAppInst failed: %s", errstr)
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
			return fmt.Errorf("RefreshAppInst recv failed: %s", errstr)
		}
		if cli.OutputStream {
			c.WriteOutput(obj, cli.OutputFormat)
			continue
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
func RefreshAppInsts(c *cli.Command, data []edgeproto.AppInst, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("RefreshAppInst %v\n", data[ii])
		myerr := RefreshAppInst(c, &data[ii])
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
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
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
			return fmt.Errorf("UpdateAppInst recv failed: %s", errstr)
		}
		if cli.OutputStream {
			c.WriteOutput(obj, cli.OutputFormat)
			continue
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
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
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
			errstr := err.Error()
			st, ok := status.FromError(err)
			if ok {
				errstr = st.Message()
			}
			return fmt.Errorf("ShowAppInst recv failed: %s", errstr)
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
	RefreshAppInstCmd.GenCmd(),
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
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
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
			errstr := err.Error()
			st, ok := status.FromError(err)
			if ok {
				errstr = st.Message()
			}
			return fmt.Errorf("ShowAppInstInfo recv failed: %s", errstr)
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
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
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
			errstr := err.Error()
			st, ok := status.FromError(err)
			if ok {
				errstr = st.Message()
			}
			return fmt.Errorf("ShowAppInstMetrics recv failed: %s", errstr)
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
	"appkey.organization",
	"appkey.name",
	"appkey.version",
	"clusterinstkey.clusterkey.name",
	"clusterinstkey.cloudletkey.organization",
	"clusterinstkey.cloudletkey.name",
	"clusterinstkey.organization",
}
var AppInstKeyAliasArgs = []string{}
var AppInstKeyComments = map[string]string{
	"appkey.organization":                     "App developer organization",
	"appkey.name":                             "App name",
	"appkey.version":                          "App version",
	"clusterinstkey.clusterkey.name":          "Cluster name",
	"clusterinstkey.cloudletkey.organization": "Organization of the cloudlet site",
	"clusterinstkey.cloudletkey.name":         "Name of the cloudlet",
	"clusterinstkey.organization":             "Name of Developer organization that this cluster belongs to",
}
var AppInstKeySpecialArgs = map[string]string{}
var AppInstRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
	"cloudlet-org",
	"cloudlet",
}
var AppInstOptionalArgs = []string{
	"cluster",
	"cluster-org",
	"flavor",
	"crmoverride",
	"autoclusteripaccess",
	"forceupdate",
	"updatemultiple",
	"configs:#.kind",
	"configs:#.config",
	"sharedvolumesize",
	"healthcheck",
	"privacypolicy",
	"powerstate",
	"vmflavor",
	"optres",
}
var AppInstAliasArgs = []string{
	"app-org=key.appkey.organization",
	"appname=key.appkey.name",
	"appvers=key.appkey.version",
	"cluster=key.clusterinstkey.clusterkey.name",
	"cloudlet-org=key.clusterinstkey.cloudletkey.organization",
	"cloudlet=key.clusterinstkey.cloudletkey.name",
	"cluster-org=key.clusterinstkey.organization",
	"flavor=flavor.name",
}
var AppInstComments = map[string]string{
	"fields":                         "Fields are used for the Update API to specify which fields to apply",
	"app-org":                        "App developer organization",
	"appname":                        "App name",
	"appvers":                        "App version",
	"cluster":                        "Cluster name",
	"cloudlet-org":                   "Organization of the cloudlet site",
	"cloudlet":                       "Name of the cloudlet",
	"cluster-org":                    "Name of Developer organization that this cluster belongs to",
	"cloudletloc.latitude":           "latitude in WGS 84 coordinates",
	"cloudletloc.longitude":          "longitude in WGS 84 coordinates",
	"cloudletloc.horizontalaccuracy": "horizontal accuracy (radius in meters)",
	"cloudletloc.verticalaccuracy":   "vertical accuracy (meters)",
	"cloudletloc.altitude":           "On android only lat and long are guaranteed to be supplied altitude in meters",
	"cloudletloc.course":             "course (IOS) / bearing (Android) (degrees east relative to true north)",
	"cloudletloc.speed":              "speed (IOS) / velocity (Android) (meters/sec)",
	"uri":                            "Base FQDN (not really URI) for the App. See Service FQDN for endpoint access.",
	"liveness":                       "Liveness of instance (see Liveness), one of LivenessUnknown, LivenessStatic, LivenessDynamic, LivenessAutoprov",
	"mappedports:#.proto":            "TCP (L4) or UDP (L4) protocol, one of LProtoUnknown, LProtoTcp, LProtoUdp",
	"mappedports:#.internalport":     "Container port",
	"mappedports:#.publicport":       "Public facing port for TCP/UDP (may be mapped on shared LB reverse proxy)",
	"mappedports:#.fqdnprefix":       "FQDN prefix to append to base FQDN in FindCloudlet response. May be empty.",
	"mappedports:#.endport":          "A non-zero end port indicates a port range from internal port to end port, inclusive.",
	"mappedports:#.tls":              "TLS termination for this port",
	"mappedports:#.nginx":            "use nginx proxy for this port if you really need a transparent proxy (udp only)",
	"flavor":                         "Flavor name",
	"state":                          "Current state of the AppInst on the Cloudlet, one of TrackedStateUnknown, NotPresent, CreateRequested, Creating, CreateError, Ready, UpdateRequested, Updating, UpdateError, DeleteRequested, Deleting, DeleteError, DeletePrepare, CrmInitok, CreatingDependencies",
	"errors":                         "Any errors trying to create, update, or delete the AppInst on the Cloudlet",
	"crmoverride":                    "Override actions to CRM, one of NoOverride, IgnoreCrmErrors, IgnoreCrm, IgnoreTransientState, IgnoreCrmAndTransientState",
	"runtimeinfo.containerids":       "List of container names",
	"autoclusteripaccess":            "IpAccess for auto-clusters. Ignored otherwise., one of IpAccessUnknown, IpAccessDedicated, IpAccessShared",
	"revision":                       "Revision changes each time the App is updated.  Refreshing the App Instance will sync the revision with that of the App",
	"forceupdate":                    "Force Appinst refresh even if revision number matches App revision number.",
	"updatemultiple":                 "Allow multiple instances to be updated at once",
	"configs:#.kind":                 "kind (type) of config, i.e. envVarsYaml, helmCustomizationYaml",
	"configs:#.config":               "config file contents or URI reference",
	"sharedvolumesize":               "shared volume size when creating auto cluster",
	"healthcheck":                    "Health Check status, one of HealthCheckUnknown, HealthCheckFailRootlbOffline, HealthCheckFailServerFail, HealthCheckOk",
	"privacypolicy":                  "Optional privacy policy name",
	"powerstate":                     "Power State of the AppInst, one of PowerOn, PowerOff, Reboot",
	"externalvolumesize":             "Size of external volume to be attached to nodes.  This is for the root partition",
	"availabilityzone":               "Optional Availability Zone if any",
	"vmflavor":                       "OS node flavor to use",
	"optres":                         "Optional Resources required by OS flavor if any",
}
var AppInstSpecialArgs = map[string]string{
	"errors":                   "StringArray",
	"fields":                   "StringArray",
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
	"key.appkey.organization",
	"key.appkey.name",
	"key.appkey.version",
	"key.clusterinstkey.clusterkey.name",
	"key.clusterinstkey.cloudletkey.organization",
	"key.clusterinstkey.cloudletkey.name",
	"key.clusterinstkey.organization",
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
	"powerstate",
}
var AppInstInfoAliasArgs = []string{}
var AppInstInfoComments = map[string]string{
	"fields":                                      "Fields are used for the Update API to specify which fields to apply",
	"key.appkey.organization":                     "App developer organization",
	"key.appkey.name":                             "App name",
	"key.appkey.version":                          "App version",
	"key.clusterinstkey.clusterkey.name":          "Cluster name",
	"key.clusterinstkey.cloudletkey.organization": "Organization of the cloudlet site",
	"key.clusterinstkey.cloudletkey.name":         "Name of the cloudlet",
	"key.clusterinstkey.organization":             "Name of Developer organization that this cluster belongs to",
	"notifyid":                                    "Id of client assigned by server (internal use only)",
	"state":                                       "Current state of the AppInst on the Cloudlet, one of TrackedStateUnknown, NotPresent, CreateRequested, Creating, CreateError, Ready, UpdateRequested, Updating, UpdateError, DeleteRequested, Deleting, DeleteError, DeletePrepare, CrmInitok, CreatingDependencies",
	"errors":                                      "Any errors trying to create, update, or delete the AppInst on the Cloudlet",
	"runtimeinfo.containerids":                    "List of container names",
	"powerstate":                                  "Power State of the AppInst, one of PowerOn, PowerOff, Reboot",
}
var AppInstInfoSpecialArgs = map[string]string{
	"errors":                   "StringArray",
	"fields":                   "StringArray",
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
var AppInstLookupRequiredArgs = []string{
	"key.appkey.organization",
	"key.appkey.name",
	"key.appkey.version",
	"key.clusterinstkey.clusterkey.name",
	"key.clusterinstkey.cloudletkey.organization",
	"key.clusterinstkey.cloudletkey.name",
	"key.clusterinstkey.organization",
}
var AppInstLookupOptionalArgs = []string{
	"policykey.organization",
	"policykey.name",
}
var AppInstLookupAliasArgs = []string{}
var AppInstLookupComments = map[string]string{
	"key.appkey.organization":                     "App developer organization",
	"key.appkey.name":                             "App name",
	"key.appkey.version":                          "App version",
	"key.clusterinstkey.clusterkey.name":          "Cluster name",
	"key.clusterinstkey.cloudletkey.organization": "Organization of the cloudlet site",
	"key.clusterinstkey.cloudletkey.name":         "Name of the cloudlet",
	"key.clusterinstkey.organization":             "Name of Developer organization that this cluster belongs to",
	"policykey.organization":                      "Name of the organization for the cluster that this policy will apply to",
	"policykey.name":                              "Policy name",
}
var AppInstLookupSpecialArgs = map[string]string{}
var CreateAppInstRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
	"cloudlet-org",
	"cloudlet",
}
var CreateAppInstOptionalArgs = []string{
	"cluster",
	"cluster-org",
	"flavor",
	"crmoverride",
	"autoclusteripaccess",
	"configs:#.kind",
	"configs:#.config",
	"sharedvolumesize",
	"healthcheck",
	"privacypolicy",
	"vmflavor",
	"optres",
}
var DeleteAppInstRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
	"cloudlet-org",
	"cloudlet",
}
var DeleteAppInstOptionalArgs = []string{
	"cluster",
	"cluster-org",
	"flavor",
	"crmoverride",
	"autoclusteripaccess",
	"forceupdate",
	"updatemultiple",
	"configs:#.kind",
	"configs:#.config",
	"sharedvolumesize",
	"healthcheck",
	"privacypolicy",
	"vmflavor",
	"optres",
}
var RefreshAppInstRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
}
var RefreshAppInstOptionalArgs = []string{
	"cluster",
	"cloudlet-org",
	"cloudlet",
	"cluster-org",
	"crmoverride",
	"forceupdate",
	"updatemultiple",
}
var UpdateAppInstRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
	"cloudlet-org",
	"cloudlet",
}
var UpdateAppInstOptionalArgs = []string{
	"cluster",
	"cluster-org",
	"crmoverride",
	"configs:#.kind",
	"configs:#.config",
	"powerstate",
}
