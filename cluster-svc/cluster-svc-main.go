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
	"time"

	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/objstore"
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
var externalPorts = flag.String("prometheus-ports", "", "ports to expose in form \"tcp:123,udp:123\"")
var influxDBAddr = flag.String("influxdb", "0.0.0.0:8086", "InfluxDB address to export to")
var influxDBUser = flag.String("influxdb-user", "root", "InfluxDB username")
var influxDBPass = flag.String("influxdb-pass", "root", "InfluxDB password")
var scrapeInterval = flag.Duration("scrapeInterval", time.Second*15, "Metrics collection interval")

var MEXInfraFlavorName = "x1.medium" // TODO - change to "infra.small" EDGECLOUD-391
var MEXInfraFlavor = edgeproto.Flavor{
	Key:   edgeproto.FlavorKey{Name: MEXInfraFlavorName},
	Vcpus: 1,
	Ram:   1024,
	Disk:  1,
}
var MEXMetricsExporterAppName = "MEXMetricsExporter"
var MEXMetricsExporterAppVer = "1.0"

var exporterT *template.Template

var MEXMetricsExporterEnvVars = `- name: MEX_CLUSTER_NAME
  valueFrom:
    configMapKeyRef:
      name: cluster-info
      key: ClusterName
      optional: true
- name: MEX_CLOUDLET_NAME
  valueFrom:
    configMapKeyRef:
      name: cluster-info
      key: CloudletName
      optional: true
- name: MEX_OPERATOR_NAME
  valueFrom:
    configMapKeyRef:
      name: cluster-info
      key: OperatorName
      optional: true
`
var MEXMetricsExporterEnvTempl = `- name: MEX_INFLUXDB_ADDR
  value: {{.InfluxDBAddr}}
- name: MEX_INFLUXDB_USER
  value: {{.InfluxDBUser}}
- name: MEX_INFLUXDB_PASS
  value: {{.InfluxDBPass}}
- name: MEX_SCRAPE_INTERVAL
  value: {{.Interval}}
`
var prometheusT *template.Template
var MEXPrometheusAppHelmTemplate = `prometheus:
  prometheusSpec:
    scrapeInterval: "{{.Interval}}"
kubelet:
  serviceMonitor:
    ## Enable scraping the kubelet over https. For requirements to enable this see
    ## https://github.com/coreos/prometheus-operator/issues/926
    ##
    https: true
`

var MEXMetricsExporterApp = edgeproto.App{
	Key: edgeproto.AppKey{
		Name:    MEXMetricsExporterAppName,
		Version: MEXMetricsExporterAppVer,
	},
	ImagePath:     "registry.mobiledgex.net:5000/mobiledgex/metrics-exporter:latest",
	ImageType:     edgeproto.ImageType_ImageTypeDocker,
	DefaultFlavor: edgeproto.FlavorKey{Name: MEXInfraFlavorName},
	DelOpt:        edgeproto.DeleteType_AutoDelete,
}

var MEXPrometheusAppName = "MEXPrometheusAppName"
var MEXPrometheusAppVer = "1.0"

var MEXPrometheusApp = edgeproto.App{
	Key: edgeproto.AppKey{
		Name:    MEXPrometheusAppName,
		Version: MEXPrometheusAppVer,
	},
	ImagePath:     "stable/prometheus-operator",
	Deployment:    cloudcommon.AppDeploymentTypeHelm,
	DefaultFlavor: edgeproto.FlavorKey{Name: MEXInfraFlavorName},
	DelOpt:        edgeproto.DeleteType_AutoDelete,
}

var dialOpts grpc.DialOption

var sigChan chan os.Signal

type ClusterInstHandler struct {
}

type exporterData struct {
	InfluxDBAddr string
	InfluxDBUser string
	InfluxDBPass string
	Interval     time.Duration
}

