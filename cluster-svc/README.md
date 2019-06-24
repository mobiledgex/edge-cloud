# Cluster-svc service

`cluster-svc` runs as a notifyee of MEX controller and listens to cluster instance notifications. When a cluster instance is created this service will add MEX-specific applications to this service.
Currently it creates `prometheus-operator`, and `metrics-exporter` services whenever a cluster service is created.
   - `prometheus-operator` - is a metrics collection framework and periodically scrapes the details from k8s. It is deployed as a helm chart
   - `metrics-exporter` - is a MEX service which converts and pushes metrics prometheus collects locally to influxDB. `metrics-exporter` is deployed with a custom deployments yaml where we set environment variables based on the command line arguments passed to `cluster-svc`

`cluster-svc` uses controller apis to create `edgeproto.App` and `edgeproto.AppInst` for the able two services using a pre-created flavor and developer

## Usage

A typical usage of `cluster-svc` is as a docker container running on the MEX Platform. For an example see edge-cloud/e2e-tests/setups/mexdemo/mex-cluster-svc-deploy.yml

```
LSHVARTS-MAC:edge-cloud levshvarts$ cluster-svc -h
Usage of cluster-svc:
  -ctrlAddrs string
    	address to connect to (default "127.0.0.1:55001")
  -d string
    	comma separated list of [etcd api notify dmedb dmereq locapi mexos metrics]
  -influxdb string
    	InfluxDB address to export to (default "http://0.0.0.0:8086")
  -influxdb-pass string
    	InfluxDB password (default "root")
  -influxdb-user string
    	InfluxDB username (default "root")
  -notifyAddrs string
    	Comma separated list of controller notify listener addresses (default "127.0.0.1:50001")
  -prometheus-ports string
    	ports to expose in form "tcp:123,udp:123"
  -r string
    	root directory for testing
  -scrapeInterval duration
    	Metrics collection interval (default 15s)
  -standalone
    	Standalone mode. AppInst data is pre-populated. Dme does not interact with controller. AppInsts can be created directly on Dme using controller AppInst API
  -tls string
    	server tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory
  ...
```

## TODO

1. scrapeInterval should be passed along to Prometheus app helm chart at creation time - currently defaults to 15secs
2. Need to create a separate flavor for infrastructure services not to conflict with user applications
3. Need to pass cloudlet and cluster information only when cluster is created to createAppInstCommon
   - EDGECLOUD-386
   - this will allow us to support multiple clusters/cloudlets