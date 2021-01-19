package main

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestAutoProvPolicyApi(t *testing.T) {
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

	NewDummyInfoResponder(&appInstApi.cache, &clusterInstApi.cache, &appInstInfoApi, &clusterInstInfoApi)
	reduceInfoTimeouts(t, ctx)
	InfluxUsageUnitTestSetup(t)
	defer InfluxUsageUnitTestStop()

	testutil.InternalAutoProvPolicyTest(t, "cud", &autoProvPolicyApi, testutil.AutoProvPolicyData)

	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)

	// test adding cloudlet to policy
	pc := edgeproto.AutoProvPolicyCloudlet{}
	pc.Key = testutil.AutoProvPolicyData[0].Key
	pc.CloudletKey = testutil.CloudletData[0].Key

	_, err := autoProvPolicyApi.AddAutoProvPolicyCloudlet(ctx, &pc)
	require.Nil(t, err, "add auto prov policy cloudlet")
	policy := edgeproto.AutoProvPolicy{}
	found := autoProvPolicyApi.cache.Get(&pc.Key, &policy)
	require.True(t, found, "get auto prov policy %v", pc.Key)
	require.Equal(t, 1, len(policy.Cloudlets))
	require.Equal(t, pc.CloudletKey, policy.Cloudlets[0].Key)

	// test adding another cloudlet to policy
	pc2 := edgeproto.AutoProvPolicyCloudlet{}
	pc2.Key = testutil.AutoProvPolicyData[0].Key
	pc2.CloudletKey = testutil.CloudletData[1].Key

	_, err = autoProvPolicyApi.AddAutoProvPolicyCloudlet(ctx, &pc2)
	require.Nil(t, err, "add auto prov policy cloudlet")
	found = autoProvPolicyApi.cache.Get(&pc2.Key, &policy)
	require.True(t, found, "get auto prov policy %v", pc2.Key)
	require.Equal(t, 2, len(policy.Cloudlets))
	require.Equal(t, pc2.CloudletKey, policy.Cloudlets[1].Key)

	// remove cloudlet from policy
	_, err = autoProvPolicyApi.RemoveAutoProvPolicyCloudlet(ctx, &pc)
	require.Nil(t, err, "remove auto prov policy cloudlet")
	found = autoProvPolicyApi.cache.Get(&pc.Key, &policy)
	require.True(t, found, "get auto prov policy %v", pc.Key)
	require.Equal(t, 1, len(policy.Cloudlets))
	require.Equal(t, pc2.CloudletKey, policy.Cloudlets[0].Key)

	// remove last cloudlet from policy
	_, err = autoProvPolicyApi.RemoveAutoProvPolicyCloudlet(ctx, &pc2)
	require.Nil(t, err, "remove auto prov policy cloudlet")
	found = autoProvPolicyApi.cache.Get(&pc2.Key, &policy)
	require.True(t, found, "get auto prov policy %v", pc2.Key)
	require.Equal(t, 0, len(policy.Cloudlets))

	// try to add policy for non-existent cloudlet
	pc.CloudletKey.Name = ""
	_, err = autoProvPolicyApi.AddAutoProvPolicyCloudlet(ctx, &pc)
	require.NotNil(t, err)

	// try to add a policy with min > max
	policy = edgeproto.AutoProvPolicy{}
	policy.Key.Name = "badpolicy"
	policy.MinActiveInstances = 3
	policy.MaxInstances = 2
	_, err = autoProvPolicyApi.CreateAutoProvPolicy(ctx, &policy)
	require.NotNil(t, err)

	addRemoveAutoProvPolicy(t, ctx)
	testApiChecks(t, ctx)
	dummy.Stop()
}

