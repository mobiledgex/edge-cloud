package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"

	"testing"
)

func TestResTagTableApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
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

	testutil.InternalResTagTableTest(t, "cud", &resTagTableApi, testutil.ResTagTableData)
	testutil.InternalResTagTableTest(t, "show", &resTagTableApi, testutil.ResTagTableData)

	// Non-Nominal attempt to create a table that should already exist from our cud tests
	var tab = edgeproto.ResTagTable{
		Key: testutil.Restblkeys[0],
	}
	_, err := resTagTableApi.CreateResTagTable(ctx, &tab)
	require.Equal(t, "ResTagTable key {\"name\":\"gpu\",\"organization\":\"AT\\u0026T Inc.\"} already exists", err.Error(), "create tag table EEXIST expected")

	testags := map[string]string{"tag1": "val1"} //  "tag2": "val2", "tag3": "val3"}
	multi_tag := map[string]string{"multi-tag1": "val1", "multi-tag2": "val2", "multi-tag3": "val2"}
	// create new table
	//var in edgeproto.ResTagTable
	var tbl edgeproto.ResTagTable
	var tkey edgeproto.ResTagTableKey
	tkey.Name = "gpu"
	tkey.Organization = "testOp"
	tbl.Key = tkey

	_, err = resTagTableApi.CreateResTagTable(ctx, &tbl)
	require.Nil(t, err, "Create Res Tag Table dmuus-clouldlet-1")

	// turn around and fetch the empty table we just created
	var tbl1 *edgeproto.ResTagTable
	tbl1, err = resTagTableApi.GetResTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "Get Res Table")
	require.Equal(t, len(tbl1.Tags), 0) // no tags yet

	// add some tags to tbl
	tbl.Tags = testags

	_, err = resTagTableApi.AddResTag(ctx, &tbl)
	require.Nil(t, err, "AddResTag")

	tbl1, err = resTagTableApi.GetResTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "GetResTagTable")
	require.Equal(t, 1, len(tbl1.Tags), "Num Tags error")

	// another tag
	tbl.Tags["tag2"] = "val2" // testags[1] // "tag2"
	_, err = resTagTableApi.AddResTag(ctx, &tbl)
	require.Nil(t, err, "AddResTag")

	tbl1, err = resTagTableApi.GetResTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "GgetResTagTable")
	require.Equal(t, 2, len(tbl1.Tags), "Num Tags error")

	tbl.Tags["tag3"] = "val3"

	_, err = resTagTableApi.AddResTag(ctx, &tbl)
	require.Nil(t, err, "AddResTag")

	tbl1, err = resTagTableApi.GetResTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "GgetResTagTable")
	require.Equal(t, 3, len(tbl1.Tags), "Num Tags error")

	// Nominal Delete tag
	delete(tbl.Tags, "tag1")
	delete(tbl.Tags, "tag2")
	// all that's left is tag3 which we want deleted
	_, err = resTagTableApi.RemoveResTag(ctx, &tbl)
	require.Nil(t, err, "RemoveResTag")

	tbl1, err = resTagTableApi.GetResTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "GetResTagTable")
	require.Equal(t, 2, len(tbl1.Tags), "Num Tags error")

	// and what's left should be tag1 and tag2
	require.Equal(t, tbl1.Tags["tag1"], "val1", "reamining tags")
	require.Equal(t, tbl1.Tags["tag2"], "val2", "remaining tags")
	require.Equal(t, 2, len(tbl1.Tags), "len remaining tags unexpected")

	// test multi-tag input support
	tbl.Tags = multi_tag
	_, err = resTagTableApi.AddResTag(ctx, &tbl)
	require.Nil(t, err, "Multi-Tag add")

	tbl1, err = resTagTableApi.GetResTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "GgetResTagTable")
	require.Equal(t, 5, len(tbl1.Tags), "multi-tag num tags err")

	// Multi-Tag remove option, our tbl is set with mulit_tag so use that
	_, err = resTagTableApi.RemoveResTag(ctx, &tbl)
	require.Nil(t, err, "RemoveResTag")

	// Get the table, we should have 2 tags left:  {tag1, tag2}
	tbl1, err = resTagTableApi.GetResTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "GgetResTagTable")
	require.Equal(t, 2, len(tbl1.Tags), "multi-tag remove tags err")

	// Final state of our test tbl1 should match that listed above in comment
	require.Equal(t, "val1", tbl1.Tags["tag1"], "TagTab membership mismatch")
	require.Equal(t, "val2", tbl1.Tags["tag2"], "TagTab membership mismatch")
	require.Equal(t, 2, len(tbl1.Tags), "TagTab len unexpected")

	// test update of optional availablity zone
	update := edgeproto.ResTagTable{}
	update.Key = tbl.Key
	update.Fields = make([]string, 0)
	update.Azone = "gpu_zone"
	update.Fields = append(update.Fields, edgeproto.ResTagTableFieldAzone)

	_, err = resTagTableApi.UpdateResTagTable(ctx, &update)
	require.Nil(t, err, "UpdateResTagTable")

	tbl1, err = resTagTableApi.GetResTagTable(ctx, &tbl.Key)
	require.Nil(t, err, "GgetResTagTable")
	require.Equal(t, "gpu_zone", tbl1.Azone, "UpdateResTagTable")

}
