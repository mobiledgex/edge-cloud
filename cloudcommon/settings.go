package cloudcommon

import "time"

// Shepherd settings
// Metrics collection interval for k8s and docker appInstances
var ShepherdMetricsCollectionInterval = time.Second * 5

// Number of times Health Check fails before we mark appInst down
var ShepherdHealthCheckRetries = 3

// Health Checking probing frequency
var ShepherdHealthCheckInterval = time.Second * 5
