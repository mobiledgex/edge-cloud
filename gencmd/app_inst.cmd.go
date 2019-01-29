// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app_inst.proto

package gencmd

import distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
import "strconv"
import "github.com/spf13/cobra"
import "context"
import "os"
import "io"
import "text/tabwriter"
import "github.com/spf13/pflag"
import "errors"
import "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/cmdsup"
import "google.golang.org/grpc/status"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/protocmd"
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
var AppInstApiCmd edgeproto.AppInstApiClient
var AppInstInfoApiCmd edgeproto.AppInstInfoApiClient
var AppInstMetricsApiCmd edgeproto.AppInstMetricsApiClient
var AppInstIn edgeproto.AppInst
var AppInstFlagSet = pflag.NewFlagSet("AppInst", pflag.ExitOnError)
var AppInstNoConfigFlagSet = pflag.NewFlagSet("AppInstNoConfig", pflag.ExitOnError)
var AppInstInLiveness string
var AppInstInMappedPortsProto string
var AppInstInIpAccess string
var AppInstInState string
var AppInstInCrmOverride string
var AppInstInfoIn edgeproto.AppInstInfo
var AppInstInfoFlagSet = pflag.NewFlagSet("AppInstInfo", pflag.ExitOnError)
var AppInstInfoNoConfigFlagSet = pflag.NewFlagSet("AppInstInfoNoConfig", pflag.ExitOnError)
var AppInstInfoInState string
var AppInstMetricsIn edgeproto.AppInstMetrics
var AppInstMetricsFlagSet = pflag.NewFlagSet("AppInstMetrics", pflag.ExitOnError)
var AppInstMetricsNoConfigFlagSet = pflag.NewFlagSet("AppInstMetricsNoConfig", pflag.ExitOnError)

func AppInstKeySlicer(in *edgeproto.AppInstKey) []string {
	s := make([]string, 0, 3)
	s = append(s, in.AppKey.DeveloperKey.Name)
	s = append(s, in.AppKey.Name)
	s = append(s, in.AppKey.Version)
	s = append(s, in.CloudletKey.OperatorKey.Name)
	s = append(s, in.CloudletKey.Name)
	s = append(s, strconv.FormatUint(uint64(in.Id), 10))
	return s
}

func AppInstKeyHeaderSlicer() []string {
	s := make([]string, 0, 3)
	s = append(s, "AppKey-DeveloperKey-Name")
	s = append(s, "AppKey-Name")
	s = append(s, "AppKey-Version")
	s = append(s, "CloudletKey-OperatorKey-Name")
	s = append(s, "CloudletKey-Name")
	s = append(s, "Id")
	return s
}

func AppInstKeyWriteOutputArray(objs []*edgeproto.AppInstKey) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(AppInstKeyHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(AppInstKeySlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func AppInstKeyWriteOutputOne(obj *edgeproto.AppInstKey) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(AppInstKeyHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(AppInstKeySlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}
func AppInstSlicer(in *edgeproto.AppInst) []string {
	s := make([]string, 0, 15)
	if in.Fields == nil {
		in.Fields = make([]string, 1)
	}
	s = append(s, in.Fields[0])
	s = append(s, in.Key.AppKey.DeveloperKey.Name)
	s = append(s, in.Key.AppKey.Name)
	s = append(s, in.Key.AppKey.Version)
	s = append(s, in.Key.CloudletKey.OperatorKey.Name)
	s = append(s, in.Key.CloudletKey.Name)
	s = append(s, strconv.FormatUint(uint64(in.Key.Id), 10))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLoc.Latitude), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLoc.Longitude), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLoc.HorizontalAccuracy), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLoc.VerticalAccuracy), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLoc.Altitude), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLoc.Course), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLoc.Speed), 'e', -1, 32))
	if in.CloudletLoc.Timestamp == nil {
		in.CloudletLoc.Timestamp = &distributed_match_engine.Timestamp{}
	}
	s = append(s, strconv.FormatUint(uint64(in.CloudletLoc.Timestamp.Seconds), 10))
	s = append(s, strconv.FormatUint(uint64(in.CloudletLoc.Timestamp.Nanos), 10))
	s = append(s, in.Uri)
	s = append(s, in.ClusterInstKey.ClusterKey.Name)
	s = append(s, in.ClusterInstKey.CloudletKey.OperatorKey.Name)
	s = append(s, in.ClusterInstKey.CloudletKey.Name)
	s = append(s, edgeproto.Liveness_name[int32(in.Liveness)])
	if in.MappedPorts == nil {
		in.MappedPorts = make([]distributed_match_engine.AppPort, 1)
	}
	s = append(s, distributed_match_engine.LProto_name[int32(in.MappedPorts[0].Proto)])
	s = append(s, strconv.FormatUint(uint64(in.MappedPorts[0].InternalPort), 10))
	s = append(s, strconv.FormatUint(uint64(in.MappedPorts[0].PublicPort), 10))
	s = append(s, in.MappedPorts[0].PublicPath)
	s = append(s, in.MappedPorts[0].FQDNPrefix)
	s = append(s, in.Flavor.Name)
	s = append(s, edgeproto.IpAccess_name[int32(in.IpAccess)])
	s = append(s, edgeproto.TrackedState_name[int32(in.State)])
	if in.Errors == nil {
		in.Errors = make([]string, 1)
	}
	s = append(s, in.Errors[0])
	s = append(s, edgeproto.CRMOverride_name[int32(in.CrmOverride)])
	s = append(s, in.AllocatedIp)
	s = append(s, in.Status)
	s = append(s, strconv.FormatBool(in.PreventAutoClusterInst))
	return s
}

