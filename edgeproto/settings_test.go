// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
