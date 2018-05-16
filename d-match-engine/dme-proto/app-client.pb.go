// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app-client.proto

/*
	Package distributed_match_engine is a generated protocol buffer package.

	It is generated from these files:
		app-client.proto

	It has these top-level messages:
		Match_Engine_Request
		Match_Engine_Reply
		Match_Engine_Loc_Verify
*/
package distributed_match_engine

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"

import context "golang.org/x/net/context"
import grpc "google.golang.org/grpc"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// User ID type - IMEI, MSISDN etc
type Match_Engine_Request_IDType int32

const (
	Match_Engine_Request_IMEI   Match_Engine_Request_IDType = 0
	Match_Engine_Request_MSISDN Match_Engine_Request_IDType = 1
)

var Match_Engine_Request_IDType_name = map[int32]string{
	0: "IMEI",
	1: "MSISDN",
}
var Match_Engine_Request_IDType_value = map[string]int32{
	"IMEI":   0,
	"MSISDN": 1,
}

func (x Match_Engine_Request_IDType) String() string {
	return proto.EnumName(Match_Engine_Request_IDType_name, int32(x))
}
func (Match_Engine_Request_IDType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptorAppClient, []int{0, 0}
}

// Status of the reply
type Match_Engine_Loc_Verify_Tower_Status int32

const (
	Match_Engine_Loc_Verify_UNKNOWN                          Match_Engine_Loc_Verify_Tower_Status = 0
	Match_Engine_Loc_Verify_CONNECTED_TO_SPECIFIED_TOWER     Match_Engine_Loc_Verify_Tower_Status = 1
	Match_Engine_Loc_Verify_NOT_CONNECTED_TO_SPECIFIED_TOWER Match_Engine_Loc_Verify_Tower_Status = 2
)

var Match_Engine_Loc_Verify_Tower_Status_name = map[int32]string{
	0: "UNKNOWN",
	1: "CONNECTED_TO_SPECIFIED_TOWER",
	2: "NOT_CONNECTED_TO_SPECIFIED_TOWER",
}
var Match_Engine_Loc_Verify_Tower_Status_value = map[string]int32{
	"UNKNOWN":                          0,
	"CONNECTED_TO_SPECIFIED_TOWER":     1,
	"NOT_CONNECTED_TO_SPECIFIED_TOWER": 2,
}

func (x Match_Engine_Loc_Verify_Tower_Status) String() string {
	return proto.EnumName(Match_Engine_Loc_Verify_Tower_Status_name, int32(x))
}
func (Match_Engine_Loc_Verify_Tower_Status) EnumDescriptor() ([]byte, []int) {
	return fileDescriptorAppClient, []int{2, 0}
}

type Match_Engine_Loc_Verify_GPS_Location_Status int32

const (
	Match_Engine_Loc_Verify_LOC_UNKNOWN    Match_Engine_Loc_Verify_GPS_Location_Status = 0
	Match_Engine_Loc_Verify_LOC_WITHIN_1M  Match_Engine_Loc_Verify_GPS_Location_Status = 1
	Match_Engine_Loc_Verify_LOC_WITHIN_10M Match_Engine_Loc_Verify_GPS_Location_Status = 2
)

var Match_Engine_Loc_Verify_GPS_Location_Status_name = map[int32]string{
	0: "LOC_UNKNOWN",
	1: "LOC_WITHIN_1M",
	2: "LOC_WITHIN_10M",
}
var Match_Engine_Loc_Verify_GPS_Location_Status_value = map[string]int32{
	"LOC_UNKNOWN":    0,
	"LOC_WITHIN_1M":  1,
	"LOC_WITHIN_10M": 2,
}

func (x Match_Engine_Loc_Verify_GPS_Location_Status) String() string {
	return proto.EnumName(Match_Engine_Loc_Verify_GPS_Location_Status_name, int32(x))
}
func (Match_Engine_Loc_Verify_GPS_Location_Status) EnumDescriptor() ([]byte, []int) {
	return fileDescriptorAppClient, []int{2, 1}
}