func AppInstHeaderSlicer() []string {
	s := make([]string, 0, 15)
	s = append(s, "Fields")
	s = append(s, "Key-AppKey-DeveloperKey-Name")
	s = append(s, "Key-AppKey-Name")
	s = append(s, "Key-AppKey-Version")
	s = append(s, "Key-CloudletKey-OperatorKey-Name")
	s = append(s, "Key-CloudletKey-Name")
	s = append(s, "Key-Id")
	s = append(s, "CloudletLoc-Latitude")
	s = append(s, "CloudletLoc-Longitude")
	s = append(s, "CloudletLoc-HorizontalAccuracy")
	s = append(s, "CloudletLoc-VerticalAccuracy")
	s = append(s, "CloudletLoc-Altitude")
	s = append(s, "CloudletLoc-Course")
	s = append(s, "CloudletLoc-Speed")
	s = append(s, "CloudletLoc-Timestamp-Seconds")
	s = append(s, "CloudletLoc-Timestamp-Nanos")
	s = append(s, "Uri")
	s = append(s, "ClusterInstKey-ClusterKey-Name")
	s = append(s, "ClusterInstKey-CloudletKey-OperatorKey-Name")
	s = append(s, "ClusterInstKey-CloudletKey-Name")
	s = append(s, "Liveness")
	s = append(s, "MappedPorts-Proto")
	s = append(s, "MappedPorts-InternalPort")
	s = append(s, "MappedPorts-PublicPort")
	s = append(s, "MappedPorts-PublicPath")
	s = append(s, "MappedPorts-FQDNPrefix")
	s = append(s, "Flavor-Name")
	s = append(s, "IpAccess")
	s = append(s, "State")
	s = append(s, "Errors")
	s = append(s, "CrmOverride")
	s = append(s, "AllocatedIp")
	s = append(s, "Status")
	s = append(s, "PreventAutoClusterInst")
	return s
}

