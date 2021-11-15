package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type FlavorApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.FlavorStore
	cache edgeproto.FlavorCache
}

func NewFlavorApi(sync *Sync, all *AllApis) *FlavorApi {
	flavorApi := FlavorApi{}
	flavorApi.all = all
	flavorApi.sync = sync
	flavorApi.store = edgeproto.NewFlavorStore(sync.store)
	edgeproto.InitFlavorCache(&flavorApi.cache)
	sync.RegisterCache(&flavorApi.cache)
	return &flavorApi
}

func (s *FlavorApi) HasFlavor(key *edgeproto.FlavorKey) bool {
	return s.cache.HasKey(key)
}

func (s *FlavorApi) CreateFlavor(ctx context.Context, in *edgeproto.Flavor) (*edgeproto.Result, error) {

	if err := in.Validate(edgeproto.FlavorAllFieldsMap); err != nil {
		return &edgeproto.Result{}, err
	}

	if in.OptResMap != nil {
		if ok, err := s.all.resTagTableApi.ValidateOptResMapValues(in.OptResMap); !ok {
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

func (s *FlavorApi) DeleteFlavor(ctx context.Context, in *edgeproto.Flavor) (res *edgeproto.Result, reterr error) {
	// if settings.MasterNodeFlavor == in.Key.Name it must remain
	// until first removed from settings.MasterNodeFlavor
	settings := edgeproto.Settings{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.Flavor{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		if cur.DeletePrepare {
			return fmt.Errorf("Flavor already being deleted")
		}
		if !s.all.settingsApi.store.STMGet(stm, &edgeproto.SettingsKeySingular, &settings) {
			// should never happen (initDefaults)
			return edgeproto.SettingsKeySingular.NotFoundError()
		}
		if settings.MasterNodeFlavor == in.Key.Name {
			return fmt.Errorf("Flavor in use by Settings MasterNodeFlavor, change Settings.MasterNodeFlavor first")
		}
		cur.DeletePrepare = true
		s.store.STMPut(stm, &cur)
		return nil
	})
	if err != nil {
		return &edgeproto.Result{}, err
	}
	defer func() {
		if reterr == nil {
			return
		}
		undoErr := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			cur := edgeproto.Flavor{}
			if !s.store.STMGet(stm, &in.Key, &cur) {
				return nil
			}
			if cur.DeletePrepare {
				cur.DeletePrepare = false
				s.store.STMPut(stm, &cur)
			}
			return nil
		})
		if undoErr != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to undo delete prepare", "key", in.Key, "err", undoErr)
		}
	}()

	if k := s.all.clusterInstApi.UsesFlavor(&in.Key); k != nil {
		return &edgeproto.Result{}, fmt.Errorf("Flavor in use by ClusterInst %s", k.GetKeyString())
	}
	if k := s.all.appApi.UsesFlavor(&in.Key); k != nil {
		return &edgeproto.Result{}, fmt.Errorf("Flavor in use by App %s", k.GetKeyString())
	}
	if k := s.all.appInstApi.UsesFlavor(&in.Key); k != nil {
		return &edgeproto.Result{}, fmt.Errorf("Flavor in use by AppInst %s", k.GetKeyString())
	}
	if k := s.all.cloudletApi.UsesFlavor(&in.Key); k != nil {
		return &edgeproto.Result{}, fmt.Errorf("Flavor in use by Cloudlet %s", k.GetKeyString())
	}

	res, err = s.store.Delete(ctx, in, s.sync.syncWait)
	// clean up auto-apps using flavor
	s.all.appApi.AutoDeleteApps(ctx, &in.Key)
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
			if err, ok := s.all.resTagTableApi.ValidateResName(ctx, res); !ok {
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
