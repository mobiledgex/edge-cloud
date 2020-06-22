package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

type stateTransition struct {
	triggerState   edgeproto.CloudletState
	triggerVersion string
	expectedState  edgeproto.TrackedState
	ignoreState    bool
}

const (
	crm_v1 = "2001-01-31"
	crm_v2 = "2002-01-31"
)

func TestCloudletApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi | log.DebugLevelNotify)
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testinit()

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	// create flavors
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)

	testutil.InternalCloudletTest(t, "cud", &cloudletApi, testutil.CloudletData)

	// test invalid location values
	clbad := testutil.CloudletData[0]
	clbad.Key.Name = "bad loc"
	testBadLat(t, ctx, &clbad, []float64{0, 90.1, -90.1, -1323213, 1232334}, "create")
	testBadLong(t, ctx, &clbad, []float64{0, 180.1, -180.1, -1323213, 1232334}, "create")

	clbad = testutil.CloudletData[0]
	clbad.Key.Name = "test num dyn ips"
	err := cloudletApi.CreateCloudlet(&clbad, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)
	clbad.NumDynamicIps = 0
	clbad.Fields = []string{edgeproto.CloudletFieldNumDynamicIps}
	err = cloudletApi.UpdateCloudlet(&clbad, testutil.NewCudStreamoutCloudlet(ctx))
	require.NotNil(t, err)

	cl := testutil.CloudletData[1]
	cl.Key.Name = "test invalid lat-long"
	err = cloudletApi.CreateCloudlet(&cl, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)
	testBadLat(t, ctx, &cl, []float64{0, 90.1, -90.1, -1323213, 1232334}, "update")
	testBadLong(t, ctx, &cl, []float64{0, 180.1, -180.1, -1323213, 1232334}, "update")

	// Resource Mapping tests

	testResMapKeysApi(t, ctx, &cl)
	testGpuResourceMapping(t, ctx, &cl)

	// Cloudlet state tests
	testCloudletStates(t, ctx)
	testManualBringup(t, ctx)

	dummy.Stop()
}

func testBadLat(t *testing.T, ctx context.Context, clbad *edgeproto.Cloudlet, lats []float64, action string) {
	for _, lat := range lats {
		clbad.Location.Latitude = lat
		clbad.Fields = []string{edgeproto.CloudletFieldLocationLatitude}
		switch action {
		case "create":
			err := cloudletApi.CreateCloudlet(clbad, testutil.NewCudStreamoutCloudlet(ctx))
			require.NotNil(t, err, "create cloudlet bad latitude")
		case "update":
			err := cloudletApi.UpdateCloudlet(clbad, testutil.NewCudStreamoutCloudlet(ctx))
			require.NotNil(t, err, "update cloudlet bad latitude")
		}
	}
}

func testBadLong(t *testing.T, ctx context.Context, clbad *edgeproto.Cloudlet, longs []float64, action string) {
	for _, long := range longs {
		clbad.Location.Longitude = long
		clbad.Fields = []string{edgeproto.CloudletFieldLocationLongitude}
		switch action {
		case "create":
			err := cloudletApi.CreateCloudlet(clbad, testutil.NewCudStreamoutCloudlet(ctx))
			require.NotNil(t, err, "create cloudlet bad longitude")
		case "update":
			err := cloudletApi.CreateCloudlet(clbad, testutil.NewCudStreamoutCloudlet(ctx))
			require.NotNil(t, err, "update cloudlet bad longitude")
		}
	}
}

