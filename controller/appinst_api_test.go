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

func GetAppInstStreamMsgs(t *testing.T, ctx context.Context, key *edgeproto.AppInstKey, pass bool) []edgeproto.Result {
	// Verify stream appInst
	streamAppInst := NewStreamoutMsg(ctx)
	err := streamObjApi.StreamAppInst(key, streamAppInst)
	if pass {
		require.Nil(t, err, "stream appinst")
		require.Greater(t, len(streamAppInst.Msgs), 0, "contains stream messages")
	} else {
		require.NotNil(t, err, "stream appinst should return error for key %s", *key)
	}
	return streamAppInst.Msgs
}

func GetClusterInstStreamMsgs(t *testing.T, ctx context.Context, key *edgeproto.ClusterInstKey, pass bool) []edgeproto.Result {
	// Verify stream clusterInst
	streamClusterInst := NewStreamoutMsg(ctx)
	err := streamObjApi.StreamClusterInst(key, streamClusterInst)
	if pass {
		require.Nil(t, err, "stream clusterinst")
		require.Greater(t, len(streamClusterInst.Msgs), 0, "contains stream messages")
	} else {
		require.NotNil(t, err, "stream clusterinst should return error")
	}
	return streamClusterInst.Msgs
}

func GetCloudletStreamMsgs(t *testing.T, ctx context.Context, key *edgeproto.CloudletKey) []edgeproto.Result {
	// Verify stream cloudlet
	streamCloudlet := NewStreamoutMsg(ctx)
	err := streamObjApi.StreamCloudlet(key, streamCloudlet)
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

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()
	responder := NewDummyInfoResponder(&appInstApi.cache, &clusterInstApi.cache,
		&appInstInfoApi, &clusterInstInfoApi)

	reduceInfoTimeouts(t, ctx)

	InfluxUsageUnitTestSetup(t)
	defer InfluxUsageUnitTestStop()

	// cannote create instances without apps and cloudlets
	for _, obj := range testutil.AppInstData {
		err := appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.NotNil(t, err, "Create app inst without apps/cloudlets")
		// Verify stream AppInst fails
		GetAppInstStreamMsgs(t, ctx, &obj.Key, Fail)
	}

	// create supporting data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalGPUDriverCreate(t, &gpuDriverApi, testutil.GPUDriverData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData())
	insertCloudletInfo(ctx, testutil.CloudletInfoData)
	testutil.InternalAutoProvPolicyCreate(t, &autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalAutoScalePolicyCreate(t, &autoScalePolicyApi, testutil.AutoScalePolicyData)
	testutil.InternalAppCreate(t, &appApi, testutil.AppData)
	testutil.InternalClusterInstCreate(t, &clusterInstApi, testutil.ClusterInstData)
	testutil.InternalCloudletRefsTest(t, "show", &cloudletRefsApi, testutil.CloudletRefsData)
	clusterInstCnt := len(clusterInstApi.cache.Objs)
	require.Equal(t, len(testutil.ClusterInstData), clusterInstCnt)

	// Set responder to fail. This should clean up the object after
	// the fake crm returns a failure. If it doesn't, the next test to
	// create all the app insts will fail.
	responder.SetSimulateAppCreateFailure(true)
	// clean up on failure may find ports inconsistent
	RequireAppInstPortConsistency = false
	for ii, obj := range testutil.AppInstData {
		if testutil.IsAutoClusterAutoDeleteApp(&obj.Key) {
			continue
		}
		err := appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.NotNil(t, err, "Create app inst responder failures")
		// make sure error matches responder
		// if app-inst triggers auto-cluster, the error will be for a cluster
		if strings.Contains(err.Error(), "cluster inst") {
			require.Equal(t, "Encountered failures: crm create cluster inst failed", err.Error(), "AppInst[%d]: %v", ii, obj.Key)
		} else {
			require.Equal(t, "Encountered failures: crm create app inst failed", err.Error(), "AppInst[%d]: %v", ii, obj.Key)
		}
		// As there was some progress, there should be some messages in stream
		msgs := GetAppInstStreamMsgs(t, ctx, &obj.Key, Fail)
		require.Greater(t, len(msgs), 0, "some progress messages before failure")
	}
	responder.SetSimulateAppCreateFailure(false)
	RequireAppInstPortConsistency = true
	require.Equal(t, 0, len(appInstApi.cache.Objs))
	require.Equal(t, clusterInstCnt, len(clusterInstApi.cache.Objs))
	testutil.InternalCloudletRefsTest(t, "show", &cloudletRefsApi, testutil.CloudletRefsData)

	testutil.InternalAppInstTest(t, "cud", &appInstApi, testutil.AppInstData, testutil.WithCreatedAppInstTestData(testutil.CreatedAppInstData()))
	InternalAppInstCachedFieldsTest(t, ctx)
	// check cluster insts created (includes explicit and auto)
	testutil.InternalClusterInstTest(t, "show", &clusterInstApi,
		append(testutil.ClusterInstData, testutil.ClusterInstAutoData...))
	require.Equal(t, len(testutil.ClusterInstData)+len(testutil.ClusterInstAutoData), len(clusterInstApi.cache.Objs))

	// after app insts create, check that cloudlet refs data is correct.
	// Note this refs data is a second set after app insts were created.
	testutil.InternalCloudletRefsTest(t, "show", &cloudletRefsApi, testutil.CloudletRefsWithAppInstsData)
	testutil.InternalAppInstRefsTest(t, "show", &appInstRefsApi, testutil.GetAppInstRefsData())

	commonApi := testutil.NewInternalAppInstApi(&appInstApi)

	// Set responder to fail delete.
	responder.SetSimulateAppDeleteFailure(true)
	obj := testutil.AppInstData[0]
	err := appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.NotNil(t, err, "Delete AppInst responder failure")
	responder.SetSimulateAppDeleteFailure(false)
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_READY)
	testutil.InternalAppInstRefsTest(t, "show", &appInstRefsApi, testutil.GetAppInstRefsData())
	// As there was some progress, there should be some messages in stream
	msgs := GetAppInstStreamMsgs(t, ctx, &obj.Key, Fail)
	require.Greater(t, len(msgs), 0, "some progress messages before failure")

	obj = testutil.AppInstData[0]
	// check override of error DELETE_ERROR
	err = forceAppInstState(ctx, &obj, edgeproto.TrackedState_DELETE_ERROR, responder)
	require.Nil(t, err, "force state")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_DELETE_ERROR)
	err = appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "create overrides delete error")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_READY)
	testutil.InternalAppInstRefsTest(t, "show", &appInstRefsApi, testutil.GetAppInstRefsData())
	// As there was progress, there should be some messages in stream
	msgs = GetAppInstStreamMsgs(t, ctx, &obj.Key, Pass)
	require.Greater(t, len(msgs), 0, "progress messages")

	// check override of error CREATE_ERROR
	err = forceAppInstState(ctx, &obj, edgeproto.TrackedState_CREATE_ERROR, responder)
	require.Nil(t, err, "force state")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_CREATE_ERROR)
	err = appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
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
	testutil.InternalAppInstRefsTest(t, "show", &appInstRefsApi, appInstRefsDeleted)
	// As there was some progress, there should be some messages in stream
	msgs = GetAppInstStreamMsgs(t, ctx, &obj.Key, Pass)
	require.Greater(t, len(msgs), 0, "some progress messages")

	// check override of error UPDATE_ERROR
	err = appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "create appinst")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_READY)
	err = forceAppInstState(ctx, &obj, edgeproto.TrackedState_UPDATE_ERROR, responder)
	require.Nil(t, err, "force state")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_UPDATE_ERROR)
	err = appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "delete overrides create error")
	checkAppInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_NOT_PRESENT)
	testutil.InternalAppInstRefsTest(t, "show", &appInstRefsApi, appInstRefsDeleted)

	// override CRM error
	responder.SetSimulateAppCreateFailure(true)
	responder.SetSimulateAppDeleteFailure(true)
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "override crm error")
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "override crm error")

	// ignore CRM
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
	err = appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "ignore crm")
	obj = testutil.AppInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
	err = appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
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
		err = appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "create AppInst")
		err = forceAppInstState(ctx, &obj, state, responder)
		require.Nil(t, err, "force state")
		checkAppInstState(t, ctx, commonApi, &obj, state)
		obj = testutil.AppInstData[0]
		obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_AND_TRANSIENT_STATE
		err = appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "override crm and transient state %s", stateName)
	}
	responder.SetSimulateAppDeleteFailure(false)

	testAppInstOverrideTransientDelete(t, ctx, commonApi, responder)

	// Test Fqdn prefix
	for _, data := range appInstApi.cache.Objs {
		obj := data.Obj
		app_name := util.K8SSanitize(obj.Key.AppKey.Name + obj.Key.AppKey.Version)
		if obj.Key.AppKey.Name == "helmApp" || obj.Key.AppKey.Name == "vm lb" {
			continue
		}
		cloudlet := edgeproto.Cloudlet{}
		found := cloudletApi.cache.Get(&obj.Key.ClusterInstKey.CloudletKey, &cloudlet)
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

	// delete all AppInsts and Apps and check that refs are empty
	for ii, obj := range testutil.AppInstData {
		if ii == 0 {
			// skip AppInst[0], it was deleted earlier in the test
			continue
		}
		if testutil.IsAutoClusterAutoDeleteApp(&obj.Key) {
			continue
		}
		err := appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "Delete app inst %d failed", ii)
	}

	testAppFlavorRequest(t, ctx, commonApi, responder)
	testDeprecatedSharedRootLBFQDN(t, ctx)

	// cleanup unused reservable auto clusters
	clusterInstApi.cleanupIdleReservableAutoClusters(ctx, time.Duration(0))
	clusterInstApi.cleanupWorkers.WaitIdle()

	for ii, obj := range testutil.AppData {
		_, err := appApi.DeleteApp(ctx, &obj)
		require.Nil(t, err, "Delete app %d: %s failed", ii, obj.Key.GetKeyString())
	}
	testutil.InternalAppInstRefsTest(t, "show", &appInstRefsApi, []edgeproto.AppInstRefs{})

	dummy.Stop()
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

