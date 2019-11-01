package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type ResTagTableApi struct {
	sync  *Sync
	store edgeproto.ResTagTableStore
	cache edgeproto.ResTagTableCache
}

var resTagTableApi = ResTagTableApi{}

func InitResTagTableApi(sync *Sync) {
	resTagTableApi.sync = sync
	resTagTableApi.store = edgeproto.NewResTagTableStore(sync.store)
	edgeproto.InitResTagTableCache(&resTagTableApi.cache)
	sync.RegisterCache(&resTagTableApi.cache)
}

func (s *ResTagTableApi) CreateResTagTable(ctx context.Context, in *edgeproto.ResTagTable) (*edgeproto.Result, error) {

	if err := in.Validate(edgeproto.ResTagTableAllFieldsMap); err != nil {
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

func (s *ResTagTableApi) DeleteResTagTable(ctx context.Context, in *edgeproto.ResTagTable) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, nil) {
			return objstore.ErrKVStoreKeyNotFound
		}
		s.store.STMDel(stm, &in.Key)
		return nil
	})

	return &edgeproto.Result{}, err
}

func (s *ResTagTableApi) GetResTagTable(ctx context.Context, in *edgeproto.ResTagTableKey) (*edgeproto.ResTagTable, error) {
	var tbl edgeproto.ResTagTable
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, in, &tbl) {
			return objstore.ErrKVStoreKeyNotFound
		}
		return nil
	})

	return &tbl, err
}

func (s *ResTagTableApi) ShowResTagTable(in *edgeproto.ResTagTable, cb edgeproto.ResTagTableApi_ShowResTagTableServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.ResTagTable) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *ResTagTableApi) UpdateResTagTable(ctx context.Context, in *edgeproto.ResTagTable) (*edgeproto.Result, error) {
	// currently unused, we may use for a "table rename" at some point
	return &edgeproto.Result{}, nil

}

func (s *ResTagTableApi) AddResTag(ctx context.Context, in *edgeproto.ResTagTable) (*edgeproto.Result, error) {
	var tbl edgeproto.ResTagTable

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

func (s *ResTagTableApi) RemoveResTag(ctx context.Context, in *edgeproto.ResTagTable) (*edgeproto.Result, error) {
	var tbl edgeproto.ResTagTable
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

func (s *ResTagTableApi) AddZone(ctx context.Context, in *edgeproto.ResTagTable) (*edgeproto.Result, error) {
	var tbl edgeproto.ResTagTable

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &tbl) {
			return objstore.ErrKVStoreKeyNotFound
		}
		tbl.Azone = in.Azone
		s.store.STMPut(stm, &tbl)
		return nil
	})

	return &edgeproto.Result{}, err
}

func (s *ResTagTableApi) RemoveZone(ctx context.Context, in *edgeproto.ResTagTable) (*edgeproto.Result, error) {
	var tbl edgeproto.ResTagTable
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &tbl) {
			return objstore.ErrKVStoreKeyNotFound
		}
		tbl.Azone = ""
		s.store.STMPut(stm, &tbl)
		return nil
	})

	return &edgeproto.Result{}, err

}
