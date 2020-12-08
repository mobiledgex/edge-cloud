package dmecommon

import (
	"time"

	grpcstats "github.com/mobiledgex/edge-cloud/metrics/grpc"
)

// Filled in by DME. Added to EdgeEventStatCall to update stats
type CustomStatInfo struct {
	Name             string
	Value            float64
	SessionCookieKey CookieKey // SessionCookie to identify unique clients for EdgeEvents
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

func (c *CustomStats) Update(info *CustomStatInfo) {
	stat, ok := c.Stats[info.Name]
	if !ok {
		stat = NewCustomStat()
	}
	stat.Count++
	stat.RollingStatistics.UpdateRollingStatistics(info.SessionCookieKey.UniqueId, info.Value)
}
