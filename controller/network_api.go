package main

import (
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type NetworkApi struct {
	sync  *Sync
	store edgeproto.NetworkStore
	cache edgeproto.NetworkCache
}

var networkApi = NetworkApi{}

func InitNetworkApi(sync *Sync) {
	networkApi.sync = sync
	networkApi.store = edgeproto.NewNetworkStore(sync.store)
	edgeproto.InitNetworkCache(&networkApi.cache)
	sync.RegisterCache(&networkApi.cache)
}

func (s *NetworkApi) CreateNetwork(in *edgeproto.Network, cb edgeproto.NetworkApi_CreateNetworkServer) error {
	ctx := cb.Context()
	log.SpanLog(ctx, log.DebugLevelApi, "CreateNetwork", "network", in)
	if err := in.Validate(nil); err != nil {
		return err
	}
	_, err := s.store.Create(ctx, in, s.sync.syncWait)
	return err

}

func (s *NetworkApi) UpdateNetwork(in *edgeproto.Network, cb edgeproto.NetworkApi_UpdateNetworkServer) error {
	ctx := cb.Context()
	cur := edgeproto.Network{}
	changed := 0

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
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

func (s *NetworkApi) DeleteNetwork(in *edgeproto.Network, cb edgeproto.NetworkApi_DeleteNetworkServer) error {
	ctx := cb.Context()
	if !s.cache.HasKey(&in.Key) {
		return in.Key.NotFoundError()
	}
	_, err := s.store.Delete(ctx, in, s.sync.syncWait)
	return err
}

func (s *NetworkApi) ShowNetwork(in *edgeproto.Network, cb edgeproto.NetworkApi_ShowNetworkServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Network) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *NetworkApi) STMFind(stm concurrency.STM, name, org string, network *edgeproto.Network) error {
	key := edgeproto.NetworkKey{}
	key.Name = name
	key.Organization = org
	if !s.store.STMGet(stm, &key, network) {
		return fmt.Errorf("Network %s for organization %s not found", name, org)
	}
	return nil
}

func (s *NetworkApi) GetNetworks(networks map[edgeproto.NetworkKey]*edgeproto.Network) {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for _, data := range s.cache.Objs {
		net := data.Obj
		copy := &edgeproto.Network{}
		copy.DeepCopyIn(net)
		networks[net.Key] = copy
	}
}
