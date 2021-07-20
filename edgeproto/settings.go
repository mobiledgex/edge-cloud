package edgeproto

import (
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud/objstore"
)

type SettingsKey string

// There is only one settings object allowed
var settingsKey = "settings"
var SettingsKeySingular = SettingsKey(settingsKey)

func (m SettingsKey) GetKeyString() string {
	return settingsKey
}

func (m *SettingsKey) Matches(o *SettingsKey) bool {
	return true
}

func (m SettingsKey) ValidateKey() error {
	return nil
}

func (m SettingsKey) NotFoundError() error {
	// n/a
	return nil
}

func (m SettingsKey) ExistsError() error {
	// n/a
	return nil
}

func (m SettingsKey) GetTags() map[string]string {
	return map[string]string{}
}

func (m *Settings) GetObjKey() objstore.ObjKey {
	return m.GetKey()
}

func (s *Settings) GetKey() *SettingsKey {
	return &SettingsKeySingular
}

func (s *Settings) GetKeyVal() SettingsKey {
	return SettingsKeySingular
}

func (s *Settings) SetKey(key *SettingsKey) {}

func SettingsKeyStringParse(str string, obj *Settings) {}

func (s *Settings) Validate(fields map[string]struct{}) error {
	dur0 := Duration(0)
	v := NewFieldValidator(SettingsAllFieldsStringMap)
	for f, _ := range fields {
		switch f {
		case SettingsFieldShepherdMetricsCollectionInterval:
			v.CheckGT(f, s.ShepherdMetricsCollectionInterval, dur0)
		case SettingsFieldShepherdAlertEvaluationInterval:
			v.CheckGT(f, s.ShepherdAlertEvaluationInterval, dur0)
		case SettingsFieldShepherdHealthCheckRetries:
			v.CheckGT(f, s.ShepherdHealthCheckRetries, int32(0))
		case SettingsFieldShepherdHealthCheckInterval:
			v.CheckGT(f, s.ShepherdHealthCheckInterval, dur0)
		case SettingsFieldAutoDeployIntervalSec:
			v.CheckGT(f, s.AutoDeployIntervalSec, float64(0))
		case SettingsFieldAutoDeployOffsetSec:
			v.CheckGTE(f, s.AutoDeployOffsetSec, float64(0))
		case SettingsFieldAutoDeployMaxIntervals:
			v.CheckGT(f, s.AutoDeployMaxIntervals, uint32(0))
		case SettingsFieldCreateAppInstTimeout:
			v.CheckGT(f, s.CreateAppInstTimeout, dur0)
		case SettingsFieldUpdateAppInstTimeout:
			v.CheckGT(f, s.UpdateAppInstTimeout, dur0)
		case SettingsFieldDeleteAppInstTimeout:
			v.CheckGT(f, s.DeleteAppInstTimeout, dur0)
		case SettingsFieldCreateClusterInstTimeout:
			v.CheckGT(f, s.CreateClusterInstTimeout, dur0)
		case SettingsFieldUpdateClusterInstTimeout:
			v.CheckGT(f, s.UpdateClusterInstTimeout, dur0)
		case SettingsFieldDeleteClusterInstTimeout:
			v.CheckGT(f, s.DeleteClusterInstTimeout, dur0)
		case SettingsFieldCreateCloudletTimeout:
			v.CheckGT(f, s.CreateCloudletTimeout, dur0)
		case SettingsFieldUpdateCloudletTimeout:
			v.CheckGT(f, s.UpdateCloudletTimeout, dur0)
		case SettingsFieldMasterNodeFlavor:
			// no validation
		case SettingsFieldMaxTrackedDmeClients:
			v.CheckGT(f, s.MaxTrackedDmeClients, int32(0))
		case SettingsFieldChefClientInterval:
			v.CheckGT(f, s.ChefClientInterval, dur0)
		case SettingsFieldCloudletMaintenanceTimeout:
			v.CheckGT(f, s.CloudletMaintenanceTimeout, dur0)
		case SettingsFieldEdgeEventsMetricsContinuousQueriesCollectionIntervalsInterval:
			// no validation
		case SettingsFieldInfluxDbMetricsRetention:
			// no validation
		case SettingsFieldInfluxDbCloudletUsageMetricsRetention:
			// no validation
		case SettingsFieldUpdateVmPoolTimeout:
			v.CheckGT(f, s.UpdateVmPoolTimeout, dur0)
		case SettingsFieldUpdateTrustPolicyTimeout:
			v.CheckGT(f, s.UpdateTrustPolicyTimeout, dur0)
		case SettingsFieldDmeApiMetricsCollectionInterval:
			v.CheckGT(f, s.DmeApiMetricsCollectionInterval, dur0)
		case SettingsFieldCleanupReservableAutoClusterIdletime:
			v.CheckGT(f, s.CleanupReservableAutoClusterIdletime, Duration(30*time.Second))
		case SettingsFieldAppinstClientCleanupInterval:
			v.CheckGT(f, s.AppinstClientCleanupInterval, Duration(2*time.Second))
		case SettingsFieldEdgeEventsMetricsCollectionInterval:
			v.CheckGT(f, s.EdgeEventsMetricsCollectionInterval, dur0)
		case SettingsFieldEdgeEventsMetricsContinuousQueriesCollectionIntervals:
			for _, val := range s.EdgeEventsMetricsContinuousQueriesCollectionIntervals {
				v.CheckGT(f, val.Interval, dur0)
			}
		case SettingsFieldInfluxDbEdgeEventsMetricsRetention:
			// no validation
		case SettingsFieldInfluxDbDownsampledMetricsRetention:
			// no validation
		case SettingsFieldLocationTileSideLengthKm:
			v.CheckGT(f, s.LocationTileSideLengthKm, int64(0))
		case SettingsFieldClusterAutoScaleAveragingDurationSec:
			v.CheckGT(f, s.ClusterAutoScaleAveragingDurationSec, int64(0))
		case SettingsFieldClusterAutoScaleRetryDelay:
			v.CheckGT(f, s.ClusterAutoScaleRetryDelay, dur0)
		case SettingsFieldUserDefinedAlertMinTriggerTime:
			v.CheckGT(f, s.UserDefinedAlertMinTriggerTime, dur0)
		case SettingsFieldDisableRateLimit:
			// no validation
		case SettingsFieldMaxNumPerIpRateLimiters:
			v.CheckGT(f, s.MaxNumPerIpRateLimiters, int64(0))
		default:
			// If this is a setting field (and not "fields"), ensure there is an entry in the switch
			// above.  If no validation is to be done for a field, make an empty case entry
			_, ok := SettingsAllFieldsMap[f]
			if ok {
				return fmt.Errorf("No validation set for settings field: %s - %s", v.fieldDesc[f], f)
			}
		}
	}
	return v.err
}

