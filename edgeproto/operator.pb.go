// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: operator.proto

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
import "sync"
import "github.com/mobiledgex/edge-cloud/util"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
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
func (m *OperatorCode) String() string            { return proto.CompactTextString(m) }
func (*OperatorCode) ProtoMessage()               {}
func (*OperatorCode) Descriptor() ([]byte, []int) { return fileDescriptorOperator, []int{0} }

type OperatorKey struct {
	// Company or Organization name of the operator
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (m *OperatorKey) Reset()                    { *m = OperatorKey{} }
func (m *OperatorKey) String() string            { return proto.CompactTextString(m) }
func (*OperatorKey) ProtoMessage()               {}
func (*OperatorKey) Descriptor() ([]byte, []int) { return fileDescriptorOperator, []int{1} }

type Operator struct {
	Fields []byte `protobuf:"bytes,1,opt,name=fields,proto3" json:"fields,omitempty"`
	// Unique identifier key
	Key OperatorKey `protobuf:"bytes,2,opt,name=key" json:"key"`
}

func (m *Operator) Reset()                    { *m = Operator{} }
func (m *Operator) String() string            { return proto.CompactTextString(m) }
func (*Operator) ProtoMessage()               {}
func (*Operator) Descriptor() ([]byte, []int) { return fileDescriptorOperator, []int{2} }

func init() {
	proto.RegisterType((*OperatorCode)(nil), "edgeproto.OperatorCode")
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
	key, err := json.Marshal(m)
	if err != nil {
		util.FatalLog("Failed to marshal OperatorKey key string", "obj", m)
	}
	return string(key)
}

func OperatorKeyStringParse(str string, key *OperatorKey) {
	err := json.Unmarshal([]byte(str), key)
	if err != nil {
		util.FatalLog("Failed to unmarshal OperatorKey key string", "str", str)
	}
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

func (s *Operator) HasFields() bool {
	return true
}

type OperatorStore struct {
	objstore     objstore.ObjStore
	listOperator map[OperatorKey]struct{}
}

func NewOperatorStore(objstore objstore.ObjStore) OperatorStore {
	return OperatorStore{objstore: objstore}
}

type OperatorCacher interface {
	SyncOperatorUpdate(m *Operator, rev int64)
	SyncOperatorDelete(m *Operator, rev int64)
	SyncOperatorPrune(current map[OperatorKey]struct{})
	SyncOperatorRevOnly(rev int64)
}

func (s *OperatorStore) Create(m *Operator, wait func(int64)) (*Result, error) {
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

func (s *OperatorStore) Update(m *Operator, wait func(int64)) (*Result, error) {
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
	rev, err := s.objstore.Update(key, string(val), vers)
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

type OperatorCb func(m *Operator) error

func (s *OperatorStore) LoadAll(cb OperatorCb) error {
	loadkey := objstore.DbKeyPrefixString(&OperatorKey{})
	err := s.objstore.List(loadkey, func(key, val []byte, rev int64) error {
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

func (s *OperatorStore) LoadOne(key string) (*Operator, int64, error) {
	val, rev, err := s.objstore.Get(key)
	if err != nil {
		return nil, 0, err
	}
	var obj Operator
	err = json.Unmarshal(val, &obj)
	if err != nil {
		util.DebugLog(util.DebugLevelApi, "Failed to parse Operator data", "val", string(val))
		return nil, 0, err
	}
	return &obj, rev, nil
}

// Sync will sync changes for any Operator objects.
func (s *OperatorStore) Sync(ctx context.Context, cacher OperatorCacher) error {
	str := objstore.DbKeyPrefixString(&OperatorKey{})
	return s.objstore.Sync(ctx, str, func(in *objstore.SyncCbData) {
		obj := Operator{}
		// Even on parse error, we should still call back to keep
		// the revision numbers in sync so no caller hangs on wait.
		action := in.Action
		if action == objstore.SyncUpdate || action == objstore.SyncList {
			err := json.Unmarshal(in.Value, &obj)
			if err != nil {
				util.WarnLog("Failed to parse Operator data", "val", string(in.Value))
				action = objstore.SyncRevOnly
			}
		} else if action == objstore.SyncDelete {
			keystr := objstore.DbKeyPrefixRemove(string(in.Key))
			OperatorKeyStringParse(keystr, obj.GetKey())
		}
		util.DebugLog(util.DebugLevelApi, "Sync cb", "action", objstore.SyncActionStrs[in.Action], "key", string(in.Key), "value", string(in.Value), "rev", in.Rev)
		switch action {
		case objstore.SyncUpdate:
			cacher.SyncOperatorUpdate(&obj, in.Rev)
		case objstore.SyncDelete:
			cacher.SyncOperatorDelete(&obj, in.Rev)
		case objstore.SyncListStart:
			s.listOperator = make(map[OperatorKey]struct{})
		case objstore.SyncList:
			s.listOperator[obj.Key] = struct{}{}
			cacher.SyncOperatorUpdate(&obj, in.Rev)
		case objstore.SyncListEnd:
			cacher.SyncOperatorPrune(s.listOperator)
			s.listOperator = nil
		case objstore.SyncRevOnly:
			cacher.SyncOperatorRevOnly(in.Rev)
		}
	})
}

// OperatorCache caches Operator objects in memory in a hash table
// and keeps them in sync with the database.
type OperatorCache struct {
	Store      *OperatorStore
	Objs       map[OperatorKey]*Operator
	Rev        int64
	Mux        util.Mutex
	Cond       sync.Cond
	initWait   bool
	syncDone   bool
	syncCancel context.CancelFunc
	notifyCb   func(obj *OperatorKey)
}

func NewOperatorCache(store *OperatorStore) *OperatorCache {
	cache := OperatorCache{
		Store:    store,
		Objs:     make(map[OperatorKey]*Operator),
		initWait: true,
	}
	cache.Mux.InitCond(&cache.Cond)

	ctx, cancel := context.WithCancel(context.Background())
	cache.syncCancel = cancel
	go func() {
		err := cache.Store.Sync(ctx, &cache)
		if err != nil {
			util.WarnLog("Operator Sync failed", "err", err)
		}
		cache.syncDone = true
		cache.Cond.Broadcast()
	}()
	return &cache
}

func (c *OperatorCache) WaitInitSyncDone() {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for c.initWait {
		c.Cond.Wait()
	}
}

func (c *OperatorCache) Done() {
	c.syncCancel()
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for !c.syncDone {
		c.Cond.Wait()
	}
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

func (c *OperatorCache) SyncOperatorUpdate(in *Operator, rev int64) {
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

func (c *OperatorCache) SyncOperatorDelete(in *Operator, rev int64) {
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

func (c *OperatorCache) SyncOperatorPrune(current map[OperatorKey]struct{}) {
	deleted := make(map[OperatorKey]struct{})
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

func (c *OperatorCache) SyncOperatorRevOnly(rev int64) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	c.Rev = rev
	util.DebugLog(util.DebugLevelApi, "SyncRevOnly", "rev", rev)
	c.Cond.Broadcast()
}

func (c *OperatorCache) SyncWait(rev int64) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	util.DebugLog(util.DebugLevelApi, "SyncWait", "cache-rev", c.Rev, "wait-rev", rev)
	for c.Rev < rev {
		c.Cond.Wait()
	}
}

func (c *OperatorCache) Show(filter *Operator, cb func(ret *Operator) error) error {
	util.DebugLog(util.DebugLevelApi, "Show Operator", "count", len(c.Objs))
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for _, obj := range c.Objs {
		if !obj.Matches(filter) {
			continue
		}
		util.DebugLog(util.DebugLevelApi, "Show Operator", "obj", obj)
		err := cb(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *OperatorCache) SetNotifyCb(fn func(obj *OperatorKey)) {
	c.notifyCb = fn
}

func (m *Operator) GetKey() *OperatorKey {
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

func init() { proto.RegisterFile("operator.proto", fileDescriptorOperator) }

var fileDescriptorOperator = []byte{
	// 402 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x91, 0xbf, 0x4a, 0xf3, 0x50,
	0x18, 0xc6, 0x7b, 0xda, 0x52, 0xbe, 0x9e, 0x86, 0xd2, 0x2f, 0xe5, 0x2b, 0xf9, 0xaa, 0xa4, 0x92,
	0x49, 0x84, 0xe6, 0x48, 0x5d, 0x4a, 0x37, 0x1b, 0x37, 0x51, 0x21, 0xd2, 0xdd, 0xa4, 0x39, 0x4d,
	0x83, 0x69, 0xde, 0x90, 0x3f, 0xd4, 0x6e, 0xe2, 0x2d, 0x78, 0x03, 0x5e, 0x82, 0x97, 0xd1, 0x51,
	0x70, 0x17, 0x2d, 0x0e, 0xba, 0x09, 0x75, 0x70, 0x94, 0x9c, 0xa4, 0x35, 0x94, 0xe2, 0xd2, 0x25,
	0x3c, 0xef, 0xcb, 0xf3, 0xfc, 0x9e, 0x97, 0x13, 0x5c, 0x06, 0x97, 0x7a, 0x5a, 0x00, 0x9e, 0xec,
	0x7a, 0x10, 0x00, 0x5f, 0xa4, 0x86, 0x49, 0x99, 0xac, 0x6f, 0x9b, 0x00, 0xa6, 0x4d, 0x89, 0xe6,
	0x5a, 0x44, 0x73, 0x1c, 0x08, 0xb4, 0xc0, 0x02, 0xc7, 0x8f, 0x8d, 0xf5, 0xb6, 0x69, 0x05, 0xc3,
	0x50, 0x97, 0xfb, 0x30, 0x22, 0x23, 0xd0, 0x2d, 0x3b, 0x0a, 0x5e, 0x91, 0xe8, 0xdb, 0xec, 0xdb,
	0x10, 0x1a, 0x84, 0xf9, 0x4c, 0xea, 0x2c, 0x45, 0x92, 0xe4, 0x3c, 0xea, 0x87, 0x76, 0x90, 0x4c,
	0xcd, 0x14, 0xc7, 0x04, 0x13, 0x62, 0xb7, 0x1e, 0x0e, 0xd8, 0xc4, 0x06, 0xa6, 0x62, 0xbb, 0xd4,
	0xc6, 0xdc, 0x59, 0x72, 0xb1, 0x02, 0x06, 0xe5, 0x2b, 0x38, 0x77, 0x72, 0xaa, 0x08, 0x68, 0x07,
	0xed, 0x16, 0xd5, 0x48, 0xb2, 0x8d, 0xa2, 0x08, 0xd9, 0x64, 0xa3, 0x28, 0x9d, 0xfc, 0xdb, 0x5c,
	0x40, 0x12, 0xc1, 0xa5, 0x45, 0xf2, 0x98, 0x4e, 0x78, 0x1e, 0xe7, 0x1d, 0x6d, 0x44, 0x93, 0x24,
	0xd3, 0x1d, 0x2e, 0x32, 0x7e, 0xcd, 0x05, 0x74, 0x7f, 0xd7, 0x40, 0xd2, 0x05, 0xfe, 0xb3, 0x08,
	0xf0, 0x35, 0x5c, 0x18, 0x58, 0xd4, 0x36, 0x7c, 0xe6, 0xe7, 0xd4, 0x64, 0xe2, 0x65, 0x9c, 0xbb,
	0xa4, 0x13, 0x56, 0x56, 0x6a, 0xd5, 0xe4, 0xe5, 0xe3, 0xc9, 0xa9, 0xaa, 0x6e, 0x7e, 0xfa, 0xd4,
	0xc8, 0xa8, 0x91, 0x31, 0x6e, 0xf8, 0x98, 0x0b, 0xe8, 0xfa, 0x53, 0x40, 0xad, 0xf7, 0xec, 0xcf,
	0x4d, 0x87, 0xae, 0xc5, 0xf7, 0x70, 0x59, 0xf1, 0xa8, 0x16, 0xd0, 0x65, 0x6f, 0x75, 0x0d, 0xb2,
	0xfe, 0x37, 0xb5, 0x54, 0xd9, 0x5b, 0x4a, 0x5b, 0x37, 0x8f, 0xaf, 0xb7, 0xd9, 0x7f, 0x52, 0x85,
	0xf4, 0x19, 0x80, 0x2c, 0x7e, 0x6b, 0x07, 0xed, 0x45, 0xd8, 0x23, 0x6a, 0xd3, 0x8d, 0xb0, 0x06,
	0x03, 0xac, 0x62, 0x7b, 0xae, 0xb1, 0xd9, 0xb5, 0x21, 0x03, 0xac, 0x60, 0xb9, 0xf3, 0x21, 0x8c,
	0x7f, 0x87, 0xae, 0x5b, 0x4a, 0xff, 0x19, 0xb6, 0x2a, 0x95, 0x89, 0x3f, 0x84, 0x71, 0x1a, 0xba,
	0x8f, 0xba, 0x95, 0xe9, 0x8b, 0x98, 0x99, 0xce, 0x44, 0xf4, 0x30, 0x13, 0xd1, 0xf3, 0x4c, 0x44,
	0x7a, 0x81, 0xc5, 0x0f, 0xbe, 0x03, 0x00, 0x00, 0xff, 0xff, 0xa5, 0x59, 0x31, 0x89, 0x03, 0x03,
	0x00, 0x00,
}
