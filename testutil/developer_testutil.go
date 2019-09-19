// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: developer.proto

package testutil

import "google.golang.org/grpc"
import "github.com/mobiledgex/edge-cloud/edgeproto"
import "io"
import "testing"
import "context"
import "time"
import "github.com/stretchr/testify/require"
import "github.com/mobiledgex/edge-cloud/log"
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
	Ctx context.Context
}

func (x *ShowDeveloper) Init() {
	x.Data = make(map[string]edgeproto.Developer)
}

func (x *ShowDeveloper) Send(m *edgeproto.Developer) error {
	x.Data[m.GetKey().GetKeyString()] = *m
	return nil
}

func (x *ShowDeveloper) Context() context.Context {
	return x.Ctx
}

var DeveloperShowExtraCount = 0

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
		x.Data[obj.GetKey().GetKeyString()] = *obj
	}
}

func (x *ShowDeveloper) CheckFound(obj *edgeproto.Developer) bool {
	_, found := x.Data[obj.GetKey().GetKeyString()]
	return found
}

func (x *ShowDeveloper) AssertFound(t *testing.T, obj *edgeproto.Developer) {
	check, found := x.Data[obj.GetKey().GetKeyString()]
	require.True(t, found, "find Developer %s", obj.GetKey().GetKeyString())
	if found && !check.Matches(obj, edgeproto.MatchIgnoreBackend(), edgeproto.MatchSortArrayedKeys()) {
		require.Equal(t, *obj, check, "Developer are equal")
	}
	if found {
		// remove in case there are dups in the list, so the
		// same object cannot be used again
		delete(x.Data, obj.GetKey().GetKeyString())
	}
}

func (x *ShowDeveloper) AssertNotFound(t *testing.T, obj *edgeproto.Developer) {
	_, found := x.Data[obj.GetKey().GetKeyString()]
	require.False(t, found, "do not find Developer %s", obj.GetKey().GetKeyString())
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
	copy := &edgeproto.Developer{}
	*copy = *in
	if x.internal_api != nil {
		return x.internal_api.CreateDeveloper(ctx, copy)
	} else {
		return x.client_api.CreateDeveloper(ctx, copy)
	}
}

func (x *DeveloperCommonApi) DeleteDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	copy := &edgeproto.Developer{}
	*copy = *in
	if x.internal_api != nil {
		return x.internal_api.DeleteDeveloper(ctx, copy)
	} else {
		return x.client_api.DeleteDeveloper(ctx, copy)
	}
}

func (x *DeveloperCommonApi) UpdateDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	copy := &edgeproto.Developer{}
	*copy = *in
	if x.internal_api != nil {
		return x.internal_api.UpdateDeveloper(ctx, copy)
	} else {
		return x.client_api.UpdateDeveloper(ctx, copy)
	}
}

func (x *DeveloperCommonApi) ShowDeveloper(ctx context.Context, filter *edgeproto.Developer, showData *ShowDeveloper) error {
	if x.internal_api != nil {
		showData.Ctx = ctx
		return x.internal_api.ShowDeveloper(filter, showData)
	} else {
		stream, err := x.client_api.ShowDeveloper(ctx, filter)
		showData.ReadStream(stream, err)
		return err
	}
}

func NewInternalDeveloperApi(api edgeproto.DeveloperApiServer) *DeveloperCommonApi {
	apiWrap := DeveloperCommonApi{}
	apiWrap.internal_api = api
	return &apiWrap
}

func NewClientDeveloperApi(api edgeproto.DeveloperApiClient) *DeveloperCommonApi {
	apiWrap := DeveloperCommonApi{}
	apiWrap.client_api = api
	return &apiWrap
}

func InternalDeveloperTest(t *testing.T, test string, api edgeproto.DeveloperApiServer, testData []edgeproto.Developer) {
	span := log.StartSpan(log.DebugLevelApi, "InternalDeveloperTest")
	defer span.Finish()
	ctx := log.ContextWithSpan(context.Background(), span)

	switch test {
	case "cud":
		basicDeveloperCudTest(t, ctx, NewInternalDeveloperApi(api), testData)
	case "show":
		basicDeveloperShowTest(t, ctx, NewInternalDeveloperApi(api), testData)
	}
}

