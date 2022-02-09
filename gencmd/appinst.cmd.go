// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: appinst.proto

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
func AppInstHideTags(in *edgeproto.AppInst) {
	if cli.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cli.HideTags, ",") {
		tags[tag] = struct{}{}
	}
	for i0 := 0; i0 < len(in.MappedPorts); i0++ {
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
	if _, found := tags["timestamp"]; found {
		in.UpdatedAt = distributed_match_engine.Timestamp{}
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
	obj.Fields = cli.GetSpecifiedFields(jsonMap, c.ReqData)
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
	c.WriteOutput(c.CobraCmd.OutOrStdout(), objs, cli.OutputFormat)
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
	c.WriteOutput(c.CobraCmd.OutOrStdout(), objs, cli.OutputFormat)
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
	c.WriteOutput(c.CobraCmd.OutOrStdout(), objs, cli.OutputFormat)
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

var AppInstLatencyApiCmd edgeproto.AppInstLatencyApiClient

var RequestAppInstLatencyCmd = &cli.Command{
	Use:          "RequestAppInstLatency",
	RequiredArgs: strings.Join(AppInstLatencyRequiredArgs, " "),
	OptionalArgs: strings.Join(AppInstLatencyOptionalArgs, " "),
	AliasArgs:    strings.Join(AppInstLatencyAliasArgs, " "),
	SpecialArgs:  &AppInstLatencySpecialArgs,
	Comments:     AppInstLatencyComments,
	ReqData:      &edgeproto.AppInstLatency{},
	ReplyData:    &edgeproto.Result{},
	Run:          runRequestAppInstLatency,
}

func runRequestAppInstLatency(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*edgeproto.AppInstLatency)
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	return RequestAppInstLatency(c, obj)
}

func RequestAppInstLatency(c *cli.Command, in *edgeproto.AppInstLatency) error {
	if AppInstLatencyApiCmd == nil {
		return fmt.Errorf("AppInstLatencyApi client not initialized")
	}
	ctx := context.Background()
	obj, err := AppInstLatencyApiCmd.RequestAppInstLatency(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("RequestAppInstLatency failed: %s", errstr)
	}
	c.WriteOutput(c.CobraCmd.OutOrStdout(), obj, cli.OutputFormat)
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
func RequestAppInstLatencys(c *cli.Command, data []edgeproto.AppInstLatency, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("RequestAppInstLatency %v\n", data[ii])
		myerr := RequestAppInstLatency(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AppInstLatencyApiCmds = []*cobra.Command{
	RequestAppInstLatencyCmd.GenCmd(),
}

var VirtualClusterInstKeyRequiredArgs = []string{}
var VirtualClusterInstKeyOptionalArgs = []string{
	"clusterkey.name",
	"cloudletkey.organization",
	"cloudletkey.name",
	"cloudletkey.federatedorganization",
	"organization",
}
var VirtualClusterInstKeyAliasArgs = []string{}
var VirtualClusterInstKeyComments = map[string]string{
	"clusterkey.name":                   "Cluster name",
	"cloudletkey.organization":          "Organization of the cloudlet site",
	"cloudletkey.name":                  "Name of the cloudlet",
	"cloudletkey.federatedorganization": "Federated operator organization who shared this cloudlet",
	"organization":                      "Name of Developer organization that this cluster belongs to",
}
var VirtualClusterInstKeySpecialArgs = map[string]string{}
var AppInstKeyRequiredArgs = []string{}
var AppInstKeyOptionalArgs = []string{
	"appkey.organization",
	"appkey.name",
	"appkey.version",
	"clusterinstkey.clusterkey.name",
	"clusterinstkey.cloudletkey.organization",
	"clusterinstkey.cloudletkey.name",
	"clusterinstkey.cloudletkey.federatedorganization",
	"clusterinstkey.organization",
}
var AppInstKeyAliasArgs = []string{}
var AppInstKeyComments = map[string]string{
	"appkey.organization":                              "App developer organization",
	"appkey.name":                                      "App name",
	"appkey.version":                                   "App version",
	"clusterinstkey.clusterkey.name":                   "Cluster name",
	"clusterinstkey.cloudletkey.organization":          "Organization of the cloudlet site",
	"clusterinstkey.cloudletkey.name":                  "Name of the cloudlet",
	"clusterinstkey.cloudletkey.federatedorganization": "Federated operator organization who shared this cloudlet",
	"clusterinstkey.organization":                      "Name of Developer organization that this cluster belongs to",
}
var AppInstKeySpecialArgs = map[string]string{}
var AppInstRequiredArgs = []string{
	"apporg",
	"appname",
	"appvers",
	"cloudletorg",
	"cloudlet",
}
var AppInstOptionalArgs = []string{
	"cluster",
	"federatedorg",
	"clusterorg",
	"flavor",
	"crmoverride",
	"forceupdate",
	"updatemultiple",
	"configs:empty",
	"configs:#.kind",
	"configs:#.config",
	"healthcheck",
	"powerstate",
	"realclustername",
	"dedicatedip",
}
var AppInstAliasArgs = []string{
	"apporg=key.appkey.organization",
	"appname=key.appkey.name",
	"appvers=key.appkey.version",
	"cluster=key.clusterinstkey.clusterkey.name",
	"cloudletorg=key.clusterinstkey.cloudletkey.organization",
	"cloudlet=key.clusterinstkey.cloudletkey.name",
	"federatedorg=key.clusterinstkey.cloudletkey.federatedorganization",
	"clusterorg=key.clusterinstkey.organization",
	"flavor=flavor.name",
}
var AppInstComments = map[string]string{
	"fields":                         "Fields are used for the Update API to specify which fields to apply",
	"apporg":                         "App developer organization",
	"appname":                        "App name",
	"appvers":                        "App version",
	"cluster":                        "Cluster name",
	"cloudletorg":                    "Organization of the cloudlet site",
	"cloudlet":                       "Name of the cloudlet",
	"federatedorg":                   "Federated operator organization who shared this cloudlet",
	"clusterorg":                     "Name of Developer organization that this cluster belongs to",
	"cloudletloc.latitude":           "Latitude in WGS 84 coordinates",
	"cloudletloc.longitude":          "Longitude in WGS 84 coordinates",
	"cloudletloc.horizontalaccuracy": "Horizontal accuracy (radius in meters)",
	"cloudletloc.verticalaccuracy":   "Vertical accuracy (meters)",
	"cloudletloc.altitude":           "On android only lat and long are guaranteed to be supplied Altitude in meters",
	"cloudletloc.course":             "Course (IOS) / bearing (Android) (degrees east relative to true north)",
	"cloudletloc.speed":              "Speed (IOS) / velocity (Android) (meters/sec)",
	"cloudletloc.timestamp":          "Timestamp",
	"uri":                            "Base FQDN (not really URI) for the App. See Service FQDN for endpoint access.",
	"liveness":                       "Liveness of instance (see Liveness), one of Unknown, Static, Dynamic, Autoprov",
	"mappedports:empty":              "For instances accessible via a shared load balancer, defines the external ports on the shared load balancer that map to the internal ports External ports should be appended to the Uri for L4 access., specify mappedports:empty=true to clear",
	"mappedports:#.proto":            "TCP (L4) or UDP (L4) protocol, one of Unknown, Tcp, Udp",
	"mappedports:#.internalport":     "Container port",
	"mappedports:#.publicport":       "Public facing port for TCP/UDP (may be mapped on shared LB reverse proxy)",
	"mappedports:#.fqdnprefix":       "skip 4 to preserve the numbering. 4 was path_prefix but was removed since we dont need it after removed http FQDN prefix to append to base FQDN in FindCloudlet response. May be empty.",
	"mappedports:#.endport":          "A non-zero end port indicates a port range from internal port to end port, inclusive.",
	"mappedports:#.tls":              "TLS termination for this port",
	"mappedports:#.nginx":            "Use nginx proxy for this port if you really need a transparent proxy (udp only)",
	"mappedports:#.maxpktsize":       "Maximum datagram size (udp only)",
	"flavor":                         "Flavor name",
	"state":                          "Current state of the AppInst on the Cloudlet, one of TrackedStateUnknown, NotPresent, CreateRequested, Creating, CreateError, Ready, UpdateRequested, Updating, UpdateError, DeleteRequested, Deleting, DeleteError, DeletePrepare, CrmInitok, CreatingDependencies, DeleteDone",
	"errors":                         "Any errors trying to create, update, or delete the AppInst on the Cloudlet, specify errors:empty=true to clear",
	"crmoverride":                    "Override actions to CRM, one of NoOverride, IgnoreCrmErrors, IgnoreCrm, IgnoreTransientState, IgnoreCrmAndTransientState",
	"runtimeinfo.containerids":       "List of container names, specify runtimeinfo.containerids:empty=true to clear",
	"createdat":                      "Created at time",
	"autoclusteripaccess":            "(Deprecated) IpAccess for auto-clusters. Ignored otherwise., one of Unknown, Dedicated, Shared",
	"revision":                       "Revision changes each time the App is updated.  Refreshing the App Instance will sync the revision with that of the App",
	"forceupdate":                    "Force Appinst refresh even if revision number matches App revision number.",
	"updatemultiple":                 "Allow multiple instances to be updated at once",
	"configs:empty":                  "Customization files passed through to implementing services, specify configs:empty=true to clear",
	"configs:#.kind":                 "Kind (type) of config, i.e. envVarsYaml, helmCustomizationYaml",
	"configs:#.config":               "Config file contents or URI reference",
	"healthcheck":                    "Health Check status, one of Unknown, RootlbOffline, ServerFail, Ok, CloudletOffline",
	"powerstate":                     "Power State of the AppInst, one of PowerOn, PowerOff, Reboot",
	"externalvolumesize":             "Size of external volume to be attached to nodes.  This is for the root partition",
	"availabilityzone":               "Optional Availability Zone if any",
	"vmflavor":                       "OS node flavor to use",
	"optres":                         "Optional Resources required by OS flavor if any",
	"updatedat":                      "Updated at time",
	"realclustername":                "Real ClusterInst name",
	"internalporttolbip":             "mapping of ports to load balancer IPs, specify internalporttolbip:empty=true to clear",
	"dedicatedip":                    "Dedicated IP assigns an IP for this AppInst but requires platform support",
	"uniqueid":                       "A unique id for the AppInst within the region to be used by platforms",
	"dnslabel":                       "DNS label that is unique within the cloudlet and among other AppInsts/ClusterInsts",
}
var AppInstSpecialArgs = map[string]string{
	"errors":                   "StringArray",
	"fields":                   "StringArray",
	"internalporttolbip":       "StringToString",
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
	"key.clusterinstkey.cloudletkey.federatedorganization",
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
	"status.msgcount",
	"status.msgs",
	"powerstate",
	"uri",
}
var AppInstInfoAliasArgs = []string{}
var AppInstInfoComments = map[string]string{
	"fields":                                               "Fields are used for the Update API to specify which fields to apply",
	"key.appkey.organization":                              "App developer organization",
	"key.appkey.name":                                      "App name",
	"key.appkey.version":                                   "App version",
	"key.clusterinstkey.clusterkey.name":                   "Cluster name",
	"key.clusterinstkey.cloudletkey.organization":          "Organization of the cloudlet site",
	"key.clusterinstkey.cloudletkey.name":                  "Name of the cloudlet",
	"key.clusterinstkey.cloudletkey.federatedorganization": "Federated operator organization who shared this cloudlet",
	"key.clusterinstkey.organization":                      "Name of Developer organization that this cluster belongs to",
	"notifyid":                                             "Id of client assigned by server (internal use only)",
	"state":                                                "Current state of the AppInst on the Cloudlet, one of TrackedStateUnknown, NotPresent, CreateRequested, Creating, CreateError, Ready, UpdateRequested, Updating, UpdateError, DeleteRequested, Deleting, DeleteError, DeletePrepare, CrmInitok, CreatingDependencies, DeleteDone",
	"errors":                                               "Any errors trying to create, update, or delete the AppInst on the Cloudlet",
	"runtimeinfo.containerids":                             "List of container names",
	"powerstate":                                           "Power State of the AppInst, one of PowerOn, PowerOff, Reboot",
	"uri":                                                  "Base FQDN for the App based on the cloudlet platform",
}
var AppInstInfoSpecialArgs = map[string]string{
	"errors":                   "StringArray",
	"fields":                   "StringArray",
	"runtimeinfo.containerids": "StringArray",
	"status.msgs":              "StringArray",
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
	"key.clusterinstkey.cloudletkey.federatedorganization",
	"key.clusterinstkey.organization",
}
var AppInstLookupOptionalArgs = []string{
	"policykey.organization",
	"policykey.name",
}
var AppInstLookupAliasArgs = []string{}
var AppInstLookupComments = map[string]string{
	"key.appkey.organization":                              "App developer organization",
	"key.appkey.name":                                      "App name",
	"key.appkey.version":                                   "App version",
	"key.clusterinstkey.clusterkey.name":                   "Cluster name",
	"key.clusterinstkey.cloudletkey.organization":          "Organization of the cloudlet site",
	"key.clusterinstkey.cloudletkey.name":                  "Name of the cloudlet",
	"key.clusterinstkey.cloudletkey.federatedorganization": "Federated operator organization who shared this cloudlet",
	"key.clusterinstkey.organization":                      "Name of Developer organization that this cluster belongs to",
	"policykey.organization":                               "Name of the organization for the cluster that this policy will apply to",
	"policykey.name":                                       "Policy name",
}
var AppInstLookupSpecialArgs = map[string]string{}
var AppInstLookup2RequiredArgs = []string{
	"key.appkey.organization",
	"key.appkey.name",
	"key.appkey.version",
	"key.clusterinstkey.clusterkey.name",
	"key.clusterinstkey.cloudletkey.organization",
	"key.clusterinstkey.cloudletkey.name",
	"key.clusterinstkey.cloudletkey.federatedorganization",
	"key.clusterinstkey.organization",
}
var AppInstLookup2OptionalArgs = []string{
	"cloudletkey.organization",
	"cloudletkey.name",
	"cloudletkey.federatedorganization",
}
var AppInstLookup2AliasArgs = []string{}
var AppInstLookup2Comments = map[string]string{
	"key.appkey.organization":                              "App developer organization",
	"key.appkey.name":                                      "App name",
	"key.appkey.version":                                   "App version",
	"key.clusterinstkey.clusterkey.name":                   "Cluster name",
	"key.clusterinstkey.cloudletkey.organization":          "Organization of the cloudlet site",
	"key.clusterinstkey.cloudletkey.name":                  "Name of the cloudlet",
	"key.clusterinstkey.cloudletkey.federatedorganization": "Federated operator organization who shared this cloudlet",
	"key.clusterinstkey.organization":                      "Name of Developer organization that this cluster belongs to",
	"cloudletkey.organization":                             "Organization of the cloudlet site",
	"cloudletkey.name":                                     "Name of the cloudlet",
	"cloudletkey.federatedorganization":                    "Federated operator organization who shared this cloudlet",
}
var AppInstLookup2SpecialArgs = map[string]string{}
var AppInstLatencyRequiredArgs = []string{
	"apporg",
	"appname",
	"appvers",
	"cloudletorg",
	"cloudlet",
}
var AppInstLatencyOptionalArgs = []string{
	"cluster",
	"federatedorg",
	"clusterorg",
}
var AppInstLatencyAliasArgs = []string{
	"apporg=key.appkey.organization",
	"appname=key.appkey.name",
	"appvers=key.appkey.version",
	"cluster=key.clusterinstkey.clusterkey.name",
	"cloudletorg=key.clusterinstkey.cloudletkey.organization",
	"cloudlet=key.clusterinstkey.cloudletkey.name",
	"federatedorg=key.clusterinstkey.cloudletkey.federatedorganization",
	"clusterorg=key.clusterinstkey.organization",
}
var AppInstLatencyComments = map[string]string{
	"apporg":       "App developer organization",
	"appname":      "App name",
	"appvers":      "App version",
	"cluster":      "Cluster name",
	"cloudletorg":  "Organization of the cloudlet site",
	"cloudlet":     "Name of the cloudlet",
	"federatedorg": "Federated operator organization who shared this cloudlet",
	"clusterorg":   "Name of Developer organization that this cluster belongs to",
}
var AppInstLatencySpecialArgs = map[string]string{}
var CreateAppInstRequiredArgs = []string{
	"apporg",
	"appname",
	"appvers",
	"cloudletorg",
	"cloudlet",
}
var CreateAppInstOptionalArgs = []string{
	"cluster",
	"federatedorg",
	"clusterorg",
	"flavor",
	"crmoverride",
	"configs:#.kind",
	"configs:#.config",
	"healthcheck",
	"realclustername",
	"dedicatedip",
}
var DeleteAppInstRequiredArgs = []string{
	"apporg",
	"appname",
	"appvers",
	"cloudletorg",
	"cloudlet",
}
var DeleteAppInstOptionalArgs = []string{
	"cluster",
	"federatedorg",
	"clusterorg",
	"flavor",
	"crmoverride",
	"forceupdate",
	"updatemultiple",
	"configs:#.kind",
	"configs:#.config",
	"healthcheck",
	"realclustername",
	"dedicatedip",
}
var RefreshAppInstRequiredArgs = []string{
	"apporg",
	"appname",
	"appvers",
}
var RefreshAppInstOptionalArgs = []string{
	"cluster",
	"cloudletorg",
	"cloudlet",
	"federatedorg",
	"clusterorg",
	"crmoverride",
	"forceupdate",
	"updatemultiple",
	"realclustername",
	"dedicatedip",
}
var UpdateAppInstRequiredArgs = []string{
	"apporg",
	"appname",
	"appvers",
	"cloudletorg",
	"cloudlet",
}
var UpdateAppInstOptionalArgs = []string{
	"cluster",
	"federatedorg",
	"clusterorg",
	"crmoverride",
	"configs:empty",
	"configs:#.kind",
	"configs:#.config",
	"powerstate",
	"realclustername",
	"dedicatedip",
}
