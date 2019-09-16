package gensupport

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/protogen"
)

// GenerateHideTags requires "strings" and "cli" packages to be imported.
func GenerateHideTags(g *generator.Generator, support *PluginSupport, desc *generator.Descriptor) {
	msgName := desc.DescriptorProto.Name
	g.P("func ", msgName, "HideTags(in *", support.FQTypeName(g, desc), ") {")
	g.P("if cli.HideTags == \"\" { return }")
	g.P("tags := make(map[string]struct{})")
	g.P("for _, tag := range strings.Split(cli.HideTags, \",\") {")
	g.P("tags[tag] = struct{}{}")
	g.P("}")
	visited := make([]*generator.Descriptor, 0)
	generateHideTagFields(g, support, make([]string, 0), desc, visited)
	g.P("}")
	g.P()
}

func generateHideTagFields(g *generator.Generator, support *PluginSupport, parents []string, desc *generator.Descriptor, visited []*generator.Descriptor) {
	if WasVisited(desc, visited) {
		return
	}
	msg := desc.DescriptorProto
	for _, field := range msg.Field {
		if field.Type == nil {
			continue
		}
		if field.OneofIndex != nil {
			// not supported yet
			continue
		}
		mapType := support.GetMapType(g, field)
		if mapType != nil {
			// not supported
			continue
		}
		tag := GetHideTag(field)
		if tag == "" && *field.Type != descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			continue
		}
		name := generator.CamelCase(*field.Name)
		hierField := strings.Join(append(parents, name), ".")

		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			subDesc := GetDesc(g, field.GetTypeName())
			if tag != "" {
				g.P("if _, found := tags[\"", tag, "\"]; found {")
				if IsRepeated(field) || gogoproto.IsNullable(field) {
					g.P("in.", hierField, " = nil")
				} else {
					subType := support.FQTypeName(g, subDesc)
					g.P("in.", hierField, " = ", subType, "{}")
				}
				g.P("}")
				continue
			}
			idx := ""
			if IsRepeated(field) {
				ii := fmt.Sprintf("i%d", len(parents))
				g.P("for ", ii, " := 0; ", ii, " < len(in.", hierField, "); ", ii, "++ {")
				idx = "[" + ii + "]"
			}
			generateHideTagFields(g, support, append(parents, name+idx),
				subDesc, append(visited, desc))
			if IsRepeated(field) {
				g.P("}")
			}
		} else {
			val := "0"
			if IsRepeated(field) {
				val = "nil"
			} else {
				switch *field.Type {
				case descriptor.FieldDescriptorProto_TYPE_BOOL:
					val = "false"
				case descriptor.FieldDescriptorProto_TYPE_STRING:
					val = "\"\""
				case descriptor.FieldDescriptorProto_TYPE_BYTES:
					val = "nil"
				}
			}
			g.P("if _, found := tags[\"", tag, "\"]; found {")
			g.P("in.", hierField, " = ", val)
			g.P("}")
		}
	}
}

func GetHideTag(field *descriptor.FieldDescriptorProto) string {
	return GetStringExtension(field.Options, protogen.E_Hidetag, "")
}
