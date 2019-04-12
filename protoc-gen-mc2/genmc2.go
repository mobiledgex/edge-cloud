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

type GenMC2 struct {
	*generator.Generator
	support         gensupport.PluginSupport
	tmpl            *template.Template
	tmplMethodTest  *template.Template
	tmplMessageTest *template.Template
	regionStructs   map[string]struct{}
	firstFile       bool
	gentest         bool
	importEcho      bool
	importHttp      bool
	importContext   bool
	importIO        bool
	importJson      bool
	importTesting   bool
	importRequire   bool
	importOrmclient bool
}

func (g *GenMC2) Name() string {
	return "GenMC2"
}

func (g *GenMC2) Init(gen *generator.Generator) {
	g.Generator = gen
	g.tmpl = template.Must(template.New("mc2").Parse(tmpl))
	g.tmplMethodTest = template.Must(template.New("methodtest").Parse(tmplMethodTest))
	g.tmplMessageTest = template.Must(template.New("messagetest").Parse(tmplMessageTest))
	g.regionStructs = make(map[string]struct{})
	g.firstFile = true
}

func (g *GenMC2) GenerateImports(file *generator.FileDescriptor) {
	g.support.PrintUsedImports(g.Generator)
	if g.importEcho {
		g.PrintImport("", "github.com/labstack/echo")
	}
	if g.importHttp {
		g.PrintImport("", "net/http")
	}
	if g.importContext {
		g.PrintImport("", "context")
	}
	if g.importIO {
		g.PrintImport("", "io")
	}
	if g.importJson {
		g.PrintImport("", "encoding/json")
	}
	if g.importTesting {
		g.PrintImport("", "testing")
	}
	if g.importRequire {
		g.PrintImport("", "github.com/stretchr/testify/require")
	}
	if g.importOrmclient {
		g.PrintImport("", "github.com/mobiledgex/edge-cloud/mc/ormclient")
	}
}

func (g *GenMC2) Generate(file *generator.FileDescriptor) {
	g.importEcho = false
	g.importHttp = false
	g.importContext = false
	g.importIO = false
	g.importJson = false
	g.importTesting = false
	g.importRequire = false
	g.importOrmclient = false

	g.support.InitFile()
	if !g.support.GenFile(*file.FileDescriptorProto.Name) {
		return
	}
	if !genFile(file) {
		return
	}

	g.P(gensupport.AutoGenComment)
	if _, found := g.Generator.Param["gentest"]; found {
		// generate test code
		g.gentest = true
	}

	for _, service := range file.FileDescriptorProto.Service {
		g.generateService(service)
	}
	if g.firstFile {
		if !g.gentest {
			g.generatePosts()
		}
		g.firstFile = false
	}
	if g.gentest {
		for _, msg := range file.Messages() {
			if GetGenerateCud(msg.DescriptorProto) &&
				!GetGenerateShowTest(msg.DescriptorProto) {
				g.generateMessageTest(msg)
			}
		}
	}
}

func genFile(file *generator.FileDescriptor) bool {
	if len(file.FileDescriptorProto.Service) != 0 {
		for _, service := range file.FileDescriptorProto.Service {
			if len(service.Method) == 0 {
				continue
			}
			for _, method := range service.Method {
				if GetMc2Api(method) != "" {
					return true
				}
			}
		}
	}
	return false
}

func (g *GenMC2) generatePosts() {
	g.P("func addControllerApis(group *echo.Group) {")

	for _, file := range g.Generator.Request.ProtoFile {
		if !g.support.GenFile(*file.Name) {
			continue
		}
		if len(file.Service) == 0 {
			continue
		}
		for _, service := range file.Service {
			if len(service.Method) == 0 {
				continue
			}
			for _, method := range service.Method {
				if GetMc2Api(method) == "" {
					continue
				}
				g.P("group.POST(\"/ctrl/", method.Name,
					"\", ", method.Name, ")")
			}
		}
	}
	g.P("}")
	g.P()
}

func (g *GenMC2) generateService(service *descriptor.ServiceDescriptorProto) {
	if len(service.Method) == 0 {
		return
	}
	for _, method := range service.Method {
		g.generateMethod(*service.Name, method)
	}
}

