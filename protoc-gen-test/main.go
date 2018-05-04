package main

import (
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/vanity"
	"github.com/gogo/protobuf/vanity/command"
)

const AutoGenComment = "// Auto-generate Cud Tests"

func main() {
	req := command.Read()
	files := req.GetProtoFile()
	files = vanity.FilterFiles(files, vanity.NotGoogleProtobufDescriptorProto)

	// override package name
	pkg := "testutil"
	for _, protofile := range req.ProtoFile {
		if protofile.Options == nil {
			protofile.Options = &descriptor.FileOptions{}
		}
		protofile.Options.GoPackage = &pkg
	}

	testcud := TestCud{}
	resp := command.GeneratePlugin(req, &testcud, "_testutil.go")

	// not really any better way to avoid printing files with no
	// test output (files are not empty due to some header stuff)
	if len(resp.File) > 0 {
		ii := 0
		for _, file := range resp.File {
			if !strings.Contains(*resp.File[ii].Content, AutoGenComment) {
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
