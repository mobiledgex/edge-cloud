package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gogo/protobuf/types"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	pfutils "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/utils"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/mobiledgex/edge-cloud/vmspec"
	"google.golang.org/grpc"
)

type CloudletApi struct {
	sync  *Sync
	store edgeproto.CloudletStore
	cache edgeproto.CloudletCache
}

// Vault roles for all services
type VaultRoles struct {
	DmeRoleID    string `json:"dmeroleid"`
	DmeSecretID  string `json:"dmesecretid"`
	CRMRoleID    string `json:"crmroleid"`
	CRMSecretID  string `json:"crmsecretid"`
	CtrlRoleID   string `json:"controllerroleid"`
	CtrlSecretID string `json:"controllersecretid"`
}

var (
	cloudletApi           = CloudletApi{}
	DefaultPlatformFlavor = edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "DefaultPlatformFlavor",
		},
		Vcpus: 2,
		Ram:   4096,
		Disk:  20,
	}
)

const (
	PlatformInitTimeout = 20 * time.Minute
)

type updateCloudletCallback struct {
	in       *edgeproto.Cloudlet
	callback edgeproto.CloudletApi_CreateCloudletServer
}

func (s *updateCloudletCallback) cb(updateType edgeproto.CacheUpdateType, value string) {
	ctx := s.callback.Context()
	switch updateType {
	case edgeproto.UpdateTask:
		log.SpanLog(ctx, log.DebugLevelApi, "SetStatusTask", "key", s.in.Key, "taskName", value)
		s.in.Status.SetTask(value)
		s.callback.Send(&edgeproto.Result{Message: s.in.Status.ToString()})
	case edgeproto.UpdateStep:
		log.SpanLog(ctx, log.DebugLevelApi, "SetStatusStep", "key", s.in.Key, "stepName", value)
		s.in.Status.SetStep(value)
		s.callback.Send(&edgeproto.Result{Message: s.in.Status.ToString()})
	}
}

func ignoreCRMState(cctx *CallContext) bool {
	if cctx.Override == edgeproto.CRMOverride_IGNORE_CRM ||
		cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_AND_TRANSIENT_STATE {
		return true
	}
	return false
}

func InitCloudletApi(sync *Sync) {
	cloudletApi.sync = sync
	cloudletApi.store = edgeproto.NewCloudletStore(sync.store)
	edgeproto.InitCloudletCache(&cloudletApi.cache)
	sync.RegisterCache(&cloudletApi.cache)
}

func (s *CloudletApi) Get(key *edgeproto.CloudletKey, buf *edgeproto.Cloudlet) bool {
	return s.cache.Get(key, buf)
}

func (s *CloudletApi) HasKey(key *edgeproto.CloudletKey) bool {
	return s.cache.HasKey(key)
}

func (s *CloudletApi) ReplaceErrorState(ctx context.Context, in *edgeproto.Cloudlet, newState edgeproto.TrackedState) {
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		inst := edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}

		if inst.State != edgeproto.TrackedState_CREATE_ERROR &&
			inst.State != edgeproto.TrackedState_DELETE_ERROR &&
			inst.State != edgeproto.TrackedState_UPDATE_ERROR {
			return nil
		}
		if newState == edgeproto.TrackedState_NOT_PRESENT {
			s.store.STMDel(stm, &in.Key)
		} else {
			inst.State = newState
			inst.Errors = nil
			s.store.STMPut(stm, &inst)
		}
		return nil
	})
}

func getCrmEnv(vars map[string]string) {
	// For Vault approle, CRM will have its own role/secret
	if val, ok := os.LookupEnv("VAULT_CRM_ROLE_ID"); ok {
		vars["VAULT_ROLE_ID"] = val
	}
	if val, ok := os.LookupEnv("VAULT_CRM_SECRET_ID"); ok {
		vars["VAULT_SECRET_ID"] = val
	}

	for _, key := range []string{
		"GITHUB_ID",
		"VAULT_TOKEN",
		"JAEGER_ENDPOINT",
	} {
		if val, ok := os.LookupEnv(key); ok {
			vars[key] = val
		}
	}
}