func (g *GenMC2) generateMethod(service string, method *descriptor.MethodDescriptorProto) {
	api := GetMc2Api(method)
	if api == "" {
		return
	}
	apiVals := strings.Split(api, ",")
	if len(apiVals) != 3 {
		g.Fail("invalid mc2_api string, expected ResourceType,Action,OrgNameField")
	}
	in := gensupport.GetDesc(g.Generator, method.GetInputType())
	out := gensupport.GetDesc(g.Generator, method.GetOutputType())
	g.support.FQTypeName(g.Generator, in)
	inname := *in.DescriptorProto.Name
	_, found := g.regionStructs[inname]
	args := tmplArgs{
		Service:    service,
		MethodName: *method.Name,
		InName:     inname,
		OutName:    *out.DescriptorProto.Name,
		GenStruct:  !found,
		Resource:   apiVals[0],
		Action:     apiVals[1],
		OrgField:   apiVals[2],
		Org:        "obj." + apiVals[2],
		ShowOrg:    "res." + apiVals[2],
		Outstream:  gensupport.ServerStreaming(method),
	}
	if apiVals[2] == "" {
		args.Org = `""`
		args.ShowOrg = `""`
	}
	if apiVals[2] == "skipenforce" {
		args.SkipEnforce = true
	}
	if args.Action == "ActionView" && strings.HasPrefix(args.MethodName, "Show") {
		args.Show = true
	}
	if g.gentest {
		err := g.tmplMethodTest.Execute(g, &args)
		if err != nil {
			g.Fail("Failed to execute method test template: ", err.Error())
		}
		g.importOrmclient = true
	} else {
		err := g.tmpl.Execute(g, &args)
		if err != nil {
			g.Fail("Failed to execute method template: ", err.Error())
		}
		g.importEcho = true
		g.importHttp = true
		g.importContext = true
		if args.Outstream {
			g.importIO = true
			g.importJson = true
		}
	}
	if !found {
		g.regionStructs[inname] = struct{}{}
	}
}

type tmplArgs struct {
	Service     string
	MethodName  string
	InName      string
	OutName     string
	GenStruct   bool
	Resource    string
	Action      string
	OrgField    string
	Org         string
	ShowOrg     string
	Outstream   bool
	SkipEnforce bool
	Show        bool
}

var tmpl = `
{{- if .GenStruct}}
type Region{{.InName}} struct {
	Region string
	{{.InName}} edgeproto.{{.InName}}
}

{{- end}}
func {{.MethodName}}(c echo.Context) error {
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.claims = claims

	in := Region{{.InName}}{}
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	rc.region = in.Region
{{- if .Outstream}}
	// stream func may return "forbidden", so don't write
	// header until we know it's ok.
	wroteHeader := false
	err = {{.MethodName}}Stream(rc, &in.{{.InName}}, func(res *edgeproto.{{.OutName}}) {
		if !wroteHeader {
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c.Response().WriteHeader(http.StatusOK)
		}
		json.NewEncoder(c.Response()).Encode(res)
		c.Response().Flush()
	})
	if err != nil {
		if wroteHeader {
			json.NewEncoder(c.Response()).Encode(MsgErr(err))
			c.Response().Flush()
			return nil
		} else {
			return setReply(c, err, Msg("ok"))
		}
	}
	return nil
{{- else}}
	err = {{.MethodName}}Obj(rc, &in.{{.InName}})
	return setReply(c, err, Msg("ok"))
{{- end}}
}

{{if .Outstream}}
func {{.MethodName}}Stream(rc *RegionContext, obj *edgeproto.{{.InName}}, cb func(res *edgeproto.{{.OutName}})) error {
{{- else}}
func {{.MethodName}}Obj(rc *RegionContext, obj *edgeproto.{{.InName}}) error {
{{- end}}
{{- if and (not .Show) (not .SkipEnforce)}}
	if !enforcer.Enforce(rc.claims.Username, {{.Org}},
		{{.Resource}}, {{.Action}}) {
		return echo.ErrForbidden
	}
{{- end}}
	if rc.conn == nil {
		conn, err := connectController(rc.region)
		if err != nil {
			return err
		}
		rc.conn = conn
		defer func() {
			rc.conn.Close()
			rc.conn = nil
		}()
	}
	api := edgeproto.New{{.Service}}Client(rc.conn)
	ctx := context.Background()
{{- if .Outstream}}
	stream, err := api.{{.MethodName}}(ctx, obj)
	if err != nil {
		return err
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return err
		}
{{- if and (.Show) (not .SkipEnforce)}}
		if !enforcer.Enforce(rc.claims.Username, {{.ShowOrg}},
			{{.Resource}}, {{.Action}}) {
			continue
		}
{{- end}}
		cb(res)
	}
	return nil
{{- else}}
	_, err := api.{{.MethodName}}(ctx, obj)
	return err
{{- end}}
}

{{ if .Outstream}}
func {{.MethodName}}Obj(rc *RegionContext, obj *edgeproto.{{.InName}}) ([]edgeproto.{{.OutName}}, error) {
	arr := []edgeproto.{{.OutName}}{}
	err := {{.MethodName}}Stream(rc, obj, func(res *edgeproto.{{.OutName}}) {
		arr = append(arr, *res)
	})
	return arr, err
}
{{- end}}

`

