package main

import (
	"context"
	"testing"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestCloudletApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
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

	// create operators
	testutil.InternalOperatorCreate(t, &operatorApi, testutil.OperatorData)
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

func testGpuResourceMapping(t *testing.T, ctx context.Context, cl *edgeproto.Cloudlet) {
	// Cloudlet now has a map key'ed by resource name whose value is a resource tag map key
	// We init this map, and create a resource table, and place its key into this map
	// and pass this map to the matcher routine, this allows the matcher to have access
	// to all optional resource tag maps present in the cloudlet.
	if cl.ResTagMap == nil {
		cl.ResTagMap = make(map[string]*edgeproto.ResTagTableKey)
	}

	var gputab = edgeproto.ResTagTable{
		Key: edgeproto.ResTagTableKey{
			Name: "gpumap",
		},
		Tags: []string{"nvidia-63"},
	}
	//
	//testags := []string{"nvidia-63", "nvidia-72", "tag3"}

	//	tbl.Tags = []string{testags[0]}
	_, err := resTagTableApi.CreateResTagTable(ctx, &gputab)
	require.Nil(t, nil, err, "CreateResTagTable")

	cl.ResTagMap["gpu"] = &gputab.Key

	clusterInstApi.GetCloudletResourceMap("gpu", cl.ResTagMap)

	// Test the flavor matcher modifications.
	// We have 2 new extra flavors in test_data.go to
	// mock a couple of FlavorInfo structs representing what some openstack ops have offered.
	// One will have "gpu" in the flavor name itself, another will has vgpu=nvidia-63 as a property.

	// We also  need a list of edgeproto.FlavorInfo structs
	// which it so happens we have in the testutils.CloudletInfoData.Flavors array

	// Now, the Users MEX Flavor contains the key.Name of the context (cloudlet) in which it is to be
	// looked up within. So if tmus-clouldlet-1 is the clouldlet.Key.Name, we'll expect to find
	// a ResTagTable with that name. (If we don't, that's perfectly fine, either no gpus are offered
	// or all such flavaors have "gpu" in the flavor name) The GetVMSpec is happy to be passed an
	// nil ResTagTable.
	//
	// FlavorData[4].ResTabCtx = 'tbl1' so let's add a tag to that table that
	// will match a property of one of our CloudletInfoData[0].Flavors to that table.

	//	_, err = resTagTableApi.AddResTag(ctx, &gputab)
	//	require.Nil(t, err, "AddResTag")
	tbl1, err := resTagTableApi.GetResTagTable(ctx, &gputab.Key)
	require.Nil(t, err, "GetResTagTable")
	require.Equal(t, 1, len(tbl1.Tags), "tag count mismatch")

	/*  our test user defined this MEX meta flavor:
		edgeproto.Flavor{
			Key: edgeproto.FlavorKey{
				Name: "x1.large-mex",
			},
			Ram:   8192,
			Vcpus: 8,
			Disk:  40,
			Ress:  1,
		},

	    our mock OpenStack flavors are defined as:
				// restagtable tests

				&edgeproto.FlavorInfo{
					Name:       "flavor.large",
					Vcpus:      uint64(10),
					Ram:        uint64(8192),
					Disk:       uint64(40),
					Properties: "vgpu=nvidia-63",
				},
				&edgeproto.FlavorInfo{
					Name:  "flavor.large-gpu",
					Vcpus: uint64(8),
					Ram:   uint64(8192),
					Disk:  uint64(40),
				},

	*/
	spec, vmerr := clusterInstApi.GetVMSpec(testutil.CloudletInfoData[0].Flavors, testutil.FlavorData[4], cl.ResTagMap)
	require.Nil(t, vmerr, "GetVmSpec")
	require.Equal(t, "flavor.large-gpu", spec.FlavorName)

	// now to force vmspec.GetVMSpec() to actually look into the given tag table. We
	// ask for more Vcpus which will reject flavor.large-gpu (8 vcpus), but still requesting a GPU
	// resource, so the table will be searched for a matching tag in flavor.large-gpu (10 vcpus) properties.
	testutil.FlavorData[4].Vcpus = 10
	spec, vmerr = clusterInstApi.GetVMSpec(testutil.CloudletInfoData[0].Flavors, testutil.FlavorData[4], cl.ResTagMap)
	require.Nil(t, vmerr, "GetVmSpec")
	require.Equal(t, "flavor.large", spec.FlavorName)

	// and finally, make sure GetVMSpec ignores a nil tbl if none exist or desired, behavior
	// is only a flavor with 'gpu' in the name will trigger a gpu request match.
	spec, vmerr = clusterInstApi.GetVMSpec(testutil.CloudletInfoData[0].Flavors, testutil.FlavorData[4], nil)
	require.Equal(t, "no suitable platform flavor found for x1.large-mex, please try a smaller flavor", vmerr.Error(), "nil table")

}

func testResMapKeysApi(t *testing.T, ctx context.Context, cl *edgeproto.Cloudlet) {
	// We can add/remove edgeproto.ResTagTableKey values to the cl.ResTagMap map
	// which then can be used in the GetVMSpec call when matching our meta-resource specificer
	// to a deployments actual resources/flavrs.

	// test_data contains sample resource tag maps, add them to the cloudlet
	// verify, and remove them. ClI should follow suit.
	if cl.ResTagMap == nil {
		cl.ResTagMap = make(map[string]*edgeproto.ResTagTableKey)
	}
	if cl.Mapping == nil {
		cl.Mapping = make(map[string]string)
	}
	// use the restblkey.Names as clould.ResTagMap[key] = tblkey in test
	// gpu, nas-storage, nic-offload are the tbl names
	// setup the test map using the test_data objects
	// The AddCloudResMapKey is setup to accept multiple res tbl keys at once
	// but we're doing it one by one.

	cl.Mapping[testutil.Restblkeys[0].Name] = testutil.Restblkeys[0].Name
	_, err := cloudletApi.AddCloudletResMapping(ctx, cl)
	require.Nil(t, err, "AddCloudletResMapKey")

	cl.Mapping[testutil.Restblkeys[1].Name] = testutil.Restblkeys[1].Name
	_, err = cloudletApi.AddCloudletResMapping(ctx, cl)
	require.Nil(t, err, "AddCloudletResMapKey")

	cl.Mapping[testutil.Restblkeys[2].Name] = testutil.Restblkeys[2].Name
	_, err = cloudletApi.AddCloudletResMapping(ctx, cl)
	require.Nil(t, err, "AddCloudletResMapKey")

	testcl := &edgeproto.Cloudlet{}
	// now it's all stored, fetch a copy of the cloudlet and verify
	err = cloudletApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !cloudletApi.store.STMGet(stm, &cl.Key, testcl) {
			return objstore.ErrKVStoreKeyNotFound
		}
		return err
	})
	// what's in our testcl? Check the resource map
	tkey := testcl.ResTagMap[testutil.Restblkeys[0].Name]
	require.Equal(t, testutil.Restblkeys[0].Name, tkey.Name, "AddCloudletResMapKey")

	tkey = testcl.ResTagMap[testutil.Restblkeys[1].Name]
	require.Equal(t, testutil.Restblkeys[1].Name, tkey.Name, "AddCloudletResMapKey")

	tkey = testcl.ResTagMap[testutil.Restblkeys[2].Name]
	require.Equal(t, testutil.Restblkeys[2].Name, tkey.Name, "AddCloudletResMapKey")

	// and the actual keys should match as well
	require.Equal(t, testutil.Restblkeys[0], *testcl.ResTagMap[testutil.Restblkeys[0].Name], "AddCloudletResMapKey")
	require.Equal(t, testutil.Restblkeys[1], *testcl.ResTagMap[testutil.Restblkeys[1].Name], "AddCloudletResMapKey")
	require.Equal(t, testutil.Restblkeys[2], *testcl.ResTagMap[testutil.Restblkeys[2].Name], "AddCloudletResMapKey")

	cl1 := edgeproto.Cloudlet{}
	cl1.Mapping = make(map[string]string)
	cl1.Mapping[testutil.Restblkeys[2].Name] = testutil.Restblkeys[2].Name
	cl1.Mapping[testutil.Restblkeys[1].Name] = testutil.Restblkeys[1].Name
	cl1.Key = cl.Key
	_, err = cloudletApi.RmCloudletResMapping(ctx, &cl1)
	require.Nil(t, err, "RmCloudletResMapKey")

	rmcl := &edgeproto.Cloudlet{}
	err = cloudletApi.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !cloudletApi.store.STMGet(stm, &cl.Key, rmcl) {
			return objstore.ErrKVStoreKeyNotFound
		}
		return err
	})
	require.Nil(t, err, "STMGet failure")
	// and check the maps len = 1
	require.Equal(t, 1, len(rmcl.ResTagMap), "RmCloudletResMapKey")
	// and might as well check the key "gpu" exists
	_, ok := rmcl.ResTagMap[testutil.Restblkeys[0].Name]
	require.Equal(t, true, ok, "RmCloudletResMapKey")
}
