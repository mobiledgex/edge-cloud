package gensupport

import (
	"bufio"
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	plugin "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"github.com/gogo/protobuf/vanity"
	"github.com/gogo/protobuf/vanity/command"
	"github.com/mobiledgex/edge-cloud/edgeprotogen"
	"github.com/mobiledgex/edge-cloud/protogen"
)

const AutoGenComment = "// Auto-generated code: DO NOT EDIT"

// PluginSupport provides support data and functions for the
// generator.Plugin struct that will generate the code.
// The generator.Plugin that will use it should include an
// instance of it and pass it to the RunMain function.
type PluginSupport struct {
	// PackageImportPort is the import path of the proto file being
	// generated
	PackageImportPath string
	// ProtoFiles are all of the proto files that support or possibly
	// are dependencies of the proto file being generated
	ProtoFiles []*descriptor.FileDescriptorProto
	// ProtoFilesGen are all of the proto files in the request to
	// generate.
	ProtoFilesGen map[string]struct{}
	// MessageTypesGen are all the message types that are defined
	// in this package.
	MessageTypesGen map[string]struct{}
	// Map of all packages used from calls to FQTypeName
	// Can be used to generate imports.
	UsedPkgs map[string]*descriptor.FileDescriptorProto
	// Current package, used for plugins adding code to .pb.go.
	// For plugins that are generating files to separate directory
	// and package, this is not needed.
	PbGoPackage string
	// Lookup by file and location of comments
	Comments map[string]map[string]*descriptor.SourceCodeInfo_Location
}

func (s *PluginSupport) Init(req *plugin.CodeGeneratorRequest) {
	// PackageImportPath is the path used in the import statement for
	// structs generated from the proto files.
	// This scheme requires that protoc is called in the Makefile from the
	// same directory where the .proto files exist.
	s.PackageImportPath, _ = os.Getwd()
	// import path is under src
	index := strings.Index(s.PackageImportPath, "/src/")
	if index != -1 {
		s.PackageImportPath = s.PackageImportPath[index+5:]
	}

	s.ProtoFiles = make([]*descriptor.FileDescriptorProto, 0)
	s.ProtoFilesGen = make(map[string]struct{})
	s.MessageTypesGen = make(map[string]struct{})
	s.Comments = make(map[string]map[string]*descriptor.SourceCodeInfo_Location)
	if req != nil {
		for _, filename := range req.FileToGenerate {
			s.ProtoFilesGen[filename] = struct{}{}
		}
		for _, protofile := range req.ProtoFile {
			s.ProtoFiles = append(s.ProtoFiles, protofile)
			if _, ok := s.ProtoFilesGen[protofile.GetName()]; ok {
				for _, desc := range protofile.GetMessageType() {
					s.MessageTypesGen["."+*protofile.Package+"."+*desc.Name] = struct{}{}
				}
			}
			comments := make(map[string]*descriptor.SourceCodeInfo_Location)
			// This replicates what is done in the generator, but
			// allows us access to comments from all files, which
			// the generator does not provide.
			s.Comments[protofile.GetName()] = comments
			for _, loc := range protofile.GetSourceCodeInfo().GetLocation() {
				if loc.LeadingComments == nil {
					continue
				}
				var p []string
				for _, n := range loc.Path {
					p = append(p, strconv.Itoa(int(n)))
				}
				comments[strings.Join(p, ",")] = loc
			}
		}
	}
}

// InitFile should be called by the plugin whenever a new file is being
// generated.
func (s *PluginSupport) InitFile() {
	s.UsedPkgs = make(map[string]*descriptor.FileDescriptorProto)
}

func (s *PluginSupport) GenFile(filename string) bool {
	_, found := s.ProtoFilesGen[filename]
	return found
}

// DottedName gets the proto-style name.
// For normal objects, name should be singular.
// For nested objects, there will be multiple names to include parent objects.
func DottedName(packageName string, names ...string) string {
	all := []string{}
	if packageName != "" {
		all = append(all, packageName)
	}
	all = append(all, names...)
	return "." + strings.Join(all, ".")
}

