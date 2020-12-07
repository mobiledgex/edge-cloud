package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestCloudletPoolApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testinit()
	cplookup := &node.CloudletPoolCache{}
	cplookup.Init()
	nodeMgr.CloudletPoolLookup = cplookup

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	// create supporting data
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)

	testutil.InternalCloudletPoolTest(t, "cud", &cloudletPoolApi, testutil.CloudletPoolData)

	// create test cloudlet
	cloudlet := edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			Name:         "testcloudlet",
			Organization: testutil.CloudletPoolData[0].Key.Organization,
		},
		NumDynamicIps: 100,
		Location: dme.Loc{
			Latitude:  40.712776,
			Longitude: -74.005974,
		},
		CrmOverride: edgeproto.CRMOverride_IGNORE_CRM,
	}
	err := cloudletApi.CreateCloudlet(&cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)

	// set up test data
	poolKey := testutil.CloudletPoolData[0].Key
	member := edgeproto.CloudletPoolMember{}
	member.Key = poolKey
	member.CloudletName = cloudlet.Key.Name
	pool := edgeproto.CloudletPool{}

	// add member to pool
	_, err = cloudletPoolApi.AddCloudletPoolMember(ctx, &member)
	require.Nil(t, err)
	found := cloudletPoolApi.cache.Get(&poolKey, &pool)
	require.True(t, found, "get pool %v", poolKey)
	require.Equal(t, 2, len(pool.Cloudlets))

	// add duplicate should fail
	_, err = cloudletPoolApi.AddCloudletPoolMember(ctx, &member)
	require.NotNil(t, err)

	// remove member from pool
	_, err = cloudletPoolApi.RemoveCloudletPoolMember(ctx, &member)
	require.Nil(t, err)
	found = cloudletPoolApi.cache.Get(&poolKey, &pool)
	require.True(t, found, "get pool %v", poolKey)
	require.Equal(t, 1, len(pool.Cloudlets))

	// use update to set members for next test
	poolUpdate := testutil.CloudletPoolData[0]
	poolUpdate.Cloudlets = append(poolUpdate.Cloudlets, member.CloudletName)
	poolUpdate.Fields = []string{edgeproto.CloudletPoolFieldCloudlets}
	require.Equal(t, 2, len(poolUpdate.Cloudlets))
	_, err = cloudletPoolApi.UpdateCloudletPool(ctx, &poolUpdate)
	require.Nil(t, err)
	found = cloudletPoolApi.cache.Get(&poolKey, &pool)
	require.True(t, found, "get pool %v", poolKey)
	require.Equal(t, 2, len(pool.Cloudlets))

	// delete cloudlet, see it gets removed from pool
	err = cloudletApi.DeleteCloudlet(&cloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)
	found = cloudletPoolApi.cache.Get(&poolKey, &pool)
	require.True(t, found, "get pool %v", poolKey)
	require.Equal(t, 1, len(pool.Cloudlets))

	// add cloudlet that doesn't exist, should fail
	_, err = cloudletPoolApi.AddCloudletPoolMember(ctx, &member)
	require.NotNil(t, err)
}
