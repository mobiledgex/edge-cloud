package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	influxq "github.com/mobiledgex/edge-cloud/controller/influxq_client"
	"github.com/mobiledgex/edge-cloud/controller/influxq_client/influxq_testutil"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestClusterInstApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi | log.DebugLevelNotify)
	log.InitTracer(nil)
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
	responder := NewDummyInfoResponder(&apis.appInstApi.cache, &apis.clusterInstApi.cache,
		apis.appInstInfoApi, apis.clusterInstInfoApi)

	reduceInfoTimeouts(t, ctx, apis)

	// cannot create insts without cluster/cloudlet
	for _, obj := range testutil.ClusterInstData {
		err := apis.clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
		require.NotNil(t, err, "Create ClusterInst without cloudlet")
	}

	// create support data
	cloudletData := testutil.CloudletData()
	testutil.InternalFlavorCreate(t, apis.flavorApi, testutil.FlavorData)
	testutil.InternalGPUDriverCreate(t, apis.gpuDriverApi, testutil.GPUDriverData)
	testutil.InternalCloudletCreate(t, apis.cloudletApi, cloudletData)
	insertCloudletInfo(ctx, apis, testutil.CloudletInfoData)
	testutil.InternalAutoProvPolicyCreate(t, apis.autoProvPolicyApi, testutil.AutoProvPolicyData)
	testutil.InternalAutoScalePolicyCreate(t, apis.autoScalePolicyApi, testutil.AutoScalePolicyData)

	// Set responder to fail. This should clean up the object after
	// the fake crm returns a failure. If it doesn't, the next test to
	// create all the cluster insts will fail.
	responder.SetSimulateClusterCreateFailure(true)
	for _, obj := range testutil.ClusterInstData {
		err := apis.clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
		require.NotNil(t, err, "Create ClusterInst responder failures")
		// make sure error matches responder
		require.Equal(t, "Encountered failures: crm create ClusterInst failed", err.Error())
	}
	responder.SetSimulateClusterCreateFailure(false)
	require.Equal(t, 0, len(apis.clusterInstApi.cache.Objs))

	testutil.InternalClusterInstTest(t, "cud", apis.clusterInstApi, testutil.ClusterInstData)
	// after cluster insts create, check that cloudlet refs data is correct.
	testutil.InternalCloudletRefsTest(t, "show", apis.cloudletRefsApi, testutil.CloudletRefsData)

	commonApi := testutil.NewInternalClusterInstApi(apis.clusterInstApi)

	// Set responder to fail delete.
	responder.SetSimulateClusterDeleteFailure(true)
	obj := testutil.ClusterInstData[0]
	err := apis.clusterInstApi.DeleteClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.NotNil(t, err, "Delete ClusterInst responder failure")
	responder.SetSimulateClusterDeleteFailure(false)
	checkClusterInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_READY)

	// check override of error DELETE_ERROR
	err = forceClusterInstState(ctx, &obj, edgeproto.TrackedState_DELETE_ERROR, responder, apis)
	require.Nil(t, err, "force state")
	checkClusterInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_DELETE_ERROR)
	err = apis.clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "create overrides delete error")
	checkClusterInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_READY)
	// progress message should exist
	msgs := GetClusterInstStreamMsgs(t, ctx, &obj.Key, apis, Pass)
	require.Greater(t, len(msgs), 0, "some progress messages")

	// check override of error CREATE_ERROR
	err = forceClusterInstState(ctx, &obj, edgeproto.TrackedState_CREATE_ERROR, responder, apis)
	require.Nil(t, err, "force state")
	checkClusterInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_CREATE_ERROR)
	err = apis.clusterInstApi.DeleteClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "delete overrides create error")
	checkClusterInstState(t, ctx, commonApi, &obj, edgeproto.TrackedState_NOT_PRESENT)
	// progress message should exist
	msgs = GetClusterInstStreamMsgs(t, ctx, &obj.Key, apis, Pass)
	require.Greater(t, len(msgs), 0, "some progress messages")

	// test update of autoscale policy
	obj = testutil.ClusterInstData[0]
	obj.Key.Organization = testutil.AutoScalePolicyData[1].Key.Organization
	err = apis.clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "create ClusterInst")
	check := edgeproto.ClusterInst{}
	found := apis.clusterInstApi.cache.Get(&obj.Key, &check)
	require.True(t, found)
	require.Equal(t, 2, int(check.NumNodes))
	// progress message should exist
	msgs = GetClusterInstStreamMsgs(t, ctx, &obj.Key, apis, Pass)
	require.Greater(t, len(msgs), 0, "some progress messages")

	obj.AutoScalePolicy = testutil.AutoScalePolicyData[1].Key.Name
	obj.Fields = []string{edgeproto.ClusterInstFieldAutoScalePolicy}
	err = apis.clusterInstApi.UpdateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err)
	check = edgeproto.ClusterInst{}
	found = apis.clusterInstApi.cache.Get(&obj.Key, &check)
	require.True(t, found)
	require.Equal(t, testutil.AutoScalePolicyData[1].Key.Name, check.AutoScalePolicy)
	require.Equal(t, 4, int(check.NumNodes))
	// progress message should exist
	msgs = GetClusterInstStreamMsgs(t, ctx, &obj.Key, apis, Pass)
	require.Greater(t, len(msgs), 0, "some progress messages")

	// override CRM error
	responder.SetSimulateClusterCreateFailure(true)
	responder.SetSimulateClusterDeleteFailure(true)
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = apis.clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "override crm error")
	// progress message should exist
	msgs = GetClusterInstStreamMsgs(t, ctx, &obj.Key, apis, Pass)
	require.Greater(t, len(msgs), 0, "some progress messages")
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_ERRORS
	err = apis.clusterInstApi.DeleteClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "override crm error")
	// progress message should exist
	msgs = GetClusterInstStreamMsgs(t, ctx, &obj.Key, apis, Pass)
	require.Greater(t, len(msgs), 0, "some progress messages")

	// ignore CRM
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
	err = apis.clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "ignore crm")
	obj = testutil.ClusterInstData[0]
	obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
	err = apis.clusterInstApi.DeleteClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "ignore crm")

	// inavailability of matching node flavor
	obj = testutil.ClusterInstData[0]
	obj.Flavor = testutil.FlavorData[0].Key
	err = apis.clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.NotNil(t, err, "flavor not available")

	// Create appInst with autocluster should fail as cluster create
	// responder is set to fail. But post failure, clusterInst object
	// created internally should be cleaned up
	targetCloudletKey := cloudletData[1].Key
	targetApp := testutil.AppData[11]
	testReservableClusterInstExists := func(cloudletKey edgeproto.CloudletKey) {
		foundCluster := false
		for cKey, cCache := range apis.clusterInstApi.cache.Objs {
			if cKey.CloudletKey == cloudletKey &&
				cCache.Obj.Reservable {
				foundCluster = true
			}
		}
		require.False(t, foundCluster, "no reservable cluster exists on this cloudlet")
	}
	// 1. Ensure no reservable clusterinst is there on our target cloudlet
	testReservableClusterInstExists(targetCloudletKey)
	// 2. Create AppInst and ensure it fails
	_, err = apis.appApi.CreateApp(ctx, &targetApp)
	require.Nil(t, err, "create App")
	appinstTest := edgeproto.AppInst{}
	appinstTest.Key.AppKey = targetApp.Key
	appinstTest.Key.ClusterInstKey.CloudletKey = targetCloudletKey
	appinstTest.Key.ClusterInstKey.ClusterKey.Name = "autoclustertest"
	appinstTest.Key.ClusterInstKey.Organization = cloudcommon.OrganizationMobiledgeX
	err = apis.appInstApi.CreateAppInst(&appinstTest, testutil.NewCudStreamoutAppInst(ctx))
	require.NotNil(t, err)
	// 3. Ensure no reservable clusterinst exist on the target cloudlet
	testReservableClusterInstExists(targetCloudletKey)
	// 4. Clean up created app
	_, err = apis.appApi.DeleteApp(ctx, &targetApp)
	require.Nil(t, err, "delete App")

	responder.SetSimulateClusterCreateFailure(false)
	responder.SetSimulateClusterDeleteFailure(false)

	testReservableClusterInst(t, ctx, commonApi, apis)
	testClusterInstOverrideTransientDelete(t, ctx, commonApi, responder, apis)

	testClusterInstResourceUsage(t, ctx, apis)
	testClusterInstGPUFlavor(t, ctx, apis)

	dummy.Stop()
}

