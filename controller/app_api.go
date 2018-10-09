// app config

package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/coreos/etcd/clientv3/concurrency"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/mobiledgex/edge-cloud/util"
)

// Should only be one of these instantiated in main
type AppApi struct {
	sync  *Sync
	store edgeproto.AppStore
	cache edgeproto.AppCache
}

var appApi = AppApi{}

func InitAppApi(sync *Sync) {
	appApi.sync = sync
	appApi.store = edgeproto.NewAppStore(sync.store)
	edgeproto.InitAppCache(&appApi.cache)
	sync.RegisterCache(&appApi.cache)
	appApi.cache.SetUpdatedCb(appApi.UpdatedCb)
}

func (s *AppApi) HasApp(key *edgeproto.AppKey) bool {
	return s.cache.HasKey(key)
}

func (s *AppApi) Get(key *edgeproto.AppKey, buf *edgeproto.App) bool {
	return s.cache.Get(key, buf)
}

func (s *AppApi) UsesDeveloper(in *edgeproto.DeveloperKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for key, _ := range s.cache.Objs {
		if key.DeveloperKey.Matches(in) {
			return true
		}
	}
	return false
}

func (s *AppApi) UsesFlavor(key *edgeproto.FlavorKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, app := range s.cache.Objs {
		if app.DefaultFlavor.Matches(key) {
			return true
		}
	}
	return false
}

func (s *AppApi) UsesCluster(key *edgeproto.ClusterKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, app := range s.cache.Objs {
		if app.Cluster.Matches(key) {
			return true
		}
	}
	return false
}

// helper function to help with the fields validation of the App object
func (s *AppApi) validateApp(app *edgeproto.App) error {
	var err error
	if app.ImageType == edgeproto.ImageType_ImageTypeUnknown {
		return errors.New("Please specify Image Type")
	}
	_, ok := edgeproto.ImageType_name[int32(app.ImageType)]
	if !ok {
		return errors.New("invalid Image Type")
	}
	_, ok = edgeproto.IpAccess_name[int32(app.IpAccess)]
	if !ok {
		return errors.New("invalid IpAccess Type")
	}
	if app.AccessPorts == "" {
		return errors.New("Please specify access ports")
	}
	_, err = parseAppPorts(app.AccessPorts)
	if err != nil {
		return err
	}

	return nil
}

func (s *AppApi) CreateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	var err error
	err = appApi.validateApp(in)
	if err != nil {
		return &edgeproto.Result{}, err
	}

	if in.IpAccess == edgeproto.IpAccess_IpAccessUnknown {
		// default to shared
		in.IpAccess = edgeproto.IpAccess_IpAccessShared
	}
	if in.ImagePath == "" {
		if in.ImageType == edgeproto.ImageType_ImageTypeDocker {
			in.ImagePath = "mobiledgex_" +
				util.DockerSanitize(in.Key.DeveloperKey.Name) + "/" +
				util.DockerSanitize(in.Key.Name) + ":" +
				util.DockerSanitize(in.Key.Version)
		} else {
			in.ImagePath = "qcow path not determined yet"
		}
	}

	// make sure cluster exists
	// This is a separate STM to avoid ordering issues between
	// auto-cluster create and app create in watch cb.
	if in.Cluster.Name == "" {
		if in.DefaultFlavor.Name == "" {
			return &edgeproto.Result{}, errors.New("DefaultFlavor is required if Cluster is not specified")
		}
		err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
			if s.store.STMGet(stm, &in.Key, nil) {
				return objstore.ErrKVStoreKeyExists
			}
			// cluster not specified, create one automatically
			in.Cluster.Name = fmt.Sprintf("%s%s", ClusterAutoPrefix,
				in.Key.Name)
			in.Cluster.Name = util.K8SSanitize(in.Cluster.Name)
			cluster := edgeproto.Cluster{}
			cluster.Key = in.Cluster
			clusterFlavorKey, lookup := GetClusterFlavorForFlavor(&in.DefaultFlavor)
			if lookup != nil {
				return lookup
			}
			cluster.DefaultFlavor = *clusterFlavorKey
			cluster.Auto = true
			// it may be possible that cluster already exists
			if !clusterApi.store.STMGet(stm, &cluster.Key, nil) {
				log.DebugLog(log.DebugLevelApi,
					"Create auto-cluster",
					"key", cluster.Key,
					"app", in)
				clusterApi.store.STMPut(stm, &cluster)
			}
			return nil
		})
		if err != nil {
			return &edgeproto.Result{}, err
		}
		defer func() {
			if err != nil {
				s.deleteClusterAuto(&in.Cluster)
			}
		}()
	}

	err = s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		if err := in.Validate(edgeproto.AppAllFieldsMap); err != nil {
			return err
		}
		if !developerApi.store.STMGet(stm, &in.Key.DeveloperKey, nil) {
			return errors.New("Specified developer not found")
		}
		if !flavorApi.store.STMGet(stm, &in.DefaultFlavor, nil) {
			return errors.New("Specified default flavor not found")
		}
		if s.store.STMGet(stm, &in.Key, nil) {
			return objstore.ErrKVStoreKeyExists
		}
		if !clusterApi.store.STMGet(stm, &in.Cluster, nil) {
			return errors.New("Specified Cluster not found")
		}
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *AppApi) UpdateApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	return s.store.Update(in, s.sync.syncWait)
}

