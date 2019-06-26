// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: controller.proto

package edgeproto

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/gogo/protobuf/gogoproto"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/protocmd"

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

// ControllerKey uniquely defines a Controller
type ControllerKey struct {
	// external API address
	Addr string `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
}

func (m *ControllerKey) Reset()                    { *m = ControllerKey{} }
func (m *ControllerKey) String() string            { return proto.CompactTextString(m) }
func (*ControllerKey) ProtoMessage()               {}
func (*ControllerKey) Descriptor() ([]byte, []int) { return fileDescriptorController, []int{0} }

// A Controller is a service that manages the edge-cloud data and controls other edge-cloud micro-services.
type Controller struct {
	// Fields are used for the Update API to specify which fields to apply
	Fields []string `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	// Unique identifier key
	Key ControllerKey `protobuf:"bytes,2,opt,name=key" json:"key"`
}

func (m *Controller) Reset()                    { *m = Controller{} }
func (m *Controller) String() string            { return proto.CompactTextString(m) }
func (*Controller) ProtoMessage()               {}
func (*Controller) Descriptor() ([]byte, []int) { return fileDescriptorController, []int{1} }

func init() {
	proto.RegisterType((*ControllerKey)(nil), "edgeproto.ControllerKey")
	proto.RegisterType((*Controller)(nil), "edgeproto.Controller")
}
func (this *ControllerKey) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 5)
	s = append(s, "&edgeproto.ControllerKey{")
	s = append(s, "Addr: "+fmt.Sprintf("%#v", this.Addr)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringController(v interface{}, typ string) string {
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

// Client API for ControllerApi service

type ControllerApiClient interface {
	// Show Controllers
	ShowController(ctx context.Context, in *Controller, opts ...grpc.CallOption) (ControllerApi_ShowControllerClient, error)
}

type controllerApiClient struct {
	cc *grpc.ClientConn
}

func NewControllerApiClient(cc *grpc.ClientConn) ControllerApiClient {
	return &controllerApiClient{cc}
}

func (c *controllerApiClient) ShowController(ctx context.Context, in *Controller, opts ...grpc.CallOption) (ControllerApi_ShowControllerClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_ControllerApi_serviceDesc.Streams[0], c.cc, "/edgeproto.ControllerApi/ShowController", opts...)
	if err != nil {
		return nil, err
	}
	x := &controllerApiShowControllerClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ControllerApi_ShowControllerClient interface {
	Recv() (*Controller, error)
	grpc.ClientStream
}

type controllerApiShowControllerClient struct {
	grpc.ClientStream
}

func (x *controllerApiShowControllerClient) Recv() (*Controller, error) {
	m := new(Controller)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for ControllerApi service

type ControllerApiServer interface {
	// Show Controllers
	ShowController(*Controller, ControllerApi_ShowControllerServer) error
}

func RegisterControllerApiServer(s *grpc.Server, srv ControllerApiServer) {
	s.RegisterService(&_ControllerApi_serviceDesc, srv)
}

func _ControllerApi_ShowController_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Controller)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ControllerApiServer).ShowController(m, &controllerApiShowControllerServer{stream})
}

type ControllerApi_ShowControllerServer interface {
	Send(*Controller) error
	grpc.ServerStream
}

type controllerApiShowControllerServer struct {
	grpc.ServerStream
}

func (x *controllerApiShowControllerServer) Send(m *Controller) error {
	return x.ServerStream.SendMsg(m)
}

var _ControllerApi_serviceDesc = grpc.ServiceDesc{
	ServiceName: "edgeproto.ControllerApi",
	HandlerType: (*ControllerApiServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ShowController",
			Handler:       _ControllerApi_ShowController_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "controller.proto",
}

func (m *ControllerKey) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ControllerKey) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Addr) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintController(dAtA, i, uint64(len(m.Addr)))
		i += copy(dAtA[i:], m.Addr)
	}
	return i, nil
}

func (m *Controller) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Controller) MarshalTo(dAtA []byte) (int, error) {
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
	i = encodeVarintController(dAtA, i, uint64(m.Key.Size()))
	n1, err := m.Key.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	return i, nil
}

