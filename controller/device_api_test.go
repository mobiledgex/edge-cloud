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
	"time"

	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestDeviceApi(t *testing.T) {
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

	log.SpanLog(ctx, log.DebugLevelApi, "Starting tests")
	// Test Update of the platform device
	for _, obj := range testutil.PlarformDeviceClientData {
		apis.deviceApi.Update(ctx, &obj, 0)
	}
	testutil.InternalDeviceTest(t, "show", apis.deviceApi, testutil.PlarformDeviceClientData)
	// Add the existing platform device with the new timestamp
	dev := testutil.PlarformDeviceClientData[0]
	dev.FirstSeen = testutil.GetTimestamp(time.Date(2009, time.November, 11, 23, 0, 0, 0, time.UTC))
	apis.deviceApi.Update(ctx, &dev, 0)
	testutil.InternalDeviceTest(t, "show", apis.deviceApi, testutil.PlarformDeviceClientData)
	// Test Update of a platform device without uniqueID
	dev = testutil.PlarformDeviceClientData[0]
	dev.Key.UniqueId = ""
	apis.deviceApi.Update(ctx, &dev, 0)
	testutil.InternalDeviceTest(t, "show", apis.deviceApi, testutil.PlarformDeviceClientData)
	// Test that flush doesn't remove the entries
	apis.deviceApi.Flush(ctx, 0)
	testutil.InternalDeviceTest(t, "show", apis.deviceApi, testutil.PlarformDeviceClientData)
	// Test report to show only a single device in December
	report := edgeproto.DeviceReport{
		Begin: testutil.GetTimestamp(time.Date(2009, time.December, 1, 23, 0, 0, 0, time.UTC)),
		End:   testutil.GetTimestamp(time.Date(2009, time.December, 31, 23, 0, 0, 0, time.UTC)),
	}
	show := testutil.ShowDevice{}
	show.Init()
	err := apis.deviceApi.ShowDeviceReport(&report, &show)
	require.Nil(t, err)
	require.Equal(t, 2, len(show.Data))
	// Verify that the two devices got were correct
	for _, dev := range show.Data {
		if dev.Key.UniqueIdType == testutil.PlarformDeviceClientData[2].Key.UniqueIdType {
			require.Equal(t, testutil.PlarformDeviceClientData[2], dev)
		} else {
			require.Equal(t, testutil.PlarformDeviceClientData[4], dev)
		}
	}

	// I want to call testutil.ApiClient.ShowDeviceReport()...

	// Evict all the platform device
	for _, obj := range testutil.PlarformDeviceClientData {
		apis.deviceApi.EvictDevice(ctx, &obj)
	}
	testutil.InternalDeviceTest(t, "show", apis.deviceApi, []edgeproto.Device{})
	// Test Inject of the platform devices
	for _, obj := range testutil.PlarformDeviceClientData {
		apis.deviceApi.InjectDevice(ctx, &obj)
	}
	testutil.InternalDeviceTest(t, "show", apis.deviceApi, testutil.PlarformDeviceClientData)

	dummy.Stop()
}
