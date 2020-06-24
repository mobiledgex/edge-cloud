package cloudcommon

// Alert names
var AlertAutoScaleUp = "AutoScaleUp"
var AlertAutoScaleDown = "AutoScaleDown"
var AlertAppInstDown = "AppInstDown"
var AlertAutoProvDown = "AutoProvDown"

// Alert annotation keys
// for autoscale:
var AlertKeyNodeCount = "nodecount"
var AlertKeyLowCpuNodeCount = "lowcpunodecount"
var AlertKeyMinNodes = "minnodes"

// for healthCheck:
var AlertHealthCheckStatus = "status"
