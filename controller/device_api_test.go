package main

import (
	"context"
	"time"

	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
)

func TestDeviceApi(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelEtcd | log.DebugLevelApi)
	testinit()
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	dummy := dummyEtcd{}
	dummy.Start()

	sync := InitSync(&dummy)
	InitApis(sync)
	sync.Start()
	defer sync.Done()

	log.SpanLog(ctx, log.DebugLevelApi, "Starting tests")
	// Test Update of the platform device
	for _, obj := range testutil.PlarformDeviceClientData {
		deviceApi.Update(ctx, &obj, 0)
	}
	testutil.InternalDeviceTest(t, "show", &deviceApi, testutil.PlarformDeviceClientData)
	// Add the existing platform device with the new timestamp
	dev := testutil.PlarformDeviceClientData[0]
	dev.FirstSeen = testutil.GetTimestamp(time.Date(2009, time.November, 11, 23, 0, 0, 0, time.UTC))
	deviceApi.Update(ctx, &dev, 0)
	testutil.InternalDeviceTest(t, "show", &deviceApi, testutil.PlarformDeviceClientData)
	// Test Update of a platform device without uniqueID
	dev = testutil.PlarformDeviceClientData[0]
	dev.Key.UniqueId = ""
	deviceApi.Update(ctx, &dev, 0)
	testutil.InternalDeviceTest(t, "show", &deviceApi, testutil.PlarformDeviceClientData)
	// Test that flush doesn't remove the entries
	deviceApi.Flush(ctx, 0)
	testutil.InternalDeviceTest(t, "show", &deviceApi, testutil.PlarformDeviceClientData)
	// Test report to show only a single device in December
	/*
		report := edgeproto.DeviceReport{
			Begin: testutil.GetTimestamp(time.Date(2009, time.December, 1, 23, 0, 0, 0, time.UTC)),
			End:   testutil.GetTimestamp(time.Date(2009, time.December, 31, 23, 0, 0, 0, time.UTC)),
		}
	*/
	// TODO- run report
	//api := testutil.NewInternalDeviceApi(&deviceApi)
	// I want to call testutil.ApiClient.ShowDeviceReport()...

	// Evict all the platform device
	for _, obj := range testutil.PlarformDeviceClientData {
		deviceApi.EvictDevice(ctx, &obj)
	}
	testutil.InternalDeviceTest(t, "show", &deviceApi, []edgeproto.Device{})
	// Test Inject of the platform devices
	for _, obj := range testutil.PlarformDeviceClientData {
		deviceApi.InjectDevice(ctx, &obj)
	}
	testutil.InternalDeviceTest(t, "show", &deviceApi, testutil.PlarformDeviceClientData)

	dummy.Stop()
}
