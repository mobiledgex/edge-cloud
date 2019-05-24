package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AppInstApi struct {
	sync  *Sync
	store edgeproto.AppInstStore
	cache edgeproto.AppInstCache
}

const RootLBSharedPortBegin int32 = 10000

var appInstApi = AppInstApi{}

// TODO: these timeouts should be adjust based on target platform,
// as some platforms (azure, etc) may take much longer.
// These timeouts should be at least long enough for the controller and
// CRM to exchange an initial set of messages (i.e. 10 sec or so).
var CreateAppInstTimeout = 30 * time.Minute
var UpdateAppInstTimeout = 20 * time.Minute
var DeleteAppInstTimeout = 20 * time.Minute

// Transition states indicate states in which the CRM is still busy.
var CreateAppInstTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_CREATING: struct{}{},
}
var UpdateAppInstTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_UPDATING: struct{}{},
}
var DeleteAppInstTransitions = map[edgeproto.TrackedState]struct{}{
	edgeproto.TrackedState_DELETING: struct{}{},
}

func InitAppInstApi(sync *Sync) {
	appInstApi.sync = sync
	appInstApi.store = edgeproto.NewAppInstStore(sync.store)
	edgeproto.InitAppInstCache(&appInstApi.cache)
	sync.RegisterCache(&appInstApi.cache)
	if *shortTimeouts {
		CreateAppInstTimeout = 3 * time.Second
		UpdateAppInstTimeout = 2 * time.Second
		DeleteAppInstTimeout = 2 * time.Second
	}
}

func (s *AppInstApi) GetAllKeys(keys map[edgeproto.AppInstKey]struct{}) {
	s.cache.GetAllKeys(keys)
}

func (s *AppInstApi) Get(key *edgeproto.AppInstKey, val *edgeproto.AppInst) bool {
	return s.cache.Get(key, val)
}

func (s *AppInstApi) HasKey(key *edgeproto.AppInstKey) bool {
	return s.cache.HasKey(key)
}

func (s *AppInstApi) UsesCloudlet(in *edgeproto.CloudletKey, dynInsts map[edgeproto.AppInstKey]struct{}) bool {
	var app edgeproto.App
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	static := false
	for key, val := range s.cache.Objs {
		if key.ClusterInstKey.CloudletKey.Matches(in) && appApi.Get(&val.Key.AppKey, &app) {
			if (val.Liveness == edgeproto.Liveness_LIVENESS_STATIC) && (app.DelOpt == edgeproto.DeleteType_NO_AUTO_DELETE) {
				static = true
				//if can autodelete it then also add it to the dynInsts to be deleted later
			} else if (val.Liveness == edgeproto.Liveness_LIVENESS_DYNAMIC) || (app.DelOpt == edgeproto.DeleteType_AUTO_DELETE) {
				dynInsts[key] = struct{}{}
			}
		}
	}
	return static
}

func (s *AppInstApi) UsesApp(in *edgeproto.AppKey, dynInsts map[edgeproto.AppInstKey]struct{}) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	static := false
	for key, val := range s.cache.Objs {
		if key.AppKey.Matches(in) {
			if val.Liveness == edgeproto.Liveness_LIVENESS_STATIC {
				static = true
			} else if val.Liveness == edgeproto.Liveness_LIVENESS_DYNAMIC {
				dynInsts[key] = struct{}{}
			}
		}
	}
	return static
}

func (s *AppInstApi) UsesClusterInst(in *edgeproto.ClusterInstKey) bool {
	var app edgeproto.App
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for key, val := range s.cache.Objs {
		if key.ClusterInstKey.Matches(in) && appApi.Get(&val.Key.AppKey, &app) {
			log.DebugLog(log.DebugLevelApi, "AppInst found for clusterInst", "app", app.Key.Name,
				"autodelete", app.DelOpt.String())
			if app.DelOpt == edgeproto.DeleteType_NO_AUTO_DELETE {
				return true
			}
		}
	}
	return false
}

