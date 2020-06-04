package cloudcommon

// Alert names
var AlertAutoScaleUp = "AutoScaleUp"
var AlertAutoScaleDown = "AutoScaleDown"
var AlertAppInstDown = "AppInstDown"
var AlertAutoProvDown = "AutoProvDown"

// Alert label keys
var AlertLabelClusterOrg = "clusterorg"
var AlertLabelCloudletOrg = "cloudletorg"
var AlertLabelCloudlet = "cloudlet"
var AlertLabelCluster = "cluster"
var AlertLabelApp = "app"
var AlertLabelAppVer = "appver"
var AlertLabelAppOrg = "apporg"

// Alert annotation keys
// for autoscale:
var AlertKeyNodeCount = "nodecount"
var AlertKeyLowCpuNodeCount = "lowcpunodecount"
var AlertKeyMinNodes = "minnodes"

// for healthCheck:
var AlertHealthCheckStatus = "status"
