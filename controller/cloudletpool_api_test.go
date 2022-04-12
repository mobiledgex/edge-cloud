// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"testing"

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
	testSvcs := testinit(ctx, t)
	defer testfinish(testSvcs)

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	apis := NewAllApis(sync)
	sync.Start()
	defer sync.Done()

	// create supporting data
	testutil.InternalFlavorCreate(t, apis.flavorApi, testutil.FlavorData)
	testutil.InternalGPUDriverCreate(t, apis.gpuDriverApi, testutil.GPUDriverData)
	testutil.InternalResTagTableCreate(t, apis.resTagTableApi, testutil.ResTagTableData)
	testutil.InternalCloudletCreate(t, apis.cloudletApi, testutil.CloudletData())

	testutil.InternalCloudletPoolTest(t, "cud", apis.cloudletPoolApi, testutil.CloudletPoolData)

	// create test cloudlet
	testcloudlet := edgeproto.Cloudlet{
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
	err := apis.cloudletApi.CreateCloudlet(&testcloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)
	fedcloudlet := edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			Name:                  "testfedcloudlet",
			Organization:          testutil.CloudletPoolData[0].Key.Organization,
			FederatedOrganization: "FedOrg",
		},
		NumDynamicIps: 100,
		Location: dme.Loc{
			Latitude:  40.712776,
			Longitude: -74.005974,
		},
		CrmOverride: edgeproto.CRMOverride_IGNORE_CRM,
	}
	err = apis.cloudletApi.CreateCloudlet(&fedcloudlet, testutil.NewCudStreamoutCloudlet(ctx))
	require.Nil(t, err)

	testcloudlets := []edgeproto.Cloudlet{testcloudlet, fedcloudlet}

	count := 1
	for _, cloudlet := range testcloudlets {
		// set up test data
		poolKey := testutil.CloudletPoolData[0].Key
		member := edgeproto.CloudletPoolMember{}
		member.Key = poolKey
		member.Cloudlet = cloudlet.Key
		pool := edgeproto.CloudletPool{}

		// add member to pool
		_, err = apis.cloudletPoolApi.AddCloudletPoolMember(ctx, &member)
		require.Nil(t, err)
		count++
		found := apis.cloudletPoolApi.cache.Get(&poolKey, &pool)
		require.True(t, found, "get pool %v", poolKey)
		require.Equal(t, count, len(pool.Cloudlets))

		// add duplicate should fail
		_, err = apis.cloudletPoolApi.AddCloudletPoolMember(ctx, &member)
		require.NotNil(t, err)

		// remove member from pool
		_, err = apis.cloudletPoolApi.RemoveCloudletPoolMember(ctx, &member)
		require.Nil(t, err)
		count--
		found = apis.cloudletPoolApi.cache.Get(&poolKey, &pool)
		require.True(t, found, "get pool %v", poolKey)
		require.Equal(t, count, len(pool.Cloudlets))

		// use update to set members for next test
		poolUpdate := pool
		poolUpdate.Cloudlets = append(poolUpdate.Cloudlets, member.Cloudlet)
		poolUpdate.Fields = []string{edgeproto.CloudletPoolFieldCloudlets}
		_, err = apis.cloudletPoolApi.UpdateCloudletPool(ctx, &poolUpdate)
		require.Nil(t, err)
		count++
		found = apis.cloudletPoolApi.cache.Get(&poolKey, &pool)
		require.True(t, found, "get pool %v", poolKey)
		require.Equal(t, count, len(pool.Cloudlets))

		// add cloudlet that doesn't exist, should fail
		_, err = apis.cloudletPoolApi.AddCloudletPoolMember(ctx, &member)
		require.NotNil(t, err)
	}
}
