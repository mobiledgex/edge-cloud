// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app-client.proto

/*
Package gencmd is a generated protocol buffer package.

It is generated from these files:
	app-client.proto

It has these top-level messages:
	Match_Engine_Request
	Match_Engine_Reply
	Match_Engine_Loc_Verify
*/
package gencmd

import distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import google_protobuf "github.com/gogo/protobuf/types"
import "time"
import "strconv"
import "github.com/spf13/cobra"
import "context"
import "github.com/spf13/pflag"
import "errors"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/mobiledgex/edge-cloud/edgeproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
var Match_Engine_ApiCmd distributed_match_engine.Match_Engine_ApiClient
var Match_Engine_RequestIn distributed_match_engine.Match_Engine_Request
var Match_Engine_RequestFlagSet = pflag.NewFlagSet("Match_Engine_Request", pflag.ExitOnError)
var Match_Engine_RequestInIdType string

func Match_Engine_RequestSlicer(in *distributed_match_engine.Match_Engine_Request) []string {
	s := make([]string, 0, 12)
	s = append(s, strconv.FormatUint(uint64(in.Ver), 10))
	s = append(s, distributed_match_engine.Match_Engine_Request_IDType_name[int32(in.IdType)])
	s = append(s, in.Id)
	s = append(s, strconv.FormatUint(uint64(in.Carrier), 10))
	s = append(s, strconv.FormatUint(uint64(in.Tower), 10))
	if in.GpsLocation == nil {
		in.GpsLocation = &edgeproto.Loc{}
	}
	s = append(s, strconv.FormatFloat(float64(in.GpsLocation.Lat), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.GpsLocation.Long), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.GpsLocation.HorizontalAccuracy), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.GpsLocation.VerticalAccuracy), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.GpsLocation.Altitude), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.GpsLocation.Course), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.GpsLocation.Speed), 'e', -1, 32))
	if in.GpsLocation.Timestamp == nil {
		in.GpsLocation.Timestamp = &google_protobuf.Timestamp{}
	}
	_GpsLocation_TimestampTime := time.Unix(in.GpsLocation.Timestamp.Seconds, int64(in.GpsLocation.Timestamp.Nanos))
	s = append(s, _GpsLocation_TimestampTime.String())
	s = append(s, strconv.FormatUint(uint64(in.AppId), 10))
	s = append(s, "")
	for _, b := range in.Protocol {
		s[len(s)-1] += fmt.Sprintf("%v", b)
	}
	s = append(s, "")
	for _, b := range in.ServerPort {
		s[len(s)-1] += fmt.Sprintf("%v", b)
	}
	s = append(s, in.DevName)
	s = append(s, in.AppName)
	s = append(s, in.AppVers)
	return s
}

func Match_Engine_RequestHeaderSlicer() []string {
	s := make([]string, 0, 12)
	s = append(s, "Ver")
	s = append(s, "IdType")
	s = append(s, "Id")
	s = append(s, "Carrier")
	s = append(s, "Tower")
	s = append(s, "GpsLocation-Lat")
	s = append(s, "GpsLocation-Long")
	s = append(s, "GpsLocation-HorizontalAccuracy")
	s = append(s, "GpsLocation-VerticalAccuracy")
	s = append(s, "GpsLocation-Altitude")
	s = append(s, "GpsLocation-Course")
	s = append(s, "GpsLocation-Speed")
	s = append(s, "GpsLocation-Timestamp")
	s = append(s, "AppId")
	s = append(s, "Protocol")
	s = append(s, "ServerPort")
	s = append(s, "DevName")
	s = append(s, "AppName")
	s = append(s, "AppVers")
	return s
}

