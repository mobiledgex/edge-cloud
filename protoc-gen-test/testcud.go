// Modeled after gogo/protobuf/plugin/testgen/testgen.go testText plugin

package main

import (
	"text/template"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/mobiledgex/edge-cloud/gensupport"
	"github.com/mobiledgex/edge-cloud/protogen"
)

const edgeproto = "edgeproto"

type TestCud struct {
	*generator.Generator
	support         gensupport.PluginSupport
	cudTmpl         *template.Template
	methodTmpl      *template.Template
	firstFile       bool
	importEdgeproto bool
	importIO        bool
	importTesting   bool
	importContext   bool
	importTime      bool
	importAssert    bool
	importGrpc      bool
}

func (t *TestCud) Name() string {
	return "TestCud"
}

func (t *TestCud) Init(g *generator.Generator) {
	t.Generator = g
	t.support.Init(g.Request)
	t.cudTmpl = template.Must(template.New("cud").Parse(tmpl))
	t.methodTmpl = template.Must(template.New("method").Parse(methodTmpl))
	t.firstFile = true
}

type cudFunc struct {
	Func      string
	Pkg       string
	Name      string
	KeyName   string
	Streamout bool
}

type tmplArgs struct {
	Pkg         string
	Name        string
	KeyName     string
	UpdateField string
	UpdateValue string
	ShowOnly    bool
	Streamout   bool
	CudFuncs    []cudFunc
}

