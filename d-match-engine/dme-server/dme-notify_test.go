package main

import (
	"context"
	"testing"
	"time"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

var (
	appInstUsable    bool = true
	appInstNotUsable bool = false
)

func TestNotify(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelNotify)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	eehandler, err := initEdgeEventsPlugin(ctx, "standalone")
	require.Nil(t, err, "init edge events plugin")
	dmecommon.SetupMatchEngine(eehandler)
	initRateLimitMgr()
	dmecommon.InitAppInstClients()
	defer dmecommon.StopAppInstClients()
	apps := dmetest.GenerateApps()
	appInsts := dmetest.GenerateAppInsts()

	// test dummy server sending notices to dme
	addr := "127.0.0.1:60001"

	// dummy server side
	serverHandler := notify.NewDummyHandler()
	serverMgr := notify.ServerMgr{}
	serverHandler.RegisterServer(&serverMgr)
	serverMgr.Start("ctrl", addr, nil)

	// client (dme) side
	client := initNotifyClient(ctx, addr, grpc.WithInsecure())
	client.Start()

	// create data on server side
	for _, app := range apps {
		serverHandler.AppCache.Update(ctx, app, 0)
	}
	for _, appInst := range appInsts {
		serverHandler.AppInstCache.Update(ctx, appInst, 0)
	}
	// wait for the last appInst data to show up locally
	last := len(appInsts) - 1
	waitForAppInst(appInsts[last])
	// check that all data is present
	checkAllData(t, appInsts)

	// remove one appinst
	remaining := appInsts[:last]
	serverHandler.AppInstCache.Delete(ctx, appInsts[last], 0)
	// wait for it to be gone locally
	waitForNoAppInst(appInsts[last])
	// check new data
	checkAllData(t, remaining)
	// add it back
	serverHandler.AppInstCache.Update(ctx, appInsts[last], 0)
	// wait for it to be present again
	waitForAppInst(appInsts[last])
	checkAllData(t, appInsts)

	// update cloudletInfo for a single cloudlet and make sure it gets propagated to appInsts
	cloudletInfo := edgeproto.CloudletInfo{
		Key: edgeproto.CloudletKey{
			Organization: dmetest.Cloudlets[2].CarrierName,
			Name:         dmetest.Cloudlets[2].Name,
		},
		State: dme.CloudletState_CLOUDLET_STATE_OFFLINE,
	}
	cloudlet := edgeproto.Cloudlet{
		Key: cloudletInfo.Key,
	}
	serverHandler.CloudletCache.Update(ctx, &cloudlet, 0)
	serverHandler.CloudletInfoCache.Update(ctx, &cloudletInfo, 0)
	// check that the appInsts on that cloudlet are not available
	waitAndCheckCloudletforApps(t, &cloudletInfo.Key, appInstNotUsable)

	// update cloudletInfo for a single cloudlet and make sure it gets propagated to appInsts
	cloudletInfo.State = dme.CloudletState_CLOUDLET_STATE_READY
	serverHandler.CloudletInfoCache.Update(ctx, &cloudletInfo, 0)
	// check that the appInsts on that cloudlet are available
	waitAndCheckCloudletforApps(t, &cloudletInfo.Key, appInstUsable)

	// mark cloudlet under maintenance state just for cloudlet object
	cloudlet.MaintenanceState = dme.MaintenanceState_UNDER_MAINTENANCE
	serverHandler.CloudletCache.Update(ctx, &cloudlet, 0)
	waitAndCheckCloudletforApps(t, &cloudletInfo.Key, appInstNotUsable)

	// mark cloudlet operational just for cloudlet object
	cloudlet.MaintenanceState = dme.MaintenanceState_NORMAL_OPERATION
	serverHandler.CloudletCache.Update(ctx, &cloudlet, 0)
	waitAndCheckCloudletforApps(t, &cloudletInfo.Key, appInstUsable)

	// set cloudletInfo maintenance state in maintenance mode,
	// should not affect appInst
	cloudletInfo.MaintenanceState = dme.MaintenanceState_CRM_UNDER_MAINTENANCE
	serverHandler.CloudletInfoCache.Update(ctx, &cloudletInfo, 0)
	waitAndCheckCloudletforApps(t, &cloudletInfo.Key, appInstUsable)

	// delete cloudlet object appInst should not be usable, even though
	serverHandler.CloudletCache.Delete(ctx, &cloudlet, 0)
	waitAndCheckCloudletforApps(t, &cloudletInfo.Key, appInstNotUsable)

	// stop client, delete appInst on server, then start client.
	// This checks that client deletes locally data
	// that was deleted while the connection was down.
	client.Stop()
	serverHandler.AppInstCache.Delete(ctx, appInsts[last], 0)
	client.Start()
	waitForNoAppInst(appInsts[last])
	checkAllData(t, remaining)

	// add a new device - see that it makes it to the server
	for _, reg := range dmetest.DeviceData {
		dmecommon.RecordDevice(ctx, &reg)
	}
	// verify the devices were added to the server
	count := len(dmetest.DeviceData) - 1 // Since one is a duplicate
	// verify that devices are in local cache
	assert.Equal(t, count, len(dmecommon.PlatformClientsCache.Objs))
	serverHandler.WaitForDevices(count)
	assert.Equal(t, count, len(serverHandler.DeviceCache.Objs))
	// Delete all elements from local cache directly
	for _, data := range dmecommon.PlatformClientsCache.Objs {
		obj := data.Obj
		delete(dmecommon.PlatformClientsCache.Objs, obj.GetKeyVal())
		delete(dmecommon.PlatformClientsCache.List, obj.GetKeyVal())
	}
	assert.Equal(t, 0, len(dmecommon.PlatformClientsCache.Objs))
	assert.Equal(t, count, len(serverHandler.DeviceCache.Objs))
	// Add a single device - make sure count in local cache is updated
	dmecommon.RecordDevice(ctx, &dmetest.DeviceData[0])
	assert.Equal(t, 1, len(dmecommon.PlatformClientsCache.Objs))
	// Make sure that count in the server cache is the same
	assert.Equal(t, count, len(serverHandler.DeviceCache.Objs))
	// Add the same device, check that nothing is updated
	dmecommon.RecordDevice(ctx, &dmetest.DeviceData[0])
	assert.Equal(t, 1, len(dmecommon.PlatformClientsCache.Objs))
	assert.Equal(t, count, len(serverHandler.DeviceCache.Objs))

	serverMgr.Stop()
	client.Stop()
}

