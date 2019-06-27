// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: predictiveqos.proto

package gencmd

import distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import "strings"
import "strconv"
import "github.com/spf13/cobra"
import "context"
import "os"
import "io"
import "text/tabwriter"
import "github.com/spf13/pflag"
import "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/cmdsup"
import "google.golang.org/grpc/status"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
var QueryQoSCmd distributed_match_engine.QueryQoSClient
var HealthCmd distributed_match_engine.HealthClient
var HealthCheckRequestIn distributed_match_engine.HealthCheckRequest
var HealthCheckRequestFlagSet = pflag.NewFlagSet("HealthCheckRequest", pflag.ExitOnError)
var HealthCheckRequestNoConfigFlagSet = pflag.NewFlagSet("HealthCheckRequestNoConfig", pflag.ExitOnError)
var QoSKPIRequestIn distributed_match_engine.QoSKPIRequest
var QoSKPIRequestFlagSet = pflag.NewFlagSet("QoSKPIRequest", pflag.ExitOnError)
var QoSKPIRequestNoConfigFlagSet = pflag.NewFlagSet("QoSKPIRequestNoConfig", pflag.ExitOnError)
var ServingStatusStrings = []string{
	"Unknown",
	"Serving",
	"NotServing",
}

func QoSKPIRequestSlicer(in *distributed_match_engine.QoSKPIRequest) []string {
	s := make([]string, 0, 2)
	s = append(s, strconv.FormatUint(uint64(in.Requestid), 10))
	if in.Requests == nil {
		in.Requests = make([]*distributed_match_engine.PositionKpiRequest, 1)
	}
	if in.Requests[0] == nil {
		in.Requests[0] = &distributed_match_engine.PositionKpiRequest{}
	}
	s = append(s, strconv.FormatUint(uint64(in.Requests[0].Positionid), 10))
	s = append(s, strconv.FormatFloat(float64(in.Requests[0].Latitude), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.Requests[0].Longitude), 'e', -1, 32))
	s = append(s, strconv.FormatUint(uint64(in.Requests[0].Timestamp), 10))
	s = append(s, strconv.FormatFloat(float64(in.Requests[0].Altitude), 'e', -1, 32))
	return s
}

func QoSKPIRequestHeaderSlicer() []string {
	s := make([]string, 0, 2)
	s = append(s, "Requestid")
	s = append(s, "Requests-Positionid")
	s = append(s, "Requests-Latitude")
	s = append(s, "Requests-Longitude")
	s = append(s, "Requests-Timestamp")
	s = append(s, "Requests-Altitude")
	return s
}

