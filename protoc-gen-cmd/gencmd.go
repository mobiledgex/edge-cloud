package main

import (
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/gensupport"
	"github.com/mobiledgex/edge-cloud/protoc-gen-cmd/protocmd"
	"github.com/spf13/cobra"
)

var _ cobra.Command

type EnumArg struct {
	field     *descriptor.FieldDescriptorProto
	inVar     string
	msgName   string
	hierField string
	repeated  bool
}

type GenCmd struct {
	*generator.Generator
	support            gensupport.PluginSupport
	packageName        string
	tmpl               *template.Template
	fieldTmpl          *template.Template
	inMessages         map[string]*generator.Descriptor
	enumArgs           map[string][]*EnumArg
	hideTags           map[string]struct{}
	importTime         bool
	importStrconv      bool
	importStrings      bool
	importCobra        bool
	importContext      bool
	importOS           bool
	importIO           bool
	importTabwriter    bool
	importBuiltinTypes bool
	importPflag        bool
	importErrors       bool
	importOutputGen    bool
	importCmdSup       bool
}

func (g *GenCmd) Name() string {
	return "GenCmd"
}

func (g *GenCmd) Init(gen *generator.Generator) {
	g.Generator = gen
	g.tmpl = template.Must(template.New("cmd").Parse(tmpl))
	g.fieldTmpl = template.Must(template.New("field").Parse(fieldTmpl))
}

func (g *GenCmd) GenerateImports(file *generator.FileDescriptor) {
	g.support.PrintUsedImports(g.Generator)
	if g.importStrings {
		g.PrintImport("", "strings")
	}
	if g.importTime {
		g.PrintImport("", "time")
	}
	if g.importStrconv {
		g.PrintImport("", "strconv")
	}
	if g.importCobra {
		g.PrintImport("", "github.com/spf13/cobra")
	}
	if g.importContext {
		g.PrintImport("", "context")
	}
	if g.importOS {
		g.PrintImport("", "os")
	}
	if g.importIO {
		g.PrintImport("", "io")
	}
	if g.importTabwriter {
		g.PrintImport("", "text/tabwriter")
	}
	if g.importBuiltinTypes {
		g.PrintImport("google_protobuf1", "github.com/gogo/protobuf/types")
	}
	if g.importPflag {
		g.PrintImport("", "github.com/spf13/pflag")
	}
	if g.importErrors {
		g.PrintImport("", "errors")
	}
	if g.importOutputGen {
		g.importCmdSup = true
		g.PrintImport("", "encoding/json")
		g.PrintImport("", "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/yaml")
	}
	if g.importCmdSup {
		g.PrintImport("", "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/cmdsup")
	}
}

