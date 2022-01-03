package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
)

func TestAddRefsChecks(t *testing.T) {
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

	dataGen := AddRefsDataGen{}
	allAddRefsChecks(t, ctx, apis, &dataGen)
}

type AddRefsDataGen struct{}

func (s *AddRefsDataGen) GetAddAppAlertPolicyTestObj() (*edgeproto.AppAlertPolicy, *testSupportData) {
	app := testutil.AppData[0]
	app.AlertPolicies = nil
	alertPolicy := testutil.AlertPolicyData[0]

	testObj := edgeproto.AppAlertPolicy{
		AppKey:      app.Key,
		AlertPolicy: alertPolicy.Key.Name,
	}
	supportData := &testSupportData{}
	supportData.Apps = []edgeproto.App{app}
	supportData.AlertPolicies = []edgeproto.AlertPolicy{alertPolicy}
	return &testObj, supportData
}

func (s *AddRefsDataGen) GetAddAppAutoProvPolicyTestObj() (*edgeproto.AppAutoProvPolicy, *testSupportData) {
	app := testutil.AppData[0]
	app.Deployment = cloudcommon.DeploymentTypeKubernetes
	app.AutoProvPolicies = nil
	autoProvPolicy := testutil.AutoProvPolicyData[0]

	testObj := edgeproto.AppAutoProvPolicy{
		AppKey:         app.Key,
		AutoProvPolicy: autoProvPolicy.Key.Name,
	}
	supportData := &testSupportData{}
	supportData.Apps = []edgeproto.App{app}
	supportData.AutoProvPolicies = []edgeproto.AutoProvPolicy{autoProvPolicy}
	return &testObj, supportData
}

func (s *AddRefsDataGen) GetAddAutoProvPolicyCloudletTestObj() (*edgeproto.AutoProvPolicyCloudlet, *testSupportData) {
	cloudlet := testutil.CloudletData()[0]
	autoProvPolicy := testutil.AutoProvPolicyData[0]
	autoProvPolicy.Cloudlets = nil

	testObj := edgeproto.AutoProvPolicyCloudlet{
		Key:         autoProvPolicy.Key,
		CloudletKey: cloudlet.Key,
	}
	supportData := &testSupportData{}
	supportData.Cloudlets = []edgeproto.Cloudlet{cloudlet}
	supportData.AutoProvPolicies = []edgeproto.AutoProvPolicy{autoProvPolicy}
	return &testObj, supportData
}

func (s *AddRefsDataGen) GetAddCloudletPoolMemberTestObj() (*edgeproto.CloudletPoolMember, *testSupportData) {
	cloudlet := testutil.CloudletData()[0]
	cloudletPool := testutil.CloudletPoolData[0]
	cloudletPool.Key.Organization = cloudlet.Key.Organization
	cloudletPool.Cloudlets = nil

	testObj := edgeproto.CloudletPoolMember{
		Key:          cloudletPool.Key,
		CloudletName: cloudlet.Key.Name,
	}
	supportData := &testSupportData{}
	supportData.Cloudlets = []edgeproto.Cloudlet{cloudlet}
	supportData.CloudletPools = []edgeproto.CloudletPool{cloudletPool}
	return &testObj, supportData
}

func (s *AddRefsDataGen) GetAddCloudletResMappingTestObj() (*edgeproto.CloudletResMap, *testSupportData) {
	cloudlet := testutil.CloudletData()[0]
	cloudlet.ResTagMap = nil
	resTagTable := testutil.ResTagTableData[0]

	testObj := edgeproto.CloudletResMap{
		Key: cloudlet.Key,
		Mapping: map[string]string{
			"gpu": resTagTable.Key.Name,
		},
	}
	supportData := &testSupportData{}
	supportData.Cloudlets = []edgeproto.Cloudlet{cloudlet}
	supportData.ResTagTables = []edgeproto.ResTagTable{resTagTable}
	return &testObj, supportData
}

