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
	// check durations
	v := NewFieldValidator(SettingsAllFieldsStringMap)
	for f, _ := range fields {
		switch f {
		case SettingsFieldShepherdMetricsCollectionInterval:
			v.CheckGT(f, int64(s.ShepherdMetricsCollectionInterval), 0)
		case SettingsFieldShepherdAlertEvaluationInterval:
			v.CheckGT(f, int64(s.ShepherdAlertEvaluationInterval), 0)
		case SettingsFieldShepherdHealthCheckRetries:
			v.CheckGT(f, int64(s.ShepherdHealthCheckRetries), 0)
		case SettingsFieldShepherdHealthCheckInterval:
			v.CheckGT(f, int64(s.ShepherdHealthCheckInterval), 0)
		case SettingsFieldAutoDeployIntervalSec:
			v.CheckFloatGT(f, s.AutoDeployIntervalSec, 0)
		case SettingsFieldAutoDeployOffsetSec:
			v.CheckFloatGE(f, s.AutoDeployOffsetSec, 0)
		case SettingsFieldAutoDeployMaxIntervals:
			v.CheckGT(f, int64(s.AutoDeployMaxIntervals), 0)
		case SettingsFieldLoadBalancerMaxPortRange:
			v.CheckGT(f, int64(s.LoadBalancerMaxPortRange), 0)
			v.CheckLT(f, int64(s.LoadBalancerMaxPortRange), 65536)
		case SettingsFieldCreateAppInstTimeout:
			v.CheckGT(f, int64(s.CreateAppInstTimeout), 0)
		case SettingsFieldUpdateAppInstTimeout:
			v.CheckGT(f, int64(s.UpdateAppInstTimeout), 0)
		case SettingsFieldDeleteAppInstTimeout:
			v.CheckGT(f, int64(s.DeleteAppInstTimeout), 0)
		case SettingsFieldCreateClusterInstTimeout:
			v.CheckGT(f, int64(s.CreateClusterInstTimeout), 0)
		case SettingsFieldUpdateClusterInstTimeout:
			v.CheckGT(f, int64(s.UpdateClusterInstTimeout), 0)
		case SettingsFieldDeleteClusterInstTimeout:
			v.CheckGT(f, int64(s.DeleteClusterInstTimeout), 0)
		case SettingsFieldCreateCloudletTimeout:
			v.CheckGT(f, int64(s.CreateCloudletTimeout), 0)
		case SettingsFieldUpdateCloudletTimeout:
			v.CheckGT(f, int64(s.UpdateCloudletTimeout), 0)
		case SettingsFieldMasterNodeFlavor:
			// no validation
		case SettingsFieldMaxTrackedDmeClients:
			v.CheckGT(f, int64(s.MaxTrackedDmeClients), 0)
		case SettingsFieldChefClientInterval:
			v.CheckGT(f, int64(s.ChefClientInterval), 0)
		case SettingsFieldCloudletMaintenanceTimeout:
			v.CheckGT(f, int64(s.CloudletMaintenanceTimeout), 0)
		case SettingsFieldInfluxDbMetricsRetention:
			fallthrough
		case SettingsFieldInfluxDbCloudletUsageMetricsRetention:
			// no validation
		case SettingsFieldUpdateVmPoolTimeout:
			v.CheckGT(f, int64(s.UpdateVmPoolTimeout), 0)
		case SettingsFieldUpdateTrustPolicyTimeout:
			v.CheckGT(f, int64(s.UpdateTrustPolicyTimeout), 0)
		case SettingsFieldDmeApiMetricsCollectionInterval:
			v.CheckGT(f, int64(s.DmeApiMetricsCollectionInterval), 0)
		case SettingsFieldCleanupReservableAutoClusterIdletime:
			v.CheckGT(f, int64(s.CleanupReservableAutoClusterIdletime), 0)
		case SettingsFieldEdgeEventsMetricsCollectionInterval:
			v.CheckGT(f, int64(s.EdgeEventsMetricsCollectionInterval), 0)
		case SettingsFieldEdgeEventsMetricsContinuousQueriesCollectionIntervals:
			for _, val := range s.EdgeEventsMetricsContinuousQueriesCollectionIntervals {
				v.CheckGT(f, int64(val.Interval), 0)
			}
		case SettingsFieldEdgeEventsMetricsContinuousQueriesCollectionIntervalsInterval:
			// no validation
		case SettingsFieldInfluxDbEdgeEventsMetricsRetention:
			// no validation
		case SettingsFieldInfluxDbDownsampledMetricsRetention:
			// no validation
		case SettingsFieldLocationTileSideLengthKm:
			v.CheckGT(f, int64(s.LocationTileSideLengthKm), 0)
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
	s.LoadBalancerMaxPortRange = 50
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
	return &s
}

func (s *SettingsCache) Singular() *Settings {
	cur := Settings{}
	if s.Get(&SettingsKeySingular, &cur) {
		return &cur
	}
	return GetDefaultSettings()
}