func (g *GenCmd) Generate(file *generator.FileDescriptor) {
	// generate files with messages, because other files may rely
	// on TSV function.
	g.importStrings = false
	g.importTime = false
	g.importStrconv = false
	g.importCobra = false
	g.importContext = false
	g.importOS = false
	g.importIO = false
	g.importTabwriter = false
	g.importBuiltinTypes = false
	g.importPflag = false
	g.importErrors = false
	g.importOutputGen = false
	g.importCmdSup = false
	g.inMessages = make(map[string]*generator.Descriptor)
	g.enumArgs = make(map[string][]*EnumArg)
	g.hideTags = make(map[string]struct{})
	g.packageName = *file.FileDescriptorProto.Package
	g.support.InitFile()

	g.P(gensupport.AutoGenComment)

	// Generate service vars which must assigned by the main function.
	// Also generate input vars which will be used to capture input args.
	if len(file.FileDescriptorProto.Service) > 0 {
		for _, service := range file.FileDescriptorProto.Service {
			g.generateServiceVars(file.FileDescriptorProto, service)
		}
	}
	g.generateInputVars()

	for _, en := range file.Enums() {
		g.generateEnumValSlice(en)
	}

	// Generate slicer functions for writing output
	for _, desc := range file.Messages() {
		if desc.File() != file.FileDescriptorProto {
			continue
		}
		g.generateSlicer(desc)
	}

	// Generate hidetags functions
	for _, desc := range file.Messages() {
		if desc.File() != file.FileDescriptorProto {
			continue
		}
		visited := make([]*generator.Descriptor, 0)
		if g.hasHideTags(desc, visited) {
			g.hideTags[*desc.DescriptorProto.Name] = struct{}{}
			g.generateHideTags(desc)
		}
	}

	// Generate cobra command definitions
	if len(file.FileDescriptorProto.Service) > 0 {
		for _, service := range file.FileDescriptorProto.Service {
			g.generateServiceCmd(file.FileDescriptorProto, service)
		}
	}

	// Generate the flag sets which capture user input args.
	// These are set up on a per struct basis, and are added to
	// commands later.
	g.P("func init() {")
	for _, desc := range file.Messages() {
		msgName := *desc.DescriptorProto.Name
		if _, found := g.inMessages[msgName]; !found {
			continue
		}
		noconfig := make(map[string]struct{})
		for _, str := range strings.Split(GetNoConfig(desc.DescriptorProto), ",") {
			noconfig[str] = struct{}{}
		}
		visited := make([]*generator.Descriptor, 0)
		parentNoConfig := false
		g.generateVarFlags(msgName, make([]string, 0), desc, noconfig, visited, parentNoConfig)
	}
	// Add per input struct flag sets to the commands.
	initNoConfig := false
	if len(file.FileDescriptorProto.Service) > 0 {
		for _, service := range file.FileDescriptorProto.Service {
			g.addCmdFlags(service, initNoConfig)
		}
	}
	g.P("}")
	g.P()

	initNoConfig = true
	if len(file.FileDescriptorProto.Service) > 0 {
		for _, service := range file.FileDescriptorProto.Service {
			g.P("func ", service.Name, "AllowNoConfig() {")
			g.addCmdFlags(service, initNoConfig)
			g.P("}")
			g.P()
		}
	}

	for _, desc := range file.Messages() {
		msgName := *desc.DescriptorProto.Name
		if _, found := g.inMessages[msgName]; !found {
			continue
		}
		if !HasGrpcFields(desc.DescriptorProto) {
			continue
		}
		g.P("func ", msgName, "SetFields() {")
		g.P(msgName, "In.Fields = make([]string, 0)")
		parents := make([]string, 0)
		nums := make([]string, 0)
		visited := make([]*generator.Descriptor, 0)
		g.generateSetFields(msgName, parents, nums, desc, visited)
		g.P("}")
		g.P()
	}

	// Because we cannot define a enum arg flag for enums,
	// we take a string argument, then must assign the correct
	// enum value for the string.
	// We have to be careful about embedded enum values that are
	// defined within a message. They end up having a type name of
	// .proto.AppInst.Liveness, for example. They also end up having
	// a generated type of AppInst_Liveness.
	// sort array to prevent generated file from changing all the time
	strs := make([]string, 0, len(g.enumArgs))
	for msgName, _ := range g.enumArgs {
		strs = append(strs, msgName)
	}
	sort.Strings(strs)
	for _, msgName := range strs {
		enumList := g.enumArgs[msgName]
		g.generateParseEnums(msgName, enumList)
	}
	gensupport.RunParseCheck(g.Generator, file)
}

func (g *GenCmd) hasHideTags(desc *generator.Descriptor, visited []*generator.Descriptor) bool {
	if gensupport.WasVisited(desc, visited) {
		return false
	}
	msg := desc.DescriptorProto
	for _, field := range msg.Field {
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			subDesc := g.GetDesc(field.GetTypeName())
			if g.hasHideTags(subDesc, append(visited, desc)) {
				return true
			}
		}
		if GetHideTag(field) != "" {
			return true
		}
	}
	return false
}

func (g *GenCmd) generateHideTags(desc *generator.Descriptor) {
	msgName := desc.DescriptorProto.Name
	g.P("func ", msgName, "HideTags(in *", g.FQTypeName(desc), ") {")
	g.P("if cmdsup.HideTags == \"\" { return }")
	g.P("tags := make(map[string]struct{})")
	g.P("for _, tag := range strings.Split(cmdsup.HideTags, \",\") {")
	g.P("tags[tag] = struct{}{}")
	g.P("}")
	visited := make([]*generator.Descriptor, 0)
	g.generateHideTagFields(make([]string, 0), desc, visited)
	g.P("}")
	g.P()
	g.importStrings = true
	g.importCmdSup = true
}

