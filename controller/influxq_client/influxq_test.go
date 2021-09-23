package influxq

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/controller/influxq_client/influxq_testutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfluxQ(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelMetrics)
	log.InitTracer(nil)
	defer log.FinishTracer()

	// lower the interval so we don't have to wait so long
	InfluxQPushInterval = 10 * time.Millisecond
	InfluxQReconnectDelay = 10 * time.Millisecond
	ctx := log.StartTestSpan(context.Background())

	p := influxq_testutil.StartInfluxd(t)
	defer p.StopLocal()

	addr := "http://" + p.HttpAddr
	q := NewInfluxQ("metrics", "", "")
	err := q.Start(addr)
	require.Nil(t, err, "new influx q")
	defer q.Stop()

	connected := q.WaitConnected()
	assert.True(t, connected, "connected")
	require.Nil(t, q.WaitCreated())

	// clear test metrics
	_, err = q.QueryDB(`DROP SERIES FROM "test-metric"`)
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
	assert.Equal(t, uint64(0), q.ErrBatch, "batch errors")
	assert.Equal(t, uint64(0), q.ErrPoint, "point errors")
	assert.Equal(t, uint64(0), q.Qfull, "Qfulls")

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
	testAutoProvCounts(t, ctx, q)
	testRetentionPolicyAndContinuousQuery(t, ctx, q, addr)
}

// Test pushing auto prov counts to influxdb and reading back.
func testAutoProvCounts(t *testing.T, ctx context.Context, q *InfluxQ) {
	_, err := q.QueryDB(fmt.Sprintf(`DROP SERIES FROM "%s"`, cloudcommon.AutoProvMeasurement))
	require.Nil(t, err, "clear test metrics")

	ap := edgeproto.AutoProvCount{}
	ap.AppKey.Organization = "dev1"
	ap.AppKey.Name = "app1"
	ap.AppKey.Version = "1.0.0"
	ap.CloudletKey.Organization = "oper1"
	ap.CloudletKey.Name = "cloudlet1"
	ap.Count = 42

	msg := edgeproto.AutoProvCounts{}
	msg.DmeNodeName = "dmeid"
	tsp, err := types.TimestampProto(time.Now())
	require.Nil(t, err, "timestamp proto")
	msg.Timestamp = *tsp
	msg.Counts = []*edgeproto.AutoProvCount{&ap}

	ts, err := types.TimestampFromProto(&msg.Timestamp)
	require.Nil(t, err, "timestamp from proto")

	err = q.PushAutoProvCounts(ctx, &msg)
	require.Nil(t, err, "push auto prov counts")

	res, err := q.QueryDB(fmt.Sprintf(`select * from "%s"`, cloudcommon.AutoProvMeasurement))
	require.Nil(t, err, "select %s", cloudcommon.AutoProvMeasurement)
	require.Equal(t, 1, len(res), "num results")
	require.Equal(t, 1, len(res[0].Series))
	row := res[0].Series[0]
	require.Equal(t, 1, len(row.Values))
	apCheck, dmeid, tsCheck, err := ParseAutoProvCount(row.Columns, row.Values[0])
	require.Nil(t, err, "parse auto prov counts")
	require.Equal(t, msg.DmeNodeName, dmeid, "check dmeid")
	require.Equal(t, ts, tsCheck, "check timestmap")
	require.Equal(t, ap, *apCheck, "check auto prov stats")
}

