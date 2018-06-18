// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: developer.proto

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
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

type ShowDeveloper struct {
	Data map[string]edgeproto.Developer
	grpc.ServerStream
}

func (x *ShowDeveloper) Init() {
	x.Data = make(map[string]edgeproto.Developer)
}

func (x *ShowDeveloper) Send(m *edgeproto.Developer) error {
	x.Data[m.Key.GetKeyString()] = *m
	return nil
}

func (x *ShowDeveloper) ReadStream(stream edgeproto.DeveloperApi_ShowDeveloperClient, err error) {
	x.Data = make(map[string]edgeproto.Developer)
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

func (x *ShowDeveloper) CheckFound(obj *edgeproto.Developer) bool {
	_, found := x.Data[obj.Key.GetKeyString()]
	return found
}

func (x *ShowDeveloper) AssertFound(t *testing.T, obj *edgeproto.Developer) {
	check, found := x.Data[obj.Key.GetKeyString()]
	assert.True(t, found, "find Developer %s", obj.Key.GetKeyString())
	if found {
		assert.Equal(t, *obj, check, "Developer are equal")
	}
}

func (x *ShowDeveloper) AssertNotFound(t *testing.T, obj *edgeproto.Developer) {
	_, found := x.Data[obj.Key.GetKeyString()]
	assert.False(t, found, "do not find Developer %s", obj.Key.GetKeyString())
}

func WaitAssertFoundDeveloper(t *testing.T, api edgeproto.DeveloperApiClient, obj *edgeproto.Developer, count int, retry time.Duration) {
	show := ShowDeveloper{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowDeveloper(ctx, obj)
		show.ReadStream(stream, err)
		cancel()
		if show.CheckFound(obj) {
			break
		}
		time.Sleep(retry)
	}
	show.AssertFound(t, obj)
}

func WaitAssertNotFoundDeveloper(t *testing.T, api edgeproto.DeveloperApiClient, obj *edgeproto.Developer, count int, retry time.Duration) {
	show := ShowDeveloper{}
	filterNone := edgeproto.Developer{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowDeveloper(ctx, &filterNone)
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
type DeveloperCommonApi struct {
	internal_api edgeproto.DeveloperApiServer
	client_api   edgeproto.DeveloperApiClient
}

func (x *DeveloperCommonApi) CreateDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.CreateDeveloper(ctx, in)
	} else {
		return x.client_api.CreateDeveloper(ctx, in)
	}
}

func (x *DeveloperCommonApi) UpdateDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.UpdateDeveloper(ctx, in)
	} else {
		return x.client_api.UpdateDeveloper(ctx, in)
	}
}

func (x *DeveloperCommonApi) DeleteDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.DeleteDeveloper(ctx, in)
	} else {
		return x.client_api.DeleteDeveloper(ctx, in)
	}
}

func (x *DeveloperCommonApi) ShowDeveloper(ctx context.Context, filter *edgeproto.Developer, showData *ShowDeveloper) error {
	if x.internal_api != nil {
		return x.internal_api.ShowDeveloper(filter, showData)
	} else {
		stream, err := x.client_api.ShowDeveloper(ctx, filter)
		showData.ReadStream(stream, err)
		return err
	}
}

func InternalDeveloperCudTest(t *testing.T, api edgeproto.DeveloperApiServer, testData []edgeproto.Developer) {
	apiWrap := DeveloperCommonApi{}
	apiWrap.internal_api = api
	basicDeveloperCudTest(t, &apiWrap, testData)
}

func ClientDeveloperCudTest(t *testing.T, api edgeproto.DeveloperApiClient, testData []edgeproto.Developer) {
	apiWrap := DeveloperCommonApi{}
	apiWrap.client_api = api
	basicDeveloperCudTest(t, &apiWrap, testData)
}

func basicDeveloperCudTest(t *testing.T, api *DeveloperCommonApi, testData []edgeproto.Developer) {
	var err error
	ctx := context.TODO()

	if len(testData) < 3 {
		assert.True(t, false, "Need at least 3 test data objects")
		return
	}

	// test create
	for _, obj := range testData {
		_, err = api.CreateDeveloper(ctx, &obj)
		assert.Nil(t, err, "Create Developer %s", obj.Key.GetKeyString())
	}
	_, err = api.CreateDeveloper(ctx, &testData[0])
	assert.NotNil(t, err, "Create duplicate Developer")

	// test show all items
	show := ShowDeveloper{}
	show.Init()
	filterNone := edgeproto.Developer{}
	err = api.ShowDeveloper(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	for _, obj := range testData {
		show.AssertFound(t, &obj)
	}
	assert.Equal(t, len(testData), len(show.Data), "Show count")

	// test delete
	_, err = api.DeleteDeveloper(ctx, &testData[0])
	assert.Nil(t, err, "delete Developer %s", testData[0].Key.GetKeyString())
	show.Init()
	err = api.ShowDeveloper(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	assert.Equal(t, len(testData)-1, len(show.Data), "Show count")
	show.AssertNotFound(t, &testData[0])
	// test update of missing object
	_, err = api.UpdateDeveloper(ctx, &testData[0])
	assert.NotNil(t, err, "Update missing object")
	// create it back
	_, err = api.CreateDeveloper(ctx, &testData[0])
	assert.Nil(t, err, "Create Developer %s", testData[0].Key.GetKeyString())

	// test invalid keys
	bad := edgeproto.Developer{}
	_, err = api.CreateDeveloper(ctx, &bad)
	assert.NotNil(t, err, "Create Developer with no key info")

	// test update
	updater := edgeproto.Developer{}
	updater.Key = testData[0].Key
	updater.Email = "update just this"
	updater.Fields = make([]string, 0)
	updater.Fields = append(updater.Fields, edgeproto.DeveloperFieldEmail)
	_, err = api.UpdateDeveloper(ctx, &updater)
	assert.Nil(t, err, "Update Developer %s", testData[0].Key.GetKeyString())

	show.Init()
	updater = testData[0]
	updater.Email = "update just this"
	err = api.ShowDeveloper(ctx, &filterNone, &show)
	assert.Nil(t, err, "show Developer")
	show.AssertFound(t, &updater)

	// revert change
	updater.Email = testData[0].Email
	_, err = api.UpdateDeveloper(ctx, &updater)
	assert.Nil(t, err, "Update back Developer")
}
