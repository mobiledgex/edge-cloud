// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: flavor.proto

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

type FlavorKey struct {
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (m *FlavorKey) Reset()                    { *m = FlavorKey{} }
func (m *FlavorKey) String() string            { return proto.CompactTextString(m) }
func (*FlavorKey) ProtoMessage()               {}
func (*FlavorKey) Descriptor() ([]byte, []int) { return fileDescriptorFlavor, []int{0} }

type Flavor struct {
	Fields []string `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	// Unique key
	Key FlavorKey `protobuf:"bytes,2,opt,name=key" json:"key"`
	// RAM in MB
	Ram uint64 `protobuf:"varint,3,opt,name=ram,proto3" json:"ram,omitempty"`
	// VCPU cores
	Vcpus uint64 `protobuf:"varint,4,opt,name=vcpus,proto3" json:"vcpus,omitempty"`
	// amount of disk in GB
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
	CreateFlavor(ctx context.Context, in *Flavor, opts ...grpc.CallOption) (*Result, error)
	DeleteFlavor(ctx context.Context, in *Flavor, opts ...grpc.CallOption) (*Result, error)
	UpdateFlavor(ctx context.Context, in *Flavor, opts ...grpc.CallOption) (*Result, error)
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
	CreateFlavor(context.Context, *Flavor) (*Result, error)
	DeleteFlavor(context.Context, *Flavor) (*Result, error)
	UpdateFlavor(context.Context, *Flavor) (*Result, error)
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

func (s *FlavorStore) Create(m *Flavor, wait func(int64)) (*Result, error) {
	err := m.Validate(FlavorAllFieldsMap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Flavor", m.GetKey())
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

func (s *FlavorStore) Update(m *Flavor, wait func(int64)) (*Result, error) {
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
	rev, err := s.kvstore.Update(key, string(val), vers)
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *FlavorStore) Put(m *Flavor, wait func(int64)) (*Result, error) {
	fmap := MakeFieldMap(m.Fields)
	err := m.Validate(fmap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Flavor", m.GetKey())
	var val []byte
	curBytes, _, _, err := s.kvstore.Get(key)
	if err == nil {
		var cur Flavor
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

func (s *FlavorStore) Delete(m *Flavor, wait func(int64)) (*Result, error) {
	err := m.GetKey().Validate()
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Flavor", m.GetKey())
	rev, err := s.kvstore.Delete(key)
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

func (s *FlavorStore) STMPut(stm concurrency.STM, obj *Flavor) {
	keystr := objstore.DbKeyString("Flavor", obj.GetKey())
	val, _ := json.Marshal(obj)
	stm.Put(keystr, string(val))
}

func (s *FlavorStore) STMDel(stm concurrency.STM, key *FlavorKey) {
	keystr := objstore.DbKeyString("Flavor", key)
	stm.Del(keystr)
}

type FlavorKeyWatcher struct {
	cb func()
}

// FlavorCache caches Flavor objects in memory in a hash table
// and keeps them in sync with the database.
type FlavorCache struct {
	Objs        map[FlavorKey]*Flavor
	Mux         util.Mutex
	List        map[FlavorKey]struct{}
	NotifyCb    func(obj *FlavorKey)
	UpdatedCb   func(old *Flavor, new *Flavor)
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

func (c *FlavorCache) GetAllKeys(keys map[FlavorKey]struct{}) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for key, _ := range c.Objs {
		keys[key] = struct{}{}
	}
}

func (c *FlavorCache) Update(in *Flavor, rev int64) {
	c.Mux.Lock()
	if c.UpdatedCb != nil {
		old := c.Objs[in.Key]
		new := &Flavor{}
		*new = *in
		defer c.UpdatedCb(old, new)
	}
	c.Objs[in.Key] = in
	log.DebugLog(log.DebugLevelApi, "SyncUpdate Flavor", "obj", in, "rev", rev)
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		c.NotifyCb(&in.Key)
	}
	c.TriggerKeyWatchers(&in.Key)
}

func (c *FlavorCache) Delete(in *Flavor, rev int64) {
	c.Mux.Lock()
	delete(c.Objs, in.Key)
	log.DebugLog(log.DebugLevelApi, "SyncDelete Flavor", "key", in.Key, "rev", rev)
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		c.NotifyCb(&in.Key)
	}
	c.TriggerKeyWatchers(&in.Key)
}

func (c *FlavorCache) Prune(validKeys map[FlavorKey]struct{}) {
	notify := make(map[FlavorKey]struct{})
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
	for key, _ := range notify {
		if c.NotifyCb != nil {
			c.NotifyCb(&key)
		}
		c.TriggerKeyWatchers(&key)
	}
}

func (c *FlavorCache) GetCount() int {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	return len(c.Objs)
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

func (c *FlavorCache) SetNotifyCb(fn func(obj *FlavorKey)) {
	c.NotifyCb = fn
}

func (c *FlavorCache) SetUpdatedCb(fn func(old *Flavor, new *Flavor)) {
	c.UpdatedCb = fn
}

func (c *FlavorCache) WatchKey(key *FlavorKey, cb func()) context.CancelFunc {
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

func (c *FlavorCache) TriggerKeyWatchers(key *FlavorKey) {
	watchers := make([]*FlavorKeyWatcher, 0)
	c.Mux.Lock()
	if list, ok := c.KeyWatchers[*key]; ok {
		watchers = append(watchers, list...)
	}
	c.Mux.Unlock()
	for ii, _ := range watchers {
		watchers[ii].cb()
	}
}
func (c *FlavorCache) SyncUpdate(key, val []byte, rev int64) {
	obj := Flavor{}
	err := json.Unmarshal(val, &obj)
	if err != nil {
		log.WarnLog("Failed to parse Flavor data", "val", string(val))
		return
	}
	c.Update(&obj, rev)
	c.Mux.Lock()
	if c.List != nil {
		c.List[obj.Key] = struct{}{}
	}
	c.Mux.Unlock()
}

func (c *FlavorCache) SyncDelete(key []byte, rev int64) {
	obj := Flavor{}
	keystr := objstore.DbKeyPrefixRemove(string(key))
	FlavorKeyStringParse(keystr, &obj.Key)
	c.Delete(&obj, rev)
}

func (c *FlavorCache) SyncListStart() {
	c.List = make(map[FlavorKey]struct{})
}

func (c *FlavorCache) SyncListEnd() {
	deleted := make(map[FlavorKey]struct{})
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
			c.TriggerKeyWatchers(&key)
		}
	}
}

func (m *Flavor) GetKey() *FlavorKey {
	return &m.Key
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
	// 437 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x92, 0x3f, 0x8b, 0xd4, 0x40,
	0x18, 0xc6, 0xef, 0xdd, 0xdd, 0x5b, 0xc8, 0x18, 0x64, 0x6f, 0x3c, 0x74, 0x5c, 0x24, 0x17, 0x52,
	0x05, 0x31, 0x19, 0x39, 0x1b, 0xd9, 0xce, 0x53, 0xb4, 0x10, 0x11, 0x22, 0x7e, 0x80, 0xfc, 0x99,
	0x9d, 0x0d, 0x97, 0x64, 0x42, 0xfe, 0xdc, 0x79, 0x9d, 0xf8, 0x15, 0x6c, 0x2c, 0x2c, 0xfc, 0x08,
	0x7e, 0x8c, 0x6d, 0x04, 0xc1, 0x5e, 0x74, 0xb1, 0xb0, 0x14, 0xf6, 0x0a, 0x4b, 0xc9, 0x9b, 0xdc,
	0x1a, 0xb1, 0x39, 0xb6, 0x09, 0xcf, 0xf3, 0xf2, 0x3c, 0xbf, 0xbc, 0x33, 0x0c, 0xd1, 0xe7, 0x89,
	0x7f, 0xa2, 0x0a, 0x37, 0x2f, 0x54, 0xa5, 0xa8, 0x26, 0x22, 0x29, 0x50, 0x4e, 0x6f, 0x49, 0xa5,
	0x64, 0x22, 0xb8, 0x9f, 0xc7, 0xdc, 0xcf, 0x32, 0x55, 0xf9, 0x55, 0xac, 0xb2, 0xb2, 0x0d, 0x4e,
	0xef, 0xcb, 0xb8, 0x5a, 0xd4, 0x81, 0x1b, 0xaa, 0x94, 0xa7, 0x2a, 0x88, 0x93, 0xa6, 0xf8, 0x8a,
	0x37, 0x5f, 0x27, 0x4c, 0x54, 0x1d, 0x71, 0xcc, 0x49, 0x91, 0x6d, 0x44, 0xd7, 0x7c, 0x72, 0xb9,
	0x66, 0xe8, 0x48, 0x91, 0x39, 0x61, 0x7a, 0x61, 0x7b, 0xa2, 0x03, 0xe9, 0x85, 0x28, 0xeb, 0xa4,
	0xea, 0x9c, 0xd3, 0xc3, 0x4a, 0x25, 0x55, 0x9b, 0x0e, 0xea, 0x39, 0x3a, 0x34, 0xa8, 0xda, 0xb8,
	0xe5, 0x10, 0xed, 0x31, 0x1e, 0xfc, 0xa9, 0x38, 0xa3, 0x94, 0x8c, 0x32, 0x3f, 0x15, 0x0c, 0x4c,
	0xb0, 0x35, 0x0f, 0xf5, 0x4c, 0xff, 0xb9, 0x66, 0xf0, 0x7b, 0xcd, 0xe0, 0xe3, 0x87, 0x03, 0xb0,
	0xde, 0x03, 0x19, 0xb7, 0x79, 0x7a, 0x9d, 0x8c, 0xe7, 0xb1, 0x48, 0xa2, 0x92, 0x81, 0x39, 0xb4,
	0x35, 0xaf, 0x73, 0xf4, 0x0e, 0x19, 0x1e, 0x8b, 0x33, 0x36, 0x30, 0xc1, 0xbe, 0x72, 0xb8, 0xef,
	0x6e, 0x2e, 0xd2, 0xdd, 0xfc, 0xe7, 0x68, 0xb4, 0xfc, 0x7a, 0xb0, 0xe3, 0x35, 0x31, 0x3a, 0x21,
	0xc3, 0xc2, 0x4f, 0xd9, 0xd0, 0x04, 0x7b, 0xe4, 0x35, 0x92, 0xee, 0x93, 0xdd, 0x93, 0x30, 0xaf,
	0x4b, 0x36, 0xc2, 0x59, 0x6b, 0x9a, 0xd5, 0xa2, 0xb8, 0x3c, 0x66, 0xbb, 0x38, 0x44, 0x3d, 0x9b,
	0x34, 0xab, 0xfd, 0x5a, 0x33, 0x78, 0x7d, 0xce, 0xe0, 0xdd, 0x39, 0x83, 0xc3, 0x4f, 0x83, 0x8b,
	0xe3, 0x3c, 0xc8, 0x63, 0xfa, 0x9c, 0xe8, 0x0f, 0x0b, 0xe1, 0x57, 0xa2, 0xdb, 0x78, 0xef, 0xbf,
	0x65, 0xa6, 0xfd, 0x91, 0x87, 0xd7, 0x68, 0xdd, 0x7c, 0xf3, 0xe5, 0xc7, 0xdb, 0xc1, 0x35, 0xeb,
	0x2a, 0x0f, 0xb1, 0xcc, 0xdb, 0x87, 0x31, 0x83, 0xdb, 0x0d, 0xf0, 0x91, 0x48, 0xc4, 0xd6, 0xc0,
	0x08, 0xcb, 0xff, 0x02, 0x5f, 0xe6, 0xd1, 0xf6, 0x1b, 0xd6, 0x58, 0xee, 0x01, 0x9f, 0x11, 0xf2,
	0x62, 0xa1, 0x4e, 0x2f, 0x87, 0x6b, 0x47, 0xd6, 0x0d, 0xc4, 0xed, 0x59, 0x3a, 0x2f, 0x17, 0xea,
	0xf4, 0x2f, 0xec, 0x2e, 0x1c, 0x4d, 0x96, 0xdf, 0x8d, 0x9d, 0xe5, 0xca, 0x80, 0xcf, 0x2b, 0x03,
	0xbe, 0xad, 0x0c, 0x08, 0xc6, 0x58, 0xbd, 0xf7, 0x27, 0x00, 0x00, 0xff, 0xff, 0xc0, 0xe6, 0x4f,
	0xd3, 0x2f, 0x03, 0x00, 0x00,
}