func testRetentionPolicyAndContinuousQuery(t *testing.T, ctx context.Context, q *InfluxQ, addr string) {
	// Create Downsampled DB
	qd := NewInfluxQ(cloudcommon.DownsampledMetricsDbName, "", "")
	// Start downsample db
	err := qd.Start(addr)
	require.Nil(t, err, "new influx q")
	defer qd.Stop()
	connected := qd.WaitConnected()
	assert.True(t, connected, "connected downsampled db")
	createdErr := qd.WaitCreated()
	require.Nil(t, createdErr)

	// Update default retention policy to downsampled db (this will be used for continuous query fully qualified measurement)
	rpdef := time.Duration(1 * time.Hour)
	err = qd.CreateRetentionPolicy(rpdef, DefaultRetentionPolicy)
	// Check that retention policy has been created
	assert.Nil(t, err)

	// Create a Continuous Query (no retention policy, so will use default)
	cqs := &ContinuousQuerySettings{
		Measurement: "test-metric",
		AggregationFunctions: map[string]string{
			"val1": "sum(\"val1\")",
		},
		CollectionInterval: time.Duration(10 * time.Millisecond),
	}
	err = CreateContinuousQuery(q, qd, cqs)
	assert.Nil(t, err, "create cq")
	time.Sleep(1 * time.Second)
	// Add some more data for continuous query to aggregate
	for ii := 0; ii < 2; ii++ {
		tmst, _ := types.TimestampProto(time.Now())
		metric := edgeproto.Metric{}
		metric.Name = "test-metric"
		metric.Timestamp = *tmst
		metric.Tags = make([]*edgeproto.MetricTag, 0)
		metric.Tags = append(metric.Tags, &edgeproto.MetricTag{
			Name: "tag1",
			Val:  "someval1",
		})
		metric.Tags = append(metric.Tags, &edgeproto.MetricTag{
			Name: "tag2",
			Val:  "someval2",
		})
		metric.Vals = make([]*edgeproto.MetricVal, 0)
		metric.Vals = append(metric.Vals, &edgeproto.MetricVal{
			Name: "val1",
			Value: &edgeproto.MetricVal_Ival{
				Ival: uint64(5),
			},
		})
		q.AddMetric(&metric)
		time.Sleep(time.Microsecond)
	}
	// Check that continuous query has aggregated data
	time.Sleep(2 * time.Second)
	query := fmt.Sprintf("select * from \"test-metric-10ms\"")
	res, err := qd.QueryDB(query)
	require.Nil(t, err, "select *")
	require.True(t, len(res) > 0)
	require.True(t, len(res[0].Series) > 0)
	require.True(t, len(res[0].Series[0].Values) > 0, "num results")

	// Create non-default retention policy to downsampled db (this will be used for continuous query fully qualified measurement)
	rpnondef := time.Duration(2 * time.Hour)
	err = qd.CreateRetentionPolicy(rpnondef, NonDefaultRetentionPolicy)
	// Check that retention policy has been created
	assert.Nil(t, err)

	// Create a Continuous Query on new rp
	cqs = &ContinuousQuerySettings{
		Measurement: "test-metric",
		AggregationFunctions: map[string]string{
			"val2": "sum(\"val2\")",
		},
		CollectionInterval:  time.Duration(5 * time.Millisecond),
		RetentionPolicyTime: 2 * time.Hour,
	}
	err = CreateContinuousQuery(q, qd, cqs)
	assert.Nil(t, err, "create cq")
	time.Sleep(2 * time.Second)
	// Add some more data for new continuous query to aggregate
	for ii := 0; ii < 2; ii++ {
		tmst, _ := types.TimestampProto(time.Now())
		metric := edgeproto.Metric{}
		metric.Name = "test-metric"
		metric.Timestamp = *tmst
		metric.Tags = make([]*edgeproto.MetricTag, 0)
		metric.Tags = append(metric.Tags, &edgeproto.MetricTag{
			Name: "tag1",
			Val:  "someval1",
		})
		metric.Tags = append(metric.Tags, &edgeproto.MetricTag{
			Name: "tag2",
			Val:  "someval2",
		})
		metric.Vals = make([]*edgeproto.MetricVal, 0)
		metric.Vals = append(metric.Vals, &edgeproto.MetricVal{
			Name: "val2",
			Value: &edgeproto.MetricVal_Dval{
				Dval: float64(6.2),
			},
		})
		q.AddMetric(&metric)
		time.Sleep(time.Microsecond)
	}
	// Check that new continuous query has aggregated data
	time.Sleep(3 * time.Second)
	measurementName := CreateInfluxFullyQualifiedMeasurementName(cloudcommon.DownsampledMetricsDbName, "test-metric", 5*time.Millisecond, 2*time.Hour)
	query = fmt.Sprintf("select * from %s", measurementName)
	res, err = qd.QueryDB(query)
	require.Nil(t, err, "select *")
	require.True(t, len(res) > 0)
	require.True(t, len(res[0].Series) > 0)
	require.True(t, len(res[0].Series[0].Values) > 0, "num results")
}
