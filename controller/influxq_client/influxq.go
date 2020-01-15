package influxq

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/mobiledgex/edge-cloud/cloudcommon/influxsup"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// Each write to the Influx DB is an HTTP Post method.
// To avoid the overhead of opening a connection for every data point,
// Influx DB provides a way to batch a bunch of "points" (metrics)
// at once. To utilize this, the controller buffers incoming data
// and writes it after a certain period of time (interval) or number of
// buffered metrics (count trigger).

var InfluxQPushInterval time.Duration = time.Second
var InfluxQPushCountTrigger = 50
var InfluxQPushCountMax = 5000
var InfluxQPrecision = "us"
var InfluxQReconnectDelay time.Duration = time.Second

type InfluxQ struct {
	dbName    string
	user      string
	password  string
	client    client.Client
	data      []*edgeproto.Metric
	done      bool
	dbcreated bool
	doPush    chan bool
	mux       sync.Mutex
	wg        sync.WaitGroup
	ErrBatch  uint64
	ErrPoint  uint64
	Qfull     uint64
}

func NewInfluxQ(DBName, username, password string) *InfluxQ {
	q := InfluxQ{}
	q.dbName = DBName
	q.data = make([]*edgeproto.Metric, 0)
	q.doPush = make(chan bool, 1)
	q.user = username
	q.password = password
	return &q
}

func (q *InfluxQ) Start(addr string) error {
	cl, err := influxsup.GetClient(addr, q.user, q.password)
	if err != nil {
		return err
	}
	q.mux.Lock()
	defer q.mux.Unlock()
	q.client = cl
	q.done = false
	q.wg.Add(1)
	go q.RunPush()
	return nil
}

func (q *InfluxQ) Stop() {
	q.done = true
	q.DoPush() // wake up thread
	q.wg.Wait()
	q.client.Close()
}

func (q *InfluxQ) RunPush() {
	for !q.done {
		if !q.dbcreated {
			// make sure main db is created otherwise
			// batch point writes will fail
			_, err := q.QueryDB(fmt.Sprintf("create database %s", q.dbName))
			if err != nil {
				if _, ok := err.(net.Error); !ok {
					// only log for non-network errors
					log.DebugLog(log.DebugLevelMetrics,
						"create database", "err", err)
				}
				time.Sleep(InfluxQReconnectDelay)
				continue
			}
			q.dbcreated = true
			fmt.Printf("asdf database created: %+v", q)
		}

		select {
		case <-q.doPush:
		case <-time.After(InfluxQPushInterval):
		}
		if q.done {
			break
		}
		q.mux.Lock()
		if len(q.data) == 0 {
			q.mux.Unlock()
			continue
		}
		data := q.data
		q.data = make([]*edgeproto.Metric, 0)
		q.mux.Unlock()

		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  q.dbName,
			Precision: InfluxQPrecision,
		})
		if err != nil {
			log.DebugLog(log.DebugLevelMetrics, "create batch points",
				"err", err)
			atomic.AddUint64(&q.ErrBatch, 1)
			continue
		}

		for _, metric := range data {
			tags := make(map[string]string)
			for _, mtag := range metric.Tags {
				tags[mtag.Name] = mtag.Val
			}
			fields := make(map[string]interface{})
			for _, mval := range metric.Vals {
				switch mval.Value.(type) {
				case *edgeproto.MetricVal_Dval:
					fields[mval.Name] = mval.GetDval()
				case *edgeproto.MetricVal_Ival:
					fields[mval.Name] = int64(mval.GetIval())
				case *edgeproto.MetricVal_Bval:
					fields[mval.Name] = mval.GetBval()
				case *edgeproto.MetricVal_Sval:
					fields[mval.Name] = mval.GetSval()
				}
			}
			ts, err := types.TimestampFromProto(&metric.Timestamp)
			if err != nil {
				log.DebugLog(log.DebugLevelMetrics, "set timestamp",
					"timestamp", &metric.Timestamp, "err", err)
				atomic.AddUint64(&q.ErrPoint, 1)
				continue
			}
			pt, err := client.NewPoint(metric.Name, tags, fields, ts)
			if err != nil {
				log.DebugLog(log.DebugLevelMetrics,
					"metric new point", "err", err)
				atomic.AddUint64(&q.ErrPoint, 1)
				continue
			}
			bp.AddPoint(pt)
		}
		err = q.client.Write(bp)
		if err != nil {
			log.DebugLog(log.DebugLevelMetrics, "write batch points",
				"err", err)
			atomic.AddUint64(&q.ErrBatch, 1)
		}
	}
	q.wg.Done()
}

func (q *InfluxQ) Recv(ctx context.Context, metric *edgeproto.Metric) {
	q.AddMetric(metric)
}

func (q *InfluxQ) AddMetric(metrics ...*edgeproto.Metric) {
	q.mux.Lock()
	defer q.mux.Unlock()
	if len(q.data) > InfluxQPushCountMax {
		// limit len to prevent out of memory if
		// q is not reachable
		q.Qfull++
		return
	}
	for ii, _ := range metrics {
		q.data = append(q.data, metrics[ii])
	}
	if len(q.data) > InfluxQPushCountTrigger {
		q.DoPush()
	}
}

func (q *InfluxQ) DoPush() {
	select {
	case q.doPush <- true:
	default:
		// already triggered
	}
}

func (q *InfluxQ) QueryDB(cmd string) ([]client.Result, error) {
	query := client.Query{
		Command:  cmd,
		Database: q.dbName,
	}
	resp, err := q.client.Query(query)
	if err != nil {
		return nil, err
	}
	if resp.Error() != nil {
		return nil, resp.Error()
	}
	return resp.Results, nil
}

func (q *InfluxQ) WaitConnected() bool {
	// wait till db online
	for ii := 0; ii < 200; ii++ {
		if _, _, err := q.client.Ping(0); err == nil {
			return true
		}
		time.Sleep(25 * time.Millisecond)
	}
	return false
}
