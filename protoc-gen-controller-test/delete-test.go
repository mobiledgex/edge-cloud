// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"sort"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/edgexr/edge-cloud/gensupport"
	"github.com/edgexr/edge-cloud/util"
)

type deleteArgs struct {
	Type               string
	KeyType            string
	ApiObj             string
	DeletePrepareField string
	Streamout          bool
	RefBys             []refByArgs
	TrackedBys         []trackedByArgs
}

type refByArgs struct {
	Type     string
	ObjField string
	ApiObj   string
}

type trackedByArgs struct {
	Type    string
	RefName string
	ApiObj  string
}

var dataGenTmpl = `
{{- define "dataGen"}}
// Caller must write by hand the test data generator.
// Each Ref object should only have a single reference to the key,
// in order to properly test each reference (i.e. don't have a single
// object that has multiple references).
type {{.Type}}DeleteDataGen interface {
	Get{{.Type}}TestObj() (*edgeproto.{{.Type}}, *testSupportData)
{{- range .RefBys}}
	Get{{.ObjField}}Ref(key *edgeproto.{{$.KeyType}}) (*edgeproto.{{.Type}}, *testSupportData)
{{- end}}
{{- range .TrackedBys}}
	Get{{.RefName}}Ref(key *edgeproto.{{$.KeyType}}) (*edgeproto.{{.Type}}, *testSupportData)
{{- end}}
}
{{- end}}
`

var runDeleteTmpl = `
{{- define "runDelete"}}
	testObj, _ = dataGen.Get{{.Type}}TestObj()
{{- if .Streamout}}
	err = api.Delete{{.Type}}(testObj, testutil.NewCudStreamout{{.Type}}(ctx))
{{- else}}
	_, err = api.Delete{{.Type}}(ctx, testObj)
{{- end}}
{{- end}}
`