func reduceInfoTimeouts(t *testing.T, ctx context.Context, apis *AllApis) {
	apis.settingsApi.initDefaults(ctx)

	settings, err := apis.settingsApi.ShowSettings(ctx, &edgeproto.Settings{})
	require.Nil(t, err)

	settings.CreateClusterInstTimeout = edgeproto.Duration(1 * time.Second)
	settings.UpdateClusterInstTimeout = edgeproto.Duration(1 * time.Second)
	settings.DeleteClusterInstTimeout = edgeproto.Duration(1 * time.Second)
	settings.CreateAppInstTimeout = edgeproto.Duration(1 * time.Second)
	settings.UpdateAppInstTimeout = edgeproto.Duration(1 * time.Second)
	settings.DeleteAppInstTimeout = edgeproto.Duration(1 * time.Second)
	settings.CloudletMaintenanceTimeout = edgeproto.Duration(2 * time.Second)

	settings.Fields = []string{
		edgeproto.SettingsFieldCreateAppInstTimeout,
		edgeproto.SettingsFieldUpdateAppInstTimeout,
		edgeproto.SettingsFieldDeleteAppInstTimeout,
		edgeproto.SettingsFieldCreateClusterInstTimeout,
		edgeproto.SettingsFieldUpdateClusterInstTimeout,
		edgeproto.SettingsFieldDeleteClusterInstTimeout,
		edgeproto.SettingsFieldCloudletMaintenanceTimeout,
	}
	_, err = apis.settingsApi.UpdateSettings(ctx, settings)
	require.Nil(t, err)

	updated, err := apis.settingsApi.ShowSettings(ctx, &edgeproto.Settings{})
	updated.Fields = []string{}
	settings.Fields = []string{}
	require.Equal(t, settings, updated)
}

func checkClusterInstState(t *testing.T, ctx context.Context, api *testutil.ClusterInstCommonApi, in *edgeproto.ClusterInst, state edgeproto.TrackedState) {
	out := edgeproto.ClusterInst{}
	found := testutil.GetClusterInst(t, ctx, api, &in.Key, &out)
	log.SpanLog(ctx, log.DebugLevelInfo, "check ClusterInst state", "state", state)
	if state == edgeproto.TrackedState_NOT_PRESENT {
		require.False(t, found, "get cluster inst")
	} else {
		require.True(t, found, "get cluster inst")
		require.Equal(t, state, out.State, "cluster inst state")
	}
}

func forceClusterInstState(ctx context.Context, in *edgeproto.ClusterInst, state edgeproto.TrackedState, responder *DummyInfoResponder, apis *AllApis) error {
	log.SpanLog(ctx, log.DebugLevelInfo, "force ClusterInst state", "state", state)
	if responder != nil {
		// disable responder, otherwise it will respond to certain states
		// and change the current state
		responder.enable = false
		defer func() {
			responder.enable = true
		}()
	}
	err := apis.clusterInstApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		obj := edgeproto.ClusterInst{}
		if !apis.clusterInstApi.store.STMGet(stm, &in.Key, &obj) {
			return in.Key.NotFoundError()
		}
		obj.State = state
		apis.clusterInstApi.store.STMPut(stm, &obj)
		return nil
	})
	return err
}

