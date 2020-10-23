// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: common.proto

package edgeproto

import (
	"encoding/json"
	"errors"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	"github.com/mobiledgex/edge-cloud/util"
	io "io"
	math "math"
	math_bits "math/bits"
	"strconv"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Liveness Options
//
// Liveness indicates if an object was created statically via an external API call, or dynamically via an internal algorithm.
type Liveness int32

const (
	// Unknown liveness
	Liveness_LIVENESS_UNKNOWN Liveness = 0
	// Object managed by external entity
	Liveness_LIVENESS_STATIC Liveness = 1
	// Object managed internally
	Liveness_LIVENESS_DYNAMIC Liveness = 2
	// Object created by Auto Provisioning, treated like Static except when deleting App
	Liveness_LIVENESS_AUTOPROV Liveness = 3
)

var Liveness_name = map[int32]string{
	0: "LIVENESS_UNKNOWN",
	1: "LIVENESS_STATIC",
	2: "LIVENESS_DYNAMIC",
	3: "LIVENESS_AUTOPROV",
}

var Liveness_value = map[string]int32{
	"LIVENESS_UNKNOWN":  0,
	"LIVENESS_STATIC":   1,
	"LIVENESS_DYNAMIC":  2,
	"LIVENESS_AUTOPROV": 3,
}

func (x Liveness) String() string {
	return proto.EnumName(Liveness_name, int32(x))
}

func (Liveness) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_555bd8c177793206, []int{0}
}

// Type of public IP support
//
// Static IP support indicates a set of static public IPs are available for use, and managed by the Controller. Dynamic indicates the Cloudlet uses a DHCP server to provide public IP addresses, and the controller has no control over which IPs are assigned.
type IpSupport int32

const (
	// Unknown IP support
	IpSupport_IP_SUPPORT_UNKNOWN IpSupport = 0
	// Static IP addresses are provided to and managed by Controller
	IpSupport_IP_SUPPORT_STATIC IpSupport = 1
	// IP addresses are dynamically provided by an Operator's DHCP server
	IpSupport_IP_SUPPORT_DYNAMIC IpSupport = 2
)

var IpSupport_name = map[int32]string{
	0: "IP_SUPPORT_UNKNOWN",
	1: "IP_SUPPORT_STATIC",
	2: "IP_SUPPORT_DYNAMIC",
}

var IpSupport_value = map[string]int32{
	"IP_SUPPORT_UNKNOWN": 0,
	"IP_SUPPORT_STATIC":  1,
	"IP_SUPPORT_DYNAMIC": 2,
}

func (x IpSupport) String() string {
	return proto.EnumName(IpSupport_name, int32(x))
}

func (IpSupport) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_555bd8c177793206, []int{1}
}

// IpAccess Options
//
// IpAccess indicates the type of RootLB that Developer requires for their App
type IpAccess int32

const (
	// Unknown IP access
	IpAccess_IP_ACCESS_UNKNOWN IpAccess = 0
	// Dedicated RootLB
	IpAccess_IP_ACCESS_DEDICATED IpAccess = 1
	// Shared RootLB
	IpAccess_IP_ACCESS_SHARED IpAccess = 3
)

var IpAccess_name = map[int32]string{
	0: "IP_ACCESS_UNKNOWN",
	1: "IP_ACCESS_DEDICATED",
	3: "IP_ACCESS_SHARED",
}

var IpAccess_value = map[string]int32{
	"IP_ACCESS_UNKNOWN":   0,
	"IP_ACCESS_DEDICATED": 1,
	"IP_ACCESS_SHARED":    3,
}

func (x IpAccess) String() string {
	return proto.EnumName(IpAccess_name, int32(x))
}

func (IpAccess) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_555bd8c177793206, []int{2}
}

// Tracked States
//
// TrackedState is used to track the state of an object on a remote node,
// i.e. track the state of a ClusterInst object on the CRM (Cloudlet).
type TrackedState int32

const (
	// Unknown state
	TrackedState_TRACKED_STATE_UNKNOWN TrackedState = 0
	// Not present (does not exist)
	TrackedState_NOT_PRESENT TrackedState = 1
	// Create requested
	TrackedState_CREATE_REQUESTED TrackedState = 2
	// Creating
	TrackedState_CREATING TrackedState = 3
	// Create error
	TrackedState_CREATE_ERROR TrackedState = 4
	// Ready
	TrackedState_READY TrackedState = 5
	// Update requested
	TrackedState_UPDATE_REQUESTED TrackedState = 6
	// Updating
	TrackedState_UPDATING TrackedState = 7
	// Update error
	TrackedState_UPDATE_ERROR TrackedState = 8
	// Delete requested
	TrackedState_DELETE_REQUESTED TrackedState = 9
	// Deleting
	TrackedState_DELETING TrackedState = 10
	// Delete error
	TrackedState_DELETE_ERROR TrackedState = 11
	// Delete prepare (extra state used by controller to block other changes)
	TrackedState_DELETE_PREPARE TrackedState = 12
	// CRM INIT OK
	TrackedState_CRM_INITOK TrackedState = 13
	// Creating dependencies (state used to tracked dependent object change progress)
	TrackedState_CREATING_DEPENDENCIES TrackedState = 14
)

var TrackedState_name = map[int32]string{
	0:  "TRACKED_STATE_UNKNOWN",
	1:  "NOT_PRESENT",
	2:  "CREATE_REQUESTED",
	3:  "CREATING",
	4:  "CREATE_ERROR",
	5:  "READY",
	6:  "UPDATE_REQUESTED",
	7:  "UPDATING",
	8:  "UPDATE_ERROR",
	9:  "DELETE_REQUESTED",
	10: "DELETING",
	11: "DELETE_ERROR",
	12: "DELETE_PREPARE",
	13: "CRM_INITOK",
	14: "CREATING_DEPENDENCIES",
}