func getPlatformConfig(ctx context.Context, cloudlet *edgeproto.Cloudlet) (*edgeproto.PlatformConfig, error) {
	pfConfig := edgeproto.PlatformConfig{}
	pfConfig.PlatformTag = cloudlet.ContainerVersion
	pfConfig.TlsCertFile = nodeMgr.TlsCertFile
	pfConfig.VaultAddr = nodeMgr.VaultAddr
	pfConfig.UseVaultCas = nodeMgr.InternalPki.UseVaultCAs
	pfConfig.UseVaultCerts = nodeMgr.InternalPki.UseVaultCerts
	pfConfig.ContainerRegistryPath = *cloudletRegistryPath
	pfConfig.CloudletVmImagePath = *cloudletVMImagePath
	pfConfig.TestMode = *testMode
	pfConfig.EnvVar = make(map[string]string)
	pfConfig.Region = *region
	pfConfig.CommercialCerts = *commercialCerts
	getCrmEnv(pfConfig.EnvVar)
	addrObjs := strings.Split(*notifyAddr, ":")
	if len(addrObjs) != 2 {
		return nil, fmt.Errorf("unable to fetch notify addr of the controller")
	}
	pfConfig.NotifyCtrlAddrs = *publicAddr + ":" + addrObjs[1]
	pfConfig.Span = log.SpanToString(ctx)

	return &pfConfig, nil
}

func isOperatorInfraCloudlet(in *edgeproto.Cloudlet) bool {
	if !in.DeploymentLocal && in.PlatformType == edgeproto.PlatformType_PLATFORM_TYPE_OPENSTACK {
		return true
	}
	return false
}

func (s *CloudletApi) CreateCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_CreateCloudletServer) error {
	if in.IpSupport == edgeproto.IpSupport_IP_SUPPORT_UNKNOWN {
		in.IpSupport = edgeproto.IpSupport_IP_SUPPORT_DYNAMIC
	}
	// TODO: support static IP assignment.
	if in.IpSupport != edgeproto.IpSupport_IP_SUPPORT_DYNAMIC {
		return errors.New("Only dynamic IPs are supported currently")
	}
	if in.IpSupport == edgeproto.IpSupport_IP_SUPPORT_STATIC {
		// TODO: Validate static ips
	} else {
		// dynamic
		if in.NumDynamicIps < 1 {
			return errors.New("Must specify at least one dynamic public IP available")
		}
	}
	if in.Location.Latitude == 0 && in.Location.Longitude == 0 {
		// user forgot to specify location
		return errors.New("location is missing; 0,0 is not a valid location")
	}

	// If notifysrvaddr is empty, set default value
	if in.NotifySrvAddr == "" {
		in.NotifySrvAddr = "127.0.0.1:0"
	}

	if in.PhysicalName == "" {
		in.PhysicalName = in.Key.Name
		cb.Send(&edgeproto.Result{Message: "Setting physicalname to match cloudlet name"})
	}

	if in.ContainerVersion == "" {
		in.ContainerVersion = *versionTag
	}

	if in.DeploymentLocal {
		if in.AccessVars != nil {
			return errors.New("Access vars is not supported for local deployment")
		}
		if in.VmImageVersion != "" {
			return errors.New("VM Image version is not supported for local deployment")
		}
		if in.PackageVersion != "" {
			return errors.New("Package version is not supported for local deployment")
		}
	}

	if in.VmImageVersion != "" {
		if in.PackageVersion != "" {
			return errors.New("Cannot set package version during creation")
		}
		in.PackageVersion = in.VmImageVersion
	}

	return s.createCloudletInternal(DefCallContext(), in, cb)
}