func waitForState(key *edgeproto.CloudletKey, state edgeproto.TrackedState) error {
	lastState := edgeproto.TrackedState_TRACKED_STATE_UNKNOWN
	for i := 0; i < 10; i++ {
		cloudlet := edgeproto.Cloudlet{}
		if cloudletApi.cache.Get(key, &cloudlet) {
			if cloudlet.State == state {
				return nil
			}
			lastState = cloudlet.State
		}
		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("Unable to get desired cloudlet state, actual state %s, desired state %s", lastState, state)
}

func forceCloudletInfoState(ctx context.Context, key *edgeproto.CloudletKey, state edgeproto.CloudletState, version string) {
	info := edgeproto.CloudletInfo{}
	info.Key = *key
	info.State = state
	info.ContainerVersion = version
	cloudletInfoApi.Update(ctx, &info, 0)
}

func testNotifyId(t *testing.T, ctrlHandler *notify.DummyHandler, key *edgeproto.CloudletKey, nodeCount, notifyId int, crmVersion string) {
	require.Equal(t, nodeCount, len(ctrlHandler.NodeCache.Objs), "node count matches")
	nodeVersion, nodeNotifyId, err := ctrlHandler.GetCloudletDetails(key)
	require.Nil(t, err, "get cloudlet version & notifyId from node cache")
	require.Equal(t, crmVersion, nodeVersion, "node version matches")
	require.Equal(t, int64(notifyId), nodeNotifyId, "node notifyId matches")
}

func testCloudletStates(t *testing.T, ctx context.Context) {
	ctrlHandler := notify.NewDummyHandler()
	ctrlMgr := notify.ServerMgr{}
	ctrlHandler.RegisterServer(&ctrlMgr)
	ctrlMgr.Start("127.0.0.1:50001", nil)
	defer ctrlMgr.Stop()

	crm_notifyaddr := "127.0.0.1:0"
	cloudlet := testutil.CloudletData[2]
	cloudlet.ContainerVersion = crm_v1
	cloudlet.Key.Name = "testcloudletstates"
	cloudlet.NotifySrvAddr = crm_notifyaddr
	pfConfig, err := getPlatformConfig(ctx, &cloudlet)
	require.Nil(t, err, "get platform config")

	err = cloudcommon.StartCRMService(ctx, &cloudlet, pfConfig)
	require.Nil(t, err, "start cloudlet")
	defer func() {
		// Delete CRM
		err = cloudcommon.StopCRMService(ctx, &cloudlet)
		require.Nil(t, err, "stop cloudlet")
	}()

	err = ctrlHandler.WaitForCloudletState(&cloudlet.Key, edgeproto.CloudletState_CLOUDLET_STATE_INIT, crm_v1)
	require.Nil(t, err, "cloudlet state transition")

	cloudlet.State = edgeproto.TrackedState_CRM_INITOK
	ctrlHandler.CloudletCache.Update(ctx, &cloudlet, 0)

	err = ctrlHandler.WaitForCloudletState(&cloudlet.Key, edgeproto.CloudletState_CLOUDLET_STATE_READY, crm_v1)
	require.Nil(t, err, "cloudlet state transition")

	cloudlet.State = edgeproto.TrackedState_READY
	ctrlHandler.CloudletCache.Update(ctx, &cloudlet, 0)
}

func testManualBringup(t *testing.T, ctx context.Context) {
	var err error
	cloudlet := testutil.CloudletData[2]
	cloudlet.Key.Name = "crmmanualbringup"
	cloudlet.ContainerVersion = crm_v1
	err = cloudletApi.CreateCloudlet(&cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)

	err = waitForState(&cloudlet.Key, edgeproto.TrackedState_READY)
	require.Nil(t, err, "cloudlet obj created")

	forceCloudletInfoState(ctx, &cloudlet.Key, edgeproto.CloudletState_CLOUDLET_STATE_INIT, crm_v2)
	err = waitForState(&cloudlet.Key, edgeproto.TrackedState_CRM_INITOK)
	require.Nil(t, err, fmt.Sprintf("cloudlet state transtions"))

	forceCloudletInfoState(ctx, &cloudlet.Key, edgeproto.CloudletState_CLOUDLET_STATE_READY, crm_v2)
	err = waitForState(&cloudlet.Key, edgeproto.TrackedState_READY)
	require.Nil(t, err, fmt.Sprintf("cloudlet state transtions"))

	err = cloudletApi.DeleteCloudlet(&cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)
}

func testResMapKeysApi(t *testing.T, ctx context.Context, cl *edgeproto.Cloudlet) {
	// We can add/remove edgeproto.ResTagTableKey values to the cl.ResTagMap map
	// which then can be used in the GetVMSpec call when matching our meta-resource specificer
	// to a deployments actual resources/flavrs.
	resmap := edgeproto.CloudletResMap{}
	resmap.Key = cl.Key
	// test_data contains sample resource tag maps, add them to the cloudlet
	// verify, and remove them. ClI should follow suit.
	if cl.ResTagMap == nil {
		cl.ResTagMap = make(map[string]*edgeproto.ResTagTableKey)
	}
	if resmap.Mapping == nil {
		resmap.Mapping = make(map[string]string)
	}

	// use the OptResNames as clould.ResTagMap[key] = tblkey in test
	// gpu, nas and nic are the current set of Resource Names.
	// setup the test map using the test_data objects
	// The AddCloudResMapKey is setup to accept multiple res tbl keys at once
	// but we're doing it one by one.

	resmap.Mapping[strings.ToLower(edgeproto.OptResNames_name[0])] = testutil.Restblkeys[0].Name
	_, err := cloudletApi.AddCloudletResMapping(ctx, &resmap)
	require.Nil(t, err, "AddCloudletResMapKey")

	resmap.Mapping[strings.ToLower(edgeproto.OptResNames_name[1])] = testutil.Restblkeys[1].Name
	_, err = cloudletApi.AddCloudletResMapping(ctx, &resmap)
	require.Nil(t, err, "AddCloudletResMapKey")

	resmap.Mapping[strings.ToLower(edgeproto.OptResNames_name[2])] = testutil.Restblkeys[2].Name
	_, err = cloudletApi.AddCloudletResMapping(ctx, &resmap)
	require.Nil(t, err, "AddCloudletResMapKey")

	testcl := &edgeproto.Cloudlet{}
	// now it's all stored, fetch a copy of the cloudlet and verify
	err = cloudletApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !cloudletApi.store.STMGet(stm, &cl.Key, testcl) {
			return cl.Key.NotFoundError()
		}
		return err
	})

	// what's in our testcl? Check the resource map
	tkey := testcl.ResTagMap[strings.ToLower(edgeproto.OptResNames_name[0])]
	require.Equal(t, testutil.Restblkeys[0].Name, tkey.Name, "AddCloudletResMapKey")
	tkey = testcl.ResTagMap[strings.ToLower(edgeproto.OptResNames_name[1])]
	require.Equal(t, testutil.Restblkeys[1].Name, tkey.Name, "AddCloudletResMapKey")
	tkey = testcl.ResTagMap[strings.ToLower(edgeproto.OptResNames_name[2])]
	require.Equal(t, testutil.Restblkeys[2].Name, tkey.Name, "AddCloudletResMapKey")

	// and the actual keys should match as well
	require.Equal(t, testutil.Restblkeys[0], *testcl.ResTagMap[testutil.Restblkeys[0].Name], "AddCloudletResMapKey")
	require.Equal(t, testutil.Restblkeys[1], *testcl.ResTagMap[testutil.Restblkeys[1].Name], "AddCloudletResMapKey")
	require.Equal(t, testutil.Restblkeys[2], *testcl.ResTagMap[testutil.Restblkeys[2].Name], "AddCloudletResMapKey")

	resmap1 := edgeproto.CloudletResMap{}
	resmap1.Mapping = make(map[string]string)
	resmap1.Mapping[strings.ToLower(edgeproto.OptResNames_name[2])] = testutil.Restblkeys[2].Name
	resmap1.Mapping[strings.ToLower(edgeproto.OptResNames_name[1])] = testutil.Restblkeys[1].Name
	resmap1.Key = cl.Key

	_, err = cloudletApi.RemoveCloudletResMapping(ctx, &resmap1)
	require.Nil(t, err, "RemoveCloudletResMapKey")

	rmcl := &edgeproto.Cloudlet{}
	if rmcl.ResTagMap == nil {
		rmcl.ResTagMap = make(map[string]*edgeproto.ResTagTableKey)
	}
	rmcl.Key = resmap1.Key

	err = cloudletApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !cloudletApi.store.STMGet(stm, &cl.Key, rmcl) {
			return cl.Key.NotFoundError()
		}
		return err
	})

	require.Nil(t, err, "STMGet failure")
	// and check the maps len = 1
	require.Equal(t, 1, len(rmcl.ResTagMap), "RemoveCloudletResMapKey")
	// and might as well check the key "gpu" exists
	_, ok := rmcl.ResTagMap[testutil.Restblkeys[0].Name]
	require.Equal(t, true, ok, "RemoveCloudletResMapKey")
}