var TrackedState_value = map[string]int32{
	"TRACKED_STATE_UNKNOWN": 0,
	"NOT_PRESENT":           1,
	"CREATE_REQUESTED":      2,
	"CREATING":              3,
	"CREATE_ERROR":          4,
	"READY":                 5,
	"UPDATE_REQUESTED":      6,
	"UPDATING":              7,
	"UPDATE_ERROR":          8,
	"DELETE_REQUESTED":      9,
	"DELETING":              10,
	"DELETE_ERROR":          11,
	"DELETE_PREPARE":        12,
	"CRM_INITOK":            13,
	"CREATING_DEPENDENCIES": 14,
}

func (x TrackedState) String() string {
	return proto.EnumName(TrackedState_name, int32(x))
}

func (TrackedState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_555bd8c177793206, []int{3}
}

// Overrides default CRM behaviour
//
// CRMOverride can be applied to commands that issue requests to the CRM.
// It should only be used by administrators when bugs have caused the
// Controller and CRM to get out of sync. It allows commands from the
// Controller to ignore errors from the CRM, or ignore the CRM completely
// (messages will not be sent to CRM).
type CRMOverride int32

const (
	// No override
	CRMOverride_NO_OVERRIDE CRMOverride = 0
	// Ignore errors from CRM
	CRMOverride_IGNORE_CRM_ERRORS CRMOverride = 1
	// Ignore CRM completely (does not inform CRM of operation)
	CRMOverride_IGNORE_CRM CRMOverride = 2
	// Ignore Transient State (only admin should use if CRM crashed)
	CRMOverride_IGNORE_TRANSIENT_STATE CRMOverride = 3
	// Ignore CRM and Transient State
	CRMOverride_IGNORE_CRM_AND_TRANSIENT_STATE CRMOverride = 4
)

var CRMOverride_name = map[int32]string{
	0: "NO_OVERRIDE",
	1: "IGNORE_CRM_ERRORS",
	2: "IGNORE_CRM",
	3: "IGNORE_TRANSIENT_STATE",
	4: "IGNORE_CRM_AND_TRANSIENT_STATE",
}

var CRMOverride_value = map[string]int32{
	"NO_OVERRIDE":                    0,
	"IGNORE_CRM_ERRORS":              1,
	"IGNORE_CRM":                     2,
	"IGNORE_TRANSIENT_STATE":         3,
	"IGNORE_CRM_AND_TRANSIENT_STATE": 4,
}

func (x CRMOverride) String() string {
	return proto.EnumName(CRMOverride_name, int32(x))
}

func (CRMOverride) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_555bd8c177793206, []int{4}
}

// Cloudlet Maintenance States
//
// Maintenance allows for planned downtimes of Cloudlets.
// These states involve message exchanges between the Controller,
// the AutoProv service, and the CRM. Certain states are only set
// by certain actors.
type MaintenanceState int32

const (
	// Normal operational state
	MaintenanceState_NORMAL_OPERATION MaintenanceState = 0
	// Request start of maintenance
	MaintenanceState_MAINTENANCE_START MaintenanceState = 1
	// Trigger failover for any HA AppInsts
	MaintenanceState_FAILOVER_REQUESTED MaintenanceState = 2
	// Failover done
	MaintenanceState_FAILOVER_DONE MaintenanceState = 3
	// Some errors encountered during maintenance failover
	MaintenanceState_FAILOVER_ERROR MaintenanceState = 4
	// Request start of maintenance without AutoProv failover
	MaintenanceState_MAINTENANCE_START_NO_FAILOVER MaintenanceState = 5
	// Request CRM to transition to maintenance
	MaintenanceState_CRM_REQUESTED MaintenanceState = 6
	// CRM request done and under maintenance
	MaintenanceState_CRM_UNDER_MAINTENANCE MaintenanceState = 7
	// CRM failed to go into maintenance
	MaintenanceState_CRM_ERROR MaintenanceState = 8
	// Request CRM to transition to normal operation
	MaintenanceState_NORMAL_OPERATION_INIT MaintenanceState = 9
	// Under maintenance
	MaintenanceState_UNDER_MAINTENANCE MaintenanceState = 31
)

var MaintenanceState_name = map[int32]string{
	0:  "NORMAL_OPERATION",
	1:  "MAINTENANCE_START",
	2:  "FAILOVER_REQUESTED",
	3:  "FAILOVER_DONE",
	4:  "FAILOVER_ERROR",
	5:  "MAINTENANCE_START_NO_FAILOVER",
	6:  "CRM_REQUESTED",
	7:  "CRM_UNDER_MAINTENANCE",
	8:  "CRM_ERROR",
	9:  "NORMAL_OPERATION_INIT",
	31: "UNDER_MAINTENANCE",
}

var MaintenanceState_value = map[string]int32{
	"NORMAL_OPERATION":              0,
	"MAINTENANCE_START":             1,
	"FAILOVER_REQUESTED":            2,
	"FAILOVER_DONE":                 3,
	"FAILOVER_ERROR":                4,
	"MAINTENANCE_START_NO_FAILOVER": 5,
	"CRM_REQUESTED":                 6,
	"CRM_UNDER_MAINTENANCE":         7,
	"CRM_ERROR":                     8,
	"NORMAL_OPERATION_INIT":         9,
	"UNDER_MAINTENANCE":             31,
}

