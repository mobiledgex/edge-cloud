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
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type NetworkApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.NetworkStore
	cache edgeproto.NetworkCache
}

func NewNetworkApi(sync *Sync, all *AllApis) *NetworkApi {
	networkApi := NetworkApi{}
	networkApi.all = all
	networkApi.sync = sync
	networkApi.store = edgeproto.NewNetworkStore(sync.store)
	edgeproto.InitNetworkCache(&networkApi.cache)
	sync.RegisterCache(&networkApi.cache)
	return &networkApi
}

func (s *NetworkApi) CreateNetwork(in *edgeproto.Network, cb edgeproto.NetworkApi_CreateNetworkServer) error {
	ctx := cb.Context()
	log.SpanLog(ctx, log.DebugLevelApi, "CreateNetwork", "network", in)
	if err := in.Validate(nil); err != nil {
		return err
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.Network{}
		if s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.ExistsError()
		}
		cloudlet := edgeproto.Cloudlet{}
		if !s.all.cloudletApi.store.STMGet(stm, &in.Key.CloudletKey, &cloudlet) {
			return in.Key.CloudletKey.NotFoundError()
		}
		if cloudlet.DeletePrepare {
			return in.Key.CloudletKey.BeingDeletedError()
		}
		s.store.STMPut(stm, in)
		return nil
	})
	return err
}

func (s *NetworkApi) UpdateNetwork(in *edgeproto.Network, cb edgeproto.NetworkApi_UpdateNetworkServer) error {
	ctx := cb.Context()
	cur := edgeproto.Network{}

	if s.all.clusterInstApi.UsesNetwork(&in.Key) != nil {
		return errors.New("Network cannot be modified while associated with a Cluster Instance")
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		changed := 0
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		changed = cur.CopyInFields(in)
		if err := cur.Validate(nil); err != nil {
			return err
		}
		if changed == 0 {
			return nil
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *NetworkApi) DeleteNetwork(in *edgeproto.Network, cb edgeproto.NetworkApi_DeleteNetworkServer) (reterr error) {
	ctx := cb.Context()
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.Network{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		if cur.DeletePrepare {
			return in.Key.BeingDeletedError()
		}
		cur.DeletePrepare = true
		s.store.STMPut(stm, &cur)
		return nil
	})
	if err != nil {
		return err
	}
	defer func() {
		if reterr == nil {
			return
		}
		undoErr := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			cur := edgeproto.Network{}
			if !s.store.STMGet(stm, &in.Key, &cur) {
				return nil
			}
			if cur.DeletePrepare {
				cur.DeletePrepare = false
				s.store.STMPut(stm, &cur)
			}
			return nil
		})
		if undoErr != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to undo delete prepare", "key", in.Key, "err", undoErr)
		}
	}()

	if k := s.all.clusterInstApi.UsesNetwork(&in.Key); k != nil {
		return fmt.Errorf("Network in use by ClusterInst %s", k.GetKeyString())
	}
	_, err = s.store.Delete(ctx, in, s.sync.syncWait)
	return err
}

func (s *NetworkApi) ShowNetwork(in *edgeproto.Network, cb edgeproto.NetworkApi_ShowNetworkServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Network) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *NetworkApi) UsesCloudlet(in *edgeproto.CloudletKey) *edgeproto.NetworkKey {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for key := range s.cache.Objs {
		if key.CloudletKey.Matches(in) {
			return &key
		}
	}
	return nil
}
