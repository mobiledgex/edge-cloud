package cloudcommon

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateSeverity(t *testing.T) {
	require.True(t, IsAlertSeverityValid("info"))
	require.True(t, IsAlertSeverityValid("warning"))
	require.True(t, IsAlertSeverityValid("error"))
	require.False(t, IsAlertSeverityValid("invaid"))
	require.False(t, IsAlertSeverityValid(""))
}

func TestValidateMonitoredAlert(t *testing.T) {
	require.True(t, IsMonitoredAlert(AlertAutoScaleUp))
	require.True(t, IsMonitoredAlert(AlertAutoScaleDown))
	require.True(t, IsMonitoredAlert(AlertAppInstDown))
	require.True(t, IsMonitoredAlert(AlertAutoUndeploy))
	require.False(t, IsMonitoredAlert(""))
	require.False(t, IsMonitoredAlert("UnmonitoredAlert"))
}
