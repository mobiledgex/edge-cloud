package main

import (
	"context"
	"testing"
	"time"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
)

func TestNotify(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelNotify)
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	dmecommon.SetupMatchEngine()
	apps := dmetest.GenerateApps()
	appInsts := dmetest.GenerateAppInsts()

	// test dummy server sending notices to dme
	addr := "127.0.0.1:60001"

	// dummy server side
	serverHandler := notify.NewDummyHandler()
	serverMgr := notify.ServerMgr{}
	serverHandler.RegisterServer(&serverMgr)
	serverMgr.Start(addr, "")

	// client (dme) side
	client := initNotifyClient(addr, "")
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

	// stop client, delete appInst on server, then start client.
	// This checks that client deletes locally data
	// that was deleted while the connection was down.
	client.Stop()
	serverHandler.AppInstCache.Delete(ctx, appInsts[last], 0)
	client.Start()
	waitForNoAppInst(appInsts[last])
	checkAllData(t, remaining)

	serverMgr.Stop()
	client.Stop()
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
