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
	if found {
		assert.Equal(t, *obj, check, "AppInst are equal")
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
	if x.internal_api != nil {
		return x.internal_api.CreateAppInst(ctx, in)
	} else {
		return x.client_api.CreateAppInst(ctx, in)
	}
}

func (x *AppInstCommonApi) UpdateAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.UpdateAppInst(ctx, in)
	} else {
		return x.client_api.UpdateAppInst(ctx, in)
	}
}

func (x *AppInstCommonApi) DeleteAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.DeleteAppInst(ctx, in)
	} else {
		return x.client_api.DeleteAppInst(ctx, in)
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
func InternalAppInstCudTest(t *testing.T, api edgeproto.AppInstApiServer, testData []edgeproto.AppInst) {
	basicAppInstCudTest(t, NewInternalAppInstApi(api), testData)
}

func ClientAppInstCudTest(t *testing.T, api edgeproto.AppInstApiClient, testData []edgeproto.AppInst) {
	basicAppInstCudTest(t, NewClientAppInstApi(api), testData)
}

func basicAppInstCudTest(t *testing.T, api *AppInstCommonApi, testData []edgeproto.AppInst) {
	var err error
	ctx := context.TODO()

	if len(testData) < 3 {
		assert.True(t, false, "Need at least 3 test data objects")
		return
	}

	// test create
	for _, obj := range testData {
		_, err = api.CreateAppInst(ctx, &obj)
		assert.Nil(t, err, "Create AppInst %s", obj.Key.GetKeyString())
	}
	_, err = api.CreateAppInst(ctx, &testData[0])
	assert.NotNil(t, err, "Create duplicate AppInst")

	// test show all items
	show := ShowAppInst{}
	show.Init()
	filterNone := edgeproto.AppInst{}
	err = api.ShowAppInst(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	for _, obj := range testData {
		show.AssertFound(t, &obj)
	}
	assert.Equal(t, len(testData), len(show.Data), "Show count")

	// test delete
	_, err = api.DeleteAppInst(ctx, &testData[0])
	assert.Nil(t, err, "delete AppInst %s", testData[0].Key.GetKeyString())
	show.Init()
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

	// test update
	updater := edgeproto.AppInst{}
	updater.Key = testData[0].Key
	updater.Uri = "update just this"
	updater.Fields = make([]string, 0)
	updater.Fields = append(updater.Fields, edgeproto.AppInstFieldUri)
	_, err = api.UpdateAppInst(ctx, &updater)
	assert.Nil(t, err, "Update AppInst %s", testData[0].Key.GetKeyString())

	show.Init()
	updater = testData[0]
	updater.Uri = "update just this"
	err = api.ShowAppInst(ctx, &filterNone, &show)
	assert.Nil(t, err, "show AppInst")
	show.AssertFound(t, &updater)

	// revert change
	updater.Uri = testData[0].Uri
	_, err = api.UpdateAppInst(ctx, &updater)
	assert.Nil(t, err, "Update back AppInst")
}

// Auto-generated code: DO NOT EDIT

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
	if found {
		assert.Equal(t, *obj, check, "AppInstInfo are equal")
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
