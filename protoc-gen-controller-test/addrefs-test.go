package main

import (
	"sort"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/mobiledgex/edge-cloud/gensupport"
)

type addRefsApiArgs struct {
	Type      string
	StoreType string
	ApiObj    string
	Api       string
	Streamout bool
	Create    bool
	RefTos    []refToArgs
}

type refToArgs struct {
	Type               string
	ApiObj             string
	ObjField           string
	DeletePrepareField string
}

var runApiTmpl = `
{{- define "runApi"}}
	testObj, _ = dataGen.Get{{.Api}}TestObj()
{{- if .Streamout}}
	err = all.{{.ApiObj}}.{{.Api}}(testObj, testutil.NewCudStreamout{{.Type}}(ctx))
{{- else}}
	_, err = all.{{.ApiObj}}.{{.Api}}(ctx, testObj)
{{- end}}
{{- end}}
`

var addRefsApiTmpl = `
func {{.Api}}AddRefsChecks(t *testing.T, ctx context.Context, all *AllApis, dataGen AllAddRefsDataGen) {
	var err error

	testObj, supportData := dataGen.Get{{.Api}}TestObj()
	supportData.put(t, ctx, all)

{{- range .RefTos}}
	{
	// set delete_prepare on referenced {{.Type}}
	ref := supportData.getOne{{.Type}}()
	require.NotNil(t, ref, "support data must include one referenced {{.Type}}")
	ref.{{.DeletePrepareField}} = true
	_, err = all.{{.ApiObj}}.store.Put(ctx, ref, all.{{.ApiObj}}.sync.syncWait)
	require.Nil(t, err)
	// api call must fail with object being deleted
{{- template "runApi" $}}
	require.NotNil(t, err, "{{$.Api}} must fail with {{.Type}}.{{.DeletePrepareField}} set")
	require.Equal(t, ref.GetKey().BeingDeletedError().Error(), err.Error())
	// reset delete_prepare on referenced {{.Type}}
	ref.{{.DeletePrepareField}} = false
	_, err = all.{{.ApiObj}}.store.Put(ctx, ref, all.{{.ApiObj}}.sync.syncWait)
	require.Nil(t, err)
	}
{{- end}}

	// wrap the stores so we can make sure all checks and changes
	// happen in the same STM.
	{{.ApiObj}}Store, {{.ApiObj}}Unwrap := wrap{{.StoreType}}TrackerStore(all.{{.ApiObj}})
	defer {{.ApiObj}}Unwrap()
{{- range .RefTos}}
	{{.ApiObj}}Store, {{.ApiObj}}Unwrap := wrap{{.Type}}TrackerStore(all.{{.ApiObj}})
	defer {{.ApiObj}}Unwrap()
{{- end}}

	// {{.Api}} should succeed if no references are in delete_prepare
{{- template "runApi" $}}
	require.Nil(t, err, "{{.Api}} should succeed if no references are in delete prepare")
	// make sure everything ran in the same STM
	require.NotNil(t, {{.ApiObj}}Store.putSTM, "{{.Api}} put {{.StoreType}} must be done in STM")
{{- range .RefTos}}
	require.NotNil(t, {{.ApiObj}}Store.getSTM, "{{$.Api}} check {{.Type}} ref must be done in STM")
	require.Equal(t, {{$.ApiObj}}Store.putSTM, {{.ApiObj}}Store.getSTM, "{{$.Api}} check {{.Type}} ref must be done in same STM as {{$.StoreType}} put")
{{- end}}

	// clean up
{{- if .Create}}
	// delete created test obj
	testObj, _ = dataGen.Get{{.Api}}TestObj()
{{- if .Streamout}}
	err = all.{{.ApiObj}}.Delete{{.Type}}(testObj, testutil.NewCudStreamout{{.Type}}(ctx))
{{- else}}
	_, err = all.{{.ApiObj}}.Delete{{.Type}}(ctx, testObj)
{{- end}}
	require.Nil(t, err)
{{- end}}
	supportData.delete(t, ctx, all)
}

`

