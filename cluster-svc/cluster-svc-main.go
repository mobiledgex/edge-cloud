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
	"strconv"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
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
var externalPorts = flag.String("prometheus-ports", "tcp:9090", "ports to expose in form \"tcp:123,udp:123\"")
var scrapeInterval = flag.Duration("scrapeInterval", time.Second*15, "Metrics collection interval")
var appFlavor = flag.String("flavor", "x1.medium", "App flavor for cluster-svc applications")
var upgradeInstances = flag.Bool("updateAll", false, "Upgrade all Instances of Prometheus operator")

var exporterT *template.Template
var prometheusT *template.Template

var MEXPrometheusAppHelmTemplate = `prometheus:
  prometheusSpec:
    scrapeInterval: "{{.Interval}}"
  service:
    type: LoadBalancer
kubelet:
  serviceMonitor:
    ## Enable scraping the kubelet over https. For requirements to enable this see
    ## https://github.com/coreos/prometheus-operator/issues/926
    ##
    https: true
grafana:
  enabled: false
`

var MEXPrometheusAppName = cloudcommon.MEXPrometheusAppName
var MEXPrometheusAppVer = "1.0"

var MEXPrometheusAppKey = edgeproto.AppKey{
	Name:    MEXPrometheusAppName,
	Version: MEXPrometheusAppVer,
	DeveloperKey: edgeproto.DeveloperKey{
		Name: cloudcommon.DeveloperMobiledgeX,
	},
}

var MEXPrometheusApp = edgeproto.App{
	Key:           MEXPrometheusAppKey,
	ImagePath:     "stable/prometheus-operator",
	Deployment:    cloudcommon.AppDeploymentTypeHelm,
	DefaultFlavor: edgeproto.FlavorKey{Name: *appFlavor},
	DelOpt:        edgeproto.DeleteType_AUTO_DELETE,
	InternalPorts: true,
}

var dialOpts grpc.DialOption

var sigChan chan os.Signal

type ClusterInstHandler struct {
}

type exporterData struct {
	InfluxDBAddr string
	InfluxDBUser string
	InfluxDBPass string
	Interval     string
}

// Process updates from notify framework about cluster instances
// Create app/appInst when clusterInst transitions to a 'ready' state
func (c *ClusterInstHandler) Update(ctx context.Context, in *edgeproto.ClusterInst, rev int64) {
	var err error
	log.DebugLog(log.DebugLevelNotify, "cluster update", "cluster", in.Key.ClusterKey.Name,
		"cloudlet", in.Key.CloudletKey.Name, "state", edgeproto.TrackedState_name[int32(in.State)])
	// Need to create a connection to server, as passed to us by commands
	if in.State == edgeproto.TrackedState_READY {
		// Create Prometheus on the cluster after creation
		if err = createMEXPromInst(dialOpts, in.Key); err != nil {
			log.DebugLog(log.DebugLevelMexos, "Prometheus-operator inst create failed", "cluster", in.Key.ClusterKey.Name,
				"error", err.Error())
		}
	}
}

// Callback for clusterInst deletion - we don't need to do anything here
// Applications created by cluster service are created as auto-delete and will be removed
// when clusterInstance goes away
func (c *ClusterInstHandler) Delete(ctx context.Context, in *edgeproto.ClusterInst, rev int64) {
	log.DebugLog(log.DebugLevelNotify, "clusterInst delete", "cluster", in.Key.ClusterKey.Name, "state",
		edgeproto.TrackedState_name[int32(in.State)])
	// don't need to do anything really if a cluster instance is getting deleted
	// - all the pods in the cluster will be stopped anyways
}

// Don't need to do anything here - same as Delete
func (c *ClusterInstHandler) Prune(ctx context.Context, keys map[edgeproto.ClusterInstKey]struct{}) {
	log.DebugLog(log.DebugLevelNotify, "clusterInst prune")
}

func (c *ClusterInstHandler) Flush(ctx context.Context, notifyId int64) {}

func init() {
	prometheusT = template.Must(template.New("prometheus").Parse(MEXPrometheusAppHelmTemplate))
}

func initNotifyClient(addrs string, tlsCertFile string) *notify.Client {
	notifyClient := notify.NewClient(strings.Split(addrs, ","), tlsCertFile)
	notifyClient.RegisterRecv(notify.NewClusterInstRecv(&ClusterInstHandler{}))
	log.InfoLog("notify client to", "addrs", addrs)
	return notifyClient
}

func appInstCreateApi(apiClient edgeproto.AppInstApiClient, appInst edgeproto.AppInst) (*edgeproto.Result, error) {
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
	return res, err
}

func appInstUpdateApi(apiClient edgeproto.AppInstApiClient, appInst edgeproto.AppInst) (*edgeproto.Result, error) {
	ctx := context.TODO()
	stream, err := apiClient.UpdateAppInst(ctx, &appInst)
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
	return res, err
}

