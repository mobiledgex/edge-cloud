// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: notice.proto

package edgeproto

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf2 "github.com/gogo/protobuf/types"
import _ "github.com/gogo/protobuf/gogoproto"

import context "golang.org/x/net/context"
import grpc "google.golang.org/grpc"

import "github.com/mobiledgex/edge-cloud/util"
import "errors"
import "strconv"
import "encoding/json"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// NoticeAction denotes what kind of action this notification is for.
type NoticeAction int32

const (
	// No action
	NoticeAction_NONE NoticeAction = 0
	// Update the object
	NoticeAction_UPDATE NoticeAction = 1
	// Delete the object
	NoticeAction_DELETE NoticeAction = 2
	// Version exchange negotitation message
	NoticeAction_VERSION NoticeAction = 3
	// Initial send all finished message
	NoticeAction_SENDALL_END NoticeAction = 4
)

var NoticeAction_name = map[int32]string{
	0: "NONE",
	1: "UPDATE",
	2: "DELETE",
	3: "VERSION",
	4: "SENDALL_END",
}
var NoticeAction_value = map[string]int32{
	"NONE":        0,
	"UPDATE":      1,
	"DELETE":      2,
	"VERSION":     3,
	"SENDALL_END": 4,
}

func (x NoticeAction) String() string {
	return proto.EnumName(NoticeAction_name, int32(x))
}
func (NoticeAction) EnumDescriptor() ([]byte, []int) { return fileDescriptorNotice, []int{0} }

type Notice struct {
	// Action to perform
	Action NoticeAction `protobuf:"varint,1,opt,name=action,proto3,enum=edgeproto.NoticeAction" json:"action,omitempty"`
	// Protocol version supported by sender
	Version uint32 `protobuf:"varint,2,opt,name=version,proto3" json:"version,omitempty"`
	// Data
	Any google_protobuf2.Any `protobuf:"bytes,3,opt,name=any" json:"any"`
	// Wanted Objects
	WantObjs []string `protobuf:"bytes,4,rep,name=want_objs,json=wantObjs" json:"want_objs,omitempty"`
	// Filter by cloudlet key
	FilterCloudletKey bool `protobuf:"varint,5,opt,name=filter_cloudlet_key,json=filterCloudletKey,proto3" json:"filter_cloudlet_key,omitempty"`
	// Opentracing span
	Span string `protobuf:"bytes,6,opt,name=span,proto3" json:"span,omitempty"`
}

func (m *Notice) Reset()                    { *m = Notice{} }
func (m *Notice) String() string            { return proto.CompactTextString(m) }
func (*Notice) ProtoMessage()               {}
func (*Notice) Descriptor() ([]byte, []int) { return fileDescriptorNotice, []int{0} }

