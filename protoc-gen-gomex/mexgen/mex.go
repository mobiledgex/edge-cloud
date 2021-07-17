package mexgen

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/gensupport"
	"github.com/mobiledgex/edge-cloud/protogen"
	"github.com/mobiledgex/edge-cloud/util"
)

func RegisterMex() {
	generator.RegisterPlugin(new(mex))
}

func init() {
	generator.RegisterPlugin(new(mex))
}

type mex struct {
	gen                    *generator.Generator
	msgs                   map[string]*descriptor.DescriptorProto
	cudTemplate            *template.Template
	fieldsValTemplate      *template.Template
	enumTemplate           *template.Template
	cacheTemplate          *template.Template
	keysTemplate           *template.Template
	sublistLookupTemplate  *template.Template
	subfieldLookupTemplate *template.Template
	importUtil             bool
	importLog              bool
	importStrings          bool
	importErrors           bool
	importStrconv          bool
	importSort             bool
	importTime             bool
	importCmp              bool
	importReflect          bool
	importJson             bool
	firstFile              string
	support                gensupport.PluginSupport
	keyMessages            []descriptor.DescriptorProto
}

func (m *mex) Name() string {
	return "mex"
}

func (m *mex) Init(gen *generator.Generator) {
	m.gen = gen
	m.msgs = make(map[string]*descriptor.DescriptorProto)
	m.cudTemplate = template.Must(template.New("cud").Parse(cudTemplateIn))
	m.fieldsValTemplate = template.Must(template.New("fieldsVal").Parse(fieldsValTemplate))
	m.enumTemplate = template.Must(template.New("enum").Parse(enumTemplateIn))
	m.cacheTemplate = template.Must(template.New("cache").Parse(cacheTemplateIn))
	m.keysTemplate = template.Must(template.New("keys").Parse(keysTemplateIn))
	m.sublistLookupTemplate = template.Must(template.New("sublist").Parse(sublistLookupTemplateIn))
	m.subfieldLookupTemplate = template.Must(template.New("subfield").Parse(subfieldLookupTemplateIn))
	m.support.Init(gen.Request)
	m.firstFile = gensupport.GetFirstFile(gen)
}

// P forwards to g.gen.P
func (m *mex) P(args ...interface{}) {
	m.gen.P(args...)
}

func (m *mex) getAllKeyMessages() {
	for _, file := range m.gen.Request.ProtoFile {
		for _, desc := range file.MessageType {
			if GetObjKey(desc) {
				m.keyMessages = append(m.keyMessages, *desc)
			}
		}
	}
}

func (m *mex) Generate(file *generator.FileDescriptor) {
	m.support.InitFile()
	m.support.SetPbGoPackage(file.GetPackage())
	m.importUtil = false
	m.importLog = false
	m.importStrings = false
	m.importErrors = false
	m.importStrconv = false
	m.importSort = false
	m.importTime = false
	m.importCmp = false
	m.importReflect = false
	m.importJson = false
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
		m.generateEnumDecodeHook()
		m.generateShowCheck()
		m.generateAllKeyTags()
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
	if m.importJson {
		m.gen.PrintImport("", "encoding/json")
	}
	if m.importSort {
		m.gen.PrintImport("", "sort")
	}
	if m.importTime {
		m.gen.PrintImport("", "time")
	}
	if m.importReflect {
		m.gen.PrintImport("reflect", "reflect")
	}
	if m.importCmp {
		m.gen.PrintImport("", "github.com/google/go-cmp/cmp")
		m.gen.PrintImport("", "github.com/google/go-cmp/cmp/cmpopts")
	}
	m.support.PrintUsedImports(m.gen)
}

func (m *mex) generateUpgradeFuncs(enum *descriptor.EnumDescriptorProto) {
	m.P("var ", enum.Name, "_UpgradeFuncs = map[int32]VersionUpgradeFunc{")
	for _, e := range enum.Value {
		if GetUpgradeFunc(e) != "" {
			m.P(e.Number, ": ", GetUpgradeFunc(e), ",")
		} else {
			m.P(e.Number, ": nil,")
		}
	}
	m.P("}")
}
func (m *mex) generateUpgradeFuncNames(enum *descriptor.EnumDescriptorProto) {
	m.P("var ", enum.Name, "_UpgradeFuncNames = map[int32]string{")
	for _, e := range enum.Value {
		m.P(e.Number, ": ", strconv.Quote(GetUpgradeFunc(e)), ",")
	}
	m.P("}")
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
	// generate camel case maps
	fqname := m.support.FQTypeName(m.gen, desc)
	m.P("var ", fqname, "_CamelName = map[int32]string{")
	for _, val := range en.Value {
		m.P("// ", val.Name, " -> ", util.CamelCase(*val.Name))
		m.P(val.Number, ": \"", util.CamelCase(*val.Name), "\",")
	}
	m.P("}")
	m.P("var ", fqname, "_CamelValue = map[string]int32{")
	for _, val := range en.Value {
		m.P("\"", util.CamelCase(*val.Name), "\": ", val.Number, ",")
	}
	m.P("}")
	m.P()

	args := enumTempl{
		Name:         m.support.FQTypeName(m.gen, desc),
		CommonPrefix: gensupport.GetEnumCommonPrefix(en),
	}
	m.enumTemplate.Execute(m.gen.Buffer, args)
	m.importErrors = true
	m.importStrconv = true
	m.importJson = true
	m.importUtil = true
	if len(args.CommonPrefix) > 0 {
		m.importStrings = true
		m.P("var ", en.Name, "CommonPrefix = \"", args.CommonPrefix, "\"")
		m.P()
	}

	if GetVersionHashOpt(en) {
		// Collect all key objects
		m.getAllKeyMessages()
		salt := GetVersionHashSalt(en)
		hashStr := fmt.Sprintf("%x", getKeyVersionHash(m.keyMessages, salt))
		// Generate a hash of all the key messages.
		m.generateVersionString(hashStr)
		// Generate an array with function names
		m.generateUpgradeFuncs(en)
		m.generateUpgradeFuncNames(en)
		// Generate version check code for version message
		validateVersionHash(en, hashStr, file)
	}
}

type enumTempl struct {
	Name         string
	CommonPrefix string
}

