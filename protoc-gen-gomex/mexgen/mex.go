package mexgen

import (
	"fmt"
	"sort"
	"strconv"
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
	gen            *generator.Generator
	msgs           map[string]*descriptor.DescriptorProto
	cudTemplate    *template.Template
	enumTemplate   *template.Template
	cacheTemplate  *template.Template
	sqlTemplate    *template.Template
	importUtil     bool
	importStrings  bool
	importErrors   bool
	importStrconv  bool
	importSort     bool
	importTime     bool
	importSql      bool
	importJson     bool
	importLog      bool
	importObjstore bool
	firstFile      string
	support        gensupport.PluginSupport
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
	m.sqlTemplate = template.Must(template.New("sql").Parse(sqlTemplateIn))
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
	m.importSql = false
	m.importJson = false
	m.importLog = false
	m.importObjstore = false
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
	if m.importJson {
		m.gen.PrintImport("", "encoding/json")
	}
	if m.importObjstore {
		m.gen.PrintImport("", "github.com/mobiledgex/edge-cloud/objstore")
	}
	if hasGenerateCud {
		m.gen.PrintImport("", "github.com/coreos/etcd/clientv3/concurrency")
	}
	if m.importUtil {
		m.gen.PrintImport("", "github.com/mobiledgex/edge-cloud/util")
	}
	if m.importLog {
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
	if m.importSql {
		m.gen.PrintImport("", "database/sql")
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

func (s *{{.Name}}Store) Put(m *{{.Name}}, wait func(int64), ops ...objstore.KVOp) (*Result, error) {
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
	rev, err := s.kvstore.Put(key, string(val), ops...)
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
	c.UpdateModFunc(&in.Key, rev, func(old *{{.Name}}) (*{{.Name}}, bool) {
		return in, true
	})
}

func (c *{{.Name}}Cache) UpdateModFunc(key *{{.KeyType}}, rev int64, modFunc func(old *{{.Name}}) (new *{{.Name}}, changed bool)) {
	c.Mux.Lock()
	old := c.Objs[*key]
	new, changed := modFunc(old)
	if !changed {
		c.Mux.Unlock()
		return
	}
	if c.UpdatedCb != nil || c.NotifyCb != nil {
		if c.UpdatedCb != nil {
			newCopy := &{{.Name}}{}
			*newCopy = *new
			defer c.UpdatedCb(old, newCopy)
		}
		if c.NotifyCb != nil {
			defer c.NotifyCb(&new.Key, old)
		}
	}
	c.Objs[new.Key] = new
	log.DebugLog(log.DebugLevelApi, "SyncUpdate {{.Name}}", "obj", new, "rev", rev)
	c.Mux.Unlock()
	c.TriggerKeyWatchers(&new.Key)
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

func {{.Name}}GenericNotifyCb(fn func(key *{{.KeyType}}, old *{{.Name}})) func(objstore.ObjKey, objstore.Obj) {
	return func(objkey objstore.ObjKey, obj objstore.Obj) {
		fn(objkey.(*{{.KeyType}}), obj.(*{{.Name}}))
	}
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
func (c *{{.Name}}Cache) WaitForState(ctx context.Context, key *{{.KeyType}}, targetState {{.WaitForState}}, transitionStates map[{{.WaitForState}}]struct{}, errorState {{.WaitForState}}, timeout time.Duration, successMsg string, send func(*Result) error) error {
	curState := {{.WaitForState}}_{{.WaitForState}}Unknown
	done := make(chan bool, 1)
	failed := make(chan bool, 1)
	var err error

	cancel := c.WatchKey(key, func() {
		info := {{.Name}}{}
		if c.Get(key, &info) {
			curState = info.State
		} else {
			curState = {{.WaitForState}}_NotPresent
		}
		if send != nil {
			msg := {{.WaitForState}}_name[int32(curState)]
			send(&Result{Message: msg})
		}
		log.DebugLog(log.DebugLevelApi, "Watch event for {{.Name}}", "key", key, "state", {{.WaitForState}}_name[int32(curState)])
		if curState == errorState {
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
		curState = {{.WaitForState}}_NotPresent
	}
	if curState == targetState {
		done <- true
	}

	select {
	case <-done:
		err = nil
		if successMsg != "" {
			send(&Result{Message: successMsg})
		}
	case <-failed:
		if c.Get(key, &info) {
			err = fmt.Errorf("Encountered failures: %v", info.Errors)
		} else {
			// this shouldn't happen, since only way to get here
			// is if info state is set to Error
			err = errors.New("Unknown failure")
		}
	case <-time.After(timeout):
		hasInfo := c.Get(key, &info)
		if hasInfo && info.State == errorState {
			// error may have been sent back before watch started
			err = fmt.Errorf("Encountered failures: %v", info.Errors)
		} else if _, found := transitionStates[info.State]; hasInfo && found {
			// no success response, but state is a valid transition
			// state. That means work is still in progress.
			// Notify user that this is not an error.
			// Do not undo since CRM is still busy.
			msg := fmt.Sprintf("Timed out while work still in progress state %s. Please use Show{{.Name}} to check current status", {{.WaitForState}}_name[int32(info.State)])
			send(&Result{Message: msg})
			err = nil
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

type sqlTemplateArgs struct {
	Name             string
	KeyType          string
	Table            string
	HasGrpcFields    bool
	PrimaryKey       *SqlField
	Fields           []SqlField
	CreateTableLines []string
	AllFields        string
	CreateFields     string
	CreateVars       string
	CreateRefs       string
	GetFields        string
}

type SqlField struct {
	Name       string
	Refname    string
	Fieldnum   string
	Typ        string
	Primarykey bool
	SerialTyp  bool
	Option     string
}

var sqlTemplateIn = `
type {{.Name}}Sql struct {
	DB *sql.DB
	Table string
}

func (s *{{.Name}}Sql) Init(db *sql.DB) {
	s.DB = db
	s.Table = "{{.Table}}"
}

func Create{{.Name}}SqlTable(db *sql.DB) (sql.Result, error) {
	return db.Exec("CREATE TABLE IF NOT EXISTS {{.Table}}(" +
	{{- range .CreateTableLines}}
		{{.}}
	{{- end}}
}

func (s *{{.Name}}Sql) Create(m *{{.Name}}) (sql.Result, error) {
{{- if (.HasGrpcFields)}}
	err := m.Validate({{.Name}}AllFieldsMap)
{{- else}}
	err := m.Validate()
{{- end}}
	if err != nil { return nil, err}
	stmt := "INSERT INTO {{.Table}}({{.CreateFields}}) VALUES({{.CreateVars}})"
	log.DebugLog(log.DebugLevelApi, "create {{.Name}}", "stmt", stmt, "vals", []interface{}{ {{.CreateRefs}} })
	return s.DB.Exec(stmt, {{.CreateRefs}})
}

func (s *{{.Name}}Sql) Delete(m *{{.Name}}) (sql.Result, error) {
	err := m.GetKey().Validate()
	if err != nil { return nil, err }
	stmt := "DELETE FROM {{.Table}} WHERE {{.PrimaryKey.Name}} = $1"
	log.DebugLog(log.DebugLevelApi, "delete {{.Name}}", "stmt", stmt, "key", m.{{.PrimaryKey.Refname}})
	return s.DB.Exec(stmt, m.{{.PrimaryKey.Refname}})
}

func (s *{{.Name}}Sql) HasKey(key *{{.KeyType}}) (bool, error) {
	buf := {{.Name}}{}
	return s.Get(key, &buf)
}

func (s *{{.Name}}Sql) Get(key *{{.KeyType}}, buf *{{.Name}}) (bool, error) {
	stmt := "SELECT {{.AllFields}} FROM {{.Table}} WHERE {{.PrimaryKey.Name}} = $1"
	row := s.DB.QueryRow(stmt, key.{{.PrimaryKey.Name}})
	err := row.Scan({{.GetFields}})
	if err == nil {
		return true, nil
	} else if err == sql.ErrNoRows {
		return false, nil
	}
	return false, err
}

func (s *{{.Name}}Sql) Show(filter *{{.Name}}, cb func(ret *{{.Name}}) error) error {
	stmt := "SELECT {{.AllFields}} FROM {{.Table}}"
	rows, err := s.DB.Query(stmt)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		buf := {{.Name}}{}
		err = rows.Scan({{.GetFields}})
		if err != nil {
			return fmt.Errorf("sql scan failed, %s", err.Error())
		}
		err = cb(&buf)
		if err != nil {
			return err
		}
	}
	// return any error encountered during iteration
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

{{- if .HasGrpcFields}}
func (s *{{.Name}}Sql) Update(m *{{.Name}}) (sql.Result, error) {
	fmap := MakeFieldMap(m.Fields)
	err := m.Validate(fmap)
	if err != nil {
		return nil, err
	}
	vars := make([]interface{}, 0)
	sets := make([]string, 0)

	vars = append(vars, m.{{.PrimaryKey.Refname}}) // $1
	{{- range .Fields}}
	{{- if not .SerialTyp}}
	if _, found := fmap["{{.Fieldnum}}"]; found {
		vars = append(vars, m.{{.Refname}})
		sets = append(sets, "{{.Name}} = $" + strconv.Itoa(len(vars)))
	}
	{{- end}}
	{{- end}}
	if len(sets) == 0 {
		// nothing specified to update
		return nil, fmt.Errorf("No fields specified to update")
	}

	stmt := fmt.Sprintf("UPDATE {{.Table}} SET %s WHERE {{.PrimaryKey.Name}} = $1", strings.Join(sets, ", "))
	log.DebugLog(log.DebugLevelApi, "update {{.Name}}", "stmt", stmt, "vars", vars)
	res, err := s.DB.Exec(stmt, vars...)
	if err != nil {
		return res, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return res, err
	}
	if count != 1 {
		return res, fmt.Errorf("Updated %d rows instead of 1", count)
	}
	return res, nil
}

func (s *{{.Name}}Sql) GetMatch(filter *{{.Name}}, buf *{{.Name}}) (bool, error) {
	fmap := MakeFieldMap(filter.Fields)
	vars := make([]interface{}, 0)
	reqs := make([]string, 0)

	{{- range .Fields}}
	if _, found := fmap["{{.Fieldnum}}"]; found {
		vars = append(vars, filter.{{.Refname}})
		reqs = append(reqs, "{{.Name}} = $"+strconv.Itoa(len(vars)))
	}
	{{- end}}

	stmt := fmt.Sprintf("SELECT {{.AllFields}} FROM {{.Table}} WHERE %s", strings.Join(reqs, " AND "))
	log.DebugLog(log.DebugLevelApi, "get match {{.Name}}", "stmt", stmt, "vars", vars)
	row := s.DB.QueryRow(stmt, vars...)
	err := row.Scan({{.GetFields}})
	if err == nil {
		return true, nil
	} else if err == sql.ErrNoRows {
		return false, nil
	}
	return false, err
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
	if gensupport.HasGrpcFields(message) {
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
	if gensupport.HasGrpcFields(message) {
		m.P("fmap := MakeFieldMap(src.Fields)")
	}
	m.generateCopyIn(make([]string, 0), make([]string, 0), desc, make([]*generator.Descriptor, 0), gensupport.HasGrpcFields(message))
	m.P("}")
	m.P("")

	if GetGenerateCud(message) {
		keyField := gensupport.GetMessageKey(message)
		if keyField == nil {
			m.gen.Fail("message", *message.Name, "needs a unique key field named key of type", *message.Name+"Key", "for option generate_cud")
		}
		args := cudTemplateArgs{
			Name:      *message.Name,
			CudName:   *message.Name + "Cud",
			KeyType:   m.support.GoType(m.gen, keyField),
			HasFields: gensupport.HasGrpcFields(message),
		}
		m.cudTemplate.Execute(m.gen.Buffer, args)
		m.importJson = true
		m.importLog = true
		m.importObjstore = true
	}
	if GetGenerateCache(message) {
		keyField := gensupport.GetMessageKey(message)
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
		m.importLog = true
		m.importObjstore = true
		if args.WaitForState != "" {
			m.importErrors = true
			m.importTime = true
		}
		if args.CudCache {
			m.importJson = true
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
		m.importLog = true
		m.importJson = true
	}
	if field := gensupport.GetMessageKey(message); field != nil {
		m.P("func (m *", message.Name, ") GetKey() objstore.ObjKey {")
		m.P("return &m.Key")
		m.P("}")
		m.P("")
		m.importObjstore = true
	}
	if GetGenerateSql(message) {
		keyField := gensupport.GetMessageKey(message)
		if keyField == nil {
			m.gen.Fail("sql message ", *message.Name, " has no Key field")
			return
		}
		keyDesc := gensupport.GetDesc(m.gen, keyField.GetTypeName())
		if len(keyDesc.DescriptorProto.Field) != 1 {
			m.gen.Fail("sql message ", *message.Name, "'s Key ", *keyDesc.Name, " must only have 1 field to be used as the primary key")
			return
		}

		spec := make([]SqlField, 0)
		m.getSqlFields(make([]string, 0), make([]string, 0), desc, make([]*generator.Descriptor, 0), keyDesc, &spec)

		args := sqlTemplateArgs{
			Name:          *message.Name,
			KeyType:       m.support.GoType(m.gen, keyField),
			Table:         *message.Name + "s",
			HasGrpcFields: gensupport.HasGrpcFields(message),
			Fields:        spec,
		}
		args.CreateTableLines = make([]string, 0)

		allFields := make([]string, 0)
		createFields := make([]string, 0)
		createVars := make([]string, 0)
		createRefs := make([]string, 0)
		getFields := make([]string, 0)
		for ii, _ := range spec {
			tail := ", \" +"
			if len(spec) == ii+1 {
				// last one
				tail = ")\")"
			}
			if spec[ii].Primarykey {
				spec[ii].Option += " primary key"
				args.PrimaryKey = &spec[ii]
			}
			line := fmt.Sprintf("\"%s %s %s%s", spec[ii].Name,
				spec[ii].Typ, spec[ii].Option, tail)
			args.CreateTableLines = append(args.CreateTableLines, line)

			allFields = append(allFields, spec[ii].Name)
			getFields = append(getFields, "&buf."+spec[ii].Refname)
			if !strings.Contains(strings.ToLower(spec[ii].Typ), "serial") {
				createFields = append(createFields, spec[ii].Name)
				createVars = append(createVars, "$"+strconv.Itoa(len(allFields)))
				createRefs = append(createRefs, "m."+spec[ii].Refname)
			}
		}
		args.AllFields = strings.Join(allFields, ", ")
		args.CreateFields = strings.Join(createFields, ", ")
		args.CreateVars = strings.Join(createVars, ", ")
		args.CreateRefs = strings.Join(createRefs, ", ")
		args.GetFields = strings.Join(getFields, ", ")

		m.sqlTemplate.Execute(m.gen.Buffer, args)
		m.importSql = true
		if args.HasGrpcFields {
			m.importStrconv = true
		}
	}

	//Generate enum values validation
	m.generateEnumValidation(message, desc)
}

// Generate a single check for an enum
func (m *mex) generateEnumCheck(field *descriptor.FieldDescriptorProto, elem string) {
	m.P("if _, ok := ", m.support.GoType(m.gen, field), "_name[int32(", elem,
		")]; !ok {")
	m.P("return errors.New(\"invalid ", generator.CamelCase(*field.Name),
		"\")")
	m.P("}")
}

func (m *mex) generateMessageEnumCheck(elem string) {
	m.P("if err := ", elem, ".ValidateEnums(); err != nil {")
	m.P("return err")
	m.P("}")
}

// Generate enum validation method for each message
// NOTE: we don't check for set fields. This is ok as
// long as enums start at 0 and unset fields are zeroed out
func (m *mex) generateEnumValidation(message *descriptor.DescriptorProto, desc *generator.Descriptor) {
	m.P("// Helper method to check that enums have valid values")
	if gensupport.HasGrpcFields(message) {
		m.P("// NOTE: ValidateEnums checks all Fields even if some are not set")
	}
	msgtyp := m.gen.TypeName(desc)
	m.P("func (m *", msgtyp, ") ValidateEnums() error {")
	for _, field := range message.Field {
		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_ENUM:
			// could be an array of enums
			if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				m.P("for _, e := range m.", generator.CamelCase(*field.Name), " {")
				m.generateEnumCheck(field, "e")
				m.P("}")
			} else {
				m.generateEnumCheck(field, "m."+generator.CamelCase(*field.Name))
			}
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			// Don't try to generate a call to a vlidation for external package
			if _, ok := m.support.MessageTypesGen[field.GetTypeName()]; !ok {
				continue
			}
			// Not supported OneOf types
			if field.OneofIndex != nil {
				continue
			}
			// could be an array of messages
			if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				m.P("for _, e := range m.", generator.CamelCase(*field.Name), " {")
				m.generateMessageEnumCheck("e")
				m.P("}")
			} else {
				m.generateMessageEnumCheck("m." + generator.CamelCase(*field.Name))
			}
		}
	}
	m.P("return nil")
	m.P("}")
	m.P("")
}

func (m *mex) getSqlFields(parents, nums []string, desc *generator.Descriptor, visited []*generator.Descriptor, keyDesc *generator.Descriptor, spec *[]SqlField) {
	if gensupport.WasVisited(desc, visited) {
		return
	}
	for ii, field := range desc.DescriptorProto.Field {
		if ii == 0 && *field.Name == "fields" {
			continue
		}
		if GetSqlSkip(field) {
			continue
		}
		name := generator.CamelCase(*field.Name)
		num := fmt.Sprintf("%d", *field.Number)
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			subDesc := gensupport.GetDesc(m.gen, field.GetTypeName())
			m.getSqlFields(append(parents, name), append(nums, num), subDesc, append(visited, desc), keyDesc, spec)
			continue
		}
		sqlField := SqlField{
			Name:     strings.Join(append(parents, name), ""),
			Refname:  strings.Join(append(parents, name), "."),
			Fieldnum: strings.Join(append(nums, num), "."),
			Typ:      m.support.SqlType(m.gen, field),
			Option:   GetSqlOption(field),
		}
		if desc == keyDesc {
			// don't bother pre-pending Key struct name
			sqlField.Name = name
			sqlField.Primarykey = true
		}
		if str := GetSqlType(field); str != "" {
			// override type specified
			sqlField.Typ = str
		}
		if strings.Contains(strings.ToLower(sqlField.Typ), "serial") {
			sqlField.SerialTyp = true
		}
		*spec = append(*spec, sqlField)
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

func GetGenerateMatches(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateMatches, false)
}

func GetGenerateCud(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateCud, false)
}

func GetGenerateSql(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateSql, false)
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

func GetSqlSkip(field *descriptor.FieldDescriptorProto) bool {
	return proto.GetBoolExtension(field.Options, protogen.E_SqlSkip, false)
}

func GetSqlType(field *descriptor.FieldDescriptorProto) string {
	return gensupport.GetStringExtension(field.Options, protogen.E_SqlType, "")
}

func GetSqlOption(field *descriptor.FieldDescriptorProto) string {
	return gensupport.GetStringExtension(field.Options, protogen.E_SqlOption, "")
}
