package main

import (
	"strings"
	"text/template"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/gensupport"
	"github.com/spf13/cobra"
)

var _ cobra.Command

type EnumArg struct {
	field     *descriptor.FieldDescriptorProto
	inVar     string
	msgName   string
	hierField string
}

type GenCmd struct {
	*generator.Generator
	support            gensupport.PluginSupport
	packageName        string
	tmpl               *template.Template
	fieldTmpl          *template.Template
	inMessages         map[string]*generator.Descriptor
	enumArgs           map[string][]*EnumArg
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
	g.inMessages = make(map[string]*generator.Descriptor)
	g.enumArgs = make(map[string][]*EnumArg)
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

	// Generate slicer functions for writing output
	for _, desc := range file.Messages() {
		if desc.File() != file.FileDescriptorProto {
			continue
		}
		g.generateSlicer(desc)
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
		visited := make(map[*generator.Descriptor]bool)
		g.generateVarFlags(msgName, make([]string, 0), desc, visited)
	}
	// Add per input struct flag sets to the commands.
	if len(file.FileDescriptorProto.Service) > 0 {
		for _, service := range file.FileDescriptorProto.Service {
			g.addCmdFlags(service)
		}
	}
	g.P("}")
	g.P()

	// Because we cannot define a enum arg flag for enums,
	// we take a string argument, then must assign the correct
	// enum value for the string.
	// We have to be careful about embedded enum values that are
	// defined within a message. They end up having a type name of
	// .proto.AppInst.Liveness, for example. They also end up having
	// a generated type of AppInst_Liveness.
	for msgName, enumList := range g.enumArgs {
		g.generateParseEnums(msgName, enumList)
	}
	gensupport.RunParseCheck(g.Generator, file)
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
	for flatType, desc := range g.inMessages {
		g.importPflag = true
		g.P("var ", flatType, "In ", g.FQTypeName(desc))
		g.P("var ", flatType, "FlagSet = pflag.NewFlagSet(\"", flatType, "\", pflag.ExitOnError)")
		g.generateEnumVars(flatType, desc, make([]string, 0), make(map[*generator.Descriptor]bool))
	}
}