func (s *CloudletApi) createCloudletInternal(cctx *CallContext, in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_CreateCloudletServer) (reterr error) {
	cctx.SetOverride(&in.CrmOverride)
	ctx := cb.Context()

	defer func() {
		if reterr == nil {
			RecordCloudletEvent(ctx, &in.Key, cloudcommon.CREATED, cloudcommon.InstanceUp)
		}
	}()

	pfConfig, err := getPlatformConfig(ctx, in)
	if err != nil {
		return err
	}

	pfFlavor := edgeproto.Flavor{}
	if in.Flavor.Name == "" {
		in.Flavor = DefaultPlatformFlavor.Key
		pfFlavor = DefaultPlatformFlavor
	}

	accessVars := make(map[string]string)
	if in.AccessVars != nil {
		accessVars = in.AccessVars
		in.AccessVars = nil
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			if !cctx.Undo {
				if in.State == edgeproto.TrackedState_CREATE_ERROR {
					cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous create failed, %v", in.Errors)})
					cb.Send(&edgeproto.Result{Message: "Use DeleteCloudlet to remove and try again"})
				}
				return in.Key.ExistsError()
			}
			in.Errors = nil
		}
		if in.Flavor.Name != "" && in.Flavor.Name != DefaultPlatformFlavor.Key.Name {
			if !flavorApi.store.STMGet(stm, &in.Flavor, &pfFlavor) {
				return fmt.Errorf("Platform Flavor %s not found", in.Flavor.Name)
			}
		}
		err := in.Validate(edgeproto.CloudletAllFieldsMap)
		if err != nil {
			return err
		}

		if ignoreCRMState(cctx) {
			in.State = edgeproto.TrackedState_READY
		} else {
			in.State = edgeproto.TrackedState_CREATING
		}

		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return err
	}

	if ignoreCRMState(cctx) {
		return nil
	}

	var cloudletPlatform pf.Platform
	deleteAccessVars := false
	updatecb := updateCloudletCallback{in, cb}

	if in.DeploymentLocal {
		updatecb.cb(edgeproto.UpdateTask, "Starting CRMServer")
		err = cloudcommon.StartCRMService(ctx, in, pfConfig)
	} else {
		cloudletPlatform, err = pfutils.GetPlatform(ctx, in.PlatformType.String())
		if err == nil {
			if len(accessVars) > 0 {
				err = cloudletPlatform.SaveCloudletAccessVars(ctx, in, accessVars, pfConfig, updatecb.cb)
			}
			if err == nil {
				err = cloudletPlatform.CreateCloudlet(ctx, in, pfConfig, &pfFlavor, updatecb.cb)
				if err != nil && len(accessVars) > 0 {
					deleteAccessVars = true
				}
			}
		}
	}

	if err != nil && cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Create Cloudlet ignoring CRM failure: %s", err.Error())})
		s.ReplaceErrorState(ctx, in, edgeproto.TrackedState_READY)
		cb.Send(&edgeproto.Result{Message: "Created Cloudlet successfully"})
		return nil
	}

	if err == nil {
		// Wait for CRM to connect to controller
		err = s.WaitForCloudlet(
			ctx, &in.Key,
			edgeproto.TrackedState_CREATE_ERROR, // Set error state
			"Created Cloudlet successfully",     // Set success message
			PlatformInitTimeout, updatecb.cb,
		)
	} else {
		cb.Send(&edgeproto.Result{Message: err.Error()})
	}

	if err != nil {
		cb.Send(&edgeproto.Result{Message: "Deleting Cloudlet due to failures"})
		undoErr := s.deleteCloudletInternal(cctx.WithUndo(), in, pfConfig, cb)
		if undoErr != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Undo create Cloudlet", "undoErr", undoErr)
		}
	}
	if deleteAccessVars {
		err1 := cloudletPlatform.DeleteCloudletAccessVars(ctx, in, pfConfig, updatecb.cb)
		if err1 != nil {
			cb.Send(&edgeproto.Result{Message: err1.Error()})
		}
	}
	return err
}

func isVersionConflict(ctx context.Context, localVersion, remoteVersion string) bool {
	if localVersion == "" {
		log.SpanLog(ctx, log.DebugLevelApi, "Ignoring cloudlet validity check as local cloudlet version is missing")
	} else if remoteVersion == "" {
		log.SpanLog(ctx, log.DebugLevelApi, "Ignoring cloudlet validity check as remote cloudlet version is missing")
	} else if localVersion != remoteVersion {
		log.SpanLog(ctx, log.DebugLevelApi, "Ignoring cloudlet info from old cloudlet", "localVersion", localVersion, "remoteVersion", remoteVersion)
		return true
	}
	return false
}

func (s *CloudletApi) UpdateCloudletState(ctx context.Context, key *edgeproto.CloudletKey, newState edgeproto.TrackedState) error {
	cloudlet := edgeproto.Cloudlet{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, key, &cloudlet) {
			return key.NotFoundError()
		}
		cloudlet.State = newState
		s.store.STMPut(stm, &cloudlet)
		return nil
	})
	return err
}

