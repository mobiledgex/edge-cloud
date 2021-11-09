package main

import (
	"errors"

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
	_, err := s.store.Create(ctx, in, s.sync.syncWait)
	return err
}

func (s *NetworkApi) UpdateNetwork(in *edgeproto.Network, cb edgeproto.NetworkApi_UpdateNetworkServer) error {
	ctx := cb.Context()
	cur := edgeproto.Network{}

	if s.all.clusterInstApi.UsesNetwork(&in.Key) {
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

func (s *NetworkApi) DeleteNetwork(in *edgeproto.Network, cb edgeproto.NetworkApi_DeleteNetworkServer) error {
	ctx := cb.Context()
	if !s.cache.HasKey(&in.Key) {
		return in.Key.NotFoundError()
	}
	if s.all.clusterInstApi.UsesNetwork(&in.Key) {
		return errors.New("Network in use by Cluster Instance")
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