func InternalAppInstCachedFieldsTest(t *testing.T, ctx context.Context) {
	cAppApi := testutil.NewInternalAppApi(&appApi)
	cCloudletApi := testutil.NewInternalCloudletApi(&cloudletApi)
	cAppInstApi := testutil.NewInternalAppInstApi(&appInstApi)
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

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()
	NewDummyInfoResponder(&appInstApi.cache, &clusterInstApi.cache,
		&appInstInfoApi, &clusterInstInfoApi)

	reduceInfoTimeouts(t, ctx)
	InfluxUsageUnitTestSetup(t)
	defer InfluxUsageUnitTestStop()

	// create supporting data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalGPUDriverCreate(t, &gpuDriverApi, testutil.GPUDriverData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData())
	insertCloudletInfo(ctx, testutil.CloudletInfoData)
	testutil.InternalAutoProvPolicyCreate(t, &autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalAppCreate(t, &appApi, testutil.AppData)
	// multi-tenant ClusterInst
	mt := testutil.ClusterInstData[8]
	require.True(t, mt.MultiTenant)
	// negative tests
	// bad Organization
	mtBad := mt
	mtBad.Key.Organization = "foo"
	err := clusterInstApi.CreateClusterInst(&mtBad, testutil.NewCudStreamoutClusterInst(ctx))
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Only MobiledgeX ClusterInsts may be multi-tenant")
	// bad deployment type
	mtBad = mt
	mtBad.Deployment = cloudcommon.DeploymentTypeDocker
	err = clusterInstApi.CreateClusterInst(&mtBad, testutil.NewCudStreamoutClusterInst(ctx))
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Multi-tenant clusters must be of deployment type Kubernetes")

	// Create multi-tenant ClusterInst before reservable tests.
	// Reservable tests should pass without using multi-tenant because
	// the App's SupportMultiTenant is false.
	err = clusterInstApi.CreateClusterInst(&mt, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err)

	checkReserved := func(cloudletKey edgeproto.CloudletKey, found bool, id, reservedBy string) {
		key := &edgeproto.ClusterInstKey{}
		key.ClusterKey.Name = cloudcommon.ReservableClusterPrefix + id
		key.CloudletKey = cloudletKey
		key.Organization = cloudcommon.OrganizationMobiledgeX
		// look up reserved ClusterInst
		clusterInst := edgeproto.ClusterInst{}
		actualFound := clusterInstApi.Get(key, &clusterInst)
		require.Equal(t, found, actualFound, "lookup %s", key.GetKeyString())
		if !found {
			return
		}
		require.True(t, clusterInst.Auto, "clusterinst is auto")
		require.True(t, clusterInst.Reservable, "clusterinst is reservable")
		require.Equal(t, reservedBy, clusterInst.ReservedBy, "reserved by matches")
		// Progress message should be there for cluster instance itself
		msgs := GetClusterInstStreamMsgs(t, ctx, key, Pass)
		require.Greater(t, len(msgs), 0, "some progress messages")
	}
	createAutoClusterAppInst := func(copy edgeproto.AppInst, expectedId string) {
		// since cluster inst does not exist, it will be auto-created
		copy.Key.ClusterInstKey.ClusterKey.Name = cloudcommon.AutoClusterPrefix + expectedId
		copy.Key.ClusterInstKey.Organization = cloudcommon.OrganizationMobiledgeX
		err := appInstApi.CreateAppInst(&copy, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "create app inst")
		// As there was some progress, there should be some messages in stream
		msgs := GetAppInstStreamMsgs(t, ctx, &copy.Key, Pass)
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
		err := appInstApi.DeleteAppInst(&copy, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "delete app inst")
		checkReserved(copy.Key.ClusterInstKey.CloudletKey, true, id, "")
	}
	checkReservedIds := func(key edgeproto.CloudletKey, expected uint64) {
		refs := edgeproto.CloudletRefs{}
		found := cloudletRefsApi.cache.Get(&key, &refs)
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
	clusterInstApi.cleanupIdleReservableAutoClusters(ctx, time.Duration(0))
	clusterInstApi.cleanupWorkers.WaitIdle()
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
	clusterInstApi.cleanupIdleReservableAutoClusters(ctx, time.Duration(0))
	clusterInstApi.cleanupWorkers.WaitIdle()
	checkReserved(cloudletKey, false, "0", "")
	checkReserved(cloudletKey, false, "1", "")
	checkReserved(cloudletKey, false, "2", "")
	checkReservedIds(cloudletKey, 0)

	// Autocluster AppInst with AutoDelete delete option should fail
	autoDeleteAppInst := testutil.AppInstData[10]
	autoDeleteAppInst.RealClusterName = ""
	autoDeleteAppInst.Key.ClusterInstKey.ClusterKey.Name = cloudcommon.AutoClusterPrefix + "foo"
	err = appInstApi.CreateAppInst(&autoDeleteAppInst, testutil.NewCudStreamoutAppInst(ctx))
	require.NotNil(t, err, "create autodelete appInst")
	require.Contains(t, err.Error(), "Sidecar AppInst (AutoDelete App) must specify the RealClusterName field to deploy to the virtual cluster")

	err = clusterInstApi.DeleteClusterInst(&mt, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err)

	testDeprecatedAutoCluster(t, ctx)
	dummy.Stop()
}

func testDeprecatedAutoCluster(t *testing.T, ctx context.Context) {
	// downgrade cloudlets to older crm compatibility version
	cloudletInfos := []edgeproto.CloudletInfo{}
	for _, info := range testutil.CloudletInfoData {
		info.CompatibilityVersion = 0
		cloudletInfos = append(cloudletInfos, info)
	}
	insertCloudletInfo(ctx, cloudletInfos)

	// existing AppInst creates should fail because the ClusterInst org
	// must match the AppInst org.
	for ii, obj := range testutil.AppInstData {
		if !strings.HasPrefix(obj.Key.ClusterInstKey.ClusterKey.Name, cloudcommon.AutoClusterPrefix) {
			continue
		}
		// there is one App that has dev org as MobiledgeX, which matches
		// the Reservable auto-cluster's org, so skip it.
		if obj.Key.AppKey.Organization == obj.Key.ClusterInstKey.Organization {
			continue
		}
		err := appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.NotNil(t, err, "Create autocluster appinst[%d]: %v", ii, obj.Key)
		require.Contains(t, err.Error(), "Developer name mismatch")
	}

	appInsts := []edgeproto.AppInst{}
	clusterInsts := []edgeproto.ClusterInst{}
	for ii, obj := range testutil.AppInstData {
		if !strings.HasPrefix(obj.Key.ClusterInstKey.ClusterKey.Name, cloudcommon.AutoClusterPrefix) {
			continue
		}
		obj.Key.ClusterInstKey.Organization = obj.Key.AppKey.Organization
		appInsts = append(appInsts, obj)

		err := appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
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
	testutil.InternalClusterInstTest(t, "show", &clusterInstApi, clusterInsts)
	// delete AppInsts should delete ClusterInsts
	for _, obj := range appInsts {
		err := appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "Delete autocluster appinst %v", obj.Key)
	}
	testutil.InternalClusterInstTest(t, "show", &clusterInstApi, []edgeproto.ClusterInst{})
	// restore CRM version
	for ii := range cloudletInfos {
		cloudletInfos[ii].CompatibilityVersion = cloudcommon.GetCRMCompatibilityVersion()
	}
	insertCloudletInfo(ctx, cloudletInfos)
}

func testDeprecatedSharedRootLBFQDN(t *testing.T, ctx context.Context) {
	// downgrade cloudlets to older crm compatibility version
	cloudletInfos := []edgeproto.CloudletInfo{}
	for _, info := range testutil.CloudletInfoData {
		info.CompatibilityVersion = 0
		cloudletInfos = append(cloudletInfos, info)
	}
	insertCloudletInfo(ctx, cloudletInfos)

	// create appInst with internalports set to true, this means
	// that this appInst will not need access from external network
	appInstData := testutil.AppInstData[9]
	err := appInstApi.CreateAppInst(&appInstData, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "Create appinst: %v", appInstData.Key)

	show := testutil.ShowAppInst{}
	show.Init()
	filter := edgeproto.AppInst{}
	filter.Key = appInstData.Key

	// For older version of CRM, URI should exist
	err = appInstApi.ShowAppInst(&filter, &show)
	require.Nil(t, err, "show app inst data")
	require.True(t, len(show.Data) == 1, "matches app inst")
	for _, appInstOut := range show.Data {
		require.True(t, appInstOut.Uri != "", "app URI exists")
	}

	// delete appinst
	err = appInstApi.DeleteAppInst(&appInstData, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "Delete appinst: %v", appInstData.Key)

	// restore CRM version
	for ii := range cloudletInfos {
		cloudletInfos[ii].CompatibilityVersion = cloudcommon.GetCRMCompatibilityVersion()
	}
	insertCloudletInfo(ctx, cloudletInfos)

	// create appinst again
	err = appInstApi.CreateAppInst(&appInstData, testutil.NewCudStreamoutAppInst(ctx))
	require.Nil(t, err, "Create appinst: %v", appInstData.Key)

	// For newer version of CRM, URI should not exist
	err = appInstApi.ShowAppInst(&filter, &show)
	require.Nil(t, err, "show app inst data")
	require.True(t, len(show.Data) == 1, "matches app inst")
	for _, appInstOut := range show.Data {
		require.True(t, appInstOut.Uri == "", "app URI doesn't exist")
	}

	// final cleanup: delete appinst
	err = appInstApi.DeleteAppInst(&appInstData, testutil.NewCudStreamoutAppInst(ctx))
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

func forceAppInstState(ctx context.Context, in *edgeproto.AppInst, state edgeproto.TrackedState, responder *DummyInfoResponder) error {
	if responder != nil {
		// disable responder, otherwise it will respond to certain states
		// and change the current state
		responder.enable = false
		defer func() {
			responder.enable = true
		}()
	}
	err := appInstApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		obj := edgeproto.AppInst{}
		if !appInstApi.store.STMGet(stm, &in.Key, &obj) {
			return in.Key.NotFoundError()
		}
		obj.State = state
		appInstApi.store.STMPut(stm, &obj)
		return nil
	})
	return err
}

func testAppFlavorRequest(t *testing.T, ctx context.Context, api *testutil.AppInstCommonApi, responder *DummyInfoResponder) {
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
	_, err := flavorApi.CreateFlavor(ctx, &testflavor)
	require.Nil(t, err, "CreateFlavor")
	nonNomApp := testutil.AppInstData[2]
	nonNomApp.Flavor = testflavor.Key
	err = appInstApi.CreateAppInst(&nonNomApp, testutil.NewCudStreamoutAppInst(ctx))
	require.NotNil(t, err, "non-nom-app-create")
	require.Equal(t, "Cloudlet New York Site doesn't support GPU", err.Error())
}

// Test that Crm Override for Delete App overrides any failures
// on both side-car auto-apps and an underlying auto-cluster.
func testAppInstOverrideTransientDelete(t *testing.T, ctx context.Context, api *testutil.AppInstCommonApi, responder *DummyInfoResponder) {
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
	clustApi := testutil.NewInternalClusterInstApi(&clusterInstApi)

	responder.SetSimulateAppDeleteFailure(true)
	responder.SetSimulateClusterDeleteFailure(true)
	for val, stateName := range edgeproto.TrackedState_name {
		state := edgeproto.TrackedState(val)
		if !edgeproto.IsTransientState(state) {
			continue
		}
		// create app (also creates clusterinst)
		obj = ai
		err = appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "create AppInst")
		err = forceAppInstState(ctx, &obj, state, responder)
		require.Nil(t, err, "force state")
		checkAppInstState(t, ctx, api, &obj, state)

		// set aiauto cluster name from real cluster name of create autocluster
		clKey := obj.ClusterInstKey()
		obj = aiauto
		obj.Key.ClusterInstKey = *clKey.Virtual("")
		// create auto app
		err = appInstApi.CreateAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "create AppInst on cluster %v", obj.Key.ClusterInstKey)
		err = forceAppInstState(ctx, &obj, state, responder)
		require.Nil(t, err, "force state")
		checkAppInstState(t, ctx, api, &obj, state)

		clust = edgeproto.ClusterInst{}
		clust.Key = *clKey
		err = forceClusterInstState(ctx, &clust, state, responder)
		require.Nil(t, err, "force state")
		checkClusterInstState(t, ctx, clustApi, &clust, state)

		// delete app (to be able to delete reservable cluster)
		obj = ai
		obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_AND_TRANSIENT_STATE
		log.SpanLog(ctx, log.DebugLevelInfo, "test run appinst delete")
		err = appInstApi.DeleteAppInst(&obj, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "override crm and transient state %s", stateName)
		log.SpanLog(ctx, log.DebugLevelInfo, "test appinst deleted")

		// delete cluster (should also delete auto app)
		clust.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_AND_TRANSIENT_STATE
		log.SpanLog(ctx, log.DebugLevelInfo, "test run ClusterInst delete")
		err = clusterInstApi.DeleteClusterInst(&clust, testutil.NewCudStreamoutClusterInst(ctx))
		require.Nil(t, err, "override crm and transient state %s", stateName)
		log.SpanLog(ctx, log.DebugLevelInfo, "test ClusterInst deleted")
		// make sure cluster got deleted (means apps also were deleted)
		found := testutil.GetClusterInst(t, ctx, clustApi, &clust.Key, &edgeproto.ClusterInst{})
		require.False(t, found)
	}

	responder.SetSimulateAppDeleteFailure(false)
	responder.SetSimulateClusterDeleteFailure(false)

}
