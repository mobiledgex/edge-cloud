// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

package edgeproto

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/protocmd"
import distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

import strings "strings"
import reflect "reflect"

import context "golang.org/x/net/context"
import grpc "google.golang.org/grpc"

import "encoding/json"
import "github.com/mobiledgex/edge-cloud/objstore"
import "github.com/mobiledgex/edge-cloud/util"
import google_protobuf "github.com/gogo/protobuf/types"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type CloudletState int32

const (
	CloudletState_Unknown               CloudletState = 0
	CloudletState_ConfiguringOpenstack  CloudletState = 1
	CloudletState_ConfiguringKubernetes CloudletState = 2
	CloudletState_Ready                 CloudletState = 3
)

var CloudletState_name = map[int32]string{
	0: "Unknown",
	1: "ConfiguringOpenstack",
	2: "ConfiguringKubernetes",
	3: "Ready",
}
var CloudletState_value = map[string]int32{
	"Unknown":               0,
	"ConfiguringOpenstack":  1,
	"ConfiguringKubernetes": 2,
	"Ready":                 3,
}

func (x CloudletState) String() string {
	return proto.EnumName(CloudletState_name, int32(x))
}
func (CloudletState) EnumDescriptor() ([]byte, []int) { return fileDescriptorCloudlet, []int{0} }

