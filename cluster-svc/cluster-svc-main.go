package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"text/template"
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

// If an auto-scale profile is set, we add two measurements and two alerts.
// The measurements count the number of nodes that are above the high cpu
// threshold and below the low cpu threshold. These measurements assume
// the master node is not schedulable and other nodes are schedulable, which
// may not be true if custom taints are used.
// The scale up alert fires when all nodes are above the high cpu threshold,
// and there are less than the max number of nodes.
// The scale down alert files when any node is below the low cpu threshold,
// and there are more than the min number of nodes.
var MEXPrometheusAutoScaleT = `additionalPrometheusRules:
- name: [[.AutoScalePolicy]]
  groups:
  - name: autoscale.rules
    rules:
    - expr: sum(node:node_cpu_utilisation:avg1m{node=~"[[.NodePrefix]].*"} > bool .[[.ScaleUpCpuThresh]])
      record: 'node_cpu_high_count'
    - expr: sum(node:node_cpu_utilisation:avg1m{node=~"[[.NodePrefix]].*"} < bool .[[.ScaleDownCpuThresh]])
      record: 'node_cpu_low_count'
    - expr: count(kube_node_info) - count(kube_node_spec_taint)
      record: 'node_count'
    - alert: [[.AutoScaleUpName]]
      expr: node_cpu_high_count == node_count and node_count < [[.MaxNodes]]
      for: [[.TriggerTimeSec]]s
      labels:
        severity: none
      annotations:
        message: High cpu greater than [[.ScaleUpCpuThresh]]% for all nodes
        [[.NodeCountName]]: '{{ with query "node_count" }}{{ . | first | value | humanize }}{{ end }}'
    - alert: [[.AutoScaleDownName]]
      expr: node_cpu_low_count > 0 and node_count > [[.MinNodes]]
      for: [[.TriggerTimeSec]]s
      labels:
        severity: none
      annotations:
        message: Low cpu less than [[.ScaleDownCpuThresh]]% for some nodes
        [[.LowCpuNodeCountName]]: '{{ $value }}'
        [[.NodeCountName]]: '{{ with query "node_count" }}{{ . | first | value | humanize }}{{ end }}'
        [[.MinNodesName]]: '[[.MinNodes]]'
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

// Define prometheus operator App.
// Set version to avoid changes in behavior as helm and the operator
// code are not very stable.
var MEXPrometheusApp = edgeproto.App{
	Key:           MEXPrometheusAppKey,
	ImagePath:     "stable/prometheus-operator",
	Deployment:    cloudcommon.AppDeploymentTypeHelm,
	DefaultFlavor: edgeproto.FlavorKey{Name: *appFlavor},
	DelOpt:        edgeproto.DeleteType_AUTO_DELETE,
	InternalPorts: true,
	Annotations:   "version=6.7.3",
}

var dialOpts grpc.DialOption

var sigChan chan os.Signal

var AutoScalePolicyCache edgeproto.AutoScalePolicyCache
var ClusterInstCache edgeproto.ClusterInstCache

type promCustomizations struct {
	Interval string
}

type AppInstArgs struct {
	AutoScalePolicy     string
	AutoScaleUpName     string
	AutoScaleDownName   string
	ScaleUpCpuThresh    uint32
	ScaleDownCpuThresh  uint32
	TriggerTimeSec      uint32
	MaxNodes            int
	MinNodes            int
	NodeCountName       string
	LowCpuNodeCountName string
	MinNodesName        string
	NodePrefix          string // master node must have different prefix
}

// Process updates from notify framework about cluster instances
// Create app/appInst when clusterInst transitions to a 'ready' state
func clusterInstCb(ctx context.Context, old *edgeproto.ClusterInst, new *edgeproto.ClusterInst) {
	var err error
	log.SpanLog(ctx, log.DebugLevelNotify, "cluster update", "cluster", new.Key.ClusterKey.Name,
		"cloudlet", new.Key.CloudletKey.Name, "state", edgeproto.TrackedState_name[int32(new.State)])
	// Need to create a connection to server, as passed to us by commands
	if new.State == edgeproto.TrackedState_READY {
		// Create Prometheus on the cluster after creation
		if err = createMEXPromInst(ctx, dialOpts, new.Key, new.AutoScalePolicy); err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Prometheus-operator inst create failed", "cluster",
				new.Key.ClusterKey.Name, "error", err.Error())
		}
	}
}

func init() {
	prometheusT = template.Must(template.New("prometheus").Parse(MEXPrometheusAppHelmTemplate))
}

func initNotifyClient(ctx context.Context, addrs string, tlsCertFile string) *notify.Client {
	notifyClient := notify.NewClient(strings.Split(addrs, ","), tlsCertFile)
	edgeproto.InitAutoScalePolicyCache(&AutoScalePolicyCache)
	edgeproto.InitClusterInstCache(&ClusterInstCache)
	ClusterInstCache.SetUpdatedCb(clusterInstCb)
	log.SpanLog(ctx, log.DebugLevelInfo, "notify client to", "addrs", addrs)
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

func getAppInstConfigs(instKey edgeproto.ClusterInstKey, autoScalePolicy string) ([]*edgeproto.ConfigFile, error) {
	configs := []*edgeproto.ConfigFile{}
	if autoScalePolicy != "" {
		policy := edgeproto.AutoScalePolicy{}
		policyKey := edgeproto.PolicyKey{}
		policyKey.Developer = instKey.Developer
		policyKey.Name = autoScalePolicy
		if !AutoScalePolicyCache.Get(&policyKey, &policy) {
			return nil, fmt.Errorf("Auto scale policy %s not found for ClusterInst %s", autoScalePolicy, instKey.GetKeyString())
		}
		// change delims because Prometheus triggers off of golang delims
		t := template.Must(template.New("policy").Delims("[[", "]]").Parse(MEXPrometheusAutoScaleT))
		args := AppInstArgs{
			AutoScalePolicy:     autoScalePolicy,
			AutoScaleUpName:     cloudcommon.AlertAutoScaleUp,
			AutoScaleDownName:   cloudcommon.AlertAutoScaleDown,
			ScaleUpCpuThresh:    policy.ScaleUpCpuThresh,
			ScaleDownCpuThresh:  policy.ScaleDownCpuThresh,
			TriggerTimeSec:      policy.TriggerTimeSec,
			MaxNodes:            int(policy.MaxNodes),
			MinNodes:            int(policy.MinNodes),
			NodeCountName:       cloudcommon.AlertKeyNodeCount,
			LowCpuNodeCountName: cloudcommon.AlertKeyLowCpuNodeCount,
			MinNodesName:        cloudcommon.AlertKeyMinNodes,
			NodePrefix:          cloudcommon.MexNodePrefix,
		}
		buf := bytes.Buffer{}
		err := t.Execute(&buf, &args)
		if err != nil {
			return nil, err
		}
		policyConfig := &edgeproto.ConfigFile{
			Kind:   k8smgmt.AppConfigHelmYaml,
			Config: buf.String(),
		}
		configs = append(configs, policyConfig)
	}
	return configs, nil
}

// create an appInst as a clustersvc
func createAppInstCommon(ctx context.Context, dialOpts grpc.DialOption, instKey edgeproto.ClusterInstKey, app *edgeproto.App, autoScalePolicy string) error {
	configs, err := getAppInstConfigs(instKey, autoScalePolicy)
	if err != nil {
		return err
	}

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
		Configs: configs,
	}

	res, err := appInstCreateApi(apiClient, appInst)
	if err != nil {
		// Handle non-fatal errors
		if strings.Contains(err.Error(), objstore.ErrKVStoreKeyExists.Error()) {
			log.SpanLog(ctx, log.DebugLevelApi, "appinst already exists", "app", app.String(), "cluster", instKey.String())
			return nil
		}
		if strings.Contains(err.Error(), edgeproto.ErrEdgeApiAppNotFound.Error()) {
			log.SpanLog(ctx, log.DebugLevelApi, "app doesn't exist, create it first", "app", app.String())
			// Create the app
			if err = createAppCommon(ctx, dialOpts, app); err == nil {
				if res, err = appInstCreateApi(apiClient, appInst); err == nil {
					log.SpanLog(ctx, log.DebugLevelApi, "create appinst", "appinst", appInst.String(), "result", res.String())
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
	log.SpanLog(ctx, log.DebugLevelApi, "create appinst", "appinst", appInst.String(), "result", res.String())
	return nil

}

func createMEXPromInst(ctx context.Context, dialOpts grpc.DialOption, instKey edgeproto.ClusterInstKey, autoScalePolicy string) error {
	return createAppInstCommon(ctx, dialOpts, instKey, &MEXPrometheusApp, autoScalePolicy)
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
		ex := promCustomizations{
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

func createAppCommon(ctx context.Context, dialOpts grpc.DialOption, app *edgeproto.App) error {
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
	res, err := apiClient.CreateApp(ctx, app)
	if err != nil {
		// Handle non-fatal errors
		if strings.Contains(err.Error(), objstore.ErrKVStoreKeyExists.Error()) {
			log.SpanLog(ctx, log.DebugLevelApi, "app already exists", "app", app.String())
			return nil
		}
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateApp failed: %s", errstr)
	}
	log.SpanLog(ctx, log.DebugLevelApi, "create app", "app", app.String(), "result", res.String())
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

func updatePrometheusInsts() {
	span := log.StartSpan(log.DebugLevelApi, "updatePrometheusInsts")
	defer span.Finish()
	ctx := log.ContextWithSpan(context.Background(), span)
	conn, err := grpc.Dial(*ctrlAddr, dialOpts, grpc.WithBlock(), grpc.WithWaitForHandshake())
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "Connect to server failed", "server", *ctrlAddr, "error", err.Error())
		return
	}
	defer conn.Close()

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
			log.SpanLog(ctx, log.DebugLevelApi, "Unable to update appinst",
				"appinst", appInst, "error", err.Error())
		}
		log.SpanLog(ctx, log.DebugLevelApi, "update appinst", "appinst", appInst.String(), "result", res.String())
	}
}

// Check if we are running the correct revision of prometheus app, and if not, upgrade it
func validatePrometheusRevision(ctx context.Context) error {
	conn, err := grpc.Dial(*ctrlAddr, dialOpts, grpc.WithBlock(), grpc.WithWaitForHandshake())
	if err != nil {
		return fmt.Errorf("Connect to server %s failed: %s", *ctrlAddr, err.Error())
	}
	defer conn.Close()
	apiClient := edgeproto.NewAppApiClient(conn)

	// Get existing prometheus App
	currentApp, err := getPrometheusAppFromController(ctx, apiClient)
	if err == io.EOF {
		// No app exists yet - just create it when we need to
		log.SpanLog(ctx, log.DebugLevelApi, "app doesn't exist", "app", MEXPrometheusAppKey)
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
		log.SpanLog(ctx, log.DebugLevelApi, "update app", "app", newApp.String(), "result", currentApp.String())
	}
	return nil
}

func main() {
	var err error
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	log.InitTracer(*tlsCertFile)
	defer log.FinishTracer()
	span := log.StartSpan(log.DebugLevelInfo, "main")
	ctx := log.ContextWithSpan(context.Background(), span)

	dialOpts, err = tls.GetTLSClientDialOption(*ctrlAddr, *tlsCertFile, false)
	if err != nil {
		log.FatalLog("get TLS Credentials", "error", err)
	}

	if err = validatePrometheusRevision(ctx); err != nil {
		log.FatalLog("Validate Prometheus version", "error", err)
	}
	// Update prometheus instances in a separate go routine
	go updatePrometheusInsts()

	notifyClient := initNotifyClient(ctx, *notifyAddrs, *tlsCertFile)
	notifyClient.RegisterRecvClusterInstCache(&ClusterInstCache)
	notifyClient.RegisterRecvAutoScalePolicyCache(&AutoScalePolicyCache)
	notifyClient.Start()
	defer notifyClient.Stop()

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	log.SpanLog(ctx, log.DebugLevelMetrics, "Ready")
	span.Finish()
	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}