func (s *CloudletApi) WaitForCloudlet(ctx context.Context, key *edgeproto.CloudletKey, errorState edgeproto.TrackedState, successMsg string, timeout time.Duration, updateCallback edgeproto.CacheUpdateCallback) error {
	lastStatusId := uint32(0)
	done := make(chan bool, 1)
	failed := make(chan bool, 1)
	fatal := make(chan bool, 1)

	var err error

	go func() {
		err := cloudcommon.CrmServiceWait(*key)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "failed to cleanup crm service", "err", err)
			fatal <- true
		}
	}()

	log.SpanLog(ctx, log.DebugLevelApi, "wait for cloudlet state", "key", key, "errorState", errorState)

	info := edgeproto.CloudletInfo{}
	if cloudletInfoApi.cache.Get(key, &info) {
		lastStatusId = info.Status.TaskNumber
	}

	checkState := func(key *edgeproto.CloudletKey) {
		cloudlet := edgeproto.Cloudlet{}
		if !cloudletApi.cache.Get(key, &cloudlet) {
			return
		}
		cloudletInfo := edgeproto.CloudletInfo{}
		if !cloudletInfoApi.cache.Get(key, &cloudletInfo) {
			return
		}

		curState := cloudletInfo.State
		localVersion := cloudlet.ContainerVersion
		remoteVersion := cloudletInfo.ContainerVersion

		if curState == edgeproto.CloudletState_CLOUDLET_STATE_ERRORS {
			failed <- true
			return
		}

		if !isVersionConflict(ctx, localVersion, remoteVersion) {
			if curState == edgeproto.CloudletState_CLOUDLET_STATE_READY &&
				(cloudlet.State != edgeproto.TrackedState_UPDATE_REQUESTED && cloudlet.State != edgeproto.TrackedState_CREATING) {
				done <- true
			}
		}
	}

	log.SpanLog(ctx, log.DebugLevelApi, "watch event for CloudletInfo")
	cancel := cloudletInfoApi.cache.WatchKey(key, func(ctx context.Context) {
		if !cloudletInfoApi.cache.Get(key, &info) {
			return
		}

		if info.Status.TaskNumber == 0 {
			lastStatusId = 0
		} else if lastStatusId == info.Status.TaskNumber {
			// got repeat message
			lastStatusId = info.Status.TaskNumber

		} else if info.Status.TaskNumber > lastStatusId {
			if info.Status.StepName != "" {
				updateCallback(edgeproto.UpdateTask, info.Status.StepName)
			} else {
				updateCallback(edgeproto.UpdateTask, info.Status.TaskName)
			}
			lastStatusId = info.Status.TaskNumber
		}
		checkState(key)
	})

	// After setting up watch, check current state,
	// as it may have already changed to target state
	checkState(key)

	select {
	case <-done:
		err = nil
		if successMsg != "" {
			updateCallback(edgeproto.UpdateTask, successMsg)
		}
	case <-failed:
		if cloudletInfoApi.cache.Get(key, &info) {
			errs := strings.Join(info.Errors, ", ")
			err = fmt.Errorf("Encountered failures: %s", errs)
		} else {
			err = fmt.Errorf("Unknown failure")
		}
	case <-fatal:
		out := ""
		out, err = cloudcommon.GetCloudletLog(ctx, key)
		if err != nil || out == "" {
			out = fmt.Sprintf("Please look at %s for more details", cloudcommon.GetCloudletLogFile(key.Name))
		} else {
			out = fmt.Sprintf("Failure: %s", out)
		}
		updateCallback(edgeproto.UpdateTask, out)
		err = errors.New(out)
	case <-time.After(timeout):
		err = fmt.Errorf("Timed out waiting for cloudlet state to be Ready")
		updateCallback(edgeproto.UpdateTask, "platform bringup timed out")
	}

	cancel()
	// note: do not close done/failed, garbage collector will deal with it.

	cloudlet := edgeproto.Cloudlet{}
	err1 := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, key, &cloudlet) {
			return key.NotFoundError()
		}
		cloudlet.AccessVars = nil
		if err == nil {
			cloudlet.Errors = nil
			cloudlet.State = edgeproto.TrackedState_READY
		} else {
			cloudlet.Errors = []string{err.Error()}
			cloudlet.State = errorState
		}

		s.store.STMPut(stm, &cloudlet)
		return nil
	})
	if err1 != nil {
		return err1
	}

	return err
}

type ShowNode struct {
	Data map[string]edgeproto.Node
	grpc.ServerStream
	Ctx context.Context
}

func (x *ShowNode) Send(m *edgeproto.Node) error {
	x.Data[m.GetKey().GetKeyString()] = *m
	return nil
}

func (x *ShowNode) Context() context.Context {
	return x.Ctx
}

func (x *ShowNode) Init() {
	x.Data = make(map[string]edgeproto.Node)
}

func getCloudletVersion(ctx context.Context, key *edgeproto.CloudletKey) (string, error) {
	show := ShowNode{}
	show.Init()
	show.Ctx = ctx
	filter := edgeproto.Node{}
	err := nodeApi.ShowNode(&filter, &show)
	if err != nil {
		return "", fmt.Errorf("Unable to find Cloudlet node: %v", err)
	}
	for _, obj := range show.Data {
		if obj.Key.Type != node.NodeTypeCRM {
			continue
		}
		if obj.Key.CloudletKey != *key {
			continue
		}
		return obj.ContainerVersion, nil
	}
	return "", fmt.Errorf("Unable to find Cloudlet node")
}

