// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: clusterinst.proto

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

type ShowClusterInst struct {
	Data map[string]edgeproto.ClusterInst
	grpc.ServerStream
}

func (x *ShowClusterInst) Init() {
	x.Data = make(map[string]edgeproto.ClusterInst)
}

func (x *ShowClusterInst) Send(m *edgeproto.ClusterInst) error {
	x.Data[m.Key.GetKeyString()] = *m
	return nil
}

type CudStreamoutClusterInst struct {
	grpc.ServerStream
}

func (x *CudStreamoutClusterInst) Send(res *edgeproto.Result) error {
	fmt.Println(res)
	return nil
}
func (x *CudStreamoutClusterInst) Context() context.Context {
	return context.TODO()
}

type ClusterInstStream interface {
	Recv() (*edgeproto.Result, error)
}

func ClusterInstReadResultStream(stream ClusterInstStream, err error) error {
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

func (x *ShowClusterInst) ReadStream(stream edgeproto.ClusterInstApi_ShowClusterInstClient, err error) {
	x.Data = make(map[string]edgeproto.ClusterInst)
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

func (x *ShowClusterInst) CheckFound(obj *edgeproto.ClusterInst) bool {
	_, found := x.Data[obj.Key.GetKeyString()]
	return found
}

func (x *ShowClusterInst) AssertFound(t *testing.T, obj *edgeproto.ClusterInst) {
	check, found := x.Data[obj.Key.GetKeyString()]
	assert.True(t, found, "find ClusterInst %s", obj.Key.GetKeyString())
	if found && !check.Matches(obj, edgeproto.MatchIgnoreBackend(), edgeproto.MatchSortArrayedKeys()) {
		assert.Equal(t, *obj, check, "ClusterInst are equal")
	}
	if found {
		// remove in case there are dups in the list, so the
		// same object cannot be used again
		delete(x.Data, obj.Key.GetKeyString())
	}
}

func (x *ShowClusterInst) AssertNotFound(t *testing.T, obj *edgeproto.ClusterInst) {
	_, found := x.Data[obj.Key.GetKeyString()]
	assert.False(t, found, "do not find ClusterInst %s", obj.Key.GetKeyString())
}

func WaitAssertFoundClusterInst(t *testing.T, api edgeproto.ClusterInstApiClient, obj *edgeproto.ClusterInst, count int, retry time.Duration) {
	show := ShowClusterInst{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowClusterInst(ctx, obj)
		show.ReadStream(stream, err)
		cancel()
		if show.CheckFound(obj) {
			break
		}
		time.Sleep(retry)
	}
	show.AssertFound(t, obj)
}

func WaitAssertNotFoundClusterInst(t *testing.T, api edgeproto.ClusterInstApiClient, obj *edgeproto.ClusterInst, count int, retry time.Duration) {
	show := ShowClusterInst{}
	filterNone := edgeproto.ClusterInst{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowClusterInst(ctx, &filterNone)
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
type ClusterInstCommonApi struct {
	internal_api edgeproto.ClusterInstApiServer
	client_api   edgeproto.ClusterInstApiClient
}

func (x *ClusterInstCommonApi) CreateClusterInst(ctx context.Context, in *edgeproto.ClusterInst) (*edgeproto.Result, error) {
	copy := &edgeproto.ClusterInst{}
	*copy = *in
	if x.internal_api != nil {
		err := x.internal_api.CreateClusterInst(copy, &CudStreamoutClusterInst{})
		return &edgeproto.Result{}, err
	} else {
		stream, err := x.client_api.CreateClusterInst(ctx, copy)
		err = ClusterInstReadResultStream(stream, err)
		return &edgeproto.Result{}, err
	}
}

func (x *ClusterInstCommonApi) UpdateClusterInst(ctx context.Context, in *edgeproto.ClusterInst) (*edgeproto.Result, error) {
	copy := &edgeproto.ClusterInst{}
	*copy = *in
	if x.internal_api != nil {
		err := x.internal_api.UpdateClusterInst(copy, &CudStreamoutClusterInst{})
		return &edgeproto.Result{}, err
	} else {
		stream, err := x.client_api.UpdateClusterInst(ctx, copy)
		err = ClusterInstReadResultStream(stream, err)
		return &edgeproto.Result{}, err
	}
}

func (x *ClusterInstCommonApi) DeleteClusterInst(ctx context.Context, in *edgeproto.ClusterInst) (*edgeproto.Result, error) {
	copy := &edgeproto.ClusterInst{}
	*copy = *in
	if x.internal_api != nil {
		err := x.internal_api.DeleteClusterInst(copy, &CudStreamoutClusterInst{})
		return &edgeproto.Result{}, err
	} else {
		stream, err := x.client_api.DeleteClusterInst(ctx, copy)
		err = ClusterInstReadResultStream(stream, err)
		return &edgeproto.Result{}, err
	}
}

func (x *ClusterInstCommonApi) ShowClusterInst(ctx context.Context, filter *edgeproto.ClusterInst, showData *ShowClusterInst) error {
	if x.internal_api != nil {
		return x.internal_api.ShowClusterInst(filter, showData)
	} else {
		stream, err := x.client_api.ShowClusterInst(ctx, filter)
		showData.ReadStream(stream, err)
		return err
	}
}

func NewInternalClusterInstApi(api edgeproto.ClusterInstApiServer) *ClusterInstCommonApi {
	apiWrap := ClusterInstCommonApi{}
	apiWrap.internal_api = api
	return &apiWrap
}

func NewClientClusterInstApi(api edgeproto.ClusterInstApiClient) *ClusterInstCommonApi {
	apiWrap := ClusterInstCommonApi{}
	apiWrap.client_api = api
	return &apiWrap
}

func InternalClusterInstTest(t *testing.T, test string, api edgeproto.ClusterInstApiServer, testData []edgeproto.ClusterInst) {
	switch test {
	case "cud":
		basicClusterInstCudTest(t, NewInternalClusterInstApi(api), testData)
	case "show":
		basicClusterInstShowTest(t, NewInternalClusterInstApi(api), testData)
	}
}

func ClientClusterInstTest(t *testing.T, test string, api edgeproto.ClusterInstApiClient, testData []edgeproto.ClusterInst) {
	switch test {
	case "cud":
		basicClusterInstCudTest(t, NewClientClusterInstApi(api), testData)
	case "show":
		basicClusterInstShowTest(t, NewClientClusterInstApi(api), testData)
	}
}

func basicClusterInstShowTest(t *testing.T, api *ClusterInstCommonApi, testData []edgeproto.ClusterInst) {
	var err error
	ctx := context.TODO()

	show := ShowClusterInst{}
	show.Init()
	filterNone := edgeproto.ClusterInst{}
	err = api.ShowClusterInst(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	assert.Equal(t, len(testData), len(show.Data), "Show count")
	for _, obj := range testData {
		show.AssertFound(t, &obj)
	}
}

func GetClusterInst(t *testing.T, api *ClusterInstCommonApi, key *edgeproto.ClusterInstKey, out *edgeproto.ClusterInst) bool {
	var err error
	ctx := context.TODO()

	show := ShowClusterInst{}
	show.Init()
	filter := edgeproto.ClusterInst{}
	filter.Key = *key
	err = api.ShowClusterInst(ctx, &filter, &show)
	assert.Nil(t, err, "show data")
	obj, found := show.Data[key.GetKeyString()]
	if found {
		*out = obj
	}
	return found
}

func basicClusterInstCudTest(t *testing.T, api *ClusterInstCommonApi, testData []edgeproto.ClusterInst) {
	var err error
	ctx := context.TODO()

	if len(testData) < 3 {
		assert.True(t, false, "Need at least 3 test data objects")
		return
	}

	// test create
	createClusterInstData(t, api, testData)

	// test duplicate create - should fail
	_, err = api.CreateClusterInst(ctx, &testData[0])
	assert.NotNil(t, err, "Create duplicate ClusterInst")

	// test show all items
	basicClusterInstShowTest(t, api, testData)

	// test delete
	_, err = api.DeleteClusterInst(ctx, &testData[0])
	assert.Nil(t, err, "delete ClusterInst %s", testData[0].Key.GetKeyString())
	show := ShowClusterInst{}
	show.Init()
	filterNone := edgeproto.ClusterInst{}
	err = api.ShowClusterInst(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	assert.Equal(t, len(testData)-1, len(show.Data), "Show count")
	show.AssertNotFound(t, &testData[0])
	// test update of missing object
	_, err = api.UpdateClusterInst(ctx, &testData[0])
	assert.NotNil(t, err, "Update missing object")
	// create it back
	_, err = api.CreateClusterInst(ctx, &testData[0])
	assert.Nil(t, err, "Create ClusterInst %s", testData[0].Key.GetKeyString())

	// test invalid keys
	bad := edgeproto.ClusterInst{}
	_, err = api.CreateClusterInst(ctx, &bad)
	assert.NotNil(t, err, "Create ClusterInst with no key info")

}

func InternalClusterInstCreate(t *testing.T, api edgeproto.ClusterInstApiServer, testData []edgeproto.ClusterInst) {
	createClusterInstData(t, NewInternalClusterInstApi(api), testData)
}

func ClientClusterInstCreate(t *testing.T, api edgeproto.ClusterInstApiClient, testData []edgeproto.ClusterInst) {
	createClusterInstData(t, NewClientClusterInstApi(api), testData)
}

func createClusterInstData(t *testing.T, api *ClusterInstCommonApi, testData []edgeproto.ClusterInst) {
	var err error
	ctx := context.TODO()

	for _, obj := range testData {
		_, err = api.CreateClusterInst(ctx, &obj)
		assert.Nil(t, err, "Create ClusterInst %s", obj.Key.GetKeyString())
	}
}

// Auto-generated code: DO NOT EDIT

type ShowClusterInstInfo struct {
	Data map[string]edgeproto.ClusterInstInfo
	grpc.ServerStream
}

func (x *ShowClusterInstInfo) Init() {
	x.Data = make(map[string]edgeproto.ClusterInstInfo)
}

func (x *ShowClusterInstInfo) Send(m *edgeproto.ClusterInstInfo) error {
	x.Data[m.Key.GetKeyString()] = *m
	return nil
}

func (x *ShowClusterInstInfo) ReadStream(stream edgeproto.ClusterInstInfoApi_ShowClusterInstInfoClient, err error) {
	x.Data = make(map[string]edgeproto.ClusterInstInfo)
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

func (x *ShowClusterInstInfo) CheckFound(obj *edgeproto.ClusterInstInfo) bool {
	_, found := x.Data[obj.Key.GetKeyString()]
	return found
}

func (x *ShowClusterInstInfo) AssertFound(t *testing.T, obj *edgeproto.ClusterInstInfo) {
	check, found := x.Data[obj.Key.GetKeyString()]
	assert.True(t, found, "find ClusterInstInfo %s", obj.Key.GetKeyString())
	if found && !check.Matches(obj, edgeproto.MatchIgnoreBackend(), edgeproto.MatchSortArrayedKeys()) {
		assert.Equal(t, *obj, check, "ClusterInstInfo are equal")
	}
	if found {
		// remove in case there are dups in the list, so the
		// same object cannot be used again
		delete(x.Data, obj.Key.GetKeyString())
	}
}

func (x *ShowClusterInstInfo) AssertNotFound(t *testing.T, obj *edgeproto.ClusterInstInfo) {
	_, found := x.Data[obj.Key.GetKeyString()]
	assert.False(t, found, "do not find ClusterInstInfo %s", obj.Key.GetKeyString())
}

func WaitAssertFoundClusterInstInfo(t *testing.T, api edgeproto.ClusterInstInfoApiClient, obj *edgeproto.ClusterInstInfo, count int, retry time.Duration) {
	show := ShowClusterInstInfo{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowClusterInstInfo(ctx, obj)
		show.ReadStream(stream, err)
		cancel()
		if show.CheckFound(obj) {
			break
		}
		time.Sleep(retry)
	}
	show.AssertFound(t, obj)
}

func WaitAssertNotFoundClusterInstInfo(t *testing.T, api edgeproto.ClusterInstInfoApiClient, obj *edgeproto.ClusterInstInfo, count int, retry time.Duration) {
	show := ShowClusterInstInfo{}
	filterNone := edgeproto.ClusterInstInfo{}
	for ii := 0; ii < count; ii++ {
		ctx, cancel := context.WithTimeout(context.Background(), retry)
		stream, err := api.ShowClusterInstInfo(ctx, &filterNone)
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
type ClusterInstInfoCommonApi struct {
	internal_api edgeproto.ClusterInstInfoApiServer
	client_api   edgeproto.ClusterInstInfoApiClient
}

func (x *ClusterInstInfoCommonApi) ShowClusterInstInfo(ctx context.Context, filter *edgeproto.ClusterInstInfo, showData *ShowClusterInstInfo) error {
	if x.internal_api != nil {
		return x.internal_api.ShowClusterInstInfo(filter, showData)
	} else {
		stream, err := x.client_api.ShowClusterInstInfo(ctx, filter)
		showData.ReadStream(stream, err)
		return err
	}
}

func NewInternalClusterInstInfoApi(api edgeproto.ClusterInstInfoApiServer) *ClusterInstInfoCommonApi {
	apiWrap := ClusterInstInfoCommonApi{}
	apiWrap.internal_api = api
	return &apiWrap
}

func NewClientClusterInstInfoApi(api edgeproto.ClusterInstInfoApiClient) *ClusterInstInfoCommonApi {
	apiWrap := ClusterInstInfoCommonApi{}
	apiWrap.client_api = api
	return &apiWrap
}

func InternalClusterInstInfoTest(t *testing.T, test string, api edgeproto.ClusterInstInfoApiServer, testData []edgeproto.ClusterInstInfo) {
	switch test {
	case "show":
		basicClusterInstInfoShowTest(t, NewInternalClusterInstInfoApi(api), testData)
	}
}

func ClientClusterInstInfoTest(t *testing.T, test string, api edgeproto.ClusterInstInfoApiClient, testData []edgeproto.ClusterInstInfo) {
	switch test {
	case "show":
		basicClusterInstInfoShowTest(t, NewClientClusterInstInfoApi(api), testData)
	}
}

func basicClusterInstInfoShowTest(t *testing.T, api *ClusterInstInfoCommonApi, testData []edgeproto.ClusterInstInfo) {
	var err error
	ctx := context.TODO()

	show := ShowClusterInstInfo{}
	show.Init()
	filterNone := edgeproto.ClusterInstInfo{}
	err = api.ShowClusterInstInfo(ctx, &filterNone, &show)
	assert.Nil(t, err, "show data")
	assert.Equal(t, len(testData), len(show.Data), "Show count")
	for _, obj := range testData {
		show.AssertFound(t, &obj)
	}
}

func GetClusterInstInfo(t *testing.T, api *ClusterInstInfoCommonApi, key *edgeproto.ClusterInstKey, out *edgeproto.ClusterInstInfo) bool {
	var err error
	ctx := context.TODO()

	show := ShowClusterInstInfo{}
	show.Init()
	filter := edgeproto.ClusterInstInfo{}
	filter.Key = *key
	err = api.ShowClusterInstInfo(ctx, &filter, &show)
	assert.Nil(t, err, "show data")
	obj, found := show.Data[key.GetKeyString()]
	if found {
		*out = obj
	}
	return found
}
