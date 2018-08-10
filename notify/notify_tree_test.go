package notify

import (
	"fmt"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
)

type nodeType int32

const (
	none nodeType = iota
	dme
	crm
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
	top := newNode("127.0.0.1:60002", nil, none)
	// two mid nodes
	mid1 := newNode("127.0.0.1:60003", []string{"127.0.0.1:60002"}, dme)
	mid2 := newNode("127.0.0.1:60004", []string{"127.0.0.1:60002"}, crm)
	// four low nodes
	low11 := newNode("", []string{"127.0.0.1:60003"}, dme)
	low12 := newNode("", []string{"127.0.0.1:60003"}, dme)
	low21 := newNode("", []string{"127.0.0.1:60004"}, crm)
	low22 := newNode("", []string{"127.0.0.1:60004"}, crm)

	// Note that data distributed to dme and crm are different.

	servers := []*node{top, mid1, mid2}
	clients := []*node{mid1, mid2, low11, low12, low21, low22}
	dmes := []*node{mid1, low11, low12}

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

	// crms need to send cloudletInfo up to trigger sending of
	// data down.
	low21.handler.CloudletInfoCache.Update(&testutil.CloudletInfoData[0], 0)
	low22.handler.CloudletInfoCache.Update(&testutil.CloudletInfoData[1], 0)

	// Add data to top server
	top.handler.FlavorCache.Update(&testutil.FlavorData[0], 0)
	top.handler.FlavorCache.Update(&testutil.FlavorData[1], 0)
	top.handler.FlavorCache.Update(&testutil.FlavorData[2], 0)
	for ii, _ := range testutil.ClusterInstData {
		top.handler.ClusterInstCache.Update(&testutil.ClusterInstData[ii], 0)
	}
	for ii, _ := range testutil.AppInstData {
		top.handler.AppInstCache.Update(&testutil.AppInstData[ii], 0)
	}
	top.handler.CloudletCache.Update(&testutil.CloudletData[0], 0)
	top.handler.CloudletCache.Update(&testutil.CloudletData[1], 0)
	// dmes should get all app insts but no cloudlets
	for _, n := range dmes {
		checkClientCache(t, n, 0, 0, 5, 0)
	}
	// crms at all levels get all flavors
	// mid level gets the appInsts and clusterInsts for all below it.
	// low level only gets the ones for itself.
	checkClientCache(t, low21, 3, 2, 2, 1)
	checkClientCache(t, low22, 3, 2, 2, 1)
	checkClientCache(t, mid2, 3, 4, 4, 2)

	// Add info objs to low nodes
	low11.handler.AppInstInfoCache.Update(&testutil.AppInstInfoData[0], 0)
	low12.handler.AppInstInfoCache.Update(&testutil.AppInstInfoData[1], 0)
	low21.handler.AppInstInfoCache.Update(&testutil.AppInstInfoData[2], 0)
	low22.handler.AppInstInfoCache.Update(&testutil.AppInstInfoData[3], 0)
	// check that each mid go two
	mid1.handler.WaitForAppInstInfo(2)
	mid1.handler.WaitForCloudletInfo(2)
	assert.Equal(t, 2, len(mid1.handler.AppInstInfoCache.Objs), "AppInstInfos")
	assert.Equal(t, 0, len(mid1.handler.CloudletInfoCache.Objs), "CloudletInfos")
	mid2.handler.WaitForAppInstInfo(2)
	mid2.handler.WaitForCloudletInfo(2)
	assert.Equal(t, 2, len(mid2.handler.AppInstInfoCache.Objs), "AppInstInfos")
	assert.Equal(t, 2, len(mid2.handler.CloudletInfoCache.Objs), "CloudletInfos")
	// check that top go all four
	top.handler.WaitForAppInstInfo(4)
	top.handler.WaitForCloudletInfo(4)
	assert.Equal(t, 4, len(top.handler.AppInstInfoCache.Objs), "AppInstInfos")
	assert.Equal(t, 2, len(top.handler.CloudletInfoCache.Objs), "CloudletInfos")

	// Check flush functionality
	// Disconnecting one of the low nodes should flush both mid and top
	// nodes of infos associated with the disconnected low node.
	fmt.Println("========== stopping client")
	low21.stopClient()
	mid2.handler.WaitForAppInstInfo(1)
	mid1.handler.WaitForCloudletInfo(1)
	assert.Equal(t, 1, len(mid2.handler.AppInstInfoCache.Objs), "AppInstInfos")
	assert.Equal(t, 1, len(mid2.handler.CloudletInfoCache.Objs), "CloudletInfos")
	_, found = mid2.handler.AppInstInfoCache.Objs[testutil.AppInstInfoData[2].Key]
	assert.False(t, found, "disconnected AppInstInfo")
	_, found = mid2.handler.CloudletInfoCache.Objs[testutil.CloudletInfoData[0].Key]
	assert.False(t, found, "disconnected CloudletInfo")

	top.handler.WaitForAppInstInfo(3)
	top.handler.WaitForCloudletInfo(1)
	assert.Equal(t, 3, len(top.handler.AppInstInfoCache.Objs), "AppInstInfos")
	assert.Equal(t, 1, len(top.handler.CloudletInfoCache.Objs), "CloudletInfos")
	_, found = top.handler.AppInstInfoCache.Objs[testutil.AppInstInfoData[2].Key]
	assert.False(t, found, "disconnected AppInstInfo")
	_, found = top.handler.CloudletInfoCache.Objs[testutil.CloudletInfoData[0].Key]
	assert.False(t, found, "disconnected CloudletInfo")
	fmt.Println("========== done")

	for _, n := range clients {
		n.stopClient()
	}
	for _, n := range servers {
		n.stopServer()
	}
}

func checkClientCache(t *testing.T, n *node, flavors int, clusterInsts int, appInsts int, cloudlets int) {
	n.handler.WaitForFlavors(flavors)
	n.handler.WaitForClusterInsts(clusterInsts)
	n.handler.WaitForAppInsts(appInsts)
	n.handler.WaitForCloudlets(cloudlets)
	assert.Equal(t, flavors, len(n.handler.FlavorCache.Objs), "num flavors")
	assert.Equal(t, clusterInsts, len(n.handler.ClusterInstCache.Objs), "num clusterinsts")
	assert.Equal(t, appInsts, len(n.handler.AppInstCache.Objs), "num appinsts")
	assert.Equal(t, cloudlets, len(n.handler.CloudletCache.Objs), "num cloudlets")
}

type node struct {
	handler    *DummyHandler
	serverMgr  *ServerMgr
	client     *Client
	listenAddr string
}

func newNode(listenAddr string, connectAddrs []string, typ nodeType) *node {
	n := &node{}
	n.handler = NewDummyHandler()
	n.listenAddr = listenAddr
	if listenAddr != "" {
		n.serverMgr = &ServerMgr{}
		n.handler.SetServerCb(n.serverMgr)
	}
	if connectAddrs != nil {
		if typ == crm {
			n.client = NewCRMClient(connectAddrs, n.handler)
		} else {
			n.client = NewDMEClient(connectAddrs, n.handler)
		}
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
