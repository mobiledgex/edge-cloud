// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: operatorcode.proto

package edgeproto

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"

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

// OperatorCode maps a carrier code to an Operator organization name
type OperatorCode struct {
	// MCC plus MNC code, or custom carrier code designation.
	Code string `protobuf:"bytes,1,opt,name=code,proto3" json:"code,omitempty"`
	// Operator Organization name
	Organization string `protobuf:"bytes,2,opt,name=organization,proto3" json:"organization,omitempty"`
}

func (m *OperatorCode) Reset()                    { *m = OperatorCode{} }
func (m *OperatorCode) String() string            { return proto.CompactTextString(m) }
func (*OperatorCode) ProtoMessage()               {}
func (*OperatorCode) Descriptor() ([]byte, []int) { return fileDescriptorOperatorcode, []int{0} }

func init() {
	proto.RegisterType((*OperatorCode)(nil), "edgeproto.OperatorCode")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for OperatorCodeApi service

type OperatorCodeApiClient interface {
	// Create Operator Code. Create a code for an Operator.
	CreateOperatorCode(ctx context.Context, in *OperatorCode, opts ...grpc.CallOption) (*Result, error)
	// Delete Operator Code. Delete a code for an Operator.
	DeleteOperatorCode(ctx context.Context, in *OperatorCode, opts ...grpc.CallOption) (*Result, error)
	// Show Operator Code. Show Codes for an Operator.
	ShowOperatorCode(ctx context.Context, in *OperatorCode, opts ...grpc.CallOption) (OperatorCodeApi_ShowOperatorCodeClient, error)
}

type operatorCodeApiClient struct {
	cc *grpc.ClientConn
}

func NewOperatorCodeApiClient(cc *grpc.ClientConn) OperatorCodeApiClient {
	return &operatorCodeApiClient{cc}
}

func (c *operatorCodeApiClient) CreateOperatorCode(ctx context.Context, in *OperatorCode, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.OperatorCodeApi/CreateOperatorCode", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *operatorCodeApiClient) DeleteOperatorCode(ctx context.Context, in *OperatorCode, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/edgeproto.OperatorCodeApi/DeleteOperatorCode", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *operatorCodeApiClient) ShowOperatorCode(ctx context.Context, in *OperatorCode, opts ...grpc.CallOption) (OperatorCodeApi_ShowOperatorCodeClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_OperatorCodeApi_serviceDesc.Streams[0], c.cc, "/edgeproto.OperatorCodeApi/ShowOperatorCode", opts...)
	if err != nil {
		return nil, err
	}
	x := &operatorCodeApiShowOperatorCodeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type OperatorCodeApi_ShowOperatorCodeClient interface {
	Recv() (*OperatorCode, error)
	grpc.ClientStream
}

type operatorCodeApiShowOperatorCodeClient struct {
	grpc.ClientStream
}

func (x *operatorCodeApiShowOperatorCodeClient) Recv() (*OperatorCode, error) {
	m := new(OperatorCode)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for OperatorCodeApi service

type OperatorCodeApiServer interface {
	// Create Operator Code. Create a code for an Operator.
	CreateOperatorCode(context.Context, *OperatorCode) (*Result, error)
	// Delete Operator Code. Delete a code for an Operator.
	DeleteOperatorCode(context.Context, *OperatorCode) (*Result, error)
	// Show Operator Code. Show Codes for an Operator.
	ShowOperatorCode(*OperatorCode, OperatorCodeApi_ShowOperatorCodeServer) error
}

func RegisterOperatorCodeApiServer(s *grpc.Server, srv OperatorCodeApiServer) {
	s.RegisterService(&_OperatorCodeApi_serviceDesc, srv)
}

func _OperatorCodeApi_CreateOperatorCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OperatorCode)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OperatorCodeApiServer).CreateOperatorCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.OperatorCodeApi/CreateOperatorCode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OperatorCodeApiServer).CreateOperatorCode(ctx, req.(*OperatorCode))
	}
	return interceptor(ctx, in, info, handler)
}

