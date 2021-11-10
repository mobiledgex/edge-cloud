package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

var (
	Pass bool = true
	Fail bool = false
)

type StreamoutMsg struct {
	Msgs []edgeproto.Result
	grpc.ServerStream
	Ctx context.Context
}

func (x *StreamoutMsg) Send(msg *edgeproto.Result) error {
	x.Msgs = append(x.Msgs, *msg)
	return nil
}

func (x *StreamoutMsg) Context() context.Context {
	return x.Ctx
}

func NewStreamoutMsg(ctx context.Context) *StreamoutMsg {
	return &StreamoutMsg{
		Ctx: ctx,
	}
}

func GetAppInstStreamMsgs(t *testing.T, ctx context.Context, key *edgeproto.AppInstKey, apis *AllApis, pass bool) []edgeproto.Result {
	// Verify stream appInst
	streamAppInst := NewStreamoutMsg(ctx)
	err := apis.streamObjApi.StreamAppInst(key, streamAppInst)
	if pass {
		require.Nil(t, err, "stream appinst")
		require.Greater(t, len(streamAppInst.Msgs), 0, "contains stream messages")
	} else {
		require.NotNil(t, err, "stream appinst should return error for key %s", *key)
	}
	return streamAppInst.Msgs
}

func GetClusterInstStreamMsgs(t *testing.T, ctx context.Context, key *edgeproto.ClusterInstKey, apis *AllApis, pass bool) []edgeproto.Result {
	// Verify stream clusterInst
	streamClusterInst := NewStreamoutMsg(ctx)
	err := apis.streamObjApi.StreamClusterInst(key, streamClusterInst)
	if pass {
		require.Nil(t, err, "stream clusterinst")
		require.Greater(t, len(streamClusterInst.Msgs), 0, "contains stream messages")
	} else {
		require.NotNil(t, err, "stream clusterinst should return error")
	}
	return streamClusterInst.Msgs
}

func GetCloudletStreamMsgs(t *testing.T, ctx context.Context, key *edgeproto.CloudletKey, apis *AllApis) []edgeproto.Result {
	// Verify stream cloudlet
	streamCloudlet := NewStreamoutMsg(ctx)
	err := apis.streamObjApi.StreamCloudlet(key, streamCloudlet)
	require.Nil(t, err, "stream cloudlet")
	require.Greater(t, len(streamCloudlet.Msgs), 0, "contains stream messages")
	return streamCloudlet.Msgs
}

func TestAppInstApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testinit()
	defer testfinish()

	dummy := dummyEtcd{}
	dummy.Start()
	defer dummy.Stop()

	sync := InitSync(&dummy)
	apis := NewAllApis(sync)
	sync.Start()
	defer sync.Done()
	responder := NewDummyInfoResponder(&apis.appInstApi.cache, &apis.clusterInstApi.cache, apis.appInstInfoApi, apis.clusterInstInfoApi)

	reduceInfoTimeouts(t, ctx, apis)

	// cannote create instances without apps and cloudlets
	for _, data := range testutil.AppInstData {
		obj := data
		err := apis.appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.NotNil(t, err, "Create app inst without apps/cloudlets")
		// Verify stream AppInst fails
		GetAppInstStreamMsgs(t, ctx, &obj.Key, apis, Fail)
	}

	// create supporting data
	testutil.InternalFlavorCreate(t, apis.flavorApi, testutil.FlavorData)
	testutil.InternalGPUDriverCreate(t, apis.gpuDriverApi, testutil.GPUDriverData)
	testutil.InternalCloudletCreate(t, apis.cloudletApi, testutil.CloudletData())
	insertCloudletInfo(ctx, apis, testutil.CloudletInfoData)
	testutil.InternalAutoProvPolicyCreate(t, apis.autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalAutoScalePolicyCreate(t, apis.autoScalePolicyApi, testutil.AutoScalePolicyData)
	testutil.InternalAppCreate(t, apis.appApi, testutil.AppData)
	testutil.InternalClusterInstCreate(t, apis.clusterInstApi, testutil.ClusterInstData)
	testutil.InternalCloudletRefsTest(t, "show", apis.cloudletRefsApi, testutil.CloudletRefsData)
	clusterInstCnt := len(apis.clusterInstApi.cache.Objs)
	require.Equal(t, len(testutil.ClusterInstData), clusterInstCnt)

	// Set responder to fail. This should clean up the object after
	// the fake crm returns a failure. If it doesn't, the next test to
	// create all the app insts will fail.
	responder.SetSimulateAppCreateFailure(true)
	// clean up on failure may find ports inconsistent
	RequireAppInstPortConsistency = false
	for ii, data := range testutil.AppInstData {
		obj := data // make new copy since range variable gets reused each iter
		if testutil.IsAutoClusterAutoDeleteApp(&obj.Key) {
			continue
		}
		err := apis.appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.NotNil(t, err, "Create app inst responder failures")
		// make sure error matches responder
		// if app-inst triggers auto-cluster, the error will be for a cluster
		if strings.Contains(err.Error(), "cluster inst") {
			require.Equal(t, "Encountered failures: crm create cluster inst failed", err.Error(), "AppInst[%d]: %v", ii, obj.Key)
		} else {
			require.Equal(t, "Encountered failures: crm create app inst failed", err.Error(), "AppInst[%d]: %v", ii, obj.Key)
		}
		// As there was some progress, there should be some messages in stream
		msgs := GetAppInstStreamMsgs(t, ctx, &obj.Key, apis, Fail)
		require.Greater(t, len(msgs), 0, "some progress messages before failure")
	}
	responder.SetSimulateAppCreateFailure(false)
	RequireAppInstPortConsistency = true
	require.Equal(t, 0, len(apis.appInstApi.cache.Objs))
	require.Equal(t, clusterInstCnt, len(apis.clusterInstApi.cache.Objs))
	testutil.InternalCloudletRefsTest(t, "show", apis.cloudletRefsApi, testutil.CloudletRefsData)

	testutil.InternalAppInstTest(t, "cud", apis.appInstApi, testutil.AppInstData, testutil.WithCreatedAppInstTestData(testutil.CreatedAppInstData()))
	InternalAppInstCachedFieldsTest(t, ctx, apis)
	// check cluster insts created (includes explicit and auto)
	testutil.InternalClusterInstTest(t, "show", apis.clusterInstApi,
		append(testutil.ClusterInstData, testutil.ClusterInstAutoData...))
	require.Equal(t, len(testutil.ClusterInstData)+len(testutil.ClusterInstAutoData), len(apis.clusterInstApi.cache.Objs))

	// after app insts create, check that cloudlet refs data is correct.
	// Note this refs data is a second set after app insts were created.
	testutil.InternalCloudletRefsTest(t, "show", apis.cloudletRefsApi, testutil.CloudletRefsWithAppInstsData)
	testutil.InternalAppInstRefsTest(t, "show", apis.appInstRefsApi, testutil.GetAppInstRefsData())

	commonApi := testutil.NewInternalAppInstApi(apis.appInstApi)

	// Set responder to fail delete.
	responder.SetSimulateAppDeleteFailure(true)
	obj := testutil.AppInstData[0]
	err := apis.appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.NotNil(t, err, "Delete AppInst responder failure")
	responder.SetSimulateAppDeleteFailure(false)
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_READY)
	testutil.InternalAppInstRefsTest(t, "show", apis.appInstRefsApi, testutil.GetAppInstRefsData())
	// As there was some progress, there should be some messages in stream
	msgs := GetAppInstStreamMsgs(t, ctx, &obj.Key, apis, Fail)
	require.Greater(t, len(msgs), 0, "some progress messages before failure")

	obj = testutil.AppInstData[0]
	// check override of error DELETE_ERROR
	err = forceAppInstState(ctx, &obj, edgeproto.TrackedState_DELETE_ERROR, responder, apis)
	require.Nil(t, err, "force state")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_DELETE_ERROR)
	err = apis.appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "create overrides delete error")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_READY)
	testutil.InternalAppInstRefsTest(t, "show", apis.appInstRefsApi, testutil.GetAppInstRefsData())
	// As there was progress, there should be some messages in stream
	msgs = GetAppInstStreamMsgs(t, ctx, &obj.Key, apis, Pass)
	require.Greater(t, len(msgs), 0, "progress messages")

	// check override of error CREATE_ERROR
	err = forceAppInstState(ctx, &obj, edgeproto.TrackedState_CREATE_ERROR, responder, apis)
	require.Nil(t, err, "force state")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_CREATE_ERROR)
	err = apis.appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "delete overrides create error")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_NOT_PRESENT)
	// create copy of refs without the deleted AppInst
	appInstRefsDeleted := append([]edgeproto.AppInstRefs{}, testutil.GetAppInstRefsData()...)
	appInstRefsDeleted[0].Insts = make(map[string]uint32)
	for k, v := range testutil.GetAppInstRefsData()[0].Insts {
		if k == obj.Key.GetKeyString() {
			continue
		}
		appInstRefsDeleted[0].Insts[k] = v
	}
	testutil.InternalAppInstRefsTest(t, "show", apis.appInstRefsApi, appInstRefsDeleted)
	// As there was some progress, there should be some messages in stream
	msgs = GetAppInstStreamMsgs(t, ctx, &obj.Key, apis, Pass)
	require.Greater(t, len(msgs), 0, "some progress messages")

	// check override of error UPDATE_ERROR
	err = apis.appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "create appinst")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_READY)
	err = forceAppInstState(ctx, &obj, edgeproto.TrackedState_UPDATE_ERROR, responder, apis)
	require.Nil(t, err, "force state")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_UPDATE_ERROR)
	err = apis.appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "delete overrides create error")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_NOT_PRESENT)
	testutil.InternalAppInstRefsTest(t, "show", apis.appInstRefsApi, appInstRefsDeleted)

	// override CRM error
	responder.SetSimulateAppCreateFailure(true)
	responder.SetSimulateAppDeleteFailure(true)
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = apis.appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "override crm error")
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = apis.appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "override crm error")

	// ignore CRM
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
	err = apis.appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "ignore crm")
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
	err = apis.appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "ignore crm")
	responder.SetSimulateAppCreateFailure(false)
	responder.SetSimulateAppDeleteFailure(false)

	// ignore CRM and transient state on delete of AppInst
	responder.SetSimulateAppDeleteFailure(true)
	for val, stateName := range edgeproto.TrackedState_name {
		state := edgeproto.TrackedState(val)
		if !edgeproto.IsTransientState(state) {
			continue
		}
		obj = testutil.AppInstData[0]
		err = apis.appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "create AppInst")
		err = forceAppInstState(ctx, &obj, state, responder, apis)
		require.Nil(t, err, "force state")
		checkAppInstState(t, ctx, commonApi, &obj, state)
		obj = testutil.AppInstData[0]
		obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_AND_TRANSIENT_STATE
		err = apis.appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "override crm and transient state %s", stateName)
	}
	responder.SetSimulateAppDeleteFailure(false)

	testAppInstOverrideTransientDelete(t, ctx, commonApi, responder, apis)

	// Test Fqdn prefix
	for _, data := range apis.appInstApi.cache.Objs {
		obj := data.Obj
		app_name := util.K8SSanitize(obj.Key.AppKey.Name + obj.Key.AppKey.Version)
		if obj.Key.AppKey.Name == "helmApp" || obj.Key.AppKey.Name == "vm lb" {
			continue
		}
		cloudlet := edgeproto.Cloudlet{}
		found := apis.cloudletApi.cache.Get(&obj.Key.ClusterInstKey.CloudletKey, &cloudlet)
		require.True(t, found)
		features := platform.Features{}
		operator := obj.Key.ClusterInstKey.CloudletKey.Organization

		for _, port := range obj.MappedPorts {
			lproto, err := edgeproto.LProtoStr(port.Proto)
			if err != nil {
				continue
			}
			if lproto == "http" {
				continue
			}
			test_prefix := ""
			if isIPAllocatedPerService(ctx, cloudlet.PlatformType, &features, operator) {
				test_prefix = fmt.Sprintf("%s-%s-", util.DNSSanitize(app_name), lproto)
			}
			require.Equal(t, test_prefix, port.FqdnPrefix, "check port fqdn prefix")
		}
	}

	// test appint create with overlapping ports
	obj = testutil.AppInstData[0]
	obj.Key.AppKey = testutil.AppData[1].Key
	err = apis.appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.NotNil(t, err, "Overlapping ports would trigger an app inst create failure")
	require.Contains(t, err.Error(), "port 80 is already in use")

	// delete all AppInsts and Apps and check that refs are empty
	for ii, data := range testutil.AppInstData {
		obj := data
		if ii == 0 {
			// skip AppInst[0], it was deleted earlier in the test
			continue
		}
		if testutil.IsAutoClusterAutoDeleteApp(&obj.Key) {
			continue
		}
		err := apis.appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "Delete app inst %d failed", ii)
	}

	testAppFlavorRequest(t, ctx, commonApi, responder, apis)
	testDeprecatedSharedRootLBFQDN(t, ctx, apis)
	testSingleKubernetesCloudlet(t, ctx, apis)

	// cleanup unused reservable auto clusters
	apis.clusterInstApi.cleanupIdleReservableAutoClusters(ctx, time.Duration(0))
	apis.clusterInstApi.cleanupWorkers.WaitIdle()

	for ii, data := range testutil.AppData {
		obj := data
		_, err := apis.appApi.DeleteApp(ctx, &obj)
		require.Nil(t, err, "Delete app %d: %s failed", ii, obj.Key.GetKeyString())
	}
	testutil.InternalAppInstRefsTest(t, "show", apis.appInstRefsApi, []edgeproto.AppInstRefs{})
}