func init() {
	proto.RegisterType((*Notice)(nil), "edgeproto.Notice")
	proto.RegisterEnum("edgeproto.NoticeAction", NoticeAction_name, NoticeAction_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for NotifyApi service

type NotifyApiClient interface {
	// Bidrectional stream for exchanging data between controller and DME/CRM
	StreamNotice(ctx context.Context, opts ...grpc.CallOption) (NotifyApi_StreamNoticeClient, error)
}

type notifyApiClient struct {
	cc *grpc.ClientConn
}

func NewNotifyApiClient(cc *grpc.ClientConn) NotifyApiClient {
	return &notifyApiClient{cc}
}

func (c *notifyApiClient) StreamNotice(ctx context.Context, opts ...grpc.CallOption) (NotifyApi_StreamNoticeClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_NotifyApi_serviceDesc.Streams[0], c.cc, "/edgeproto.NotifyApi/StreamNotice", opts...)
	if err != nil {
		return nil, err
	}
	x := &notifyApiStreamNoticeClient{stream}
	return x, nil
}

type NotifyApi_StreamNoticeClient interface {
	Send(*Notice) error
	Recv() (*Notice, error)
	grpc.ClientStream
}

type notifyApiStreamNoticeClient struct {
	grpc.ClientStream
}

func (x *notifyApiStreamNoticeClient) Send(m *Notice) error {
	return x.ClientStream.SendMsg(m)
}

func (x *notifyApiStreamNoticeClient) Recv() (*Notice, error) {
	m := new(Notice)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for NotifyApi service

type NotifyApiServer interface {
	// Bidrectional stream for exchanging data between controller and DME/CRM
	StreamNotice(NotifyApi_StreamNoticeServer) error
}

func RegisterNotifyApiServer(s *grpc.Server, srv NotifyApiServer) {
	s.RegisterService(&_NotifyApi_serviceDesc, srv)
}

func _NotifyApi_StreamNotice_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(NotifyApiServer).StreamNotice(&notifyApiStreamNoticeServer{stream})
}

type NotifyApi_StreamNoticeServer interface {
	Send(*Notice) error
	Recv() (*Notice, error)
	grpc.ServerStream
}

type notifyApiStreamNoticeServer struct {
	grpc.ServerStream
}

func (x *notifyApiStreamNoticeServer) Send(m *Notice) error {
	return x.ServerStream.SendMsg(m)
}

func (x *notifyApiStreamNoticeServer) Recv() (*Notice, error) {
	m := new(Notice)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _NotifyApi_serviceDesc = grpc.ServiceDesc{
	ServiceName: "edgeproto.NotifyApi",
	HandlerType: (*NotifyApiServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamNotice",
			Handler:       _NotifyApi_StreamNotice_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "notice.proto",
}

func (m *Notice) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Notice) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Action != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintNotice(dAtA, i, uint64(m.Action))
	}
	if m.Version != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintNotice(dAtA, i, uint64(m.Version))
	}
	dAtA[i] = 0x1a
	i++
	i = encodeVarintNotice(dAtA, i, uint64(m.Any.Size()))
	n1, err := m.Any.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	if len(m.WantObjs) > 0 {
		for _, s := range m.WantObjs {
			dAtA[i] = 0x22
			i++
			l = len(s)
			for l >= 1<<7 {
				dAtA[i] = uint8(uint64(l)&0x7f | 0x80)
				l >>= 7
				i++
			}
			dAtA[i] = uint8(l)
			i++
			i += copy(dAtA[i:], s)
		}
	}
	if m.FilterCloudletKey {
		dAtA[i] = 0x28
		i++
		if m.FilterCloudletKey {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	if len(m.Span) > 0 {
		dAtA[i] = 0x32
		i++
		i = encodeVarintNotice(dAtA, i, uint64(len(m.Span)))
		i += copy(dAtA[i:], m.Span)
	}
	return i, nil
}

func encodeVarintNotice(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Notice) CopyInFields(src *Notice) int {
	changed := 0
	if m.Action != src.Action {
		m.Action = src.Action
		changed++
	}
	if m.Version != src.Version {
		m.Version = src.Version
		changed++
	}
	if m.Any.TypeUrl != src.Any.TypeUrl {
		m.Any.TypeUrl = src.Any.TypeUrl
		changed++
	}
	if m.Any.Value == nil || len(m.Any.Value) != len(src.Any.Value) {
		m.Any.Value = make([]byte, len(src.Any.Value))
		changed++
	}
	copy(m.Any.Value, src.Any.Value)
	changed++
	if m.WantObjs == nil || len(m.WantObjs) != len(src.WantObjs) {
		m.WantObjs = make([]string, len(src.WantObjs))
		changed++
	}
	copy(m.WantObjs, src.WantObjs)
	changed++
	if m.FilterCloudletKey != src.FilterCloudletKey {
		m.FilterCloudletKey = src.FilterCloudletKey
		changed++
	}
	if m.Span != src.Span {
		m.Span = src.Span
		changed++
	}
	return changed
}

// Helper method to check that enums have valid values
func (m *Notice) ValidateEnums() error {
	if _, ok := NoticeAction_name[int32(m.Action)]; !ok {
		return errors.New("invalid Action")
	}
	return nil
}

var NoticeActionStrings = []string{
	"NONE",
	"UPDATE",
	"DELETE",
	"VERSION",
	"SENDALL_END",
}

const (
	NoticeActionNONE        uint64 = 1 << 0
	NoticeActionUPDATE      uint64 = 1 << 1
	NoticeActionDELETE      uint64 = 1 << 2
	NoticeActionVERSION     uint64 = 1 << 3
	NoticeActionSENDALL_END uint64 = 1 << 4
)

var NoticeAction_CamelName = map[int32]string{
	// NONE -> None
	0: "None",
	// UPDATE -> Update
	1: "Update",
	// DELETE -> Delete
	2: "Delete",
	// VERSION -> Version
	3: "Version",
	// SENDALL_END -> SendallEnd
	4: "SendallEnd",
}
var NoticeAction_CamelValue = map[string]int32{
	"None":       0,
	"Update":     1,
	"Delete":     2,
	"Version":    3,
	"SendallEnd": 4,
}

func (e *NoticeAction) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	val, ok := NoticeAction_CamelValue[util.CamelCase(str)]
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = NoticeAction_CamelName[val]
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("No enum value for %s", str))
	}
	*e = NoticeAction(val)
	return nil
}

