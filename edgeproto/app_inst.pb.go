// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app_inst.proto

package edgeproto

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/protocmd"
import _ "github.com/gogo/protobuf/gogoproto"

import strings "strings"
import reflect "reflect"

import context "golang.org/x/net/context"
import grpc "google.golang.org/grpc"

import binary "encoding/binary"

import "encoding/json"
import "github.com/mobiledgex/edge-cloud/objstore"
import "sync"
import "github.com/mobiledgex/edge-cloud/util"
import google_protobuf "github.com/gogo/protobuf/types"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// type of instance
type AppInst_Liveness int32

const (
	AppInst_UNKNOWN AppInst_Liveness = 0
	AppInst_STATIC  AppInst_Liveness = 1
	AppInst_DYNAMIC AppInst_Liveness = 2
)

var AppInst_Liveness_name = map[int32]string{
	0: "UNKNOWN",
	1: "STATIC",
	2: "DYNAMIC",
}
var AppInst_Liveness_value = map[string]int32{
	"UNKNOWN": 0,
	"STATIC":  1,
	"DYNAMIC": 2,
}

func (x AppInst_Liveness) String() string {
	return proto.EnumName(AppInst_Liveness_name, int32(x))
}
func (AppInst_Liveness) EnumDescriptor() ([]byte, []int) { return fileDescriptorAppInst, []int{1, 0} }

type AppInstKey struct {
	// App key
	AppKey AppKey `protobuf:"bytes,1,opt,name=app_key,json=appKey" json:"app_key"`
	// Cloudlet it's on
	CloudletKey CloudletKey `protobuf:"bytes,2,opt,name=cloudlet_key,json=cloudletKey" json:"cloudlet_key"`
	// inst id
	Id uint64 `protobuf:"fixed64,3,opt,name=id,proto3" json:"id,omitempty"`
}

func (m *AppInstKey) Reset()                    { *m = AppInstKey{} }
func (m *AppInstKey) String() string            { return proto.CompactTextString(m) }
func (*AppInstKey) ProtoMessage()               {}
func (*AppInstKey) Descriptor() ([]byte, []int) { return fileDescriptorAppInst, []int{0} }

