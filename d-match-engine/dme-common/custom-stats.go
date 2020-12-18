package dmecommon

import (
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	grpcstats "github.com/mobiledgex/edge-cloud/metrics/grpc"
)

// Filled in by DME. Added to EdgeEventStatCall to update stats
type CustomStatInfo struct {
	Name             string
	Samples          []*dme.Sample
	SessionCookieKey *CookieKey // SessionCookie to identify unique clients for EdgeEvents
}

type CustomStat struct {
	Count             uint64
	RollingStatistics *grpcstats.RollingStatistics
}

type CustomStats struct {
	Stats     map[string]*CustomStat
	StartTime time.Time
}

func NewCustomStat() *CustomStat {
	c := new(CustomStat)
	c.RollingStatistics = grpcstats.NewRollingStatistics()
	return c
}

func NewCustomStats() *CustomStats {
	c := new(CustomStats)
	c.Stats = make(map[string]*CustomStat)
	c.StartTime = time.Now()
	return c
}

func RecordCustomStatCall(appInstKey *edgeproto.AppInstKey, sessionCookieKey *CookieKey, eventName string, samples []*dme.Sample) {
	if EEStats == nil {
		return
	}
	call := EdgeEventStatCall{}
	call.Key.AppInstKey = *appInstKey
	call.Key.Metric = cloudcommon.CustomMetric // override method name
	call.CustomStatInfo = &CustomStatInfo{
		Name:             eventName,
		Samples:          samples,
		SessionCookieKey: sessionCookieKey,
	}
	EEStats.RecordEdgeEventStatCall(&call)
}

func (c *CustomStats) Update(info *CustomStatInfo) {
	stat, ok := c.Stats[info.Name]
	if !ok {
		stat = NewCustomStat()
	}
	stat.Count++
	statistics := grpcstats.CalculateStatistics(info.Samples)
	stat.RollingStatistics.UpdateRollingStatistics(info.SessionCookieKey.UniqueId, statistics.Avg)
	c.Stats[info.Name] = stat
}
