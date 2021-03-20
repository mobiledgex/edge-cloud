package dmecommon

import (
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestAddClients(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelDmereq)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	Settings.MaxTrackedDmeClients = 2

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

	// test timeout of the appInstances
	tsOld := cloudcommon.TimeToTimestamp(time.Now().Add(-1 * time.Minute))
	data := dmetest.AppInstClientData[3]
	data.Location.Timestamp = &tsOld
	UpdateClientsBuffer(ctx, &data)
	tsFuture := cloudcommon.TimeToTimestamp(time.Now().Add(1 * time.Minute))
	data = dmetest.AppInstClientData[4]
	data.Location.Timestamp = &tsFuture
	UpdateClientsBuffer(ctx, &data)
	list, found = clientsMap.clientsByApp[dmetest.AppInstClientData[4].ClientKey.AppInstKey.AppKey]
	require.True(t, found)
	require.Equal(t, 2, len(list))
	// set quick timeout
	Settings.AppinstClientCleanupInterval = edgeproto.Duration(1 * time.Second)
	clientsMap.UpdateClientTimeout(Settings.AppinstClientCleanupInterval)
	// give the thread a bit of time to run
	time.Sleep(2 * time.Second)
	// Check to see that one of them got deleted
	list, found = clientsMap.clientsByApp[dmetest.AppInstClientData[4].ClientKey.AppInstKey.AppKey]
	require.True(t, found)
	require.Equal(t, 1, len(list))
	require.Equal(t, data, list[0])
}
