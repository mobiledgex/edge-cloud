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
import "sync"
import "github.com/mobiledgex/edge-cloud/util"
import google_protobuf "github.com/gogo/protobuf/types"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

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
	// IP to use to connect to and control cloudlet site
	AccessIp []byte `protobuf:"bytes,4,opt,name=access_ip,json=accessIp,proto3" json:"access_ip,omitempty"`
	// Location of the cloudlet site (lat, long?)
	Location distributed_match_engine.Loc `protobuf:"bytes,5,opt,name=location" json:"location"`
}

func (m *Cloudlet) Reset()                    { *m = Cloudlet{} }
func (m *Cloudlet) String() string            { return proto.CompactTextString(m) }
func (*Cloudlet) ProtoMessage()               {}
func (*Cloudlet) Descriptor() ([]byte, []int) { return fileDescriptorCloudlet, []int{1} }

func init() {
	proto.RegisterType((*CloudletKey)(nil), "edgeproto.CloudletKey")
	proto.RegisterType((*Cloudlet)(nil), "edgeproto.Cloudlet")
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
	if len(m.AccessIp) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintCloudlet(dAtA, i, uint64(len(m.AccessIp)))
		i += copy(dAtA[i:], m.AccessIp)
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
	return true
}

const CloudletFieldKeyOperatorKeyName = "2.1.1"
const CloudletFieldKeyName = "2.2"
const CloudletFieldAccessIp = "4"
const CloudletFieldLocationLat = "5.1"
const CloudletFieldLocationLong = "5.2"
const CloudletFieldLocationHorizontalAccuracy = "5.3"
const CloudletFieldLocationVerticalAccuracy = "5.4"
const CloudletFieldLocationAltitude = "5.5"
const CloudletFieldLocationCourse = "5.6"
const CloudletFieldLocationSpeed = "5.7"
const CloudletFieldLocationTimestampSeconds = "5.8.1"
const CloudletFieldLocationTimestampNanos = "5.8.2"

var CloudletAllFields = []string{
	CloudletFieldKeyOperatorKeyName,
	CloudletFieldKeyName,
	CloudletFieldAccessIp,
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

func (m *Cloudlet) CopyInFields(src *Cloudlet) {
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
				m.Key.OperatorKey.Name = src.Key.OperatorKey.Name
			}
		}
		if _, set := fmap["2.2"]; set {
			m.Key.Name = src.Key.Name
		}
	}
	if _, set := fmap["4"]; set {
		if m.AccessIp == nil || len(m.AccessIp) < len(src.AccessIp) {
			m.AccessIp = make([]byte, len(src.AccessIp))
		}
		copy(m.AccessIp, src.AccessIp)
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
	objstore     objstore.ObjStore
	listCloudlet map[CloudletKey]struct{}
}

func NewCloudletStore(objstore objstore.ObjStore) CloudletStore {
	return CloudletStore{objstore: objstore}
}

type CloudletCacher interface {
	SyncCloudletUpdate(m *Cloudlet, rev int64)
	SyncCloudletDelete(m *Cloudlet, rev int64)
	SyncCloudletPrune(current map[CloudletKey]struct{})
	SyncCloudletRevOnly(rev int64)
}