func (x MaintenanceState) String() string {
	return proto.EnumName(MaintenanceState_name, int32(x))
}

func (MaintenanceState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_555bd8c177793206, []int{5}
}

// Status Information
//
// Used to track status of create/delete/update for resources that are being modified
// by the controller via the CRM.  Tasks are the high level jobs that are to be completed.
// Steps are work items within a task. Within the clusterinst and appinst objects this
// is converted to a string
type StatusInfo struct {
	TaskNumber uint32 `protobuf:"varint,1,opt,name=task_number,json=taskNumber,proto3" json:"task_number,omitempty"`
	MaxTasks   uint32 `protobuf:"varint,2,opt,name=max_tasks,json=maxTasks,proto3" json:"max_tasks,omitempty"`
	TaskName   string `protobuf:"bytes,3,opt,name=task_name,json=taskName,proto3" json:"task_name,omitempty"`
	StepName   string `protobuf:"bytes,4,opt,name=step_name,json=stepName,proto3" json:"step_name,omitempty"`
}

func (m *StatusInfo) Reset()         { *m = StatusInfo{} }
func (m *StatusInfo) String() string { return proto.CompactTextString(m) }
func (*StatusInfo) ProtoMessage()    {}
func (*StatusInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_555bd8c177793206, []int{0}
}
func (m *StatusInfo) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *StatusInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_StatusInfo.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *StatusInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StatusInfo.Merge(m, src)
}
func (m *StatusInfo) XXX_Size() int {
	return m.Size()
}
func (m *StatusInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_StatusInfo.DiscardUnknown(m)
}

var xxx_messageInfo_StatusInfo proto.InternalMessageInfo

func init() {
	proto.RegisterEnum("edgeproto.Liveness", Liveness_name, Liveness_value)
	proto.RegisterEnum("edgeproto.IpSupport", IpSupport_name, IpSupport_value)
	proto.RegisterEnum("edgeproto.IpAccess", IpAccess_name, IpAccess_value)
	proto.RegisterEnum("edgeproto.TrackedState", TrackedState_name, TrackedState_value)
	proto.RegisterEnum("edgeproto.CRMOverride", CRMOverride_name, CRMOverride_value)
	proto.RegisterEnum("edgeproto.MaintenanceState", MaintenanceState_name, MaintenanceState_value)
	proto.RegisterType((*StatusInfo)(nil), "edgeproto.StatusInfo")
}

func init() { proto.RegisterFile("common.proto", fileDescriptor_555bd8c177793206) }

