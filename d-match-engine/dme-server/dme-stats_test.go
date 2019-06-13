package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/stretchr/testify/assert"
)

type testdb struct {
	stats map[StatKey]*ApiStat
	mux   sync.Mutex
}

func (n *testdb) Init() {
	n.stats = make(map[StatKey]*ApiStat)
}

func (n *testdb) send(metric *edgeproto.Metric) bool {
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
			key := StatKey{}
			key.AppKey.DeveloperKey.Name = "dev" + strconv.Itoa(id)
			key.AppKey.Name = "app"
			key.AppKey.Version = "1.0.0"
			key.method = "findLocation"

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
