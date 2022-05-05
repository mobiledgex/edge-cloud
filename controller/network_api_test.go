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

	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/log"
	"github.com/edgexr/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestNetworkApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	testSvcs := testinit(ctx, t)
	defer testfinish(testSvcs)
	dummy := dummyEtcd{}
	dummy.Start()
	defer dummy.Stop()
	sync := InitSync(&dummy)
	apis := NewAllApis(sync)
	sync.Start()
	defer sync.Done()

	testutil.InternalFlavorCreate(t, apis.flavorApi, testutil.FlavorData)
	testutil.InternalGPUDriverCreate(t, apis.gpuDriverApi, testutil.GPUDriverData)
	testutil.InternalResTagTableCreate(t, apis.resTagTableApi, testutil.ResTagTableData)
	testutil.InternalCloudletCreate(t, apis.cloudletApi, testutil.CloudletData())

	testutil.InternalNetworkTest(t, "cud", apis.networkApi, testutil.NetworkData)
	// error cases
	expectCreateNetworkError(t, ctx, apis, &testutil.NetworkErrorData[0], "Invalid route destination cidr")
	expectCreateNetworkError(t, ctx, apis, &testutil.NetworkErrorData[1], "Invalid next hop")
	expectCreateNetworkError(t, ctx, apis, &testutil.NetworkErrorData[2], "Invalid connection type")

}

func expectCreateNetworkError(t *testing.T, ctx context.Context, apis *AllApis, in *edgeproto.Network, msg string) {
	err := apis.networkApi.CreateNetwork(in, testutil.NewCudStreamoutNetwork(ctx))
	require.NotNil(t, err, "create %v", in)
	require.Contains(t, err.Error(), msg, "error %v contains %s", err, msg)
}
