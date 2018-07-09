package mexgen

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/gensupport"
	"github.com/mobiledgex/edge-cloud/protogen"
)

func RegisterMex() {
	generator.RegisterPlugin(new(mex))
}

func init() {
	generator.RegisterPlugin(new(mex))
}

type mex struct {
	gen           *generator.Generator
	msgs          map[string]*descriptor.DescriptorProto
	cudTemplate   *template.Template
	enumTemplate  *template.Template
	cacheTemplate *template.Template
	importUtil    bool
	importStrings bool
	importErrors  bool
	importStrconv bool
	support       gensupport.PluginSupport
}

func (m *mex) Name() string {
	return "mex"
}

func (m *mex) Init(gen *generator.Generator) {
	m.gen = gen
	m.msgs = make(map[string]*descriptor.DescriptorProto)
	m.cudTemplate = template.Must(template.New("cud").Parse(cudTemplateIn))
	m.enumTemplate = template.Must(template.New("enum").Parse(enumTemplateIn))
	m.cacheTemplate = template.Must(template.New("cache").Parse(cacheTemplateIn))
	m.support.Init(nil)
}

// P forwards to g.gen.P
func (m *mex) P(args ...interface{}) {
	m.gen.P(args...)
}

func (m *mex) Generate(file *generator.FileDescriptor) {
	m.support.InitFile()
	m.support.SetPbGoPackage(file.GetPackage())
	m.importUtil = false
	m.importStrings = false
	m.importErrors = false
	m.importStrconv = false
	for _, desc := range file.Messages() {
		m.generateMessage(file, desc)
	}
	for _, desc := range file.Enums() {
		m.generateEnum(file, desc)
	}
	if len(file.FileDescriptorProto.Service) != 0 {
		for _, service := range file.FileDescriptorProto.Service {
			m.generateService(file, service)
		}
	}
}

func (m *mex) GenerateImports(file *generator.FileDescriptor) {
	hasGenerateCud := false
	for _, desc := range file.Messages() {
		msg := desc.DescriptorProto
		if GetGenerateCud(msg) {
			hasGenerateCud = true
		}
		m.msgs[*msg.Name] = msg
	}
	if hasGenerateCud {
		m.gen.PrintImport("", "encoding/json")
		m.gen.PrintImport("", "github.com/mobiledgex/edge-cloud/objstore")
	}
	if m.importUtil {
		m.gen.PrintImport("", "github.com/mobiledgex/edge-cloud/util")
		m.gen.PrintImport("", "github.com/mobiledgex/edge-cloud/log")
	}
	if m.importStrings {
		m.gen.PrintImport("strings", "strings")
	}
	if m.importErrors {
		m.gen.PrintImport("", "errors")
	}
	if m.importStrconv {
		m.gen.PrintImport("", "strconv")
	}
	m.support.PrintUsedImports(m.gen)
}

func (m *mex) generateEnum(file *generator.FileDescriptor, desc *generator.EnumDescriptor) {
	en := desc.EnumDescriptorProto
	m.P("var ", en.Name, "Strings = []string{")
	for _, val := range en.Value {
		m.P("\"", val.Name, "\",")
	}
	m.P("}")
	m.P()
	// generate bit map for debug levels
	if len(en.Value) <= 64 {
		m.P("const (")
		for ii, val := range en.Value {
			m.P(en.Name, generator.CamelCase(*val.Name), " uint64 = 1 << ", ii)
		}
		m.P(")")
		m.P()
	}
	args := enumTempl{Name: m.support.FQTypeName(m.gen, desc)}
	m.enumTemplate.Execute(m.gen.Buffer, args)
	m.importErrors = true
	m.importStrconv = true
}

type enumTempl struct {
	Name string
}

var enumTemplateIn = `
func (e *{{.Name}}) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil { return err }
	val, ok := {{.Name}}_value[str]
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = {{.Name}}_name[val]
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("No enum value for %s", str))
	}
	*e = {{.Name}}(val)
	return nil
}

func (e {{.Name}}) MarshalYAML() (interface{}, error) {
	return e.String(), nil
}

`