var tmpl = `
type Show{{.Name}} struct {
	Data map[string]{{.Pkg}}.{{.Name}}
	grpc.ServerStream
}

func (x *Show{{.Name}}) Init() {
	x.Data = make(map[string]{{.Pkg}}.{{.Name}})
}

func (x *Show{{.Name}}) Send(m *{{.Pkg}}.{{.Name}}) error {
	x.Data[m.Key.GetKeyString()] = *m
	return nil
}

{{- if .Streamout}}
type CudStreamout{{.Name}} struct {
	grpc.ServerStream
}

func (x *CudStreamout{{.Name}}) Send(res *{{.Pkg}}.Result) error {
	fmt.Println(res)
	return nil
}
func (x *CudStreamout{{.Name}}) Context() context.Context {
	return context.TODO()
}

type {{.Name}}Stream interface {
	Recv() (*{{.Pkg}}.Result, error)
}

func {{.Name}}ReadResultStream(stream {{.Name}}Stream, err error) error {
	if err != nil {
		return err
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Println(res)
	}
}

{{- end}}

func (x *Show{{.Name}}) ReadStream(stream {{.Pkg}}.{{.Name}}Api_Show{{.Name}}Client, err error) {
	x.Data = make(map[string]{{.Pkg}}.{{.Name}})
	if err != nil {
		return
	}
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		x.Data[obj.Key.GetKeyString()] = *obj
	}
}

func (x *Show{{.Name}}) CheckFound(obj *{{.Pkg}}.{{.Name}}) bool {
	_, found := x.Data[obj.Key.GetKeyString()]
	return found
}

func (x *Show{{.Name}}) AssertFound(t *testing.T, obj *{{.Pkg}}.{{.Name}}) {
	check, found := x.Data[obj.Key.GetKeyString()]
	assert.True(t, found, "find {{.Name}} %s", obj.Key.GetKeyString())
	if found && !check.Matches(obj, {{.Pkg}}.MatchIgnoreBackend(), {{.Pkg}}.MatchSortArrayedKeys()) {
		assert.Equal(t, *obj, check, "{{.Name}} are equal")
	}
	if found {
		// remove in case there are dups in the list, so the
		// same object cannot be used again
		delete(x.Data, obj.Key.GetKeyString())
	}
}

func (x *Show{{.Name}}) AssertNotFound(t *testing.T, obj *{{.Pkg}}.{{.Name}}) {
	_, found := x.Data[obj.Key.GetKeyString()]
	assert.False(t, found, "do not find {{.Name}} %s", obj.Key.GetKeyString())
}

func WaitAssertFound{{.Name}}(t *testing.T, api {{.Pkg}}.{{.Name}}ApiClient, obj *{{.Pkg}}.{{.Name}}, count int, retry time.Duration) {
	show := Show{{.Name}}{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.Show{{.Name}}(ctx, obj)
		show.ReadStream(stream, err)
		cancel()
		if show.CheckFound(obj) {
			break
		}
		time.Sleep(retry)
	}
	show.AssertFound(t, obj)
}

func WaitAssertNotFound{{.Name}}(t *testing.T, api {{.Pkg}}.{{.Name}}ApiClient, obj *{{.Pkg}}.{{.Name}}, count int, retry time.Duration) {
	show := Show{{.Name}}{}
	filterNone := {{.Pkg}}.{{.Name}}{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.Show{{.Name}}(ctx, &filterNone)
		show.ReadStream(stream, err)
		cancel()
		if !show.CheckFound(obj) {
			break
		}
		time.Sleep(retry)
	}
	show.AssertNotFound(t, obj)
}

// Wrap the api with a common interface
type {{.Name}}CommonApi struct {
	internal_api {{.Pkg}}.{{.Name}}ApiServer
	client_api {{.Pkg}}.{{.Name}}ApiClient
}

{{- if not .ShowOnly}}
{{range .CudFuncs}}
func (x *{{.Name}}CommonApi) {{.Func}}{{.Name}}(ctx context.Context, in *{{.Pkg}}.{{.Name}}) (*{{.Pkg}}.Result, error) {
	copy := &{{.Pkg}}.{{.Name}}{}
	*copy = *in
{{- if .Streamout}}
	if x.internal_api != nil {
		err := x.internal_api.{{.Func}}{{.Name}}(copy, &CudStreamout{{.Name}}{})
		return &{{.Pkg}}.Result{}, err
	} else {
		stream, err := x.client_api.{{.Func}}{{.Name}}(ctx, copy)
		err = {{.Name}}ReadResultStream(stream, err)
		return &{{.Pkg}}.Result{}, err
	}
{{- else}}
	if x.internal_api != nil {
		return x.internal_api.{{.Func}}{{.Name}}(ctx, copy)
	} else {
		return x.client_api.{{.Func}}{{.Name}}(ctx, copy)
	}
{{- end}}
}
{{end}}
{{- end}}

func (x *{{.Name}}CommonApi) Show{{.Name}}(ctx context.Context, filter *{{.Pkg}}.{{.Name}}, showData *Show{{.Name}}) error {
	if x.internal_api != nil {
		return x.internal_api.Show{{.Name}}(filter, showData)
	} else {
		stream, err := x.client_api.Show{{.Name}}(ctx, filter)
		showData.ReadStream(stream, err)
		return err
	}
}

func NewInternal{{.Name}}Api(api {{.Pkg}}.{{.Name}}ApiServer) *{{.Name}}CommonApi {
	apiWrap := {{.Name}}CommonApi{}
	apiWrap.internal_api = api
	return &apiWrap
}

func NewClient{{.Name}}Api(api {{.Pkg}}.{{.Name}}ApiClient) *{{.Name}}CommonApi {
	apiWrap := {{.Name}}CommonApi{}
	apiWrap.client_api = api
	return &apiWrap
}

func Internal{{.Name}}Test(t *testing.T, test string, api {{.Pkg}}.{{.Name}}ApiServer, testData []{{.Pkg}}.{{.Name}}) {
	switch test {
{{- if not .ShowOnly}}
	case "cud":
		basic{{.Name}}CudTest(t, NewInternal{{.Name}}Api(api), testData)
{{- end}}
	case "show":
		basic{{.Name}}ShowTest(t, NewInternal{{.Name}}Api(api), testData)
	}
}

func Client{{.Name}}Test(t *testing.T, test string, api {{.Pkg}}.{{.Name}}ApiClient, testData []{{.Pkg}}.{{.Name}}) {
	switch test {
{{- if not .ShowOnly}}
	case "cud":
		basic{{.Name}}CudTest(t, NewClient{{.Name}}Api(api), testData)
{{- end}}
	case "show":
		basic{{.Name}}ShowTest(t, NewClient{{.Name}}Api(api), testData)
	}
}

func basic{{.Name}}ShowTest(t *testing.T, api *{{.Name}}CommonApi, testData []{{.Pkg}}.{{.Name}}) {
	var err error
	ctx := context.TODO()

	show := Show{{.Name}}{}
	show.Init()
	filterNone := {{.Pkg}}.{{.Name}}{}
	err = api.Show{{.Name}}(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	assert.Equal(t, len(testData), len(show.Data), "Show count")
	for _, obj := range testData {
		show.AssertFound(t, &obj)
	}
}

func Get{{.Name}}(t *testing.T, api *{{.Name}}CommonApi, key *{{.KeyName}}, out *{{.Pkg}}.{{.Name}}) bool {
	var err error
	ctx := context.TODO()

	show := Show{{.Name}}{}
	show.Init()
	filter := {{.Pkg}}.{{.Name}}{}
	filter.Key = *key
	err = api.Show{{.Name}}(ctx, &filter, &show)
	assert.Nil(t, err, "show data")
	obj, found := show.Data[key.GetKeyString()]
	if found {
		*out = obj
	}
	return found
}

{{ if not .ShowOnly}}
func basic{{.Name}}CudTest(t *testing.T, api *{{.Name}}CommonApi, testData []{{.Pkg}}.{{.Name}}) {
	var err error
	ctx := context.TODO()

	if len(testData) < 3 {
		assert.True(t, false, "Need at least 3 test data objects")
		return
	}

	// test create
	create{{.Name}}Data(t, api, testData)

	// test duplicate create - should fail
	_, err = api.Create{{.Name}}(ctx, &testData[0])
	assert.NotNil(t, err, "Create duplicate {{.Name}}")

	// test show all items
	basic{{.Name}}ShowTest(t, api, testData)

	// test delete
	_, err = api.Delete{{.Name}}(ctx, &testData[0])
	assert.Nil(t, err, "delete {{.Name}} %s", testData[0].Key.GetKeyString())
	show := Show{{.Name}}{}
	show.Init()
	filterNone := {{.Pkg}}.{{.Name}}{}
	err = api.Show{{.Name}}(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	assert.Equal(t, len(testData) - 1, len(show.Data), "Show count")
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
	updater.{{.UpdateField}} = {{.UpdateValue}}
	updater.Fields = make([]string, 0)
	updater.Fields = append(updater.Fields, {{.Pkg}}.{{.Name}}Field{{.UpdateField}})
	_, err = api.Update{{.Name}}(ctx, &updater)
	assert.Nil(t, err, "Update {{.Name}} %s", testData[0].Key.GetKeyString())

	show.Init()
	updater = testData[0]
	updater.{{.UpdateField}} = {{.UpdateValue}}
	err = api.Show{{.Name}}(ctx, &filterNone, &show)
	assert.Nil(t, err, "show {{.Name}}")
	show.AssertFound(t, &updater)

	// revert change
	updater.{{.UpdateField}} = testData[0].{{.UpdateField}}
	_, err = api.Update{{.Name}}(ctx, &updater)
	assert.Nil(t, err, "Update back {{.Name}}")
{{- end}}
}

func Internal{{.Name}}Create(t *testing.T, api {{.Pkg}}.{{.Name}}ApiServer, testData []{{.Pkg}}.{{.Name}}) {
	create{{.Name}}Data(t, NewInternal{{.Name}}Api(api), testData)
}

func Client{{.Name}}Create(t *testing.T, api {{.Pkg}}.{{.Name}}ApiClient, testData []{{.Pkg}}.{{.Name}}) {
	create{{.Name}}Data(t, NewClient{{.Name}}Api(api), testData)
}

func create{{.Name}}Data(t *testing.T, api *{{.Name}}CommonApi, testData []{{.Pkg}}.{{.Name}}) {
	var err error
	ctx := context.TODO()

	for _, obj := range testData {
		_, err = api.Create{{.Name}}(ctx, &obj)
		assert.Nil(t, err, "Create {{.Name}} %s", obj.Key.GetKeyString())
	}
}
{{- end}}
`