func _OperatorCodeApi_DeleteOperatorCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OperatorCode)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OperatorCodeApiServer).DeleteOperatorCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/edgeproto.OperatorCodeApi/DeleteOperatorCode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OperatorCodeApiServer).DeleteOperatorCode(ctx, req.(*OperatorCode))
	}
	return interceptor(ctx, in, info, handler)
}

func _OperatorCodeApi_ShowOperatorCode_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(OperatorCode)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(OperatorCodeApiServer).ShowOperatorCode(m, &operatorCodeApiShowOperatorCodeServer{stream})
}

type OperatorCodeApi_ShowOperatorCodeServer interface {
	Send(*OperatorCode) error
	grpc.ServerStream
}

type operatorCodeApiShowOperatorCodeServer struct {
	grpc.ServerStream
}

func (x *operatorCodeApiShowOperatorCodeServer) Send(m *OperatorCode) error {
	return x.ServerStream.SendMsg(m)
}

var _OperatorCodeApi_serviceDesc = grpc.ServiceDesc{
	ServiceName: "edgeproto.OperatorCodeApi",
	HandlerType: (*OperatorCodeApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateOperatorCode",
			Handler:    _OperatorCodeApi_CreateOperatorCode_Handler,
		},
		{
			MethodName: "DeleteOperatorCode",
			Handler:    _OperatorCodeApi_DeleteOperatorCode_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ShowOperatorCode",
			Handler:       _OperatorCodeApi_ShowOperatorCode_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "operatorcode.proto",
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
	if len(m.Code) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintOperatorcode(dAtA, i, uint64(len(m.Code)))
		i += copy(dAtA[i:], m.Code)
	}
	if len(m.Organization) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintOperatorcode(dAtA, i, uint64(len(m.Organization)))
		i += copy(dAtA[i:], m.Organization)
	}
	return i, nil
}

func encodeVarintOperatorcode(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *OperatorCode) Matches(o *OperatorCode, fopts ...MatchOpt) bool {
	opts := MatchOptions{}
	applyMatchOptions(&opts, fopts...)
	if o == nil {
		if opts.Filter {
			return true
		}
		return false
	}
	if !opts.Filter || o.Code != "" {
		if o.Code != m.Code {
			return false
		}
	}
	if !opts.Filter || o.Organization != "" {
		if o.Organization != m.Organization {
			return false
		}
	}
	return true
}

func (m *OperatorCode) CopyInFields(src *OperatorCode) int {
	changed := 0
	if m.Code != src.Code {
		m.Code = src.Code
		changed++
	}
	if m.Organization != src.Organization {
		m.Organization = src.Organization
		changed++
	}
	return changed
}

func (m *OperatorCode) DeepCopyIn(src *OperatorCode) {
	m.Code = src.Code
	m.Organization = src.Organization
}

func (s *OperatorCode) HasFields() bool {
	return false
}

type OperatorCodeStore struct {
	kvstore objstore.KVStore
}

func NewOperatorCodeStore(kvstore objstore.KVStore) OperatorCodeStore {
	return OperatorCodeStore{kvstore: kvstore}
}

