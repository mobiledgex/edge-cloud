// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: operator.proto

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

// OperatorKey uniquely identifies an Operator
type OperatorKey struct {
	// Company or Organization name of the operator
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (m *OperatorKey) Reset()                    { *m = OperatorKey{} }
func (m *OperatorKey) String() string            { return proto.CompactTextString(m) }
func (*OperatorKey) ProtoMessage()               {}
func (*OperatorKey) Descriptor() ([]byte, []int) { return fileDescriptorOperator, []int{0} }

// An Operator supplies compute resources.
// For example, telecommunications provider such as AT&T is an Operator
type Operator struct {
	// Fields are used for the Update API to specify which fields to apply
	Fields []string `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	// Unique identifier key
	Key OperatorKey `protobuf:"bytes,2,opt,name=key" json:"key"`
}

func (m *Operator) Reset()                    { *m = Operator{} }
func (m *Operator) String() string            { return proto.CompactTextString(m) }
func (*Operator) ProtoMessage()               {}
func (*Operator) Descriptor() ([]byte, []int) { return fileDescriptorOperator, []int{1} }

func init() {
	proto.RegisterType((*OperatorKey)(nil), "edgeproto.OperatorKey")
	proto.RegisterType((*Operator)(nil), "edgeproto.Operator")
}
func (this *OperatorKey) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 5)
	s = append(s, "&edgeproto.OperatorKey{")
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
	// Create an Operator
	CreateOperator(ctx context.Context, in *Operator, opts ...grpc.CallOption) (*Result, error)
	// Delete an Operator
	DeleteOperator(ctx context.Context, in *Operator, opts ...grpc.CallOption) (*Result, error)
	// Update an Operator
	UpdateOperator(ctx context.Context, in *Operator, opts ...grpc.CallOption) (*Result, error)
	// Show Operators
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
	err := grpc.Invoke(ctx, "/edgeproto.OperatorApi/CreateOperator", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *operatorApiClient) DeleteOperator(ctx context.Context, in *Operator, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.OperatorApi/DeleteOperator", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *operatorApiClient) UpdateOperator(ctx context.Context, in *Operator, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.OperatorApi/UpdateOperator", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *operatorApiClient) ShowOperator(ctx context.Context, in *Operator, opts ...grpc.CallOption) (OperatorApi_ShowOperatorClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_OperatorApi_serviceDesc.Streams[0], c.cc, "/edgeproto.OperatorApi/ShowOperator", opts...)
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
	// Create an Operator
	CreateOperator(context.Context, *Operator) (*Result, error)
	// Delete an Operator
	DeleteOperator(context.Context, *Operator) (*Result, error)
	// Update an Operator
	UpdateOperator(context.Context, *Operator) (*Result, error)
	// Show Operators
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
		FullMethod: "/edgeproto.OperatorApi/CreateOperator",
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
		FullMethod: "/edgeproto.OperatorApi/DeleteOperator",
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
		FullMethod: "/edgeproto.OperatorApi/UpdateOperator",
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
	ServiceName: "edgeproto.OperatorApi",
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
func (m *OperatorKey) Matches(o *OperatorKey, fopts ...MatchOpt) bool {
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

func (m *OperatorKey) CopyInFields(src *OperatorKey) {
	m.Name = src.Name
}

func (m *OperatorKey) GetKeyString() string {
	key, err := json.Marshal(m)
	if err != nil {
		log.FatalLog("Failed to marshal OperatorKey key string", "obj", m)
	}
	return string(key)
}

func OperatorKeyStringParse(str string, key *OperatorKey) {
	err := json.Unmarshal([]byte(str), key)
	if err != nil {
		log.FatalLog("Failed to unmarshal OperatorKey key string", "str", str)
	}
}

// Helper method to check that enums have valid values
func (m *OperatorKey) ValidateEnums() error {
	return nil
}

func (m *Operator) Matches(o *Operator, fopts ...MatchOpt) bool {
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
	return true
}

const OperatorFieldKey = "2"
const OperatorFieldKeyName = "2.1"

var OperatorAllFields = []string{
	OperatorFieldKeyName,
}

var OperatorAllFieldsMap = map[string]struct{}{
	OperatorFieldKeyName: struct{}{},
}

func (m *Operator) DiffFields(o *Operator, fields map[string]struct{}) {
	if m.Key.Name != o.Key.Name {
		fields[OperatorFieldKeyName] = struct{}{}
		fields[OperatorFieldKey] = struct{}{}
	}
}

func (m *Operator) CopyInFields(src *Operator) {
	fmap := MakeFieldMap(src.Fields)
	if _, set := fmap["2"]; set {
		if _, set := fmap["2.1"]; set {
			m.Key.Name = src.Key.Name
		}
	}
}

func (s *Operator) HasFields() bool {
	return true
}

type OperatorStore struct {
	kvstore objstore.KVStore
}

func NewOperatorStore(kvstore objstore.KVStore) OperatorStore {
	return OperatorStore{kvstore: kvstore}
}

func (s *OperatorStore) Create(m *Operator, wait func(int64)) (*Result, error) {
	err := m.Validate(OperatorAllFieldsMap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Operator", m.GetKey())
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

func (s *OperatorStore) Update(m *Operator, wait func(int64)) (*Result, error) {
	fmap := MakeFieldMap(m.Fields)
	err := m.Validate(fmap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Operator", m.GetKey())
	var vers int64 = 0
	curBytes, vers, _, err := s.kvstore.Get(key)
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
	rev, err := s.kvstore.Update(key, string(val), vers)
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *OperatorStore) Put(m *Operator, wait func(int64), ops ...objstore.KVOp) (*Result, error) {
	fmap := MakeFieldMap(m.Fields)
	err := m.Validate(fmap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Operator", m.GetKey())
	var val []byte
	curBytes, _, _, err := s.kvstore.Get(key)
	if err == nil {
		var cur Operator
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
	rev, err := s.kvstore.Put(key, string(val), ops...)
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *OperatorStore) Delete(m *Operator, wait func(int64)) (*Result, error) {
	err := m.GetKey().Validate()
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Operator", m.GetKey())
	rev, err := s.kvstore.Delete(key)
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *OperatorStore) LoadOne(key string) (*Operator, int64, error) {
	val, rev, _, err := s.kvstore.Get(key)
	if err != nil {
		return nil, 0, err
	}
	var obj Operator
	err = json.Unmarshal(val, &obj)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "Failed to parse Operator data", "val", string(val))
		return nil, 0, err
	}
	return &obj, rev, nil
}

func (s *OperatorStore) STMGet(stm concurrency.STM, key *OperatorKey, buf *Operator) bool {
	keystr := objstore.DbKeyString("Operator", key)
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

func (s *OperatorStore) STMPut(stm concurrency.STM, obj *Operator, ops ...objstore.KVOp) {
	keystr := objstore.DbKeyString("Operator", obj.GetKey())
	val, err := json.Marshal(obj)
	if err != nil {
		log.InfoLog("Operator json marsahal failed", "obj", obj, "err", err)
	}
	v3opts := GetSTMOpts(ops...)
	stm.Put(keystr, string(val), v3opts...)
}

func (s *OperatorStore) STMDel(stm concurrency.STM, key *OperatorKey) {
	keystr := objstore.DbKeyString("Operator", key)
	stm.Del(keystr)
}

type OperatorKeyWatcher struct {
	cb func()
}

// OperatorCache caches Operator objects in memory in a hash table
// and keeps them in sync with the database.
type OperatorCache struct {
	Objs        map[OperatorKey]*Operator
	Mux         util.Mutex
	List        map[OperatorKey]struct{}
	NotifyCb    func(obj *OperatorKey, old *Operator)
	UpdatedCb   func(old *Operator, new *Operator)
	KeyWatchers map[OperatorKey][]*OperatorKeyWatcher
}

func NewOperatorCache() *OperatorCache {
	cache := OperatorCache{}
	InitOperatorCache(&cache)
	return &cache
}

func InitOperatorCache(cache *OperatorCache) {
	cache.Objs = make(map[OperatorKey]*Operator)
	cache.KeyWatchers = make(map[OperatorKey][]*OperatorKeyWatcher)
}

func (c *OperatorCache) GetTypeString() string {
	return "Operator"
}

func (c *OperatorCache) Get(key *OperatorKey, valbuf *Operator) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	inst, found := c.Objs[*key]
	if found {
		*valbuf = *inst
	}
	return found
}

func (c *OperatorCache) HasKey(key *OperatorKey) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	_, found := c.Objs[*key]
	return found
}

func (c *OperatorCache) GetAllKeys(keys map[OperatorKey]struct{}) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for key, _ := range c.Objs {
		keys[key] = struct{}{}
	}
}

func (c *OperatorCache) Update(in *Operator, rev int64) {
	c.UpdateModFunc(&in.Key, rev, func(old *Operator) (*Operator, bool) {
		return in, true
	})
}

func (c *OperatorCache) UpdateModFunc(key *OperatorKey, rev int64, modFunc func(old *Operator) (new *Operator, changed bool)) {
	c.Mux.Lock()
	old := c.Objs[*key]
	new, changed := modFunc(old)
	if !changed {
		c.Mux.Unlock()
		return
	}
	if c.UpdatedCb != nil || c.NotifyCb != nil {
		if c.UpdatedCb != nil {
			newCopy := &Operator{}
			*newCopy = *new
			defer c.UpdatedCb(old, newCopy)
		}
		if c.NotifyCb != nil {
			defer c.NotifyCb(&new.Key, old)
		}
	}
	c.Objs[new.Key] = new
	log.DebugLog(log.DebugLevelApi, "SyncUpdate Operator", "obj", new, "rev", rev)
	c.Mux.Unlock()
	c.TriggerKeyWatchers(&new.Key)
}

func (c *OperatorCache) Delete(in *Operator, rev int64) {
	c.Mux.Lock()
	old := c.Objs[in.Key]
	delete(c.Objs, in.Key)
	log.DebugLog(log.DebugLevelApi, "SyncDelete Operator", "key", in.Key, "rev", rev)
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		c.NotifyCb(&in.Key, old)
	}
	c.TriggerKeyWatchers(&in.Key)
}

func (c *OperatorCache) Prune(validKeys map[OperatorKey]struct{}) {
	notify := make(map[OperatorKey]*Operator)
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
			c.NotifyCb(&key, old)
		}
		c.TriggerKeyWatchers(&key)
	}
}

func (c *OperatorCache) GetCount() int {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	return len(c.Objs)
}

func (c *OperatorCache) Flush(notifyId int64) {
}

func (c *OperatorCache) Show(filter *Operator, cb func(ret *Operator) error) error {
	log.DebugLog(log.DebugLevelApi, "Show Operator", "count", len(c.Objs))
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for _, obj := range c.Objs {
		if !obj.Matches(filter, MatchFilter()) {
			continue
		}
		log.DebugLog(log.DebugLevelApi, "Show Operator", "obj", obj)
		err := cb(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func OperatorGenericNotifyCb(fn func(key *OperatorKey, old *Operator)) func(objstore.ObjKey, objstore.Obj) {
	return func(objkey objstore.ObjKey, obj objstore.Obj) {
		fn(objkey.(*OperatorKey), obj.(*Operator))
	}
}

func (c *OperatorCache) SetNotifyCb(fn func(obj *OperatorKey, old *Operator)) {
	c.NotifyCb = fn
}

func (c *OperatorCache) SetUpdatedCb(fn func(old *Operator, new *Operator)) {
	c.UpdatedCb = fn
}

func (c *OperatorCache) WatchKey(key *OperatorKey, cb func()) context.CancelFunc {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	list, ok := c.KeyWatchers[*key]
	if !ok {
		list = make([]*OperatorKeyWatcher, 0)
	}
	watcher := OperatorKeyWatcher{cb: cb}
	c.KeyWatchers[*key] = append(list, &watcher)
	log.DebugLog(log.DebugLevelApi, "Watching Operator", "key", key)
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

func (c *OperatorCache) TriggerKeyWatchers(key *OperatorKey) {
	watchers := make([]*OperatorKeyWatcher, 0)
	c.Mux.Lock()
	if list, ok := c.KeyWatchers[*key]; ok {
		watchers = append(watchers, list...)
	}
	c.Mux.Unlock()
	for ii, _ := range watchers {
		watchers[ii].cb()
	}
}
func (c *OperatorCache) SyncUpdate(key, val []byte, rev int64) {
	obj := Operator{}
	err := json.Unmarshal(val, &obj)
	if err != nil {
		log.WarnLog("Failed to parse Operator data", "val", string(val))
		return
	}
	c.Update(&obj, rev)
	c.Mux.Lock()
	if c.List != nil {
		c.List[obj.Key] = struct{}{}
	}
	c.Mux.Unlock()
}

func (c *OperatorCache) SyncDelete(key []byte, rev int64) {
	obj := Operator{}
	keystr := objstore.DbKeyPrefixRemove(string(key))
	OperatorKeyStringParse(keystr, &obj.Key)
	c.Delete(&obj, rev)
}

func (c *OperatorCache) SyncListStart() {
	c.List = make(map[OperatorKey]struct{})
}

func (c *OperatorCache) SyncListEnd() {
	deleted := make(map[OperatorKey]*Operator)
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
			c.NotifyCb(&key, val)
			c.TriggerKeyWatchers(&key)
		}
	}
}

func (m *Operator) GetKey() objstore.ObjKey {
	return &m.Key
}

func CmpSortOperator(a Operator, b Operator) bool {
	return a.Key.GetKeyString() < b.Key.GetKeyString()
}

// Helper method to check that enums have valid values
// NOTE: ValidateEnums checks all Fields even if some are not set
func (m *Operator) ValidateEnums() error {
	if err := m.Key.ValidateEnums(); err != nil {
		return err
	}
	return nil
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
	if len(m.Fields) > 0 {
		for _, s := range m.Fields {
			l = len(s)
			n += 1 + l + sovOperator(uint64(l))
		}
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
			m.Fields = append(m.Fields, string(dAtA[iNdEx:postIndex]))
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

func init() { proto.RegisterFile("operator.proto", fileDescriptorOperator) }

var fileDescriptorOperator = []byte{
	// 407 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xcb, 0x2f, 0x48, 0x2d,
	0x4a, 0x2c, 0xc9, 0x2f, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x4c, 0x4d, 0x49, 0x4f,
	0x05, 0x33, 0xa5, 0x64, 0xd2, 0xf3, 0xf3, 0xd3, 0x73, 0x52, 0xf5, 0x13, 0x0b, 0x32, 0xf5, 0x13,
	0xf3, 0xf2, 0xf2, 0x4b, 0x12, 0x4b, 0x32, 0xf3, 0xf3, 0x8a, 0x21, 0x0a, 0xa5, 0x2c, 0xd2, 0x33,
	0x4b, 0x32, 0x4a, 0x93, 0xf4, 0x92, 0xf3, 0x73, 0xf5, 0x73, 0xf3, 0x93, 0x32, 0x73, 0x40, 0x1a,
	0x2b, 0xf4, 0x41, 0xa4, 0x6e, 0x72, 0x4e, 0x7e, 0x69, 0x8a, 0x3e, 0x58, 0x5d, 0x7a, 0x6a, 0x1e,
	0x9c, 0x01, 0xd5, 0xe9, 0x4e, 0x9c, 0xce, 0x64, 0xdd, 0xf4, 0xd4, 0x3c, 0xdd, 0xe4, 0x5c, 0x18,
	0x17, 0x89, 0x01, 0x35, 0x88, 0xa7, 0x28, 0xb5, 0xb8, 0x34, 0xa7, 0x04, 0xca, 0x13, 0x49, 0xcf,
	0x4f, 0xcf, 0x07, 0x33, 0xf5, 0x41, 0x2c, 0x88, 0xa8, 0x92, 0x3e, 0x17, 0xb7, 0x3f, 0xd4, 0x87,
	0xde, 0xa9, 0x95, 0x42, 0x42, 0x5c, 0x2c, 0x79, 0x89, 0xb9, 0xa9, 0x12, 0x8c, 0x0a, 0x8c, 0x1a,
	0x9c, 0x41, 0x60, 0xb6, 0x15, 0xcf, 0x8b, 0xcf, 0x12, 0x8c, 0x3f, 0x3e, 0x4b, 0x30, 0x6e, 0x58,
	0x20, 0xcf, 0xa8, 0x54, 0xca, 0xc5, 0x01, 0xd3, 0x20, 0x24, 0xc6, 0xc5, 0x96, 0x96, 0x99, 0x9a,
	0x93, 0x52, 0x2c, 0xc1, 0xa8, 0xc0, 0xac, 0xc1, 0x19, 0x04, 0xe5, 0x09, 0xe9, 0x71, 0x31, 0x67,
	0xa7, 0x56, 0x4a, 0x30, 0x29, 0x30, 0x6a, 0x70, 0x1b, 0x89, 0xe9, 0xc1, 0x83, 0x4c, 0x0f, 0xc9,
	0x2a, 0x27, 0x96, 0x13, 0xf7, 0xe4, 0x19, 0x82, 0x40, 0x0a, 0xad, 0x14, 0x41, 0x36, 0x7c, 0xf8,
	0x2c, 0xc1, 0xd8, 0xf0, 0x45, 0x82, 0x71, 0xc6, 0x17, 0x09, 0xc6, 0x49, 0x9b, 0x24, 0x79, 0x41,
	0x76, 0xdb, 0x7a, 0xa7, 0x56, 0xea, 0xf9, 0x25, 0xe6, 0xa6, 0x1a, 0xbd, 0x64, 0x42, 0x38, 0xd4,
	0xb1, 0x20, 0x53, 0x28, 0x94, 0x8b, 0xcf, 0xb9, 0x28, 0x35, 0xb1, 0x24, 0x15, 0xee, 0x18, 0x61,
	0x2c, 0xf6, 0x48, 0x09, 0x22, 0x09, 0x06, 0x81, 0x43, 0x43, 0x49, 0xba, 0xe9, 0xf2, 0x93, 0xc9,
	0x4c, 0xa2, 0x4a, 0x02, 0xfa, 0xc9, 0x60, 0x03, 0xf4, 0x61, 0x31, 0x6c, 0xc5, 0xa8, 0x05, 0x32,
	0xd6, 0x25, 0x35, 0x27, 0x95, 0x22, 0x63, 0x53, 0xc0, 0x06, 0xa0, 0x1b, 0x1b, 0x5a, 0x90, 0x42,
	0x99, 0x6b, 0x4b, 0xc1, 0x06, 0xa0, 0x19, 0xcb, 0x13, 0x9c, 0x91, 0x5f, 0x8e, 0xdf, 0x50, 0x6c,
	0x82, 0x4a, 0x92, 0x60, 0x63, 0x85, 0x95, 0xf8, 0xf4, 0x8b, 0x33, 0xf2, 0xcb, 0x91, 0x0d, 0x35,
	0x60, 0x74, 0x12, 0x38, 0xf1, 0x50, 0x8e, 0xe1, 0xc4, 0x23, 0x39, 0xc6, 0x0b, 0x8f, 0xe4, 0x18,
	0x1f, 0x3c, 0x92, 0x63, 0x4c, 0x62, 0x03, 0x6b, 0x37, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0xe9,
	0x0f, 0x70, 0x80, 0x0e, 0x03, 0x00, 0x00,
}