func (g *GenCmd) generateHideTagFields(parents []string, desc *generator.Descriptor, visited []*generator.Descriptor) {
	if gensupport.WasVisited(desc, visited) {
		return
	}
	msg := desc.DescriptorProto
	for _, field := range msg.Field {
		if !supportedField(field) {
			continue
		}
		tag := GetHideTag(field)
		if tag == "" {
			continue
		}
		name := generator.CamelCase(*field.Name)
		hierField := strings.Join(append(parents, name), ".")

		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			idx := ""
			if gensupport.IsRepeated(field) {
				ii := fmt.Sprintf("i%d", len(parents))
				g.P("for ", ii, " := 0; ", ii, " < len(in.", hierField, "); ", ii, "++ {")
				idx = "[" + ii + "]"
			}
			subDesc := g.GetDesc(field.GetTypeName())
			g.generateHideTagFields(append(parents, name+idx),
				subDesc, append(visited, desc))
			if gensupport.IsRepeated(field) {
				g.P("}")
			}
		} else {
			val := "0"
			if gensupport.IsRepeated(field) {
				val = "nil"
			} else {
				switch *field.Type {
				case descriptor.FieldDescriptorProto_TYPE_BOOL:
					val = "false"
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

func (g *GenCmd) generateServiceVars(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	if len(service.Method) > 0 {
		g.P("var ", service.Name, "Cmd ", file.Package, ".", *service.Name, "Client")
		g.support.RegisterUsedPkg(*file.Package, file)
		for _, method := range service.Method {
			in := g.GetDesc(method.GetInputType())
			if in == nil || clientStreaming(method) || hasOneof(in) {
				continue
			}
			g.inMessages[g.flatTypeName(*method.InputType)] = in
		}
	}
}

func (g *GenCmd) generateInputVars() {
	// sort array to prevent generated file from changing all the time
	strs := make([]string, 0, len(g.inMessages))
	for flatType, _ := range g.inMessages {
		strs = append(strs, flatType)
	}
	sort.Strings(strs)
	for _, flatType := range strs {
		desc := g.inMessages[flatType]
		g.importPflag = true
		g.P("var ", flatType, "In ", g.FQTypeName(desc))
		g.P("var ", flatType, "FlagSet = pflag.NewFlagSet(\"", flatType, "\", pflag.ExitOnError)")
		g.P("var ", flatType, "NoConfigFlagSet = pflag.NewFlagSet(\"", flatType, "NoConfig\", pflag.ExitOnError)")
		g.generateEnumVars(flatType, desc, make([]string, 0), make([]*generator.Descriptor, 0))
	}
}

func (g *GenCmd) generateEnumValSlice(desc *generator.EnumDescriptor) {
	// this code is a duplicate of code in protoc-gen-gomex to
	// remove any dependency of gencmd on the protoc-gen-gomex,
	// in case we want to open source gencmd as its own project.
	en := desc.EnumDescriptorProto
	g.P("var ", en.Name, "Strings = []string{")
	for _, val := range en.Value {
		g.P("\"", val.Name, "\",")
	}
	g.P("}")
	g.P()
}

func (g *GenCmd) generateEnumVars(flatType string, desc *generator.Descriptor, parents []string, visited []*generator.Descriptor) {
	if gensupport.WasVisited(desc, visited) {
		return
	}
	for _, field := range desc.DescriptorProto.Field {
		if field.Type == nil {
			continue
		}
		name := generator.CamelCase(*field.Name)
		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			subDesc := g.GetDesc(field.GetTypeName())
			g.generateEnumVars(flatType, subDesc, append(parents, name), append(visited, desc))
		case descriptor.FieldDescriptorProto_TYPE_ENUM:
			inVar := flatType + "In" + strings.Join(append(parents, name), "")
			repeated := *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED
			g.P("var ", inVar, " string")
			enumArg := EnumArg{
				field:     field,
				inVar:     inVar,
				msgName:   flatType,
				hierField: strings.Join(append(parents, name), "."),
				repeated:  repeated,
			}
			enumList, found := g.enumArgs[flatType]
			if !found {
				enumList = make([]*EnumArg, 0)
			}
			g.enumArgs[flatType] = append(enumList, &enumArg)
		}
	}
}

type fieldArgs struct {
	MsgName  string
	Ref      string
	Field    string
	Type     string
	DefValue string
	Arg      string
	Usage    string
}

var fieldTmpl = `{{.MsgName}}FlagSet.{{.Type}}Var({{.Ref}}, "{{.Arg}}", {{.DefValue}}, {{.Usage}})
`

func (g *GenCmd) generateVarFlags(msgName string, parents []string, desc *generator.Descriptor, noconfig map[string]struct{}, visited []*generator.Descriptor, parentNoConfig bool) {
	if gensupport.WasVisited(desc, visited) {
		// Break recursion. Googleapis HttpRule
		// includes itself, so is a recursive
		// definition.
		return
	}
	msg := desc.DescriptorProto
	for _, field := range msg.Field {
		if !supportedField(field) {
			continue
		}
		idx := ""
		if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			idx = "[0]"
		}

		name := generator.CamelCase(*field.Name)
		if name == "Fields" {
			continue
		}
		hierField := strings.Join(append(parents, name), ".")
		fargs := &fieldArgs{
			MsgName:  msgName,
			Ref:      "&" + msgName + "In" + "." + hierField + idx,
			Field:    hierField,
			Arg:      argName(append(parents, name)),
			Type:     "String",
			DefValue: "\"\"",
			Usage:    "\"" + hierField + "\"",
		}
		if _, found := noconfig[hierField]; found || parentNoConfig {
			fargs.MsgName = msgName + "NoConfig"
		}
		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			subDesc := g.GetDesc(field.GetTypeName())
			if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				ref := ""
				if gogoproto.IsNullable(field) {
					ref = "*"
				}
				g.P(msgName, "In.", hierField, " = make([]", ref, g.FQTypeName(subDesc), ", 1)")
			}
			if gogoproto.IsNullable(field) {
				subType := g.FQTypeName(subDesc)
				g.P(msgName, "In.", hierField, idx, " = &", subType, "{}")
			}
			subNoConfig := false
			if _, found := noconfig[hierField]; found || parentNoConfig {
				subNoConfig = true
			}
			g.generateVarFlags(msgName, append(parents, name+idx), subDesc, noconfig, append(visited, desc), subNoConfig)
			continue
		case descriptor.FieldDescriptorProto_TYPE_SINT64:
			fallthrough
		case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
			fallthrough
		case descriptor.FieldDescriptorProto_TYPE_INT64:
			fargs.Type = "Int64"
			fargs.DefValue = "0"
		case descriptor.FieldDescriptorProto_TYPE_FIXED64:
			fallthrough
		case descriptor.FieldDescriptorProto_TYPE_UINT64:
			fargs.Type = "Uint64"
			fargs.DefValue = "0"
		case descriptor.FieldDescriptorProto_TYPE_SINT32:
			fallthrough
		case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
			fallthrough
		case descriptor.FieldDescriptorProto_TYPE_INT32:
			fargs.Type = "Int32"
			fargs.DefValue = "0"
		case descriptor.FieldDescriptorProto_TYPE_FIXED32:
			fallthrough
		case descriptor.FieldDescriptorProto_TYPE_UINT32:
			fargs.Type = "Uint32"
			fargs.DefValue = "0"
		case descriptor.FieldDescriptorProto_TYPE_ENUM:
			fargs.Field = msgName + "In" + strings.Join(append(parents, name), "")
			fargs.Ref = "&" + fargs.Field
			enumDesc := g.GetEnumDesc(field.GetTypeName())
			en := enumDesc.EnumDescriptorProto
			strs := make([]string, 0, len(en.Value))
			for _, val := range en.Value {
				strs = append(strs, *val.Name)
			}
			text := "one of"
			if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				text = "comma-separated list of"
			}
			fargs.Usage = fmt.Sprintf("\"%s %v\"", text, strs)
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			fargs.Type = "Bool"
			fargs.DefValue = "false"
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
			fargs.Type = "Float64"
			fargs.DefValue = "0"
		case descriptor.FieldDescriptorProto_TYPE_FLOAT:
			fargs.Type = "Float32"
			fargs.DefValue = "0"
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			if false && strings.HasSuffix(name, "Ip") {
				fargs.Type = "IP"
			} else {
				fargs.Type = "BytesHex"
			}
			fargs.DefValue = "nil"
		}
		if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED &&
			*field.Type != descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			typ := g.support.GoType(g.Generator, field)
			g.P(msgName, "In.", hierField, " = make([]", typ, ", 1)")
		}
		err := g.fieldTmpl.Execute(g, fargs)
		if err != nil {
			g.Fail("Failed to execute flag template for ", msgName, ", field ", name, ": ", err.Error(), "\n")
			return
		}
	}
}