func (s *AppInstApi) AutoDeleteAppInsts(key *edgeproto.ClusterInstKey, cb edgeproto.ClusterInstApi_DeleteClusterInstServer) error {
	var app edgeproto.App
	var err error
	apps := make(map[edgeproto.AppInstKey]*edgeproto.AppInst)
	log.DebugLog(log.DebugLevelApi, "Auto-deleting appinsts ", "cluster", key.ClusterKey.Name)
	s.cache.Mux.Lock()
	for k, val := range s.cache.Objs {
		if k.ClusterInstKey.Matches(key) && appApi.Get(&val.Key.AppKey, &app) {
			if app.DelOpt == edgeproto.DeleteType_AUTO_DELETE {
				apps[k] = val
			}
		}
	}
	s.cache.Mux.Unlock()

	//Spin in case cluster was just created and apps are still in the creation process and cannot be deleted
	var spinTime time.Duration
	start := time.Now()
	for _, val := range apps {
		log.DebugLog(log.DebugLevelApi, "Auto-deleting appinst ", "appinst", val.Key.AppKey.Name)
		cb.Send(&edgeproto.Result{Message: "Autodeleting appinst " + val.Key.AppKey.Name})
		for {
			err = s.deleteAppInstInternal(DefCallContext(), val, cb)
			if err != nil && err.Error() == "AppInst busy, cannot delete" {
				spinTime = time.Since(start)
				if spinTime > DeleteAppInstTimeout {
					log.DebugLog(log.DebugLevelApi, "Timeout while waiting for app", "appName", val.Key.AppKey.Name)
					return err
				}
				log.DebugLog(log.DebugLevelApi, "Appinst busy, retrying in 0.5s...", "appName", val.Key.AppKey.Name)
				time.Sleep(500 * time.Millisecond)
			} else { //if its anything other than an appinst busy error, break out of the spin
				break
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *AppInstApi) UsesFlavor(key *edgeproto.FlavorKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, app := range s.cache.Objs {
		if app.Flavor.Matches(key) {
			return true
		}
	}
	return false
}

func (s *AppInstApi) CreateAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_CreateAppInstServer) error {
	in.Liveness = edgeproto.Liveness_LIVENESS_STATIC
	return s.createAppInstInternal(DefCallContext(), in, cb)
}

func getProtocolBitMap(proto dme.LProto) (int32, error) {
	var bitmap int32
	switch proto {
	//put all "TCP" protocols below here
	case dme.LProto_L_PROTO_HTTP:
		fallthrough
	case dme.LProto_L_PROTO_TCP:
		bitmap = 1 //01
		break
	//put all "UDP" protocols below here
	case dme.LProto_L_PROTO_UDP:
		bitmap = 2 //10
		break
	default:
		return 0, errors.New("Unknown protocol in use for this app")
	}
	return bitmap, nil
}

func protocolInUse(protocolsToCheck int32, usedProtocols int32) bool {
	return (protocolsToCheck & usedProtocols) != 0
}

func addProtocol(protos int32, protocolToAdd int32) int32 {
	return protos | protocolToAdd
}

func removeProtocol(protos int32, protocolToRemove int32) int32 {
	return protos & (^protocolToRemove)
}

// createAppInstInternal is used to create dynamic app insts internally,
// bypassing static assignment.
func (s *AppInstApi) createAppInstInternal(cctx *CallContext, in *edgeproto.AppInst, cb edgeproto.AppInstApi_CreateAppInstServer) (reterr error) {
	if in.Liveness == edgeproto.Liveness_LIVENESS_UNKNOWN {
		in.Liveness = edgeproto.Liveness_LIVENESS_DYNAMIC
	}
	cctx.SetOverride(&in.CrmOverride)
	if !ignoreCRM(cctx) {
		if err := cloudletInfoApi.checkCloudletReady(&in.Key.ClusterInstKey.CloudletKey); err != nil {
			return err
		}
	}

	var autocluster bool
	// See if we need to create auto-cluster.
	// This also sets up the correct ClusterInstKey in "in".

	// indicates special default cloudlet maintained by the developer
	var defaultCloudlet bool

	if err := in.Key.AppKey.Validate(); err != nil {
		return err
	}

	if in.Key.ClusterInstKey.CloudletKey == cloudcommon.DefaultCloudletKey {
		log.DebugLog(log.DebugLevelApi, "special default public cloud case", "appinst", in)
		defaultCloudlet = true
		if in.Key.ClusterInstKey.Developer == "" {
			in.Key.ClusterInstKey.Developer = in.Key.AppKey.DeveloperKey.Name
		}
	}
	if in.Key.ClusterInstKey.Developer == "" {
		// This is allowed for:
		// 1) older clusters that were created before it was required
		// 2) autoclusters
		// We do ClusterInst lookup for both cases below.
	} else if in.Key.AppKey.DeveloperKey.Name != cloudcommon.DeveloperMobiledgeX &&
		in.Key.AppKey.DeveloperKey.Name != in.Key.ClusterInstKey.Developer {
		// both are specified, make sure they match
		return fmt.Errorf("Developer name mismatch between app: %s and cluster inst: %s", in.Key.AppKey.DeveloperKey.Name, in.Key.ClusterInstKey.Developer)
	}

	if !defaultCloudlet {
		err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
			autocluster = false
			if s.store.STMGet(stm, &in.Key, in) {
				if !cctx.Undo && in.State != edgeproto.TrackedState_DELETE_ERROR && !ignoreTransient(cctx, in.State) {
					if in.State == edgeproto.TrackedState_CREATE_ERROR {
						cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous create failed, %v", in.Errors)})
						cb.Send(&edgeproto.Result{Message: "Use DeleteAppInst to remove and try again"})
					}
					return objstore.ErrKVStoreKeyExists
				}
				in.Errors = nil
				if !defaultCloudlet {
					// must reset Uri
					in.Uri = ""
				}
			} else {
				err := in.Validate(edgeproto.AppInstAllFieldsMap)
				if err != nil {
					return err
				}
			}
			// make sure cloudlet exists
			if !cloudletApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, nil) {
				return errors.New("Specified cloudlet not found")
			}

			cikey := &in.Key.ClusterInstKey
			// Explicit auto-cluster requirement
			if cikey.ClusterKey.Name == "" {
				return fmt.Errorf("No cluster name specified. Create one first or use \"%s\" as the name to automatically create a ClusterInst", ClusterAutoPrefix)
			}
			var app edgeproto.App
			if !appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
				return edgeproto.ErrEdgeApiAppNotFound
			}
			// Check if specified ClusterInst exists
			if !strings.HasPrefix(cikey.ClusterKey.Name, ClusterAutoPrefix) && cloudcommon.IsClusterInstReqd(&app) {
				found := clusterInstApi.store.STMGet(stm, &in.Key.ClusterInstKey, nil)
				if !found && in.Key.ClusterInstKey.Developer == "" {
					// developer may not be specified
					// in clusterinst.
					in.Key.ClusterInstKey.Developer = in.Key.AppKey.DeveloperKey.Name
					found = clusterInstApi.store.STMGet(stm, &in.Key.ClusterInstKey, nil)
					if found {
						cb.Send(&edgeproto.Result{Message: "Setting ClusterInst developer to match App developer"})
					}
				}
				if !found {
					return errors.New("Specified ClusterInst not found")
				}
				// cluster inst exists so we're good.
				return nil
			}
			if cloudcommon.IsClusterInstReqd(&app) {
				// Auto-cluster
				if clusterInstApi.store.STMGet(stm, &in.Key.ClusterInstKey, nil) {
					// if it already exists, this means we just want to spawn more apps into it
					return nil
				}
				cikey.ClusterKey.Name = util.K8SSanitize(cikey.ClusterKey.Name)
				if cikey.Developer == "" {
					cikey.Developer = in.Key.AppKey.DeveloperKey.Name
				}
				autocluster = true
			}

			if in.Flavor.Name == "" {
				// find flavor from app
				in.Flavor = app.DefaultFlavor
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	if autocluster {
		// auto-create cluster inst
		clusterInst := edgeproto.ClusterInst{}
		clusterInst.Key = in.Key.ClusterInstKey
		clusterInst.Auto = true
		log.DebugLog(log.DebugLevelApi,
			"Create auto-clusterinst",
			"key", clusterInst.Key,
			"appinst", in)

		clusterInst.Flavor = in.Flavor
		clusterInst.IpAccess = in.AutoClusterIpAccess
		clusterInst.NumMasters = 1
		clusterInst.NumNodes = 1 // TODO support 1 master, zero nodes
		err := clusterInstApi.createClusterInstInternal(cctx, &clusterInst, cb)
		if err != nil {
			return err
		}
		defer func() {
			if reterr != nil && !cctx.Undo {
				cb.Send(&edgeproto.Result{Message: "Deleting auto-ClusterInst due to failure"})
				undoErr := clusterInstApi.deleteClusterInstInternal(cctx.WithUndo(), &clusterInst, cb)
				if undoErr != nil {
					log.DebugLog(log.DebugLevelApi,
						"Undo create auto-clusterinst failed",
						"key", clusterInst.Key,
						"undoErr", undoErr)
				}
			}
		}()
	}

	// these checks cannot be in ApplySTMWait funcs since
	// that func may run several times if the database changes while
	// it is running, and the stm func modifies in.Uri.
	if in.Uri == "" && defaultCloudlet {
		return errors.New("URI (Public Fqdn) is required for default cloudlet")
	} else if in.Uri != "" && !defaultCloudlet {
		return fmt.Errorf("Cannot specify URI %s for non-default cloudlet", in.Uri)
	}

	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		buf := in
		if !defaultCloudlet {
			// lookup already done, don't overwrite changes
			buf = nil
		}
		if s.store.STMGet(stm, &in.Key, buf) {
			if !cctx.Undo && in.State != edgeproto.TrackedState_DELETE_ERROR {
				if in.State == edgeproto.TrackedState_CREATE_ERROR {
					cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Previous create failed, %v", in.Errors)})
					cb.Send(&edgeproto.Result{Message: "Use DeleteAppInst to remove and try again"})
				}
				return objstore.ErrKVStoreKeyExists
			}
			in.Errors = nil
		} else {
			err := in.Validate(edgeproto.AppInstAllFieldsMap)
			if err != nil {
				return err
			}
		}

		// cache location of cloudlet in app inst
		var cloudlet edgeproto.Cloudlet

		if !defaultCloudlet {
			if !cloudletApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &cloudlet) {
				return errors.New("Specified cloudlet not found")
			}
			in.CloudletLoc = cloudlet.Location
		} else {
			in.CloudletLoc = dme.Loc{}
		}

		var app edgeproto.App
		if !appApi.store.STMGet(stm, &in.Key.AppKey, &app) {
			return edgeproto.ErrEdgeApiAppNotFound
		}

		if in.Flavor.Name == "" {
			in.Flavor = app.DefaultFlavor
		}

		if !flavorApi.store.STMGet(stm, &in.Flavor, nil) {
			return fmt.Errorf("Flavor %s not found", in.Flavor.Name)
		}

		var clusterKey *edgeproto.ClusterKey
		ipaccess := edgeproto.IpAccess_IP_ACCESS_SHARED
		if !defaultCloudlet && cloudcommon.IsClusterInstReqd(&app) {
			clusterInst := edgeproto.ClusterInst{}
			if !clusterInstApi.store.STMGet(stm, &in.Key.ClusterInstKey, &clusterInst) {
				return errors.New("Cluster instance does not exist for app")
			}
			if clusterInst.State != edgeproto.TrackedState_READY {
				return fmt.Errorf("ClusterInst %s not ready", clusterInst.Key.GetKeyString())
			}

			ipaccess = clusterInst.IpAccess
			clusterKey = &clusterInst.Key.ClusterKey
		}

		cloudletRefs := edgeproto.CloudletRefs{}
		cloudletRefsChanged := false
		if !cloudletRefsApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &cloudletRefs) {
			initCloudletRefs(&cloudletRefs, &in.Key.ClusterInstKey.CloudletKey)
		}

		ports, _ := edgeproto.ParseAppPorts(app.AccessPorts)
		if defaultCloudlet || !cloudcommon.IsClusterInstReqd(&app) {
			// nothing to do
		} else if ipaccess == edgeproto.IpAccess_IP_ACCESS_SHARED {
			in.Uri = cloudcommon.GetRootLBFQDN(&in.Key.ClusterInstKey.CloudletKey)
			if cloudletRefs.RootLbPorts == nil {
				cloudletRefs.RootLbPorts = make(map[int32]int32)
			}

			for ii, _ := range ports {
				if setL7Port(&ports[ii], &in.Key) {
					log.DebugLog(log.DebugLevelApi,
						"skip L7 port", "port", ports[ii])
					continue
				}
				// Samsung enabling layer ignores port mapping.
				// Attempt to use the internal port as the
				// external port so port remap is not required.
				protocolBits, err := getProtocolBitMap(ports[ii].Proto)
				if err != nil {
					return err
				}
				iport := ports[ii].InternalPort
				eport := int32(-1)
				if usedProtocols, found := cloudletRefs.RootLbPorts[iport]; !found || !protocolInUse(protocolBits, usedProtocols) {

					// rootLB has its own ports it uses
					// before any apps are even present.
					iport := ports[ii].InternalPort
					if iport != 22 &&
						iport != cloudcommon.RootLBL7Port {
						eport = iport
					}
				}
				for p := RootLBSharedPortBegin; p < 65000 && eport == int32(-1); p++ {
					// each kubernetes service gets its own
					// nginx proxy that runs in the rootLB,
					// and http ports are also mapped to it,
					// so there is no shared L7 port + path.
					if usedProtocols, found := cloudletRefs.RootLbPorts[p]; found && protocolInUse(protocolBits, usedProtocols) {

						continue
					}
					eport = p
				}
				if eport == int32(-1) {
					return errors.New("no free external ports")
				}
				ports[ii].PublicPort = eport
				existingProtoBits, _ := cloudletRefs.RootLbPorts[eport]
				cloudletRefs.RootLbPorts[eport] = addProtocol(protocolBits, existingProtoBits)

				cloudletRefsChanged = true
			}
		} else {
			if isIPAllocatedPerService(in.Key.ClusterInstKey.CloudletKey.OperatorKey.Name) {
				//dedicated access in which each service gets a different ip
				in.Uri = cloudcommon.GetAppFQDN(&in.Key, &in.Key.ClusterInstKey.CloudletKey, clusterKey)
				for ii, _ := range ports {
					// No rootLB to do L7 muxing, and each
					// service has it's own IP anyway so
					// no muxing is needed. Treat http as tcp.
					if ports[ii].Proto == dme.LProto_L_PROTO_HTTP {
						ports[ii].Proto = dme.LProto_L_PROTO_TCP
					}
					ports[ii].PublicPort = ports[ii].InternalPort
				}
			} else {
				//dedicated access in which IP is that of the LB
				in.Uri = cloudcommon.GetDedicatedLBFQDN(&in.Key.ClusterInstKey.CloudletKey, clusterKey)
				for ii, _ := range ports {
					if setL7Port(&ports[ii], &in.Key) {
						continue
					}
					ports[ii].PublicPort = ports[ii].InternalPort
				}
			}
		}
		if len(ports) > 0 {
			in.MappedPorts = ports
			if !defaultCloudlet {
				setPortFQDNPrefixes(in, &app)
			}
		}

		// TODO: Make sure resources are available
		if cloudletRefsChanged {
			cloudletRefsApi.store.STMPut(stm, &cloudletRefs)
		}
		in.CreatedAt = cloudcommon.TimeToTimestamp(time.Now())

		if ignoreCRM(cctx) || defaultCloudlet {
			in.State = edgeproto.TrackedState_READY
		} else {
			in.State = edgeproto.TrackedState_CREATE_REQUESTED
		}
		s.store.STMPut(stm, in)
		return nil
	})
	if err != nil {
		return err
	}
	if ignoreCRM(cctx) || defaultCloudlet {
		cb.Send(&edgeproto.Result{Message: "Created successfully"})
		return nil
	}
	err = appInstApi.cache.WaitForState(cb.Context(), &in.Key, edgeproto.TrackedState_READY, CreateAppInstTransitions, edgeproto.TrackedState_CREATE_ERROR, CreateAppInstTimeout, "Created successfully", cb.Send)
	if err != nil && cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Create AppInst ignoring CRM failure: %s", err.Error())})
		s.ReplaceErrorState(in, edgeproto.TrackedState_READY)
		cb.Send(&edgeproto.Result{Message: "Created AppInst successfully"})
		err = nil
	}
	if err != nil {
		// XXX should probably track mod revision ID and only undo
		// if no other changes were made to appInst in the meantime.
		// crm failed or some other err, undo
		cb.Send(&edgeproto.Result{Message: "Deleting AppInst due to failure"})
		undoErr := s.deleteAppInstInternal(cctx.WithUndo(), in, cb)
		if undoErr != nil {
			log.InfoLog("Undo create appinst", "undoErr", undoErr)
		}
	}
	return err
}