func Match_Engine_ReplySlicer(in *distributed_match_engine.Match_Engine_Reply) []string {
	s := make([]string, 0, 6)
	s = append(s, strconv.FormatUint(uint64(in.Ver), 10))
	s = append(s, in.Uri)
	s = append(s, "")
	for i, b := range in.ServiceIp {
		s[len(s)-1] += fmt.Sprintf("%v", b)
		if i < 3 {
			s[len(s)-1] += "."
		}
	}
	s = append(s, strconv.FormatUint(uint64(in.ServicePort), 10))
	if in.CloudletLocation == nil {
		in.CloudletLocation = &edgeproto.Loc{}
	}
	s = append(s, strconv.FormatFloat(float64(in.CloudletLocation.Lat), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLocation.Long), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLocation.HorizontalAccuracy), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLocation.VerticalAccuracy), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLocation.Altitude), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLocation.Course), 'e', -1, 32))
	s = append(s, strconv.FormatFloat(float64(in.CloudletLocation.Speed), 'e', -1, 32))
	if in.CloudletLocation.Timestamp == nil {
		in.CloudletLocation.Timestamp = &google_protobuf.Timestamp{}
	}
	_CloudletLocation_TimestampTime := time.Unix(in.CloudletLocation.Timestamp.Seconds, int64(in.CloudletLocation.Timestamp.Nanos))
	s = append(s, _CloudletLocation_TimestampTime.String())
	s = append(s, strconv.FormatBool(in.Status))
	return s
}

func Match_Engine_ReplyHeaderSlicer() []string {
	s := make([]string, 0, 6)
	s = append(s, "Ver")
	s = append(s, "Uri")
	s = append(s, "ServiceIp")
	s = append(s, "ServicePort")
	s = append(s, "CloudletLocation-Lat")
	s = append(s, "CloudletLocation-Long")
	s = append(s, "CloudletLocation-HorizontalAccuracy")
	s = append(s, "CloudletLocation-VerticalAccuracy")
	s = append(s, "CloudletLocation-Altitude")
	s = append(s, "CloudletLocation-Course")
	s = append(s, "CloudletLocation-Speed")
	s = append(s, "CloudletLocation-Timestamp")
	s = append(s, "Status")
	return s
}

func Match_Engine_Loc_VerifySlicer(in *distributed_match_engine.Match_Engine_Loc_Verify) []string {
	s := make([]string, 0, 3)
	s = append(s, strconv.FormatUint(uint64(in.Ver), 10))
	s = append(s, distributed_match_engine.Match_Engine_Loc_Verify_Tower_Status_name[int32(in.TowerStatus)])
	s = append(s, distributed_match_engine.Match_Engine_Loc_Verify_GPS_Location_Status_name[int32(in.GpsLocationStatus)])
	return s
}

func Match_Engine_Loc_VerifyHeaderSlicer() []string {
	s := make([]string, 0, 3)
	s = append(s, "Ver")
	s = append(s, "TowerStatus")
	s = append(s, "GpsLocationStatus")
	return s
}