// this relies on field flags generated by gomex
func (g *GenCmd) generateSetFields(msgName string, parents, nums []string, desc *generator.Descriptor, visited []*generator.Descriptor) {
	if gensupport.WasVisited(desc, visited) {
		return
	}
	msg := desc.DescriptorProto
	for _, field := range msg.Field {
		if !supportedField(field) {
			continue
		}
		name := generator.CamelCase(*field.Name)
		if name == "Fields" {
			continue
		}
		num := fmt.Sprintf("%d", *field.Number)

		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			subDesc := g.GetDesc(field.GetTypeName())
			g.generateSetFields(msgName, append(parents, name), append(nums, num), subDesc, append(visited, desc))
		default:
			g.P("if ", msgName, "FlagSet.Lookup(\"", argName(append(parents, name)), "\").Changed {")
			g.P(msgName, "In.Fields = append(", msgName, "In.Fields, \"", strings.Join(append(nums, num), "."), "\")")
			g.P("}")
		}
	}
}

func (g *GenCmd) generateParseEnums(msgName string, enumList []*EnumArg) {
	g.P("func parse", msgName, "Enums() error {")
	for _, enumArg := range enumList {
		en := g.GetEnumDesc(enumArg.field.GetTypeName())
		if en == nil {
			g.Fail("Enum for ", enumArg.inVar, " not found")
		}
		typeName := g.FQTypeName(en)
		g.P("if ", enumArg.inVar, " != \"\" {")
		enumVar := enumArg.msgName + "In." + enumArg.hierField
		if enumArg.repeated {
			g.P("for _, str := range strings.Split(", enumArg.inVar, ", \",\") {")
			g.P("switch str {")
			for _, val := range en.Value {
				g.P("case \"", *val.Name, "\":")

				g.P(enumVar, " = append(", enumVar, ", ", typeName, "(", val.Number, "))")
			}
			g.P("default:")
			g.P("return errors.New(\"Invalid value for ", enumArg.inVar, "\")")
			g.P("}")
			g.P("}")
		} else {
			g.P("switch ", enumArg.inVar, " {")
			for _, val := range en.Value {
				g.P("case \"", *val.Name, "\":")
				g.P(enumVar, " = ", typeName, "(", val.Number, ")")
			}
			g.P("default:")
			g.P("return errors.New(\"Invalid value for ", enumArg.inVar, "\")")
			g.P("}")
		}
		g.P("}")
		g.importErrors = true
	}
	g.P("return nil")
	g.P("}")
	g.P()
}