// GetGeneratorFiles gets the wrapped files of the generator.
// This requires that Generator.WrapTypes() and Generator.BuildTypeNameMap()
// have already been called.
// Generator doesn't provide a way to access the wrapped files
// directly, which are useful if generating something besides golang.
// This gets the files in a back-door kind of way.
func (s *PluginSupport) GetGeneratorFiles(g *generator.Generator) []*generator.FileDescriptor {
	files := []*generator.FileDescriptor{}
	for _, protoFile := range g.Request.ProtoFile {
		if !s.GenFile(protoFile.GetName()) {
			continue
		}
		// find an object in the file
		var name string
		if len(protoFile.MessageType) > 0 {
			name = *protoFile.MessageType[0].Name
		} else if len(protoFile.EnumType) > 0 {
			name = *protoFile.EnumType[0].Name
		} else {
			g.Fail("GetGeneratorFiles: Can't find object for file", *protoFile.Name)
			continue
		}
		obj := g.ObjectNamed(DottedName(protoFile.GetPackage(), name))
		// can get generator file from object
		files = append(files, obj.File())
	}
	return files
}

// RegisterUsedPkg adds the package to the list
func (s *PluginSupport) RegisterUsedPkg(pkg string, file *descriptor.FileDescriptorProto) {
	pkg = strings.Replace(pkg, ".", "_", -1)
	s.UsedPkgs[pkg] = file
}

// SetPbGoPackage should be called when using support to help generate .pb.go,
// with the current package, to prevent generating an import for the
// current package.
func (s *PluginSupport) SetPbGoPackage(pkgName string) {
	s.PbGoPackage = pkgName
}

func (s *PluginSupport) GetPackageName(obj generator.Object) string {
	pkg := *obj.File().Package
	if pkg == s.PbGoPackage {
		pkg = ""
	}
	return strings.Replace(pkg, ".", "_", -1)
}

func (s *PluginSupport) GetPackage(obj generator.Object, ops ...Op) string {
	opts := getOptions(ops)
	pkg := s.GetPackageName(obj)
	if pkg != "" {
		if !opts.noImport {
			s.UsedPkgs[pkg] = obj.File().FileDescriptorProto
		}
		pkg += "."
	}
	return pkg
}

// FQTypeName returns the fully qualified type name (includes package
// and parents for nested definitions) for the given generator.Object.
// This also adds the package to a list of used packages for PrintUsedImports().
func (s *PluginSupport) FQTypeName(g *generator.Generator, obj generator.Object, ops ...Op) string {
	pkg := s.GetPackage(obj, ops...)
	return pkg + generator.CamelCaseSlice(obj.TypeName())
}

// PrintUsedImports will print imports based on calls to FQTypeName() and
// RegisterUsedPkg().
func (s *PluginSupport) PrintUsedImports(g *generator.Generator) {
	// sort used packages so file doesn't change if recompiling
	pkgsSorted := make([]string, len(s.UsedPkgs))
	ii := 0
	for pkg, _ := range s.UsedPkgs {
		pkgsSorted[ii] = pkg
		ii++
	}
	sort.Strings(pkgsSorted)
	for _, pkg := range pkgsSorted {
		if pkg == s.PbGoPackage {
			continue
		}
		file := s.UsedPkgs[pkg]
		ipath := path.Dir(*file.Name)
		if ipath == "." {
			ipath = s.PackageImportPath
		} else if builtinPath, found := g.ImportMap[*file.Name]; found {
			// this handles google/protobuf builtin paths for
			// Timestamp, Empty, etc.
			ipath = builtinPath
		}
		g.PrintImport(generator.GoPackageName(pkg), generator.GoImportPath(ipath))
	}
}

// See generator.GetComments(). The path is a comma separated list of integers.
func (s *PluginSupport) GetComments(fileName string, path string) string {
	comments, ok := s.Comments[fileName]
	if !ok {
		return ""
	}
	loc, ok := comments[path]
	if !ok {
		return ""
	}
	text := strings.TrimSuffix(loc.GetLeadingComments(), "\n")
	return text
}

// GetDesc returns the Descriptor based on the protoc type name
// referenced in Fields and Methods.
func GetDesc(g *generator.Generator, typeName string) *generator.Descriptor {
	obj := g.TypeNameByObject(typeName)
	desc, ok := obj.(*generator.Descriptor)
	if ok {
		return desc
	}
	panic(typeName + " is not of type Descriptor")
}

