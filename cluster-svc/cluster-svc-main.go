package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Command line options
var rootDir = flag.String("r", "", "root directory for testing")
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var ctrlAddr = flag.String("ctrlAddrs", "127.0.0.1:55001", "address to connect to")
var standalone = flag.Bool("standalone", false, "Standalone mode. AppInst data is pre-populated. Dme does not interact with controller. AppInsts can be created directly on Dme using controller AppInst API")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var tlsCertFile = flag.String("tls", "", "server tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory")
var influxDBAddr = flag.String("influxdb", "0.0.0.0:8086", "InfluxDB address to export to")
var influxDBUser = flag.String("influxdb-user", "root", "InfluxDB username")
var influxDBPass = flag.String("influxdb-pass", "root", "InfluxDB password")

// Hard coded username - TODO to move to user db
var MEXDeveloper = "mexinfradev_"
var MEXDevUsername = "_mexinfradev"
var MEXDevPass = "_mexdeveloperpass"
var MEXdev = edgeproto.Developer{
	Key:      edgeproto.DeveloperKey{Name: MEXDeveloper},
	Username: MEXDevUsername,
}

var MEXMetricsExporterAppName = "MEXMetricsExporter"
var MEXMetricsExporterAppVer = "1.0"

var exporterT *template.Template

var MEXMetricsExporterTemplate = `apiVersion: v1
kind: Service
metadata:
  name: mexmetricsexporter-udp
  labels:
    run: mexmetricsexporter
spec:
  type: LoadBalancer
  ports:
  - name: udp12345
    protocol: UDP
    port: 12345
    targetPort: 12345
  selector:
    run: mexmetricsexporter
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mexmetricsexporter-deployment
spec:
  selector:
    matchLabels:
      run: mexmetricsexporter
  replicas: 1
  template:
    metadata:
      labels:
        run: mexmetricsexporter
    spec:
      volumes:
      imagePullSecrets:
      - name: mexregistrysecret
      containers:
      - name: mexmetricsexporter
        image: registry.mobiledgex.net:5000/mobiledgex/metrics-exporter:latest
        imagePullPolicy: Always
        env:
        - name: MEX_CLUSTER_NAME
          value: {{.Cluster}}
        - name: MEX_INFLUXDB_ADDR
          value: {{.InfluxDBAddr}}
        - name: MEX_INFLUXDB_USER
          value: {{.InfluxDBUser}}
        - name: MEX_INFLUXDB_PASS
          value: {{.InfluxDBPass}}
        ports:
`

var MEXMetricsExporterApp = edgeproto.App{
	Key: edgeproto.AppKey{
		DeveloperKey: MEXdev.Key,
		Name:         MEXMetricsExporterAppName,
		Version:      MEXMetricsExporterAppVer,
	},
	ImagePath:     "registry.mobiledgex.net:5000/mobiledgex/metrics-exporter:latest",
	ImageType:     edgeproto.ImageType_ImageTypeDocker,
	DefaultFlavor: edgeproto.FlavorKey{Name: "x1.medium"}, // TODO flavor
	DelOpt:        edgeproto.DeleteType_AutoDelete,
}

var MEXPrometheusAppName = "MEXPrometheusAppName"
var MEXPrometheusAppVer = "1.0"

var MEXPrometheusApp = edgeproto.App{
	Key: edgeproto.AppKey{
		DeveloperKey: MEXdev.Key,
		Name:         MEXPrometheusAppName,
		Version:      MEXPrometheusAppVer,
	},
	ImagePath:     "stable/prometheus-operator",
	Deployment:    cloudcommon.AppDeploymentTypeHelm,
	DefaultFlavor: edgeproto.FlavorKey{Name: "x1.medium"}, // TODO flavor
	DelOpt:        edgeproto.DeleteType_AutoDelete,
}

var dialOpts grpc.DialOption

var sigChan chan os.Signal

type ClusterInstHandler struct {
}

