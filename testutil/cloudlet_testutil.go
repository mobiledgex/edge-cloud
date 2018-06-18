// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

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

type ShowCloudlet struct {
	Data map[string]edgeproto.Cloudlet
	grpc.ServerStream
}

func (x *ShowCloudlet) Init() {
	x.Data = make(map[string]edgeproto.Cloudlet)
}

func (x *ShowCloudlet) Send(m *edgeproto.Cloudlet) error {
	x.Data[m.Key.GetKeyString()] = *m
	return nil
}

func (x *ShowCloudlet) ReadStream(stream edgeproto.CloudletApi_ShowCloudletClient, err error) {
	x.Data = make(map[string]edgeproto.Cloudlet)
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

func (x *ShowCloudlet) CheckFound(obj *edgeproto.Cloudlet) bool {
	_, found := x.Data[obj.Key.GetKeyString()]
	return found
}

func (x *ShowCloudlet) AssertFound(t *testing.T, obj *edgeproto.Cloudlet) {
	check, found := x.Data[obj.Key.GetKeyString()]
	assert.True(t, found, "find Cloudlet %s", obj.Key.GetKeyString())
	if found {
		assert.Equal(t, *obj, check, "Cloudlet are equal")
	}
}

func (x *ShowCloudlet) AssertNotFound(t *testing.T, obj *edgeproto.Cloudlet) {
	_, found := x.Data[obj.Key.GetKeyString()]
	assert.False(t, found, "do not find Cloudlet %s", obj.Key.GetKeyString())
}

func WaitAssertFoundCloudlet(t *testing.T, api edgeproto.CloudletApiClient, obj *edgeproto.Cloudlet, count int, retry time.Duration) {
	show := ShowCloudlet{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowCloudlet(ctx, obj)
		show.ReadStream(stream, err)
		cancel()
		if show.CheckFound(obj) {
			break
		}
		time.Sleep(retry)
	}
	show.AssertFound(t, obj)
}

func WaitAssertNotFoundCloudlet(t *testing.T, api edgeproto.CloudletApiClient, obj *edgeproto.Cloudlet, count int, retry time.Duration) {
	show := ShowCloudlet{}
	filterNone := edgeproto.Cloudlet{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowCloudlet(ctx, &filterNone)
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
type CloudletCommonApi struct {
	internal_api edgeproto.CloudletApiServer
	client_api   edgeproto.CloudletApiClient
}

func (x *CloudletCommonApi) CreateCloudlet(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.CreateCloudlet(ctx, in)
	} else {
		return x.client_api.CreateCloudlet(ctx, in)
	}
}

func (x *CloudletCommonApi) UpdateCloudlet(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.UpdateCloudlet(ctx, in)
	} else {
		return x.client_api.UpdateCloudlet(ctx, in)
	}
}

func (x *CloudletCommonApi) DeleteCloudlet(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.DeleteCloudlet(ctx, in)
	} else {
		return x.client_api.DeleteCloudlet(ctx, in)
	}
}

func (x *CloudletCommonApi) ShowCloudlet(ctx context.Context, filter *edgeproto.Cloudlet, showData *ShowCloudlet) error {
	if x.internal_api != nil {
		return x.internal_api.ShowCloudlet(filter, showData)
	} else {
		stream, err := x.client_api.ShowCloudlet(ctx, filter)
		showData.ReadStream(stream, err)
		return err
	}
}

func InternalCloudletCudTest(t *testing.T, api edgeproto.CloudletApiServer, testData []edgeproto.Cloudlet) {
	apiWrap := CloudletCommonApi{}
	apiWrap.internal_api = api
	basicCloudletCudTest(t, &apiWrap, testData)
}

func ClientCloudletCudTest(t *testing.T, api edgeproto.CloudletApiClient, testData []edgeproto.Cloudlet) {
	apiWrap := CloudletCommonApi{}
	apiWrap.client_api = api
	basicCloudletCudTest(t, &apiWrap, testData)
}

func basicCloudletCudTest(t *testing.T, api *CloudletCommonApi, testData []edgeproto.Cloudlet) {
	var err error
	ctx := context.TODO()

	if len(testData) < 3 {
		assert.True(t, false, "Need at least 3 test data objects")
		return
	}

	// test create
	for _, obj := range testData {
		_, err = api.CreateCloudlet(ctx, &obj)
		assert.Nil(t, err, "Create Cloudlet %s", obj.Key.GetKeyString())
	}
	_, err = api.CreateCloudlet(ctx, &testData[0])
	assert.NotNil(t, err, "Create duplicate Cloudlet")

	// test show all items
	show := ShowCloudlet{}
	show.Init()
	filterNone := edgeproto.Cloudlet{}
	err = api.ShowCloudlet(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	for _, obj := range testData {
		show.AssertFound(t, &obj)
	}
	assert.Equal(t, len(testData), len(show.Data), "Show count")

	// test delete
	_, err = api.DeleteCloudlet(ctx, &testData[0])
	assert.Nil(t, err, "delete Cloudlet %s", testData[0].Key.GetKeyString())
	show.Init()
	err = api.ShowCloudlet(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	assert.Equal(t, len(testData)-1, len(show.Data), "Show count")
	show.AssertNotFound(t, &testData[0])
	// test update of missing object
	_, err = api.UpdateCloudlet(ctx, &testData[0])
	assert.NotNil(t, err, "Update missing object")
	// create it back
	_, err = api.CreateCloudlet(ctx, &testData[0])
	assert.Nil(t, err, "Create Cloudlet %s", testData[0].Key.GetKeyString())

	// test invalid keys
	bad := edgeproto.Cloudlet{}
	_, err = api.CreateCloudlet(ctx, &bad)
	assert.NotNil(t, err, "Create Cloudlet with no key info")

}
