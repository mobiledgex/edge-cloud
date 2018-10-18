// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app.proto

/*
Package testutil is a generated protocol buffer package.

It is generated from these files:
	app.proto
	app_inst.proto
	cloud-resource-manager.proto
	cloudlet.proto
	cluster.proto
	clusterflavor.proto
	clusterinst.proto
	common.proto
	controller.proto
	developer.proto
	flavor.proto
	metric.proto
	node.proto
	notice.proto
	operator.proto
	refs.proto
	result.proto
	user.proto

It has these top-level messages:
	AppKey
	App
	AppInstKey
	AppInst
	AppInstInfo
	AppInstMetrics
	CloudResource
	EdgeCloudApp
	EdgeCloudApplication
	CloudletKey
	Cloudlet
	CloudletInfo
	CloudletMetrics
	ClusterKey
	Cluster
	ClusterFlavorKey
	ClusterFlavor
	ClusterInstKey
	ClusterInst
	ClusterInstInfo
	ControllerKey
	Controller
	DeveloperKey
	Developer
	FlavorKey
	Flavor
	MetricTag
	MetricVal
	Metric
	NodeKey
	Node
	NoticeReply
	NoticeRequest
	OperatorKey
	Operator
	CloudletRefs
	ClusterRefs
	Result
	UserKey
	User
	RoleKey
	Role
*/
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
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

type ShowApp struct {
	Data map[string]edgeproto.App
	grpc.ServerStream
}

func (x *ShowApp) Init() {
	x.Data = make(map[string]edgeproto.App)
}

func (x *ShowApp) Send(m *edgeproto.App) error {
	x.Data[m.Key.GetKeyString()] = *m
	return nil
}

func (x *ShowApp) ReadStream(stream edgeproto.AppApi_ShowAppClient, err error) {
	x.Data = make(map[string]edgeproto.App)
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

func (x *ShowApp) CheckFound(obj *edgeproto.App) bool {
	_, found := x.Data[obj.Key.GetKeyString()]
	return found
}

func (x *ShowApp) AssertFound(t *testing.T, obj *edgeproto.App) {
	check, found := x.Data[obj.Key.GetKeyString()]
	assert.True(t, found, "find App %s", obj.Key.GetKeyString())
	if found && !check.Matches(obj, edgeproto.MatchIgnoreBackend(), edgeproto.MatchSortArrayedKeys()) {
		assert.Equal(t, *obj, check, "App are equal")
	}
	if found {
		// remove in case there are dups in the list, so the
		// same object cannot be used again
		delete(x.Data, obj.Key.GetKeyString())
	}
}

func (x *ShowApp) AssertNotFound(t *testing.T, obj *edgeproto.App) {
	_, found := x.Data[obj.Key.GetKeyString()]
	assert.False(t, found, "do not find App %s", obj.Key.GetKeyString())
}

func WaitAssertFoundApp(t *testing.T, api edgeproto.AppApiClient, obj *edgeproto.App, count int, retry time.Duration) {
	show := ShowApp{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowApp(ctx, obj)
		show.ReadStream(stream, err)
		cancel()
		if show.CheckFound(obj) {
			break
		}
		time.Sleep(retry)
	}
	show.AssertFound(t, obj)
}

func WaitAssertNotFoundApp(t *testing.T, api edgeproto.AppApiClient, obj *edgeproto.App, count int, retry time.Duration) {
	show := ShowApp{}
	filterNone := edgeproto.App{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowApp(ctx, &filterNone)
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
type AppCommonApi struct {
	internal_api edgeproto.AppApiServer
	client_api   edgeproto.AppApiClient
}

func (x *AppCommonApi) CreateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.CreateApp(ctx, in)
	} else {
		return x.client_api.CreateApp(ctx, in)
	}
}

func (x *AppCommonApi) UpdateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.UpdateApp(ctx, in)
	} else {
		return x.client_api.UpdateApp(ctx, in)
	}
}

func (x *AppCommonApi) DeleteApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	if x.internal_api != nil {
		return x.internal_api.DeleteApp(ctx, in)
	} else {
		return x.client_api.DeleteApp(ctx, in)
	}
}

