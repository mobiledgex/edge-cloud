// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"

	"github.com/edgexr/edge-cloud/edgeproto"
)

type OperatorCodeApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.OperatorCodeStore
	cache edgeproto.OperatorCodeCache
}

func NewOperatorCodeApi(sync *Sync, all *AllApis) *OperatorCodeApi {
	operatorCodeApi := OperatorCodeApi{}
	operatorCodeApi.all = all
	operatorCodeApi.sync = sync
	operatorCodeApi.store = edgeproto.NewOperatorCodeStore(sync.store)
	edgeproto.InitOperatorCodeCache(&operatorCodeApi.cache)
	sync.RegisterCache(&operatorCodeApi.cache)
	return &operatorCodeApi
}

func (s *OperatorCodeApi) CreateOperatorCode(ctx context.Context, in *edgeproto.OperatorCode) (*edgeproto.Result, error) {
	if err := in.Validate(nil); err != nil {
		return &edgeproto.Result{}, err
	}
	return s.store.Create(ctx, in, s.sync.syncWait)
}

func (s *OperatorCodeApi) DeleteOperatorCode(ctx context.Context, in *edgeproto.OperatorCode) (*edgeproto.Result, error) {
	if !s.cache.HasKey(in.GetKey()) {
		return &edgeproto.Result{}, in.GetKey().NotFoundError()
	}
	return s.store.Delete(ctx, in, s.sync.syncWait)
}

func (s *OperatorCodeApi) ShowOperatorCode(in *edgeproto.OperatorCode, cb edgeproto.OperatorCodeApi_ShowOperatorCodeServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.OperatorCode) error {
		err := cb.Send(obj)
		return err
	})
	return err
}
