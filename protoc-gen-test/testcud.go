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

type TestCud struct {
	*generator.Generator
	support         gensupport.PluginSupport
	cudTmpl         *template.Template
	methodTmpl      *template.Template
	clientTmpl      *template.Template
	recvStreamFuncs map[string]struct{}
	methodGroups    map[string]*gensupport.MethodGroup
	firstFile       bool
	importProtoPkg  bool
	importIO        bool
	importTesting   bool
	importContext   bool
	importTime      bool
	importRequire   bool
	importGrpc      bool
	importLog       bool
	importCli       bool
	importWrapper   bool
}

func (t *TestCud) Name() string {
	return "TestCud"
}

func (t *TestCud) Init(g *generator.Generator) {
	t.Generator = g
	t.support.Init(g.Request)
	t.cudTmpl = template.Must(template.New("cud").Parse(tmpl))
	t.methodTmpl = template.Must(template.New("method").Parse(methodTmpl))
	t.clientTmpl = template.Must(template.New("client").Parse(clientTmpl))
	t.recvStreamFuncs = make(map[string]struct{})
	t.methodGroups = make(map[string]*gensupport.MethodGroup)
	t.firstFile = true
	for _, file := range t.Generator.Request.ProtoFile {
		if len(file.Service) == 0 {
			continue
		}
		for _, service := range file.Service {
			groups := gensupport.GetMethodGroups(g, service)
			for _, group := range groups {
				if _, found := t.methodGroups[group.InType]; found {
					continue
				}
				t.methodGroups[group.InType] = group
			}
		}
	}
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

func {{.Name}}ReadResultStream(stream ResultStream, err error) error {
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
	if t.importProtoPkg {
		t.PrintImport("", generator.GoImportPath(t.support.PackageImportPath))
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
	if t.importWrapper {
		t.PrintImport("", "github.com/mobiledgex/edge-cloud/edgectl/wrapper")
	}
}

func (t *TestCud) Generate(file *generator.FileDescriptor) {
	t.importGrpc = false
	t.importProtoPkg = false
	t.importIO = false
	t.importTesting = false
	t.importContext = false
	t.importTime = false
	t.importRequire = false
	t.importLog = false
	t.importCli = false
	t.importWrapper = false
	t.support.InitFile()
	if !t.support.GenFile(*file.FileDescriptorProto.Name) {
		return
	}
	hasMethod := false
	if len(file.FileDescriptorProto.Service) > 0 {
		for _, service := range file.FileDescriptorProto.Service {
			if hasSupportedMethod(service) {
				hasMethod = true
				break
			}
		}
	}
	hasE2edata := false
	for _, msg := range file.Messages() {
		if gensupport.GetE2edata(msg.DescriptorProto) {
			hasE2edata = true
			break
		}
	}
	if !hasMethod && !hasE2edata {
		return
	}

	t.P(gensupport.AutoGenComment)
	for _, msg := range file.Messages() {
		if GetGenerateCudTest(msg.DescriptorProto) ||
			GetGenerateShowTest(msg.DescriptorProto) {
			t.generateCudTest(msg)
		}
		if gensupport.GetE2edata(msg.DescriptorProto) {
			t.genE2edata(msg)
		}
	}
	for _, service := range file.FileDescriptorProto.Service {
		if len(service.Method) == 0 {
			continue
		}
		t.generateRunApi(file.FileDescriptorProto, service)
		for _, method := range service.Method {
			t.genDummyMethod(*service.Name, method)
		}
	}
	for _, service := range file.FileDescriptorProto.Service {
		if len(service.Method) == 0 {
			continue
		}
		t.genClientInterface(service)
	}
	if t.firstFile {
		t.genDummyServer()
		t.genClient()
		t.firstFile = false
	}
	gensupport.RunParseCheck(t.Generator, file)
}

func hasSupportedMethod(service *descriptor.ServiceDescriptorProto) bool {
	if len(service.Method) == 0 {
		return false
	}
	for _, method := range service.Method {
		if gensupport.ClientStreaming(method) {
			continue
		}
		return true
	}
	return false
}

func (t *TestCud) generateCudTest(desc *generator.Descriptor) {
	message := desc.DescriptorProto
	keystr, err := t.support.GetMessageKeyType(t.Generator, desc)
	if err != nil {
		keystr = "key not found"
	}
	args := tmplArgs{
		Pkg:       t.support.GetPackageName(desc),
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
	t.importProtoPkg = true
	t.importIO = true
	t.importTesting = true
	t.importContext = true
	t.importTime = true
	t.importRequire = true
	t.importLog = true
}

func (t *TestCud) generateRunApi(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	// group methods by input type
	groups := gensupport.GetMethodGroups(t.Generator, service)
	for _, group := range groups {
		t.generateRunGroupApi(file, service, group)
	}
}

func (t *TestCud) generateRunGroupApi(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, group *gensupport.MethodGroup) {
	apiName := group.ApiName()
	inType := group.InType
	inPkg := t.support.GetPackage(group.In)
	dataIn := "data *[]" + inPkg + inType
	if group.SingularData {
		dataIn = "obj *" + inPkg + inType
	}

	t.P()
	t.P("func (r *Run) ", apiName, "(", dataIn, ", dataMap interface{}, dataOut interface{}) {")
	t.P("log.DebugLog(log.DebugLevelApi, \"API for ", inType, "\", \"mode\", r.Mode)")
	t.importLog = true
	looped := false
	for _, mInfo := range group.MethodInfos {
		if !mInfo.IsShow {
			continue
		}
		t.P("if r.Mode == \"show\" {")
		if group.SingularData {
			t.P("obj = &", inPkg, inType, "{}")
		} else {
			t.P("obj := &", inPkg, inType, "{}")
		}
		t.runApiOutput(apiName, mInfo, group, looped)
		t.P("return")
		t.P("}")
		break
	}

	if group.SingularData {
		t.P("if obj == nil { return }")
	} else {
		t.P("for ii, objD := range *data {")
		t.P("obj := &objD")
		looped = true
	}
	t.P("switch r.Mode {")
	for _, mInfo := range group.MethodInfos {
		prefix := strings.ToLower(mInfo.Prefix)
		if mInfo.IsShow {
			prefix = "showfiltered"
		}
		t.P("case \"", prefix, "\":")
		if group.SingularData && mInfo.Prefix == "Update" {
			t.P("fallthrough")
			t.P("case \"create\":")
		}
		if group.SingularData && mInfo.Prefix == "Reset" {
			t.P("fallthrough")
			t.P("case \"delete\":")
		}
		if mInfo.IsUpdate {
			t.importCli = true
			t.P("// set specified fields")
			if group.SingularData {
				t.P("objMap, err := cli.GetGenericObj(dataMap)")
			} else {
				t.P("objMap, err := cli.GetGenericObjFromList(dataMap, ii)")
			}
			t.P("if err != nil {")
			t.P("log.DebugLog(log.DebugLevelApi, \"bad dataMap for ", inType, "\", \"err\", err)")
			t.P("*r.Rc = false")
			t.P("return")
			t.P("}")
			t.P("obj.Fields = cli.GetSpecifiedFields(objMap, obj, cli.YamlNamespace)")
			t.P()
		}
		t.runApiOutput(apiName, mInfo, group, looped)
	}
	t.P("}")
	if !group.SingularData {
		t.P("}")
	}
	t.P("}")
}

func (t *TestCud) runApiOutput(apiName string, info *gensupport.MethodInfo, group *gensupport.MethodGroup, looped bool) {
	hasKey := gensupport.GetMessageKey(group.In.DescriptorProto) != nil || gensupport.GetObjAndKey(group.In.DescriptorProto)
	t.P("out, err := r.client.", info.Name, "(r.ctx, obj)")
	t.P("if err != nil {")
	if !info.IsShow && hasKey {
		t.P("err = ignoreExpectedErrors(r.Mode, obj.GetKey(), err)")
	}
	desc := "\"" + apiName + "\""
	if looped {
		desc = "fmt.Sprintf(\"" + apiName + "[%d]\", ii)"
	}
	t.P("r.logErr(", desc, ", err)")
	t.P("} else {")
	outType := "*" + t.getOutputType(info, group)
	t.P("outp, ok := dataOut.(", outType, ")")
	t.P("if !ok {")
	t.P("panic(fmt.Sprintf(\"Run", apiName, " expected dataOut type ", outType, ", but was %T\", dataOut))")
	t.P("}")
	if group.SingularData {
		t.P("*outp = out")
	} else if info.IsShow {
		t.P("*outp = append(*outp, out...)")
	} else if info.Stream {
		t.P("*outp = append(*outp, out)")
	} else {
		t.P("*outp = append(*outp, *out)")
	}
	t.P("}")
}

func (t *TestCud) getOutputType(info *gensupport.MethodInfo, group *gensupport.MethodGroup) string {
	outPkg := t.support.GetPackage(info.Out)
	outType := outPkg + info.OutType
	if group.SingularData {
		outType = "*" + outType
	} else {
		outType = "[]" + outType
	}
	if info.Stream && !info.IsShow {
		outType = "[]" + outType
	}
	return outType
}

type methodArgs struct {
	Pkg            string
	Service        string
	Method         string
	InName         string
	OutName        string
	Outstream      bool
	OutList        bool
	HasCache       bool
	Show           bool
	CacheFunc      string
	OutData        string
	RecvStreamFunc bool
}

var methodTmpl = `
{{- if .Outstream}}
func (s *DummyServer) {{.Method}}(in *{{.Pkg}}.{{.InName}}, server {{.Pkg}}.{{.Service}}_{{.Method}}Server) error {
	var err error
{{- if .CacheFunc}}
	s.{{.InName}}Cache.{{.CacheFunc}}(server.Context(), in, 0)
{{- end}}
{{- if (eq .InName .OutName)}}
	obj := &{{.Pkg}}.{{.OutName}}{}
	if obj.Matches(in, {{.Pkg}}.MatchFilter()) {
{{- else}}
	if true {
{{- end}}
		for ii := 0; ii < s.ShowDummyCount; ii++ {
			server.Send(&{{.Pkg}}.{{.OutName}}{})
		}
	}
{{- if and .OutList .HasCache}}
	err = s.{{.InName}}Cache.Show(in, func(obj *{{.Pkg}}.{{.InName}}) error {
		err := server.Send(obj)
		return err
	})
{{- end}}
	return err
}
{{- else}}
func (s *DummyServer) {{.Method}}(ctx context.Context, in *{{.Pkg}}.{{.InName}}) (*{{.Pkg}}.{{.OutName}}, error) {
	if s.CudNoop {
		return &{{.Pkg}}.{{.OutName}}{}, nil	
	}
{{- if .CacheFunc}}
	s.{{.InName}}Cache.{{.CacheFunc}}(ctx, in, 0)
{{- end}}
	return &{{.Pkg}}.{{.OutName}}{}, nil	
}
{{- end}}

`

func (t *TestCud) getMethodArgs(service string, method *descriptor.MethodDescriptorProto) *methodArgs {
	in := gensupport.GetDesc(t.Generator, method.GetInputType())
	out := gensupport.GetDesc(t.Generator, method.GetOutputType())
	args := methodArgs{
		Pkg:       t.support.GetPackageName(in),
		Service:   service,
		Method:    *method.Name,
		InName:    *in.DescriptorProto.Name,
		OutName:   *out.DescriptorProto.Name,
		Outstream: gensupport.ServerStreaming(method),
		OutList:   GetGenerateCud(out.DescriptorProto),
		HasCache:  GetGenerateCache(in.DescriptorProto),
		Show:      gensupport.IsShow(method),
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
	args.OutData = "*" + args.Pkg + "." + args.OutName
	if args.Outstream {
		args.OutData = "[]" + args.Pkg + "." + args.OutName
	}
	return &args
}

func (t *TestCud) genDummyMethod(service string, method *descriptor.MethodDescriptorProto) {
	in := gensupport.GetDesc(t.Generator, method.GetInputType())
	if !GetGenerateCud(in.DescriptorProto) {
		return
	}
	args := t.getMethodArgs(service, method)
	err := t.methodTmpl.Execute(t, args)
	if err != nil {
		t.Fail("Failed to execute method template: ", err.Error())
	}
	t.importProtoPkg = true
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

var clientTmpl = `
{{- if .Outstream}}
{{- if .RecvStreamFunc}}
type {{.OutName}}Stream interface {
	Recv() (*{{.Pkg}}.{{.OutName}}, error)
}

func {{.OutName}}ReadStream(stream {{.OutName}}Stream) ([]{{.Pkg}}.{{.OutName}}, error) {
	output := []{{.Pkg}}.{{.OutName}}{}
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return output, fmt.Errorf("read {{.OutName}} stream failed, %v", err)
		}
		output = append(output, *obj)
	}
	return output, nil
}
{{- end}}

func (s *ApiClient) {{.Method}}(ctx context.Context, in *{{.Pkg}}.{{.InName}}) ([]{{.Pkg}}.{{.OutName}}, error) {
	api := {{.Pkg}}.New{{.Service}}Client(s.Conn)
	stream, err := api.{{.Method}}(ctx, in)
	if err != nil {
		return nil, err
	}
	return {{.OutName}}ReadStream(stream)
}

func (s *CliClient) {{.Method}}(ctx context.Context, in *{{.Pkg}}.{{.InName}}) ([]{{.Pkg}}.{{.OutName}}, error) {
	output := []{{.Pkg}}.{{.OutName}}{}
	args := append(s.BaseArgs, "controller", "{{.Method}}")
	err := wrapper.RunEdgectlObjs(args, in, &output, s.RunOps...)
	return output, err
}
{{- else}}
func (s *ApiClient) {{.Method}}(ctx context.Context, in *{{.Pkg}}.{{.InName}}) (*{{.Pkg}}.{{.OutName}}, error) {
	api := {{.Pkg}}.New{{.Service}}Client(s.Conn)
	return api.{{.Method}}(ctx, in)
}

func (s *CliClient) {{.Method}}(ctx context.Context, in *{{.Pkg}}.{{.InName}}) (*{{.Pkg}}.{{.OutName}}, error) {
	out := {{.Pkg}}.{{.OutName}}{}
	args := append(s.BaseArgs, "controller", "{{.Method}}")
	err := wrapper.RunEdgectlObjs(args, in, &out, s.RunOps...)
	return &out, err
}
{{- end}}

`

func (t *TestCud) genClientInterface(service *descriptor.ServiceDescriptorProto) {
	if len(service.Method) == 0 {
		return
	}
	methods := []*methodArgs{}

	for _, method := range service.Method {
		if gensupport.ClientStreaming(method) {
			continue
		}
		args := t.getMethodArgs(*service.Name, method)
		methods = append(methods, args)
		if args.Outstream {
			if _, found := t.recvStreamFuncs[args.OutName]; !found {
				args.RecvStreamFunc = true
				t.recvStreamFuncs[args.OutName] = struct{}{}
			}
		}
		err := t.clientTmpl.Execute(t, args)
		if err != nil {
			t.Fail("Failed to execute client template: ", err.Error())
		}
		t.importWrapper = true
		t.importContext = true
		t.importProtoPkg = true
		if args.RecvStreamFunc {
			t.importIO = true
		}
	}
	if len(methods) == 0 {
		return
	}

	t.P("type ", service.Name, "Client interface {")
	for _, args := range methods {
		t.P(args.Method, "(ctx context.Context, in *", args.Pkg, ".", args.InName, ") (", args.OutData, ", error)")
	}
	t.P("}")
	t.P()
}

func (t *TestCud) genClient() {
	t.P("type ApiClient struct {")
	t.P("Conn *grpc.ClientConn")
	t.P("}")
	t.P()
	t.P("type CliClient struct{")
	t.P("BaseArgs []string")
	t.P("RunOps []wrapper.RunOp")
	t.P("}")
	t.P()
	t.P("type Client interface {")
	for _, file := range t.Generator.Request.ProtoFile {
		if len(file.Service) == 0 {
			continue
		}
		for _, service := range file.Service {
			if hasSupportedMethod(service) {
				t.P(service.Name, "Client")
			}
		}
	}
	t.P("}")
	t.P()
	t.importGrpc = true
	t.importWrapper = true
}

type e2eFieldInfo struct {
	fieldName string
	ref       string
	group     *gensupport.MethodGroup
	info      *gensupport.MethodInfo
}

func (t *TestCud) genE2edata(desc *generator.Descriptor) {
	message := desc.DescriptorProto
	pkg := t.support.GetPackage(desc)
	t.importProtoPkg = true

	// Get groups per field. For output data struct, use first
	// method. Show data assumes input struct type is used to store
	// output show data.
	fieldInfos := []e2eFieldInfo{}
	showFieldInfos := []e2eFieldInfo{}
	for _, field := range message.Field {
		fieldDesc := gensupport.GetDesc(t.Generator, field.GetTypeName())
		inType := *fieldDesc.DescriptorProto.Name
		group, found := t.methodGroups[inType]
		if !found {
			continue
		}
		foundShow := false
		foundInfo := false
		for _, info := range group.MethodInfos {
			finfo := e2eFieldInfo{
				fieldName: generator.CamelCase(*field.Name),
				group:     group,
				info:      info,
				ref:       "&",
			}
			if group.SingularData {
				finfo.ref = ""
			}
			if info.IsShow && !foundShow {
				showFieldInfos = append(showFieldInfos, finfo)
				foundShow = true
			}
			if !info.IsShow && !foundInfo {
				fieldInfos = append(fieldInfos, finfo)
				foundInfo = true
			}
			if foundShow && foundInfo {
				break
			}
		}
	}

	outstruct := *message.Name + "Out"
	t.P("type ", outstruct, " struct {")
	for _, finfo := range fieldInfos {
		outType := t.getOutputType(finfo.info, finfo.group)
		t.P(finfo.fieldName, " ", outType)
	}
	t.P("Errors []Err")
	t.P("}")
	t.P()

	t.P("func Run", message.Name, "Apis(run *Run, in *", pkg, message.Name, ", inMap map[string]interface{}, out *", outstruct, ") {")
	for _, finfo := range fieldInfos {
		t.P("run.", finfo.group.ApiName(), "(", finfo.ref, "in.", finfo.fieldName, ", inMap[\"", strings.ToLower(finfo.fieldName), "\"], &out.", finfo.fieldName, ")")
	}
	t.P("out.Errors = run.Errs")
	t.P("}")
	t.P()

	t.P("func Run", message.Name, "ReverseApis(run *Run, in *", pkg, message.Name, ", inMap map[string]interface{}, out *", outstruct, ") {")
	for ii := len(fieldInfos) - 1; ii >= 0; ii-- {
		finfo := fieldInfos[ii]
		t.P("run.", finfo.group.ApiName(), "(", finfo.ref, "in.", finfo.fieldName, ", inMap[\"", strings.ToLower(finfo.fieldName), "\"], &out.", finfo.fieldName, ")")
	}
	t.P("out.Errors = run.Errs")
	t.P("}")
	t.P()

	t.P("func Run", message.Name, "ShowApis(run *Run, in *", pkg, message.Name, ", out *", pkg, message.Name, ") {")
	for _, finfo := range showFieldInfos {
		t.P("run.", finfo.group.ApiName(), "(", finfo.ref, "in.", finfo.fieldName, ", nil, &out.", finfo.fieldName, ")")
	}
	t.P("}")
	t.P()
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
