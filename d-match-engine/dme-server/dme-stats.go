package main

import (
	"sync"
	"time"

	"github.com/cespare/xxhash"
	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	grpcstats "github.com/mobiledgex/edge-cloud/metrics/grpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var LatencyTimes = []time.Duration{
	5 * time.Millisecond,
	10 * time.Millisecond,
	25 * time.Millisecond,
	50 * time.Millisecond,
	100 * time.Millisecond,
}

// Stats are collected per App per Cloudlet and per method name (verifylocation, etc).
type StatKey struct {
	AppKey        edgeproto.AppKey
	CloudletFound edgeproto.CloudletKey
	method        string
}

type ApiStatCall struct {
	key     StatKey
	fail    bool
	latency time.Duration
}

type ApiStat struct {
	reqs    uint64
	errs    uint64
	latency grpcstats.LatencyMetric
	mux     sync.Mutex
}

type MapShard struct {
	apiStatMap map[StatKey]*ApiStat
	notify     bool
	mux        sync.Mutex
}

type DmeStats struct {
	shards    []MapShard
	numShards uint
	mux       sync.Mutex
	interval  time.Duration
	send      func(ctx context.Context, metric *edgeproto.Metric) bool
	waitGroup sync.WaitGroup
	stop      chan struct{}
}

func NewDmeStats(interval time.Duration, numShards uint, send func(ctx context.Context, metric *edgeproto.Metric) bool) *DmeStats {
	s := DmeStats{}
	s.shards = make([]MapShard, numShards, numShards)
	s.numShards = numShards
	for ii, _ := range s.shards {
		s.shards[ii].apiStatMap = make(map[StatKey]*ApiStat)
	}
	s.interval = interval
	s.send = send
	return &s
}

func (s *DmeStats) Start() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.stop = make(chan struct{})
	s.waitGroup.Add(1)
	go s.RunNotify()
}

func (s *DmeStats) Stop() {
	close(s.stop)
	s.waitGroup.Wait()
}

func (s *DmeStats) RecordApiStatCall(call *ApiStatCall) {
	hash := xxhash.Sum64([]byte(call.key.method + call.key.AppKey.DeveloperKey.Name + call.key.AppKey.Name))
	idx := hash % uint64(s.numShards)

	shard := &s.shards[idx]
	shard.mux.Lock()
	stat, found := shard.apiStatMap[call.key]
	if !found {
		stat = &ApiStat{}
		grpcstats.InitLatencyMetric(&stat.latency, LatencyTimes)
		shard.apiStatMap[call.key] = stat
	}
	stat.reqs++
	if call.fail {
		stat.errs++
	}
	stat.latency.AddLatency(call.latency)
	shard.mux.Unlock()
}

// RunNotify walks the stats periodically, and uploads the current
// stats to the controller.
func (s *DmeStats) RunNotify() {
	done := false
	// for now, no tracing of stats
	ctx := context.Background()
	for !done {
		select {
		case <-time.After(s.interval):
			ts, _ := types.TimestampProto(time.Now())
			for ii, _ := range s.shards {
				s.shards[ii].mux.Lock()
				for key, stat := range s.shards[ii].apiStatMap {
					s.send(ctx, ApiStatToMetric(ts, &key, stat))
				}
				s.shards[ii].mux.Unlock()
			}
		case <-s.stop:
			done = true
		}
	}
	s.waitGroup.Done()
}

func ApiStatToMetric(ts *types.Timestamp, key *StatKey, stat *ApiStat) *edgeproto.Metric {
	metric := edgeproto.Metric{}
	metric.Timestamp = *ts
	metric.Name = "dme-api"
	metric.AddTag("dev", key.AppKey.DeveloperKey.Name)
	metric.AddTag("app", key.AppKey.Name)
	metric.AddTag("ver", key.AppKey.Version)
	metric.AddTag("oper", myCloudletKey.OperatorKey.Name)
	metric.AddTag("cloudlet", myCloudletKey.Name)
	metric.AddTag("id", *scaleID)
	metric.AddTag("method", key.method)
	metric.AddIntVal("reqs", stat.reqs)
	metric.AddIntVal("errs", stat.errs)
	metric.AddTag("foundCloudlet", key.CloudletFound.Name)
	metric.AddTag("foundOperator", key.CloudletFound.OperatorKey.Name)
	stat.latency.AddToMetric(&metric)
	return &metric
}

func MetricToStat(metric *edgeproto.Metric) (*StatKey, *ApiStat) {
	key := &StatKey{}
	stat := &ApiStat{}
	for _, tag := range metric.Tags {
		switch tag.Name {
		case "dev":
			key.AppKey.DeveloperKey.Name = tag.Val
		case "app":
			key.AppKey.Name = tag.Val
		case "ver":
			key.AppKey.Version = tag.Val
		case "method":
			key.method = tag.Val
		}
	}
	for _, val := range metric.Vals {
		switch val.Name {
		case "reqs":
			stat.reqs = val.GetIval()
		case "errs":
			stat.errs = val.GetIval()
		}
	}
	stat.latency.FromMetric(metric)
	return key, stat
}

func (s *DmeStats) UnaryStatsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	// call the handler
	resp, err := handler(ctx, req)

	call := ApiStatCall{}
	_, call.key.method = cloudcommon.ParseGrpcMethod(info.FullMethod)

	switch typ := req.(type) {
	case *dme.RegisterClientRequest:
		call.key.AppKey.DeveloperKey.Name = typ.DevName
		call.key.AppKey.Name = typ.AppName
		call.key.AppKey.Version = typ.AppVers
	default:
		// All other API calls besides RegisterClient
		// have the app info in the session cookie key.
		ckey, ok := dmecommon.CookieFromContext(ctx)
		if !ok {
			return resp, err
		}
		call.key.AppKey.DeveloperKey.Name = ckey.DevName
		call.key.AppKey.Name = ckey.AppName
		call.key.AppKey.Version = ckey.AppVers
	}

	// For the FindCloudlet api we should record what we returned
	switch typ := resp.(type) {
	case *dme.FindCloudletReply:
		call.key.CloudletFound.Name = typ.CloudletName
		call.key.CloudletFound.OperatorKey.Name = typ.CarrierName
	default:
		break
	}

	if err != nil {
		call.fail = true
	}
	call.latency = time.Since(start)
	s.RecordApiStatCall(&call)

	return resp, err
}
