package main

import (
	"testing"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestAddClients(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	dmecommon.Settings.MaxTrackedDmeClients = 2

	InitAppInstClients()

	UpdateClientsBuffer(ctx, &dmetest.AppInstClientData[0])
	// check that this client is added correctly
	list, found := clientsMap.clientsByApp[dmetest.AppInstClientData[0].ClientKey.AppInstKey.AppKey]
	require.True(t, found)
	require.Equal(t, 1, len(list))
	require.Equal(t, dmetest.AppInstClientData[0], list[0])

	UpdateClientsBuffer(ctx, &dmetest.AppInstClientData[1])
	// check that this client is added correctly
	list, found = clientsMap.clientsByApp[dmetest.AppInstClientData[1].ClientKey.AppInstKey.AppKey]
	require.True(t, found)
	require.Equal(t, 1, len(list))
	require.Equal(t, dmetest.AppInstClientData[1], list[0])

	UpdateClientsBuffer(ctx, &dmetest.AppInstClientData[2])
	// check that this client is added correctly and replaced the original one that was there
	list, found = clientsMap.clientsByApp[dmetest.AppInstClientData[0].ClientKey.AppInstKey.AppKey]
	require.True(t, found)
	require.Equal(t, 1, len(list))
	require.Equal(t, dmetest.AppInstClientData[2], list[0])

	// Add couple other clients to trigger the eviction of the first client
	UpdateClientsBuffer(ctx, &dmetest.AppInstClientData[3])
	UpdateClientsBuffer(ctx, &dmetest.AppInstClientData[4])
	// check that this client is added correctly and replaced the original one that was there
	list, found = clientsMap.clientsByApp[dmetest.AppInstClientData[0].ClientKey.AppInstKey.AppKey]
	require.True(t, found)
	require.Equal(t, 2, len(list))
	for _, c := range list {
		require.NotEqual(t, c, dmetest.AppInstClientData[2])
	}

	require.Equal(t, 2, len(clientsMap.clientsByApp))

	// test deletion of AppInstances
	PurgeAppInstClients(ctx, &dmetest.AppInstClientData[1].ClientKey.AppInstKey)
	list, found = clientsMap.clientsByApp[dmetest.AppInstClientData[1].ClientKey.AppInstKey.AppKey]
	require.True(t, found)
	require.Equal(t, 0, len(list))

	PurgeAppInstClients(ctx, &dmetest.AppInstClientData[0].ClientKey.AppInstKey)
	list, found = clientsMap.clientsByApp[dmetest.AppInstClientData[0].ClientKey.AppInstKey.AppKey]
	require.True(t, found)
	require.Equal(t, 0, len(list))
}
