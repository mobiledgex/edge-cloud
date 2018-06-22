package main

import (
	"testing"
	"time"

	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/util"
)

func TestNotify(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelNotify)
	setupMatchEngine()
	appInsts := dmetest.GenerateAppInsts()

	// test dummy server sending notices to dme
	addr := "127.0.0.1:60001"

	// dummy server side
	serverHandler := notify.NewDummyServerHandler()
	notify.ServerMgrStart(addr, serverHandler)

	// client (dme) side
	clientHandler := &NotifyHandler{}
	client := initNotifyClient(addr, clientHandler)
	go client.Run()

	// create data on server side
	for _, appInst := range appInsts {
		serverHandler.CreateAppInst(appInst)
	}
	// wait for the last appInst data to show up locally
	last := len(appInsts) - 1
	waitForAppInst(appInsts[last])
	// check that all data is present
	checkAllData(t, appInsts)

	// remove one appinst
	remaining := appInsts[:last]
	serverHandler.DeleteAppInst(appInsts[last])
	// wait for it to be gone locally
	waitForNoAppInst(appInsts[last])
	// check new data
	checkAllData(t, remaining)
	// add it back
	serverHandler.CreateAppInst(appInsts[last])
	// wait for it to be present again
	waitForAppInst(appInsts[last])
	checkAllData(t, appInsts)

	// stop client, delete appInst on server, then start client.
	// This checks that client deletes locally data
	// that was deleted while the connection was down.
	client.Stop()
	serverHandler.DeleteAppInst(appInsts[last])
	go client.Run()
	waitForNoAppInst(appInsts[last])
	checkAllData(t, remaining)

	notify.ServerMgrDone()
	client.Stop()
}

func waitForAppInst(appInst *edgeproto.AppInst) {
	tbl := carrierAppTbl

	key := carrierAppKey{}
	setCarrierAppKey(appInst, &key)

	for i := 0; i < 10; i++ {
		if app, found := tbl.apps[key]; found {
			if _, found := app.insts[appInst.Key.CloudletKey]; found {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func waitForNoAppInst(appInst *edgeproto.AppInst) {
	tbl := carrierAppTbl

	key := carrierAppKey{}
	setCarrierAppKey(appInst, &key)

	for i := 0; i < 10; i++ {
		app, found := tbl.apps[key]
		if !found {
			break
		}
		if _, found := app.insts[appInst.Key.CloudletKey]; !found {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}
