package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type OperatorApi struct {
	sql edgeproto.OperatorSql
}

var operatorApi = OperatorApi{}

func InitOperatorApi(sql *sql.DB) {
	operatorApi.sql.Init(sql)
}

func (s *OperatorApi) RequireOperator(key *edgeproto.OperatorKey) error {
	found, err := s.sql.HasKey(key)
	if err != nil {
		return fmt.Errorf("look up operator %s failed, %s",
			key.Name, err.Error())
	}
	if !found {
		return fmt.Errorf("operator %s not found", key.Name)
	}
	return nil
}

func (s *OperatorApi) CreateOperator(ctx context.Context, in *edgeproto.Operator) (*edgeproto.Result, error) {
	if in.Key.Name == cloudcommon.OperatorDeveloper {
		return nil, errors.New("Cannot create operator with name = " + cloudcommon.OperatorDeveloper)
	}
	_, err := s.sql.Create(in)
	return &edgeproto.Result{}, err
}

func (s *OperatorApi) UpdateOperator(ctx context.Context, in *edgeproto.Operator) (*edgeproto.Result, error) {
	_, err := s.sql.Update(in)
	return &edgeproto.Result{}, err
}

func (s *OperatorApi) DeleteOperator(ctx context.Context, in *edgeproto.Operator) (*edgeproto.Result, error) {
	if cloudletApi.UsesOperator(&in.Key) {
		return &edgeproto.Result{}, errors.New("Operator in use by Cloudlet")
	}
	_, err := s.sql.Delete(in)
	return &edgeproto.Result{}, err
}

func (s *OperatorApi) ShowOperator(in *edgeproto.Operator, cb edgeproto.OperatorApi_ShowOperatorServer) error {
	err := s.sql.Show(in, func(obj *edgeproto.Operator) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