var enumTemplateIn = `
func (e *{{.Name}}) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil { return err }
	val, ok := {{.Name}}_CamelValue[util.CamelCase(str)]
{{- if .CommonPrefix}}
	if !ok {
		// may have omitted common prefix
		val, ok = {{.Name}}_CamelValue["{{.CommonPrefix}}"+util.CamelCase(str)]
	}
{{- end}}
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = {{.Name}}_CamelName[val]
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("No enum value for %s", str))
	}
	*e = {{.Name}}(val)
	return nil
}

func (e {{.Name}}) MarshalYAML() (interface{}, error) {
	str := proto.EnumName({{.Name}}_CamelName, int32(e))
{{- if .CommonPrefix}}
	str = strings.TrimPrefix(str, "{{.CommonPrefix}}")
{{- end}}
	return str, nil
}

// custom JSON encoding/decoding
func (e *{{.Name}}) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		val, ok := {{.Name}}_CamelValue[util.CamelCase(str)]
{{- if .CommonPrefix}}
		if !ok {
			// may have omitted common prefix
			val, ok = {{.Name}}_CamelValue["{{.CommonPrefix}}"+util.CamelCase(str)]
		}
{{- end}}
		if !ok {
			// may be int value instead of enum name
			ival, err := strconv.Atoi(str)
			val = int32(ival)
			if err == nil {
				_, ok = {{.Name}}_CamelName[val]
			}
		}
		if !ok {
			return errors.New(fmt.Sprintf("No enum value for %s", str))
		}
		*e = {{.Name}}(val)
		return nil
	}
	var val int32
	err = json.Unmarshal(b, &val)
	if err == nil {
		*e = {{.Name}}(val)
		return nil
	}
	return fmt.Errorf("No enum value for %v", b)
}

/*
 * This is removed because we do not have enough time in
 * release 3.0 to update the SDK, UI, and documentation for this
 * change. It should be done in 3.1.
func (e {{.Name}}) MarshalJSON() ([]byte, error) {
	str := proto.EnumName({{.Name}}_CamelName, int32(e))
{{- if .CommonPrefix}}
	str = strings.TrimPrefix(str, "{{.CommonPrefix}}")
{{- end}}
	return json.Marshal(str)
}
*/
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
	repeated := false
	name := generator.CamelCase(*field.Name)
	if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED ||
		*field.Type == descriptor.FieldDescriptorProto_TYPE_BYTES {
		nilval = "nil"
		repeated = true
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
		if repeated {
			m.P("if len(m.", name, ") == 0 && len(o.", name, ") > 0 || len(m.", name, ") > 0 && len(o.", name, ") == 0 {")
		} else {
			m.P("if m.", name, " == nil && o.", name, " != nil || m.", name, " != nil && o.", name, " == nil {")
		}
		m.P("return false")
		m.P("} else if m.", name, " != nil && o.", name, "!= nil {")
	}

	mapType := m.support.GetMapType(m.gen, field)
	oName := ""
	mName := ""
	if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED ||
		*field.Type == descriptor.FieldDescriptorProto_TYPE_BYTES {
		m.P("if !opts.Filter && len(m.", name, ") != len(o.", name, ") {")
		m.P("return false")
		m.P("}")
		if mapType == nil {
			skipMatch := false
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
				if !GetGenerateMatches(subDesc.DescriptorProto) {
					skipMatch = true
				}
			}
			if !skipMatch {
				m.P("found := 0")
				m.P("for oIndex, _ := range o.", name, " {")
				m.P("for mIndex, _ := range m.", name, " {")
				oName = name + "[oIndex]"
				mName = name + "[mIndex]"
			}
		} else {
			m.P("for k, _ := range o.", name, " {")
			m.P("_, ok := m.", name, "[k]")
			m.P("if !ok {")
			m.P("return false")
			m.P("}")
			name = name + "[k]"
			field = mapType.ValField
		}
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
			if oName != "" && mName != "" {
				m.P("if m.", mName, ".Matches(", ref, "o.", oName, ", fopts...) {")
				m.P("found++")
				m.P("break")
				m.P("}")
				printedCheck = false
			} else {
				m.P("if !m.", name, ".Matches(", ref, "o.", name, ", fopts...) {")
			}
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
		if oName != "" && mName != "" {
			m.P("if o.", oName, " == m.", mName, "{")
			m.P("found++")
			m.P("break")
			m.P("}")
		} else {
			m.P("if o.", name, " != m.", name, "{")
			m.P("return false")
			m.P("}")
		}
	}
	if repeated {
		if oName != "" && mName != "" {
			m.P("}")
			m.P("}")
			m.P("if found != len(o.", name, ") {")
			m.P("return false")
			m.P("}")
		}
		if mapType != nil {
			m.P("}")
		}
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

func (m *mex) getInvalidMethodFields(names []string, subAllInvalidFields bool, desc *generator.Descriptor, method *descriptor.MethodDescriptorProto) {
	message := desc.DescriptorProto
	noconfig := gensupport.GetNoConfig(message, method)
	noconfigMap := make(map[string]string)
	for _, nc := range strings.Split(noconfig, ",") {
		if nc == "" {
			continue
		}
		noconfigMap["."+nc] = "0"
	}
	for ii, field := range message.Field {
		if ii == 0 && *field.Name == "fields" {
			continue
		}
		if keyField := gensupport.GetMessageKey(message); keyField != nil {
			if *keyField.Name == *field.Name {
				continue
			}
		}
		nilval := "0"
		nilcheck := true
		name := generator.CamelCase(*field.Name)
		if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED ||
			*field.Type == descriptor.FieldDescriptorProto_TYPE_BYTES {
			nilval = "nil"
		} else {
			switch *field.Type {
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				nilval = "nil"
				if !gogoproto.IsNullable(field) {
					nilcheck = false
				}
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				nilval = "\"\""
			case descriptor.FieldDescriptorProto_TYPE_BOOL:
				nilval = "false"
			}
		}
		fieldName := strings.Join(append(names, name), ".")
		nullableMessage := false
		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			if nilcheck {
				if _, ok := noconfigMap[fieldName]; !ok {
					for ncField, _ := range noconfigMap {
						if strings.HasPrefix(ncField, fieldName) {
							m.P("if m", fieldName, " != ", nilval, " {")
							nullableMessage = true
							break
						}
					}
				}
				if _, ok := noconfigMap[fieldName]; ok || subAllInvalidFields {
					m.P("if m", fieldName, " != ", nilval, " {")
					argStr := strings.TrimLeft(fieldName, ".")
					m.P("return fmt.Errorf(\"Invalid field specified: ", argStr, ", this field is only for internal use\")")
					m.P("}")
					continue
				}
			}
			subAllInvalidFields := false
			if _, ok := noconfigMap[fieldName]; ok {
				subAllInvalidFields = true
			}
			subDesc := gensupport.GetDesc(m.gen, field.GetTypeName())
			m.getInvalidMethodFields(append(names, name), subAllInvalidFields, subDesc, method)
			if nullableMessage {
				m.P("}")
			}
		default:
			if _, ok := noconfigMap[fieldName]; ok || subAllInvalidFields {
				m.P("if m", fieldName, " != ", nilval, " {")
				argStr := strings.TrimLeft(fieldName, ".")
				m.P("return fmt.Errorf(\"Invalid field specified: ", argStr, ", this field is only for internal use\")")
				m.P("}")
			}
		}
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

// generator.EnumDescriptor as formal arg type ?
func (m *mex) generateIsKeyField(parents, names []string, desc *generator.Descriptor) {
	message := desc.DescriptorProto
	for ii, field := range message.Field {
		if ii == 0 && *field.Name == "fields" {
			continue
		}
		name := generator.CamelCase(*field.Name)
		fieldKey := strings.Join(append(names, name), "")
		m.P("return strings.HasPrefix(s, ", fieldKey, "+\".\") || s == ", fieldKey)
		m.importStrings = true
		return
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
		nullableMessage := false
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE && gogoproto.IsNullable(field) {
			nullableMessage = true
		}
		if nullableMessage {
			m.P("if m.", hierName, " != nil && o.", hierName, " != nil {")
		}

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
		if nullableMessage {
			m.P("} else if (m.", hierName, " != nil && o.", hierName, " == nil) || (m.", hierName, " == nil && o.", hierName, " != nil) {")
			m.markDiff(names, name)
			m.P("}")
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

func (m *mex) generateMethodFields(fieldPrefix string, names []string, noconfigMap map[string]struct{}, desc *generator.Descriptor, method *descriptor.MethodDescriptorProto) {
	message := desc.DescriptorProto
	noconfig := gensupport.GetNoConfig(message, method)
	for _, nc := range strings.Split(noconfig, ",") {
		name := strings.Replace(fieldPrefix+nc, ".", "", -1)
		noconfigMap[name] = struct{}{}
	}
	for ii, field := range message.Field {
		if ii == 0 && *field.Name == "fields" {
			continue
		}
		if keyField := gensupport.GetMessageKey(message); keyField != nil {
			if *keyField.Name == *field.Name {
				continue
			}
		}
		name := generator.CamelCase(*field.Name)
		fieldName := strings.Join(append(names, name), "")
		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			if _, ok := noconfigMap[fieldName]; ok {
				continue
			}
			m.P(fieldName, ": struct{}{},")
			subDesc := gensupport.GetDesc(m.gen, field.GetTypeName())
			m.generateMethodFields(fieldPrefix, append(names, name), noconfigMap, subDesc, method)
		default:
			if _, ok := noconfigMap[fieldName]; !ok {
				m.P(fieldName, ": struct{}{},")
			}
		}
	}
}

// Generate a simple string map to use in user-friendly error messages EC-608
func (m *mex) generateAllStringFieldsMap(afg AllFieldsGen, names, nums []string, fprefix string, desc *generator.Descriptor) {

	message := desc.DescriptorProto
	for ii, field := range message.Field {
		if ii == 0 && *field.Name == "fields" {
			continue
		}
		name := generator.CamelCase(*field.Name)
		pname := name
		num := fmt.Sprintf("%d", *field.Number)

		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			subDesc := gensupport.GetDesc(m.gen, field.GetTypeName())
			m.generateAllStringFieldsMap(afg, append(names, name), append(nums, num), fprefix, subDesc)
		default:

			switch afg {

			case AllFieldsGenSlice:
				m.P(strings.Join(append(names, name), ""), ",")

			case AllFieldsGenMap:
				var readable []string
				pname = strings.Join(append(names, name, ""), "")
				m.P(fprefix, strings.Join(append(names, name), ""), ":")

				l := 0
				// take the camelcase name and insert " " before
				// each capital letter, use as the value of map
				//
				for s := pname; s != ""; s = s[l:] {
					l = strings.IndexFunc(s[1:], unicode.IsUpper) + 1
					if l <= 0 {
						l = len(s)
					}
					readable = append(readable, s[:l])
				}
				pstr := strings.Join(readable, " ") // readable?
				m.P("\"", pstr, "\"", ",")
				readable = nil
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
			m.P("if _, set := fmap[\"", numStr, "\"]; set {")
		}
		if nullableMessage || *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			m.P("if src.", hierName, " != nil {")
		}
		if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			m.printCopyInMakeArray(hierName, desc, field)
			if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
				depth := fmt.Sprintf("%d", len(parents))
				if mapType == nil {
					skipMap = true
					m.P("m.", hierName, " = src.", hierName)
					m.P("changed++")
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
			m.P("if src.", hierName, " != nil {")
			m.P("m.", hierName, " = src.", hierName)
			m.P("changed++")
			m.P("}")
		default:
			if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				m.P("m.", hierName, " = src.", hierName)
				m.P("changed++")
			} else {
				m.P("if m.", hierName, " != src.", hierName, "{")
				m.P("m.", hierName, " = src.", hierName)
				m.P("changed++")
				m.P("}")
			}
		}
		if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED && *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			if mapType != nil {
				if !mapType.ValIsMessage {
					m.P("changed++")
				}
				m.P("}")
			}
		}
		if nullableMessage || *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			m.P("} else if m.", hierName, " != nil {")
			m.P("m.", hierName, " = nil")
			m.P("changed++")
			m.P("}")
		}
		if hasGrpcFields {
			m.P("}")
		}
	}
}

func (m *mex) generateDeepCopyIn(desc *generator.Descriptor) {
	msgtyp := m.gen.TypeName(desc)
	m.P("func (m *", msgtyp, ") DeepCopyIn(src *", msgtyp, ") {")
	for ii, field := range desc.DescriptorProto.Field {
		if ii == 0 && *field.Name == "fields" {
			continue
		}
		if field.OneofIndex != nil {
			// no support
			continue
		}
		name := generator.CamelCase(*field.Name)
		nullable := false
		checkField := field
		mapType := m.support.GetMapType(m.gen, field)
		if mapType != nil {
			checkField = mapType.ValField
		}
		if *checkField.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			nullable = gogoproto.IsNullable(field)
		}
		ptr := ""
		if nullable {
			ptr = "*"
		}
		nilCheck := false
		if nullable || *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			m.P("if src.", name, " != nil {")
			nilCheck = true
		}
		if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			ftype := m.support.GoType(m.gen, field)
			if mapType == nil {
				m.P("m.", name, " = make([]", ptr, ftype, ", len(src.", name, "), len(src.", name, "))")
				m.P("for ii, s := range src.", name, " {")
				to := "m." + name + "[ii]"
				m.printCopyVar(field, to, "s", nullable, mapType)
			} else {
				m.P("m.", name, " = make(map[", mapType.KeyType, "]", ptr, mapType.ValType, ")")
				m.P("for k, v := range src.", name, " {")
				to := "m." + name + "[k]"
				m.printCopyVar(mapType.ValField, to, "v", nullable, mapType)
			}
			m.P("}")
		} else {
			m.printCopyVar(field, "m."+name, "src."+name, nullable, mapType)
		}
		if nilCheck {
			m.P("} else {")
			m.P("m.", name, " = nil")
			m.P("}")
		}
	}
	m.P("}")
	m.P()
}

func (m *mex) printCopyVar(field *descriptor.FieldDescriptorProto, to, from string, nullable bool, mapType *gensupport.MapType) {
	deepCopy := false
	if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		desc := gensupport.GetDesc(m.gen, field.GetTypeName())
		// check if this object is from our proto files
		// (as opposed to something like google_protobuf.Timestamp)
		if m.support.GenFile(*desc.File().Name) {
			deepCopy = true
		}
	}
	tmp := to
	deepCopyRef := "&"
	copyDeref := ""
	if nullable {
		// use temp var
		tmp = strings.TrimPrefix(from, "src.")
		tmp = "tmp_" + tmp
		ftype := m.support.GoType(m.gen, field)
		m.P("var ", tmp, " ", ftype)
		deepCopyRef = ""
		copyDeref = "*"
	}
	if deepCopy {
		m.P(tmp, ".DeepCopyIn(", deepCopyRef, from, ")")
	} else {
		m.P(tmp, " = ", copyDeref, from)
	}
	if nullable {
		m.P(to, " = &", tmp)
	}
}

type cudTemplateArgs struct {
	Name        string
	KeyType     string
	CudName     string
	HasFields   bool
	GenCache    bool
	NotifyCache bool
	ObjAndKey   bool
}

var fieldsValTemplate = `
func (m *{{.Name}}) ValidateUpdateFields() error {
	if m.Fields == nil {
		return fmt.Errorf("nothing specified to update")
	}
	fmap := MakeFieldMap(m.Fields)
	badFieldStrs := []string{}
	for field, _ := range fmap {
		if m.IsKeyField(field) {
			continue
		}
		if _, ok := Update{{.Name}}FieldsMap[field]; !ok {
			if _, ok := {{.Name}}AllFieldsStringMap[field]; !ok {
				continue
			}
			badFieldStrs = append(badFieldStrs, {{.Name}}AllFieldsStringMap[field])
		}
	}
	if len(badFieldStrs) > 0 {
		return fmt.Errorf("specified field(s) %s cannot be modified", strings.Join(badFieldStrs, ","))
	}
	return nil
}

