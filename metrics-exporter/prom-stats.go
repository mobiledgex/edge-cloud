package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type MetricAppInstKey struct {
	cluster   string
	pod       string
	developer string
}

type PromStat struct {
	cpu     float64
	mem     uint64
	netSend uint64
	netRecv uint64
}

type PromStats struct {
	promAddr string
	interval time.Duration
	statsMap map[MetricAppInstKey]*PromStat
	send     func(metric *edgeproto.Metric)
	waitGrp  sync.WaitGroup
	stop     chan struct{}
}

type PromResp struct {
	Status string   `json:"status,omitempty"`
	Data   PromData `json:"data,omitempty"`
}
type PromData struct {
	ResType string       `json:"resultType,omitempty"`
	Result  []PromMetric `json:"result,omitempty"`
}
type PromMetric struct {
	Labels PromLables    `json:"metric,omitempty"`
	Values []interface{} `json:"value,omitempty"`
}
type PromLables struct {
	PodName string `json:"pod_name,omitempty"`
	///Others
}

func NewPromStats(promAddr string, interval time.Duration, send func(metric *edgeproto.Metric)) *PromStats {
	p := PromStats{}
	p.promAddr = promAddr
	p.statsMap = make(map[MetricAppInstKey]*PromStat)
	p.interval = interval
	p.send = send
	return &p
}

func getPromMetrics(addr string, query string) (*PromResp, error) {
	reqURI := "http://" + addr + "/api/v1/query?query=" + query
	resp, err := http.Get(reqURI)
	if err != nil {
		DebugPrint("Failed to run <%s>\n",
			"http://"+addr+"/api/v1/query?query="+query)
		return nil, err
	}
	defer resp.Body.Close()

	promResp := &PromResp{}
	if err = json.NewDecoder(resp.Body).Decode(promResp); err != nil {
		return nil, err
	}
	return promResp, nil
}

func (p *PromStats) CollectPromStats() error {
	appKey := MetricAppInstKey{
		cluster:   "myclust",
		developer: "",
	}
	// Get Pod CPU usage percentage
	resp, err := getPromMetrics(p.promAddr, promQCpuPod)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.pod = metric.Labels.PodName
			stat, found := p.statsMap[appKey]
			if !found {
				stat = &PromStat{}
				p.statsMap[appKey] = stat
			}
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				stat.cpu = val
			}
		}
	}
	// Get Pod Mem usage percentage
	resp, err = getPromMetrics(p.promAddr, promQMemPod)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.pod = metric.Labels.PodName
			stat, found := p.statsMap[appKey]
			if !found {
				stat = &PromStat{}
				p.statsMap[appKey] = stat
			}
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				stat.mem = val
			}
		}
	}
	return nil
}

func (p *PromStats) Start() {
	p.stop = make(chan struct{})
	p.waitGrp.Add(1)
	go p.RunNotify()
}

func (p *PromStats) Stop() {
	DebugPrint("Stopping PromStats thread\n")
	close(p.stop)
	p.waitGrp.Wait()
}

func (p *PromStats) RunNotify() {
	DebugPrint("Started PromStats thread\n")
	done := false
	for !done {
		select {
		case <-time.After(p.interval):
			ts, _ := types.TimestampProto(time.Now())
			if p.CollectPromStats() != nil {
				continue
			}
			for key, stat := range p.statsMap {
				p.send(StatToMetric(ts, &key, stat))
			}
		case <-p.stop:
			done = true
		}
	}
	p.waitGrp.Done()
}

func StatToMetric(ts *types.Timestamp, key *MetricAppInstKey, stat *PromStat) *edgeproto.Metric {
	metric := edgeproto.Metric{}
	metric.Timestamp = *ts
	metric.Name = "crm-appinst"
	metric.AddTag("dev", key.developer)
	metric.AddTag("cluster", key.cluster)
	metric.AddTag("app", key.pod)
	metric.AddDoubleVal("cpu", stat.cpu)
	metric.AddIntVal("mem", stat.mem)
	metric.AddIntVal("sendBytes", stat.netSend)
	metric.AddIntVal("recvBytes", stat.netRecv)
	return &metric
}