func AppInstWriteOutputArray(objs []*edgeproto.AppInst) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(AppInstHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(AppInstSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func AppInstWriteOutputOne(obj *edgeproto.AppInst) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(AppInstHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(AppInstSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}
func AppInstInfoSlicer(in *edgeproto.AppInstInfo) []string {
	s := make([]string, 0, 5)
	if in.Fields == nil {
		in.Fields = make([]string, 1)
	}
	s = append(s, in.Fields[0])
	s = append(s, in.Key.AppKey.DeveloperKey.Name)
	s = append(s, in.Key.AppKey.Name)
	s = append(s, in.Key.AppKey.Version)
	s = append(s, in.Key.CloudletKey.OperatorKey.Name)
	s = append(s, in.Key.CloudletKey.Name)
	s = append(s, strconv.FormatUint(uint64(in.Key.Id), 10))
	s = append(s, strconv.FormatUint(uint64(in.NotifyId), 10))
	s = append(s, edgeproto.TrackedState_name[int32(in.State)])
	if in.Errors == nil {
		in.Errors = make([]string, 1)
	}
	s = append(s, in.Errors[0])
	return s
}

func AppInstInfoHeaderSlicer() []string {
	s := make([]string, 0, 5)
	s = append(s, "Fields")
	s = append(s, "Key-AppKey-DeveloperKey-Name")
	s = append(s, "Key-AppKey-Name")
	s = append(s, "Key-AppKey-Version")
	s = append(s, "Key-CloudletKey-OperatorKey-Name")
	s = append(s, "Key-CloudletKey-Name")
	s = append(s, "Key-Id")
	s = append(s, "NotifyId")
	s = append(s, "State")
	s = append(s, "Errors")
	return s
}

func AppInstInfoWriteOutputArray(objs []*edgeproto.AppInstInfo) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(AppInstInfoHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(AppInstInfoSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func AppInstInfoWriteOutputOne(obj *edgeproto.AppInstInfo) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(AppInstInfoHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(AppInstInfoSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}
func AppInstMetricsSlicer(in *edgeproto.AppInstMetrics) []string {
	s := make([]string, 0, 1)
	s = append(s, strconv.FormatUint(uint64(in.Something), 10))
	return s
}

func AppInstMetricsHeaderSlicer() []string {
	s := make([]string, 0, 1)
	s = append(s, "Something")
	return s
}

func AppInstMetricsWriteOutputArray(objs []*edgeproto.AppInstMetrics) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(AppInstMetricsHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(AppInstMetricsSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func AppInstMetricsWriteOutputOne(obj *edgeproto.AppInstMetrics) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(AppInstMetricsHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(AppInstMetricsSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}
func AppInstStatusSlicer(in *edgeproto.AppInstStatus) []string {
	s := make([]string, 0, 3)
	if in.Fields == nil {
		in.Fields = make([]string, 1)
	}
	s = append(s, in.Fields[0])
	s = append(s, in.Key.AppKey.DeveloperKey.Name)
	s = append(s, in.Key.AppKey.Name)
	s = append(s, in.Key.AppKey.Version)
	s = append(s, in.Key.CloudletKey.OperatorKey.Name)
	s = append(s, in.Key.CloudletKey.Name)
	s = append(s, strconv.FormatUint(uint64(in.Key.Id), 10))
	s = append(s, in.Status)
	return s
}

func AppInstStatusHeaderSlicer() []string {
	s := make([]string, 0, 3)
	s = append(s, "Fields")
	s = append(s, "Key-AppKey-DeveloperKey-Name")
	s = append(s, "Key-AppKey-Name")
	s = append(s, "Key-AppKey-Version")
	s = append(s, "Key-CloudletKey-OperatorKey-Name")
	s = append(s, "Key-CloudletKey-Name")
	s = append(s, "Key-Id")
	s = append(s, "Status")
	return s
}

func AppInstStatusWriteOutputArray(objs []*edgeproto.AppInstStatus) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(AppInstStatusHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(AppInstStatusSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func AppInstStatusWriteOutputOne(obj *edgeproto.AppInstStatus) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(AppInstStatusHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(AppInstStatusSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}
func AppInstHideTags(in *edgeproto.AppInst) {
	if cmdsup.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cmdsup.HideTags, ",") {
		tags[tag] = struct{}{}
	}
	if _, found := tags["nocmp"]; found {
		in.MappedPorts = nil
	}
	if _, found := tags["nocmp"]; found {
		in.IpAccess = 0
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
		in.Status = ""
	}
}

func AppInstInfoHideTags(in *edgeproto.AppInstInfo) {
	if cmdsup.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cmdsup.HideTags, ",") {
		tags[tag] = struct{}{}
	}
	if _, found := tags["nocmp"]; found {
		in.NotifyId = 0
	}
}

var CreateAppInstCmd = &cobra.Command{
	Use: "CreateAppInst",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if we got this far, usage has been met.
		cmd.SilenceUsage = true
		err := parseAppInstEnums()
		if err != nil {
			return fmt.Errorf("CreateAppInst failed: %s", err.Error())
		}
		return CreateAppInst(&AppInstIn)
	},
}

func CreateAppInst(in *edgeproto.AppInst) error {
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
		ResultWriteOutputOne(obj)
	}
	return nil
}

func CreateAppInsts(data []edgeproto.AppInst, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("CreateAppInst %v\n", data[ii])
		myerr := CreateAppInst(&data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var DeleteAppInstCmd = &cobra.Command{
	Use: "DeleteAppInst",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if we got this far, usage has been met.
		cmd.SilenceUsage = true
		err := parseAppInstEnums()
		if err != nil {
			return fmt.Errorf("DeleteAppInst failed: %s", err.Error())
		}
		return DeleteAppInst(&AppInstIn)
	},
}

func DeleteAppInst(in *edgeproto.AppInst) error {
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
		ResultWriteOutputOne(obj)
	}
	return nil
}

func DeleteAppInsts(data []edgeproto.AppInst, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("DeleteAppInst %v\n", data[ii])
		myerr := DeleteAppInst(&data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var UpdateAppInstCmd = &cobra.Command{
	Use: "UpdateAppInst",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if we got this far, usage has been met.
		cmd.SilenceUsage = true
		err := parseAppInstEnums()
		if err != nil {
			return fmt.Errorf("UpdateAppInst failed: %s", err.Error())
		}
		AppInstSetFields()
		return UpdateAppInst(&AppInstIn)
	},
}

func UpdateAppInst(in *edgeproto.AppInst) error {
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
		ResultWriteOutputOne(obj)
	}
	return nil
}

func UpdateAppInsts(data []edgeproto.AppInst, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("UpdateAppInst %v\n", data[ii])
		myerr := UpdateAppInst(&data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var ShowAppInstCmd = &cobra.Command{
	Use: "ShowAppInst",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if we got this far, usage has been met.
		cmd.SilenceUsage = true
		err := parseAppInstEnums()
		if err != nil {
			return fmt.Errorf("ShowAppInst failed: %s", err.Error())
		}
		return ShowAppInst(&AppInstIn)
	},
}

func ShowAppInst(in *edgeproto.AppInst) error {
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
	AppInstWriteOutputArray(objs)
	return nil
}

func ShowAppInsts(data []edgeproto.AppInst, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowAppInst %v\n", data[ii])
		myerr := ShowAppInst(&data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AppInstApiCmds = []*cobra.Command{
	CreateAppInstCmd,
	DeleteAppInstCmd,
	UpdateAppInstCmd,
	ShowAppInstCmd,
}

var ShowAppInstInfoCmd = &cobra.Command{
	Use: "ShowAppInstInfo",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if we got this far, usage has been met.
		cmd.SilenceUsage = true
		err := parseAppInstInfoEnums()
		if err != nil {
			return fmt.Errorf("ShowAppInstInfo failed: %s", err.Error())
		}
		return ShowAppInstInfo(&AppInstInfoIn)
	},
}

func ShowAppInstInfo(in *edgeproto.AppInstInfo) error {
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
	AppInstInfoWriteOutputArray(objs)
	return nil
}

func ShowAppInstInfos(data []edgeproto.AppInstInfo, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowAppInstInfo %v\n", data[ii])
		myerr := ShowAppInstInfo(&data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AppInstInfoApiCmds = []*cobra.Command{
	ShowAppInstInfoCmd,
}

var ShowAppInstMetricsCmd = &cobra.Command{
	Use: "ShowAppInstMetrics",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if we got this far, usage has been met.
		cmd.SilenceUsage = true
		return ShowAppInstMetrics(&AppInstMetricsIn)
	},
}

func ShowAppInstMetrics(in *edgeproto.AppInstMetrics) error {
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
	AppInstMetricsWriteOutputArray(objs)
	return nil
}

func ShowAppInstMetricss(data []edgeproto.AppInstMetrics, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("ShowAppInstMetrics %v\n", data[ii])
		myerr := ShowAppInstMetrics(&data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var AppInstMetricsApiCmds = []*cobra.Command{
	ShowAppInstMetricsCmd,
}

func init() {
	AppInstFlagSet.StringVar(&AppInstIn.Key.AppKey.DeveloperKey.Name, "key-appkey-developerkey-name", "", "Key.AppKey.DeveloperKey.Name")
	AppInstFlagSet.StringVar(&AppInstIn.Key.AppKey.Name, "key-appkey-name", "", "Key.AppKey.Name")
	AppInstFlagSet.StringVar(&AppInstIn.Key.AppKey.Version, "key-appkey-version", "", "Key.AppKey.Version")
	AppInstFlagSet.StringVar(&AppInstIn.Key.CloudletKey.OperatorKey.Name, "key-cloudletkey-operatorkey-name", "", "Key.CloudletKey.OperatorKey.Name")
	AppInstFlagSet.StringVar(&AppInstIn.Key.CloudletKey.Name, "key-cloudletkey-name", "", "Key.CloudletKey.Name")
	AppInstFlagSet.Uint64Var(&AppInstIn.Key.Id, "key-id", 0, "Key.Id")
	AppInstNoConfigFlagSet.Float64Var(&AppInstIn.CloudletLoc.Latitude, "cloudletloc-latitude", 0, "CloudletLoc.Latitude")
	AppInstNoConfigFlagSet.Float64Var(&AppInstIn.CloudletLoc.Longitude, "cloudletloc-longitude", 0, "CloudletLoc.Longitude")
	AppInstNoConfigFlagSet.Float64Var(&AppInstIn.CloudletLoc.HorizontalAccuracy, "cloudletloc-horizontalaccuracy", 0, "CloudletLoc.HorizontalAccuracy")
	AppInstNoConfigFlagSet.Float64Var(&AppInstIn.CloudletLoc.VerticalAccuracy, "cloudletloc-verticalaccuracy", 0, "CloudletLoc.VerticalAccuracy")
	AppInstNoConfigFlagSet.Float64Var(&AppInstIn.CloudletLoc.Altitude, "cloudletloc-altitude", 0, "CloudletLoc.Altitude")
	AppInstNoConfigFlagSet.Float64Var(&AppInstIn.CloudletLoc.Course, "cloudletloc-course", 0, "CloudletLoc.Course")
	AppInstNoConfigFlagSet.Float64Var(&AppInstIn.CloudletLoc.Speed, "cloudletloc-speed", 0, "CloudletLoc.Speed")
	AppInstIn.CloudletLoc.Timestamp = &distributed_match_engine.Timestamp{}
	AppInstNoConfigFlagSet.Int64Var(&AppInstIn.CloudletLoc.Timestamp.Seconds, "cloudletloc-timestamp-seconds", 0, "CloudletLoc.Timestamp.Seconds")
	AppInstNoConfigFlagSet.Int32Var(&AppInstIn.CloudletLoc.Timestamp.Nanos, "cloudletloc-timestamp-nanos", 0, "CloudletLoc.Timestamp.Nanos")
	AppInstFlagSet.StringVar(&AppInstIn.Uri, "uri", "", "Uri")
	AppInstNoConfigFlagSet.StringVar(&AppInstIn.ClusterInstKey.ClusterKey.Name, "clusterinstkey-clusterkey-name", "", "ClusterInstKey.ClusterKey.Name")
	AppInstNoConfigFlagSet.StringVar(&AppInstIn.ClusterInstKey.CloudletKey.OperatorKey.Name, "clusterinstkey-cloudletkey-operatorkey-name", "", "ClusterInstKey.CloudletKey.OperatorKey.Name")
	AppInstNoConfigFlagSet.StringVar(&AppInstIn.ClusterInstKey.CloudletKey.Name, "clusterinstkey-cloudletkey-name", "", "ClusterInstKey.CloudletKey.Name")
	AppInstNoConfigFlagSet.StringVar(&AppInstInLiveness, "liveness", "", "one of [LivenessUnknown LivenessStatic LivenessDynamic]")
	AppInstFlagSet.StringVar(&AppInstIn.Flavor.Name, "flavor-name", "", "Flavor.Name")
	AppInstFlagSet.StringVar(&AppInstInIpAccess, "ipaccess", "", "one of [IpAccessUnknown IpAccessDedicated IpAccessDedicatedOrShared IpAccessShared]")
	AppInstFlagSet.StringVar(&AppInstInState, "state", "", "one of [TrackedStateUnknown NotPresent CreateRequested Creating CreateError Ready UpdateRequested Updating UpdateError DeleteRequested Deleting DeleteError]")
	AppInstFlagSet.StringVar(&AppInstInCrmOverride, "crmoverride", "", "one of [NoOverride IgnoreCRMErrors IgnoreCRM IgnoreTransientState IgnoreCRMandTransientState]")
	AppInstFlagSet.StringVar(&AppInstIn.AllocatedIp, "allocatedip", "", "AllocatedIp")
	AppInstFlagSet.StringVar(&AppInstIn.Status, "status", "", "Status")
	AppInstFlagSet.BoolVar(&AppInstIn.PreventAutoClusterInst, "preventautoclusterinst", false, "PreventAutoClusterInst")
	AppInstInfoFlagSet.StringVar(&AppInstInfoIn.Key.AppKey.DeveloperKey.Name, "key-appkey-developerkey-name", "", "Key.AppKey.DeveloperKey.Name")
	AppInstInfoFlagSet.StringVar(&AppInstInfoIn.Key.AppKey.Name, "key-appkey-name", "", "Key.AppKey.Name")
	AppInstInfoFlagSet.StringVar(&AppInstInfoIn.Key.AppKey.Version, "key-appkey-version", "", "Key.AppKey.Version")
	AppInstInfoFlagSet.StringVar(&AppInstInfoIn.Key.CloudletKey.OperatorKey.Name, "key-cloudletkey-operatorkey-name", "", "Key.CloudletKey.OperatorKey.Name")
	AppInstInfoFlagSet.StringVar(&AppInstInfoIn.Key.CloudletKey.Name, "key-cloudletkey-name", "", "Key.CloudletKey.Name")
	AppInstInfoFlagSet.Uint64Var(&AppInstInfoIn.Key.Id, "key-id", 0, "Key.Id")
	AppInstInfoFlagSet.Int64Var(&AppInstInfoIn.NotifyId, "notifyid", 0, "NotifyId")
	AppInstInfoFlagSet.StringVar(&AppInstInfoInState, "state", "", "one of [TrackedStateUnknown NotPresent CreateRequested Creating CreateError Ready UpdateRequested Updating UpdateError DeleteRequested Deleting DeleteError]")
	AppInstMetricsFlagSet.Uint64Var(&AppInstMetricsIn.Something, "something", 0, "Something")
	CreateAppInstCmd.Flags().AddFlagSet(AppInstFlagSet)
	DeleteAppInstCmd.Flags().AddFlagSet(AppInstFlagSet)
	UpdateAppInstCmd.Flags().AddFlagSet(AppInstFlagSet)
	ShowAppInstCmd.Flags().AddFlagSet(AppInstFlagSet)
	ShowAppInstInfoCmd.Flags().AddFlagSet(AppInstInfoFlagSet)
	ShowAppInstMetricsCmd.Flags().AddFlagSet(AppInstMetricsFlagSet)
}

func AppInstApiAllowNoConfig() {
	CreateAppInstCmd.Flags().AddFlagSet(AppInstNoConfigFlagSet)
	DeleteAppInstCmd.Flags().AddFlagSet(AppInstNoConfigFlagSet)
	UpdateAppInstCmd.Flags().AddFlagSet(AppInstNoConfigFlagSet)
	ShowAppInstCmd.Flags().AddFlagSet(AppInstNoConfigFlagSet)
}

func AppInstInfoApiAllowNoConfig() {
	ShowAppInstInfoCmd.Flags().AddFlagSet(AppInstInfoNoConfigFlagSet)
}

func AppInstMetricsApiAllowNoConfig() {
	ShowAppInstMetricsCmd.Flags().AddFlagSet(AppInstMetricsNoConfigFlagSet)
}

func AppInstSetFields() {
	AppInstIn.Fields = make([]string, 0)
	if AppInstFlagSet.Lookup("key-appkey-developerkey-name").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "2.1.1.2")
	}
	if AppInstFlagSet.Lookup("key-appkey-name").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "2.1.2")
	}
	if AppInstFlagSet.Lookup("key-appkey-version").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "2.1.3")
	}
	if AppInstFlagSet.Lookup("key-cloudletkey-operatorkey-name").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "2.2.1.1")
	}
	if AppInstFlagSet.Lookup("key-cloudletkey-name").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "2.2.2")
	}
	if AppInstFlagSet.Lookup("key-id").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "2.3")
	}
	if AppInstNoConfigFlagSet.Lookup("cloudletloc-latitude").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "3.1")
	}
	if AppInstNoConfigFlagSet.Lookup("cloudletloc-longitude").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "3.2")
	}
	if AppInstNoConfigFlagSet.Lookup("cloudletloc-horizontalaccuracy").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "3.3")
	}
	if AppInstNoConfigFlagSet.Lookup("cloudletloc-verticalaccuracy").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "3.4")
	}
	if AppInstNoConfigFlagSet.Lookup("cloudletloc-altitude").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "3.5")
	}
	if AppInstNoConfigFlagSet.Lookup("cloudletloc-course").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "3.6")
	}
	if AppInstNoConfigFlagSet.Lookup("cloudletloc-speed").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "3.7")
	}
	if AppInstNoConfigFlagSet.Lookup("cloudletloc-timestamp-seconds").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "3.8.1")
	}
	if AppInstNoConfigFlagSet.Lookup("cloudletloc-timestamp-nanos").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "3.8.2")
	}
	if AppInstFlagSet.Lookup("uri").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "4")
	}
	if AppInstNoConfigFlagSet.Lookup("clusterinstkey-clusterkey-name").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "5.1.1")
	}
	if AppInstNoConfigFlagSet.Lookup("clusterinstkey-cloudletkey-operatorkey-name").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "5.2.1.1")
	}
	if AppInstNoConfigFlagSet.Lookup("clusterinstkey-cloudletkey-name").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "5.2.2")
	}
	if AppInstNoConfigFlagSet.Lookup("liveness").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "6")
	}
	if AppInstFlagSet.Lookup("flavor-name").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "12.1")
	}
	if AppInstFlagSet.Lookup("ipaccess").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "13")
	}
	if AppInstFlagSet.Lookup("state").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "14")
	}
	if AppInstFlagSet.Lookup("crmoverride").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "16")
	}
	if AppInstFlagSet.Lookup("allocatedip").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "17")
	}
	if AppInstFlagSet.Lookup("status").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "18")
	}
	if AppInstFlagSet.Lookup("preventautoclusterinst").Changed {
		AppInstIn.Fields = append(AppInstIn.Fields, "19")
	}
}