func isCloudletUpgradeRequired(ctx context.Context, cloudlet *edgeproto.Cloudlet) error {
	cloudletVersion, err := getCloudletVersion(ctx, &cloudlet.Key)
	if err != nil {
		return fmt.Errorf("unable to fetch cloudlet version: %v", err)
	}

	if cloudletVersion == "" {
		return nil
	}

	ctrl_vers, err := util.ContainerVersionParse(*versionTag)
	if err != nil {
		return err
	}

	cloudlet_vers, err := util.ContainerVersionParse(cloudletVersion)
	if err != nil {
		return err
	}

	new_vers, err := util.ContainerVersionParse(cloudlet.ContainerVersion)
	if err != nil {
		return err
	}

	diff := ctrl_vers.Sub(*new_vers)
	if diff > 0 {
		return fmt.Errorf("cannot upgrade cloudlet to a version below controller version %s", *versionTag)
	}

	diff = cloudlet_vers.Sub(*new_vers)
	if diff > 0 {
		return fmt.Errorf("downgrade from version %s to %s is not supported", cloudlet_vers, new_vers)
	} else if diff < 0 {
		return nil
	} else {
		// Allow users to retry upgrade to same version on an update error
		if cloudlet.State != edgeproto.TrackedState_UPDATE_ERROR {
			return fmt.Errorf("no upgrade required, cloudlet is already of version %s", cloudletVersion)
		}
		return nil
	}
}

func (s *CloudletApi) UpdateCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_UpdateCloudletServer) error {
	ctx := cb.Context()
	fmap := edgeproto.MakeFieldMap(in.Fields)
	if _, found := fmap[edgeproto.CloudletFieldNumDynamicIps]; found {
		staticSet := false
		if _, staticFound := fmap[edgeproto.CloudletFieldIpSupport]; staticFound {
			if in.IpSupport == edgeproto.IpSupport_IP_SUPPORT_STATIC {
				staticSet = true
			}
		}
		if in.NumDynamicIps < 1 && !staticSet {
			return errors.New("Cannot specify less than one dynamic IP unless Ip Support Static is specified")
		}
	}

	err := in.Validate(fmap)
	if err != nil {
		return err
	}

	cur := &edgeproto.Cloudlet{}
	if !cloudletApi.cache.Get(&in.Key, cur) {
		return in.Key.NotFoundError()
	}
	cur.CopyInFields(in)

	upgrade := false
	if _, found := fmap[edgeproto.CloudletFieldContainerVersion]; found {
		err = isCloudletUpgradeRequired(ctx, cur)
		if err != nil {
			return err
		}
		// verify if image is available in registry
		registry_path := *cloudletRegistryPath + ":" + in.ContainerVersion
		err = cloudcommon.ValidateDockerRegistryPath(ctx, registry_path, vaultConfig)
		if err != nil {
			if *testMode {
				log.SpanLog(ctx, log.DebugLevelInfo, "Failed to validate cloudlet registry path", "err", err)
			} else {
				return err
			}
		}
		upgrade = true
	}

	if _, found := fmap[edgeproto.CloudletFieldVmImageVersion]; found {
		return errors.New("Cloudlet VM baseimage version upgrade is not yet supported")
	}

	if _, found := fmap[edgeproto.CloudletFieldPackageVersion]; found {
		err = util.ValidateImageVersion(in.PackageVersion)
		if err != nil {
			return err
		}
		upgrade = true
	}

	accessVars := make(map[string]string)
	if _, found := fmap[edgeproto.CloudletFieldAccessVars]; found {
		if !upgrade {
			return fmt.Errorf("Access vars can only be updated during upgrade")
		}
		if in.DeploymentLocal {
			return errors.New("Access vars is not supported for local deployment")
		}
		accessVars = in.AccessVars
		in.AccessVars = nil
	}

	cctx := DefCallContext()
	cctx.SetOverride(&in.CrmOverride)

	if !ignoreCRMState(cctx) && upgrade {
		if in.DeploymentLocal {
			return fmt.Errorf("upgrade is not supported for local deployments")
		}
		updatecb := updateCloudletCallback{in, cb}
		var cloudletPlatform pf.Platform
		cloudletPlatform, err := pfutils.GetPlatform(ctx, in.PlatformType.String())
		if err != nil {
			return err
		}
		pfConfig, err := getPlatformConfig(ctx, in)
		if err != nil {
			return err
		}
		if len(accessVars) > 0 {
			err = cloudletPlatform.SaveCloudletAccessVars(ctx, in, accessVars, pfConfig, updatecb.cb)
			if err != nil {
				return err
			}
		}
		err = s.UpgradeCloudlet(ctx, in, cb)
		if err != nil {
			if cloudletPlatform != nil && len(accessVars) > 0 {
				err1 := cloudletPlatform.DeleteCloudletAccessVars(ctx, in, pfConfig, updatecb.cb)
				if err1 != nil {
					cb.Send(&edgeproto.Result{Message: err1.Error()})
				}
			}
			return err
		}
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, cur) {
			return in.Key.NotFoundError()
		}
		cur.CopyInFields(in)
		// In case we need to set TrackedState to ready
		// maybe after a manual upgrade of cloudlet
		if ignoreCRMState(cctx) {
			cur.State = edgeproto.TrackedState_READY
		}
		if cur.State == edgeproto.TrackedState_READY {
			cur.Errors = nil
		}
		s.store.STMPut(stm, cur)
		return nil
	})

	if err != nil {
		return err
	}

	// after the cloudlet change is committed, if the location changed,
	// update app insts as well.
	s.UpdateAppInstLocations(ctx, in)
	return nil
}