func (x *AppCommonApi) ShowApp(ctx context.Context, filter *edgeproto.App, showData *ShowApp) error {
	if x.internal_api != nil {
		return x.internal_api.ShowApp(filter, showData)
	} else {
		stream, err := x.client_api.ShowApp(ctx, filter)
		showData.ReadStream(stream, err)
		return err
	}
}

func NewInternalAppApi(api edgeproto.AppApiServer) *AppCommonApi {
	apiWrap := AppCommonApi{}
	apiWrap.internal_api = api
	return &apiWrap
}

func NewClientAppApi(api edgeproto.AppApiClient) *AppCommonApi {
	apiWrap := AppCommonApi{}
	apiWrap.client_api = api
	return &apiWrap
}

func InternalAppTest(t *testing.T, test string, api edgeproto.AppApiServer, testData []edgeproto.App) {
	switch test {
	case "cud":
		basicAppCudTest(t, NewInternalAppApi(api), testData)
	case "show":
		basicAppShowTest(t, NewInternalAppApi(api), testData)
	}
}

func ClientAppTest(t *testing.T, test string, api edgeproto.AppApiClient, testData []edgeproto.App) {
	switch test {
	case "cud":
		basicAppCudTest(t, NewClientAppApi(api), testData)
	case "show":
		basicAppShowTest(t, NewClientAppApi(api), testData)
	}
}

func basicAppShowTest(t *testing.T, api *AppCommonApi, testData []edgeproto.App) {
	var err error
	ctx := context.TODO()

	show := ShowApp{}
	show.Init()
	filterNone := edgeproto.App{}
	err = api.ShowApp(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	assert.Equal(t, len(testData), len(show.Data), "Show count")
	for _, obj := range testData {
		show.AssertFound(t, &obj)
	}
}

func GetApp(t *testing.T, api *AppCommonApi, key *edgeproto.AppKey, out *edgeproto.App) bool {
	var err error
	ctx := context.TODO()

	show := ShowApp{}
	show.Init()
	filter := edgeproto.App{}
	filter.Key = *key
	err = api.ShowApp(ctx, &filter, &show)
	assert.Nil(t, err, "show data")
	obj, found := show.Data[key.GetKeyString()]
	if found {
		*out = obj
	}
	return found
}

func basicAppCudTest(t *testing.T, api *AppCommonApi, testData []edgeproto.App) {
	var err error
	ctx := context.TODO()

	if len(testData) < 3 {
		assert.True(t, false, "Need at least 3 test data objects")
		return
	}

	// test create
	createAppData(t, api, testData)

	// test duplicate create - should fail
	_, err = api.CreateApp(ctx, &testData[0])
	assert.NotNil(t, err, "Create duplicate App")

	// test show all items
	basicAppShowTest(t, api, testData)

	// test delete
	_, err = api.DeleteApp(ctx, &testData[0])
	assert.Nil(t, err, "delete App %s", testData[0].Key.GetKeyString())
	show := ShowApp{}
	show.Init()
	filterNone := edgeproto.App{}
	err = api.ShowApp(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	assert.Equal(t, len(testData)-1, len(show.Data), "Show count")
	show.AssertNotFound(t, &testData[0])
	// create it back
	_, err = api.CreateApp(ctx, &testData[0])
	assert.Nil(t, err, "Create App %s", testData[0].Key.GetKeyString())

	// test invalid keys
	bad := edgeproto.App{}
	_, err = api.CreateApp(ctx, &bad)
	assert.NotNil(t, err, "Create App with no key info")

}

func InternalAppCreate(t *testing.T, api edgeproto.AppApiServer, testData []edgeproto.App) {
	createAppData(t, NewInternalAppApi(api), testData)
}

func ClientAppCreate(t *testing.T, api edgeproto.AppApiClient, testData []edgeproto.App) {
	createAppData(t, NewClientAppApi(api), testData)
}

func createAppData(t *testing.T, api *AppCommonApi, testData []edgeproto.App) {
	var err error
	ctx := context.TODO()

	for _, obj := range testData {
		_, err = api.CreateApp(ctx, &obj)
		assert.Nil(t, err, "Create App %s", obj.Key.GetKeyString())
	}
}