func GetDefaultSettings() *Settings {
	s := Settings{}
	// Set default values
	s.ShepherdMetricsCollectionInterval = Duration(5 * time.Second)
	s.ShepherdAlertEvaluationInterval = Duration(15 * time.Second)
	s.ShepherdHealthCheckRetries = 3
	s.ShepherdHealthCheckInterval = Duration(5 * time.Second)
	s.AutoDeployIntervalSec = 300
	s.AutoDeployOffsetSec = 20
	s.AutoDeployMaxIntervals = 10
	s.CreateAppInstTimeout = Duration(30 * time.Minute)
	s.UpdateAppInstTimeout = Duration(30 * time.Minute)
	s.DeleteAppInstTimeout = Duration(20 * time.Minute)
	s.CreateClusterInstTimeout = Duration(30 * time.Minute)
	s.UpdateClusterInstTimeout = Duration(20 * time.Minute)
	s.DeleteClusterInstTimeout = Duration(20 * time.Minute)
	s.CreateCloudletTimeout = Duration(30 * time.Minute)
	s.UpdateCloudletTimeout = Duration(20 * time.Minute)
	s.MasterNodeFlavor = ""
	s.MaxTrackedDmeClients = 100
	s.ChefClientInterval = Duration(10 * time.Minute)
	s.CloudletMaintenanceTimeout = Duration(5 * time.Minute)
	s.UpdateVmPoolTimeout = Duration(20 * time.Minute)
	s.UpdateTrustPolicyTimeout = Duration(10 * time.Minute)
	s.DmeApiMetricsCollectionInterval = Duration(30 * time.Second)
	s.InfluxDbMetricsRetention = Duration(672 * time.Hour) // 28 days is a default
	s.CleanupReservableAutoClusterIdletime = Duration(30 * time.Minute)
	s.InfluxDbCloudletUsageMetricsRetention = Duration(8760 * time.Hour) // 1 year
	s.AppinstClientCleanupInterval = Duration(24 * time.Hour)            // 24 hours, dme's cookieExpiration
	s.LocationTileSideLengthKm = 2
	s.EdgeEventsMetricsCollectionInterval = Duration(1 * time.Hour)  // Collect every hour
	s.InfluxDbEdgeEventsMetricsRetention = Duration(672 * time.Hour) // 28 days
	s.EdgeEventsMetricsContinuousQueriesCollectionIntervals = []*CollectionInterval{
		&CollectionInterval{
			Interval: Duration(24 * time.Hour), // Downsample into daily intervals
		},
		&CollectionInterval{
			Interval: Duration(168 * time.Hour), // Downsample into weekly intervals
		},
		&CollectionInterval{
			Interval: Duration(672 * time.Hour), // Downsample into monthly intervals
		},
	}
	s.InfluxDbDownsampledMetricsRetention = Duration(8760 * time.Hour) // 1 year
	s.ClusterAutoScaleAveragingDurationSec = 60
	s.ClusterAutoScaleRetryDelay = Duration(time.Minute)
	s.UserDefinedAlertMinTriggerTime = Duration(30 * time.Second)
	s.DisableRateLimit = false
	s.MaxNumPerIpRateLimiters = 10000
	return &s
}

func (s *SettingsCache) Singular() *Settings {
	cur := Settings{}
	if s.Get(&SettingsKeySingular, &cur) {
		return &cur
	}
	return GetDefaultSettings()
}
