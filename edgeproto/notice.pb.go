// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: notice.proto

package edgeproto

import (
	context "context"
	"encoding/json"
	"errors"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	types "github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/util"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

func (NoticeAction) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_642492014393dbdb, []int{0}
}

type Notice struct {
	// Action to perform
	Action NoticeAction `protobuf:"varint,1,opt,name=action,proto3,enum=edgeproto.NoticeAction" json:"action,omitempty"`
	// Protocol version supported by sender
	Version uint32 `protobuf:"varint,2,opt,name=version,proto3" json:"version,omitempty"`
	// Data
	Any types.Any `protobuf:"bytes,3,opt,name=any,proto3" json:"any"`
	// Wanted Objects
	WantObjs []string `protobuf:"bytes,4,rep,name=want_objs,json=wantObjs,proto3" json:"want_objs,omitempty"`
	// Filter by cloudlet key
	FilterCloudletKey bool `protobuf:"varint,5,opt,name=filter_cloudlet_key,json=filterCloudletKey,proto3" json:"filter_cloudlet_key,omitempty"`
	// Opentracing span
	Span string `protobuf:"bytes,6,opt,name=span,proto3" json:"span,omitempty"`
	// Database revision for which object was last modified
	ModRev int64 `protobuf:"varint,7,opt,name=mod_rev,json=modRev,proto3" json:"mod_rev,omitempty"`
	// Extra tags
	Tags map[string]string `protobuf:"bytes,8,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (m *Notice) Reset()         { *m = Notice{} }
func (m *Notice) String() string { return proto.CompactTextString(m) }
func (*Notice) ProtoMessage()    {}
func (*Notice) Descriptor() ([]byte, []int) {
	return fileDescriptor_642492014393dbdb, []int{0}
}
func (m *Notice) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Notice) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Notice.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Notice) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Notice.Merge(m, src)
}
func (m *Notice) XXX_Size() int {
	return m.Size()
}
func (m *Notice) XXX_DiscardUnknown() {
	xxx_messageInfo_Notice.DiscardUnknown(m)
}

var xxx_messageInfo_Notice proto.InternalMessageInfo

func init() {
	proto.RegisterEnum("edgeproto.NoticeAction", NoticeAction_name, NoticeAction_value)
	proto.RegisterType((*Notice)(nil), "edgeproto.Notice")
	proto.RegisterMapType((map[string]string)(nil), "edgeproto.Notice.TagsEntry")
}

func init() { proto.RegisterFile("notice.proto", fileDescriptor_642492014393dbdb) }

var fileDescriptor_642492014393dbdb = []byte{
	// 455 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x91, 0x41, 0x8f, 0xd2, 0x40,
	0x14, 0xc7, 0x3b, 0xb4, 0x5b, 0xe8, 0x03, 0xb5, 0x3b, 0x92, 0x6c, 0x65, 0x4d, 0x6d, 0xf6, 0xd4,
	0x18, 0xd3, 0x1a, 0x3c, 0x68, 0xf6, 0xd6, 0x95, 0xc6, 0x18, 0x49, 0x31, 0x03, 0x7a, 0x6d, 0x06,
	0x18, 0x1a, 0xd6, 0xd2, 0x21, 0xed, 0x80, 0xe9, 0xb7, 0xf0, 0x63, 0x71, 0xdc, 0xa3, 0x27, 0xa3,
	0x10, 0xbf, 0x87, 0xe9, 0x94, 0xdd, 0x98, 0x70, 0xfb, 0xbd, 0xff, 0xfb, 0xbf, 0x37, 0xef, 0xbd,
	0x81, 0x4e, 0xc6, 0xc5, 0x72, 0xc6, 0xbc, 0x75, 0xce, 0x05, 0xc7, 0x06, 0x9b, 0x27, 0x4c, 0x62,
	0xef, 0x59, 0xc2, 0x79, 0x92, 0x32, 0x5f, 0x46, 0xd3, 0xcd, 0xc2, 0xa7, 0x59, 0x59, 0xbb, 0x7a,
	0xdd, 0x84, 0x27, 0x5c, 0xa2, 0x5f, 0x51, 0xad, 0x5e, 0xfd, 0x6d, 0x80, 0x1e, 0xc9, 0x66, 0xd8,
	0x07, 0x9d, 0xce, 0xc4, 0x92, 0x67, 0x16, 0x72, 0x90, 0xfb, 0xb8, 0x7f, 0xe1, 0x3d, 0xf4, 0xf5,
	0x6a, 0x4b, 0x20, 0xd3, 0xe4, 0x68, 0xc3, 0x16, 0x34, 0xb7, 0x2c, 0x2f, 0xaa, 0x8a, 0x86, 0x83,
	0xdc, 0x47, 0xe4, 0x3e, 0xc4, 0xaf, 0x40, 0xa5, 0x59, 0x69, 0xa9, 0x0e, 0x72, 0xdb, 0xfd, 0xae,
	0x57, 0x0f, 0xe5, 0xdd, 0x0f, 0xe5, 0x05, 0x59, 0x79, 0xa3, 0xed, 0x7e, 0xbd, 0x50, 0x48, 0x65,
	0xc3, 0x97, 0x60, 0x7c, 0xa7, 0x99, 0x88, 0xf9, 0xf4, 0xb6, 0xb0, 0x34, 0x47, 0x75, 0x0d, 0xd2,
	0xaa, 0x84, 0xd1, 0xf4, 0xb6, 0xc0, 0x1e, 0x3c, 0x5d, 0x2c, 0x53, 0xc1, 0xf2, 0x78, 0x96, 0xf2,
	0xcd, 0x3c, 0x65, 0x22, 0xfe, 0xc6, 0x4a, 0xeb, 0xcc, 0x41, 0x6e, 0x8b, 0x9c, 0xd7, 0xa9, 0xf7,
	0xc7, 0xcc, 0x27, 0x56, 0x62, 0x0c, 0x5a, 0xb1, 0xa6, 0x99, 0xa5, 0x3b, 0xc8, 0x35, 0x88, 0x64,
	0x7c, 0x01, 0xcd, 0x15, 0x9f, 0xc7, 0x39, 0xdb, 0x5a, 0x4d, 0x07, 0xb9, 0x2a, 0xd1, 0x57, 0x7c,
	0x4e, 0xd8, 0x16, 0xfb, 0xa0, 0x09, 0x9a, 0x14, 0x56, 0xcb, 0x51, 0xdd, 0x76, 0xff, 0xf2, 0x64,
	0x61, 0x6f, 0x42, 0x93, 0x22, 0xcc, 0x44, 0x5e, 0x12, 0x69, 0xec, 0xbd, 0x05, 0xe3, 0x41, 0xc2,
	0x26, 0xa8, 0xd5, 0x28, 0x48, 0xbe, 0x54, 0x21, 0xee, 0xc2, 0xd9, 0x96, 0xa6, 0x1b, 0x26, 0xef,
	0x61, 0x90, 0x3a, 0xb8, 0x6e, 0xbc, 0x43, 0x2f, 0x23, 0xe8, 0xfc, 0x7f, 0x43, 0xdc, 0x02, 0x2d,
	0x1a, 0x45, 0xa1, 0xa9, 0x60, 0x00, 0xfd, 0xcb, 0xe7, 0x41, 0x30, 0x09, 0x4d, 0x54, 0xf1, 0x20,
	0x1c, 0x86, 0x93, 0xd0, 0x6c, 0xe0, 0x36, 0x34, 0xbf, 0x86, 0x64, 0xfc, 0x71, 0x14, 0x99, 0x2a,
	0x7e, 0x02, 0xed, 0x71, 0x18, 0x0d, 0x82, 0xe1, 0x30, 0x0e, 0xa3, 0x81, 0xa9, 0xf5, 0x3f, 0x80,
	0x51, 0xf5, 0x5b, 0x94, 0xc1, 0x7a, 0x89, 0xaf, 0xa1, 0x33, 0x16, 0x39, 0xa3, 0xab, 0xe3, 0x4f,
	0x9e, 0x9f, 0x2c, 0xd2, 0x3b, 0x95, 0xae, 0x14, 0x17, 0xbd, 0x46, 0x37, 0xcf, 0x77, 0x7f, 0x6c,
	0x65, 0xb7, 0xb7, 0xd1, 0xdd, 0xde, 0x46, 0xbf, 0xf7, 0x36, 0xfa, 0x71, 0xb0, 0x95, 0xbb, 0x83,
	0xad, 0xfc, 0x3c, 0xd8, 0xca, 0x54, 0x97, 0x15, 0x6f, 0xfe, 0x05, 0x00, 0x00, 0xff, 0xff, 0x14,
	0x4b, 0xa4, 0x82, 0x71, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// NotifyApiClient is the client API for NotifyApi service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
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
	stream, err := c.cc.NewStream(ctx, &_NotifyApi_serviceDesc.Streams[0], "/edgeproto.NotifyApi/StreamNotice", opts...)
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

// NotifyApiServer is the server API for NotifyApi service.
type NotifyApiServer interface {
	// Bidrectional stream for exchanging data between controller and DME/CRM
	StreamNotice(NotifyApi_StreamNoticeServer) error
}

// UnimplementedNotifyApiServer can be embedded to have forward compatible implementations.
type UnimplementedNotifyApiServer struct {
}

func (*UnimplementedNotifyApiServer) StreamNotice(srv NotifyApi_StreamNoticeServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamNotice not implemented")
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
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Notice) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Notice) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Tags) > 0 {
		for k := range m.Tags {
			v := m.Tags[k]
			baseI := i
			i -= len(v)
			copy(dAtA[i:], v)
			i = encodeVarintNotice(dAtA, i, uint64(len(v)))
			i--
			dAtA[i] = 0x12
			i -= len(k)
			copy(dAtA[i:], k)
			i = encodeVarintNotice(dAtA, i, uint64(len(k)))
			i--
			dAtA[i] = 0xa
			i = encodeVarintNotice(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0x42
		}
	}
	if m.ModRev != 0 {
		i = encodeVarintNotice(dAtA, i, uint64(m.ModRev))
		i--
		dAtA[i] = 0x38
	}
	if len(m.Span) > 0 {
		i -= len(m.Span)
		copy(dAtA[i:], m.Span)
		i = encodeVarintNotice(dAtA, i, uint64(len(m.Span)))
		i--
		dAtA[i] = 0x32
	}
	if m.FilterCloudletKey {
		i--
		if m.FilterCloudletKey {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x28
	}
	if len(m.WantObjs) > 0 {
		for iNdEx := len(m.WantObjs) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.WantObjs[iNdEx])
			copy(dAtA[i:], m.WantObjs[iNdEx])
			i = encodeVarintNotice(dAtA, i, uint64(len(m.WantObjs[iNdEx])))
			i--
			dAtA[i] = 0x22
		}
	}
	{
		size, err := m.Any.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintNotice(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if m.Version != 0 {
		i = encodeVarintNotice(dAtA, i, uint64(m.Version))
		i--
		dAtA[i] = 0x10
	}
	if m.Action != 0 {
		i = encodeVarintNotice(dAtA, i, uint64(m.Action))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintNotice(dAtA []byte, offset int, v uint64) int {
	offset -= sovNotice(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
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
	if src.Any.Value != nil {
		m.Any.Value = src.Any.Value
		changed++
	}
	if src.WantObjs != nil {
		m.WantObjs = src.WantObjs
		changed++
	} else if m.WantObjs != nil {
		m.WantObjs = nil
		changed++
	}
	if m.FilterCloudletKey != src.FilterCloudletKey {
		m.FilterCloudletKey = src.FilterCloudletKey
		changed++
	}
	if m.Span != src.Span {
		m.Span = src.Span
		changed++
	}
	if m.ModRev != src.ModRev {
		m.ModRev = src.ModRev
		changed++
	}
	if src.Tags != nil {
		m.Tags = make(map[string]string)
		for k0, _ := range src.Tags {
			m.Tags[k0] = src.Tags[k0]
			changed++
		}
	} else if m.Tags != nil {
		m.Tags = nil
		changed++
	}
	return changed
}

func (m *Notice) DeepCopyIn(src *Notice) {
	m.Action = src.Action
	m.Version = src.Version
	m.Any = src.Any
	if src.WantObjs != nil {
		m.WantObjs = make([]string, len(src.WantObjs), len(src.WantObjs))
		for ii, s := range src.WantObjs {
			m.WantObjs[ii] = s
		}
	} else {
		m.WantObjs = nil
	}
	m.FilterCloudletKey = src.FilterCloudletKey
	m.Span = src.Span
	m.ModRev = src.ModRev
	if src.Tags != nil {
		m.Tags = make(map[string]string)
		for k, v := range src.Tags {
			m.Tags[k] = v
		}
	} else {
		m.Tags = nil
	}
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
		return fmt.Errorf("Invalid NoticeAction value %q", str)
	}
	*e = NoticeAction(val)
	return nil
}

func (e NoticeAction) MarshalYAML() (interface{}, error) {
	str := proto.EnumName(NoticeAction_CamelName, int32(e))
	return str, nil
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
			return fmt.Errorf("Invalid NoticeAction value %q", str)
		}
		*e = NoticeAction(val)
		return nil
	}
	var val int32
	err = json.Unmarshal(b, &val)
	if err == nil {
		_, ok := NoticeAction_CamelName[val]
		if !ok {
			return fmt.Errorf("Invalid NoticeAction value %d", val)
		}
		*e = NoticeAction(val)
		return nil
	}
	return fmt.Errorf("Invalid NoticeAction value %v", b)
}

/*
 * This is removed because we do not have enough time in
 * release 3.0 to update the SDK, UI, and documentation for this
 * change. It should be done in 3.1.
func (e NoticeAction) MarshalJSON() ([]byte, error) {
	str := proto.EnumName(NoticeAction_CamelName, int32(e))
	return json.Marshal(str)
}
*/
func (m *Notice) IsValidArgsForStreamNotice() error {
	return nil
}

func (m *Notice) Size() (n int) {
	if m == nil {
		return 0
	}
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
	if m.ModRev != 0 {
		n += 1 + sovNotice(uint64(m.ModRev))
	}
	if len(m.Tags) > 0 {
		for k, v := range m.Tags {
			_ = k
			_ = v
			mapEntrySize := 1 + len(k) + sovNotice(uint64(len(k))) + 1 + len(v) + sovNotice(uint64(len(v)))
			n += mapEntrySize + 1 + sovNotice(uint64(mapEntrySize))
		}
	}
	return n
}

func sovNotice(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
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
			wire |= uint64(b&0x7F) << shift
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
				m.Action |= NoticeAction(b&0x7F) << shift
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
				m.Version |= uint32(b&0x7F) << shift
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
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthNotice
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthNotice
			}
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
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthNotice
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNotice
			}
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
				v |= int(b&0x7F) << shift
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
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthNotice
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNotice
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Span = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ModRev", wireType)
			}
			m.ModRev = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNotice
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ModRev |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Tags", wireType)
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
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthNotice
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthNotice
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Tags == nil {
				m.Tags = make(map[string]string)
			}
			var mapkey string
			var mapvalue string
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
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
					wire |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				fieldNum := int32(wire >> 3)
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowNotice
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return ErrInvalidLengthNotice
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey < 0 {
						return ErrInvalidLengthNotice
					}
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var stringLenmapvalue uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowNotice
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapvalue |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapvalue := int(stringLenmapvalue)
					if intStringLenmapvalue < 0 {
						return ErrInvalidLengthNotice
					}
					postStringIndexmapvalue := iNdEx + intStringLenmapvalue
					if postStringIndexmapvalue < 0 {
						return ErrInvalidLengthNotice
					}
					if postStringIndexmapvalue > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = string(dAtA[iNdEx:postStringIndexmapvalue])
					iNdEx = postStringIndexmapvalue
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipNotice(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if skippy < 0 {
						return ErrInvalidLengthNotice
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.Tags[mapkey] = mapvalue
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
			if (iNdEx + skippy) < 0 {
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
	depth := 0
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
		case 1:
			iNdEx += 8
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
			if length < 0 {
				return 0, ErrInvalidLengthNotice
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupNotice
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthNotice
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthNotice        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowNotice          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupNotice = fmt.Errorf("proto: unexpected end of group")
)