type Match_Engine_Request struct {
	Ver    uint32                      `protobuf:"varint,1,opt,name=ver,proto3" json:"ver,omitempty"`
	IdType Match_Engine_Request_IDType `protobuf:"varint,2,opt,name=id_type,json=idType,proto3,enum=distributed_match_engine.Match_Engine_Request_IDType" json:"id_type,omitempty"`
	// Actual ID
	Id string `protobuf:"bytes,3,opt,name=id,proto3" json:"id,omitempty"`
	// The carrier that user is connected to
	Carrier uint64 `protobuf:"varint,4,opt,name=carrier,proto3" json:"carrier,omitempty"`
	// The tower that user is currently connected to
	Tower uint64 `protobuf:"varint,5,opt,name=tower,proto3" json:"tower,omitempty"`
	// The GPS location of the user
	GpsLocation *edgeproto.Loc `protobuf:"bytes,6,opt,name=gps_location,json=gpsLocation" json:"gps_location,omitempty"`
	// Edge-cloud assigned application ID
	AppId uint64 `protobuf:"varint,7,opt,name=app_id,json=appId,proto3" json:"app_id,omitempty"`
	// Protocol application uses
	Protocol []byte `protobuf:"bytes,8,opt,name=protocol,proto3" json:"protocol,omitempty"`
	// The protocol port on the server side
	ServerPort []byte `protobuf:"bytes,9,opt,name=server_port,json=serverPort,proto3" json:"server_port,omitempty"`
}

func (m *Match_Engine_Request) Reset()                    { *m = Match_Engine_Request{} }
func (m *Match_Engine_Request) String() string            { return proto.CompactTextString(m) }
func (*Match_Engine_Request) ProtoMessage()               {}
func (*Match_Engine_Request) Descriptor() ([]byte, []int) { return fileDescriptorAppClient, []int{0} }

type Match_Engine_Reply struct {
	Ver uint32 `protobuf:"varint,1,opt,name=ver,proto3" json:"ver,omitempty"`
	// ip of the app service
	ServiceIp []byte `protobuf:"bytes,2,opt,name=service_ip,json=serviceIp,proto3" json:"service_ip,omitempty"`
	// port of the app service?
	ServerPort uint32 `protobuf:"varint,3,opt,name=server_port,json=serverPort,proto3" json:"server_port,omitempty"`
	// location of the cloudlet?
	CloudletLocation *edgeproto.Loc `protobuf:"bytes,4,opt,name=cloudlet_location,json=cloudletLocation" json:"cloudlet_location,omitempty"`
}

func (m *Match_Engine_Reply) Reset()                    { *m = Match_Engine_Reply{} }
func (m *Match_Engine_Reply) String() string            { return proto.CompactTextString(m) }
func (*Match_Engine_Reply) ProtoMessage()               {}
func (*Match_Engine_Reply) Descriptor() ([]byte, []int) { return fileDescriptorAppClient, []int{1} }

type Match_Engine_Loc_Verify struct {
	Ver               uint32                                      `protobuf:"varint,1,opt,name=ver,proto3" json:"ver,omitempty"`
	TowerStatus       Match_Engine_Loc_Verify_Tower_Status        `protobuf:"varint,2,opt,name=tower_status,json=towerStatus,proto3,enum=distributed_match_engine.Match_Engine_Loc_Verify_Tower_Status" json:"tower_status,omitempty"`
	GpsLocationStatus Match_Engine_Loc_Verify_GPS_Location_Status `protobuf:"varint,3,opt,name=gps_location_status,json=gpsLocationStatus,proto3,enum=distributed_match_engine.Match_Engine_Loc_Verify_GPS_Location_Status" json:"gps_location_status,omitempty"`
}

func (m *Match_Engine_Loc_Verify) Reset()                    { *m = Match_Engine_Loc_Verify{} }
func (m *Match_Engine_Loc_Verify) String() string            { return proto.CompactTextString(m) }
func (*Match_Engine_Loc_Verify) ProtoMessage()               {}
func (*Match_Engine_Loc_Verify) Descriptor() ([]byte, []int) { return fileDescriptorAppClient, []int{2} }

