package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"time"

	influxq "github.com/mobiledgex/edge-cloud/controller/influxq_clinet"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var promAddress = flag.String("apiAddr", "0.0.0.0:9090", "Prometheus address to bind to")
var influxdb = flag.String("influxdb", "0.0.0.0:8086", "InfluxDB address to export to")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var clusterName = flag.String("cluster", "myclust", "Cluster Name")

var promQCpuClust = "sum(rate(container_cpu_usage_seconds_total%7Bid%3D%22%2F%22%7D%5B1m%5D))%2Fsum(machine_cpu_cores)*100"
var promQMemClust = "sum(container_memory_working_set_bytes%7Bid%3D%22%2F%22%7D)%2Fsum(machine_memory_bytes)*100"
var promQDiskClust = "sum(container_fs_usage_bytes%7Bdevice%3D~%22%5E%2Fdev%2F%5Bsv%5Dd%5Ba-z%5D%5B1-9%5D%24%22%2Cid%3D%22%2F%22%7D)%2Fsum(container_fs_limit_bytes%7Bdevice%3D~%22%5E%2Fdev%2F%5Bsv%5Dd%5Ba-z%5D%5B1-9%5D%24%22%2Cid%3D%22%2F%22%7D)*100"
var promQSendBytesRateClust = "sum(irate(container_network_transmit_bytes_total%5B1m%5D))"
var promQRecvBytesRateClust = "sum(irate(container_network_receive_bytes_total%5B1m%5D))"

var promQCpuPod = "sum(rate(container_cpu_usage_seconds_total%7Bimage!%3D%22%22%7D%5B1m%5D))by(pod_name)"
var promQMemPod = "sum(container_memory_working_set_bytes%7Bimage!%3D%22%22%7D)by(pod_name)"
var promQNetRecvRate = "sum(irate(container_network_receive_bytes_total%7Bimage!%3D%22%22%7D%5B1m%5D))by(pod_name)"
var promQNetSendRate = "sum(irate(container_network_transmit_bytes_total%7Bimage!%3D%22%22%7D%5B1m%5D))by(pod_name)"

var Env = map[string]string{
	"INFLUXDB_USER": "root",
	"INFLUXDB_PASS": "root",
}

var InfluxDBName = "clusterstats"
var influxQ *influxq.InfluxQ

var sigChan chan os.Signal

func getIPfromEnv() (string, error) {
	re := regexp.MustCompile(".*PROMETHEUS_PORT_9090_TCP_ADDR=(.*)")
	for _, e := range os.Environ() {
		match := re.FindStringSubmatch(e)
		if len(match) > 1 {
			return match[1], nil
		}
	}
	return "", errors.New("No Prometheus is running")
}

func initEnv() {
	val := os.Getenv("MEX_CLUSTER_NAME")
	if val != "" {
		*clusterName = val
	}
	val = os.Getenv("MEX_INFLUXDB_ADDR")
	if val != "" {
		*influxdb = val
	}
	val = os.Getenv("MEX_INFLUXDB_USER")
	if val != "" {
		Env["INFLUXDB_USER"] = val
	}
	val = os.Getenv("MEX_INFLUXDB_PASS")
	if val != "" {
		Env["INFLUXDB_PASS"] = val
	}
}

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	fmt.Printf("Starting metrics exporter with Prometheus addr %s\n", *promAddress)
	clustIP, err := getIPfromEnv()
	if err == nil {
		*promAddress = clustIP + ":9090"
	}
	fmt.Printf("Found Prometheus running on: %s\n", *promAddress)
	initEnv()
	influxQ = influxq.NewInfluxQ(InfluxDBName)
	err = influxQ.Start(*influxdb)
	if err != nil {
		log.FatalLog("Failed to start influx queue",
			"address", *influxdb, "err", err)
	}
	defer influxQ.Stop()

	stats := NewPromStats(*promAddress, time.Second*15, sendMetric)
	stats.Start()
	defer stats.Stop()

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	DebugPrint("Ready\n")

	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}

func DebugPrint(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func sendMetric(metric *edgeproto.Metric) {
	influxQ.AddMetric(metric)
}
