package notify

import (
	"fmt"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestNotifyBasic(t *testing.T) {
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
	serverHandler.RegisterServer(&serverMgr)
	serverMgr.Start(addr, "")

	// Set up client DME
	dmeHandler := NewDummyHandler()
	clientDME := NewClient(serverAddrs, "")
	dmeHandler.RegisterDMEClient(clientDME)
	clientDME.Start()

	// Set up client CRM
	crmHandler := NewDummyHandler()
	clientCRM := NewClient(serverAddrs, "")
	crmHandler.RegisterCRMClient(clientCRM)
	clientCRM.Start()

	// It takes a little while for the Run thread to start up
	// Wait until it's connected
	clientDME.WaitForConnect(1)
	clientCRM.WaitForConnect(1)
	require.Equal(t, 0, len(dmeHandler.AppCache.Objs), "num Apps")
	require.Equal(t, 0, len(dmeHandler.AppInstCache.Objs), "num appInsts")
	require.Equal(t, uint64(0), clientDME.sendrecv.stats.Recv, "num updates")
	require.Equal(t, NotifyVersion, clientDME.version, "version")
	require.Equal(t, uint64(1), clientDME.sendrecv.stats.Connects, "connects")
	require.Equal(t, uint64(1), clientCRM.sendrecv.stats.Connects, "connects")
	checkServerConnections(t, &serverMgr, 2)

	// Create some app insts which will trigger updates
	serverHandler.AppCache.Update(&testutil.AppData[0], 0)
	serverHandler.AppCache.Update(&testutil.AppData[1], 0)
	serverHandler.AppCache.Update(&testutil.AppData[2], 0)
	serverHandler.AppCache.Update(&testutil.AppData[3], 0)
	serverHandler.AppCache.Update(&testutil.AppData[4], 0)
	dmeHandler.WaitForAppInsts(5)
	require.Equal(t, 5, len(dmeHandler.AppCache.Objs), "num Apps")
	stats := serverMgr.GetStats(clientDME.GetLocalAddr())
	require.Equal(t, uint64(5), stats.ObjSend["App"])

	// Create some app insts which will trigger updates
	serverHandler.AppInstCache.Update(&testutil.AppInstData[0], 0)
	serverHandler.AppInstCache.Update(&testutil.AppInstData[1], 0)
	serverHandler.AppInstCache.Update(&testutil.AppInstData[2], 0)
	dmeHandler.WaitForAppInsts(3)
	require.Equal(t, 3, len(dmeHandler.AppInstCache.Objs), "num appInsts")
	clientDME.GetStats(stats)
	require.Equal(t, uint64(3), stats.ObjRecv["AppInst"], "app inst updates")
	require.Equal(t, uint64(8), stats.Recv, "num updates")
	stats = serverMgr.GetStats(clientDME.GetLocalAddr())
	require.Equal(t, uint64(3), stats.ObjSend["AppInst"])

	// Kill connection out from under the code, forcing reconnect
	fmt.Println("DME cancel")
	clientDME.cancel()
	// wait for it to reconnect
	clientDME.WaitForConnect(2)
	require.Equal(t, uint64(2), clientDME.sendrecv.stats.Connects, "connects")
	checkServerConnections(t, &serverMgr, 2)

	// All cloudlets and all app insts will be sent again
	// Note on server side, this is a new connection so stats are reset
	serverHandler.AppInstCache.Update(&testutil.AppInstData[3], 0)
	dmeHandler.WaitForAppInsts(4)
	require.Equal(t, 4, len(dmeHandler.AppInstCache.Objs), "num appInsts")
	require.Equal(t, uint64(17), clientDME.sendrecv.stats.Recv, "num updates")
	stats = serverMgr.GetStats(clientDME.GetLocalAddr())
	require.Equal(t, uint64(4), stats.ObjSend["AppInst"])
	require.Equal(t, uint64(5), stats.ObjSend["App"])

	// Delete an inst
	serverHandler.AppInstCache.Delete(&testutil.AppInstData[0], 0)
	dmeHandler.WaitForAppInsts(3)
	require.Equal(t, 3, len(dmeHandler.AppInstCache.Objs), "num appInsts")
	require.Equal(t, uint64(18), clientDME.sendrecv.stats.Recv, "num updates")
	clientDME.GetStats(stats)
	require.Equal(t, uint64(8), stats.ObjRecv["AppInst"], "app inst updates")
	stats = serverMgr.GetStats(clientDME.GetLocalAddr())
	require.Equal(t, uint64(5), stats.ObjSend["AppInst"])
	require.Equal(t, uint64(5), stats.ObjSend["App"])

	// Stop DME, check that server closes connection as well
	fmt.Println("DME stop")
	clientDME.Stop()
	serverMgr.WaitServerCount(1)
	checkServerConnections(t, &serverMgr, 1)
	// reset data in handler, check that is it restored on reconnect
	edgeproto.InitAppInstCache(&dmeHandler.AppInstCache)
	clientDME.Start()
	clientDME.WaitForConnect(3)
	require.Equal(t, uint64(3), clientDME.sendrecv.stats.Connects, "connects")
	dmeHandler.WaitForAppInsts(3)
	require.Equal(t, 3, len(dmeHandler.AppInstCache.Objs), "num appInsts")

	// This time stop server, delete an inst, then start the
	// receiver again. The dmeHandler remains the same so none of
	// the data/stats changes. This tests that a delete during
	// disconnect is properly accounted for during the handling
	// of the sendall done command by removing the stale entry.
	fmt.Println("ServerMgr done")
	serverMgr.Stop()
	serverHandler.AppInstCache.Delete(&testutil.AppInstData[1], 0)
	serverMgr.Start(addr, "")
	clientDME.WaitForConnect(4)
	dmeHandler.WaitForAppInsts(2)
	require.Equal(t, uint64(4), clientDME.sendrecv.stats.Connects, "connects")
	require.Equal(t, 2, len(dmeHandler.AppInstCache.Objs), "num appInsts")
	clientDME.GetStats(stats)
	require.Equal(t, uint64(13), stats.ObjRecv["AppInst"], "app inst updates")
	stats = serverMgr.GetStats(clientDME.GetLocalAddr())
	require.Equal(t, uint64(2), stats.ObjSend["AppInst"])

	// set ClusterInst and AppInst state to CreateRequested so they get
	// sent to the CRM.
	for i, _ := range testutil.ClusterInstData {
		testutil.ClusterInstData[i].State = edgeproto.TrackedState_CreateRequested
	}
	for i, _ := range testutil.AppInstData {
		testutil.AppInstData[i].State = edgeproto.TrackedState_CreateRequested
	}

	// Now test cloudlets. Use the same receiver, but register it
	// as a cloudlet mananger. CRM should have 2 connects since
	// server was restarted.
	clientCRM.WaitForConnect(1)
	require.Equal(t, uint64(2), clientCRM.sendrecv.stats.Connects, "connects")
	// crm must send cloudletinfo to receive clusterInsts and appInsts
	crmHandler.CloudletInfoCache.Update(&testutil.CloudletInfoData[0], 0)
	serverHandler.CloudletCache.Update(&testutil.CloudletData[0], 0)
	serverHandler.CloudletCache.Update(&testutil.CloudletData[1], 0)
	serverHandler.FlavorCache.Update(&testutil.FlavorData[0], 0)
	serverHandler.FlavorCache.Update(&testutil.FlavorData[1], 0)
	serverHandler.FlavorCache.Update(&testutil.FlavorData[2], 0)
	serverHandler.ClusterFlavorCache.Update(&testutil.ClusterFlavorData[0], 0)
	serverHandler.ClusterFlavorCache.Update(&testutil.ClusterFlavorData[1], 0)
	serverHandler.ClusterFlavorCache.Update(&testutil.ClusterFlavorData[2], 0)
	serverHandler.ClusterInstCache.Update(&testutil.ClusterInstData[0], 0)
	serverHandler.ClusterInstCache.Update(&testutil.ClusterInstData[1], 0)
	serverHandler.ClusterInstCache.Update(&testutil.ClusterInstData[2], 0)
	serverHandler.ClusterInstCache.Update(&testutil.ClusterInstData[3], 0)
	serverHandler.AppInstCache.Update(&testutil.AppInstData[0], 0)
	serverHandler.AppInstCache.Update(&testutil.AppInstData[1], 0)
	serverHandler.AppInstCache.Update(&testutil.AppInstData[2], 0)
	serverHandler.AppInstCache.Update(&testutil.AppInstData[3], 0)
	// Note: only ClusterInsts and AppInsts with cloudlet keys that
	// match the CRM's cloudletinfo will be sent.
	crmHandler.WaitForCloudlets(1)
	crmHandler.WaitForFlavors(3)
	crmHandler.WaitForClusterFlavors(3)
	crmHandler.WaitForClusterInsts(2)
	crmHandler.WaitForApps(1)
	crmHandler.WaitForAppInsts(2)
	require.Equal(t, 1, len(crmHandler.CloudletCache.Objs), "num cloudlets")
	require.Equal(t, 3, len(crmHandler.FlavorCache.Objs), "num flavors")
	require.Equal(t, 3, len(crmHandler.ClusterFlavorCache.Objs), "num cluster flavors")
	require.Equal(t, 2, len(crmHandler.ClusterInstCache.Objs), "num clusterInsts")
	require.Equal(t, 1, len(crmHandler.AppCache.Objs), "num apps")
	require.Equal(t, 2, len(crmHandler.AppInstCache.Objs), "num appInsts")
	serverHandler.FlavorCache.Delete(&testutil.FlavorData[1], 0)
	serverHandler.ClusterFlavorCache.Delete(&testutil.ClusterFlavorData[1], 0)
	serverHandler.ClusterInstCache.Delete(&testutil.ClusterInstData[0], 0)
	serverHandler.AppInstCache.Delete(&testutil.AppInstData[0], 0)
	crmHandler.WaitForFlavors(2)
	crmHandler.WaitForClusterFlavors(2)
	crmHandler.WaitForClusterInsts(1)
	crmHandler.WaitForAppInsts(1)
	require.Equal(t, 2, len(crmHandler.FlavorCache.Objs), "num flavors")
	require.Equal(t, 2, len(crmHandler.ClusterFlavorCache.Objs), "num cluster flavors")
	require.Equal(t, 1, len(crmHandler.ClusterInstCache.Objs), "num clusterInsts")
	require.Equal(t, 1, len(crmHandler.AppInstCache.Objs), "num appInsts")
	clientCRM.GetStats(stats)
	require.Equal(t, uint64(1), stats.ObjRecv["Cloudlet"], "cloudlet updates")
	require.Equal(t, uint64(4), stats.ObjRecv["Flavor"], "flavor updates")
	require.Equal(t, uint64(4), stats.ObjRecv["ClusterFlavor"], "cluster flavor updates")
	require.Equal(t, uint64(3), stats.ObjRecv["ClusterInst"], "clusterInst updates")
	require.Equal(t, uint64(3), stats.ObjRecv["AppInst"], "appInst updates")
	stats = serverMgr.GetStats(clientCRM.GetLocalAddr())
	require.Equal(t, uint64(1), stats.ObjSend["Cloudlet"], "sent cloudlets")
	require.Equal(t, uint64(4), stats.ObjSend["Flavor"], "sent flavors")
	require.Equal(t, uint64(4), stats.ObjSend["ClusterFlavor"], "sent cluster flavors")
	require.Equal(t, uint64(3), stats.ObjSend["ClusterInst"], "sent clusterInsts")
	require.Equal(t, uint64(3), stats.ObjSend["AppInst"], "sent appInsts")
	checkServerConnections(t, &serverMgr, 2)

	// Send data from CRM to server
	fmt.Println("Create AppInstInfo")
	for _, ai := range testutil.AppInstData {
		info := edgeproto.AppInstInfo{}
		info.Key = ai.Key
		crmHandler.AppInstInfoCache.Update(&info, 0)
	}
	serverHandler.WaitForAppInstInfo(len(testutil.AppInstData))
	require.Equal(t, len(testutil.AppInstData),
		len(serverHandler.AppInstInfoCache.Objs),
		"sent appInstInfo")

	for _, ci := range testutil.ClusterInstData {
		info := edgeproto.ClusterInstInfo{}
		info.Key = ci.Key
		crmHandler.ClusterInstInfoCache.Update(&info, 0)
	}
	serverHandler.WaitForClusterInstInfo(len(testutil.ClusterInstData))
	require.Equal(t, len(testutil.ClusterInstData),
		len(serverHandler.ClusterInstInfoCache.Objs),
		"sent clusterInstInfo")

	for _, cl := range testutil.CloudletData {
		info := edgeproto.CloudletInfo{}
		info.Key = cl.Key
		crmHandler.CloudletInfoCache.Update(&info, 0)
	}
	serverHandler.WaitForCloudletInfo(len(testutil.CloudletData))
	require.Equal(t, len(testutil.CloudletData),
		len(serverHandler.CloudletInfoCache.Objs),
		"sent cloudletInfo")

	clientDME.Stop()
	clientCRM.Stop()
}

func checkServerConnections(t *testing.T, serverMgr *ServerMgr, expected int) {
	serverMgr.mux.Lock()
	for addr, server := range serverMgr.table {
		log.DebugLog(log.DebugLevelNotify, "server connections", "client", addr, "stats", server.sendrecv.stats)
	}
	require.Equal(t, expected, len(serverMgr.table), "num server connections")
	serverMgr.mux.Unlock()
}
