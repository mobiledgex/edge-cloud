package main

import (
	"strings"
	"text/template"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/gensupport"
	"github.com/mobiledgex/edge-cloud/protogen"
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
	inMessages         map[string]*generator.Descriptor
	enumArgs           map[string][]*EnumArg
	hideTags           map[string]struct{}
	noconfigFields     map[string]struct{}
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
	importStatus       bool
	importCli          bool
}

func (g *GenCmd) Name() string {
	return "GenCmd"
}

func (g *GenCmd) Init(gen *generator.Generator) {
	g.Generator = gen
	g.tmpl = template.Must(template.New("cmd").Parse(tmpl))
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
	if g.importCli {
		g.PrintImport("", "github.com/mobiledgex/edge-cloud/cli")
	}
	if g.importStatus {
		g.PrintImport("", "google.golang.org/grpc/status")
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
	g.importStatus = false
	g.importCli = false
	g.inMessages = make(map[string]*generator.Descriptor)
	g.enumArgs = make(map[string][]*EnumArg)
	g.hideTags = make(map[string]struct{})
	g.noconfigFields = make(map[string]struct{})
	g.packageName = *file.FileDescriptorProto.Package
	g.support.InitFile()
	if !g.support.GenFile(*file.FileDescriptorProto.Name) {
		return
	}

	g.P(gensupport.AutoGenComment)

	// Generate hidetags functions
	for _, desc := range file.Messages() {
		if desc.File() != file {
			continue
		}
		visited := make([]*generator.Descriptor, 0)
		if gensupport.HasHideTags(g.Generator, desc, protogen.E_Hidetag, visited) {
			g.hideTags[*desc.DescriptorProto.Name] = struct{}{}
			gensupport.GenerateHideTags(g.Generator, &g.support, desc)
			g.importStrings = true
			g.importCli = true
		}
	}

	// Generate service vars which must assigned by the main function.
	// Also generate input vars which will be used to capture input args.
	if len(file.FileDescriptorProto.Service) > 0 {
		for _, service := range file.FileDescriptorProto.Service {
			g.generateServiceVars(file.FileDescriptorProto, service)
			g.generateServiceCmd(file.FileDescriptorProto, service)
		}
	}
	for ii, desc := range file.Messages() {
		if desc.File() != file {
			continue
		}
		gensupport.GenerateMessageArgs(g.Generator, &g.support, desc, false, ii)
	}
	if len(file.FileDescriptorProto.Service) > 0 {
		for _, service := range file.FileDescriptorProto.Service {
			if len(service.Method) == 0 {
				continue
			}
			for ii, method := range service.Method {
				gensupport.GenerateMethodArgs(g.Generator, &g.support, method, false, ii)
			}
		}
	}
	gensupport.RunParseCheck(g.Generator, file)
}

func (g *GenCmd) generateServiceVars(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	if len(service.Method) > 0 {
		g.P("var ", service.Name, "Cmd ", file.Package, ".", *service.Name, "Client")
		g.support.RegisterUsedPkg(*file.Package, file)
		for _, method := range service.Method {
			in := g.GetDesc(method.GetInputType())
			if in == nil || gensupport.ClientStreaming(method) || hasOneof(in) {
				continue
			}
			g.inMessages[g.flatTypeName(*method.InputType)] = in
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
			g.P(method.Name, "Cmd.GenCmd(),")
		}
		g.P("}")
		g.P()
		g.importCobra = true
	}
}

type tmplArgs struct {
	Service              string
	Method               string
	InType               string
	OutType              string
	FQInType             string
	FQOutType            string
	ServerStream         bool
	HasEnums             bool
	SetFields            bool
	OutHideTags          bool
	StreamOutIncremental bool
	Show                 bool
	InputRequired        bool
	HasMethodArgs        bool
	SingularData         bool
}

var tmpl = `
var {{.Method}}Cmd = &cli.Command{
	Use: "{{.Method}}",
{{- if and .Show (not .InputRequired)}}
	OptionalArgs: strings.Join(append({{.InType}}RequiredArgs, {{.InType}}OptionalArgs...), " "),
{{- else if .HasMethodArgs}}
	RequiredArgs: strings.Join({{.Method}}RequiredArgs, " "),
	OptionalArgs: strings.Join({{.Method}}OptionalArgs, " "),
{{- else}}
	RequiredArgs: strings.Join({{.InType}}RequiredArgs, " "),
	OptionalArgs: strings.Join({{.InType}}OptionalArgs, " "),
{{- end}}
	AliasArgs: strings.Join({{.InType}}AliasArgs, " "),
	SpecialArgs: &{{.InType}}SpecialArgs,
	Comments: {{.InType}}Comments,
	ReqData: &{{.FQInType}}{},
	ReplyData: &{{.FQOutType}}{},
	Run: run{{.Method}},
}

func run{{.Method}}(c *cli.Command, args []string) error {
	if cli.SilenceUsage {
		c.CobraCmd.SilenceUsage = true
	}
	obj := c.ReqData.(*{{.FQInType}})
{{- if .SetFields}}
	jsonMap, err := c.ParseInput(args)
	if err != nil {
		return err
	}
	obj.Fields = cli.GetSpecifiedFields(jsonMap, c.ReqData, cli.JsonNamespace)
{{- else}}
	_, err := c.ParseInput(args)
	if err != nil {
		return err
	}
{{- end}}
	return {{.Method}}(c, obj)
}

func {{.Method}}(c *cli.Command, in *{{.FQInType}}) error {
	if {{.Service}}Cmd == nil {
		return fmt.Errorf("{{.Service}} client not initialized")
	}
	ctx := context.Background()
{{- if .ServerStream}}
	stream, err := {{.Service}}Cmd.{{.Method}}(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("{{.Method}} failed: %s", errstr)
	}

	objs := make([]*{{.FQOutType}}, 0)
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			errstr := err.Error()
			st, ok := status.FromError(err)
			if ok {
				errstr = st.Message()
			}
			return fmt.Errorf("{{.Method}} recv failed: %s", errstr)
		}
	{{- if .OutHideTags}}
		{{.OutType}}HideTags(obj)
	{{- end}}
	{{- if .StreamOutIncremental}}
		if cli.OutputStream {
			c.WriteOutput(obj, cli.OutputFormat)
			continue
		}
	{{- end}}
		objs = append(objs, obj)
	}
	if len(objs) == 0 {
		return nil
	}
	c.WriteOutput(objs, cli.OutputFormat)
{{- else}}
	obj, err := {{.Service}}Cmd.{{.Method}}(ctx, in)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("{{.Method}} failed: %s", errstr)
	}
	{{- if .OutHideTags}}
	{{.OutType}}HideTags(obj)
	{{- end}}
	c.WriteOutput(obj, cli.OutputFormat)
{{- end}}
	return nil
}

// this supports "Create" and "Delete" commands on ApplicationData
{{- if .SingularData}}
func {{.Method}}Batch(c *cli.Command, data *{{.FQInType}}, err *error) {
	if *err != nil || data == nil {
		return
	}
	fmt.Printf("{{.Method}} %v\n", data)
	myerr := {{.Method}}(c, data)
	if myerr != nil {
		*err = myerr
	}
}
{{- else}}
func {{.Method}}s(c *cli.Command, data []{{.FQInType}}, err *error) {
	if *err != nil {
		return
	}
	for ii, _ := range data {
		fmt.Printf("{{.Method}} %v\n", data[ii])
		myerr := {{.Method}}(c, &data[ii])
		if myerr != nil {
			*err = myerr
			break
		}
	}
}
{{- end}}

`

func (g *GenCmd) generateMethodCmd(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) bool {
	in := g.GetDesc(method.GetInputType())
	out := g.GetDesc(method.GetOutputType())
	if in == nil || gensupport.ClientStreaming(method) || hasOneof(in) {
		// not supported yet
		return false
	}

	g.importContext = true
	g.importStatus = true
	g.importCli = true
	g.importStrings = true
	_, hasEnums := g.enumArgs[*in.DescriptorProto.Name]
	cmd := &tmplArgs{
		Service:              *service.Name,
		Method:               *method.Name,
		InType:               g.flatTypeName(*method.InputType),
		OutType:              g.flatTypeName(*method.OutputType),
		FQInType:             g.FQTypeName(in),
		FQOutType:            g.FQTypeName(out),
		ServerStream:         gensupport.ServerStreaming(method),
		HasEnums:             hasEnums,
		StreamOutIncremental: gensupport.GetStreamOutIncremental(method),
		InputRequired:        gensupport.GetInputRequired(method),
		HasMethodArgs:        gensupport.HasMethodArgs(method),
		SingularData:         gensupport.GetSingularData(in.DescriptorProto),
	}
	if gensupport.IsShow(method) {
		cmd.Show = true
	}
	if strings.HasPrefix(*method.Name, "Update"+cmd.InType) && gensupport.HasGrpcFields(in.DescriptorProto) {
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
