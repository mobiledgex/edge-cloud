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

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestTrustPolicyApi(t *testing.T) {
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

	testutil.InternalTrustPolicyTest(t, "cud", apis.trustPolicyApi, testutil.TrustPolicyData)
	// error cases
	expectCreatePolicyError(t, ctx, apis, &testutil.TrustPolicyErrorData[0], "cannot be higher than max")
	expectCreatePolicyError(t, ctx, apis, &testutil.TrustPolicyErrorData[1], "invalid CIDR")
	expectCreatePolicyError(t, ctx, apis, &testutil.TrustPolicyErrorData[2], "Invalid min port: 0")

	dummy.Stop()
}

func expectCreatePolicyError(t *testing.T, ctx context.Context, apis *AllApis, in *edgeproto.TrustPolicy, msg string) {
	err := apis.trustPolicyApi.CreateTrustPolicy(in, testutil.NewCudStreamoutTrustPolicy(ctx))
	require.NotNil(t, err, "create %v", in)
	require.Contains(t, err.Error(), msg, "error %v contains %s", err, msg)
}