func (g *GenCmd) generateSlicer(desc *generator.Descriptor) {
	message := desc.DescriptorProto
	g.P("func ", gensupport.GetMsgName(desc), "Slicer(in *", g.FQTypeName(desc), ") []string {")
	g.P("s := make([]string, 0, ", len(message.Field), ")")
	g.generateSlicerFields(desc, make([]string, 0), make([]*generator.Descriptor, 0))
	g.P("return s")
	g.P("}")
	g.P()

	g.P("func ", gensupport.GetMsgName(desc), "HeaderSlicer() []string {")
	g.P("s := make([]string, 0, ", len(message.Field), ")")
	g.generateHeaderSlicerFields(desc, make([]string, 0), make([]*generator.Descriptor, 0))
	g.P("return s")
	g.P("}")
	g.P()

}

func (g *GenCmd) generateSlicerFields(desc *generator.Descriptor, parents []string, visited []*generator.Descriptor) {
	if gensupport.WasVisited(desc, visited) {
		return
	}
	for _, field := range desc.DescriptorProto.Field {
		if !supportedField(field) {
			continue
		}
		idx := ""
		name := generator.CamelCase(*field.Name)
		hierName := strings.Join(append(parents, name), ".")

		if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			idx = "[0]"
			g.P("if in.", hierName, " == nil {")
			ref := ""
			if gogoproto.IsNullable(field) && *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
				ref = "*"
			}
			typ := g.support.GoType(g.Generator, field)
			g.P("in.", hierName, " = make([]", ref, typ, ", 1)")
			g.P("}")
		}

		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			subDesc := g.GetDesc(field.GetTypeName())
			if gogoproto.IsNullable(field) {
				g.P("if in.", hierName, idx, " == nil {")
				g.P("in.", hierName, idx, " = &", g.FQTypeName(subDesc), "{}")
				g.P("}")
			}
			if *field.TypeName == ".google.protobuf.Timestamp" {
				g.importTime = true
				tempField := "_" + strings.Map(strFixer, hierName)
				g.P(tempField, "Time := time.Unix(in.", hierName, idx, ".Seconds, int64(in.", hierName, idx, ".Nanos))")
				g.P("s = append(s, ", tempField, "Time.String())")
				break
			}
			g.generateSlicerFields(subDesc, append(parents, name+idx), append(visited, desc))
		case descriptor.FieldDescriptorProto_TYPE_GROUP:
			// deprecated in proto3
		case descriptor.FieldDescriptorProto_TYPE_BYTES:

			g.P("s = append(s, \"\")")
			// append the to the current s index.  If this is an Ip address, separate with dots
			if strings.HasSuffix(hierName, "Ip") {
				g.P("for i,b := range in.", hierName, idx, " {")
				g.P("s[len(s)-1] += fmt.Sprintf(\"%v\", b)")
				g.P("if i < 3 {")
				g.P("s[len(s)-1] += \".\"")
				g.P("}")
				g.P("}")
			} else {
				g.P("for _,b := range in.", hierName, idx, " {")
				g.P("s[len(s)-1] += fmt.Sprintf(\"%v\", b)")
				g.P("}")
			}
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			g.importStrconv = true
			g.P("s = append(s, strconv.FormatBool(in.", hierName, idx, "))")
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
			fallthrough
		case descriptor.FieldDescriptorProto_TYPE_FLOAT:
			g.importStrconv = true
			g.P("s = append(s, strconv.FormatFloat(float64(in.", hierName, idx, "), 'e', -1, 32))")
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			g.P("s = append(s, in.", hierName, idx, ")")
		case descriptor.FieldDescriptorProto_TYPE_ENUM:
			en := g.GetEnumDesc(field.GetTypeName())
			if en == nil {
				g.Fail("Enum for ", *desc.DescriptorProto.Name, " field ", name, " not found")
			}
			g.P("s = append(s, ", g.FQTypeName(en), "_name[int32(in.", hierName, idx, ")])")
		default:
			// all integers
			g.importStrconv = true
			g.P("s = append(s, strconv.FormatUint(uint64(in.", hierName, idx, "),10))")
		}
	}
}