func addRemoveAutoProvPolicy(t *testing.T, ctx context.Context) {
	// add app with multiple policies
	app := testutil.AppData[11]
	require.True(t, len(app.AutoProvPolicies) > 1)
	_, err := appApi.CreateApp(ctx, &app)
	require.Nil(t, err)

	// new policy (copy)
	ap := testutil.AutoProvPolicyData[3]
	ap.Key.Name = "test-policy"
	_, err = autoProvPolicyApi.CreateAutoProvPolicy(ctx, &ap)
	require.Nil(t, err)

	// add new policy to app
	appPolicy := edgeproto.AppAutoProvPolicy{
		AppKey:         app.Key,
		AutoProvPolicy: ap.Key.Name,
	}
	_, err = appApi.AddAppAutoProvPolicy(ctx, &appPolicy)
	require.Nil(t, err)

	appCheck := edgeproto.App{}
	found := appApi.Get(&app.Key, &appCheck)
	require.True(t, found)
	require.Equal(t, 3, len(appCheck.AutoProvPolicies))
	found = false
	for _, str := range appCheck.AutoProvPolicies {
		if str == ap.Key.Name {
			found = true
		}
	}
	require.True(t, found)

	// remove policy from app
	_, err = appApi.RemoveAppAutoProvPolicy(ctx, &appPolicy)
	require.Nil(t, err)

	found = appApi.Get(&app.Key, &appCheck)
	require.True(t, found)
	require.Equal(t, 2, len(appCheck.AutoProvPolicies))
	found = false
	for _, str := range appCheck.AutoProvPolicies {
		if str == ap.Key.Name {
			found = true
		}
	}
	require.False(t, found)

	_, err = appApi.DeleteApp(ctx, &app)
	require.Nil(t, err)
}

