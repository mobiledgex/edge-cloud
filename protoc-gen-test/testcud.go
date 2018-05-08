// Modeled after gogo/protobuf/plugin/testgen/testgen.go testText plugin

package main

import (
	"text/template"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/protogen"
	"github.com/mobiledgex/edge-cloud/util"
)

const edgeproto = "edgeproto"

type TestCud struct {
	*generator.Generator
	cudTmpl *template.Template
}

func (t *TestCud) Name() string {
	return "TestCud"
}

func (t *TestCud) Init(g *generator.Generator) {
	t.Generator = g
	t.cudTmpl = template.Must(template.New("cud").Parse(tmpl))
}

type tmplArgs struct {
	Pkg         string
	Name        string
	KeyName     string
	UpdateField string
}

var tmpl = `
type Show{{.Name}} struct {
	data map[string]{{.Pkg}}.{{.Name}}
	grpc.ServerStream
}

func (x *Show{{.Name}}) Init() {
	x.data = make(map[string]{{.Pkg}}.{{.Name}})
}

func (x *Show{{.Name}}) Send(m *{{.Pkg}}.{{.Name}}) error {
	x.data[m.Key.GetKeyString()] = *m
	return nil
}

func (x *Show{{.Name}}) ReadStream(stream {{.Pkg}}.{{.Name}}Api_Show{{.Name}}Client, err error) {
	x.data = make(map[string]{{.Pkg}}.{{.Name}})
	if err != nil {
		return
	}
	for {
		obj, err := stream.Recv()
		if (err == io.EOF) {
			break
		}
		//util.InfoLog("show {{.Name}}", "key", obj.Key.GetKeyString())
		if err != nil {
			break
		}
		x.data[obj.Key.GetKeyString()] = *obj
	}
}

func (x *Show{{.Name}}) AssertFound(t *testing.T, obj *{{.Pkg}}.{{.Name}}) {
	check, found := x.data[obj.Key.GetKeyString()]
	assert.True(t, found, "find {{.Name}} %s", obj.Key.GetKeyString())
	if found {
		assert.Equal(t, *obj, check, "{{.Name}} are equal")
	}
}

func (x *Show{{.Name}}) AssertNotFound(t *testing.T, obj *{{.Pkg}}.{{.Name}}) {
	_, found := x.data[obj.Key.GetKeyString()]
	assert.False(t, found, "do not find {{.Name}} %s", obj.Key.GetKeyString())
}

// Wrap the api with a common interface
type {{.Name}}CommonApi struct {
	internal_api {{.Pkg}}.{{.Name}}ApiServer
	client_api {{.Pkg}}.{{.Name}}ApiClient
}

func (x *{{.Name}}CommonApi) Create{{.Name}}(ctx context.Context, in *{{.Pkg}}.{{.Name}}) (*{{.Pkg}}.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.Create{{.Name}}(ctx, in)
	} else {
		return x.client_api.Create{{.Name}}(ctx, in)
	}
}

func (x *{{.Name}}CommonApi) Update{{.Name}}(ctx context.Context, in *{{.Pkg}}.{{.Name}}) (*{{.Pkg}}.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.Update{{.Name}}(ctx, in)
	} else {
		return x.client_api.Update{{.Name}}(ctx, in)
	}
}

func (x *{{.Name}}CommonApi) Delete{{.Name}}(ctx context.Context, in *{{.Pkg}}.{{.Name}}) (*{{.Pkg}}.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.Delete{{.Name}}(ctx, in)
	} else {
		return x.client_api.Delete{{.Name}}(ctx, in)
	}
}

func (x *{{.Name}}CommonApi) Show{{.Name}}(ctx context.Context, filter *{{.Pkg}}.{{.Name}}, showData *Show{{.Name}}) error {
	if x.internal_api != nil {
		return x.internal_api.Show{{.Name}}(filter, showData)
	} else {
		stream, err := x.client_api.Show{{.Name}}(ctx, filter)
		showData.ReadStream(stream, err)
		return err
	}
}

func Internal{{.Name}}CudTest(t *testing.T, api {{.Pkg}}.{{.Name}}ApiServer, testData []{{.Pkg}}.{{.Name}}) {
	apiWrap := {{.Name}}CommonApi{}
	apiWrap.internal_api = api
	basic{{.Name}}CudTest(t, &apiWrap, testData)
}

func Client{{.Name}}CudTest(t *testing.T, api {{.Pkg}}.{{.Name}}ApiClient, testData []{{.Pkg}}.{{.Name}}) {
	apiWrap := {{.Name}}CommonApi{}
	apiWrap.client_api = api
	basic{{.Name}}CudTest(t, &apiWrap, testData)
}

func basic{{.Name}}CudTest(t *testing.T, api *{{.Name}}CommonApi, testData []{{.Pkg}}.{{.Name}}) {
	var err error
	ctx := context.TODO()

	if len(testData) < 3 {
		assert.True(t, false, "Need at least 3 test data objects")
		return
	}

	// test create
	for _, obj := range testData {
		_, err = api.Create{{.Name}}(ctx, &obj)
		assert.Nil(t, err, "Create {{.Name}} %s", obj.Key.GetKeyString())
	}
	_, err = api.Create{{.Name}}(ctx, &testData[0])
	assert.NotNil(t, err, "Create duplicate {{.Name}}")

	// test show all items
	show := Show{{.Name}}{}
	show.Init()
	filterNone := {{.Pkg}}.{{.Name}}{}
	err = api.Show{{.Name}}(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	for _, obj := range testData {
		show.AssertFound(t, &obj)
	}
	assert.Equal(t, len(testData), len(show.data), "Show count")

	// test delete
	_, err = api.Delete{{.Name}}(ctx, &testData[0])
	assert.Nil(t, err, "delete {{.Name}} %s", testData[0].Key.GetKeyString())
	show.Init()
	err = api.Show{{.Name}}(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	assert.Equal(t, len(testData) - 1, len(show.data), "Show count")
	show.AssertNotFound(t, &testData[0])
	// test update of missing object
	_, err = api.Update{{.Name}}(ctx, &testData[0])
	assert.NotNil(t, err, "Update missing object")
	// create it back
	_, err = api.Create{{.Name}}(ctx, &testData[0])
	assert.Nil(t, err, "Create {{.Name}} %s", testData[0].Key.GetKeyString())

	// test invalid keys
	bad := {{.Pkg}}.{{.Name}}{}
	_, err = api.Create{{.Name}}(ctx, &bad)
	assert.NotNil(t, err, "Create {{.Name}} with no key info")

{{if .UpdateField}}
	// test update
	updater := {{.Pkg}}.{{.Name}}{}
	updater.Key = testData[0].Key
	updater.{{.UpdateField}} = "update just this"
	updater.Fields = util.GrpcFieldsNew()
	util.GrpcFieldsSet(updater.Fields, {{.Pkg}}.{{.Name}}Field{{.UpdateField}})
	_, err = api.Update{{.Name}}(ctx, &updater)
	assert.Nil(t, err, "Update {{.Name}} %s", testData[0].Key.GetKeyString())

	show.Init()
	updater = testData[0]
	updater.{{.UpdateField}} = "update just this"
	err = api.Show{{.Name}}(ctx, &filterNone, &show)
	assert.Nil(t, err, "show {{.Name}}")
	show.AssertFound(t, &updater)
{{- end}}
}

`

