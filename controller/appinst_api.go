package main

import (
	"context"
	"errors"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/util"
)

type AppInstApi struct {
	sync  *Sync
	store edgeproto.AppInstStore
	cache edgeproto.AppInstCache
}

var appInstApi = AppInstApi{}

func InitAppInstApi(sync *Sync) {
	appInstApi.sync = sync
	appInstApi.store = edgeproto.NewAppInstStore(sync.store)
	edgeproto.InitAppInstCache(&appInstApi.cache)
	appInstApi.cache.SetNotifyCb(notify.ServerMgrOne.UpdateAppInst)
	sync.RegisterCache(&appInstApi.cache)
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

func (s *AppInstApi) GetAppInstsForCloudlets(cloudlets map[edgeproto.CloudletKey]struct{}, appInsts map[edgeproto.AppInstKey]struct{}) {
	s.cache.GetAppInstsForCloudlets(cloudlets, appInsts)
}

func (s *AppInstApi) UsesCloudlet(in *edgeproto.CloudletKey, dynInsts map[edgeproto.AppInstKey]struct{}) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	static := false
	for key, val := range s.cache.Objs {
		if key.CloudletKey.Matches(in) {
			if val.Liveness == edgeproto.Liveness_LivenessStatic {
				static = true
			} else if val.Liveness == edgeproto.Liveness_LivenessDynamic {
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
			if val.Liveness == edgeproto.Liveness_LivenessStatic {
				static = true
			} else if val.Liveness == edgeproto.Liveness_LivenessDynamic {
				dynInsts[key] = struct{}{}
			}
		}
	}
	return static
}

func (s *AppInstApi) UsesClusterInst(key *edgeproto.ClusterInstKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, val := range s.cache.Objs {
		if val.ClusterInstKey.Matches(key) {
			return true
		}
	}
	return false
}

func (s *AppInstApi) CreateAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	in.Liveness = edgeproto.Liveness_LivenessStatic
	return s.createAppInstInternal(in)
}

// createAppInstInternal is used to create dynamic app insts internally,
// bypassing static assignment.
func (s *AppInstApi) createAppInstInternal(in *edgeproto.AppInst) (*edgeproto.Result, error) {
	if in.Liveness == edgeproto.Liveness_LivenessUnknown {
		in.Liveness = edgeproto.Liveness_LivenessDynamic
	}

	// cache location of cloudlet in app inst
	var cloudlet edgeproto.Cloudlet
	if !cloudletApi.Get(&in.Key.CloudletKey, &cloudlet) {
		return &edgeproto.Result{}, errors.New("Specified cloudlet not found")
	}
	in.CloudletLoc = cloudlet.Location

	// cache app path in app inst
	var app edgeproto.App
	if !appApi.Get(&in.Key.AppKey, &app) {
		return &edgeproto.Result{}, errors.New("Specified app not found")
	}
	in.ImagePath = app.ImagePath
	in.ImageType = app.ImageType
	in.ConfigMap = app.ConfigMap
	in.Flavor = app.Flavor

	// CloudletKey is duplicated in both the ClusterInstKey and the
	// AppInstKey. It's inconsistent to create an AppInst on a ClusterInst
	// when the cloudlet keys don't match, so don't require the user to
	// specify the CloudletKey in the cluster during AppInst create, just
	// take it from the AppInst Key.
	in.ClusterInstKey.CloudletKey = in.Key.CloudletKey

	if in.ClusterInstKey.ClusterKey.Name == "" {
		// cluster inst unspecified, create one automatically
		clusterInst := edgeproto.ClusterInst{}
		clusterInst.Key.ClusterKey = app.Cluster
		clusterInst.Key.CloudletKey = in.Key.CloudletKey
		// it may be possible that cluster already exists
		if !clusterInstApi.HasKey(&clusterInst.Key) {
			resp, err := clusterInstApi.createClusterInstInternal(&clusterInst)
			if err != nil {
				return resp, err
			}
			defer func() {
				if err != nil {
					clusterInstApi.deleteClusterInstInternal(&clusterInst)
				}
			}()
		}
		in.ClusterInstKey = clusterInst.Key
	} else if !clusterInstApi.HasKey(&in.ClusterInstKey) {
		return &edgeproto.Result{}, errors.New("Specified ClusterInst not found")
	}

	// set up URI for this instance
	in.Uri = util.DNSSanitize(in.Key.AppKey.Name) + "." +
		util.DNSSanitize(in.Key.CloudletKey.Name) + "." +
		util.AppDNSRoot
	// TODO:
	// Allocate mapped ports and mapped path(s)
	// Reserve resources
	in.MappedPorts = app.AccessPorts
	in.MappedPath = util.DNSSanitize(app.Key.Name)

	resp, err := s.store.Create(in, s.sync.syncWait)
	return resp, err
}

func (s *AppInstApi) UpdateAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	// don't allow updates to cached fields
	if in.Fields != nil {
		for _, field := range in.Fields {
			if field == edgeproto.AppInstFieldImagePath {
				return &edgeproto.Result{}, errors.New("Cannot specify image path as it is inherited from specified app")
			} else if strings.HasPrefix(field, edgeproto.AppInstFieldCloudletLoc) {
				return &edgeproto.Result{}, errors.New("Cannot specify cloudlet location fields as they are inherited from specified cloudlet")
			}
		}
	}
	return s.store.Update(in, s.sync.syncWait)
}

func (s *AppInstApi) DeleteAppInst(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	resp, err := s.store.Delete(in, s.sync.syncWait)
	// also delete associated info
	appInstInfoApi.Del(&in.Key, s.sync.syncWait)
	return resp, err
}

func (s *AppInstApi) ShowAppInst(in *edgeproto.AppInst, cb edgeproto.AppInstApi_ShowAppInstServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.AppInst) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
