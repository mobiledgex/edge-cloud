// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: common.proto

package gencmd

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
import "strconv"
import "os"
import "text/tabwriter"
import "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/cmdsup"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
var LivenessStrings = []string{
	"LivenessUnknown",
	"LivenessStatic",
	"LivenessDynamic",
}

var IpSupportStrings = []string{
	"IpSupportUnknown",
	"IpSupportStatic",
	"IpSupportDynamic",
}

var IpAccessStrings = []string{
	"IpAccessUnknown",
	"IpAccessDedicated",
	"IpAccessDedicatedOrShared",
	"IpAccessShared",
}

var TrackedStateStrings = []string{
	"TrackedStateUnknown",
	"NotPresent",
	"CreateRequested",
	"Creating",
	"CreateError",
	"Ready",
	"UpdateRequested",
	"Updating",
	"UpdateError",
	"DeleteRequested",
	"Deleting",
	"DeleteError",
	"DeletePrepare",
}

var CRMOverrideStrings = []string{
	"NoOverride",
	"IgnoreCrmErrors",
	"IgnoreCrm",
	"IgnoreTransientState",
	"IgnoreCrmAndTransientState",
}

func StatusInfoSlicer(in *edgeproto.StatusInfo) []string {
	s := make([]string, 0, 3)
	s = append(s, strconv.FormatUint(uint64(in.TaskNumber), 10))
	s = append(s, in.TaskName)
	s = append(s, in.StepName)
	return s
}

func StatusInfoHeaderSlicer() []string {
	s := make([]string, 0, 3)
	s = append(s, "TaskNumber")
	s = append(s, "TaskName")
	s = append(s, "StepName")
	return s
}

func StatusInfoWriteOutputArray(objs []*edgeproto.StatusInfo) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(StatusInfoHeaderSlicer(), "\t"))
		for _, obj := range objs {
			fmt.Fprintln(output, strings.Join(StatusInfoSlicer(obj), "\t"))
		}
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(objs)
	}
}

func StatusInfoWriteOutputOne(obj *edgeproto.StatusInfo) {
	if cmdsup.OutputFormat == cmdsup.OutputFormatTable {
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(output, strings.Join(StatusInfoHeaderSlicer(), "\t"))
		fmt.Fprintln(output, strings.Join(StatusInfoSlicer(obj), "\t"))
		output.Flush()
	} else {
		cmdsup.WriteOutputGeneric(obj)
	}
}
func init() {
}
