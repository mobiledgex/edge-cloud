package main

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/vmspec"
	"github.com/stretchr/testify/require"

	"testing"
)

func dumptbl(tbl edgeproto.GpuTagTable) {

	fmt.Printf("dump tbl %s\n", tbl.Key.Name)
	for i, tag := range tbl.Tags {
		fmt.Printf("\ttag[%d] = %s\n", i, tag)
	}
}

func TestGpuTagTableApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	objstore.InitRegion(1)

	tMode := true
	testMode = &tMode

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	testutil.InternalGpuTagTableTest(t, "cud", &gpuTagTableApi, testutil.GpuTagTableData)
	testutil.InternalGpuTagTableTest(t, "show", &gpuTagTableApi, testutil.GpuTagTableData)

	// Non-Nominal attempt to create a table that should already exist from our cud tests
	var tab = edgeproto.GpuTagTable{
		Key: testutil.Gputblkeys[0],
	}
	_, err := gpuTagTableApi.CreateGpuTagTable(ctx, &tab)
	require.Equal(t, "Key already exists", err.Error(), "create tag table EEXIST expected")

	testags := []string{"tag1", "tag2", "tag3"}
	multi_tag := []string{"multi-tag1", "multi-tag2", "multi-tag3"}
	testkeys := []string{"tbl1", "tbl2", "tbl3"}
	nonnom_multi_tag := []string{"mtag1", "mtag2", "mtag2"}

	// create new table
	var tbl edgeproto.GpuTagTable

	//	var key edgeproto.GpuTagTableKey
	tbl.Key.Name = testkeys[0]
	_, err = gpuTagTableApi.CreateGpuTagTable(ctx, &tbl)
	require.Nil(t, err, "Create Gpu Tag Table tmus-clouldlet-1")

	// turn around and fetch the empty table we just created
	var tbl1 *edgeproto.GpuTagTable
	tbl1, err = gpuTagTableApi.GetGpuTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "Get Gpu Table")
	require.Equal(t, len(tbl1.Tags), 0) // no tags yet

	// add some tags to tbl
	tbl.Tags = []string{testags[0]}

	_, err = gpuTagTableApi.AddGpuTag(ctx, &tbl)
	require.Nil(t, err, "AddGpuTag")

	tbl1, err = gpuTagTableApi.GetGpuTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "GetGpuTagTable")
	require.Equal(t, len(tbl1.Tags), 1, "Num Tags error")

	// another tag
	tbl.Tags[0] = testags[1] // "tag2"
	_, err = gpuTagTableApi.AddGpuTag(ctx, &tbl)
	require.Nil(t, err, "AddGpuTag")

	tbl1, err = gpuTagTableApi.GetGpuTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "GgetGpuTagTable")
	require.Equal(t, len(tbl1.Tags), 2, "Num Tags error")

	tbl.Tags[0] = testags[2] // "tag3"

	_, err = gpuTagTableApi.AddGpuTag(ctx, &tbl)
	require.Nil(t, err, "AddGpuTag")

	tbl1, err = gpuTagTableApi.GetGpuTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "GgetGpuTagTable")
	require.Equal(t, len(tbl1.Tags), 3, "Num Tags error")

	// Non-nominal add duplicate tag
	_, err = gpuTagTableApi.AddGpuTag(ctx, &tbl)
	require.Equal(t, "Duplicate Tag Found tag3", err.Error(), "AddGpuTag dup tag")

	// Nominal Delete tag
	_, err = gpuTagTableApi.RemoveGpuTag(ctx, &tbl)
	require.Nil(t, err, "RemoveGpuTag")

	tbl1, err = gpuTagTableApi.GetGpuTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "GgetGpuTagTable")
	require.Equal(t, len(tbl1.Tags), 2, "Num Tags error")

	// and what's left should be tag1 and tag2
	require.Equal(t, tbl1.Tags[0], "tag1", "reamining tags")
	require.Equal(t, tbl1.Tags[1], "tag2", "remaining tags")
	require.Equal(t, 2, len(tbl1.Tags), "len remaining tags unexpected")

	// Test the flavor matcher modifications.
	// We have 2 new extra flavors in test_data.go to
	// mock a couple of FlavorInfo structs representing what some openstack ops have offered.
	// One will have "gpu" in the flavor name itself, another will has vgpu=nvidia-63 as a property.

	// We also  need a list of edgeproto.FlavorInfo structs
	// which it so happens we have in the testutils.CloudletInfoData.Flavors array
	require.Equal(t, len(testutil.CloudletInfoData[0].Flavors), 6, "num test_data.cloudletInfoData.Flavors")
	require.Equal(t, len(testutil.FlavorData), 5, "num test_data.FlavorData")

	// Now, the Users MEX Flavor contains the key.Name of the context (cloudlet) in which it is to be
	// looked up within. So if tmus-clouldlet-1 is the clouldlet.Key.Name, we'll expect to find
	// a GpuTagTable with that name. (If we don't, that's perfectly fine, either no gpus are offered
	// or all such flavaors have "gpu" in the flavor name) The GetVMSpec is happy to be passed an
	// empty GpuTagTable.
	//
	// FlavorData[4].GpuTabCtx = 'tbl1' so let's add a tag to that table that
	// will match a property of one of our CloudletInfoData[0].Flavors to that table.

	tbl.Tags[0] = "nvidia-63" // vcpu=nvidia-63 could be any property value
	_, err = gpuTagTableApi.AddGpuTag(ctx, &tbl)
	require.Nil(t, err, "AddGpuTag")

	tbl1, err = gpuTagTableApi.GetGpuTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "GgetGpuTagTable")
	require.Equal(t, 3, len(tbl1.Tags), "tag count mismatch")
	/*  our test user defined this MEX meta flavor:
		edgeproto.Flavor{
			Key: edgeproto.FlavorKey{
				Name: "x1.large-mex",
			},
			Ram:   8192,
			Vcpus: 8,
			Disk:  40,
			Gpus:  1,
		},

	    our mock OpenStack flavors are defined as:
				// gputagtable tests

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
	spec, vmerr := vmspec.GetVMSpec(testutil.CloudletInfoData[0].Flavors, testutil.FlavorData[4], *tbl1)
	require.Nil(t, vmerr, "GetVmSpec")
	require.Equal(t, "flavor.large-gpu", spec.FlavorName)

	// now to force vmspec.GetVMSpec() to actually look into the given tag table. We
	// ask for more Vcpus which will reject flavor.large-gpu (8 vcpus), but still requesting a GPU
	// resource, so the table will be searched for a matching tag in flavor.large-gpu (10 vcpus) properties.
	testutil.FlavorData[4].Vcpus = 10
	spec, vmerr = vmspec.GetVMSpec(testutil.CloudletInfoData[0].Flavors, testutil.FlavorData[4], *tbl1)
	require.Nil(t, vmerr, "GetVmSpec")
	require.Equal(t, "flavor.large", spec.FlavorName)

	// test multi-tag input support
	tbl.Tags = multi_tag
	_, err = gpuTagTableApi.AddGpuTag(ctx, &tbl)
	require.Nil(t, err, "Multi-Tag add")

	tbl1, err = gpuTagTableApi.GetGpuTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "GgetGpuTagTable")
	require.Equal(t, 6, len(tbl1.Tags), "multi-tag num tags err")
	dumptbl(*tbl1)

	// Multi-Tag remove option, our tbl is set with mulit_tag so use that
	_, err = gpuTagTableApi.RemoveGpuTag(ctx, &tbl)
	require.Nil(t, err, "RemoveGpuTag")

	// Get the table, we should have 3 tags left:  {tag1, tag2 and nvidia-63}
	tbl1, err = gpuTagTableApi.GetGpuTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "GgetGpuTagTable")
	require.Equal(t, 3, len(tbl1.Tags), "multi-tag remove tags err")
	dumptbl(*tbl1)

	// Non-nominal multi-tag input test, duplicate tag in input tag set
	tbl.Tags = nonnom_multi_tag
	_, err = gpuTagTableApi.AddGpuTag(ctx, &tbl)
	require.Equal(t, "Duplicate Tag Found mtag2 in multi-tag input", err.Error(), "Dup Multi-Tag add")

	// Final state of our test tbl1 should match that listed above in comment
	require.Equal(t, "tag1", tbl1.Tags[0], "TagTab membership mismatch")
	require.Equal(t, "tag2", tbl1.Tags[1], "TagTab membership mismatch")
	require.Equal(t, "nvidia-63", tbl1.Tags[2], "TagTab membership mismatch")
	require.Equal(t, 3, len(tbl1.Tags), "TagTab len unexpected")

	// and finally, make sure GetVMSpec ignores an empty tbl if none exist or desired, behavior
	// is only a flavor with 'gpu' in the name will trigger a gpu request match.
	tbl1.Key.Name = ""
	spec, vmerr = vmspec.GetVMSpec(testutil.CloudletInfoData[0].Flavors, testutil.FlavorData[4], *tbl1)
	require.Equal(t, "no suitable platform flavor found for x1.large-mex, please try a smaller flavor", vmerr.Error(), "nil table")

}
