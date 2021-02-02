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
		if cur.CloudletMaintenanceTimeout == 0 {
			cur.CloudletMaintenanceTimeout = edgeproto.GetDefaultSettings().CloudletMaintenanceTimeout
			modified = true
		}
		if cur.ShepherdAlertEvaluationInterval == 0 {
			cur.ShepherdAlertEvaluationInterval = edgeproto.GetDefaultSettings().ShepherdAlertEvaluationInterval
			modified = true
		}
		if cur.UpdateVmPoolTimeout == 0 {
			cur.UpdateVmPoolTimeout = edgeproto.GetDefaultSettings().UpdateVmPoolTimeout
			modified = true
		}
		if cur.UpdateTrustPolicyTimeout == 0 {
			cur.UpdateTrustPolicyTimeout = edgeproto.GetDefaultSettings().UpdateTrustPolicyTimeout
			modified = true
		}
		if cur.InfluxDbMetricsRetention == 0 {
			cur.InfluxDbMetricsRetention = edgeproto.GetDefaultSettings().InfluxDbMetricsRetention
			modified = true
		}
		if cur.CleanupReservableAutoClusterIdletime == 0 {
			cur.CleanupReservableAutoClusterIdletime = edgeproto.GetDefaultSettings().CleanupReservableAutoClusterIdletime
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
					continue
				}
				// check the value used for MasterNodeFlavor currently
				// exists as a flavor, error if not.

				flav := edgeproto.Flavor{}
				flav.Key.Name = in.MasterNodeFlavor
				if !flavorApi.store.STMGet(stm, &(flav.Key), &flav) {
					return fmt.Errorf("Flavor must preexist")
				}
			} else if field == edgeproto.SettingsFieldInfluxDbMetricsRetention {
				log.SpanLog(ctx, log.DebugLevelApi, "update influxdb retention policy", "timer", in.InfluxDbMetricsRetention)
				err1 := services.influxQ.UpdateDefaultRetentionPolicy(in.InfluxDbMetricsRetention.TimeDuration())
				if err1 != nil {
					return err1
				}
			}
		}
		s.store.STMPut(stm, &cur)
		return nil
	})

	return &edgeproto.Result{}, err
}

func (s *SettingsApi) ResetSettings(ctx context.Context, in *edgeproto.Settings) (*edgeproto.Result, error) {
	settings := edgeproto.GetDefaultSettings()
	// Set all the fields
	settings.Fields = edgeproto.SettingsAllFields
	return s.UpdateSettings(ctx, settings)
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