func init() {
	proto.RegisterType((*Match_Engine_Request)(nil), "distributed_match_engine.Match_Engine_Request")
	proto.RegisterType((*Match_Engine_Reply)(nil), "distributed_match_engine.Match_Engine_Reply")
	proto.RegisterType((*Match_Engine_Loc_Verify)(nil), "distributed_match_engine.Match_Engine_Loc_Verify")
	proto.RegisterEnum("distributed_match_engine.Match_Engine_Request_IDType", Match_Engine_Request_IDType_name, Match_Engine_Request_IDType_value)
	proto.RegisterEnum("distributed_match_engine.Match_Engine_Loc_Verify_Tower_Status", Match_Engine_Loc_Verify_Tower_Status_name, Match_Engine_Loc_Verify_Tower_Status_value)
	proto.RegisterEnum("distributed_match_engine.Match_Engine_Loc_Verify_GPS_Location_Status", Match_Engine_Loc_Verify_GPS_Location_Status_name, Match_Engine_Loc_Verify_GPS_Location_Status_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Match_Engine_Api service

type Match_Engine_ApiClient interface {
	FindCloudlet(ctx context.Context, in *Match_Engine_Request, opts ...grpc.CallOption) (*Match_Engine_Reply, error)
	VerifyLocation(ctx context.Context, in *Match_Engine_Request, opts ...grpc.CallOption) (*Match_Engine_Loc_Verify, error)
}

type match_Engine_ApiClient struct {
	cc *grpc.ClientConn
}

func NewMatch_Engine_ApiClient(cc *grpc.ClientConn) Match_Engine_ApiClient {
	return &match_Engine_ApiClient{cc}
}

func (c *match_Engine_ApiClient) FindCloudlet(ctx context.Context, in *Match_Engine_Request, opts ...grpc.CallOption) (*Match_Engine_Reply, error) {
	out := new(Match_Engine_Reply)
	err := grpc.Invoke(ctx, "/distributed_match_engine.Match_Engine_Api/FindCloudlet", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *match_Engine_ApiClient) VerifyLocation(ctx context.Context, in *Match_Engine_Request, opts ...grpc.CallOption) (*Match_Engine_Loc_Verify, error) {
	out := new(Match_Engine_Loc_Verify)
	err := grpc.Invoke(ctx, "/distributed_match_engine.Match_Engine_Api/VerifyLocation", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Match_Engine_Api service

type Match_Engine_ApiServer interface {
	FindCloudlet(context.Context, *Match_Engine_Request) (*Match_Engine_Reply, error)
	VerifyLocation(context.Context, *Match_Engine_Request) (*Match_Engine_Loc_Verify, error)
}

func RegisterMatch_Engine_ApiServer(s *grpc.Server, srv Match_Engine_ApiServer) {
	s.RegisterService(&_Match_Engine_Api_serviceDesc, srv)
}

func _Match_Engine_Api_FindCloudlet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Match_Engine_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(Match_Engine_ApiServer).FindCloudlet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/distributed_match_engine.Match_Engine_Api/FindCloudlet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(Match_Engine_ApiServer).FindCloudlet(ctx, req.(*Match_Engine_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Match_Engine_Api_VerifyLocation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Match_Engine_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(Match_Engine_ApiServer).VerifyLocation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/distributed_match_engine.Match_Engine_Api/VerifyLocation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(Match_Engine_ApiServer).VerifyLocation(ctx, req.(*Match_Engine_Request))
	}
	return interceptor(ctx, in, info, handler)
}

var _Match_Engine_Api_serviceDesc = grpc.ServiceDesc{
	ServiceName: "distributed_match_engine.Match_Engine_Api",
	HandlerType: (*Match_Engine_ApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FindCloudlet",
			Handler:    _Match_Engine_Api_FindCloudlet_Handler,
		},
		{
			MethodName: "VerifyLocation",
			Handler:    _Match_Engine_Api_VerifyLocation_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "app-client.proto",
}

func (m *Match_Engine_Request) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Match_Engine_Request) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Ver != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(m.Ver))
	}
	if m.IdType != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(m.IdType))
	}
	if len(m.Id) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(len(m.Id)))
		i += copy(dAtA[i:], m.Id)
	}
	if m.Carrier != 0 {
		dAtA[i] = 0x20
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(m.Carrier))
	}
	if m.Tower != 0 {
		dAtA[i] = 0x28
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(m.Tower))
	}
	if m.GpsLocation != nil {
		dAtA[i] = 0x32
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(m.GpsLocation.Size()))
		n1, err := m.GpsLocation.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if m.AppId != 0 {
		dAtA[i] = 0x38
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(m.AppId))
	}
	if len(m.Protocol) > 0 {
		dAtA[i] = 0x42
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(len(m.Protocol)))
		i += copy(dAtA[i:], m.Protocol)
	}
	if len(m.ServerPort) > 0 {
		dAtA[i] = 0x4a
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(len(m.ServerPort)))
		i += copy(dAtA[i:], m.ServerPort)
	}
	return i, nil
}