`

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

func (s *{{.Name}}Store) Create(ctx context.Context, m *{{.Name}}, wait func(int64)) (*Result, error) {
{{- if (.ObjAndKey)}}
	err := m.ValidateKey()
{{- else if (.HasFields)}}
	err := m.Validate({{.Name}}AllFieldsMap)
{{- else}}
	err := m.Validate(nil)
{{- end}}
	if err != nil { return nil, err }
	key := objstore.DbKeyString("{{.Name}}", m.GetKey())
	val, err := json.Marshal(m)
	if err != nil { return nil, err }
	rev, err := s.kvstore.Create(ctx, key, string(val))
	if err != nil { return nil, err }
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *{{.Name}}Store) Update(ctx context.Context, m *{{.Name}}, wait func(int64)) (*Result, error) {
{{- if (.ObjAndKey)}}
	err := m.ValidateKey()
{{- else if (.HasFields)}}
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
	rev, err := s.kvstore.Update(ctx, key, string(val), vers)
	if err != nil { return nil, err }
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *{{.Name}}Store) Put(ctx context.Context, m *{{.Name}}, wait func(int64), ops ...objstore.KVOp) (*Result, error) {
{{- if (.ObjAndKey)}}
	err := m.ValidateKey()
{{- else if (.HasFields)}}
	err := m.Validate({{.Name}}AllFieldsMap)
	m.Fields = nil
{{- else}}
	err := m.Validate(nil)
{{- end}}
	if err != nil { return nil, err }
	key := objstore.DbKeyString("{{.Name}}", m.GetKey())
	var val []byte
	val, err = json.Marshal(m)
	if err != nil { return nil, err }
	rev, err := s.kvstore.Put(ctx, key, string(val), ops...)
	if err != nil { return nil, err }
	if wait != nil {
		wait(rev)
	}
	return &Result{}, err
}

func (s *{{.Name}}Store) Delete(ctx context.Context, m *{{.Name}}, wait func(int64)) (*Result, error) {
	err := m.GetKey().ValidateKey()
	if err != nil { return nil, err }
	key := objstore.DbKeyString("{{.Name}}", m.GetKey())
	rev, err := s.kvstore.Delete(ctx, key)
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
		log.DebugLog(log.DebugLevelApi, "Failed to parse {{.Name}} data", "val", string(val), "err", err)
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

func (s *{{.Name}}Store) STMPut(stm concurrency.STM, obj *{{.Name}}, ops ...objstore.KVOp) {
	keystr := objstore.DbKeyString("{{.Name}}", obj.GetKey())
	val, err := json.Marshal(obj)
	if err != nil {
		log.InfoLog("{{.Name}} json marsahal failed", "obj", obj, "err", err)
	}
	v3opts := GetSTMOpts(ops...)
	stm.Put(keystr, string(val), v3opts...)
}

func (s *{{.Name}}Store) STMDel(stm concurrency.STM, key *{{.KeyType}}) {
	keystr := objstore.DbKeyString("{{.Name}}", key)
	stm.Del(keystr)
}

`

type cacheTemplateArgs struct {
	Name          string
	KeyType       string
	CudCache      bool
	NotifyCache   bool
	NotifyFlush   bool
	WaitForState  string
	ObjAndKey     bool
	CustomKeyType string
}