func GetDescKey(g *generator.Generator, msg *generator.Descriptor) *generator.Descriptor {
	msgProto := msg.DescriptorProto
	if GetObjAndKey(msgProto) {
		return msg
	} else if keyField := GetMessageKey(msgProto); keyField != nil {
		return GetDesc(g, keyField.GetTypeName())
	}
	g.Fail("Key type not found for message ", *msg.Name)
	return nil
}

// GetEnumDesc returns the EnumDescriptor based on the protoc type name
// referenced in Fields.
func GetEnumDesc(g *generator.Generator, typeName string) *generator.EnumDescriptor {
	obj := g.TypeNameByObject(typeName)
	desc, ok := obj.(*generator.EnumDescriptor)
	if ok {
		return desc
	}
	panic(typeName + " is not of type EnumDescriptor")
}

// This assumes the Message is in the same package as the given file,
// but possibly in a different file.
func GetPackageDesc(g *generator.Generator, file *generator.FileDescriptor, name string) *generator.Descriptor {
	dottedName := DottedName(file.GetPackage(), name)
	return GetDesc(g, dottedName)
}

// GetMsgName returns the hierarchical type name of the Message without package
func GetMsgName(msg *generator.Descriptor) string {
	return strings.Join(msg.TypeName(), "_")
}

// GetEnumName returns the hierarchical type name of the Enum without package
func GetEnumName(en *generator.EnumDescriptor) string {
	return strings.Join(en.TypeName(), "_")
}

func WasVisited(desc *generator.Descriptor, visited []*generator.Descriptor) bool {
	for _, d := range visited {
		if desc == d {
			return true
		}
	}
	return false
}

type Options struct {
	noImport bool
}

type Op func(opts *Options)

func WithNoImport() Op {
	return func(opts *Options) { opts.noImport = true }
}

func getOptions(ops []Op) *Options {
	options := Options{}
	for _, op := range ops {
		op(&options)
	}
	return &options
}

// Similar to generator.GoType(), but does not prepend any array or pointer
// references (* or &).
func (s *PluginSupport) GoType(g *generator.Generator, field *descriptor.FieldDescriptorProto, ops ...Op) string {
	typ := ""
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		typ = "float64"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		typ = "float32"
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		typ = "int64"
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		typ = "uint64"
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		typ = "int32"
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		typ = "uint32"
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		typ = "uint64"
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		typ = "uint32"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		typ = "bool"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		typ = "string"
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		g.Fail("group type not allowed")
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		desc := GetDesc(g, field.GetTypeName())
		typ = s.FQTypeName(g, desc, ops...)
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		typ = "[]byte"
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		desc := GetEnumDesc(g, field.GetTypeName())
		typ = s.FQTypeName(g, desc, ops...)
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		typ = "int32"
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		typ = "int64"
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		typ = "int32"
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		typ = "int64"
	default:
		g.Fail("unknown type for", field.GetName())
	}
	return typ
}

// ConvTypeNames takes a protoc format type name (as used in Fields and
// Methods) and returns the package plus a Go-ified type name.
// The protoc format is .package.Name or .package.Parent.Name for nested
// types.
func ConvTypeName(typeName string) (string, string) {
	if typeName[0] != '.' {
		return "", strings.Replace(typeName, ".", "_", -1)
	}
	typeName = typeName[1:]
	index := strings.Index(typeName, ".")
	if index == -1 {
		return "", typeName
	}
	pkg := typeName[:index]
	return pkg, strings.Replace(typeName[index+1:], ".", "_", -1)
}

func GetStringExtension(pb proto.Message, extension *proto.ExtensionDesc, def string) string {
	if reflect.ValueOf(pb).IsNil() {
		return def
	}
	value, err := proto.GetExtension(pb, extension)
	if err == nil && value.(*string) != nil {
		return *(value.(*string))
	}
	return def
}

func FindStringExtension(pb proto.Message, extension *proto.ExtensionDesc) (string, bool) {
	if reflect.ValueOf(pb).IsNil() {
		return "", false
	}
	value, err := proto.GetExtension(pb, extension)
	if err == nil && value.(*string) != nil {
		return *(value.(*string)), true
	}
	return "", false
}

func HasExtension(pb proto.Message, extension *proto.ExtensionDesc) bool {
	if reflect.ValueOf(pb).IsNil() {
		return false
	}
	value, err := proto.GetExtension(pb, extension)
	return err == nil && value != nil
}

