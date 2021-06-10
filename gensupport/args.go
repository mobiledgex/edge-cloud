package gensupport

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/protogen"
	"github.com/mobiledgex/edge-cloud/util"
)

type Arg struct {
	Name    string
	Comment string
}

// Get all message types that are used as input to a method in any
// of the files.
func GetInputMessages(g *generator.Generator, support *PluginSupport) map[string]*generator.Descriptor {
	allInputDescs := make(map[string]*generator.Descriptor)
	for _, protofile := range support.ProtoFiles {
		if !support.GenFile(protofile.GetName()) {
			continue
		}
		for _, svc := range protofile.GetService() {
			if len(svc.Method) == 0 {
				continue
			}
			for _, method := range svc.Method {
				desc := GetDesc(g, method.GetInputType())
				allInputDescs[*desc.DescriptorProto.Name] = desc
			}
		}
	}
	return allInputDescs
}

func GenerateMessageArgs(g *generator.Generator, support *PluginSupport, desc *generator.Descriptor, prefixMessageToAlias bool, count int) {
	generateArgs(g, support, desc, nil, prefixMessageToAlias, count)
}

func GenerateMethodArgs(g *generator.Generator, support *PluginSupport, method *descriptor.MethodDescriptorProto, prefixMessageToAlias bool, count int) {
	if !HasMethodArgs(method) {
		return
	}
	desc := GetDesc(g, method.GetInputType())
	generateArgs(g, support, desc, method, prefixMessageToAlias, count)
}

func generateArgs(g *generator.Generator, support *PluginSupport, desc *generator.Descriptor, method *descriptor.MethodDescriptorProto, prefixMessageToAlias bool, count int) {
	if desc.Options != nil && desc.Options.MapEntry != nil && *desc.Options.MapEntry == true {
		// descriptor for map entry, skip
		return
	}

	message := desc.DescriptorProto
	msgname := *message.Name
	if method != nil {
		msgname = *method.Name
	}

	aliasSpec := GetAlias(message)
	aliasMap := make(map[string]string)
	for _, a := range strings.Split(aliasSpec, ",") {
		// format is alias=real
		kv := strings.SplitN(strings.TrimSpace(a), "=", 2)
		if len(kv) != 2 {
			continue
		}
		// real -> alias
		aliasMap[kv[1]] = kv[0]
	}
	noconfig := GetNoConfig(message, method)
	noconfigMap := make(map[string]struct{})
	for _, nc := range strings.Split(noconfig, ",") {
		noconfigMap[nc] = struct{}{}
	}
	notreq := GetNotreq(message, method)
	notreqMap := make(map[string]struct{})
	for _, nr := range strings.Split(notreq, ",") {
		notreqMap[nr] = struct{}{}
	}
	alsoreq := GetAlsoRequired(message, method)
	alsoreqMap := make(map[string]struct{})
	for _, ar := range strings.Split(alsoreq, ",") {
		alsoreqMap[ar] = struct{}{}
	}

	// find all possible args
	allargs, specialArgs := GetArgs(g, support, []string{}, desc)

	// generate required args (set by Key)
	requiredMap := make(map[string]struct{})
	g.P("var ", msgname, "RequiredArgs = []string{")
	for _, arg := range allargs {
		if argSpecified(arg.Name, notreqMap) {
			// explicity not required
			continue
		} else if _, found := noconfigMap[arg.Name]; found {
			// part of no config
			continue
		} else if argSpecified(arg.Name, alsoreqMap) {
			// explicity required
		} else if strings.HasPrefix(arg.Name, "Key.") || GetObjAndKey(message) {
			// key field, or entire struct is key, so all fields
			// are implicitly required
		} else {
			// default: implicitly not required
			continue
		}

		requiredMap[arg.Name] = struct{}{}
		// use alias if exists
		str, ok := aliasMap[arg.Name]
		if !ok {
			str = arg.Name
		}
		g.P("\"", strings.ToLower(str), "\",")
	}
	g.P("}")

	// generate optional args
	g.P("var ", msgname, "OptionalArgs = []string{")
	for _, arg := range allargs {
		if arg.Name == "Fields" {
			continue
		}
		if _, found := requiredMap[arg.Name]; found {
			continue
		}
		parts := strings.Split(arg.Name, ".")
		checkStr := ""
		noconfigFound := false
		for _, part := range parts {
			if checkStr == "" {
				checkStr = part
			} else {
				checkStr = checkStr + "." + part
			}
			if _, found := noconfigMap[checkStr]; found {
				noconfigFound = true
				break
			}
		}
		if noconfigFound {
			continue
		}
		parts = strings.Split(arg.Name, ":")
		if _, found := noconfigMap[parts[0]]; found {
			continue
		}
		str, ok := aliasMap[arg.Name]
		if !ok {
			str = arg.Name
		}
		g.P("\"", strings.ToLower(str), "\",")
	}
	g.P("}")

	if method != nil {
		// aliases, comments, etc should be same for methods
		return
	}

	// generate aliases
	g.P("var ", msgname, "AliasArgs = []string{")
	for _, arg := range allargs {
		// keep noconfig ones here because aliases
		// may be used for tabular output later.

		alias, ok := aliasMap[arg.Name]
		if !ok {
			if !prefixMessageToAlias {
				continue
			}
			alias = arg.Name
		}
		name := strings.ToLower(arg.Name)
		alias = strings.ToLower(alias)

		prefix := ""
		if prefixMessageToAlias {
			prefix = strings.ToLower(*message.Name) + "."
		}

		g.P("\"", alias, "=", prefix, name, "\",")
	}
	g.P("}")

	// generate comments
	g.P("var ", msgname, "Comments = map[string]string{")
	for _, arg := range allargs {
		if arg.Comment == "" {
			continue
		}
		alias, ok := aliasMap[arg.Name]
		if !ok {
			alias = arg.Name
		}
		alias = strings.ToLower(alias)
		g.P("\"", alias, "\": \"", arg.Comment, "\",")
	}
	g.P("}")

	// generate special args
	g.P("var ", msgname, "SpecialArgs = map[string]string{")
	keys := make([]string, 0, len(specialArgs))
	for arg, _ := range specialArgs {
		keys = append(keys, arg)
	}
	sort.Strings(keys)
	for _, arg := range keys {
		argType := specialArgs[arg]
		if prefixMessageToAlias {
			arg = *message.Name + "." + arg
		}
		g.P("\"", strings.ToLower(arg), "\": \"", argType, "\",")
	}
	g.P("}")
}