func appInstCachedFieldsTest(t *testing.T, ctx context.Context, cAppApi *testutil.AppCommonApi, cCloudletApi *testutil.CloudletCommonApi, cAppInstApi *testutil.AppInstCommonApi) {
	// test assumes test data has already been loaded

	// update app and check that app insts are updated
	updater := edgeproto.App{}
	updater.Key = testutil.AppData[0].Key
	newPath := "resources: a new config"
	updater.AndroidPackageName = newPath
	updater.Fields = make([]string, 0)
	updater.Fields = append(updater.Fields, edgeproto.AppFieldAndroidPackageName)
	_, err := cAppApi.UpdateApp(ctx, &updater)
	require.Nil(t, err, "Update app")

	show := testutil.ShowAppInst{}
	show.Init()
	filter := edgeproto.AppInst{}
	filter.Key.AppKey = testutil.AppData[0].Key
	err = cAppInstApi.ShowAppInst(ctx, &filter, &show)
	require.Nil(t, err, "show app inst data")
	require.True(t, len(show.Data) > 0, "number of matching app insts")

	// update cloudlet and check that app insts are updated
	updater2 := edgeproto.Cloudlet{}
	updater2.Key = testutil.CloudletData()[0].Key
	newLat := 52.84583
	updater2.Location.Latitude = newLat
	updater2.Fields = make([]string, 0)
	updater2.Fields = append(updater2.Fields, edgeproto.CloudletFieldLocationLatitude)
	_, err = cCloudletApi.UpdateCloudlet(ctx, &updater2)
	require.Nil(t, err, "Update cloudlet")

	show.Init()
	filter = edgeproto.AppInst{}
	filter.Key.ClusterInstKey.CloudletKey = testutil.CloudletData()[0].Key
	err = cAppInstApi.ShowAppInst(ctx, &filter, &show)
	require.Nil(t, err, "show app inst data")
	for _, inst := range show.Data {
		require.Equal(t, newLat, inst.CloudletLoc.Latitude, "check app inst latitude")
	}
	require.True(t, len(show.Data) > 0, "number of matching app insts")
}

func InternalAppInstCachedFieldsTest(t *testing.T, ctx context.Context, apis *AllApis) {
	cAppApi := testutil.NewInternalAppApi(apis.appApi)
	cCloudletApi := testutil.NewInternalCloudletApi(apis.cloudletApi)
	cAppInstApi := testutil.NewInternalAppInstApi(apis.appInstApi)
	appInstCachedFieldsTest(t, ctx, cAppApi, cCloudletApi, cAppInstApi)
}

