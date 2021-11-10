package main

import (
	"text/template"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/gensupport"
	"github.com/mobiledgex/edge-cloud/protogen"
)

type ControllerTest struct {
	*generator.Generator
	support           gensupport.PluginSupport
	firstFile         bool
	firstGen          bool
	genFile           bool
	refData           *gensupport.RefData
	deleteTmpl        *template.Template
	trackerTmpl       *template.Template
	dataStoreTmpl     *template.Template
	importContext     bool
	importTesting     bool
	importEdgeproto   bool
	importObjstore    bool
	importRequire     bool
	importConcurrency bool
	importTestutil    bool
}

func (s *ControllerTest) Name() string {
	return "ControllerTest"
}

func (s *ControllerTest) Init(g *generator.Generator) {
	s.Generator = g
	s.support.Init(g.Request)
	s.firstFile = true
	s.firstGen = true
	s.deleteTmpl = template.Must(template.New("delete").Parse(deleteTmpl))
	s.deleteTmpl = template.Must(s.deleteTmpl.Parse(runDeleteTmpl))
	s.deleteTmpl = template.Must(s.deleteTmpl.Parse(dataGenTmpl))
	s.trackerTmpl = template.Must(template.New("tracker").Parse(trackerTmpl))
	s.dataStoreTmpl = template.Must(template.New("dataStore").Parse(dataStoreTmpl))
}

func (s *ControllerTest) GenerateImports(file *generator.FileDescriptor) {
	if s.importContext {
		s.PrintImport("", "context")
	}
	if s.importTesting {
		s.PrintImport("", "testing")
	}
	if s.importEdgeproto {
		s.PrintImport("", "github.com/mobiledgex/edge-cloud/edgeproto")
	}
	if s.importObjstore {
		s.PrintImport("", "github.com/mobiledgex/edge-cloud/objstore")
	}
	if s.importRequire {
		s.PrintImport("", "github.com/stretchr/testify/require")
	}
	if s.importConcurrency {
		s.PrintImport("", "github.com/coreos/etcd/clientv3/concurrency")
	}
	if s.importTestutil {
		s.PrintImport("", "github.com/mobiledgex/edge-cloud/testutil")
	}
}

func (s *ControllerTest) Generate(file *generator.FileDescriptor) {
	s.genFile = false
	s.importContext = false
	s.importTesting = false
	s.importEdgeproto = false
	s.importObjstore = false
	s.importRequire = false
	s.importConcurrency = false
	s.importTestutil = false
	s.support.InitFile()
	if !s.support.GenFile(*file.FileDescriptorProto.Name) {
		return
	}
	if s.firstFile {
		s.refData = s.support.GatherRefData(s.Generator)
		s.firstFile = false
	}
	for _, desc := range file.Messages() {
		if _, ok := s.refData.RefTos[*desc.Name]; ok {
			s.genFile = true
			break
		}
		if _, ok := s.refData.Trackers[*desc.Name]; ok {
			s.genFile = true
			break
		}
		if *desc.Name == AllDataName {
			s.genFile = true
			break
		}
	}
	if !s.genFile {
		return
	}
	s.P(gensupport.AutoGenComment)

	for _, desc := range file.Messages() {
		if refToGroup, ok := s.refData.RefTos[*desc.Name]; ok {
			s.generateDeleteTest(refToGroup)
		}
		if tracker, ok := s.refData.Trackers[*desc.Name]; ok {
			s.generateDeleteTracker(tracker)
		}
		if *desc.Name == AllDataName {
			s.generateDataStore(desc)
		}
	}
	if s.firstGen {
		s.generateAllDeleteTest()
		s.firstGen = false
	}

	gensupport.RunParseCheck(s.Generator, file)

}

func GetGenerateCudStreamout(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateCudStreamout, false)
}

func GetControllerApiStruct(message *descriptor.DescriptorProto) string {
	return gensupport.GetStringExtension(message.Options, protogen.E_ControllerApiStruct, "")
}

func GetGenerateCud(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateCud, false)
}
