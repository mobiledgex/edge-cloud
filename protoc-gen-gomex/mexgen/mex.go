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
	for _, desc := range file.Messages() {
		msg := desc.DescriptorProto
		if GetGenerateCud(msg) {
			hasGenerateCud = true
		}
		if HasGrpcFields(msg) {
			hasGrpcFields = true
		}
		m.msgs[*msg.Name] = msg
	}
	if hasGenerateCud {
		m.P("import \"encoding/json\"")
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

func (m *mex) generateFieldKeyString(message *descriptor.DescriptorProto, field *descriptor.FieldDescriptorProto) {
	if field.Type == nil {
		return
	}
	name := generator.CamelCase(*field.Name)
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		m.P("s = append(s, m.", name, ".GetKeyString())")
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		// deprecated in proto3
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		// TODO
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		m.P("s = append(s, m.", name, ")")
	default:
		m.P("s = append(s, string(m.", name, "))")
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
}

var cudTemplateIn = `
type {{.CudName}} interface {
	// Validate all fields for create/update
	Validate(in *{{.Name}}) error
	// Validate only key fields for delete
	ValidateKey(key *{{.KeyName}}) error
	// Get key string for saving to persistent object storage
	GetObjStoreKeyString(key *{{.KeyName}}) string
	// Object storage IO interface
	ObjStore
	// Refresh is called after create/update/delete to update in-memory cache
	Refresh(in *{{.Name}}, key string) error
	// Get key string for loading all objects of this type
	GetLoadKeyString() string
}

func (m *{{.Name}}) Create(cud {{.CudName}}) (*Result, error) {
	err := cud.Validate(m)
	if err != nil { return nil, err }
	key := cud.GetObjStoreKeyString(&m.Key)
	val, err := json.Marshal(m)
	if err != nil { return nil, err }
	err = cud.Create(key, string(val))
	if err != nil { return nil, err }
	err = cud.Refresh(m, key)
	return &Result{}, err
}

func (m *{{.Name}}) Update(cud {{.CudName}}) (*Result, error) {
	err := cud.Validate(m)
	if err != nil { return nil, err }
	key := cud.GetObjStoreKeyString(&m.Key)
	var vers int64 = 0
{{- if (.HasFields)}}
	curBytes, vers, err := cud.Get(key)
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
	err = cud.Update(key, string(val), vers)
	if err != nil { return nil, err }
	err = cud.Refresh(m, key)
	return &Result{}, err
}

func (m *{{.Name}}) Delete(cud {{.CudName}}) (*Result, error) {
	err := cud.ValidateKey(&m.Key)
	if err != nil { return nil, err }
	key := cud.GetObjStoreKeyString(&m.Key)
	err = cud.Delete(key)
	if err != nil { return nil, err }
	err = cud.Refresh(m, key)
	return &Result{}, err
}

type LoadAll{{.Name}}sCb func(m *{{.Name}}) error

func LoadAll{{.Name}}s(cud {{.CudName}}, cb LoadAll{{.Name}}sCb) error {
	loadkey := cud.GetLoadKeyString()
	err := cud.List(loadkey, func(key, val []byte) error {
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

func LoadOne{{.Name}}(cud {{.CudName}}, key string) (*{{.Name}}, error) {
	val, _, err := cud.Get(key)
	if err != nil {
		return nil, err
	}
	var obj {{.Name}}
	err = json.Unmarshal(val, &obj)
	if err != nil {
		util.DebugLog(util.DebugLevelApi, "Failed to parse {{.Name}} data", "val", string(val))
		return nil, err
	}
	return &obj, nil
}

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
		}
		m.cudTemplate.Execute(m.gen.Buffer, args)
	}
	if GetObjKey(message) {
		m.P("func (m *", message.Name, ") GetKeyString() string {")
		m.P("s := make([]string, 0, ", len(message.Field), ")")
		for _, field := range message.Field {
			m.generateFieldKeyString(message, field)
		}
		m.P("return strings.Join(s, \"/\")")
		m.P("}")
		m.P("")
	}
	if HasMessageKey(message) {
		m.P("func (m *", message.Name, ") GetKey() ObjKey {")
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

func GetObjKey(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_ObjKey, false)
}
