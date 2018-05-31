package mexgen

import (
	"fmt"
	"text/template"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/protogen"
)

func RegisterMex() {
	generator.RegisterPlugin(new(mex))
}

func init() {
	generator.RegisterPlugin(new(mex))
}

type mex struct {
	gen         *generator.Generator
	msgs        map[string]*descriptor.DescriptorProto
	cudTemplate *template.Template
}

func (m *mex) Name() string {
	return "mex"
}

func (m *mex) Init(gen *generator.Generator) {
	m.gen = gen
	m.msgs = make(map[string]*descriptor.DescriptorProto)
	m.cudTemplate = template.Must(template.New("cud").Parse(cudTemplateIn))
}

// P forwards to g.gen.P
func (m *mex) P(args ...interface{}) {
	m.gen.P(args...)
}

func (m *mex) Generate(file *generator.FileDescriptor) {
	for _, desc := range file.Messages() {
		m.generateMessage(file, desc)
	}
	if len(file.FileDescriptorProto.Service) != 0 {
		for _, service := range file.FileDescriptorProto.Service {
			m.generateService(file, service)
		}
	}
}

func (m *mex) GenerateImports(file *generator.FileDescriptor) {
	hasGenerateCud := false
	hasGrpcFields := false
	hasGenerateCache := false
	for _, desc := range file.Messages() {
		msg := desc.DescriptorProto
		if GetGenerateCud(msg) {
			hasGenerateCud = true
			if GetGenerateCache(msg) {
				hasGenerateCache = true
			}
		}
		if HasGrpcFields(msg) {
			hasGrpcFields = true
		}
		m.msgs[*msg.Name] = msg
	}
	if hasGenerateCud {
		m.P("import \"encoding/json\"")
		m.P("import \"github.com/mobiledgex/edge-cloud/objstore\"")
	}
	if hasGenerateCache {
		m.P("import \"sync\"")
	}
	if hasGrpcFields {
		m.P("import \"github.com/mobiledgex/edge-cloud/util\"")
	}
}

func (m *mex) generateFieldMatches(message *descriptor.DescriptorProto, field *descriptor.FieldDescriptorProto) {
	if field.Type == nil {
		return
	}
	name := generator.CamelCase(*field.Name)
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		nullcheck := ""
		ref := "&"
		if gogoproto.IsNullable(field) {
			nullcheck = fmt.Sprintf("filter.%s != nil && m.%s != nil && ", name, name)
			ref = ""
		}
		if *field.TypeName == ".google.protobuf.Timestamp" {
			m.P("if ", nullcheck, "(m.", name, ".Seconds != filter.", name, ".Seconds || m.", name, ".Nanos != filter.", name, ".Nanos) {")
		} else {
			m.P("if ", nullcheck, "!m.", name, ".Matches(", ref, "filter.", name, ") {")
		}
		m.P("return false")
		m.P("}")
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		// deprecated in proto3
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		// TODO
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		m.P("if filter.", name, " != \"\" && filter.", name, " != m.", name, "{")
		m.P("return false")
		m.P("}")
	default:
		m.P("if filter.", name, " != 0 && filter.", name, " != m.", name, "{")
		m.P("return false")
		m.P("}")
	}
}

func (m *mex) printCopyInMakeArray(desc *generator.Descriptor, field *descriptor.FieldDescriptorProto) {
	name := generator.CamelCase(*field.Name)
	typ, _ := m.gen.GoType(desc, field)
	m.P("if m.", name, " == nil {")
	m.P("m.", name, " = make(", typ, ", len(src.", name, "))")
	m.P("}")
}

func (m *mex) getFieldDesc(field *descriptor.FieldDescriptorProto) *generator.Descriptor {
	obj := m.gen.ObjectNamed(field.GetTypeName())
	if obj == nil {
		return nil
	}
	desc, ok := obj.(*generator.Descriptor)
	if ok {
		return desc
	}
	return nil
}