func (s *CloudletApi) UpgradeCloudlet(ctx context.Context, in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_UpdateCloudletServer) (reterr error) {
	updatecb := updateCloudletCallback{in, cb}

	if appInstApi.UsingCloudlet(&in.Key) {
		return fmt.Errorf("AppInst is busy on the cloudlet")
	}
	if clusterInstApi.UsingCloudlet(&in.Key) {
		return fmt.Errorf("ClusterInst is busy on the cloudlet")
	}

	if err := cloudletInfoApi.checkCloudletReady(&in.Key); err != nil {
		if in.State == edgeproto.TrackedState_UPDATE_ERROR {
			info := edgeproto.CloudletInfo{}
			if cloudletInfoApi.cache.Get(&in.Key, &info) &&
				info.State == edgeproto.CloudletState_CLOUDLET_STATE_READY {
				// Allow upgrade even on UPDATE_ERROR, so that end-users can retry upgrade
			} else {
				return err
			}
		} else {
			return err
		}
	}

	log.SpanLog(ctx, log.DebugLevelApi, "fetch platform config")
	pfConfig, err := getPlatformConfig(ctx, in)
	if err != nil {
		return err
	}

	// cleanup old crms post upgrade
	pfConfig.CleanupMode = true
	cloudlet := &edgeproto.Cloudlet{}
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, cloudlet) {
			return in.Key.NotFoundError()
		}
		cloudlet.CopyInFields(in)
		cloudlet.Config = *pfConfig
		cloudlet.State = edgeproto.TrackedState_UPDATE_REQUESTED

		s.store.STMPut(stm, cloudlet)
		return nil
	})
	if err != nil {
		return err
	}

	RecordCloudletEvent(ctx, &in.Key, cloudcommon.UPDATE_START, cloudcommon.InstanceDown)
	defer func() {
		info := edgeproto.CloudletInfo{}
		if cloudletInfoApi.cache.Get(&in.Key, &info) {
			if reterr == nil {
				RecordCloudletEvent(ctx, &in.Key, cloudcommon.UPDATE_COMPLETE, cloudcommon.InstanceUp)
			} else if info.State == edgeproto.CloudletState_CLOUDLET_STATE_READY { // update failed but cloudlet is still up
				RecordCloudletEvent(ctx, &in.Key, cloudcommon.UPDATE_ERROR, cloudcommon.InstanceUp)
			} else { // error and cloudlet went down
				RecordCloudletEvent(ctx, &in.Key, cloudcommon.UPDATE_ERROR, cloudcommon.InstanceDown)
			}
		}
	}()

	updatecb.cb(edgeproto.UpdateTask, "Upgrading Cloudlet")

	// Wait for cloudlet to finish upgrading
	err = s.WaitForCloudlet(
		ctx, &in.Key,
		edgeproto.TrackedState_UPDATE_ERROR, // Set error state
		"Upgraded Cloudlet successfully",    // Set success message
		PlatformInitTimeout, updatecb.cb,
	)
	return err
}

func (s *CloudletApi) DeleteCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_DeleteCloudletServer) error {
	pfConfig, err := getPlatformConfig(cb.Context(), in)
	if err != nil {
		return err
	}
	return s.deleteCloudletInternal(DefCallContext(), in, pfConfig, cb)
}

