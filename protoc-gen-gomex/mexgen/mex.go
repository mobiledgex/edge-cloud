package mexgen

import (
	"fmt"
	"sort"
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
	importSort    bool
	importTime    bool
	firstFile     string
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
	m.support.Init(gen.Request)
	// Generator passes us all files (some of which are builtin
	// like google/api/http). To determine the first file to generate
	// one-off code, sort by request files which are the subset of
	// files we will generate code for.
	files := make([]string, len(gen.Request.FileToGenerate))
	copy(files, gen.Request.FileToGenerate)
	sort.Strings(files)
	if len(files) > 0 {
		m.firstFile = files[0]
	}
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
	m.importSort = false
	m.importTime = false
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
	if m.firstFile == *file.FileDescriptorProto.Name {
		m.P(matchOptions)
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
		m.gen.PrintImport("", "github.com/coreos/etcd/clientv3/concurrency")
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
	if m.importSort {
		m.gen.PrintImport("", "sort")
	}
	if m.importTime {
		m.gen.PrintImport("", "time")
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

type MatchType int

const (
	FieldMatch MatchType = iota
	ExactMatch
	IgnoreBackendMatch
)

var matchOptions = `
type MatchOptions struct {
	// Filter will ignore 0 or nil fields on the passed in object
	Filter bool
	// IgnoreBackend will ignore fields that were marked backend in .proto
	IgnoreBackend bool
	// Sort repeated (arrays) of Key objects so matching does not
	// fail due to order.
	SortArrayedKeys bool
}

type MatchOpt func(*MatchOptions)

func MatchFilter() MatchOpt {
	return func(opts *MatchOptions) {
		opts.Filter = true
	}
}

func MatchIgnoreBackend() MatchOpt {
	return func(opts *MatchOptions) {
		opts.IgnoreBackend = true
	}
}

func MatchSortArrayedKeys() MatchOpt {
	return func(opts *MatchOptions) {
		opts.SortArrayedKeys = true
	}
}

func applyMatchOptions(opts *MatchOptions, args ...MatchOpt) {
	for _, f := range args {
		f(opts)
	}
}

`

func (m *mex) generateFieldMatches(message *descriptor.DescriptorProto, field *descriptor.FieldDescriptorProto) {
	if field.Type == nil {
		return
	}
	backend := GetBackend(field)
	if backend {
		m.P("if !opts.IgnoreBackend {")
	}

	// ignore field if filter was specified and o.name is 0 or nil
	nilval := "0"
	nilCheck := true
	name := generator.CamelCase(*field.Name)
	if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED ||
		*field.Type == descriptor.FieldDescriptorProto_TYPE_BYTES {
		nilval = "nil"
	} else {
		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			if !gogoproto.IsNullable(field) {
				nilCheck = false
			}
			nilval = "nil"
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			nilval = "\"\""
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			nilval = "false"
		}
	}
	if nilCheck {
		m.P("if !opts.Filter || o.", name, " != ", nilval, " {")
	}
	if nilCheck && nilval == "nil" {
		m.P("if m.", name, " == nil && o.", name, " != nil || m.", name, " != nil && o.", name, " == nil {")
		m.P("return false")
		m.P("} else if m.", name, " != nil && o.", name, "!= nil {")
	}

	mapType := m.support.GetMapType(m.gen, field)
	repeated := false
	if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED ||
		*field.Type == descriptor.FieldDescriptorProto_TYPE_BYTES {
		m.P("if len(m.", name, ") != len(o.", name, ") {")
		m.P("return false")
		m.P("}")
		if mapType == nil {
			if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
				subDesc := gensupport.GetDesc(m.gen, field.GetTypeName())
				if GetObjKey(subDesc.DescriptorProto) {
					m.P("if opts.SortArrayedKeys {")
					m.P("sort.Slice(m.", name, ", func(i, j int) bool {")
					m.P("return m.", name, "[i].GetKeyString() < m.", name, "[j].GetKeyString()")
					m.P("})")
					m.P("sort.Slice(o.", name, ", func(i, j int) bool {")
					m.P("return o.", name, "[i].GetKeyString() < o.", name, "[j].GetKeyString()")
					m.P("})")
					m.P("}")
					m.importSort = true
				}
			}
			m.P("for i := 0; i < len(m.", name, "); i++ {")
			name = name + "[i]"
		} else {
			m.P("for k, _ := range m.", name, " {")
			m.P("_, ok := o.", name, "[k]")
			m.P("if !ok {")
			m.P("return false")
			m.P("}")
			name = name + "[k]"
			field = mapType.ValField
		}
		repeated = true
	}
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		ref := "&"
		if gogoproto.IsNullable(field) {
			ref = ""
		}
		subDesc := gensupport.GetDesc(m.gen, field.GetTypeName())
		printedCheck := true
		if *field.TypeName == ".google.protobuf.Timestamp" {
			m.P("if m.", name, ".Seconds != o.", name, ".Seconds || m.", name, ".Nanos != o.", name, ".Nanos {")
		} else if GetGenerateMatches(subDesc.DescriptorProto) {
			m.P("if !m.", name, ".Matches(", ref, "o.", name, ", fopts...) {")
		} else {
			printedCheck = false
		}
		if printedCheck {
			m.P("return false")
			m.P("}")
		}
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		// deprecated in proto3
	default:
		m.P("if o.", name, " != m.", name, "{")
		m.P("return false")
		m.P("}")
	}
	if repeated {
		m.P("}")
	}
	if nilCheck && nilval == "nil" {
		m.P("}")
	}
	if nilCheck {
		m.P("}")
	}
	if backend {
		m.P("}")
	}
}

