package edgeproto

import (
	fmt "fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNotifyOrder(t *testing.T) {
	no := NewNotifyOrder()

	// this is just for debug
	nodes := make([]*NotifyOrderNode, 0)
	for _, node := range no.objs {
		nodes = append(nodes, node)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].order < nodes[j].order
	})
	for _, node := range nodes {
		fmt.Printf("%d: %s\n", node.order, node.typeName)
	}

	require.True(t, no.Less("Flavor", "Cloudlet"))
	require.True(t, no.Less("App", "AppInst"))
	require.True(t, no.Less("App", "TrustPolicyException"))
	require.True(t, no.Less("Flavor", "AppInstRefs"))
	require.True(t, no.Less("App", "AppInstRefs"))
	require.True(t, no.Less("AppInst", "AppInstRefs"))
	require.True(t, no.Less("ClusterInst", "AppInstRefs"))
	require.False(t, no.Less("AppInst", "App"))
	require.False(t, no.Less("CloudletPool", "Flavor"))
	// nodes that don't exist should have order 0
	require.False(t, no.Less("foo", "Flavor"))
	require.True(t, no.Less("foo", "Cloudlet"))
}