var deleteTmpl = `
{{- template "dataGen" .}}

// {{.Type}}DeleteStore wraps around the usual
// store to instrument checks and inject data while
// the delete api code is running.
type {{.Type}}DeleteStore struct {
	edgeproto.{{.Type}}Store
	t *testing.T
	allApis *AllApis
	putDeletePrepare bool
	putDeletePrepareCb func()
	putDeletePrepareSTM concurrency.STM
}

func (s *{{.Type}}DeleteStore) Put(ctx context.Context, m *edgeproto.{{.Type}}, wait func(int64), ops ...objstore.KVOp) (*edgeproto.Result, error) {
	if wait != nil {
		s.putDeletePrepare = m.{{.DeletePrepareField}}
	}
	res, err := s.{{.Type}}Store.Put(ctx, m, wait, ops...)
	if s.putDeletePrepare && s.putDeletePrepareCb != nil {
		s.putDeletePrepareCb()
	}
	return res, err
}

func (s *{{.Type}}DeleteStore) STMPut(stm concurrency.STM, obj *edgeproto.{{.Type}}, ops ...objstore.KVOp) {
	// there's an assumption that this is run within an ApplySTMWait,
	// where we wait for the caches to be updated with the transaction.
	if obj.{{.DeletePrepareField}} {
		s.putDeletePrepare = true
		s.putDeletePrepareSTM = stm
	} else {
		s.putDeletePrepare = false
		s.putDeletePrepareSTM = nil
	}
	s.{{.Type}}Store.STMPut(stm, obj, ops...)
	if s.putDeletePrepare && s.putDeletePrepareCb != nil {
		s.putDeletePrepareCb()
	}
}

func (s *{{.Type}}DeleteStore) Delete(ctx context.Context, m *edgeproto.{{.Type}}, wait func(int64)) (*edgeproto.Result, error) {
	require.True(s.t, s.putDeletePrepare, "{{.DeletePrepareField}} must be comitted to database with a sync.Wait before deleting")
	return s.{{.Type}}Store.Delete(ctx, m, wait)
}

func (s *{{.Type}}DeleteStore) STMDel(stm concurrency.STM, key *edgeproto.{{.KeyType}}){
	require.True(s.t, s.putDeletePrepare, "{{.DeletePrepareField}} must be comitted to database with a sync.Wait before deleting")
	s.{{.Type}}Store.STMDel(stm, key)
}

func (s *{{.Type}}DeleteStore) requireUndoDeletePrepare(ctx context.Context, obj *edgeproto.{{.Type}}) {
	deletePrepare := s.getDeletePrepare(ctx, obj)
	require.False(s.t, deletePrepare, "must undo delete prepare field on failure")
}

func (s *{{.Type}}DeleteStore) getDeletePrepare(ctx context.Context, obj *edgeproto.{{.Type}}) bool {
	buf := edgeproto.{{.Type}}{}
	found := s.Get(ctx, obj.GetKey(), &buf)
	require.True(s.t, found, "expected test object to be found")
	return buf.{{.DeletePrepareField}}
}

func delete{{.Type}}Checks(t *testing.T, ctx context.Context, all *AllApis, dataGen {{.Type}}DeleteDataGen) {
	var err error
	// override store so we can inject data and check data
	api := all.{{.ApiObj}}
	origStore := api.store
	deleteStore := &{{.Type}}DeleteStore{
		{{.Type}}Store: origStore,
		t: t,
		allApis: all,
	}
	api.store = deleteStore
{{- range .TrackedBys}}
	{{.ApiObj}}Store, {{.ApiObj}}Unwrap := wrap{{.Type}}TrackerStore(all.{{.ApiObj}})
{{- end}}
	defer func() {
		api.store = origStore
{{- range .TrackedBys}}
		{{.ApiObj}}Unwrap()
{{- end}}
	}()

	// inject testObj directly, bypassing create checks/deps
	testObj, supportData := dataGen.Get{{.Type}}TestObj()
	supportData.put(t, ctx, all)
	defer supportData.delete(t, ctx, all)
	origStore.Put(ctx, testObj, api.sync.syncWait)

	// Positive test, delete should succeed without any references.
	// The overrided store checks that delete prepare was set on the
	// object in the database before actually doing the delete.
{{- if .TrackedBys}}
	// This call back checks that any refs lookups are done in the
	// same stm as the delete prepare is set.
	deleteStore.putDeletePrepareCb = func() {
		// make sure ref objects reads happen in same stm
		// as delete prepare is set
		require.NotNil(t, deleteStore.putDeletePrepareSTM, "must set delete prepare in STM")
{{- range .TrackedBys}}
		require.NotNil(t, {{.ApiObj}}Store.getSTM, "must check for refs from {{.Type}} in STM")
		require.Equal(t, deleteStore.putDeletePrepareSTM, {{.ApiObj}}Store.getSTM, "delete prepare and ref check for {{.Type}} must be done in the same STM")
{{- end}}
	}
{{- end}}
{{- template "runDelete" .}}
	require.Nil(t, err, "delete must succeed with no refs")
{{- if .TrackedBys}}
	deleteStore.putDeletePrepareCb = nil
{{- end}}

	// Negative test, inject testObj with delete prepare already set.
	testObj, _ = dataGen.Get{{.Type}}TestObj()
	testObj.{{.DeletePrepareField}} = true
	origStore.Put(ctx, testObj, api.sync.syncWait)
	// delete should fail with already being deleted
{{- template "runDelete" .}}
	require.NotNil(t, err, "delete must fail if already being deleted")
	require.Equal(t, testObj.GetKey().BeingDeletedError().Error(), err.Error())
	// failed delete must not interfere with existing delete prepare state
	require.True(t, deleteStore.getDeletePrepare(ctx, testObj), "delete prepare must not be modified by failed delete")

	// inject testObj for ref tests
	testObj, _ = dataGen.Get{{.Type}}TestObj()
	origStore.Put(ctx, testObj, api.sync.syncWait)
{{range .RefBys}}
	{
	// Negative test, {{.Type}} refers to {{$.Type}}.
	// The cb will inject refBy obj after delete prepare has been set.
	refBy, supportData := dataGen.Get{{.ObjField}}Ref(testObj.GetKey())
	supportData.put(t, ctx, all)
	deleteStore.putDeletePrepareCb = func() {
		all.{{.ApiObj}}.store.Put(ctx, refBy, all.{{.ApiObj}}.sync.syncWait)
	}
{{- template "runDelete" $}}
	require.NotNil(t, err, "must fail delete with ref from {{.Type}}")
	require.Contains(t, err.Error(), "in use")
	// check that delete prepare was reset
	deleteStore.requireUndoDeletePrepare(ctx, testObj)
	// remove {{.Type}} obj
	_, err = all.{{.ApiObj}}.store.Delete(ctx, refBy, all.{{.ApiObj}}.sync.syncWait)
	require.Nil(t, err, "cleanup ref from {{.Type}} must succeed")
	deleteStore.putDeletePrepareCb = nil
	supportData.delete(t, ctx, all)
	}
{{- end}}

{{- range .TrackedBys}}
	{
	// Negative test, {{.Type}} refers to {{$.Type}} via refs object.
	// Inject the refs object to trigger an "in use" error.
	refBy, supportData := dataGen.Get{{.RefName}}Ref(testObj.GetKey())
	supportData.put(t, ctx, all)
	_, err = all.{{.ApiObj}}.store.Put(ctx, refBy, all.{{.ApiObj}}.sync.syncWait)
	require.Nil(t, err)
{{- template "runDelete" $}}
	require.NotNil(t, err, "delete with ref from {{.Type}} must fail")
	require.Contains(t, err.Error(), "in use")
	// check that delete prepare was reset
	deleteStore.requireUndoDeletePrepare(ctx, testObj)
	// remove {{.Type}} obj
	_, err = all.{{.ApiObj}}.store.Delete(ctx, refBy, all.{{.ApiObj}}.sync.syncWait)
	require.Nil(t, err, "cleanup ref from {{.Type}} must succeed")
	supportData.delete(t, ctx, all)
	}
{{- end}}

	// clean up testObj
{{- template "runDelete" .}}
	require.Nil(t, err, "cleanup must succeed")
}

`

