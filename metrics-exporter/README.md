# Metrics-exporter service 

`metrics-exporter` service collects prometheus metrics on a cluster and writes them into an influxDB.

The service collects cluster-wide as well as per-pod metrics. Service will create database `clusterstats` in the influxDB if one doesn't exist.
Cluster metrics are collected and stored in `crm-cluster` measurement in `clusterstats` database. It includes the following metrics:
   - `cpu` - cluster CPU utilization percentage
   - `mem` - cluster memory utilization percentage
   - `disk` - cluster filesystem utilization percentage
   - `sendBytes` - cluster tx traffic rate averaged over 1 minute
   - `recvBytes` - cluster rx traffic rate averaged over 1 minute
   - `tcpConns` - total number of established TCP connections on this cluster
   - `tcpRetrans` - total number of TCP retransmissions on this cluster
   - `udpRecv` - total number of rx UDP datagrams on this cluster
   - `udpSend` - total number of tx UDP datagrams on this cluster
   - `udpRecvErr` - tatal number of UDP errors received on this cluster
In addition to the above values `cluster` tag is added to each measurement with the name of a cluster.
Per-pod metrics are collected and stored in `crm-appinst` measurement in `clusterstats` database. The following metrics are collected:
   - `cpu` - CPU utilization of this pod as a percentage of total available CPU
   - `mem` - current memory footprint of a given pod in bytes
   - `disk` - filesystem usage for a given pod
   - `sendBytes` - tx traffic rate averaged over 1 minute for a given pod
   - `recvBytes` - rx traffic rate averaged over 1 minute for a given pod
In addition to the above values `cluster`, `dev`, and `app` tags are added to the measurement to uniquely identify a particular time series.

The collection of the above metrics happens every set interval by running queries agains a prometheus running in a different pod on the same cluster. See Usage section for addition details of how to configure interval/influxDB address/Prometheus address. In a typical deployment `cluster-svc` will set the appropriate environement variables for `metrics-exporter` when it deploys it on a given cluster. 

## Usage

In a most usual case this service is started by `cluster-svc` as a pod on a kubernetes cluster, but a process can be started locally with the following usage.

```
$ metrics-exporter -h
Usage of metrics-exporter:
  -apiAddr string
    	Prometheus address to bind to (default "0.0.0.0:9090")
  -cloudlet string
    	Cloudlet Name (default "local")
  -cluster string
    	Cluster Name (default "myclust")
  -d string
    	comma separated list of [etcd api notify dmedb dmereq locapi mexos metrics]
  -influxdb string
    	InfluxDB address to export to (default "0.0.0.0:8086")
  -interval duration
    	Metrics collection interval (default 15s)
  -operator string
    	Cloudlet Operator Name (default "local")
```

The following environment variable will take precedence over command-line arguments:

   - MEX_OPERATOR_NAME - equivalent to `-operator`
   - MEX_CLOUDLET_NAME - equivalent to `-cloudlet`
   - MEX_CLUSTER_NAME - equivalent to `-cluster`
   - MEX_INFLUXDB_ADDR - equivalent to `-influxdb`
   - MEX_INFLUXDB_USER - username that is used to connect to Influxdb, defaults to `root`
   - MEX_INFLUXDB_PASS - password that is used to connect to Influxdb, defaults to `root`
   - MEX_SCRAPE_INTERVAL - equivalent to `-interval`

## Docker Image

`make build-docker` will build, create(from Dockerfile) and upload the image to our repository at:
`registry.mobiledgex.net:5000/mobiledgex/metrics-exporter:latest`
This is the intended deployment packaging of this service


## TODO

1. Need to add different tag to image in the docker registry - currently only have `latest`
2. Long term we need to retire `metrics-exporter` as it's a too rigid to provide configurable metrics
   - ideally a federated prometheus system with a global prometheus scraping cluster prometheuses

