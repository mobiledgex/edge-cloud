package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type SettingsApi struct {
	sync  *Sync
	store edgeproto.SettingsStore
	cache edgeproto.SettingsCache
}

var settingsApi = SettingsApi{}

func InitSettingsApi(sync *Sync) {
	settingsApi.sync = sync
	settingsApi.store = edgeproto.NewSettingsStore(sync.store)
	edgeproto.InitSettingsCache(&settingsApi.cache)
	sync.RegisterCache(&settingsApi.cache)
}

func (s *SettingsApi) initDefaults(ctx context.Context) error {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := &edgeproto.Settings{}
		modified := false
		if !s.store.STMGet(stm, &edgeproto.SettingsKeySingular, cur) {
			cur = edgeproto.GetDefaultSettings()
			modified = true
		}
		if cur.ChefClientInterval == 0 {
			cur.ChefClientInterval = edgeproto.GetDefaultSettings().ChefClientInterval
			modified = true
		}
		if modified {
			s.store.STMPut(stm, cur)
		}
		return nil
	})
	return err
}

func (s *SettingsApi) UpdateSettings(ctx context.Context, in *edgeproto.Settings) (*edgeproto.Result, error) {
	if err := in.Validate(edgeproto.MakeFieldMap(in.Fields)); err != nil {
		return &edgeproto.Result{}, err
	}
	log.SpanLog(ctx, log.DebugLevelApi, "update settings", "in", in)

	cur := edgeproto.Settings{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &edgeproto.SettingsKeySingular, &cur) {
			// should never happen due to initDefaults
			log.SpanLog(ctx, log.DebugLevelApi, "settings not found")
			return in.GetKey().NotFoundError()
		}
		changeCount := cur.CopyInFields(in)
		log.SpanLog(ctx, log.DebugLevelApi, "update settings", "changed", changeCount)
		if changeCount == 0 {
			// nothing changed
			return nil
		}
		for _, field := range in.Fields {
			if field == edgeproto.SettingsFieldMasterNodeFlavor {
				if in.MasterNodeFlavor == "" {
					// allow a 'clear setting' operation
					s.store.STMPut(stm, &cur)
					return nil
				}
				// check the value used for MasterNodeFlavor currently
				// exists as a flavor, error if not.

				flav := edgeproto.Flavor{}
				flav.Key.Name = in.MasterNodeFlavor
				if !flavorApi.store.STMGet(stm, &(flav.Key), &flav) {
					return fmt.Errorf("Flavor must preexist")
				}
			} else if field == edgeproto.SettingsFieldInfluxDbMetricsRetention {
				services.influxQ.UpdateDefaultRetentionPolicy(in.InfluxDbMetricsRetention.TimeDuration())
			}
		}
		s.store.STMPut(stm, &cur)
		return nil
	})

	return &edgeproto.Result{}, err
}

func (s *SettingsApi) ResetSettings(ctx context.Context, in *edgeproto.Settings) (*edgeproto.Result, error) {
	return s.store.Put(ctx, edgeproto.GetDefaultSettings(), s.sync.syncWait)
}

func (s *SettingsApi) ShowSettings(ctx context.Context, in *edgeproto.Settings) (*edgeproto.Settings, error) {
	cur := edgeproto.Settings{}
	if !s.cache.Get(&edgeproto.SettingsKeySingular, &cur) {
		return &edgeproto.Settings{}, fmt.Errorf("no settings found")
	}
	return &cur, nil
}

func (s *SettingsApi) Get() *edgeproto.Settings {
	return s.cache.Singular()
}