func (s *ControllerTest) getDeleteRefTos(desc *generator.Descriptor) *gensupport.RefToGroup {
	refToGroup, ok := s.refData.RefTos[*desc.Name]
	if !ok {
		return nil
	}
	if !refToGroup.To.GenerateCud {
		return nil
	}
	return refToGroup
}

func (s *ControllerTest) generateDeleteTest(desc *generator.Descriptor) {
	refToGroup := s.getDeleteRefTos(desc)
	if refToGroup == nil {
		return
	}
	message := desc.DescriptorProto
	args := deleteArgs{
		Type:               refToGroup.To.Type,
		KeyType:            refToGroup.To.KeyType,
		DeletePrepareField: gensupport.GetDeletePrepareField(s.Generator, desc),
		Streamout:          GetGenerateCudStreamout(message),
		ApiObj:             getApiObj(desc),
	}
	// for tracked refs, we inject a refs object into the db
	tracked := map[string]struct{}{}
	for _, tracker := range s.refData.Trackers {
		if tracker.To.Type != refToGroup.To.Type {
			continue
		}
		for _, byObjField := range tracker.Bys {
			byArgs := trackedByArgs{}
			byArgs.Type = tracker.Type
			byArgs.RefName = tracker.To.Type + byObjField.By.Type + byObjField.Field.HierName
			byArgs.ApiObj = getApiObj(tracker.TypeDesc)
			tracked[byObjField.By.Type] = struct{}{}
			args.TrackedBys = append(args.TrackedBys, byArgs)
		}
	}
	sort.Slice(args.TrackedBys, func(i, j int) bool {
		return args.TrackedBys[i].Type < args.TrackedBys[j].Type
	})
	// for untracked refs, we inject an object into the db
	for _, byObjField := range refToGroup.Bys {
		if _, ok := tracked[byObjField.By.Type]; ok {
			continue
		}
		if !byObjField.By.GenerateCud {
			continue
		}
		byArgs := refByArgs{}
		byArgs.Type = byObjField.By.Type
		byArgs.ObjField = byObjField.By.Type + strings.Replace(byObjField.Field.HierName, ".", "", -1)
		byArgs.ApiObj = getApiObj(byObjField.By.TypeDesc)
		args.RefBys = append(args.RefBys, byArgs)
	}
	sort.Slice(args.RefBys, func(i, j int) bool {
		return args.RefBys[i].Type < args.RefBys[j].Type
	})

	s.deleteTmpl.Execute(s.Generator, args)
	s.importContext = true
	s.importTesting = true
	s.importEdgeproto = true
	s.importObjstore = true
	s.importRequire = true
	s.importConcurrency = true
	if args.Streamout {
		s.importTestutil = true
	}
}

