// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app_inst.proto

package testutil

import "google.golang.org/grpc"
import "github.com/mobiledgex/edge-cloud/edgeproto"
import "io"
import "testing"
import "context"
import "time"
import "github.com/stretchr/testify/assert"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/protocmd"
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

type ShowAppInst struct {
	Data map[string]edgeproto.AppInst
	grpc.ServerStream
}

func (x *ShowAppInst) Init() {
	x.Data = make(map[string]edgeproto.AppInst)
}

func (x *ShowAppInst) Send(m *edgeproto.AppInst) error {
	x.Data[m.Key.GetKeyString()] = *m
	return nil
}

type CudStreamoutAppInst struct {
	grpc.ServerStream
}

func (x *CudStreamoutAppInst) Send(res *edgeproto.Result) error {
	fmt.Println(res)
	return nil
}
func (x *CudStreamoutAppInst) Context() context.Context {
	return context.TODO()
}

type AppInstStream interface {
	Recv() (*edgeproto.Result, error)
}

func AppInstReadResultStream(stream AppInstStream, err error) error {
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

func (x *ShowAppInst) ReadStream(stream edgeproto.AppInstApi_ShowAppInstClient, err error) {
	x.Data = make(map[string]edgeproto.AppInst)
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

func (x *ShowAppInst) CheckFound(obj *edgeproto.AppInst) bool {
	_, found := x.Data[obj.Key.GetKeyString()]
	return found
}

func (x *ShowAppInst) AssertFound(t *testing.T, obj *edgeproto.AppInst) {
	check, found := x.Data[obj.Key.GetKeyString()]
	assert.True(t, found, "find AppInst %s", obj.Key.GetKeyString())
	if found && !check.Matches(obj, edgeproto.MatchIgnoreBackend(), edgeproto.MatchSortArrayedKeys()) {
		assert.Equal(t, *obj, check, "AppInst are equal")
	}
	if found {
		// remove in case there are dups in the list, so the
		// same object cannot be used again
		delete(x.Data, obj.Key.GetKeyString())
	}
}

func (x *ShowAppInst) AssertNotFound(t *testing.T, obj *edgeproto.AppInst) {
	_, found := x.Data[obj.Key.GetKeyString()]
	assert.False(t, found, "do not find AppInst %s", obj.Key.GetKeyString())
}

func WaitAssertFoundAppInst(t *testing.T, api edgeproto.AppInstApiClient, obj *edgeproto.AppInst, count int, retry time.Duration) {
	show := ShowAppInst{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowAppInst(ctx, obj)
		show.ReadStream(stream, err)
		cancel()
		if show.CheckFound(obj) {
			break
		}
		time.Sleep(retry)
	}
	show.AssertFound(t, obj)
}

func WaitAssertNotFoundAppInst(t *testing.T, api edgeproto.AppInstApiClient, obj *edgeproto.AppInst, count int, retry time.Duration) {
	show := ShowAppInst{}
	filterNone := edgeproto.AppInst{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowAppInst(ctx, &filterNone)
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
type AppInstCommonApi struct {
	internal_api edgeproto.AppInstApiServer
	client_api   edgeproto.AppInstApiClient
}

func (x *AppInstCommonApi) CreateAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	copy := &edgeproto.AppInst{}
	*copy = *in
	if x.internal_api != nil {
		err := x.internal_api.CreateAppInst(copy, &CudStreamoutAppInst{})
		return &edgeproto.Result{}, err
	} else {
		stream, err := x.client_api.CreateAppInst(ctx, copy)
		err = AppInstReadResultStream(stream, err)
		return &edgeproto.Result{}, err
	}
}

func (x *AppInstCommonApi) UpdateAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	copy := &edgeproto.AppInst{}
	*copy = *in
	if x.internal_api != nil {
		err := x.internal_api.UpdateAppInst(copy, &CudStreamoutAppInst{})
		return &edgeproto.Result{}, err
	} else {
		stream, err := x.client_api.UpdateAppInst(ctx, copy)
		err = AppInstReadResultStream(stream, err)
		return &edgeproto.Result{}, err
	}
}

func (x *AppInstCommonApi) DeleteAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	copy := &edgeproto.AppInst{}
	*copy = *in
	if x.internal_api != nil {
		err := x.internal_api.DeleteAppInst(copy, &CudStreamoutAppInst{})
		return &edgeproto.Result{}, err
	} else {
		stream, err := x.client_api.DeleteAppInst(ctx, copy)
		err = AppInstReadResultStream(stream, err)
		return &edgeproto.Result{}, err
	}
}

func (x *AppInstCommonApi) ShowAppInst(ctx context.Context, filter *edgeproto.AppInst, showData *ShowAppInst) error {
	if x.internal_api != nil {
		return x.internal_api.ShowAppInst(filter, showData)
	} else {
		stream, err := x.client_api.ShowAppInst(ctx, filter)
		showData.ReadStream(stream, err)
		return err
	}
}

func NewInternalAppInstApi(api edgeproto.AppInstApiServer) *AppInstCommonApi {
	apiWrap := AppInstCommonApi{}
	apiWrap.internal_api = api
	return &apiWrap
}

func NewClientAppInstApi(api edgeproto.AppInstApiClient) *AppInstCommonApi {
	apiWrap := AppInstCommonApi{}
	apiWrap.client_api = api
	return &apiWrap
}

func InternalAppInstTest(t *testing.T, test string, api edgeproto.AppInstApiServer, testData []edgeproto.AppInst) {
	switch test {
	case "cud":
		basicAppInstCudTest(t, NewInternalAppInstApi(api), testData)
	case "show":
		basicAppInstShowTest(t, NewInternalAppInstApi(api), testData)
	}
}

func ClientAppInstTest(t *testing.T, test string, api edgeproto.AppInstApiClient, testData []edgeproto.AppInst) {
	switch test {
	case "cud":
		basicAppInstCudTest(t, NewClientAppInstApi(api), testData)
	case "show":
		basicAppInstShowTest(t, NewClientAppInstApi(api), testData)
	}
}

func basicAppInstShowTest(t *testing.T, api *AppInstCommonApi, testData []edgeproto.AppInst) {
	var err error
	ctx := context.TODO()

	show := ShowAppInst{}
	show.Init()
	filterNone := edgeproto.AppInst{}
	err = api.ShowAppInst(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	assert.Equal(t, len(testData), len(show.Data), "Show count")
	for _, obj := range testData {
		show.AssertFound(t, &obj)
	}
}

func GetAppInst(t *testing.T, api *AppInstCommonApi, key *edgeproto.AppInstKey, out *edgeproto.AppInst) bool {
	var err error
	ctx := context.TODO()

	show := ShowAppInst{}
	show.Init()
	filter := edgeproto.AppInst{}
	filter.Key = *key
	err = api.ShowAppInst(ctx, &filter, &show)
	assert.Nil(t, err, "show data")
	obj, found := show.Data[key.GetKeyString()]
	if found {
		*out = obj
	}
	return found
}

func basicAppInstCudTest(t *testing.T, api *AppInstCommonApi, testData []edgeproto.AppInst) {
	var err error
	ctx := context.TODO()

	if len(testData) < 3 {
		assert.True(t, false, "Need at least 3 test data objects")
		return
	}

	// test create
	createAppInstData(t, api, testData)

	// test duplicate create - should fail
	_, err = api.CreateAppInst(ctx, &testData[0])
	assert.NotNil(t, err, "Create duplicate AppInst")

	// test show all items
	basicAppInstShowTest(t, api, testData)

	// test delete
	_, err = api.DeleteAppInst(ctx, &testData[0])
	assert.Nil(t, err, "delete AppInst %s", testData[0].Key.GetKeyString())
	show := ShowAppInst{}
	show.Init()
	filterNone := edgeproto.AppInst{}
	err = api.ShowAppInst(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	assert.Equal(t, len(testData)-1, len(show.Data), "Show count")
	show.AssertNotFound(t, &testData[0])
	// test update of missing object
	_, err = api.UpdateAppInst(ctx, &testData[0])
	assert.NotNil(t, err, "Update missing object")
	// create it back
	_, err = api.CreateAppInst(ctx, &testData[0])
	assert.Nil(t, err, "Create AppInst %s", testData[0].Key.GetKeyString())

	// test invalid keys
	bad := edgeproto.AppInst{}
	_, err = api.CreateAppInst(ctx, &bad)
	assert.NotNil(t, err, "Create AppInst with no key info")

}

func InternalAppInstCreate(t *testing.T, api edgeproto.AppInstApiServer, testData []edgeproto.AppInst) {
	createAppInstData(t, NewInternalAppInstApi(api), testData)
}

func ClientAppInstCreate(t *testing.T, api edgeproto.AppInstApiClient, testData []edgeproto.AppInst) {
	createAppInstData(t, NewClientAppInstApi(api), testData)
}

func createAppInstData(t *testing.T, api *AppInstCommonApi, testData []edgeproto.AppInst) {
	var err error
	ctx := context.TODO()

	for _, obj := range testData {
		_, err = api.CreateAppInst(ctx, &obj)
		assert.Nil(t, err, "Create AppInst %s", obj.Key.GetKeyString())
	}
}

type ShowAppInstInfo struct {
	Data map[string]edgeproto.AppInstInfo
	grpc.ServerStream
}

func (x *ShowAppInstInfo) Init() {
	x.Data = make(map[string]edgeproto.AppInstInfo)
}

func (x *ShowAppInstInfo) Send(m *edgeproto.AppInstInfo) error {
	x.Data[m.Key.GetKeyString()] = *m
	return nil
}

func (x *ShowAppInstInfo) ReadStream(stream edgeproto.AppInstInfoApi_ShowAppInstInfoClient, err error) {
	x.Data = make(map[string]edgeproto.AppInstInfo)
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

func (x *ShowAppInstInfo) CheckFound(obj *edgeproto.AppInstInfo) bool {
	_, found := x.Data[obj.Key.GetKeyString()]
	return found
}

func (x *ShowAppInstInfo) AssertFound(t *testing.T, obj *edgeproto.AppInstInfo) {
	check, found := x.Data[obj.Key.GetKeyString()]
	assert.True(t, found, "find AppInstInfo %s", obj.Key.GetKeyString())
	if found && !check.Matches(obj, edgeproto.MatchIgnoreBackend(), edgeproto.MatchSortArrayedKeys()) {
		assert.Equal(t, *obj, check, "AppInstInfo are equal")
	}
	if found {
		// remove in case there are dups in the list, so the
		// same object cannot be used again
		delete(x.Data, obj.Key.GetKeyString())
	}
}

func (x *ShowAppInstInfo) AssertNotFound(t *testing.T, obj *edgeproto.AppInstInfo) {
	_, found := x.Data[obj.Key.GetKeyString()]
	assert.False(t, found, "do not find AppInstInfo %s", obj.Key.GetKeyString())
}

func WaitAssertFoundAppInstInfo(t *testing.T, api edgeproto.AppInstInfoApiClient, obj *edgeproto.AppInstInfo, count int, retry time.Duration) {
	show := ShowAppInstInfo{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowAppInstInfo(ctx, obj)
		show.ReadStream(stream, err)
		cancel()
		if show.CheckFound(obj) {
			break
		}
		time.Sleep(retry)
	}
	show.AssertFound(t, obj)
}

func WaitAssertNotFoundAppInstInfo(t *testing.T, api edgeproto.AppInstInfoApiClient, obj *edgeproto.AppInstInfo, count int, retry time.Duration) {
	show := ShowAppInstInfo{}
	filterNone := edgeproto.AppInstInfo{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowAppInstInfo(ctx, &filterNone)
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
type AppInstInfoCommonApi struct {
	internal_api edgeproto.AppInstInfoApiServer
	client_api   edgeproto.AppInstInfoApiClient
}

func (x *AppInstInfoCommonApi) ShowAppInstInfo(ctx context.Context, filter *edgeproto.AppInstInfo, showData *ShowAppInstInfo) error {
	if x.internal_api != nil {
		return x.internal_api.ShowAppInstInfo(filter, showData)
	} else {
		stream, err := x.client_api.ShowAppInstInfo(ctx, filter)
		showData.ReadStream(stream, err)
		return err
	}
}

func NewInternalAppInstInfoApi(api edgeproto.AppInstInfoApiServer) *AppInstInfoCommonApi {
	apiWrap := AppInstInfoCommonApi{}
	apiWrap.internal_api = api
	return &apiWrap
}

func NewClientAppInstInfoApi(api edgeproto.AppInstInfoApiClient) *AppInstInfoCommonApi {
	apiWrap := AppInstInfoCommonApi{}
	apiWrap.client_api = api
	return &apiWrap
}

func InternalAppInstInfoTest(t *testing.T, test string, api edgeproto.AppInstInfoApiServer, testData []edgeproto.AppInstInfo) {
	switch test {
	case "show":
		basicAppInstInfoShowTest(t, NewInternalAppInstInfoApi(api), testData)
	}
}

func ClientAppInstInfoTest(t *testing.T, test string, api edgeproto.AppInstInfoApiClient, testData []edgeproto.AppInstInfo) {
	switch test {
	case "show":
		basicAppInstInfoShowTest(t, NewClientAppInstInfoApi(api), testData)
	}
}

func basicAppInstInfoShowTest(t *testing.T, api *AppInstInfoCommonApi, testData []edgeproto.AppInstInfo) {
	var err error
	ctx := context.TODO()

	show := ShowAppInstInfo{}
	show.Init()
	filterNone := edgeproto.AppInstInfo{}
	err = api.ShowAppInstInfo(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	assert.Equal(t, len(testData), len(show.Data), "Show count")
	for _, obj := range testData {
		show.AssertFound(t, &obj)
	}
}

func GetAppInstInfo(t *testing.T, api *AppInstInfoCommonApi, key *edgeproto.AppInstKey, out *edgeproto.AppInstInfo) bool {
	var err error
	ctx := context.TODO()

	show := ShowAppInstInfo{}
	show.Init()
	filter := edgeproto.AppInstInfo{}
	filter.Key = *key
	err = api.ShowAppInstInfo(ctx, &filter, &show)
	assert.Nil(t, err, "show data")
	obj, found := show.Data[key.GetKeyString()]
	if found {
		*out = obj
	}
	return found
}

func (s *DummyServer) CreateAppInst(in *edgeproto.AppInst, server edgeproto.AppInstApi_CreateAppInstServer) error {
	if true {
		server.Send(&edgeproto.Result{})
		server.Send(&edgeproto.Result{})
		server.Send(&edgeproto.Result{})
	}
	return nil
}

func (s *DummyServer) DeleteAppInst(in *edgeproto.AppInst, server edgeproto.AppInstApi_DeleteAppInstServer) error {
	if true {
		server.Send(&edgeproto.Result{})
		server.Send(&edgeproto.Result{})
		server.Send(&edgeproto.Result{})
	}
	return nil
}

func (s *DummyServer) UpdateAppInst(in *edgeproto.AppInst, server edgeproto.AppInstApi_UpdateAppInstServer) error {
	if true {
		server.Send(&edgeproto.Result{})
		server.Send(&edgeproto.Result{})
		server.Send(&edgeproto.Result{})
	}
	return nil
}

func (s *DummyServer) ShowAppInst(in *edgeproto.AppInst, server edgeproto.AppInstApi_ShowAppInstServer) error {
	obj := &edgeproto.AppInst{}
	if obj.Matches(in, edgeproto.MatchFilter()) {
		server.Send(&edgeproto.AppInst{})
		server.Send(&edgeproto.AppInst{})
		server.Send(&edgeproto.AppInst{})
	}
	for _, out := range s.AppInsts {
		if !out.Matches(in, edgeproto.MatchFilter()) {
			continue
		}
		server.Send(&out)
	}
	return nil
}

func (s *DummyServer) ShowAppInstInfo(in *edgeproto.AppInstInfo, server edgeproto.AppInstInfoApi_ShowAppInstInfoServer) error {
	obj := &edgeproto.AppInstInfo{}
	if obj.Matches(in, edgeproto.MatchFilter()) {
		server.Send(&edgeproto.AppInstInfo{})
		server.Send(&edgeproto.AppInstInfo{})
		server.Send(&edgeproto.AppInstInfo{})
	}
	for _, out := range s.AppInstInfos {
		if !out.Matches(in, edgeproto.MatchFilter()) {
			continue
		}
		server.Send(&out)
	}
	return nil
}
