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
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	pfutils "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/utils"
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
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var tlsCertFile = flag.String("tls", "", "server tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory")
var externalPorts = flag.String("prometheus-ports", "tcp:9090", "ports to expose in form \"tcp:123,udp:123\"")
var scrapeInterval = flag.Duration("scrapeInterval", time.Second*15, "Metrics collection interval")
var appFlavor = flag.String("flavor", "x1.medium", "App flavor for cluster-svc applications")
var upgradeInstances = flag.Bool("updateAll", false, "Upgrade all Instances of Prometheus operator")
var pluginRequired = flag.Bool("pluginRequired", false, "Require plugin")

var exporterT *template.Template
var prometheusT *template.Template

var clusterSvcPlugin pf.ClusterSvc

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
defaultRules:
  create: true
  rules:
    alertmanager: false
    etcd: false
    general: false
    k8s: true
    kubeApiserver: false
    kubePrometheusNodeAlerting: false
    kubePrometheusNodeRecording: true
    kubernetesAbsent: true
    kubernetesApps: true
    kubernetesResources: true
    kubernetesStorage: true
    kubernetesSystem: true
    kubeScheduler: true
    network: true
    node: true
    prometheus: true
    prometheusOperator: true
    time: true
grafana:
  enabled: false
alertmanager:
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

// Define prometheus operator App.
// Version 7.1.1 tested with helm 2.15 and kubernetes 1.16
var MEXPrometheusApp = edgeproto.App{
	Key:           MEXPrometheusAppKey,
	ImagePath:     "stable/prometheus-operator",
	Deployment:    cloudcommon.AppDeploymentTypeHelm,
	DefaultFlavor: edgeproto.FlavorKey{Name: *appFlavor},
	DelOpt:        edgeproto.DeleteType_AUTO_DELETE,
	InternalPorts: true,
	Annotations:   "version=7.1.1",
}

var dialOpts grpc.DialOption

var sigChan chan os.Signal

var AutoScalePolicyCache edgeproto.AutoScalePolicyCache
var ClusterInstCache edgeproto.ClusterInstCache

type promCustomizations struct {
	Interval string
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
		if err = createMEXPromInst(ctx, dialOpts, new); err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Prometheus-operator inst create failed", "cluster",
				new.Key.ClusterKey.Name, "error", err.Error())
		}
	}
}

