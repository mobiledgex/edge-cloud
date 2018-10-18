package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type DeveloperApi struct {
	sql edgeproto.DeveloperSql
}

var developerApi = DeveloperApi{}

func InitDeveloperApi(sql *sql.DB) {
	developerApi.sql.Init(sql)
}

func (s *DeveloperApi) RequireDeveloper(key *edgeproto.DeveloperKey) error {
	found, err := s.sql.HasKey(key)
	if err != nil {
		return fmt.Errorf("look up developer %s failed, %s",
			key.Name, err.Error())
	}
	if !found {
		return fmt.Errorf("developer %s not found", key.Name)
	}
	return nil
}

func (s *DeveloperApi) CreateDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	_, err := s.sql.Create(in)
	return &edgeproto.Result{}, err
}

func (s *DeveloperApi) UpdateDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	_, err := s.sql.Update(in)
	return &edgeproto.Result{}, err
}

func (s *DeveloperApi) DeleteDeveloper(ctx context.Context, in *edgeproto.Developer) (*edgeproto.Result, error) {
	if appApi.UsesDeveloper(&in.Key) {
		return &edgeproto.Result{}, errors.New("Developer in use by Application")
	}
	_, err := s.sql.Delete(in)
	return &edgeproto.Result{}, err
}

func (s *DeveloperApi) ShowDeveloper(in *edgeproto.Developer, cb edgeproto.DeveloperApi_ShowDeveloperServer) error {
	err := s.sql.Show(in, func(obj *edgeproto.Developer) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
