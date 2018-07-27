// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: notice.proto

package gencmd

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
import "strconv"
import "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/cmdsup"
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
var NotifyApiCmd edgeproto.NotifyApiClient
var NoticeActionStrings = []string{
	"NONE",
	"UPDATE",
	"DELETE",
	"VERSION",
	"SENDALL_END",
}

var NoticeRequestorStrings = []string{
	"NoticeRequestorNone",
	"NoticeRequestorDME",
	"NoticeRequestorCRM",
}

func NoticeReplySlicer(in *edgeproto.NoticeReply) []string {
	s := make([]string, 0, 6)
	s = append(s, edgeproto.NoticeAction_name[int32(in.Action)])
	s = append(s, strconv.FormatUint(uint64(in.Version), 10))
	return s
}

func NoticeReplyHeaderSlicer() []string {
	s := make([]string, 0, 6)
	s = append(s, "Action")
	s = append(s, "Version")
	return s
}

func NoticeRequestSlicer(in *edgeproto.NoticeRequest) []string {
	s := make([]string, 0, 7)
	s = append(s, edgeproto.NoticeAction_name[int32(in.Action)])
	s = append(s, strconv.FormatUint(uint64(in.Version), 10))
	s = append(s, edgeproto.NoticeRequestor_name[int32(in.Requestor)])
	s = append(s, strconv.FormatUint(uint64(in.Revision), 10))
	return s
}

func NoticeRequestHeaderSlicer() []string {
	s := make([]string, 0, 7)
	s = append(s, "Action")
	s = append(s, "Version")
	s = append(s, "Requestor")
	s = append(s, "Revision")
	return s
}

func NoticeRequestHideTags(in *edgeproto.NoticeRequest) {
	if cmdsup.HideTags == "" {
		return
	}
	tags := make(map[string]struct{})
	for _, tag := range strings.Split(cmdsup.HideTags, ",") {
		tags[tag] = struct{}{}
	}
}

func init() {
}

func NotifyApiAllowNoConfig() {
}
