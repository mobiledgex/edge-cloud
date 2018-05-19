// copyright

package gensupport

import (
	"fmt"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

// DumpProtoFiles will dump as comments the proto data passed in
// by protoc. Note this does not dump wrapped data that the generator
// builds on top of the proto data.
//
// This is useful to help understand what exactly is the data that
// is passed in and available to the generator. This function skips the
// descriptor.proto file because the format definitions are not
// useful and quite lengthly.
func (s *PluginSupport) DumpProtoFiles(g *generator.Generator) {
	for _, file := range s.ProtoFiles {
		if *file.Name == "google/protobuf/descriptor.proto" {
			continue
		}
		if *file.Name == "github.com/gogo/protobuf/gogoproto/gogo.proto" {
			// this file is just added options from gogo
			continue
		}
		dumpFile(g, file)
	}
}

func dumpFile(g *generator.Generator, file *descriptor.FileDescriptorProto) {
	messages := make(map[*descriptor.DescriptorProto]bool)

	Pnil(g, "// file ", file.Name, " package ", file.Package)
	if len(file.Dependency) > 0 {
		for _, dep := range file.Dependency {
			Pnil(g, "//   dependency ", dep)
		}
	}
	if len(file.PublicDependency) > 0 {
		for _, pub := range file.PublicDependency {
			Pnil(g, "//   public dep ", &pub)
		}
	}
	if len(file.WeakDependency) > 0 {
		for _, weak := range file.WeakDependency {
			Pnil(g, "//   weak dep ", &weak)
		}
	}
	if len(file.MessageType) > 0 {
		for _, msg := range file.MessageType {
			dumpMessage(g, "", msg, messages)
		}
	}
	if len(file.EnumType) > 0 {
		for _, en := range file.EnumType {
			dumpEnum(g, "", en)
		}
	}
	if len(file.Service) > 0 {
		for _, service := range file.Service {
			dumpService(g, "", service)
		}
	}
	if len(file.Extension) > 0 {
		Pnil(g, "// extensions: ")
		for _, ext := range file.Extension {
			dumpField(g, "  ", ext)
		}
	}
	Pnil(g, "// file ", file.Name, " done")
	Pnil(g, "")
}

func dumpMessage(g *generator.Generator, indent string, msg *descriptor.DescriptorProto, messages map[*descriptor.DescriptorProto]bool) {
	if _, found := messages[msg]; found {
		Pnil(g, "//", indent, " recursive definition of ", msg.Name)
		return
	}

	messages[msg] = true
	Pnil(g, "//", indent, " message ", msg.Name)
	if len(msg.Field) > 0 {
		for _, field := range msg.Field {
			dumpField(g, indent+"  ", field)
		}
	}
	if len(msg.NestedType) > 0 {
		for _, nested := range msg.NestedType {
			dumpMessage(g, indent+"  ", nested, messages)
		}
	}
	if len(msg.EnumType) > 0 {
		for _, en := range msg.EnumType {
			dumpEnum(g, indent+"  ", en)
		}
	}
	if len(msg.OneofDecl) > 0 {
		for _, oneof := range msg.OneofDecl {
			dumpOneof(g, indent+"  ", oneof)
		}
	}
}

func dumpField(g *generator.Generator, indent string, field *descriptor.FieldDescriptorProto) {
	ftype := "none"
	if field.Type != nil {
		ftype = field.Type.String()
	}
	Pnil(g, "//", indent, " field ", field.Name, " type ", ftype, " typename ", field.TypeName, " extendee ", field.Extendee, " label ", GetLabelString(field.Label), " oneof-index ", field.OneofIndex)
}

func dumpEnum(g *generator.Generator, indent string, en *descriptor.EnumDescriptorProto) {
	Pnil(g, "//", indent, " enum ", en.Name)
	if len(en.Value) > 0 {
		for _, val := range en.Value {
			dumpEnumValue(g, indent+"  ", val)
		}
	}
}

func dumpOneof(g *generator.Generator, indent string, oneof *descriptor.OneofDescriptorProto) {
	Pnil(g, "//", indent, " oneof ", oneof.Name)
}

func dumpEnumValue(g *generator.Generator, indent string, val *descriptor.EnumValueDescriptorProto) {
	Pnil(g, "//", indent, " enum-value ", val.Name, " number ", val.Number)
}

func dumpService(g *generator.Generator, indent string, service *descriptor.ServiceDescriptorProto) {
	Pnil(g, "//", indent, " service ", service.Name)
	if len(service.Method) > 0 {
		for _, method := range service.Method {
			dumpMethod(g, indent+"  ", method)
		}
	}
}

func dumpMethod(g *generator.Generator, indent string, method *descriptor.MethodDescriptorProto) {
	Pnil(g, "//", indent, " method ", method.Name, " input-type ", method.InputType, " output-type ", method.OutputType, " client-streaming ", method.ClientStreaming, " server-streaming ", method.ServerStreaming)
}

func GetLabelString(label *descriptor.FieldDescriptorProto_Label) string {
	if label == nil {
		return "none"
	}
	switch *label {
	case descriptor.FieldDescriptorProto_LABEL_OPTIONAL:
		return "optional"
	case descriptor.FieldDescriptorProto_LABEL_REQUIRED:
		return "required"
	case descriptor.FieldDescriptorProto_LABEL_REPEATED:
		return "repeated"
	}
	return "unknown"
}

// Pnil is a copy of generator.P(), but allows nil objects to be
// passed in. The normal generator.P() will crash on nil objects.
func Pnil(g *generator.Generator, str ...interface{}) {
	for _, v := range str {
		switch s := v.(type) {
		case string:
			g.WriteString(s)
		case *string:
			if v.(*string) == nil {
				g.WriteString("nil")
				continue
			}
			g.WriteString(*s)
		case bool:
			fmt.Fprintf(g, "%t", s)
		case *bool:
			if v.(*bool) == nil {
				g.WriteString("nil")
				continue
			}
			fmt.Fprintf(g, "%t", *s)
		case int:
			fmt.Fprintf(g, "%d", s)
		case *int32:
			if v.(*int32) == nil {
				g.WriteString("nil")
				continue
			}
			fmt.Fprintf(g, "%d", *s)
		case *int64:
			if v.(*int64) == nil {
				g.WriteString("nil")
				continue
			}
			fmt.Fprintf(g, "%d", *s)
		case float64:
			fmt.Fprintf(g, "%g", s)
		case *float64:
			if v.(*float64) == nil {
				g.WriteString("nil")
				continue
			}
			fmt.Fprintf(g, "%g", *s)
		default:
			g.Fail(fmt.Sprintf("unknown type in printer: %T", v))
		}
	}
	g.WriteByte('\n')
}

// Field type name is only set if type is ENUM, MESSAGE, or GROUP.
// Otherwise it may be an invalid pointer (not even nil)
func FieldTypeNameValid(field *descriptor.FieldDescriptorProto) bool {
	if field.Type == nil {
		return false
	}
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		fallthrough
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		return true
	}
	return false
}