func (s *OperatorCodeStore) Create(ctx context.Context, m *OperatorCode, wait func(int64)) (*Result, error) {
	err := m.Validate(nil)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("OperatorCode", m.GetKey())
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

func (s *OperatorCodeStore) Update(ctx context.Context, m *OperatorCode, wait func(int64)) (*Result, error) {
	err := m.Validate(nil)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("OperatorCode", m.GetKey())
	var vers int64 = 0
	val, err := json.Marshal(m)
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

func (s *OperatorCodeStore) Put(ctx context.Context, m *OperatorCode, wait func(int64), ops ...objstore.KVOp) (*Result, error) {
	err := m.Validate(nil)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("OperatorCode", m.GetKey())
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

func (s *OperatorCodeStore) Delete(ctx context.Context, m *OperatorCode, wait func(int64)) (*Result, error) {
	err := m.GetKey().ValidateKey()
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("OperatorCode", m.GetKey())
	rev, err := s.kvstore.Delete(ctx, key)
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *OperatorCodeStore) LoadOne(key string) (*OperatorCode, int64, error) {
	val, rev, _, err := s.kvstore.Get(key)
	if err != nil {
		return nil, 0, err
	}
	var obj OperatorCode
	err = json.Unmarshal(val, &obj)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "Failed to parse OperatorCode data", "val", string(val), "err", err)
		return nil, 0, err
	}
	return &obj, rev, nil
}

func (s *OperatorCodeStore) STMGet(stm concurrency.STM, key *OperatorCodeKey, buf *OperatorCode) bool {
	keystr := objstore.DbKeyString("OperatorCode", key)
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

func (s *OperatorCodeStore) STMPut(stm concurrency.STM, obj *OperatorCode, ops ...objstore.KVOp) {
	keystr := objstore.DbKeyString("OperatorCode", obj.GetKey())
	val, err := json.Marshal(obj)
	if err != nil {
		log.InfoLog("OperatorCode json marsahal failed", "obj", obj, "err", err)
	}
	v3opts := GetSTMOpts(ops...)
	stm.Put(keystr, string(val), v3opts...)
}

func (s *OperatorCodeStore) STMDel(stm concurrency.STM, key *OperatorCodeKey) {
	keystr := objstore.DbKeyString("OperatorCode", key)
	stm.Del(keystr)
}

type OperatorCodeKeyWatcher struct {
	cb func(ctx context.Context)
}

type OperatorCodeCacheData struct {
	Obj    *OperatorCode
	ModRev int64
}

// OperatorCodeCache caches OperatorCode objects in memory in a hash table
// and keeps them in sync with the database.
type OperatorCodeCache struct {
	Objs          map[OperatorCodeKey]*OperatorCodeCacheData
	Mux           util.Mutex
	List          map[OperatorCodeKey]struct{}
	FlushAll      bool
	NotifyCb      func(ctx context.Context, obj *OperatorCodeKey, old *OperatorCode, modRev int64)
	UpdatedCbs    []func(ctx context.Context, old *OperatorCode, new *OperatorCode)
	DeletedCbs    []func(ctx context.Context, old *OperatorCode)
	KeyWatchers   map[OperatorCodeKey][]*OperatorCodeKeyWatcher
	UpdatedKeyCbs []func(ctx context.Context, key *OperatorCodeKey)
	DeletedKeyCbs []func(ctx context.Context, key *OperatorCodeKey)
}

func NewOperatorCodeCache() *OperatorCodeCache {
	cache := OperatorCodeCache{}
	InitOperatorCodeCache(&cache)
	return &cache
}

func InitOperatorCodeCache(cache *OperatorCodeCache) {
	cache.Objs = make(map[OperatorCodeKey]*OperatorCodeCacheData)
	cache.KeyWatchers = make(map[OperatorCodeKey][]*OperatorCodeKeyWatcher)
	cache.NotifyCb = nil
	cache.UpdatedCbs = nil
	cache.DeletedCbs = nil
	cache.UpdatedKeyCbs = nil
	cache.DeletedKeyCbs = nil
}

func (c *OperatorCodeCache) GetTypeString() string {
	return "OperatorCode"
}

func (c *OperatorCodeCache) Get(key *OperatorCodeKey, valbuf *OperatorCode) bool {
	var modRev int64
	return c.GetWithRev(key, valbuf, &modRev)
}

func (c *OperatorCodeCache) GetWithRev(key *OperatorCodeKey, valbuf *OperatorCode, modRev *int64) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	inst, found := c.Objs[*key]
	if found {
		valbuf.DeepCopyIn(inst.Obj)
		*modRev = inst.ModRev
	}
	return found
}

func (c *OperatorCodeCache) HasKey(key *OperatorCodeKey) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	_, found := c.Objs[*key]
	return found
}

func (c *OperatorCodeCache) GetAllKeys(ctx context.Context, cb func(key *OperatorCodeKey, modRev int64)) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for key, data := range c.Objs {
		cb(&key, data.ModRev)
	}
}

func (c *OperatorCodeCache) Update(ctx context.Context, in *OperatorCode, modRev int64) {
	c.UpdateModFunc(ctx, in.GetKey(), modRev, func(old *OperatorCode) (*OperatorCode, bool) {
		return in, true
	})
}