var cacheTemplateIn = `
type {{.Name}}KeyWatcher struct {
	cb func(ctx context.Context)
}

type {{.Name}}CacheData struct {
	Obj *{{.Name}}
	ModRev int64
}

// {{.Name}}Cache caches {{.Name}} objects in memory in a hash table
// and keeps them in sync with the database.
type {{.Name}}Cache struct {
	Objs map[{{.KeyType}}]*{{.Name}}CacheData
	Mux util.Mutex
	List map[{{.KeyType}}]struct{}
	FlushAll bool
	NotifyCbs []func(ctx context.Context, obj *{{.KeyType}}, old *{{.Name}}, modRev int64)
	UpdatedCbs []func(ctx context.Context, old *{{.Name}}, new *{{.Name}})
	DeletedCbs []func(ctx context.Context, old *{{.Name}})
	KeyWatchers map[{{.KeyType}}][]*{{.Name}}KeyWatcher
	UpdatedKeyCbs []func(ctx context.Context, key *{{.KeyType}})
	DeletedKeyCbs []func(ctx context.Context, key *{{.KeyType}})
}

func New{{.Name}}Cache() *{{.Name}}Cache {
	cache := {{.Name}}Cache{}
	Init{{.Name}}Cache(&cache)
	return &cache
}

func Init{{.Name}}Cache(cache *{{.Name}}Cache) {
	cache.Objs = make(map[{{.KeyType}}]*{{.Name}}CacheData)
	cache.KeyWatchers = make(map[{{.KeyType}}][]*{{.Name}}KeyWatcher)
	cache.NotifyCbs = nil
	cache.UpdatedCbs = nil
	cache.DeletedCbs = nil
	cache.UpdatedKeyCbs = nil
	cache.DeletedKeyCbs = nil
}

func (c *{{.Name}}Cache) GetTypeString() string {
	return "{{.Name}}"
}

func (c *{{.Name}}Cache) Get(key *{{.KeyType}}, valbuf *{{.Name}}) bool {
	var modRev int64
	return c.GetWithRev(key, valbuf, &modRev)
}

func (c *{{.Name}}Cache) GetWithRev(key *{{.KeyType}}, valbuf *{{.Name}}, modRev *int64) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	inst, found := c.Objs[*key]
	if found {
		valbuf.DeepCopyIn(inst.Obj)
		*modRev = inst.ModRev
	}
	return found
}

func (c *{{.Name}}Cache) HasKey(key *{{.KeyType}}) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	_, found := c.Objs[*key]
	return found
}

func (c *{{.Name}}Cache) GetAllKeys(ctx context.Context, cb func(key *{{.KeyType}}, modRev int64)) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for key, data := range c.Objs {
		cb(&key, data.ModRev)
	}
}

func (c *{{.Name}}Cache) Update(ctx context.Context, in *{{.Name}}, modRev int64) {
	c.UpdateModFunc(ctx, in.GetKey(), modRev, func(old *{{.Name}}) (*{{.Name}}, bool) {
		return in, true
	})
}

func (c *{{.Name}}Cache) UpdateModFunc(ctx context.Context, key *{{.KeyType}}, modRev int64, modFunc func(old *{{.Name}}) (new *{{.Name}}, changed bool)) {
	c.Mux.Lock()
	var old *{{.Name}}
	if oldData, found := c.Objs[*key]; found {
		old = oldData.Obj
	}
	new, changed := modFunc(old)
	if !changed {
		c.Mux.Unlock()
		return
	}
	for _, cb := range c.UpdatedCbs {
		newCopy := &{{.Name}}{}
		newCopy.DeepCopyIn(new)
		defer cb(ctx, old, newCopy)
	}
	for _, cb := range c.NotifyCbs {
		if cb != nil {
			defer cb(ctx, new.GetKey(), old, modRev)
		}
	}
	for _, cb := range c.UpdatedKeyCbs {
		defer cb(ctx, key)
	}
	store := &{{.Name}}{}
	store.DeepCopyIn(new)
	c.Objs[new.GetKeyVal()] = &{{.Name}}CacheData{
		Obj: store,
		ModRev: modRev,
	}
	log.SpanLog(ctx, log.DebugLevelApi, "cache update", "new", store)
	c.Mux.Unlock()
	c.TriggerKeyWatchers(ctx, new.GetKey())
}

func (c *{{.Name}}Cache) Delete(ctx context.Context, in *{{.Name}}, modRev int64) {
	c.DeleteCondFunc(ctx, in, modRev, func(old *{{.Name}}) bool {
		return true
	})
}

func (c *{{.Name}}Cache) DeleteCondFunc(ctx context.Context, in *{{.Name}}, modRev int64, condFunc func(old *{{.Name}}) bool) {
	c.Mux.Lock()
	var old *{{.Name}}
	oldData, found := c.Objs[in.GetKeyVal()]
	if found {
		old = oldData.Obj
		if !condFunc(old) {
			c.Mux.Unlock()
			return
		}
	}
	delete(c.Objs, in.GetKeyVal())
	log.SpanLog(ctx, log.DebugLevelApi, "cache delete")
	c.Mux.Unlock()
	for _, cb := range c.NotifyCbs {
		if cb != nil {
			cb(ctx, in.GetKey(), old, modRev)
		}
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

func (c *{{.Name}}Cache) Prune(ctx context.Context, validKeys map[{{.KeyType}}]struct{}) {
	notify := make(map[{{.KeyType}}]*{{.Name}}CacheData)
	c.Mux.Lock()
	for key, _ := range c.Objs {
		if _, ok := validKeys[key]; !ok {
			if len(c.NotifyCbs) > 0 || len(c.DeletedKeyCbs) > 0 || len(c.DeletedCbs) > 0 {
				notify[key] = c.Objs[key]
			}
			delete(c.Objs, key)
		}
	}
	c.Mux.Unlock()
	for key, old := range notify {
	        for _, cb := range c.NotifyCbs {
			if cb != nil {
				cb(ctx, &key, old.Obj, old.ModRev)
			}
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

func (c *{{.Name}}Cache) GetCount() int {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	return len(c.Objs)
}

func (c *{{.Name}}Cache) Flush(ctx context.Context, notifyId int64) {
{{- if .NotifyFlush}}
	log.SpanLog(ctx, log.DebugLevelApi, "CacheFlush {{.Name}}", "notifyId", notifyId, "FlushAll", c.FlushAll)
	flushed := make(map[{{.KeyType}}]*{{.Name}}CacheData)
	c.Mux.Lock()
	for key, val := range c.Objs {
		if !c.FlushAll && val.Obj.NotifyId != notifyId {
			continue
		}
		flushed[key] = c.Objs[key]
		log.SpanLog(ctx, log.DebugLevelApi, "CacheFlush {{.Name}} delete", "key", key)
		delete(c.Objs, key)
	}
	c.Mux.Unlock()
	if len(flushed) > 0 {
		for key, old := range flushed {
		        for _, cb := range c.NotifyCbs {
				if cb != nil {
					cb(ctx, &key, old.Obj, old.ModRev)
				}
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
{{- end}}
}

func (c *{{.Name}}Cache) Show(filter *{{.Name}}, cb func(ret *{{.Name}}) error) error {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	for _, data := range c.Objs {
{{- if .CudCache}}
		if !data.Obj.Matches(filter, MatchFilter()) {
			continue
		}
{{- end}}
		err := cb(data.Obj)
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


func (c *{{.Name}}Cache) SetNotifyCb(fn func(ctx context.Context, obj *{{.KeyType}}, old *{{.Name}}, modRev int64)) {
	c.NotifyCbs = []func(ctx context.Context, obj *{{.KeyType}}, old *{{.Name}}, modRev int64){fn}
}

func (c *{{.Name}}Cache) SetUpdatedCb(fn func(ctx context.Context, old *{{.Name}}, new *{{.Name}})) {
	c.UpdatedCbs = []func(ctx context.Context, old *{{.Name}}, new *{{.Name}}){fn}
}

func (c *{{.Name}}Cache) SetDeletedCb(fn func(ctx context.Context, old *{{.Name}})) {
	c.DeletedCbs = []func(ctx context.Context, old *{{.Name}}){fn}
}

func (c *{{.Name}}Cache) SetUpdatedKeyCb(fn func(ctx context.Context, key *{{.KeyType}})) {
	c.UpdatedKeyCbs = []func(ctx context.Context, key *{{.KeyType}}){fn}
}

func (c *{{.Name}}Cache) SetDeletedKeyCb(fn func(ctx context.Context, key *{{.KeyType}})) {
	c.DeletedKeyCbs = []func(ctx context.Context, key *{{.KeyType}}){fn}
}

func (c *{{.Name}}Cache) AddUpdatedCb(fn func(ctx context.Context, old *{{.Name}}, new *{{.Name}})) {
	c.UpdatedCbs = append(c.UpdatedCbs, fn)
}

func (c *{{.Name}}Cache) AddDeletedCb(fn func(ctx context.Context, old *{{.Name}})) {
	c.DeletedCbs = append(c.DeletedCbs, fn)
}

func (c *{{.Name}}Cache) AddNotifyCb(fn func(ctx context.Context, obj *{{.KeyType}}, old *{{.Name}}, modRev int64)) {
	c.NotifyCbs = append(c.NotifyCbs, fn)
}

func (c *{{.Name}}Cache) AddUpdatedKeyCb(fn func(ctx context.Context, key *{{.KeyType}})) {
	c.UpdatedKeyCbs = append(c.UpdatedKeyCbs, fn)
}

func (c *{{.Name}}Cache) AddDeletedKeyCb(fn func(ctx context.Context, key *{{.KeyType}})) {
	c.DeletedKeyCbs = append(c.DeletedKeyCbs, fn)
}

func (c *{{.Name}}Cache) SetFlushAll() {
	c.FlushAll = true
}

func (c *{{.Name}}Cache) WatchKey(key *{{.KeyType}}, cb func(ctx context.Context)) context.CancelFunc {
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

func (c *{{.Name}}Cache) TriggerKeyWatchers(ctx context.Context, key *{{.KeyType}}) {
	watchers := make([]*{{.Name}}KeyWatcher, 0)
	c.Mux.Lock()
	if list, ok := c.KeyWatchers[*key]; ok {
		watchers = append(watchers, list...)
	}
	c.Mux.Unlock()
	for ii, _ := range watchers {
		watchers[ii].cb(ctx)
	}
}


{{- if .CudCache}}
// Note that we explicitly ignore the global revision number, because of the way
// the notify framework sends updates (by hashing keys and doing lookups, instead
// of sequentially through a history buffer), updates may be done out-of-order
// or multiple updates compressed into one update, so the state of the cache at
// any point in time may not by in sync with a particular database revision number.

func (c *{{.Name}}Cache) SyncUpdate(ctx context.Context, key, val []byte, rev, modRev int64) {
	obj := {{.Name}}{}
	err := json.Unmarshal(val, &obj)
	if err != nil {
		log.WarnLog("Failed to parse {{.Name}} data", "val", string(val), "err", err)
		return
	}
	c.Update(ctx, &obj, modRev)
	c.Mux.Lock()
	if c.List != nil {
		c.List[obj.GetKeyVal()] = struct{}{}
	}
	c.Mux.Unlock()
}

func (c *{{.Name}}Cache) SyncDelete(ctx context.Context, key []byte, rev, modRev int64) {
	obj := {{.Name}}{}
	keystr := objstore.DbKeyPrefixRemove(string(key))
{{- if .CustomKeyType}}
	{{.KeyType}}StringParse(keystr, &obj)
{{- else}}
	{{.KeyType}}StringParse(keystr, obj.GetKey())
{{- end}}
	c.Delete(ctx, &obj, modRev)
}

func (c *{{.Name}}Cache) SyncListStart(ctx context.Context) {
	c.List = make(map[{{.KeyType}}]struct{})
}

func (c *{{.Name}}Cache) SyncListEnd(ctx context.Context) {
	deleted := make(map[{{.KeyType}}]*{{.Name}}CacheData)
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
	        for _, cb := range c.NotifyCbs {
			if cb != nil {
				cb(ctx, &key, val.Obj, val.ModRev)
			}
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
{{- end}}

{{if ne (.WaitForState) ("")}}
func (c *{{.Name}}Cache) WaitForState(ctx context.Context, key *{{.KeyType}}, targetState {{.WaitForState}}, transitionStates map[{{.WaitForState}}]struct{}, errorState {{.WaitForState}}, timeout time.Duration, successMsg string, send func(*Result) error, opts ...WaitStateOps) error {
	curState := {{.WaitForState}}_TRACKED_STATE_UNKNOWN
	done := make(chan string, 1)
	failed := make(chan bool, 1)
	var lastMsgCnt int
	var err error

	var wSpec WaitStateSpec
	for _, op := range opts {
		if err := op(&wSpec); err != nil {
		       return err
		}
	}

	var streamCancel context.CancelFunc
        if wSpec.StreamCache != nil {
		checkStreamMsg := func() {
			streamObj := StreamObj{}
			if !wSpec.StreamCache.Get(wSpec.StreamKey, &streamObj) {
				return
			}
			if len(streamObj.Status.Msgs) > 0 || streamObj.Status.MsgCount > 0 {
				if lastMsgCnt < int(streamObj.Status.MsgCount) {
					for ii := 0; ii < len(streamObj.Status.Msgs); ii++ {
						send(&Result{Message: streamObj.Status.Msgs[ii]})
						lastMsgCnt++
					}
				}
			}
		}

                streamCancel = wSpec.StreamCache.WatchKey(wSpec.StreamKey, func(ctx context.Context) {
			checkStreamMsg()
                })

		// After setting up watch, check if any status messages were received in the meantime
		checkStreamMsg()
        }

	cancel := c.WatchKey(key, func(ctx context.Context) {
		info := {{.Name}}{}
		if c.Get(key, &info) {
			curState = info.State
		} else {
			curState = {{.WaitForState}}_NOT_PRESENT
		}
		log.SpanLog(ctx, log.DebugLevelApi, "Watch event for {{.Name}}", "key", key, "state", {{.WaitForState}}_CamelName[int32(curState)])
		if curState == errorState {
			failed <- true
		} else if curState == targetState {
			msg := ""
			if curState == {{.WaitForState}}_NOT_PRESENT {
				msg = {{.WaitForState}}_CamelName[int32(curState)]
			}
			done <- msg
		}
	})
	// After setting up watch, check current state,
	// as it may have already changed to target state
	info := {{.Name}}{}
	if c.Get(key, &info) {
		curState = info.State
	} else {
		curState = {{.WaitForState}}_NOT_PRESENT
	}
	if curState == targetState {
		msg := ""
		if curState == {{.WaitForState}}_NOT_PRESENT {
			msg = {{.WaitForState}}_CamelName[int32(curState)]
		}
		done <- msg
	}

	select {
	case doneMsg := <-done:
		if doneMsg != "" {
			send(&Result{Message: doneMsg})
		}
		err = nil
		if successMsg != "" && send != nil {
			send(&Result{Message: successMsg})
		}
	case <-failed:
		if c.Get(key, &info) {
			errs := strings.Join(info.Errors, ", ")
			err = fmt.Errorf("Encountered failures: %s", errs)
		} else {
			// this shouldn't happen, since only way to get here
			// is if info state is set to Error
			err = errors.New("Unknown failure")
		}
	case <-time.After(timeout):
		hasInfo := c.Get(key, &info)
		if hasInfo && info.State == errorState {
			// error may have been sent back before watch started
			errs := strings.Join(info.Errors, ", ")
			err = fmt.Errorf("Encountered failures: %s", errs)
		} else if _, found := transitionStates[info.State]; hasInfo && found {
			// no success response, but state is a valid transition
			// state. That means work is still in progress.
			// Notify user that this is not an error.
			// Do not undo since CRM is still busy.
			if send != nil {
				msg := fmt.Sprintf("Timed out while work still in progress state %s. Please use Show{{.Name}} to check current status", {{.WaitForState}}_CamelName[int32(info.State)])
				send(&Result{Message: msg})
			}
			err = nil
		} else {
			err = fmt.Errorf("Timed out; expected state %s but is %s",
				{{.WaitForState}}_CamelName[int32(targetState)],
				{{.WaitForState}}_CamelName[int32(curState)])
		}
	}
	cancel()
	if streamCancel != nil {
		streamCancel()
	}
	// note: do not close done/failed, garbage collector will deal with it.
	return err
}
{{- end}}

`