func ClientDeveloperTest(t *testing.T, test string, api edgeproto.DeveloperApiClient, testData []edgeproto.Developer) {
	span := log.StartSpan(log.DebugLevelApi, "ClientDeveloperTest")
	defer span.Finish()
	ctx := log.ContextWithSpan(context.Background(), span)

	switch test {
	case "cud":
		basicDeveloperCudTest(t, ctx, NewClientDeveloperApi(api), testData)
	case "show":
		basicDeveloperShowTest(t, ctx, NewClientDeveloperApi(api), testData)
	}
}

func basicDeveloperShowTest(t *testing.T, ctx context.Context, api *DeveloperCommonApi, testData []edgeproto.Developer) {
	var err error

	show := ShowDeveloper{}
	show.Init()
	filterNone := edgeproto.Developer{}
	err = api.ShowDeveloper(ctx, &filterNone, &show)
	require.Nil(t, err, "show data")
	require.Equal(t, len(testData)+DeveloperShowExtraCount, len(show.Data), "Show count")
	for _, obj := range testData {
		show.AssertFound(t, &obj)
	}
}

func GetDeveloper(t *testing.T, ctx context.Context, api *DeveloperCommonApi, key *edgeproto.DeveloperKey, out *edgeproto.Developer) bool {
	var err error

	show := ShowDeveloper{}
	show.Init()
	filter := edgeproto.Developer{}
	filter.Key = *key
	err = api.ShowDeveloper(ctx, &filter, &show)
	require.Nil(t, err, "show data")
	obj, found := show.Data[key.GetKeyString()]
	if found {
		*out = obj
	}
	return found
}

func basicDeveloperCudTest(t *testing.T, ctx context.Context, api *DeveloperCommonApi, testData []edgeproto.Developer) {
	var err error

	if len(testData) < 3 {
		require.True(t, false, "Need at least 3 test data objects")
		return
	}

	// test create
	CreateDeveloperData(t, ctx, api, testData)

	// test duplicate Create - should fail
	_, err = api.CreateDeveloper(ctx, &testData[0])
	require.NotNil(t, err, "Create duplicate Developer")

	// test show all items
	basicDeveloperShowTest(t, ctx, api, testData)

	// test Delete
	_, err = api.DeleteDeveloper(ctx, &testData[0])
	require.Nil(t, err, "Delete Developer %s", testData[0].GetKey().GetKeyString())
	show := ShowDeveloper{}
	show.Init()
	filterNone := edgeproto.Developer{}
	err = api.ShowDeveloper(ctx, &filterNone, &show)
	require.Nil(t, err, "show data")
	require.Equal(t, len(testData)-1+DeveloperShowExtraCount, len(show.Data), "Show count")
	show.AssertNotFound(t, &testData[0])
	// test update of missing object
	_, err = api.UpdateDeveloper(ctx, &testData[0])
	require.NotNil(t, err, "Update missing object")
	// Create it back
	_, err = api.CreateDeveloper(ctx, &testData[0])
	require.Nil(t, err, "Create Developer %s", testData[0].GetKey().GetKeyString())

	// test invalid keys
	bad := edgeproto.Developer{}
	_, err = api.CreateDeveloper(ctx, &bad)
	require.NotNil(t, err, "Create Developer with no key info")

}

func InternalDeveloperCreate(t *testing.T, api edgeproto.DeveloperApiServer, testData []edgeproto.Developer) {
	span := log.StartSpan(log.DebugLevelApi, "InternalDeveloperCreate")
	defer span.Finish()
	ctx := log.ContextWithSpan(context.Background(), span)

	CreateDeveloperData(t, ctx, NewInternalDeveloperApi(api), testData)
}

func ClientDeveloperCreate(t *testing.T, api edgeproto.DeveloperApiClient, testData []edgeproto.Developer) {
	span := log.StartSpan(log.DebugLevelApi, "ClientDeveloperCreate")
	defer span.Finish()
	ctx := log.ContextWithSpan(context.Background(), span)

	CreateDeveloperData(t, ctx, NewClientDeveloperApi(api), testData)
}