func (s *ControllerTest) getAddRefsApiArgs(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) *addRefsApiArgs {
	prefix := gensupport.GetCamelCasePrefix(*method.Name)
	if prefix != gensupport.PrefixCreate && prefix != gensupport.PrefixUpdate && prefix != gensupport.PrefixAdd {
		return nil
	}
	in := gensupport.GetDesc(s.Generator, method.GetInputType())
	refByGroup := s.refData.RefBys[*in.Name]
	if refByGroup == nil {
		return nil
	}
	if in.Field == nil || len(in.Field) == 0 {
		// no fields with refers_to
		return nil
	}
	storeDesc := in
	firstRefIsStore := false
	if prefix == gensupport.PrefixAdd {
		// The input object may be the object to add to, or it
		// may be an intermediate object to just carry the key
		// for the object to add to, and the data to add.
		// In the second case, the first field should have a
		// refers_to that denotes the object the key field refers to.
		if gensupport.GetRefersTo(in.Field[0]) != "" {
			storeDesc = refByGroup.Tos[0].To.TypeDesc
			firstRefIsStore = true
		}
	}

	noconfig := map[string]struct{}{}
	if nc := GetMethodNoconfig(method); nc != "" {
		fields := strings.Split(nc, ",")
		for _, name := range fields {
			noconfig[name] = struct{}{}
		}
	}

	args := addRefsApiArgs{
		Type:      *in.Name,
		ApiObj:    getApiObjService(service),
		Api:       *method.Name,
		Streamout: GetGenerateCudStreamout(in.DescriptorProto),
		Create:    prefix == "Create",
		StoreType: *storeDesc.Name,
	}
	// If object refers to other objects, those look ups must check
	// the ref object's delete prepare
	for ii, byFieldTo := range refByGroup.Tos {
		if firstRefIsStore && ii == 0 {
			continue
		}
		if _, found := noconfig[byFieldTo.Field.HierName]; found {
			continue
		}
		if prefix == gensupport.PrefixUpdate && byFieldTo.Field.InKey {
			// can't update key values
			continue
		}
		toArgs := refToArgs{
			Type:               byFieldTo.To.Type,
			ApiObj:             getApiObj(byFieldTo.To.TypeDesc),
			ObjField:           refByGroup.By.Type + strings.Replace(byFieldTo.Field.HierName, ".", "", -1),
			DeletePrepareField: gensupport.GetDeletePrepareField(s.Generator, byFieldTo.To.TypeDesc),
		}
		args.RefTos = append(args.RefTos, toArgs)
	}
	if len(args.RefTos) == 0 {
		return nil
	}
	return &args
}

func (s *ControllerTest) generateAddRefsApiTest(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	args := s.getAddRefsApiArgs(service, method)
	if args == nil {
		return
	}
	s.addRefsApiTmpl.Execute(s.Generator, args)
	s.importContext = true
	s.importTesting = true
	s.importRequire = true
	if args.Streamout {
		s.importTestutil = true
	}
}

func (s *ControllerTest) generateAllAddRefsTest() {
	apis := []*addRefsApiArgs{}
	files := s.support.GetGeneratorFiles(s.Generator)
	for _, file := range files {
		for _, service := range file.Service {
			if len(service.Method) == 0 {
				continue
			}
			for _, method := range service.Method {
				args := s.getAddRefsApiArgs(service, method)
				if args == nil {
					continue
				}
				api := addRefsApiArgs{
					Type: args.Type,
					Api:  *method.Name,
				}
				apis = append(apis, &api)
			}
		}
	}
	sort.Slice(apis, func(i, j int) bool {
		return apis[i].Api < apis[j].Api
	})

	s.P()
	s.P("type AllAddRefsDataGen interface {")
	for _, api := range apis {
		s.P("Get" + api.Api + "TestObj() (*edgeproto." + api.Type + ", *testSupportData)")
	}
	s.P("}")

	s.P()
	s.P("func allAddRefsChecks(t *testing.T, ctx context.Context, all *AllApis, dataGen AllAddRefsDataGen) {")
	for _, api := range apis {
		s.P(api.Api + "AddRefsChecks(t, ctx, all, dataGen)")
	}
	s.P("}")
	s.importEdgeproto = true
	s.importTesting = true
	s.importContext = true
}