func AppInstInfoSetFields() {
	AppInstInfoIn.Fields = make([]string, 0)
	if AppInstInfoFlagSet.Lookup("key-appkey-developerkey-name").Changed {
		AppInstInfoIn.Fields = append(AppInstInfoIn.Fields, "2.1.1.2")
	}
	if AppInstInfoFlagSet.Lookup("key-appkey-name").Changed {
		AppInstInfoIn.Fields = append(AppInstInfoIn.Fields, "2.1.2")
	}
	if AppInstInfoFlagSet.Lookup("key-appkey-version").Changed {
		AppInstInfoIn.Fields = append(AppInstInfoIn.Fields, "2.1.3")
	}
	if AppInstInfoFlagSet.Lookup("key-cloudletkey-operatorkey-name").Changed {
		AppInstInfoIn.Fields = append(AppInstInfoIn.Fields, "2.2.1.1")
	}
	if AppInstInfoFlagSet.Lookup("key-cloudletkey-name").Changed {
		AppInstInfoIn.Fields = append(AppInstInfoIn.Fields, "2.2.2")
	}
	if AppInstInfoFlagSet.Lookup("key-id").Changed {
		AppInstInfoIn.Fields = append(AppInstInfoIn.Fields, "2.3")
	}
	if AppInstInfoFlagSet.Lookup("notifyid").Changed {
		AppInstInfoIn.Fields = append(AppInstInfoIn.Fields, "3")
	}
	if AppInstInfoFlagSet.Lookup("state").Changed {
		AppInstInfoIn.Fields = append(AppInstInfoIn.Fields, "4")
	}
}

