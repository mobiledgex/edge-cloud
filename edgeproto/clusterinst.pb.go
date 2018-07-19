// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: clusterinst.proto

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

import "encoding/json"
import "github.com/mobiledgex/edge-cloud/objstore"
import "github.com/mobiledgex/edge-cloud/util"
import "github.com/mobiledgex/edge-cloud/log"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type ClusterInstKey struct {
	// cluster key
	ClusterKey ClusterKey `protobuf:"bytes,1,opt,name=cluster_key,json=clusterKey" json:"cluster_key"`
	// cloudlet it's on
	CloudletKey CloudletKey `protobuf:"bytes,2,opt,name=cloudlet_key,json=cloudletKey" json:"cloudlet_key"`
}

func (m *ClusterInstKey) Reset()                    { *m = ClusterInstKey{} }
func (m *ClusterInstKey) String() string            { return proto.CompactTextString(m) }
func (*ClusterInstKey) ProtoMessage()               {}
func (*ClusterInstKey) Descriptor() ([]byte, []int) { return fileDescriptorClusterinst, []int{0} }

type ClusterInst struct {
	Fields []string `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	// Unique key
	Key ClusterInstKey `protobuf:"bytes,2,opt,name=key" json:"key"`
	// flavor (cached from cluster)
	Flavor FlavorKey `protobuf:"bytes,3,opt,name=flavor" json:"flavor"`
	// number of nodes in the cluster (cached from Cluster)
	Nodes int32 `protobuf:"varint,4,opt,name=nodes,proto3" json:"nodes,omitempty"`
	// Future: policy options on where this cluster can be created.
	// type of instance
	Liveness Liveness `protobuf:"varint,9,opt,name=liveness,proto3,enum=edgeproto.Liveness" json:"liveness,omitempty"`
}

func (m *ClusterInst) Reset()                    { *m = ClusterInst{} }
func (m *ClusterInst) String() string            { return proto.CompactTextString(m) }
func (*ClusterInst) ProtoMessage()               {}
func (*ClusterInst) Descriptor() ([]byte, []int) { return fileDescriptorClusterinst, []int{1} }

func init() {
	proto.RegisterType((*ClusterInstKey)(nil), "edgeproto.ClusterInstKey")
	proto.RegisterType((*ClusterInst)(nil), "edgeproto.ClusterInst")
}
func (this *ClusterInstKey) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 6)
	s = append(s, "&edgeproto.ClusterInstKey{")
	s = append(s, "ClusterKey: "+strings.Replace(this.ClusterKey.GoString(), `&`, ``, 1)+",\n")
	s = append(s, "CloudletKey: "+strings.Replace(this.CloudletKey.GoString(), `&`, ``, 1)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringClusterinst(v interface{}, typ string) string {
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

// Client API for ClusterInstApi service

type ClusterInstApiClient interface {
	CreateClusterInst(ctx context.Context, in *ClusterInst, opts ...grpc.CallOption) (*Result, error)
	DeleteClusterInst(ctx context.Context, in *ClusterInst, opts ...grpc.CallOption) (*Result, error)
	UpdateClusterInst(ctx context.Context, in *ClusterInst, opts ...grpc.CallOption) (*Result, error)
	ShowClusterInst(ctx context.Context, in *ClusterInst, opts ...grpc.CallOption) (ClusterInstApi_ShowClusterInstClient, error)
}

type clusterInstApiClient struct {
	cc *grpc.ClientConn
}

func NewClusterInstApiClient(cc *grpc.ClientConn) ClusterInstApiClient {
	return &clusterInstApiClient{cc}
}

func (c *clusterInstApiClient) CreateClusterInst(ctx context.Context, in *ClusterInst, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.ClusterInstApi/CreateClusterInst", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterInstApiClient) DeleteClusterInst(ctx context.Context, in *ClusterInst, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.ClusterInstApi/DeleteClusterInst", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterInstApiClient) UpdateClusterInst(ctx context.Context, in *ClusterInst, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.ClusterInstApi/UpdateClusterInst", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterInstApiClient) ShowClusterInst(ctx context.Context, in *ClusterInst, opts ...grpc.CallOption) (ClusterInstApi_ShowClusterInstClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_ClusterInstApi_serviceDesc.Streams[0], c.cc, "/edgeproto.ClusterInstApi/ShowClusterInst", opts...)
	if err != nil {
		return nil, err
	}
	x := &clusterInstApiShowClusterInstClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ClusterInstApi_ShowClusterInstClient interface {
	Recv() (*ClusterInst, error)
	grpc.ClientStream
}

type clusterInstApiShowClusterInstClient struct {
	grpc.ClientStream
}

func (x *clusterInstApiShowClusterInstClient) Recv() (*ClusterInst, error) {
	m := new(ClusterInst)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for ClusterInstApi service

type ClusterInstApiServer interface {
	CreateClusterInst(context.Context, *ClusterInst) (*Result, error)
	DeleteClusterInst(context.Context, *ClusterInst) (*Result, error)
	UpdateClusterInst(context.Context, *ClusterInst) (*Result, error)
	ShowClusterInst(*ClusterInst, ClusterInstApi_ShowClusterInstServer) error
}

func RegisterClusterInstApiServer(s *grpc.Server, srv ClusterInstApiServer) {
	s.RegisterService(&_ClusterInstApi_serviceDesc, srv)
}

func _ClusterInstApi_CreateClusterInst_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClusterInst)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterInstApiServer).CreateClusterInst(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.ClusterInstApi/CreateClusterInst",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterInstApiServer).CreateClusterInst(ctx, req.(*ClusterInst))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClusterInstApi_DeleteClusterInst_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClusterInst)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterInstApiServer).DeleteClusterInst(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.ClusterInstApi/DeleteClusterInst",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterInstApiServer).DeleteClusterInst(ctx, req.(*ClusterInst))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClusterInstApi_UpdateClusterInst_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClusterInst)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterInstApiServer).UpdateClusterInst(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.ClusterInstApi/UpdateClusterInst",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterInstApiServer).UpdateClusterInst(ctx, req.(*ClusterInst))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClusterInstApi_ShowClusterInst_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ClusterInst)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ClusterInstApiServer).ShowClusterInst(m, &clusterInstApiShowClusterInstServer{stream})
}

type ClusterInstApi_ShowClusterInstServer interface {
	Send(*ClusterInst) error
	grpc.ServerStream
}

type clusterInstApiShowClusterInstServer struct {
	grpc.ServerStream
}

func (x *clusterInstApiShowClusterInstServer) Send(m *ClusterInst) error {
	return x.ServerStream.SendMsg(m)
}

var _ClusterInstApi_serviceDesc = grpc.ServiceDesc{
	ServiceName: "edgeproto.ClusterInstApi",
	HandlerType: (*ClusterInstApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateClusterInst",
			Handler:    _ClusterInstApi_CreateClusterInst_Handler,
		},
		{
			MethodName: "DeleteClusterInst",
			Handler:    _ClusterInstApi_DeleteClusterInst_Handler,
		},
		{
			MethodName: "UpdateClusterInst",
			Handler:    _ClusterInstApi_UpdateClusterInst_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ShowClusterInst",
			Handler:       _ClusterInstApi_ShowClusterInst_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "clusterinst.proto",
}

func (m *ClusterInstKey) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ClusterInstKey) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintClusterinst(dAtA, i, uint64(m.ClusterKey.Size()))
	n1, err := m.ClusterKey.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	dAtA[i] = 0x12
	i++
	i = encodeVarintClusterinst(dAtA, i, uint64(m.CloudletKey.Size()))
	n2, err := m.CloudletKey.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n2
	return i, nil
}

func (m *ClusterInst) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ClusterInst) MarshalTo(dAtA []byte) (int, error) {
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
	i = encodeVarintClusterinst(dAtA, i, uint64(m.Key.Size()))
	n3, err := m.Key.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n3
	dAtA[i] = 0x1a
	i++
	i = encodeVarintClusterinst(dAtA, i, uint64(m.Flavor.Size()))
	n4, err := m.Flavor.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n4
	if m.Nodes != 0 {
		dAtA[i] = 0x20
		i++
		i = encodeVarintClusterinst(dAtA, i, uint64(m.Nodes))
	}
	if m.Liveness != 0 {
		dAtA[i] = 0x48
		i++
		i = encodeVarintClusterinst(dAtA, i, uint64(m.Liveness))
	}
	return i, nil
}

func encodeVarintClusterinst(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *ClusterInstKey) Matches(filter *ClusterInstKey) bool {
	if filter == nil {
		return true
	}
	if !m.ClusterKey.Matches(&filter.ClusterKey) {
		return false
	}
	if !m.CloudletKey.Matches(&filter.CloudletKey) {
		return false
	}
	return true
}

func (m *ClusterInstKey) MatchesIgnoreBackend(filter *ClusterInstKey) bool {
	if filter == nil {
		return true
	}
	if !m.ClusterKey.MatchesIgnoreBackend(&filter.ClusterKey) {
		return false
	}
	if !m.CloudletKey.MatchesIgnoreBackend(&filter.CloudletKey) {
		return false
	}
	return true
}

func (m *ClusterInstKey) CopyInFields(src *ClusterInstKey) {
	m.ClusterKey.Name = src.ClusterKey.Name
	m.CloudletKey.OperatorKey.Name = src.CloudletKey.OperatorKey.Name
	m.CloudletKey.Name = src.CloudletKey.Name
}

func (m *ClusterInstKey) GetKeyString() string {
	key, err := json.Marshal(m)
	if err != nil {
		log.FatalLog("Failed to marshal ClusterInstKey key string", "obj", m)
	}
	return string(key)
}

func ClusterInstKeyStringParse(str string, key *ClusterInstKey) {
	err := json.Unmarshal([]byte(str), key)
	if err != nil {
		log.FatalLog("Failed to unmarshal ClusterInstKey key string", "str", str)
	}
}

func (m *ClusterInst) Matches(filter *ClusterInst) bool {
	if filter == nil {
		return true
	}
	if !m.Key.Matches(&filter.Key) {
		return false
	}
	if !m.Flavor.Matches(&filter.Flavor) {
		return false
	}
	if filter.Nodes != 0 && filter.Nodes != m.Nodes {
		return false
	}
	if filter.Liveness != 0 && filter.Liveness != m.Liveness {
		return false
	}
	return true
}

func (m *ClusterInst) MatchesIgnoreBackend(filter *ClusterInst) bool {
	if filter == nil {
		return true
	}
	if !m.Key.MatchesIgnoreBackend(&filter.Key) {
		return false
	}
	return true
}

const ClusterInstFieldKey = "2"
const ClusterInstFieldKeyClusterKey = "2.1"
const ClusterInstFieldKeyClusterKeyName = "2.1.1"
const ClusterInstFieldKeyCloudletKey = "2.2"
const ClusterInstFieldKeyCloudletKeyOperatorKey = "2.2.1"
const ClusterInstFieldKeyCloudletKeyOperatorKeyName = "2.2.1.1"
const ClusterInstFieldKeyCloudletKeyName = "2.2.2"
const ClusterInstFieldFlavor = "3"
const ClusterInstFieldFlavorName = "3.1"
const ClusterInstFieldNodes = "4"
const ClusterInstFieldLiveness = "9"

var ClusterInstAllFields = []string{
	ClusterInstFieldKeyClusterKeyName,
	ClusterInstFieldKeyCloudletKeyOperatorKeyName,
	ClusterInstFieldKeyCloudletKeyName,
	ClusterInstFieldFlavorName,
	ClusterInstFieldNodes,
	ClusterInstFieldLiveness,
}

var ClusterInstAllFieldsMap = map[string]struct{}{
	ClusterInstFieldKeyClusterKeyName:             struct{}{},
	ClusterInstFieldKeyCloudletKeyOperatorKeyName: struct{}{},
	ClusterInstFieldKeyCloudletKeyName:            struct{}{},
	ClusterInstFieldFlavorName:                    struct{}{},
	ClusterInstFieldNodes:                         struct{}{},
	ClusterInstFieldLiveness:                      struct{}{},
}

func (m *ClusterInst) DiffFields(o *ClusterInst, fields map[string]struct{}) {
	if m.Key.ClusterKey.Name != o.Key.ClusterKey.Name {
		fields[ClusterInstFieldKeyClusterKeyName] = struct{}{}
		fields[ClusterInstFieldKeyClusterKey] = struct{}{}
		fields[ClusterInstFieldKey] = struct{}{}
	}
	if m.Key.CloudletKey.OperatorKey.Name != o.Key.CloudletKey.OperatorKey.Name {
		fields[ClusterInstFieldKeyCloudletKeyOperatorKeyName] = struct{}{}
		fields[ClusterInstFieldKeyCloudletKeyOperatorKey] = struct{}{}
		fields[ClusterInstFieldKeyCloudletKey] = struct{}{}
		fields[ClusterInstFieldKey] = struct{}{}
	}
	if m.Key.CloudletKey.Name != o.Key.CloudletKey.Name {
		fields[ClusterInstFieldKeyCloudletKeyName] = struct{}{}
		fields[ClusterInstFieldKeyCloudletKey] = struct{}{}
		fields[ClusterInstFieldKey] = struct{}{}
	}
	if m.Flavor.Name != o.Flavor.Name {
		fields[ClusterInstFieldFlavorName] = struct{}{}
		fields[ClusterInstFieldFlavor] = struct{}{}
	}
	if m.Nodes != o.Nodes {
		fields[ClusterInstFieldNodes] = struct{}{}
	}
	if m.Liveness != o.Liveness {
		fields[ClusterInstFieldLiveness] = struct{}{}
	}
}

func (m *ClusterInst) CopyInFields(src *ClusterInst) {
	fmap := MakeFieldMap(src.Fields)
	if _, set := fmap["2"]; set {
		if _, set := fmap["2.1"]; set {
			if _, set := fmap["2.1.1"]; set {
				m.Key.ClusterKey.Name = src.Key.ClusterKey.Name
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
	}
	if _, set := fmap["3"]; set {
		if _, set := fmap["3.1"]; set {
			m.Flavor.Name = src.Flavor.Name
		}
	}
	if _, set := fmap["4"]; set {
		m.Nodes = src.Nodes
	}
	if _, set := fmap["9"]; set {
		m.Liveness = src.Liveness
	}
}

func (s *ClusterInst) HasFields() bool {
	return true
}

type ClusterInstStore struct {
	kvstore objstore.KVStore
}

func NewClusterInstStore(kvstore objstore.KVStore) ClusterInstStore {
	return ClusterInstStore{kvstore: kvstore}
}

func (s *ClusterInstStore) Create(m *ClusterInst, wait func(int64)) (*Result, error) {
	err := m.Validate(ClusterInstAllFieldsMap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("ClusterInst", m.GetKey())
	val, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	rev, err := s.kvstore.Create(key, string(val))
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *ClusterInstStore) Update(m *ClusterInst, wait func(int64)) (*Result, error) {
	fmap := MakeFieldMap(m.Fields)
	err := m.Validate(fmap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("ClusterInst", m.GetKey())
	var vers int64 = 0
	curBytes, vers, err := s.kvstore.Get(key)
	if err != nil {
		return nil, err
	}
	var cur ClusterInst
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
	rev, err := s.kvstore.Update(key, string(val), vers)
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *ClusterInstStore) Put(m *ClusterInst, wait func(int64)) (*Result, error) {
	fmap := MakeFieldMap(m.Fields)
	err := m.Validate(fmap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("ClusterInst", m.GetKey())
	var val []byte
	curBytes, _, err := s.kvstore.Get(key)
	if err == nil {
		var cur ClusterInst
		err = json.Unmarshal(curBytes, &cur)
		if err != nil {
			return nil, err
		}
		cur.CopyInFields(m)
		// never save fields
		cur.Fields = nil
		val, err = json.Marshal(cur)
	} else {
		m.Fields = nil
		val, err = json.Marshal(m)
	}
	if err != nil {
		return nil, err
	}
	rev, err := s.kvstore.Put(key, string(val))
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *ClusterInstStore) Delete(m *ClusterInst, wait func(int64)) (*Result, error) {
	err := m.GetKey().Validate()
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("ClusterInst", m.GetKey())
	rev, err := s.kvstore.Delete(key)
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *ClusterInstStore) LoadOne(key string) (*ClusterInst, int64, error) {
	val, rev, err := s.kvstore.Get(key)
	if err != nil {
		return nil, 0, err
	}
	var obj ClusterInst
	err = json.Unmarshal(val, &obj)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "Failed to parse ClusterInst data", "val", string(val))
		return nil, 0, err
	}
	return &obj, rev, nil
}

// ClusterInstCache caches ClusterInst objects in memory in a hash table
// and keeps them in sync with the database.
type ClusterInstCache struct {
	Objs      map[ClusterInstKey]*ClusterInst
	Mux       util.Mutex
	List      map[ClusterInstKey]struct{}
	NotifyCb  func(obj *ClusterInstKey)
	UpdatedCb func(old *ClusterInst, new *ClusterInst)
}

func NewClusterInstCache() *ClusterInstCache {
	cache := ClusterInstCache{}
	InitClusterInstCache(&cache)
	return &cache
}

func InitClusterInstCache(cache *ClusterInstCache) {
	cache.Objs = make(map[ClusterInstKey]*ClusterInst)
}

func (c *ClusterInstCache) GetTypeString() string {
	return "ClusterInst"
}

func (c *ClusterInstCache) Get(key *ClusterInstKey, valbuf *ClusterInst) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	inst, found := c.Objs[*key]
	if found {
		*valbuf = *inst
	}
	return found
}

func (c *ClusterInstCache) HasKey(key *ClusterInstKey) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	_, found := c.Objs[*key]
	return found
}

func (c *ClusterInstCache) GetAllKeys(keys map[ClusterInstKey]struct{}) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for key, _ := range c.Objs {
		keys[key] = struct{}{}
	}
}

func (c *ClusterInstCache) Update(in *ClusterInst, rev int64) {
	c.Mux.Lock()
	if c.UpdatedCb != nil {
		old := c.Objs[in.Key]
		new := &ClusterInst{}
		*new = *in
		defer c.UpdatedCb(old, new)
	}
	c.Objs[in.Key] = in
	log.DebugLog(log.DebugLevelApi, "SyncUpdate", "obj", in, "rev", rev)
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		c.NotifyCb(&in.Key)
	}
}

func (c *ClusterInstCache) Delete(in *ClusterInst, rev int64) {
	c.Mux.Lock()
	delete(c.Objs, in.Key)
	log.DebugLog(log.DebugLevelApi, "SyncUpdate", "key", in.Key, "rev", rev)
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		c.NotifyCb(&in.Key)
	}
}

func (c *ClusterInstCache) Prune(validKeys map[ClusterInstKey]struct{}) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for key, _ := range c.Objs {
		if _, ok := validKeys[key]; !ok {
			delete(c.Objs, key)
			if c.NotifyCb != nil {
				c.NotifyCb(&key)
			}
		}
	}
}

func (c *ClusterInstCache) GetCount() int {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	return len(c.Objs)
}

func (c *ClusterInstCache) Show(filter *ClusterInst, cb func(ret *ClusterInst) error) error {
	log.DebugLog(log.DebugLevelApi, "Show ClusterInst", "count", len(c.Objs))
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for _, obj := range c.Objs {
		if !obj.Matches(filter) {
			continue
		}
		log.DebugLog(log.DebugLevelApi, "Show ClusterInst", "obj", obj)
		err := cb(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ClusterInstCache) SetNotifyCb(fn func(obj *ClusterInstKey)) {
	c.NotifyCb = fn
}

func (c *ClusterInstCache) SetUpdatedCb(fn func(old *ClusterInst, new *ClusterInst)) {
	c.UpdatedCb = fn
}
func (c *ClusterInstCache) SyncUpdate(key, val []byte, rev int64) {
	obj := ClusterInst{}
	err := json.Unmarshal(val, &obj)
	if err != nil {
		log.WarnLog("Failed to parse ClusterInst data", "val", string(val))
		return
	}
	c.Update(&obj, rev)
	c.Mux.Lock()
	if c.List != nil {
		c.List[obj.Key] = struct{}{}
	}
	c.Mux.Unlock()
}

func (c *ClusterInstCache) SyncDelete(key []byte, rev int64) {
	obj := ClusterInst{}
	keystr := objstore.DbKeyPrefixRemove(string(key))
	ClusterInstKeyStringParse(keystr, &obj.Key)
	c.Delete(&obj, rev)
}

func (c *ClusterInstCache) SyncListStart() {
	c.List = make(map[ClusterInstKey]struct{})
}

func (c *ClusterInstCache) SyncListEnd() {
	deleted := make(map[ClusterInstKey]struct{})
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
func (m *ClusterInst) GetKey() *ClusterInstKey {
	return &m.Key
}

func (m *ClusterInstKey) Size() (n int) {
	var l int
	_ = l
	l = m.ClusterKey.Size()
	n += 1 + l + sovClusterinst(uint64(l))
	l = m.CloudletKey.Size()
	n += 1 + l + sovClusterinst(uint64(l))
	return n
}

func (m *ClusterInst) Size() (n int) {
	var l int
	_ = l
	if len(m.Fields) > 0 {
		for _, s := range m.Fields {
			l = len(s)
			n += 1 + l + sovClusterinst(uint64(l))
		}
	}
	l = m.Key.Size()
	n += 1 + l + sovClusterinst(uint64(l))
	l = m.Flavor.Size()
	n += 1 + l + sovClusterinst(uint64(l))
	if m.Nodes != 0 {
		n += 1 + sovClusterinst(uint64(m.Nodes))
	}
	if m.Liveness != 0 {
		n += 1 + sovClusterinst(uint64(m.Liveness))
	}
	return n
}

func sovClusterinst(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozClusterinst(x uint64) (n int) {
	return sovClusterinst(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ClusterInstKey) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowClusterinst
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
			return fmt.Errorf("proto: ClusterInstKey: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ClusterInstKey: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClusterKey", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowClusterinst
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
				return ErrInvalidLengthClusterinst
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ClusterKey.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
					return ErrIntOverflowClusterinst
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
				return ErrInvalidLengthClusterinst
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.CloudletKey.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipClusterinst(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthClusterinst
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
func (m *ClusterInst) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowClusterinst
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
			return fmt.Errorf("proto: ClusterInst: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ClusterInst: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Fields", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowClusterinst
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
				return ErrInvalidLengthClusterinst
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
					return ErrIntOverflowClusterinst
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
				return ErrInvalidLengthClusterinst
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
				return fmt.Errorf("proto: wrong wireType = %d for field Flavor", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowClusterinst
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
				return ErrInvalidLengthClusterinst
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Flavor.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Nodes", wireType)
			}
			m.Nodes = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowClusterinst
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Nodes |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 9:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Liveness", wireType)
			}
			m.Liveness = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowClusterinst
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Liveness |= (Liveness(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipClusterinst(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthClusterinst
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
func skipClusterinst(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowClusterinst
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
					return 0, ErrIntOverflowClusterinst
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
					return 0, ErrIntOverflowClusterinst
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
				return 0, ErrInvalidLengthClusterinst
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowClusterinst
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
				next, err := skipClusterinst(dAtA[start:])
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
	ErrInvalidLengthClusterinst = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowClusterinst   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("clusterinst.proto", fileDescriptorClusterinst) }

var fileDescriptorClusterinst = []byte{
	// 554 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x93, 0x4f, 0x6b, 0x13, 0x4f,
	0x18, 0xc7, 0x3b, 0x49, 0x13, 0x7e, 0x99, 0xe4, 0x17, 0xcd, 0xd6, 0x86, 0x69, 0x2c, 0x69, 0xd8,
	0x53, 0x94, 0x26, 0xab, 0x11, 0x54, 0x82, 0x20, 0xa6, 0xa2, 0x48, 0xc5, 0x43, 0xc4, 0xab, 0xb2,
	0xd9, 0x7d, 0xb2, 0x59, 0x9c, 0xdd, 0x09, 0x3b, 0xbb, 0xad, 0xb9, 0x89, 0xde, 0xbc, 0x49, 0x2f,
	0x5e, 0x04, 0x5f, 0x82, 0xf8, 0x2a, 0x72, 0x14, 0xbc, 0x8b, 0x06, 0x0f, 0x3d, 0x0a, 0xe9, 0x41,
	0x6f, 0xb2, 0xb3, 0xb3, 0x9b, 0x4d, 0xad, 0xd0, 0x43, 0x2f, 0xe1, 0xf9, 0xf3, 0xfd, 0x7e, 0x9e,
	0x67, 0x9e, 0x24, 0xb8, 0x62, 0xd0, 0x80, 0xfb, 0xe0, 0xd9, 0x2e, 0xf7, 0xdb, 0x63, 0x8f, 0xf9,
	0x4c, 0x29, 0x80, 0x69, 0x81, 0x08, 0x6b, 0x9b, 0x16, 0x63, 0x16, 0x05, 0x4d, 0x1f, 0xdb, 0x9a,
	0xee, 0xba, 0xcc, 0xd7, 0x7d, 0x9b, 0xb9, 0x3c, 0x12, 0xd6, 0x6e, 0x5a, 0xb6, 0x3f, 0x0a, 0x06,
	0x6d, 0x83, 0x39, 0x9a, 0xc3, 0x06, 0x36, 0x0d, 0x8d, 0x2f, 0xb4, 0xf0, 0xb3, 0x65, 0x50, 0x16,
	0x98, 0x9a, 0xd0, 0x59, 0xe0, 0x26, 0x81, 0x74, 0xde, 0x3f, 0x9d, 0xd3, 0x68, 0x59, 0xe0, 0xb6,
	0x0c, 0x27, 0x4e, 0x53, 0x81, 0x04, 0x95, 0x3c, 0xe0, 0x01, 0xf5, 0xe3, 0x6c, 0x48, 0xf5, 0x3d,
	0xe6, 0xc9, 0xec, 0x7f, 0xf9, 0x34, 0x99, 0x96, 0x05, 0x98, 0x42, 0x22, 0x36, 0x98, 0xe3, 0xb0,
	0x78, 0xa3, 0x56, 0x6a, 0x23, 0x8b, 0x59, 0x2c, 0x1a, 0x34, 0x08, 0x86, 0x22, 0x13, 0x89, 0x88,
	0x22, 0xb9, 0xfa, 0x1e, 0xe1, 0xf2, 0x4e, 0x84, 0x7f, 0xe0, 0x72, 0x7f, 0x17, 0x26, 0xca, 0x2d,
	0x5c, 0x94, 0x03, 0x9f, 0x3d, 0x87, 0x09, 0x41, 0x0d, 0xd4, 0x2c, 0x76, 0xd6, 0xdb, 0xc9, 0x31,
	0xdb, 0x52, 0xbf, 0x0b, 0x93, 0xde, 0xea, 0xf4, 0xeb, 0xd6, 0x4a, 0x1f, 0x1b, 0x49, 0x45, 0xb9,
	0x8d, 0x4b, 0xf1, 0x7e, 0xc2, 0x9e, 0x11, 0xf6, 0xea, 0x92, 0x3d, 0x6a, 0x2f, 0xfc, 0x45, 0x63,
	0x51, 0xea, 0x96, 0x0e, 0xe7, 0x04, 0xfd, 0x9a, 0x13, 0xf4, 0xf1, 0xc3, 0x16, 0x52, 0xdf, 0x66,
	0x70, 0x31, 0xb5, 0x9f, 0x52, 0xc5, 0xf9, 0xa1, 0x0d, 0xd4, 0xe4, 0x04, 0x35, 0xb2, 0xcd, 0x42,
	0x5f, 0x66, 0xca, 0x55, 0x9c, 0x5d, 0x4c, 0xdb, 0xf8, 0x7b, 0x59, 0xf9, 0x38, 0x39, 0x30, 0xd4,
	0x2a, 0x37, 0x70, 0x3e, 0x3a, 0x33, 0xc9, 0x0a, 0xd7, 0x85, 0x94, 0xeb, 0x9e, 0x68, 0x84, 0x86,
	0x42, 0x68, 0x38, 0x7c, 0xfd, 0x1b, 0xa1, 0xbe, 0x94, 0x2b, 0x17, 0x71, 0xce, 0x65, 0x26, 0x70,
	0xb2, 0xda, 0x40, 0xcd, 0x5c, 0x2f, 0x17, 0x75, 0xa3, 0x9a, 0x72, 0x1d, 0xff, 0x47, 0xed, 0x3d,
	0x70, 0x81, 0x73, 0x52, 0x68, 0xa0, 0x66, 0xb9, 0xb3, 0x96, 0xe2, 0x3e, 0x94, 0xad, 0xd8, 0x94,
	0x68, 0xbb, 0x97, 0xc2, 0x67, 0xff, 0x9c, 0x13, 0xf4, 0xf2, 0x88, 0xa0, 0x77, 0x47, 0x04, 0xbd,
	0xf9, 0xb4, 0xb1, 0x1e, 0xad, 0xb1, 0xfd, 0x28, 0x24, 0x6f, 0xc7, 0xde, 0xce, 0x41, 0x76, 0xe9,
	0x3b, 0xbb, 0x33, 0xb6, 0x95, 0xa7, 0xb8, 0xb2, 0xe3, 0x81, 0xee, 0xc3, 0xd2, 0xad, 0x4e, 0x3e,
	0x43, 0xad, 0x92, 0xaa, 0xf7, 0xc5, 0xcf, 0x4e, 0xad, 0xbf, 0xfa, 0xf2, 0xe3, 0x20, 0x43, 0xd4,
	0x35, 0xcd, 0x10, 0x18, 0x2d, 0xf5, 0x6f, 0xea, 0xa2, 0xcb, 0x21, 0xff, 0x2e, 0x50, 0x38, 0x03,
	0xbe, 0x29, 0x30, 0x27, 0xf0, 0x9f, 0x8c, 0xcd, 0xb3, 0xd8, 0x3f, 0x10, 0x98, 0xe3, 0x7c, 0x1d,
	0x9f, 0x7b, 0x3c, 0x62, 0xfb, 0xa7, 0xa1, 0xff, 0xa3, 0xae, 0x6e, 0x8a, 0x11, 0x55, 0xb5, 0xa2,
	0xf1, 0x11, 0xdb, 0x3f, 0x36, 0xe0, 0x0a, 0xea, 0x9d, 0x9f, 0x7e, 0xaf, 0xaf, 0x4c, 0x67, 0x75,
	0xf4, 0x79, 0x56, 0x47, 0xdf, 0x66, 0x75, 0x34, 0xc8, 0x0b, 0xc8, 0xb5, 0x3f, 0x01, 0x00, 0x00,
	0xff, 0xff, 0x6b, 0xd0, 0x16, 0x90, 0x9b, 0x04, 0x00, 0x00,
}
