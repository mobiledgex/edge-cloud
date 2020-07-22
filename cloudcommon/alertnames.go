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

func IsMonitoredAlert(alertName string) bool {
	if alertName == AlertAutoScaleUp ||
		alertName == AlertAutoScaleDown ||
		alertName == AlertAppInstDown ||
		alertName == AlertAutoUndeploy {
		return true
	}
	return false
}