func (c *OperatorCodeCache) UpdateModFunc(ctx context.Context, key *OperatorCodeKey, modRev int64, modFunc func(old *OperatorCode) (new *OperatorCode, changed bool)) {
	c.Mux.Lock()
	var old *OperatorCode
	if oldData, found := c.Objs[*key]; found {
		old = oldData.Obj
	}
	new, changed := modFunc(old)
	if !changed {
		c.Mux.Unlock()
		return
	}
	for _, cb := range c.UpdatedCbs {
		newCopy := &OperatorCode{}
		newCopy.DeepCopyIn(new)
		defer cb(ctx, old, newCopy)
	}
	if c.NotifyCb != nil {
		defer c.NotifyCb(ctx, new.GetKey(), old, modRev)
	}
	for _, cb := range c.UpdatedKeyCbs {
		defer cb(ctx, key)
	}
	store := &OperatorCode{}
	store.DeepCopyIn(new)
	c.Objs[new.GetKeyVal()] = &OperatorCodeCacheData{
		Obj:    store,
		ModRev: modRev,
	}
	log.SpanLog(ctx, log.DebugLevelApi, "cache update", "new", store)
	c.Mux.Unlock()
	c.TriggerKeyWatchers(ctx, new.GetKey())
}

func (c *OperatorCodeCache) Delete(ctx context.Context, in *OperatorCode, modRev int64) {
	c.Mux.Lock()
	var old *OperatorCode
	oldData, found := c.Objs[in.GetKeyVal()]
	if found {
		old = oldData.Obj
	}
	delete(c.Objs, in.GetKeyVal())
	log.SpanLog(ctx, log.DebugLevelApi, "cache delete")
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		c.NotifyCb(ctx, in.GetKey(), old, modRev)
	}
	if old != nil {
		for _, cb := range c.DeletedCbs {
			cb(ctx, old)
		}
	}
	for _, cb := range c.DeletedKeyCbs {
		cb(ctx, in.GetKey())
	}
	c.TriggerKeyWatchers(ctx, in.GetKey())
}

func (c *OperatorCodeCache) Prune(ctx context.Context, validKeys map[OperatorCodeKey]struct{}) {
	notify := make(map[OperatorCodeKey]*OperatorCodeCacheData)
	c.Mux.Lock()
	for key, _ := range c.Objs {
		if _, ok := validKeys[key]; !ok {
			if c.NotifyCb != nil || len(c.DeletedKeyCbs) > 0 || len(c.DeletedCbs) > 0 {
				notify[key] = c.Objs[key]
			}
			delete(c.Objs, key)
		}
	}
	c.Mux.Unlock()
	for key, old := range notify {
		if c.NotifyCb != nil {
			c.NotifyCb(ctx, &key, old.Obj, old.ModRev)
		}
		for _, cb := range c.DeletedKeyCbs {
			cb(ctx, &key)
		}
		if old.Obj != nil {
			for _, cb := range c.DeletedCbs {
				cb(ctx, old.Obj)
			}
		}
		c.TriggerKeyWatchers(ctx, &key)
	}
}

func (c *OperatorCodeCache) GetCount() int {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	return len(c.Objs)
}

func (c *OperatorCodeCache) Flush(ctx context.Context, notifyId int64) {
}

