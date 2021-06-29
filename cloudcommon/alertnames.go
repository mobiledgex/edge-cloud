package cloudcommon

// Alert names
const (
	AlertAutoScaleUp             = "AutoScaleUp"
	AlertAutoScaleDown           = "AutoScaleDown"
	AlertClusterAutoScale        = "ClusterAutoScale"
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
	AlertSeverityLabel         = "severity"
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

// Map represents severities for the specific alerts that the platfrom generates
var AlertSeverityValues = map[string]string{
	AlertAppInstDown:           AlertSeverityError,
	AlertCloudletDown:          AlertSeverityError,
	AlertCloudletResourceUsage: AlertSeverityWarn,
}

func GetSeverityForAlert(alertname string) string {
	if severity, found := AlertSeverityValues[alertname]; found {
		return severity
	}
	// default to "info"
	return AlertSeverityInfo
}

func IsMonitoredAlert(alertName string) bool {
	if alertName == AlertClusterAutoScale ||
		alertName == AlertAutoScaleUp ||
		alertName == AlertAutoScaleDown ||
		alertName == AlertAppInstDown ||
		alertName == AlertCloudletDown ||
		alertName == AlertAutoUndeploy ||
		alertName == AlertCloudletResourceUsage {
		return true
	}
	return false
}

func IsInternalAlert(alertName string) bool {
	if alertName == AlertAppInstDown ||
		alertName == AlertCloudletDown ||
		alertName == AlertCloudletResourceUsage {
		return false
	}
	return true
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