// Process updates from notify framework about cluster instances
// Create app/appInst when clusterInst transitions to a 'ready' state
// TODO - Once App is decoupled from Cluster, we should create Apps on a startup
func (c *ClusterInstHandler) Update(in *edgeproto.ClusterInst, rev int64) {
	var err error
	log.DebugLog(log.DebugLevelNotify, "cluster update", "cluster", in.Key.ClusterKey.Name,
		"cloudlet", in.Key.CloudletKey.Name, "state", edgeproto.TrackedState_name[int32(in.State)])
	// Need to create a connection to server, as passed to us by commands
	if in.State == edgeproto.TrackedState_Ready {
		// Create Applications
		// TODO - this should really be done on a strartup of cluster service
		if err = createMEXPrometheus(dialOpts, in.Key.ClusterKey); err != nil {
			log.DebugLog(log.DebugLevelMexos, "Prometheus-operator app create failed", "cluster", in.Key.ClusterKey.Name,
				"error", err.Error())
		}
		//TODO - need to wait a bit before running this command
		if err = createMEXMetricsExporter(dialOpts, in.Key.ClusterKey); err != nil {
			log.DebugLog(log.DebugLevelMexos, "metrics Exporter app create failed", "cluster", in.Key.ClusterKey.Name,
				"error", err.Error())
		}
		// Create Two applications on the cluster after creation
		//   - Prometheus and MetricsExporter
		if err = createMEXPromInst(dialOpts, in.Key); err != nil {
			log.DebugLog(log.DebugLevelMexos, "Prometheus-operator inst create failed", "cluster", in.Key.ClusterKey.Name,
				"error", err.Error())
		}
		if err = createMEXMetricsExporterInst(dialOpts, in.Key); err != nil {
			log.DebugLog(log.DebugLevelMexos, "Prometheus-operator inst create failed", "cluster", in.Key.ClusterKey.Name,
				"error", err.Error())
		}
	}
}

// Callback for clusterInst deletion - we don't need to do anything here
// Applications created by cluster service are created as auto-delete and will be removed
// when clusterInstance goes away
func (c *ClusterInstHandler) Delete(in *edgeproto.ClusterInst, rev int64) {
	log.DebugLog(log.DebugLevelNotify, "clusterInst delete", "cluster", in.Key.ClusterKey.Name, "state",
		edgeproto.TrackedState_name[int32(in.State)])
	// don't need to do anything really if a cluster instance is getting deleted
	// - all the pods in the cluster will be stopped anyways
}

// Don't need to do anything here - same as Delete
func (c *ClusterInstHandler) Prune(keys map[edgeproto.ClusterInstKey]struct{}) {
	log.DebugLog(log.DebugLevelNotify, "clusterInst prune")
}

func (c *ClusterInstHandler) Flush(notifyId int64) {}

func init() {
	exporterT = template.Must(template.New("exporter").Parse(MEXMetricsExporterTemplate))
}

func initNotifyClient(addrs string, tlsCertFile string) *notify.Client {
	notifyClient := notify.NewClient(strings.Split(addrs, ","), tlsCertFile)
	notifyClient.RegisterRecv(notify.NewClusterInstRecv(&ClusterInstHandler{}))
	log.InfoLog("notify client to", "addrs", addrs)
	return notifyClient
}

// create an appInst as a clustersvcdev
func createAppInstCommon(dialOpts grpc.DialOption, instKey edgeproto.ClusterInstKey, appKey edgeproto.AppKey) error {
	conn, err := grpc.Dial(*ctrlAddr, dialOpts, grpc.WithBlock(), grpc.WithWaitForHandshake())
	if err != nil {
		return fmt.Errorf("Connect to server %s failed: %s", *ctrlAddr, err.Error())
	}
	defer conn.Close()
	apiClient := edgeproto.NewAppInstApiClient(conn)

	appInst := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      appKey,
			CloudletKey: instKey.CloudletKey,
			Id:          1,
		},
		ClusterInstKey: instKey,
	}

	ctx := context.TODO()
	stream, err := apiClient.CreateAppInst(ctx, &appInst)
	var res *edgeproto.Result
	if err == nil {
		for {
			res, err = stream.Recv()
			if err == io.EOF {
				err = nil
				break
			}
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateAppInst failed: %s", errstr)
	}
	log.DebugLog(log.DebugLevelMexos, "create appinst", "appinst", appInst.String(), "result", res.String())
	return nil

}