func IsRepeated(field *descriptor.FieldDescriptorProto) bool {
	return *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED
}

func HasGrpcFields(message *descriptor.DescriptorProto) bool {
	if message.Field != nil && len(message.Field) > 0 && *message.Field[0].Name == "fields" && *message.Field[0].Type == descriptor.FieldDescriptorProto_TYPE_STRING {
		return true
	}
	return false
}

func FindField(message *descriptor.DescriptorProto, camelCaseName string) *descriptor.FieldDescriptorProto {
	for _, field := range message.Field {
		name := generator.CamelCase(*field.Name)
		if name == camelCaseName {
			return field
		}
	}
	return nil
}

func FindHierField(g *generator.Generator, message *descriptor.DescriptorProto, camelCaseHierName string) (*descriptor.DescriptorProto, *descriptor.FieldDescriptorProto, error) {
	names := strings.Split(camelCaseHierName, ".")
	for {
		name := names[0]
		field := FindField(message, name)
		if field == nil {
			return message, field, fmt.Errorf("No field %s on %s", *field.Name, *message.Name)
		}
		if len(names) == 1 {
			return message, field, nil
		}
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			subDesc := GetDesc(g, field.GetTypeName())
			message = subDesc.DescriptorProto
			names = names[1:]
		} else {
			return message, field, fmt.Errorf("Field %s on %s is not a message and has no children", strings.Join(names, "."), *message.Name)
		}
	}
}

func GetObjAndKey(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_ObjAndKey, false)
}

func GetE2edata(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_E2Edata, false)
}

func GetCustomKeyType(message *descriptor.DescriptorProto) string {
	return GetStringExtension(message.Options, protogen.E_CustomKeyType, "")
}

func GetSingularData(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_SingularData, false)
}

func HasHideTags(g *generator.Generator, desc *generator.Descriptor, hideTag *proto.ExtensionDesc, visited []*generator.Descriptor) bool {
	if WasVisited(desc, visited) {
		return false
	}
	msg := desc.DescriptorProto
	for _, field := range msg.Field {
		if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			subDesc := GetDesc(g, field.GetTypeName())
			if HasHideTags(g, subDesc, hideTag, append(visited, desc)) {
				return true
			}
		}
		if GetStringExtension(field.Options, hideTag, "") != "" {
			return true
		}
	}
	return false
}

func GetMessageKey(message *descriptor.DescriptorProto) *descriptor.FieldDescriptorProto {
	if message.Field == nil {
		return nil
	}
	if len(message.Field) > 0 && *message.Field[0].Name == "key" {
		return message.Field[0]
	}
	if len(message.Field) > 1 && HasGrpcFields(message) && *message.Field[1].Name == "key" {
		return message.Field[1]
	}
	return nil
}

func (s *PluginSupport) GetMessageKeyType(g *generator.Generator, desc *generator.Descriptor) (string, error) {
	message := desc.DescriptorProto
	if field := GetMessageKey(message); field != nil {
		if typ := gogoproto.GetCastType(field); typ != "" {
			return typ, nil
		}
		return s.GoType(g, field), nil
	} else if typ := GetCustomKeyType(message); typ != "" {
		pkg := s.GetPackage(desc)
		return pkg + typ, nil
	} else if GetObjAndKey(message) {
		return s.FQTypeName(g, desc), nil
	}
	return "", fmt.Errorf("No Key field for %s, use field named Key or protogen.obj_and_key option or protogen.custom_key_type option", *message.Name)
}

type MapType struct {
	KeyField     *descriptor.FieldDescriptorProto
	ValField     *descriptor.FieldDescriptorProto
	KeyType      string
	ValType      string
	ValIsMessage bool
	FlagType     string
	DefValue     string
}

func (s *PluginSupport) GetMapType(g *generator.Generator, field *descriptor.FieldDescriptorProto, ops ...Op) *MapType {
	if *field.Type != descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		return nil
	}
	desc := GetDesc(g, field.GetTypeName())
	if !desc.GetOptions().GetMapEntry() {
		return nil
	}
	m := MapType{}
	m.KeyField = desc.Field[0]
	m.ValField = desc.Field[1]
	m.KeyType = s.GoType(g, m.KeyField, ops...)
	m.ValType = s.GoType(g, m.ValField, ops...)
	if *m.ValField.Type == descriptor.FieldDescriptorProto_TYPE_STRING &&
		*m.KeyField.Type == descriptor.FieldDescriptorProto_TYPE_STRING {
		m.FlagType = "StringToString"
		m.DefValue = "map[string]string{}"
		return &m
	}
	if *m.ValField.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		m.ValIsMessage = true
	}
	return &m
}