func testReservableClusterInst(t *testing.T, ctx context.Context, api *testutil.ClusterInstCommonApi, apis *AllApis) {
	cinst := testutil.ClusterInstData[7]
	checkReservedBy(t, ctx, api, &cinst.Key, "")

	// create test app
	for _, app := range testutil.AppData {
		_, err := apis.appApi.CreateApp(ctx, &app)
		require.Nil(t, err, "create App")
	}

	// Should be able to create a developer AppInst on the ClusterInst
	streamOut := testutil.NewCudStreamoutAppInst(ctx)
	appinst := edgeproto.AppInst{}
	appinst.Key.AppKey = testutil.AppData[0].Key
	appinst.Key.ClusterInstKey = *cinst.Key.Virtual("")
	err := apis.appInstApi.CreateAppInst(&appinst, streamOut)
	require.Nil(t, err, "create AppInst")
	checkReservedBy(t, ctx, api, &cinst.Key, appinst.Key.AppKey.Organization)

	// Cannot create another AppInst on it from different developer
	appinst2 := edgeproto.AppInst{}
	appinst2.Key.AppKey = testutil.AppData[10].Key
	appinst2.Key.ClusterInstKey = *cinst.Key.Virtual("")
	appinst2.Flavor = appinst.Flavor
	require.NotEqual(t, appinst.Key.AppKey.Organization, appinst2.Key.AppKey.Organization)
	err = apis.appInstApi.CreateAppInst(&appinst2, streamOut)
	require.NotNil(t, err, "create AppInst on already reserved ClusterInst")
	// Cannot create another AppInst on it from the same developer
	appinst3 := edgeproto.AppInst{}
	appinst3.Key.AppKey = testutil.AppData[1].Key
	appinst3.Key.ClusterInstKey = *cinst.Key.Virtual("")
	require.Equal(t, appinst.Key.AppKey.Organization, appinst3.Key.AppKey.Organization)
	err = apis.appInstApi.CreateAppInst(&appinst3, streamOut)
	require.NotNil(t, err, "create AppInst on already reserved ClusterInst")

	// Make sure above changes have not affected ReservedBy setting
	checkReservedBy(t, ctx, api, &cinst.Key, appinst.Key.AppKey.Organization)

	// Deleting AppInst should removed ReservedBy
	err = apis.appInstApi.DeleteAppInst(&appinst, streamOut)
	require.Nil(t, err, "delete AppInst")
	checkReservedBy(t, ctx, api, &cinst.Key, "")

	// Can now create AppInst from different developer
	err = apis.appInstApi.CreateAppInst(&appinst2, streamOut)
	require.Nil(t, err, "create AppInst on reservable ClusterInst")
	checkReservedBy(t, ctx, api, &cinst.Key, appinst2.Key.AppKey.Organization)

	// Delete AppInst
	err = apis.appInstApi.DeleteAppInst(&appinst2, streamOut)
	require.Nil(t, err, "delete AppInst on reservable ClusterInst")
	checkReservedBy(t, ctx, api, &cinst.Key, "")

	// Cannot create VM with autocluster
	appinstBad := edgeproto.AppInst{}
	appinstBad.Key.AppKey = testutil.AppData[12].Key
	appinstBad.Key.ClusterInstKey.CloudletKey = testutil.CloudletData()[0].Key
	appinstBad.Key.ClusterInstKey.ClusterKey.Name = "autoclusterBad"
	appinstBad.Key.ClusterInstKey.Organization = cloudcommon.OrganizationMobiledgeX
	err = apis.appInstApi.CreateAppInst(&appinstBad, streamOut)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "No cluster required for App deployment type vm")

	// Cannot create VM with autocluster and realclustername
	appinstBad.RealClusterName = cinst.Key.ClusterKey.Name
	err = apis.appInstApi.CreateAppInst(&appinstBad, streamOut)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "No cluster required for App deployment type vm")

	// Delete App
	for _, app := range testutil.AppData {
		_, err = apis.appApi.DeleteApp(ctx, &app)
		require.Nil(t, err, "delete App")
	}
	checkReservedBy(t, ctx, api, &cinst.Key, "")
}

func checkReservedBy(t *testing.T, ctx context.Context, api *testutil.ClusterInstCommonApi, key *edgeproto.ClusterInstKey, expected string) {
	cinst := edgeproto.ClusterInst{}
	found := testutil.GetClusterInst(t, ctx, api, key, &cinst)
	require.True(t, found, "get ClusterInst")
	require.True(t, cinst.Reservable)
	require.Equal(t, expected, cinst.ReservedBy)
	require.Equal(t, cloudcommon.OrganizationMobiledgeX, cinst.Key.Organization)
}

