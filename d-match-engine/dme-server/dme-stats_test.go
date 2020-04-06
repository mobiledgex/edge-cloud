package main

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/stretchr/testify/assert"
)

type testdb struct {
	stats map[dmecommon.StatKey]*ApiStat
	mux   sync.Mutex
}

func (n *testdb) Init() {
	n.stats = make(map[dmecommon.StatKey]*ApiStat)
}

func (n *testdb) send(ctx context.Context, metric *edgeproto.Metric) bool {
	key, stat := MetricToStat(metric)
	n.mux.Lock()
	n.stats[*key] = stat
	n.mux.Unlock()
	return true
}

func TestStatDrops(t *testing.T) {
	db := testdb{}
	db.Init()
	notifyInterval := 20 * time.Millisecond
	numThreads := 100
	stats := NewDmeStats(notifyInterval, 10, db.send)

	stats.Start()
	defer stats.Stop()
	count := uint64(0)
	wg := sync.WaitGroup{}

	for ii := 0; ii < numThreads; ii++ {
		wg.Add(1)
		go func(id int) {
			key := dmecommon.StatKey{}
			key.AppKey.Organization = "dev" + strconv.Itoa(id)
			key.AppKey.Name = "app"
			key.AppKey.Version = "1.0.0"
			key.Method = "findCloudlet"
			key.CloudletFound.Name = "UnitTest"
			key.CloudletFound.Organization = "unittest"
			key.CellId = 1000

			ch := time.After(10 * notifyInterval)
			done := false
			for !done {
				select {
				case <-ch:
					done = true
				default:
					stats.RecordApiStatCall(&ApiStatCall{
						key:     key,
						fail:    rand.Intn(2) == 1,
						latency: time.Duration(rand.Intn(200)) * time.Millisecond,
					})
					atomic.AddUint64(&count, 1)
					time.Sleep(100 * time.Microsecond)
				}
			}
			wg.Done()
		}(ii)
	}
	wg.Wait()

	// sleep one more interval to get the final stats
	time.Sleep(notifyInterval)

	dbCount := uint64(0)
	db.mux.Lock()
	for _, stat := range db.stats {
		dbCount += stat.reqs
	}
	assert.Equal(t, numThreads, len(db.stats), "stat count")
	db.mux.Unlock()

	assert.Equal(t, count, dbCount, "api requests")
	fmt.Printf("served %d requests\n", count)
}

func TestStatChanged(t *testing.T) {
	db := testdb{}
	db.Init()
	notifyInterval := 20 * time.Millisecond
	stats := NewDmeStats(notifyInterval, 1, db.send)
	stats.Start()
	defer stats.Stop()
	var mux = &sync.Mutex{}

	key := dmecommon.StatKey{}
	key.AppKey.Organization = "dev"
	key.AppKey.Name = "app"
	key.AppKey.Version = "1.0.0"
	key.Method = "findCloudlet"
	key.CloudletFound.Name = "UnitTest"
	key.CloudletFound.Organization = "unittest"
	key.CellId = 1000

	mux.Lock()
	stats.RecordApiStatCall(&ApiStatCall{
		key:     key,
		fail:    false,
		latency: 50 * time.Millisecond,
	})
	time.Sleep(100 * time.Microsecond)
	assert.True(t, stats.shards[0].apiStatMap[key].changed)
	mux.Unlock()

	// sleep two intervals to make sure that stats are uploaded to the controller
	time.Sleep(2 * notifyInterval)
	assert.False(t, stats.shards[0].apiStatMap[key].changed)
}

func TestDeviceRecord(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelNotify)
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())
	dmecommon.SetupMatchEngine()

	// test dummy server sending notices to dme
	addr := "127.0.0.1:60001"

	// dummy server side
	serverHandler := notify.NewDummyHandler()
	serverMgr := notify.ServerMgr{}
	serverHandler.RegisterServer(&serverMgr)
	serverMgr.Start(addr, nil)

	// client (dme) side
	client := initNotifyClient(addr, nil)
	client.Start()

	// add a new device - see that it makes it to the server
	for _, reg := range dmetest.DeviceData {
		recordDevice(ctx, &reg)
	}
	// verify the devices were added to the server
	count := len(dmetest.DeviceData) - 1 // Since one is a duplicate
	// verify that devices are in local cache
	assert.Equal(t, count, len(platformClientsCache.Objs))
	serverHandler.WaitForDevices(count)
	assert.Equal(t, count, len(serverHandler.DeviceCache.Objs))
	// Delete all elements from local cache directly
	for _, obj := range platformClientsCache.Objs {
		delete(platformClientsCache.Objs, obj.GetKeyVal())
		delete(platformClientsCache.List, obj.GetKeyVal())
	}
	assert.Equal(t, 0, len(platformClientsCache.Objs))
	assert.Equal(t, count, len(serverHandler.DeviceCache.Objs))
	// Add a single device - make sure count in local cache is updated
	recordDevice(ctx, &dmetest.DeviceData[0])
	assert.Equal(t, 1, len(platformClientsCache.Objs))
	// Make sure that count in the server cache is the same
	assert.Equal(t, count, len(serverHandler.DeviceCache.Objs))
	// Add the same device, check that nothing is updated
	recordDevice(ctx, &dmetest.DeviceData[0])
	assert.Equal(t, 1, len(platformClientsCache.Objs))
	assert.Equal(t, count, len(serverHandler.DeviceCache.Objs))
	serverMgr.Stop()
	client.Stop()
}