func createMEXPromInst(dialOpts grpc.DialOption, instKey edgeproto.ClusterInstKey) error {
	return createAppInstCommon(dialOpts, instKey, MEXPrometheusApp.Key)
}

func createMEXMetricsExporterInst(dialOpts grpc.DialOption, instKey edgeproto.ClusterInstKey) error {
	return createAppInstCommon(dialOpts, instKey, MEXMetricsExporterApp.Key)
}

func createMEXInfraDeveloper(dialOpts grpc.DialOption) error {
	var err error

	conn, err := grpc.Dial(*ctrlAddr, dialOpts, grpc.WithBlock(), grpc.WithWaitForHandshake())
	if err != nil {
		return fmt.Errorf("Connect to server %s failed: %s", *ctrlAddr, err.Error())
	}
	defer conn.Close()
	apiClient := edgeproto.NewDeveloperApiClient(conn)

	ctx := context.TODO()
	res, err := apiClient.CreateDeveloper(ctx, &MEXdev)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateDeveloper failed: %s", errstr)
	}
	log.DebugLog(log.DebugLevelMexos, "create dev", "developer", MEXdev.String(), "result", res.String())
	return nil
}

func createAppCommon(dialOpts grpc.DialOption, app *edgeproto.App, cluster edgeproto.ClusterKey) error {
	conn, err := grpc.Dial(*ctrlAddr, dialOpts, grpc.WithBlock(), grpc.WithWaitForHandshake())
	if err != nil {
		return fmt.Errorf("Connect to server %s failed: %s", *ctrlAddr, err.Error())
	}
	defer conn.Close()
	apiClient := edgeproto.NewAppApiClient(conn)

	app.Cluster = cluster

	ctx := context.TODO()
	res, err := apiClient.CreateApp(ctx, app)
	if err != nil {
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateApp failed: %s", errstr)
	}
	log.DebugLog(log.DebugLevelMexos, "create app", "app", app.String(), "result", res.String())
	return nil
}
func createMEXPrometheus(dialOpts grpc.DialOption, cluster edgeproto.ClusterKey) error {
	return createAppCommon(dialOpts, &MEXPrometheusApp, cluster)
}

type exporterData struct {
	Cluster      string
	InfluxDBAddr string
	InfluxDBUser string
	InfluxDBPass string
}

func createMEXMetricsExporter(dialOpts grpc.DialOption, cluster edgeproto.ClusterKey) error {
	app := MEXMetricsExporterApp

	ex := exporterData{
		Cluster:      cluster.Name,
		InfluxDBAddr: *influxDBAddr,
		InfluxDBUser: *influxDBUser,
		InfluxDBPass: *influxDBPass,
	}
	buf := bytes.Buffer{}
	err := exporterT.Execute(&buf, &ex)
	if err != nil {
		return err
	}
	app.DeploymentManifest = buf.String()
	return createAppCommon(dialOpts, &app, cluster)
}

func main() {
	var err error
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)

	if *standalone {
		fmt.Printf("Running in Standalone Mode with test instances\n")
	} else {
		notifyClient := initNotifyClient(*notifyAddrs, *tlsCertFile)
		notifyClient.Start()
		defer notifyClient.Stop()
	}

	if *standalone {
		// TODO - unit tests see cluster-svc_test.go
	}
	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	dialOpts, err = tls.GetTLSClientDialOption(*ctrlAddr, *tlsCertFile)
	if err != nil {
		log.FatalLog("get TLS Credentials", "error", err)
	}

	log.InfoLog("Ready")
	if err = createMEXInfraDeveloper(dialOpts); err != nil {
		fmt.Printf("Failed to create a developer\n")
	}

	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}