// RunParseCheck will run the parser to check for parse errors in the
// generated code. While the gogo generator does this as well, if there
// is a failure it does not generate line numbers, which makes it very
// difficult to locate the line with the error. This function can be
// called at the end of the Generate() function to check the generated code.
// At that point the generated code will be missing the imports and
// some of the other header code generated by the gogo generator code,
// but that is the last place it can be called before the gogo generator
// parser runs.
func RunParseCheck(g *generator.Generator, file *generator.FileDescriptor) {
	if g.Buffer.Len() == 0 {
		return
	}
	content := g.Buffer
	g.Buffer = new(bytes.Buffer)
	g.P("package ", file.FileDescriptorProto.GetPackage())
	g.Write(content.Bytes())

	fset := token.NewFileSet()
	raw := g.Bytes()
	_, err := parser.ParseFile(fset, "", g, parser.ParseComments)
	if err != nil {
		var src bytes.Buffer
		s := bufio.NewScanner(bytes.NewReader(raw))
		for line := 1; s.Scan(); line++ {
			fmt.Fprintf(&src, "%5d\t%s\n", line, s.Bytes())
		}
		if serr := s.Err(); serr == nil {
			g.Fail("bad Go source code was generated:", err.Error(), "\n"+src.String())
		}
	}
	g.Reset()
	g.Write(content.Bytes())
}

// RunMain should be called by the main function with the plugin
// that will be used to generate the code. The pkg string is
// the name of the package used in the generated files.
// The fileSuffix will replace .pb.go as the generated file suffix.
// The target directory of the generated files is controlled by
// the call to protoc, and cannot be manipulated here.
// If a PluginSupport is provided, it will be initialized so that
// support functions can be used by the plugin.
func RunMain(pkg, fileSuffix string, p generator.Plugin, support *PluginSupport) {
	req := command.Read()
	files := req.GetProtoFile()
	files = vanity.FilterFiles(files, vanity.NotGoogleProtobufDescriptorProto)

	if support != nil {
		support.Init(req)
	}

	args := strings.Split(req.GetParameter(), ",")
	for _, arg := range args {
		kv := strings.Split(arg, "=")
		if len(kv) == 2 {
			switch kv[0] {
			case "suffix":
				fileSuffix = kv[1]
			case "pkg":
				pkg = kv[1]
			}
		}
	}

	// override package name
	for _, protofile := range req.ProtoFile {
		if protofile.Options == nil {
			protofile.Options = &descriptor.FileOptions{}
		}
		protofile.Options.GoPackage = &pkg
	}

	resp := command.GeneratePlugin(req, p, fileSuffix)

	// not really any better way to avoid printing files with no
	// test output (files are not empty due to some header stuff)
	if len(resp.File) > 0 {
		ii := 0
		for _, file := range resp.File {
			if !strings.Contains(*file.Content, AutoGenComment) {
				continue
			}
			// copy and increment index
			resp.File[ii] = file
			ii++
		}
		resp.File = resp.File[:ii]
	}
	command.Write(resp)
}

func ClientStreaming(method *descriptor.MethodDescriptorProto) bool {
	if method.ClientStreaming == nil {
		return false
	}
	return *method.ClientStreaming
}

func ServerStreaming(method *descriptor.MethodDescriptorProto) bool {
	if method.ServerStreaming == nil {
		return false
	}
	return *method.ServerStreaming
}

func IsShow(method *descriptor.MethodDescriptorProto) bool {
	if GetNonStandardShow(method) {
		return false
	}
	return GetCamelCasePrefix(*method.Name) == "Show"
}

func GetEnumBackend(enumVal *descriptor.EnumValueDescriptorProto) bool {
	return proto.GetBoolExtension(enumVal.Options, edgeprotogen.E_EnumBackend, false)
}

func GetNonStandardShow(method *descriptor.MethodDescriptorProto) bool {
	return proto.GetBoolExtension(method.Options, protogen.E_NonStandardShow, false)
}

