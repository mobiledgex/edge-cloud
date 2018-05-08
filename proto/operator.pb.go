// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: operator.proto

package proto

import proto1 "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/gogo/protobuf/gogoproto"

import strings "strings"
import reflect "reflect"

import context "golang.org/x/net/context"
import grpc "google.golang.org/grpc"

import "encoding/json"
import "github.com/mobiledgex/edge-cloud/util"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type OperatorCode struct {
	// Operator code consists of two pars, a mobile network code (MNC)
	// and a mobile country code (MCC). These are strings instead of
	// integers to preserve leading zeros which have meaning.
	// A single operator (like AT&T) may have multiple operator codes
	// across countries and in the same country for different wireless bands.
	MNC string `protobuf:"bytes,1,opt,name=MNC,proto3" json:"MNC,omitempty"`
	MCC string `protobuf:"bytes,2,opt,name=MCC,proto3" json:"MCC,omitempty"`
}

func (m *OperatorCode) Reset()                    { *m = OperatorCode{} }
func (m *OperatorCode) String() string            { return proto1.CompactTextString(m) }
func (*OperatorCode) ProtoMessage()               {}
func (*OperatorCode) Descriptor() ([]byte, []int) { return fileDescriptorOperator, []int{0} }

type OperatorKey struct {
	// Company or Organization name of the operator
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (m *OperatorKey) Reset()                    { *m = OperatorKey{} }
func (m *OperatorKey) String() string            { return proto1.CompactTextString(m) }
func (*OperatorKey) ProtoMessage()               {}
func (*OperatorKey) Descriptor() ([]byte, []int) { return fileDescriptorOperator, []int{1} }

type Operator struct {
	Fields []byte `protobuf:"bytes,1,opt,name=fields,proto3" json:"fields,omitempty"`
	// Unique identifier key
	Key OperatorKey `protobuf:"bytes,2,opt,name=key" json:"key"`
}

func (m *Operator) Reset()                    { *m = Operator{} }
func (m *Operator) String() string            { return proto1.CompactTextString(m) }
func (*Operator) ProtoMessage()               {}
func (*Operator) Descriptor() ([]byte, []int) { return fileDescriptorOperator, []int{2} }

func init() {
	proto1.RegisterType((*OperatorCode)(nil), "proto.OperatorCode")
	proto1.RegisterType((*OperatorKey)(nil), "proto.OperatorKey")
	proto1.RegisterType((*Operator)(nil), "proto.Operator")
}
func (this *OperatorKey) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 5)
	s = append(s, "&proto.OperatorKey{")
	s = append(s, "Name: "+fmt.Sprintf("%#v", this.Name)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringOperator(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for OperatorApi service

type OperatorApiClient interface {
	CreateOperator(ctx context.Context, in *Operator, opts ...grpc.CallOption) (*Result, error)
	DeleteOperator(ctx context.Context, in *Operator, opts ...grpc.CallOption) (*Result, error)
	UpdateOperator(ctx context.Context, in *Operator, opts ...grpc.CallOption) (*Result, error)
	ShowOperator(ctx context.Context, in *Operator, opts ...grpc.CallOption) (OperatorApi_ShowOperatorClient, error)
}

type operatorApiClient struct {
	cc *grpc.ClientConn
}

func NewOperatorApiClient(cc *grpc.ClientConn) OperatorApiClient {
	return &operatorApiClient{cc}
}

func (c *operatorApiClient) CreateOperator(ctx context.Context, in *Operator, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/proto.OperatorApi/CreateOperator", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *operatorApiClient) DeleteOperator(ctx context.Context, in *Operator, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/proto.OperatorApi/DeleteOperator", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *operatorApiClient) UpdateOperator(ctx context.Context, in *Operator, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/proto.OperatorApi/UpdateOperator", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *operatorApiClient) ShowOperator(ctx context.Context, in *Operator, opts ...grpc.CallOption) (OperatorApi_ShowOperatorClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_OperatorApi_serviceDesc.Streams[0], c.cc, "/proto.OperatorApi/ShowOperator", opts...)
	if err != nil {
		return nil, err
	}
	x := &operatorApiShowOperatorClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type OperatorApi_ShowOperatorClient interface {
	Recv() (*Operator, error)
	grpc.ClientStream
}

type operatorApiShowOperatorClient struct {
	grpc.ClientStream
}

func (x *operatorApiShowOperatorClient) Recv() (*Operator, error) {
	m := new(Operator)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for OperatorApi service

type OperatorApiServer interface {
	CreateOperator(context.Context, *Operator) (*Result, error)
	DeleteOperator(context.Context, *Operator) (*Result, error)
	UpdateOperator(context.Context, *Operator) (*Result, error)
	ShowOperator(*Operator, OperatorApi_ShowOperatorServer) error
}

func RegisterOperatorApiServer(s *grpc.Server, srv OperatorApiServer) {
	s.RegisterService(&_OperatorApi_serviceDesc, srv)
}

func _OperatorApi_CreateOperator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Operator)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OperatorApiServer).CreateOperator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.OperatorApi/CreateOperator",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OperatorApiServer).CreateOperator(ctx, req.(*Operator))
	}
	return interceptor(ctx, in, info, handler)
}