type sublistLookupTemplateArgs struct {
	Name       string
	KeyType    string
	LookupType string
	LookupName string
}

var sublistLookupTemplateIn = `
type {{.Name}}By{{.LookupName}} struct {
	{{.LookupName}}s map[{{.LookupType}}]map[{{.KeyType}}]struct{}
	Mux util.Mutex
}

func (s *{{.Name}}By{{.LookupName}}) Init() {
	s.{{.LookupName}}s = make(map[{{.LookupType}}]map[{{.KeyType}}]struct{})
}

func (s *{{.Name}}By{{.LookupName}}) Updated(old *{{.Name}}, new *{{.Name}}) map[{{.LookupType}}]struct{} {
	// the below func must be implemented by the user:
	// {{.Name}}.Get{{.LookupName}}s() map[{{.LookupType}}]struct{}
	old{{.LookupName}}s := make(map[{{.LookupType}}]struct{})
	if old != nil {
		old{{.LookupName}}s = old.Get{{.LookupName}}s()
	}
	new{{.LookupName}}s := new.Get{{.LookupName}}s()

	for lookup, _ := range old{{.LookupName}}s {
		if _, found := new{{.LookupName}}s[lookup]; found {
			delete(old{{.LookupName}}s, lookup)
			delete(new{{.LookupName}}s, lookup)
		}
	}

	s.Mux.Lock()
	defer s.Mux.Unlock()

	changed := make(map[{{.LookupType}}]struct{})
	for lookup, _ := range old{{.LookupName}}s {
		// remove
		s.removeRef(lookup, old.GetKeyVal())
		changed[lookup] = struct{}{}
	}
	for lookup, _ := range new{{.LookupName}}s {
		// add
		s.addRef(lookup, new.GetKeyVal())
		changed[lookup] = struct{}{}
	}
	return changed
}

func (s *{{.Name}}By{{.LookupName}}) Deleted(old *{{.Name}}) {
	old{{.LookupName}}s := old.Get{{.LookupName}}s()

	s.Mux.Lock()
	defer s.Mux.Unlock()

	for lookup, _ := range old{{.LookupName}}s {
		s.removeRef(lookup, old.GetKeyVal())
	}
}

func (s *{{.Name}}By{{.LookupName}}) addRef(lookup {{.LookupType}}, key {{.KeyType}}) {
	{{.KeyType}}s, found := s.{{.LookupName}}s[lookup]
	if !found {
		{{.KeyType}}s = make(map[{{.KeyType}}]struct{})
		s.{{.LookupName}}s[lookup] = {{.KeyType}}s
	}
	{{.KeyType}}s[key] = struct{}{}
}

func (s *{{.Name}}By{{.LookupName}}) removeRef(lookup {{.LookupType}}, key {{.KeyType}}) {
	{{.KeyType}}s, found := s.{{.LookupName}}s[lookup]
	if found {
		delete({{.KeyType}}s, key)
		if len({{.KeyType}}s) == 0 {
			delete(s.{{.LookupName}}s, lookup)
		}
	}
}

func (s *{{.Name}}By{{.LookupName}}) Find(lookup {{.LookupType}}) []{{.KeyType}} {
	s.Mux.Lock()
	defer s.Mux.Unlock()

	list := []{{.KeyType}}{}
	for k, _ := range s.{{.LookupName}}s[lookup] {
		list = append(list, k)
	}
	return list
}

func (s *{{.Name}}By{{.LookupName}}) HasRef(lookup {{.LookupType}}) bool {
	s.Mux.Lock()
	defer s.Mux.Unlock()

	_, found := s.{{.LookupName}}s[lookup]
	return found
}

// Convert to dumpable format. JSON cannot marshal maps with struct keys.
func (s *{{.Name}}By{{.LookupName}}) Dumpable() map[string]interface{} {
	s.Mux.Lock()
	defer s.Mux.Unlock()

	dat := make(map[string]interface{})
	for lookup, keys := range s.{{.LookupName}}s {
		keystrs := make(map[string]interface{})
		for k, _ := range keys {
			keystrs[k.GetKeyString()] = struct{}{}
		}
		dat[lookup.GetKeyString()] = keystrs
	}
	return dat
}

`

