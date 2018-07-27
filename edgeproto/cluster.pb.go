// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cluster.proto

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
import "github.com/coreos/etcd/clientv3/concurrency"
import "github.com/mobiledgex/edge-cloud/util"
import "github.com/mobiledgex/edge-cloud/log"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type ClusterKey struct {
	// cluster name
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (m *ClusterKey) Reset()                    { *m = ClusterKey{} }
func (m *ClusterKey) String() string            { return proto.CompactTextString(m) }
func (*ClusterKey) ProtoMessage()               {}
func (*ClusterKey) Descriptor() ([]byte, []int) { return fileDescriptorCluster, []int{0} }

// Developer creates a cluster so that apps can be assigned to it.
// Clusters are not location specific. They may be instantiated on
// 0 or more Cloudlets. When the controller (or the user) decides
// create an AppInst (create an App on a Cloudlet), it sends the
// Cluster to the CRM to create the cluster.
type Cluster struct {
	Fields []string `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	// Unique key
	Key ClusterKey `protobuf:"bytes,2,opt,name=key" json:"key"`
	// flavor of each node the cluster
	Flavor FlavorKey `protobuf:"bytes,3,opt,name=flavor" json:"flavor"`
	// number of nodes in the cluster
	Nodes int32 `protobuf:"varint,4,opt,name=nodes,proto3" json:"nodes,omitempty"`
	// auto set to true when automatically created by back-end
	Auto bool `protobuf:"varint,5,opt,name=auto,proto3" json:"auto,omitempty"`
}

func (m *Cluster) Reset()                    { *m = Cluster{} }
func (m *Cluster) String() string            { return proto.CompactTextString(m) }
func (*Cluster) ProtoMessage()               {}
func (*Cluster) Descriptor() ([]byte, []int) { return fileDescriptorCluster, []int{1} }

func init() {
	proto.RegisterType((*ClusterKey)(nil), "edgeproto.ClusterKey")
	proto.RegisterType((*Cluster)(nil), "edgeproto.Cluster")
}
func (this *ClusterKey) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 5)
	s = append(s, "&edgeproto.ClusterKey{")
	s = append(s, "Name: "+fmt.Sprintf("%#v", this.Name)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringCluster(v interface{}, typ string) string {
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

// Client API for ClusterApi service

type ClusterApiClient interface {
	CreateCluster(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (*Result, error)
	DeleteCluster(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (*Result, error)
	UpdateCluster(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (*Result, error)
	ShowCluster(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (ClusterApi_ShowClusterClient, error)
}

type clusterApiClient struct {
	cc *grpc.ClientConn
}

func NewClusterApiClient(cc *grpc.ClientConn) ClusterApiClient {
	return &clusterApiClient{cc}
}

func (c *clusterApiClient) CreateCluster(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.ClusterApi/CreateCluster", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterApiClient) DeleteCluster(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.ClusterApi/DeleteCluster", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterApiClient) UpdateCluster(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.ClusterApi/UpdateCluster", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterApiClient) ShowCluster(ctx context.Context, in *Cluster, opts ...grpc.CallOption) (ClusterApi_ShowClusterClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_ClusterApi_serviceDesc.Streams[0], c.cc, "/edgeproto.ClusterApi/ShowCluster", opts...)
	if err != nil {
		return nil, err
	}
	x := &clusterApiShowClusterClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ClusterApi_ShowClusterClient interface {
	Recv() (*Cluster, error)
	grpc.ClientStream
}

type clusterApiShowClusterClient struct {
	grpc.ClientStream
}

func (x *clusterApiShowClusterClient) Recv() (*Cluster, error) {
	m := new(Cluster)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for ClusterApi service

type ClusterApiServer interface {
	CreateCluster(context.Context, *Cluster) (*Result, error)
	DeleteCluster(context.Context, *Cluster) (*Result, error)
	UpdateCluster(context.Context, *Cluster) (*Result, error)
	ShowCluster(*Cluster, ClusterApi_ShowClusterServer) error
}

func RegisterClusterApiServer(s *grpc.Server, srv ClusterApiServer) {
	s.RegisterService(&_ClusterApi_serviceDesc, srv)
}

func _ClusterApi_CreateCluster_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Cluster)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterApiServer).CreateCluster(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.ClusterApi/CreateCluster",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterApiServer).CreateCluster(ctx, req.(*Cluster))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClusterApi_DeleteCluster_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Cluster)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterApiServer).DeleteCluster(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.ClusterApi/DeleteCluster",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterApiServer).DeleteCluster(ctx, req.(*Cluster))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClusterApi_UpdateCluster_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Cluster)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterApiServer).UpdateCluster(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.ClusterApi/UpdateCluster",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterApiServer).UpdateCluster(ctx, req.(*Cluster))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClusterApi_ShowCluster_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Cluster)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ClusterApiServer).ShowCluster(m, &clusterApiShowClusterServer{stream})
}

type ClusterApi_ShowClusterServer interface {
	Send(*Cluster) error
	grpc.ServerStream
}

type clusterApiShowClusterServer struct {
	grpc.ServerStream
}

func (x *clusterApiShowClusterServer) Send(m *Cluster) error {
	return x.ServerStream.SendMsg(m)
}

var _ClusterApi_serviceDesc = grpc.ServiceDesc{
	ServiceName: "edgeproto.ClusterApi",
	HandlerType: (*ClusterApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateCluster",
			Handler:    _ClusterApi_CreateCluster_Handler,
		},
		{
			MethodName: "DeleteCluster",
			Handler:    _ClusterApi_DeleteCluster_Handler,
		},
		{
			MethodName: "UpdateCluster",
			Handler:    _ClusterApi_UpdateCluster_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ShowCluster",
			Handler:       _ClusterApi_ShowCluster_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "cluster.proto",
}

func (m *ClusterKey) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ClusterKey) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintCluster(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	return i, nil
}

func (m *Cluster) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Cluster) MarshalTo(dAtA []byte) (int, error) {
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
	i = encodeVarintCluster(dAtA, i, uint64(m.Key.Size()))
	n1, err := m.Key.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	dAtA[i] = 0x1a
	i++
	i = encodeVarintCluster(dAtA, i, uint64(m.Flavor.Size()))
	n2, err := m.Flavor.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n2
	if m.Nodes != 0 {
		dAtA[i] = 0x20
		i++
		i = encodeVarintCluster(dAtA, i, uint64(m.Nodes))
	}
	if m.Auto {
		dAtA[i] = 0x28
		i++
		if m.Auto {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	return i, nil
}

func encodeVarintCluster(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *ClusterKey) Matches(filter *ClusterKey) bool {
	if filter == nil {
		return true
	}
	if filter.Name != "" && filter.Name != m.Name {
		return false
	}
	return true
}

func (m *ClusterKey) MatchesIgnoreBackend(filter *ClusterKey) bool {
	if filter == nil {
		return true
	}
	if filter.Name != "" && filter.Name != m.Name {
		return false
	}
	return true
}

func (m *ClusterKey) CopyInFields(src *ClusterKey) {
	m.Name = src.Name
}

func (m *ClusterKey) GetKeyString() string {
	key, err := json.Marshal(m)
	if err != nil {
		log.FatalLog("Failed to marshal ClusterKey key string", "obj", m)
	}
	return string(key)
}

func ClusterKeyStringParse(str string, key *ClusterKey) {
	err := json.Unmarshal([]byte(str), key)
	if err != nil {
		log.FatalLog("Failed to unmarshal ClusterKey key string", "str", str)
	}
}

func (m *Cluster) Matches(filter *Cluster) bool {
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
	if filter.Auto != false && filter.Auto != m.Auto {
		return false
	}
	return true
}

func (m *Cluster) MatchesIgnoreBackend(filter *Cluster) bool {
	if filter == nil {
		return true
	}
	if !m.Key.MatchesIgnoreBackend(&filter.Key) {
		return false
	}
	if !m.Flavor.MatchesIgnoreBackend(&filter.Flavor) {
		return false
	}
	if filter.Nodes != 0 && filter.Nodes != m.Nodes {
		return false
	}
	return true
}

const ClusterFieldKey = "2"
const ClusterFieldKeyName = "2.1"
const ClusterFieldFlavor = "3"
const ClusterFieldFlavorName = "3.1"
const ClusterFieldNodes = "4"
const ClusterFieldAuto = "5"

var ClusterAllFields = []string{
	ClusterFieldKeyName,
	ClusterFieldFlavorName,
	ClusterFieldNodes,
	ClusterFieldAuto,
}

var ClusterAllFieldsMap = map[string]struct{}{
	ClusterFieldKeyName:    struct{}{},
	ClusterFieldFlavorName: struct{}{},
	ClusterFieldNodes:      struct{}{},
	ClusterFieldAuto:       struct{}{},
}

func (m *Cluster) DiffFields(o *Cluster, fields map[string]struct{}) {
	if m.Key.Name != o.Key.Name {
		fields[ClusterFieldKeyName] = struct{}{}
		fields[ClusterFieldKey] = struct{}{}
	}
	if m.Flavor.Name != o.Flavor.Name {
		fields[ClusterFieldFlavorName] = struct{}{}
		fields[ClusterFieldFlavor] = struct{}{}
	}
	if m.Nodes != o.Nodes {
		fields[ClusterFieldNodes] = struct{}{}
	}
	if m.Auto != o.Auto {
		fields[ClusterFieldAuto] = struct{}{}
	}
}

func (m *Cluster) CopyInFields(src *Cluster) {
	fmap := MakeFieldMap(src.Fields)
	if _, set := fmap["2"]; set {
		if _, set := fmap["2.1"]; set {
			m.Key.Name = src.Key.Name
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
	if _, set := fmap["5"]; set {
		m.Auto = src.Auto
	}
}

func (s *Cluster) HasFields() bool {
	return true
}

type ClusterStore struct {
	kvstore objstore.KVStore
}

func NewClusterStore(kvstore objstore.KVStore) ClusterStore {
	return ClusterStore{kvstore: kvstore}
}

func (s *ClusterStore) Create(m *Cluster, wait func(int64)) (*Result, error) {
	err := m.Validate(ClusterAllFieldsMap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Cluster", m.GetKey())
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

func (s *ClusterStore) Update(m *Cluster, wait func(int64)) (*Result, error) {
	fmap := MakeFieldMap(m.Fields)
	err := m.Validate(fmap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Cluster", m.GetKey())
	var vers int64 = 0
	curBytes, vers, _, err := s.kvstore.Get(key)
	if err != nil {
		return nil, err
	}
	var cur Cluster
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

func (s *ClusterStore) Put(m *Cluster, wait func(int64)) (*Result, error) {
	fmap := MakeFieldMap(m.Fields)
	err := m.Validate(fmap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Cluster", m.GetKey())
	var val []byte
	curBytes, _, _, err := s.kvstore.Get(key)
	if err == nil {
		var cur Cluster
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

func (s *ClusterStore) Delete(m *Cluster, wait func(int64)) (*Result, error) {
	err := m.GetKey().Validate()
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Cluster", m.GetKey())
	rev, err := s.kvstore.Delete(key)
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *ClusterStore) LoadOne(key string) (*Cluster, int64, error) {
	val, rev, _, err := s.kvstore.Get(key)
	if err != nil {
		return nil, 0, err
	}
	var obj Cluster
	err = json.Unmarshal(val, &obj)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "Failed to parse Cluster data", "val", string(val))
		return nil, 0, err
	}
	return &obj, rev, nil
}

func (s *ClusterStore) STMGet(stm concurrency.STM, key *ClusterKey, buf *Cluster) bool {
	keystr := objstore.DbKeyString("Cluster", key)
	valstr := stm.Get(keystr)
	if valstr == "" {
		return false
	}
	if buf != nil {
		err := json.Unmarshal([]byte(valstr), buf)
		if err != nil {
			return false
		}
	}
	return true
}

func (s *ClusterStore) STMPut(stm concurrency.STM, obj *Cluster) {
	keystr := objstore.DbKeyString("Cluster", obj.GetKey())
	val, _ := json.Marshal(obj)
	stm.Put(keystr, string(val))
}

func (s *ClusterStore) STMDel(stm concurrency.STM, key *ClusterKey) {
	keystr := objstore.DbKeyString("Cluster", key)
	stm.Del(keystr)
}

// ClusterCache caches Cluster objects in memory in a hash table
// and keeps them in sync with the database.
type ClusterCache struct {
	Objs      map[ClusterKey]*Cluster
	Mux       util.Mutex
	List      map[ClusterKey]struct{}
	NotifyCb  func(obj *ClusterKey)
	UpdatedCb func(old *Cluster, new *Cluster)
}

func NewClusterCache() *ClusterCache {
	cache := ClusterCache{}
	InitClusterCache(&cache)
	return &cache
}

func InitClusterCache(cache *ClusterCache) {
	cache.Objs = make(map[ClusterKey]*Cluster)
}

func (c *ClusterCache) GetTypeString() string {
	return "Cluster"
}

func (c *ClusterCache) Get(key *ClusterKey, valbuf *Cluster) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	inst, found := c.Objs[*key]
	if found {
		*valbuf = *inst
	}
	return found
}

func (c *ClusterCache) HasKey(key *ClusterKey) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	_, found := c.Objs[*key]
	return found
}

func (c *ClusterCache) GetAllKeys(keys map[ClusterKey]struct{}) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for key, _ := range c.Objs {
		keys[key] = struct{}{}
	}
}

func (c *ClusterCache) Update(in *Cluster, rev int64) {
	c.Mux.Lock()
	if c.UpdatedCb != nil {
		old := c.Objs[in.Key]
		new := &Cluster{}
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

func (c *ClusterCache) Delete(in *Cluster, rev int64) {
	c.Mux.Lock()
	delete(c.Objs, in.Key)
	log.DebugLog(log.DebugLevelApi, "SyncUpdate", "key", in.Key, "rev", rev)
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		c.NotifyCb(&in.Key)
	}
}

func (c *ClusterCache) Prune(validKeys map[ClusterKey]struct{}) {
	notify := make(map[ClusterKey]struct{})
	c.Mux.Lock()
	for key, _ := range c.Objs {
		if _, ok := validKeys[key]; !ok {
			delete(c.Objs, key)
			if c.NotifyCb != nil {
				notify[key] = struct{}{}
			}
		}
	}
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		for key, _ := range notify {
			c.NotifyCb(&key)
		}
	}
}

func (c *ClusterCache) GetCount() int {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	return len(c.Objs)
}

func (c *ClusterCache) Show(filter *Cluster, cb func(ret *Cluster) error) error {
	log.DebugLog(log.DebugLevelApi, "Show Cluster", "count", len(c.Objs))
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for _, obj := range c.Objs {
		if !obj.Matches(filter) {
			continue
		}
		log.DebugLog(log.DebugLevelApi, "Show Cluster", "obj", obj)
		err := cb(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ClusterCache) SetNotifyCb(fn func(obj *ClusterKey)) {
	c.NotifyCb = fn
}

func (c *ClusterCache) SetUpdatedCb(fn func(old *Cluster, new *Cluster)) {
	c.UpdatedCb = fn
}
func (c *ClusterCache) SyncUpdate(key, val []byte, rev int64) {
	obj := Cluster{}
	err := json.Unmarshal(val, &obj)
	if err != nil {
		log.WarnLog("Failed to parse Cluster data", "val", string(val))
		return
	}
	c.Update(&obj, rev)
	c.Mux.Lock()
	if c.List != nil {
		c.List[obj.Key] = struct{}{}
	}
	c.Mux.Unlock()
}

func (c *ClusterCache) SyncDelete(key []byte, rev int64) {
	obj := Cluster{}
	keystr := objstore.DbKeyPrefixRemove(string(key))
	ClusterKeyStringParse(keystr, &obj.Key)
	c.Delete(&obj, rev)
}

func (c *ClusterCache) SyncListStart() {
	c.List = make(map[ClusterKey]struct{})
}

func (c *ClusterCache) SyncListEnd() {
	deleted := make(map[ClusterKey]struct{})
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
func (m *Cluster) GetKey() *ClusterKey {
	return &m.Key
}

func (m *ClusterKey) Size() (n int) {
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovCluster(uint64(l))
	}
	return n
}

func (m *Cluster) Size() (n int) {
	var l int
	_ = l
	if len(m.Fields) > 0 {
		for _, s := range m.Fields {
			l = len(s)
			n += 1 + l + sovCluster(uint64(l))
		}
	}
	l = m.Key.Size()
	n += 1 + l + sovCluster(uint64(l))
	l = m.Flavor.Size()
	n += 1 + l + sovCluster(uint64(l))
	if m.Nodes != 0 {
		n += 1 + sovCluster(uint64(m.Nodes))
	}
	if m.Auto {
		n += 2
	}
	return n
}

func sovCluster(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozCluster(x uint64) (n int) {
	return sovCluster(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ClusterKey) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCluster
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
			return fmt.Errorf("proto: ClusterKey: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ClusterKey: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCluster
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
				return ErrInvalidLengthCluster
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCluster(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCluster
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
func (m *Cluster) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCluster
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
			return fmt.Errorf("proto: Cluster: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Cluster: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Fields", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCluster
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
				return ErrInvalidLengthCluster
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
					return ErrIntOverflowCluster
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
				return ErrInvalidLengthCluster
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
					return ErrIntOverflowCluster
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
				return ErrInvalidLengthCluster
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
					return ErrIntOverflowCluster
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
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Auto", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCluster
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
			m.Auto = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipCluster(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCluster
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
func skipCluster(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCluster
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
					return 0, ErrIntOverflowCluster
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
					return 0, ErrIntOverflowCluster
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
				return 0, ErrInvalidLengthCluster
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowCluster
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
				next, err := skipCluster(dAtA[start:])
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
	ErrInvalidLengthCluster = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCluster   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("cluster.proto", fileDescriptorCluster) }

var fileDescriptorCluster = []byte{
	// 478 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x92, 0xb1, 0xab, 0xd3, 0x5e,
	0x14, 0xc7, 0xdf, 0x69, 0xd3, 0xfe, 0x7e, 0xbd, 0xaf, 0xe5, 0xe9, 0xa5, 0xca, 0x35, 0x48, 0x5f,
	0xc8, 0x14, 0x84, 0x24, 0x52, 0x17, 0xe9, 0xf6, 0xfa, 0x44, 0x07, 0x27, 0x23, 0xfe, 0x01, 0x69,
	0x72, 0x9a, 0x06, 0xd3, 0x9c, 0xd0, 0x24, 0xef, 0xf9, 0x36, 0x71, 0x74, 0x75, 0x71, 0x74, 0x74,
	0x14, 0xff, 0x8a, 0x4e, 0x22, 0x38, 0x0a, 0xa2, 0xc5, 0xc1, 0x51, 0x68, 0x07, 0x47, 0xc9, 0x4d,
	0xda, 0x17, 0x54, 0x44, 0x78, 0x4b, 0x38, 0xdf, 0x93, 0xef, 0xe7, 0x7b, 0xef, 0x39, 0x09, 0xeb,
	0x79, 0x51, 0x9e, 0x66, 0xb8, 0xb0, 0x92, 0x05, 0x65, 0xc4, 0x3b, 0xe8, 0x07, 0x28, 0x4b, 0xf5,
	0x7a, 0x40, 0x14, 0x44, 0x68, 0xbb, 0x49, 0x68, 0xbb, 0x71, 0x4c, 0x99, 0x9b, 0x85, 0x14, 0xa7,
	0xa5, 0x51, 0xbd, 0x1d, 0x84, 0xd9, 0x2c, 0x9f, 0x58, 0x1e, 0xcd, 0xed, 0x39, 0x4d, 0xc2, 0xa8,
	0x00, 0x9f, 0xd8, 0xc5, 0xd3, 0xf4, 0x22, 0xca, 0x7d, 0x5b, 0xfa, 0x02, 0x8c, 0x77, 0x45, 0x45,
	0xde, 0xfb, 0x37, 0xd2, 0x33, 0x03, 0x8c, 0x4d, 0x6f, 0xbe, 0x95, 0xb5, 0xa2, 0x0a, 0xea, 0x2e,
	0x30, 0xcd, 0xa3, 0xac, 0x52, 0x07, 0x3e, 0x9e, 0x60, 0x44, 0xc9, 0x76, 0x14, 0xb5, 0x3b, 0x8d,
	0xdc, 0x13, 0xda, 0x2a, 0xb3, 0x76, 0x6a, 0x40, 0x01, 0x95, 0x61, 0x93, 0x7c, 0x2a, 0x95, 0x14,
	0xb2, 0x2a, 0xed, 0xba, 0xc5, 0xd8, 0x71, 0xb9, 0x98, 0xfb, 0x78, 0xc6, 0x39, 0x53, 0x62, 0x77,
	0x8e, 0x02, 0x34, 0x30, 0x3a, 0x8e, 0xac, 0x47, 0xdd, 0x6f, 0x6b, 0x01, 0x3f, 0xd6, 0x02, 0xde,
	0xbc, 0x3a, 0x04, 0xfd, 0x1d, 0xb0, 0xff, 0x2a, 0x80, 0x5f, 0x65, 0xed, 0x69, 0x88, 0x91, 0x9f,
	0x0a, 0xd0, 0x9a, 0x46, 0xc7, 0xa9, 0x14, 0x37, 0x59, 0xf3, 0x31, 0x9e, 0x89, 0x86, 0x06, 0xc6,
	0xfe, 0xf0, 0x8a, 0xb5, 0xdb, 0xb4, 0x75, 0x7e, 0xd2, 0x58, 0x59, 0x7e, 0x3a, 0xdc, 0x73, 0x0a,
	0x1f, 0x1f, 0xb2, 0x76, 0x39, 0x81, 0x68, 0x4a, 0xa2, 0x5f, 0x23, 0xee, 0xca, 0x17, 0xe7, 0x40,
	0xe5, 0xe4, 0x7d, 0xd6, 0x8a, 0xc9, 0xc7, 0x54, 0x28, 0x1a, 0x18, 0x2d, 0xa7, 0x14, 0x5c, 0x30,
	0xc5, 0xcd, 0x33, 0x12, 0x2d, 0x0d, 0x8c, 0xff, 0xc7, 0xca, 0xeb, 0x8d, 0x00, 0x47, 0x76, 0x46,
	0xa2, 0x18, 0xe2, 0xfb, 0x5a, 0xc0, 0xd3, 0x8d, 0x80, 0x97, 0x1b, 0x01, 0xcf, 0xdf, 0x5e, 0x53,
	0x8e, 0xf2, 0x8c, 0x86, 0x1f, 0x1b, 0xbb, 0x0d, 0x1c, 0x25, 0x21, 0x77, 0x58, 0xef, 0x78, 0x81,
	0x6e, 0x86, 0xdb, 0x21, 0xf9, 0xef, 0xf7, 0x57, 0x2f, 0xd7, 0x7a, 0x8e, 0xfc, 0x36, 0xba, 0xfa,
	0xec, 0xc3, 0xd7, 0x17, 0x8d, 0xbe, 0x7e, 0x60, 0x7b, 0x12, 0xb7, 0xab, 0xdf, 0x6d, 0x04, 0x37,
	0x8a, 0xcc, 0x3b, 0x18, 0xe1, 0x05, 0x32, 0x7d, 0x89, 0xff, 0x92, 0xf9, 0x28, 0xf1, 0x2f, 0x72,
	0xcf, 0x5c, 0xe2, 0xf5, 0xcc, 0x07, 0x6c, 0xff, 0xe1, 0x8c, 0x4e, 0xff, 0x96, 0xf8, 0x87, 0x9e,
	0x2e, 0x64, 0x24, 0xd7, 0x7b, 0x76, 0x3a, 0xa3, 0xd3, 0x5a, 0xe0, 0x4d, 0x18, 0x5f, 0x5a, 0x7e,
	0x19, 0xec, 0x2d, 0x57, 0x03, 0x78, 0xbf, 0x1a, 0xc0, 0xe7, 0xd5, 0x00, 0x26, 0x6d, 0x09, 0xdf,
	0xfa, 0x19, 0x00, 0x00, 0xff, 0xff, 0xc2, 0x7b, 0x67, 0xe2, 0x90, 0x03, 0x00, 0x00,
}