func (m *mex) generateFieldCopyIn(desc *generator.Descriptor, ii int, field *descriptor.FieldDescriptorProto) {
	message := desc.DescriptorProto
	if ii == 0 && *field.Name == "fields" {
		// skip fields
		return
	}
	if field.OneofIndex != nil {
		// no support for copy OneOf fields
		return
	}
	name := generator.CamelCase(*field.Name)
	if HasGrpcFields(message) {
		m.P("if set, _ := util.GrpcFieldsGet(src.Fields, ", field.Number, "); set == true {")
	}
	if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		m.printCopyInMakeArray(desc, field)
	}
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		fieldDesc := m.getFieldDesc(field)
		msgtyp := m.gen.TypeName(fieldDesc)
		idx := ""
		if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			idx = "[ii]"
			m.P("for ii := 0; ii < len(m.", name, ") && ii < len(src.", name, "); ii++ {")
		}
		if gogoproto.IsNullable(field) {
			m.P("if src.", name, idx, " != nil {")
			m.P("if m.", name, idx, " == nil {")
			m.P("m.", name, idx, " = &", msgtyp, "{}")
			m.P("}")
		}
		if fieldDesc == nil || !HasGrpcFields(fieldDesc.DescriptorProto) {
			if gogoproto.IsNullable(field) {
				m.P("*m.", name, idx, " = *src.", name, idx)
			} else {
				m.P("m.", name, idx, " = src.", name, idx)
			}
		} else {
			if gogoproto.IsNullable(field) {
				m.P("m.", name, idx, ".CopyInFields(src.", name, idx, ")")
			} else {
				m.P("m.", name, idx, ".CopyInFields(&src.", name, idx, ")")
			}
		}
		if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			m.P("}")
		}
		if gogoproto.IsNullable(field) {
			m.P("}")
		}
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		// deprecated in proto3
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		m.printCopyInMakeArray(desc, field)
		m.P("copy(m.", name, ", src.", name, ")")
	default:
		if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			m.P("copy(m.", name, ", src.", name, ")")
		} else {
			m.P("m.", name, " = src.", name)
		}
	}
	if HasGrpcFields(message) {
		m.P("}")
	}
}

type cudTemplateArgs struct {
	Name      string
	KeyName   string
	CudName   string
	HasFields bool
	GenCache  bool
}