func (s *CloudletApi) deleteCloudletInternal(cctx *CallContext, in *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, cb edgeproto.CloudletApi_DeleteCloudletServer) (reterr error) {
	ctx := cb.Context()

	defer func() {
		if reterr == nil {
			RecordCloudletEvent(ctx, &in.Key, cloudcommon.DELETED, cloudcommon.InstanceDown)
		}
	}()

	dynInsts := make(map[edgeproto.AppInstKey]struct{})
	if appInstApi.UsesCloudlet(&in.Key, dynInsts) {
		// disallow delete if static instances are present
		return errors.New("Cloudlet in use by static AppInst")
	}

	clDynInsts := make(map[edgeproto.ClusterInstKey]struct{})
	if clusterInstApi.UsesCloudlet(&in.Key, clDynInsts) {
		return errors.New("Cloudlet in use by static ClusterInst")
	}

	cctx.SetOverride(&in.CrmOverride)

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			return in.Key.NotFoundError()
		}
		if ignoreCRMState(cctx) {
			// delete happens later, this STM just checks for existence
			return nil
		}
		if !cctx.Undo {
			if in.State == edgeproto.TrackedState_CREATE_REQUESTED ||
				in.State == edgeproto.TrackedState_CREATING ||
				in.State == edgeproto.TrackedState_UPDATE_REQUESTED ||
				in.State == edgeproto.TrackedState_UPDATING {
				return errors.New("Cloudlet busy, cannot be deleted")
			}
			if in.State == edgeproto.TrackedState_DELETE_ERROR &&
				cctx.Override != edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
				cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous delete failed, %v", in.Errors)})
				cb.Send(&edgeproto.Result{Message: "Use CreateCloudlet to rebuild, and try again"})
			}
			if in.State == edgeproto.TrackedState_DELETE_REQUESTED ||
				in.State == edgeproto.TrackedState_DELETING ||
				in.State == edgeproto.TrackedState_DELETE_PREPARE {
				return errors.New("Cloudlet busy, already under deletion")
			}
		}
		in.State = edgeproto.TrackedState_DELETE_PREPARE
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return err
	}

	if !ignoreCRMState(cctx) {
		updatecb := updateCloudletCallback{in, cb}

		if in.DeploymentLocal {
			updatecb.cb(edgeproto.UpdateTask, "Stopping CRMServer")
			err = cloudcommon.StopCRMService(ctx, in)
		} else {
			var cloudletPlatform pf.Platform
			cloudletPlatform, err = pfutils.GetPlatform(ctx, in.PlatformType.String())
			if err == nil {
				err = cloudletPlatform.DeleteCloudlet(ctx, in, pfConfig, updatecb.cb)
			}
		}
		if err != nil && cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
			cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Delete Cloudlet ignoring CRM failure: %s", err.Error())})
			s.ReplaceErrorState(ctx, in, edgeproto.TrackedState_NOT_PRESENT)
			err = nil
		}
	}

	updateCloudlet := edgeproto.Cloudlet{}
	err1 := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &updateCloudlet) {
			return in.Key.NotFoundError()
		}
		if err != nil {
			updateCloudlet.State = edgeproto.TrackedState_DELETE_ERROR
			s.store.STMPut(stm, &updateCloudlet)
			return nil
		}
		s.store.STMDel(stm, &in.Key)
		cloudletRefsApi.store.STMDel(stm, &in.Key)
		cb.Send(&edgeproto.Result{Message: "Deleted Cloudlet successfully"})
		return nil
	})
	if err1 != nil {
		return err1
	}

	if err != nil {
		return err
	}

	// also delete associated info
	// Note: don't delete cloudletinfo, that will get deleted once CRM
	// disconnects. Otherwise if admin deletes/recreates Cloudlet with
	// CRM connected the whole time, we will end up without cloudletInfo.
	// also delete dynamic instances
	if len(dynInsts) > 0 {
		// delete dynamic instances
		for key, _ := range dynInsts {
			appInst := edgeproto.AppInst{Key: key}
			derr := appInstApi.deleteAppInstInternal(DefCallContext(), &appInst, cb)
			if derr != nil {
				log.DebugLog(log.DebugLevelApi,
					"Failed to delete dynamic app inst",
					"key", key, "err", derr)
			}
		}
	}
	if len(clDynInsts) > 0 {
		for key, _ := range clDynInsts {
			clInst := edgeproto.ClusterInst{Key: key}
			derr := clusterInstApi.deleteClusterInstInternal(DefCallContext(), &clInst, cb)
			if derr != nil {
				log.DebugLog(log.DebugLevelApi,
					"Failed to delete dynamic ClusterInst",
					"key", key, "err", derr)
			}
		}
	}
	cloudletPoolMemberApi.cloudletDeleted(ctx, &in.Key)
	cloudletInfoApi.cleanupCloudletInfo(ctx, &in.Key)
	return nil
}

func (s *CloudletApi) ShowCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_ShowCloudletServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Cloudlet) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *CloudletApi) RemoveCloudletResMapping(ctx context.Context, in *edgeproto.CloudletResMap) (*edgeproto.Result, error) {
	var err error
	cl := edgeproto.Cloudlet{}
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &cl) {
			return in.Key.NotFoundError()
		}

		for resource, _ := range in.Mapping {
			delete(cl.ResTagMap, resource)
		}
		s.store.STMPut(stm, &cl)
		return err
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletApi) AddCloudletResMapping(ctx context.Context, in *edgeproto.CloudletResMap) (*edgeproto.Result, error) {

	var err error
	cl := edgeproto.Cloudlet{}
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &cl) {
			return in.Key.NotFoundError()
		} else {
			if cl.ResTagMap == nil {
				cl.ResTagMap = make(map[string]*edgeproto.ResTagTableKey)
			}
		}

		return err
	})
	if err != nil {
		return &edgeproto.Result{}, err
	}

	for resource, tblname := range in.Mapping {
		if valerr, ok := resTagTableApi.ValidateResName(resource); !ok {
			return &edgeproto.Result{}, valerr
		}
		resource = strings.ToLower(resource)
		var key edgeproto.ResTagTableKey
		key.Name = tblname
		key.Organization = in.Key.Organization
		tbl, err := resTagTableApi.GetResTagTable(ctx, &key)

		if err != nil && err.Error() == key.NotFoundError().Error() {
			// auto-create empty
			tbl.Key = key
			_, err = resTagTableApi.CreateResTagTable(ctx, tbl)
			if err != nil {
				return &edgeproto.Result{}, err
			}
		}
		cl.ResTagMap[resource] = &key
	}

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &cl) {
			return in.Key.NotFoundError()
		}
		for resource, tblname := range in.Mapping {
			key := edgeproto.ResTagTableKey{
				Name:         tblname,
				Organization: in.Key.Organization,
			}
			cl.ResTagMap[resource] = &key
		}
		s.store.STMPut(stm, &cl)
		return err
	})

	return &edgeproto.Result{}, err
}

