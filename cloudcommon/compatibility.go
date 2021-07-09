package cloudcommon

const (
	CRMCompatibilityAutoReservableCluster uint32 = 1
	CRMCompatibilitySharedRootLBFQDN      uint32 = 2
)

// This should always return the highest compatibility version
func GetCRMCompatibilityVersion() uint32 {
	return CRMCompatibilitySharedRootLBFQDN
}
