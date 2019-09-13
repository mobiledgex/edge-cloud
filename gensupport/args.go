package gensupport

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/protoc-gen-cmd/protocmd"
	"github.com/mobiledgex/edge-cloud/util"
)

type Arg struct {
	Name    string
	Comment string
}

func GenerateMessageArgs(g *generator.Generator, support *PluginSupport, desc *generator.Descriptor, prefixMessageToAlias bool, count int) {
	message := desc.DescriptorProto

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
	noconfig := GetNoConfig(message)
	noconfigMap := make(map[string]struct{})
	for _, nc := range strings.Split(noconfig, ",") {
		noconfigMap[nc] = struct{}{}
	}
	notreq := GetNotreq(message)
	notreqMap := make(map[string]struct{})
	for _, nr := range strings.Split(notreq, ",") {
		notreqMap[nr] = struct{}{}
	}

	// find all possible args
	allargs, specialArgs := GetArgs(g, support, []string{}, desc)

	// generate required args (set by Key)
	requiredMap := make(map[string]struct{})
	g.P("var ", message.Name, "RequiredArgs = []string{")
	for _, arg := range allargs {
		if !strings.HasPrefix(arg.Name, "Key.") {
			continue
		}
		if _, found := notreqMap[arg.Name]; found {
			continue
		}
		parts := strings.Split(arg.Name, ".")
		if _, found := notreqMap[parts[0]]; found {
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
	g.P("var ", message.Name, "OptionalArgs = []string{")
	for _, arg := range allargs {
		if arg.Name == "Fields" {
			continue
		}
		if _, found := requiredMap[arg.Name]; found {
			continue
		}
		if _, found := noconfigMap[arg.Name]; found {
			continue
		}
		parts := strings.Split(arg.Name, ".")
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

	// generate aliases
	g.P("var ", message.Name, "AliasArgs = []string{")
	for _, arg := range allargs {
		if arg.Name == "Fields" {
			continue
		}
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
	g.P("var ", message.Name, "Comments = map[string]string{")
	for _, arg := range allargs {
		if arg.Name == "Fields" || arg.Comment == "" {
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
	g.P("var ", message.Name, "SpecialArgs = map[string]string{")
	for arg, argType := range specialArgs {
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
		comment := support.GetComments(desc.File(), fmt.Sprintf("%s,2,%d", desc.Path(), i))
		comment = strings.TrimSpace(strings.Map(removeNewLines, comment))
		arg := Arg{
			Name:    hierName,
			Comment: comment,
		}
		mapType := support.GetMapType(g, field)
		if mapType != nil && mapType.FlagType != "" {
			specialArgs[hierName] = mapType.FlagType
			allargs = append(allargs, arg)
		} else if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			subDesc := GetDesc(g, field.GetTypeName())
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
				strs = append(strs, util.CamelCase(*val.Name))
			}
			text := "one of"
			if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				text = "comma-separated list of"
			}
			arg.Comment = arg.Comment + ", " + text + " " + strings.Join(strs, ", ")
			allargs = append(allargs, arg)
		} else {
			allargs = append(allargs, arg)
		}
	}
	return allargs, specialArgs
}

func removeNewLines(r rune) rune {
	if r == '\n' || r == '\r' || r == '\t' || r == '"' || r == '\'' {
		return -1
	}
	return r
}

func GetStreamOutIncremental(method *descriptor.MethodDescriptorProto) bool {
	return proto.GetBoolExtension(method.Options, protocmd.E_StreamOutIncremental, false)
}

func GetNoConfig(message *descriptor.DescriptorProto) string {
	return GetStringExtension(message.Options, protocmd.E_Noconfig, "")
}

func GetNotreq(message *descriptor.DescriptorProto) string {
	return GetStringExtension(message.Options, protocmd.E_Notreq, "")
}

func GetAlias(message *descriptor.DescriptorProto) string {
	return GetStringExtension(message.Options, protocmd.E_Alias, "")
}