func getApiObj(desc *generator.Descriptor) string {
	apiObj := GetControllerApiStruct(desc.DescriptorProto)
	if apiObj == "" {
		apiObj = util.UncapitalizeMessage(*desc.Name) + "Api"
	}
	return apiObj
}

func getApiObjDef(desc *generator.Descriptor) string {
	return *desc.Name + "Api"
}

func getApiObjService(service *descriptor.ServiceDescriptorProto) string {
	return util.UncapitalizeMessage(*service.Name)
}

func (s *ControllerTest) generateAllDeleteTest() {
	names := []string{}
	for name := range s.refData.RefTos {
		names = append(names, name)
	}
	sort.Strings(names)

	s.P()
	s.P("type AllDeleteDataGen interface {")
	for _, name := range names {
		s.P(name, "DeleteDataGen")
	}
	s.P("}")

	s.P()
	s.P("func allDeleteChecks(t *testing.T, ctx context.Context, all *AllApis, dataGen AllDeleteDataGen) {")

	for _, name := range names {
		s.P("delete", name, "Checks(t, ctx, all, dataGen)")
	}
	s.P("}")
}

type trackerArgs struct {
	Type      string
	KeyType   string
	ApiObjDef string
}

var trackerTmpl = `
// {{.Type}}StoreTracker wraps around the usual
// store to track the STM used for gets/puts.
type {{.Type}}StoreTracker struct {
	edgeproto.{{.Type}}Store
	getSTM concurrency.STM
	putSTM concurrency.STM
}

// Wrap the Api's store with a tracker store.
// Returns the tracker store, and the unwrap function to defer.
func wrap{{.Type}}TrackerStore(api *{{.ApiObjDef}}) (*{{.Type}}StoreTracker, func()) {
	orig := api.store
	tracker := &{{.Type}}StoreTracker{
		{{.Type}}Store: api.store,
	}
	api.store = tracker
	unwrap := func() {
		api.store = orig
	}
	return tracker, unwrap
}

func (s *{{.Type}}StoreTracker) STMGet(stm concurrency.STM, key *edgeproto.{{.KeyType}}, buf *edgeproto.{{.Type}}) bool {
	found := s.{{.Type}}Store.STMGet(stm, key, buf)
	if s.getSTM == nil {
		s.getSTM = stm
	}
	return found
}

func (s *{{.Type}}StoreTracker) STMPut(stm concurrency.STM, obj *edgeproto.{{.Type}}, ops ...objstore.KVOp) {
	s.{{.Type}}Store.STMPut(stm, obj, ops...)
	if s.putSTM == nil {
		s.putSTM = stm
	}
}

`

func (s *ControllerTest) generateStoreTracker(desc *generator.Descriptor) {
	name := *desc.Name
	if s.refData.RefBys[name] == nil && s.refData.RefTos[name] == nil && s.refData.Trackers[name] == nil {
		return
	}
	if !GetGenerateCud(desc.DescriptorProto) {
		return
	}
	keyDesc := gensupport.GetDescKey(s.Generator, desc)
	args := trackerArgs{
		Type:      name,
		KeyType:   *keyDesc.Name,
		ApiObjDef: getApiObjDef(desc),
	}
	s.trackerTmpl.Execute(s.Generator, args)
	s.importEdgeproto = true
	s.importConcurrency = true
	s.importObjstore = true
}