type subfieldLookupTemplateArgs struct {
	Name        string
	KeyType     string
	LookupType  string
	LookupName  string
	LookupField string
}

var subfieldLookupTemplateIn = `
type {{.Name}}By{{.LookupName}} struct {
	{{.LookupName}}s map[{{.LookupType}}]map[{{.KeyType}}]struct{}
	Mux util.Mutex
}

func (s *{{.Name}}By{{.LookupName}}) Init() {
	s.{{.LookupName}}s = make(map[{{.LookupType}}]map[{{.KeyType}}]struct{})
}

func (s *{{.Name}}By{{.LookupName}}) Updated(obj *{{.Name}}) {
	lookup := obj.{{.LookupField}}

	s.Mux.Lock()
	defer s.Mux.Unlock()

	{{.KeyType}}s, found := s.{{.LookupName}}s[lookup]
	if !found {
		{{.KeyType}}s = make(map[{{.KeyType}}]struct{})
		s.{{.LookupName}}s[lookup] = {{.KeyType}}s
	}
	{{.KeyType}}s[obj.GetKeyVal()] = struct{}{}
}

func (s *{{.Name}}By{{.LookupName}}) Deleted(obj *{{.Name}}) {
	lookup := obj.{{.LookupField}}

	s.Mux.Lock()
	defer s.Mux.Unlock()

	{{.KeyType}}s, found := s.{{.LookupName}}s[lookup]
	if found {
		delete({{.KeyType}}s, obj.GetKeyVal())
		if len({{.KeyType}}s) == 0 {
			delete(s.{{.LookupName}}s, lookup)
		}
	}
}

func (s *{{.Name}}By{{.LookupName}}) Find(lookup {{.LookupType}}) []{{.KeyType}} {
	s.Mux.Lock()
	defer s.Mux.Unlock()

	list := []{{.KeyType}}{}
	for k, _ := range s.{{.LookupName}}s[lookup] {
		list = append(list, k)
	}
	return list
}

`

type keysTemplateArgs struct {
	Name      string
	KeyType   string
	ObjAndKey bool
}

var keysTemplateIn = `
func (m *{{.Name}}) GetObjKey() objstore.ObjKey {
	return m.GetKey()
}

func (m *{{.Name}}) GetKey() *{{.KeyType}} {
{{- if .ObjAndKey}}
	return m
{{- else}}
	return &m.Key
{{- end}}
}

func (m *{{.Name}}) GetKeyVal() {{.KeyType}} {
{{- if .ObjAndKey}}
	return *m
{{- else}}
	return m.Key
{{- end}}
}

func (m *{{.Name}}) SetKey(key *{{.KeyType}}) {
{{- if .ObjAndKey}}
	*m = *key
{{- else}}
	m.Key = *key
{{- end}}
}

func CmpSort{{.Name}}(a {{.Name}}, b {{.Name}}) bool {
{{- if .ObjAndKey}}
	return a.GetKeyString() < b.GetKeyString()
{{- else}}
	return a.Key.GetKeyString() < b.Key.GetKeyString()
{{- end}}
}


`

type ugpradeError struct {
	CurHash        string
	CurHashEnumVal int32
	NewHash        string
	NewHashEnumVal int32
}