func ClientAppInstCachedFieldsTest(t *testing.T, ctx context.Context, appApi edgeproto.AppApiClient, cloudletApi edgeproto.CloudletApiClient, appInstApi edgeproto.AppInstApiClient) {
	cAppApi := testutil.NewClientAppApi(appApi)
	cCloudletApi := testutil.NewClientCloudletApi(cloudletApi)
	cAppInstApi := testutil.NewClientAppInstApi(appInstApi)
	appInstCachedFieldsTest(t, ctx, cAppApi, cCloudletApi, cAppInstApi)
}

func TestAutoClusterInst(t *testing.T) {
	log.InitTracer(nil)
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testinit()
	defer testfinish()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	apis := NewAllApis(sync)
	sync.Start()
	defer sync.Done()
	NewDummyInfoResponder(&apis.appInstApi.cache, &apis.clusterInstApi.cache,
		apis.appInstInfoApi, apis.clusterInstInfoApi)

	reduceInfoTimeouts(t, ctx, apis)

	// create supporting data
	testutil.InternalFlavorCreate(t, apis.flavorApi, testutil.FlavorData)
	testutil.InternalGPUDriverCreate(t, apis.gpuDriverApi, testutil.GPUDriverData)
	testutil.InternalCloudletCreate(t, apis.cloudletApi, testutil.CloudletData())
	insertCloudletInfo(ctx, apis, testutil.CloudletInfoData)
	testutil.InternalAutoProvPolicyCreate(t, apis.autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalAppCreate(t, apis.appApi, testutil.AppData)
	// multi-tenant ClusterInst
	mt := testutil.ClusterInstData[8]
	require.True(t, mt.MultiTenant)
	// negative tests
	// bad Organization
	mtBad := mt
	mtBad.Key.Organization = "foo"
	err := apis.clusterInstApi.CreateClusterInst(&mtBad, testutil.NewCudStreamoutClusterInst(ctx))
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Only MobiledgeX ClusterInsts may be multi-tenant")
	// bad deployment type
	mtBad = mt
	mtBad.Deployment = cloudcommon.DeploymentTypeDocker
	err = apis.clusterInstApi.CreateClusterInst(&mtBad, testutil.NewCudStreamoutClusterInst(ctx))
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Multi-tenant clusters must be of deployment type Kubernetes")

	// Create multi-tenant ClusterInst before reservable tests.
	// Reservable tests should pass without using multi-tenant because
	// the App's SupportMultiTenant is false.
	err = apis.clusterInstApi.CreateClusterInst(&mt, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err)

	checkReserved := func(cloudletKey edgeproto.CloudletKey, found bool, id, reservedBy string) {
		key := &edgeproto.ClusterInstKey{}
		key.ClusterKey.Name = cloudcommon.ReservableClusterPrefix + id
		key.CloudletKey = cloudletKey
		key.Organization = cloudcommon.OrganizationMobiledgeX
		// look up reserved ClusterInst
		clusterInst := edgeproto.ClusterInst{}
		actualFound := apis.clusterInstApi.Get(key, &clusterInst)
		require.Equal(t, found, actualFound, "lookup %s", key.GetKeyString())
		if !found {
			return
		}
		require.True(t, clusterInst.Auto, "clusterinst is auto")
		require.True(t, clusterInst.Reservable, "clusterinst is reservable")
		require.Equal(t, reservedBy, clusterInst.ReservedBy, "reserved by matches")
		// Progress message should be there for cluster instance itself
		msgs := GetClusterInstStreamMsgs(t, ctx, key, apis, Pass)
		require.Greater(t, len(msgs), 0, "some progress messages")
	}
	createAutoClusterAppInst := func(copy edgeproto.AppInst, expectedId string) {
		// since cluster inst does not exist, it will be auto-created
		copy.Key.ClusterInstKey.ClusterKey.Name = cloudcommon.AutoClusterPrefix + expectedId
		copy.Key.ClusterInstKey.Organization = cloudcommon.OrganizationMobiledgeX
		err := apis.appInstApi.CreateAppInst(&copy, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "create app inst")
		// As there was some progress, there should be some messages in stream
		msgs := GetAppInstStreamMsgs(t, ctx, &copy.Key, apis, Pass)
		require.Greater(t, len(msgs), 0, "some progress messages")
		// Check that reserved ClusterInst was created
		checkReserved(copy.Key.ClusterInstKey.CloudletKey, true, expectedId, copy.Key.AppKey.Organization)
		// check for expected cluster name.
		require.Equal(t, cloudcommon.AutoClusterPrefix+expectedId, copy.Key.ClusterInstKey.ClusterKey.Name)
	}
	deleteAutoClusterAppInst := func(copy edgeproto.AppInst, id string) {
		// delete appinst
		copy.Key.ClusterInstKey.ClusterKey.Name = cloudcommon.AutoClusterPrefix + id
		copy.Key.ClusterInstKey.Organization = cloudcommon.OrganizationMobiledgeX
		err := apis.appInstApi.DeleteAppInst(&copy, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "delete app inst")
		checkReserved(copy.Key.ClusterInstKey.CloudletKey, true, id, "")
	}
	checkReservedIds := func(key edgeproto.CloudletKey, expected uint64) {
		refs := edgeproto.CloudletRefs{}
		found := apis.cloudletRefsApi.cache.Get(&key, &refs)
		require.True(t, found)
		require.Equal(t, expected, refs.ReservedAutoClusterIds)
	}

	// create auto-cluster AppInsts
	appInst := testutil.AppInstData[0]
	appInst.Key.AppKey = testutil.AppData[1].Key // does not support multi-tenant
	cloudletKey := appInst.Key.ClusterInstKey.CloudletKey
	createAutoClusterAppInst(appInst, "0")
	checkReservedIds(cloudletKey, 1)
	createAutoClusterAppInst(appInst, "1")
	checkReservedIds(cloudletKey, 3)
	createAutoClusterAppInst(appInst, "2")
	checkReservedIds(cloudletKey, 7)
	// delete one
	deleteAutoClusterAppInst(appInst, "1")
	checkReservedIds(cloudletKey, 7) // clusterinst doesn't get deleted
	// create again, should reuse existing free ClusterInst
	createAutoClusterAppInst(appInst, "1")
	checkReservedIds(cloudletKey, 7)
	// delete one again
	deleteAutoClusterAppInst(appInst, "1")
	checkReservedIds(cloudletKey, 7) // clusterinst doesn't get deleted
	// cleanup unused reservable auto clusters
	apis.clusterInstApi.cleanupIdleReservableAutoClusters(ctx, time.Duration(0))
	apis.clusterInstApi.cleanupWorkers.WaitIdle()
	checkReserved(cloudletKey, false, "1", "")
	checkReservedIds(cloudletKey, 5)
	// create again, should create new ClusterInst with next free id
	createAutoClusterAppInst(appInst, "1")
	checkReservedIds(cloudletKey, 7)
	// delete all of them
	deleteAutoClusterAppInst(appInst, "0")
	deleteAutoClusterAppInst(appInst, "1")
	deleteAutoClusterAppInst(appInst, "2")
	checkReservedIds(cloudletKey, 7)
	// cleanup unused reservable auto clusters
	apis.clusterInstApi.cleanupIdleReservableAutoClusters(ctx, time.Duration(0))
	apis.clusterInstApi.cleanupWorkers.WaitIdle()
	checkReserved(cloudletKey, false, "0", "")
	checkReserved(cloudletKey, false, "1", "")
	checkReserved(cloudletKey, false, "2", "")
	checkReservedIds(cloudletKey, 0)

	// Autocluster AppInst with AutoDelete delete option should fail
	autoDeleteAppInst := testutil.AppInstData[10]
	autoDeleteAppInst.RealClusterName = ""
	autoDeleteAppInst.Key.ClusterInstKey.ClusterKey.Name = cloudcommon.AutoClusterPrefix + "foo"
	err = apis.appInstApi.CreateAppInst(&autoDeleteAppInst, testutil.NewCudStreamoutAppInst(ctx))
	require.NotNil(t, err, "create autodelete appInst")
	require.Contains(t, err.Error(), "Sidecar AppInst (AutoDelete App) must specify the RealClusterName field to deploy to the virtual cluster")

	err = apis.clusterInstApi.DeleteClusterInst(&mt, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err)

	testDeprecatedAutoCluster(t, ctx, apis)
	dummy.Stop()
}

func testDeprecatedAutoCluster(t *testing.T, ctx context.Context, apis *AllApis) {
	// downgrade cloudlets to older crm compatibility version
	cloudletInfos := []edgeproto.CloudletInfo{}
	for _, info := range testutil.CloudletInfoData {
		info.CompatibilityVersion = 0
		cloudletInfos = append(cloudletInfos, info)
	}
	insertCloudletInfo(ctx, apis, cloudletInfos)

	// existing AppInst creates should fail because the ClusterInst org
	// must match the AppInst org.
	for ii, data := range testutil.AppInstData {
		obj := data
		if !strings.HasPrefix(obj.Key.ClusterInstKey.ClusterKey.Name, cloudcommon.AutoClusterPrefix) {
			continue
		}
		// there is one App that has dev org as MobiledgeX, which matches
		// the Reservable auto-cluster's org, so skip it.
		if obj.Key.AppKey.Organization == obj.Key.ClusterInstKey.Organization {
			continue
		}
		err := apis.appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.NotNil(t, err, "Create autocluster appinst[%d]: %v", ii, obj.Key)
		require.Contains(t, err.Error(), "Developer name mismatch")
	}

	appInsts := []edgeproto.AppInst{}
	clusterInsts := []edgeproto.ClusterInst{}
	for ii, data := range testutil.AppInstData {
		obj := data
		if !strings.HasPrefix(obj.Key.ClusterInstKey.ClusterKey.Name, cloudcommon.AutoClusterPrefix) {
			continue
		}
		obj.Key.ClusterInstKey.Organization = obj.Key.AppKey.Organization
		appInsts = append(appInsts, obj)

		err := apis.appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "Create autocluster appinst[%d]: %v", ii, obj.Key)
		// find app for AppInst
		var app *edgeproto.App
		for _, a := range testutil.AppData {
			if a.Key.Matches(&obj.Key.AppKey) {
				app = &a
				break
			}
		}
		require.NotNil(t, app, "find App for AppInst")
		// build expected ClusterInst
		cinst := edgeproto.ClusterInst{}
		cinst.Key = *obj.Key.ClusterInstKey.Real("")
		cinst.Flavor = app.DefaultFlavor
		cinst.NumMasters = 1
		cinst.NumNodes = 1
		cinst.Auto = true
		cinst.State = edgeproto.TrackedState_READY
		clusterInsts = append(clusterInsts, cinst)
	}
	testutil.InternalClusterInstTest(t, "show", apis.clusterInstApi, clusterInsts)
	// delete AppInsts should delete ClusterInsts
	for _, obj := range appInsts {
		err := apis.appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "Delete autocluster appinst %v", obj.Key)
	}
	testutil.InternalClusterInstTest(t, "show", apis.clusterInstApi, []edgeproto.ClusterInst{})
	// restore CRM version
	for ii := range cloudletInfos {
		cloudletInfos[ii].CompatibilityVersion = cloudcommon.GetCRMCompatibilityVersion()
	}
	insertCloudletInfo(ctx, apis, cloudletInfos)
}