func (m *Match_Engine_Reply) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Match_Engine_Reply) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Ver != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(m.Ver))
	}
	if len(m.ServiceIp) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(len(m.ServiceIp)))
		i += copy(dAtA[i:], m.ServiceIp)
	}
	if m.ServerPort != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(m.ServerPort))
	}
	if m.CloudletLocation != nil {
		dAtA[i] = 0x22
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(m.CloudletLocation.Size()))
		n2, err := m.CloudletLocation.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	return i, nil
}

func (m *Match_Engine_Loc_Verify) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Match_Engine_Loc_Verify) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Ver != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(m.Ver))
	}
	if m.TowerStatus != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(m.TowerStatus))
	}
	if m.GpsLocationStatus != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintAppClient(dAtA, i, uint64(m.GpsLocationStatus))
	}
	return i, nil
}

func encodeVarintAppClient(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Match_Engine_Request) CopyInFields(src *Match_Engine_Request) {
	m.Ver = src.Ver
	m.IdType = src.IdType
	m.Id = src.Id
	m.Carrier = src.Carrier
	m.Tower = src.Tower
	if m.GpsLocation != nil && src.GpsLocation != nil {
		*m.GpsLocation = *src.GpsLocation
	}
	m.AppId = src.AppId
	copy(m.Protocol, src.Protocol)
	copy(m.ServerPort, src.ServerPort)
}

func (m *Match_Engine_Reply) CopyInFields(src *Match_Engine_Reply) {
	m.Ver = src.Ver
	copy(m.ServiceIp, src.ServiceIp)
	m.ServerPort = src.ServerPort
	if m.CloudletLocation != nil && src.CloudletLocation != nil {
		*m.CloudletLocation = *src.CloudletLocation
	}
}

func (m *Match_Engine_Loc_Verify) CopyInFields(src *Match_Engine_Loc_Verify) {
	m.Ver = src.Ver
	m.TowerStatus = src.TowerStatus
	m.GpsLocationStatus = src.GpsLocationStatus
}

func (m *Match_Engine_Request) Size() (n int) {
	var l int
	_ = l
	if m.Ver != 0 {
		n += 1 + sovAppClient(uint64(m.Ver))
	}
	if m.IdType != 0 {
		n += 1 + sovAppClient(uint64(m.IdType))
	}
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + sovAppClient(uint64(l))
	}
	if m.Carrier != 0 {
		n += 1 + sovAppClient(uint64(m.Carrier))
	}
	if m.Tower != 0 {
		n += 1 + sovAppClient(uint64(m.Tower))
	}
	if m.GpsLocation != nil {
		l = m.GpsLocation.Size()
		n += 1 + l + sovAppClient(uint64(l))
	}
	if m.AppId != 0 {
		n += 1 + sovAppClient(uint64(m.AppId))
	}
	l = len(m.Protocol)
	if l > 0 {
		n += 1 + l + sovAppClient(uint64(l))
	}
	l = len(m.ServerPort)
	if l > 0 {
		n += 1 + l + sovAppClient(uint64(l))
	}
	return n
}

