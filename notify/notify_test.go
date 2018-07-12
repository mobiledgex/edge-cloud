package notify

import (
	"fmt"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNotify(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelNotify)

	// override retry time
	NotifyRetryTime = 10 * time.Millisecond

	// This tests the server sending notices to
	// a client.
	addr := "127.0.0.1:61234"
	serverAddrs := []string{addr}

	// Set up server
	serverHandler := NewDummyHandler()
	serverMgr := ServerMgr{}
	serverHandler.SetServerCb(&serverMgr)
	serverMgr.Start(addr, serverHandler)

	// Set up client DME
	dmeHandler := NewDummyHandler()
	clientDME := NewDMEClient(serverAddrs, dmeHandler)
	clientDME.Start()

	// Set up client CRM
	crmHandler := NewDummyHandler()
	clientCRM := NewCRMClient(serverAddrs, crmHandler)
	crmHandler.SetClientCb(clientCRM)
	clientCRM.Start()

	// It takes a little while for the Run thread to start up
	// Wait until it's connected
	clientDME.WaitForConnect(1)
	clientCRM.WaitForConnect(1)
	assert.Equal(t, 0, len(dmeHandler.AppInstCache.Objs), "num appInsts")
	assert.Equal(t, uint64(0), clientDME.stats.Recv, "num updates")
	assert.Equal(t, NotifyVersion, clientDME.version, "version")
	assert.Equal(t, uint64(1), clientDME.stats.Connects, "connects")
	assert.Equal(t, uint64(1), clientCRM.stats.Connects, "connects")
	checkServerConnections(t, &serverMgr, 2)

	// Create some app insts which will trigger updates
	serverHandler.AppInstCache.Update(&testutil.AppInstData[0], 0)
	serverHandler.AppInstCache.Update(&testutil.AppInstData[1], 0)
	serverHandler.AppInstCache.Update(&testutil.AppInstData[2], 0)
	dmeHandler.WaitForAppInsts(3)
	assert.Equal(t, 3, len(dmeHandler.AppInstCache.Objs), "num appInsts")
	assert.Equal(t, uint64(3), clientDME.stats.AppInstRecv, "app inst updates")
	assert.Equal(t, uint64(3), clientDME.stats.Recv, "num updates")
	stats := serverMgr.GetStats(clientDME.GetLocalAddr())
	assert.Equal(t, uint64(3), stats.AppInstsSent)

	// Kill connection out from under the code, forcing reconnect
	fmt.Println("DME cancel")
	clientDME.cancel()
	// wait for it to reconnect
	clientDME.WaitForConnect(2)
	assert.Equal(t, uint64(2), clientDME.stats.Connects, "connects")
	checkServerConnections(t, &serverMgr, 2)

	// All cloudlets and all app insts will be sent again
	// Note on server side, this is a new connection so stats are reset
	serverHandler.AppInstCache.Update(&testutil.AppInstData[3], 0)
	dmeHandler.WaitForAppInsts(4)
	assert.Equal(t, 4, len(dmeHandler.AppInstCache.Objs), "num appInsts")
	assert.Equal(t, uint64(7), clientDME.stats.Recv, "num updates")
	stats = serverMgr.GetStats(clientDME.GetLocalAddr())
	assert.Equal(t, uint64(4), stats.AppInstsSent)

	// Delete an inst
	serverHandler.AppInstCache.Delete(&testutil.AppInstData[0], 0)
	dmeHandler.WaitForAppInsts(3)
	assert.Equal(t, 3, len(dmeHandler.AppInstCache.Objs), "num appInsts")
	assert.Equal(t, uint64(8), clientDME.stats.Recv, "num updates")
	assert.Equal(t, uint64(8), clientDME.stats.AppInstRecv, "app inst updates")
	stats = serverMgr.GetStats(clientDME.GetLocalAddr())
	assert.Equal(t, uint64(5), stats.AppInstsSent)

	// Stop DME, check that server closes connection as well
	fmt.Println("DME stop")
	clientDME.Stop()
	serverMgr.WaitServerCount(1)
	checkServerConnections(t, &serverMgr, 1)
	// reset data in handler, check that is it restored on reconnect
	edgeproto.InitAppInstCache(&dmeHandler.AppInstCache)
	clientDME.Start()
	clientDME.WaitForConnect(3)
	assert.Equal(t, uint64(3), clientDME.stats.Connects, "connects")
	dmeHandler.WaitForAppInsts(3)
	assert.Equal(t, 3, len(dmeHandler.AppInstCache.Objs), "num appInsts")

	// This time stop server, delete an inst, then start the
	// receiver again. The dmeHandler remains the same so none of
	// the data/stats changes. This tests that a delete during
	// disconnect is properly accounted for during the handling
	// of the sendall done command by removing the stale entry.
	fmt.Println("ServerMgr done")
	serverMgr.Stop()
	serverHandler.AppInstCache.Delete(&testutil.AppInstData[1], 0)
	serverMgr.Start(addr, serverHandler)
	clientDME.WaitForConnect(4)
	dmeHandler.WaitForAppInsts(2)
	assert.Equal(t, uint64(4), clientDME.stats.Connects, "connects")
	assert.Equal(t, 2, len(dmeHandler.AppInstCache.Objs), "num appInsts")
	assert.Equal(t, uint64(13), clientDME.stats.AppInstRecv, "app inst updates")
	stats = serverMgr.GetStats(clientDME.GetLocalAddr())
	assert.Equal(t, uint64(2), stats.AppInstsSent)
	fmt.Printf("stats for %s: %v\n", clientDME.GetLocalAddr(), stats)

	// Now test cloudlets. Use the same receiver, but register it
	// as a cloudlet mananger. CRM should have 2 connects since
	// server was restarted.
	clientCRM.WaitForConnect(1)
	assert.Equal(t, uint64(2), clientCRM.stats.Connects, "connects")
	serverHandler.CloudletCache.Update(&testutil.CloudletData[0], 0)
	serverHandler.CloudletCache.Update(&testutil.CloudletData[1], 0)
	crmHandler.WaitForCloudlets(2)
	assert.Equal(t, 2, len(crmHandler.CloudletCache.Objs), "num cloudlets")
	serverHandler.CloudletCache.Delete(&testutil.CloudletData[1], 0)
	crmHandler.WaitForCloudlets(1)
	assert.Equal(t, 1, len(crmHandler.CloudletCache.Objs), "num cloudlets")
	assert.Equal(t, uint64(3), clientCRM.stats.CloudletRecv, "cloudlet updates")
	stats = serverMgr.GetStats(clientCRM.GetLocalAddr())
	assert.Equal(t, uint64(3), stats.CloudletsSent, "sent cloudlets")
	checkServerConnections(t, &serverMgr, 2)

	// Send data from CRM to server
	fmt.Println("Create AppInstInfo")
	for _, ai := range testutil.AppInstData {
		info := edgeproto.AppInstInfo{}
		info.Key = ai.Key
		crmHandler.AppInstInfoCache.Update(&info, 0)
	}
	serverHandler.WaitForAppInstInfo(len(testutil.AppInstData))
	assert.Equal(t, len(testutil.AppInstData),
		len(serverHandler.AppInstInfoCache.Objs),
		"sent appInstInfo")

	for _, cl := range testutil.CloudletData {
		info := edgeproto.CloudletInfo{}
		info.Key = cl.Key
		crmHandler.CloudletInfoCache.Update(&info, 0)
	}
	serverHandler.WaitForCloudletInfo(len(testutil.CloudletData))
	assert.Equal(t, len(testutil.CloudletData),
		len(serverHandler.CloudletInfoCache.Objs),
		"sent cloudletInfo")

	clientDME.Stop()
	clientCRM.Stop()
}

func checkServerConnections(t *testing.T, serverMgr *ServerMgr, expected int) {
	serverMgr.mux.Lock()
	for addr, server := range serverMgr.table {
		log.DebugLog(log.DebugLevelNotify, "server connections", "client", addr, "stats", server.stats)
	}
	assert.Equal(t, expected, len(serverMgr.table), "num server connections")
	serverMgr.mux.Unlock()
}
