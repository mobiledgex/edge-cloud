package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"strings"
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

func (s *FlavorApi) CreateFlavor(ctx context.Context, in *edgeproto.Flavor) (*edgeproto.Result, error) {

	if err := in.Validate(edgeproto.FlavorAllFieldsMap); err != nil {
		return &edgeproto.Result{}, err
	}

	if in.OptResMap != nil {
		if ok, err := resTagTableApi.ValidateOptResMapValues(in.OptResMap); !ok {
			return &edgeproto.Result{}, err
		}
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.ExistsError()
		}
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *FlavorApi) UpdateFlavor(ctx context.Context, in *edgeproto.Flavor) (*edgeproto.Result, error) {
	// Unsupported for now
	return &edgeproto.Result{}, errors.New("Update Flavor not supported")
	//return s.store.Update(in, s.sync.syncWait)
}

func (s *FlavorApi) DeleteFlavor(ctx context.Context, in *edgeproto.Flavor) (*edgeproto.Result, error) {
	if !flavorApi.HasFlavor(&in.Key) {
		// key doesn't exist
		return &edgeproto.Result{}, in.Key.NotFoundError()
	}
	if clusterInstApi.UsesFlavor(&in.Key) {
		return &edgeproto.Result{}, errors.New("Flavor in use by ClusterInst")
	}
	if appApi.UsesFlavor(&in.Key) {
		return &edgeproto.Result{}, errors.New("Flavor in use by App")
	}
	if appInstApi.UsesFlavor(&in.Key) {
		return &edgeproto.Result{}, errors.New("Flavor in use by AppInst")
	}
	// if settings.MasterNodeFlavor == in.Key.Name it must remain
	// until first removed from settings.MasterNodeFlavor
	settings := edgeproto.Settings{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !settingsApi.store.STMGet(stm, &edgeproto.SettingsKeySingular, &settings) {
			// should never happen (initDefaults)
			return edgeproto.SettingsKeySingular.NotFoundError()
		}
		if settings.MasterNodeFlavor == in.Key.Name {
			return fmt.Errorf("Flavor in use by Settings MasterNodeFlavor, change Settings.MasterNodeFlavor first")
		}
		return nil
	})
	if err != nil {
		return &edgeproto.Result{}, err
	}

	res, err := s.store.Delete(ctx, in, s.sync.syncWait)
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

func (s *FlavorApi) AddFlavorRes(ctx context.Context, in *edgeproto.Flavor) (*edgeproto.Result, error) {

	var flav edgeproto.Flavor
	var err error

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &flav) {
			return in.Key.NotFoundError()
		}
		if flav.OptResMap == nil {
			flav.OptResMap = make(map[string]string)
		}
		for res, val := range in.OptResMap {
			// validate the resname(s)
			if err, ok := resTagTableApi.ValidateResName(ctx, res); !ok {
				return err
			}
			in.Key.Name = strings.ToLower(in.Key.Name)
			flav.OptResMap[res] = val
		}
		s.store.STMPut(stm, &flav)
		return nil
	})

	return &edgeproto.Result{}, err
}

func (s *FlavorApi) RemoveFlavorRes(ctx context.Context, in *edgeproto.Flavor) (*edgeproto.Result, error) {
	var flav edgeproto.Flavor
	var err error

	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &flav) {
			return in.Key.NotFoundError()
		}
		for res, _ := range in.OptResMap {
			delete(flav.OptResMap, res)
		}
		s.store.STMPut(stm, &flav)
		return nil
	})

	return &edgeproto.Result{}, err
}
