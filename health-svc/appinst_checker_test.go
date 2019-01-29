package main

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestAppInstChecker(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelNotify | log.DebugLevelApi)

	addr := "127.0.0.1:61235"
	saddrs := []string{addr}

	serverH := notify.NewDummyHandler()
	serverMgr := notify.ServerMgr{}
	serverH.RegisterServer(&serverMgr)
	serverMgr.Start(addr, "")

	client := notify.NewClient(saddrs, "")
	aic := &AppInstCheck{}
	aic.init(client)

	client.Start()
	defer client.Stop()
	client.WaitForConnect(1)

	// Use the unpopulated AppInst test data.
	// This data is not the same as what comes from the
	// controller, because the controller will populate
	// all the fields based on the App data.
	// But we don't want the ports fields anyway, so that
	// the AppInst status check function doesn't try to
	// connect to anything.
	appInsts := testutil.AppInstData
	for ii, _ := range appInsts {
		// state needs to be ready
		appInsts[ii].State = edgeproto.TrackedState_Ready
		serverH.AppInstCache.Update(&appInsts[ii], 0)
	}
	serverH.WaitForAppInstStatuses(len(appInsts))
	require.Equal(t, 0, len(aic.tagged))
	require.Equal(t, len(appInsts), len(serverH.AppInstStatusMsgQueue.Objs))
}
