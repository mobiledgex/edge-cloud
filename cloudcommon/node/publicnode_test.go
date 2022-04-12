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

package node

import (
	"context"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
)

func TestPublicCertManager(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	api := &cloudcommon.TestPublicCertApi{}
	mgr, err := NewPublicCertManager("localhost", api, "", "")
	require.Nil(t, err)
	_, err = mgr.GetServerTlsConfig(ctx)
	require.Nil(t, err)
	require.Equal(t, 1, api.GetCount)

	// force refresh
	mgr.expiresAt = time.Now()
	mgr.refreshThreshold = time.Hour
	mgr.refreshRetryDelay = time.Millisecond
	mgr.StartRefresh()
	// wait until refresh done
	for ii := 0; ii < 10; ii++ {
		if api.GetCount == 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.Equal(t, 2, api.GetCount)
	mgr.StopRefresh()
}
