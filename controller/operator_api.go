package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type OperatorApi struct {
	sync  *Sync
	store edgeproto.OperatorStore
	cache edgeproto.OperatorCache
}

var operatorApi = OperatorApi{}

func InitOperatorApi(sync *Sync) {
	operatorApi.sync = sync
	operatorApi.store = edgeproto.NewOperatorStore(sync.store)
	edgeproto.InitOperatorCache(&operatorApi.cache)
	sync.RegisterCache(&operatorApi.cache)
}

func (s *OperatorApi) HasOperator(key *edgeproto.OperatorKey) bool {
	return s.cache.HasKey(key)
}

func (s *OperatorApi) CreateOperator(ctx context.Context, in *edgeproto.Operator) (*edgeproto.Result, error) {
	return s.store.Create(in, s.sync.syncWait)
}

func (s *OperatorApi) UpdateOperator(ctx context.Context, in *edgeproto.Operator) (*edgeproto.Result, error) {
	return s.store.Update(in, s.sync.syncWait)
}

func (s *OperatorApi) DeleteOperator(ctx context.Context, in *edgeproto.Operator) (*edgeproto.Result, error) {
	if cloudletApi.UsesOperator(&in.Key) {
		return &edgeproto.Result{}, errors.New("Operator in use by Cloudlet")
	}
	return s.store.Delete(in, s.sync.syncWait)
}

func (s *OperatorApi) ShowOperator(in *edgeproto.Operator, cb edgeproto.OperatorApi_ShowOperatorServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Operator) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
