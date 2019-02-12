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

type PodPromStat struct {
	cpu     float64
	mem     uint64
	disk    float64
	netSend uint64
	netRecv uint64
}

type ClustPromStat struct {
	cpu        float64
	mem        float64
	disk       float64
	netSend    uint64
	netRecv    uint64
	tcpConns   uint64
	tcpRetrans uint64
	udpSend    uint64
	udpRecv    uint64
	udpRecvErr uint64
}

type PromStats struct {
	promAddr    string
	interval    time.Duration
	appStatsMap map[MetricAppInstKey]*PodPromStat
	clusterStat *ClustPromStat
	send        func(metric *edgeproto.Metric)
	waitGrp     sync.WaitGroup
	stop        chan struct{}
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
}

func NewPromStats(promAddr string, interval time.Duration, send func(metric *edgeproto.Metric)) *PromStats {
	p := PromStats{}
	p.promAddr = promAddr
	p.appStatsMap = make(map[MetricAppInstKey]*PodPromStat)
	p.clusterStat = &ClustPromStat{}
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
		cluster:   *clusterName,
		developer: "",
	}
	// Get Pod CPU usage percentage
	resp, err := getPromMetrics(p.promAddr, promQCpuPod)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.pod = metric.Labels.PodName
			stat, found := p.appStatsMap[appKey]
			if !found {
				stat = &PodPromStat{}
				p.appStatsMap[appKey] = stat
			}
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				stat.cpu = val
			}
		}
	}
	// Get Pod Mem usage
	resp, err = getPromMetrics(p.promAddr, promQMemPod)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.pod = metric.Labels.PodName
			stat, found := p.appStatsMap[appKey]
			if !found {
				stat = &PodPromStat{}
				p.appStatsMap[appKey] = stat
			}
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				stat.mem = val
			}
		}
	}
	// Get Pod NetRecv bytes rate averaged over 1m
	resp, err = getPromMetrics(p.promAddr, promQNetRecvRate)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.pod = metric.Labels.PodName
			stat, found := p.appStatsMap[appKey]
			if !found {
				stat = &PodPromStat{}
				p.appStatsMap[appKey] = stat
			}
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				stat.netRecv = uint64(val)
			}
		}
	}
	// Get Pod NetRecv bytes rate averaged over 1m
	resp, err = getPromMetrics(p.promAddr, promQNetSendRate)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.pod = metric.Labels.PodName
			stat, found := p.appStatsMap[appKey]
			if !found {
				stat = &PodPromStat{}
				p.appStatsMap[appKey] = stat
			}
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				stat.netSend = uint64(val)
			}
		}
	}

	// Get Cluster CPU usage
	resp, err = getPromMetrics(p.promAddr, promQCpuClust)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				p.clusterStat.cpu = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster Mem usage
	resp, err = getPromMetrics(p.promAddr, promQMemClust)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				p.clusterStat.mem = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster Disk usage percentage
	resp, err = getPromMetrics(p.promAddr, promQDiskClust)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				p.clusterStat.disk = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster NetRecv bytes rate averaged over 1m
	resp, err = getPromMetrics(p.promAddr, promQRecvBytesRateClust)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				p.clusterStat.netRecv = uint64(val)
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster NetSend bytes rate averaged over 1m
	resp, err = getPromMetrics(p.promAddr, promQSendBytesRateClust)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				p.clusterStat.netSend = uint64(val)
				// We should have only one value here
				break
			}
		}
	}

	// Get Cluster Established TCP connections
	resp, err = getPromMetrics(p.promAddr, promQTcpConnClust)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				p.clusterStat.tcpConns = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster TCP retransmissions
	resp, err = getPromMetrics(p.promAddr, promQTcpRetransClust)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				p.clusterStat.tcpRetrans = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster UDP Send Datagrams
	resp, err = getPromMetrics(p.promAddr, promQUdpSendPktsClust)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				p.clusterStat.udpSend = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster UDP Recv Datagrams
	resp, err = getPromMetrics(p.promAddr, promQUdpRecvPktsClust)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				p.clusterStat.udpRecv = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster UDP Recv Errors
	resp, err = getPromMetrics(p.promAddr, promQUdpRecvErr)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				p.clusterStat.udpRecvErr = val
				// We should have only one value here
				break
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
			DebugPrint("Sending metrics for %s with timestamp %s\n", *clusterName, ts.String())
			for key, stat := range p.appStatsMap {
				p.send(PodStatToMetric(ts, &key, stat))
			}
			p.send(ClusterStatToMetric(ts, p.clusterStat))
		case <-p.stop:
			done = true
		}
	}
	p.waitGrp.Done()
}

func ClusterStatToMetric(ts *types.Timestamp, stat *ClustPromStat) *edgeproto.Metric {
	metric := edgeproto.Metric{}
	metric.Timestamp = *ts
	metric.Name = "crm-cluster"
	metric.AddTag("cluster", *clusterName)
	metric.AddDoubleVal("cpu", stat.cpu)
	metric.AddDoubleVal("mem", stat.mem)
	metric.AddDoubleVal("disk", stat.disk)
	metric.AddIntVal("sendBytes", stat.netSend)
	metric.AddIntVal("recvBytes", stat.netRecv)
	metric.AddIntVal("tcpConns", stat.tcpConns)
	metric.AddIntVal("tcpRetrans", stat.tcpRetrans)
	metric.AddIntVal("udpSend", stat.udpSend)
	metric.AddIntVal("udpRecv", stat.udpRecv)
	metric.AddIntVal("udpRecvErr", stat.udpRecvErr)
	return &metric
}

func PodStatToMetric(ts *types.Timestamp, key *MetricAppInstKey, stat *PodPromStat) *edgeproto.Metric {
	metric := edgeproto.Metric{}
	metric.Timestamp = *ts
	metric.Name = "crm-appinst"
	metric.AddTag("dev", key.developer)
	metric.AddTag("cluster", key.cluster)
	metric.AddTag("app", key.pod)
	metric.AddDoubleVal("cpu", stat.cpu)
	metric.AddIntVal("mem", stat.mem)
	metric.AddDoubleVal("disk", stat.disk)
	metric.AddIntVal("sendBytes", stat.netSend)
	metric.AddIntVal("recvBytes", stat.netRecv)
	return &metric
}
