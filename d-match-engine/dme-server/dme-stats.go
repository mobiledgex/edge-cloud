package main

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	grpcstats "github.com/mobiledgex/edge-cloud/metrics/grpc"
	"github.com/mobiledgex/edge-cloud/util"
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

type ApiStatCall struct {
	key     dmecommon.StatKey
	fail    bool
	latency time.Duration
}

type ApiStat struct {
	reqs    uint64
	errs    uint64
	latency grpcstats.LatencyMetric
	mux     sync.Mutex
	changed bool
}

type MapShard struct {
	apiStatMap map[dmecommon.StatKey]*ApiStat
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
		s.shards[ii].apiStatMap = make(map[dmecommon.StatKey]*ApiStat)
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
	idx := util.GetShardIndex(call.key.Method+call.key.AppKey.Organization+call.key.AppKey.Name, s.numShards)

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
	stat.changed = true
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
					if stat.changed {
						s.send(ctx, ApiStatToMetric(ts, &key, stat))
						stat.changed = false
					}
				}
				s.shards[ii].mux.Unlock()
			}
		case <-s.stop:
			done = true
		}
	}
	s.waitGroup.Done()
}

func ApiStatToMetric(ts *types.Timestamp, key *dmecommon.StatKey, stat *ApiStat) *edgeproto.Metric {
	metric := edgeproto.Metric{}
	metric.Timestamp = *ts
	metric.Name = "dme-api"
	metric.AddTag("apporg", key.AppKey.Organization)
	metric.AddTag("app", key.AppKey.Name)
	metric.AddTag("ver", key.AppKey.Version)
	metric.AddTag("cloudletorg", myCloudletKey.Organization)
	metric.AddTag("cloudlet", myCloudletKey.Name)
	metric.AddTag("id", *scaleID)
	metric.AddTag("method", key.Method)
	metric.AddIntVal("reqs", stat.reqs)
	metric.AddIntVal("errs", stat.errs)
	metric.AddTag("foundCloudlet", key.CloudletFound.Name)
	metric.AddTag("foundOperator", key.CloudletFound.Organization)
	// Cell ID is just a unique number - keep it as a string
	metric.AddTag("cellID", strconv.FormatUint(uint64(key.CellId), 10))
	stat.latency.AddToMetric(&metric)
	return &metric
}

