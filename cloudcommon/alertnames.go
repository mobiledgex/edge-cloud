package cloudcommon

// Alert names
const (
	AlertAutoScaleUp             = "AutoScaleUp"
	AlertAutoScaleDown           = "AutoScaleDown"
	AlertAppInstDown             = "AppInstDown"
	AlertAutoUndeploy            = "AutoProvUndeploy"
	AlertCloudletDown            = "CloudletDown"
	AlertCloudletDownDescription = "Cloudlet resource manager is offline"
	AlertCloudletResourceUsage   = "CloudletResourceUsage"
)

// Alert types
const (
	AlertAnnotationTitle       = "title"
	AlertAnnotationDescription = "description"
	AlertScopeTypeTag          = "scope"
	AlertScopeApp              = "Application"
	AlertScopeCloudlet         = "Cloudlet"
)

// Alert annotation keys
// for autoscale:
var AlertKeyNodeCount = "nodecount"
var AlertKeyLowCpuNodeCount = "lowcpunodecount"
var AlertKeyMinNodes = "minnodes"

// for healthCheck:
var AlertHealthCheckStatus = "status"

const (
	AlertSeverityError = "error"
	AlertSeverityWarn  = "warning"
	AlertSeverityInfo  = "info"
	// List in the order of increasing severity
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
		alertName == AlertCloudletDown ||
		alertName == AlertAutoUndeploy ||
		alertName == AlertCloudletResourceUsage {
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