// Test that Crm Override for Delete ClusterInst overrides any failures
// on side-car auto-apps.
func testClusterInstOverrideTransientDelete(t *testing.T, ctx context.Context, api *testutil.ClusterInstCommonApi, responder *DummyInfoResponder, apis *AllApis) {
	clust := testutil.ClusterInstData[0]
	clust.Key.ClusterKey.Name = "crmoverride"

	// autoapp
	app := testutil.AppData[9] // auto-delete app
	require.Equal(t, edgeproto.DeleteType_AUTO_DELETE, app.DelOpt)
	_, err := apis.appApi.CreateApp(ctx, &app)
	require.Nil(t, err, "create App")

	aiauto := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:         app.Key,
			ClusterInstKey: *clust.Key.Virtual(""),
		},
	}

	var obj edgeproto.ClusterInst
	var ai edgeproto.AppInst
	appCommon := testutil.NewInternalAppInstApi(apis.appInstApi)

	responder.SetSimulateClusterDeleteFailure(true)
	responder.SetSimulateAppDeleteFailure(true)
	for val, stateName := range edgeproto.TrackedState_name {
		state := edgeproto.TrackedState(val)
		if !edgeproto.IsTransientState(state) {
			continue
		}
		// create cluster
		obj = clust
		err = apis.clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
		require.Nil(t, err, "create ClusterInst")
		// create autoapp

		ai = aiauto
		err = apis.appInstApi.CreateAppInst(&ai, testutil.NewCudStreamoutAppInst(ctx))
		require.Nil(t, err, "create auto AppInst")
		// force bad states
		err = forceAppInstState(ctx, &ai, state, responder, apis)
		require.Nil(t, err, "force state")
		checkAppInstState(t, ctx, appCommon, &ai, state)
		err = forceClusterInstState(ctx, &obj, state, responder, apis)
		require.Nil(t, err, "force state")
		checkClusterInstState(t, ctx, api, &obj, state)
		// delete cluster
		obj = clust
		obj.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM_AND_TRANSIENT_STATE
		err = apis.clusterInstApi.DeleteClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
		require.Nil(t, err, "override crm and transient state %s", stateName)
	}
	responder.SetSimulateClusterDeleteFailure(false)
	responder.SetSimulateAppDeleteFailure(false)

	_, err = apis.appApi.DeleteApp(ctx, &app)
	require.Nil(t, err, "delete App")

	// cleanup unused reservable auto clusters
	apis.clusterInstApi.cleanupIdleReservableAutoClusters(ctx, time.Duration(0))
	apis.clusterInstApi.cleanupWorkers.WaitIdle()
}

func getMetricCounts(t *testing.T, ctx context.Context, cloudlet *edgeproto.Cloudlet, apis *AllApis) *ResourceMetrics {
	var metrics []*edgeproto.Metric
	var err error
	err = apis.clusterInstApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		metrics, err = apis.clusterInstApi.getCloudletResourceMetric(ctx, stm, &cloudlet.Key)
		require.Nil(t, err, "get cloudlet resource metrics")
		require.Greater(t, len(metrics), 0, "metrics")
		return nil
	})
	require.Nil(t, err)
	pfType := pf.GetType(cloudlet.PlatformType.String())
	resMeasurement := cloudcommon.GetCloudletResourceUsageMeasurement(pfType)
	ramUsed := uint64(0)
	vcpusUsed := uint64(0)
	gpusUsed := uint64(0)
	externalIPsUsed := uint64(0)
	flavorCnt := make(map[string]uint64)
	for _, metric := range metrics {
		if metric.Name == resMeasurement {
			for _, val := range metric.Vals {
				switch val.Name {
				case cloudcommon.ResourceMetricRamMB:
					out := val.Value.(*edgeproto.MetricVal_Ival)
					ramUsed = out.Ival
				case cloudcommon.ResourceMetricVcpus:
					out := val.Value.(*edgeproto.MetricVal_Ival)
					vcpusUsed = out.Ival
				case cloudcommon.ResourceMetricsGpus:
					out := val.Value.(*edgeproto.MetricVal_Ival)
					gpusUsed = out.Ival
				case cloudcommon.ResourceMetricExternalIPs:
					out := val.Value.(*edgeproto.MetricVal_Ival)
					externalIPsUsed = out.Ival
				}
			}
		} else if metric.Name == cloudcommon.CloudletFlavorUsageMeasurement {
			fName := ""
			for _, tag := range metric.Tags {
				if tag.Name != "flavor" {
					continue
				}
				fName = tag.Val
				if _, ok := flavorCnt[fName]; !ok {
					flavorCnt[fName] = uint64(0)
				}
				break
			}
			require.NotEmpty(t, fName, "flavor name found")
			for _, val := range metric.Vals {
				if val.Name != "count" {
					continue
				}
				out := val.Value.(*edgeproto.MetricVal_Ival)
				_, ok := flavorCnt[fName]
				require.True(t, ok, "flavor name found")
				flavorCnt[fName] += out.Ival
				break
			}
		}
	}
	return &ResourceMetrics{
		ramUsed:         ramUsed,
		vcpusUsed:       vcpusUsed,
		gpusUsed:        gpusUsed,
		externalIpsUsed: externalIPsUsed,
		flavorCnt:       flavorCnt,
	}
}

func getMetricsDiff(old *ResourceMetrics, new *ResourceMetrics) *ResourceMetrics {
	diffRam := new.ramUsed - old.ramUsed
	diffVcpus := new.vcpusUsed - old.vcpusUsed
	diffGpus := new.gpusUsed - old.gpusUsed
	diffExternalIps := new.externalIpsUsed - old.externalIpsUsed
	diffFlavorCnt := make(map[string]uint64)
	for fName, fCnt := range new.flavorCnt {
		if oldVal, ok := old.flavorCnt[fName]; ok {
			diffCnt := fCnt - oldVal
			if diffCnt > 0 {
				diffFlavorCnt[fName] = diffCnt
			}
		} else {
			diffFlavorCnt[fName] = fCnt
		}
	}
	return &ResourceMetrics{
		ramUsed:         diffRam,
		vcpusUsed:       diffVcpus,
		gpusUsed:        diffGpus,
		externalIpsUsed: diffExternalIps,
		flavorCnt:       diffFlavorCnt,
	}
}

