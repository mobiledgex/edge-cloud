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
	// currently unused, we may use for a "table rename" at some point
	return &edgeproto.Result{}, nil

}

func (s *GpuTagTableApi) AddGpuTag(ctx context.Context, in *edgeproto.GpuTagTable) (*edgeproto.Result, error) {
	var tbl edgeproto.GpuTagTable

	// validate input, don't accept dup tag values in any  multi-tag input
	if len(in.Tags) > 1 {
		for i, ctag := range in.Tags {
			for _, ttag := range in.Tags[i+1 : len(in.Tags)] {
				if ctag == ttag {
					return &edgeproto.Result{}, fmt.Errorf("Duplicate Tag Found %s in multi-tag input", ctag)
				}
				if i == len(in.Tags)-1 {
					break
				}
			}
		}
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &tbl) {
			return objstore.ErrKVStoreKeyNotFound
		}

		for _, t := range in.Tags {
			for _, tt := range tbl.Tags {
				if t == tt {
					return fmt.Errorf("Duplicate Tag Found %s", t)
				}
			}
		}
		for _, t := range in.Tags {
			tbl.Tags = append(tbl.Tags, t)
		}
		s.store.STMPut(stm, &tbl)
		return nil
	})

	return &edgeproto.Result{}, err
}

func (s *GpuTagTableApi) RemoveGpuTag(ctx context.Context, in *edgeproto.GpuTagTable) (*edgeproto.Result, error) {
	var tbl edgeproto.GpuTagTable
	// validate input, reject dup tags in list
	if len(in.Tags) > 1 {
		for i, ctag := range in.Tags {
			for _, ttag := range in.Tags[i+1 : len(in.Tags)] {
				if ctag == ttag {
					return &edgeproto.Result{}, fmt.Errorf("Duplicate Tag Found %s in multi-tag input", ctag)
				}
				if i == len(in.Tags)-1 {
					break
				}
			}
		}
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &tbl) {
			return objstore.ErrKVStoreKeyNotFound
		}
		for _, t := range in.Tags {
			for j, tag := range tbl.Tags {

				if t == tag {
					tbl.Tags[j] = tbl.Tags[len(tbl.Tags)-1]
					tbl.Tags[len(tbl.Tags)-1] = ""
					tbl.Tags = tbl.Tags[:len(tbl.Tags)-1]
				}
			}
		}
		s.store.STMPut(stm, &tbl)
		return nil
	})

	return &edgeproto.Result{}, err

}
