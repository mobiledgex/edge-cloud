package util

import (
	"bufio"
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/gogo/protobuf/vanity"
	"github.com/gogo/protobuf/vanity/command"
)

const AutoGenComment = "// Auto-generated code: DO NOT EDIT"

// gogo protobuf plugin parse check does not generate line numbers
// if there is a parse error which makes it really hard to track down
// errors in generated code. This function should be called from
// your plugin on the generator after generation is done.
func RunParseCheck(g *generator.Generator, p generator.Plugin, file *generator.FileDescriptor) {
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
			g.Fail("baddd Go source code was generated:", err.Error(), "\n"+src.String())
		}
	}
	g.Reset()
	g.Write(content.Bytes())
}

func RunMain(pkg, fileSuffix string, p generator.Plugin) {
	req := command.Read()
	files := req.GetProtoFile()
	files = vanity.FilterFiles(files, vanity.NotGoogleProtobufDescriptorProto)

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
				//log.Print("no auto-gen comment: " + *file.Content)
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
