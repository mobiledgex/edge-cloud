// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dmecommon

import (
	"sync"
	"time"

	dme "github.com/edgexr/edge-cloud/d-match-engine/dme-proto"
	"github.com/edgexr/edge-cloud/edgeproto"
	grpcstats "github.com/edgexr/edge-cloud/metrics/grpc"
)

// Filled in by DME. Added to EdgeEventStatCall to update stats
type LatencyStatInfo struct {
	Samples []*dme.Sample
}

func GetLatencyStatKey(appInstKey edgeproto.AppInstKey, deviceInfo *DeviceInfo, loc *dme.Loc, tileLength int) LatencyStatKey {
	statKey := LatencyStatKey{
		AppInstKey:   appInstKey,
		LocationTile: GetLocationTileFromGpsLocation(loc, tileLength),
	}

	if deviceInfo.DeviceInfoStatic != nil {
		statKey.DeviceOs = deviceInfo.DeviceInfoStatic.DeviceOs
		statKey.DeviceModel = deviceInfo.DeviceInfoStatic.DeviceModel
	}
	if deviceInfo.DeviceInfoDynamic != nil {
		statKey.DeviceCarrier = deviceInfo.DeviceInfoDynamic.CarrierName
		statKey.DataNetworkType = deviceInfo.DeviceInfoDynamic.DataNetworkType
		statKey.SignalStrength = uint64(deviceInfo.DeviceInfoDynamic.SignalStrength)
	}
	return statKey
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
