package main

import (
	"testing"

	distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
)

var AppInstClientKey1 = edgeproto.AppInstClientKey{
	Key: testutil.AppInstData[0].Key,
}

var AppInstClient1 = edgeproto.AppInstClient{
	ClientKey: AppInstClientKey1,
	Status:    edgeproto.AppInstClient_FIND_FOUND,
	Location: distributed_match_engine.Loc{
		Latitude:  1.0,
		Longitude: 1.0,
	},
}
var AppInstClient12 = edgeproto.AppInstClient{
	ClientKey: AppInstClientKey1,
	Status:    edgeproto.AppInstClient_FIND_FOUND,
	Location: distributed_match_engine.Loc{
		Latitude:  1.0,
		Longitude: 2.0,
	},
}
var AppInstClient13 = edgeproto.AppInstClient{
	ClientKey: AppInstClientKey1,
	Status:    edgeproto.AppInstClient_FIND_FOUND,
	Location: distributed_match_engine.Loc{
		Latitude:  1.0,
		Longitude: 3.0,
	},
}

var AppInstClientKey2 = edgeproto.AppInstClientKey{
	Key: testutil.AppInstData[3].Key,
}

var AppInstClient2 = edgeproto.AppInstClient{
	ClientKey: AppInstClientKey2,
	Status:    edgeproto.AppInstClient_FIND_FOUND,
	Location: distributed_match_engine.Loc{
		Latitude:  1.0,
		Longitude: 2.0,
	},
}

func TestClientQ(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelMetrics)
	log.InitTracer("")
	defer log.FinishTracer()
	// Create queue
	q := NewAppInstClientQ()
	assert.NotNil(t, q)
	// Add a channel for appInst1
	ch1 := make(chan edgeproto.AppInstClient, AppInstClientQMaxClients)
	q.SetRecvChan(AppInstClientKey1.Key, ch1)
	// Add Client for AppInst1
	q.AddClient(&AppInstClient1)
	// Check client is received in the channel
	appInstClient := <-ch1
	assert.Equal(t, AppInstClient1, appInstClient)
	// 3. Add a second channel for the different appInst
	ch2 := make(chan edgeproto.AppInstClient, AppInstClientQMaxClients)
	q.SetRecvChan(AppInstClientKey2.Key, ch2)
	// Add a Client for AppInst2
	q.AddClient(&AppInstClient2)
	// Check client is received in the channel 2
	appInstClient = <-ch2
	assert.Equal(t, AppInstClient2, appInstClient)
	// Delete channel 2 - check that list is 0
	count := q.ClearRecvChan(AppInstClientKey2.Key, ch2)
	assert.Equal(t, 0, count)
	// Delete non-existent channel, return is -1
	count = q.ClearRecvChan(AppInstClientKey2.Key, ch2)
	assert.Equal(t, -1, count)
	// Add a second Channel for AppInst1
	ch12 := make(chan edgeproto.AppInstClient, AppInstClientQMaxClients)
	q.SetRecvChan(AppInstClientKey1.Key, ch12)
	// Check that AppInst1 client is received
	appInstClient = <-ch12
	assert.Equal(t, AppInstClient1, appInstClient)
	// Add a client 2 for AppInst1
	q.AddClient(&AppInstClient12)
	// Check that both of the channels recieve the AppInstClient
	appInstClient = <-ch1
	assert.Equal(t, AppInstClient12, appInstClient)
	appInstClient = <-ch12
	assert.Equal(t, AppInstClient12, appInstClient)
	// Delete Channel 1 - verify that count is 1
	count = q.ClearRecvChan(AppInstClientKey1.Key, ch1)
	assert.Equal(t, 1, count)
	// Add client 3 for AppInst1
	q.AddClient(&AppInstClient13)
	// Verify that it's received in channel 2 for appInst1
	appInstClient = <-ch12
	assert.Equal(t, AppInstClient13, appInstClient)
	// Delete channel 2 - verify that count is 0
	count = q.ClearRecvChan(AppInstClientKey1.Key, ch12)
	assert.Equal(t, 0, count)
}
