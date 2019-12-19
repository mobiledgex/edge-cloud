package cloudcommon

import "time"

// Shepherd settings
// Metrics collection interval for k8s and docker appInstances
var ShepherdMetricsCollectionInterval = time.Second * 5

// Number of times Health Check fails before we mark appInst down
var ShepherdHealthCheckRetries = 3

// Health Checking probing frequency
var ShepherdHealthCheckInterval = time.Second * 5

// Auto Provisioning Stats push and analysis interval
var AutoDeployIntervalSec float64 = 300

// Auto Provisioning analysis offset from interval
var AutoDeployOffsetSec float64 = 20

// Auto Provisioning Policy max allowed intervals
var AutoDeployMaxIntervals uint32 = 10