func (s *AppApi) DeleteApp(ctx context.Context, in *edgeproto.App) (*edgeproto.Result, error) {
	if !appApi.HasApp(&in.Key) {
		// key doesn't exist
		return &edgeproto.Result{}, objstore.ErrKVStoreKeyNotFound
	}
	dynInsts := make(map[edgeproto.AppInstKey]struct{})
	if appInstApi.UsesApp(&in.Key, dynInsts) {
		// disallow delete if static instances are present
		return &edgeproto.Result{}, errors.New("Application in use by static Application Instance")
	}
	clusterKey := edgeproto.ClusterKey{}
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		app := edgeproto.App{}
		if !s.store.STMGet(stm, &in.Key, &app) {
			// already deleted
			return nil
		}
		clusterKey = app.Cluster
		// delete app
		s.store.STMDel(stm, &in.Key)
		return nil
	})
	if err == nil {
		// delete cluster afterwards if it was auto-created
		s.deleteClusterAuto(&clusterKey)
	}
	if err == nil && len(dynInsts) > 0 {
		// delete dynamic instances
		for key, _ := range dynInsts {
			appInst := edgeproto.AppInst{Key: key}
			derr := appInstApi.DeleteAppInst(&appInst, nil)
			if derr != nil {
				log.DebugLog(log.DebugLevelApi,
					"Failed to delete dynamic app inst",
					"err", derr)
			}
		}
	}
	return &edgeproto.Result{}, err
}

func (s *AppApi) deleteClusterAuto(key *edgeproto.ClusterKey) {
	err := s.sync.ApplySTMWait(func(stm concurrency.STM) error {
		cluster := edgeproto.Cluster{}
		if clusterApi.store.STMGet(stm, key, &cluster) && cluster.Auto {
			clusterApi.store.STMDel(stm, key)
		}
		return nil
	})
	if err != nil {
		log.InfoLog("Failed to delete auto cluster",
			"clusterInst", key, "err", err)
	}
}

func (s *AppApi) ShowApp(in *edgeproto.App, cb edgeproto.AppApi_ShowAppServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.App) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *AppApi) UpdatedCb(old *edgeproto.App, new *edgeproto.App) {
	if old == nil {
		return
	}
	if old.ImagePath != new.ImagePath || old.ImageType != new.ImageType ||
		old.Config != new.Config || old.AppTemplate != new.AppTemplate {
		log.DebugLog(log.DebugLevelApi, "updating image path")
		appInstApi.cache.Mux.Lock()
		for _, inst := range appInstApi.cache.Objs {
			if inst.Key.AppKey.Matches(&new.Key) {
				old := inst
				inst.ImagePath = new.ImagePath
				inst.ImageType = new.ImageType
				inst.Config = new.Config
				inst.AppTemplate = new.AppTemplate
				// TODO: update mapped ports if needed
				if appInstApi.cache.NotifyCb != nil {
					appInstApi.cache.NotifyCb(&inst.Key, old)
				}
			}
		}
		appInstApi.cache.Mux.Unlock()
	}
}

func GetLProto(s string) (dme.LProto, error) {
	s = strings.ToLower(s)
	switch s {
	case "tcp":
		return dme.LProto_LProtoTCP, nil
	case "udp":
		return dme.LProto_LProtoUDP, nil
	case "http":
		return dme.LProto_LProtoHTTP, nil
	}
	return 0, fmt.Errorf("%s is not a supported Protocol", s)
}

func parseAppPorts(ports string) ([]dme.AppPort, error) {
	appports := make([]dme.AppPort, 0)
	strs := strings.Split(ports, ",")
	for _, str := range strs {
		vals := strings.Split(str, ":")
		if len(vals) != 2 {
			return nil, fmt.Errorf("Invalid Access Ports format, expected proto:port but was %s", vals[0])
		}
		proto, err := GetLProto(vals[0])
		if err != nil {
			return nil, err
		}
		port, err := strconv.ParseInt(vals[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("Failed to convert port %s to integer: %s", vals[1], err)
		}
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("Port %s out of range", vals[1])
		}
		p := dme.AppPort{
			Proto:        proto,
			InternalPort: int32(port),
		}
		appports = append(appports, p)
	}
	return appports, nil
}

// GetClusterFlavorForFlavor finds the smallest cluster flavor whose
// node flavor matches the passed in flavor.
func GetClusterFlavorForFlavor(flavorKey *edgeproto.FlavorKey) (*edgeproto.ClusterFlavorKey, error) {
	matchingFlavors := make([]*edgeproto.ClusterFlavor, 0)

	clusterFlavorApi.cache.Mux.Lock()
	defer clusterFlavorApi.cache.Mux.Unlock()
	for _, val := range clusterFlavorApi.cache.Objs {
		if val.NodeFlavor.Matches(flavorKey) {
			matchingFlavors = append(matchingFlavors, val)
		}
	}
	if len(matchingFlavors) == 0 {
		return nil, fmt.Errorf("No cluster flavors matching flavor %s found", flavorKey.Name)
	}
	sort.Slice(matchingFlavors, func(i, j int) bool {
		if matchingFlavors[i].MaxNodes < matchingFlavors[j].MaxNodes {
			return true
		}
		if matchingFlavors[i].NumNodes < matchingFlavors[j].NumNodes {
			return true
		}
		if matchingFlavors[i].NumMasters < matchingFlavors[j].NumMasters {
			return true
		}
		return false
	})
	return &matchingFlavors[0].Key, nil
}
