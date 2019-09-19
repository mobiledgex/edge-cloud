// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: flavor.proto

package edgeproto

import proto "github.com/gogo/protobuf/proto"
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
import "github.com/mobiledgex/edge-cloud/objstore"
import "github.com/coreos/etcd/clientv3/concurrency"
import "github.com/mobiledgex/edge-cloud/util"
import "github.com/mobiledgex/edge-cloud/log"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// FlavorKey uniquely identifies a Flavor.
type FlavorKey struct {
	// Flavor name
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (m *FlavorKey) Reset()                    { *m = FlavorKey{} }
func (m *FlavorKey) String() string            { return proto.CompactTextString(m) }
func (*FlavorKey) ProtoMessage()               {}
func (*FlavorKey) Descriptor() ([]byte, []int) { return fileDescriptorFlavor, []int{0} }

type Flavor struct {
	// Fields are used for the Update API to specify which fields to apply
	Fields []string `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	// Unique key for the new flavor.
	Key FlavorKey `protobuf:"bytes,2,opt,name=key" json:"key"`
	// RAM in megabytes
	Ram uint64 `protobuf:"varint,3,opt,name=ram,proto3" json:"ram,omitempty"`
	// Number of virtual CPUs
	Vcpus uint64 `protobuf:"varint,4,opt,name=vcpus,proto3" json:"vcpus,omitempty"`
	// Amount of disk space in gigabytes
	Disk uint64 `protobuf:"varint,5,opt,name=disk,proto3" json:"disk,omitempty"`
}

func (m *Flavor) Reset()                    { *m = Flavor{} }
func (m *Flavor) String() string            { return proto.CompactTextString(m) }
func (*Flavor) ProtoMessage()               {}
func (*Flavor) Descriptor() ([]byte, []int) { return fileDescriptorFlavor, []int{1} }

func init() {
	proto.RegisterType((*FlavorKey)(nil), "edgeproto.FlavorKey")
	proto.RegisterType((*Flavor)(nil), "edgeproto.Flavor")
}
func (this *FlavorKey) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 5)
	s = append(s, "&edgeproto.FlavorKey{")
	s = append(s, "Name: "+fmt.Sprintf("%#v", this.Name)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringFlavor(v interface{}, typ string) string {
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

// Client API for FlavorApi service

type FlavorApiClient interface {
	// Create a Flavor
	CreateFlavor(ctx context.Context, in *Flavor, opts ...grpc.CallOption) (*Result, error)
	// Delete a Flavor
	DeleteFlavor(ctx context.Context, in *Flavor, opts ...grpc.CallOption) (*Result, error)
	// Update a Flavor
	UpdateFlavor(ctx context.Context, in *Flavor, opts ...grpc.CallOption) (*Result, error)
	// Show Flavors
	ShowFlavor(ctx context.Context, in *Flavor, opts ...grpc.CallOption) (FlavorApi_ShowFlavorClient, error)
}

type flavorApiClient struct {
	cc *grpc.ClientConn
}

func NewFlavorApiClient(cc *grpc.ClientConn) FlavorApiClient {
	return &flavorApiClient{cc}
}

func (c *flavorApiClient) CreateFlavor(ctx context.Context, in *Flavor, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.FlavorApi/CreateFlavor", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *flavorApiClient) DeleteFlavor(ctx context.Context, in *Flavor, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.FlavorApi/DeleteFlavor", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *flavorApiClient) UpdateFlavor(ctx context.Context, in *Flavor, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.FlavorApi/UpdateFlavor", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *flavorApiClient) ShowFlavor(ctx context.Context, in *Flavor, opts ...grpc.CallOption) (FlavorApi_ShowFlavorClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_FlavorApi_serviceDesc.Streams[0], c.cc, "/edgeproto.FlavorApi/ShowFlavor", opts...)
	if err != nil {
		return nil, err
	}
	x := &flavorApiShowFlavorClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type FlavorApi_ShowFlavorClient interface {
	Recv() (*Flavor, error)
	grpc.ClientStream
}

type flavorApiShowFlavorClient struct {
	grpc.ClientStream
}

func (x *flavorApiShowFlavorClient) Recv() (*Flavor, error) {
	m := new(Flavor)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for FlavorApi service

type FlavorApiServer interface {
	// Create a Flavor
	CreateFlavor(context.Context, *Flavor) (*Result, error)
	// Delete a Flavor
	DeleteFlavor(context.Context, *Flavor) (*Result, error)
	// Update a Flavor
	UpdateFlavor(context.Context, *Flavor) (*Result, error)
	// Show Flavors
	ShowFlavor(*Flavor, FlavorApi_ShowFlavorServer) error
}

func RegisterFlavorApiServer(s *grpc.Server, srv FlavorApiServer) {
	s.RegisterService(&_FlavorApi_serviceDesc, srv)
}

func _FlavorApi_CreateFlavor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Flavor)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FlavorApiServer).CreateFlavor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.FlavorApi/CreateFlavor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FlavorApiServer).CreateFlavor(ctx, req.(*Flavor))
	}
	return interceptor(ctx, in, info, handler)
}

func _FlavorApi_DeleteFlavor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Flavor)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FlavorApiServer).DeleteFlavor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.FlavorApi/DeleteFlavor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FlavorApiServer).DeleteFlavor(ctx, req.(*Flavor))
	}
	return interceptor(ctx, in, info, handler)
}

func _FlavorApi_UpdateFlavor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Flavor)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FlavorApiServer).UpdateFlavor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.FlavorApi/UpdateFlavor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FlavorApiServer).UpdateFlavor(ctx, req.(*Flavor))
	}
	return interceptor(ctx, in, info, handler)
}

func _FlavorApi_ShowFlavor_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Flavor)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(FlavorApiServer).ShowFlavor(m, &flavorApiShowFlavorServer{stream})
}

type FlavorApi_ShowFlavorServer interface {
	Send(*Flavor) error
	grpc.ServerStream
}

type flavorApiShowFlavorServer struct {
	grpc.ServerStream
}

func (x *flavorApiShowFlavorServer) Send(m *Flavor) error {
	return x.ServerStream.SendMsg(m)
}

var _FlavorApi_serviceDesc = grpc.ServiceDesc{
	ServiceName: "edgeproto.FlavorApi",
	HandlerType: (*FlavorApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateFlavor",
			Handler:    _FlavorApi_CreateFlavor_Handler,
		},
		{
			MethodName: "DeleteFlavor",
			Handler:    _FlavorApi_DeleteFlavor_Handler,
		},
		{
			MethodName: "UpdateFlavor",
			Handler:    _FlavorApi_UpdateFlavor_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ShowFlavor",
			Handler:       _FlavorApi_ShowFlavor_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "flavor.proto",
}

func (m *FlavorKey) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FlavorKey) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintFlavor(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	return i, nil
}

func (m *Flavor) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Flavor) MarshalTo(dAtA []byte) (int, error) {
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
	i = encodeVarintFlavor(dAtA, i, uint64(m.Key.Size()))
	n1, err := m.Key.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	if m.Ram != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintFlavor(dAtA, i, uint64(m.Ram))
	}
	if m.Vcpus != 0 {
		dAtA[i] = 0x20
		i++
		i = encodeVarintFlavor(dAtA, i, uint64(m.Vcpus))
	}
	if m.Disk != 0 {
		dAtA[i] = 0x28
		i++
		i = encodeVarintFlavor(dAtA, i, uint64(m.Disk))
	}
	return i, nil
}

func encodeVarintFlavor(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *FlavorKey) Matches(o *FlavorKey, fopts ...MatchOpt) bool {
	opts := MatchOptions{}
	applyMatchOptions(&opts, fopts...)
	if o == nil {
		if opts.Filter {
			return true
		}
		return false
	}
	if !opts.Filter || o.Name != "" {
		if o.Name != m.Name {
			return false
		}
	}
	return true
}

func (m *FlavorKey) CopyInFields(src *FlavorKey) {
	m.Name = src.Name
}

func (m *FlavorKey) GetKeyString() string {
	key, err := json.Marshal(m)
	if err != nil {
		log.FatalLog("Failed to marshal FlavorKey key string", "obj", m)
	}
	return string(key)
}

func FlavorKeyStringParse(str string, key *FlavorKey) {
	err := json.Unmarshal([]byte(str), key)
	if err != nil {
		log.FatalLog("Failed to unmarshal FlavorKey key string", "str", str)
	}
}

// Helper method to check that enums have valid values
func (m *FlavorKey) ValidateEnums() error {
	return nil
}

func (m *Flavor) Matches(o *Flavor, fopts ...MatchOpt) bool {
	opts := MatchOptions{}
	applyMatchOptions(&opts, fopts...)
	if o == nil {
		if opts.Filter {
			return true
		}
		return false
	}
	if !m.Key.Matches(&o.Key, fopts...) {
		return false
	}
	if !opts.Filter || o.Ram != 0 {
		if o.Ram != m.Ram {
			return false
		}
	}
	if !opts.Filter || o.Vcpus != 0 {
		if o.Vcpus != m.Vcpus {
			return false
		}
	}
	if !opts.Filter || o.Disk != 0 {
		if o.Disk != m.Disk {
			return false
		}
	}
	return true
}

const FlavorFieldKey = "2"
const FlavorFieldKeyName = "2.1"
const FlavorFieldRam = "3"
const FlavorFieldVcpus = "4"
const FlavorFieldDisk = "5"

var FlavorAllFields = []string{
	FlavorFieldKeyName,
	FlavorFieldRam,
	FlavorFieldVcpus,
	FlavorFieldDisk,
}

var FlavorAllFieldsMap = map[string]struct{}{
	FlavorFieldKeyName: struct{}{},
	FlavorFieldRam:     struct{}{},
	FlavorFieldVcpus:   struct{}{},
	FlavorFieldDisk:    struct{}{},
}

var FlavorAllFieldsStringMap = map[string]string{
	FlavorFieldKeyName: "Flavor Field Key Name",
	FlavorFieldRam:     "Flavor Field Ram",
	FlavorFieldVcpus:   "Flavor Field Vcpus",
	FlavorFieldDisk:    "Flavor Field Disk",
}

func (m *Flavor) IsKeyField(s string) bool {
	return strings.HasPrefix(s, FlavorFieldKey+".")
}

func (m *Flavor) DiffFields(o *Flavor, fields map[string]struct{}) {
	if m.Key.Name != o.Key.Name {
		fields[FlavorFieldKeyName] = struct{}{}
		fields[FlavorFieldKey] = struct{}{}
	}
	if m.Ram != o.Ram {
		fields[FlavorFieldRam] = struct{}{}
	}
	if m.Vcpus != o.Vcpus {
		fields[FlavorFieldVcpus] = struct{}{}
	}
	if m.Disk != o.Disk {
		fields[FlavorFieldDisk] = struct{}{}
	}
}

func (m *Flavor) CopyInFields(src *Flavor) {
	fmap := MakeFieldMap(src.Fields)
	if _, set := fmap["2"]; set {
		if _, set := fmap["2.1"]; set {
			m.Key.Name = src.Key.Name
		}
	}
	if _, set := fmap["3"]; set {
		m.Ram = src.Ram
	}
	if _, set := fmap["4"]; set {
		m.Vcpus = src.Vcpus
	}
	if _, set := fmap["5"]; set {
		m.Disk = src.Disk
	}
}

func (s *Flavor) HasFields() bool {
	return true
}

type FlavorStore struct {
	kvstore objstore.KVStore
}

func NewFlavorStore(kvstore objstore.KVStore) FlavorStore {
	return FlavorStore{kvstore: kvstore}
}

func (s *FlavorStore) Create(ctx context.Context, m *Flavor, wait func(int64)) (*Result, error) {
	err := m.Validate(FlavorAllFieldsMap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Flavor", m.GetKey())
	val, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	rev, err := s.kvstore.Create(ctx, key, string(val))
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *FlavorStore) Update(ctx context.Context, m *Flavor, wait func(int64)) (*Result, error) {
	fmap := MakeFieldMap(m.Fields)
	err := m.Validate(fmap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Flavor", m.GetKey())
	var vers int64 = 0
	curBytes, vers, _, err := s.kvstore.Get(key)
	if err != nil {
		return nil, err
	}
	var cur Flavor
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
	rev, err := s.kvstore.Update(ctx, key, string(val), vers)
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *FlavorStore) Put(ctx context.Context, m *Flavor, wait func(int64), ops ...objstore.KVOp) (*Result, error) {
	err := m.Validate(FlavorAllFieldsMap)
	m.Fields = nil
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Flavor", m.GetKey())
	var val []byte
	val, err = json.Marshal(m)
	if err != nil {
		return nil, err
	}
	rev, err := s.kvstore.Put(ctx, key, string(val), ops...)
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *FlavorStore) Delete(ctx context.Context, m *Flavor, wait func(int64)) (*Result, error) {
	err := m.GetKey().ValidateKey()
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Flavor", m.GetKey())
	rev, err := s.kvstore.Delete(ctx, key)
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *FlavorStore) LoadOne(key string) (*Flavor, int64, error) {
	val, rev, _, err := s.kvstore.Get(key)
	if err != nil {
		return nil, 0, err
	}
	var obj Flavor
	err = json.Unmarshal(val, &obj)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "Failed to parse Flavor data", "val", string(val))
		return nil, 0, err
	}
	return &obj, rev, nil
}

func (s *FlavorStore) STMGet(stm concurrency.STM, key *FlavorKey, buf *Flavor) bool {
	keystr := objstore.DbKeyString("Flavor", key)
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

func (s *FlavorStore) STMPut(stm concurrency.STM, obj *Flavor, ops ...objstore.KVOp) {
	keystr := objstore.DbKeyString("Flavor", obj.GetKey())
	val, err := json.Marshal(obj)
	if err != nil {
		log.InfoLog("Flavor json marsahal failed", "obj", obj, "err", err)
	}
	v3opts := GetSTMOpts(ops...)
	stm.Put(keystr, string(val), v3opts...)
}

func (s *FlavorStore) STMDel(stm concurrency.STM, key *FlavorKey) {
	keystr := objstore.DbKeyString("Flavor", key)
	stm.Del(keystr)
}

func (m *Flavor) getKey() *FlavorKey {
	return &m.Key
}

func (m *Flavor) getKeyVal() FlavorKey {
	return m.Key
}

type FlavorKeyWatcher struct {
	cb func(ctx context.Context)
}

// FlavorCache caches Flavor objects in memory in a hash table
// and keeps them in sync with the database.
type FlavorCache struct {
	Objs        map[FlavorKey]*Flavor
	Mux         util.Mutex
	List        map[FlavorKey]struct{}
	NotifyCb    func(ctx context.Context, obj *FlavorKey, old *Flavor)
	UpdatedCb   func(ctx context.Context, old *Flavor, new *Flavor)
	KeyWatchers map[FlavorKey][]*FlavorKeyWatcher
}

func NewFlavorCache() *FlavorCache {
	cache := FlavorCache{}
	InitFlavorCache(&cache)
	return &cache
}

func InitFlavorCache(cache *FlavorCache) {
	cache.Objs = make(map[FlavorKey]*Flavor)
	cache.KeyWatchers = make(map[FlavorKey][]*FlavorKeyWatcher)
}

func (c *FlavorCache) GetTypeString() string {
	return "Flavor"
}

func (c *FlavorCache) Get(key *FlavorKey, valbuf *Flavor) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	inst, found := c.Objs[*key]
	if found {
		*valbuf = *inst
	}
	return found
}

func (c *FlavorCache) HasKey(key *FlavorKey) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	_, found := c.Objs[*key]
	return found
}

func (c *FlavorCache) GetAllKeys(ctx context.Context, keys map[FlavorKey]context.Context) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for key, _ := range c.Objs {
		keys[key] = ctx
	}
}

func (c *FlavorCache) Update(ctx context.Context, in *Flavor, rev int64) {
	c.UpdateModFunc(ctx, in.getKey(), rev, func(old *Flavor) (*Flavor, bool) {
		return in, true
	})
}

func (c *FlavorCache) UpdateModFunc(ctx context.Context, key *FlavorKey, rev int64, modFunc func(old *Flavor) (new *Flavor, changed bool)) {
	c.Mux.Lock()
	old := c.Objs[*key]
	new, changed := modFunc(old)
	if !changed {
		c.Mux.Unlock()
		return
	}
	if c.UpdatedCb != nil || c.NotifyCb != nil {
		if c.UpdatedCb != nil {
			newCopy := &Flavor{}
			*newCopy = *new
			defer c.UpdatedCb(ctx, old, newCopy)
		}
		if c.NotifyCb != nil {
			defer c.NotifyCb(ctx, new.getKey(), old)
		}
	}
	c.Objs[new.getKeyVal()] = new
	log.SpanLog(ctx, log.DebugLevelApi, "cache update", "new", new)
	log.DebugLog(log.DebugLevelApi, "SyncUpdate Flavor", "obj", new, "rev", rev)
	c.Mux.Unlock()
	c.TriggerKeyWatchers(ctx, new.getKey())
}

func (c *FlavorCache) Delete(ctx context.Context, in *Flavor, rev int64) {
	c.Mux.Lock()
	old := c.Objs[in.getKeyVal()]
	delete(c.Objs, in.getKeyVal())
	log.SpanLog(ctx, log.DebugLevelApi, "cache delete")
	log.DebugLog(log.DebugLevelApi, "SyncDelete Flavor", "key", in.getKey(), "rev", rev)
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		c.NotifyCb(ctx, in.getKey(), old)
	}
	c.TriggerKeyWatchers(ctx, in.getKey())
}

func (c *FlavorCache) Prune(ctx context.Context, validKeys map[FlavorKey]struct{}) {
	notify := make(map[FlavorKey]*Flavor)
	c.Mux.Lock()
	for key, _ := range c.Objs {
		if _, ok := validKeys[key]; !ok {
			if c.NotifyCb != nil {
				notify[key] = c.Objs[key]
			}
			delete(c.Objs, key)
		}
	}
	c.Mux.Unlock()
	for key, old := range notify {
		if c.NotifyCb != nil {
			c.NotifyCb(ctx, &key, old)
		}
		c.TriggerKeyWatchers(ctx, &key)
	}
}

func (c *FlavorCache) GetCount() int {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	return len(c.Objs)
}

func (c *FlavorCache) Flush(ctx context.Context, notifyId int64) {
}

func (c *FlavorCache) Show(filter *Flavor, cb func(ret *Flavor) error) error {
	log.DebugLog(log.DebugLevelApi, "Show Flavor", "count", len(c.Objs))
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for _, obj := range c.Objs {
		if !obj.Matches(filter, MatchFilter()) {
			continue
		}
		log.DebugLog(log.DebugLevelApi, "Show Flavor", "obj", obj)
		err := cb(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func FlavorGenericNotifyCb(fn func(key *FlavorKey, old *Flavor)) func(objstore.ObjKey, objstore.Obj) {
	return func(objkey objstore.ObjKey, obj objstore.Obj) {
		fn(objkey.(*FlavorKey), obj.(*Flavor))
	}
}

func (c *FlavorCache) SetNotifyCb(fn func(ctx context.Context, obj *FlavorKey, old *Flavor)) {
	c.NotifyCb = fn
}

func (c *FlavorCache) SetUpdatedCb(fn func(ctx context.Context, old *Flavor, new *Flavor)) {
	c.UpdatedCb = fn
}

func (c *FlavorCache) WatchKey(key *FlavorKey, cb func(ctx context.Context)) context.CancelFunc {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	list, ok := c.KeyWatchers[*key]
	if !ok {
		list = make([]*FlavorKeyWatcher, 0)
	}
	watcher := FlavorKeyWatcher{cb: cb}
	c.KeyWatchers[*key] = append(list, &watcher)
	log.DebugLog(log.DebugLevelApi, "Watching Flavor", "key", key)
	return func() {
		c.Mux.Lock()
		defer c.Mux.Unlock()
		list, ok := c.KeyWatchers[*key]
		if !ok {
			return
		}
		for ii, _ := range list {
			if list[ii] != &watcher {
				continue
			}
			if len(list) == 1 {
				delete(c.KeyWatchers, *key)
				return
			}
			list[ii] = list[len(list)-1]
			list[len(list)-1] = nil
			c.KeyWatchers[*key] = list[:len(list)-1]
			return
		}
	}
}

func (c *FlavorCache) TriggerKeyWatchers(ctx context.Context, key *FlavorKey) {
	watchers := make([]*FlavorKeyWatcher, 0)
	c.Mux.Lock()
	if list, ok := c.KeyWatchers[*key]; ok {
		watchers = append(watchers, list...)
	}
	c.Mux.Unlock()
	for ii, _ := range watchers {
		watchers[ii].cb(ctx)
	}
}
func (c *FlavorCache) SyncUpdate(ctx context.Context, key, val []byte, rev int64) {
	obj := Flavor{}
	err := json.Unmarshal(val, &obj)
	if err != nil {
		log.WarnLog("Failed to parse Flavor data", "val", string(val))
		return
	}
	c.Update(ctx, &obj, rev)
	c.Mux.Lock()
	if c.List != nil {
		c.List[obj.getKeyVal()] = struct{}{}
	}
	c.Mux.Unlock()
}

func (c *FlavorCache) SyncDelete(ctx context.Context, key []byte, rev int64) {
	obj := Flavor{}
	keystr := objstore.DbKeyPrefixRemove(string(key))
	FlavorKeyStringParse(keystr, obj.getKey())
	c.Delete(ctx, &obj, rev)
}

func (c *FlavorCache) SyncListStart(ctx context.Context) {
	c.List = make(map[FlavorKey]struct{})
}

func (c *FlavorCache) SyncListEnd(ctx context.Context) {
	deleted := make(map[FlavorKey]*Flavor)
	c.Mux.Lock()
	for key, val := range c.Objs {
		if _, found := c.List[key]; !found {
			deleted[key] = val
			delete(c.Objs, key)
		}
	}
	c.List = nil
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		for key, val := range deleted {
			c.NotifyCb(ctx, &key, val)
			c.TriggerKeyWatchers(ctx, &key)
		}
	}
}

func (m *Flavor) GetKey() objstore.ObjKey {
	return &m.Key
}

func CmpSortFlavor(a Flavor, b Flavor) bool {
	return a.Key.GetKeyString() < b.Key.GetKeyString()
}

// Helper method to check that enums have valid values
// NOTE: ValidateEnums checks all Fields even if some are not set
func (m *Flavor) ValidateEnums() error {
	if err := m.Key.ValidateEnums(); err != nil {
		return err
	}
	return nil
}

func (m *FlavorKey) Size() (n int) {
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovFlavor(uint64(l))
	}
	return n
}

func (m *Flavor) Size() (n int) {
	var l int
	_ = l
	if len(m.Fields) > 0 {
		for _, s := range m.Fields {
			l = len(s)
			n += 1 + l + sovFlavor(uint64(l))
		}
	}
	l = m.Key.Size()
	n += 1 + l + sovFlavor(uint64(l))
	if m.Ram != 0 {
		n += 1 + sovFlavor(uint64(m.Ram))
	}
	if m.Vcpus != 0 {
		n += 1 + sovFlavor(uint64(m.Vcpus))
	}
	if m.Disk != 0 {
		n += 1 + sovFlavor(uint64(m.Disk))
	}
	return n
}

func sovFlavor(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozFlavor(x uint64) (n int) {
	return sovFlavor(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *FlavorKey) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowFlavor
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
			return fmt.Errorf("proto: FlavorKey: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FlavorKey: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFlavor
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
				return ErrInvalidLengthFlavor
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipFlavor(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthFlavor
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
func (m *Flavor) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowFlavor
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
			return fmt.Errorf("proto: Flavor: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Flavor: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Fields", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFlavor
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
				return ErrInvalidLengthFlavor
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
					return ErrIntOverflowFlavor
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
				return ErrInvalidLengthFlavor
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
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ram", wireType)
			}
			m.Ram = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFlavor
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Ram |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Vcpus", wireType)
			}
			m.Vcpus = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFlavor
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Vcpus |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Disk", wireType)
			}
			m.Disk = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFlavor
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Disk |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipFlavor(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthFlavor
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
func skipFlavor(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowFlavor
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
					return 0, ErrIntOverflowFlavor
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
					return 0, ErrIntOverflowFlavor
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
				return 0, ErrInvalidLengthFlavor
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowFlavor
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
				next, err := skipFlavor(dAtA[start:])
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
	ErrInvalidLengthFlavor = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowFlavor   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("flavor.proto", fileDescriptorFlavor) }

var fileDescriptorFlavor = []byte{
	// 492 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x91, 0x41, 0x8b, 0xd3, 0x40,
	0x14, 0xc7, 0x77, 0xb6, 0xdd, 0x42, 0xc7, 0xb8, 0xec, 0xc6, 0x22, 0x43, 0xd1, 0x6e, 0xc9, 0x41,
	0x8a, 0x64, 0x33, 0xb2, 0x1e, 0x94, 0x82, 0x87, 0xae, 0x8b, 0x97, 0x45, 0x0f, 0x11, 0xf7, 0x3e,
	0x4d, 0x5e, 0xd3, 0xa1, 0x49, 0x26, 0x64, 0x92, 0xad, 0xbd, 0x89, 0x9f, 0x40, 0xf0, 0x22, 0x9e,
	0xfc, 0x08, 0x7e, 0x07, 0x2f, 0xc5, 0x93, 0xe0, 0xcd, 0x83, 0x68, 0xf1, 0x20, 0x7b, 0x12, 0x9a,
	0x05, 0x8f, 0x32, 0x93, 0x58, 0x57, 0xf4, 0xb0, 0x97, 0xbd, 0x84, 0xff, 0xfb, 0xf3, 0xde, 0xfb,
	0xfd, 0x5f, 0x06, 0x1b, 0xa3, 0x90, 0x1d, 0x8b, 0xd4, 0x49, 0x52, 0x91, 0x09, 0xb3, 0x09, 0x7e,
	0x00, 0x5a, 0xb6, 0xaf, 0x05, 0x42, 0x04, 0x21, 0x50, 0x96, 0x70, 0xca, 0xe2, 0x58, 0x64, 0x2c,
	0xe3, 0x22, 0x96, 0x65, 0x63, 0xfb, 0x6e, 0xc0, 0xb3, 0x71, 0x3e, 0x74, 0x3c, 0x11, 0xd1, 0x48,
	0x0c, 0x79, 0xa8, 0x06, 0x9f, 0x52, 0xf5, 0xdd, 0xf5, 0x42, 0x91, 0xfb, 0x54, 0xf7, 0x05, 0x10,
	0xaf, 0x44, 0x35, 0x69, 0xa4, 0x20, 0xf3, 0x30, 0xab, 0xaa, 0x56, 0x20, 0x02, 0xa1, 0x25, 0x55,
	0xaa, 0x74, 0xad, 0x5d, 0xdc, 0x7c, 0xa0, 0x63, 0x1d, 0xc2, 0xcc, 0x34, 0x71, 0x3d, 0x66, 0x11,
	0x10, 0xd4, 0x45, 0xbd, 0xa6, 0xab, 0x75, 0xdf, 0xf8, 0xbe, 0x24, 0xe8, 0xe7, 0x92, 0xa0, 0xb7,
	0x6f, 0x76, 0x90, 0xf5, 0x0e, 0xe1, 0x46, 0xd9, 0x6f, 0x5e, 0xc5, 0x8d, 0x11, 0x87, 0xd0, 0x97,
	0x04, 0x75, 0x6b, 0xbd, 0xa6, 0x5b, 0x55, 0xa6, 0x8d, 0x6b, 0x13, 0x98, 0x91, 0xf5, 0x2e, 0xea,
	0x5d, 0xda, 0x6b, 0x39, 0xab, 0x33, 0x9d, 0x15, 0x67, 0xbf, 0x3e, 0xff, 0xbc, 0xb3, 0xe6, 0xaa,
	0x36, 0x73, 0x0b, 0xd7, 0x52, 0x16, 0x91, 0x5a, 0x17, 0xf5, 0xea, 0xae, 0x92, 0x66, 0x0b, 0x6f,
	0x1c, 0x7b, 0x49, 0x2e, 0x49, 0x5d, 0x7b, 0x65, 0xa1, 0xa2, 0xf9, 0x5c, 0x4e, 0xc8, 0x86, 0x36,
	0xb5, 0xee, 0xdf, 0x51, 0xd1, 0x7e, 0x2c, 0x09, 0x7a, 0x56, 0x10, 0xf4, 0xa2, 0x20, 0xe8, 0x55,
	0x41, 0xd0, 0xeb, 0x53, 0x72, 0x59, 0x05, 0xbf, 0x77, 0x08, 0x33, 0xe7, 0x11, 0x8b, 0xe0, 0xfd,
	0x29, 0xd9, 0x74, 0x59, 0x64, 0x1f, 0xa9, 0x3d, 0xf6, 0x01, 0x97, 0x93, 0xbd, 0x4f, 0xb5, 0xdf,
	0x57, 0x0f, 0x12, 0x6e, 0x26, 0xd8, 0xb8, 0x9f, 0x02, 0xcb, 0xa0, 0x3a, 0x6c, 0xfb, 0x9f, 0xcc,
	0xed, 0xb3, 0x96, 0xab, 0x7f, 0xaa, 0xd5, 0x3f, 0x29, 0xc8, 0x75, 0x17, 0xa4, 0xc8, 0x53, 0xaf,
	0x9a, 0x94, 0xf6, 0xc0, 0x53, 0x2f, 0xf7, 0x90, 0xc5, 0x2c, 0x00, 0xfb, 0xf9, 0xc7, 0x6f, 0x2f,
	0xd7, 0xaf, 0x58, 0x9b, 0xd4, 0xd3, 0xdb, 0x69, 0xf9, 0xfc, 0x7d, 0x74, 0x53, 0x11, 0x0f, 0x20,
	0x84, 0x8b, 0x23, 0xfa, 0x7a, 0xfb, 0xdf, 0xc4, 0x27, 0x89, 0x7f, 0x81, 0x37, 0xe6, 0x7a, 0xfb,
	0x19, 0xe2, 0x14, 0xe3, 0xc7, 0x63, 0x31, 0x3d, 0x1f, 0xaf, 0xb4, 0xac, 0xc1, 0x49, 0x41, 0x6e,
	0xfc, 0x9f, 0x77, 0xc4, 0x61, 0x6a, 0xcb, 0x09, 0x4f, 0x20, 0x1e, 0x89, 0xd4, 0x03, 0x0d, 0xde,
	0xb6, 0x0c, 0x2a, 0xc7, 0x62, 0xfa, 0x07, 0x7b, 0x0b, 0xed, 0x6f, 0xcd, 0xbf, 0x76, 0xd6, 0xe6,
	0x8b, 0x0e, 0xfa, 0xb0, 0xe8, 0xa0, 0x2f, 0x8b, 0x0e, 0x1a, 0x36, 0x34, 0xe4, 0xf6, 0xaf, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x6b, 0x45, 0x66, 0x5e, 0x81, 0x03, 0x00, 0x00,
}