func (s *AddRefsDataGen) GetCreateAppTestObj() (*edgeproto.App, *testSupportData) {
	flavor := testutil.FlavorData[0]
	autoProvPolicy := testutil.AutoProvPolicyData[0]
	alertPolicy := testutil.AlertPolicyData[0]

	app := testutil.AppData[0]
	app.DefaultFlavor = flavor.Key
	app.AutoProvPolicies = []string{autoProvPolicy.Key.Name}
	app.AlertPolicies = []string{alertPolicy.Key.Name}

	supportData := &testSupportData{}
	supportData.Flavors = []edgeproto.Flavor{flavor}
	supportData.AutoProvPolicies = []edgeproto.AutoProvPolicy{autoProvPolicy}
	supportData.AlertPolicies = []edgeproto.AlertPolicy{alertPolicy}
	return &app, supportData
}

func (s *AddRefsDataGen) GetCreateAppInstTestObj() (*edgeproto.AppInst, *testSupportData) {
	app := testutil.AppData[0]
	cloudlet := testutil.CloudletData()[0]
	cloudletInfo := testutil.CloudletInfoData[0]
	clusterInst := testutil.ClusterInstData[0]
	clusterInst.Key.CloudletKey = cloudlet.Key
	clusterInst.State = edgeproto.TrackedState_READY
	flavor := testutil.FlavorData[0]

	appInst := testutil.AppInstData[0]
	appInst.Key.AppKey = app.Key
	appInst.Key.ClusterInstKey = *clusterInst.Key.Virtual("")
	appInst.Flavor = flavor.Key
	appInst.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM

	supportData := &testSupportData{}
	supportData.Apps = []edgeproto.App{app}
	supportData.Cloudlets = []edgeproto.Cloudlet{cloudlet}
	supportData.CloudletInfos = []edgeproto.CloudletInfo{cloudletInfo}
	supportData.ClusterInsts = []edgeproto.ClusterInst{clusterInst}
	supportData.Flavors = []edgeproto.Flavor{flavor}
	return &appInst, supportData
}

func (s *AddRefsDataGen) GetCreateAutoProvPolicyTestObj() (*edgeproto.AutoProvPolicy, *testSupportData) {
	cloudlet := testutil.CloudletData()[0]

	autoProvPolicy := testutil.AutoProvPolicyData[0]
	autoProvPolicy.Cloudlets = []*edgeproto.AutoProvCloudlet{
		&edgeproto.AutoProvCloudlet{
			Key: cloudlet.Key,
		},
	}

	supportData := &testSupportData{}
	supportData.Cloudlets = []edgeproto.Cloudlet{cloudlet}
	return &autoProvPolicy, supportData
}

func (s *AddRefsDataGen) GetCreateCloudletTestObj() (*edgeproto.Cloudlet, *testSupportData) {
	// must use Cloudlet[2] because TrustPolicy validation does not
	// allow special characters in org name.
	cloudlet := testutil.CloudletData()[2]

	flavor := testutil.FlavorData[0]
	resTagTable := testutil.ResTagTableData[0]
	resTagTable.Key.Organization = cloudlet.Key.Organization
	trustPolicy := testutil.TrustPolicyData[0]
	trustPolicy.Key.Organization = cloudlet.Key.Organization
	gpuDriver := testutil.GPUDriverData[0]
	gpuDriver.Key.Organization = cloudlet.Key.Organization
	vmpool := testutil.VMPoolData[0]
	vmpool.Key.Organization = cloudlet.Key.Organization

	cloudlet.Flavor = flavor.Key
	cloudlet.ResTagMap = map[string]*edgeproto.ResTagTableKey{"gpu": &resTagTable.Key}
	cloudlet.TrustPolicy = trustPolicy.Key.Name
	cloudlet.GpuConfig.Driver = gpuDriver.Key
	cloudlet.VmPool = vmpool.Key.Name
	cloudlet.PlatformType = edgeproto.PlatformType_PLATFORM_TYPE_FAKE_VM_POOL
	cloudlet.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM

	supportData := &testSupportData{}
	supportData.Flavors = []edgeproto.Flavor{flavor}
	supportData.ResTagTables = []edgeproto.ResTagTable{resTagTable}
	supportData.TrustPolicies = []edgeproto.TrustPolicy{trustPolicy}
	supportData.GpuDrivers = []edgeproto.GPUDriver{gpuDriver}
	supportData.VmPools = []edgeproto.VMPool{vmpool}
	return &cloudlet, supportData
}