func (m *Match_Engine_Reply) Size() (n int) {
	var l int
	_ = l
	if m.Ver != 0 {
		n += 1 + sovAppClient(uint64(m.Ver))
	}
	l = len(m.ServiceIp)
	if l > 0 {
		n += 1 + l + sovAppClient(uint64(l))
	}
	if m.ServerPort != 0 {
		n += 1 + sovAppClient(uint64(m.ServerPort))
	}
	if m.CloudletLocation != nil {
		l = m.CloudletLocation.Size()
		n += 1 + l + sovAppClient(uint64(l))
	}
	return n
}

func (m *Match_Engine_Loc_Verify) Size() (n int) {
	var l int
	_ = l
	if m.Ver != 0 {
		n += 1 + sovAppClient(uint64(m.Ver))
	}
	if m.TowerStatus != 0 {
		n += 1 + sovAppClient(uint64(m.TowerStatus))
	}
	if m.GpsLocationStatus != 0 {
		n += 1 + sovAppClient(uint64(m.GpsLocationStatus))
	}
	return n
}

func sovAppClient(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozAppClient(x uint64) (n int) {
	return sovAppClient(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Match_Engine_Request) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAppClient
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Match_Engine_Request: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Match_Engine_Request: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ver", wireType)
			}
			m.Ver = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Ver |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IdType", wireType)
			}
			m.IdType = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.IdType |= (Match_Engine_Request_IDType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAppClient
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Carrier", wireType)
			}
			m.Carrier = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Carrier |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Tower", wireType)
			}
			m.Tower = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Tower |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field GpsLocation", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAppClient
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.GpsLocation == nil {
				m.GpsLocation = &edgeproto.Loc{}
			}
			if err := m.GpsLocation.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AppId", wireType)
			}
			m.AppId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.AppId |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Protocol", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthAppClient
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Protocol = append(m.Protocol[:0], dAtA[iNdEx:postIndex]...)
			if m.Protocol == nil {
				m.Protocol = []byte{}
			}
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ServerPort", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthAppClient
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ServerPort = append(m.ServerPort[:0], dAtA[iNdEx:postIndex]...)
			if m.ServerPort == nil {
				m.ServerPort = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAppClient(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthAppClient
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
func (m *Match_Engine_Reply) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAppClient
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Match_Engine_Reply: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Match_Engine_Reply: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ver", wireType)
			}
			m.Ver = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Ver |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ServiceIp", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthAppClient
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ServiceIp = append(m.ServiceIp[:0], dAtA[iNdEx:postIndex]...)
			if m.ServiceIp == nil {
				m.ServiceIp = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ServerPort", wireType)
			}
			m.ServerPort = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ServerPort |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CloudletLocation", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAppClient
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.CloudletLocation == nil {
				m.CloudletLocation = &edgeproto.Loc{}
			}
			if err := m.CloudletLocation.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAppClient(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthAppClient
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
func (m *Match_Engine_Loc_Verify) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAppClient
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Match_Engine_Loc_Verify: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Match_Engine_Loc_Verify: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ver", wireType)
			}
			m.Ver = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Ver |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TowerStatus", wireType)
			}
			m.TowerStatus = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TowerStatus |= (Match_Engine_Loc_Verify_Tower_Status(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field GpsLocationStatus", wireType)
			}
			m.GpsLocationStatus = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.GpsLocationStatus |= (Match_Engine_Loc_Verify_GPS_Location_Status(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipAppClient(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthAppClient
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
func skipAppClient(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowAppClient
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
					return 0, ErrIntOverflowAppClient
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowAppClient
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
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthAppClient
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowAppClient
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipAppClient(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthAppClient = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowAppClient   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("app-client.proto", fileDescriptorAppClient) }

var fileDescriptorAppClient = []byte{
	// 633 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x53, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0xcd, 0x26, 0x69, 0xd2, 0x4e, 0xd2, 0xe0, 0x6e, 0x8b, 0xb0, 0x2a, 0x48, 0x2d, 0x8b, 0x43,
	0x0e, 0xd4, 0xa5, 0x05, 0x71, 0x41, 0x42, 0x82, 0xd4, 0x05, 0x8b, 0xc4, 0xa9, 0x9c, 0x40, 0x8f,
	0x2b, 0xc7, 0x5e, 0xd2, 0x95, 0xdc, 0xec, 0x62, 0xaf, 0x0b, 0xf9, 0x1e, 0xbe, 0x82, 0x3f, 0xe8,
	0x91, 0x4f, 0xa0, 0xfd, 0x0b, 0x6e, 0xa8, 0xeb, 0xb8, 0x49, 0xab, 0x80, 0x9a, 0x8b, 0xb5, 0x33,
	0x9e, 0x7d, 0xef, 0xcd, 0x9b, 0x59, 0xd0, 0x7c, 0x21, 0x76, 0x83, 0x88, 0xd1, 0xb1, 0xb4, 0x44,
	0xcc, 0x25, 0xc7, 0x7a, 0xc8, 0x12, 0x19, 0xb3, 0x61, 0x2a, 0x69, 0x48, 0xce, 0x7c, 0x19, 0x9c,
	0x12, 0x3a, 0x1e, 0xb1, 0x31, 0xdd, 0x7e, 0x39, 0x62, 0xf2, 0x34, 0x1d, 0x5a, 0x01, 0x3f, 0xdb,
	0x3b, 0xe3, 0x43, 0x16, 0xd1, 0x70, 0x44, 0xbf, 0xef, 0x5d, 0x7f, 0x77, 0x83, 0x88, 0xa7, 0xa1,
	0x3a, 0x2a, 0x94, 0xbd, 0x88, 0x07, 0x19, 0x9e, 0x79, 0x59, 0x84, 0xad, 0xae, 0x82, 0xb1, 0x15,
	0x0c, 0xf1, 0xe8, 0xd7, 0x94, 0x26, 0x12, 0x6b, 0x50, 0x3a, 0xa7, 0xb1, 0x8e, 0x0c, 0xd4, 0x5a,
	0xf7, 0xae, 0x8f, 0xb8, 0x07, 0x55, 0x16, 0x12, 0x39, 0x11, 0x54, 0x2f, 0x1a, 0xa8, 0xd5, 0x38,
	0x78, 0x65, 0xfd, 0x4b, 0x8c, 0xb5, 0x08, 0xd2, 0x72, 0x0e, 0xd5, 0x6d, 0xaf, 0xc2, 0xc2, 0xc1,
	0x44, 0x50, 0xdc, 0x80, 0x22, 0x0b, 0xf5, 0x92, 0x81, 0x5a, 0x6b, 0x5e, 0x91, 0x85, 0x58, 0x87,
	0x6a, 0xe0, 0xc7, 0x31, 0xa3, 0xb1, 0x5e, 0x36, 0x50, 0xab, 0xec, 0xe5, 0x21, 0xde, 0x82, 0x15,
	0xc9, 0xbf, 0xd1, 0x58, 0x5f, 0x51, 0xf9, 0x2c, 0xc0, 0xfb, 0x50, 0x1f, 0x89, 0x84, 0x44, 0x3c,
	0xf0, 0x25, 0xe3, 0x63, 0xbd, 0x62, 0xa0, 0x56, 0xed, 0xa0, 0x61, 0xdd, 0xf4, 0x69, 0x75, 0x78,
	0xe0, 0xd5, 0x46, 0x22, 0xe9, 0x4c, 0x4b, 0xf0, 0x43, 0xa8, 0xf8, 0x42, 0x10, 0x16, 0xea, 0xd5,
	0x0c, 0xc9, 0x17, 0xc2, 0x09, 0xf1, 0x36, 0xac, 0xaa, 0x0b, 0x01, 0x8f, 0xf4, 0x55, 0x03, 0xb5,
	0xea, 0xde, 0x4d, 0x8c, 0x77, 0xa0, 0x96, 0xd0, 0xf8, 0x9c, 0xc6, 0x44, 0xf0, 0x58, 0xea, 0x6b,
	0xea, 0x37, 0x64, 0xa9, 0x63, 0x1e, 0x4b, 0x73, 0x07, 0xaa, 0xd3, 0xce, 0xf0, 0x2a, 0x94, 0x9d,
	0xae, 0xed, 0x68, 0x05, 0x0c, 0x50, 0xe9, 0xf6, 0x9d, 0xfe, 0xa1, 0xab, 0x21, 0xf3, 0x07, 0x02,
	0x7c, 0xc7, 0x10, 0x11, 0x4d, 0x16, 0x38, 0xfc, 0x04, 0x14, 0x2e, 0x0b, 0x28, 0x61, 0x42, 0x99,
	0x5c, 0xf7, 0xd6, 0xa6, 0x19, 0x47, 0xdc, 0x55, 0x52, 0x52, 0x17, 0xe7, 0x94, 0xe0, 0xd7, 0xb0,
	0xa1, 0x26, 0x1d, 0x51, 0x39, 0x73, 0xa5, 0xbc, 0xd0, 0x15, 0x2d, 0x2f, 0xcc, 0xad, 0x31, 0x7f,
	0x96, 0xe0, 0xd1, 0x2d, 0x95, 0x1d, 0x1e, 0x90, 0xcf, 0x34, 0x66, 0x5f, 0x16, 0x49, 0xf5, 0xa1,
	0xae, 0x86, 0x40, 0x12, 0xe9, 0xcb, 0x34, 0x99, 0x6e, 0xc4, 0x9b, 0x7b, 0x6e, 0xc4, 0x0c, 0xda,
	0x1a, 0x28, 0x94, 0xbe, 0x42, 0xf1, 0x6a, 0x0a, 0x33, 0x0b, 0x70, 0x0a, 0x9b, 0xf3, 0xe3, 0xcd,
	0x99, 0x4a, 0x8a, 0xc9, 0x5e, 0x9e, 0xe9, 0xfd, 0x71, 0x9f, 0xe4, 0xdd, 0xe6, 0x84, 0x1b, 0x73,
	0xcb, 0x91, 0xa5, 0xcc, 0x00, 0xea, 0xf3, 0x9a, 0x70, 0x0d, 0xaa, 0x9f, 0xdc, 0x8f, 0x6e, 0xef,
	0xc4, 0xd5, 0x0a, 0xd8, 0x80, 0xc7, 0xed, 0x9e, 0xeb, 0xda, 0xed, 0x81, 0x7d, 0x48, 0x06, 0x3d,
	0xd2, 0x3f, 0xb6, 0xdb, 0xce, 0x91, 0xa3, 0x82, 0x13, 0xdb, 0xd3, 0x10, 0x7e, 0x0a, 0x86, 0xdb,
	0x1b, 0x90, 0xff, 0x56, 0x15, 0xcd, 0x2e, 0x6c, 0x2e, 0x90, 0x83, 0x1f, 0x40, 0xad, 0xd3, 0x6b,
	0x93, 0x19, 0xdf, 0x06, 0xac, 0x5f, 0x27, 0x4e, 0x9c, 0xc1, 0x07, 0xc7, 0x25, 0xfb, 0x5d, 0x0d,
	0x61, 0x0c, 0x8d, 0xf9, 0xd4, 0xf3, 0xae, 0x56, 0x3c, 0xf8, 0x83, 0x40, 0xbb, 0xd5, 0xf6, 0x5b,
	0xc1, 0xf0, 0x18, 0xea, 0x47, 0x6c, 0x1c, 0xb6, 0xa7, 0x83, 0xc6, 0xd6, 0x72, 0xcf, 0x75, 0xfb,
	0xd9, 0xbd, 0xeb, 0x45, 0x34, 0x31, 0x0b, 0x38, 0x85, 0x46, 0xe6, 0xf4, 0xcd, 0x6b, 0x5b, 0x96,
	0x71, 0x7f, 0xe9, 0xa1, 0x9a, 0x85, 0x77, 0xda, 0xc5, 0x65, 0xb3, 0x70, 0x71, 0xd5, 0x44, 0xbf,
	0xae, 0x9a, 0xe8, 0xf7, 0x55, 0x13, 0x0d, 0x2b, 0x6a, 0xcd, 0x5f, 0xfc, 0x0d, 0x00, 0x00, 0xff,
	0xff, 0xcc, 0x72, 0x15, 0x9c, 0x3e, 0x05, 0x00, 0x00,
}
