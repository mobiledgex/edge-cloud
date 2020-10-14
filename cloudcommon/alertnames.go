package cloudcommon

// Alert names
var AlertAutoScaleUp = "AutoScaleUp"
var AlertAutoScaleDown = "AutoScaleDown"
var AlertAppInstDown = "AppInstDown"
var AlertAutoUndeploy = "AutoProvUndeploy"

// Alert annotation keys
// for autoscale:
var AlertKeyNodeCount = "nodecount"
var AlertKeyLowCpuNodeCount = "lowcpunodecount"
var AlertKeyMinNodes = "minnodes"

// for healthCheck:
var AlertHealthCheckStatus = "status"

const (
	AlertSeverityError       = "error"
	AlertSeverityWarn        = "warning"
	AlertSeverityInfo        = "info"
	ValidAlertSeverityString = `"info", "warning", "error"`
)

var AlertSeverityTypes = map[string]struct{}{
	AlertSeverityError: struct{}{},
	AlertSeverityWarn:  struct{}{},
	AlertSeverityInfo:  struct{}{},
}

func IsMonitoredAlert(alertName string) bool {
	if alertName == AlertAutoScaleUp ||
		alertName == AlertAutoScaleDown ||
		alertName == AlertAppInstDown ||
		alertName == AlertAutoUndeploy {
		return true
	}
	return false
}

func IsAlertSeverityValid(severity string) bool {
	if _, found := AlertSeverityTypes[severity]; found {
		return true
	}
	return false
}

// Helper function - returns the string representations of all valid severities
func GetValidAlertSeverityString() string {
	return ValidAlertSeverityString
}
