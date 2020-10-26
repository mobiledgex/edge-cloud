// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: appcommon.proto

package distributed_match_engine

import (
	"encoding/json"
	"errors"
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
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

// LProto indicates which protocol to use for accessing an application on a particular port. This is required by Kubernetes for port mapping.
type LProto int32

const (
	// Unknown protocol
	LProto_L_PROTO_UNKNOWN LProto = 0
	// TCP (L4) protocol
	LProto_L_PROTO_TCP LProto = 1
	// UDP (L4) protocol
	LProto_L_PROTO_UDP LProto = 2
)

var LProto_name = map[int32]string{
	0: "L_PROTO_UNKNOWN",
	1: "L_PROTO_TCP",
	2: "L_PROTO_UDP",
}

var LProto_value = map[string]int32{
	"L_PROTO_UNKNOWN": 0,
	"L_PROTO_TCP":     1,
	"L_PROTO_UDP":     2,
}

func (x LProto) String() string {
	return proto.EnumName(LProto_name, int32(x))
}

func (LProto) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_fdc58d2114e550de, []int{0}
}

// Health check status
//
// Health check status gets set by external, or rootLB health check
type HealthCheck int32

const (
	// Health Check is unknown
	HealthCheck_HEALTH_CHECK_UNKNOWN HealthCheck = 0
	// Health Check failure due to RootLB being offline
	HealthCheck_HEALTH_CHECK_FAIL_ROOTLB_OFFLINE HealthCheck = 1
	// Health Check failure due to Backend server being unavailable
	HealthCheck_HEALTH_CHECK_FAIL_SERVER_FAIL HealthCheck = 2
	// Health Check is ok
	HealthCheck_HEALTH_CHECK_OK HealthCheck = 3
)

var HealthCheck_name = map[int32]string{
	0: "HEALTH_CHECK_UNKNOWN",
	1: "HEALTH_CHECK_FAIL_ROOTLB_OFFLINE",
	2: "HEALTH_CHECK_FAIL_SERVER_FAIL",
	3: "HEALTH_CHECK_OK",
}

var HealthCheck_value = map[string]int32{
	"HEALTH_CHECK_UNKNOWN":             0,
	"HEALTH_CHECK_FAIL_ROOTLB_OFFLINE": 1,
	"HEALTH_CHECK_FAIL_SERVER_FAIL":    2,
	"HEALTH_CHECK_OK":                  3,
}

func (x HealthCheck) String() string {
	return proto.EnumName(HealthCheck_name, int32(x))
}

func (HealthCheck) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_fdc58d2114e550de, []int{1}
}

// CloudletState is the state of the Cloudlet.
type CloudletState int32

const (
	// Unknown
	CloudletState_CLOUDLET_STATE_UNKNOWN CloudletState = 0
	// Create/Delete/Update encountered errors (see Errors field of CloudletInfo)
	CloudletState_CLOUDLET_STATE_ERRORS CloudletState = 1
	// Cloudlet is created and ready
	CloudletState_CLOUDLET_STATE_READY CloudletState = 2
	// Cloudlet is offline (unreachable)
	CloudletState_CLOUDLET_STATE_OFFLINE CloudletState = 3
	// Cloudlet is not present
	CloudletState_CLOUDLET_STATE_NOT_PRESENT CloudletState = 4
	// Cloudlet is initializing
	CloudletState_CLOUDLET_STATE_INIT CloudletState = 5
	// Cloudlet is upgrading
	CloudletState_CLOUDLET_STATE_UPGRADE CloudletState = 6
	// Cloudlet needs data to synchronize
	CloudletState_CLOUDLET_STATE_NEED_SYNC CloudletState = 7
)

var CloudletState_name = map[int32]string{
	0: "CLOUDLET_STATE_UNKNOWN",
	1: "CLOUDLET_STATE_ERRORS",
	2: "CLOUDLET_STATE_READY",
	3: "CLOUDLET_STATE_OFFLINE",
	4: "CLOUDLET_STATE_NOT_PRESENT",
	5: "CLOUDLET_STATE_INIT",
	6: "CLOUDLET_STATE_UPGRADE",
	7: "CLOUDLET_STATE_NEED_SYNC",
}