func parseAppInstEnums() error {
	if AppInstInLiveness != "" {
		switch AppInstInLiveness {
		case "LivenessUnknown":
			AppInstIn.Liveness = edgeproto.Liveness(0)
		case "LivenessStatic":
			AppInstIn.Liveness = edgeproto.Liveness(1)
		case "LivenessDynamic":
			AppInstIn.Liveness = edgeproto.Liveness(2)
		default:
			return errors.New("Invalid value for AppInstInLiveness")
		}
	}
	if AppInstInMappedPortsProto != "" {
		switch AppInstInMappedPortsProto {
		case "LProtoUnknown":
			AppInstIn.MappedPorts[0].Proto = distributed_match_engine.LProto(0)
		case "LProtoTCP":
			AppInstIn.MappedPorts[0].Proto = distributed_match_engine.LProto(1)
		case "LProtoUDP":
			AppInstIn.MappedPorts[0].Proto = distributed_match_engine.LProto(2)
		case "LProtoHTTP":
			AppInstIn.MappedPorts[0].Proto = distributed_match_engine.LProto(3)
		default:
			return errors.New("Invalid value for AppInstInMappedPortsProto")
		}
	}
	if AppInstInIpAccess != "" {
		switch AppInstInIpAccess {
		case "IpAccessUnknown":
			AppInstIn.IpAccess = edgeproto.IpAccess(0)
		case "IpAccessDedicated":
			AppInstIn.IpAccess = edgeproto.IpAccess(1)
		case "IpAccessDedicatedOrShared":
			AppInstIn.IpAccess = edgeproto.IpAccess(2)
		case "IpAccessShared":
			AppInstIn.IpAccess = edgeproto.IpAccess(3)
		default:
			return errors.New("Invalid value for AppInstInIpAccess")
		}
	}
	if AppInstInState != "" {
		switch AppInstInState {
		case "TrackedStateUnknown":
			AppInstIn.State = edgeproto.TrackedState(0)
		case "NotPresent":
			AppInstIn.State = edgeproto.TrackedState(1)
		case "CreateRequested":
			AppInstIn.State = edgeproto.TrackedState(2)
		case "Creating":
			AppInstIn.State = edgeproto.TrackedState(3)
		case "CreateError":
			AppInstIn.State = edgeproto.TrackedState(4)
		case "Ready":
			AppInstIn.State = edgeproto.TrackedState(5)
		case "UpdateRequested":
			AppInstIn.State = edgeproto.TrackedState(6)
		case "Updating":
			AppInstIn.State = edgeproto.TrackedState(7)
		case "UpdateError":
			AppInstIn.State = edgeproto.TrackedState(8)
		case "DeleteRequested":
			AppInstIn.State = edgeproto.TrackedState(9)
		case "Deleting":
			AppInstIn.State = edgeproto.TrackedState(10)
		case "DeleteError":
			AppInstIn.State = edgeproto.TrackedState(11)
		default:
			return errors.New("Invalid value for AppInstInState")
		}
	}
	if AppInstInCrmOverride != "" {
		switch AppInstInCrmOverride {
		case "NoOverride":
			AppInstIn.CrmOverride = edgeproto.CRMOverride(0)
		case "IgnoreCRMErrors":
			AppInstIn.CrmOverride = edgeproto.CRMOverride(1)
		case "IgnoreCRM":
			AppInstIn.CrmOverride = edgeproto.CRMOverride(2)
		case "IgnoreTransientState":
			AppInstIn.CrmOverride = edgeproto.CRMOverride(3)
		case "IgnoreCRMandTransientState":
			AppInstIn.CrmOverride = edgeproto.CRMOverride(4)
		default:
			return errors.New("Invalid value for AppInstInCrmOverride")
		}
	}
	return nil
}