var fileDescriptor_555bd8c177793206 = []byte{
	// 739 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x54, 0xbf, 0x72, 0xe2, 0x46,
	0x18, 0x47, 0x80, 0xef, 0xd0, 0x07, 0xe6, 0xd6, 0x6b, 0xfb, 0x8e, 0x70, 0x17, 0xf9, 0x72, 0xd5,
	0x0d, 0x33, 0x31, 0x45, 0x9a, 0xb4, 0x1b, 0xe9, 0x8b, 0xb3, 0x63, 0x58, 0x29, 0x2b, 0xe1, 0x8c,
	0x2b, 0x8d, 0x00, 0x85, 0x30, 0xb6, 0x24, 0x06, 0x84, 0xc7, 0x0f, 0x90, 0xf4, 0x7e, 0x8a, 0x3c,
	0x47, 0x4a, 0x97, 0x2e, 0x53, 0x26, 0xf6, 0x3b, 0xa4, 0xce, 0xac, 0x24, 0x0b, 0x38, 0x37, 0x9e,
	0xd5, 0xef, 0xdf, 0xf7, 0xed, 0xfe, 0x3c, 0x40, 0x6b, 0x92, 0x44, 0x51, 0x12, 0x9f, 0x2e, 0x96,
	0x49, 0x9a, 0x50, 0x3d, 0x9c, 0xce, 0xc2, 0xec, 0xd8, 0xfd, 0x7e, 0x36, 0x4f, 0x7f, 0x5b, 0x8f,
	0x4f, 0x27, 0x49, 0xd4, 0x8f, 0x92, 0xf1, 0xfc, 0x5a, 0x51, 0xb7, 0x7d, 0xf5, 0xf7, 0xdb, 0xc9,
	0x75, 0xb2, 0x9e, 0xf6, 0x33, 0xdd, 0x2c, 0x8c, 0xcb, 0x43, 0x1e, 0xd2, 0x3d, 0x9a, 0x25, 0xb3,
	0x24, 0x3b, 0xf6, 0xd5, 0x29, 0x47, 0x3f, 0xfd, 0xae, 0x01, 0xb8, 0x69, 0x90, 0xae, 0x57, 0x3c,
	0xfe, 0x35, 0xa1, 0x27, 0xd0, 0x4c, 0x83, 0xd5, 0x95, 0x1f, 0xaf, 0xa3, 0x71, 0xb8, 0xec, 0x68,
	0x1f, 0xb5, 0xcf, 0xfb, 0x12, 0x14, 0x24, 0x32, 0x84, 0xbe, 0x07, 0x3d, 0x0a, 0x6e, 0x7d, 0x85,
	0xac, 0x3a, 0xd5, 0x8c, 0x6e, 0x44, 0xc1, 0xad, 0xa7, 0xbe, 0x15, 0x99, 0xbb, 0x83, 0x28, 0xec,
	0xd4, 0x3e, 0x6a, 0x9f, 0x75, 0xd9, 0xc8, 0xbc, 0x41, 0x14, 0x2a, 0x72, 0x95, 0x86, 0x8b, 0x9c,
	0xac, 0xe7, 0xa4, 0x02, 0x14, 0xd9, 0x1b, 0x43, 0x63, 0x30, 0xbf, 0x09, 0xe3, 0x70, 0xb5, 0xa2,
	0x47, 0x40, 0x06, 0xfc, 0x02, 0x05, 0xba, 0xae, 0x3f, 0x12, 0xe7, 0xc2, 0xfe, 0x45, 0x90, 0x0a,
	0x3d, 0x84, 0x37, 0x25, 0xea, 0x7a, 0xcc, 0xe3, 0x26, 0xd1, 0x76, 0xa4, 0xd6, 0xa5, 0x60, 0x43,
	0x6e, 0x92, 0x2a, 0x3d, 0x86, 0x83, 0x12, 0x65, 0x23, 0xcf, 0x76, 0xa4, 0x7d, 0x41, 0x6a, 0x3d,
	0x09, 0x3a, 0x5f, 0xb8, 0xeb, 0xc5, 0x22, 0x59, 0xa6, 0xf4, 0x2d, 0x50, 0xee, 0xf8, 0xee, 0xc8,
	0x71, 0x6c, 0xe9, 0x6d, 0x8d, 0x39, 0x86, 0x83, 0x2d, 0xbc, 0x1c, 0xb4, 0x2b, 0x2f, 0x47, 0xf5,
	0x1c, 0x68, 0xf0, 0x05, 0x9b, 0x4c, 0xd4, 0xde, 0xb9, 0x95, 0x99, 0xe6, 0xee, 0xe2, 0xef, 0xe0,
	0x70, 0x03, 0x5b, 0x68, 0x71, 0x93, 0x79, 0x68, 0xe5, 0xcb, 0x6f, 0x08, 0xf7, 0x27, 0x26, 0xd1,
	0x22, 0xb5, 0xde, 0x9f, 0x55, 0x68, 0x79, 0xcb, 0x60, 0x72, 0x15, 0x4e, 0x55, 0x2f, 0x21, 0xfd,
	0x0a, 0x8e, 0x3d, 0xc9, 0xcc, 0x73, 0xb4, 0xb2, 0x75, 0x70, 0x2b, 0xfa, 0x0d, 0x34, 0x85, 0xed,
	0xf9, 0x8e, 0x44, 0x17, 0x85, 0x97, 0x47, 0x9a, 0x12, 0x95, 0x48, 0xe2, 0xcf, 0x23, 0x74, 0xd5,
	0xa0, 0x2a, 0x6d, 0x41, 0x23, 0x43, 0xb9, 0x38, 0x23, 0x35, 0x4a, 0xa0, 0x55, 0x68, 0x50, 0x4a,
	0x5b, 0x92, 0x3a, 0xd5, 0x61, 0x4f, 0x22, 0xb3, 0x2e, 0xc9, 0x9e, 0x0a, 0x18, 0x39, 0xd6, 0x6e,
	0xc0, 0x2b, 0x15, 0x90, 0xa1, 0x2a, 0xe0, 0xb5, 0x0a, 0x28, 0x34, 0x79, 0x40, 0x43, 0xb9, 0x2c,
	0x1c, 0xe0, 0x8e, 0x4b, 0x57, 0xae, 0x0c, 0x55, 0x2e, 0x50, 0xae, 0x42, 0x93, 0xbb, 0x9a, 0x94,
	0x42, 0xbb, 0x40, 0x1c, 0x89, 0x0e, 0x93, 0x48, 0x5a, 0xb4, 0x0d, 0x60, 0xca, 0xa1, 0xcf, 0x05,
	0xf7, 0xec, 0x73, 0xb2, 0xaf, 0x2e, 0xff, 0xbc, 0xba, 0x6f, 0xa1, 0x83, 0xc2, 0x42, 0x61, 0x72,
	0x74, 0x49, 0xbb, 0xf7, 0x87, 0x06, 0x4d, 0x53, 0x0e, 0xed, 0x9b, 0x70, 0xb9, 0x9c, 0x4f, 0xc3,
	0xfc, 0x31, 0x7c, 0xfb, 0x02, 0xa5, 0xe4, 0x16, 0x16, 0x55, 0x9e, 0x09, 0x5b, 0xa2, 0xaf, 0x22,
	0xb3, 0xa9, 0x2e, 0xd1, 0xd4, 0x88, 0x0d, 0x4c, 0xaa, 0xb4, 0x0b, 0x6f, 0x8b, 0x6f, 0x4f, 0x32,
	0xe1, 0x72, 0x14, 0x79, 0xef, 0x48, 0x6a, 0xf4, 0x13, 0x18, 0x5b, 0x11, 0x4c, 0x58, 0x2f, 0x34,
	0xf5, 0xde, 0x5f, 0x55, 0x20, 0xc3, 0x60, 0x1e, 0xa7, 0x61, 0x1c, 0xc4, 0x93, 0x30, 0x2f, 0xed,
	0x08, 0x88, 0xb0, 0xe5, 0x90, 0x0d, 0x7c, 0xdb, 0x41, 0xc9, 0x3c, 0x6e, 0x17, 0xff, 0x5c, 0x43,
	0xc6, 0x85, 0x87, 0x82, 0x09, 0x13, 0x55, 0x82, 0x54, 0xad, 0x7d, 0x00, 0xfa, 0x23, 0xe3, 0x03,
	0xb5, 0xfa, 0x76, 0x6f, 0xdd, 0xfa, 0xdd, 0x7f, 0x1d, 0x8d, 0xbe, 0x83, 0xfd, 0x92, 0xb5, 0x6c,
	0x81, 0xa4, 0x56, 0x10, 0x1d, 0x68, 0x97, 0x44, 0x51, 0x65, 0xc1, 0x7c, 0x03, 0x5f, 0xbf, 0x98,
	0xe3, 0x0b, 0xdb, 0x7f, 0x96, 0x93, 0x3d, 0x95, 0xaa, 0xae, 0xb4, 0xd5, 0x72, 0xe1, 0x3d, 0x51,
	0x2f, 0x3e, 0xf4, 0x47, 0xc2, 0x42, 0xe9, 0x6f, 0xa5, 0x90, 0xd7, 0x85, 0xe0, 0x10, 0xf4, 0xf2,
	0x3d, 0x49, 0x63, 0xe3, 0xfa, 0xf2, 0xbe, 0x59, 0x89, 0x44, 0x2f, 0x04, 0xef, 0xe1, 0xe0, 0x65,
	0xe4, 0x49, 0x4e, 0xfe, 0xf0, 0xe1, 0xfe, 0x5f, 0xa3, 0x72, 0xff, 0x68, 0x68, 0x0f, 0x8f, 0x86,
	0xf6, 0xcf, 0xa3, 0xa1, 0xdd, 0x3d, 0x19, 0x95, 0x87, 0x27, 0xa3, 0xf2, 0xf7, 0x93, 0x51, 0x19,
	0xbf, 0xca, 0x7e, 0xa9, 0xbe, 0xfb, 0x3f, 0x00, 0x00, 0xff, 0xff, 0xc0, 0x27, 0x00, 0x95, 0x14,
	0x05, 0x00, 0x00,
}