func (t *TestCud) GenerateImports(file *generator.FileDescriptor) {
	if t.importGrpc {
		t.PrintImport("", "google.golang.org/grpc")
	}
	if t.importEdgeproto {
		t.PrintImport("", "github.com/mobiledgex/edge-cloud/edgeproto")
	}
	if t.importIO {
		t.PrintImport("", "io")
	}
	if t.importTesting {
		t.PrintImport("", "testing")
	}
	if t.importContext {
		t.PrintImport("", "context")
	}
	if t.importTime {
		t.PrintImport("", "time")
	}
	if t.importAssert {
		t.PrintImport("", "github.com/stretchr/testify/assert")
	}
}

func (t *TestCud) Generate(file *generator.FileDescriptor) {
	t.importGrpc = false
	t.importEdgeproto = false
	t.importIO = false
	t.importTesting = false
	t.importContext = false
	t.importTime = false
	t.importAssert = false
	t.support.InitFile()
	if !t.support.GenFile(*file.FileDescriptorProto.Name) {
		return
	}
	hasCudMethod := false
	if len(file.FileDescriptorProto.Service) != 0 {
		for _, service := range file.FileDescriptorProto.Service {
			if len(service.Method) > 0 {
				for _, method := range service.Method {
					in := gensupport.GetDesc(t.Generator,
						method.GetInputType())
					if GetGenerateCud(in.DescriptorProto) {
						hasCudMethod = true
						break
					}
				}
			}
		}
	}
	hasGenerateCudTest := false
	for _, msg := range file.Messages() {
		if GetGenerateCudTest(msg.DescriptorProto) ||
			GetGenerateShowTest(msg.DescriptorProto) {
			hasGenerateCudTest = true
			break
		}
	}
	if !hasGenerateCudTest && !hasCudMethod {
		return
	}
	t.P(gensupport.AutoGenComment)
	for _, msg := range file.Messages() {
		if GetGenerateCudTest(msg.DescriptorProto) ||
			GetGenerateShowTest(msg.DescriptorProto) {
			t.generateCudTest(msg.DescriptorProto)
		}
	}
	if len(file.FileDescriptorProto.Service) != 0 {
		for _, service := range file.FileDescriptorProto.Service {
			if len(service.Method) == 0 {
				continue
			}
			for _, method := range service.Method {
				t.genDummyMethod(*service.Name, method)
			}
		}
	}
	if t.firstFile {
		t.genDummyServer()
		t.firstFile = false
	}
}