func _OperatorApi_DeleteOperator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Operator)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OperatorApiServer).DeleteOperator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.OperatorApi/DeleteOperator",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OperatorApiServer).DeleteOperator(ctx, req.(*Operator))
	}
	return interceptor(ctx, in, info, handler)
}

func _OperatorApi_UpdateOperator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Operator)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OperatorApiServer).UpdateOperator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.OperatorApi/UpdateOperator",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OperatorApiServer).UpdateOperator(ctx, req.(*Operator))
	}
	return interceptor(ctx, in, info, handler)
}

func _OperatorApi_ShowOperator_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Operator)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(OperatorApiServer).ShowOperator(m, &operatorApiShowOperatorServer{stream})
}

type OperatorApi_ShowOperatorServer interface {
	Send(*Operator) error
	grpc.ServerStream
}

type operatorApiShowOperatorServer struct {
	grpc.ServerStream
}

func (x *operatorApiShowOperatorServer) Send(m *Operator) error {
	return x.ServerStream.SendMsg(m)
}

var _OperatorApi_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.OperatorApi",
	HandlerType: (*OperatorApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateOperator",
			Handler:    _OperatorApi_CreateOperator_Handler,
		},
		{
			MethodName: "DeleteOperator",
			Handler:    _OperatorApi_DeleteOperator_Handler,
		},
		{
			MethodName: "UpdateOperator",
			Handler:    _OperatorApi_UpdateOperator_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ShowOperator",
			Handler:       _OperatorApi_ShowOperator_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "operator.proto",
}

func (m *OperatorCode) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *OperatorCode) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.MNC) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintOperator(dAtA, i, uint64(len(m.MNC)))
		i += copy(dAtA[i:], m.MNC)
	}
	if len(m.MCC) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintOperator(dAtA, i, uint64(len(m.MCC)))
		i += copy(dAtA[i:], m.MCC)
	}
	return i, nil
}

func (m *OperatorKey) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *OperatorKey) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintOperator(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	return i, nil
}

func (m *Operator) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Operator) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Fields) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintOperator(dAtA, i, uint64(len(m.Fields)))
		i += copy(dAtA[i:], m.Fields)
	}
	dAtA[i] = 0x12
	i++
	i = encodeVarintOperator(dAtA, i, uint64(m.Key.Size()))
	n1, err := m.Key.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	return i, nil
}

func encodeVarintOperator(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *OperatorCode) Matches(filter *OperatorCode) bool {
	if filter == nil {
		return true
	}
	if filter.MNC != "" && filter.MNC != m.MNC {
		return false
	}
	if filter.MCC != "" && filter.MCC != m.MCC {
		return false
	}
	return true
}