var cudTemplateIn = `
func (s *{{.Name}}) HasFields() bool {
{{- if (.HasFields)}}
	return true
{{- else}}
	return false
{{- end}}
}

type {{.Name}}Store struct {
	objstore objstore.ObjStore
	list{{.Name}} map[{{.Name}}Key]struct{}
}

func New{{.Name}}Store(objstore objstore.ObjStore) {{.Name}}Store {
	return {{.Name}}Store{objstore: objstore}
}

type {{.Name}}Cacher interface {
	Sync{{.Name}}Update(m *{{.Name}}, rev int64)
	Sync{{.Name}}Delete(m *{{.Name}}, rev int64)
	Sync{{.Name}}Prune(current map[{{.Name}}Key]struct{})
	Sync{{.Name}}RevOnly(rev int64)
}

func (s *{{.Name}}Store) Create(m *{{.Name}}, wait func(int64)) (*Result, error) {
	err := m.Validate()
	if err != nil { return nil, err }
	key := objstore.DbKeyString(m.GetKey())
	val, err := json.Marshal(m)
	if err != nil { return nil, err }
	rev, err := s.objstore.Create(key, string(val))
	if err != nil { return nil, err }
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *{{.Name}}Store) Update(m *{{.Name}}, wait func(int64)) (*Result, error) {
	err := m.Validate()
	if err != nil { return nil, err }
	key := objstore.DbKeyString(m.GetKey())
	var vers int64 = 0
{{- if (.HasFields)}}
	curBytes, vers, err := s.objstore.Get(key)
	if err != nil { return nil, err }
	var cur {{.Name}}
	err = json.Unmarshal(curBytes, &cur)
	if err != nil { return nil, err }
	cur.CopyInFields(m)
	// never save fields
	cur.Fields = nil
	val, err := json.Marshal(cur)
{{- else}}
	val, err := json.Marshal(m)
{{- end}}
	if err != nil { return nil, err }
	rev, err := s.objstore.Update(key, string(val), vers)
	if err != nil { return nil, err }
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *{{.Name}}Store) Delete(m *{{.Name}}, wait func(int64)) (*Result, error) {
	err := m.GetKey().Validate()
	if err != nil { return nil, err }
	key := objstore.DbKeyString(m.GetKey())
	rev, err := s.objstore.Delete(key)
	if err != nil { return nil, err }
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

type {{.Name}}Cb func(m *{{.Name}}) error

func (s *{{.Name}}Store) LoadAll(cb {{.Name}}Cb) error {
	loadkey := objstore.DbKeyPrefixString(&{{.Name}}Key{})
	err := s.objstore.List(loadkey, func(key, val []byte, rev int64) error {
		var obj {{.Name}}
		err := json.Unmarshal(val, &obj)
		if err != nil {
			util.WarnLog("Failed to parse {{.Name}} data", "val", string(val))
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

func (s *{{.Name}}Store) LoadOne(key string) (*{{.Name}}, int64, error) {
	val, rev, err := s.objstore.Get(key)
	if err != nil {
		return nil, 0, err
	}
	var obj {{.Name}}
	err = json.Unmarshal(val, &obj)
	if err != nil {
		util.DebugLog(util.DebugLevelApi, "Failed to parse {{.Name}} data", "val", string(val))
		return nil, 0, err
	}
	return &obj, rev, nil
}

// Sync will sync changes for any {{.Name}} objects.
func (s *{{.Name}}Store) Sync(ctx context.Context, cacher {{.Name}}Cacher) error {
	str := objstore.DbKeyPrefixString(&{{.Name}}Key{})
	return s.objstore.Sync(ctx, str, func(in *objstore.SyncCbData) {
		obj := {{.Name}}{}
		// Even on parse error, we should still call back to keep
		// the revision numbers in sync so no caller hangs on wait.
		action := in.Action
		if action == objstore.SyncUpdate || action == objstore.SyncList {
			err := json.Unmarshal(in.Value, &obj)
			if err != nil {
				util.WarnLog("Failed to parse {{.Name}} data", "val", string(in.Value))
				action = objstore.SyncRevOnly
			}
		} else if action == objstore.SyncDelete {
			keystr := objstore.DbKeyPrefixRemove(string(in.Key))
			{{.Name}}KeyStringParse(keystr, obj.GetKey())
		}
		util.DebugLog(util.DebugLevelApi, "Sync cb", "action", objstore.SyncActionStrs[in.Action], "key", string(in.Key), "value", string(in.Value), "rev", in.Rev)
		switch action {
		case objstore.SyncUpdate:
			cacher.Sync{{.Name}}Update(&obj, in.Rev)
		case objstore.SyncDelete:
			cacher.Sync{{.Name}}Delete(&obj, in.Rev)
		case objstore.SyncListStart:
			s.list{{.Name}} = make(map[{{.Name}}Key]struct{})
		case objstore.SyncList:
			s.list{{.Name}}[obj.Key] = struct{}{}
			cacher.Sync{{.Name}}Update(&obj, in.Rev)
		case objstore.SyncListEnd:
			cacher.Sync{{.Name}}Prune(s.list{{.Name}})
			s.list{{.Name}} = nil
		case objstore.SyncRevOnly:
			cacher.Sync{{.Name}}RevOnly(in.Rev)
		}
	})
}

{{if (.GenCache)}}
// {{.Name}}Cache caches {{.Name}} objects in memory in a hash table
// and keeps them in sync with the database.
type {{.Name}}Cache struct {
	Store *{{.Name}}Store
	Objs map[{{.Name}}Key]*{{.Name}}
	Rev int64
	Mux util.Mutex
	Cond sync.Cond
	initWait bool
	syncDone bool
	syncCancel context.CancelFunc
	notifyCb func(obj *{{.Name}}Key)
}

func New{{.Name}}Cache(store *{{.Name}}Store) *{{.Name}}Cache {
	cache := {{.Name}}Cache{
		Store: store,
		Objs: make(map[{{.Name}}Key]*{{.Name}}),
		initWait: true,
	}
	cache.Mux.InitCond(&cache.Cond)

	ctx, cancel := context.WithCancel(context.Background())
	cache.syncCancel = cancel
	go func() {
		err := cache.Store.Sync(ctx, &cache)
		if err != nil {
			util.WarnLog("{{.Name}} Sync failed", "err", err)
		}
		cache.syncDone = true
		cache.Cond.Broadcast()
	}()
	return &cache
}

func (c *{{.Name}}Cache) WaitInitSyncDone() {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for c.initWait {
		c.Cond.Wait()
	}
}

func (c *{{.Name}}Cache) Done() {
	c.syncCancel()
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for !c.syncDone {
		c.Cond.Wait()
	}
}

func (c *{{.Name}}Cache) Get(key *{{.Name}}Key, valbuf *{{.Name}}) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	inst, found := c.Objs[*key]
	if found {
		*valbuf = *inst
	}
	return found
}

func (c *{{.Name}}Cache) HasKey(key *{{.Name}}Key) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	_, found := c.Objs[*key]
	return found
}

func (c *{{.Name}}Cache) GetAllKeys(keys map[{{.Name}}Key]struct{}) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for key, _ := range c.Objs {
		keys[key] = struct{}{}
	}
}

func (c *{{.Name}}Cache) Sync{{.Name}}Update(in *{{.Name}}, rev int64) {
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

func (c *{{.Name}}Cache) Sync{{.Name}}Delete(in *{{.Name}}, rev int64) {
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

func (c *{{.Name}}Cache) Sync{{.Name}}Prune(current map[{{.Name}}Key]struct{}) {
	deleted := make(map[{{.Name}}Key]struct{})
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

func (c *{{.Name}}Cache) Sync{{.Name}}RevOnly(rev int64) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	c.Rev = rev
	util.DebugLog(util.DebugLevelApi, "SyncRevOnly", "rev", rev)
	c.Cond.Broadcast()
}

func (c *{{.Name}}Cache) SyncWait(rev int64) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	util.DebugLog(util.DebugLevelApi, "SyncWait", "cache-rev", c.Rev, "wait-rev", rev)
	for c.Rev < rev {
		c.Cond.Wait()
	}
}

func (c *{{.Name}}Cache) Show(filter *{{.Name}}, cb func(ret *{{.Name}}) error) error {
	util.DebugLog(util.DebugLevelApi, "Show {{.Name}}", "count", len(c.Objs))
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for _, obj := range c.Objs {
		if !obj.Matches(filter) {
			continue
		}
		util.DebugLog(util.DebugLevelApi, "Show {{.Name}}", "obj", obj)
		err := cb(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *{{.Name}}Cache) SetNotifyCb(fn func(obj *{{.Name}}Key)) {
	c.notifyCb = fn
}
{{- end}}

`

