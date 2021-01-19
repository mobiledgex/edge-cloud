package grpcstats

import (
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
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
