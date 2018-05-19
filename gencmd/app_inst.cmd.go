// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app_inst.proto

package gencmd

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import google_protobuf "github.com/gogo/protobuf/types"
import "strings"
import "time"
import "strconv"
import "github.com/spf13/cobra"
import "context"
import "os"
import "io"
import "text/tabwriter"
import "github.com/spf13/pflag"
import "errors"
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
var AppInstApiCmd edgeproto.AppInstApiClient
var AppInstIn edgeproto.AppInst
var AppInstFlagSet = pflag.NewFlagSet("AppInst", pflag.ExitOnError)
var AppInstInLiveness string

func AppInstKeySlicer(in *edgeproto.AppInstKey) []string {
	s := make([]string, 0, 3)
	s = append(s, in.AppKey.DeveloperKey.Name)
	s = append(s, in.AppKey.Name)
	s = append(s, in.AppKey.Version)
	s = append(s, in.CloudletKey.OperatorKey.Name)
	s = append(s, in.CloudletKey.Name)
	s = append(s, string(in.Id))
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

func AppInstSlicer(in *edgeproto.AppInst) []string {
	s := make([]string, 0, 6)
	s = append(s, in.Key.AppKey.DeveloperKey.Name)
	s = append(s, in.Key.AppKey.Name)
	s = append(s, in.Key.AppKey.Version)
	s = append(s, in.Key.CloudletKey.OperatorKey.Name)
	s = append(s, in.Key.CloudletKey.Name)
	s = append(s, string(in.Key.Id))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLoc.Lat), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLoc.Long), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLoc.HorizontalAccuracy), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLoc.VerticalAccuracy), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLoc.Altitude), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLoc.Course), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLoc.Speed), 'e', -1, 32))
	if in.CloudletLoc.Timestamp == nil {
		in.CloudletLoc.Timestamp = &google_protobuf.Timestamp{}
	}
	timestampTime := time.Unix(in.CloudletLoc.Timestamp.Seconds, int64(in.CloudletLoc.Timestamp.Nanos))
	s = append(s, timestampTime.String())
	s = append(s, string(in.Port))
	s = append(s, string(in.Liveness))
	return s
}

func AppInstHeaderSlicer() []string {
	s := make([]string, 0, 6)
	s = append(s, "Key-AppKey-DeveloperKey-Name")
	s = append(s, "Key-AppKey-Name")
	s = append(s, "Key-AppKey-Version")
	s = append(s, "Key-CloudletKey-OperatorKey-Name")
	s = append(s, "Key-CloudletKey-Name")
	s = append(s, "Key-Id")
	s = append(s, "CloudletLoc-Lat")
	s = append(s, "CloudletLoc-Long")
	s = append(s, "CloudletLoc-HorizontalAccuracy")
	s = append(s, "CloudletLoc-VerticalAccuracy")
	s = append(s, "CloudletLoc-Altitude")
	s = append(s, "CloudletLoc-Course")
	s = append(s, "CloudletLoc-Speed")
	s = append(s, "CloudletLoc-Timestamp")
	s = append(s, "Port")
	s = append(s, "Liveness")
	return s
}