func (s *AddRefsDataGen) GetCreateCloudletPoolTestObj() (*edgeproto.CloudletPool, *testSupportData) {
	cloudlet := testutil.CloudletData()[0]

	cloudletPool := testutil.CloudletPoolData[0]
	cloudletPool.Key.Organization = cloudlet.Key.Organization
	cloudletPool.Cloudlets = []string{cloudlet.Key.Name}

	supportData := &testSupportData{}
	supportData.Cloudlets = []edgeproto.Cloudlet{cloudlet}
	return &cloudletPool, supportData
}

func (s *AddRefsDataGen) GetCreateClusterInstTestObj() (*edgeproto.ClusterInst, *testSupportData) {
	cloudlet := testutil.CloudletData()[0]
	cloudletInfo := testutil.CloudletInfoData[0]
	flavor := testutil.FlavorData[0]
	autoScalePolicy := testutil.AutoScalePolicyData[0]
	network := testutil.NetworkData[0]
	network.Key.CloudletKey = cloudlet.Key

	clusterInst := testutil.ClusterInstData[0]
	clusterInst.Key.CloudletKey = cloudlet.Key
	clusterInst.Flavor = flavor.Key
	clusterInst.AutoScalePolicy = autoScalePolicy.Key.Name
	clusterInst.Networks = []string{network.Key.Name}
	clusterInst.CrmOverride = edgeproto.CRMOverride_IGNORE_CRM
	clusterInst.Deployment = cloudcommon.DeploymentTypeKubernetes
	clusterInst.State = edgeproto.TrackedState_READY

	supportData := &testSupportData{}
	supportData.Cloudlets = []edgeproto.Cloudlet{cloudlet}
	supportData.CloudletInfos = []edgeproto.CloudletInfo{cloudletInfo}
	supportData.Flavors = []edgeproto.Flavor{flavor}
	supportData.AutoScalePolicies = []edgeproto.AutoScalePolicy{autoScalePolicy}
	supportData.Networks = []edgeproto.Network{network}
	return &clusterInst, supportData
}

func (s *AddRefsDataGen) GetCreateNetworkTestObj() (*edgeproto.Network, *testSupportData) {
	cloudlet := testutil.CloudletData()[0]

	network := testutil.NetworkData[0]
	network.Key.CloudletKey = cloudlet.Key

	supportData := &testSupportData{}
	supportData.Cloudlets = []edgeproto.Cloudlet{cloudlet}
	return &network, supportData
}

func (s *AddRefsDataGen) GetCreateTrustPolicyExceptionTestObj() (*edgeproto.TrustPolicyException, *testSupportData) {
	cloudletPool := testutil.CloudletPoolData[0]
	app := testutil.AppData[0]

	tpe := testutil.TrustPolicyExceptionData[0]
	tpe.Key.AppKey = app.Key
	tpe.Key.CloudletPoolKey = cloudletPool.Key

	supportData := &testSupportData{}
	supportData.CloudletPools = []edgeproto.CloudletPool{cloudletPool}
	supportData.Apps = []edgeproto.App{app}
	return &tpe, supportData
}