func testDeprecatedSharedRootLBFQDN(t *testing.T, ctx context.Context, apis *AllApis) {
	// downgrade cloudlets to older crm compatibility version
	cloudletInfos := []edgeproto.CloudletInfo{}
	for _, info := range testutil.CloudletInfoData {
		info.CompatibilityVersion = 0
		cloudletInfos = append(cloudletInfos, info)
	}
	insertCloudletInfo(ctx, apis, cloudletInfos)

	// create appInst with internalports set to true, this means
	// that this appInst will not need access from external network
	appInstData := testutil.AppInstData[9]
	err := apis.appInstApi.CreateAppInst(&appInstData, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "Create appinst: %v", appInstData.Key)

	show := testutil.ShowAppInst{}
	show.Init()
	filter := edgeproto.AppInst{}
	filter.Key = appInstData.Key

	// For older version of CRM, URI should exist
	err = apis.appInstApi.ShowAppInst(&filter, &show)
	require.Nil(t, err, "show app inst data")
	require.True(t, len(show.Data) == 1, "matches app inst")
	for _, appInstOut := range show.Data {
		require.True(t, appInstOut.Uri != "", "app URI exists")
	}

	// delete appinst
	err = apis.appInstApi.DeleteAppInst(&appInstData, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "Delete appinst: %v", appInstData.Key)

	// restore CRM version
	for ii := range cloudletInfos {
		cloudletInfos[ii].CompatibilityVersion = cloudcommon.GetCRMCompatibilityVersion()
	}
	insertCloudletInfo(ctx, apis, cloudletInfos)

	// create appinst again
	err = apis.appInstApi.CreateAppInst(&appInstData, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "Create appinst: %v", appInstData.Key)

	// For newer version of CRM, URI should not exist
	err = apis.appInstApi.ShowAppInst(&filter, &show)
	require.Nil(t, err, "show app inst data")
	require.True(t, len(show.Data) == 1, "matches app inst")
	for _, appInstOut := range show.Data {
		require.True(t, appInstOut.Uri == "", "app URI doesn't exist")
	}

	// final cleanup: delete appinst
	err = apis.appInstApi.DeleteAppInst(&appInstData, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "Delete appinst: %v", appInstData.Key)
}