func testApiChecks(t *testing.T, ctx context.Context) {
	var err error
	flavor := testutil.FlavorData[0]
	app := edgeproto.App{}
	app.Key.Name = "checkDemand"
	app.Key.Organization = "org"
	app.Key.Version = "1.0.0"
	app.ImageType = edgeproto.ImageType_IMAGE_TYPE_DOCKER
	app.DefaultFlavor = flavor.Key

	numCloudlets1 := 6
	pt1 := newAutoProvPolicyTest("policy1", app.Key.Organization, numCloudlets1, &flavor)
	pt1.policy.MinActiveInstances = 2
	pt1.policy.MaxInstances = 4

	numCloudlets2 := 10
	pt2 := newAutoProvPolicyTest("policy2", app.Key.Organization, numCloudlets2, &flavor)
	pt2.policy.MinActiveInstances = 5
	pt2.policy.MaxInstances = 8

	// pt3 is used to test limit of one AppInst per cloudlet
	pt3 := newAutoProvPolicyTest("policy3", app.Key.Organization, 1, &flavor)
	pt3.policy.MinActiveInstances = 0
	pt3.policy.DeployClientCount = 1
	pt3.policy.MaxInstances = 20
	// Add extra reservable ClusterInsts on the Cloudlet
	for ii := 0; ii < 3; ii++ {
		// copy ClusterInst
		cl := pt3.clusterInsts[0]
		cl.Key.ClusterKey.Name = fmt.Sprintf("extra-%d", ii)
		pt3.clusterInsts = append(pt3.clusterInsts, cl)
	}

	app.AutoProvPolicies = append(app.AutoProvPolicies,
		pt1.policy.Key.Name,
		pt2.policy.Key.Name,
		pt3.policy.Key.Name)

	// create all supporting data
	pt1.create(t, ctx)
	pt2.create(t, ctx)
	pt3.create(t, ctx)
	_, err = appApi.CreateApp(ctx, &app)
	require.Nil(t, err)

	// *** Check Demand Reason ***

	// spawn threads to create AppInsts on every cloudlet
	// the checks should limit the creates to MaxInstances
	pt1.goDoAppInsts(t, ctx, &app, cloudcommon.Create, cloudcommon.AutoProvReasonDemand)
	pt1.expectAppInsts(t, ctx, &app, int(pt1.policy.MaxInstances))

	pt2.goDoAppInsts(t, ctx, &app, cloudcommon.Create, cloudcommon.AutoProvReasonDemand)
	pt2.expectAppInsts(t, ctx, &app, int(pt2.policy.MaxInstances))

	// spawn threads to delete AppInsts on every cloudlet
	// the checks should limit the deletes to MinActiveInstances
	pt1.goDoAppInsts(t, ctx, &app, cloudcommon.Delete, cloudcommon.AutoProvReasonDemand)
	pt1.expectAppInsts(t, ctx, &app, int(pt1.policy.MinActiveInstances))

	pt2.goDoAppInsts(t, ctx, &app, cloudcommon.Delete, cloudcommon.AutoProvReasonDemand)
	pt2.expectAppInsts(t, ctx, &app, int(pt2.policy.MinActiveInstances))

	// remove all AppInsts to prep for next test (as if done by user)
	pt1.goDoAppInsts(t, ctx, &app, cloudcommon.Delete, "")
	pt1.expectAppInsts(t, ctx, &app, 0)

	pt2.goDoAppInsts(t, ctx, &app, cloudcommon.Delete, "")
	pt2.expectAppInsts(t, ctx, &app, 0)

	// *** Check MinMax Reason ***

	// spawn threads to create AppInsts on every cloudlet
	// the checks should limit the creates to MinActiveInstances
	pt1.goDoAppInsts(t, ctx, &app, cloudcommon.Create, cloudcommon.AutoProvReasonMinMax)
	pt1.expectAppInsts(t, ctx, &app, int(pt1.policy.MinActiveInstances))

	pt2.goDoAppInsts(t, ctx, &app, cloudcommon.Create, cloudcommon.AutoProvReasonMinMax)
	pt2.expectAppInsts(t, ctx, &app, int(pt2.policy.MinActiveInstances))

	// create instances manually on all cloudlets
	pt1.goDoAppInsts(t, ctx, &app, cloudcommon.Create, "")
	pt1.expectAppInsts(t, ctx, &app, numCloudlets1)

	pt2.goDoAppInsts(t, ctx, &app, cloudcommon.Create, "")
	pt2.expectAppInsts(t, ctx, &app, numCloudlets2)

	// spawn threads to delete AppInsts on every cloudlet
	// the checks should limit the deletes to MaxInstances
	pt1.goDoAppInsts(t, ctx, &app, cloudcommon.Delete, cloudcommon.AutoProvReasonMinMax)
	pt1.expectAppInsts(t, ctx, &app, int(pt1.policy.MaxInstances))

	pt2.goDoAppInsts(t, ctx, &app, cloudcommon.Delete, cloudcommon.AutoProvReasonMinMax)
	pt2.expectAppInsts(t, ctx, &app, int(pt2.policy.MaxInstances))

	// remove all AppInsts to prep for next test (as if done by user)
	pt1.goDoAppInsts(t, ctx, &app, cloudcommon.Delete, "")
	pt1.expectAppInsts(t, ctx, &app, 0)

	pt2.goDoAppInsts(t, ctx, &app, cloudcommon.Delete, "")
	pt2.expectAppInsts(t, ctx, &app, 0)

	// *** Check Orphaned Reason ***

	// create instances manually on all cloudlets
	pt1.goDoAppInsts(t, ctx, &app, cloudcommon.Create, "")
	pt1.expectAppInsts(t, ctx, &app, numCloudlets1)

	pt2.goDoAppInsts(t, ctx, &app, cloudcommon.Create, "")
	pt2.expectAppInsts(t, ctx, &app, numCloudlets2)

	// remove cloudlets from policies
	rmCount1 := 3
	pt1.policy.Cloudlets = pt1.policy.Cloudlets[rmCount1:]
	pt1.policy.Fields = []string{
		edgeproto.AutoProvPolicyFieldCloudlets,
		edgeproto.AutoProvPolicyFieldCloudletsKey,
		edgeproto.AutoProvPolicyFieldCloudletsKeyOrganization,
		edgeproto.AutoProvPolicyFieldCloudletsKeyName}
	_, err = autoProvPolicyApi.UpdateAutoProvPolicy(ctx, &pt1.policy)
	require.Nil(t, err)

	rmCount2 := 7
	pt2.policy.Cloudlets = pt2.policy.Cloudlets[rmCount2:]
	pt2.policy.Fields = pt1.policy.Fields
	_, err = autoProvPolicyApi.UpdateAutoProvPolicy(ctx, &pt2.policy)
	require.Nil(t, err)

	// spawn threads to delete
	// this should delete all the AppInsts that are on removed
	// cloudlets, but leave the rest untouched.
	pt1.goDoAppInsts(t, ctx, &app, cloudcommon.Delete, cloudcommon.AutoProvReasonOrphaned)
	pt1.expectAppInsts(t, ctx, &app, numCloudlets1-rmCount1)

	pt2.goDoAppInsts(t, ctx, &app, cloudcommon.Delete, cloudcommon.AutoProvReasonOrphaned)
	pt2.expectAppInsts(t, ctx, &app, numCloudlets2-rmCount2)

	// remove all AppInsts for clean up
	pt1.goDoAppInsts(t, ctx, &app, cloudcommon.Delete, "")
	pt1.expectAppInsts(t, ctx, &app, 0)

	pt2.goDoAppInsts(t, ctx, &app, cloudcommon.Delete, "")
	pt2.expectAppInsts(t, ctx, &app, 0)

	// Check limit only 1 App instance can be auto-provisioned per cloudlet,
	// regardless of the number of reservable ClusterInsts available.
	pt3.expectAppInsts(t, ctx, &app, 0)
	// Reason Demand
	pt3.goDoAppInsts(t, ctx, &app, cloudcommon.Create, cloudcommon.AutoProvReasonDemand)
	pt3.expectAppInsts(t, ctx, &app, len(pt3.cloudlets))
	pt3.goDoAppInsts(t, ctx, &app, cloudcommon.Delete, "")
	pt3.expectAppInsts(t, ctx, &app, 0)
	// Reason MinMax
	pt3.goDoAppInsts(t, ctx, &app, cloudcommon.Create, cloudcommon.AutoProvReasonMinMax)
	pt3.expectAppInsts(t, ctx, &app, len(pt3.cloudlets))
	pt3.goDoAppInsts(t, ctx, &app, cloudcommon.Delete, "")
	pt3.expectAppInsts(t, ctx, &app, 0)
	// Manual create should not limit them.
	pt3.goDoAppInsts(t, ctx, &app, cloudcommon.Create, "")
	pt3.expectAppInsts(t, ctx, &app, len(pt3.clusterInsts))
	pt3.goDoAppInsts(t, ctx, &app, cloudcommon.Delete, "")
	pt3.expectAppInsts(t, ctx, &app, 0)

	// cleanup all data
	_, err = appApi.DeleteApp(ctx, &app)
	require.Nil(t, err)

	pt2.delete(t, ctx)
	pt1.delete(t, ctx)
}