var CloudletState_value = map[string]int32{
	"CLOUDLET_STATE_UNKNOWN":     0,
	"CLOUDLET_STATE_ERRORS":      1,
	"CLOUDLET_STATE_READY":       2,
	"CLOUDLET_STATE_OFFLINE":     3,
	"CLOUDLET_STATE_NOT_PRESENT": 4,
	"CLOUDLET_STATE_INIT":        5,
	"CLOUDLET_STATE_UPGRADE":     6,
	"CLOUDLET_STATE_NEED_SYNC":   7,
}

func (x CloudletState) String() string {
	return proto.EnumName(CloudletState_name, int32(x))
}

func (CloudletState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_fdc58d2114e550de, []int{2}
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
	return fileDescriptor_fdc58d2114e550de, []int{3}
}

// Application Port
//
// AppPort describes an L4 or L7 public access port/path mapping. This is used to track external to internal mappings for access via a shared load balancer or reverse proxy.
type AppPort struct {
	// TCP (L4) or UDP (L4) protocol
	Proto LProto `protobuf:"varint,1,opt,name=proto,proto3,enum=distributed_match_engine.LProto" json:"proto,omitempty"`
	// Container port
	InternalPort int32 `protobuf:"varint,2,opt,name=internal_port,json=internalPort,proto3" json:"internal_port,omitempty"`
	// Public facing port for TCP/UDP (may be mapped on shared LB reverse proxy)
	PublicPort int32 `protobuf:"varint,3,opt,name=public_port,json=publicPort,proto3" json:"public_port,omitempty"`
	// skip 4 to preserve the numbering. 4 was path_prefix but was removed since we dont need it after removed http
	// FQDN prefix to append to base FQDN in FindCloudlet response. May be empty.
	FqdnPrefix string `protobuf:"bytes,5,opt,name=fqdn_prefix,json=fqdnPrefix,proto3" json:"fqdn_prefix,omitempty"`
	// A non-zero end port indicates a port range from internal port to end port, inclusive.
	EndPort int32 `protobuf:"varint,6,opt,name=end_port,json=endPort,proto3" json:"end_port,omitempty"`
	// TLS termination for this port
	Tls bool `protobuf:"varint,7,opt,name=tls,proto3" json:"tls,omitempty"`
	// use nginx proxy for this port if you really need a transparent proxy (udp only)
	Nginx                bool     `protobuf:"varint,8,opt,name=nginx,proto3" json:"nginx,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AppPort) Reset()         { *m = AppPort{} }
func (m *AppPort) String() string { return proto.CompactTextString(m) }
func (*AppPort) ProtoMessage()    {}
func (*AppPort) Descriptor() ([]byte, []int) {
	return fileDescriptor_fdc58d2114e550de, []int{0}
}
func (m *AppPort) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *AppPort) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_AppPort.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *AppPort) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AppPort.Merge(m, src)
}
func (m *AppPort) XXX_Size() int {
	return m.Size()
}
func (m *AppPort) XXX_DiscardUnknown() {
	xxx_messageInfo_AppPort.DiscardUnknown(m)
}

var xxx_messageInfo_AppPort proto.InternalMessageInfo

var E_EnumBackend = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.EnumValueOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51042,
	Name:          "distributed_match_engine.enum_backend",
	Tag:           "varint,51042,opt,name=enum_backend",
	Filename:      "appcommon.proto",
}

func init() {
	proto.RegisterEnum("distributed_match_engine.LProto", LProto_name, LProto_value)
	proto.RegisterEnum("distributed_match_engine.HealthCheck", HealthCheck_name, HealthCheck_value)
	proto.RegisterEnum("distributed_match_engine.CloudletState", CloudletState_name, CloudletState_value)
	proto.RegisterEnum("distributed_match_engine.MaintenanceState", MaintenanceState_name, MaintenanceState_value)
	proto.RegisterType((*AppPort)(nil), "distributed_match_engine.AppPort")
	proto.RegisterExtension(E_EnumBackend)
}

func init() { proto.RegisterFile("appcommon.proto", fileDescriptor_fdc58d2114e550de) }

var fileDescriptor_fdc58d2114e550de = []byte{
	// 692 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x54, 0x4d, 0x6f, 0xe2, 0x46,
	0x18, 0xde, 0x81, 0xf0, 0x91, 0x97, 0xcd, 0x66, 0x76, 0x92, 0x74, 0xbd, 0x74, 0x97, 0x90, 0xb6,
	0x07, 0xc4, 0x81, 0x95, 0x5a, 0xa9, 0x87, 0x5e, 0x2a, 0xc7, 0x1e, 0x0a, 0x8a, 0xf1, 0xb8, 0x83,
	0xd9, 0x2a, 0xa7, 0x91, 0xc1, 0x93, 0xc4, 0x8a, 0xb1, 0x5d, 0x63, 0xa4, 0xfc, 0x81, 0x5e, 0xab,
	0xfe, 0xa4, 0x1e, 0x73, 0xec, 0x4f, 0x68, 0xb9, 0xf6, 0xd0, 0x5b, 0xcf, 0xd5, 0xd8, 0x90, 0x12,
	0x92, 0xbd, 0x0d, 0xcf, 0xd7, 0x3c, 0xef, 0x2b, 0x3c, 0x70, 0xe8, 0x25, 0xc9, 0x2c, 0x9e, 0xcf,
	0xe3, 0xa8, 0x97, 0xa4, 0x71, 0x16, 0x13, 0xcd, 0x0f, 0x16, 0x59, 0x1a, 0x4c, 0x97, 0x99, 0xf4,
	0xc5, 0xdc, 0xcb, 0x66, 0x37, 0x42, 0x46, 0xd7, 0x41, 0x24, 0x9b, 0xed, 0xeb, 0x38, 0xbe, 0x0e,
	0xe5, 0x87, 0x5c, 0x37, 0x5d, 0x5e, 0x7d, 0xf0, 0xe5, 0x62, 0x96, 0x06, 0x49, 0x16, 0xa7, 0x85,
	0xf7, 0x8b, 0xbf, 0x11, 0xd4, 0xf4, 0x24, 0x71, 0xe2, 0x34, 0x23, 0xdf, 0x42, 0x25, 0x07, 0x35,
	0xd4, 0x46, 0x9d, 0x57, 0x5f, 0xb7, 0x7b, 0x9f, 0xca, 0xed, 0x59, 0x8e, 0xd2, 0xf1, 0x42, 0x4e,
	0xbe, 0x84, 0x83, 0x20, 0xca, 0x64, 0x1a, 0x79, 0xa1, 0x48, 0xe2, 0x34, 0xd3, 0x4a, 0x6d, 0xd4,
	0xa9, 0xf0, 0x97, 0x1b, 0x30, 0x0f, 0x3f, 0x85, 0x46, 0xb2, 0x9c, 0x86, 0xc1, 0xac, 0x90, 0x94,
	0x73, 0x09, 0x14, 0xd0, 0x46, 0x70, 0xf5, 0xb3, 0x1f, 0x89, 0x24, 0x95, 0x57, 0xc1, 0x9d, 0x56,
	0x69, 0xa3, 0xce, 0x3e, 0x07, 0x05, 0x39, 0x39, 0x42, 0xde, 0x42, 0x5d, 0x46, 0x7e, 0x61, 0xaf,
	0xe6, 0xf6, 0x9a, 0x8c, 0xfc, 0xdc, 0x8b, 0xa1, 0x9c, 0x85, 0x0b, 0xad, 0xd6, 0x46, 0x9d, 0x3a,
	0x57, 0x47, 0x72, 0x0c, 0x15, 0x55, 0xf5, 0x4e, 0xab, 0xe7, 0x58, 0xf1, 0xa3, 0xfb, 0x3d, 0x54,
	0x8b, 0xea, 0xe4, 0x08, 0x0e, 0x2d, 0xe1, 0x70, 0xe6, 0x32, 0x31, 0xb1, 0x2f, 0x6c, 0xf6, 0x93,
	0x8d, 0x5f, 0x90, 0x43, 0x68, 0x6c, 0x40, 0xd7, 0x70, 0x30, 0xda, 0x06, 0x26, 0xa6, 0x83, 0x4b,
	0xdd, 0x5f, 0x10, 0x34, 0x06, 0xd2, 0x0b, 0xb3, 0x1b, 0xe3, 0x46, 0xce, 0x6e, 0x89, 0x06, 0xc7,
	0x03, 0xaa, 0x5b, 0xee, 0x40, 0x18, 0x03, 0x6a, 0x5c, 0x6c, 0x65, 0x7d, 0x05, 0xed, 0x47, 0x4c,
	0x5f, 0x1f, 0x5a, 0x82, 0x33, 0xe6, 0x5a, 0xe7, 0x82, 0xf5, 0xfb, 0xd6, 0xd0, 0xa6, 0x18, 0x91,
	0x33, 0x78, 0xff, 0x54, 0x35, 0xa6, 0xfc, 0x23, 0xe5, 0xf9, 0x19, 0x97, 0x54, 0xd3, 0x47, 0x12,
	0x76, 0x81, 0xcb, 0xdd, 0x7f, 0x10, 0x1c, 0x18, 0x61, 0xbc, 0xf4, 0x43, 0x99, 0x8d, 0x33, 0x2f,
	0x93, 0xa4, 0x09, 0x9f, 0x19, 0x16, 0x9b, 0x98, 0x16, 0x75, 0xc5, 0xd8, 0xd5, 0x5d, 0xba, 0xd5,
	0xe5, 0x2d, 0x9c, 0xec, 0x70, 0x94, 0x73, 0xc6, 0xc7, 0x18, 0xa9, 0x01, 0x76, 0x28, 0x4e, 0x75,
	0xf3, 0x12, 0x97, 0x9e, 0x09, 0xdc, 0xd4, 0x2e, 0x93, 0x16, 0x34, 0x77, 0x38, 0x9b, 0xb9, 0xc2,
	0xe1, 0x74, 0x4c, 0x6d, 0x17, 0xef, 0x91, 0x37, 0x70, 0xb4, 0xc3, 0x0f, 0xed, 0xa1, 0x8b, 0x2b,
	0xcf, 0xb5, 0x74, 0x7e, 0xe0, 0xba, 0x49, 0x71, 0x95, 0xbc, 0x03, 0x6d, 0x37, 0x94, 0x52, 0x53,
	0x8c, 0x2f, 0x6d, 0x03, 0xd7, 0xba, 0xbf, 0x97, 0x00, 0x8f, 0x3c, 0xf5, 0x97, 0x8a, 0xbc, 0x68,
	0x26, 0x8b, 0xa1, 0x8f, 0x01, 0xdb, 0x8c, 0x8f, 0x74, 0x4b, 0x30, 0x87, 0x72, 0xdd, 0x1d, 0x32,
	0x35, 0xee, 0x09, 0xbc, 0x1e, 0xe9, 0x43, 0xdb, 0xa5, 0xb6, 0x6e, 0x1b, 0x54, 0x65, 0x71, 0x17,
	0x23, 0xf2, 0x0e, 0x88, 0x5a, 0x29, 0x53, 0xbb, 0xe5, 0xf4, 0xc7, 0x09, 0x1d, 0xbb, 0xd4, 0xc4,
	0xa5, 0xe6, 0xde, 0x6f, 0xff, 0x6a, 0x88, 0xbc, 0x81, 0x83, 0x07, 0xd6, 0x64, 0x6a, 0xca, 0x35,
	0xa1, 0xc1, 0xab, 0x07, 0x22, 0x5f, 0x1b, 0xde, 0x5b, 0x33, 0x67, 0xf0, 0xfe, 0xc9, 0x3d, 0xc2,
	0x66, 0x62, 0x23, 0xc7, 0x15, 0x95, 0x6a, 0xf0, 0xd1, 0xd6, 0x75, 0xd5, 0xb5, 0xf7, 0x14, 0x4e,
	0x14, 0x31, 0xb1, 0x4d, 0xca, 0xc5, 0x56, 0x0a, 0xae, 0xad, 0x05, 0x47, 0xb0, 0xaf, 0x04, 0xc5,
	0x8d, 0xf5, 0xff, 0x5d, 0xbb, 0xf3, 0x16, 0x9b, 0xdd, 0x5f, 0x0b, 0x3e, 0x87, 0xd7, 0x4f, 0x23,
	0x4f, 0x0b, 0xf2, 0xbb, 0x3e, 0xbc, 0x94, 0xd1, 0x72, 0x2e, 0xa6, 0xde, 0xec, 0x56, 0x46, 0x3e,
	0x39, 0xeb, 0x15, 0xcf, 0x43, 0x6f, 0xf3, 0x3c, 0xf4, 0x68, 0xb4, 0x9c, 0x7f, 0xf4, 0xc2, 0xa5,
	0x64, 0x49, 0x16, 0xc4, 0xd1, 0x42, 0x5b, 0xfd, 0x5a, 0xce, 0xbf, 0x9f, 0x86, 0x32, 0x9e, 0x17,
	0xbe, 0x73, 0x7c, 0xff, 0x57, 0xeb, 0xc5, 0xfd, 0xaa, 0x85, 0xfe, 0x58, 0xb5, 0xd0, 0x9f, 0xab,
	0x16, 0x9a, 0x56, 0xf3, 0x84, 0x6f, 0xfe, 0x0b, 0x00, 0x00, 0xff, 0xff, 0x21, 0x31, 0x8e, 0xd3,
	0x9b, 0x04, 0x00, 0x00,
}

func (m *AppPort) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AppPort) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *AppPort) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.Nginx {
		i--
		if m.Nginx {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x40
	}
	if m.Tls {
		i--
		if m.Tls {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x38
	}
	if m.EndPort != 0 {
		i = encodeVarintAppcommon(dAtA, i, uint64(m.EndPort))
		i--
		dAtA[i] = 0x30
	}
	if len(m.FqdnPrefix) > 0 {
		i -= len(m.FqdnPrefix)
		copy(dAtA[i:], m.FqdnPrefix)
		i = encodeVarintAppcommon(dAtA, i, uint64(len(m.FqdnPrefix)))
		i--
		dAtA[i] = 0x2a
	}
	if m.PublicPort != 0 {
		i = encodeVarintAppcommon(dAtA, i, uint64(m.PublicPort))
		i--
		dAtA[i] = 0x18
	}
	if m.InternalPort != 0 {
		i = encodeVarintAppcommon(dAtA, i, uint64(m.InternalPort))
		i--
		dAtA[i] = 0x10
	}
	if m.Proto != 0 {
		i = encodeVarintAppcommon(dAtA, i, uint64(m.Proto))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintAppcommon(dAtA []byte, offset int, v uint64) int {
	offset -= sovAppcommon(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *AppPort) CopyInFields(src *AppPort) int {
	changed := 0
	if m.Proto != src.Proto {
		m.Proto = src.Proto
		changed++
	}
	if m.InternalPort != src.InternalPort {
		m.InternalPort = src.InternalPort
		changed++
	}
	if m.PublicPort != src.PublicPort {
		m.PublicPort = src.PublicPort
		changed++
	}
	if m.FqdnPrefix != src.FqdnPrefix {
		m.FqdnPrefix = src.FqdnPrefix
		changed++
	}
	if m.EndPort != src.EndPort {
		m.EndPort = src.EndPort
		changed++
	}
	if m.Tls != src.Tls {
		m.Tls = src.Tls
		changed++
	}
	if m.Nginx != src.Nginx {
		m.Nginx = src.Nginx
		changed++
	}
	return changed
}

func (m *AppPort) DeepCopyIn(src *AppPort) {
	m.Proto = src.Proto
	m.InternalPort = src.InternalPort
	m.PublicPort = src.PublicPort
	m.FqdnPrefix = src.FqdnPrefix
	m.EndPort = src.EndPort
	m.Tls = src.Tls
	m.Nginx = src.Nginx
}

// Helper method to check that enums have valid values
func (m *AppPort) ValidateEnums() error {
	if _, ok := LProto_name[int32(m.Proto)]; !ok {
		return errors.New("invalid Proto")
	}
	return nil
}

var LProtoStrings = []string{
	"L_PROTO_UNKNOWN",
	"L_PROTO_TCP",
	"L_PROTO_UDP",
}

const (
	LProtoL_PROTO_UNKNOWN uint64 = 1 << 0
	LProtoL_PROTO_TCP     uint64 = 1 << 1
	LProtoL_PROTO_UDP     uint64 = 1 << 2
)

var LProto_CamelName = map[int32]string{
	// L_PROTO_UNKNOWN -> LProtoUnknown
	0: "LProtoUnknown",
	// L_PROTO_TCP -> LProtoTcp
	1: "LProtoTcp",
	// L_PROTO_UDP -> LProtoUdp
	2: "LProtoUdp",
}
var LProto_CamelValue = map[string]int32{
	"LProtoUnknown": 0,
	"LProtoTcp":     1,
	"LProtoUdp":     2,
}

func (e *LProto) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	val, ok := LProto_CamelValue[util.CamelCase(str)]
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = LProto_CamelName[val]
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("No enum value for %s", str))
	}
	*e = LProto(val)
	return nil
}

func (e LProto) MarshalYAML() (interface{}, error) {
	return proto.EnumName(LProto_CamelName, int32(e)), nil
}

// custom JSON encoding/decoding
func (e *LProto) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		val, ok := LProto_CamelValue[util.CamelCase(str)]
		if !ok {
			// may be int value instead of enum name
			ival, err := strconv.Atoi(str)
			val = int32(ival)
			if err == nil {
				_, ok = LProto_CamelName[val]
			}
		}
		if !ok {
			return errors.New(fmt.Sprintf("No enum value for %s", str))
		}
		*e = LProto(val)
		return nil
	}
	var val int32
	err = json.Unmarshal(b, &val)
	if err == nil {
		*e = LProto(val)
		return nil
	}
	return fmt.Errorf("No enum value for %v", b)
}

var HealthCheckStrings = []string{
	"HEALTH_CHECK_UNKNOWN",
	"HEALTH_CHECK_FAIL_ROOTLB_OFFLINE",
	"HEALTH_CHECK_FAIL_SERVER_FAIL",
	"HEALTH_CHECK_OK",
}

const (
	HealthCheckHEALTH_CHECK_UNKNOWN             uint64 = 1 << 0
	HealthCheckHEALTH_CHECK_FAIL_ROOTLB_OFFLINE uint64 = 1 << 1
	HealthCheckHEALTH_CHECK_FAIL_SERVER_FAIL    uint64 = 1 << 2
	HealthCheckHEALTH_CHECK_OK                  uint64 = 1 << 3
)

var HealthCheck_CamelName = map[int32]string{
	// HEALTH_CHECK_UNKNOWN -> HealthCheckUnknown
	0: "HealthCheckUnknown",
	// HEALTH_CHECK_FAIL_ROOTLB_OFFLINE -> HealthCheckFailRootlbOffline
	1: "HealthCheckFailRootlbOffline",
	// HEALTH_CHECK_FAIL_SERVER_FAIL -> HealthCheckFailServerFail
	2: "HealthCheckFailServerFail",
	// HEALTH_CHECK_OK -> HealthCheckOk
	3: "HealthCheckOk",
}
var HealthCheck_CamelValue = map[string]int32{
	"HealthCheckUnknown":           0,
	"HealthCheckFailRootlbOffline": 1,
	"HealthCheckFailServerFail":    2,
	"HealthCheckOk":                3,
}

func (e *HealthCheck) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	val, ok := HealthCheck_CamelValue[util.CamelCase(str)]
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = HealthCheck_CamelName[val]
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("No enum value for %s", str))
	}
	*e = HealthCheck(val)
	return nil
}

func (e HealthCheck) MarshalYAML() (interface{}, error) {
	return proto.EnumName(HealthCheck_CamelName, int32(e)), nil
}

// custom JSON encoding/decoding
func (e *HealthCheck) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		val, ok := HealthCheck_CamelValue[util.CamelCase(str)]
		if !ok {
			// may be int value instead of enum name
			ival, err := strconv.Atoi(str)
			val = int32(ival)
			if err == nil {
				_, ok = HealthCheck_CamelName[val]
			}
		}
		if !ok {
			return errors.New(fmt.Sprintf("No enum value for %s", str))
		}
		*e = HealthCheck(val)
		return nil
	}
	var val int32
	err = json.Unmarshal(b, &val)
	if err == nil {
		*e = HealthCheck(val)
		return nil
	}
	return fmt.Errorf("No enum value for %v", b)
}

var CloudletStateStrings = []string{
	"CLOUDLET_STATE_UNKNOWN",
	"CLOUDLET_STATE_ERRORS",
	"CLOUDLET_STATE_READY",
	"CLOUDLET_STATE_OFFLINE",
	"CLOUDLET_STATE_NOT_PRESENT",
	"CLOUDLET_STATE_INIT",
	"CLOUDLET_STATE_UPGRADE",
	"CLOUDLET_STATE_NEED_SYNC",
}

const (
	CloudletStateCLOUDLET_STATE_UNKNOWN     uint64 = 1 << 0
	CloudletStateCLOUDLET_STATE_ERRORS      uint64 = 1 << 1
	CloudletStateCLOUDLET_STATE_READY       uint64 = 1 << 2
	CloudletStateCLOUDLET_STATE_OFFLINE     uint64 = 1 << 3
	CloudletStateCLOUDLET_STATE_NOT_PRESENT uint64 = 1 << 4
	CloudletStateCLOUDLET_STATE_INIT        uint64 = 1 << 5
	CloudletStateCLOUDLET_STATE_UPGRADE     uint64 = 1 << 6
	CloudletStateCLOUDLET_STATE_NEED_SYNC   uint64 = 1 << 7
)

var CloudletState_CamelName = map[int32]string{
	// CLOUDLET_STATE_UNKNOWN -> CloudletStateUnknown
	0: "CloudletStateUnknown",
	// CLOUDLET_STATE_ERRORS -> CloudletStateErrors
	1: "CloudletStateErrors",
	// CLOUDLET_STATE_READY -> CloudletStateReady
	2: "CloudletStateReady",
	// CLOUDLET_STATE_OFFLINE -> CloudletStateOffline
	3: "CloudletStateOffline",
	// CLOUDLET_STATE_NOT_PRESENT -> CloudletStateNotPresent
	4: "CloudletStateNotPresent",
	// CLOUDLET_STATE_INIT -> CloudletStateInit
	5: "CloudletStateInit",
	// CLOUDLET_STATE_UPGRADE -> CloudletStateUpgrade
	6: "CloudletStateUpgrade",
	// CLOUDLET_STATE_NEED_SYNC -> CloudletStateNeedSync
	7: "CloudletStateNeedSync",
}
var CloudletState_CamelValue = map[string]int32{
	"CloudletStateUnknown":    0,
	"CloudletStateErrors":     1,
	"CloudletStateReady":      2,
	"CloudletStateOffline":    3,
	"CloudletStateNotPresent": 4,
	"CloudletStateInit":       5,
	"CloudletStateUpgrade":    6,
	"CloudletStateNeedSync":   7,
}

func (e *CloudletState) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	val, ok := CloudletState_CamelValue[util.CamelCase(str)]
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = CloudletState_CamelName[val]
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("No enum value for %s", str))
	}
	*e = CloudletState(val)
	return nil
}

func (e CloudletState) MarshalYAML() (interface{}, error) {
	return proto.EnumName(CloudletState_CamelName, int32(e)), nil
}

// custom JSON encoding/decoding
func (e *CloudletState) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		val, ok := CloudletState_CamelValue[util.CamelCase(str)]
		if !ok {
			// may be int value instead of enum name
			ival, err := strconv.Atoi(str)
			val = int32(ival)
			if err == nil {
				_, ok = CloudletState_CamelName[val]
			}
		}
		if !ok {
			return errors.New(fmt.Sprintf("No enum value for %s", str))
		}
		*e = CloudletState(val)
		return nil
	}
	var val int32
	err = json.Unmarshal(b, &val)
	if err == nil {
		*e = CloudletState(val)
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
func (m *AppPort) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Proto != 0 {
		n += 1 + sovAppcommon(uint64(m.Proto))
	}
	if m.InternalPort != 0 {
		n += 1 + sovAppcommon(uint64(m.InternalPort))
	}
	if m.PublicPort != 0 {
		n += 1 + sovAppcommon(uint64(m.PublicPort))
	}
	l = len(m.FqdnPrefix)
	if l > 0 {
		n += 1 + l + sovAppcommon(uint64(l))
	}
	if m.EndPort != 0 {
		n += 1 + sovAppcommon(uint64(m.EndPort))
	}
	if m.Tls {
		n += 2
	}
	if m.Nginx {
		n += 2
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovAppcommon(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozAppcommon(x uint64) (n int) {
	return sovAppcommon(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *AppPort) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAppcommon
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
			return fmt.Errorf("proto: AppPort: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AppPort: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Proto", wireType)
			}
			m.Proto = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppcommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Proto |= LProto(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field InternalPort", wireType)
			}
			m.InternalPort = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppcommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.InternalPort |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PublicPort", wireType)
			}
			m.PublicPort = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppcommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PublicPort |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FqdnPrefix", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppcommon
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
				return ErrInvalidLengthAppcommon
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAppcommon
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.FqdnPrefix = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field EndPort", wireType)
			}
			m.EndPort = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppcommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.EndPort |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Tls", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppcommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Tls = bool(v != 0)
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Nginx", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppcommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Nginx = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipAppcommon(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthAppcommon
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthAppcommon
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipAppcommon(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowAppcommon
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
					return 0, ErrIntOverflowAppcommon
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
					return 0, ErrIntOverflowAppcommon
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
				return 0, ErrInvalidLengthAppcommon
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupAppcommon
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthAppcommon
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthAppcommon        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowAppcommon          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupAppcommon = fmt.Errorf("proto: unexpected end of group")
)