func getClusterInstMetricCounts(t *testing.T, ctx context.Context, clusterInst *edgeproto.ClusterInst, apis *AllApis) *ResourceMetrics {
	var err error
	var nodeFlavor *edgeproto.FlavorInfo
	var masterNodeFlavor *edgeproto.FlavorInfo
	var lbFlavor *edgeproto.FlavorInfo
	err = apis.clusterInstApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		existingCl := edgeproto.ClusterInst{}
		found := apis.clusterInstApi.store.STMGet(stm, &clusterInst.Key, &existingCl)
		require.True(t, found, "cluster inst exists")

		cloudletKey := existingCl.Key.CloudletKey
		cloudlet := edgeproto.Cloudlet{}
		found = apis.cloudletApi.store.STMGet(stm, &cloudletKey, &cloudlet)
		require.True(t, found, "cloudlet exists")

		cloudletInfo := edgeproto.CloudletInfo{}
		found = apis.cloudletInfoApi.store.STMGet(stm, &cloudletKey, &cloudletInfo)
		require.True(t, found, "cloudlet info exists")

		for _, flavor := range cloudletInfo.Flavors {
			if flavor.Name == existingCl.NodeFlavor {
				nodeFlavor = flavor
			}
			if flavor.Name == existingCl.MasterNodeFlavor {
				masterNodeFlavor = flavor
			}
		}
		lbFlavor, err = apis.clusterInstApi.GetRootLBFlavorInfo(ctx, stm, &cloudlet, &cloudletInfo)
		require.Nil(t, err, "found rootlb flavor")
		return nil
	})
	require.Nil(t, err)
	require.NotNil(t, nodeFlavor, "found node flavor")
	require.NotNil(t, masterNodeFlavor, "found master node flavor")
	require.NotNil(t, lbFlavor, "found rootlb flavor")
	ramUsed := uint64(clusterInst.NumNodes)*nodeFlavor.Ram +
		uint64(clusterInst.NumMasters)*masterNodeFlavor.Ram +
		lbFlavor.Ram
	vcpusUsed := uint64(clusterInst.NumNodes)*nodeFlavor.Vcpus +
		uint64(clusterInst.NumMasters)*masterNodeFlavor.Vcpus +
		lbFlavor.Vcpus
	externalIPsUsed := uint64(1) // 1 dedicated Flavor
	gpusUsed := uint64(0)
	if clusterInst.OptRes == "gpu" {
		gpusUsed = uint64(clusterInst.NumNodes)
		if clusterInst.NodeFlavor == clusterInst.MasterNodeFlavor {
			gpusUsed += uint64(clusterInst.NumMasters)
		}
	}
	flavorCnt := make(map[string]uint64)
	if _, ok := flavorCnt[nodeFlavor.Name]; !ok {
		flavorCnt[nodeFlavor.Name] = uint64(0)
	}
	flavorCnt[nodeFlavor.Name] += uint64(clusterInst.NumNodes)
	if _, ok := flavorCnt[masterNodeFlavor.Name]; !ok {
		flavorCnt[masterNodeFlavor.Name] = uint64(0)
	}
	flavorCnt[masterNodeFlavor.Name] += uint64(clusterInst.NumMasters)
	if _, ok := flavorCnt[lbFlavor.Name]; !ok {
		flavorCnt[lbFlavor.Name] = uint64(0)
	}
	flavorCnt[lbFlavor.Name] += uint64(1)
	return &ResourceMetrics{
		ramUsed:         ramUsed,
		vcpusUsed:       vcpusUsed,
		gpusUsed:        gpusUsed,
		externalIpsUsed: externalIPsUsed,
		flavorCnt:       flavorCnt,
	}
}

type ResourceMetrics struct {
	ramUsed         uint64
	vcpusUsed       uint64
	gpusUsed        uint64
	externalIpsUsed uint64
	flavorCnt       map[string]uint64
}

func validateClusterInstMetrics(t *testing.T, ctx context.Context, cloudlet *edgeproto.Cloudlet, clusterInst *edgeproto.ClusterInst, oldResUsage *ResourceMetrics, apis *AllApis) {
	// get resource usage after clusterInst creation
	newResUsage := getMetricCounts(t, ctx, cloudlet, apis)
	// get resource usage of clusterInst
	actualResUsage := getMetricsDiff(oldResUsage, newResUsage)
	// validate that metrics output shows expected clusterinst resources
	expectedResUsage := getClusterInstMetricCounts(t, ctx, clusterInst, apis)
	require.Equal(t, actualResUsage.ramUsed, expectedResUsage.ramUsed, "ram metric matches")
	require.Equal(t, actualResUsage.vcpusUsed, expectedResUsage.vcpusUsed, "vcpus metric matches")
	require.Equal(t, actualResUsage.gpusUsed, expectedResUsage.gpusUsed, "gpus metric matches")
	require.Equal(t, actualResUsage.externalIpsUsed, expectedResUsage.externalIpsUsed, "externalips metric matches")
	for efName, efCnt := range expectedResUsage.flavorCnt {
		afCnt, found := actualResUsage.flavorCnt[efName]
		require.True(t, found, "expected flavor found")
		require.Equal(t, afCnt, efCnt, "flavor count matches")
	}
}

