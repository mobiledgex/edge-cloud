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
	labels := map[string]string{}
	require.False(t, IsMonitoredAlert(labels))

	labels = map[string]string{AlertScopeTypeTag: "invalidScope"}
	require.False(t, IsMonitoredAlert(labels))

	labels = map[string]string{"alertname": "SomeUserAlert", AlertScopeTypeTag: AlertScopeApp}
	require.True(t, IsMonitoredAlert(labels))

	labels = map[string]string{"alertname": "UnmonitoredAlert", AlertScopeTypeTag: AlertScopeCloudlet}
	require.True(t, IsMonitoredAlert(labels))

	labels = map[string]string{"alertname": AlertAutoScaleUp, AlertScopeTypeTag: AlertScopeCloudlet}
	require.True(t, IsMonitoredAlert(labels))

	labels = map[string]string{"alertname": AlertAutoScaleDown}
	require.True(t, IsMonitoredAlert(labels))

	labels = map[string]string{"alertname": AlertAppInstDown, AlertScopeTypeTag: AlertScopeCloudlet}
	require.True(t, IsMonitoredAlert(labels))

	labels = map[string]string{"alertname": AlertAutoUndeploy, AlertScopeTypeTag: AlertScopeCloudlet}
	require.True(t, IsMonitoredAlert(labels))

	labels = map[string]string{"alertname": "", AlertScopeTypeTag: AlertScopeCloudlet}
	require.False(t, IsMonitoredAlert(labels))

	labels = map[string]string{"alertname": "UnmonitoredAlert"}
	require.False(t, IsMonitoredAlert(labels))
}
