package cloudcommon

// Alert names
var AlertAutoScaleUp = "AutoScaleUp"
var AlertAutoScaleDown = "AutoScaleDown"
var AlertAppInstDown = "AppInstDown"

// Alert label keys
var AlertLabelDev = "dev"
var AlertLabelOperator = "operator"
var AlertLabelCloudlet = "cloudlet"
var AlertLabelCluster = "cluster"
var AlertLabelApp = "app"
var AlertLabelAppVer = "appver"

// Alert annotation keys
// for autoscale:
var AlertKeyNodeCount = "nodecount"
var AlertKeyLowCpuNodeCount = "lowcpunodecount"
var AlertKeyMinNodes = "minnodes"

// for healthCheck:
var AlertHealthCheckStatus = "status"