type MethodInfo struct {
	Name     string
	Prefix   string
	Stream   bool
	Mc2Api   bool
	IsShow   bool
	IsUpdate bool
	Method   *descriptor.MethodDescriptorProto
	Out      *generator.Descriptor
	OutType  string
}

type MethodGroup struct {
	ServiceName  string
	MethodInfos  []*MethodInfo
	InType       string
	In           *generator.Descriptor
	HasStream    bool
	HasUpdate    bool
	HasMc2Api    bool
	HasShow      bool
	SingularData bool
	Suffix       string
}

func (m *MethodGroup) ApiName() string {
	return m.ServiceName + m.Suffix
}

func GetCamelCasePrefix(name string) string {
	if name == "" {
		return ""
	}
	for ii := 1; ii < len(name); ii++ {
		if unicode.IsUpper(rune(name[ii])) {
			return name[:ii]
		}
	}
	return name
}

func GetMethodInfo(g *generator.Generator, method *descriptor.MethodDescriptorProto) (*generator.Descriptor, *MethodInfo) {
	in := GetDesc(g, method.GetInputType())

	info := MethodInfo{}
	info.Name = *method.Name
	prefix := GetCamelCasePrefix(*method.Name)
	if *method.Name == prefix+*in.DescriptorProto.Name {
		// use prefixes only when prefixed against input object type
		info.Prefix = prefix
	} else {
		info.Prefix = *method.Name
	}
	info.Out = GetDesc(g, method.GetOutputType())
	info.OutType = *info.Out.DescriptorProto.Name
	info.Method = method
	if ServerStreaming(method) {
		info.Stream = true
	}
	if GetStringExtension(method.Options, protogen.E_Mc2Api, "") != "" {
		info.Mc2Api = true
	}
	if IsShow(method) {
		info.IsShow = true
	}
	if info.Prefix == "Update" {
		info.IsUpdate = true
	}
	return in, &info
}

// group methods by input type
func CollectMethodGroups(g *generator.Generator, service *descriptor.ServiceDescriptorProto, groups map[string]*MethodGroup) {
	for _, method := range service.Method {
		in, info := GetMethodInfo(g, method)
		if info == nil {
			continue
		}
		inType := *in.DescriptorProto.Name
		group, found := groups[inType]
		if !found {
			group = &MethodGroup{}
			group.ServiceName = *service.Name
			group.In = in
			group.InType = inType
			group.SingularData = GetSingularData(in.DescriptorProto)
			if inType+"Api" != *service.Name {
				group.Suffix = "_" + inType
			}
			groups[inType] = group
		}
		group.MethodInfos = append(group.MethodInfos, info)
		if info.Stream {
			group.HasStream = true
		}
		if info.IsUpdate {
			group.HasUpdate = true
		}
		if info.Mc2Api {
			group.HasMc2Api = true
		}
		if info.IsShow {
			group.HasShow = true
		}
	}
}

func GetMethodGroups(g *generator.Generator, service *descriptor.ServiceDescriptorProto) []*MethodGroup {
	groups := make(map[string]*MethodGroup)
	CollectMethodGroups(g, service, groups)
	// convert map into sorted list
	groupsSorted := make([]*MethodGroup, 0)
	for _, group := range groups {
		groupsSorted = append(groupsSorted, group)
	}
	sort.Slice(groupsSorted, func(i, j int) bool {
		return groupsSorted[i].InType < groupsSorted[j].InType
	})
	return groupsSorted
}

func GetAllMethodGroups(g *generator.Generator, support *PluginSupport) map[string]*MethodGroup {
	groups := make(map[string]*MethodGroup)
	for _, protofile := range support.ProtoFiles {
		if !support.GenFile(protofile.GetName()) {
			continue
		}
		for _, svc := range protofile.GetService() {
			if len(svc.Method) == 0 {
				continue
			}
			CollectMethodGroups(g, svc, groups)
		}
	}
	return groups
}

func GetFirstFile(gen *generator.Generator) string {
	// Generator passes us all files (some of which are builtin
	// like google/api/http). To determine the first file to generate
	// one-off code, sort by request files which are the subset of
	// files we will generate code for.
	files := make([]string, len(gen.Request.FileToGenerate))
	copy(files, gen.Request.FileToGenerate)
	sort.Strings(files)
	if len(files) > 0 {
		return files[0]
	}
	return ""
}
