package dmecommon

import (
	"flag"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmeutil "github.com/mobiledgex/edge-cloud/d-match-engine/dme-util"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	grpcstats "github.com/mobiledgex/edge-cloud/metrics/grpc"
	"github.com/mobiledgex/edge-cloud/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// MyCloudlet is the information for the cloudlet in which the DME is instantiated.
// The key for MyCloudlet is provided as a configuration - either command line or
// from a file.
var MyCloudletKey edgeproto.CloudletKey

var PlatformClientsCache edgeproto.DeviceCache

var ScaleID = flag.String("scaleID", "", "ID to distinguish multiple DMEs in the same cloudlet. Defaults to hostname if unspecified.")
var monitorUuidType = flag.String("monitorUuidType", "MobiledgeXMonitorProbe", "AppInstClient UUID Type used for monitoring purposes")

const EdgeEventLatencyMethod = "appinst-latency"

var Stats *DmeStats

var LatencyTimes = []time.Duration{
	5 * time.Millisecond,
	10 * time.Millisecond,
	25 * time.Millisecond,
	50 * time.Millisecond,
	100 * time.Millisecond,
}

type ApiStatCall struct {
	Key           StatKey
	Fail          bool
	Latency       time.Duration
	Samples       []float64 // Latency samples for EdgeEvents
	SessionCookie string    // SessionCookie to identify unique clients for EdgeEvents
}

type ApiStat struct {
	Reqs                uint64
	Errs                uint64
	Latency             grpcstats.LatencyMetric
	RollingLatencyTemp  *dmeutil.RollingLatency // Temporary rolling statistics for EdgeEvents latency measurements (resets after 10 min)
	RollingLatencyTotal *dmeutil.RollingLatency // Rolling statistics for EdgeEvents latency measurements to be stored in influx
	Mux                 sync.Mutex
	Changed             bool
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

func (s *DmeStats) LookupApiStatCall(call *ApiStatCall) (*ApiStat, bool) {
	idx := util.GetShardIndex(call.Key.Method+call.Key.AppKey.Organization+call.Key.AppKey.Name, s.numShards)

	shard := &s.shards[idx]
	shard.mux.Lock()
	defer shard.mux.Unlock()
	stat, found := shard.apiStatMap[call.Key]
	return stat, found
}

func (s *DmeStats) RecordApiStatCall(call *ApiStatCall) {
	idx := util.GetShardIndex(call.Key.Method+call.Key.AppKey.Organization+call.Key.AppKey.Name, s.numShards)

	shard := &s.shards[idx]
	shard.mux.Lock()
	stat, found := shard.apiStatMap[call.Key]
	if !found {
		stat = &ApiStat{}
		grpcstats.InitLatencyMetric(&stat.Latency, LatencyTimes)
		shard.apiStatMap[call.Key] = stat
	}
	stat.Reqs++
	if call.Fail {
		stat.Errs++
	}
	stat.Latency.AddLatency(call.Latency)
	if call.Key.Method == EdgeEventLatencyMethod {
		// Update RollingLatency and RollingLatencyTotal statistics
		if stat.RollingLatencyTemp == nil {
			stat.RollingLatencyTemp = dmeutil.NewRollingLatency()
		}
		if stat.RollingLatencyTotal == nil {
			stat.RollingLatencyTotal = dmeutil.NewRollingLatency()
		}
		stat.RollingLatencyTemp.UpdateRollingLatency(call.Samples, call.SessionCookie)
		stat.RollingLatencyTotal.UpdateRollingLatency(call.Samples, call.SessionCookie)
	}
	stat.Changed = true
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
					if stat.Changed {
						if key.Method == EdgeEventLatencyMethod {
							s.send(ctx, EdgeEventStatToMetric(ts, &key, stat))
						} else {
							s.send(ctx, ApiStatToMetric(ts, &key, stat))
						}
						stat.Changed = false
					}
					// Reset RollingLatencyTemp every 10 minutes, so that latency values are current
					if key.Method == EdgeEventLatencyMethod {
						if stat.RollingLatencyTemp == nil || stat.RollingLatencyTemp.NumUniqueClients == 0 {
							continue
						}
						t := cloudcommon.TimestampToTime(*stat.RollingLatencyTemp.Latency.Timestamp)
						if time.Since(t) > time.Minute*10 {
							stat.RollingLatencyTemp = dmeutil.NewRollingLatency()
						}

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

func ApiStatToMetric(ts *types.Timestamp, key *StatKey, stat *ApiStat) *edgeproto.Metric {
	metric := edgeproto.Metric{}
	metric.Timestamp = *ts
	metric.Name = "dme-api"
	metric.AddTag("apporg", key.AppKey.Organization)
	metric.AddTag("app", key.AppKey.Name)
	metric.AddTag("ver", key.AppKey.Version)
	metric.AddTag("cloudletorg", MyCloudletKey.Organization)
	metric.AddTag("cloudlet", MyCloudletKey.Name)
	metric.AddTag("id", *ScaleID)
	metric.AddTag("method", key.Method)
	metric.AddIntVal("reqs", stat.Reqs)
	metric.AddIntVal("errs", stat.Errs)
	metric.AddTag("foundCloudlet", key.CloudletFound.Name)
	metric.AddTag("foundOperator", key.CloudletFound.Organization)
	// Cell ID is just a unique number - keep it as a string
	metric.AddTag("cellID", strconv.FormatUint(uint64(key.CellId), 10))
	stat.Latency.AddToMetric(&metric)
	return &metric
}

func EdgeEventStatToMetric(ts *types.Timestamp, key *StatKey, stat *ApiStat) *edgeproto.Metric {
	metric := edgeproto.Metric{}
	metric.Timestamp = *ts
	metric.Name = EdgeEventLatencyMethod
	metric.AddTag("dmecloudlet", MyCloudletKey.Name)
	metric.AddTag("dmecloudletorg", MyCloudletKey.Organization)
	metric.AddIntVal("reqs", stat.Reqs)
	metric.AddIntVal("errs", stat.Errs)
	// AppInst information
	metric.AddTag("app", key.AppKey.Name)
	metric.AddTag("apporg", key.AppKey.Organization)
	metric.AddTag("ver", key.AppKey.Version)
	metric.AddTag("cloudlet", key.CloudletFound.Name)
	metric.AddTag("cloudletorg", key.CloudletFound.Organization)
	metric.AddTag("cluster", key.ClusterKey.Name)
	metric.AddTag("clusterorg", key.ClusterInstOrg)
	// Latency information (store RollingLatencyTotal in influx)
	metric.AddIntVal("numsamples", stat.RollingLatencyTotal.Latency.NumSamples)
	metric.AddIntVal("numclients", stat.RollingLatencyTotal.NumUniqueClients)
	metric.AddDoubleVal("avg", stat.RollingLatencyTotal.Latency.Avg)
	metric.AddDoubleVal("stddev", stat.RollingLatencyTotal.Latency.StdDev)
	metric.AddDoubleVal("min", stat.RollingLatencyTotal.Latency.Min)
	metric.AddDoubleVal("max", stat.RollingLatencyTotal.Latency.Max)

	stat.Latency.AddToMetric(&metric)
	return &metric
}

func MetricToStat(metric *edgeproto.Metric) (*StatKey, *ApiStat) {
	key := &StatKey{}
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
		case "clusterinstorg":
			key.ClusterInstOrg = tag.Val
		case "clustername":
			key.ClusterKey.Name = tag.Val
		}
	}
	for _, val := range metric.Vals {
		switch val.Name {
		case "reqs":
			stat.Reqs = val.GetIval()
		case "errs":
			stat.Errs = val.GetIval()
		}
	}
	stat.Latency.FromMetric(metric)
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
func RecordDevice(ctx context.Context, req *dme.RegisterClientRequest) {
	devKey := edgeproto.DeviceKey{
		UniqueId:     req.UniqueId,
		UniqueIdType: req.UniqueIdType,
	}
	if PlatformClientsCache.HasKey(&devKey) {
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
	PlatformClientsCache.Update(ctx, &dev, 0)
}

func (s *DmeStats) UnaryStatsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	call := ApiStatCall{}
	ctx = context.WithValue(ctx, StatKeyContextKey, &call.Key)

	// call the handler
	resp, err := handler(ctx, req)

	_, call.Key.Method = cloudcommon.ParseGrpcMethod(info.FullMethod)

	updateClient := false
	var loc *dme.Loc

	switch typ := req.(type) {
	case *dme.RegisterClientRequest:
		call.Key.AppKey.Organization = typ.OrgName
		call.Key.AppKey.Name = typ.AppName
		call.Key.AppKey.Version = typ.AppVers
		// For platform App clients we need to do accounting of devices
		if err == nil {
			if cloudcommon.IsPlatformApp(typ.OrgName, typ.AppName) ||
				strings.Contains(strings.ToLower(typ.UniqueIdType), strings.ToLower(cloudcommon.OrganizationSamsung)) {
				go RecordDevice(ctx, typ)
			}
		}

	case *dme.PlatformFindCloudletRequest:
		token := req.(*dme.PlatformFindCloudletRequest).ClientToken
		// cannot collect any stats without a token
		if token != "" {
			tokdata, tokerr := GetClientDataFromToken(token)
			if tokerr != nil {
				return resp, tokerr
			}
			call.Key.AppKey = tokdata.AppKey
			loc = &tokdata.Location
			updateClient = true
		}

	case *dme.FindCloudletRequest:

		ckey, ok := CookieFromContext(ctx)
		if !ok {
			return resp, err
		}
		call.Key.AppKey.Organization = ckey.OrgName
		call.Key.AppKey.Name = ckey.AppName
		call.Key.AppKey.Version = ckey.AppVers
		loc = req.(*dme.FindCloudletRequest).GpsLocation
		updateClient = true

	default:
		// All other API calls besides RegisterClient
		// have the app info in the session cookie key.
		ckey, ok := CookieFromContext(ctx)
		if !ok {
			return resp, err
		}
		call.Key.AppKey.Organization = ckey.OrgName
		call.Key.AppKey.Name = ckey.AppName
		call.Key.AppKey.Version = ckey.AppVers
	}
	call.Key.CellId = getCellIdFromDmeReq(req)
	if err != nil {
		call.Fail = true
	}
	call.Latency = time.Since(start)

	if updateClient {
		ckey, ok := CookieFromContext(ctx)
		if !ok {
			return resp, err
		}
		// skip platform monitoring FindCloudletCalls, or if we didn't find the cloudlet
		createClient := true
		if err != nil ||
			ckey.UniqueIdType == *monitorUuidType ||
			getResultFromFindCloudletReply(resp.(*dme.FindCloudletReply)) != dme.FindCloudletReply_FIND_FOUND {
			createClient = false
		}

		// Update clients cache if we found the cloudlet
		if createClient {
			client := getAppInstClient(call.Key.AppKey.Name, call.Key.AppKey.Version, call.Key.AppKey.Organization, loc)
			if client != nil {
				client.ClientKey.Key.ClusterInstKey.CloudletKey = call.Key.CloudletFound
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
