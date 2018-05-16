package main

import (
	"context"
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
)

type OperatorApi struct {
	edgeproto.ObjStore
	operators map[edgeproto.OperatorKey]*edgeproto.Operator
	mux       util.Mutex
}

func InitOperatorApi(objStore edgeproto.ObjStore) *OperatorApi {
	api := &OperatorApi{ObjStore: objStore}
	api.operators = make(map[edgeproto.OperatorKey]*edgeproto.Operator)

	api.mux.Lock()
	defer api.mux.Unlock()

	err := edgeproto.LoadAllOperators(api, func(obj *edgeproto.Operator) error {
		api.operators[obj.Key] = obj
		return nil
	})
	if err != nil && err == context.DeadlineExceeded {
		util.WarnLog("Init Operators failed", "error", err)
	}
	return api
}

func (s *OperatorApi) ValidateKey(key *edgeproto.OperatorKey) error {
	if key == nil {
		return errors.New("Operator key not specified")
	}
	if !util.ValidName(key.Name) {
		return errors.New("Invalid operator name")
	}
	return nil
}

func (s *OperatorApi) Validate(in *edgeproto.Operator) error {
	return s.ValidateKey(&in.Key)
}

func (s *OperatorApi) GetObjStoreKeyString(key *edgeproto.OperatorKey) string {
	return GetObjStoreKey(OperatorType, key.GetKeyString())
}

func (s *OperatorApi) GetLoadKeyString() string {
	return GetObjStoreKey(OperatorType, "")
}

func (s *OperatorApi) HasOperator(key *edgeproto.OperatorKey) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	_, found := s.operators[*key]
	return found
}

func (s *OperatorApi) Refresh(in *edgeproto.Operator, key string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	obj, err := edgeproto.LoadOneOperator(s, key)
	if err == nil {
		s.operators[in.Key] = obj
	} else if err == edgeproto.ObjStoreErrKeyNotFound {
		delete(s.operators, in.Key)
		err = nil
	}
	return err
}

func (s *OperatorApi) CreateOperator(ctx context.Context, in *edgeproto.Operator) (*edgeproto.Result, error) {
	return in.Create(s)
}

func (s *OperatorApi) UpdateOperator(ctx context.Context, in *edgeproto.Operator) (*edgeproto.Result, error) {
	return in.Update(s)
}

func (s *OperatorApi) DeleteOperator(ctx context.Context, in *edgeproto.Operator) (*edgeproto.Result, error) {
	return in.Delete(s)
}

func (s *OperatorApi) ShowOperator(in *edgeproto.Operator, cb edgeproto.OperatorApi_ShowOperatorServer) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	for _, myOper := range s.operators {
		if !myOper.Matches(in) {
			continue
		}
		err := cb.Send(myOper)
		if err != nil {
			return err
		}
	}
	return nil
}