func encodeVarintController(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *ControllerKey) Matches(o *ControllerKey, fopts ...MatchOpt) bool {
	opts := MatchOptions{}
	applyMatchOptions(&opts, fopts...)
	if o == nil {
		if opts.Filter {
			return true
		}
		return false
	}
	if !opts.Filter || o.Addr != "" {
		if o.Addr != m.Addr {
			return false
		}
	}
	return true
}

func (m *ControllerKey) CopyInFields(src *ControllerKey) {
	m.Addr = src.Addr
}

func (m *ControllerKey) GetKeyString() string {
	key, err := json.Marshal(m)
	if err != nil {
		log.FatalLog("Failed to marshal ControllerKey key string", "obj", m)
	}
	return string(key)
}

func ControllerKeyStringParse(str string, key *ControllerKey) {
	err := json.Unmarshal([]byte(str), key)
	if err != nil {
		log.FatalLog("Failed to unmarshal ControllerKey key string", "str", str)
	}
}

// Helper method to check that enums have valid values
func (m *ControllerKey) ValidateEnums() error {
	return nil
}

func (m *Controller) Matches(o *Controller, fopts ...MatchOpt) bool {
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

const ControllerFieldKey = "2"
const ControllerFieldKeyAddr = "2.1"

var ControllerAllFields = []string{
	ControllerFieldKeyAddr,
}

var ControllerAllFieldsMap = map[string]struct{}{
	ControllerFieldKeyAddr: struct{}{},
}

var ControllerAllFieldsStringMap = map[string]string{
	ControllerFieldKeyAddr: "Controller Field Key Addr",
}

func (m *Controller) DiffFields(o *Controller, fields map[string]struct{}) {
	if m.Key.Addr != o.Key.Addr {
		fields[ControllerFieldKeyAddr] = struct{}{}
		fields[ControllerFieldKey] = struct{}{}
	}
}

func (m *Controller) CopyInFields(src *Controller) {
	fmap := MakeFieldMap(src.Fields)
	if _, set := fmap["2"]; set {
		if _, set := fmap["2.1"]; set {
			m.Key.Addr = src.Key.Addr
		}
	}
}

func (s *Controller) HasFields() bool {
	return true
}

type ControllerStore struct {
	kvstore objstore.KVStore
}

func NewControllerStore(kvstore objstore.KVStore) ControllerStore {
	return ControllerStore{kvstore: kvstore}
}

func (s *ControllerStore) Create(m *Controller, wait func(int64)) (*Result, error) {
	err := m.Validate(ControllerAllFieldsMap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Controller", m.GetKey())
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

func (s *ControllerStore) Update(m *Controller, wait func(int64)) (*Result, error) {
	fmap := MakeFieldMap(m.Fields)
	err := m.Validate(fmap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Controller", m.GetKey())
	var vers int64 = 0
	curBytes, vers, _, err := s.kvstore.Get(key)
	if err != nil {
		return nil, err
	}
	var cur Controller
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

func (s *ControllerStore) Put(m *Controller, wait func(int64), ops ...objstore.KVOp) (*Result, error) {
	fmap := MakeFieldMap(m.Fields)
	err := m.Validate(fmap)
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Controller", m.GetKey())
	var val []byte
	curBytes, _, _, err := s.kvstore.Get(key)
	if err == nil {
		var cur Controller
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

func (s *ControllerStore) Delete(m *Controller, wait func(int64)) (*Result, error) {
	err := m.GetKey().Validate()
	if err != nil {
		return nil, err
	}
	key := objstore.DbKeyString("Controller", m.GetKey())
	rev, err := s.kvstore.Delete(key)
	if err != nil {
		return nil, err
	}
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *ControllerStore) LoadOne(key string) (*Controller, int64, error) {
	val, rev, _, err := s.kvstore.Get(key)
	if err != nil {
		return nil, 0, err
	}
	var obj Controller
	err = json.Unmarshal(val, &obj)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "Failed to parse Controller data", "val", string(val))
		return nil, 0, err
	}
	return &obj, rev, nil
}

func (s *ControllerStore) STMGet(stm concurrency.STM, key *ControllerKey, buf *Controller) bool {
	keystr := objstore.DbKeyString("Controller", key)
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

func (s *ControllerStore) STMPut(stm concurrency.STM, obj *Controller, ops ...objstore.KVOp) {
	keystr := objstore.DbKeyString("Controller", obj.GetKey())
	val, err := json.Marshal(obj)
	if err != nil {
		log.InfoLog("Controller json marsahal failed", "obj", obj, "err", err)
	}
	v3opts := GetSTMOpts(ops...)
	stm.Put(keystr, string(val), v3opts...)
}

func (s *ControllerStore) STMDel(stm concurrency.STM, key *ControllerKey) {
	keystr := objstore.DbKeyString("Controller", key)
	stm.Del(keystr)
}

type ControllerKeyWatcher struct {
	cb func()
}

// ControllerCache caches Controller objects in memory in a hash table
// and keeps them in sync with the database.
type ControllerCache struct {
	Objs        map[ControllerKey]*Controller
	Mux         util.Mutex
	List        map[ControllerKey]struct{}
	NotifyCb    func(obj *ControllerKey, old *Controller)
	UpdatedCb   func(old *Controller, new *Controller)
	KeyWatchers map[ControllerKey][]*ControllerKeyWatcher
}

func NewControllerCache() *ControllerCache {
	cache := ControllerCache{}
	InitControllerCache(&cache)
	return &cache
}

func InitControllerCache(cache *ControllerCache) {
	cache.Objs = make(map[ControllerKey]*Controller)
	cache.KeyWatchers = make(map[ControllerKey][]*ControllerKeyWatcher)
}

func (c *ControllerCache) GetTypeString() string {
	return "Controller"
}

func (c *ControllerCache) Get(key *ControllerKey, valbuf *Controller) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	inst, found := c.Objs[*key]
	if found {
		*valbuf = *inst
	}
	return found
}

func (c *ControllerCache) HasKey(key *ControllerKey) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	_, found := c.Objs[*key]
	return found
}

func (c *ControllerCache) GetAllKeys(keys map[ControllerKey]struct{}) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for key, _ := range c.Objs {
		keys[key] = struct{}{}
	}
}

func (c *ControllerCache) Update(in *Controller, rev int64) {
	c.UpdateModFunc(&in.Key, rev, func(old *Controller) (*Controller, bool) {
		return in, true
	})
}

func (c *ControllerCache) UpdateModFunc(key *ControllerKey, rev int64, modFunc func(old *Controller) (new *Controller, changed bool)) {
	c.Mux.Lock()
	old := c.Objs[*key]
	new, changed := modFunc(old)
	if !changed {
		c.Mux.Unlock()
		return
	}
	if c.UpdatedCb != nil || c.NotifyCb != nil {
		if c.UpdatedCb != nil {
			newCopy := &Controller{}
			*newCopy = *new
			defer c.UpdatedCb(old, newCopy)
		}
		if c.NotifyCb != nil {
			defer c.NotifyCb(&new.Key, old)
		}
	}
	c.Objs[new.Key] = new
	log.DebugLog(log.DebugLevelApi, "SyncUpdate Controller", "obj", new, "rev", rev)
	c.Mux.Unlock()
	c.TriggerKeyWatchers(&new.Key)
}

func (c *ControllerCache) Delete(in *Controller, rev int64) {
	c.Mux.Lock()
	old := c.Objs[in.Key]
	delete(c.Objs, in.Key)
	log.DebugLog(log.DebugLevelApi, "SyncDelete Controller", "key", in.Key, "rev", rev)
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		c.NotifyCb(&in.Key, old)
	}
	c.TriggerKeyWatchers(&in.Key)
}

func (c *ControllerCache) Prune(validKeys map[ControllerKey]struct{}) {
	notify := make(map[ControllerKey]*Controller)
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

func (c *ControllerCache) GetCount() int {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	return len(c.Objs)
}

func (c *ControllerCache) Flush(notifyId int64) {
}

func (c *ControllerCache) Show(filter *Controller, cb func(ret *Controller) error) error {
	log.DebugLog(log.DebugLevelApi, "Show Controller", "count", len(c.Objs))
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for _, obj := range c.Objs {
		if !obj.Matches(filter, MatchFilter()) {
			continue
		}
		log.DebugLog(log.DebugLevelApi, "Show Controller", "obj", obj)
		err := cb(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func ControllerGenericNotifyCb(fn func(key *ControllerKey, old *Controller)) func(objstore.ObjKey, objstore.Obj) {
	return func(objkey objstore.ObjKey, obj objstore.Obj) {
		fn(objkey.(*ControllerKey), obj.(*Controller))
	}
}

func (c *ControllerCache) SetNotifyCb(fn func(obj *ControllerKey, old *Controller)) {
	c.NotifyCb = fn
}

func (c *ControllerCache) SetUpdatedCb(fn func(old *Controller, new *Controller)) {
	c.UpdatedCb = fn
}

func (c *ControllerCache) WatchKey(key *ControllerKey, cb func()) context.CancelFunc {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	list, ok := c.KeyWatchers[*key]
	if !ok {
		list = make([]*ControllerKeyWatcher, 0)
	}
	watcher := ControllerKeyWatcher{cb: cb}
	c.KeyWatchers[*key] = append(list, &watcher)
	log.DebugLog(log.DebugLevelApi, "Watching Controller", "key", key)
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

func (c *ControllerCache) TriggerKeyWatchers(key *ControllerKey) {
	watchers := make([]*ControllerKeyWatcher, 0)
	c.Mux.Lock()
	if list, ok := c.KeyWatchers[*key]; ok {
		watchers = append(watchers, list...)
	}
	c.Mux.Unlock()
	for ii, _ := range watchers {
		watchers[ii].cb()
	}
}
func (c *ControllerCache) SyncUpdate(key, val []byte, rev int64) {
	obj := Controller{}
	err := json.Unmarshal(val, &obj)
	if err != nil {
		log.WarnLog("Failed to parse Controller data", "val", string(val))
		return
	}
	c.Update(&obj, rev)
	c.Mux.Lock()
	if c.List != nil {
		c.List[obj.Key] = struct{}{}
	}
	c.Mux.Unlock()
}

func (c *ControllerCache) SyncDelete(key []byte, rev int64) {
	obj := Controller{}
	keystr := objstore.DbKeyPrefixRemove(string(key))
	ControllerKeyStringParse(keystr, &obj.Key)
	c.Delete(&obj, rev)
}

func (c *ControllerCache) SyncListStart() {
	c.List = make(map[ControllerKey]struct{})
}

func (c *ControllerCache) SyncListEnd() {
	deleted := make(map[ControllerKey]*Controller)
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

func (m *Controller) GetKey() objstore.ObjKey {
	return &m.Key
}

func CmpSortController(a Controller, b Controller) bool {
	return a.Key.GetKeyString() < b.Key.GetKeyString()
}

// Helper method to check that enums have valid values
// NOTE: ValidateEnums checks all Fields even if some are not set
func (m *Controller) ValidateEnums() error {
	if err := m.Key.ValidateEnums(); err != nil {
		return err
	}
	return nil
}

func (m *ControllerKey) Size() (n int) {
	var l int
	_ = l
	l = len(m.Addr)
	if l > 0 {
		n += 1 + l + sovController(uint64(l))
	}
	return n
}

func (m *Controller) Size() (n int) {
	var l int
	_ = l
	if len(m.Fields) > 0 {
		for _, s := range m.Fields {
			l = len(s)
			n += 1 + l + sovController(uint64(l))
		}
	}
	l = m.Key.Size()
	n += 1 + l + sovController(uint64(l))
	return n
}

func sovController(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozController(x uint64) (n int) {
	return sovController(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ControllerKey) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowController
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
			return fmt.Errorf("proto: ControllerKey: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ControllerKey: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Addr", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowController
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
				return ErrInvalidLengthController
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Addr = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipController(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthController
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
func (m *Controller) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowController
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
			return fmt.Errorf("proto: Controller: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Controller: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Fields", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowController
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
				return ErrInvalidLengthController
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
					return ErrIntOverflowController
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
				return ErrInvalidLengthController
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
			skippy, err := skipController(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthController
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
func skipController(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowController
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
					return 0, ErrIntOverflowController
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
					return 0, ErrIntOverflowController
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
				return 0, ErrInvalidLengthController
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowController
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
				next, err := skipController(dAtA[start:])
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
	ErrInvalidLengthController = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowController   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("controller.proto", fileDescriptorController) }

var fileDescriptorController = []byte{
	// 324 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x48, 0xce, 0xcf, 0x2b,
	0x29, 0xca, 0xcf, 0xc9, 0x49, 0x2d, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x4c, 0x4d,
	0x49, 0x4f, 0x05, 0x33, 0xa5, 0x64, 0xd2, 0xf3, 0xf3, 0xd3, 0x73, 0x52, 0xf5, 0x13, 0x0b, 0x32,
	0xf5, 0x13, 0xf3, 0xf2, 0xf2, 0x4b, 0x12, 0x4b, 0x32, 0xf3, 0xf3, 0x8a, 0x21, 0x0a, 0xa5, 0x44,
	0xd2, 0xf3, 0xd3, 0xf3, 0xc1, 0x4c, 0x7d, 0x10, 0x0b, 0x2a, 0x6a, 0x91, 0x9e, 0x59, 0x92, 0x51,
	0x9a, 0xa4, 0x97, 0x9c, 0x9f, 0xab, 0x9f, 0x9b, 0x9f, 0x94, 0x99, 0x03, 0x32, 0xae, 0x42, 0x1f,
	0x44, 0xea, 0x26, 0xe7, 0xe4, 0x97, 0xa6, 0xe8, 0x83, 0xd5, 0xa5, 0xa7, 0xe6, 0xc1, 0x19, 0x50,
	0x9d, 0xee, 0xc4, 0xe9, 0x4c, 0xd6, 0x4d, 0x4f, 0xcd, 0xd3, 0x4d, 0xce, 0x85, 0x71, 0x91, 0x18,
	0x10, 0x83, 0x94, 0x0c, 0xb9, 0x78, 0x9d, 0xe1, 0xbe, 0xf2, 0x4e, 0xad, 0x14, 0x12, 0xe2, 0x62,
	0x49, 0x4c, 0x49, 0x29, 0x92, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x0c, 0x02, 0xb3, 0xad, 0x78, 0x5e,
	0x7c, 0x96, 0x60, 0xfc, 0xf1, 0x59, 0x82, 0x71, 0xc3, 0x02, 0x79, 0x46, 0xa5, 0x14, 0x2e, 0x2e,
	0x84, 0x16, 0x21, 0x31, 0x2e, 0xb6, 0xb4, 0xcc, 0xd4, 0x9c, 0x94, 0x62, 0x09, 0x46, 0x05, 0x66,
	0x0d, 0xce, 0x20, 0x28, 0x4f, 0xc8, 0x80, 0x8b, 0x39, 0x3b, 0xb5, 0x52, 0x82, 0x49, 0x81, 0x51,
	0x83, 0xdb, 0x48, 0x42, 0x0f, 0x1e, 0x50, 0x7a, 0x28, 0xd6, 0x39, 0xb1, 0x9c, 0xb8, 0x27, 0xcf,
	0x10, 0x04, 0x52, 0x0a, 0xb1, 0xe5, 0xc3, 0x67, 0x09, 0xc6, 0x86, 0x2f, 0x12, 0x8c, 0x46, 0x79,
	0xc8, 0x0e, 0x73, 0x2c, 0xc8, 0x14, 0x8a, 0xe5, 0xe2, 0x0b, 0xce, 0xc8, 0x2f, 0x47, 0xb2, 0x5a,
	0x14, 0xab, 0xa9, 0x52, 0xd8, 0x85, 0x95, 0xa4, 0x9b, 0x2e, 0x3f, 0x99, 0xcc, 0x24, 0xaa, 0x24,
	0xa0, 0x5f, 0x9c, 0x91, 0x5f, 0xae, 0x8f, 0x88, 0x4b, 0x2b, 0x46, 0x2d, 0x03, 0x46, 0x27, 0x81,
	0x13, 0x0f, 0xe5, 0x18, 0x4e, 0x3c, 0x92, 0x63, 0xbc, 0xf0, 0x48, 0x8e, 0xf1, 0xc1, 0x23, 0x39,
	0xc6, 0x24, 0x36, 0xb0, 0x11, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x4b, 0x4c, 0x35, 0xd4,
	0xf7, 0x01, 0x00, 0x00,
}