func (g *GenCmd) generateEnumVars(flatType string, desc *generator.Descriptor, parents []string, visited map[*generator.Descriptor]bool) {
	visited[desc] = true
	for _, field := range desc.DescriptorProto.Field {
		if field.Type == nil {
			continue
		}
		name := generator.CamelCase(*field.Name)
		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			subDesc := g.GetDesc(field.GetTypeName())
			g.generateEnumVars(flatType, subDesc, append(parents, name), visited)
		case descriptor.FieldDescriptorProto_TYPE_ENUM:
			inVar := flatType + "In" + strings.Join(append(parents, name), "")
			g.P("var ", inVar, " string")
			enumArg := EnumArg{
				field:     field,
				inVar:     inVar,
				msgName:   flatType,
				hierField: strings.Join(append(parents, name), "."),
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
}

var fieldTmpl = `{{.MsgName}}FlagSet.{{.Type}}Var({{.Ref}}, "{{.Arg}}", {{.DefValue}}, "{{.Field}}")
`

func (g *GenCmd) generateVarFlags(msgName string, parents []string, desc *generator.Descriptor, visited map[*generator.Descriptor]bool) {
	visited[desc] = true
	msg := desc.DescriptorProto
	for _, field := range msg.Field {
		if !supportedField(field) {
			continue
		}

		name := generator.CamelCase(*field.Name)
		if name == "Fields" {
			continue
		}
		hierField := strings.Join(append(parents, name), ".")
		fargs := &fieldArgs{
			MsgName:  msgName,
			Ref:      "&" + msgName + "In" + "." + hierField,
			Field:    hierField,
			Arg:      argName(append(parents, name)),
			Type:     "String",
			DefValue: "\"\"",
		}
		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			subDesc := g.GetDesc(field.GetTypeName())
			if _, found := visited[subDesc]; found {
				// Break recursion. Googleapis HttpRule
				// includes itself, so is a recursive
				// definition.
				continue
			}
			subType := g.FQTypeName(subDesc)
			if gogoproto.IsNullable(field) {
				g.P(msgName, "In.", hierField, " = &", subType, "{}")
			}
			g.generateVarFlags(msgName, append(parents, name), subDesc, visited)
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
		err := g.fieldTmpl.Execute(g, fargs)
		if err != nil {
			g.Fail("Failed to execute flag template for ", msgName, ", field ", name, ": ", err.Error(), "\n")
			return
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
		g.P("switch ", enumArg.inVar, " {")
		for _, val := range en.Value {
			g.P("case \"", strings.ToLower(*val.Name), "\":")
			g.P(enumArg.msgName, "In.", enumArg.hierField, " = ", typeName, "(", val.Number, ")")
		}
		g.P("default:")
		g.P("return errors.New(\"Invalid value for ", enumArg.inVar, "\")")
		g.P("}")
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
	g.generateSlicerFields(desc, make([]string, 0), make(map[*generator.Descriptor]bool))
	g.P("return s")
	g.P("}")
	g.P()

	g.P("func ", gensupport.GetMsgName(desc), "HeaderSlicer() []string {")
	g.P("s := make([]string, 0, ", len(message.Field), ")")
	g.generateHeaderSlicerFields(desc, make([]string, 0), make(map[*generator.Descriptor]bool))
	g.P("return s")
	g.P("}")
	g.P()

}

func (g *GenCmd) generateSlicerFields(desc *generator.Descriptor, parents []string, visited map[*generator.Descriptor]bool) {
	visited[desc] = true
	for _, field := range desc.DescriptorProto.Field {
		if !supportedField(field) {
			continue
		}

		name := generator.CamelCase(*field.Name)
		hierName := strings.Join(append(parents, name), ".")
		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			subDesc := g.GetDesc(field.GetTypeName())
			if _, found := visited[subDesc]; found {
				// break recursion
				continue
			}
			if gogoproto.IsNullable(field) {
				g.P("if in.", hierName, " == nil {")
				g.P("in.", hierName, " = &", g.FQTypeName(subDesc), "{}")
				g.P("}")
			}
			if *field.TypeName == ".google.protobuf.Timestamp" {
				g.importTime = true
				g.P(field.Name, "Time := time.Unix(in.", hierName, ".Seconds, int64(in.", hierName, ".Nanos))")
				g.P("s = append(s, ", field.Name, "Time.String())")
				break
			}
			g.generateSlicerFields(subDesc, append(parents, name), visited)
		case descriptor.FieldDescriptorProto_TYPE_GROUP:
			// deprecated in proto3
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			// TODO
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			g.P("s = append(s, strconv.FormatBool(in.", hierName, "))")
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
			fallthrough
		case descriptor.FieldDescriptorProto_TYPE_FLOAT:
			g.importStrconv = true
			g.P("s = append(s, strconv.FormatFloat(float64(in.", hierName, "), 'e', -1, 32))")
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			g.P("s = append(s, in.", hierName, ")")
		default:
			g.P("s = append(s, string(in.", hierName, "))")
		}
	}
}

func (g *GenCmd) generateHeaderSlicerFields(desc *generator.Descriptor, parents []string, visited map[*generator.Descriptor]bool) {
	visited[desc] = true
	for _, field := range desc.DescriptorProto.Field {
		if !supportedField(field) {
			continue
		}

		name := generator.CamelCase(*field.Name)
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_GROUP || *field.Type == descriptor.FieldDescriptorProto_TYPE_BYTES {
			continue
		}
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE && *field.TypeName != ".google.protobuf.Timestamp" {
			subDesc := g.GetDesc(field.GetTypeName())
			if _, found := visited[subDesc]; found {
				// break recursion
				continue
			}
			g.generateHeaderSlicerFields(subDesc, append(parents, name), visited)
		} else {
			g.P("s = append(s, \"", strings.Join(append(parents, name), "-"), "\")")
		}
	}
}

func (g *GenCmd) generateServiceCmd(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	if len(service.Method) > 0 {
		for _, method := range service.Method {
			g.generateMethodCmd(file, service, method)
		}
	}
}

type tmplArgs struct {
	Service      string
	Method       string
	InType       string
	OutType      string
	ServerStream bool
	HasEnums     bool
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
		ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
{{- if .ServerStream}}
		output := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		count := 0
		fmt.Fprintln(output, strings.Join({{.OutType}}HeaderSlicer(), "\t"))
		defer cancel()
		stream, err := {{.Service}}Cmd.{{.Method}}(ctx, &{{.InType}}In)
		if err != nil {
			fmt.Println("{{.Method}} failed: ", err)
			return
		}
		for {
			obj, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("{{.Method}} recv failed: ", err)
				break
			}
			fmt.Fprintln(output, strings.Join({{.OutType}}Slicer(obj), "\t"))
			count++
		}
		if count > 0 {
			output.Flush()
		}
{{- else}}
		out, err := {{.Service}}Cmd.{{.Method}}(ctx, &{{.InType}}In)
		cancel()
		if err != nil {
			fmt.Println("{{.Method}} failed: ", err)
		} else {
			headers := {{.OutType}}HeaderSlicer()
			data := {{.OutType}}Slicer(out)
			for ii := 0; ii < len(headers) && ii < len(data); ii++ {
				fmt.Println(headers[ii] + ": " + data[ii])
			}
		}
{{- end}}
	},
}
`

func (g *GenCmd) generateMethodCmd(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	in := g.GetDesc(method.GetInputType())
	if in == nil || clientStreaming(method) || hasOneof(in) {
		// not supported yet
		return
	}

	g.importCobra = true
	g.importContext = true
	g.importTime = true
	_, hasEnums := g.enumArgs[*in.DescriptorProto.Name]
	cmd := &tmplArgs{
		Service:      *service.Name,
		Method:       *method.Name,
		InType:       g.flatTypeName(*method.InputType),
		OutType:      g.flatTypeName(*method.OutputType),
		ServerStream: serverStreaming(method),
		HasEnums:     hasEnums,
	}
	if cmd.ServerStream {
		g.importTabwriter = true
		g.importOS = true
		g.importIO = true
		g.importStrings = true
	}
	err := g.tmpl.Execute(g, cmd)
	if err != nil {
		g.Fail("Failed to execute cmdTemplate for ", *method.Name, ": ", err.Error(), "\n")
		return
	}
}

func (g *GenCmd) addCmdFlags(service *descriptor.ServiceDescriptorProto) {
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
		g.P(method.Name, "Cmd.Flags().AddFlagSet(", flatType, "FlagSet)")
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
	if *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		// not supported yet
		return false
	}
	return true
}

func argName(strs []string) string {
	str := strings.Join(strs, "-")
	return strings.ToLower(str)
}
