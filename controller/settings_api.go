package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	influxq "github.com/mobiledgex/edge-cloud/controller/influxq_client"
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
	def := edgeproto.GetDefaultSettings()
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := &edgeproto.Settings{}
		modified := false
		if !s.store.STMGet(stm, &edgeproto.SettingsKeySingular, cur) {
			cur = def
			modified = true
		}
		if cur.ChefClientInterval == 0 {
			cur.ChefClientInterval = def.ChefClientInterval
			modified = true
		}
		if cur.CloudletMaintenanceTimeout == 0 {
			cur.CloudletMaintenanceTimeout = def.CloudletMaintenanceTimeout
			modified = true
		}
		if cur.ShepherdAlertEvaluationInterval == 0 {
			cur.ShepherdAlertEvaluationInterval = def.ShepherdAlertEvaluationInterval
			modified = true
		}
		if cur.ShepherdMetricsScrapeInterval == 0 {
			cur.ShepherdMetricsScrapeInterval = def.ShepherdMetricsScrapeInterval
			if cur.ShepherdAlertEvaluationInterval < cur.ShepherdMetricsScrapeInterval {
				// eval interval cannot be less than scrape interval
				cur.ShepherdMetricsScrapeInterval = cur.ShepherdAlertEvaluationInterval
			}
			modified = true
		}
		if cur.UpdateVmPoolTimeout == 0 {
			cur.UpdateVmPoolTimeout = def.UpdateVmPoolTimeout
			modified = true
		}
		if cur.UpdateTrustPolicyTimeout == 0 {
			cur.UpdateTrustPolicyTimeout = def.UpdateTrustPolicyTimeout
			modified = true
		}
		if cur.InfluxDbMetricsRetention == 0 {
			cur.InfluxDbMetricsRetention = def.InfluxDbMetricsRetention
			modified = true
		}
		if cur.CleanupReservableAutoClusterIdletime == 0 {
			cur.CleanupReservableAutoClusterIdletime = def.CleanupReservableAutoClusterIdletime
			modified = true
		}
		if cur.InfluxDbCloudletUsageMetricsRetention == 0 {
			cur.InfluxDbCloudletUsageMetricsRetention = def.InfluxDbCloudletUsageMetricsRetention
			modified = true
		}
		if cur.MaxTrackedDmeClients == 0 {
			cur.MaxTrackedDmeClients = def.MaxTrackedDmeClients
			modified = true
		}
		if cur.DmeApiMetricsCollectionInterval == 0 {
			cur.DmeApiMetricsCollectionInterval = def.DmeApiMetricsCollectionInterval
			modified = true
		}
		if cur.EdgeEventsMetricsCollectionInterval == 0 {
			cur.EdgeEventsMetricsCollectionInterval = def.EdgeEventsMetricsCollectionInterval
			modified = true
		}
		if cur.InfluxDbEdgeEventsMetricsRetention == 0 {
			cur.InfluxDbEdgeEventsMetricsRetention = def.InfluxDbEdgeEventsMetricsRetention
			modified = true
		}
		if cur.InfluxDbDownsampledMetricsRetention == 0 {
			cur.InfluxDbDownsampledMetricsRetention = def.InfluxDbDownsampledMetricsRetention
			modified = true
		}
		if cur.LocationTileSideLengthKm == 0 {
			cur.LocationTileSideLengthKm = def.LocationTileSideLengthKm
			modified = true
		}
		if cur.EdgeEventsMetricsContinuousQueriesCollectionIntervals == nil || len(cur.EdgeEventsMetricsContinuousQueriesCollectionIntervals) == 0 {
			cur.EdgeEventsMetricsContinuousQueriesCollectionIntervals = def.EdgeEventsMetricsContinuousQueriesCollectionIntervals
			modified = true
		}
		if cur.CreateCloudletTimeout == 0 {
			cur.CreateCloudletTimeout = def.CreateCloudletTimeout
			modified = true
		}
		if cur.UpdateCloudletTimeout == 0 {
			cur.UpdateCloudletTimeout = def.UpdateCloudletTimeout
			modified = true
		}
		if cur.AppinstClientCleanupInterval == 0 {
			cur.AppinstClientCleanupInterval = def.AppinstClientCleanupInterval
			modified = true
		}
		if cur.ClusterAutoScaleAveragingDurationSec == 0 {
			cur.ClusterAutoScaleAveragingDurationSec = def.ClusterAutoScaleAveragingDurationSec
			modified = true
		}
		if cur.ClusterAutoScaleRetryDelay == 0 {
			cur.ClusterAutoScaleRetryDelay = def.ClusterAutoScaleRetryDelay
			modified = true
		}
		if cur.AlertPolicyMinTriggerTime == 0 {
			cur.AlertPolicyMinTriggerTime = def.AlertPolicyMinTriggerTime
			modified = true
		}
		if cur.RateLimitMaxTrackedIps == 0 {
			cur.RateLimitMaxTrackedIps = edgeproto.GetDefaultSettings().RateLimitMaxTrackedIps
			modified = true
		}
		if cur.ResourceSnapshotThreadInterval == 0 {
			cur.ResourceSnapshotThreadInterval = edgeproto.GetDefaultSettings().ResourceSnapshotThreadInterval
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
		oldSettings := cur
		changeCount := cur.CopyInFields(in)
		log.SpanLog(ctx, log.DebugLevelApi, "update settings", "changed", changeCount)
		if changeCount == 0 {
			// nothing changed
			return nil
		}
		if cur.ShepherdAlertEvaluationInterval < cur.ShepherdMetricsScrapeInterval {
			return fmt.Errorf("Shepherd alert evaluation interval cannot be less than Shepherd metrics scrape interval")
		}
		newCqs := false
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
				err := services.influxQ.CreateRetentionPolicy(in.InfluxDbMetricsRetention.TimeDuration(), influxq.DefaultRetentionPolicy)
				if err != nil {
					return err
				}
			} else if field == edgeproto.SettingsFieldInfluxDbCloudletUsageMetricsRetention {
				log.SpanLog(ctx, log.DebugLevelApi, "update influxdb cloudlet usage metrics retention policy", "timer", in.InfluxDbCloudletUsageMetricsRetention)
				err := services.cloudletResourcesInfluxQ.CreateRetentionPolicy(in.InfluxDbCloudletUsageMetricsRetention.TimeDuration(), influxq.DefaultRetentionPolicy)
				if err != nil {
					return err
				}
			} else if field == edgeproto.SettingsFieldInfluxDbEdgeEventsMetricsRetention {
				log.SpanLog(ctx, log.DebugLevelApi, "update influxdb edge events metrics retention policy", "timer", in.InfluxDbEdgeEventsMetricsRetention)
				err := services.edgeEventsInfluxQ.CreateRetentionPolicy(in.InfluxDbEdgeEventsMetricsRetention.TimeDuration(), influxq.DefaultRetentionPolicy)
				if err != nil {
					return err
				}
			} else if field == edgeproto.SettingsFieldInfluxDbDownsampledMetricsRetention {
				log.SpanLog(ctx, log.DebugLevelApi, "update influxdb downsampled metrics retention policy", "timer", in.InfluxDbDownsampledMetricsRetention)
				err := services.downsampledMetricsInfluxQ.CreateRetentionPolicy(in.InfluxDbDownsampledMetricsRetention.TimeDuration(), influxq.DefaultRetentionPolicy)
				if err != nil {
					return err
				}
			} else if field == edgeproto.SettingsFieldEdgeEventsMetricsContinuousQueriesCollectionIntervals || field == edgeproto.SettingsFieldEdgeEventsMetricsCollectionInterval {
				// make sure cq intervals are all greater than the collection interval
				for _, cq := range cur.EdgeEventsMetricsContinuousQueriesCollectionIntervals {
					if cq.Interval <= cur.EdgeEventsMetricsCollectionInterval {
						return fmt.Errorf("All EdgeEvents continuous query collection intervals must be greater than the EdgeEventsMetricsCollectionInterval")
					}
				}
				newCqs = true
			} else if field == edgeproto.SettingsFieldEdgeEventsMetricsContinuousQueriesCollectionIntervalsRetention {
				newCqs = true
			} else if field == edgeproto.SettingsFieldEdgeEventsMetricsContinuousQueriesCollectionIntervalsInterval {
				newCqs = true
			}
		}
		if newCqs {
			// Drop old cqs
			for _, collectioninterval := range oldSettings.EdgeEventsMetricsContinuousQueriesCollectionIntervals {
				interval := time.Duration(collectioninterval.Interval)
				retention := time.Duration(collectioninterval.Retention)
				if errl := influxq.DropContinuousQuery(services.edgeEventsInfluxQ, services.downsampledMetricsInfluxQ, cloudcommon.LatencyMetric, interval, retention); errl != nil {
					return errl
				}
				if errd := influxq.DropContinuousQuery(services.edgeEventsInfluxQ, services.downsampledMetricsInfluxQ, cloudcommon.DeviceMetric, interval, retention); errd != nil {
					return errd
				}
			}
			// Create new cqs
			for _, collectioninterval := range in.EdgeEventsMetricsContinuousQueriesCollectionIntervals {
				interval := time.Duration(collectioninterval.Interval)
				retention := time.Duration(collectioninterval.Retention)
				latencyCqSettings := influxq.CreateLatencyContinuousQuerySettings(interval, retention)
				err := influxq.CreateContinuousQuery(services.edgeEventsInfluxQ, services.downsampledMetricsInfluxQ, latencyCqSettings)
				if err != nil && !strings.Contains(err.Error(), "already exists") {
					return err
				}
				deviceCqSettings := influxq.CreateDeviceInfoContinuousQuerySettings(interval, retention)
				err = influxq.CreateContinuousQuery(services.edgeEventsInfluxQ, services.downsampledMetricsInfluxQ, deviceCqSettings)
				if err != nil && !strings.Contains(err.Error(), "already exists") {
					return err
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