func (m *mex) generateMessage(file *generator.FileDescriptor, desc *generator.Descriptor) {
	message := desc.DescriptorProto
	if GetGenerateMatches(message) && message.Field != nil {
		m.P("func (m *", message.Name, ") Matches(filter *", message.Name, ") bool {")
		m.P("if filter == nil { return true }")
		for _, field := range message.Field {
			m.generateFieldMatches(message, field)
		}
		m.P("return true")
		m.P("}")
		m.P("")
	}
	if HasGrpcFields(message) {
		for _, field := range message.Field {
			if field.Name == nil || field.Number == nil {
				continue
			}
			name := generator.CamelCase(*field.Name)
			m.P("const ", message.Name, "Field", name, " uint = ", field.Number)
		}
		m.P("")
	}
	msgtyp := m.gen.TypeName(desc)
	m.P("func (m *", msgtyp, ") CopyInFields(src *", msgtyp, ") {")
	for ii, field := range message.Field {
		m.generateFieldCopyIn(desc, ii, field)
	}
	m.P("}")
	m.P("")

	if GetGenerateCud(message) {
		if !HasMessageKey(message) {
			m.gen.Fail("message", *message.Name, "needs a unique key field named key of type", *message.Name+"Key", "for option generate_cud")
		}
		args := cudTemplateArgs{
			Name:      *message.Name,
			CudName:   *message.Name + "Cud",
			KeyName:   *message.Name + "Key",
			HasFields: HasGrpcFields(message),
			GenCache:  GetGenerateCache(message),
		}
		m.cudTemplate.Execute(m.gen.Buffer, args)
	}
	if GetObjKey(message) {
		m.P("func (m *", message.Name, ") GetKeyString() string {")
		m.P("key, err := json.Marshal(m)")
		m.P("if err != nil {")
		m.P("util.FatalLog(\"Failed to marshal ", message.Name, " key string\", \"obj\", m)")
		m.P("}")
		m.P("return string(key)")
		m.P("}")
		m.P("")

		m.P("func ", message.Name, "StringParse(str string, key *", message.Name, ") {")
		m.P("err := json.Unmarshal([]byte(str), key)")
		m.P("if err != nil {")
		m.P("util.FatalLog(\"Failed to unmarshal ", message.Name, " key string\", \"str\", str)")
		m.P("}")
		m.P("}")
		m.P("")
	}
	if HasMessageKey(message) {
		m.P("func (m *", message.Name, ") GetKey() *", message.Name, "Key {")
		m.P("return &m.Key")
		m.P("}")
		m.P("")
	}
}

func (m *mex) generateService(file *generator.FileDescriptor, service *descriptor.ServiceDescriptorProto) {
	if len(service.Method) != 0 {
		for _, method := range service.Method {
			m.generateMethod(file, service, method)
		}
	}
}

func (m *mex) generateMethod(file *generator.FileDescriptor, service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {

}

func HasGrpcFields(message *descriptor.DescriptorProto) bool {
	if message.Field != nil && len(message.Field) > 0 && *message.Field[0].Name == "fields" && *message.Field[0].Type == descriptor.FieldDescriptorProto_TYPE_BYTES {
		return true
	}
	return false
}

func HasMessageKey(message *descriptor.DescriptorProto) bool {
	if message.Field == nil {
		return false
	}
	if len(message.Field) > 0 && *message.Field[0].Name == "key" {
		return true
	}
	if len(message.Field) > 1 && HasGrpcFields(message) && *message.Field[1].Name == "key" {
		return true
	}
	return false
}

func GetGenerateMatches(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateMatches, false)
}

func GetGenerateCud(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateCud, false)
}

func GetGenerateCache(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateCache, false)
}

func GetObjKey(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_ObjKey, false)
}