// Process updates from notify framework about cluster instances
// Create app/appInst when clusterInst transitions to a 'ready' state
func (c *ClusterInstHandler) Update(in *edgeproto.ClusterInst, rev int64) {
	var err error
	log.DebugLog(log.DebugLevelNotify, "cluster update", "cluster", in.Key.ClusterKey.Name,
		"cloudlet", in.Key.CloudletKey.Name, "state", edgeproto.TrackedState_name[int32(in.State)])
	// Need to create a connection to server, as passed to us by commands
	if in.State == edgeproto.TrackedState_Ready {
		// Create Two applications on the cluster after creation
		//   - Prometheus and MetricsExporter
		if err = createMEXPromInst(dialOpts, in.Key); err != nil {
			log.DebugLog(log.DebugLevelMexos, "Prometheus-operator inst create failed", "cluster", in.Key.ClusterKey.Name,
				"error", err.Error())
		}
		if err = createMEXMetricsExporterInst(dialOpts, in.Key); err != nil {
			log.DebugLog(log.DebugLevelMexos, "Metrics-exporter inst create failed", "cluster", in.Key.ClusterKey.Name,
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
	exporterT = template.Must(template.New("exporter").Parse(MEXMetricsExporterEnvTempl))
	prometheusT = template.Must(template.New("prometheus").Parse(MEXPrometheusAppHelmTemplate))
}

func initNotifyClient(addrs string, tlsCertFile string) *notify.Client {
	notifyClient := notify.NewClient(strings.Split(addrs, ","), tlsCertFile)
	notifyClient.RegisterRecv(notify.NewClusterInstRecv(&ClusterInstHandler{}))
	log.InfoLog("notify client to", "addrs", addrs)
	return notifyClient
}

// create an appInst as a clustersvc
func createAppInstCommon(dialOpts grpc.DialOption, instKey edgeproto.ClusterInstKey, app *edgeproto.App) error {
	conn, err := grpc.Dial(*ctrlAddr, dialOpts, grpc.WithBlock(), grpc.WithWaitForHandshake())
	if err != nil {
		return fmt.Errorf("Connect to server %s failed: %s", *ctrlAddr, err.Error())
	}
	defer conn.Close()
	apiClient := edgeproto.NewAppInstApiClient(conn)

	appInst := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      app.Key,
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
		// Handle non-fatal errors
		if strings.Contains(err.Error(), objstore.ErrKVStoreKeyExists.Error()) {
			log.DebugLog(log.DebugLevelMexos, "appinst already exists", "app", app.String(), "cluster", instKey.String())
			return nil
		}
		if strings.Contains(err.Error(), edgeproto.ErrEdgeApiAppNotFound.Error()) {
			// Create the app
			if err = createAppCommon(dialOpts, app); err == nil {
				log.DebugLog(log.DebugLevelMexos, "app doesn't exist, create it first", "app", app.String())
				return createAppInstCommon(dialOpts, instKey, app)
			}
		}
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
	return createAppInstCommon(dialOpts, instKey, &MEXPrometheusApp)
}

func createMEXMetricsExporterInst(dialOpts grpc.DialOption, instKey edgeproto.ClusterInstKey) error {
	return createAppInstCommon(dialOpts, instKey, &MEXMetricsExporterApp)
}

func createMEXInfraFlavor(dialOpts grpc.DialOption) error {
	conn, err := grpc.Dial(*ctrlAddr, dialOpts, grpc.WithBlock(), grpc.WithWaitForHandshake())
	if err != nil {
		return fmt.Errorf("Connect to server %s failed: %s", *ctrlAddr, err.Error())
	}
	defer conn.Close()
	apiClient := edgeproto.NewFlavorApiClient(conn)

	ctx := context.TODO()
	res, err := apiClient.CreateFlavor(ctx, &MEXInfraFlavor)
	if err != nil {
		if err == objstore.ErrKVStoreKeyExists {
			log.DebugLog(log.DebugLevelMexos, "flavor already exists", "flavor", MEXInfraFlavor.String())
			return nil
		}
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateFlavor failed: %s", errstr)
	}
	log.DebugLog(log.DebugLevelMexos, "create flavor", "flavor", MEXInfraFlavor.String(), "result", res.String())
	return nil

}

func fillAppConfigs(app *edgeproto.App) error {
	switch app.Key.Name {
	case MEXMetricsExporterAppName:
		ex := exporterData{
			InfluxDBAddr: *influxDBAddr,
			InfluxDBUser: *influxDBUser,
			InfluxDBPass: *influxDBPass,
			Interval:     *scrapeInterval,
		}
		buf := bytes.Buffer{}
		err := exporterT.Execute(&buf, &ex)
		if err != nil {
			return err
		}
		paramConf := edgeproto.ConfigFile{
			Kind:   mexos.AppConfigEnvYaml,
			Config: buf.String(),
		}
		envConf := edgeproto.ConfigFile{
			Kind:   mexos.AppConfigEnvYaml,
			Config: MEXMetricsExporterEnvVars,
		}

		app.Configs = []*edgeproto.ConfigFile{&paramConf, &envConf}
	case MEXPrometheusAppName:
		ex := exporterData{
			Interval: *scrapeInterval,
		}
		buf := bytes.Buffer{}
		err := prometheusT.Execute(&buf, &ex)
		if err != nil {
			return err
		}
		// Now add this yaml to the prometheus AppYamls
		config := edgeproto.ConfigFile{
			Kind:   mexos.AppConfigHemYaml,
			Config: buf.String(),
		}
		app.Configs = []*edgeproto.ConfigFile{&config}
		app.AccessPorts = *externalPorts
	default:
		return fmt.Errorf("Unrecognized app %s", app.Key.Name)
	}
	return nil
}

func createAppCommon(dialOpts grpc.DialOption, app *edgeproto.App) error {
	conn, err := grpc.Dial(*ctrlAddr, dialOpts, grpc.WithBlock(), grpc.WithWaitForHandshake())
	if err != nil {
		return fmt.Errorf("Connect to server %s failed: %s", *ctrlAddr, err.Error())
	}
	defer conn.Close()

	// add app customizations
	if err = fillAppConfigs(app); err != nil {
		return err
	}
	apiClient := edgeproto.NewAppApiClient(conn)
	ctx := context.TODO()
	res, err := apiClient.CreateApp(ctx, app)
	if err != nil {
		// Handle non-fatal errors
		if strings.Contains(err.Error(), objstore.ErrKVStoreKeyExists.Error()) {
			log.DebugLog(log.DebugLevelMexos, "app already exists", "app", app.String())
			return nil
		}
		// TODO - uncomment below along with fixes to yaml files for EDGECLOUD-391
		/*
			if strings.Contains(err.Error(), edgeproto.ErrEdgeApiFlavorNotFound.Error()) {
				if err = createMEXInfraFlavor(dialOpts); err == nil {
					log.DebugLog(log.DebugLevelMexos, "flavor doesn't exist, create it first", "flavor", MEXInfraFlavor.String())
					return createAppCommon(dialOpts, app)
				}
			}
		*/
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
	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}