func (t *TestCud) GenerateImports(file *generator.FileDescriptor) {
	hasGenerateCud := false
	hasTestUpdate := false
	for _, msg := range file.Messages() {
		if GetGenerateCud(msg.DescriptorProto) {
			hasGenerateCud = true
		}
	}
	if !hasGenerateCud {
		return
	}
	t.PrintImport("", "google.golang.org/grpc")
	t.PrintImport("edgeproto", "github.com/mobiledgex/edge-cloud/proto")
	t.PrintImport("", "io")
	t.PrintImport("", "testing")
	t.PrintImport("", "context")
	t.PrintImport("", "github.com/stretchr/testify/assert")
	for _, msg := range file.Messages() {
		for _, field := range msg.DescriptorProto.Field {
			if GetTestUpdate(field) {
				hasTestUpdate = true
			}
		}
	}
	if hasTestUpdate {
		t.PrintImport("", "github.com/mobiledgex/edge-cloud/util")
	}
}

func (t *TestCud) Generate(file *generator.FileDescriptor) {
	hasGenerateCud := false
	for _, msg := range file.Messages() {
		if GetGenerateCud(msg.DescriptorProto) {
			hasGenerateCud = true
		}
	}
	if !hasGenerateCud {
		return
	}
	for _, msg := range file.Messages() {
		if GetGenerateCud(msg.DescriptorProto) {
			t.generateTestCud(msg.DescriptorProto)
		}
	}
}

func (t *TestCud) generateTestCud(message *descriptor.DescriptorProto) {
	args := tmplArgs{
		Pkg:     edgeproto,
		Name:    *message.Name,
		KeyName: *message.Name + "Key",
	}
	for _, field := range message.Field {
		if GetTestUpdate(field) {
			args.UpdateField = generator.CamelCase(*field.Name)
		}
	}
	t.P(util.AutoGenComment)
	t.cudTmpl.Execute(t, args)
}

func GetGenerateCud(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateCud, false)
}

func GetTestUpdate(field *descriptor.FieldDescriptorProto) bool {
	return proto.GetBoolExtension(field.Options, protogen.E_TestUpdate, false)
}