func (m *mex) generateFieldMatches(message *descriptor.DescriptorProto, field *descriptor.FieldDescriptorProto) {
	if field.Type == nil {
		return
	}
	if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		// TODO: matches support for repeated fields
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
		subDesc := gensupport.GetDesc(m.gen, field.GetTypeName())
		printedCheck := true
		if *field.TypeName == ".google.protobuf.Timestamp" {
			m.P("if ", nullcheck, "(m.", name, ".Seconds != filter.", name, ".Seconds || m.", name, ".Nanos != filter.", name, ".Nanos) {")
		} else if GetGenerateMatches(subDesc.DescriptorProto) {
			m.P("if ", nullcheck, "!m.", name, ".Matches(", ref, "filter.", name, ") {")
		} else {
			printedCheck = false
		}
		if printedCheck {
			m.P("return false")
			m.P("}")
		}
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

func (m *mex) printCopyInMakeArray(name string, desc *generator.Descriptor, field *descriptor.FieldDescriptorProto) {
	typ, _ := m.gen.GoType(desc, field)
	m.P("if m.", name, " == nil || len(m.", name, ") < len(src.", name, ") {")
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

func (m *mex) generateFields(names, nums []string, desc *generator.Descriptor) {
	message := desc.DescriptorProto
	for ii, field := range message.Field {
		if ii == 0 && *field.Name == "fields" {
			continue
		}
		name := generator.CamelCase(*field.Name)
		num := fmt.Sprintf("%d", *field.Number)
		m.P("const ", strings.Join(append(names, name), ""), " = \"", strings.Join(append(nums, num), "."), "\"")
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			subDesc := gensupport.GetDesc(m.gen, field.GetTypeName())
			m.generateFields(append(names, name), append(nums, num), subDesc)
		}
	}
}

type AllFieldsGen int

const (
	AllFieldsGenSlice = iota
	AllFieldsGenMap
)

func (m *mex) generateAllFields(afg AllFieldsGen, names, nums []string, desc *generator.Descriptor) {
	message := desc.DescriptorProto
	for ii, field := range message.Field {
		if ii == 0 && *field.Name == "fields" {
			continue
		}
		name := generator.CamelCase(*field.Name)
		num := fmt.Sprintf("%d", *field.Number)
		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			subDesc := gensupport.GetDesc(m.gen, field.GetTypeName())
			m.generateAllFields(afg, append(names, name), append(nums, num), subDesc)
		default:
			switch afg {
			case AllFieldsGenSlice:
				m.P(strings.Join(append(names, name), ""), ",")
			case AllFieldsGenMap:
				m.P(strings.Join(append(names, name), ""), ": struct{}{},")
			}
		}
	}
}

func (m *mex) generateCopyIn(parents, nums []string, desc *generator.Descriptor, visited []*generator.Descriptor, hasGrpcFields bool) {
	if gensupport.WasVisited(desc, visited) {
		return
	}
	for ii, field := range desc.DescriptorProto.Field {
		if ii == 0 && *field.Name == "fields" {
			continue
		}
		if field.OneofIndex != nil {
			// no support for copy OneOf fields
			continue
		}

		name := generator.CamelCase(*field.Name)
		hierName := strings.Join(append(parents, name), ".")
		num := fmt.Sprintf("%d", *field.Number)
		idx := ""
		nullableMessage := false
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE && gogoproto.IsNullable(field) {
			nullableMessage = true
		}

		if hasGrpcFields {
			numStr := strings.Join(append(nums, num), ".")
			nilCheck := ""
			if nullableMessage {
				nilCheck = " && src." + hierName + " != nil"
			}
			m.P("if _, set := fmap[\"", numStr, "\"]; set", nilCheck, " {")
		} else if nullableMessage {
			m.P("if src.", hierName, " != nil {")
		}
		if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			m.printCopyInMakeArray(hierName, desc, field)
			if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
				incr := fmt.Sprintf("i%d", len(parents))
				m.P("for ", incr, " := 0; ", incr, " < len(src.", hierName, "); ", incr, "++ {")
				idx = "[" + incr + "]"
			}
		}
		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			subDesc := gensupport.GetDesc(m.gen, field.GetTypeName())
			if gogoproto.IsNullable(field) {
				typ := m.support.FQTypeName(m.gen, subDesc)
				m.P("m.", hierName, idx, " = &", typ, "{}")
			}
			m.generateCopyIn(append(parents, name+idx), append(nums, num), subDesc, append(visited, desc), hasGrpcFields)
		case descriptor.FieldDescriptorProto_TYPE_GROUP:
			// deprecated in proto3
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			m.printCopyInMakeArray(hierName, desc, field)
			m.P("copy(m.", hierName, ", src.", hierName, ")")
		default:
			if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				m.P("copy(m.", hierName, ", src.", hierName, ")")
			} else {
				m.P("m.", hierName, " = src.", hierName)
			}
		}
		if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED && *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			m.P("}")
		}
		if hasGrpcFields || nullableMessage {
			m.P("}")
		}
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
}