var FindCloudletCmd = &cobra.Command{
	Use: "FindCloudlet",
	Run: func(cmd *cobra.Command, args []string) {
		if Match_Engine_ApiCmd == nil {
			fmt.Println("Match_Engine_Api client not initialized")
			return
		}
		var err error
		err = parseMatch_Engine_RequestEnums()
		if err != nil {
			fmt.Println("FindCloudlet: ", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		out, err := Match_Engine_ApiCmd.FindCloudlet(ctx, &Match_Engine_RequestIn)
		cancel()
		if err != nil {
			fmt.Println("FindCloudlet failed: ", err)
		} else {
			headers := Match_Engine_ReplyHeaderSlicer()
			data := Match_Engine_ReplySlicer(out)
			for ii := 0; ii < len(headers) && ii < len(data); ii++ {
				fmt.Println(headers[ii] + ": " + data[ii])
			}
		}
	},
}

var VerifyLocationCmd = &cobra.Command{
	Use: "VerifyLocation",
	Run: func(cmd *cobra.Command, args []string) {
		if Match_Engine_ApiCmd == nil {
			fmt.Println("Match_Engine_Api client not initialized")
			return
		}
		var err error
		err = parseMatch_Engine_RequestEnums()
		if err != nil {
			fmt.Println("VerifyLocation: ", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		out, err := Match_Engine_ApiCmd.VerifyLocation(ctx, &Match_Engine_RequestIn)
		cancel()
		if err != nil {
			fmt.Println("VerifyLocation failed: ", err)
		} else {
			headers := Match_Engine_Loc_VerifyHeaderSlicer()
			data := Match_Engine_Loc_VerifySlicer(out)
			for ii := 0; ii < len(headers) && ii < len(data); ii++ {
				fmt.Println(headers[ii] + ": " + data[ii])
			}
		}
	},
}

func init() {
	Match_Engine_RequestFlagSet.Uint32Var(&Match_Engine_RequestIn.Ver, "ver", 0, "Ver")
	Match_Engine_RequestFlagSet.StringVar(&Match_Engine_RequestInIdType, "idtype", "", "Match_Engine_RequestInIdType")
	Match_Engine_RequestFlagSet.StringVar(&Match_Engine_RequestIn.Id, "id", "", "Id")
	Match_Engine_RequestFlagSet.Uint64Var(&Match_Engine_RequestIn.Carrier, "carrier", 0, "Carrier")
	Match_Engine_RequestFlagSet.Uint64Var(&Match_Engine_RequestIn.Tower, "tower", 0, "Tower")
	Match_Engine_RequestIn.GpsLocation = &edgeproto.Loc{}
	Match_Engine_RequestFlagSet.Float64Var(&Match_Engine_RequestIn.GpsLocation.Lat, "gpslocation-lat", 0, "GpsLocation.Lat")
	Match_Engine_RequestFlagSet.Float64Var(&Match_Engine_RequestIn.GpsLocation.Long, "gpslocation-long", 0, "GpsLocation.Long")
	Match_Engine_RequestFlagSet.Float64Var(&Match_Engine_RequestIn.GpsLocation.HorizontalAccuracy, "gpslocation-horizontalaccuracy", 0, "GpsLocation.HorizontalAccuracy")
	Match_Engine_RequestFlagSet.Float64Var(&Match_Engine_RequestIn.GpsLocation.VerticalAccuracy, "gpslocation-verticalaccuracy", 0, "GpsLocation.VerticalAccuracy")
	Match_Engine_RequestFlagSet.Float64Var(&Match_Engine_RequestIn.GpsLocation.Altitude, "gpslocation-altitude", 0, "GpsLocation.Altitude")
	Match_Engine_RequestFlagSet.Float64Var(&Match_Engine_RequestIn.GpsLocation.Course, "gpslocation-course", 0, "GpsLocation.Course")
	Match_Engine_RequestFlagSet.Float64Var(&Match_Engine_RequestIn.GpsLocation.Speed, "gpslocation-speed", 0, "GpsLocation.Speed")
	Match_Engine_RequestIn.GpsLocation.Timestamp = &google_protobuf.Timestamp{}
	Match_Engine_RequestFlagSet.Int64Var(&Match_Engine_RequestIn.GpsLocation.Timestamp.Seconds, "gpslocation-timestamp-seconds", 0, "GpsLocation.Timestamp.Seconds")
	Match_Engine_RequestFlagSet.Int32Var(&Match_Engine_RequestIn.GpsLocation.Timestamp.Nanos, "gpslocation-timestamp-nanos", 0, "GpsLocation.Timestamp.Nanos")
	Match_Engine_RequestFlagSet.Uint64Var(&Match_Engine_RequestIn.AppId, "appid", 0, "AppId")
	Match_Engine_RequestFlagSet.BytesHexVar(&Match_Engine_RequestIn.Protocol, "protocol", nil, "Protocol")
	Match_Engine_RequestFlagSet.BytesHexVar(&Match_Engine_RequestIn.ServerPort, "serverport", nil, "ServerPort")
	Match_Engine_RequestFlagSet.StringVar(&Match_Engine_RequestIn.DevName, "devname", "", "DevName")
	Match_Engine_RequestFlagSet.StringVar(&Match_Engine_RequestIn.AppName, "appname", "", "AppName")
	Match_Engine_RequestFlagSet.StringVar(&Match_Engine_RequestIn.AppVers, "appvers", "", "AppVers")
	FindCloudletCmd.Flags().AddFlagSet(Match_Engine_RequestFlagSet)
	VerifyLocationCmd.Flags().AddFlagSet(Match_Engine_RequestFlagSet)
}

func parseMatch_Engine_RequestEnums() error {
	if Match_Engine_RequestInIdType != "" {
		switch Match_Engine_RequestInIdType {
		case "imei":
			Match_Engine_RequestIn.IdType = distributed_match_engine.Match_Engine_Request_IDType(0)
		case "msisdn":
			Match_Engine_RequestIn.IdType = distributed_match_engine.Match_Engine_Request_IDType(1)
		case "ipaddr":
			Match_Engine_RequestIn.IdType = distributed_match_engine.Match_Engine_Request_IDType(2)
		default:
			return errors.New("Invalid value for Match_Engine_RequestInIdType")
		}
	}
	return nil
}
