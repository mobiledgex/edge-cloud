// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: operator.proto

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

type ShowOperator struct {
	Data map[string]edgeproto.Operator
	grpc.ServerStream
}

func (x *ShowOperator) Init() {
	x.Data = make(map[string]edgeproto.Operator)
}

func (x *ShowOperator) Send(m *edgeproto.Operator) error {
	x.Data[m.Key.GetKeyString()] = *m
	return nil
}

func (x *ShowOperator) ReadStream(stream edgeproto.OperatorApi_ShowOperatorClient, err error) {
	x.Data = make(map[string]edgeproto.Operator)
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

func (x *ShowOperator) CheckFound(obj *edgeproto.Operator) bool {
	_, found := x.Data[obj.Key.GetKeyString()]
	return found
}

func (x *ShowOperator) AssertFound(t *testing.T, obj *edgeproto.Operator) {
	check, found := x.Data[obj.Key.GetKeyString()]
	assert.True(t, found, "find Operator %s", obj.Key.GetKeyString())
	if found {
		assert.Equal(t, *obj, check, "Operator are equal")
	}
}

func (x *ShowOperator) AssertNotFound(t *testing.T, obj *edgeproto.Operator) {
	_, found := x.Data[obj.Key.GetKeyString()]
	assert.False(t, found, "do not find Operator %s", obj.Key.GetKeyString())
}

func WaitAssertFoundOperator(t *testing.T, api edgeproto.OperatorApiClient, obj *edgeproto.Operator, count int, retry time.Duration) {
	show := ShowOperator{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowOperator(ctx, obj)
		show.ReadStream(stream, err)
		cancel()
		if show.CheckFound(obj) {
			break
		}
		time.Sleep(retry)
	}
	show.AssertFound(t, obj)
}

func WaitAssertNotFoundOperator(t *testing.T, api edgeproto.OperatorApiClient, obj *edgeproto.Operator, count int, retry time.Duration) {
	show := ShowOperator{}
	filterNone := edgeproto.Operator{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowOperator(ctx, &filterNone)
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
type OperatorCommonApi struct {
	internal_api edgeproto.OperatorApiServer
	client_api   edgeproto.OperatorApiClient
}

func (x *OperatorCommonApi) CreateOperator(ctx context.Context, in *edgeproto.Operator) (*edgeproto.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.CreateOperator(ctx, in)
	} else {
		return x.client_api.CreateOperator(ctx, in)
	}
}

func (x *OperatorCommonApi) UpdateOperator(ctx context.Context, in *edgeproto.Operator) (*edgeproto.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.UpdateOperator(ctx, in)
	} else {
		return x.client_api.UpdateOperator(ctx, in)
	}
}

func (x *OperatorCommonApi) DeleteOperator(ctx context.Context, in *edgeproto.Operator) (*edgeproto.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.DeleteOperator(ctx, in)
	} else {
		return x.client_api.DeleteOperator(ctx, in)
	}
}

func (x *OperatorCommonApi) ShowOperator(ctx context.Context, filter *edgeproto.Operator, showData *ShowOperator) error {
	if x.internal_api != nil {
		return x.internal_api.ShowOperator(filter, showData)
	} else {
		stream, err := x.client_api.ShowOperator(ctx, filter)
		showData.ReadStream(stream, err)
		return err
	}
}

func InternalOperatorCudTest(t *testing.T, api edgeproto.OperatorApiServer, testData []edgeproto.Operator) {
	apiWrap := OperatorCommonApi{}
	apiWrap.internal_api = api
	basicOperatorCudTest(t, &apiWrap, testData)
}

func ClientOperatorCudTest(t *testing.T, api edgeproto.OperatorApiClient, testData []edgeproto.Operator) {
	apiWrap := OperatorCommonApi{}
	apiWrap.client_api = api
	basicOperatorCudTest(t, &apiWrap, testData)
}

func basicOperatorCudTest(t *testing.T, api *OperatorCommonApi, testData []edgeproto.Operator) {
	var err error
	ctx := context.TODO()

	if len(testData) < 3 {
		assert.True(t, false, "Need at least 3 test data objects")
		return
	}

	// test create
	for _, obj := range testData {
		_, err = api.CreateOperator(ctx, &obj)
		assert.Nil(t, err, "Create Operator %s", obj.Key.GetKeyString())
	}
	_, err = api.CreateOperator(ctx, &testData[0])
	assert.NotNil(t, err, "Create duplicate Operator")

	// test show all items
	show := ShowOperator{}
	show.Init()
	filterNone := edgeproto.Operator{}
	err = api.ShowOperator(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	for _, obj := range testData {
		show.AssertFound(t, &obj)
	}
	assert.Equal(t, len(testData), len(show.Data), "Show count")

	// test delete
	_, err = api.DeleteOperator(ctx, &testData[0])
	assert.Nil(t, err, "delete Operator %s", testData[0].Key.GetKeyString())
	show.Init()
	err = api.ShowOperator(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	assert.Equal(t, len(testData)-1, len(show.Data), "Show count")
	show.AssertNotFound(t, &testData[0])
	// test update of missing object
	_, err = api.UpdateOperator(ctx, &testData[0])
	assert.NotNil(t, err, "Update missing object")
	// create it back
	_, err = api.CreateOperator(ctx, &testData[0])
	assert.Nil(t, err, "Create Operator %s", testData[0].Key.GetKeyString())

	// test invalid keys
	bad := edgeproto.Operator{}
	_, err = api.CreateOperator(ctx, &bad)
	assert.NotNil(t, err, "Create Operator with no key info")

}