func (e NoticeAction) MarshalYAML() (interface{}, error) {
	return proto.EnumName(NoticeAction_CamelName, int32(e)), nil
}

// custom JSON encoding/decoding
func (e *NoticeAction) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		val, ok := NoticeAction_CamelValue[util.CamelCase(str)]
		if !ok {
			// may be int value instead of enum name
			ival, err := strconv.Atoi(str)
			val = int32(ival)
			if err == nil {
				_, ok = NoticeAction_CamelName[val]
			}
		}
		if !ok {
			return errors.New(fmt.Sprintf("No enum value for %s", str))
		}
		*e = NoticeAction(val)
		return nil
	}
	var val int32
	err = json.Unmarshal(b, &val)
	if err == nil {
		*e = NoticeAction(val)
		return nil
	}
	return fmt.Errorf("No enum value for %v", b)
}
func (m *Notice) Size() (n int) {
	var l int
	_ = l
	if m.Action != 0 {
		n += 1 + sovNotice(uint64(m.Action))
	}
	if m.Version != 0 {
		n += 1 + sovNotice(uint64(m.Version))
	}
	l = m.Any.Size()
	n += 1 + l + sovNotice(uint64(l))
	if len(m.WantObjs) > 0 {
		for _, s := range m.WantObjs {
			l = len(s)
			n += 1 + l + sovNotice(uint64(l))
		}
	}
	if m.FilterCloudletKey {
		n += 2
	}
	l = len(m.Span)
	if l > 0 {
		n += 1 + l + sovNotice(uint64(l))
	}
	return n
}