func waitAndCheckCloudletforApps(t *testing.T, key *edgeproto.CloudletKey, isAppInstUsable bool) {
	var still_enabled bool

	tbl := dmecommon.DmeAppTbl
	carrier := key.Organization
	for i := 0; i < 10; i++ {
		still_enabled = false
		for _, app := range tbl.Apps {
			if c, found := app.Carriers[carrier]; found {
				for clusterInstKey, appInst := range c.Insts {
					if clusterInstKey.CloudletKey.GetKeyString() == key.GetKeyString() &&
						dmecommon.IsAppInstUsable(appInst) {
						still_enabled = true
					}
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	if isAppInstUsable {
		assert.True(t, still_enabled, "Notify message should have propagated")
	} else {
		assert.False(t, still_enabled, "Notify message did not propagate")
	}
}

func waitForAppInst(appInst *edgeproto.AppInst) {
	tbl := dmecommon.DmeAppTbl

	appkey := appInst.Key.AppKey
	for i := 0; i < 10; i++ {
		if app, found := tbl.Apps[appkey]; found {
			for _, c := range app.Carriers {
				if _, found := c.Insts[appInst.Key.ClusterInstKey]; found {
					break
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func waitForNoAppInst(appInst *edgeproto.AppInst) {
	tbl := dmecommon.DmeAppTbl

	appkey := appInst.Key.AppKey
	for i := 0; i < 10; i++ {
		app, found := tbl.Apps[appkey]
		if !found {
			break
		}
		for _, c := range app.Carriers {
			if _, found := c.Insts[appInst.Key.ClusterInstKey]; !found {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}
