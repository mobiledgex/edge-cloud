package cloudcommon

const (
	CRMCompatibilityAutoReservableCluster uint32 = 1
)

// This should always return the highest compatibility version
func GetCRMCompatibilityVersion() uint32 {
	return CRMCompatibilityAutoReservableCluster
}
