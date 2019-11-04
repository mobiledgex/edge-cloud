package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestCloudletPoolApi(t *testing.T) {
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

	// create supporting data
	testutil.InternalOperatorCreate(t, &operatorApi, testutil.OperatorData)
	testutil.InternalFlavorCreate(t, &flavorApi, testutil.FlavorData)
	testutil.InternalCloudletCreate(t, &cloudletApi, testutil.CloudletData)

	testutil.InternalCloudletPoolTest(t, "cud", &cloudletPoolApi, testutil.CloudletPoolData)

	testutil.InternalCloudletPoolMemberTest(t, "cud", &cloudletPoolMemberApi, testutil.CloudletPoolMemberData)

	// test pools for cloudlets api
	for _, cloudlet := range testutil.CloudletData {
		expected := expectedPoolsForCloudlet(t, &cloudlet.Key)
		show := testutil.ShowCloudletPool{}
		show.Init()
		show.Ctx = ctx
		err := cloudletPoolMemberApi.ShowPoolsForCloudlet(&cloudlet.Key, &show)
		require.Nil(t, err, "show pools for cloudlet key %v", cloudlet.Key)
		require.Equal(t, len(expected), len(show.Data), "num pools for cloudlet key %v", cloudlet.Key)
		for _, pool := range expected {
			show.AssertFound(t, pool)
		}
	}

	// test ShowCloudletsForPools api
	for _, pool := range testutil.CloudletPoolData {
		expected := expectedCloudletsForPools(t, pool.Key)
		show := testutil.ShowCloudlet{}
		show.Init()
		show.Ctx = ctx
		err := cloudletPoolMemberApi.ShowCloudletsForPool(&pool.Key, &show)
		require.Nil(t, err, "show cloudlets for pool %v", pool.Key)
		require.Equal(t, len(expected), len(show.Data), "num cloudlets for pool key %v", pool.Key)
		for _, cloudlet := range expected {
			show.AssertFound(t, cloudlet)
		}
	}

	// delete cloudlet, check that it cleans up members
	{
		in := testutil.CloudletData[3]
		// first check that there's something to clean up
		count := countMembersForCloudlet(t, ctx, &in.Key)
		require.True(t, count > 0, "members exist to clean up")

		out := testutil.NewCudStreamoutCloudlet(ctx)
		err := cloudletApi.DeleteCloudlet(&in, out)
		require.Nil(t, err, "delete cloudlet")
		count = countMembersForCloudlet(t, ctx, &in.Key)
		require.Equal(t, 0, count, "members deleted")
	}
	// delete pool, check that it cleans up members
	{
		in := testutil.CloudletPoolData[1]
		// first check that there's something to clean up
		count := countMembersForPool(t, ctx, &in.Key)
		require.True(t, count > 0, "members exist to clean up")

		_, err := cloudletPoolApi.DeleteCloudletPool(ctx, &in)
		require.Nil(t, err, "delete cloudlet pool")
		count = countMembersForPool(t, ctx, &in.Key)
		require.Equal(t, 0, count, "members deleted")
	}
}

func expectedPoolsForCloudlet(t *testing.T, cloudletKey *edgeproto.CloudletKey) map[edgeproto.CloudletPoolKey]*edgeproto.CloudletPool {
	pools := make(map[edgeproto.CloudletPoolKey]*edgeproto.CloudletPool)
	for _, member := range testutil.CloudletPoolMemberData {
		if !cloudletKey.Matches(&member.CloudletKey) {
			continue
		}
		pool, found := testutil.FindCloudletPoolData(&member.PoolKey, testutil.CloudletPoolData)
		require.True(t, found, "find cloudlet pool %v", &member.PoolKey)
		pools[pool.Key] = pool
	}
	return pools
}

func expectedCloudletsForPools(t *testing.T, poolKeys ...edgeproto.CloudletPoolKey) map[edgeproto.CloudletKey]*edgeproto.Cloudlet {
	keysMap := make(map[edgeproto.CloudletPoolKey]struct{})
	for _, poolKey := range poolKeys {
		keysMap[poolKey] = struct{}{}
	}
	cloudlets := make(map[edgeproto.CloudletKey]*edgeproto.Cloudlet)
	for _, member := range testutil.CloudletPoolMemberData {
		if _, found := keysMap[member.PoolKey]; !found {
			continue
		}
		cloudlet, found := testutil.FindCloudletData(&member.CloudletKey, testutil.CloudletData)
		require.True(t, found, "find cloudlet %v", &member.CloudletKey)
		cloudlets[cloudlet.Key] = cloudlet
	}
	return cloudlets
}

func countMembersForCloudlet(t *testing.T, ctx context.Context, key *edgeproto.CloudletKey) int {
	show := testutil.ShowCloudletPoolMember{}
	show.Init()
	show.Ctx = ctx
	filter := edgeproto.CloudletPoolMember{
		CloudletKey: *key,
	}
	err := cloudletPoolMemberApi.ShowCloudletPoolMember(&filter, &show)
	require.Nil(t, err, "show cloudlet pool member")
	return len(show.Data)
}

func countMembersForPool(t *testing.T, ctx context.Context, key *edgeproto.CloudletPoolKey) int {
	show := testutil.ShowCloudletPoolMember{}
	show.Init()
	show.Ctx = ctx
	filter := edgeproto.CloudletPoolMember{
		PoolKey: *key,
	}
	err := cloudletPoolMemberApi.ShowCloudletPoolMember(&filter, &show)
	require.Nil(t, err, "show cloudlet pool member")
	return len(show.Data)
}