func (c *OperatorCodeCache) Show(filter *OperatorCode, cb func(ret *OperatorCode) error) error {
	log.DebugLog(log.DebugLevelApi, "Show OperatorCode", "count", len(c.Objs))
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for _, data := range c.Objs {
		log.DebugLog(log.DebugLevelApi, "Compare OperatorCode", "filter", filter, "data", data)
		if !data.Obj.Matches(filter, MatchFilter()) {
			continue
		}
		log.DebugLog(log.DebugLevelApi, "Show OperatorCode", "obj", data.Obj)
		err := cb(data.Obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func OperatorCodeGenericNotifyCb(fn func(key *OperatorCodeKey, old *OperatorCode)) func(objstore.ObjKey, objstore.Obj) {
	return func(objkey objstore.ObjKey, obj objstore.Obj) {
		fn(objkey.(*OperatorCodeKey), obj.(*OperatorCode))
	}
}

func (c *OperatorCodeCache) SetNotifyCb(fn func(ctx context.Context, obj *OperatorCodeKey, old *OperatorCode, modRev int64)) {
	c.NotifyCb = fn
}

func (c *OperatorCodeCache) SetUpdatedCb(fn func(ctx context.Context, old *OperatorCode, new *OperatorCode)) {
	c.UpdatedCbs = []func(ctx context.Context, old *OperatorCode, new *OperatorCode){fn}
}

func (c *OperatorCodeCache) SetDeletedCb(fn func(ctx context.Context, old *OperatorCode)) {
	c.DeletedCbs = []func(ctx context.Context, old *OperatorCode){fn}
}

func (c *OperatorCodeCache) SetUpdatedKeyCb(fn func(ctx context.Context, key *OperatorCodeKey)) {
	c.UpdatedKeyCbs = []func(ctx context.Context, key *OperatorCodeKey){fn}
}

func (c *OperatorCodeCache) SetDeletedKeyCb(fn func(ctx context.Context, key *OperatorCodeKey)) {
	c.DeletedKeyCbs = []func(ctx context.Context, key *OperatorCodeKey){fn}
}

func (c *OperatorCodeCache) AddUpdatedCb(fn func(ctx context.Context, old *OperatorCode, new *OperatorCode)) {
	c.UpdatedCbs = append(c.UpdatedCbs, fn)
}

func (c *OperatorCodeCache) AddDeletedCb(fn func(ctx context.Context, old *OperatorCode)) {
	c.DeletedCbs = append(c.DeletedCbs, fn)
}

func (c *OperatorCodeCache) AddUpdatedKeyCb(fn func(ctx context.Context, key *OperatorCodeKey)) {
	c.UpdatedKeyCbs = append(c.UpdatedKeyCbs, fn)
}

func (c *OperatorCodeCache) AddDeletedKeyCb(fn func(ctx context.Context, key *OperatorCodeKey)) {
	c.DeletedKeyCbs = append(c.DeletedKeyCbs, fn)
}

func (c *OperatorCodeCache) SetFlushAll() {
	c.FlushAll = true
}

func (c *OperatorCodeCache) WatchKey(key *OperatorCodeKey, cb func(ctx context.Context)) context.CancelFunc {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	list, ok := c.KeyWatchers[*key]
	if !ok {
		list = make([]*OperatorCodeKeyWatcher, 0)
	}
	watcher := OperatorCodeKeyWatcher{cb: cb}
	c.KeyWatchers[*key] = append(list, &watcher)
	log.DebugLog(log.DebugLevelApi, "Watching OperatorCode", "key", key)
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

func (c *OperatorCodeCache) TriggerKeyWatchers(ctx context.Context, key *OperatorCodeKey) {
	watchers := make([]*OperatorCodeKeyWatcher, 0)
	c.Mux.Lock()
	if list, ok := c.KeyWatchers[*key]; ok {
		watchers = append(watchers, list...)
	}
	c.Mux.Unlock()
	for ii, _ := range watchers {
		watchers[ii].cb(ctx)
	}
}

// Note that we explicitly ignore the global revision number, because of the way
// the notify framework sends updates (by hashing keys and doing lookups, instead
// of sequentially through a history buffer), updates may be done out-of-order
// or multiple updates compressed into one update, so the state of the cache at
// any point in time may not by in sync with a particular database revision number.

func (c *OperatorCodeCache) SyncUpdate(ctx context.Context, key, val []byte, rev, modRev int64) {
	obj := OperatorCode{}
	err := json.Unmarshal(val, &obj)
	if err != nil {
		log.WarnLog("Failed to parse OperatorCode data", "val", string(val), "err", err)
		return
	}
	c.Update(ctx, &obj, modRev)
	c.Mux.Lock()
	if c.List != nil {
		c.List[obj.GetKeyVal()] = struct{}{}
	}
	c.Mux.Unlock()
}

func (c *OperatorCodeCache) SyncDelete(ctx context.Context, key []byte, rev, modRev int64) {
	obj := OperatorCode{}
	keystr := objstore.DbKeyPrefixRemove(string(key))
	OperatorCodeKeyStringParse(keystr, &obj)
	c.Delete(ctx, &obj, modRev)
}

func (c *OperatorCodeCache) SyncListStart(ctx context.Context) {
	c.List = make(map[OperatorCodeKey]struct{})
}

func (c *OperatorCodeCache) SyncListEnd(ctx context.Context) {
	deleted := make(map[OperatorCodeKey]*OperatorCodeCacheData)
	c.Mux.Lock()
	for key, val := range c.Objs {
		if _, found := c.List[key]; !found {
			deleted[key] = val
			delete(c.Objs, key)
		}
	}
	c.List = nil
	c.Mux.Unlock()
	for key, val := range deleted {
		if c.NotifyCb != nil {
			c.NotifyCb(ctx, &key, val.Obj, val.ModRev)
		}
		for _, cb := range c.DeletedKeyCbs {
			cb(ctx, &key)
		}
		if val.Obj != nil {
			for _, cb := range c.DeletedCbs {
				cb(ctx, val.Obj)
			}
		}
		c.TriggerKeyWatchers(ctx, &key)
	}
}

func (c *OperatorCodeCache) UsesOrg(org string) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for _, val := range c.Objs {
		if val.Obj.Organization == org {
			return true
		}
	}
	return false
}

// Helper method to check that enums have valid values
func (m *OperatorCode) ValidateEnums() error {
	return nil
}

func (m *OperatorCode) Size() (n int) {
	var l int
	_ = l
	l = len(m.Code)
	if l > 0 {
		n += 1 + l + sovOperatorcode(uint64(l))
	}
	l = len(m.Organization)
	if l > 0 {
		n += 1 + l + sovOperatorcode(uint64(l))
	}
	return n
}

func sovOperatorcode(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozOperatorcode(x uint64) (n int) {
	return sovOperatorcode(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *OperatorCode) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowOperatorcode
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
				return fmt.Errorf("proto: wrong wireType = %d for field Code", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOperatorcode
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
				return ErrInvalidLengthOperatorcode
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Code = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Organization", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOperatorcode
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
				return ErrInvalidLengthOperatorcode
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Organization = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipOperatorcode(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthOperatorcode
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
func skipOperatorcode(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowOperatorcode
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
					return 0, ErrIntOverflowOperatorcode
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
					return 0, ErrIntOverflowOperatorcode
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
				return 0, ErrInvalidLengthOperatorcode
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowOperatorcode
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
				next, err := skipOperatorcode(dAtA[start:])
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
	ErrInvalidLengthOperatorcode = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowOperatorcode   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("operatorcode.proto", fileDescriptorOperatorcode) }

var fileDescriptorOperatorcode = []byte{
	// 407 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xca, 0x2f, 0x48, 0x2d,
	0x4a, 0x2c, 0xc9, 0x2f, 0x4a, 0xce, 0x4f, 0x49, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2,
	0x4c, 0x4d, 0x49, 0x4f, 0x05, 0x33, 0xa5, 0x64, 0xd2, 0xf3, 0xf3, 0xd3, 0x73, 0x52, 0xf5, 0x13,
	0x0b, 0x32, 0xf5, 0x13, 0xf3, 0xf2, 0xf2, 0x4b, 0x12, 0x4b, 0x32, 0xf3, 0xf3, 0x8a, 0x21, 0x0a,
	0xa5, 0x2c, 0xd2, 0x33, 0x4b, 0x32, 0x4a, 0x93, 0xf4, 0x92, 0xf3, 0x73, 0xf5, 0x73, 0xf3, 0x93,
	0x32, 0x73, 0x40, 0x1a, 0x2b, 0xf4, 0x41, 0xa4, 0x6e, 0x72, 0x4e, 0x7e, 0x69, 0x8a, 0x3e, 0x58,
	0x5d, 0x7a, 0x6a, 0x1e, 0x9c, 0x01, 0xd5, 0xc9, 0x53, 0x94, 0x5a, 0x5c, 0x9a, 0x53, 0x02, 0xe1,
	0x29, 0xb5, 0x33, 0x72, 0xf1, 0xf8, 0x43, 0xdd, 0xe1, 0x9c, 0x9f, 0x92, 0x2a, 0x24, 0xc4, 0xc5,
	0x02, 0x72, 0x8f, 0x04, 0xa3, 0x02, 0xa3, 0x06, 0x67, 0x10, 0x98, 0x2d, 0xa4, 0xc4, 0xc5, 0x93,
	0x5f, 0x94, 0x9e, 0x98, 0x97, 0x59, 0x05, 0x76, 0x83, 0x04, 0x13, 0x58, 0x0e, 0x45, 0xcc, 0xca,
	0xfe, 0xc5, 0x67, 0x09, 0xc6, 0x0f, 0x9f, 0x25, 0x18, 0x1b, 0xbe, 0x48, 0x30, 0x4e, 0xf8, 0x22,
	0xc1, 0x38, 0xe3, 0x8b, 0x04, 0xe3, 0x86, 0xaf, 0x12, 0x0c, 0x9f, 0xbe, 0x4a, 0xf0, 0x23, 0xdb,
	0xe0, 0x9d, 0x5a, 0xb9, 0xe9, 0x9b, 0x84, 0x40, 0x59, 0x62, 0x8e, 0xad, 0x3f, 0x92, 0x01, 0x46,
	0x77, 0x98, 0xb9, 0x50, 0xd4, 0x39, 0x16, 0x64, 0x0a, 0x2d, 0x60, 0xe4, 0x12, 0x72, 0x2e, 0x4a,
	0x4d, 0x2c, 0x49, 0x45, 0x71, 0xa3, 0xb8, 0x1e, 0x3c, 0x98, 0xf4, 0x90, 0x25, 0xa4, 0x04, 0x91,
	0x24, 0x82, 0xc0, 0xde, 0x54, 0x8a, 0x7b, 0xf5, 0x45, 0x42, 0x3b, 0x28, 0xb5, 0x38, 0xbf, 0xb4,
	0x28, 0x39, 0xd5, 0x19, 0x14, 0x32, 0x39, 0xa9, 0x25, 0xc5, 0x3a, 0x8e, 0xc9, 0x20, 0x4b, 0x7d,
	0x13, 0xf3, 0x12, 0xd3, 0x53, 0x75, 0x90, 0xdd, 0xb1, 0xea, 0x9b, 0x04, 0x0f, 0x32, 0xbf, 0xe9,
	0xf2, 0x93, 0xc9, 0x4c, 0x92, 0x4a, 0x22, 0xfa, 0xc9, 0x60, 0x77, 0xe8, 0x23, 0x47, 0x9a, 0x15,
	0xa3, 0x96, 0xd0, 0x04, 0x46, 0x2e, 0x21, 0x97, 0xd4, 0x9c, 0x54, 0x0a, 0x9c, 0xe8, 0x47, 0xa2,
	0x13, 0xe1, 0x4e, 0x4a, 0x01, 0xdb, 0x8b, 0xe1, 0xa4, 0x49, 0x8c, 0x5c, 0x02, 0xc1, 0x19, 0xf9,
	0xe5, 0xc4, 0x39, 0x08, 0x97, 0x84, 0x92, 0xd7, 0xab, 0x2f, 0x12, 0x9a, 0xb8, 0x9c, 0x15, 0x96,
	0x99, 0x5a, 0x8e, 0xe9, 0x28, 0x71, 0x25, 0x21, 0xfd, 0xe2, 0x8c, 0xfc, 0x72, 0x74, 0x27, 0x19,
	0x30, 0x3a, 0x09, 0x9c, 0x78, 0x28, 0xc7, 0x70, 0xe2, 0x91, 0x1c, 0xe3, 0x85, 0x47, 0x72, 0x8c,
	0x0f, 0x1e, 0xc9, 0x31, 0x26, 0xb1, 0x81, 0xed, 0x34, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0xb3,
	0xf5, 0xa3, 0xad, 0x08, 0x03, 0x00, 0x00,
}