func (m *StatusInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *StatusInfo) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *StatusInfo) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.StepName) > 0 {
		i -= len(m.StepName)
		copy(dAtA[i:], m.StepName)
		i = encodeVarintCommon(dAtA, i, uint64(len(m.StepName)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.TaskName) > 0 {
		i -= len(m.TaskName)
		copy(dAtA[i:], m.TaskName)
		i = encodeVarintCommon(dAtA, i, uint64(len(m.TaskName)))
		i--
		dAtA[i] = 0x1a
	}
	if m.MaxTasks != 0 {
		i = encodeVarintCommon(dAtA, i, uint64(m.MaxTasks))
		i--
		dAtA[i] = 0x10
	}
	if m.TaskNumber != 0 {
		i = encodeVarintCommon(dAtA, i, uint64(m.TaskNumber))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintCommon(dAtA []byte, offset int, v uint64) int {
	offset -= sovCommon(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *StatusInfo) CopyInFields(src *StatusInfo) int {
	changed := 0
	if m.TaskNumber != src.TaskNumber {
		m.TaskNumber = src.TaskNumber
		changed++
	}
	if m.MaxTasks != src.MaxTasks {
		m.MaxTasks = src.MaxTasks
		changed++
	}
	if m.TaskName != src.TaskName {
		m.TaskName = src.TaskName
		changed++
	}
	if m.StepName != src.StepName {
		m.StepName = src.StepName
		changed++
	}
	return changed
}

func (m *StatusInfo) DeepCopyIn(src *StatusInfo) {
	m.TaskNumber = src.TaskNumber
	m.MaxTasks = src.MaxTasks
	m.TaskName = src.TaskName
	m.StepName = src.StepName
}

// Helper method to check that enums have valid values
func (m *StatusInfo) ValidateEnums() error {
	return nil
}

var LivenessStrings = []string{
	"LIVENESS_UNKNOWN",
	"LIVENESS_STATIC",
	"LIVENESS_DYNAMIC",
	"LIVENESS_AUTOPROV",
}

const (
	LivenessLIVENESS_UNKNOWN  uint64 = 1 << 0
	LivenessLIVENESS_STATIC   uint64 = 1 << 1
	LivenessLIVENESS_DYNAMIC  uint64 = 1 << 2
	LivenessLIVENESS_AUTOPROV uint64 = 1 << 3
)

var Liveness_CamelName = map[int32]string{
	// LIVENESS_UNKNOWN -> LivenessUnknown
	0: "LivenessUnknown",
	// LIVENESS_STATIC -> LivenessStatic
	1: "LivenessStatic",
	// LIVENESS_DYNAMIC -> LivenessDynamic
	2: "LivenessDynamic",
	// LIVENESS_AUTOPROV -> LivenessAutoprov
	3: "LivenessAutoprov",
}
var Liveness_CamelValue = map[string]int32{
	"LivenessUnknown":  0,
	"LivenessStatic":   1,
	"LivenessDynamic":  2,
	"LivenessAutoprov": 3,
}

func (e *Liveness) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	val, ok := Liveness_CamelValue[util.CamelCase(str)]
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = Liveness_CamelName[val]
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("No enum value for %s", str))
	}
	*e = Liveness(val)
	return nil
}

func (e Liveness) MarshalYAML() (interface{}, error) {
	return proto.EnumName(Liveness_CamelName, int32(e)), nil
}

// custom JSON encoding/decoding
func (e *Liveness) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		val, ok := Liveness_CamelValue[util.CamelCase(str)]
		if !ok {
			// may be int value instead of enum name
			ival, err := strconv.Atoi(str)
			val = int32(ival)
			if err == nil {
				_, ok = Liveness_CamelName[val]
			}
		}
		if !ok {
			return errors.New(fmt.Sprintf("No enum value for %s", str))
		}
		*e = Liveness(val)
		return nil
	}
	var val int32
	err = json.Unmarshal(b, &val)
	if err == nil {
		*e = Liveness(val)
		return nil
	}
	return fmt.Errorf("No enum value for %v", b)
}