// create an appInst as a clustersvc
func createAppInstCommon(dialOpts grpc.DialOption, instKey edgeproto.ClusterInstKey, app *edgeproto.App) error {
	//update flavor
	app.DefaultFlavor = edgeproto.FlavorKey{Name: *appFlavor}
	conn, err := grpc.Dial(*ctrlAddr, dialOpts, grpc.WithBlock(), grpc.WithWaitForHandshake())
	if err != nil {
		return fmt.Errorf("Connect to server %s failed: %s", *ctrlAddr, err.Error())
	}
	defer conn.Close()
	apiClient := edgeproto.NewAppInstApiClient(conn)

	appInst := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:         app.Key,
			ClusterInstKey: instKey,
		},
	}

	res, err := appInstCreateApi(apiClient, appInst)
	if err != nil {
		// Handle non-fatal errors
		if strings.Contains(err.Error(), objstore.ErrKVStoreKeyExists.Error()) {
			log.DebugLog(log.DebugLevelMexos, "appinst already exists", "app", app.String(), "cluster", instKey.String())
			return nil
		}
		if strings.Contains(err.Error(), edgeproto.ErrEdgeApiAppNotFound.Error()) {
			log.DebugLog(log.DebugLevelMexos, "app doesn't exist, create it first", "app", app.String())
			// Create the app
			if err = createAppCommon(dialOpts, app); err == nil {
				if res, err = appInstCreateApi(apiClient, appInst); err == nil {
					log.DebugLog(log.DebugLevelMexos, "create appinst", "appinst", appInst.String(), "result", res.String())
					return nil
				}
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

func scrapeIntervalInSeconds(scrapeInterval time.Duration) string {
	var secs = int(scrapeInterval.Seconds()) //round it to the second
	var scrapeStr = strconv.Itoa(secs) + "s"
	return scrapeStr
}

func fillAppConfigs(app *edgeproto.App, interval time.Duration) error {
	var scrapeStr = scrapeIntervalInSeconds(interval)
	switch app.Key.Name {
	case MEXPrometheusAppName:
		ex := exporterData{
			Interval: scrapeStr,
		}
		buf := bytes.Buffer{}
		err := prometheusT.Execute(&buf, &ex)
		if err != nil {
			return err
		}
		// Now add this yaml to the prometheus AppYamls
		config := edgeproto.ConfigFile{
			Kind:   k8smgmt.AppConfigHelmYaml,
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
	if err = fillAppConfigs(app, *scrapeInterval); err != nil {
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

func getPrometheusAppFromController(ctx context.Context, apiClient edgeproto.AppApiClient) (*edgeproto.App, error) {
	stream, err := apiClient.ShowApp(ctx, &edgeproto.App{
		Key: MEXPrometheusAppKey,
	})
	if err != nil {
		return nil, err
	}
	// There should only be one app
	return stream.Recv()
}

func getPrometheusAppFromClusterSvc() (*edgeproto.App, error) {
	app := MEXPrometheusApp
	// add app customizations
	if err := fillAppConfigs(&app, *scrapeInterval); err != nil {
		return nil, err
	}
	return &app, nil
}

func setPrometheusAppDiffFields(src *edgeproto.App, dst *edgeproto.App) {
	fields := make(map[string]struct{})
	src.DiffFields(dst, fields)

	if _, found := fields[edgeproto.AppFieldImagePath]; found {
		dst.Fields = append(dst.Fields, edgeproto.AppFieldImagePath)
	}
	if _, found := fields[edgeproto.AppFieldConfigs]; found {
		dst.Fields = append(dst.Fields, edgeproto.AppFieldConfigs,
			edgeproto.AppFieldConfigsKind, edgeproto.AppFieldConfigsConfig)
	}
}

// Check if we are running the correct revision of prometheus app, and if not, upgrade it
func validatePrometheusRevision() error {
	conn, err := grpc.Dial(*ctrlAddr, dialOpts, grpc.WithBlock(), grpc.WithWaitForHandshake())
	if err != nil {
		return fmt.Errorf("Connect to server %s failed: %s", *ctrlAddr, err.Error())
	}
	defer conn.Close()
	ctx := context.TODO()
	apiClient := edgeproto.NewAppApiClient(conn)

	// Get Exusting prometheus App
	currentApp, err := getPrometheusAppFromController(ctx, apiClient)
	if err == io.EOF {
		// No app exists yet - just create it when we need to
		log.DebugLog(log.DebugLevelMexos, "app doesn't exist", "app", MEXPrometheusAppKey)
		return nil
	}
	if err != nil {
		return err
	}

	newApp, err := getPrometheusAppFromClusterSvc()
	if err != nil {
		return err
	}

	// Set the fields we want to update
	setPrometheusAppDiffFields(currentApp, newApp)
	if len(newApp.Fields) > 0 {
		_, err := apiClient.UpdateApp(ctx, newApp)
		if err != nil {
			errstr := err.Error()
			st, ok := status.FromError(err)
			if ok {
				errstr = st.Message()
			}
			return fmt.Errorf("UpdateApp failed: %s", errstr)
		}
		log.DebugLog(log.DebugLevelMexos, "update app", "app", newApp.String(), "result", currentApp.String())
	}
	// Update all appInstances of the Prometheus App
	if *upgradeInstances {
		apiClient := edgeproto.NewAppInstApiClient(conn)
		appInst := edgeproto.AppInst{
			Key: edgeproto.AppInstKey{
				AppKey: MEXPrometheusAppKey,
			},
			UpdateMultiple: true,
		}
		res, err := appInstUpdateApi(apiClient, appInst)
		if err != nil {
			return err
		}
		log.DebugLog(log.DebugLevelMexos, "update appinst", "appinst", appInst.String(), "result", res.String())
	}
	return nil
}

func main() {
	var err error
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	log.InitTracer(*tlsCertFile)
	defer log.FinishTracer()

	dialOpts, err = tls.GetTLSClientDialOption(*ctrlAddr, *tlsCertFile, false)
	if err != nil {
		log.FatalLog("get TLS Credentials", "error", err)
	}

	if err = validatePrometheusRevision(); err != nil {
		log.FatalLog("Validate Prometheus version", "error", err)
	}

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

	log.InfoLog("Ready")
	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}
