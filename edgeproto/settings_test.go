package edgeproto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSettingsValidate(t *testing.T) {
	// This exercises all the Validate checks, to make sure
	// there aren't any mismatched types being passed to the check
	// functions. Because those type checks are runtime checks,
	// but we want to catch any type mismatches at compile-time.
	settings := GetDefaultSettings()
	err := settings.Validate(SettingsAllFieldsMap)
	require.Nil(t, err)

	// Check output format of float values (use %g instead of %f to avoid 0.0000)
	settings = GetDefaultSettings()
	settings.AutoDeployIntervalSec = -1
	err = settings.Validate(SettingsAllFieldsMap)
	require.NotNil(t, err)
	require.Equal(t, "Auto Deploy Interval Sec must be greater than 0", err.Error())

	// Check output format of duration values
	// (make sure format is in human-readable string, instead of raw nanosec)
	settings = GetDefaultSettings()
	settings.AppinstClientCleanupInterval = Duration(time.Duration(time.Second))
	err = settings.Validate(SettingsAllFieldsMap)
	require.NotNil(t, err)
	require.Equal(t, "Appinst Client Cleanup Interval must be greater than 2s", err.Error())
	settings = GetDefaultSettings()
	settings.CreateAppInstTimeout = Duration(0)
	err = settings.Validate(SettingsAllFieldsMap)
	require.NotNil(t, err)
	require.Equal(t, "Create App Inst Timeout must be greater than 0s", err.Error())
}
