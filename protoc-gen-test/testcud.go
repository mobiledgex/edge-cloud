// Modeled after gogo/protobuf/plugin/testgen/testgen.go testText plugin

package main

import (
	"strings"
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
	importRequire   bool
	importGrpc      bool
	importLog       bool
	importCli       bool
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
	HasUpdate   bool
	UpdateField string
	UpdateValue string
	ShowOnly    bool
	Streamout   bool
	Create      string
	Delete      string
	CudFuncs    []cudFunc
	ObjAndKey   bool
}

var tmpl = `
type Show{{.Name}} struct {
	Data map[string]{{.Pkg}}.{{.Name}}
	grpc.ServerStream
	Ctx context.Context
}

func (x *Show{{.Name}}) Init() {
	x.Data = make(map[string]{{.Pkg}}.{{.Name}})
}

func (x *Show{{.Name}}) Send(m *{{.Pkg}}.{{.Name}}) error {
	x.Data[m.GetKey().GetKeyString()] = *m
	return nil
}

func (x *Show{{.Name}}) Context() context.Context {
	return x.Ctx
}

var {{.Name}}ShowExtraCount = 0

{{- if .Streamout}}
type CudStreamout{{.Name}} struct {
	grpc.ServerStream
	Ctx context.Context
}

func (x *CudStreamout{{.Name}}) Send(res *{{.Pkg}}.Result) error {
	fmt.Println(res)
	return nil
}

func (x *CudStreamout{{.Name}}) Context() context.Context {
	return x.Ctx
}

func NewCudStreamout{{.Name}}(ctx context.Context) *CudStreamout{{.Name}} {
	return &CudStreamout{{.Name}}{
		Ctx: ctx,
	}
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
		x.Data[obj.GetKey().GetKeyString()] = *obj
	}
}

func (x *Show{{.Name}}) CheckFound(obj *{{.Pkg}}.{{.Name}}) bool {
	_, found := x.Data[obj.GetKey().GetKeyString()]
	return found
}

func (x *Show{{.Name}}) AssertFound(t *testing.T, obj *{{.Pkg}}.{{.Name}}) {
	check, found := x.Data[obj.GetKey().GetKeyString()]
	require.True(t, found, "find {{.Name}} %s", obj.GetKey().GetKeyString())
	if found && !check.Matches(obj, {{.Pkg}}.MatchIgnoreBackend(), {{.Pkg}}.MatchSortArrayedKeys()) {
		require.Equal(t, *obj, check, "{{.Name}} are equal")
	}
	if found {
		// remove in case there are dups in the list, so the
		// same object cannot be used again
		delete(x.Data, obj.GetKey().GetKeyString())
	}
}

func (x *Show{{.Name}}) AssertNotFound(t *testing.T, obj *{{.Pkg}}.{{.Name}}) {
	_, found := x.Data[obj.GetKey().GetKeyString()]
	require.False(t, found, "do not find {{.Name}} %s", obj.GetKey().GetKeyString())
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
		err := x.internal_api.{{.Func}}{{.Name}}(copy, NewCudStreamout{{.Name}}(ctx))
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
		showData.Ctx = ctx
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
	span := log.StartSpan(log.DebugLevelApi, "Internal{{.Name}}Test")
	defer span.Finish()
	ctx := log.ContextWithSpan(context.Background(), span)

	switch test {
{{- if not .ShowOnly}}
	case "cud":
		basic{{.Name}}CudTest(t, ctx, NewInternal{{.Name}}Api(api), testData)
{{- end}}
	case "show":
		basic{{.Name}}ShowTest(t, ctx, NewInternal{{.Name}}Api(api), testData)
	}
}

func Client{{.Name}}Test(t *testing.T, test string, api {{.Pkg}}.{{.Name}}ApiClient, testData []{{.Pkg}}.{{.Name}}) {
	span := log.StartSpan(log.DebugLevelApi, "Client{{.Name}}Test")
	defer span.Finish()
	ctx := log.ContextWithSpan(context.Background(), span)

	switch test {
{{- if not .ShowOnly}}
	case "cud":
		basic{{.Name}}CudTest(t, ctx, NewClient{{.Name}}Api(api), testData)
{{- end}}
	case "show":
		basic{{.Name}}ShowTest(t, ctx, NewClient{{.Name}}Api(api), testData)
	}
}

func basic{{.Name}}ShowTest(t *testing.T, ctx context.Context, api *{{.Name}}CommonApi, testData []{{.Pkg}}.{{.Name}}) {
	var err error

	show := Show{{.Name}}{}
	show.Init()
	filterNone := {{.Pkg}}.{{.Name}}{}
	err = api.Show{{.Name}}(ctx, &filterNone, &show)
	require.Nil(t, err, "show data")
	require.Equal(t, len(testData) + {{.Name}}ShowExtraCount, len(show.Data), "Show count")
	for _, obj := range testData {
		show.AssertFound(t, &obj)
	}
}

func Get{{.Name}}(t *testing.T, ctx context.Context, api *{{.Name}}CommonApi, key *{{.KeyName}}, out *{{.Pkg}}.{{.Name}}) bool {
	var err error

	show := Show{{.Name}}{}
	show.Init()
	filter := {{.Pkg}}.{{.Name}}{}
	filter.SetKey(key)
	err = api.Show{{.Name}}(ctx, &filter, &show)
	require.Nil(t, err, "show data")
	obj, found := show.Data[key.GetKeyString()]
	if found {
		*out = obj
	}
	return found
}

{{ if not .ShowOnly}}
func basic{{.Name}}CudTest(t *testing.T, ctx context.Context, api *{{.Name}}CommonApi, testData []{{.Pkg}}.{{.Name}}) {
	var err error

	if len(testData) < 3 {
		require.True(t, false, "Need at least 3 test data objects")
		return
	}

	// test create
	{{.Create}}{{.Name}}Data(t, ctx, api, testData)

	// test duplicate {{.Create}} - should fail
	_, err = api.{{.Create}}{{.Name}}(ctx, &testData[0])
	require.NotNil(t, err, "{{.Create}} duplicate {{.Name}}")

	// test show all items
	basic{{.Name}}ShowTest(t, ctx, api, testData)

	// test {{.Delete}}
	_, err = api.{{.Delete}}{{.Name}}(ctx, &testData[0])
	require.Nil(t, err, "{{.Delete}} {{.Name}} %s", testData[0].GetKey().GetKeyString())
	show := Show{{.Name}}{}
	show.Init()
	filterNone := {{.Pkg}}.{{.Name}}{}
	err = api.Show{{.Name}}(ctx, &filterNone, &show)
	require.Nil(t, err, "show data")
	require.Equal(t, len(testData) - 1 + {{.Name}}ShowExtraCount, len(show.Data), "Show count")
	show.AssertNotFound(t, &testData[0])
{{- if .HasUpdate}}
	// test update of missing object
	_, err = api.Update{{.Name}}(ctx, &testData[0])
	require.NotNil(t, err, "Update missing object")
{{- end}}
	// {{.Create}} it back
	_, err = api.{{.Create}}{{.Name}}(ctx, &testData[0])
	require.Nil(t, err, "{{.Create}} {{.Name}} %s", testData[0].GetKey().GetKeyString())

	// test invalid keys
	bad := {{.Pkg}}.{{.Name}}{}
	_, err = api.{{.Create}}{{.Name}}(ctx, &bad)
	require.NotNil(t, err, "{{.Create}} {{.Name}} with no key info")

{{if .UpdateField}}
	// test update
	updater := {{.Pkg}}.{{.Name}}{}
	updater.Key = testData[0].Key
	updater.{{.UpdateField}} = {{.UpdateValue}}
	updater.Fields = make([]string, 0)
	updater.Fields = append(updater.Fields, {{.Pkg}}.{{.Name}}Field{{.UpdateField}})
	_, err = api.Update{{.Name}}(ctx, &updater)
	require.Nil(t, err, "Update {{.Name}} %s", testData[0].GetKey().GetKeyString())

	show.Init()
	updater = testData[0]
	updater.{{.UpdateField}} = {{.UpdateValue}}
	err = api.Show{{.Name}}(ctx, &filterNone, &show)
	require.Nil(t, err, "show {{.Name}}")
	show.AssertFound(t, &updater)

	// revert change
	updater.{{.UpdateField}} = testData[0].{{.UpdateField}}
	_, err = api.Update{{.Name}}(ctx, &updater)
	require.Nil(t, err, "Update back {{.Name}}")
{{- end}}
}

func Internal{{.Name}}{{.Create}}(t *testing.T, api {{.Pkg}}.{{.Name}}ApiServer, testData []{{.Pkg}}.{{.Name}}) {
	span := log.StartSpan(log.DebugLevelApi, "Internal{{.Name}}{{.Create}}")
	defer span.Finish()
	ctx := log.ContextWithSpan(context.Background(), span)

	{{.Create}}{{.Name}}Data(t, ctx, NewInternal{{.Name}}Api(api), testData)
}

func Client{{.Name}}{{.Create}}(t *testing.T, api {{.Pkg}}.{{.Name}}ApiClient, testData []{{.Pkg}}.{{.Name}}) {
	span := log.StartSpan(log.DebugLevelApi, "Client{{.Name}}{{.Create}}")
	defer span.Finish()
	ctx := log.ContextWithSpan(context.Background(), span)

	{{.Create}}{{.Name}}Data(t, ctx, NewClient{{.Name}}Api(api), testData)
}

func {{.Create}}{{.Name}}Data(t *testing.T, ctx context.Context, api *{{.Name}}CommonApi, testData []{{.Pkg}}.{{.Name}}) {
	var err error

	for _, obj := range testData {
		_, err = api.{{.Create}}{{.Name}}(ctx, &obj)
		require.Nil(t, err, "{{.Create}} {{.Name}} %s", obj.GetKey().GetKeyString())
	}
}
{{- end}}

func Find{{.Name}}Data(key *{{.KeyName}}, testData []{{.Pkg}}.{{.Name}}) (*{{.Pkg}}.{{.Name}}, bool) {
	for ii, _ := range testData {
		if testData[ii].GetKey().Matches(key) {
			return &testData[ii], true
		}
	}
	return nil, false
}
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
	if t.importRequire {
		t.PrintImport("", "github.com/stretchr/testify/require")
	}
	if t.importLog {
		t.PrintImport("", "github.com/mobiledgex/edge-cloud/log")
	}
	if t.importCli {
		t.PrintImport("", "github.com/mobiledgex/edge-cloud/cli")
	}
}

func (t *TestCud) Generate(file *generator.FileDescriptor) {
	t.importGrpc = false
	t.importEdgeproto = false
	t.importIO = false
	t.importTesting = false
	t.importContext = false
	t.importTime = false
	t.importRequire = false
	t.importLog = false
	t.importCli = false
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
			t.generateCudTest(msg)
		}
	}
	if len(file.FileDescriptorProto.Service) != 0 {
		for _, service := range file.FileDescriptorProto.Service {
			if len(service.Method) == 0 {
				continue
			}
			t.generateRunApi(file.FileDescriptorProto, service)
			for _, method := range service.Method {
				t.genDummyMethod(*service.Name, method)
			}
		}
	}
	if t.firstFile {
		t.genDummyServer()
		t.firstFile = false
	}
	gensupport.RunParseCheck(t.Generator, file)
}

func (t *TestCud) generateCudTest(desc *generator.Descriptor) {
	message := desc.DescriptorProto
	keystr, err := t.support.GetMessageKeyType(t.Generator, desc)
	if err != nil {
		keystr = "key not found"
	}
	args := tmplArgs{
		Pkg:       edgeproto,
		Name:      *message.Name,
		KeyName:   keystr,
		ShowOnly:  GetGenerateShowTest(message),
		Streamout: GetGenerateCudStreamout(message),
		HasUpdate: GetGenerateCudTestUpdate(message),
		ObjAndKey: gensupport.GetObjAndKey(message),
	}
	fncs := []string{}
	if GetGenerateAddrmTest(message) {
		args.Create = "Add"
		args.Delete = "Remove"
		fncs = []string{"Add", "Remove"}
	} else {
		args.Create = "Create"
		args.Delete = "Delete"
		fncs = []string{"Create", "Delete"}
	}
	if args.HasUpdate {
		fncs = append(fncs, "Update")
	}
	cudFuncs := make([]cudFunc, 0)
	for _, str := range fncs {
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
	t.importRequire = true
	t.importLog = true
}

func (t *TestCud) generateRunApi(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	// group methods by input type
	groups := gensupport.GetMethodGroups(t.Generator, service, nil)
	for inType, group := range groups {
		t.generateRunGroupApi(file, service, inType, group)
	}
}

func (t *TestCud) generateRunGroupApi(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, inType string, group *gensupport.MethodGroup) {
	specialKeys := map[string]string{
		"CloudletPoolMember": "PoolKey",
	}
	apiName := *service.Name + group.Suffix

	t.P()
	t.P("func Run", apiName, "(conn *grpc.ClientConn, ctx context.Context, data *[]edgeproto.", inType, ", dataMap []map[string]interface{}, mode string) error {")
	t.P("var err error")
	objApiStr := strings.ToLower(string(inType[0])) + string(inType[1:len(inType)]) + "Api"
	objKey := "obj.Key"
	if newKey, ok := specialKeys[inType]; ok {
		objKey = "obj." + newKey
	}
	t.P(objApiStr, " := edgeproto.New", service.Name, "Client(conn)")
	if group.HasUpdate {
		t.P("for ii, obj := range *data {")
	} else {
		t.P("for _, obj := range *data {")
	}
	t.P("log.DebugLog(log.DebugLevelApi, \"API %v for ", inType, ": %v\", mode, ", objKey, ")")
	if group.HasStream {
		t.P("var stream ", inType, "Stream")
	}
	t.P("switch mode {")
	for _, mInfo := range group.MethodInfos {
		t.P("case \"", mInfo.Action, "\":")
		if mInfo.Action == "update" {
			t.importCli = true
			t.P("obj.Fields = cli.GetSpecifiedFields(dataMap[ii], &obj, cli.YamlNamespace)")
		}
		if mInfo.Stream {
			t.P("stream, err = ", objApiStr, ".", mInfo.Name, "(ctx, &obj)")
		} else {
			t.P("_, err = ", objApiStr, ".", mInfo.Name, "(ctx, &obj)")
		}
	}
	t.P("default:")
	t.P("log.DebugLog(log.DebugLevelApi, \"Unsupported API %v for ", inType, ": %v\", mode, ", objKey, ")")
	t.P("return nil")
	t.P("}")
	if group.HasStream {
		t.P("err = ", inType, "ReadResultStream(stream, err)")
	}
	t.P("err = ignoreExpectedErrors(mode, &", objKey, ", err)")
	t.P("if err != nil {")
	t.P("return fmt.Errorf(\"API %s failed for %v -- err %v\", mode, ", objKey, ", err)")
	t.P("}")
	t.P("}")
	t.P("return nil")
	t.P("}")
	t.importGrpc = true
}

type methodArgs struct {
	Service   string
	Method    string
	InName    string
	OutName   string
	Outstream bool
	OutList   bool
	HasCache  bool
	Show      bool
	CacheFunc string
}

var methodTmpl = `
{{- if .Outstream}}
func (s *DummyServer) {{.Method}}(in *edgeproto.{{.InName}}, server edgeproto.{{.Service}}_{{.Method}}Server) error {
	var err error
{{- if .CacheFunc}}
	s.{{.InName}}Cache.{{.CacheFunc}}(server.Context(), in, 0)
{{- end}}
{{- if (eq .InName .OutName)}}
	obj := &edgeproto.{{.OutName}}{}
	if obj.Matches(in, edgeproto.MatchFilter()) {
{{- else}}
	if true {
{{- end}}
		for ii := 0; ii < s.ShowDummyCount; ii++ {
			server.Send(&edgeproto.{{.OutName}}{})
		}
	}
{{- if and .OutList .HasCache}}
	err = s.{{.InName}}Cache.Show(in, func(obj *edgeproto.{{.InName}}) error {
		err := server.Send(obj)
		return err
	})
{{- end}}
	return err
}
{{- else}}
func (s *DummyServer) {{.Method}}(ctx context.Context, in *edgeproto.{{.InName}}) (*edgeproto.{{.OutName}}, error) {
	if s.CudNoop {
		return &edgeproto.{{.OutName}}{}, nil	
	}
{{- if .CacheFunc}}
	s.{{.InName}}Cache.{{.CacheFunc}}(ctx, in, 0)
{{- end}}
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
		OutList:   GetGenerateCud(out.DescriptorProto),
		HasCache:  GetGenerateCache(in.DescriptorProto),
		Show:      strings.HasPrefix(*method.Name, "Show"),
	}
	if args.HasCache {
		if strings.HasPrefix(*method.Name, "Create") {
			args.CacheFunc = "Update"
		} else if strings.HasPrefix(*method.Name, "Delete") {
			args.CacheFunc = "Delete"
		} else if strings.HasPrefix(*method.Name, "Update") {
			args.CacheFunc = "Update"
		}
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
	t.P("type DummyServer struct {")

	for _, file := range t.Generator.Request.ProtoFile {
		for _, desc := range file.MessageType {
			if !GetGenerateCud(desc) {
				continue
			}
			if !GetGenerateCache(desc) {
				continue
			}
			t.P(desc.Name, "Cache edgeproto.", desc.Name, "Cache")
		}
	}
	t.P("ShowDummyCount int")
	t.P("CudNoop bool")
	t.P("}")
	t.P()

	t.P("func RegisterDummyServer(server *grpc.Server) *DummyServer {")
	t.P("d := &DummyServer{}")

	for _, file := range t.Generator.Request.ProtoFile {
		for _, desc := range file.MessageType {
			if !GetGenerateCud(desc) {
				continue
			}
			if !GetGenerateCache(desc) {
				continue
			}
			t.P("edgeproto.Init", desc.Name, "Cache(&d.", desc.Name, "Cache)")
		}
	}
	for _, file := range t.Generator.Request.ProtoFile {
		if len(file.Service) == 0 {
			continue
		}
		for _, service := range file.Service {
			if len(service.Method) == 0 {
				continue
			}
			if !GetDummyServer(service) {
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
	t.P("return d")
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

func GetGenerateCudTestUpdate(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateCudTestUpdate, true)
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

func GetDummyServer(service *descriptor.ServiceDescriptorProto) bool {
	return proto.GetBoolExtension(service.Options, protogen.E_DummyServer, true)
}

func GetGenerateAddrmTest(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateAddrmTest, false)
}

func GetGenerateCache(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateCache, false)
}
