package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type FlavorApi struct {
	sync  *Sync
	store edgeproto.FlavorStore
	cache edgeproto.FlavorCache
}

var flavorApi = FlavorApi{}

func InitFlavorApi(sync *Sync) {
	flavorApi.sync = sync
	flavorApi.store = edgeproto.NewFlavorStore(sync.store)
	edgeproto.InitFlavorCache(&flavorApi.cache)
	sync.RegisterCache(&flavorApi.cache)
}

func (s *FlavorApi) HasFlavor(key *edgeproto.FlavorKey) bool {
	return s.cache.HasKey(key)
}

func (s *FlavorApi) Get(key *edgeproto.FlavorKey, buf *edgeproto.Flavor) bool {
	return s.cache.Get(key, buf)
}

func (s *FlavorApi) GetAllKeys(keys map[edgeproto.FlavorKey]struct{}) {
	s.cache.GetAllKeys(keys)
}

func (s *FlavorApi) CreateFlavor(ctx context.Context, in *edgeproto.Flavor) (*edgeproto.Result, error) {
	return s.store.Create(in, s.sync.syncWait)
}

func (s *FlavorApi) UpdateFlavor(ctx context.Context, in *edgeproto.Flavor) (*edgeproto.Result, error) {
	// Unsupported for now
	return &edgeproto.Result{}, errors.New("Update flavor not supported")
	//return s.store.Update(in, s.sync.syncWait)
}

func (s *FlavorApi) DeleteFlavor(ctx context.Context, in *edgeproto.Flavor) (*edgeproto.Result, error) {
	if clusterApi.UsesFlavor(&in.Key) {
		return &edgeproto.Result{}, errors.New("Flavor in use by Cluster")
	}
	if appApi.UsesFlavor(&in.Key) {
		return &edgeproto.Result{}, errors.New("Flavor in use by App")
	}
	if appInstApi.UsesFlavor(&in.Key) {
		return &edgeproto.Result{}, errors.New("Flavor in use by App Instance")
	}
	res, err := s.store.Delete(in, s.sync.syncWait)
	// clean up auto-apps using flavor
	appApi.AutoDeleteApps(ctx, &in.Key)
	return res, err
}

func (s *FlavorApi) ShowFlavor(in *edgeproto.Flavor, cb edgeproto.FlavorApi_ShowFlavorServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Flavor) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