func autoScalePolicyCb(ctx context.Context, old *edgeproto.AutoScalePolicy, new *edgeproto.AutoScalePolicy) {
	log.SpanLog(ctx, log.DebugLevelNotify, "autoscalepolicy update", "new", new, "old", old)
	if new == nil || old == nil {
		// deleted, should have been removed from all clusterinsts already
		// or new policy, won't be on any clusterinsts yet
		return
	}
	fields := make(map[string]struct{})
	new.DiffFields(old, fields)
	if len(fields) == 0 {
		return
	}
	// update all prometheus AppInsts on ClusterInsts using the policy
	insts := []edgeproto.ClusterInst{}
	ClusterInstCache.Mux.Lock()
	for k, v := range ClusterInstCache.Objs {
		if new.Key.Developer == k.Developer && new.Key.Name == v.AutoScalePolicy {
			insts = append(insts, *v)
		}
	}
	ClusterInstCache.Mux.Unlock()

	for _, inst := range insts {
		err := createMEXPromInst(ctx, dialOpts, &inst)
		log.SpanLog(ctx, log.DebugLevelApi, "Updated policy", "ClusterInst", inst.Key, "AutoScalePolicy", new.Key.Name, "err", err)
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
	AutoScalePolicyCache.SetUpdatedCb(autoScalePolicyCb)
	log.SpanLog(ctx, log.DebugLevelInfo, "notify client to", "addrs", addrs)
	return notifyClient
}

func appInstCreateApi(ctx context.Context, apiClient edgeproto.AppInstApiClient, appInst edgeproto.AppInst) (*edgeproto.Result, error) {
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

func appInstUpdateApi(ctx context.Context, apiClient edgeproto.AppInstApiClient, appInst *edgeproto.AppInst) (*edgeproto.Result, error) {
	stream, err := apiClient.UpdateAppInst(ctx, appInst)
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

func appInstRefreshApi(ctx context.Context, apiClient edgeproto.AppInstApiClient, appInst *edgeproto.AppInst) (*edgeproto.Result, error) {
	stream, err := apiClient.RefreshAppInst(ctx, appInst)
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

func appInstGetApi(ctx context.Context, apiClient edgeproto.AppInstApiClient, appInst *edgeproto.AppInst) (*edgeproto.AppInst, error) {
	stream, err := apiClient.ShowAppInst(ctx, appInst)
	insts := make([]*edgeproto.AppInst, 0)
	var inst *edgeproto.AppInst
	if err == nil {
		for {
			inst, err = stream.Recv()
			if err == io.EOF {
				err = nil
				break
			}
			if err != nil {
				break
			}
			insts = append(insts, inst)
		}
	}
	if len(insts) == 1 {
		return insts[0], err
	}
	return nil, err
}

// create an appInst as a clustersvc
func createAppInstCommon(ctx context.Context, dialOpts grpc.DialOption, clusterInst *edgeproto.ClusterInst, app *edgeproto.App) error {
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
			ClusterInstKey: clusterInst.Key,
		},
	}
	if clusterSvcPlugin != nil && clusterInst.AutoScalePolicy != "" {
		policy := edgeproto.AutoScalePolicy{}
		policyKey := edgeproto.PolicyKey{}
		policyKey.Developer = clusterInst.Key.Developer
		policyKey.Name = clusterInst.AutoScalePolicy
		if !AutoScalePolicyCache.Get(&policyKey, &policy) {
			return fmt.Errorf("Auto scale policy %s not found for ClusterInst %s", clusterInst.AutoScalePolicy, clusterInst.Key.GetKeyString())
		}
		configs, err := clusterSvcPlugin.GetAppInstConfigs(ctx, clusterInst, &appInst, &policy)
		if err != nil {
			return err
		}
		appInst.Configs = configs
	}

	res, err := appInstCreateApi(ctx, apiClient, appInst)
	if err != nil {
		// Handle non-fatal errors
		if strings.Contains(err.Error(), appInst.Key.ExistsError().Error()) {
			log.SpanLog(ctx, log.DebugLevelApi, "appinst already exists", "app", app.String(), "cluster", clusterInst.Key.String())
			updateExistingAppInst(ctx, apiClient, &appInst)
			return nil
		}
		if strings.Contains(err.Error(), edgeproto.ErrEdgeApiAppNotFound.Error()) {
			log.SpanLog(ctx, log.DebugLevelApi, "app doesn't exist, create it first", "app", app.String())
			// Create the app
			if err = createAppCommon(ctx, dialOpts, app); err == nil {
				if res, err = appInstCreateApi(ctx, apiClient, appInst); err == nil {
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

func createMEXPromInst(ctx context.Context, dialOpts grpc.DialOption, inst *edgeproto.ClusterInst) error {
	return createAppInstCommon(ctx, dialOpts, inst, &MEXPrometheusApp)
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
		if strings.Contains(err.Error(), app.Key.ExistsError().Error()) {
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
	if _, found := fields[edgeproto.AppFieldAnnotations]; found {
		dst.Fields = append(dst.Fields, edgeproto.AppFieldAnnotations)
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
		res, err := appInstRefreshApi(ctx, apiClient, &appInst)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Unable to update appinst",
				"appinst", appInst, "error", err.Error())
		}
		log.SpanLog(ctx, log.DebugLevelApi, "update appinst", "appinst", appInst.String(), "result", res.String())
	}
}

// updates existing AppInst if needed
func updateExistingAppInst(ctx context.Context, apiClient edgeproto.AppInstApiClient, appInst *edgeproto.AppInst) {
	appRef := edgeproto.AppInst{
		Key: appInst.Key,
	}
	existing, err := appInstGetApi(ctx, apiClient, &appRef)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "update check: failed to show AppInst", "key", appInst.Key, "err", err)
		return
	}
	if existing == nil {
		log.SpanLog(ctx, log.DebugLevelApi, "update check: AppInst not found", "key", appInst.Key)
		return
	}
	fields := make(map[string]struct{})
	appInst.DiffFields(existing, fields)
	hasKey := func(field string) bool {
		_, found := fields[field]
		return found
	}
	newFields := []string{}
	if hasKey(edgeproto.AppInstFieldConfigs) ||
		hasKey(edgeproto.AppInstFieldConfigsKind) ||
		hasKey(edgeproto.AppInstFieldConfigsConfig) {
		newFields = append(newFields,
			edgeproto.AppInstFieldConfigs,
			edgeproto.AppInstFieldConfigsKind,
			edgeproto.AppInstFieldConfigsConfig)
	}
	if len(newFields) == 0 {
		log.SpanLog(ctx, log.DebugLevelApi, "update check: no changes")
		return
	}
	// update appinst
	appInst.Fields = newFields
	log.SpanLog(ctx, log.DebugLevelApi, "update check: updating AppInst", "key", appInst.Key, "fields", newFields)
	_, err = appInstUpdateApi(ctx, apiClient, appInst)
	log.SpanLog(ctx, log.DebugLevelApi, "update check: updated AppInst", "key", appInst.Key, "err", err)
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

	clusterSvcPlugin, err = pfutils.GetClusterSvc(ctx, *pluginRequired)
	if err != nil {
		log.FatalLog("get cluster service", "err", err)
	}

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