func testClusterInstResourceUsage(t *testing.T, ctx context.Context, apis *AllApis) {
	obj := testutil.ClusterInstData[0]
	obj.NumNodes = 10
	obj.Flavor = testutil.FlavorData[3].Key
	err := apis.clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.NotNil(t, err, "not enough resources available")
	require.Contains(t, err.Error(), "Not enough")

	// create appinst
	testutil.InternalResTagTableTest(t, "cud", apis.resTagTableApi, testutil.ResTagTableData)
	testutil.InternalResTagTableTest(t, "show", apis.resTagTableApi, testutil.ResTagTableData)
	testutil.InternalAppCreate(t, apis.appApi, []edgeproto.App{
		testutil.AppData[0], testutil.AppData[12],
	})

	// Update settings to empty master node flavor, so that we can use GPU flavor for master node
	apis.settingsApi.initDefaults(ctx)
	settings, err := apis.settingsApi.ShowSettings(ctx, &edgeproto.Settings{})
	require.Nil(t, err)
	settings.MasterNodeFlavor = ""
	settings.Fields = []string{edgeproto.SettingsFieldMasterNodeFlavor}
	_, err = apis.settingsApi.UpdateSettings(ctx, settings)
	require.Nil(t, err)

	// get resource usage before clusterInst creation
	cloudletData := testutil.CloudletData()
	oldResUsage := getMetricCounts(t, ctx, &cloudletData[0], apis)

	// create clusterInst1
	clusterInstObj := testutil.ClusterInstData[0]
	clusterInstObj.Key.ClusterKey.Name = "GPUCluster"
	clusterInstObj.Flavor = testutil.FlavorData[4].Key
	err = apis.clusterInstApi.CreateClusterInst(&clusterInstObj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "create cluster inst with gpu flavor")
	// validate clusterinst resource metrics
	validateClusterInstMetrics(t, ctx, &cloudletData[0], &clusterInstObj, oldResUsage, apis)

	// get resource usage before clusterInst creation
	oldResUsage = getMetricCounts(t, ctx, &cloudletData[0], apis)

	// create clusterInst2
	clusterInstObj2 := testutil.ClusterInstData[0]
	clusterInstObj2.Key.ClusterKey.Name = "NonGPUCluster1"
	err = apis.clusterInstApi.CreateClusterInst(&clusterInstObj2, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "create cluster inst")
	// validate clusterinst resource metrics
	validateClusterInstMetrics(t, ctx, &cloudletData[0], &clusterInstObj2, oldResUsage, apis)

	// delete clusterInst2
	err = apis.clusterInstApi.DeleteClusterInst(&clusterInstObj2, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "delete cluster inst")
	// validate clusterinst resource metrics post deletion
	delResUsage := getMetricCounts(t, ctx, &cloudletData[0], apis)
	require.Equal(t, oldResUsage.ramUsed, delResUsage.ramUsed, "ram used is same as old value")
	require.Equal(t, oldResUsage.vcpusUsed, delResUsage.vcpusUsed, "vcpus used is same as old value")
	require.Equal(t, oldResUsage.gpusUsed, delResUsage.gpusUsed, "gpus used is same as old value")
	require.Equal(t, oldResUsage.externalIpsUsed, delResUsage.externalIpsUsed, "externalIpsUsed used is same as old value")
	for efName, efCnt := range delResUsage.flavorCnt {
		afCnt, found := oldResUsage.flavorCnt[efName]
		require.True(t, found, "expected flavor found")
		require.Equal(t, afCnt, efCnt, "flavor count matches")
	}

	// get resource usage before clusterInst creation
	oldResUsage = getMetricCounts(t, ctx, &cloudletData[0], apis)

	// create clusterInst2 again
	err = apis.clusterInstApi.CreateClusterInst(&clusterInstObj2, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "create cluster inst")
	// validate clusterinst resource metrics
	validateClusterInstMetrics(t, ctx, &cloudletData[0], &clusterInstObj2, oldResUsage, apis)

	appInstObj := testutil.AppInstData[0]
	appInstObj.Key.ClusterInstKey = *clusterInstObj.Key.Virtual("")
	testutil.InternalAppInstCreate(t, apis.appInstApi, []edgeproto.AppInst{
		appInstObj, testutil.AppInstData[11],
	})

	err = apis.clusterInstApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cloudletKey := obj.Key.CloudletKey
		cloudlet := edgeproto.Cloudlet{}
		found := apis.cloudletApi.store.STMGet(stm, &cloudletKey, &cloudlet)
		require.True(t, found, "cloudlet exists")

		cloudletInfo := edgeproto.CloudletInfo{}
		found = apis.cloudletInfoApi.store.STMGet(stm, &cloudletKey, &cloudletInfo)
		require.True(t, found, "cloudlet info exists")

		cloudletRefs := edgeproto.CloudletRefs{}
		found = apis.cloudletRefsApi.store.STMGet(stm, &cloudletKey, &cloudletRefs)
		require.True(t, found, "cloudlet refs exists")

		allRes, diffRes, _, err := apis.clusterInstApi.getAllCloudletResources(ctx, stm, &cloudlet, &cloudletInfo, &cloudletRefs)
		require.Nil(t, err, "get all cloudlet resources")
		require.Equal(t, len(allRes), len(diffRes), "should match as crm resource snapshot doesn't have any tracked resources")
		clusters := make(map[edgeproto.ClusterInstKey]struct{})
		resTypeVMAppCount := 0
		for _, res := range allRes {
			if res.Key.ClusterKey.Name == cloudcommon.DefaultClust {
				resTypeVMAppCount++
				continue
			}
			existingCl := edgeproto.ClusterInst{}
			found = apis.clusterInstApi.store.STMGet(stm, &res.Key, &existingCl)
			require.True(t, found, "cluster inst from resources exists")
			clusters[res.Key] = struct{}{}
		}
		require.Equal(t, resTypeVMAppCount, 2, "two vm appinst resource exists")
		for _, ciRefKey := range cloudletRefs.ClusterInsts {
			ciKey := edgeproto.ClusterInstKey{}
			ciKey.FromClusterInstRefKey(&ciRefKey, &cloudletRefs.Key)
			existingCl := edgeproto.ClusterInst{}
			if apis.clusterInstApi.store.STMGet(stm, &ciKey, &existingCl) {
				_, found = clusters[ciKey]
				require.True(t, found, "refs clusterinst exists", ciKey)
			}
		}
		require.Equal(t, len(cloudletRefs.VmAppInsts), 1, "1 vm appinsts exists")

		// test cluster inst vm requirements
		quotaMap := make(map[string]edgeproto.ResourceQuota)
		for _, quota := range cloudlet.ResourceQuotas {
			quotaMap[quota.Name] = quota
		}
		lbFlavor, err := apis.clusterInstApi.GetRootLBFlavorInfo(ctx, stm, &cloudlet, &cloudletInfo)
		require.Nil(t, err, "found rootlb flavor")
		clusterInst := testutil.ClusterInstData[0]
		clusterInst.NumMasters = 2
		clusterInst.NumNodes = 2
		clusterInst.IpAccess = edgeproto.IpAccess_IP_ACCESS_DEDICATED
		clusterInst.Flavor = testutil.FlavorData[4].Key
		clusterInst.NodeFlavor = "flavor.large"
		nodeFlavorInfo, masterFlavorInfo, err := apis.clusterInstApi.getClusterFlavorInfo(ctx, stm, cloudletInfo.Flavors, &clusterInst)
		require.Nil(t, err, "get cluster flavor info")
		ciResources, err := cloudcommon.GetClusterInstVMRequirements(ctx, &clusterInst, nodeFlavorInfo, masterFlavorInfo, lbFlavor)
		require.Nil(t, err, "get cluster inst vm requirements")
		// number of vm resources = num_nodes + num_masters + num_of_rootLBs
		require.Equal(t, 5, len(ciResources), "matches number of vm resources")
		numNodes := 0
		numMasters := 0
		numRootLB := 0
		for _, res := range ciResources {
			if res.Type == cloudcommon.VMTypeClusterMaster {
				numMasters++
			} else if res.Type == cloudcommon.VMTypeClusterK8sNode {
				numNodes++
			} else if res.Type == cloudcommon.VMTypeRootLB {
				numRootLB++
			} else {
				require.Fail(t, "invalid resource type", "type", res.Type)
			}
			require.Equal(t, res.Key, clusterInst.Key, "resource key matches cluster inst key")
		}
		require.Equal(t, numMasters, int(clusterInst.NumMasters), "resource type count matches")
		require.Equal(t, numNodes, int(clusterInst.NumNodes), "resource type count matches")
		require.Equal(t, numRootLB, 1, "resource type count matches")

		skipInfraCheck := false
		warnings, err := apis.clusterInstApi.validateCloudletInfraResources(ctx, stm, &cloudlet, &cloudletInfo.ResourcesSnapshot, allRes, ciResources, diffRes, skipInfraCheck)
		require.NotNil(t, err, "not enough resource available error")
		require.Greater(t, len(warnings), 0, "warnings for resources", "warnings", warnings)
		for _, warning := range warnings {
			if strings.Contains(warning, "RAM") {
				quota, found := quotaMap["RAM"]
				require.True(t, found, "quota for RAM is set")
				require.Contains(t, warning, fmt.Sprintf("%d%%", quota.AlertThreshold))
			} else if strings.Contains(warning, "vCPUs") {
				quota, found := quotaMap["vCPUs"]
				require.True(t, found, "quota for vCPUs is set")
				require.Contains(t, warning, fmt.Sprintf("%d%%", quota.AlertThreshold))
			} else if strings.Contains(warning, "GPUs") {
				quota, found := quotaMap["GPUs"]
				require.True(t, found, "quota for GPUs is set")
				require.Contains(t, warning, fmt.Sprintf("%d%%", quota.AlertThreshold))
			} else {
				require.Contains(t, warning, fmt.Sprintf("%d%%", cloudlet.DefaultResourceAlertThreshold))
			}
		}

		// test vm app inst resource requirements
		appInst := testutil.AppInstData[11]
		appInst.Flavor = testutil.FlavorData[4].Key
		appInst.VmFlavor = "flavor.large"
		vmAppResources, err := cloudcommon.GetVMAppRequirements(ctx, &testutil.AppData[12], &appInst, cloudletInfo.Flavors, lbFlavor)
		require.Nil(t, err, "get app inst vm requirements")
		require.Equal(t, 2, len(vmAppResources), "matches number of vm resources")
		foundVMRes := false
		foundVMRootLBRes := false
		for _, vmRes := range vmAppResources {
			if vmRes.Type == cloudcommon.VMTypeAppVM {
				foundVMRes = true
			} else if vmRes.Type == cloudcommon.VMTypeRootLB {
				foundVMRootLBRes = true
			}
			require.Equal(t, vmAppResources[0].Key, *appInst.ClusterInstKey(), "resource key matches appinst's clusterinst key")
		}
		require.True(t, foundVMRes, "resource type app vm found")
		require.True(t, foundVMRootLBRes, "resource type vm rootlb found")

		warnings, err = apis.clusterInstApi.validateCloudletInfraResources(ctx, stm, &cloudlet, &cloudletInfo.ResourcesSnapshot, allRes, vmAppResources, diffRes, skipInfraCheck)
		require.Nil(t, err, "enough resource available")
		require.Greater(t, len(warnings), 0, "warnings for resources", "warnings", warnings)

		for _, warning := range warnings {
			if strings.HasPrefix(warning, "[Quota]") {
				continue
			} else if strings.Contains(warning, "RAM") {
				quota, found := quotaMap["RAM"]
				require.True(t, found, "quota for RAM is set")
				require.Contains(t, warning, fmt.Sprintf("%d%%", quota.AlertThreshold))
			} else if strings.Contains(warning, "vCPUs") {
				quota, found := quotaMap["vCPUs"]
				require.True(t, found, "quota for vCPUs is set")
				require.Contains(t, warning, fmt.Sprintf("%d%%", quota.AlertThreshold))
			} else if strings.Contains(warning, "GPUs") {
				quota, found := quotaMap["GPUs"]
				require.True(t, found, "quota for GPUs is set")
				require.Contains(t, warning, fmt.Sprintf("%d%%", quota.AlertThreshold))
			} else {
				require.Contains(t, warning, fmt.Sprintf("%d%%", cloudlet.DefaultResourceAlertThreshold))
			}
		}
		return nil
	})
	require.Nil(t, err)
}