func (s *CloudletStore) Create(m *Cloudlet, wait func(int64)) (*Result, error) {
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

func (s *CloudletStore) Update(m *Cloudlet, wait func(int64)) (*Result, error) {
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

// Sync will sync changes for any Cloudlet objects.
func (s *CloudletStore) Sync(ctx context.Context, cacher CloudletCacher) error {
	str := objstore.DbKeyPrefixString(&CloudletKey{})
	return s.objstore.Sync(ctx, str, func(in *objstore.SyncCbData) {
		obj := Cloudlet{}
		// Even on parse error, we should still call back to keep
		// the revision numbers in sync so no caller hangs on wait.
		action := in.Action
		if action == objstore.SyncUpdate || action == objstore.SyncList {
			err := json.Unmarshal(in.Value, &obj)
			if err != nil {
				util.WarnLog("Failed to parse Cloudlet data", "val", string(in.Value))
				action = objstore.SyncRevOnly
			}
		} else if action == objstore.SyncDelete {
			keystr := objstore.DbKeyPrefixRemove(string(in.Key))
			CloudletKeyStringParse(keystr, obj.GetKey())
		}
		util.DebugLog(util.DebugLevelApi, "Sync cb", "action", objstore.SyncActionStrs[in.Action], "key", string(in.Key), "value", string(in.Value), "rev", in.Rev)
		switch action {
		case objstore.SyncUpdate:
			cacher.SyncCloudletUpdate(&obj, in.Rev)
		case objstore.SyncDelete:
			cacher.SyncCloudletDelete(&obj, in.Rev)
		case objstore.SyncListStart:
			s.listCloudlet = make(map[CloudletKey]struct{})
		case objstore.SyncList:
			s.listCloudlet[obj.Key] = struct{}{}
			cacher.SyncCloudletUpdate(&obj, in.Rev)
		case objstore.SyncListEnd:
			cacher.SyncCloudletPrune(s.listCloudlet)
			s.listCloudlet = nil
		case objstore.SyncRevOnly:
			cacher.SyncCloudletRevOnly(in.Rev)
		}
	})
}

// CloudletCache caches Cloudlet objects in memory in a hash table
// and keeps them in sync with the database.
type CloudletCache struct {
	Store      *CloudletStore
	Objs       map[CloudletKey]*Cloudlet
	Rev        int64
	Mux        util.Mutex
	Cond       sync.Cond
	initWait   bool
	syncDone   bool
	syncCancel context.CancelFunc
	notifyCb   func(obj *CloudletKey)
}

func NewCloudletCache(store *CloudletStore) *CloudletCache {
	cache := CloudletCache{
		Store:    store,
		Objs:     make(map[CloudletKey]*Cloudlet),
		initWait: true,
	}
	cache.Mux.InitCond(&cache.Cond)

	ctx, cancel := context.WithCancel(context.Background())
	cache.syncCancel = cancel
	go func() {
		err := cache.Store.Sync(ctx, &cache)
		if err != nil {
			util.WarnLog("Cloudlet Sync failed", "err", err)
		}
		cache.syncDone = true
		cache.Cond.Broadcast()
	}()
	return &cache
}

func (c *CloudletCache) WaitInitSyncDone() {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for c.initWait {
		c.Cond.Wait()
	}
}

func (c *CloudletCache) Done() {
	c.syncCancel()
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for !c.syncDone {
		c.Cond.Wait()
	}
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

func (c *CloudletCache) SyncCloudletUpdate(in *Cloudlet, rev int64) {
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

func (c *CloudletCache) SyncCloudletDelete(in *Cloudlet, rev int64) {
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

func (c *CloudletCache) SyncCloudletPrune(current map[CloudletKey]struct{}) {
	deleted := make(map[CloudletKey]struct{})
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

func (c *CloudletCache) SyncCloudletRevOnly(rev int64) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	c.Rev = rev
	util.DebugLog(util.DebugLevelApi, "SyncRevOnly", "rev", rev)
	c.Cond.Broadcast()
}

func (c *CloudletCache) SyncWait(rev int64) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	util.DebugLog(util.DebugLevelApi, "SyncWait", "cache-rev", c.Rev, "wait-rev", rev)
	for c.Rev < rev {
		c.Cond.Wait()
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
	c.notifyCb = fn
}

func (m *Cloudlet) GetKey() *CloudletKey {
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
	l = len(m.AccessIp)
	if l > 0 {
		n += 1 + l + sovCloudlet(uint64(l))
	}
	l = m.Location.Size()
	n += 1 + l + sovCloudlet(uint64(l))
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
				return fmt.Errorf("proto: wrong wireType = %d for field AccessIp", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCloudlet
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
				return ErrInvalidLengthCloudlet
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AccessIp = append(m.AccessIp[:0], dAtA[iNdEx:postIndex]...)
			if m.AccessIp == nil {
				m.AccessIp = []byte{}
			}
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
	// 565 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x92, 0x4f, 0x6b, 0x13, 0x4f,
	0x18, 0xc7, 0x3b, 0x69, 0x7f, 0x25, 0xd9, 0x84, 0xfc, 0xec, 0x2a, 0x61, 0x8d, 0x35, 0x0d, 0x7b,
	0x0a, 0xe2, 0xee, 0x4a, 0xbd, 0x48, 0x2e, 0xa5, 0x8d, 0xa0, 0xd2, 0x82, 0x90, 0x6a, 0xaf, 0x61,
	0x33, 0xfb, 0x74, 0x33, 0x38, 0xbb, 0xcf, 0xb2, 0x3b, 0x6b, 0x8c, 0x27, 0xf1, 0xe8, 0xd5, 0x37,
	0xe0, 0xd5, 0x9b, 0xf8, 0x2a, 0x72, 0x14, 0xbc, 0x8b, 0x06, 0x0f, 0x5e, 0x04, 0x21, 0x1e, 0x3c,
	0xca, 0xce, 0x4e, 0xfe, 0x68, 0x83, 0x0a, 0xbd, 0x0c, 0xcf, 0xdf, 0xcf, 0xf3, 0xcc, 0x77, 0x46,
	0xab, 0x52, 0x8e, 0xa9, 0xc7, 0x41, 0xd8, 0x51, 0x8c, 0x02, 0xf5, 0x12, 0x78, 0x3e, 0x48, 0xb3,
	0xbe, 0xed, 0x23, 0xfa, 0x1c, 0x1c, 0x37, 0x62, 0x8e, 0x1b, 0x86, 0x28, 0x5c, 0xc1, 0x30, 0x4c,
	0xf2, 0xc2, 0xfa, 0x2d, 0x9f, 0x89, 0x41, 0xda, 0xb7, 0x29, 0x06, 0x4e, 0x80, 0x7d, 0xc6, 0xb3,
	0xc6, 0x27, 0x4e, 0x76, 0x5a, 0x92, 0xe9, 0xc8, 0x3a, 0x1f, 0xc2, 0xb9, 0xa1, 0x3a, 0xef, 0xfc,
	0x5b, 0x27, 0xb5, 0x7c, 0x08, 0x2d, 0x1a, 0xcc, 0xdc, 0x25, 0x43, 0x81, 0xaa, 0x18, 0x41, 0xec,
	0x0a, 0x8c, 0x95, 0x5f, 0x89, 0x21, 0x49, 0xb9, 0xba, 0x49, 0xbd, 0xf3, 0xd7, 0x31, 0x9e, 0x15,
	0xb8, 0x82, 0x0e, 0x2c, 0x08, 0x7d, 0x16, 0x82, 0xe3, 0x05, 0x60, 0xc9, 0x56, 0x87, 0x23, 0x55,
	0x10, 0x6b, 0x09, 0xe2, 0xa3, 0x8f, 0xf9, 0x0a, 0xfd, 0xf4, 0x54, 0x7a, 0x79, 0x75, 0x66, 0xe5,
	0xe5, 0x66, 0xa4, 0x95, 0x3b, 0x4a, 0xcf, 0x43, 0x18, 0xe9, 0x7b, 0x5a, 0x65, 0xb6, 0x62, 0xef,
	0x11, 0x8c, 0x0c, 0xd2, 0x24, 0xad, 0xf2, 0x6e, 0xcd, 0x9e, 0x6b, 0x6c, 0xdf, 0x57, 0xe9, 0x43,
	0x18, 0x1d, 0x6c, 0x8c, 0x3f, 0xec, 0xac, 0x75, 0xcb, 0xb8, 0x08, 0xe9, 0xba, 0xb6, 0x11, 0xba,
	0x01, 0x18, 0x85, 0x26, 0x69, 0x95, 0xba, 0xd2, 0x6e, 0x57, 0xbe, 0x4c, 0x0d, 0xf2, 0x63, 0x6a,
	0x90, 0x37, 0xaf, 0x76, 0x88, 0xf9, 0xba, 0xa0, 0x15, 0x67, 0x23, 0xf5, 0x9a, 0xb6, 0x79, 0xca,
	0x80, 0x7b, 0x89, 0x41, 0x9a, 0xeb, 0xad, 0x52, 0x57, 0x79, 0xba, 0xad, 0xad, 0x67, 0xe3, 0x0b,
	0x67, 0xc6, 0x2f, 0x2d, 0xab, 0xc6, 0x67, 0x85, 0xfa, 0x15, 0xad, 0xe4, 0x52, 0x0a, 0x49, 0xd2,
	0x63, 0x91, 0xb1, 0xd1, 0x24, 0xad, 0x4a, 0xb7, 0x98, 0x07, 0xee, 0x45, 0xfa, 0x9e, 0x56, 0xe4,
	0x48, 0xe5, 0x5f, 0x30, 0xfe, 0x93, 0xc4, 0xab, 0xb6, 0xc7, 0x12, 0x11, 0xb3, 0x7e, 0x2a, 0xc0,
	0xeb, 0x49, 0x4d, 0x7b, 0xb9, 0xa6, 0xf6, 0x11, 0x52, 0x05, 0x9e, 0x37, 0xb5, 0x87, 0xd9, 0x05,
	0xbe, 0x4d, 0x0d, 0xf2, 0xec, 0xbb, 0x41, 0x5e, 0xbc, 0xbd, 0xec, 0x1f, 0xa9, 0x8c, 0x7d, 0x17,
	0x63, 0xf6, 0x14, 0x43, 0xe1, 0xf2, 0x7d, 0x4a, 0xd3, 0xd8, 0xa5, 0xa3, 0xeb, 0xf3, 0xdc, 0x09,
	0xc4, 0x82, 0xd1, 0x55, 0x99, 0x0e, 0xa6, 0x71, 0x02, 0x0b, 0xff, 0x38, 0x02, 0xf0, 0x16, 0xee,
	0x03, 0x16, 0x40, 0x22, 0xdc, 0x20, 0xda, 0xfd, 0x5a, 0x58, 0x3c, 0xcf, 0x7e, 0xc4, 0xf4, 0x13,
	0xad, 0xda, 0x89, 0xc1, 0x15, 0x30, 0x17, 0xf0, 0xe2, 0x0a, 0x6d, 0xea, 0x5b, 0x4b, 0xc1, 0xae,
	0xfc, 0x61, 0xe6, 0xf6, 0xf3, 0xf7, 0x9f, 0x5f, 0x16, 0x6a, 0xe6, 0x96, 0x43, 0x25, 0xc0, 0xf1,
	0xe0, 0x31, 0xf0, 0xec, 0xe5, 0xda, 0xe4, 0x5a, 0xc6, 0xbd, 0x0d, 0x1c, 0xce, 0xc5, 0xf5, 0x24,
	0xe0, 0x0c, 0xf7, 0x61, 0xe4, 0x9d, 0x6f, 0xdf, 0x54, 0x02, 0x7e, 0xe7, 0x56, 0x8e, 0x07, 0x38,
	0xfc, 0x33, 0x75, 0x55, 0xd0, 0xac, 0x4b, 0xee, 0x25, 0xf3, 0x7f, 0x27, 0x19, 0xe0, 0xf0, 0x17,
	0xea, 0x0d, 0x72, 0x70, 0x61, 0xfc, 0xa9, 0xb1, 0x36, 0x9e, 0x34, 0xc8, 0xbb, 0x49, 0x83, 0x7c,
	0x9c, 0x34, 0x48, 0x7f, 0x53, 0xf6, 0xdf, 0xfc, 0x19, 0x00, 0x00, 0xff, 0xff, 0xe6, 0x49, 0xa7,
	0x43, 0x76, 0x04, 0x00, 0x00,
}
