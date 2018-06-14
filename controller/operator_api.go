package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/objstore"
)

type OperatorApi struct {
	store edgeproto.OperatorStore
	cache *edgeproto.OperatorCache
}

func InitOperatorApi(objStore objstore.ObjStore) *OperatorApi {
	api := &OperatorApi{
		store: edgeproto.NewOperatorStore(objStore),
	}
	api.cache = edgeproto.NewOperatorCache(&api.store)
	return api
}

func (s *OperatorApi) WaitInitDone() {
	s.cache.WaitInitSyncDone()
}

func (s *OperatorApi) HasOperator(key *edgeproto.OperatorKey) bool {
	return s.cache.HasKey(key)
}

func (s *OperatorApi) CreateOperator(ctx context.Context, in *edgeproto.Operator) (*edgeproto.Result, error) {
	return s.store.Create(in, s.cache.SyncWait)
}

func (s *OperatorApi) UpdateOperator(ctx context.Context, in *edgeproto.Operator) (*edgeproto.Result, error) {
	return s.store.Update(in, s.cache.SyncWait)
}

func (s *OperatorApi) DeleteOperator(ctx context.Context, in *edgeproto.Operator) (*edgeproto.Result, error) {
	return s.store.Delete(in, s.cache.SyncWait)
}

func (s *OperatorApi) ShowOperator(in *edgeproto.Operator, cb edgeproto.OperatorApi_ShowOperatorServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Operator) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
