package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfluxQ(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelMetrics)
	addr := "127.0.0.1:8086"
	// lower the interval so we don't have to wait so long
	InfluxQPushInterval = 10 * time.Millisecond
	InfluxQReconnectDelay = 10 * time.Millisecond

	// start influxd if not already running
	_, err := exec.Command("sh", "-c", "pgrep -x influxd").Output()
	if err != nil {
		db := NewInfluxDBServer()
		err := db.Start(addr)
		require.Nil(t, err, "start InfluxDB server")
		defer db.Stop()
	}

	q := NewInfluxQ()
	err = q.Start(addr)
	require.Nil(t, err, "new influx q")
	defer q.Stop()

	connected := q.WaitConnected()
	assert.True(t, connected, "connected")

	// clear test metrics
	_, err = q.QueryDB(`DROP SERIES FROM "test-metric"`)
	if err != nil && strings.Contains(err.Error(), ".wal: file already closed") {
		// seems like a bug with influxdb, restart it
		fmt.Println("Restarting influx due to wal file already closed error")
		exec.Command("sh", "-c", "pkill -x influxd").Output()
		db := NewInfluxDBServer()
		err = db.Start(addr)
		require.Nil(t, err, "start InfluxDB server")
		defer db.Stop()
	}
	require.Nil(t, err, "clear test metrics")

	count := 0
	iilimit := 10

	for tt := 0; tt < 10; tt++ {
		ts, _ := types.TimestampProto(time.Now())
		for ii := 0; ii < iilimit; ii++ {
			metric := edgeproto.Metric{}
			metric.Name = "test-metric"
			metric.Timestamp = *ts
			metric.Tags = make([]*edgeproto.MetricTag, 0)
			metric.Tags = append(metric.Tags, &edgeproto.MetricTag{
				Name: "tag1",
				Val:  "someval" + strconv.Itoa(ii),
			})
			metric.Tags = append(metric.Tags, &edgeproto.MetricTag{
				Name: "tag2",
				Val:  "someval",
			})
			metric.Vals = make([]*edgeproto.MetricVal, 0)
			metric.Vals = append(metric.Vals, &edgeproto.MetricVal{
				Name: "val1",
				Value: &edgeproto.MetricVal_Ival{
					Ival: uint64(ii + tt*iilimit),
				},
			})
			metric.Vals = append(metric.Vals, &edgeproto.MetricVal{
				Name: "val2",
				Value: &edgeproto.MetricVal_Dval{
					Dval: float64(ii+tt*iilimit) / 2.0,
				},
			})
			q.AddMetric(&metric)
			time.Sleep(time.Microsecond)
			count++
		}
	}

	// wait for metrics to get pushed to db
	time.Sleep(2 * InfluxQPushInterval)
	assert.Equal(t, uint64(0), q.errBatch, "batch errors")
	assert.Equal(t, uint64(0), q.errPoint, "point errors")
	assert.Equal(t, uint64(0), q.qfull, "qfulls")

	// wait for records to get updated in database and become queryable.
	num := 0
	for ii := 0; ii < 10; ii++ {
		res, err := q.QueryDB("select count(val1) from \"test-metric\"")
		if err == nil && len(res) > 0 && len(res[0].Series) > 0 && len(res[0].Series[0].Values) > 0 {
			jnum, ok := res[0].Series[0].Values[0][1].(json.Number)
			if ok {
				val, err := jnum.Int64()
				if err == nil && int(val) == count {
					num = count
					break
				}
			}

		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.Equal(t, count, num, "num unique values")

	// query all test-metrics
	res, err := q.QueryDB("select * from \"test-metric\"")
	assert.Nil(t, err, "select *")
	assert.Equal(t, 1, len(res), "num results")
	if len(res) > 0 {
		assert.Equal(t, 1, len(res[0].Series), "num series")
		if len(res[0].Series) > 0 {
			assert.Equal(t, count, len(res[0].Series[0].Values), "num values")
			// prints results if needed
			if false {
				for ii, _ := range res[0].Series[0].Values {
					fmt.Printf("%d: %v\n", ii, res[0].Series[0].Values[ii])
				}
			}
		}
	}
}