var IpSupportStrings = []string{
	"IP_SUPPORT_UNKNOWN",
	"IP_SUPPORT_STATIC",
	"IP_SUPPORT_DYNAMIC",
}

const (
	IpSupportIP_SUPPORT_UNKNOWN uint64 = 1 << 0
	IpSupportIP_SUPPORT_STATIC  uint64 = 1 << 1
	IpSupportIP_SUPPORT_DYNAMIC uint64 = 1 << 2
)

var IpSupport_CamelName = map[int32]string{
	// IP_SUPPORT_UNKNOWN -> IpSupportUnknown
	0: "IpSupportUnknown",
	// IP_SUPPORT_STATIC -> IpSupportStatic
	1: "IpSupportStatic",
	// IP_SUPPORT_DYNAMIC -> IpSupportDynamic
	2: "IpSupportDynamic",
}
var IpSupport_CamelValue = map[string]int32{
	"IpSupportUnknown": 0,
	"IpSupportStatic":  1,
	"IpSupportDynamic": 2,
}

func (e *IpSupport) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	val, ok := IpSupport_CamelValue[util.CamelCase(str)]
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = IpSupport_CamelName[val]
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("No enum value for %s", str))
	}
	*e = IpSupport(val)
	return nil
}

func (e IpSupport) MarshalYAML() (interface{}, error) {
	return proto.EnumName(IpSupport_CamelName, int32(e)), nil
}

// custom JSON encoding/decoding
func (e *IpSupport) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		val, ok := IpSupport_CamelValue[util.CamelCase(str)]
		if !ok {
			// may be int value instead of enum name
			ival, err := strconv.Atoi(str)
			val = int32(ival)
			if err == nil {
				_, ok = IpSupport_CamelName[val]
			}
		}
		if !ok {
			return errors.New(fmt.Sprintf("No enum value for %s", str))
		}
		*e = IpSupport(val)
		return nil
	}
	var val int32
	err = json.Unmarshal(b, &val)
	if err == nil {
		*e = IpSupport(val)
		return nil
	}
	return fmt.Errorf("No enum value for %v", b)
}

var IpAccessStrings = []string{
	"IP_ACCESS_UNKNOWN",
	"IP_ACCESS_DEDICATED",
	"IP_ACCESS_SHARED",
}

const (
	IpAccessIP_ACCESS_UNKNOWN   uint64 = 1 << 0
	IpAccessIP_ACCESS_DEDICATED uint64 = 1 << 1
	IpAccessIP_ACCESS_SHARED    uint64 = 1 << 2
)

var IpAccess_CamelName = map[int32]string{
	// IP_ACCESS_UNKNOWN -> IpAccessUnknown
	0: "IpAccessUnknown",
	// IP_ACCESS_DEDICATED -> IpAccessDedicated
	1: "IpAccessDedicated",
	// IP_ACCESS_SHARED -> IpAccessShared
	3: "IpAccessShared",
}
var IpAccess_CamelValue = map[string]int32{
	"IpAccessUnknown":   0,
	"IpAccessDedicated": 1,
	"IpAccessShared":    3,
}

func (e *IpAccess) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	val, ok := IpAccess_CamelValue[util.CamelCase(str)]
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = IpAccess_CamelName[val]
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("No enum value for %s", str))
	}
	*e = IpAccess(val)
	return nil
}

func (e IpAccess) MarshalYAML() (interface{}, error) {
	return proto.EnumName(IpAccess_CamelName, int32(e)), nil
}

// custom JSON encoding/decoding
func (e *IpAccess) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		val, ok := IpAccess_CamelValue[util.CamelCase(str)]
		if !ok {
			// may be int value instead of enum name
			ival, err := strconv.Atoi(str)
			val = int32(ival)
			if err == nil {
				_, ok = IpAccess_CamelName[val]
			}
		}
		if !ok {
			return errors.New(fmt.Sprintf("No enum value for %s", str))
		}
		*e = IpAccess(val)
		return nil
	}
	var val int32
	err = json.Unmarshal(b, &val)
	if err == nil {
		*e = IpAccess(val)
		return nil
	}
	return fmt.Errorf("No enum value for %v", b)
}

var TrackedStateStrings = []string{
	"TRACKED_STATE_UNKNOWN",
	"NOT_PRESENT",
	"CREATE_REQUESTED",
	"CREATING",
	"CREATE_ERROR",
	"READY",
	"UPDATE_REQUESTED",
	"UPDATING",
	"UPDATE_ERROR",
	"DELETE_REQUESTED",
	"DELETING",
	"DELETE_ERROR",
	"DELETE_PREPARE",
	"CRM_INITOK",
	"CREATING_DEPENDENCIES",
}

const (
	TrackedStateTRACKED_STATE_UNKNOWN uint64 = 1 << 0
	TrackedStateNOT_PRESENT           uint64 = 1 << 1
	TrackedStateCREATE_REQUESTED      uint64 = 1 << 2
	TrackedStateCREATING              uint64 = 1 << 3
	TrackedStateCREATE_ERROR          uint64 = 1 << 4
	TrackedStateREADY                 uint64 = 1 << 5
	TrackedStateUPDATE_REQUESTED      uint64 = 1 << 6
	TrackedStateUPDATING              uint64 = 1 << 7
	TrackedStateUPDATE_ERROR          uint64 = 1 << 8
	TrackedStateDELETE_REQUESTED      uint64 = 1 << 9
	TrackedStateDELETING              uint64 = 1 << 10
	TrackedStateDELETE_ERROR          uint64 = 1 << 11
	TrackedStateDELETE_PREPARE        uint64 = 1 << 12
	TrackedStateCRM_INITOK            uint64 = 1 << 13
	TrackedStateCREATING_DEPENDENCIES uint64 = 1 << 14
)