func checkAppInstState(t *testing.T, ctx context.Context, api *testutil.AppInstCommonApi, in *edgeproto.AppInst, state edgeproto.TrackedState) {
	out := edgeproto.AppInst{}
	found := testutil.GetAppInst(t, ctx, api, &in.Key, &out)
	if state == edgeproto.TrackedState_NOT_PRESENT {
		require.False(t, found, "get app inst")
	} else {
		require.True(t, found, "get app inst")
		require.Equal(t, state, out.State, "app inst state")
	}
}

func forceAppInstState(ctx context.Context, in *edgeproto.AppInst, state edgeproto.TrackedState, responder *DummyInfoResponder, apis *AllApis) error {
	if responder != nil {
		// disable responder, otherwise it will respond to certain states
		// and change the current state
		responder.enable = false
		defer func() {
			responder.enable = true
		}()
	}
	err := apis.appInstApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		obj := edgeproto.AppInst{}
		if !apis.appInstApi.store.STMGet(stm, &in.Key, &obj) {
			return in.Key.NotFoundError()
		}
		obj.State = state
		apis.appInstApi.store.STMPut(stm, &obj)
		return nil
	})
	return err
}

func testAppFlavorRequest(t *testing.T, ctx context.Context, api *testutil.AppInstCommonApi, responder *DummyInfoResponder, apis *AllApis) {
	// Non-nomial test, request an optional resource from a cloudlet that offers none.
	var testflavor = edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "x1.large-mex",
		},
		Ram:       8192,
		Vcpus:     8,
		Disk:      40,
		OptResMap: map[string]string{"gpu": "gpu:1"},
	}
	_, err := apis.flavorApi.CreateFlavor(ctx, &testflavor)
	require.Nil(t, err, "CreateFlavor")
	nonNomApp := testutil.AppInstData[2]
	nonNomApp.Flavor = testflavor.Key
	err = apis.appInstApi.CreateAppInst(&nonNomApp, testutil.NewCudStreamoutAppInst(ctx))
	require.NotNil(t, err, "non-nom-app-create")
	require.Equal(t, "Cloudlet New York Site doesn't support GPU", err.Error())
}

