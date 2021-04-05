package dmecommon

import (
	"sync"
	"time"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	grpcstats "github.com/mobiledgex/edge-cloud/metrics/grpc"
)

// Filled in by DME. Added to EdgeEventStatCall to update stats
type LatencyStatInfo struct {
	Samples []*dme.Sample
}

func GetLatencyStatKey(appInstKey edgeproto.AppInstKey, deviceInfo *dme.DeviceInfo, carrier string, loc *dme.Loc, tileLength int) LatencyStatKey {
	return LatencyStatKey{
		AppInstKey:      appInstKey,
		DataNetworkType: deviceInfo.DataNetworkType,
		DeviceOs:        deviceInfo.DeviceOs,
		DeviceModel:     deviceInfo.DeviceModel,
		SignalStrength:  uint64(deviceInfo.SignalStrength),
		DeviceCarrier:   carrier,
		LocationTile:    GetLocationTileFromGpsLocation(loc, tileLength),
	}
}

// Used to find corresponding LatencyStat
// Created from LatencyInfo fields
type LatencyStatKey struct {
	AppInstKey      edgeproto.AppInstKey
	DeviceCarrier   string
	LocationTile    string
	DataNetworkType string
	DeviceOs        string
	DeviceModel     string
	SignalStrength  uint64
}

// Wrapper struct that holds values for latency stats
type LatencyStat struct {
	LatencyCounts     grpcstats.LatencyMetric // buckets for counts
	LatencyBuckets    []time.Duration
	RollingStatistics *grpcstats.RollingStatistics // General stats: Avg, StdDev, Min, Max
	Mux               sync.Mutex
	Changed           bool
}

// Constants for Debug
const (
	RequestAppInstLatency = "request-appinst-latency"
)

func NewLatencyStat(latencyBuckets []time.Duration) *LatencyStat {
	l := new(LatencyStat)
	l.LatencyBuckets = latencyBuckets
	grpcstats.InitLatencyMetric(&l.LatencyCounts, latencyBuckets)
	l.RollingStatistics = grpcstats.NewRollingStatistics()
	return l
}

func (l *LatencyStat) ResetLatencyStat() {
	grpcstats.InitLatencyMetric(&l.LatencyCounts, l.LatencyBuckets)
	l.RollingStatistics = grpcstats.NewRollingStatistics()
	l.Changed = false
}

func (l *LatencyStat) Update(info *LatencyStatInfo) {
	if info != nil && info.Samples != nil && len(info.Samples) > 0 {
		// Update Latency counts and rolling statistics
		for _, sample := range info.Samples {
			if sample.Value >= 0 {
				l.Changed = true
				l.LatencyCounts.AddLatency(time.Duration(sample.Value) * time.Millisecond)
				l.RollingStatistics.UpdateRollingStatistics(sample.Value)
			}
		}
	}
}
