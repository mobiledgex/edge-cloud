package notify

import (
	"fmt"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/assert"
)

func TestNotify(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelNotify)

	// override retry time
	NotifyRetryTime = 10 * time.Millisecond

	// This tests the server sending notices to
	// a client.
	addr := "127.0.0.1:61234"
	serverAddrs := []string{addr}

	// Set up server
	serverHandler := NewDummyServerHandler()
	ServerMgrStart(addr, serverHandler)

	// Set up client DME
	dmeHandler := NewDummyClientHandler()
	clientDME := NewDMEClient(serverAddrs, dmeHandler)
	go clientDME.Run()

	// Set up client CRM
	crmHandler := NewDummyClientHandler()
	clientCRM := NewCRMClient(serverAddrs, crmHandler)
	go clientCRM.Run()

	// It takes a little while for the Run thread to start up
	// Wait until it's connected
	clientDME.WaitForConnect(1)
	clientCRM.WaitForConnect(1)
	assert.Equal(t, 0, len(dmeHandler.AppInsts), "num appInsts")
	assert.Equal(t, 0, dmeHandler.NumUpdates, "num updates")
	assert.Equal(t, NotifyVersion, clientDME.version, "version")
	assert.Equal(t, uint64(1), clientDME.stats.Connects, "connects")
	assert.Equal(t, uint64(1), clientCRM.stats.Connects, "connects")
	checkServerConnections(t, 2)

	// Create some app insts which will trigger updates
	serverHandler.CreateAppInst(&testutil.AppInstData[0])
	serverHandler.CreateAppInst(&testutil.AppInstData[1])
	serverHandler.CreateAppInst(&testutil.AppInstData[2])
	dmeHandler.WaitForAppInsts(3)
	assert.Equal(t, 3, len(dmeHandler.AppInsts), "num appInsts")
	assert.Equal(t, 3, dmeHandler.NumAppInstUpdates, "num app inst updates")
	assert.Equal(t, 3, dmeHandler.NumUpdates, "num updates")
	stats := GetServerStats(clientDME.GetLocalAddr())
	assert.Equal(t, uint64(3), stats.AppInstsSent)

	// Kill connection out from under the code, forcing reconnect
	fmt.Println("DME cancel")
	clientDME.cancel()
	// wait for it to reconnect
	clientDME.WaitForConnect(2)
	assert.Equal(t, uint64(2), clientDME.stats.Connects, "connects")
	checkServerConnections(t, 2)

	// All cloudlets and all app insts will be sent again
	// Note on server side, this is a new connection so stats are reset
	serverHandler.CreateAppInst(&testutil.AppInstData[3])
	dmeHandler.WaitForAppInsts(4)
	assert.Equal(t, 4, len(dmeHandler.AppInsts), "num appInsts")
	assert.Equal(t, 7, dmeHandler.NumUpdates, "num updates")
	stats = GetServerStats(clientDME.GetLocalAddr())
	assert.Equal(t, uint64(4), stats.AppInstsSent)

	// Delete an inst
	serverHandler.DeleteAppInst(&testutil.AppInstData[0])
	dmeHandler.WaitForAppInsts(3)
	assert.Equal(t, 3, len(dmeHandler.AppInsts), "num appInsts")
	assert.Equal(t, 8, dmeHandler.NumUpdates, "num updates")
	assert.Equal(t, 8, dmeHandler.NumAppInstUpdates, "num app inst updates")
	stats = GetServerStats(clientDME.GetLocalAddr())
	assert.Equal(t, uint64(5), stats.AppInstsSent)

	// Stop DME, check that server closes connection as well
	fmt.Println("DME stop")
	clientDME.Stop()
	WaitServerCount(1)
	checkServerConnections(t, 1)
	// reset data in handler, check that is it restored on reconnect
	dmeHandler.AppInsts = make(map[edgeproto.AppInstKey]edgeproto.AppInst)
	go clientDME.Run()
	clientDME.WaitForConnect(3)
	assert.Equal(t, uint64(3), clientDME.stats.Connects, "connects")
	dmeHandler.WaitForAppInsts(3)
	assert.Equal(t, 3, len(dmeHandler.AppInsts), "num appInsts")

	// This time stop server, delete an inst, then start the
	// receiver again. The dmeHandler remains the same so none of
	// the data/stats changes. This tests that a delete during
	// disconnect is properly accounted for during the handling
	// of the sendall done command by removing the stale entry.
	fmt.Println("ServerMgr done")
	ServerMgrDone()
	serverHandler.DeleteAppInst(&testutil.AppInstData[1])
	ServerMgrStart(addr, serverHandler)
	clientDME.WaitForConnect(4)
	dmeHandler.WaitForAppInsts(2)
	assert.Equal(t, uint64(4), clientDME.stats.Connects, "connects")
	assert.Equal(t, 2, len(dmeHandler.AppInsts), "num appInsts")
	assert.Equal(t, 13, dmeHandler.NumAppInstUpdates, "num app inst updates")
	stats = GetServerStats(clientDME.GetLocalAddr())
	assert.Equal(t, uint64(2), stats.AppInstsSent)

	// Now test cloudlets. Use the same receiver, but register it
	// as a cloudlet mananger. CRM should have 2 connects since
	// server was restarted.
	clientCRM.WaitForConnect(1)
	assert.Equal(t, uint64(2), clientCRM.stats.Connects, "connects")
	serverHandler.CreateCloudlet(&testutil.CloudletData[0])
	serverHandler.CreateCloudlet(&testutil.CloudletData[1])
	crmHandler.WaitForCloudlets(2)
	assert.Equal(t, 2, len(crmHandler.Cloudlets), "num cloudlets")
	serverHandler.DeleteCloudlet(&testutil.CloudletData[1])
	crmHandler.WaitForCloudlets(1)
	assert.Equal(t, 1, len(crmHandler.Cloudlets), "num cloudlets")
	assert.Equal(t, 3, crmHandler.NumCloudletUpdates, "num cloudlet updates")
	stats = GetServerStats(clientCRM.GetLocalAddr())
	assert.Equal(t, uint64(3), stats.CloudletsSent, "sent cloudlets")
	checkServerConnections(t, 2)

	// Send data from CRM to server
	for _, ai := range testutil.AppInstData {
		msg := edgeproto.NoticeRequest{}
		data := &edgeproto.NoticeRequest_AppInstInfo{}
		data.AppInstInfo = &edgeproto.AppInstInfo{}
		data.AppInstInfo.Key = ai.Key
		msg.Data = data
		err := clientCRM.Send(&msg)
		assert.Nil(t, err, "send to server")
	}
	serverHandler.WaitForAppInstInfo(len(testutil.AppInstData))
	assert.Equal(t, len(testutil.AppInstData), len(serverHandler.AppInstsInfo),
		"sent appInstInfo")

	for _, cl := range testutil.CloudletData {
		msg := edgeproto.NoticeRequest{}
		data := &edgeproto.NoticeRequest_CloudletInfo{}
		data.CloudletInfo = &edgeproto.CloudletInfo{}
		data.CloudletInfo.Key = cl.Key
		msg.Data = data
		err := clientCRM.Send(&msg)
		assert.Nil(t, err, "send to server")
	}
	serverHandler.WaitForCloudletInfo(len(testutil.CloudletData))
	assert.Equal(t, len(testutil.CloudletData), len(serverHandler.CloudletsInfo),
		"sent cloudletInfo")

	clientDME.Stop()
	clientCRM.Stop()
}

func checkServerConnections(t *testing.T, expected int) {
	serverMgr.mux.Lock()
	for addr, server := range serverMgr.table {
		util.DebugLog(util.DebugLevelNotify, "server connections", "client", addr, "stats", server.stats)
	}
	assert.Equal(t, expected, len(serverMgr.table), "num server connections")
	serverMgr.mux.Unlock()
}