// AppInsts are instances of an application instantiated
// on a cloudlet, like a docker or VM instance.
type AppInst struct {
	Fields []string `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	// Unique identifier key
	Key AppInstKey `protobuf:"bytes,2,opt,name=key" json:"key"`
	// Cache the location of the cloudlet
	CloudletLoc Loc `protobuf:"bytes,3,opt,name=cloudlet_loc,json=cloudletLoc" json:"cloudlet_loc"`
	// how to connect to this instance
	Ip []byte `protobuf:"bytes,4,opt,name=ip,proto3" json:"ip,omitempty"`
	// port to connect to this instance
	Port     uint32           `protobuf:"varint,5,opt,name=port,proto3" json:"port,omitempty"`
	Liveness AppInst_Liveness `protobuf:"varint,6,opt,name=liveness,proto3,enum=edgeproto.AppInst_Liveness" json:"liveness,omitempty"`
}

func (m *AppInst) Reset()                    { *m = AppInst{} }
func (m *AppInst) String() string            { return proto.CompactTextString(m) }
func (*AppInst) ProtoMessage()               {}
func (*AppInst) Descriptor() ([]byte, []int) { return fileDescriptorAppInst, []int{1} }

func init() {
	proto.RegisterType((*AppInstKey)(nil), "edgeproto.AppInstKey")
	proto.RegisterType((*AppInst)(nil), "edgeproto.AppInst")
	proto.RegisterEnum("edgeproto.AppInst_Liveness", AppInst_Liveness_name, AppInst_Liveness_value)
}
func (this *AppInstKey) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 7)
	s = append(s, "&edgeproto.AppInstKey{")
	s = append(s, "AppKey: "+strings.Replace(this.AppKey.GoString(), `&`, ``, 1)+",\n")
	s = append(s, "CloudletKey: "+strings.Replace(this.CloudletKey.GoString(), `&`, ``, 1)+",\n")
	s = append(s, "Id: "+fmt.Sprintf("%#v", this.Id)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringAppInst(v interface{}, typ string) string {
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

// Client API for AppInstApi service

type AppInstApiClient interface {
	CreateAppInst(ctx context.Context, in *AppInst, opts ...grpc.CallOption) (*Result, error)
	DeleteAppInst(ctx context.Context, in *AppInst, opts ...grpc.CallOption) (*Result, error)
	UpdateAppInst(ctx context.Context, in *AppInst, opts ...grpc.CallOption) (*Result, error)
	ShowAppInst(ctx context.Context, in *AppInst, opts ...grpc.CallOption) (AppInstApi_ShowAppInstClient, error)
}

type appInstApiClient struct {
	cc *grpc.ClientConn
}

func NewAppInstApiClient(cc *grpc.ClientConn) AppInstApiClient {
	return &appInstApiClient{cc}
}

func (c *appInstApiClient) CreateAppInst(ctx context.Context, in *AppInst, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.AppInstApi/CreateAppInst", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appInstApiClient) DeleteAppInst(ctx context.Context, in *AppInst, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.AppInstApi/DeleteAppInst", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appInstApiClient) UpdateAppInst(ctx context.Context, in *AppInst, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.AppInstApi/UpdateAppInst", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appInstApiClient) ShowAppInst(ctx context.Context, in *AppInst, opts ...grpc.CallOption) (AppInstApi_ShowAppInstClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_AppInstApi_serviceDesc.Streams[0], c.cc, "/edgeproto.AppInstApi/ShowAppInst", opts...)
	if err != nil {
		return nil, err
	}
	x := &appInstApiShowAppInstClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type AppInstApi_ShowAppInstClient interface {
	Recv() (*AppInst, error)
	grpc.ClientStream
}

type appInstApiShowAppInstClient struct {
	grpc.ClientStream
}

func (x *appInstApiShowAppInstClient) Recv() (*AppInst, error) {
	m := new(AppInst)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for AppInstApi service

type AppInstApiServer interface {
	CreateAppInst(context.Context, *AppInst) (*Result, error)
	DeleteAppInst(context.Context, *AppInst) (*Result, error)
	UpdateAppInst(context.Context, *AppInst) (*Result, error)
	ShowAppInst(*AppInst, AppInstApi_ShowAppInstServer) error
}

func RegisterAppInstApiServer(s *grpc.Server, srv AppInstApiServer) {
	s.RegisterService(&_AppInstApi_serviceDesc, srv)
}

func _AppInstApi_CreateAppInst_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppInst)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppInstApiServer).CreateAppInst(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.AppInstApi/CreateAppInst",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppInstApiServer).CreateAppInst(ctx, req.(*AppInst))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppInstApi_DeleteAppInst_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppInst)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppInstApiServer).DeleteAppInst(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.AppInstApi/DeleteAppInst",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppInstApiServer).DeleteAppInst(ctx, req.(*AppInst))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppInstApi_UpdateAppInst_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppInst)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppInstApiServer).UpdateAppInst(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.AppInstApi/UpdateAppInst",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppInstApiServer).UpdateAppInst(ctx, req.(*AppInst))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppInstApi_ShowAppInst_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(AppInst)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(AppInstApiServer).ShowAppInst(m, &appInstApiShowAppInstServer{stream})
}

type AppInstApi_ShowAppInstServer interface {
	Send(*AppInst) error
	grpc.ServerStream
}

type appInstApiShowAppInstServer struct {
	grpc.ServerStream
}

func (x *appInstApiShowAppInstServer) Send(m *AppInst) error {
	return x.ServerStream.SendMsg(m)
}

var _AppInstApi_serviceDesc = grpc.ServiceDesc{
	ServiceName: "edgeproto.AppInstApi",
	HandlerType: (*AppInstApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateAppInst",
			Handler:    _AppInstApi_CreateAppInst_Handler,
		},
		{
			MethodName: "DeleteAppInst",
			Handler:    _AppInstApi_DeleteAppInst_Handler,
		},
		{
			MethodName: "UpdateAppInst",
			Handler:    _AppInstApi_UpdateAppInst_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ShowAppInst",
			Handler:       _AppInstApi_ShowAppInst_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "app_inst.proto",
}

func (m *AppInstKey) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AppInstKey) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintAppInst(dAtA, i, uint64(m.AppKey.Size()))
	n1, err := m.AppKey.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	dAtA[i] = 0x12
	i++
	i = encodeVarintAppInst(dAtA, i, uint64(m.CloudletKey.Size()))
	n2, err := m.CloudletKey.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n2
	if m.Id != 0 {
		dAtA[i] = 0x19
		i++
		binary.LittleEndian.PutUint64(dAtA[i:], uint64(m.Id))
		i += 8
	}
	return i, nil
}

func (m *AppInst) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AppInst) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Fields) > 0 {
		for _, s := range m.Fields {
			dAtA[i] = 0xa
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
	dAtA[i] = 0x12
	i++
	i = encodeVarintAppInst(dAtA, i, uint64(m.Key.Size()))
	n3, err := m.Key.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n3
	dAtA[i] = 0x1a
	i++
	i = encodeVarintAppInst(dAtA, i, uint64(m.CloudletLoc.Size()))
	n4, err := m.CloudletLoc.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n4
	if len(m.Ip) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintAppInst(dAtA, i, uint64(len(m.Ip)))
		i += copy(dAtA[i:], m.Ip)
	}
	if m.Port != 0 {
		dAtA[i] = 0x28
		i++
		i = encodeVarintAppInst(dAtA, i, uint64(m.Port))
	}
	if m.Liveness != 0 {
		dAtA[i] = 0x30
		i++
		i = encodeVarintAppInst(dAtA, i, uint64(m.Liveness))
	}
	return i, nil
}

func encodeVarintAppInst(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *AppInstKey) Matches(filter *AppInstKey) bool {
	if filter == nil {
		return true
	}
	if !m.AppKey.Matches(&filter.AppKey) {
		return false
	}
	if !m.CloudletKey.Matches(&filter.CloudletKey) {
		return false
	}
	if filter.Id != 0 && filter.Id != m.Id {
		return false
	}
	return true
}

func (m *AppInstKey) CopyInFields(src *AppInstKey) {
	m.AppKey.DeveloperKey.Name = src.AppKey.DeveloperKey.Name
	m.AppKey.Name = src.AppKey.Name
	m.AppKey.Version = src.AppKey.Version
	m.CloudletKey.OperatorKey.Name = src.CloudletKey.OperatorKey.Name
	m.CloudletKey.Name = src.CloudletKey.Name
	m.Id = src.Id
}

func (m *AppInstKey) GetKeyString() string {
	key, err := json.Marshal(m)
	if err != nil {
		util.FatalLog("Failed to marshal AppInstKey key string", "obj", m)
	}
	return string(key)
}

func AppInstKeyStringParse(str string, key *AppInstKey) {
	err := json.Unmarshal([]byte(str), key)
	if err != nil {
		util.FatalLog("Failed to unmarshal AppInstKey key string", "str", str)
	}
}

func (m *AppInst) Matches(filter *AppInst) bool {
	if filter == nil {
		return true
	}
	if !m.Key.Matches(&filter.Key) {
		return false
	}
	if !m.CloudletLoc.Matches(&filter.CloudletLoc) {
		return false
	}
	if filter.Port != 0 && filter.Port != m.Port {
		return false
	}
	if filter.Liveness != 0 && filter.Liveness != m.Liveness {
		return false
	}
	return true
}

const AppInstFieldKeyAppKeyDeveloperKeyName = "2.1.1.2"
const AppInstFieldKeyAppKeyName = "2.1.2"
const AppInstFieldKeyAppKeyVersion = "2.1.3"
const AppInstFieldKeyCloudletKeyOperatorKeyName = "2.2.1.1"
const AppInstFieldKeyCloudletKeyName = "2.2.2"
const AppInstFieldKeyId = "2.3"
const AppInstFieldCloudletLocLat = "3.1"
const AppInstFieldCloudletLocLong = "3.2"
const AppInstFieldCloudletLocHorizontalAccuracy = "3.3"
const AppInstFieldCloudletLocVerticalAccuracy = "3.4"
const AppInstFieldCloudletLocAltitude = "3.5"
const AppInstFieldCloudletLocCourse = "3.6"
const AppInstFieldCloudletLocSpeed = "3.7"
const AppInstFieldCloudletLocTimestampSeconds = "3.8.1"
const AppInstFieldCloudletLocTimestampNanos = "3.8.2"
const AppInstFieldIp = "4"
const AppInstFieldPort = "5"
const AppInstFieldLiveness = "6"

var AppInstAllFields = []string{
	AppInstFieldKeyAppKeyDeveloperKeyName,
	AppInstFieldKeyAppKeyName,
	AppInstFieldKeyAppKeyVersion,
	AppInstFieldKeyCloudletKeyOperatorKeyName,
	AppInstFieldKeyCloudletKeyName,
	AppInstFieldKeyId,
	AppInstFieldCloudletLocLat,
	AppInstFieldCloudletLocLong,
	AppInstFieldCloudletLocHorizontalAccuracy,
	AppInstFieldCloudletLocVerticalAccuracy,
	AppInstFieldCloudletLocAltitude,
	AppInstFieldCloudletLocCourse,
	AppInstFieldCloudletLocSpeed,
	AppInstFieldCloudletLocTimestampSeconds,
	AppInstFieldCloudletLocTimestampNanos,
	AppInstFieldIp,
	AppInstFieldPort,
	AppInstFieldLiveness,
}

func (m *AppInst) CopyInFields(src *AppInst) {
	fmap := make(map[string]struct{})
	// add specified fields and parent fields
	for _, set := range src.Fields {
		for {
			fmap[set] = struct{}{}
			idx := strings.LastIndex(set, ".")
			if idx == -1 {
				break
			}
			set = set[:idx]
		}
	}
	if _, set := fmap["2"]; set {
		if _, set := fmap["2.1"]; set {
			if _, set := fmap["2.1.1"]; set {
				if _, set := fmap["2.1.1.2"]; set {
					m.Key.AppKey.DeveloperKey.Name = src.Key.AppKey.DeveloperKey.Name
				}
			}
			if _, set := fmap["2.1.2"]; set {
				m.Key.AppKey.Name = src.Key.AppKey.Name
			}
			if _, set := fmap["2.1.3"]; set {
				m.Key.AppKey.Version = src.Key.AppKey.Version
			}
		}
		if _, set := fmap["2.2"]; set {
			if _, set := fmap["2.2.1"]; set {
				if _, set := fmap["2.2.1.1"]; set {
					m.Key.CloudletKey.OperatorKey.Name = src.Key.CloudletKey.OperatorKey.Name
				}
			}
			if _, set := fmap["2.2.2"]; set {
				m.Key.CloudletKey.Name = src.Key.CloudletKey.Name
			}
		}
		if _, set := fmap["2.3"]; set {
			m.Key.Id = src.Key.Id
		}
	}
	if _, set := fmap["3"]; set {
		if _, set := fmap["3.1"]; set {
			m.CloudletLoc.Lat = src.CloudletLoc.Lat
		}
		if _, set := fmap["3.2"]; set {
			m.CloudletLoc.Long = src.CloudletLoc.Long
		}
		if _, set := fmap["3.3"]; set {
			m.CloudletLoc.HorizontalAccuracy = src.CloudletLoc.HorizontalAccuracy
		}
		if _, set := fmap["3.4"]; set {
			m.CloudletLoc.VerticalAccuracy = src.CloudletLoc.VerticalAccuracy
		}
		if _, set := fmap["3.5"]; set {
			m.CloudletLoc.Altitude = src.CloudletLoc.Altitude
		}
		if _, set := fmap["3.6"]; set {
			m.CloudletLoc.Course = src.CloudletLoc.Course
		}
		if _, set := fmap["3.7"]; set {
			m.CloudletLoc.Speed = src.CloudletLoc.Speed
		}
		if _, set := fmap["3.8"]; set && src.CloudletLoc.Timestamp != nil {
			m.CloudletLoc.Timestamp = &google_protobuf.Timestamp{}
			if _, set := fmap["3.8.1"]; set {
				m.CloudletLoc.Timestamp.Seconds = src.CloudletLoc.Timestamp.Seconds
			}
			if _, set := fmap["3.8.2"]; set {
				m.CloudletLoc.Timestamp.Nanos = src.CloudletLoc.Timestamp.Nanos
			}
		}
	}
	if _, set := fmap["4"]; set {
		if m.Ip == nil || len(m.Ip) < len(src.Ip) {
			m.Ip = make([]byte, len(src.Ip))
		}
		copy(m.Ip, src.Ip)
	}
	if _, set := fmap["5"]; set {
		m.Port = src.Port
	}
	if _, set := fmap["6"]; set {
		m.Liveness = src.Liveness
	}
}

func (s *AppInst) HasFields() bool {
	return true
}

type AppInstStore struct {
	objstore    objstore.ObjStore
	listAppInst map[AppInstKey]struct{}
}

func NewAppInstStore(objstore objstore.ObjStore) AppInstStore {
	return AppInstStore{objstore: objstore}
}

type AppInstCacher interface {
	SyncAppInstUpdate(m *AppInst, rev int64)
	SyncAppInstDelete(m *AppInst, rev int64)
	SyncAppInstPrune(current map[AppInstKey]struct{})
	SyncAppInstRevOnly(rev int64)
}

func (s *AppInstStore) Create(m *AppInst, wait func(int64)) (*Result, error) {
	err := m.Validate()
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString(m.GetKey())
	val, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	rev, err := s.objstore.Create(key, string(val))
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *AppInstStore) Update(m *AppInst, wait func(int64)) (*Result, error) {
	err := m.Validate()
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString(m.GetKey())
	var vers int64 = 0
	curBytes, vers, err := s.objstore.Get(key)
	if err != nil {
		return nil, err
	}
	var cur AppInst
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
	rev, err := s.objstore.Update(key, string(val), vers)
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *AppInstStore) Delete(m *AppInst, wait func(int64)) (*Result, error) {
	err := m.GetKey().Validate()
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString(m.GetKey())
	rev, err := s.objstore.Delete(key)
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

type AppInstCb func(m *AppInst) error

func (s *AppInstStore) LoadAll(cb AppInstCb) error {
	loadkey := objstore.DbKeyPrefixString(&AppInstKey{})
	err := s.objstore.List(loadkey, func(key, val []byte, rev int64) error {
		var obj AppInst
		err := json.Unmarshal(val, &obj)
		if err != nil {
			util.WarnLog("Failed to parse AppInst data", "val", string(val))
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

func (s *AppInstStore) LoadOne(key string) (*AppInst, int64, error) {
	val, rev, err := s.objstore.Get(key)
	if err != nil {
		return nil, 0, err
	}
	var obj AppInst
	err = json.Unmarshal(val, &obj)
	if err != nil {
		util.DebugLog(util.DebugLevelApi, "Failed to parse AppInst data", "val", string(val))
		return nil, 0, err
	}
	return &obj, rev, nil
}

// Sync will sync changes for any AppInst objects.
func (s *AppInstStore) Sync(ctx context.Context, cacher AppInstCacher) error {
	str := objstore.DbKeyPrefixString(&AppInstKey{})
	return s.objstore.Sync(ctx, str, func(in *objstore.SyncCbData) {
		obj := AppInst{}
		// Even on parse error, we should still call back to keep
		// the revision numbers in sync so no caller hangs on wait.
		action := in.Action
		if action == objstore.SyncUpdate || action == objstore.SyncList {
			err := json.Unmarshal(in.Value, &obj)
			if err != nil {
				util.WarnLog("Failed to parse AppInst data", "val", string(in.Value))
				action = objstore.SyncRevOnly
			}
		} else if action == objstore.SyncDelete {
			keystr := objstore.DbKeyPrefixRemove(string(in.Key))
			AppInstKeyStringParse(keystr, obj.GetKey())
		}
		util.DebugLog(util.DebugLevelApi, "Sync cb", "action", objstore.SyncActionStrs[in.Action], "key", string(in.Key), "value", string(in.Value), "rev", in.Rev)
		switch action {
		case objstore.SyncUpdate:
			cacher.SyncAppInstUpdate(&obj, in.Rev)
		case objstore.SyncDelete:
			cacher.SyncAppInstDelete(&obj, in.Rev)
		case objstore.SyncListStart:
			s.listAppInst = make(map[AppInstKey]struct{})
		case objstore.SyncList:
			s.listAppInst[obj.Key] = struct{}{}
			cacher.SyncAppInstUpdate(&obj, in.Rev)
		case objstore.SyncListEnd:
			cacher.SyncAppInstPrune(s.listAppInst)
			s.listAppInst = nil
		case objstore.SyncRevOnly:
			cacher.SyncAppInstRevOnly(in.Rev)
		}
	})
}

// AppInstCache caches AppInst objects in memory in a hash table
// and keeps them in sync with the database.
type AppInstCache struct {
	Store      *AppInstStore
	Objs       map[AppInstKey]*AppInst
	Rev        int64
	Mux        util.Mutex
	Cond       sync.Cond
	initWait   bool
	syncDone   bool
	syncCancel context.CancelFunc
	notifyCb   func(obj *AppInstKey)
}

func NewAppInstCache(store *AppInstStore) *AppInstCache {
	cache := AppInstCache{
		Store:    store,
		Objs:     make(map[AppInstKey]*AppInst),
		initWait: true,
	}
	cache.Mux.InitCond(&cache.Cond)

	ctx, cancel := context.WithCancel(context.Background())
	cache.syncCancel = cancel
	go func() {
		err := cache.Store.Sync(ctx, &cache)
		if err != nil {
			util.WarnLog("AppInst Sync failed", "err", err)
		}
		cache.syncDone = true
		cache.Cond.Broadcast()
	}()
	return &cache
}

func (c *AppInstCache) WaitInitSyncDone() {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for c.initWait {
		c.Cond.Wait()
	}
}

func (c *AppInstCache) Done() {
	c.syncCancel()
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for !c.syncDone {
		c.Cond.Wait()
	}
}

func (c *AppInstCache) Get(key *AppInstKey, valbuf *AppInst) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	inst, found := c.Objs[*key]
	if found {
		*valbuf = *inst
	}
	return found
}

func (c *AppInstCache) HasKey(key *AppInstKey) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	_, found := c.Objs[*key]
	return found
}

func (c *AppInstCache) GetAllKeys(keys map[AppInstKey]struct{}) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for key, _ := range c.Objs {
		keys[key] = struct{}{}
	}
}

func (c *AppInstCache) SyncAppInstUpdate(in *AppInst, rev int64) {
	c.Mux.Lock()
	c.Objs[*in.GetKey()] = in
	c.Rev = rev
	util.DebugLog(util.DebugLevelApi, "SyncUpdate", "obj", in, "rev", rev)
	c.Cond.Broadcast()
	c.Mux.Unlock()
	if c.notifyCb != nil {
		c.notifyCb(in.GetKey())
	}
}

func (c *AppInstCache) SyncAppInstDelete(in *AppInst, rev int64) {
	c.Mux.Lock()
	delete(c.Objs, *in.GetKey())
	c.Rev = rev
	util.DebugLog(util.DebugLevelApi, "SyncUpdate", "key", in.GetKey(), "rev", rev)
	c.Cond.Broadcast()
	c.Mux.Unlock()
	if c.notifyCb != nil {
		c.notifyCb(in.GetKey())
	}
}

func (c *AppInstCache) SyncAppInstPrune(current map[AppInstKey]struct{}) {
	deleted := make(map[AppInstKey]struct{})
	c.Mux.Lock()
	for key, _ := range c.Objs {
		if _, found := current[key]; !found {
			delete(c.Objs, key)
			deleted[key] = struct{}{}
		}
	}
	if c.initWait {
		c.initWait = false
		c.Cond.Broadcast()
	}
	c.Mux.Unlock()
	if c.notifyCb != nil {
		for key, _ := range deleted {
			c.notifyCb(&key)
		}
	}
}

func (c *AppInstCache) SyncAppInstRevOnly(rev int64) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	c.Rev = rev
	util.DebugLog(util.DebugLevelApi, "SyncRevOnly", "rev", rev)
	c.Cond.Broadcast()
}

func (c *AppInstCache) SyncWait(rev int64) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	util.DebugLog(util.DebugLevelApi, "SyncWait", "cache-rev", c.Rev, "wait-rev", rev)
	for c.Rev < rev {
		c.Cond.Wait()
	}
}

func (c *AppInstCache) Show(filter *AppInst, cb func(ret *AppInst) error) error {
	util.DebugLog(util.DebugLevelApi, "Show AppInst", "count", len(c.Objs))
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for _, obj := range c.Objs {
		if !obj.Matches(filter) {
			continue
		}
		util.DebugLog(util.DebugLevelApi, "Show AppInst", "obj", obj)
		err := cb(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *AppInstCache) SetNotifyCb(fn func(obj *AppInstKey)) {
	c.notifyCb = fn
}

func (m *AppInst) GetKey() *AppInstKey {
	return &m.Key
}

func (m *AppInstKey) Size() (n int) {
	var l int
	_ = l
	l = m.AppKey.Size()
	n += 1 + l + sovAppInst(uint64(l))
	l = m.CloudletKey.Size()
	n += 1 + l + sovAppInst(uint64(l))
	if m.Id != 0 {
		n += 9
	}
	return n
}

func (m *AppInst) Size() (n int) {
	var l int
	_ = l
	if len(m.Fields) > 0 {
		for _, s := range m.Fields {
			l = len(s)
			n += 1 + l + sovAppInst(uint64(l))
		}
	}
	l = m.Key.Size()
	n += 1 + l + sovAppInst(uint64(l))
	l = m.CloudletLoc.Size()
	n += 1 + l + sovAppInst(uint64(l))
	l = len(m.Ip)
	if l > 0 {
		n += 1 + l + sovAppInst(uint64(l))
	}
	if m.Port != 0 {
		n += 1 + sovAppInst(uint64(m.Port))
	}
	if m.Liveness != 0 {
		n += 1 + sovAppInst(uint64(m.Liveness))
	}
	return n
}

func sovAppInst(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozAppInst(x uint64) (n int) {
	return sovAppInst(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *AppInstKey) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAppInst
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
			return fmt.Errorf("proto: AppInstKey: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AppInstKey: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AppKey", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppInst
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
				return ErrInvalidLengthAppInst
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.AppKey.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CloudletKey", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppInst
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
				return ErrInvalidLengthAppInst
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.CloudletKey.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			m.Id = 0
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = uint64(binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
		default:
			iNdEx = preIndex
			skippy, err := skipAppInst(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthAppInst
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
func (m *AppInst) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAppInst
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
			return fmt.Errorf("proto: AppInst: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AppInst: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Fields", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppInst
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
				return ErrInvalidLengthAppInst
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Fields = append(m.Fields, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Key", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppInst
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
				return ErrInvalidLengthAppInst
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Key.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CloudletLoc", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppInst
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
				return ErrInvalidLengthAppInst
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.CloudletLoc.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ip", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppInst
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
				return ErrInvalidLengthAppInst
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Ip = append(m.Ip[:0], dAtA[iNdEx:postIndex]...)
			if m.Ip == nil {
				m.Ip = []byte{}
			}
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Port", wireType)
			}
			m.Port = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppInst
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Port |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Liveness", wireType)
			}
			m.Liveness = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppInst
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Liveness |= (AppInst_Liveness(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipAppInst(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthAppInst
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
func skipAppInst(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowAppInst
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
					return 0, ErrIntOverflowAppInst
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
					return 0, ErrIntOverflowAppInst
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
				return 0, ErrInvalidLengthAppInst
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowAppInst
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
				next, err := skipAppInst(dAtA[start:])
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
	ErrInvalidLengthAppInst = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowAppInst   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("app_inst.proto", fileDescriptorAppInst) }

var fileDescriptorAppInst = []byte{
	// 581 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x93, 0xc1, 0x8b, 0xd3, 0x40,
	0x14, 0xc6, 0x3b, 0x69, 0x4d, 0xb7, 0x93, 0xb6, 0xd6, 0x41, 0x97, 0xd8, 0x95, 0x6e, 0xc8, 0x29,
	0x08, 0x6d, 0x4a, 0x3d, 0xac, 0xf4, 0x22, 0xdd, 0x2e, 0x48, 0x69, 0xad, 0x98, 0xdd, 0x45, 0x3c,
	0x49, 0x9a, 0xcc, 0xa6, 0x83, 0x69, 0x66, 0x68, 0x52, 0xd7, 0xde, 0xc4, 0xa3, 0x57, 0xcf, 0x82,
	0x7f, 0x82, 0x08, 0xfe, 0x0f, 0x3d, 0x0a, 0x1e, 0x05, 0xd1, 0xe2, 0xc1, 0xa3, 0x50, 0x0f, 0x1e,
	0x25, 0xd3, 0x69, 0x1b, 0x77, 0x45, 0x84, 0xbd, 0x94, 0xef, 0xcd, 0xbc, 0xdf, 0xf7, 0xbe, 0xbe,
	0x69, 0x61, 0xd1, 0x66, 0xec, 0x31, 0x09, 0xc2, 0xa8, 0xc6, 0xc6, 0x34, 0xa2, 0x28, 0x87, 0x5d,
	0x0f, 0x73, 0x59, 0xbe, 0xe1, 0x51, 0xea, 0xf9, 0xd8, 0xb4, 0x19, 0x31, 0xed, 0x20, 0xa0, 0x91,
	0x1d, 0x11, 0x1a, 0x84, 0xcb, 0xc6, 0x72, 0x7e, 0x8c, 0xc3, 0x89, 0x2f, 0xb0, 0xf2, 0x6d, 0x8f,
	0x44, 0xc3, 0xc9, 0xa0, 0xe6, 0xd0, 0x91, 0x39, 0xa2, 0x03, 0xe2, 0xc7, 0x36, 0xcf, 0xcc, 0xf8,
	0xb3, 0xea, 0xf8, 0x74, 0xe2, 0x9a, 0xbc, 0xcf, 0xc3, 0xc1, 0x5a, 0x08, 0xf2, 0xee, 0xff, 0x91,
	0x4e, 0xd5, 0xc3, 0x41, 0xd5, 0x19, 0xad, 0xca, 0x84, 0x10, 0x46, 0x39, 0x9b, 0x31, 0x21, 0x8b,
	0x1c, 0xf4, 0xf1, 0x2a, 0x5d, 0xce, 0xa7, 0x8e, 0x90, 0xd5, 0xc4, 0x38, 0x8f, 0x7a, 0x74, 0xe9,
	0x32, 0x98, 0x9c, 0xf0, 0x8a, 0x17, 0x5c, 0x2d, 0xdb, 0xf5, 0xd7, 0x00, 0xc2, 0x16, 0x63, 0x9d,
	0x20, 0x8c, 0xba, 0x78, 0x8a, 0xea, 0x30, 0x1b, 0xef, 0xeb, 0x09, 0x9e, 0xaa, 0x40, 0x03, 0x86,
	0xd2, 0xb8, 0x52, 0x5b, 0xef, 0xab, 0xd6, 0x62, 0xac, 0x8b, 0xa7, 0xfb, 0x99, 0xd9, 0xe7, 0xdd,
	0x94, 0x25, 0xdb, 0xbc, 0x42, 0x77, 0x60, 0x7e, 0x15, 0x86, 0x63, 0x12, 0xc7, 0xb6, 0x13, 0x58,
	0x5b, 0x5c, 0x6f, 0x58, 0xc5, 0xd9, 0x1c, 0xa1, 0x22, 0x94, 0x88, 0xab, 0xa6, 0x35, 0x60, 0xc8,
	0x96, 0x44, 0xdc, 0x66, 0xfe, 0xfb, 0x42, 0x05, 0xbf, 0x16, 0x2a, 0x78, 0xfb, 0x66, 0x17, 0xe8,
	0xef, 0x25, 0x98, 0x15, 0xf9, 0xd0, 0x36, 0x94, 0x4f, 0x08, 0xf6, 0xdd, 0x50, 0x05, 0x5a, 0xda,
	0xc8, 0x59, 0xa2, 0x42, 0x55, 0x98, 0xde, 0x4c, 0xbe, 0xf6, 0x67, 0x60, 0xf1, 0xc5, 0xc4, 0xe0,
	0xb8, 0x0f, 0xed, 0x25, 0x12, 0xfb, 0xd4, 0xe1, 0xa3, 0x95, 0x46, 0x31, 0xc1, 0xf5, 0xa8, 0x73,
	0x36, 0x69, 0x8f, 0x3a, 0x3c, 0x29, 0x53, 0x33, 0x1a, 0x30, 0xf2, 0x96, 0x44, 0x18, 0x42, 0x30,
	0xc3, 0xe8, 0x38, 0x52, 0x2f, 0x69, 0xc0, 0x28, 0x58, 0x5c, 0xa3, 0x3d, 0xb8, 0xe5, 0x93, 0xa7,
	0x38, 0xc0, 0x61, 0xa8, 0xca, 0x1a, 0x30, 0x8a, 0x8d, 0x9d, 0xf3, 0x81, 0x6a, 0x3d, 0xd1, 0x62,
	0xad, 0x9b, 0xf5, 0x3a, 0xdc, 0x5a, 0x9d, 0x22, 0x05, 0x66, 0x8f, 0xfb, 0xdd, 0xfe, 0xfd, 0x87,
	0xfd, 0x52, 0x0a, 0x41, 0x28, 0x1f, 0x1e, 0xb5, 0x8e, 0x3a, 0xed, 0x12, 0x88, 0x2f, 0x0e, 0x1e,
	0xf5, 0x5b, 0xf7, 0x3a, 0xed, 0x92, 0xd4, 0xdc, 0x89, 0x17, 0xf5, 0x63, 0xa1, 0x82, 0xe7, 0x3f,
	0x55, 0xf0, 0xf2, 0xdd, 0x75, 0xa5, 0xbd, 0xc9, 0xda, 0xf8, 0x24, 0xad, 0xdf, 0xb5, 0xc5, 0x08,
	0xb2, 0x60, 0xa1, 0x3d, 0xc6, 0x76, 0x84, 0x57, 0xbb, 0x44, 0xe7, 0x53, 0x95, 0x93, 0x6f, 0x6d,
	0xf1, 0x1f, 0xbf, 0x5e, 0x7e, 0xf1, 0xf1, 0xdb, 0x2b, 0xe9, 0xaa, 0x7e, 0xd9, 0x74, 0x38, 0x6e,
	0xda, 0x8c, 0xc5, 0x7f, 0xa6, 0x26, 0xb8, 0x19, 0x7b, 0x1e, 0x60, 0x1f, 0x5f, 0xc0, 0xd3, 0xe5,
	0xf8, 0x19, 0xcf, 0x63, 0xe6, 0x5e, 0x24, 0xe7, 0x84, 0xe3, 0x49, 0xcf, 0x07, 0x50, 0x39, 0x1c,
	0xd2, 0xd3, 0x7f, 0x39, 0xfe, 0xe5, 0x4c, 0x57, 0xb9, 0x25, 0xd2, 0x0b, 0x66, 0x38, 0xa4, 0xa7,
	0x09, 0xc3, 0x3a, 0xd8, 0x2f, 0xcd, 0xbe, 0x56, 0x52, 0xb3, 0x79, 0x05, 0x7c, 0x98, 0x57, 0xc0,
	0x97, 0x79, 0x05, 0x0c, 0x64, 0x0e, 0xdf, 0xfa, 0x1d, 0x00, 0x00, 0xff, 0xff, 0xc8, 0xaa, 0x4b,
	0x2f, 0x6f, 0x04, 0x00, 0x00,
}