func testGpuResourceMapping(t *testing.T, ctx context.Context, cl *edgeproto.Cloudlet) {
	// Cloudlet has a map key'ed by resource name/type whose value is a res tag tbl key.
	// We init this map, and create a resource table, and place its key into this map
	// and pass this map to the matcher routine, this allows the matcher to have access
	// to all optional resource tag maps present in the cloudlet. A meta-flavor has a
	// similar map to request generic resources that need to be mapped to specific
	// platform resources. We create such a edgeproto.Flavor and set it's request
	// map to ask for a gpu and a nas storage volume. The game for the matcher/mapper
	// is to take our meta-flavor resourse request object, and return, for this
	// operator/cloudlet the closest matching available flavor to use in the eventual
	// launch of a suitable image.
	var cli edgeproto.CloudletInfo = testutil.CloudletInfoData[0]

	if cl.ResTagMap == nil {
		cl.ResTagMap = make(map[string]*edgeproto.ResTagTableKey)
	}
	var gputab = edgeproto.ResTagTable{
		Key: edgeproto.ResTagTableKey{
			Name: "gpumap",
		},
		Tags: map[string]string{"vgpu": "nvidia-63:1", "pci": "t4:1", "gpu": "T4:1"},
	}

	var nastab = edgeproto.ResTagTable{
		Key: edgeproto.ResTagTableKey{
			Name: "nasmap",
		},
		Tags: map[string]string{"nas": "ceph:1"},
	}
	_, err := resTagTableApi.CreateResTagTable(ctx, &gputab)
	require.Nil(t, nil, err, "CreateResTagTable")

	// Our clouldets resource map, maps from resource type names, to ResTagTableKeys.
	// The ResTagTableKey is a resource name, and the owning operator key.
	cl.ResTagMap["gpu"] = &gputab.Key

	// We also  need a list of edgeproto.FlavorInfo structs
	// which it so happens we have in the testutils.CloudletInfoData.Flavors array
	tbl1, err := resTagTableApi.GetResTagTable(ctx, &gputab.Key)
	require.Nil(t, err, "GetResTagTable")
	require.Equal(t, 3, len(tbl1.Tags), "tag count mismatch")

	// specify a pci pass_throuh, don't care what kind
	// should match flavor.large-pci
	var flavorPciMatch = edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "x1.large-pci-mex",
		},
		Ram:   8192,
		Vcpus: 10,
		Disk:  40,
		// This requests a passthru
		OptResMap: map[string]string{"gpu": "pci:1"},
	}

	// map to a generic nvidia vgpu type, should match flavor.large-nvidia
	var flavorVgpuMatch = edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "x1.large-vgpu-mex",
		},
		Ram:   8192,
		Vcpus: 10,
		Disk:  40,
		// This requests 2 vgpu instances, (not supported by nvidia yet)
		OptResMap: map[string]string{"gpu": "vgpu:1"},
	}
	// don't care what kind of gpu resource

	// don't care what kind of gpu resource
	var testflavor = edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "x1.large-mex",
		},
		Ram:   8192,
		Vcpus: 8,
		Disk:  40,
		// This says I want one gpu, don't care if it's vgpu or passthrough
		OptResMap: map[string]string{"gpu": "gpu:1"},
	}
	// request two optional resources
	var testflavor2 = edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "x1.large-2-Resources",
		},
		Ram:   8192,
		Vcpus: 8,
		Disk:  40,
		// This says I want one gpu, don't care if it's vgpu or passthrough
		OptResMap: map[string]string{"gpu": "gpu:1", "nas": "ceph-20:1"},
	}
	// request nas optional resource only
	var testflavorNas = edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "x1.large-2-Resources",
		},
		Ram:   8192,
		Vcpus: 8,
		Disk:  40,
		// This says I want one gpu, don't care if it's vgpu or passthrough
		OptResMap: map[string]string{"nas": "ceph-20:1"},
	}

	// test request for a specific type of pci  ( one T4 )
	// should match flavor.large from testutils.
	var testPciT4flavor = edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "mex.large-pci-T4",
		},
		Ram:   8192,
		Vcpus: 8,
		Disk:  40,
		// This says I want one gpu, don't care if it's vgpu or passthrough
		OptResMap: map[string]string{"gpu": "pci:t4:1"},
	}

	var flavorVgpuNvidiaMatch = edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "mex.large-vgpu-nvidia-63",
		},
		Ram:   8192,
		Vcpus: 10,
		Disk:  40,
		// This requests 2 vgpu instances, (not supported by nvidia yet)
		OptResMap: map[string]string{"gpu": "vgpu:nvidia-63:1"},
	}
	taz := edgeproto.OSAZone{Name: "AZ1_GPU", Status: "available"}
	timg := edgeproto.OSImage{Name: "gpu_image"}
	cli.AvailabilityZones = append(cli.AvailabilityZones, &taz)
	cli.OsImages = append(cli.OsImages, &timg)

	// testflavor wants some generic GPU resource, it should match
	// the first flavor offering some type of gpu reosurce.
	// We can direct a generic request to a given flavor though,
	// which is the case here.

	err = cloudletApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {

		spec, vmerr := resTagTableApi.GetVMSpec(ctx, stm, testflavor, *cl, cli)
		require.Nil(t, vmerr, "GetVmSpec")
		require.Equal(t, "flavor.large", spec.FlavorName)
		require.Equal(t, "AZ1_GPU", spec.AvailabilityZone)
		require.Equal(t, "gpu_image", spec.ImageName)

		spec, vmerr = resTagTableApi.GetVMSpec(ctx, stm, flavorVgpuMatch, *cl, cli)
		require.Nil(t, vmerr, "GetVmSpec wildcard request")
		require.Equal(t, "flavor.large", spec.FlavorName)

		spec, vmerr = resTagTableApi.GetVMSpec(ctx, stm, flavorPciMatch, *cl, cli)
		require.Nil(t, vmerr, "GetVMSpec")
		require.Equal(t, "flavor.large", spec.FlavorName)

		// non-nominal, ask for more resources than the would-be match supports.
		// change testflavor to request 10 gpus of any kind.
		testflavor.OptResMap["gpu"] = "gpu:10"
		spec, vmerr = resTagTableApi.GetVMSpec(ctx, stm, testflavor, *cl, cli)
		require.Equal(t, "no suitable platform flavor found for x1.large-mex, please try a smaller flavor", vmerr.Error(), "nil table")

		// specific pci passthrough
		spec, vmerr = resTagTableApi.GetVMSpec(ctx, stm, testPciT4flavor, *cl, cli)
		require.Nil(t, vmerr, "GetVmSpec")
		require.Equal(t, "flavor.large", spec.FlavorName)

		spec, vmerr = resTagTableApi.GetVMSpec(ctx, stm, flavorVgpuNvidiaMatch, *cl, cli)
		require.Nil(t, vmerr, "GetVmSpec")
		require.Equal(t, "flavor.large-nvidia", spec.FlavorName)
		uses := resTagTableApi.UsesGpu(ctx, stm, *spec.FlavorInfo, *cl)
		require.Equal(t, true, uses)

		// Now try 2 optional resources requested by one flavor, first non-nominal, no res tag table for nas tags
		spec, vmerr = resTagTableApi.GetVMSpec(ctx, stm, testflavor2, *cl, cli)
		if vmerr != nil {
			require.Equal(t, "no suitable platform flavor found for x1.large-2-Resources, please try a smaller flavor", vmerr.Error())
		}

		// now, add cloudlet mapping for nas to the cloudlet, making the above test nominal...
		cl.ResTagMap["nas"] = &nastab.Key

		// ...and actually create the new nas res tag table
		_, err := resTagTableApi.CreateResTagTable(ctx, &nastab)
		require.Nil(t, err)

		spec, vmerr = resTagTableApi.GetVMSpec(ctx, stm, testflavor2, *cl, cli)
		require.Nil(t, err, "GetVMSpec")
		require.Equal(t, "flavor.large2", spec.FlavorName)

		// Non-nominal: ask for nas only, should reject testflavor2 as there are no
		// os flavors with only a nas resource
		spec, vmerr = resTagTableApi.GetVMSpec(ctx, stm, testflavorNas, *cl, cli)
		require.Equal(t, "no suitable platform flavor found for x1.large-2-Resources, please try a smaller flavor", vmerr.Error())
		// Non-nominal: flavor requests optional resource, while cloudlet's OptResMap is nil (cloudlet supports none)
		cl.ResTagMap = nil
		spec, vmerr = resTagTableApi.GetVMSpec(ctx, stm, testflavor, *cl, cli)
		require.Equal(t, "Optional resource requested by x1.large-mex, cloudlet test invalid lat-long supports none", vmerr.Error())

		nulCL := edgeproto.Cloudlet{}
		// and finally, Non-nominal, request a resource, and cloudlet has none to give (nil cloudlet/cloudlet.ResTagMap)
		spec, vmerr = resTagTableApi.GetVMSpec(ctx, stm, testflavor, nulCL, cli)
		require.Equal(t, "Optional resource requested by x1.large-mex, cloudlet  supports none", vmerr.Error(), "nil table")
		return nil
	})
}