var TrackedState_CamelName = map[int32]string{
	// TRACKED_STATE_UNKNOWN -> TrackedStateUnknown
	0: "TrackedStateUnknown",
	// NOT_PRESENT -> NotPresent
	1: "NotPresent",
	// CREATE_REQUESTED -> CreateRequested
	2: "CreateRequested",
	// CREATING -> Creating
	3: "Creating",
	// CREATE_ERROR -> CreateError
	4: "CreateError",
	// READY -> Ready
	5: "Ready",
	// UPDATE_REQUESTED -> UpdateRequested
	6: "UpdateRequested",
	// UPDATING -> Updating
	7: "Updating",
	// UPDATE_ERROR -> UpdateError
	8: "UpdateError",
	// DELETE_REQUESTED -> DeleteRequested
	9: "DeleteRequested",
	// DELETING -> Deleting
	10: "Deleting",
	// DELETE_ERROR -> DeleteError
	11: "DeleteError",
	// DELETE_PREPARE -> DeletePrepare
	12: "DeletePrepare",
	// CRM_INITOK -> CrmInitok
	13: "CrmInitok",
	// CREATING_DEPENDENCIES -> CreatingDependencies
	14: "CreatingDependencies",
}
var TrackedState_CamelValue = map[string]int32{
	"TrackedStateUnknown":  0,
	"NotPresent":           1,
	"CreateRequested":      2,
	"Creating":             3,
	"CreateError":          4,
	"Ready":                5,
	"UpdateRequested":      6,
	"Updating":             7,
	"UpdateError":          8,
	"DeleteRequested":      9,
	"Deleting":             10,
	"DeleteError":          11,
	"DeletePrepare":        12,
	"CrmInitok":            13,
	"CreatingDependencies": 14,
}

func (e *TrackedState) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	val, ok := TrackedState_CamelValue[util.CamelCase(str)]
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = TrackedState_CamelName[val]
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("No enum value for %s", str))
	}
	*e = TrackedState(val)
	return nil
}

func (e TrackedState) MarshalYAML() (interface{}, error) {
	return proto.EnumName(TrackedState_CamelName, int32(e)), nil
}

// custom JSON encoding/decoding
func (e *TrackedState) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		val, ok := TrackedState_CamelValue[util.CamelCase(str)]
		if !ok {
			// may be int value instead of enum name
			ival, err := strconv.Atoi(str)
			val = int32(ival)
			if err == nil {
				_, ok = TrackedState_CamelName[val]
			}
		}
		if !ok {
			return errors.New(fmt.Sprintf("No enum value for %s", str))
		}
		*e = TrackedState(val)
		return nil
	}
	var val int32
	err = json.Unmarshal(b, &val)
	if err == nil {
		*e = TrackedState(val)
		return nil
	}
	return fmt.Errorf("No enum value for %v", b)
}

var CRMOverrideStrings = []string{
	"NO_OVERRIDE",
	"IGNORE_CRM_ERRORS",
	"IGNORE_CRM",
	"IGNORE_TRANSIENT_STATE",
	"IGNORE_CRM_AND_TRANSIENT_STATE",
}

const (
	CRMOverrideNO_OVERRIDE                    uint64 = 1 << 0
	CRMOverrideIGNORE_CRM_ERRORS              uint64 = 1 << 1
	CRMOverrideIGNORE_CRM                     uint64 = 1 << 2
	CRMOverrideIGNORE_TRANSIENT_STATE         uint64 = 1 << 3
	CRMOverrideIGNORE_CRM_AND_TRANSIENT_STATE uint64 = 1 << 4
)

var CRMOverride_CamelName = map[int32]string{
	// NO_OVERRIDE -> NoOverride
	0: "NoOverride",
	// IGNORE_CRM_ERRORS -> IgnoreCrmErrors
	1: "IgnoreCrmErrors",
	// IGNORE_CRM -> IgnoreCrm
	2: "IgnoreCrm",
	// IGNORE_TRANSIENT_STATE -> IgnoreTransientState
	3: "IgnoreTransientState",
	// IGNORE_CRM_AND_TRANSIENT_STATE -> IgnoreCrmAndTransientState
	4: "IgnoreCrmAndTransientState",
}
var CRMOverride_CamelValue = map[string]int32{
	"NoOverride":                 0,
	"IgnoreCrmErrors":            1,
	"IgnoreCrm":                  2,
	"IgnoreTransientState":       3,
	"IgnoreCrmAndTransientState": 4,
}

func (e *CRMOverride) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	val, ok := CRMOverride_CamelValue[util.CamelCase(str)]
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = CRMOverride_CamelName[val]
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("No enum value for %s", str))
	}
	*e = CRMOverride(val)
	return nil
}

func (e CRMOverride) MarshalYAML() (interface{}, error) {
	return proto.EnumName(CRMOverride_CamelName, int32(e)), nil
}

// custom JSON encoding/decoding
func (e *CRMOverride) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		val, ok := CRMOverride_CamelValue[util.CamelCase(str)]
		if !ok {
			// may be int value instead of enum name
			ival, err := strconv.Atoi(str)
			val = int32(ival)
			if err == nil {
				_, ok = CRMOverride_CamelName[val]
			}
		}
		if !ok {
			return errors.New(fmt.Sprintf("No enum value for %s", str))
		}
		*e = CRMOverride(val)
		return nil
	}
	var val int32
	err = json.Unmarshal(b, &val)
	if err == nil {
		*e = CRMOverride(val)
		return nil
	}
	return fmt.Errorf("No enum value for %v", b)
}