var testInfluxProc *process.Influx

func influxUsageUnitTestSetup(t *testing.T) {
	testInfluxProc = influxq_testutil.StartInfluxd(t)

	q := influxq.NewInfluxQ(cloudcommon.EventsDbName, "", "")
	err := q.Start("http://" + testInfluxProc.HttpAddr)
	if err != nil {
		influxUsageUnitTestStop()
	}
	require.Nil(t, err, "new influx q")
	services.events = q

	connected := q.WaitConnected()
	if !connected {
		influxUsageUnitTestStop()
	}
	require.True(t, connected)
}

func influxUsageUnitTestStop() {
	if services.events != nil {
		services.events.Stop()
		services.events = nil
	}
	if testInfluxProc != nil {
		testInfluxProc.StopLocal()
		testInfluxProc = nil
	}
}

func TestInflux(t *testing.T) {
	p := influxq_testutil.StartInfluxd(t)
	p.StopLocal()
	p = influxq_testutil.StartInfluxd(t)
	p.StopLocal()
}

func TestDefaultMTCluster(t *testing.T) {
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

	testutil.InternalFlavorTest(t, "cud", apis.flavorApi, testutil.FlavorData)

	cloudlet := testutil.CloudletData()[0]
	cloudlet.EnableDefaultServerlessCluster = true
	cloudlet.GpuConfig = edgeproto.GPUConfig{}
	cloudletInfo := testutil.CloudletInfoData[0]
	cloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_INIT
	apis.cloudletInfoApi.Update(ctx, &cloudletInfo, 0)

	// create cloudlet, should create cluster
	err := apis.cloudletApi.CreateCloudlet(&cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)
	// simulate ready state in info, which triggers cluster create
	cloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_READY
	apis.cloudletInfoApi.Update(ctx, &cloudletInfo, 0)
	waitDefaultMTClust(t, cloudlet.Key, apis, true)

	// update to off, should delete cluster
	cloudlet.EnableDefaultServerlessCluster = false
	cloudlet.Fields = []string{edgeproto.CloudletFieldEnableDefaultServerlessCluster}
	err = apis.cloudletApi.UpdateCloudlet(&cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)
	waitDefaultMTClust(t, cloudlet.Key, apis, false)

	// update to on, should create cluster
	cloudlet.EnableDefaultServerlessCluster = true
	err = apis.cloudletApi.UpdateCloudlet(&cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)
	waitDefaultMTClust(t, cloudlet.Key, apis, true)

	// delete cloudlet, should auto-delete cluster
	err = apis.cloudletApi.DeleteCloudlet(&cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)
	waitDefaultMTClust(t, cloudlet.Key, apis, false)
}