func sovNotice(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozNotice(x uint64) (n int) {
	return sovNotice(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Notice) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowNotice
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
			return fmt.Errorf("proto: Notice: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Notice: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Action", wireType)
			}
			m.Action = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNotice
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Action |= (NoticeAction(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Version", wireType)
			}
			m.Version = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNotice
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Version |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Any", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNotice
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
				return ErrInvalidLengthNotice
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Any.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field WantObjs", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNotice
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
				return ErrInvalidLengthNotice
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.WantObjs = append(m.WantObjs, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field FilterCloudletKey", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNotice
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.FilterCloudletKey = bool(v != 0)
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Span", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNotice
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
				return ErrInvalidLengthNotice
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Span = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipNotice(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthNotice
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
func skipNotice(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowNotice
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
					return 0, ErrIntOverflowNotice
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
					return 0, ErrIntOverflowNotice
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
				return 0, ErrInvalidLengthNotice
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowNotice
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
				next, err := skipNotice(dAtA[start:])
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
	ErrInvalidLengthNotice = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowNotice   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("notice.proto", fileDescriptorNotice) }

var fileDescriptorNotice = []byte{
	// 364 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x90, 0x51, 0x6f, 0xa2, 0x40,
	0x14, 0x85, 0x19, 0x61, 0x51, 0xae, 0xee, 0x2e, 0xce, 0x9a, 0x2c, 0xeb, 0x26, 0x2c, 0xf1, 0x89,
	0x6c, 0x36, 0xb0, 0xb1, 0x6f, 0x7d, 0xc3, 0x32, 0x69, 0x9a, 0x1a, 0x6c, 0xd0, 0xf6, 0x95, 0x80,
	0x8e, 0x04, 0x4b, 0x19, 0x03, 0xd8, 0x86, 0x7f, 0xe8, 0x63, 0x7f, 0x41, 0xd3, 0xf2, 0x4b, 0x1a,
	0x40, 0x9b, 0x26, 0xbe, 0x9d, 0x7b, 0xce, 0x37, 0x27, 0x77, 0x2e, 0xf4, 0x12, 0x96, 0x47, 0x4b,
	0x6a, 0x6c, 0x53, 0x96, 0x33, 0x2c, 0xd1, 0x55, 0x48, 0x6b, 0x39, 0xfc, 0x15, 0x32, 0x16, 0xc6,
	0xd4, 0xac, 0xa7, 0x60, 0xb7, 0x36, 0xfd, 0xa4, 0x68, 0xa8, 0xe1, 0x20, 0x64, 0x21, 0xab, 0xa5,
	0x59, 0xa9, 0xc6, 0x1d, 0x95, 0x08, 0x44, 0xa7, 0x2e, 0xc3, 0x26, 0x88, 0xfe, 0x32, 0x8f, 0x58,
	0xa2, 0x20, 0x0d, 0xe9, 0xdf, 0xc6, 0x3f, 0x8d, 0x8f, 0x5e, 0xa3, 0x41, 0xac, 0x3a, 0x76, 0x0f,
	0x18, 0x56, 0xa0, 0xfd, 0x48, 0xd3, 0xac, 0x7a, 0xd1, 0xd2, 0x90, 0xfe, 0xd5, 0x3d, 0x8e, 0xf8,
	0x1f, 0xf0, 0x7e, 0x52, 0x28, 0xbc, 0x86, 0xf4, 0xee, 0x78, 0x60, 0x34, 0x4b, 0x19, 0xc7, 0xa5,
	0x0c, 0x2b, 0x29, 0x26, 0xc2, 0xfe, 0xe5, 0x0f, 0xe7, 0x56, 0x18, 0xfe, 0x0d, 0xd2, 0x93, 0x9f,
	0xe4, 0x1e, 0x0b, 0x36, 0x99, 0x22, 0x68, 0xbc, 0x2e, 0xb9, 0x9d, 0xca, 0x98, 0x05, 0x9b, 0x0c,
	0x1b, 0xf0, 0x63, 0x1d, 0xc5, 0x39, 0x4d, 0xbd, 0x65, 0xcc, 0x76, 0xab, 0x98, 0xe6, 0xde, 0x3d,
	0x2d, 0x94, 0x2f, 0x1a, 0xd2, 0x3b, 0x6e, 0xbf, 0x89, 0x2e, 0x0e, 0xc9, 0x35, 0x2d, 0x30, 0x06,
	0x21, 0xdb, 0xfa, 0x89, 0x22, 0x6a, 0x48, 0x97, 0xdc, 0x5a, 0xff, 0x75, 0xa0, 0xf7, 0xf9, 0x03,
	0xb8, 0x03, 0x82, 0x33, 0x73, 0x88, 0xcc, 0x61, 0x00, 0xf1, 0xf6, 0xc6, 0xb6, 0x16, 0x44, 0x46,
	0x95, 0xb6, 0xc9, 0x94, 0x2c, 0x88, 0xdc, 0xc2, 0x5d, 0x68, 0xdf, 0x11, 0x77, 0x7e, 0x35, 0x73,
	0x64, 0x1e, 0x7f, 0x87, 0xee, 0x9c, 0x38, 0xb6, 0x35, 0x9d, 0x7a, 0xc4, 0xb1, 0x65, 0x61, 0x7c,
	0x09, 0x52, 0xd5, 0xb7, 0x2e, 0xac, 0x6d, 0x84, 0xcf, 0xa1, 0x37, 0xcf, 0x53, 0xea, 0x3f, 0x1c,
	0xce, 0xd8, 0x3f, 0x39, 0xdb, 0xf0, 0xd4, 0x1a, 0x71, 0x3a, 0xfa, 0x8f, 0x26, 0xf2, 0xfe, 0x4d,
	0xe5, 0xf6, 0xa5, 0x8a, 0x9e, 0x4b, 0x15, 0xbd, 0x96, 0x2a, 0x0a, 0xc4, 0x9a, 0x3a, 0x7b, 0x0f,
	0x00, 0x00, 0xff, 0xff, 0x7b, 0xe1, 0x48, 0x9d, 0xe2, 0x01, 0x00, 0x00,
}
