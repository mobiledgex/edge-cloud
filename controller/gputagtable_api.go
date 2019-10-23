package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type GpuTagTableApi struct {
	sync  *Sync
	store edgeproto.GpuTagTableStore
	cache edgeproto.GpuTagTableCache
}

var gpuTagTableApi = GpuTagTableApi{}

func InitGpuTagTableApi(sync *Sync) {
	gpuTagTableApi.sync = sync
	gpuTagTableApi.store = edgeproto.NewGpuTagTableStore(sync.store)
	edgeproto.InitGpuTagTableCache(&gpuTagTableApi.cache)
	sync.RegisterCache(&gpuTagTableApi.cache)
}

func (s *GpuTagTableApi) CreateGpuTagTable(ctx context.Context, in *edgeproto.GpuTagTable) (*edgeproto.Result, error) {

	if err := in.Validate(edgeproto.GpuTagTableAllFieldsMap); err != nil {
		return &edgeproto.Result{}, err
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return objstore.ErrKVStoreKeyExists
		}
		s.store.STMPut(stm, in)
		return nil
	})

	return &edgeproto.Result{}, err
}

func (s *GpuTagTableApi) DeleteGpuTagTable(ctx context.Context, in *edgeproto.GpuTagTable) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, nil) {
			return objstore.ErrKVStoreKeyNotFound
		}
		s.store.STMDel(stm, &in.Key)
		return nil
	})

	return &edgeproto.Result{}, err
}

func (s *GpuTagTableApi) GetGpuTagTable(ctx context.Context, in *edgeproto.GpuTagTableKey) (*edgeproto.GpuTagTable, error) {
	var tbl edgeproto.GpuTagTable
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, in, &tbl) {
			return objstore.ErrKVStoreKeyNotFound
		}
		return nil
	})

	return &tbl, err
}

func (s *GpuTagTableApi) ShowGpuTagTable(in *edgeproto.GpuTagTable, cb edgeproto.GpuTagTableApi_ShowGpuTagTableServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.GpuTagTable) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *GpuTagTableApi) UpdateGpuTagTable(ctx context.Context, in *edgeproto.GpuTagTable) (*edgeproto.Result, error) {

	for _, field := range in.Fields {
		if field == edgeproto.GpuTagTableFieldTags {
			// which to avoid duplicate tag values
			for _, f := range in.Tags {
				if field == f {
					return &edgeproto.Result{}, nil
				}
			}
		}
	}

	return s.store.Update(ctx, in, s.sync.syncWait)

}

func (s *GpuTagTableApi) AddGpuTag(ctx context.Context, in *edgeproto.GpuTagTable) (*edgeproto.Result, error) {
	var tbl edgeproto.GpuTagTable
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &tbl) {
			return objstore.ErrKVStoreKeyNotFound
		}
		for _, t := range tbl.Tags {
			if t == in.Tags[0] {
				return fmt.Errorf("Duplicate Tag Found %s", in.Tags[0])
			}
		}
		tbl.Tags = append(tbl.Tags, in.Tags[0])
		s.store.STMPut(stm, &tbl)
		return nil
	})

	return &edgeproto.Result{}, err
}

func (s *GpuTagTableApi) RemoveGpuTag(ctx context.Context, in *edgeproto.GpuTagTable) (*edgeproto.Result, error) {
	var tbl edgeproto.GpuTagTable
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &tbl) {
			return objstore.ErrKVStoreKeyNotFound
		}
		for i, t := range tbl.Tags {
			if t == in.Tags[0] {
				fmt.Printf("\n\tRemoveGpuTag found tag %s at i = %d\n", t, i)

				tbl.Tags[i] = tbl.Tags[len(tbl.Tags)-1]
				tbl.Tags[len(tbl.Tags)-1] = ""
				tbl.Tags = tbl.Tags[:len(tbl.Tags)-1]
			}
		}
		s.store.STMPut(stm, &tbl)
		return nil
	})

	return &edgeproto.Result{}, err

}