func CreateDeveloperData(t *testing.T, ctx context.Context, api *DeveloperCommonApi, testData []edgeproto.Developer) {
	var err error

	for _, obj := range testData {
		_, err = api.CreateDeveloper(ctx, &obj)
		require.Nil(t, err, "Create Developer %s", obj.GetKey().GetKeyString())
	}
}

func FindDeveloperData(key *edgeproto.DeveloperKey, testData []edgeproto.Developer) (*edgeproto.Developer, bool) {
	for ii, _ := range testData {
		if testData[ii].Key.Matches(key) {
			return &testData[ii], true
		}
	}
	return nil, false
}

func (s *DummyServer) CreateDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) DeleteDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) UpdateDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) ShowDeveloper(in *edgeproto.Developer, server edgeproto.DeveloperApi_ShowDeveloperServer) error {
	obj := &edgeproto.Developer{}
	if obj.Matches(in, edgeproto.MatchFilter()) {
		server.Send(&edgeproto.Developer{})
		server.Send(&edgeproto.Developer{})
		server.Send(&edgeproto.Developer{})
	}
	for _, out := range s.Developers {
		if !out.Matches(in, edgeproto.MatchFilter()) {
			continue
		}
		server.Send(&out)
	}
	return nil
}

type DummyServer struct {
	Developers          []edgeproto.Developer
	Flavors             []edgeproto.Flavor
	Apps                []edgeproto.App
	Operators           []edgeproto.Operator
	Cloudlets           []edgeproto.Cloudlet
	CloudletInfos       []edgeproto.CloudletInfo
	ClusterInsts        []edgeproto.ClusterInst
	ClusterInstInfos    []edgeproto.ClusterInstInfo
	AppInsts            []edgeproto.AppInst
	AppInstInfos        []edgeproto.AppInstInfo
	CloudletPools       []edgeproto.CloudletPool
	CloudletPoolMembers []edgeproto.CloudletPoolMember
	Controllers         []edgeproto.Controller
	Nodes               []edgeproto.Node
	CloudletRefss       []edgeproto.CloudletRefs
	ClusterRefss        []edgeproto.ClusterRefs
}

func RegisterDummyServer(server *grpc.Server) *DummyServer {
	d := &DummyServer{}
	d.Developers = make([]edgeproto.Developer, 0)
	d.Flavors = make([]edgeproto.Flavor, 0)
	d.Apps = make([]edgeproto.App, 0)
	d.Operators = make([]edgeproto.Operator, 0)
	d.Cloudlets = make([]edgeproto.Cloudlet, 0)
	d.CloudletInfos = make([]edgeproto.CloudletInfo, 0)
	d.ClusterInsts = make([]edgeproto.ClusterInst, 0)
	d.ClusterInstInfos = make([]edgeproto.ClusterInstInfo, 0)
	d.AppInsts = make([]edgeproto.AppInst, 0)
	d.AppInstInfos = make([]edgeproto.AppInstInfo, 0)
	d.CloudletPools = make([]edgeproto.CloudletPool, 0)
	d.CloudletPoolMembers = make([]edgeproto.CloudletPoolMember, 0)
	d.Controllers = make([]edgeproto.Controller, 0)
	d.Nodes = make([]edgeproto.Node, 0)
	d.CloudletRefss = make([]edgeproto.CloudletRefs, 0)
	d.ClusterRefss = make([]edgeproto.ClusterRefs, 0)
	edgeproto.RegisterDeveloperApiServer(server, d)
	edgeproto.RegisterFlavorApiServer(server, d)
	edgeproto.RegisterAppApiServer(server, d)
	edgeproto.RegisterOperatorApiServer(server, d)
	edgeproto.RegisterCloudletApiServer(server, d)
	edgeproto.RegisterCloudletInfoApiServer(server, d)
	edgeproto.RegisterClusterInstApiServer(server, d)
	edgeproto.RegisterClusterInstInfoApiServer(server, d)
	edgeproto.RegisterAppInstApiServer(server, d)
	edgeproto.RegisterAppInstInfoApiServer(server, d)
	edgeproto.RegisterCloudletPoolApiServer(server, d)
	edgeproto.RegisterControllerApiServer(server, d)
	edgeproto.RegisterNodeApiServer(server, d)
	edgeproto.RegisterCloudletRefsApiServer(server, d)
	edgeproto.RegisterClusterRefsApiServer(server, d)
	return d
}
