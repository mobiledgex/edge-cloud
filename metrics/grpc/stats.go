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

package grpcstats

import (
	"time"

	"github.com/edgexr/edge-cloud/edgeproto"
)

type LatencyMetric struct {
	times   []time.Duration
	buckets []uint64
}

func InitLatencyMetric(m *LatencyMetric, times []time.Duration) {
	m.times = times
	m.buckets = make([]uint64, len(times)+1)
}

func (m *LatencyMetric) AddLatency(d time.Duration) {
	ii := 0
	for ii, _ = range m.times {
		if d < m.times[ii] {
			ii--
			break
		}
	}
	if ii >= 0 {
		m.buckets[ii]++
	}
}

func (m *LatencyMetric) AddToMetric(metric *edgeproto.Metric) {
	ii := 0
	var t time.Duration
	for ii, t = range m.times {
		metric.AddIntVal(t.String(), m.buckets[ii])
	}
}

func (m *LatencyMetric) FromMetric(metric *edgeproto.Metric) {
	m.times = make([]time.Duration, 0)
	m.buckets = make([]uint64, 0)
	for _, val := range metric.Vals {
		t, err := time.ParseDuration(val.Name)
		if err != nil {
			continue
		}
		m.times = append(m.times, t)
		m.buckets = append(m.buckets, val.GetIval())
	}
}