func (s *AppInstApi) updateAppInstInternal(cctx *CallContext, in *edgeproto.AppInst, cb edgeproto.AppInstApi_UpdateAppInstServer) error {
	_, err := s.store.Update(in, s.sync.syncWait)
	return err
}

func (s *AppInstApi) UpdateAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_UpdateAppInstServer) error {
	// don't allow updates to cached fields
	if in.Fields != nil {
		for _, field := range in.Fields {
			if strings.HasPrefix(field, edgeproto.AppInstFieldCloudletLoc) {
				return errors.New("Cannot specify cloudlet location fields as they are inherited from specified cloudlet")
			}
		}
	}
	return errors.New("Update app instance not supported yet")
}

func (s *AppInstApi) DeleteAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_DeleteAppInstServer) error {
	return s.deleteAppInstInternal(DefCallContext(), in, cb)
}

func (s *AppInstApi) deleteAppInstInternal(cctx *CallContext, in *edgeproto.AppInst, cb edgeproto.AppInstApi_DeleteAppInstServer) error {
	cctx.SetOverride(&in.CrmOverride)

	var defaultCloudlet bool

	log.DebugLog(log.DebugLevelApi, "deleteAppInstInternal", "appinst", in)

	if in.Key.ClusterInstKey.CloudletKey == cloudcommon.DefaultCloudletKey {
		log.DebugLog(log.DebugLevelApi, "special public cloud case", "appinst", in)
		defaultCloudlet = true
		if in.Key.ClusterInstKey.Developer == "" {
			in.Key.ClusterInstKey.Developer = in.Key.AppKey.DeveloperKey.Name
		}
	}

	if !ignoreCRM(cctx) {
		if err := cloudletInfoApi.checkCloudletReady(&in.Key.ClusterInstKey.CloudletKey); err != nil {
			return err
		}
	}
	// check if we are deleting an autocluster instance we need to set the key correctly.
	if strings.HasPrefix(in.Key.ClusterInstKey.ClusterKey.Name, ClusterAutoPrefix) && in.Key.ClusterInstKey.Developer == "" {
		in.Key.ClusterInstKey.Developer = in.Key.AppKey.DeveloperKey.Name
	}
	clusterInstKey := edgeproto.ClusterInstKey{}
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, in) {
			if in.Key.ClusterInstKey.Developer == "" {
				// create still allows it be unset on input,
				// but will set it internally. But it is required
				// on delete just in case another ClusterInst
				// exists without it set. So be nice and
				// remind them they may have forgotten to
				// specify it.
				cb.Send(&edgeproto.Result{Message: "ClusterInstKey developer not specified, may need to specify it"})
			}
			// already deleted
			return objstore.ErrKVStoreKeyNotFound
		}

		if !cctx.Undo && in.State != edgeproto.TrackedState_READY && in.State != edgeproto.TrackedState_CREATE_ERROR && in.State != edgeproto.TrackedState_DELETE_ERROR && !ignoreTransient(cctx, in.State) {
			return errors.New("AppInst busy, cannot delete")
		}

		var cloudlet edgeproto.Cloudlet
		if !defaultCloudlet {
			if !cloudletApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &cloudlet) {
				return errors.New("Specified cloudlet not found")
			}

			cloudletRefs := edgeproto.CloudletRefs{}
			cloudletRefsChanged := false
			if cloudletRefsApi.store.STMGet(stm, &in.Key.ClusterInstKey.CloudletKey, &cloudletRefs) {
				// shared root load balancer
				for ii, _ := range in.MappedPorts {
					if in.MappedPorts[ii].Proto == dme.LProto_L_PROTO_HTTP {
						continue
					}

					p := in.MappedPorts[ii].PublicPort
					protocol, err := getProtocolBitMap(in.MappedPorts[ii].Proto)

					if err != nil {
						return err
					}
					protos, _ := cloudletRefs.RootLbPorts[p]
					if cloudletRefs.RootLbPorts != nil {
						cloudletRefs.RootLbPorts[p] = removeProtocol(protos, protocol)
						if cloudletRefs.RootLbPorts[p] == 0 {
							delete(cloudletRefs.RootLbPorts, p)
						}
					}
					cloudletRefsChanged = true
				}
			}
			clusterInstKey = in.Key.ClusterInstKey
			if cloudletRefsChanged {
				cloudletRefsApi.store.STMPut(stm, &cloudletRefs)
			}
		}
		// delete app inst
		if ignoreCRM(cctx) || defaultCloudlet {
			// CRM state should be the same as before the
			// operation failed, so just need to clean up
			// controller state.
			s.store.STMDel(stm, &in.Key)
		} else {
			in.State = edgeproto.TrackedState_DELETE_REQUESTED
			s.store.STMPut(stm, in)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if ignoreCRM(cctx) || defaultCloudlet {
		return nil
	}
	err = appInstApi.cache.WaitForState(cb.Context(), &in.Key, edgeproto.TrackedState_NOT_PRESENT, DeleteAppInstTransitions, edgeproto.TrackedState_DELETE_ERROR, DeleteAppInstTimeout, "Deleted AppInst successfully", cb.Send)
	if err != nil && cctx.Override == edgeproto.CRMOverride_IGNORE_CRM_ERRORS {
		cb.Send(&edgeproto.Result{Message: fmt.Sprintf("Delete AppInst ignoring CRM failure: %s", err.Error())})
		s.ReplaceErrorState(in, edgeproto.TrackedState_NOT_PRESENT)
		cb.Send(&edgeproto.Result{Message: "Deleted AppInst successfully"})
		err = nil
	}
	if err != nil {
		// crm failed or some other err, undo
		cb.Send(&edgeproto.Result{Message: "Recreating AppInst due to failure"})
		undoErr := s.createAppInstInternal(cctx.WithUndo(), in, cb)
		if undoErr != nil {
			log.InfoLog("Undo delete appinst", "undoErr", undoErr)
		}
		return err
	}
	// delete clusterinst afterwards if it was auto-created and nobody is left using it
	clusterInst := edgeproto.ClusterInst{}
	if clusterInstApi.Get(&clusterInstKey, &clusterInst) && clusterInst.Auto && !appInstApi.UsesClusterInst(&clusterInstKey) {
		cb.Send(&edgeproto.Result{Message: "Deleting auto-cluster inst"})
		autoerr := clusterInstApi.deleteClusterInstInternal(cctx, &clusterInst, cb)
		if autoerr != nil {
			log.InfoLog("Failed to delete auto cluster inst",
				"clusterInst", clusterInst, "err", err)
		}
	}
	return err
}