func (t *TestCud) generateCudTest(message *descriptor.DescriptorProto) {
	keystr, err := t.support.GetMessageKeyName(t.Generator, message)
	if err != nil {
		keystr = "key not found"
	}
	args := tmplArgs{
		Pkg:       edgeproto,
		Name:      *message.Name,
		KeyName:   keystr,
		ShowOnly:  GetGenerateShowTest(message),
		Streamout: GetGenerateCudStreamout(message),
	}
	cudFuncs := make([]cudFunc, 0)
	for _, str := range []string{"Create", "Update", "Delete"} {
		cf := cudFunc{
			Func:      str,
			Pkg:       args.Pkg,
			Name:      args.Name,
			KeyName:   args.KeyName,
			Streamout: args.Streamout,
		}
		cudFuncs = append(cudFuncs, cf)
	}
	args.CudFuncs = cudFuncs

	for _, field := range message.Field {
		if GetTestUpdate(field) {
			args.UpdateField = generator.CamelCase(*field.Name)
			switch *field.Type {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				args.UpdateValue = "\"update just this\""
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				args.UpdateValue = "[]byte{1,2,3,4}"
			default:
				args.UpdateValue = "1101"
			}
		}
	}
	t.cudTmpl.Execute(t, args)
	t.importGrpc = true
	t.importEdgeproto = true
	t.importIO = true
	t.importTesting = true
	t.importContext = true
	t.importTime = true
	t.importAssert = true
}

type methodArgs struct {
	Service   string
	Method    string
	InName    string
	OutName   string
	Outstream bool
}

var methodTmpl = `
{{- if .Outstream}}
func (s *DummyServer) {{.Method}}(in *edgeproto.{{.InName}}, server edgeproto.{{.Service}}_{{.Method}}Server) error {
	server.Send(&edgeproto.{{.OutName}}{})
	server.Send(&edgeproto.{{.OutName}}{})
	server.Send(&edgeproto.{{.OutName}}{})
	return nil
}
{{- else}}
func (s *DummyServer) {{.Method}}(ctx context.Context, in *edgeproto.{{.InName}}) (*edgeproto.{{.OutName}}, error) {
	return &edgeproto.{{.OutName}}{}, nil
}
{{- end}}

`

func (t *TestCud) genDummyMethod(service string, method *descriptor.MethodDescriptorProto) {
	in := gensupport.GetDesc(t.Generator, method.GetInputType())
	out := gensupport.GetDesc(t.Generator, method.GetOutputType())
	if !GetGenerateCud(in.DescriptorProto) {
		return
	}
	args := methodArgs{
		Service:   service,
		Method:    *method.Name,
		InName:    *in.DescriptorProto.Name,
		OutName:   *out.DescriptorProto.Name,
		Outstream: gensupport.ServerStreaming(method),
	}
	err := t.methodTmpl.Execute(t, &args)
	if err != nil {
		t.Fail("Failed to execute method template: ", err.Error())
	}
	t.importEdgeproto = true
	if !args.Outstream {
		t.importContext = true
	}
}

func (t *TestCud) genDummyServer() {
	t.P("type DummyServer struct {}")
	t.P()
	t.P("func RegisterDummyServer(server *grpc.Server) {")
	t.P("d := &DummyServer{}")

	for _, file := range t.Generator.Request.ProtoFile {
		if len(file.Service) == 0 {
			continue
		}
		for _, service := range file.Service {
			if len(service.Method) == 0 {
				continue
			}
			hasCudMethod := false
			for _, method := range service.Method {
				in := gensupport.GetDesc(t.Generator,
					method.GetInputType())
				if GetGenerateCud(in.DescriptorProto) {
					hasCudMethod = true
					break
				}
			}
			if hasCudMethod {
				t.P("edgeproto.Register", service.Name,
					"Server(server, d)")
			}
		}
	}
	t.P("}")
	t.P()
	t.importGrpc = true
}

func GetGenerateCud(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateCud, false)
}

func GetGenerateCudTest(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateCudTest, false)
}

func GetGenerateCudStreamout(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateCudStreamout, false)
}

func GetGenerateShowTest(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateShowTest, false)
}

func GetTestUpdate(field *descriptor.FieldDescriptorProto) bool {
	return proto.GetBoolExtension(field.Options, protogen.E_TestUpdate, false)
}