var CreateAppInstCmd = &cobra.Command{
	Use: "CreateAppInst",
	Run: func(cmd *cobra.Command, args []string) {
		if AppInstApiCmd == nil {
			fmt.Println("AppInstApi client not initialized")
			return
		}
		var err error
		err = parseAppInstEnums()
		if err != nil {
			fmt.Println("CreateAppInst: ", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		out, err := AppInstApiCmd.CreateAppInst(ctx, &AppInstIn)
		cancel()
		if err != nil {
			fmt.Println("CreateAppInst failed: ", err)
		} else {
			headers := ResultHeaderSlicer()
			data := ResultSlicer(out)
			for ii := 0; ii < len(headers) && ii < len(data); ii++ {
				fmt.Println(headers[ii] + ": " + data[ii])
			}
		}
	},
}

var DeleteAppInstCmd = &cobra.Command{
	Use: "DeleteAppInst",
	Run: func(cmd *cobra.Command, args []string) {
		if AppInstApiCmd == nil {
			fmt.Println("AppInstApi client not initialized")
			return
		}
		var err error
		err = parseAppInstEnums()
		if err != nil {
			fmt.Println("DeleteAppInst: ", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		out, err := AppInstApiCmd.DeleteAppInst(ctx, &AppInstIn)
		cancel()
		if err != nil {
			fmt.Println("DeleteAppInst failed: ", err)
		} else {
			headers := ResultHeaderSlicer()
			data := ResultSlicer(out)
			for ii := 0; ii < len(headers) && ii < len(data); ii++ {
				fmt.Println(headers[ii] + ": " + data[ii])
			}
		}
	},
}

var UpdateAppInstCmd = &cobra.Command{
	Use: "UpdateAppInst",
	Run: func(cmd *cobra.Command, args []string) {
		if AppInstApiCmd == nil {
			fmt.Println("AppInstApi client not initialized")
			return
		}
		var err error
		err = parseAppInstEnums()
		if err != nil {
			fmt.Println("UpdateAppInst: ", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		out, err := AppInstApiCmd.UpdateAppInst(ctx, &AppInstIn)
		cancel()
		if err != nil {
			fmt.Println("UpdateAppInst failed: ", err)
		} else {
			headers := ResultHeaderSlicer()
			data := ResultSlicer(out)
			for ii := 0; ii < len(headers) && ii < len(data); ii++ {
				fmt.Println(headers[ii] + ": " + data[ii])
			}
		}
	},
}

var ShowAppInstCmd = &cobra.Command{
	Use: "ShowAppInst",
	Run: func(cmd *cobra.Command, args []string) {
		if AppInstApiCmd == nil {
			fmt.Println("AppInstApi client not initialized")
			return
		}
		var err error
		err = parseAppInstEnums()
		if err != nil {
			fmt.Println("ShowAppInst: ", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		count := 0
		fmt.Fprintln(output, strings.Join(AppInstHeaderSlicer(), "\t"))
		defer cancel()
		stream, err := AppInstApiCmd.ShowAppInst(ctx, &AppInstIn)
		if err != nil {
			fmt.Println("ShowAppInst failed: ", err)
			return
		}
		for {
			obj, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("ShowAppInst recv failed: ", err)
				break
			}
			fmt.Fprintln(output, strings.Join(AppInstSlicer(obj), "\t"))
			count++
		}
		if count > 0 {
			output.Flush()
		}
	},
}

func init() {
	AppInstFlagSet.StringVar(&AppInstIn.Key.AppKey.DeveloperKey.Name, "key-appkey-developerkey-name", "", "Key.AppKey.DeveloperKey.Name")
	AppInstFlagSet.StringVar(&AppInstIn.Key.AppKey.Name, "key-appkey-name", "", "Key.AppKey.Name")
	AppInstFlagSet.StringVar(&AppInstIn.Key.AppKey.Version, "key-appkey-version", "", "Key.AppKey.Version")
	AppInstFlagSet.StringVar(&AppInstIn.Key.CloudletKey.OperatorKey.Name, "key-cloudletkey-operatorkey-name", "", "Key.CloudletKey.OperatorKey.Name")
	AppInstFlagSet.StringVar(&AppInstIn.Key.CloudletKey.Name, "key-cloudletkey-name", "", "Key.CloudletKey.Name")
	AppInstFlagSet.Uint64Var(&AppInstIn.Key.Id, "key-id", 0, "Key.Id")
	AppInstFlagSet.Float64Var(&AppInstIn.CloudletLoc.Lat, "cloudletloc-lat", 0, "CloudletLoc.Lat")
	AppInstFlagSet.Float64Var(&AppInstIn.CloudletLoc.Long, "cloudletloc-long", 0, "CloudletLoc.Long")
	AppInstFlagSet.Float64Var(&AppInstIn.CloudletLoc.HorizontalAccuracy, "cloudletloc-horizontalaccuracy", 0, "CloudletLoc.HorizontalAccuracy")
	AppInstFlagSet.Float64Var(&AppInstIn.CloudletLoc.VerticalAccuracy, "cloudletloc-verticalaccuracy", 0, "CloudletLoc.VerticalAccuracy")
	AppInstFlagSet.Float64Var(&AppInstIn.CloudletLoc.Altitude, "cloudletloc-altitude", 0, "CloudletLoc.Altitude")
	AppInstFlagSet.Float64Var(&AppInstIn.CloudletLoc.Course, "cloudletloc-course", 0, "CloudletLoc.Course")
	AppInstFlagSet.Float64Var(&AppInstIn.CloudletLoc.Speed, "cloudletloc-speed", 0, "CloudletLoc.Speed")
	AppInstIn.CloudletLoc.Timestamp = &google_protobuf.Timestamp{}
	AppInstFlagSet.Int64Var(&AppInstIn.CloudletLoc.Timestamp.Seconds, "cloudletloc-timestamp-seconds", 0, "CloudletLoc.Timestamp.Seconds")
	AppInstFlagSet.Int32Var(&AppInstIn.CloudletLoc.Timestamp.Nanos, "cloudletloc-timestamp-nanos", 0, "CloudletLoc.Timestamp.Nanos")
	AppInstFlagSet.BytesHexVar(&AppInstIn.Ip, "ip", nil, "Ip")
	AppInstFlagSet.Uint32Var(&AppInstIn.Port, "port", 0, "Port")
	AppInstFlagSet.StringVar(&AppInstInLiveness, "liveness", "", "AppInstInLiveness")
	CreateAppInstCmd.Flags().AddFlagSet(AppInstFlagSet)
	DeleteAppInstCmd.Flags().AddFlagSet(AppInstFlagSet)
	UpdateAppInstCmd.Flags().AddFlagSet(AppInstFlagSet)
	ShowAppInstCmd.Flags().AddFlagSet(AppInstFlagSet)
}

func parseAppInstEnums() error {
	if AppInstInLiveness != "" {
		switch AppInstInLiveness {
		case "unknown":
			AppInstIn.Liveness = edgeproto.AppInst_Liveness(0)
		case "static":
			AppInstIn.Liveness = edgeproto.AppInst_Liveness(1)
		case "dynamic":
			AppInstIn.Liveness = edgeproto.AppInst_Liveness(2)
		default:
			return errors.New("Invalid value for AppInstInLiveness")
		}
	}
	return nil
}