func (s *AddRefsDataGen) GetUpdateAppTestObj() (*edgeproto.App, *testSupportData) {
	testObj, supportData := s.GetCreateAppTestObj()
	// copy and clear refs
	updatable := *testObj
	updatable.DefaultFlavor = edgeproto.FlavorKey{}
	updatable.AutoProvPolicies = []string{}
	updatable.AlertPolicies = []string{}

	supportData.Apps = []edgeproto.App{updatable}

	testObj.Fields = []string{
		edgeproto.AppFieldDefaultFlavor,
		edgeproto.AppFieldDefaultFlavorName,
		edgeproto.AppFieldAutoProvPolicies,
		edgeproto.AppFieldAlertPolicies,
	}
	return testObj, supportData
}

func (s *AddRefsDataGen) GetUpdateAppInstTestObj() (*edgeproto.AppInst, *testSupportData) {
	testObj, supportData := s.GetCreateAppInstTestObj()
	// copy and clear refs
	updatable := *testObj
	updatable.Flavor = edgeproto.FlavorKey{}

	supportData.AppInstances = []edgeproto.AppInst{updatable}

	testObj.Fields = []string{
		edgeproto.AppInstFieldFlavor,
		edgeproto.AppInstFieldFlavorName,
	}
	return testObj, supportData
}

func (s *AddRefsDataGen) GetUpdateAutoProvPolicyTestObj() (*edgeproto.AutoProvPolicy, *testSupportData) {
	testObj, supportData := s.GetCreateAutoProvPolicyTestObj()
	// copy and clear refs
	updatable := *testObj
	updatable.Cloudlets = nil

	supportData.AutoProvPolicies = []edgeproto.AutoProvPolicy{updatable}

	testObj.Fields = []string{
		edgeproto.AutoProvPolicyFieldCloudlets,
		edgeproto.AutoProvPolicyFieldCloudletsKey,
		edgeproto.AutoProvPolicyFieldCloudletsKeyName,
		edgeproto.AutoProvPolicyFieldCloudletsKeyOrganization,
	}
	return testObj, supportData
}

func (s *AddRefsDataGen) GetUpdateCloudletTestObj() (*edgeproto.Cloudlet, *testSupportData) {
	testObj, supportData := s.GetCreateCloudletTestObj()
	// copy and clear refs
	updatable := *testObj
	updatable.Flavor = edgeproto.FlavorKey{}
	updatable.ResTagMap = nil
	updatable.TrustPolicy = ""
	updatable.GpuConfig.Driver = edgeproto.GPUDriverKey{}

	supportData.Cloudlets = []edgeproto.Cloudlet{updatable}
	supportData.CloudletInfos = []edgeproto.CloudletInfo{testutil.CloudletInfoData[2]}

	testObj.Fields = []string{
		edgeproto.CloudletFieldTrustPolicy,
		edgeproto.CloudletFieldGpuConfigDriver,
		edgeproto.CloudletFieldGpuConfigDriverName,
		edgeproto.CloudletFieldGpuConfigDriverOrganization,
	}
	return testObj, supportData
}

func (s *AddRefsDataGen) GetUpdateCloudletPoolTestObj() (*edgeproto.CloudletPool, *testSupportData) {
	testObj, supportData := s.GetCreateCloudletPoolTestObj()
	// copy and clear refs
	updatable := *testObj
	updatable.Cloudlets = nil

	supportData.CloudletPools = []edgeproto.CloudletPool{updatable}

	testObj.Fields = []string{edgeproto.CloudletPoolFieldCloudlets}
	return testObj, supportData
}

func (s *AddRefsDataGen) GetUpdateClusterInstTestObj() (*edgeproto.ClusterInst, *testSupportData) {
	testObj, supportData := s.GetCreateClusterInstTestObj()
	// copy and clear refs
	updatable := *testObj
	updatable.Flavor = edgeproto.FlavorKey{}
	updatable.AutoScalePolicy = ""
	updatable.Networks = nil

	supportData.ClusterInsts = []edgeproto.ClusterInst{updatable}

	testObj.Fields = []string{
		edgeproto.ClusterInstFieldAutoScalePolicy,
	}
	return testObj, supportData
}