func New{{.Name}}Store(objstore objstore.ObjStore) {{.Name}}Store {
	return {{.Name}}Store{objstore: objstore}
}

func (s *{{.Name}}Store) Create(m *{{.Name}}, wait func(int64)) (*Result, error) {
	err := m.Validate({{.Name}}AllFieldsMap)
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
	fmap := MakeFieldMap(m.Fields)
	err := m.Validate(fmap)
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
			log.WarnLog("Failed to parse {{.Name}} data", "val", string(val))
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
		log.DebugLog(log.DebugLevelApi, "Failed to parse {{.Name}} data", "val", string(val))
		return nil, 0, err
	}
	return &obj, rev, nil
}
`

type cacheTemplateArgs struct {
	Name        string
	KeyType     string
	CudCache    bool
	NotifyCache bool
}

var cacheTemplateIn = `
// {{.Name}}Cache caches {{.Name}} objects in memory in a hash table
// and keeps them in sync with the database.
type {{.Name}}Cache struct {
	Objs map[{{.KeyType}}]*{{.Name}}
	Mux util.Mutex
	List map[{{.KeyType}}]struct{}
	NotifyCb func(obj *{{.KeyType}})
	UpdatedCb func(old *{{.Name}}, new *{{.Name}})
}

func New{{.Name}}Cache() *{{.Name}}Cache {
	cache := {{.Name}}Cache{}
	Init{{.Name}}Cache(&cache)
	return &cache
}

func Init{{.Name}}Cache(cache *{{.Name}}Cache) {
	cache.Objs = make(map[{{.KeyType}}]*{{.Name}})
}

func (c *{{.Name}}Cache) Get(key *{{.KeyType}}, valbuf *{{.Name}}) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	inst, found := c.Objs[*key]
	if found {
		*valbuf = *inst
	}
	return found
}

func (c *{{.Name}}Cache) HasKey(key *{{.KeyType}}) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	_, found := c.Objs[*key]
	return found
}

func (c *{{.Name}}Cache) GetAllKeys(keys map[{{.KeyType}}]struct{}) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for key, _ := range c.Objs {
		keys[key] = struct{}{}
	}
}

