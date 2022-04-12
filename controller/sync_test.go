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

package main

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/objstore"
	"github.com/stretchr/testify/require"
)

func TestSyncInsertBatch(t *testing.T) {
	sy := Sync{}
	sy.notifyOrd = edgeproto.NewNotifyOrder()

	sy.insertBatchData(&edgeproto.AppInstRefsCache{}, &objstore.SyncCbData{
		Key: []byte("appInstRefsKey"),
	})
	sy.insertBatchData(&edgeproto.AppInstCache{}, &objstore.SyncCbData{
		Key: []byte("appInstKey"),
	})
	sy.insertBatchData(&edgeproto.AppCache{}, &objstore.SyncCbData{
		Key: []byte("appKey"),
	})
	sy.insertBatchData(&edgeproto.FlavorCache{}, &objstore.SyncCbData{
		Key: []byte("flavorKey"),
	})
	sy.insertBatchData(&edgeproto.CloudletCache{}, &objstore.SyncCbData{
		Key: []byte("cloudletKey"),
	})
	// make sure sorted order is correct, order val in comment,
	// see edgeproto TestNotifyOrder unit test.
	newOrder := []string{
		"flavorKey",      // 0
		"cloudletKey",    // 1
		"appKey",         // 3
		"appInstKey",     // 4
		"appInstRefsKey", // 5
	}
	require.Equal(t, len(newOrder), len(sy.batch))
	for ii, name := range newOrder {
		require.Equal(t, name, string(sy.batch[ii].data.Key), "%d: %s", ii, name)
	}
}