// Test that Crm Override for Delete App overrides any failures
// on both side-car auto-apps and an underlying auto-cluster.
func testAppInstOverrideTransientDelete(t *testing.T, ctx context.Context, api *testutil.AppInstCommonApi, responder *DummyInfoResponder, apis *AllApis) {
	// autocluster app
	ai := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey: testutil.AppData[0].Key,
			ClusterInstKey: edgeproto.VirtualClusterInstKey{
				ClusterKey: edgeproto.ClusterKey{
					Name: util.K8SSanitize(cloudcommon.AutoClusterPrefix + "override-clust"),
				},
				CloudletKey:  testutil.CloudletData()[1].Key,
				Organization: cloudcommon.OrganizationMobiledgeX,
			},
		},
	}
	// autoapp
	require.Equal(t, edgeproto.DeleteType_AUTO_DELETE, testutil.AppData[9].DelOpt)
	aiauto := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:         testutil.AppData[9].Key, // auto-delete app
			ClusterInstKey: ai.Key.ClusterInstKey,
		},
	}
	var err error
	var obj edgeproto.AppInst
	var clust edgeproto.ClusterInst
	clustApi := testutil.NewInternalClusterInstApi(apis.clusterInstApi)

	responder.SetSimulateAppDeleteFailure(true)
	responder.SetSimulateClusterDeleteFailure(true)
	for val, stateName := range edgeproto.TrackedState_name {
		state := edgeproto.TrackedState(val)
		if !edgeproto.IsTransientState(state) {
			continue
		}
		// create app (also creates clusterinst)
		obj = ai
		err = apis.appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "create AppInst")
		err = forceAppInstState(ctx, &obj, state, responder, apis)
		require.Nil(t, err, "force state")
		checkAppInstState(t, ctx, api, &obj, state)

		// set aiauto cluster name from real cluster name of create autocluster
		clKey := obj.ClusterInstKey()
		obj = aiauto
		obj.Key.ClusterInstKey = *clKey.Virtual("")
		// create auto app
		err = apis.appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "create AppInst on cluster %v", obj.Key.ClusterInstKey)
		err = forceAppInstState(ctx, &obj, state, responder, apis)
		require.Nil(t, err, "force state")
		checkAppInstState(t, ctx, api, &obj, state)

		clust = edgeproto.ClusterInst{}
		clust.Key = *clKey
		err = forceClusterInstState(ctx, &clust, state, responder, apis)
		require.Nil(t, err, "force state")
		checkClusterInstState(t, ctx, clustApi, &clust, state)

		// delete app (to be able to delete reservable cluster)
		obj = ai
		obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_AND_TRANSIENT_STATE
		log.SpanLog(ctx, log.DebugLevelInfo, "test run appinst delete")
		err = apis.appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "override crm and transient state %s", stateName)
		log.SpanLog(ctx, log.DebugLevelInfo, "test appinst deleted")

		// delete cluster (should also delete auto app)
		clust.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_AND_TRANSIENT_STATE
		log.SpanLog(ctx, log.DebugLevelInfo, "test run ClusterInst delete")
		err = apis.clusterInstApi.DeleteClusterInst(&clust, testutil.NewCudStreamoutClusterInst(ctx))
		require.Nil(t, err, "override crm and transient state %s", stateName)
		log.SpanLog(ctx, log.DebugLevelInfo, "test ClusterInst deleted")
		// make sure cluster got deleted (means apps also were deleted)
		found := testutil.GetClusterInst(t, ctx, clustApi, &clust.Key, &edgeproto.ClusterInst{})
		require.False(t, found)
	}

	responder.SetSimulateAppDeleteFailure(false)
	responder.SetSimulateClusterDeleteFailure(false)

}