func (m *OperatorCode) CopyInFields(src *OperatorCode) {
	m.MNC = src.MNC
	m.MCC = src.MCC
}

func (m *OperatorKey) Matches(filter *OperatorKey) bool {
	if filter == nil {
		return true
	}
	if filter.Name != "" && filter.Name != m.Name {
		return false
	}
	return true
}

func (m *OperatorKey) CopyInFields(src *OperatorKey) {
	m.Name = src.Name
}

func (m *OperatorKey) GetKeyString() string {
	s := make([]string, 0, 1)
	s = append(s, m.Name)
	return strings.Join(s, "/")
}

func (m *Operator) Matches(filter *Operator) bool {
	if filter == nil {
		return true
	}
	if !m.Key.Matches(&filter.Key) {
		return false
	}
	return true
}

const OperatorFieldFields uint = 1
const OperatorFieldKey uint = 2

func (m *Operator) CopyInFields(src *Operator) {
	if set, _ := util.GrpcFieldsGet(src.Fields, 2); set == true {
		m.Key = src.Key
	}
}

type OperatorCud interface {
	// Validate all fields for create/update
	Validate(in *Operator) error
	// Validate only key fields for delete
	ValidateKey(key *OperatorKey) error
	// Get key string for saving to persistent object storage
	GetObjStoreKeyString(key *OperatorKey) string
	// Object storage IO interface
	ObjStore
	// Refresh is called after create/update/delete to update in-memory cache
	Refresh(in *Operator, key string) error
	// Get key string for loading all objects of this type
	GetLoadKeyString() string
}

func (m *Operator) Create(cud OperatorCud) (*Result, error) {
	err := cud.Validate(m)
	if err != nil {
		return nil, err
	}
	key := cud.GetObjStoreKeyString(&m.Key)
	val, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	err = cud.Create(key, string(val))
	if err != nil {
		return nil, err
	}
	err = cud.Refresh(m, key)
	return &Result{}, err
}

func (m *Operator) Update(cud OperatorCud) (*Result, error) {
	err := cud.Validate(m)
	if err != nil {
		return nil, err
	}
	key := cud.GetObjStoreKeyString(&m.Key)
	var vers int64 = 0
	curBytes, vers, err := cud.Get(key)
	if err != nil {
		return nil, err
	}
	var cur Operator
	err = json.Unmarshal(curBytes, &cur)
	if err != nil {
		return nil, err
	}
	cur.CopyInFields(m)
	// never save fields
	cur.Fields = nil
	val, err := json.Marshal(cur)
	if err != nil {
		return nil, err
	}
	err = cud.Update(key, string(val), vers)
	if err != nil {
		return nil, err
	}
	err = cud.Refresh(m, key)
	return &Result{}, err
}

func (m *Operator) Delete(cud OperatorCud) (*Result, error) {
	err := cud.ValidateKey(&m.Key)
	if err != nil {
		return nil, err
	}
	key := cud.GetObjStoreKeyString(&m.Key)
	err = cud.Delete(key)
	if err != nil {
		return nil, err
	}
	err = cud.Refresh(m, key)
	return &Result{}, err
}

type LoadAllOperatorsCb func(m *Operator) error

