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

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestOperatorCodeApi(t *testing.T) {
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

	testutil.InternalOperatorCodeTest(t, "cud", apis.operatorCodeApi, testutil.OperatorCodeData)
	// create duplicate key should fail
	code := testutil.OperatorCodeData[0]
	_, err := apis.operatorCodeApi.CreateOperatorCode(ctx, &code)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "already exists")

	// check not found error on delete
	_, err = apis.operatorCodeApi.DeleteOperatorCode(ctx, &code)
	require.Nil(t, err)
	_, err = apis.operatorCodeApi.DeleteOperatorCode(ctx, &code)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "not found")
}