func testSingleKubernetesCloudlet(t *testing.T, ctx context.Context, apis *AllApis) {
	var err error
	var found bool
	// Single kubernetes cloudlets can be either multi-tenant,
	// or dedicated to a particular organization (which removes all
	// of the multi-tenant deployment restrictions on namespaces, etc).
	cloudletMT := edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			Organization: "unittest",
			Name:         "singlek8sMT",
		},
		IpSupport:     edgeproto.IpSupport_IP_SUPPORT_DYNAMIC,
		NumDynamicIps: 10,
		Location: dme.Loc{
			Latitude:  37.1231,
			Longitude: 94.123,
		},
		PlatformType: edgeproto.PlatformType_PLATFORM_TYPE_FAKE_SINGLE_CLUSTER,
		CrmOverride:  edgeproto.CRMOverride_IGNORE_CRM,
	}
	cloudletMTInfo := edgeproto.CloudletInfo{
		Key:                  cloudletMT.Key,
		State:                dme.CloudletState_CLOUDLET_STATE_READY,
		CompatibilityVersion: 1, // cloudcommon.GetCRMCompatibilityVersion()
	}
	mtOrg := cloudcommon.OrganizationMobiledgeX
	stOrg := testutil.AppInstData[0].Key.AppKey.Organization

	cloudletST := cloudletMT
	cloudletST.Key.Name = "singlek8sST"
	cloudletST.SingleKubernetesClusterOwner = stOrg
	cloudletSTInfo := cloudletMTInfo
	cloudletSTInfo.Key = cloudletST.Key

	setupTests := []struct {
		desc         string
		cloudlet     *edgeproto.Cloudlet
		cloudletInfo *edgeproto.CloudletInfo
		ownerOrg     string
		mt           bool
	}{
		{"mt setup", &cloudletMT, &cloudletMTInfo, "", true},
		{"st setup", &cloudletST, &cloudletSTInfo, stOrg, false},
	}
	for _, test := range setupTests {
		clusterInstKey := getDefaultClustKey(test.cloudlet.Key, test.ownerOrg)
		// create cloudlet
		err = apis.cloudletApi.CreateCloudlet(test.cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
		require.Nil(t, err, test.desc)
		apis.cloudletInfoApi.Update(ctx, test.cloudletInfo, 0)
		// creating cloudlet also creates singleton cluster for cloudlet
		clusterInst := edgeproto.ClusterInst{}
		found = apis.clusterInstApi.Get(clusterInstKey, &clusterInst)
		require.True(t, found, test.desc)
		require.Equal(t, clusterInst.MultiTenant, test.mt)

		// trying to create clusterinst against cloudlets should fail
		tryClusterInst := edgeproto.ClusterInst{
			Key: edgeproto.ClusterInstKey{
				ClusterKey: edgeproto.ClusterKey{
					Name: "someclust",
				},
				Organization: "foo",
			},
			Flavor:     testutil.FlavorData[0].Key,
			NumMasters: 1,
			NumNodes:   2,
		}
		tryClusterInst.Key.CloudletKey = test.cloudlet.Key
		err = apis.clusterInstApi.CreateClusterInst(&tryClusterInst, testutil.NewCudStreamoutClusterInst(ctx))
		require.NotNil(t, err, test.desc)
		require.Contains(t, err.Error(), "only supports AppInst creates", test.desc)
	}

	// AppInst negative and positive tests
	// TODO: resource allocation failure...
	PASS := "PASS"
	appInstCreateTests := []struct {
		desc            string
		aiIdx           int
		cloudlet        *edgeproto.Cloudlet
		clusterName     string
		clusterOrg      string
		realClusterName string
		errStr          string
	}{{
		"MT non-serverless app",
		3, &cloudletMT, "clust", mtOrg, "",
		"Target cloudlet platform only supports serverless Apps",
	}, {
		"MT bad cluster org",
		0, &cloudletMT, "clust", "foo", "",
		"ClusterInst organization must be set to MobiledgeX",
	}, {
		"MT bad real cluster name",
		0, &cloudletMT, "autocluster", mtOrg, "foo",
		"Invalid RealClusterName for single kubernetes cluster cloudlet, should be left blank",
	}, {
		"ST bad cluster org", 0,
		&cloudletST, "clust", "foo", "",
		"ClusterInst organization must be set to " + stOrg,
	}, {
		"ST bad real cluster name",
		0, &cloudletST, "autocluster", stOrg, "foo",
		"Invalid RealClusterName for single kubernetes cluster cloudlet, should be left blank",
	}, {
		"MT any clust name",
		0, &cloudletMT, "clust", mtOrg, "", PASS,
	}, {
		"MT auto clust name",
		0, &cloudletMT, "autocluster", mtOrg, "", PASS,
	}, {
		"ST any clust name",
		0, &cloudletST, "clust", stOrg, "", PASS,
	}, {
		"ST auto clust name",
		0, &cloudletST, "autocluster", stOrg, "", PASS,
	}}
	for _, test := range appInstCreateTests {
		ai := testutil.AppInstData[test.aiIdx]
		ai.Key.ClusterInstKey.CloudletKey = test.cloudlet.Key
		ai.Key.ClusterInstKey.ClusterKey.Name = test.clusterName
		ai.Key.ClusterInstKey.Organization = test.clusterOrg
		ai.RealClusterName = test.realClusterName
		err = apis.appInstApi.CreateAppInst(&ai, testutil.NewCudStreamoutAppInst(ctx))
		if test.errStr == PASS {
			require.Nil(t, err, test.desc)
			// clean up
			err = apis.appInstApi.DeleteAppInst(&ai, testutil.NewCudStreamoutAppInst(ctx))
			require.Nil(t, err, test.desc)
		} else {
			require.NotNil(t, err, test.desc)
			require.Contains(t, err.Error(), test.errStr, test.desc)
		}
	}

	cleanupTests := []struct {
		desc     string
		cloudlet *edgeproto.Cloudlet
		ownerOrg string
	}{
		{"mt cleanup test", &cloudletMT, ""},
		{"st cleanup test", &cloudletST, stOrg},
	}
	for _, test := range cleanupTests {
		clusterInstKey := getDefaultClustKey(test.cloudlet.Key, test.ownerOrg)
		// Create AppInst
		ai := testutil.AppInstData[0]
		ai.Key.ClusterInstKey.CloudletKey = test.cloudlet.Key
		ai.Key.ClusterInstKey.ClusterKey.Name = "blocker"
		ai.Key.ClusterInstKey.Organization = clusterInstKey.Organization
		ai.RealClusterName = ""
		err = apis.appInstApi.CreateAppInst(&ai, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, test.desc)

		// check refs
		refs := edgeproto.ClusterRefs{}
		found = apis.clusterRefsApi.cache.Get(clusterInstKey, &refs)
		require.True(t, found, test.desc)
		require.Equal(t, 1, len(refs.Apps), test.desc)
		refAiKey := edgeproto.AppInstKey{}
		refAiKey.FromClusterRefsAppInstKey(&refs.Apps[0], clusterInstKey)
		require.Equal(t, ai.Key, refAiKey, test.desc)

		// Test that delete cloudlet fails if AppInst exists
		test.cloudlet.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
		err = apis.cloudletApi.DeleteCloudlet(test.cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
		require.NotNil(t, err, test.desc)
		require.Contains(t, err.Error(), "Cloudlet in use by AppInst", test.desc)

		// delete AppInst
		err = apis.appInstApi.DeleteAppInst(&ai, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, test.desc)

		// now delete cloudlet should succeed
		test.cloudlet.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
		err = apis.cloudletApi.DeleteCloudlet(test.cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
		require.Nil(t, err, test.desc)

		// check that cluster and refs don't exist
		clusterInst := edgeproto.ClusterInst{}
		found = apis.clusterInstApi.Get(clusterInstKey, &clusterInst)
		require.False(t, found, test.desc)
		found = apis.clusterRefsApi.cache.Get(clusterInstKey, &refs)
		require.False(t, found, test.desc)
	}
}
