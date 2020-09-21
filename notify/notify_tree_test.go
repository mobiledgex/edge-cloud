package notify

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
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
	log.SetDebugLevel(log.DebugLevelNotify | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	// override retry time
	NotifyRetryTime = 10 * time.Millisecond
	found := false

	// one top node
	top := newNode("top", "127.0.0.1:60002", nil, none)
	// two mid nodes
	mid1 := newNode("mid1", "127.0.0.1:60003", []string{"127.0.0.1:60002"}, dme)
	mid2 := newNode("mid2", "127.0.0.1:60004", []string{"127.0.0.1:60002"}, crm)
	// four low nodes
	low11 := newNode("low11", "", []string{"127.0.0.1:60003"}, dme)
	low12 := newNode("low12", "", []string{"127.0.0.1:60003"}, dme)
	low21 := newNode("low21", "", []string{"127.0.0.1:60004"}, crm)
	low22 := newNode("low22", "", []string{"127.0.0.1:60004"}, crm)

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
	low21.handler.CloudletInfoCache.Update(ctx, &testutil.CloudletInfoData[0], 0)
	low22.handler.CloudletInfoCache.Update(ctx, &testutil.CloudletInfoData[1], 0)

	// Add data to top server
	top.handler.FlavorCache.Update(ctx, &testutil.FlavorData[0], 0)
	top.handler.FlavorCache.Update(ctx, &testutil.FlavorData[1], 0)
	top.handler.FlavorCache.Update(ctx, &testutil.FlavorData[2], 0)

	// set ClusterInst and AppInst state to CREATE_REQUESTED so they get
	// sent to the CRM.
	for ii, _ := range testutil.ClusterInstData {
		testutil.ClusterInstData[ii].State = edgeproto.TrackedState_CREATE_REQUESTED
		top.handler.ClusterInstCache.Update(ctx, &testutil.ClusterInstData[ii], 0)
	}
	for ii, _ := range testutil.AppInstData {
		testutil.AppInstData[ii].State = edgeproto.TrackedState_CREATE_REQUESTED
		top.handler.AppInstCache.Update(ctx, &testutil.AppInstData[ii], 0)
	}
	top.handler.CloudletCache.Update(ctx, &testutil.CloudletData[0], 0)
	top.handler.CloudletCache.Update(ctx, &testutil.CloudletData[1], 0)
	// dmes should get all app insts but no cloudlets
	for _, n := range dmes {
		checkClientCache(t, n, 0, 0, len(testutil.AppInstData), 0)
	}
	// crms at all levels get all flavors
	// mid level gets the appInsts and clusterInsts for all below it.
	// low level only gets the ones for itself.
	checkClientCache(t, low21, 3, 3, 6, 1)
	checkClientCache(t, low22, 3, 2, 3, 1)
	checkClientCache(t, mid2, 3, 5, 9, 2)
	checkCache(t, mid1, FreeReservableClusterInstType, 1)

	// Add info objs to low nodes
	low11.handler.AppInstInfoCache.Update(ctx, &testutil.AppInstInfoData[0], 0)
	low12.handler.AppInstInfoCache.Update(ctx, &testutil.AppInstInfoData[1], 0)
	low21.handler.AppInstInfoCache.Update(ctx, &testutil.AppInstInfoData[2], 0)
	low22.handler.AppInstInfoCache.Update(ctx, &testutil.AppInstInfoData[3], 0)
	// dme mid should get 0 because it doesn't want infos
	// crm mid should get 2, 1 from each crm low
	mid1.handler.WaitForAppInstInfo(0)
	mid1.handler.WaitForCloudletInfo(0)
	require.Equal(t, 0, len(mid1.handler.AppInstInfoCache.Objs), "AppInstInfos")
	require.Equal(t, 0, len(mid1.handler.CloudletInfoCache.Objs), "CloudletInfos")
	mid2.handler.WaitForAppInstInfo(2)
	mid2.handler.WaitForCloudletInfo(2)
	require.Equal(t, 2, len(mid2.handler.AppInstInfoCache.Objs), "AppInstInfos")
	require.Equal(t, 2, len(mid2.handler.CloudletInfoCache.Objs), "CloudletInfos")
	// check that top got 2 propagated from crm mid
	top.handler.WaitForAppInstInfo(2)
	top.handler.WaitForCloudletInfo(2)
	require.Equal(t, 2, len(top.handler.AppInstInfoCache.Objs), "AppInstInfos")
	require.Equal(t, 2, len(top.handler.CloudletInfoCache.Objs), "CloudletInfos")
	// add alerts
	checkCache(t, top, AlertType, 0)
	checkCache(t, mid1, AlertType, 0)
	checkCache(t, mid2, AlertType, 0)
	checkCache(t, low11, AlertType, 0)
	checkCache(t, low12, AlertType, 0)
	checkCache(t, low21, AlertType, 0)
	checkCache(t, low22, AlertType, 0)
	low21.handler.AlertCache.Update(ctx, &testutil.AlertData[0], 0)
	low21.handler.AlertCache.Update(ctx, &testutil.AlertData[1], 0)
	low22.handler.AlertCache.Update(ctx, &testutil.AlertData[2], 0)
	checkCache(t, top, AlertType, 3)
	checkCache(t, mid1, AlertType, 0)
	checkCache(t, mid2, AlertType, 3)
	checkCache(t, low11, AlertType, 0)
	checkCache(t, low12, AlertType, 0)
	checkCache(t, low21, AlertType, 2)
	checkCache(t, low22, AlertType, 1)

	// Check flush functionality
	// Disconnecting one of the low nodes should flush both mid and top
	// nodes of infos associated with the disconnected low node.
	fmt.Println("========== stopping client")
	low21.stopClient()
	mid2.handler.WaitForAppInstInfo(1)
	mid2.handler.WaitForCloudletInfo(1)
	require.Equal(t, 1, len(mid2.handler.AppInstInfoCache.Objs), "AppInstInfos")
	require.Equal(t, 1, len(mid2.handler.CloudletInfoCache.Objs), "CloudletInfos")
	require.Equal(t, 0, len(mid1.handler.CloudletInfoCache.Objs), "CloudletInfos")
	_, found = mid2.handler.AppInstInfoCache.Objs[testutil.AppInstInfoData[2].Key]
	require.False(t, found, "disconnected AppInstInfo")
	_, found = mid2.handler.CloudletInfoCache.Objs[testutil.CloudletInfoData[0].Key]
	require.False(t, found, "disconnected CloudletInfo")

	top.handler.WaitForAppInstInfo(1)
	top.handler.WaitForCloudletInfo(1)
	require.Equal(t, 1, len(top.handler.AppInstInfoCache.Objs), "AppInstInfos")
	require.Equal(t, 1, len(top.handler.CloudletInfoCache.Objs), "CloudletInfos")
	_, found = top.handler.AppInstInfoCache.Objs[testutil.AppInstInfoData[2].Key]
	require.False(t, found, "disconnected AppInstInfo")
	_, found = top.handler.CloudletInfoCache.Objs[testutil.CloudletInfoData[0].Key]
	require.False(t, found, "disconnected CloudletInfo")

	checkCache(t, top, AlertType, 1)
	checkCache(t, mid1, AlertType, 0)
	checkCache(t, mid2, AlertType, 1)

	low21.startClient()
	checkClientCache(t, low21, 3, 3, 6, 1)
	checkClientCache(t, low22, 3, 2, 3, 1)
	checkClientCache(t, mid2, 3, 5, 9, 2)
	checkCache(t, mid1, FreeReservableClusterInstType, 1)
	mid2.handler.WaitForCloudletInfo(2)

	fmt.Println("========== cleanup")

	// Delete objects to make sure deletes propagate and are applied
	for ii, _ := range testutil.AppInstData {
		top.handler.AppInstCache.Delete(ctx, &testutil.AppInstData[ii], 0)
	}
	for ii, _ := range testutil.ClusterInstData {
		log.SpanLog(ctx, log.DebugLevelNotify, "deleting ClusterInst", "key", testutil.ClusterInstData[ii].Key)
		top.handler.ClusterInstCache.Delete(ctx, &testutil.ClusterInstData[ii], 0)
	}
	top.handler.FlavorCache.Delete(ctx, &testutil.FlavorData[0], 0)
	top.handler.FlavorCache.Delete(ctx, &testutil.FlavorData[1], 0)
	top.handler.FlavorCache.Delete(ctx, &testutil.FlavorData[2], 0)
	top.handler.CloudletCache.Delete(ctx, &testutil.CloudletData[0], 0)
	top.handler.CloudletCache.Delete(ctx, &testutil.CloudletData[1], 0)
	checkClientCache(t, mid2, 0, 0, 0, 0)
	checkClientCache(t, low21, 0, 0, 0, 0)
	checkClientCache(t, low22, 0, 0, 0, 0)
	checkCache(t, mid1, FreeReservableClusterInstType, 0)

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
	fmt.Printf("%v counts: %d, %d, %d, %d\n", n,
		len(n.handler.FlavorCache.Objs),
		len(n.handler.ClusterInstCache.Objs),
		len(n.handler.AppInstCache.Objs),
		len(n.handler.CloudletCache.Objs))
	if n.client != nil {
		fmt.Printf("client %+v\n", n.client)
	}
	require.Equal(t, flavors, len(n.handler.FlavorCache.Objs), "num flavors")
	require.Equal(t, clusterInsts, len(n.handler.ClusterInstCache.Objs), "num clusterinsts")
	require.Equal(t, appInsts, len(n.handler.AppInstCache.Objs), "num appinsts")
	require.Equal(t, cloudlets, len(n.handler.CloudletCache.Objs), "num cloudlets")
}

func checkCache(t *testing.T, n *node, typ CacheType, count int) {
	n.handler.WaitFor(typ, count)
	require.Equal(t, count, n.handler.GetCache(typ).GetCount(), "node %s count mismatch for %s", n.name, typ.String())
}

type node struct {
	handler    *DummyHandler
	serverMgr  *ServerMgr
	client     *Client
	listenAddr string
	name       string
}

func newNode(name, listenAddr string, connectAddrs []string, typ nodeType) *node {
	n := &node{}
	n.handler = NewDummyHandler()
	n.listenAddr = listenAddr
	n.name = name
	if listenAddr != "" {
		n.serverMgr = &ServerMgr{}
		n.handler.RegisterServer(n.serverMgr)
	}
	if connectAddrs != nil {
		n.client = NewClient(name, connectAddrs, grpc.WithInsecure())
		if typ == crm {
			n.handler.RegisterCRMClient(n.client)
		} else {
			n.handler.RegisterDMEClient(n.client)
		}
	}
	return n
}

func (n *node) startServer() {
	n.serverMgr.Start(n.name, n.listenAddr, nil)
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
