package main

import (
	"strconv"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/gensupport"
	"github.com/mobiledgex/edge-cloud/protogen"
)

type ControllerGen struct {
	*generator.Generator
	support gensupport.PluginSupport
}

func (s *ControllerGen) Name() string {
	return "ControllerGen"
}

func (s *ControllerGen) Init(g *generator.Generator) {
	s.Generator = g
	s.support.Init(g.Request)
}

func (s *ControllerGen) GenerateImports(file *generator.FileDescriptor) {
	// imports, if any
}

func (s *ControllerGen) Generate(file *generator.FileDescriptor) {
	s.support.InitFile()
	if !s.support.GenFile(*file.FileDescriptorProto.Name) {
		return
	}

	for _, desc := range file.Enums() {
		s.generateEnum(file, desc)
	}

	if s.Generator.Buffer.Len() != 0 {
		s.P(gensupport.AutoGenComment)
		gensupport.RunParseCheck(s.Generator, file)
	}
}

func (s *ControllerGen) generateEnum(file *generator.FileDescriptor, desc *generator.EnumDescriptor) {
	en := desc.EnumDescriptorProto
	if GetVersionHashOpt(en) {
		s.generateUpgradeFuncs(en)
		s.generateUpgradeFuncNames(en)
	}
}

func (s *ControllerGen) generateUpgradeFuncs(enum *descriptor.EnumDescriptorProto) {
	s.P("var ", enum.Name, "_UpgradeFuncs = map[int32]VersionUpgradeFunc{")
	for _, e := range enum.Value {
		if GetUpgradeFunc(e) != "" {
			s.P(e.Number, ": ", GetUpgradeFunc(e), ",")
		} else {
			s.P(e.Number, ": nil,")
		}
	}
	s.P("}")
}
func (s *ControllerGen) generateUpgradeFuncNames(enum *descriptor.EnumDescriptorProto) {
	s.P("var ", enum.Name, "_UpgradeFuncNames = map[int32]string{")
	for _, e := range enum.Value {
		s.P(e.Number, ": ", strconv.Quote(GetUpgradeFunc(e)), ",")
	}
	s.P("}")
}

func GetVersionHashOpt(enum *descriptor.EnumDescriptorProto) bool {
	return proto.GetBoolExtension(enum.Options, protogen.E_VersionHash, false)
}

func GetUpgradeFunc(enumVal *descriptor.EnumValueDescriptorProto) string {
	return gensupport.GetStringExtension(enumVal.Options, protogen.E_UpgradeFunc, "")
}