func MetricToStat(metric *edgeproto.Metric) (*dmecommon.StatKey, *ApiStat) {
	key := &dmecommon.StatKey{}
	stat := &ApiStat{}
	for _, tag := range metric.Tags {
		switch tag.Name {
		case "apporg":
			key.AppKey.Organization = tag.Val
		case "app":
			key.AppKey.Name = tag.Val
		case "ver":
			key.AppKey.Version = tag.Val
		case "method":
			key.Method = tag.Val
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

func getCellIdFromDmeReq(req interface{}) uint32 {
	switch typ := req.(type) {
	case *dme.RegisterClientRequest:
		return typ.CellId
	case *dme.FindCloudletRequest:
		return typ.CellId
	case *dme.VerifyLocationRequest:
		return typ.CellId
	case *dme.GetLocationRequest:
		return typ.CellId
	case *dme.AppInstListRequest:
		return typ.CellId
	case *dme.FqdnListRequest:
		return typ.CellId
	case *dme.DynamicLocGroupRequest:
		return typ.CellId
	case *dme.QosPositionRequest:
		return typ.CellId
	}
	return 0
}

func getAppInstClient(appname, appver, apporg string, loc *dme.Loc) *edgeproto.AppInstClient {
	return &edgeproto.AppInstClient{
		ClientKey: edgeproto.AppInstClientKey{
			Key: edgeproto.AppInstKey{
				AppKey: edgeproto.AppKey{
					Organization: apporg,
					Name:         appname,
					Version:      appver,
				},
			},
		},
		Location: *loc,
	}
}

func getResultFromFindCloudletReply(mreq *dme.FindCloudletReply) dme.FindCloudletReply_FindStatus {
	return mreq.Status
}

// Helper function to keep track of the registered devices
func recordDevice(ctx context.Context, req *dme.RegisterClientRequest) {
	devKey := edgeproto.DeviceKey{
		UniqueId:     req.UniqueId,
		UniqueIdType: req.UniqueIdType,
	}
	if platformClientsCache.HasKey(&devKey) {
		return
	}
	ts, err := types.TimestampProto(time.Now())
	if err != nil {
		return
	}
	dev := edgeproto.Device{
		Key:       devKey,
		FirstSeen: ts,
	}
	// Update local cache, which will trigger a send to controller
	platformClientsCache.Update(ctx, &dev, 0)
}

func (s *DmeStats) UnaryStatsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	call := ApiStatCall{}
	ctx = context.WithValue(ctx, dmecommon.StatKeyContextKey, &call.key)

	// call the handler
	resp, err := handler(ctx, req)

	_, call.key.Method = cloudcommon.ParseGrpcMethod(info.FullMethod)

	updateClient := false
	var loc *dme.Loc

	switch typ := req.(type) {
	case *dme.RegisterClientRequest:
		call.key.AppKey.Organization = typ.OrgName
		call.key.AppKey.Name = typ.AppName
		call.key.AppKey.Version = typ.AppVers
		// For platform App clients we need to do accounting of devices
		if err == nil {
			if cloudcommon.IsPlatformApp(typ.OrgName, typ.AppName) ||
				strings.Contains(typ.UniqueIdType, cloudcommon.Organizationplatos) {
				go recordDevice(ctx, typ)
			}
		}

	case *dme.PlatformFindCloudletRequest:
		token := req.(*dme.PlatformFindCloudletRequest).ClientToken
		// cannot collect any stats without a token
		if token != "" {
			tokdata, tokerr := dmecommon.GetClientDataFromToken(token)
			if tokerr != nil {
				return resp, tokerr
			}
			call.key.AppKey = tokdata.AppKey
			loc = &tokdata.Location
			updateClient = true
		}

	case *dme.FindCloudletRequest:

		ckey, ok := dmecommon.CookieFromContext(ctx)
		if !ok {
			return resp, err
		}
		call.key.AppKey.Organization = ckey.OrgName
		call.key.AppKey.Name = ckey.AppName
		call.key.AppKey.Version = ckey.AppVers
		loc = req.(*dme.FindCloudletRequest).GpsLocation
		updateClient = true

	default:
		// All other API calls besides RegisterClient
		// have the app info in the session cookie key.
		ckey, ok := dmecommon.CookieFromContext(ctx)
		if !ok {
			return resp, err
		}
		call.key.AppKey.Organization = ckey.OrgName
		call.key.AppKey.Name = ckey.AppName
		call.key.AppKey.Version = ckey.AppVers
	}
	call.key.CellId = getCellIdFromDmeReq(req)
	if err != nil {
		call.fail = true
	}
	call.latency = time.Since(start)

	if updateClient {
		ckey, ok := dmecommon.CookieFromContext(ctx)
		if !ok {
			return resp, err
		}
		// Update clients cache if we found the cloudlet
		if err == nil && getResultFromFindCloudletReply(resp.(*dme.FindCloudletReply)) == dme.FindCloudletReply_FIND_FOUND {
			client := getAppInstClient(call.key.AppKey.Name, call.key.AppKey.Version, call.key.AppKey.Organization, loc)
			if client != nil {
				client.ClientKey.Key.ClusterInstKey.CloudletKey = call.key.CloudletFound
				client.ClientKey.UniqueId = ckey.UniqueId
				client.ClientKey.UniqueIdType = ckey.UniqueIdType
				// GpsLocation timestamp can carry an arbitrary system time instead of a timestamp
				client.Location.Timestamp = &dme.Timestamp{}
				ts := time.Now()
				client.Location.Timestamp.Seconds = ts.Unix()
				client.Location.Timestamp.Nanos = int32(ts.Nanosecond())
				// Update list of clients on the side and if there is a listener, send it
				go UpdateClientsBuffer(ctx, client)
			}
		}
	}

	s.RecordApiStatCall(&call)

	return resp, err
}