func (g *GenCmd) generateHeaderSlicerFields(desc *generator.Descriptor, parents []string, visited []*generator.Descriptor) {
	if gensupport.WasVisited(desc, visited) {
		return
	}
	for _, field := range desc.DescriptorProto.Field {
		if !supportedField(field) {
			continue
		}

		name := generator.CamelCase(*field.Name)
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_GROUP {
			continue
		}
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE && *field.TypeName != ".google.protobuf.Timestamp" {
			subDesc := g.GetDesc(field.GetTypeName())
			g.generateHeaderSlicerFields(subDesc, append(parents, name), append(visited, desc))
		} else {
			g.P("s = append(s, \"", strings.Join(append(parents, name), "-"), "\")")
		}
	}
}

func (g *GenCmd) generateServiceCmd(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	if len(service.Method) == 0 {
		return
	}
	count := 0
	for _, method := range service.Method {
		if g.generateMethodCmd(file, service, method) {
			count++
		}
	}
	if count > 0 {
		g.P()
		g.P("var ", service.Name, "Cmds = []*cobra.Command{")
		for _, method := range service.Method {
			g.P(method.Name, "Cmd,")
		}
		g.P("}")
		g.P()
	}
}

type tmplArgs struct {
	Service      string
	Method       string
	InType       string
	OutType      string
	FQOutType    string
	ServerStream bool
	HasEnums     bool
	SetFields    bool
	OutHideTags  bool
}