func QoSKPIRequestWriteOutputArray(objs []*distributed_match_engine.QoSKPIRequest) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(QoSKPIRequestHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(QoSKPIRequestSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func QoSKPIRequestWriteOutputOne(obj *distributed_match_engine.QoSKPIRequest) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(QoSKPIRequestHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(QoSKPIRequestSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}
func PositionKpiResultSlicer(in *distributed_match_engine.PositionKpiResult) []string {
	s := make([]string, 0, 10)
	s = append(s, strconv.FormatUint(uint64(in.Positionid), 10))
	return s
}

func PositionKpiResultHeaderSlicer() []string {
	s := make([]string, 0, 10)
	s = append(s, "Positionid")
	return s
}

func PositionKpiResultWriteOutputArray(objs []*distributed_match_engine.PositionKpiResult) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(PositionKpiResultHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(PositionKpiResultSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func PositionKpiResultWriteOutputOne(obj *distributed_match_engine.PositionKpiResult) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(PositionKpiResultHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(PositionKpiResultSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}
func PositionKpiRequestSlicer(in *distributed_match_engine.PositionKpiRequest) []string {
	s := make([]string, 0, 5)
	s = append(s, strconv.FormatUint(uint64(in.Positionid), 10))
	s = append(s, strconv.FormatFloat(float64(in.Latitude), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.Longitude), 'e', -1, 32))
	s = append(s, strconv.FormatUint(uint64(in.Timestamp), 10))
	s = append(s, strconv.FormatFloat(float64(in.Altitude), 'e', -1, 32))
	return s
}

func PositionKpiRequestHeaderSlicer() []string {
	s := make([]string, 0, 5)
	s = append(s, "Positionid")
	s = append(s, "Latitude")
	s = append(s, "Longitude")
	s = append(s, "Timestamp")
	s = append(s, "Altitude")
	return s
}

func PositionKpiRequestWriteOutputArray(objs []*distributed_match_engine.PositionKpiRequest) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(PositionKpiRequestHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(PositionKpiRequestSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func PositionKpiRequestWriteOutputOne(obj *distributed_match_engine.PositionKpiRequest) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(PositionKpiRequestHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(PositionKpiRequestSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}
func QoSKPIResponseSlicer(in *distributed_match_engine.QoSKPIResponse) []string {
	s := make([]string, 0, 2)
	s = append(s, strconv.FormatUint(uint64(in.Requestid), 10))
	if in.Results == nil {
		in.Results = make([]*distributed_match_engine.PositionKpiResult, 1)
	}
	if in.Results[0] == nil {
		in.Results[0] = &distributed_match_engine.PositionKpiResult{}
	}
	s = append(s, strconv.FormatUint(uint64(in.Results[0].Positionid), 10))
	return s
}

func QoSKPIResponseHeaderSlicer() []string {
	s := make([]string, 0, 2)
	s = append(s, "Requestid")
	s = append(s, "Results-Positionid")
	return s
}

func QoSKPIResponseWriteOutputArray(objs []*distributed_match_engine.QoSKPIResponse) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(QoSKPIResponseHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(QoSKPIResponseSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func QoSKPIResponseWriteOutputOne(obj *distributed_match_engine.QoSKPIResponse) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(QoSKPIResponseHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(QoSKPIResponseSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}
func HealthCheckRequestSlicer(in *distributed_match_engine.HealthCheckRequest) []string {
	s := make([]string, 0, 1)
	s = append(s, in.Service)
	return s
}

func HealthCheckRequestHeaderSlicer() []string {
	s := make([]string, 0, 1)
	s = append(s, "Service")
	return s
}

func HealthCheckRequestWriteOutputArray(objs []*distributed_match_engine.HealthCheckRequest) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(HealthCheckRequestHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(HealthCheckRequestSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func HealthCheckRequestWriteOutputOne(obj *distributed_match_engine.HealthCheckRequest) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(HealthCheckRequestHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(HealthCheckRequestSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}
func HealthCheckResponseSlicer(in *distributed_match_engine.HealthCheckResponse) []string {
	s := make([]string, 0, 3)
	s = append(s, distributed_match_engine.HealthCheckResponse_ServingStatus_CamelName[int32(in.Status)])
	s = append(s, strconv.FormatUint(uint64(in.Errorcode), 10))
	s = append(s, in.Modelversion)
	return s
}

func HealthCheckResponseHeaderSlicer() []string {
	s := make([]string, 0, 3)
	s = append(s, "Status")
	s = append(s, "Errorcode")
	s = append(s, "Modelversion")
	return s
}

func HealthCheckResponseWriteOutputArray(objs []*distributed_match_engine.HealthCheckResponse) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(HealthCheckResponseHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(HealthCheckResponseSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func HealthCheckResponseWriteOutputOne(obj *distributed_match_engine.HealthCheckResponse) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(HealthCheckResponseHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(HealthCheckResponseSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}

var QueryQoSKPICmd = &cobra.Command{
	Use: "QueryQoSKPI",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if we got this far, usage has been met.
		cmd.SilenceUsage = true
		return QueryQoSKPI(&QoSKPIRequestIn)
	},
}

func QueryQoSKPI(in *distributed_match_engine.QoSKPIRequest) error {
	if QueryQoSCmd == nil {
		return fmt.Errorf("QueryQoS client not initialized")
	}
	ctx := context.Background()
	stream, err := QueryQoSCmd.QueryQoSKPI(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("QueryQoSKPI failed: %s", errstr)
	}
	objs := make([]*distributed_match_engine.QoSKPIResponse, 0)
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("QueryQoSKPI recv failed: %s", err.Error())
		}
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	QoSKPIResponseWriteOutputArray(objs)
	return nil
}

func QueryQoSKPIs(data []distributed_match_engine.QoSKPIRequest, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("QueryQoSKPI %v\n", data[ii])
		myerr := QueryQoSKPI(&data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var QueryQoSCmds = []*cobra.Command{
	QueryQoSKPICmd,
}

var CheckCmd = &cobra.Command{
	Use: "Check",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if we got this far, usage has been met.
		cmd.SilenceUsage = true
		return Check(&HealthCheckRequestIn)
	},
}

func Check(in *distributed_match_engine.HealthCheckRequest) error {
	if HealthCmd == nil {
		return fmt.Errorf("Health client not initialized")
	}
	ctx := context.Background()
	obj, err := HealthCmd.Check(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("Check failed: %s", errstr)
	}
	HealthCheckResponseWriteOutputOne(obj)
	return nil
}

func Checks(data []distributed_match_engine.HealthCheckRequest, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("Check %v\n", data[ii])
		myerr := Check(&data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}

var HealthCmds = []*cobra.Command{
	CheckCmd,
}

func init() {
	QoSKPIRequestFlagSet.Int64Var(&QoSKPIRequestIn.Requestid, "requestid", 0, "Requestid")
	HealthCheckRequestFlagSet.StringVar(&HealthCheckRequestIn.Service, "service", "", "Service")
	QueryQoSKPICmd.Flags().AddFlagSet(QoSKPIRequestFlagSet)
	CheckCmd.Flags().AddFlagSet(HealthCheckRequestFlagSet)
}

func QueryQoSAllowNoConfig() {
	QueryQoSKPICmd.Flags().AddFlagSet(QoSKPIRequestNoConfigFlagSet)
}

func HealthAllowNoConfig() {
	CheckCmd.Flags().AddFlagSet(HealthCheckRequestNoConfigFlagSet)
}