func waitDefaultMTClust(t *testing.T, cloudletKey edgeproto.CloudletKey, apis *AllApis, present bool) {
	key := getDefaultMTClustKey(cloudletKey)
	cinst := edgeproto.ClusterInst{}
	var found bool
	for ii := 0; ii < 40; ii++ {
		found = apis.clusterInstApi.Get(key, &cinst)
		if present == found {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	require.Equal(t, present, found, "DefaultMTCluster presence incorrect")
}

func testClusterInstGPUFlavor(t *testing.T, ctx context.Context, apis *AllApis) {
	cloudletData := testutil.CloudletData()
	vgpuCloudlet := cloudletData[0]
	vgpuCloudlet.Key.Name = "VGPUCloudlet"
	vgpuCloudlet.GpuConfig.Driver = testutil.GPUDriverData[3].Key
	vgpuCloudlet.ResTagMap["gpu"] = &testutil.Restblkeys[0]
	err := apis.cloudletApi.CreateCloudlet(&vgpuCloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)
	cloudletInfo := testutil.CloudletInfoData[0]
	cloudletInfo.Key = vgpuCloudlet.Key
	cloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_READY
	apis.cloudletInfoApi.Update(ctx, &cloudletInfo, 0)

	obj := testutil.ClusterInstData[0]
	obj.Key.ClusterKey.Name = "GPUTestClusterFlavor"
	obj.Flavor = testutil.FlavorData[4].Key // GPU Passthrough flavor

	// Deploy GPU cluster on non-GPU cloudlet, should fail
	obj.Key.CloudletKey = cloudletData[1].Key
	err = apis.clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.NotNil(t, err, "create cluster inst with gpu flavor on vgpu cloudlet fails")
	require.Contains(t, err.Error(), "doesn't support GPU")

	// Deploy GPU passthrough cluster on vGPU cloudlet, should fail
	obj.Key.CloudletKey = vgpuCloudlet.Key
	err = apis.clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.NotNil(t, err, "create cluster inst with gpu flavor on vgpu cloudlet fails")
	require.Contains(t, err.Error(), "doesn't support GPU resource \"pci\"")

	vgpuFlavor := testutil.FlavorData[4]
	vgpuFlavor.Key.Name = "mex-vgpu-flavor"
	vgpuFlavor.OptResMap["gpu"] = "vmware:vgpu:1"
	_, err = apis.flavorApi.CreateFlavor(ctx, &vgpuFlavor)
	require.Nil(t, err, "create flavor as vgpu flavor")

	obj.Flavor = vgpuFlavor.Key
	verbose = true
	err = apis.clusterInstApi.CreateClusterInst(&obj, testutil.NewCudStreamoutClusterInst(ctx))
	require.Nil(t, err, "create cluster inst with vgpu flavor on vgpu cloudlet")
}