var tmpl = `

var {{.Method}}Cmd = &cobra.Command{
	Use: "{{.Method}}",
	Run: func(cmd *cobra.Command, args []string) {
		if {{.Service}}Cmd == nil {
			fmt.Println("{{.Service}} client not initialized")
			return
		}
		var err error
{{- if .HasEnums}}
		err = parse{{.InType}}Enums()
		if err != nil {
			fmt.Println("{{.Method}}: ", err)
			return
		}
{{- end}}
{{- if .SetFields}}
		{{.InType}}SetFields()
{{- end}}
		ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
{{- if .ServerStream}}
		defer cancel()
		stream, err := {{.Service}}Cmd.{{.Method}}(ctx, &{{.InType}}In)
		if err != nil {
			fmt.Println("{{.Method}} failed: ", err)
			return
		}
		objs := make([]*{{.FQOutType}}, 0)
		for {
			obj, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("{{.Method}} recv failed: ", err)
				break
			}
{{- if .OutHideTags}}
			{{.OutType}}HideTags(obj)
{{- end}}
			objs = append(objs, obj)
		}
		if len(objs) == 0 {
			return
		}
{{- else}}
		objs, err := {{.Service}}Cmd.{{.Method}}(ctx, &{{.InType}}In)
		cancel()
		if err != nil {
			fmt.Println("{{.Method}} failed: ", err)
			return
		}
{{- if .OutHideTags}}
		{{.OutType}}HideTags(objs)
{{- end}}
{{- end}}
		switch cmdsup.OutputFormat {
		case cmdsup.OutputFormatYaml:
			output, err := yaml.Marshal(objs)
			if err != nil {
				fmt.Printf("Yaml failed to marshal: %s\n", err)
				return
			}
			fmt.Print(string(output))
		case cmdsup.OutputFormatJson:
			output, err := json.MarshalIndent(objs, "", "  ")
			if err != nil {
				fmt.Printf("Json failed to marshal: %s\n", err)
				return
			}
			fmt.Println(string(output))
		case cmdsup.OutputFormatJsonCompact:
			output, err := json.Marshal(objs)
			if err != nil {
				fmt.Printf("Json failed to marshal: %s\n", err)
				return
			}
			fmt.Println(string(output))
		case cmdsup.OutputFormatTable:
			output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			fmt.Fprintln(output, strings.Join({{.OutType}}HeaderSlicer(), "\t"))
{{- if .ServerStream}}
			for _, obj := range objs {
				fmt.Fprintln(output, strings.Join({{.OutType}}Slicer(obj), "\t"))
			}
{{- else}}
			fmt.Fprintln(output, strings.Join({{.OutType}}Slicer(objs), "\t"))
{{- end}}
			output.Flush()
		}
	},
}
`

func (g *GenCmd) generateMethodCmd(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) bool {
	in := g.GetDesc(method.GetInputType())
	out := g.GetDesc(method.GetOutputType())
	if in == nil || clientStreaming(method) || hasOneof(in) {
		// not supported yet
		return false
	}

	g.importCobra = true
	g.importContext = true
	g.importTime = true
	g.importTabwriter = true
	g.importOS = true
	g.importStrings = true
	g.importOutputGen = true
	_, hasEnums := g.enumArgs[*in.DescriptorProto.Name]
	cmd := &tmplArgs{
		Service:      *service.Name,
		Method:       *method.Name,
		InType:       g.flatTypeName(*method.InputType),
		OutType:      g.flatTypeName(*method.OutputType),
		FQOutType:    g.FQTypeName(out),
		ServerStream: serverStreaming(method),
		HasEnums:     hasEnums,
	}
	if strings.HasPrefix(*method.Name, "Update"+cmd.InType) && HasGrpcFields(in.DescriptorProto) {
		cmd.SetFields = true
	}
	if _, found := g.hideTags[*out.DescriptorProto.Name]; found {
		cmd.OutHideTags = true
	}
	if cmd.ServerStream {
		g.importIO = true
	}
	err := g.tmpl.Execute(g, cmd)
	if err != nil {
		g.Fail("Failed to execute cmdTemplate for ", *method.Name, ": ", err.Error(), "\n")
		return false
	}
	return true
}