func GetArgs(g *generator.Generator, support *PluginSupport, parents []string, desc *generator.Descriptor) ([]Arg, map[string]string) {
	allargs := []Arg{}
	specialArgs := make(map[string]string)
	msg := desc.DescriptorProto
	for i, field := range msg.Field {
		if field.Type == nil || field.OneofIndex != nil {
			continue
		}
		name := generator.CamelCase(*field.Name)
		hierName := strings.Join(append(parents, name), ".")
		comment := support.GetComments(desc.File().GetName(), fmt.Sprintf("%s,2,%d", desc.Path(), i))
		comment = strings.TrimSpace(strings.Map(RemoveNewLines, comment))
		arg := Arg{
			Name:    hierName,
			Comment: comment,
		}
		mapType := support.GetMapType(g, field, WithNoImport())
		if mapType != nil && mapType.FlagType != "" {
			specialArgs[hierName] = mapType.FlagType
			allargs = append(allargs, arg)
		} else if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			subDesc := GetDesc(g, field.GetTypeName())
			if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				name = name + ":#"
			}
			subArgs, subSpecialArgs := GetArgs(g, support, append(parents, name), subDesc)
			allargs = append(allargs, subArgs...)
			for k, v := range subSpecialArgs {
				specialArgs[k] = v
			}
		} else if *field.Type == descriptor.FieldDescriptorProto_TYPE_ENUM {
			enumDesc := GetEnumDesc(g, field.GetTypeName())
			en := enumDesc.EnumDescriptorProto
			strs := make([]string, 0, len(en.Value))
			for _, val := range en.Value {
				if GetEnumBackend(val) {
					continue
				}
				strs = append(strs, util.CamelCase(*val.Name))
			}
			text := "one of"
			if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				text = "comma-separated list of"
			}
			arg.Comment = arg.Comment + ", " + text + " " + strings.Join(strs, ", ")
			allargs = append(allargs, arg)
		} else {
			if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED && *field.Type == descriptor.FieldDescriptorProto_TYPE_STRING {
				specialArgs[hierName] = "StringArray"
			}
			allargs = append(allargs, arg)
		}
	}
	return allargs, specialArgs
}

func RemoveNewLines(r rune) rune {
	if r == '\n' || r == '\r' || r == '\t' || r == '"' || r == '\'' {
		return -1
	}
	return r
}

// arg is hierarchical name, A.B.C. The map may contain the full
// path or a parent, so A.B.C or A.B or A, all of which specify
// A.B.C.
func argSpecified(arg string, entries map[string]struct{}) bool {
	parts := strings.Split(arg, ".")
	for ii := 1; ii <= len(parts); ii++ {
		sub := strings.Join(parts[:ii], ".")
		if _, found := entries[sub]; found {
			return true
		}
	}
	return false
}

func HasMethodArgs(method *descriptor.MethodDescriptorProto) bool {
	if HasExtension(method.Options, protogen.E_MethodNoconfig) ||
		HasExtension(method.Options, protogen.E_MethodNotRequired) ||
		HasExtension(method.Options, protogen.E_MethodAlsoRequired) {
		return true
	}
	return false
}

func GetStreamOutIncremental(method *descriptor.MethodDescriptorProto) bool {
	return proto.GetBoolExtension(method.Options, protogen.E_StreamOutIncremental, false)
}

func GetNoConfig(message *descriptor.DescriptorProto, method *descriptor.MethodDescriptorProto) string {
	noConfigStr := GetStringExtension(message.Options, protogen.E_Noconfig, "")
	if method != nil {
		str, found := FindStringExtension(method.Options, protogen.E_MethodNoconfig)
		if found {
			if noConfigStr != "" && str != "" {
				noConfigStr += ","
			}
			noConfigStr += str
		}
	}
	return noConfigStr
}

func GetNotreq(message *descriptor.DescriptorProto, method *descriptor.MethodDescriptorProto) string {
	if method != nil {
		str, found := FindStringExtension(method.Options, protogen.E_MethodNotRequired)
		if found {
			return str
		}
	}
	return GetStringExtension(message.Options, protogen.E_NotRequired, "")
}

func GetAlsoRequired(message *descriptor.DescriptorProto, method *descriptor.MethodDescriptorProto) string {
	if method != nil {
		str, found := FindStringExtension(method.Options, protogen.E_MethodAlsoRequired)
		if found {
			return str
		}
	}
	return GetStringExtension(message.Options, protogen.E_AlsoRequired, "")
}

func GetInputRequired(method *descriptor.MethodDescriptorProto) bool {
	return proto.GetBoolExtension(method.Options, protogen.E_InputRequired, false)
}

func GetAlias(message *descriptor.DescriptorProto) string {
	return GetStringExtension(message.Options, protogen.E_Alias, "")
}
