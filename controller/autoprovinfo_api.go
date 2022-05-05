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
	"context"
	"fmt"
	"time"

	dme "github.com/edgexr/edge-cloud/d-match-engine/dme-proto"
	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/log"
)

type AutoProvInfoApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.AutoProvInfoStore
	cache edgeproto.AutoProvInfoCache
}

func NewAutoProvInfoApi(sync *Sync, all *AllApis) *AutoProvInfoApi {
	autoProvInfoApi := AutoProvInfoApi{}
	autoProvInfoApi.all = all
	autoProvInfoApi.sync = sync
	autoProvInfoApi.store = edgeproto.NewAutoProvInfoStore(sync.store)
	edgeproto.InitAutoProvInfoCache(&autoProvInfoApi.cache)
	sync.RegisterCache(&autoProvInfoApi.cache)
	return &autoProvInfoApi
}

func (s *AutoProvInfoApi) Update(ctx context.Context, in *edgeproto.AutoProvInfo, rev int64) {
	s.store.Put(ctx, in, nil)
}

func (s *AutoProvInfoApi) Delete(ctx context.Context, in *edgeproto.AutoProvInfo, rev int64) {
	s.store.Delete(ctx, in, nil)
}

func (s *AutoProvInfoApi) Flush(ctx context.Context, notifyId int64) {
	// no-op
}

func (s *AutoProvInfoApi) Prune(ctx context.Context, keys map[edgeproto.CloudletKey]struct{}) {
	// no-op
}

func (s *AutoProvInfoApi) waitForMaintenanceState(ctx context.Context, key *edgeproto.CloudletKey, targetState, errorState dme.MaintenanceState, timeout time.Duration, result *edgeproto.AutoProvInfo) error {
	done := make(chan bool, 1)
	check := func(ctx context.Context) {
		if !s.cache.Get(key, result) {
			log.SpanLog(ctx, log.DebugLevelApi, "wait for AutoProvInfo state info not found", "key", key)
			return
		}
		if result.MaintenanceState == targetState || result.MaintenanceState == errorState {
			done <- true
		}
	}

	log.SpanLog(ctx, log.DebugLevelApi, "wait for AutoProvInfo state", "target", targetState)

	cancel := s.cache.WatchKey(key, check)

	// after setting up watch, check current state,
	// as it may have already changed to the target state
	check(ctx)

	var err error
	select {
	case <-done:
	case <-time.After(timeout):
		err = fmt.Errorf("timed out waiting for AutoProvInfo maintenance state")
	}
	cancel()

	return err
}
