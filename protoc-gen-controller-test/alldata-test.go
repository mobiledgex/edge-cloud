package main

import (
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/gensupport"
)

var AllDataName = "AllData"

type dataStoreArgs struct {
	Type          string
	Fields        []*dataStoreFieldArgs
	FieldsReverse []*dataStoreFieldArgs
	Op            string
}

type dataStoreFieldArgs struct {
	Name     string
	Type     string
	Repeated bool
	ApiObj   string
}

var dataStoreTmpl = `
type testSupportData edgeproto.{{.Type}}

func (s *testSupportData) put(t *testing.T, ctx context.Context, all *AllApis) {
{{- range .Fields}}
{{- if .Repeated}}
	for _, obj := range s.{{.Name}} {
		_, err := all.{{.ApiObj}}.store.Put(ctx, &obj, all.{{.ApiObj}}.sync.syncWait)
		require.Nil(t, err)
	}
{{- else}}
	if s.{{.Name}} != nil {
		_, err := all.{{.ApiObj}}.store.Put(ctx, s.{{.Name}}, all.{{.ApiObj}}.sync.syncWait)
		require.Nil(t, err)
	}
{{- end}}
{{- end}}
}

func (s *testSupportData) delete(t *testing.T, ctx context.Context, all *AllApis) {
{{- range .FieldsReverse}}
{{- if .Repeated}}
	for _, obj := range s.{{.Name}} {
		_, err := all.{{.ApiObj}}.store.Delete(ctx, &obj, all.{{.ApiObj}}.sync.syncWait)
		require.Nil(t, err)
	}
{{- else}}
	if s.{{.Name}} != nil {
		_, err := all.{{.ApiObj}}.store.Delete(ctx, s.{{.Name}}, all.{{.ApiObj}}.sync.syncWait)
		require.Nil(t, err)
	}
{{- end}}
{{- end}}
}

{{- range .Fields}}

func (s *testSupportData) getOne{{.Type}}() *edgeproto.{{.Type}} {
{{- if .Repeated}}
	if len(s.{{.Name}}) == 0 {
		return nil
	}
	return &s.{{.Name}}[0]
{{- else}}
	return s.{{.Name}}
{{- end}}
}
{{- end}}

`

func (s *ControllerTest) generateDataStore(desc *generator.Descriptor) {
	message := desc.DescriptorProto
	args := dataStoreArgs{}
	args.Type = *message.Name

	for _, field := range message.Field {
		if *field.Type != descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			continue
		}
		subDesc := gensupport.GetDesc(s.Generator, field.GetTypeName())
		if !GetGenerateCud(subDesc.DescriptorProto) {
			continue
		}
		fieldArgs := dataStoreFieldArgs{}
		fieldArgs.Name = generator.CamelCase(*field.Name)
		fieldArgs.Type = *subDesc.Name
		fieldArgs.Repeated = *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED
		fieldArgs.ApiObj = getApiObj(subDesc)
		args.Fields = append(args.Fields, &fieldArgs)
		args.FieldsReverse = append([]*dataStoreFieldArgs{&fieldArgs}, args.FieldsReverse...)
	}
	if len(args.Fields) == 0 {
		return
	}
	s.dataStoreTmpl.Execute(s.Generator, args)
	s.importEdgeproto = true
	s.importContext = true
	s.importTesting = true
	s.importRequire = true
}