var MaintenanceStateStrings = []string{
	"NORMAL_OPERATION",
	"MAINTENANCE_START",
	"FAILOVER_REQUESTED",
	"FAILOVER_DONE",
	"FAILOVER_ERROR",
	"MAINTENANCE_START_NO_FAILOVER",
	"CRM_REQUESTED",
	"CRM_UNDER_MAINTENANCE",
	"CRM_ERROR",
	"NORMAL_OPERATION_INIT",
	"UNDER_MAINTENANCE",
}

const (
	MaintenanceStateNORMAL_OPERATION              uint64 = 1 << 0
	MaintenanceStateMAINTENANCE_START             uint64 = 1 << 1
	MaintenanceStateFAILOVER_REQUESTED            uint64 = 1 << 2
	MaintenanceStateFAILOVER_DONE                 uint64 = 1 << 3
	MaintenanceStateFAILOVER_ERROR                uint64 = 1 << 4
	MaintenanceStateMAINTENANCE_START_NO_FAILOVER uint64 = 1 << 5
	MaintenanceStateCRM_REQUESTED                 uint64 = 1 << 6
	MaintenanceStateCRM_UNDER_MAINTENANCE         uint64 = 1 << 7
	MaintenanceStateCRM_ERROR                     uint64 = 1 << 8
	MaintenanceStateNORMAL_OPERATION_INIT         uint64 = 1 << 9
	MaintenanceStateUNDER_MAINTENANCE             uint64 = 1 << 10
)

var MaintenanceState_CamelName = map[int32]string{
	// NORMAL_OPERATION -> NormalOperation
	0: "NormalOperation",
	// MAINTENANCE_START -> MaintenanceStart
	1: "MaintenanceStart",
	// FAILOVER_REQUESTED -> FailoverRequested
	2: "FailoverRequested",
	// FAILOVER_DONE -> FailoverDone
	3: "FailoverDone",
	// FAILOVER_ERROR -> FailoverError
	4: "FailoverError",
	// MAINTENANCE_START_NO_FAILOVER -> MaintenanceStartNoFailover
	5: "MaintenanceStartNoFailover",
	// CRM_REQUESTED -> CrmRequested
	6: "CrmRequested",
	// CRM_UNDER_MAINTENANCE -> CrmUnderMaintenance
	7: "CrmUnderMaintenance",
	// CRM_ERROR -> CrmError
	8: "CrmError",
	// NORMAL_OPERATION_INIT -> NormalOperationInit
	9: "NormalOperationInit",
	// UNDER_MAINTENANCE -> UnderMaintenance
	31: "UnderMaintenance",
}
var MaintenanceState_CamelValue = map[string]int32{
	"NormalOperation":            0,
	"MaintenanceStart":           1,
	"FailoverRequested":          2,
	"FailoverDone":               3,
	"FailoverError":              4,
	"MaintenanceStartNoFailover": 5,
	"CrmRequested":               6,
	"CrmUnderMaintenance":        7,
	"CrmError":                   8,
	"NormalOperationInit":        9,
	"UnderMaintenance":           31,
}

func (e *MaintenanceState) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	val, ok := MaintenanceState_CamelValue[util.CamelCase(str)]
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = MaintenanceState_CamelName[val]
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("No enum value for %s", str))
	}
	*e = MaintenanceState(val)
	return nil
}

func (e MaintenanceState) MarshalYAML() (interface{}, error) {
	return proto.EnumName(MaintenanceState_CamelName, int32(e)), nil
}

// custom JSON encoding/decoding
func (e *MaintenanceState) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		val, ok := MaintenanceState_CamelValue[util.CamelCase(str)]
		if !ok {
			// may be int value instead of enum name
			ival, err := strconv.Atoi(str)
			val = int32(ival)
			if err == nil {
				_, ok = MaintenanceState_CamelName[val]
			}
		}
		if !ok {
			return errors.New(fmt.Sprintf("No enum value for %s", str))
		}
		*e = MaintenanceState(val)
		return nil
	}
	var val int32
	err = json.Unmarshal(b, &val)
	if err == nil {
		*e = MaintenanceState(val)
		return nil
	}
	return fmt.Errorf("No enum value for %v", b)
}
func (m *StatusInfo) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.TaskNumber != 0 {
		n += 1 + sovCommon(uint64(m.TaskNumber))
	}
	if m.MaxTasks != 0 {
		n += 1 + sovCommon(uint64(m.MaxTasks))
	}
	l = len(m.TaskName)
	if l > 0 {
		n += 1 + l + sovCommon(uint64(l))
	}
	l = len(m.StepName)
	if l > 0 {
		n += 1 + l + sovCommon(uint64(l))
	}
	return n
}

func sovCommon(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozCommon(x uint64) (n int) {
	return sovCommon(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *StatusInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCommon
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: StatusInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: StatusInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TaskNumber", wireType)
			}
			m.TaskNumber = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TaskNumber |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxTasks", wireType)
			}
			m.MaxTasks = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxTasks |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TaskName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCommon
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCommon
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TaskName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StepName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCommon
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCommon
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.StepName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCommon(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCommon
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthCommon
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipCommon(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCommon
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCommon
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthCommon
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupCommon
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthCommon
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthCommon        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCommon          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupCommon = fmt.Errorf("proto: unexpected end of group")
)