type CloudletKey struct {
	// Operator of the cloudlet site
	OperatorKey OperatorKey `protobuf:"bytes,1,opt,name=operator_key,json=operatorKey" json:"operator_key"`
	// Name of the cloudlet
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (m *CloudletKey) Reset()                    { *m = CloudletKey{} }
func (m *CloudletKey) String() string            { return proto.CompactTextString(m) }
func (*CloudletKey) ProtoMessage()               {}
func (*CloudletKey) Descriptor() ([]byte, []int) { return fileDescriptorCloudlet, []int{0} }

// Cloudlet Sites are created and uploaded by Operators
// This information is used to connect to and manage Cloudlets
type Cloudlet struct {
	Fields []string `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	// Unique identifier key
	Key CloudletKey `protobuf:"bytes,2,opt,name=key" json:"key"`
	// URI to use to connect to and control cloudlet site
	AccessUri string `protobuf:"bytes,4,opt,name=access_uri,json=accessUri,proto3" json:"access_uri,omitempty"`
	// Location of the cloudlet site (lat, long?)
	Location distributed_match_engine.Loc `protobuf:"bytes,5,opt,name=location" json:"location"`
}

func (m *Cloudlet) Reset()                    { *m = Cloudlet{} }
func (m *Cloudlet) String() string            { return proto.CompactTextString(m) }
func (*Cloudlet) ProtoMessage()               {}
func (*Cloudlet) Descriptor() ([]byte, []int) { return fileDescriptorCloudlet, []int{1} }

// CloudletInfo is the information CRM passes up to the controller
// about the Cloudlets it is managing.
type CloudletInfo struct {
	// Unique identifier key
	Key CloudletKey `protobuf:"bytes,1,opt,name=key" json:"key"`
	// State of cloudlet
	State CloudletState `protobuf:"varint,2,opt,name=state,proto3,enum=edgeproto.CloudletState" json:"state,omitempty"`
	// TODO: Not entirely sure how resources will be specified.
	// This is a placeholder for now.
	Resources uint64 `protobuf:"varint,3,opt,name=resources,proto3" json:"resources,omitempty"`
}

func (m *CloudletInfo) Reset()                    { *m = CloudletInfo{} }
func (m *CloudletInfo) String() string            { return proto.CompactTextString(m) }
func (*CloudletInfo) ProtoMessage()               {}
func (*CloudletInfo) Descriptor() ([]byte, []int) { return fileDescriptorCloudlet, []int{2} }

func init() {
	proto.RegisterType((*CloudletKey)(nil), "edgeproto.CloudletKey")
	proto.RegisterType((*Cloudlet)(nil), "edgeproto.Cloudlet")
	proto.RegisterType((*CloudletInfo)(nil), "edgeproto.CloudletInfo")
	proto.RegisterEnum("edgeproto.CloudletState", CloudletState_name, CloudletState_value)
}
func (this *CloudletKey) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 6)
	s = append(s, "&edgeproto.CloudletKey{")
	s = append(s, "OperatorKey: "+strings.Replace(this.OperatorKey.GoString(), `&`, ``, 1)+",\n")
	s = append(s, "Name: "+fmt.Sprintf("%#v", this.Name)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringCloudlet(v interface{}, typ string) string {
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

// Client API for CloudletApi service

type CloudletApiClient interface {
	CreateCloudlet(ctx context.Context, in *Cloudlet, opts ...grpc.CallOption) (*Result, error)
	DeleteCloudlet(ctx context.Context, in *Cloudlet, opts ...grpc.CallOption) (*Result, error)
	UpdateCloudlet(ctx context.Context, in *Cloudlet, opts ...grpc.CallOption) (*Result, error)
	ShowCloudlet(ctx context.Context, in *Cloudlet, opts ...grpc.CallOption) (CloudletApi_ShowCloudletClient, error)
}

type cloudletApiClient struct {
	cc *grpc.ClientConn
}

func NewCloudletApiClient(cc *grpc.ClientConn) CloudletApiClient {
	return &cloudletApiClient{cc}
}

func (c *cloudletApiClient) CreateCloudlet(ctx context.Context, in *Cloudlet, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.CloudletApi/CreateCloudlet", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudletApiClient) DeleteCloudlet(ctx context.Context, in *Cloudlet, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.CloudletApi/DeleteCloudlet", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudletApiClient) UpdateCloudlet(ctx context.Context, in *Cloudlet, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.CloudletApi/UpdateCloudlet", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudletApiClient) ShowCloudlet(ctx context.Context, in *Cloudlet, opts ...grpc.CallOption) (CloudletApi_ShowCloudletClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_CloudletApi_serviceDesc.Streams[0], c.cc, "/edgeproto.CloudletApi/ShowCloudlet", opts...)
	if err != nil {
		return nil, err
	}
	x := &cloudletApiShowCloudletClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type CloudletApi_ShowCloudletClient interface {
	Recv() (*Cloudlet, error)
	grpc.ClientStream
}

type cloudletApiShowCloudletClient struct {
	grpc.ClientStream
}

func (x *cloudletApiShowCloudletClient) Recv() (*Cloudlet, error) {
	m := new(Cloudlet)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for CloudletApi service

type CloudletApiServer interface {
	CreateCloudlet(context.Context, *Cloudlet) (*Result, error)
	DeleteCloudlet(context.Context, *Cloudlet) (*Result, error)
	UpdateCloudlet(context.Context, *Cloudlet) (*Result, error)
	ShowCloudlet(*Cloudlet, CloudletApi_ShowCloudletServer) error
}

func RegisterCloudletApiServer(s *grpc.Server, srv CloudletApiServer) {
	s.RegisterService(&_CloudletApi_serviceDesc, srv)
}

func _CloudletApi_CreateCloudlet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Cloudlet)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudletApiServer).CreateCloudlet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.CloudletApi/CreateCloudlet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudletApiServer).CreateCloudlet(ctx, req.(*Cloudlet))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudletApi_DeleteCloudlet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Cloudlet)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudletApiServer).DeleteCloudlet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.CloudletApi/DeleteCloudlet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudletApiServer).DeleteCloudlet(ctx, req.(*Cloudlet))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudletApi_UpdateCloudlet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Cloudlet)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudletApiServer).UpdateCloudlet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.CloudletApi/UpdateCloudlet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudletApiServer).UpdateCloudlet(ctx, req.(*Cloudlet))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudletApi_ShowCloudlet_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Cloudlet)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CloudletApiServer).ShowCloudlet(m, &cloudletApiShowCloudletServer{stream})
}

type CloudletApi_ShowCloudletServer interface {
	Send(*Cloudlet) error
	grpc.ServerStream
}

type cloudletApiShowCloudletServer struct {
	grpc.ServerStream
}

func (x *cloudletApiShowCloudletServer) Send(m *Cloudlet) error {
	return x.ServerStream.SendMsg(m)
}

var _CloudletApi_serviceDesc = grpc.ServiceDesc{
	ServiceName: "edgeproto.CloudletApi",
	HandlerType: (*CloudletApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateCloudlet",
			Handler:    _CloudletApi_CreateCloudlet_Handler,
		},
		{
			MethodName: "DeleteCloudlet",
			Handler:    _CloudletApi_DeleteCloudlet_Handler,
		},
		{
			MethodName: "UpdateCloudlet",
			Handler:    _CloudletApi_UpdateCloudlet_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ShowCloudlet",
			Handler:       _CloudletApi_ShowCloudlet_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "cloudlet.proto",
}

func (m *CloudletKey) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CloudletKey) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintCloudlet(dAtA, i, uint64(m.OperatorKey.Size()))
	n1, err := m.OperatorKey.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	if len(m.Name) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintCloudlet(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	return i, nil
}

func (m *Cloudlet) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Cloudlet) MarshalTo(dAtA []byte) (int, error) {
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
	i = encodeVarintCloudlet(dAtA, i, uint64(m.Key.Size()))
	n2, err := m.Key.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n2
	if len(m.AccessUri) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintCloudlet(dAtA, i, uint64(len(m.AccessUri)))
		i += copy(dAtA[i:], m.AccessUri)
	}
	dAtA[i] = 0x2a
	i++
	i = encodeVarintCloudlet(dAtA, i, uint64(m.Location.Size()))
	n3, err := m.Location.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n3
	return i, nil
}

func (m *CloudletInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CloudletInfo) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintCloudlet(dAtA, i, uint64(m.Key.Size()))
	n4, err := m.Key.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n4
	if m.State != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintCloudlet(dAtA, i, uint64(m.State))
	}
	if m.Resources != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintCloudlet(dAtA, i, uint64(m.Resources))
	}
	return i, nil
}

func encodeVarintCloudlet(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *CloudletKey) Matches(filter *CloudletKey) bool {
	if filter == nil {
		return true
	}
	if !m.OperatorKey.Matches(&filter.OperatorKey) {
		return false
	}
	if filter.Name != "" && filter.Name != m.Name {
		return false
	}
	return true
}

func (m *CloudletKey) CopyInFields(src *CloudletKey) {
	m.OperatorKey.Name = src.OperatorKey.Name
	m.Name = src.Name
}

func (m *CloudletKey) GetKeyString() string {
	key, err := json.Marshal(m)
	if err != nil {
		util.FatalLog("Failed to marshal CloudletKey key string", "obj", m)
	}
	return string(key)
}

func CloudletKeyStringParse(str string, key *CloudletKey) {
	err := json.Unmarshal([]byte(str), key)
	if err != nil {
		util.FatalLog("Failed to unmarshal CloudletKey key string", "str", str)
	}
}

func (m *Cloudlet) Matches(filter *Cloudlet) bool {
	if filter == nil {
		return true
	}
	if !m.Key.Matches(&filter.Key) {
		return false
	}
	if filter.AccessUri != "" && filter.AccessUri != m.AccessUri {
		return false
	}
	return true
}

const CloudletFieldKey = "2"
const CloudletFieldKeyOperatorKey = "2.1"
const CloudletFieldKeyOperatorKeyName = "2.1.1"
const CloudletFieldKeyName = "2.2"
const CloudletFieldAccessUri = "4"
const CloudletFieldLocation = "5"
const CloudletFieldLocationLat = "5.1"
const CloudletFieldLocationLong = "5.2"
const CloudletFieldLocationHorizontalAccuracy = "5.3"
const CloudletFieldLocationVerticalAccuracy = "5.4"
const CloudletFieldLocationAltitude = "5.5"
const CloudletFieldLocationCourse = "5.6"
const CloudletFieldLocationSpeed = "5.7"
const CloudletFieldLocationTimestamp = "5.8"
const CloudletFieldLocationTimestampSeconds = "5.8.1"
const CloudletFieldLocationTimestampNanos = "5.8.2"

var CloudletAllFields = []string{
	CloudletFieldKeyOperatorKeyName,
	CloudletFieldKeyName,
	CloudletFieldAccessUri,
	CloudletFieldLocationLat,
	CloudletFieldLocationLong,
	CloudletFieldLocationHorizontalAccuracy,
	CloudletFieldLocationVerticalAccuracy,
	CloudletFieldLocationAltitude,
	CloudletFieldLocationCourse,
	CloudletFieldLocationSpeed,
	CloudletFieldLocationTimestampSeconds,
	CloudletFieldLocationTimestampNanos,
}

var CloudletAllFieldsMap = map[string]struct{}{
	CloudletFieldKeyOperatorKeyName:         struct{}{},
	CloudletFieldKeyName:                    struct{}{},
	CloudletFieldAccessUri:                  struct{}{},
	CloudletFieldLocationLat:                struct{}{},
	CloudletFieldLocationLong:               struct{}{},
	CloudletFieldLocationHorizontalAccuracy: struct{}{},
	CloudletFieldLocationVerticalAccuracy:   struct{}{},
	CloudletFieldLocationAltitude:           struct{}{},
	CloudletFieldLocationCourse:             struct{}{},
	CloudletFieldLocationSpeed:              struct{}{},
	CloudletFieldLocationTimestampSeconds:   struct{}{},
	CloudletFieldLocationTimestampNanos:     struct{}{},
}

func (m *Cloudlet) CopyInFields(src *Cloudlet) {
	fmap := MakeFieldMap(src.Fields)
	if _, set := fmap["2"]; set {
		if _, set := fmap["2.1"]; set {
			if _, set := fmap["2.1.1"]; set {
				m.Key.OperatorKey.Name = src.Key.OperatorKey.Name
			}
		}
		if _, set := fmap["2.2"]; set {
			m.Key.Name = src.Key.Name
		}
	}
	if _, set := fmap["4"]; set {
		m.AccessUri = src.AccessUri
	}
	if _, set := fmap["5"]; set {
		if _, set := fmap["5.1"]; set {
			m.Location.Lat = src.Location.Lat
		}
		if _, set := fmap["5.2"]; set {
			m.Location.Long = src.Location.Long
		}
		if _, set := fmap["5.3"]; set {
			m.Location.HorizontalAccuracy = src.Location.HorizontalAccuracy
		}
		if _, set := fmap["5.4"]; set {
			m.Location.VerticalAccuracy = src.Location.VerticalAccuracy
		}
		if _, set := fmap["5.5"]; set {
			m.Location.Altitude = src.Location.Altitude
		}
		if _, set := fmap["5.6"]; set {
			m.Location.Course = src.Location.Course
		}
		if _, set := fmap["5.7"]; set {
			m.Location.Speed = src.Location.Speed
		}
		if _, set := fmap["5.8"]; set && src.Location.Timestamp != nil {
			m.Location.Timestamp = &google_protobuf.Timestamp{}
			if _, set := fmap["5.8.1"]; set {
				m.Location.Timestamp.Seconds = src.Location.Timestamp.Seconds
			}
			if _, set := fmap["5.8.2"]; set {
				m.Location.Timestamp.Nanos = src.Location.Timestamp.Nanos
			}
		}
	}
}

func (s *Cloudlet) HasFields() bool {
	return true
}

type CloudletStore struct {
	objstore objstore.ObjStore
}

func NewCloudletStore(objstore objstore.ObjStore) CloudletStore {
	return CloudletStore{objstore: objstore}
}

func (s *CloudletStore) Create(m *Cloudlet, wait func(int64)) (*Result, error) {
	err := m.Validate(CloudletAllFieldsMap)
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

func (s *CloudletStore) Update(m *Cloudlet, wait func(int64)) (*Result, error) {
	fmap := MakeFieldMap(m.Fields)
	err := m.Validate(fmap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString(m.GetKey())
	var vers int64 = 0
	curBytes, vers, err := s.objstore.Get(key)
	if err != nil {
		return nil, err
	}
	var cur Cloudlet
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

func (s *CloudletStore) Delete(m *Cloudlet, wait func(int64)) (*Result, error) {
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

type CloudletCb func(m *Cloudlet) error

func (s *CloudletStore) LoadAll(cb CloudletCb) error {
	loadkey := objstore.DbKeyPrefixString(&CloudletKey{})
	err := s.objstore.List(loadkey, func(key, val []byte, rev int64) error {
		var obj Cloudlet
		err := json.Unmarshal(val, &obj)
		if err != nil {
			util.WarnLog("Failed to parse Cloudlet data", "val", string(val))
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

func (s *CloudletStore) LoadOne(key string) (*Cloudlet, int64, error) {
	val, rev, err := s.objstore.Get(key)
	if err != nil {
		return nil, 0, err
	}
	var obj Cloudlet
	err = json.Unmarshal(val, &obj)
	if err != nil {
		util.DebugLog(util.DebugLevelApi, "Failed to parse Cloudlet data", "val", string(val))
		return nil, 0, err
	}
	return &obj, rev, nil
}

// CloudletCache caches Cloudlet objects in memory in a hash table
// and keeps them in sync with the database.
type CloudletCache struct {
	Objs      map[CloudletKey]*Cloudlet
	Mux       util.Mutex
	List      map[CloudletKey]struct{}
	NotifyCb  func(obj *CloudletKey)
	UpdatedCb func(old *Cloudlet, new *Cloudlet)
}

func InitCloudletCache(cache *CloudletCache) {
	cache.Objs = make(map[CloudletKey]*Cloudlet)
}

func (c *CloudletCache) Get(key *CloudletKey, valbuf *Cloudlet) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	inst, found := c.Objs[*key]
	if found {
		*valbuf = *inst
	}
	return found
}

func (c *CloudletCache) HasKey(key *CloudletKey) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	_, found := c.Objs[*key]
	return found
}

func (c *CloudletCache) GetAllKeys(keys map[CloudletKey]struct{}) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for key, _ := range c.Objs {
		keys[key] = struct{}{}
	}
}

func (c *CloudletCache) Update(in *Cloudlet, rev int64) {
	c.Mux.Lock()
	if c.UpdatedCb != nil {
		old := c.Objs[*in.GetKey()]
		new := &Cloudlet{}
		*new = *in
		defer c.UpdatedCb(old, new)
	}
	c.Objs[*in.GetKey()] = in
	util.DebugLog(util.DebugLevelApi, "SyncUpdate", "obj", in, "rev", rev)
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		c.NotifyCb(in.GetKey())
	}
}

func (c *CloudletCache) Delete(in *Cloudlet, rev int64) {
	c.Mux.Lock()
	delete(c.Objs, *in.GetKey())
	util.DebugLog(util.DebugLevelApi, "SyncUpdate", "key", in.GetKey(), "rev", rev)
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		c.NotifyCb(in.GetKey())
	}
}

func (c *CloudletCache) Show(filter *Cloudlet, cb func(ret *Cloudlet) error) error {
	util.DebugLog(util.DebugLevelApi, "Show Cloudlet", "count", len(c.Objs))
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for _, obj := range c.Objs {
		if !obj.Matches(filter) {
			continue
		}
		util.DebugLog(util.DebugLevelApi, "Show Cloudlet", "obj", obj)
		err := cb(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CloudletCache) SetNotifyCb(fn func(obj *CloudletKey)) {
	c.NotifyCb = fn
}

func (c *CloudletCache) SetUpdatedCb(fn func(old *Cloudlet, new *Cloudlet)) {
	c.UpdatedCb = fn
}

func (c *CloudletCache) SyncUpdate(key, val []byte, rev int64) {
	obj := Cloudlet{}
	err := json.Unmarshal(val, &obj)
	if err != nil {
		util.WarnLog("Failed to parse Cloudlet data", "val", string(val))
		return
	}
	c.Update(&obj, rev)
	c.Mux.Lock()
	if c.List != nil {
		c.List[obj.Key] = struct{}{}
	}
	c.Mux.Unlock()
}

func (c *CloudletCache) SyncDelete(key []byte, rev int64) {
	obj := Cloudlet{}
	keystr := objstore.DbKeyPrefixRemove(string(key))
	CloudletKeyStringParse(keystr, obj.GetKey())
	c.Delete(&obj, rev)
}

func (c *CloudletCache) SyncListStart() {
	c.List = make(map[CloudletKey]struct{})
}

func (c *CloudletCache) SyncListEnd() {
	deleted := make(map[CloudletKey]struct{})
	c.Mux.Lock()
	for key, _ := range c.Objs {
		if _, found := c.List[key]; !found {
			delete(c.Objs, key)
			deleted[key] = struct{}{}
		}
	}
	c.List = nil
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		for key, _ := range deleted {
			c.NotifyCb(&key)
		}
	}
}

func (m *Cloudlet) GetKey() *CloudletKey {
	return &m.Key
}

func (m *CloudletInfo) CopyInFields(src *CloudletInfo) {
	m.Key.OperatorKey.Name = src.Key.OperatorKey.Name
	m.Key.Name = src.Key.Name
	m.State = src.State
	m.Resources = src.Resources
}

func (m *CloudletInfo) GetKey() *CloudletKey {
	return &m.Key
}

func (m *CloudletKey) Size() (n int) {
	var l int
	_ = l
	l = m.OperatorKey.Size()
	n += 1 + l + sovCloudlet(uint64(l))
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovCloudlet(uint64(l))
	}
	return n
}

func (m *Cloudlet) Size() (n int) {
	var l int
	_ = l
	if len(m.Fields) > 0 {
		for _, s := range m.Fields {
			l = len(s)
			n += 1 + l + sovCloudlet(uint64(l))
		}
	}
	l = m.Key.Size()
	n += 1 + l + sovCloudlet(uint64(l))
	l = len(m.AccessUri)
	if l > 0 {
		n += 1 + l + sovCloudlet(uint64(l))
	}
	l = m.Location.Size()
	n += 1 + l + sovCloudlet(uint64(l))
	return n
}

func (m *CloudletInfo) Size() (n int) {
	var l int
	_ = l
	l = m.Key.Size()
	n += 1 + l + sovCloudlet(uint64(l))
	if m.State != 0 {
		n += 1 + sovCloudlet(uint64(m.State))
	}
	if m.Resources != 0 {
		n += 1 + sovCloudlet(uint64(m.Resources))
	}
	return n
}

func sovCloudlet(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozCloudlet(x uint64) (n int) {
	return sovCloudlet(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *CloudletKey) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCloudlet
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
			return fmt.Errorf("proto: CloudletKey: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CloudletKey: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OperatorKey", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCloudlet
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
				return ErrInvalidLengthCloudlet
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.OperatorKey.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCloudlet
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
				return ErrInvalidLengthCloudlet
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCloudlet(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCloudlet
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
func (m *Cloudlet) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCloudlet
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
			return fmt.Errorf("proto: Cloudlet: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Cloudlet: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Fields", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCloudlet
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
				return ErrInvalidLengthCloudlet
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
					return ErrIntOverflowCloudlet
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
				return ErrInvalidLengthCloudlet
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Key.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AccessUri", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCloudlet
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
				return ErrInvalidLengthCloudlet
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AccessUri = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Location", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCloudlet
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
				return ErrInvalidLengthCloudlet
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Location.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCloudlet(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCloudlet
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
func (m *CloudletInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCloudlet
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
			return fmt.Errorf("proto: CloudletInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CloudletInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Key", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCloudlet
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
				return ErrInvalidLengthCloudlet
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Key.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field State", wireType)
			}
			m.State = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCloudlet
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.State |= (CloudletState(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Resources", wireType)
			}
			m.Resources = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCloudlet
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Resources |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipCloudlet(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCloudlet
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
func skipCloudlet(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCloudlet
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
					return 0, ErrIntOverflowCloudlet
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
					return 0, ErrIntOverflowCloudlet
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
				return 0, ErrInvalidLengthCloudlet
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowCloudlet
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
				next, err := skipCloudlet(dAtA[start:])
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
	ErrInvalidLengthCloudlet = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCloudlet   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("cloudlet.proto", fileDescriptorCloudlet) }

var fileDescriptorCloudlet = []byte{
	// 685 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x93, 0x4d, 0x6b, 0x13, 0x5d,
	0x14, 0xc7, 0x7b, 0x93, 0xb4, 0x4f, 0x73, 0x93, 0x27, 0x4f, 0x3a, 0x4f, 0x2d, 0xd3, 0x50, 0xd3,
	0x10, 0x37, 0xa1, 0x98, 0x19, 0xa9, 0x1b, 0xe9, 0xa6, 0xb4, 0x11, 0x54, 0x5a, 0x28, 0x4c, 0x6d,
	0x57, 0x42, 0x98, 0xdc, 0x39, 0x99, 0x5c, 0x3a, 0x73, 0xef, 0x70, 0xe7, 0x8e, 0x31, 0xae, 0xc4,
	0x95, 0x88, 0x3b, 0xbf, 0x80, 0x1f, 0x41, 0xfd, 0x14, 0x5d, 0x0a, 0xee, 0x45, 0x8b, 0x0b, 0x37,
	0x82, 0xd0, 0x2e, 0x5c, 0xca, 0xdc, 0x99, 0xbc, 0xd4, 0x06, 0x2d, 0x74, 0x13, 0xce, 0xdb, 0xff,
	0x77, 0xce, 0x3d, 0x39, 0x83, 0x4b, 0xc4, 0xe3, 0x91, 0xe3, 0x81, 0x34, 0x02, 0xc1, 0x25, 0xd7,
	0xf2, 0xe0, 0xb8, 0xa0, 0xcc, 0xca, 0x8a, 0xcb, 0xb9, 0xeb, 0x81, 0x69, 0x07, 0xd4, 0xb4, 0x19,
	0xe3, 0xd2, 0x96, 0x94, 0xb3, 0x30, 0x29, 0xac, 0xdc, 0x71, 0xa9, 0xec, 0x45, 0x1d, 0x83, 0x70,
	0xdf, 0xf4, 0x79, 0x87, 0x7a, 0xb1, 0xf0, 0x89, 0x19, 0xff, 0x36, 0x15, 0xd3, 0x54, 0x75, 0x2e,
	0xb0, 0x91, 0x91, 0x2a, 0xef, 0x5d, 0x4e, 0x49, 0x9a, 0x2e, 0xb0, 0x26, 0xf1, 0x87, 0xee, 0x84,
	0x91, 0x82, 0x4a, 0x3c, 0x00, 0x61, 0x4b, 0x2e, 0x52, 0xbf, 0x28, 0x20, 0x8c, 0xbc, 0xf4, 0x25,
	0x95, 0xd6, 0x5f, 0xdb, 0x38, 0x4d, 0xdf, 0x96, 0xa4, 0xd7, 0x04, 0xe6, 0x52, 0x06, 0xa6, 0xe3,
	0x43, 0x53, 0x49, 0x4d, 0x8f, 0x93, 0x14, 0xd2, 0x9c, 0x80, 0xb8, 0xdc, 0xe5, 0xc9, 0x08, 0x9d,
	0xa8, 0xab, 0xbc, 0xa4, 0x3a, 0xb6, 0x92, 0xf2, 0x7a, 0x80, 0x0b, 0xad, 0x74, 0x9f, 0x3b, 0x30,
	0xd0, 0x36, 0x71, 0x71, 0x38, 0x62, 0xfb, 0x08, 0x06, 0x3a, 0xaa, 0xa1, 0x46, 0x61, 0x7d, 0xc9,
	0x18, 0xed, 0xd8, 0xd8, 0x4b, 0xd3, 0x3b, 0x30, 0xd8, 0xce, 0x1d, 0x7f, 0x5a, 0x9d, 0xb1, 0x0a,
	0x7c, 0x1c, 0xd2, 0x34, 0x9c, 0x63, 0xb6, 0x0f, 0x7a, 0xa6, 0x86, 0x1a, 0x79, 0x4b, 0xd9, 0x1b,
	0xc5, 0x6f, 0xa7, 0x3a, 0xfa, 0x79, 0xaa, 0xa3, 0xb7, 0x6f, 0x56, 0x51, 0xfd, 0x5d, 0x06, 0xcf,
	0x0f, 0x5b, 0x6a, 0x4b, 0x78, 0xae, 0x4b, 0xc1, 0x73, 0x42, 0x1d, 0xd5, 0xb2, 0x8d, 0xbc, 0x95,
	0x7a, 0x9a, 0x81, 0xb3, 0x71, 0xfb, 0xcc, 0x85, 0xf6, 0x13, 0xc3, 0xa6, 0xed, 0xe3, 0x42, 0xed,
	0x06, 0xc6, 0x36, 0x21, 0x10, 0x86, 0xed, 0x48, 0x50, 0x3d, 0x17, 0x37, 0xdf, 0xce, 0xbd, 0x38,
	0xd3, 0x91, 0x95, 0x4f, 0xe2, 0x07, 0x82, 0x6a, 0x9b, 0x78, 0xde, 0xe3, 0x44, 0xdd, 0x84, 0x3e,
	0xab, 0xc8, 0xd7, 0x0d, 0x87, 0x86, 0x52, 0xd0, 0x4e, 0x24, 0xc1, 0x69, 0xab, 0xdd, 0xb6, 0x93,
	0xdd, 0x1a, 0xbb, 0x9c, 0xa4, 0x0d, 0x46, 0xa2, 0x8d, 0x7e, 0xfc, 0x90, 0x1f, 0xa7, 0x3a, 0x7a,
	0x76, 0xa6, 0xa3, 0x97, 0xef, 0x97, 0xdd, 0xdd, 0x34, 0x63, 0xdc, 0xe7, 0x82, 0x3e, 0xe5, 0x4c,
	0xda, 0xde, 0x16, 0x21, 0x91, 0xb0, 0xc9, 0xe0, 0xe6, 0x28, 0x77, 0x08, 0x42, 0x52, 0x32, 0x2d,
	0xd3, 0xe2, 0x91, 0x08, 0x61, 0xec, 0xef, 0x07, 0x00, 0xce, 0xd8, 0x7d, 0x48, 0x7d, 0x08, 0xa5,
	0xed, 0x07, 0xf5, 0x57, 0x08, 0x17, 0x87, 0x2f, 0x7f, 0xc0, 0xba, 0x7c, 0xb8, 0x1f, 0x74, 0xd9,
	0xfd, 0x18, 0x78, 0x36, 0x94, 0xb6, 0x4c, 0xfe, 0x97, 0xd2, 0xba, 0x3e, 0x45, 0xb1, 0x1f, 0xe7,
	0xad, 0xa4, 0x4c, 0x5b, 0xc1, 0x79, 0x01, 0x21, 0x8f, 0x04, 0x81, 0x50, 0xcf, 0xd6, 0x50, 0x23,
	0x67, 0x8d, 0x03, 0x6b, 0x8f, 0xf0, 0xbf, 0xe7, 0x54, 0x5a, 0x01, 0xff, 0x73, 0xc0, 0x8e, 0x18,
	0xef, 0xb3, 0xf2, 0x8c, 0xa6, 0xe3, 0xc5, 0x16, 0x67, 0x5d, 0xea, 0x46, 0x82, 0x32, 0x77, 0x2f,
	0x00, 0x16, 0x4a, 0x9b, 0x1c, 0x95, 0x91, 0xb6, 0x8c, 0xaf, 0x4d, 0x64, 0x76, 0xa2, 0x0e, 0x08,
	0x06, 0x12, 0xc2, 0x72, 0x46, 0xcb, 0xe3, 0x59, 0x0b, 0x6c, 0x67, 0x50, 0xce, 0xae, 0x7f, 0xcf,
	0x8c, 0x6f, 0x72, 0x2b, 0xa0, 0xda, 0x21, 0x2e, 0xb5, 0x04, 0xd8, 0x12, 0x46, 0x57, 0xf3, 0xff,
	0x94, 0xf1, 0x2b, 0x0b, 0x13, 0x41, 0x4b, 0x7d, 0x56, 0xf5, 0x95, 0xe7, 0x1f, 0xbf, 0xbe, 0xce,
	0x2c, 0xd5, 0x17, 0x4c, 0xa2, 0x00, 0xa6, 0x03, 0x8f, 0xc1, 0x8b, 0xcf, 0x75, 0x03, 0xad, 0xc5,
	0xdc, 0xbb, 0xe0, 0xc1, 0x95, 0xb8, 0x8e, 0x02, 0x5c, 0xe0, 0x1e, 0x04, 0xce, 0xd5, 0xe6, 0x8d,
	0x14, 0xe0, 0x77, 0x6e, 0x71, 0xbf, 0xc7, 0xfb, 0x7f, 0xa6, 0x4e, 0x0b, 0xd6, 0x2b, 0x8a, 0xbb,
	0x58, 0xff, 0xcf, 0x0c, 0x7b, 0xbc, 0x7f, 0x8e, 0x7a, 0x0b, 0x6d, 0x97, 0x8f, 0xbf, 0x54, 0x67,
	0x8e, 0x4f, 0xaa, 0xe8, 0xc3, 0x49, 0x15, 0x7d, 0x3e, 0xa9, 0xa2, 0xce, 0x9c, 0xd2, 0xdf, 0xfe,
	0x15, 0x00, 0x00, 0xff, 0xff, 0xe1, 0x4b, 0xa8, 0x24, 0x6b, 0x05, 0x00, 0x00,
}
