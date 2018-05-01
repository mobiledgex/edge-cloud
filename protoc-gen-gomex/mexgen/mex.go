package mexgen

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/mobiledgex/edge-cloud/protogen"
)

func RegisterMex() {
	generator.RegisterPlugin(new(mex))
}

func init() {
	generator.RegisterPlugin(new(mex))
}

type mex struct {
	gen  *generator.Generator
	msgs map[string]*descriptor.DescriptorProto
}

func (m *mex) Name() string {
	return "mex"
}

func (m *mex) Init(gen *generator.Generator) {
	m.gen = gen
	m.msgs = make(map[string]*descriptor.DescriptorProto)
}

// P forwards to g.gen.P
func (m *mex) P(args ...interface{}) {
	m.gen.P(args...)
}

func (m *mex) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.MessageType) != 0 {
		for _, msg := range file.FileDescriptorProto.MessageType {
			m.generateMessage(file, msg)
		}
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
	if len(file.FileDescriptorProto.MessageType) != 0 {
		for _, msg := range file.FileDescriptorProto.MessageType {
			if GetGenerateCud(msg) {
				hasGenerateCud = true
			}
			if HasGrpcFields(msg) {
				hasGrpcFields = true
			}
			m.msgs[*msg.Name] = msg
		}
	}
	if hasGenerateCud {
		m.P("import \"encoding/json\"")
	}
	if hasGrpcFields {
		m.P("import \"github.com/mobiledgex/edge-cloud/util\"")
	}
}

func (m *mex) generateMessage(file *generator.FileDescriptor, message *descriptor.DescriptorProto) {
	if GetGenerateMatches(message) && message.Field != nil {
		m.P("func (m *", message.Name, ") Matches(filter *", message.Name, ") bool {")
		m.P("if filter == nil { return true }")
		for _, field := range message.Field {
			if field.Type == nil {
				continue
			}
			name := generator.CamelCase(*field.Name)
			switch *field.Type {
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				m.P("if filter.", name, " != nil && !filter.", name, ".Matches(m.", name, ") {")
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
	m.P("func (m *", message.Name, ") CopyInFields(src *", message.Name, ") {")
	for ii, field := range message.Field {
		if ii == 0 && *field.Name == "fields" {
			// skip fields
			continue
		}
		if field.OneofIndex != nil {
			// no support for copy OneOf fields
			continue
		}
		name := generator.CamelCase(*field.Name)
		if HasGrpcFields(message) {
			m.P("if set, _ := util.GrpcFieldsGet(src.Fields, ", field.Number, "); set == true {")
		}
		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			msg, ok := m.msgs[*field.TypeName]
			if !ok || !HasGrpcFields(msg) {
				m.P("if m.", name, " != nil && src.", name, " != nil {")
				m.P("*m.", name, " = *src.", name)
				m.P("}")
			} else {
				m.P("m.", name, ".CopyInFields(src.", name, ")")
			}
		case descriptor.FieldDescriptorProto_TYPE_GROUP:
			// deprecated in proto3
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			m.P("copy(m.", name, ", src.", name, ")")
		default:
			m.P("m.", name, " = src.", name)
		}
		if HasGrpcFields(message) {
			m.P("}")
		}
	}
	m.P("}")

	if GetGenerateCud(message) {
		m.P("type ", message.Name, "Cud interface {")
		m.P("// Validate all fields for create/update")
		m.P("Validate(in *", message.Name, ") error")
		m.P("// Validate only Key fields for delete")
		m.P("ValidateKey(in *", message.Name, ") error")
		m.P("// Get key string for etcd access")
		m.P("GetKeyString(in *", message.Name, ") string")
		m.P("//Object storage IO interface")
		m.P("ObjStore")
		m.P("// Refresh is called after create/update/delete to update in-memory cache")
		m.P("Refresh(in *", message.Name, ", key string) error")
		m.P("}")
		m.P("")

		m.P("func (m *", message.Name, ") Create(cud ", message.Name, "Cud) (*Result, error) {")
		m.P("err := cud.Validate(m)")
		m.P("if err != nil { return nil, err }")
		m.P("key := cud.GetKeyString(m)")
		m.P("val, err := json.Marshal(m)")
		m.P("if err != nil { return nil, err }")
		m.P("err = cud.Create(key, string(val))")
		m.P("if err != nil { return nil, err }")
		m.P("err = cud.Refresh(m, key)")
		m.P("return &Result{}, err")
		m.P("}")
		m.P("")

		m.P("func (m *", message.Name, ") Update(cud ", message.Name, "Cud) (*Result, error) {")
		m.P("err := cud.Validate(m)")
		m.P("if err != nil { return nil, err }")
		m.P("key := cud.GetKeyString(m)")
		m.P("var vers int64 = 0")
		if HasGrpcFields(message) {
			m.P("curBytes, vers, err := cud.Get(key)")
			m.P("if err != nil { return nil, err }")
			m.P("var cur ", message.Name)
			m.P("err = json.Unmarshal(curBytes, &cur)")
			m.P("if err != nil { return nil, err }")
			m.P("cur.CopyInFields(m)")
			m.P("// never save fields")
			m.P("cur.Fields = nil")
			m.P("val, err := json.Marshal(cur)")
		} else {
			m.P("val, err := json.Marshal(m)")
		}
		m.P("if err != nil { return nil, err }")
		m.P("err = cud.Update(key, string(val), vers)")
		m.P("if err != nil { return nil, err }")
		m.P("err = cud.Refresh(m, key)")
		m.P("return &Result{}, err")
		m.P("}")
		m.P("")

		m.P("func (m *", message.Name, ") Delete(cud ", message.Name, "Cud) (*Result, error) {")
		m.P("err := cud.ValidateKey(m)")
		m.P("if err != nil { return nil, err }")
		m.P("key := cud.GetKeyString(m)")
		m.P("err = cud.Delete(key)")
		m.P("if err != nil { return nil, err }")
		m.P("err = cud.Refresh(m, key)")
		m.P("return &Result{}, err")
		m.P("}")
		m.P("")
	}
}

func (m *mex) generateService(file *generator.FileDescriptor, service *descriptor.ServiceDescriptorProto) {
	if len(service.Method) != 0 {
		for _, method := range service.Method {
			m.generateMethod(file, method)
		}
	}
}

func (m *mex) generateMethod(file *generator.FileDescriptor, method *descriptor.MethodDescriptorProto) {

}

func HasGrpcFields(message *descriptor.DescriptorProto) bool {
	if message.Field != nil && *message.Field[0].Name == "fields" && *message.Field[0].Type == descriptor.FieldDescriptorProto_TYPE_BYTES {
		return true
	}
	return false
}

func GetGenerateMatches(message *descriptor.DescriptorProto) bool {
	return getMessageOptionBool(message, protogen.E_GenerateMatches, false)
}

func GetGenerateCud(message *descriptor.DescriptorProto) bool {
	return getMessageOptionBool(message, protogen.E_GenerateCud, false)
}

func getMethodOptionBool(method *descriptor.MethodDescriptorProto, extension *proto.ExtensionDesc, defVal bool) bool {
	if method.Options == nil {
		return defVal
	}
	ex, err := proto.GetExtension(method.Options, extension)
	if err != nil || ex == nil {
		return defVal
	}
	return *(ex.(*bool))

}

func getMessageOptionBool(message *descriptor.DescriptorProto, extension *proto.ExtensionDesc, defVal bool) bool {
	if message.Options == nil {
		return defVal
	}
	ex, err := proto.GetExtension(message.Options, extension)
	if err != nil || ex == nil {
		return defVal
	}
	return *(ex.(*bool))
}