func (s *AppInstApi) ShowAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_ShowAppInstServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.AppInst) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *AppInstApi) UpdateFromInfo(in *edgeproto.AppInstInfo) {
	log.DebugLog(log.DebugLevelApi, "Update AppInst from info", "key", in.Key, "state", in.State)
	s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		inst := edgeproto.AppInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		if inst.State == in.State {
			// already in that state
			if in.State == edgeproto.TrackedState_READY {
				// update runtime info
				inst.RuntimeInfo = in.RuntimeInfo
				s.store.STMPut(stm, &inst)
			}
			return nil
		}
		// please see state_transitions.md
		if !crmTransitionOk(inst.State, in.State) {
			log.DebugLog(log.DebugLevelApi, "Invalid state transition",
				"key", &in.Key, "cur", inst.State, "next", in.State)
			return nil
		}
		inst.State = in.State
		if in.State == edgeproto.TrackedState_CREATE_ERROR || in.State == edgeproto.TrackedState_DELETE_ERROR || in.State == edgeproto.TrackedState_UPDATE_ERROR {
			inst.Errors = in.Errors
		}
		inst.RuntimeInfo = in.RuntimeInfo
		s.store.STMPut(stm, &inst)
		return nil
	})
}

func (s *AppInstApi) DeleteFromInfo(in *edgeproto.AppInstInfo) {
	log.DebugLog(log.DebugLevelApi, "Delete AppInst from info", "key", in.Key, "state", in.State)
	s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		inst := edgeproto.AppInst{}
		if !s.store.STMGet(stm, &in.Key, &inst) {
			// got deleted in the meantime
			return nil
		}
		// please see state_transitions.md
		if inst.State != edgeproto.TrackedState_DELETING && inst.State != edgeproto.TrackedState_DELETE_REQUESTED {
			log.DebugLog(log.DebugLevelApi, "Invalid state transition",
				"key", &in.Key, "cur", inst.State,
				"next", edgeproto.TrackedState_NOT_PRESENT)
			return nil
		}
		s.store.STMDel(stm, &in.Key)
		return nil
	})
}

