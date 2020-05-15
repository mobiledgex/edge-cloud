package main

import (
	"context"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type CloudletPoolApi struct {
	sync  *Sync
	store edgeproto.CloudletPoolStore
	cache edgeproto.CloudletPoolCache
}

var cloudletPoolApi = CloudletPoolApi{}

func InitCloudletPoolApi(sync *Sync) {
	cloudletPoolApi.sync = sync
	cloudletPoolApi.store = edgeproto.NewCloudletPoolStore(sync.store)
	edgeproto.InitCloudletPoolCache(&cloudletPoolApi.cache)
	sync.RegisterCache(&cloudletPoolApi.cache)
}

func (s *CloudletPoolApi) CreateCloudletPool(ctx context.Context, in *edgeproto.CloudletPool) (*edgeproto.Result, error) {
	if err := in.Validate(edgeproto.CloudletPoolAllFieldsMap); err != nil {
		return &edgeproto.Result{}, err
	}

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.ExistsError()
		}
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) DeleteCloudletPool(ctx context.Context, in *edgeproto.CloudletPool) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.NotFoundError()
		}
		s.store.STMDel(stm, &in.Key)
		return nil
	})
	if err == nil {
		cloudletPoolMemberApi.poolDeleted(ctx, &in.Key)
	}
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) ShowCloudletPool(in *edgeproto.CloudletPool, cb edgeproto.CloudletPoolApi_ShowCloudletPoolServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.CloudletPool) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *CloudletPoolApi) showPoolsByKeys(keys map[edgeproto.CloudletPoolKey]struct{}, cb func(obj *edgeproto.CloudletPool) error) error {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()

	for key, data := range s.cache.Objs {
		if _, found := keys[key]; !found {
			continue
		}
		err := cb(data.Obj)
		if err != nil {
			return err
		}
	}
	return nil
}