type autoProvPolicyTest struct {
	policy        edgeproto.AutoProvPolicy
	cloudlets     []edgeproto.Cloudlet
	cloudletInfos []edgeproto.CloudletInfo
	clusterInsts  []edgeproto.ClusterInst
}

// AutoProvPolicy and supporting data for test
func newAutoProvPolicyTest(name, org string, count int, flavor *edgeproto.Flavor) *autoProvPolicyTest {
	s := autoProvPolicyTest{}
	s.policy.Key.Name = name
	s.policy.Key.Organization = org
	s.cloudlets = make([]edgeproto.Cloudlet, count, count)
	s.cloudletInfos = make([]edgeproto.CloudletInfo, count, count)
	s.clusterInsts = make([]edgeproto.ClusterInst, count, count)
	for ii, _ := range s.cloudlets {
		s.cloudlets[ii].Key.Name = fmt.Sprintf("%s-%d", name, ii)
		s.cloudlets[ii].Key.Organization = "op"
		s.cloudlets[ii].NumDynamicIps = 20
		s.cloudlets[ii].Location.Latitude = 1
		s.cloudlets[ii].Location.Longitude = 1
		s.cloudlets[ii].CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
		s.cloudlets[ii].PlatformType = edgeproto.PlatformType_PLATFORM_TYPE_FAKE
		s.policy.Cloudlets = append(s.policy.Cloudlets,
			&edgeproto.AutoProvCloudlet{Key: s.cloudlets[ii].Key})
	}
	for ii, _ := range s.cloudletInfos {
		s.cloudletInfos[ii].Key = s.cloudlets[ii].Key
		s.cloudletInfos[ii].State = dme.CloudletState_CLOUDLET_STATE_READY
		s.cloudletInfos[ii].Flavors = []*edgeproto.FlavorInfo{
			&edgeproto.FlavorInfo{
				Name:  flavor.Key.Name,
				Vcpus: flavor.Vcpus,
				Ram:   flavor.Ram,
				Disk:  flavor.Disk,
			},
		}
	}
	for ii, _ := range s.clusterInsts {
		s.clusterInsts[ii].Key.CloudletKey = s.cloudlets[ii].Key
		s.clusterInsts[ii].Key.ClusterKey.Name = fmt.Sprintf("%s-%d", name, ii)
		s.clusterInsts[ii].Key.Organization = org
		s.clusterInsts[ii].Flavor = flavor.Key
		s.clusterInsts[ii].NumMasters = 1
		s.clusterInsts[ii].NumNodes = 1
	}
	return &s
}

