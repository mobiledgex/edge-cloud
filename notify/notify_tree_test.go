package notify

import (
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
)

// Test a tree organization of services.
// There is one node at the Top (controller)
// Two nodes in the middle (two CRM-Fronts)
// Four nodes in the low-end (four CRM-Backs)
// Each server node has two clients connected to it, so it is
// a balanced binary tree.
func TestNotifyTree(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelNotify)

	// override retry time
	NotifyRetryTime = 10 * time.Millisecond
	found := false

	// one top node
	top := newNode("127.0.0.1:60002", nil)
	// two mid nodes
	mid1 := newNode("127.0.0.1:60003", []string{"127.0.0.1:60002"})
	mid2 := newNode("127.0.0.1:60004", []string{"127.0.0.1:60002"})
	// four low nodes
	low11 := newNode("", []string{"127.0.0.1:60003"})
	low12 := newNode("", []string{"127.0.0.1:60003"})
	low21 := newNode("", []string{"127.0.0.1:60004"})
	low22 := newNode("", []string{"127.0.0.1:60004"})

	servers := []*node{top, mid1, mid2}
	clients := []*node{mid1, mid2, low11, low12, low21, low22}

	for _, n := range servers {
		n.startServer()
	}
	for _, n := range clients {
		n.startClient()
	}

	// wait till everything is connected
	for _, n := range clients {
		n.client.WaitForConnect(1)
	}

	// Add AppInst/Cloudlet to top server
	top.handler.AppInstCache.Update(&testutil.AppInstData[0], 0)
	top.handler.AppInstCache.Update(&testutil.AppInstData[1], 0)
	top.handler.AppInstCache.Update(&testutil.AppInstData[2], 0)
	top.handler.CloudletCache.Update(&testutil.CloudletData[0], 0)
	top.handler.CloudletCache.Update(&testutil.CloudletData[1], 0)
	// check that it reached all client nodes
	for _, n := range clients {
		checkClientCache(t, n, 3, 2)
	}

	// Add info objs to low nodes
	low11.handler.AppInstInfoCache.Update(&testutil.AppInstInfoData[0], 0)
	low12.handler.AppInstInfoCache.Update(&testutil.AppInstInfoData[1], 0)
	low21.handler.AppInstInfoCache.Update(&testutil.AppInstInfoData[2], 0)
	low22.handler.AppInstInfoCache.Update(&testutil.AppInstInfoData[3], 0)
	low11.handler.CloudletInfoCache.Update(&testutil.CloudletInfoData[0], 0)
	low12.handler.CloudletInfoCache.Update(&testutil.CloudletInfoData[1], 0)
	low21.handler.CloudletInfoCache.Update(&testutil.CloudletInfoData[2], 0)
	low22.handler.CloudletInfoCache.Update(&testutil.CloudletInfoData[3], 0)
	// check that each mid go two
	mid1.handler.WaitForAppInstInfo(2)
	mid1.handler.WaitForCloudletInfo(2)
	assert.Equal(t, 2, len(mid1.handler.AppInstInfoCache.Objs), "AppInstInfos")
	assert.Equal(t, 2, len(mid1.handler.CloudletInfoCache.Objs), "CloudletInfos")
	mid2.handler.WaitForAppInstInfo(2)
	mid2.handler.WaitForCloudletInfo(2)
	assert.Equal(t, 2, len(mid2.handler.AppInstInfoCache.Objs), "AppInstInfos")
	assert.Equal(t, 2, len(mid2.handler.CloudletInfoCache.Objs), "CloudletInfos")
	// check that top go all four
	top.handler.WaitForAppInstInfo(4)
	top.handler.WaitForCloudletInfo(4)
	assert.Equal(t, 4, len(top.handler.AppInstInfoCache.Objs), "AppInstInfos")
	assert.Equal(t, 4, len(top.handler.CloudletInfoCache.Objs), "CloudletInfos")

	// Check flush functionality
	// Disconnecting one of the low nodes should flush both mid and top
	// nodes of infos associated with the disconnected low node.
	low12.stopClient()
	mid1.handler.WaitForAppInstInfo(1)
	mid1.handler.WaitForCloudletInfo(1)
	assert.Equal(t, 1, len(mid1.handler.AppInstInfoCache.Objs), "AppInstInfos")
	assert.Equal(t, 1, len(mid1.handler.CloudletInfoCache.Objs), "CloudletInfos")
	_, found = mid1.handler.AppInstInfoCache.Objs[testutil.AppInstInfoData[1].Key]
	assert.False(t, found, "disconnected AppInstInfo")
	_, found = mid1.handler.CloudletInfoCache.Objs[testutil.CloudletInfoData[1].Key]
	assert.False(t, found, "disconnected CloudletInfo")

	top.handler.WaitForAppInstInfo(3)
	top.handler.WaitForCloudletInfo(3)
	assert.Equal(t, 3, len(top.handler.AppInstInfoCache.Objs), "AppInstInfos")
	assert.Equal(t, 3, len(top.handler.CloudletInfoCache.Objs), "CloudletInfos")
	_, found = top.handler.AppInstInfoCache.Objs[testutil.AppInstInfoData[1].Key]
	assert.False(t, found, "disconnected AppInstInfo")
	_, found = top.handler.CloudletInfoCache.Objs[testutil.CloudletInfoData[1].Key]
	assert.False(t, found, "disconnected CloudletInfo")

	for _, n := range clients {
		n.stopClient()
	}
	for _, n := range servers {
		n.stopServer()
	}
}

func checkClientCache(t *testing.T, n *node, appInsts int, cloudlets int) {
	n.handler.WaitForAppInsts(appInsts)
	n.handler.WaitForCloudlets(cloudlets)
	assert.Equal(t, appInsts, len(n.handler.AppInstCache.Objs), "num appinsts")
	assert.Equal(t, cloudlets, len(n.handler.CloudletCache.Objs), "num cloudlets")
}

type node struct {
	handler    *DummyHandler
	serverMgr  *ServerMgr
	client     *Client
	listenAddr string
}

func newNode(listenAddr string, connectAddrs []string) *node {
	n := &node{}
	n.handler = NewDummyHandler()
	n.listenAddr = listenAddr
	if listenAddr != "" {
		n.serverMgr = &ServerMgr{}
		n.handler.SetServerCb(n.serverMgr)
	}
	if connectAddrs != nil {
		n.client = NewCRMClient(connectAddrs, n.handler)
		n.handler.SetClientCb(n.client)
	}
	return n
}

func (n *node) startServer() {
	n.serverMgr.Start(n.listenAddr, n.handler)
}

func (n *node) startClient() {
	n.client.Start()
}

func (n *node) stopServer() {
	n.serverMgr.Stop()
}

func (n *node) stopClient() {
	n.client.Stop()
}
