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

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	grpcstats "github.com/mobiledgex/edge-cloud/metrics/grpc"
)

// Filled in by DME. Added to EdgeEventStatCall to update stats
type CustomStatInfo struct {
	Samples []*dme.Sample
}

// Used to find corresponding CustomStat
// Created using CustomStatInfo
type CustomStatKey struct {
	AppInstKey edgeproto.AppInstKey
	Name       string
}

func GetCustomStatKey(appInstKey edgeproto.AppInstKey, statName string) CustomStatKey {
	return CustomStatKey{
		AppInstKey: appInstKey,
		Name:       statName,
	}
}

type CustomStat struct {
	Count             uint64 // number of times this custom stat has been updated
	RollingStatistics *grpcstats.RollingStatistics
	Mux               sync.Mutex
	Changed           bool
}

func NewCustomStat() *CustomStat {
	c := new(CustomStat)
	c.RollingStatistics = grpcstats.NewRollingStatistics()
	return c
}

func (c *CustomStat) Update(info *CustomStatInfo) {
	c.Changed = true
	c.Count++
	if info.Samples != nil {
		for _, sample := range info.Samples {
			c.RollingStatistics.UpdateRollingStatistics(sample.Value)
		}
	}
}