var tmplMethodTest = `
{{- if .Outstream}}
func test{{.MethodName}}(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.{{.InName}}) (int, []edgeproto.{{.OutName}}, error) {
{{- else}}
func test{{.MethodName}}(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.{{.InName}}) (int, edgeproto.{{.OutName}}, error) {
{{- end}}
	dat := &Region{{.InName}}{}
	dat.Region = region
	dat.{{.InName}} = *in
	out := edgeproto.{{.OutName}}{}
{{- if .Outstream}}
	outlist := []edgeproto.{{.OutName}}{}
	status, err := mcClient.PostJsonStreamOut(uri+"/auth/ctrl/{{.MethodName}}", token, dat, &out, func() {
		outlist = append(outlist, out)
	})
	return status, outlist, err
{{- else}}
	status, err := mcClient.PostJson(uri+"/auth/ctrl/{{.MethodName}}", token, dat, &out)
	return status, out, err
{{- end}}
}

{{- if .Outstream}}
func testPerm{{.MethodName}}(mcClient *ormclient.Client, uri, token, region, org string) (int, []edgeproto.{{.OutName}}, error) {
{{- else}}
func testPerm{{.MethodName}}(mcClient *ormclient.Client, uri, token, region, org string) (int, edgeproto.{{.OutName}}, error) {
{{- end}}
	in := &edgeproto.{{.InName}}{}
{{- if and (ne .OrgField "") (not .SkipEnforce)}}
	in.{{.OrgField}} = org
{{- end}}
	return test{{.MethodName}}(mcClient, uri, token, region, in)
}
`

func (g *GenMC2) generateMessageTest(desc *generator.Descriptor) {
	message := desc.DescriptorProto
	args := msgArgs{
		Message: *message.Name,
	}
	err := g.tmplMessageTest.Execute(g, &args)
	if err != nil {
		g.Fail("Failed to execute message test template: ", err.Error())
	}
	g.importTesting = true
	g.importRequire = true
	g.importHttp = true
}

type msgArgs struct {
	Message string
}

var tmplMessageTest = `
// This tests the user cannot modify the object because the obj belongs to
// an organization that the user does not have permissions for.
func badPermTest{{.Message}}(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	status, _, err := testPermCreate{{.Message}}(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	status, _, err = testPermUpdate{{.Message}}(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	status, _, err = testPermDelete{{.Message}}(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func badPermTestShow{{.Message}}(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	// show is allowed but won't show anything
	status, list, err := testPermShow{{.Message}}(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(list))
}

// This tests the user can modify the object because the obj belongs to
// an organization that the user has permissions for.
func goodPermTest{{.Message}}(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, showcount int) {
	status, _, err := testPermCreate{{.Message}}(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	status, _, err = testPermUpdate{{.Message}}(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	status, _, err = testPermDelete{{.Message}}(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)

	// make sure region check works
	status, _, err = testPermCreate{{.Message}}(mcClient, uri, token, "bad region", org)
	require.Nil(t, err)
	require.Equal(t, http.StatusBadRequest, status)
	status, _, err = testPermUpdate{{.Message}}(mcClient, uri, token, "bad region", org)
	require.Nil(t, err)
	require.Equal(t, http.StatusBadRequest, status)
	status, _, err = testPermDelete{{.Message}}(mcClient, uri, token, "bad region", org)
	require.Nil(t, err)
	require.Equal(t, http.StatusBadRequest, status)

	goodPermTestShow{{.Message}}(t, mcClient, uri, token, region, org, showcount)
}

func goodPermTestShow{{.Message}}(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, count int) {
	status, list, err := testPermShow{{.Message}}(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, count, len(list))

	// make sure region check works
	status, list, err = testPermShow{{.Message}}(mcClient, uri, token, "bad region", org)
	require.Nil(t, err)
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, 0, len(list))
}

// Test permissions for user with token1 who should have permissions for
// modifying obj1, and user with token2 who should have permissions for obj2.
// They should not have permissions to modify each other's objects.
func permTest{{.Message}}(t *testing.T, mcClient *ormclient.Client, uri, token1, token2, region, org1, org2 string, showcount int) {
	badPermTest{{.Message}}(t, mcClient, uri, token1, region, org2)
	badPermTestShow{{.Message}}(t, mcClient, uri, token1, region, org2)
	badPermTest{{.Message}}(t, mcClient, uri, token2, region, org1)
	badPermTestShow{{.Message}}(t, mcClient, uri, token2, region, org1)

	goodPermTest{{.Message}}(t, mcClient, uri, token1, region, org1, showcount)
	goodPermTest{{.Message}}(t, mcClient, uri, token2, region, org2, showcount)
}
`

func GetMc2Api(method *descriptor.MethodDescriptorProto) string {
	return gensupport.GetStringExtension(method.Options, protogen.E_Mc2Api, "")
}

func GetGenerateCud(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateCud, false)
}

func GetGenerateShowTest(message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, protogen.E_GenerateShowTest, false)
}