func (g *GenCmd) addCmdFlags(service *descriptor.ServiceDescriptorProto, initNoConfig bool) {
	if len(service.Method) == 0 {
		return
	}
	for _, method := range service.Method {
		flatType := g.flatTypeName(*method.InputType)
		in := g.inMessages[flatType]
		if in == nil || clientStreaming(method) || hasOneof(in) {
			// not supported yet
			continue
		}
		if initNoConfig {
			g.P(method.Name, "Cmd.Flags().AddFlagSet(", flatType,
				"NoConfigFlagSet)")
		} else {
			g.P(method.Name, "Cmd.Flags().AddFlagSet(", flatType,
				"FlagSet)")
		}
	}
}

// Get the "flat" format of the type name, for embedding in strings
// to form other names. A raw type name may be of the format:
// name
// .package.name
// .package.parent.name (for embedded structs/enums)
// This function removes any package name and converts remaining
// periods to _.
func (g *GenCmd) flatTypeName(name string) string {
	if name[0] != '.' {
		return strings.Replace(name, ".", "_", -1)
	}
	name = name[1:]
	index := strings.Index(name, ".")
	if index == -1 {
		return name
	}
	return strings.Replace(name[index+1:], ".", "_", -1)
}

// Shortcut function
func (g *GenCmd) FQTypeName(obj generator.Object) string {
	return g.support.FQTypeName(g.Generator, obj)
}

// Shortcut function
func (g *GenCmd) GetDesc(typeName string) *generator.Descriptor {
	return gensupport.GetDesc(g.Generator, typeName)
}

// Shortcut function
func (g *GenCmd) GetEnumDesc(typeName string) *generator.EnumDescriptor {
	return gensupport.GetEnumDesc(g.Generator, typeName)
}

func clientStreaming(method *descriptor.MethodDescriptorProto) bool {
	if method.ClientStreaming == nil {
		return false
	}
	return *method.ClientStreaming
}

func serverStreaming(method *descriptor.MethodDescriptorProto) bool {
	if method.ServerStreaming == nil {
		return false
	}
	return *method.ServerStreaming
}

func hasOneof(desc *generator.Descriptor) bool {
	for _, field := range desc.DescriptorProto.Field {
		if field.Type == nil {
			continue
		}
		if field.OneofIndex != nil {
			return true
		}
	}
	return false
}

func supportedField(field *descriptor.FieldDescriptorProto) bool {
	if field.Type == nil {
		return false
	}
	if field.OneofIndex != nil {
		// not supported yet
		return false
	}
	return true
}

func argName(strs []string) string {
	str := strings.Join(strs, "-")
	return strings.ToLower(strings.Replace(str, "[0]", "", -1))
}

func strFixer(r rune) rune {
	if r == '.' || r == '[' || r == ']' {
		return '_'
	}
	return r
}

func HasGrpcFields(message *descriptor.DescriptorProto) bool {
	if message.Field != nil && len(message.Field) > 0 && *message.Field[0].Name == "fields" && *message.Field[0].Type == descriptor.FieldDescriptorProto_TYPE_STRING {
		return true
	}
	return false
}

func GetNoConfig(message *descriptor.DescriptorProto) string {
	return gensupport.GetStringExtension(message.Options, protocmd.E_Noconfig, "")
}

func GetHideTag(field *descriptor.FieldDescriptorProto) string {
	return gensupport.GetStringExtension(field.Options, protocmd.E_Hidetag, "")
}