func (s *autoProvPolicyTest) create(t *testing.T, ctx context.Context) {
	for ii, _ := range s.cloudlets {
		err := cloudletApi.CreateCloudlet(&s.cloudlets[ii], testutil.NewCudStreamoutCloudlet(ctx))
		require.Nil(t, err)
	}
	for ii, _ := range s.cloudletInfos {
		cloudletInfoApi.Update(ctx, &s.cloudletInfos[ii], 0)
	}
	for ii, _ := range s.clusterInsts {
		err := clusterInstApi.CreateClusterInst(&s.clusterInsts[ii], testutil.NewCudStreamoutClusterInst(ctx))
		require.Nil(t, err)
	}
	_, err := autoProvPolicyApi.CreateAutoProvPolicy(ctx, &s.policy)
	require.Nil(t, err)
}

func (s *autoProvPolicyTest) delete(t *testing.T, ctx context.Context) {
	_, err := autoProvPolicyApi.DeleteAutoProvPolicy(ctx, &s.policy)
	require.Nil(t, err)
	for ii, _ := range s.clusterInsts {
		err := clusterInstApi.DeleteClusterInst(&s.clusterInsts[ii], testutil.NewCudStreamoutClusterInst(ctx))
		require.Nil(t, err)
	}
	for ii, _ := range s.cloudlets {
		err := cloudletApi.DeleteCloudlet(&s.cloudlets[ii], testutil.NewCudStreamoutCloudlet(ctx))
		require.Nil(t, err)
	}
	for ii, _ := range s.cloudletInfos {
		cloudletInfoApi.Delete(ctx, &s.cloudletInfos[ii], 0)
	}
}

func (s *autoProvPolicyTest) goDoAppInsts(t *testing.T, ctx context.Context, app *edgeproto.App, action cloudcommon.Action, reason string) {
	var wg sync.WaitGroup
	if reason != "" {
		// This impersonates the AutoProv service by setting the
		// grpc metadata. We also spawn threads to simulate race
		// conditions. The test expects some of these to fail so
		// we do not check the error result.
		md := metadata.Pairs(
			cloudcommon.CallerAutoProv, "",
			cloudcommon.AutoProvReason, reason,
			cloudcommon.AutoProvPolicyName, s.policy.Key.Name)
		ctx = metadata.NewIncomingContext(ctx, md)
	}
	for ii, _ := range s.clusterInsts {
		wg.Add(1)
		go func(ii int) {
			inst := edgeproto.AppInst{}
			inst.Key.AppKey = app.Key
			inst.Key.ClusterInstKey = s.clusterInsts[ii].Key
			var err error
			if action == cloudcommon.Create {
				err = appInstApi.CreateAppInst(&inst, testutil.NewCudStreamoutAppInst(ctx))
			} else if action == cloudcommon.Delete {
				err = appInstApi.DeleteAppInst(&inst, testutil.NewCudStreamoutAppInst(ctx))
			}
			log.SpanLog(ctx, log.DebugLevelApi, "goDoAppInsts", "action", action.String(), "key", inst.Key, "err", err)
			wg.Done()
		}(ii)
	}
	wg.Wait()
}

func (s *autoProvPolicyTest) expectAppInsts(t *testing.T, ctx context.Context, app *edgeproto.App, expected int) {
	// Count the number of AppInsts on Cloudlets for this policy.
	actual := 0
	for ii, _ := range s.clusterInsts {
		instKey := edgeproto.AppInstKey{}
		instKey.AppKey = app.Key
		instKey.ClusterInstKey = s.clusterInsts[ii].Key
		if appInstApi.cache.HasKey(&instKey) {
			actual++
		}
	}
	require.Equal(t, expected, actual, "expected %d insts for policy %s but found %d", expected, s.policy.Key.Name, actual)
}
