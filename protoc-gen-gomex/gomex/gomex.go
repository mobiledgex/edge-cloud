package gomex

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/mobiledgex/edge-cloud/protogen"
)

func RegisterGomex() {
	generator.RegisterPlugin(new(gomex))
}

func init() {
	generator.RegisterPlugin(new(gomex))
}

type gomex struct {
	gen  *generator.Generator
	msgs map[string]*descriptor.DescriptorProto
}

func (g *gomex) Name() string {
	return "gomex"
}

func (g *gomex) Init(gen *generator.Generator) {
	g.gen = gen
	g.msgs = make(map[string]*descriptor.DescriptorProto)
}

// P forwards to g.gen.P
func (g *gomex) P(args ...interface{}) {
	g.gen.P(args...)
}

func (g *gomex) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.MessageType) != 0 {
		for _, msg := range file.FileDescriptorProto.MessageType {
			g.generateMessage(file, msg)
		}
	}
	if len(file.FileDescriptorProto.Service) != 0 {
		for _, service := range file.FileDescriptorProto.Service {
			g.generateService(file, service)
		}
	}
}

func (g *gomex) GenerateImports(file *generator.FileDescriptor) {
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
			g.msgs[*msg.Name] = msg
		}
	}
	if hasGenerateCud {
		g.P("import \"encoding/json\"")
	}
	if hasGrpcFields {
		g.P("import \"github.com/mobiledgex/edge-cloud/util\"")
	}
}

func (g *gomex) generateMessage(file *generator.FileDescriptor, message *descriptor.DescriptorProto) {
	if GetGenerateMatches(message) && message.Field != nil {
		g.P("func (m *", message.Name, ") Matches(filter *", message.Name, ") bool {")
		g.P("if filter == nil { return true }")
		for _, field := range message.Field {
			if field.Type == nil {
				continue
			}
			name := generator.CamelCase(*field.Name)
			switch *field.Type {
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				g.P("if filter.", name, " != nil && !filter.", name, ".Matches(m.", name, ") {")
				g.P("return false")
				g.P("}")
			case descriptor.FieldDescriptorProto_TYPE_GROUP:
				// deprecated in proto3
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				// TODO
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				g.P("if filter.", name, " != \"\" && filter.", name, " != m.", name, "{")
				g.P("return false")
				g.P("}")
			default:
				g.P("if filter.", name, " != 0 && filter.", name, " != m.", name, "{")
				g.P("return false")
				g.P("}")
			}
		}
		g.P("return true")
		g.P("}")
		g.P("")
	}
	if HasGrpcFields(message) {
		for _, field := range message.Field {
			if field.Name == nil || field.Number == nil {
				continue
			}
			name := generator.CamelCase(*field.Name)
			g.P("const ", message.Name, "Field", name, " uint = ", field.Number)
		}
		g.P("")
	}
	g.P("func (m *", message.Name, ") CopyInFields(src *", message.Name, ") {")
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
			g.P("if set, _ := util.GrpcFieldsGet(src.Fields, ", field.Number, "); set == true {")
		}
		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			msg, ok := g.msgs[*field.TypeName]
			if !ok || !HasGrpcFields(msg) {
				g.P("if m.", name, " != nil && src.", name, " != nil {")
				g.P("*m.", name, " = *src.", name)
				g.P("}")
			} else {
				g.P("m.", name, ".CopyInFields(src.", name, ")")
			}
		case descriptor.FieldDescriptorProto_TYPE_GROUP:
			// deprecated in proto3
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			g.P("copy(m.", name, ", src.", name, ")")
		default:
			g.P("m.", name, " = src.", name)
		}
		if HasGrpcFields(message) {
			g.P("}")
		}
	}
	g.P("}")

	if GetGenerateCud(message) {
		g.P("type ", message.Name, "Cud interface {")
		g.P("// Validate all fields for create/update")
		g.P("Validate(in *", message.Name, ") error")
		g.P("// Validate only Key fields for delete")
		g.P("ValidateKey(in *", message.Name, ") error")
		g.P("// Get key string for etcd access")
		g.P("GetKeyString(in *", message.Name, ") string")
		g.P("//Etcd IO interface")
		g.P("EtcdIO")
		g.P("// Refresh is called after create/update/delete to update in-memory cache")
		g.P("Refresh(in *", message.Name, ", key string) error")
		g.P("}")
		g.P("")

		g.P("func (m *", message.Name, ") Create(cud ", message.Name, "Cud) (*Result, error) {")
		g.P("err := cud.Validate(m)")
		g.P("if err != nil { return nil, err }")
		g.P("key := cud.GetKeyString(m)")
		g.P("val, err := json.Marshal(m)")
		g.P("if err != nil { return nil, err }")
		g.P("err = cud.Create(key, string(val))")
		g.P("if err != nil { return nil, err }")
		g.P("err = cud.Refresh(m, key)")
		g.P("return &Result{}, err")
		g.P("}")
		g.P("")

		g.P("func (m *", message.Name, ") Update(cud ", message.Name, "Cud) (*Result, error) {")
		g.P("err := cud.Validate(m)")
		g.P("if err != nil { return nil, err }")
		g.P("key := cud.GetKeyString(m)")
		g.P("var vers int64 = 0")
		if HasGrpcFields(message) {
			g.P("curBytes, vers, err := cud.Get(key)")
			g.P("if err != nil { return nil, err }")
			g.P("var cur ", message.Name)
			g.P("err = json.Unmarshal(curBytes, &cur)")
			g.P("if err != nil { return nil, err }")
			g.P("cur.CopyInFields(m)")
			g.P("// never save fields")
			g.P("cur.Fields = nil")
			g.P("val, err := json.Marshal(cur)")
		} else {
			g.P("val, err := json.Marshal(m)")
		}
		g.P("if err != nil { return nil, err }")
		g.P("err = cud.Update(key, string(val), vers)")
		g.P("if err != nil { return nil, err }")
		g.P("err = cud.Refresh(m, key)")
		g.P("return &Result{}, err")
		g.P("}")
		g.P("")

		g.P("func (m *", message.Name, ") Delete(cud ", message.Name, "Cud) (*Result, error) {")
		g.P("err := cud.ValidateKey(m)")
		g.P("if err != nil { return nil, err }")
		g.P("key := cud.GetKeyString(m)")
		g.P("err = cud.Delete(key)")
		g.P("if err != nil { return nil, err }")
		g.P("err = cud.Refresh(m, key)")
		g.P("return &Result{}, err")
		g.P("}")
		g.P("")
	}
}

func (g *gomex) generateService(file *generator.FileDescriptor, service *descriptor.ServiceDescriptorProto) {
	if len(service.Method) != 0 {
		for _, method := range service.Method {
			g.generateMethod(file, method)
		}
	}
}

func (g *gomex) generateMethod(file *generator.FileDescriptor, method *descriptor.MethodDescriptorProto) {

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
