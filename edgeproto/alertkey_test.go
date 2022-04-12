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

	"github.com/stretchr/testify/require"
)

func TestAlertKey(t *testing.T) {
	m := map[string]string{
		"cloudlet":  "localtest",
		"cluster":   "AppCluster",
		"dev":       "MobiledgeX",
		"alertname": "DeadMansSwitch",
		"operator":  "mexdev",
		"severity":  "none",
	}
	key := MapKey(m)
	expectedKey := `{"alertname":"DeadMansSwitch","cloudlet":"localtest","cluster":"AppCluster","dev":"MobiledgeX","operator":"mexdev","severity":"none"}`
	require.Equal(t, expectedKey, key)

	a := Alert{}
	AlertKeyStringParse(key, &a)
	require.Equal(t, m, a.Labels)
}