var upgradeErrorTemplete = `
======WARNING=======
Current data model hash({{.NewHash}}) doesn't match the latest supported one({{.CurHash}}).
This is due to an upsupported change in the key of some objects in a .proto file.
In order to ensure a smooth upgrade for the production environment please make sure to add the following to version.proto file:

enum VersionHash {
	...
	{{.CurHash}} = {{.CurHashEnumVal}};
	{{.NewHash}} = {{.NewHashEnumVal}} [(protogen.upgrade_func) = "sample_upgrade_function"]; <<<===== Add this line
	...
}

Implementation of "sample_upgrade_function" should be added tp edge-cloud/upgrade/upgrade-types.go

NOTE: If no upgrade function is needed don't need to add "[(protogen.upgrade_func) = "sample_upgrade_function];" to
the VersionHash enum.

A unit test data for the automatic unit test of the upgrade function should be added to testutil/upgrade_test_data.go
   - PreUpgradeData - what key/value objects are trying to be upgraded
   - PostUpgradeData - what the resulting object store should look like
====================
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
		m.P("var ", *message.Name, "AllFieldsStringMap = map[string]string{")
		m.generateAllStringFieldsMap(AllFieldsGenMap, []string{}, []string{}, *message.Name+"Field", desc)
		m.P("}")
		m.P("")
		m.P("func (m *", *message.Name, ") IsKeyField(s string) bool {")
		m.generateIsKeyField([]string{}, []string{*message.Name + "Field"}, desc)
		m.P("}")
		m.P("")
		m.P("func (m *", message.Name, ") DiffFields(o *", message.Name, ", fields map[string]struct{}) {")
		m.generateDiffFields([]string{}, []string{*message.Name + "Field"}, desc)
		m.P("}")
		m.P("")
		for _, service := range file.Service {
			if *service.Name != *message.Name+"Api" {
				continue
			}
			if len(service.Method) == 0 {
				continue
			}
			for _, method := range service.Method {
				if gensupport.GetCamelCasePrefix(*method.Name) != "Update" {
					continue
				}
				noconfigMap := make(map[string]struct{})
				m.P("var ", *method.Name, "FieldsMap = map[string]struct{}{")
				fieldPrefix := *message.Name + "Field"
				m.generateMethodFields(fieldPrefix, []string{fieldPrefix}, noconfigMap, desc, method)
				m.P("}")
				m.P("")
				args := cudTemplateArgs{
					Name: *message.Name,
				}
				m.fieldsValTemplate.Execute(m.gen.Buffer, args)
				break
			}
		}
	}

	if desc.GetOptions().GetMapEntry() {
		return
	}

	msgtyp := m.gen.TypeName(desc)
	if GetGenerateCopyInFields(message) {
		m.P("func (m *", msgtyp, ") CopyInFields(src *", msgtyp, ") int {")
		m.P("changed := 0")
		if gensupport.HasGrpcFields(message) {
			m.P("fmap := MakeFieldMap(src.Fields)")
		}
		m.generateCopyIn(make([]string, 0), make([]string, 0), desc, make([]*generator.Descriptor, 0), gensupport.HasGrpcFields(message))
		m.P("return changed")
		m.P("}")
		m.P("")
	}

	m.generateDeepCopyIn(desc)

	if GetGenerateCud(message) {
		keyType, err := m.support.GetMessageKeyType(m.gen, desc)
		if err != nil {
			m.gen.Fail(err.Error())
		}
		args := cudTemplateArgs{
			Name:      *message.Name,
			CudName:   *message.Name + "Cud",
			HasFields: gensupport.HasGrpcFields(message),
			ObjAndKey: gensupport.GetObjAndKey(message),
			KeyType:   keyType,
		}
		m.cudTemplate.Execute(m.gen.Buffer, args)
		m.importLog = true
	}
	if GetGenerateCache(message) {
		keyType, err := m.support.GetMessageKeyType(m.gen, desc)
		if err != nil {
			m.gen.Fail(err.Error())
		}
		args := cacheTemplateArgs{
			Name:          *message.Name,
			CudCache:      GetGenerateCud(message),
			NotifyCache:   GetNotifyCache(message),
			NotifyFlush:   GetNotifyFlush(message),
			WaitForState:  GetGenerateWaitForState(message),
			ObjAndKey:     gensupport.GetObjAndKey(message),
			CustomKeyType: gensupport.GetCustomKeyType(message),
			KeyType:       keyType,
		}
		m.cacheTemplate.Execute(m.gen.Buffer, args)
		m.importUtil = true
		m.importLog = true
		if args.WaitForState != "" {
			m.importErrors = true
			m.importTime = true
			m.importStrings = true
		}
		m.generateUsesOrg(message)
	}
	if lookups := GetGenerateLookupBySublist(message); lookups != "" {
		keyType, err := m.support.GetMessageKeyType(m.gen, desc)
		if err != nil {
			m.gen.Fail(err.Error())
		}
		list := strings.Split(lookups, ",")
		for _, lookup := range list {
			lookup = strings.TrimSpace(lookup)
			nameType := strings.Split(lookup, ":")
			args := sublistLookupTemplateArgs{
				Name:       *message.Name,
				KeyType:    keyType,
				LookupType: nameType[0],
				LookupName: nameType[0],
			}
			if len(nameType) > 1 {
				args.LookupName = nameType[1]
			}
			m.sublistLookupTemplate.Execute(m.gen.Buffer, args)
			m.importUtil = true
			m.importJson = true
		}
	}
	if lookups := GetGenerateLookupBySubfield(message); lookups != "" {
		keyType, err := m.support.GetMessageKeyType(m.gen, desc)
		if err != nil {
			m.gen.Fail(err.Error())
		}
		list := strings.Split(lookups, ",")
		for _, lookup := range list {
			lookup = strings.TrimSpace(lookup)
			_, field, err := gensupport.FindHierField(m.gen, message, lookup)
			if err != nil {
				m.gen.Fail(err.Error())
				continue
			}
			args := subfieldLookupTemplateArgs{
				Name:        *message.Name,
				KeyType:     keyType,
				LookupType:  m.support.GoType(m.gen, field),
				LookupName:  m.support.GoType(m.gen, field),
				LookupField: lookup,
			}
			m.subfieldLookupTemplate.Execute(m.gen.Buffer, args)
			m.importUtil = true
		}
	}
	if GetObjKey(message) || gensupport.GetObjAndKey(message) {
		// this is a key object
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

		m.P("func (m *", message.Name, ") NotFoundError() error {")
		m.P("return fmt.Errorf(\"", strings.TrimSuffix(*message.Name, "Key"), " key %s not found\", m.GetKeyString())")
		m.P("}")
		m.P("")

		m.P("func (m *", message.Name, ") ExistsError() error {")
		m.P("return fmt.Errorf(\"", strings.TrimSuffix(*message.Name, "Key"), " key %s already exists\", m.GetKeyString())")
		m.P("}")
		m.P("")

		hasKeyTags := false
		for _, field := range message.Field {
			if field.Type == nil || field.OneofIndex != nil {
				continue
			}
			if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
				continue
			}
			tag := GetKeyTag(field)
			if tag == "" {
				m.gen.Fail(*message.Name, "field", *field.Name, "missing protogen.keytag")
			}
			fname := generator.CamelCase(*field.Name)
			m.P("var ", message.Name, "Tag", fname, " = \"", tag, "\"")
			hasKeyTags = true
		}
		if hasKeyTags {
			m.P()
		}
		m.P("func (m *", message.Name, ") GetTags() map[string]string {")
		m.P("tags := make(map[string]string)")
		m.setKeyTags([]string{}, desc, []*generator.Descriptor{})
		m.P("return tags")
		m.P("}")
		m.P()

		m.importJson = true
		m.importLog = true
	}
	if gensupport.GetMessageKey(message) != nil || gensupport.GetObjAndKey(message) {
		// this is an object that has a key field
		keyType, err := m.support.GetMessageKeyType(m.gen, desc)
		if err != nil {
			m.gen.Fail(err.Error())
		}
		args := keysTemplateArgs{
			Name:      *message.Name,
			KeyType:   keyType,
			ObjAndKey: gensupport.GetObjAndKey(message),
		}
		m.keysTemplate.Execute(m.gen.Buffer, args)
	}

	//Generate enum values validation
	m.generateEnumValidation(message, desc)

	visited := make([]*generator.Descriptor, 0)
	if gensupport.HasHideTags(m.gen, desc, protogen.E_Hidetag, visited) {
		m.generateHideTags(desc)
	}
}

func (m *mex) generateVersionString(hashStr string) {
	m.P("// Keys being hashed:")
	for _, v := range m.keyMessages {
		m.P("// ", v.Name)
	}
	m.P("var versionHashString = \"", hashStr, "\"")
	m.P("")
	m.P("func GetDataModelVersion() string {")
	m.P("return versionHashString")
	m.P("}")
}

func validateVersionHash(en *descriptor.EnumDescriptorProto, hashStr string, file *generator.FileDescriptor) {
	// We need to check the hash and verify that we have the correct one
	// If we don't have a correct one fail suggesting an upgrade function
	// Check the last one(it's the latest) and if it doesn't match fail
	lastIndex := 0
	for i, _ := range en.Value {
		if i > lastIndex {
			lastIndex = i
		}
	}
	latestVer := en.Value[lastIndex]
	// Check the substring of the value
	if !strings.Contains(*latestVer.Name, hashStr) {
		var upgradeTemplate *template.Template
		upgradeTemplate = template.Must(template.New("upgrade").Parse(upgradeErrorTemplete))
		buf := bytes.Buffer{}
		upgErr := ugpradeError{
			CurHash:        *latestVer.Name,
			CurHashEnumVal: *latestVer.Number,
			NewHash:        "HASH_" + hashStr,
			NewHashEnumVal: *latestVer.Number + 1,
		}
		if err := upgradeTemplate.Execute(&buf, &upgErr); err != nil {
			log.Fatalf("Cannot execute upgrade error template %v\n", err)
		}
		log.Fatalf("%s", buf.String())
	}
}

// Subset of the FieldDescriptorProto that is used to identify whether we need to trigger an
// incompatible upgrade or not
type FieldDescriptorProtoHashable struct {
	Name     *string
	Number   *int32
	Label    *descriptor.FieldDescriptorProto_Label
	Type     *descriptor.FieldDescriptorProto_Type
	TypeName *string
	Extendee *string
}

// Unique idenitifable message object, which is used in a version hash calculation
type HashableKey struct {
	Name  *string
	Field []FieldDescriptorProtoHashable
}

// This function generates an array of HashableKey[s] from an array of the DescriptorProto
// messages. HashableKey is defined to keep track of only the specific sets of
// fields of the DescriptorProto which make it unique
// NOTE: There is a possiblity that some of the sub-strucutres of the key messages
// is not an obj_key itself and we might miss it in a hash calculation.
// If this ever becomes a problem we should make sure to track all the sub-structs that
// are not key_obj
func getHashObjsFromMsgs(msgs []descriptor.DescriptorProto) []HashableKey {
	objs := make([]HashableKey, 0)
	for _, m := range msgs {
		o := HashableKey{
			Name: m.Name,
		}
		for _, dp := range m.Field {
			dpHash := FieldDescriptorProtoHashable{
				Name:     dp.Name,
				Number:   dp.Number,
				Label:    dp.Label,
				Type:     dp.Type,
				TypeName: dp.TypeName,
				Extendee: dp.Extendee,
			}
			o.Field = append(o.Field, dpHash)
		}
		objs = append(objs, o)
	}
	return objs
}

// Hash function for the Data Model Version
func getKeyVersionHash(msgs []descriptor.DescriptorProto, salt string) [16]byte {
	// Sort the messages to make sure we are generate repeatable hash
	sort.Slice(msgs, func(i, j int) bool {
		return *msgs[i].Name < *msgs[j].Name
	})
	// Need to build an array of HashableKeys from msgs
	hashObjs := getHashObjsFromMsgs(msgs)
	arrBytes := []byte{}
	for _, o := range hashObjs {
		jsonBytes, _ := json.Marshal(o)
		arrBytes = append(arrBytes, jsonBytes...)
	}
	// add salt
	arrBytes = append(arrBytes, []byte(salt)...)
	return md5.Sum(arrBytes)

}

// Generate a single check for an enum
func (m *mex) generateEnumCheck(field *descriptor.FieldDescriptorProto, elem string) {
	m.P("if _, ok := ", m.support.GoType(m.gen, field), "_name[int32(", elem,
		")]; !ok {")
	m.P("return errors.New(\"invalid ", generator.CamelCase(*field.Name),
		"\")")
	m.P("}")
	m.importErrors = true
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

func (m *mex) generateHideTags(desc *generator.Descriptor) {
	msgName := desc.DescriptorProto.Name
	m.P("func Ignore", msgName, "Fields(taglist string) cmp.Option {")
	m.P("names := []string{}")
	m.P("tags := make(map[string]struct{})")
	m.P("for _, tag := range strings.Split(taglist, \",\") {")
	m.P("tags[tag] = struct{}{}")
	m.P("}")
	visited := make([]*generator.Descriptor, 0)
	m.generateHideTagFields(make([]string, 0), desc, visited)
	m.P("return cmpopts.IgnoreFields(", msgName, "{}, names...)")
	m.P("}")
	m.P()
	m.importStrings = true
	m.importCmp = true
}

func (m *mex) generateHideTagFields(parents []string, desc *generator.Descriptor, visited []*generator.Descriptor) {
	if gensupport.WasVisited(desc, visited) {
		return
	}
	msg := desc.DescriptorProto
	for _, field := range msg.Field {
		if field.Type == nil || field.OneofIndex != nil {
			continue
		}
		tag := GetHideTag(field)
		if tag == "" && *field.Type != descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			continue
		}
		name := generator.CamelCase(*field.Name)
		hierField := strings.Join(append(parents, name), ".")

		if tag != "" {
			m.P("if _, found := tags[\"", tag, "\"]; found {")
			m.P("names = append(names, \"", hierField, "\")")
			m.P("}")
		}
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			subDesc := gensupport.GetDesc(m.gen, field.GetTypeName())
			m.generateHideTagFields(append(parents, name),
				subDesc, append(visited, desc))
		}
	}
}

func (m *mex) setKeyTags(parents []string, desc *generator.Descriptor, visited []*generator.Descriptor) {
	for _, field := range desc.DescriptorProto.Field {
		if field.Type == nil || field.OneofIndex != nil {
			continue
		}
		name := generator.CamelCase(*field.Name)
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			subDesc := gensupport.GetDesc(m.gen, field.GetTypeName())
			m.setKeyTags(append(parents, name),
				subDesc, append(visited, desc))
			continue
		}
		tag := GetKeyTag(field)
		hierField := strings.Join(append(parents, name), ".")
		val := "m." + hierField
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_ENUM {
			val = m.support.GoType(m.gen, field) + "_name[int32(" + val + ")]"
		}
		m.P("tags[\"", tag, "\"] = ", val)
	}
}

func (m *mex) generateEnumDecodeHook() {
	m.P("// DecodeHook for use with the mapstructure package.")
	m.P("// Allows decoding to handle protobuf enums that are")
	m.P("// represented as strings.")
	m.P("func EnumDecodeHook(from, to reflect.Type, data interface{}) (interface{}, error) {")
	m.P("if from.Kind() != reflect.String { return data, nil }")
	m.P("switch to {")
	for _, file := range m.gen.Request.ProtoFile {
		if !m.support.GenFile(*file.Name) {
			continue
		}
		for _, en := range file.EnumType {
			m.P("case reflect.TypeOf(", en.Name, "(0)):")
			m.P("if en, ok := ", en.Name, "_CamelValue[util.CamelCase(data.(string))]; ok {")
			m.P("return en, nil")
			m.P("}")
			commonPrefix := gensupport.GetEnumCommonPrefix(en)
			if commonPrefix != "" {
				m.P("if en, ok := ", en.Name, "_CamelValue[\"", commonPrefix, "\"+util.CamelCase(data.(string))]; ok {")
				m.P("return en, nil")
				m.P("}")
			}
		}
	}
	m.P("}")
	m.P("return data, nil")
	m.P("}")
	m.P()

	m.importReflect = true
}

func (m *mex) generateShowCheck() {
	m.P("var ShowMethodNames = map[string]struct{}{")
	for _, file := range m.gen.Request.ProtoFile {
		if !m.support.GenFile(*file.Name) {
			continue
		}
		if len(file.Service) == 0 {
			continue
		}
		for _, service := range file.Service {
			if len(service.Method) == 0 {
				continue
			}
			for _, method := range service.Method {
				if gensupport.IsShow(method) {
					m.P("\"", method.Name, "\": struct{}{},")
				}
			}
		}
	}
	m.P("}")
	m.P()
	m.P("func IsShow(cmd string) bool {")
	m.P("_, found := ShowMethodNames[cmd]")
	m.P("return found")
	m.P("}")
	m.P()
}

func (m *mex) generateAllKeyTags() {
	tags := make(map[string]string)
	for _, file := range m.gen.Request.ProtoFile {
		if !m.support.GenFile(*file.Name) {
			continue
		}
		for _, message := range file.MessageType {
			for _, field := range message.Field {
				if field.Type == nil || field.OneofIndex != nil {
					continue
				}
				if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
					continue
				}
				tag := GetKeyTag(field)
				if tag == "" {
					continue
				}
				if GetSkipKeyTagConflictCheck(field) {
					continue
				}

				fname := generator.CamelCase(*field.Name)
				tagLoc := *message.Name + "." + fname
				if conflict, found := tags[tag]; found {
					m.gen.Fail("KeyTag conflict for", tag, "between", tagLoc, "and", conflict)
				}
				tags[tag] = tagLoc
			}
		}
	}
	if len(tags) == 0 {
		return
	}
	list := []string{}
	for t, _ := range tags {
		list = append(list, t)
	}
	sort.Strings(list)
	m.P("var AllKeyTags = []string{")
	for _, tag := range list {
		m.P(`"`, tag, `",`)
	}
	m.P("}")
	m.P()
	m.P("var AllKeyTagsMap = map[string]struct{}{")
	for _, tag := range list {
		m.P(`"`, tag, `": struct{}{},`)
	}
	m.P("}")
	m.P()
}

func (m *mex) generateUsesOrg(message *descriptor.DescriptorProto) {
	usesOrg := GetUsesOrg(message)
	if usesOrg == "" {
		m.gen.Fail(*message.Name, "protogen.generate_cache option also requires protogen.uses_org option")
	}
	if usesOrg == "custom" {
		return
	}
	m.P()
	m.P("func (c *", message.Name, "Cache) UsesOrg(org string) bool {")
	if usesOrg == "none" {
		m.P("return false")
		m.P("}")
		return
	}
	keyIter := "_"
	valIter := "_"
	valObj := ""
	usesChecks := strings.Split(usesOrg, ",")
	kvChecks := [][]string{}
	for _, check := range usesChecks {
		kv := strings.Split(check, "=")
		if len(kv) != 2 {
			m.gen.Fail(*message.Name, "invalid uses_org check spec, expected a=b but was ", check)
			continue
		}
		if kv[0] == "key" {
			keyIter = "key"
		} else if kv[0] == "val" {
			valIter = "val"
			valObj = ".Obj"
		} else {
			m.gen.Fail(*message.Name, "invalid key in uses_org check spec, expected \"key\" or \"val\", but was ", kv[0])
		}
		kvChecks = append(kvChecks, kv)
	}
	m.P("c.Mux.Lock()")
	m.P("defer c.Mux.Unlock()")
	m.P("for ", keyIter, ", ", valIter, " := range c.Objs {")
	for _, kv := range kvChecks {
		m.P("if ", kv[0], valObj, ".", kv[1], " == org { return true }")
	}
	m.P("}")
	m.P("return false")
	m.P("}")
}

func (m *mex) generateService(file *generator.FileDescriptor, service *descriptor.ServiceDescriptorProto) {
	if len(service.Method) != 0 {
		for _, method := range service.Method {
			m.generateMethod(file, service, method)
		}
	}
}

func (m *mex) generateMethod(file *generator.FileDescriptor, service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	in := gensupport.GetDesc(m.gen, method.GetInputType())
	if !gensupport.IsShow(method) {
		m.P("func (m *", *in.DescriptorProto.Name, ") IsValidArgsFor", *method.Name, "() error {")
		m.getInvalidMethodFields([]string{""}, false, in, method)
		m.P("return nil")
		m.P("}")
		m.P("")
	}
}

func GetGenerateMatches(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateMatches, false)
}

func GetGenerateCopyInFields(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateCopyInFields, true)
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

func GetGenerateLookupBySublist(message *descriptor.DescriptorProto) string {
	return gensupport.GetStringExtension(message.Options, protogen.E_GenerateLookupBySublist, "")
}

func GetGenerateLookupBySubfield(message *descriptor.DescriptorProto) string {
	return gensupport.GetStringExtension(message.Options, protogen.E_GenerateLookupBySubfield, "")
}

func GetNotifyCache(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_NotifyCache, false)
}

func GetNotifyFlush(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_NotifyFlush, false)
}

func GetObjKey(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_ObjKey, false)
}

func GetUsesOrg(message *descriptor.DescriptorProto) string {
	return gensupport.GetStringExtension(message.Options, protogen.E_UsesOrg, "")
}

func GetBackend(field *descriptor.FieldDescriptorProto) bool {
	return proto.GetBoolExtension(field.Options, protogen.E_Backend, false)
}

func GetHideTag(field *descriptor.FieldDescriptorProto) string {
	return gensupport.GetStringExtension(field.Options, protogen.E_Hidetag, "")
}

func GetKeyTag(field *descriptor.FieldDescriptorProto) string {
	return gensupport.GetStringExtension(field.Options, protogen.E_Keytag, "")
}

func GetSkipKeyTagConflictCheck(field *descriptor.FieldDescriptorProto) bool {
	return proto.GetBoolExtension(field.Options, protogen.E_SkipKeytagConflictCheck, false)
}

func GetVersionHashOpt(enum *descriptor.EnumDescriptorProto) bool {
	return proto.GetBoolExtension(enum.Options, protogen.E_VersionHash, false)
}

func GetVersionHashSalt(enum *descriptor.EnumDescriptorProto) string {
	return gensupport.GetStringExtension(enum.Options, protogen.E_VersionHashSalt, "")
}

func GetUpgradeFunc(enumVal *descriptor.EnumValueDescriptorProto) string {
	return gensupport.GetStringExtension(enumVal.Options, protogen.E_UpgradeFunc, "")
}
