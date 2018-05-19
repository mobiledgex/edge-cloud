package gensupport

import (
	"bufio"
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	plugin "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"github.com/gogo/protobuf/vanity"
	"github.com/gogo/protobuf/vanity/command"
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
	// Map of all packages used from calls to FQTypeName
	// Can be used to generate imports.
	UsedPkgs map[string]*descriptor.FileDescriptorProto
}

func (s *PluginSupport) init(req *plugin.CodeGeneratorRequest) {
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
	for _, protofile := range req.ProtoFile {
		s.ProtoFiles = append(s.ProtoFiles, protofile)
	}
}

// InitFile should be called by the plugin whenever a new file is being
// generated.
func (s *PluginSupport) InitFile() {
	s.UsedPkgs = make(map[string]*descriptor.FileDescriptorProto)
}

// RegisterUsedPkg adds the package to the list
func (s *PluginSupport) RegisterUsedPkg(pkg string, file *descriptor.FileDescriptorProto) {
	pkg = strings.Replace(pkg, ".", "_", -1)
	s.UsedPkgs[pkg] = file
}

// FQTypeName returns the fully qualified type name (includes package
// and parents for nested definitions) for the given generator.Object.
// This also adds the package to a list of used packages for PrintUsedImports().
func (s *PluginSupport) FQTypeName(g *generator.Generator, obj generator.Object) string {
	pkg := *obj.File().Package
	pkg = strings.Replace(pkg, ".", "_", -1)
	s.UsedPkgs[pkg] = obj.File()
	if pkg != "" {
		pkg += "."
	}
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
		file := s.UsedPkgs[pkg]
		ipath := path.Dir(*file.Name)
		if ipath == "." {
			ipath = s.PackageImportPath
		} else if builtinPath, found := g.ImportMap[*file.Name]; found {
			// this handles google/protobuf builtin paths for
			// Timestamp, Empty, etc.
			ipath = builtinPath
		}
		g.PrintImport(pkg, ipath)
	}
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

// GetMsgName returns the hierarchical type name of the Message without package
func GetMsgName(msg *generator.Descriptor) string {
	return strings.Join(msg.TypeName(), "_")
}

// GetEnumName returns the hierarchical type name of the Enum without package
func GetEnumName(en *generator.EnumDescriptor) string {
	return strings.Join(en.TypeName(), "_")
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
	g.P("package ", file.PackageName())
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
		support.init(req)
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