func parseAppInstInfoEnums() error {
	if AppInstInfoInState != "" {
		switch AppInstInfoInState {
		case "TrackedStateUnknown":
			AppInstInfoIn.State = edgeproto.TrackedState(0)
		case "NotPresent":
			AppInstInfoIn.State = edgeproto.TrackedState(1)
		case "CreateRequested":
			AppInstInfoIn.State = edgeproto.TrackedState(2)
		case "Creating":
			AppInstInfoIn.State = edgeproto.TrackedState(3)
		case "CreateError":
			AppInstInfoIn.State = edgeproto.TrackedState(4)
		case "Ready":
			AppInstInfoIn.State = edgeproto.TrackedState(5)
		case "UpdateRequested":
			AppInstInfoIn.State = edgeproto.TrackedState(6)
		case "Updating":
			AppInstInfoIn.State = edgeproto.TrackedState(7)
		case "UpdateError":
			AppInstInfoIn.State = edgeproto.TrackedState(8)
		case "DeleteRequested":
			AppInstInfoIn.State = edgeproto.TrackedState(9)
		case "Deleting":
			AppInstInfoIn.State = edgeproto.TrackedState(10)
		case "DeleteError":
			AppInstInfoIn.State = edgeproto.TrackedState(11)
		default:
			return errors.New("Invalid value for AppInstInfoInState")
		}
	}
	return nil
}