func LoadAllOperators(cud OperatorCud, cb LoadAllOperatorsCb) error {
	loadkey := cud.GetLoadKeyString()
	err := cud.List(loadkey, func(key, val []byte) error {
		var obj Operator
		err := json.Unmarshal(val, &obj)
		if err != nil {
			util.WarnLog("Failed to parse Operator data", "val", string(val))
			return nil
		}
		err = cb(&obj)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func LoadOneOperator(cud OperatorCud, key string) (*Operator, error) {
	val, _, err := cud.Get(key)
	if err != nil {
		return nil, err
	}
	var obj Operator
	err = json.Unmarshal(val, &obj)
	if err != nil {
		util.DebugLog(util.DebugLevelApi, "Failed to parse Operator data", "val", string(val))
		return nil, err
	}
	return &obj, nil
}

func (m *Operator) GetKey() ObjKey {
	return &m.Key
}

func (m *OperatorCode) Size() (n int) {
	var l int
	_ = l
	l = len(m.MNC)
	if l > 0 {
		n += 1 + l + sovOperator(uint64(l))
	}
	l = len(m.MCC)
	if l > 0 {
		n += 1 + l + sovOperator(uint64(l))
	}
	return n
}

func (m *OperatorKey) Size() (n int) {
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovOperator(uint64(l))
	}
	return n
}

func (m *Operator) Size() (n int) {
	var l int
	_ = l
	l = len(m.Fields)
	if l > 0 {
		n += 1 + l + sovOperator(uint64(l))
	}
	l = m.Key.Size()
	n += 1 + l + sovOperator(uint64(l))
	return n
}

func sovOperator(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozOperator(x uint64) (n int) {
	return sovOperator(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *OperatorCode) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowOperator
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
			return fmt.Errorf("proto: OperatorCode: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: OperatorCode: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MNC", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOperator
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
				return ErrInvalidLengthOperator
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MNC = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MCC", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOperator
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
				return ErrInvalidLengthOperator
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MCC = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipOperator(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthOperator
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
func (m *OperatorKey) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowOperator
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
			return fmt.Errorf("proto: OperatorKey: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: OperatorKey: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOperator
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
				return ErrInvalidLengthOperator
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipOperator(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthOperator
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
func (m *Operator) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowOperator
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
			return fmt.Errorf("proto: Operator: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Operator: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Fields", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOperator
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
				return ErrInvalidLengthOperator
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Fields = append(m.Fields[:0], dAtA[iNdEx:postIndex]...)
			if m.Fields == nil {
				m.Fields = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Key", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOperator
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
				return ErrInvalidLengthOperator
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Key.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipOperator(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthOperator
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
func skipOperator(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowOperator
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
					return 0, ErrIntOverflowOperator
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
					return 0, ErrIntOverflowOperator
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
				return 0, ErrInvalidLengthOperator
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowOperator
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
				next, err := skipOperator(dAtA[start:])
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
	ErrInvalidLengthOperator = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowOperator   = fmt.Errorf("proto: integer overflow")
)

func init() { proto1.RegisterFile("operator.proto", fileDescriptorOperator) }

var fileDescriptorOperator = []byte{
	// 397 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x91, 0xcf, 0x4a, 0xe3, 0x50,
	0x14, 0x87, 0x7b, 0xdb, 0x4e, 0xe9, 0xdc, 0x66, 0x3a, 0xe5, 0x0e, 0x23, 0xb1, 0x4a, 0x2a, 0x59,
	0x49, 0xa1, 0xbd, 0x52, 0x37, 0x25, 0x3b, 0x1b, 0x77, 0x52, 0x85, 0x88, 0xe0, 0x36, 0x69, 0x6e,
	0xd3, 0x60, 0x9a, 0x13, 0xf2, 0x87, 0xda, 0xad, 0xaf, 0xe0, 0x0b, 0xf8, 0x08, 0x3e, 0x46, 0xdd,
	0x09, 0xee, 0x45, 0x8b, 0x0b, 0x97, 0x42, 0x37, 0x2e, 0x25, 0x37, 0x69, 0x2d, 0x01, 0x37, 0xdd,
	0x24, 0xe7, 0x1c, 0x7e, 0xdf, 0x77, 0x0e, 0x09, 0xae, 0x82, 0xc7, 0x7c, 0x3d, 0x04, 0xbf, 0xed,
	0xf9, 0x10, 0x02, 0xf9, 0xc5, 0x5f, 0xf5, 0x5d, 0x0b, 0xc0, 0x72, 0x18, 0xd5, 0x3d, 0x9b, 0xea,
	0xae, 0x0b, 0xa1, 0x1e, 0xda, 0xe0, 0x06, 0x49, 0xa8, 0xde, 0xb5, 0xec, 0x70, 0x14, 0x19, 0xed,
	0x01, 0x8c, 0xe9, 0x18, 0x0c, 0xdb, 0x61, 0xa6, 0xc5, 0xae, 0x69, 0xfc, 0x6c, 0x0d, 0x1c, 0x88,
	0x4c, 0xca, 0x73, 0x16, 0x73, 0x57, 0x45, 0x4a, 0x0a, 0x3e, 0x0b, 0x22, 0x27, 0x4c, 0xbb, 0xd6,
	0x9a, 0xc7, 0x02, 0x0b, 0x92, 0xb4, 0x11, 0x0d, 0x79, 0xc7, 0x1b, 0x5e, 0x25, 0x71, 0xb9, 0x8b,
	0x85, 0xb3, 0xf4, 0x5a, 0x15, 0x4c, 0x46, 0x6a, 0xb8, 0xd0, 0x3f, 0x55, 0x45, 0xb4, 0x87, 0xf6,
	0x7f, 0x6b, 0x71, 0xc9, 0x27, 0xaa, 0x2a, 0xe6, 0xd3, 0x89, 0xaa, 0x2a, 0xc5, 0xf7, 0x85, 0x88,
	0x64, 0x8a, 0x2b, 0x4b, 0xf2, 0x84, 0x4d, 0x09, 0xc1, 0x45, 0x57, 0x1f, 0xb3, 0x94, 0xe4, 0xb5,
	0x22, 0xc4, 0xc1, 0xcf, 0x85, 0x88, 0xee, 0xef, 0x1a, 0x48, 0xbe, 0xc4, 0xe5, 0x25, 0x40, 0xb6,
	0x70, 0x69, 0x68, 0x33, 0xc7, 0x0c, 0x78, 0x5e, 0xd0, 0xd2, 0x8e, 0x34, 0x71, 0xe1, 0x8a, 0x4d,
	0xf9, 0xb2, 0x4a, 0x87, 0x24, 0x37, 0xb6, 0xd7, 0xd6, 0xf4, 0x8a, 0xb3, 0xe7, 0x46, 0x4e, 0x8b,
	0x43, 0x4a, 0x39, 0xb6, 0x7f, 0x2c, 0x44, 0xd4, 0x79, 0xc8, 0x7f, 0xdf, 0x72, 0xe4, 0xd9, 0xa4,
	0x8f, 0xab, 0xaa, 0xcf, 0xf4, 0x90, 0xad, 0xf6, 0xfd, 0xcd, 0xa8, 0xea, 0x7f, 0xd2, 0x81, 0xc6,
	0xbf, 0x9d, 0xbc, 0x73, 0xf3, 0xf4, 0x76, 0x9b, 0xff, 0x2f, 0xd7, 0xe8, 0x80, 0x83, 0x74, 0xf9,
	0x0b, 0x15, 0xd4, 0x8c, 0x75, 0xc7, 0xcc, 0x61, 0x1b, 0xe9, 0x4c, 0x0e, 0x66, 0x75, 0x17, 0x9e,
	0xb9, 0xd9, 0x75, 0x11, 0x07, 0x33, 0x3a, 0xe1, 0x7c, 0x04, 0x93, 0x9f, 0x65, 0xd9, 0x81, 0xbc,
	0xcd, 0x75, 0xff, 0xe4, 0x2a, 0x0d, 0x46, 0x30, 0x59, 0x97, 0x1d, 0xa0, 0x5e, 0x6d, 0xf6, 0x2a,
	0xe5, 0x66, 0x73, 0x09, 0x3d, 0xce, 0x25, 0xf4, 0x32, 0x97, 0x90, 0x51, 0xe2, 0xf8, 0xe1, 0x57,
	0x00, 0x00, 0x00, 0xff, 0xff, 0x3c, 0x88, 0x71, 0x25, 0xd7, 0x02, 0x00, 0x00,
}