func (m *mex) printCopyInMakeArray(name string, desc *generator.Descriptor, field *descriptor.FieldDescriptorProto) {
	mapType := m.support.GetMapType(m.gen, field)
	if mapType != nil {
		valType := mapType.ValType
		if mapType.ValIsMessage {
			valType = "*" + valType
		}
		m.P("m.", name, " = make(map[", mapType.KeyType, "]", valType, ")")
	} else {
		typ, _ := m.gen.GoType(desc, field)
		m.P("if m.", name, " == nil || len(m.", name, ") != len(src.", name, ") {")
		m.P("m.", name, " = make(", typ, ", len(src.", name, "))")
		m.P("}")
	}
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

func (m *mex) markDiff(names []string, name string) {
	// set field and all parent fields
	names = append(names, name)
	for len(names) > 1 {
		fieldName := strings.Join(names, "")
		m.P("fields[", fieldName, "] = struct{}{}")
		names = names[:len(names)-1]
	}
}

func (m *mex) generateDiffFields(parents, names []string, desc *generator.Descriptor) {
	message := desc.DescriptorProto
	for ii, field := range message.Field {
		if ii == 0 && *field.Name == "fields" {
			continue
		}
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_GROUP {
			// deprecated in proto3
			continue
		}

		name := generator.CamelCase(*field.Name)
		hierName := strings.Join(append(parents, name), ".")
		idx := ""
		mapType := m.support.GetMapType(m.gen, field)
		loop := false
		skipMap := false

		if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED ||
			*field.Type == descriptor.FieldDescriptorProto_TYPE_BYTES {
			depth := fmt.Sprintf("%d", len(parents))
			m.P("if len(m.", hierName, ") != len(o.", hierName, ") {")
			m.markDiff(names, name)
			m.P("} else {")
			if mapType == nil {
				m.P("for i", depth, " := 0; i", depth, " < len(m.", hierName, "); i", depth, "++ {")
				idx = "[i" + depth + "]"
			} else {
				m.P("for k", depth, ", _ := range m.", hierName, " {")
				m.P("_, vok", depth, " := o.", hierName, "[k", depth, "]")
				m.P("if !vok", depth, " {")
				m.markDiff(names, name)
				m.P("} else {")
				if !mapType.ValIsMessage {
					skipMap = true
				}
				idx = "[k" + depth + "]"
			}
			loop = true
		}
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE && !skipMap {
			subDesc := gensupport.GetDesc(m.gen, field.GetTypeName())
			subNames := append(names, name)
			if mapType != nil {
				subDesc = gensupport.GetDesc(m.gen, mapType.ValField.GetTypeName())
				subNames = append(subNames, "Value")
			}
			m.generateDiffFields(append(parents, name+idx), subNames, subDesc)
		} else {
			m.P("if m.", hierName, idx, " != o.", hierName, idx, " {")
			m.markDiff(names, name)
			if loop {
				m.P("break")
			}
			m.P("}")
		}
		if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED ||
			*field.Type == descriptor.FieldDescriptorProto_TYPE_BYTES {
			m.P("}")
			m.P("}")
			if mapType != nil {
				m.P("}")
			}
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
		mapType := m.support.GetMapType(m.gen, field)
		skipMap := false

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
				depth := fmt.Sprintf("%d", len(parents))
				if mapType == nil {
					m.P("for i", depth, " := 0; i", depth, " < len(src.", hierName, "); i", depth, "++ {")
					idx = "[i" + depth + "]"
				} else {
					m.P("for k", depth, ", _ := range src.", hierName, " {")
					idx = "[k" + depth + "]"
					if !mapType.ValIsMessage {
						skipMap = true
						m.P("m.", hierName, idx, " = src.", hierName, idx)
					}
				}
			}
		}
		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			if skipMap {
				break
			}
			subDesc := gensupport.GetDesc(m.gen, field.GetTypeName())
			if mapType != nil {
				if mapType.ValIsMessage {
					m.P("m.", hierName, idx, " = &", mapType.ValType, "{}")
				}
				subDesc = gensupport.GetDesc(m.gen, mapType.ValField.GetTypeName())
			} else if gogoproto.IsNullable(field) {
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
	Name        string
	KeyType     string
	CudName     string
	HasFields   bool
	GenCache    bool
	NotifyCache bool
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
	kvstore objstore.KVStore
}

func New{{.Name}}Store(kvstore objstore.KVStore) {{.Name}}Store {
	return {{.Name}}Store{kvstore: kvstore}
}

func (s *{{.Name}}Store) Create(m *{{.Name}}, wait func(int64)) (*Result, error) {
{{- if (.HasFields)}}
	err := m.Validate({{.Name}}AllFieldsMap)
{{- else}}
	err := m.Validate(nil)
{{- end}}
	if err != nil { return nil, err }
	key := objstore.DbKeyString("{{.Name}}", m.GetKey())
	val, err := json.Marshal(m)
	if err != nil { return nil, err }
	rev, err := s.kvstore.Create(key, string(val))
	if err != nil { return nil, err }
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *{{.Name}}Store) Update(m *{{.Name}}, wait func(int64)) (*Result, error) {
{{- if (.HasFields)}}
	fmap := MakeFieldMap(m.Fields)
	err := m.Validate(fmap)
{{- else}}
	err := m.Validate(nil)
{{- end}}
	if err != nil { return nil, err }
	key := objstore.DbKeyString("{{.Name}}", m.GetKey())
	var vers int64 = 0
{{- if (.HasFields)}}
	curBytes, vers, _, err := s.kvstore.Get(key)
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
	rev, err := s.kvstore.Update(key, string(val), vers)
	if err != nil { return nil, err }
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *{{.Name}}Store) Put(m *{{.Name}}, wait func(int64)) (*Result, error) {
{{- if (.HasFields)}}
	fmap := MakeFieldMap(m.Fields)
	err := m.Validate(fmap)
{{- else}}
	err := m.Validate(nil)
{{- end}}
	if err != nil { return nil, err }
	key := objstore.DbKeyString("{{.Name}}", m.GetKey())
	var val []byte
{{- if (.HasFields)}}
	curBytes, _, _, err := s.kvstore.Get(key)
	if err == nil {
		var cur {{.Name}}
		err = json.Unmarshal(curBytes, &cur)
		if err != nil { return nil, err }
		cur.CopyInFields(m)
		// never save fields
		cur.Fields = nil
		val, err = json.Marshal(cur)
	} else {
		m.Fields = nil
		val, err = json.Marshal(m)
	}
{{- else}}
	val, err = json.Marshal(m)
{{- end}}
	if err != nil { return nil, err }
	rev, err := s.kvstore.Put(key, string(val))
	if err != nil { return nil, err }
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *{{.Name}}Store) Delete(m *{{.Name}}, wait func(int64)) (*Result, error) {
	err := m.GetKey().Validate()
	if err != nil { return nil, err }
	key := objstore.DbKeyString("{{.Name}}", m.GetKey())
	rev, err := s.kvstore.Delete(key)
	if err != nil { return nil, err }
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *{{.Name}}Store) LoadOne(key string) (*{{.Name}}, int64, error) {
	val, rev, _, err := s.kvstore.Get(key)
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

func (s *{{.Name}}Store) STMGet(stm concurrency.STM, key *{{.KeyType}}, buf *{{.Name}}) bool {
	keystr := objstore.DbKeyString("{{.Name}}", key)
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

func (s *{{.Name}}Store) STMPut(stm concurrency.STM, obj *{{.Name}}) {
	keystr := objstore.DbKeyString("{{.Name}}", obj.GetKey())
	val, _ := json.Marshal(obj)
	stm.Put(keystr, string(val))
}

func (s *{{.Name}}Store) STMDel(stm concurrency.STM, key *{{.KeyType}}) {
	keystr := objstore.DbKeyString("{{.Name}}", key)
	stm.Del(keystr)
}

`

type cacheTemplateArgs struct {
	Name         string
	KeyType      string
	CudCache     bool
	NotifyCache  bool
	WaitForState string
}

var cacheTemplateIn = `
type {{.Name}}KeyWatcher struct {
	cb func()
}

// {{.Name}}Cache caches {{.Name}} objects in memory in a hash table
// and keeps them in sync with the database.
type {{.Name}}Cache struct {
	Objs map[{{.KeyType}}]*{{.Name}}
	Mux util.Mutex
	List map[{{.KeyType}}]struct{}
	NotifyCb func(obj *{{.KeyType}}, old *{{.Name}})
	UpdatedCb func(old *{{.Name}}, new *{{.Name}})
	KeyWatchers map[{{.KeyType}}][]*{{.Name}}KeyWatcher
}

func New{{.Name}}Cache() *{{.Name}}Cache {
	cache := {{.Name}}Cache{}
	Init{{.Name}}Cache(&cache)
	return &cache
}

func Init{{.Name}}Cache(cache *{{.Name}}Cache) {
	cache.Objs = make(map[{{.KeyType}}]*{{.Name}})
	cache.KeyWatchers = make(map[{{.KeyType}}][]*{{.Name}}KeyWatcher)
}

func (c *{{.Name}}Cache) GetTypeString() string {
	return "{{.Name}}"
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
	if c.UpdatedCb != nil || c.NotifyCb != nil {
		old := c.Objs[in.Key]
		if c.UpdatedCb != nil {
			new := &{{.Name}}{}
			*new = *in
			defer c.UpdatedCb(old, new)
		}
		if c.NotifyCb != nil {
			defer c.NotifyCb(&in.Key, old)
		}
	}
	c.Objs[in.Key] = in
	log.DebugLog(log.DebugLevelApi, "SyncUpdate {{.Name}}", "obj", in, "rev", rev)
	c.Mux.Unlock()
	c.TriggerKeyWatchers(&in.Key)
}

func (c *{{.Name}}Cache) Delete(in *{{.Name}}, rev int64) {
	c.Mux.Lock()
	old := c.Objs[in.Key]
	delete(c.Objs, in.Key)
	log.DebugLog(log.DebugLevelApi, "SyncDelete {{.Name}}", "key", in.Key, "rev", rev)
	c.Mux.Unlock()
	if c.NotifyCb != nil {
		c.NotifyCb(&in.Key, old)
	}
	c.TriggerKeyWatchers(&in.Key)
}

func (c *{{.Name}}Cache) Prune(validKeys map[{{.KeyType}}]struct{}) {
	notify := make(map[{{.KeyType}}]*{{.Name}})
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

func (c *{{.Name}}Cache) GetCount() int {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	return len(c.Objs)
}

{{- if .NotifyCache}}
func (c *{{.Name}}Cache) Flush(notifyId int64) {
	flushed := make(map[{{.KeyType}}]*{{.Name}})
	c.Mux.Lock()
	for key, val := range c.Objs {
		if val.NotifyId != notifyId {
			continue
		}
		flushed[key] = c.Objs[key]
		delete(c.Objs, key)
	}
	c.Mux.Unlock()
	if len(flushed) > 0 {
		for key, old := range flushed {
			if c.NotifyCb != nil {
				c.NotifyCb(&key, old)
			}
			c.TriggerKeyWatchers(&key)
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
		if !obj.Matches(filter, MatchFilter()) {
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

func (c *{{.Name}}Cache) SetNotifyCb(fn func(obj *{{.KeyType}}, old *{{.Name}})) {
	c.NotifyCb = fn
}

func (c *{{.Name}}Cache) SetUpdatedCb(fn func(old *{{.Name}}, new *{{.Name}})) {
	c.UpdatedCb = fn
}

func (c *{{.Name}}Cache) WatchKey(key *{{.KeyType}}, cb func()) context.CancelFunc {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	list, ok := c.KeyWatchers[*key]
	if !ok {
		list = make([]*{{.Name}}KeyWatcher, 0)
	}
	watcher := {{.Name}}KeyWatcher{cb: cb}
	c.KeyWatchers[*key] = append(list, &watcher)
	log.DebugLog(log.DebugLevelApi, "Watching {{.Name}}", "key", key)
	return func() {
		c.Mux.Lock()
		defer c.Mux.Unlock()
		list, ok := c.KeyWatchers[*key]
		if !ok { return }
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

func (c *{{.Name}}Cache) TriggerKeyWatchers(key *{{.KeyType}}) {
	watchers := make([]*{{.Name}}KeyWatcher, 0)
	c.Mux.Lock()
	if list, ok := c.KeyWatchers[*key]; ok {
		watchers = append(watchers, list...)
	}
	c.Mux.Unlock()
	for ii, _ := range watchers {
		watchers[ii].cb()
	}
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
	{{.KeyType}}StringParse(keystr, &obj.Key)
	c.Delete(&obj, rev)
}

func (c *{{.Name}}Cache) SyncListStart() {
	c.List = make(map[{{.KeyType}}]struct{})
}

func (c *{{.Name}}Cache) SyncListEnd() {
	deleted := make(map[{{.KeyType}}]*{{.Name}})
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
{{- end}}

{{if ne (.WaitForState) ("")}}
func (c *{{.Name}}Cache) WaitForState(key *{{.KeyType}}, targetState {{.WaitForState}}, timeout time.Duration, send func(*Result) error) error {
	curState := {{.WaitForState}}_{{.WaitForState}}Unknown
	done := make(chan bool, 1)
	failed := make(chan bool, 1)
	var err error

	cancel := c.WatchKey(key, func() {
		info := {{.Name}}{}
		if c.Get(key, &info) {
			curState = info.State
		} else {
			curState = {{.WaitForState}}_{{.WaitForState}}NotPresent
		}
		if send != nil {
			msg := {{.WaitForState}}_name[int32(curState)]
			send(&Result{Message: msg})
		}
		log.DebugLog(log.DebugLevelApi, "Watch event for {{.Name}}", "key", key, "state", {{.WaitForState}}_name[int32(curState)])
		if curState == {{.WaitForState}}_{{.WaitForState}}Errors {
			failed <- true
		} else if curState == targetState {
			done <- true
		}
	})
	// After setting up watch, check current state,
	// as it may have already changed to target state
	info := {{.Name}}{}
	if c.Get(key, &info) {
		curState = info.State
	} else {
		curState = {{.WaitForState}}_{{.WaitForState}}NotPresent
	}
	if curState == targetState {
		done <- true
	}

	select {
	case <-done:
		err = nil
	case <-failed:
		if c.Get(key, &info) {
			err = fmt.Errorf("Encountered failures: %v", info.Errors)
		} else {
			// this shouldn't happen, since only way to get here
			// is if info state is set to Error
			err = errors.New("Unknown failure")
		}
	case <-time.After(timeout):
		if c.Get(key, &info) && info.State == {{.WaitForState}}_{{.WaitForState}}Errors {
			// error may have been sent back before watch started
			err = fmt.Errorf("Encountered failures: %v", info.Errors)
		} else {
			err = fmt.Errorf("Timed out; expected state %s but is %s",
				{{.WaitForState}}_name[int32(targetState)],
				{{.WaitForState}}_name[int32(curState)])
		}
	}
	cancel()
	// note: do not close done/failed, garbage collector will deal with it.
	return err
}
{{- end}}

`

func (m *mex) generateMessage(file *generator.FileDescriptor, desc *generator.Descriptor) {
	message := desc.DescriptorProto
	if GetGenerateMatches(message) && message.Field != nil {
		m.P("func (m *", message.Name, ") Matches(o *", message.Name, ", fopts ...MatchOpt) bool {")
		m.P("opts := MatchOptions{}")
		m.P("applyMatchOptions(&opts, fopts...)")
		m.P("if o == nil {")
		m.P("if opts.Filter { return true }")
		m.P("return false")
		m.P("}")
		for ii, field := range message.Field {
			if ii == 0 && *field.Name == "fields" {
				continue
			}
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
		m.P("func (m *", message.Name, ") DiffFields(o *", message.Name, ", fields map[string]struct{}) {")
		m.generateDiffFields([]string{}, []string{*message.Name + "Field"}, desc)
		m.P("}")
		m.P("")
	}
	if desc.GetOptions().GetMapEntry() {
		return
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
		keyField := GetMessageKey(message)
		if keyField == nil {
			m.gen.Fail("message", *message.Name, "needs a unique key field named key of type", *message.Name+"Key", "for option generate_cud")
		}
		args := cudTemplateArgs{
			Name:      *message.Name,
			CudName:   *message.Name + "Cud",
			KeyType:   m.support.GoType(m.gen, keyField),
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
			Name:         *message.Name,
			KeyType:      m.support.GoType(m.gen, keyField),
			CudCache:     GetGenerateCud(message),
			NotifyCache:  GetNotifyCache(message),
			WaitForState: GetGenerateWaitForState(message),
		}
		m.cacheTemplate.Execute(m.gen.Buffer, args)
		m.importUtil = true
		if args.WaitForState != "" {
			m.importTime = true
		}
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

func GetGenerateWaitForState(message *descriptor.DescriptorProto) string {
	return gensupport.GetStringExtension(message.Options, protogen.E_GenerateWaitForState, "")
}

func GetNotifyCache(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_NotifyCache, false)
}

func GetObjKey(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_ObjKey, false)
}

func GetBackend(field *descriptor.FieldDescriptorProto) bool {
	return proto.GetBoolExtension(field.Options, protogen.E_Backend, false)
}
