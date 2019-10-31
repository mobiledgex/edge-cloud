package cloudcommon

// Alert names
var AlertAutoScaleUp = "AutoScaleUp"
var AlertAutoScaleDown = "AutoScaleDown"

// Alert label keys
var AlertLabelDev = "dev"
var AlertLabelOperator = "operator"
var AlertLabelCloudlet = "cloudlet"
var AlertLabelCluster = "cluster"

// Alert annotation keys
// for autoscale:
var AlertKeyNodeCount = "nodecount"
var AlertKeyLowCpuNodeCount = "lowcpunodecount"
var AlertKeyMinNodes = "minnodes"