func (s *CloudletApi) UpdateAppInstLocations(ctx context.Context, in *edgeproto.Cloudlet) {
	fmap := edgeproto.MakeFieldMap(in.Fields)
	if _, found := fmap[edgeproto.CloudletFieldLocation]; !found {
		// no location fields updated
		return
	}

	// find all appinsts associated with the cloudlet
	keys := make([]edgeproto.AppInstKey, 0)
	appInstApi.cache.Mux.Lock()
	for _, data := range appInstApi.cache.Objs {
		inst := data.Obj
		if inst.Key.ClusterInstKey.CloudletKey.Matches(&in.Key) {
			keys = append(keys, inst.Key)
		}
	}
	appInstApi.cache.Mux.Unlock()

	inst := edgeproto.AppInst{}
	for ii, _ := range keys {
		inst = *appInstApi.cache.Objs[keys[ii]].Obj
		inst.Fields = make([]string, 0)
		if _, found := fmap[edgeproto.CloudletFieldLocationLatitude]; found {
			inst.CloudletLoc.Latitude = in.Location.Latitude
			inst.Fields = append(inst.Fields, edgeproto.AppInstFieldCloudletLocLatitude)
		}
		if _, found := fmap[edgeproto.CloudletFieldLocationLongitude]; found {
			inst.CloudletLoc.Longitude = in.Location.Longitude
			inst.Fields = append(inst.Fields, edgeproto.AppInstFieldCloudletLocLongitude)
		}
		if len(inst.Fields) == 0 {
			break
		}

		err := appInstApi.updateAppInstStore(ctx, &inst)
		if err != nil {
			log.DebugLog(log.DebugLevelApi, "Update AppInst Location",
				"inst", inst, "err", err)
		}
	}
}

func (s *CloudletApi) showCloudletsByKeys(keys map[edgeproto.CloudletKey]struct{}, cb func(obj *edgeproto.Cloudlet) error) error {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()

	for key, data := range s.cache.Objs {
		obj := data.Obj
		if _, found := keys[key]; !found {
			continue
		}
		err := cb(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *CloudletApi) FindFlavorMatch(ctx context.Context, in *edgeproto.FlavorMatch) (*edgeproto.FlavorMatch, error) {

	cl := edgeproto.Cloudlet{}
	var spec *vmspec.VMCreationSpec
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {

		if !cloudletApi.store.STMGet(stm, &in.Key, &cl) {
			return in.Key.NotFoundError()
		}
		cli := edgeproto.CloudletInfo{}
		if !cloudletInfoApi.store.STMGet(stm, &in.Key, &cli) {
			return in.Key.NotFoundError()
		}
		mexFlavor := edgeproto.Flavor{}
		mexFlavor.Key.Name = in.FlavorName
		if !flavorApi.store.STMGet(stm, &mexFlavor.Key, &mexFlavor) {
			return in.Key.NotFoundError()
		}
		var verr error
		spec, verr = resTagTableApi.GetVMSpec(ctx, stm, mexFlavor, cl, cli)
		if verr != nil {
			return verr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	in.FlavorName = spec.FlavorName
	in.AvailabilityZone = spec.AvailabilityZone
	return in, nil
}

func RecordCloudletEvent(ctx context.Context, cloudletKey *edgeproto.CloudletKey, event cloudcommon.InstanceEvent, serverStatus string) {
	metric := edgeproto.Metric{}
	metric.Name = cloudcommon.CloudletEvent
	ts, _ := types.TimestampProto(time.Now())
	metric.Timestamp = *ts
	metric.AddStringVal("cloudletorg", cloudletKey.Organization)
	metric.AddTag("cloudlet", cloudletKey.Name)
	metric.AddStringVal("event", string(event))
	metric.AddStringVal("status", serverStatus)

	services.events.AddMetric(&metric)
}