func (s *AppInstApi) ReplaceErrorState(in *edgeproto.AppInst, newState edgeproto.TrackedState) {
	s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		inst := edgeproto.AppInst{}
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

// public cloud k8s cluster allocates a separate IP per service.  This is a type of dedicated access
func isIPAllocatedPerService(operator string) bool {
	return operator == cloudcommon.OperatorGCP || operator == cloudcommon.OperatorAzure
}

func allocateIP(inst *edgeproto.ClusterInst, cloudlet *edgeproto.Cloudlet, refs *edgeproto.CloudletRefs) error {
	if isIPAllocatedPerService(cloudlet.Key.OperatorKey.Name) {
		// public cloud implements dedicated access
		inst.IpAccess = edgeproto.IpAccess_IP_ACCESS_DEDICATED
		return nil
	}
	if inst.IpAccess == edgeproto.IpAccess_IP_ACCESS_SHARED {
		// shared, so no allocation needed
		return nil
	}
	if inst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED_OR_SHARED {
		// set to shared, as CRM does not implement dedicated
		// ip assignment yet.
		inst.IpAccess = edgeproto.IpAccess_IP_ACCESS_SHARED
		return nil
	}

	// Allocate a dedicated IP
	if cloudlet.IpSupport == edgeproto.IpSupport_IP_SUPPORT_STATIC {
		// TODO:
		// parse cloudlet.StaticIps and refs.UsedStaticIps.
		// pick a free one, put it in refs.UsedStaticIps, and
		// set inst.AllocatedIp to the Ip.
		return errors.New("Static IPs not supported yet")
	} else if cloudlet.IpSupport == edgeproto.IpSupport_IP_SUPPORT_DYNAMIC {
		// Note one dynamic IP is reserved for Global Reverse Proxy LB.
		if refs.UsedDynamicIps+1 >= cloudlet.NumDynamicIps {
			return errors.New("No more dynamic IPs left")
		}
		refs.UsedDynamicIps++
		inst.AllocatedIp = cloudcommon.AllocatedIpDynamic
		return nil
	}
	return errors.New("Invalid IpSupport type")
}

func freeIP(inst *edgeproto.ClusterInst, cloudlet *edgeproto.Cloudlet, refs *edgeproto.CloudletRefs) {
	if inst.IpAccess == edgeproto.IpAccess_IP_ACCESS_SHARED {
		return
	}
	if cloudlet.IpSupport == edgeproto.IpSupport_IP_SUPPORT_STATIC {
		// TODO: free static ip in inst.AllocatedIp from refs.
	} else if cloudlet.IpSupport == edgeproto.IpSupport_IP_SUPPORT_DYNAMIC {
		refs.UsedDynamicIps--
		inst.AllocatedIp = ""
	}
}

func setPortFQDNPrefixes(in *edgeproto.AppInst, app *edgeproto.App) error {
	// For Kubernetes deployments, the CRM sets the
	// Fqdn based on the service (load balancer) name
	// in the kubernetes deployment manifest.
	// The Controller needs to set a matching
	// FqdnPrefix on the ports so the DME can tell the
	// App Client the correct Fqdn for a given port.
	if app.Deployment == cloudcommon.AppDeploymentTypeKubernetes {
		objs, _, err := cloudcommon.DecodeK8SYaml(app.DeploymentManifest)
		if err != nil {
			return fmt.Errorf("invalid kubernetes deployment yaml, %s", err.Error())
		}
		for ii, _ := range in.MappedPorts {
			setPortFQDNPrefix(&in.MappedPorts[ii], objs)
		}
	}
	return nil
}

func setPortFQDNPrefix(port *dme.AppPort, objs []runtime.Object) {
	for _, obj := range objs {
		ksvc, ok := obj.(*v1.Service)
		if !ok {
			continue
		}
		for _, kp := range ksvc.Spec.Ports {
			lproto, err := edgeproto.LProtoStr(port.Proto)
			if err != nil {
				return
			}
			if lproto != strings.ToLower(string(kp.Protocol)) {
				continue
			}
			if kp.TargetPort.IntValue() == int(port.InternalPort) {
				port.FqdnPrefix = cloudcommon.FqdnPrefix(ksvc.Name)
				return
			}
		}
	}
}

func setL7Port(port *dme.AppPort, key *edgeproto.AppInstKey) bool {
	if port.Proto != dme.LProto_L_PROTO_HTTP {
		return false
	}
	port.PublicPort = cloudcommon.RootLBL7Port
	port.PathPrefix = cloudcommon.GetL7Path(key, port.InternalPort)
	return true
}