func (c *{{.Name}}Cache) Update(in *{{.Name}}, rev int64) {
	c.Mux.Lock()
	if c.UpdatedCb != nil {
		old := c.Objs[in.Key]
		new := &{{.Name}}{}
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

func (c *{{.Name}}Cache) Delete(in *{{.Name}}, rev int64) {
	c.Mux.Lock()
	delete(c.Objs, in.Key)
	log.DebugLog(log.DebugLevelApi, "SyncUpdate", "key", in.Key, "rev", rev)
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		c.NotifyCb(&in.Key)
	}
}

func (c *{{.Name}}Cache) Prune(validKeys map[{{.KeyType}}]struct{}) {
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

{{- if .NotifyCache}}
func (c *{{.Name}}Cache) Flush(notifyId uint64) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for key, val := range c.Objs {
		if val.NotifyId != notifyId {
			continue
		}
		delete(c.Objs, key)
		if c.NotifyCb != nil {
			c.NotifyCb(&key)
		}
	}
}
{{- end}}

func (c *{{.Name}}Cache) Show(filter *{{.Name}}, cb func(ret *{{.Name}}) error) error {
	log.DebugLog(log.DebugLevelApi, "Show {{.Name}}", "count", len(c.Objs))
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for _, obj := range c.Objs {
{{- if .CudCache}}
		if !obj.Matches(filter) {
			continue
		}
{{- end}}
		log.DebugLog(log.DebugLevelApi, "Show {{.Name}}", "obj", obj)
		err := cb(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *{{.Name}}Cache) SetNotifyCb(fn func(obj *{{.KeyType}})) {
	c.NotifyCb = fn
}

func (c *{{.Name}}Cache) SetUpdatedCb(fn func(old *{{.Name}}, new *{{.Name}})) {
	c.UpdatedCb = fn
}

{{- if .CudCache}}
func (c *{{.Name}}Cache) SyncUpdate(key, val []byte, rev int64) {
	obj := {{.Name}}{}
	err := json.Unmarshal(val, &obj)
	if err != nil {
		log.WarnLog("Failed to parse {{.Name}} data", "val", string(val))
		return
	}
	c.Update(&obj, rev)
	c.Mux.Lock()
	if c.List != nil {
		c.List[obj.Key] = struct{}{}
	}
	c.Mux.Unlock()
}

func (c *{{.Name}}Cache) SyncDelete(key []byte, rev int64) {
	obj := {{.Name}}{}
	keystr := objstore.DbKeyPrefixRemove(string(key))
	{{.Name}}KeyStringParse(keystr, &obj.Key)
	c.Delete(&obj, rev)
}

func (c *{{.Name}}Cache) SyncListStart() {
	c.List = make(map[{{.KeyType}}]struct{})
}

func (c *{{.Name}}Cache) SyncListEnd() {
	deleted := make(map[{{.KeyType}}]struct{})
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
		m.generateFields([]string{*message.Name + "Field"}, []string{}, desc)
		m.P("")
		m.P("var ", *message.Name, "AllFields = []string{")
		m.generateAllFields(AllFieldsGenSlice, []string{*message.Name + "Field"}, []string{}, desc)
		m.P("}")
		m.P("")
		m.P("var ", *message.Name, "AllFieldsMap = map[string]struct{}{")
		m.generateAllFields(AllFieldsGenMap, []string{*message.Name + "Field"}, []string{}, desc)
		m.P("}")
		m.P("")
	}
	msgtyp := m.gen.TypeName(desc)
	m.P("func (m *", msgtyp, ") CopyInFields(src *", msgtyp, ") {")
	if HasGrpcFields(message) {
		m.P("fmap := MakeFieldMap(src.Fields)")
	}
	m.generateCopyIn(make([]string, 0), make([]string, 0), desc, make([]*generator.Descriptor, 0), HasGrpcFields(message))
	m.P("}")
	m.P("")

	if GetGenerateCud(message) {
		if GetMessageKey(message) == nil {
			m.gen.Fail("message", *message.Name, "needs a unique key field named key of type", *message.Name+"Key", "for option generate_cud")
		}
		args := cudTemplateArgs{
			Name:      *message.Name,
			CudName:   *message.Name + "Cud",
			KeyName:   *message.Name + "Key",
			HasFields: HasGrpcFields(message),
		}
		m.cudTemplate.Execute(m.gen.Buffer, args)
	}
	if GetGenerateCache(message) {
		keyField := GetMessageKey(message)
		if keyField == nil {
			m.gen.Fail("message", *message.Name, "needs a unique key field named key of type", *message.Name+"Key", "for option generate_cud")
		}
		args := cacheTemplateArgs{
			Name:        *message.Name,
			KeyType:     m.support.GoType(m.gen, keyField),
			CudCache:    GetGenerateCud(message),
			NotifyCache: GetNotifyCache(message),
		}
		m.cacheTemplate.Execute(m.gen.Buffer, args)
		m.importUtil = true
	}
	if GetObjKey(message) {
		m.P("func (m *", message.Name, ") GetKeyString() string {")
		m.P("key, err := json.Marshal(m)")
		m.P("if err != nil {")
		m.P("log.FatalLog(\"Failed to marshal ", message.Name, " key string\", \"obj\", m)")
		m.P("}")
		m.P("return string(key)")
		m.P("}")
		m.P("")

		m.P("func ", message.Name, "StringParse(str string, key *", message.Name, ") {")
		m.P("err := json.Unmarshal([]byte(str), key)")
		m.P("if err != nil {")
		m.P("log.FatalLog(\"Failed to unmarshal ", message.Name, " key string\", \"str\", str)")
		m.P("}")
		m.P("}")
		m.P("")
		m.importUtil = true
	}
	if field := GetMessageKey(message); field != nil {
		m.P("func (m *", message.Name, ") GetKey() *", m.support.GoType(m.gen, field), " {")
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
	if message.Field != nil && len(message.Field) > 0 && *message.Field[0].Name == "fields" && *message.Field[0].Type == descriptor.FieldDescriptorProto_TYPE_STRING {
		return true
	}
	return false
}

func GetMessageKey(message *descriptor.DescriptorProto) *descriptor.FieldDescriptorProto {
	if message.Field == nil {
		return nil
	}
	if len(message.Field) > 0 && *message.Field[0].Name == "key" {
		return message.Field[0]
	}
	if len(message.Field) > 1 && HasGrpcFields(message) && *message.Field[1].Name == "key" {
		return message.Field[1]
	}
	return nil
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

func GetNotifyCache(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_NotifyCache, false)
}

func GetObjKey(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_ObjKey, false)
}
