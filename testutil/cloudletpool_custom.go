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

package testutil

import (
	"context"
	fmt "fmt"

	"github.com/edgexr/edge-cloud/edgeproto"
)

func (s *DummyServer) AddCloudletPoolMember(ctx context.Context, in *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	cache := &s.CloudletPoolCache

	cache.Mux.Lock()
	defer cache.Mux.Unlock()
	data, found := cache.Objs[in.Key]
	if !found {
		return &edgeproto.Result{}, in.Key.NotFoundError()
	}
	for ii, _ := range data.Obj.Cloudlets {
		if data.Obj.Cloudlets[ii].Matches(&in.Cloudlet) {
			return &edgeproto.Result{}, fmt.Errorf("Already exists")
		}
	}
	data.Obj.Cloudlets = append(data.Obj.Cloudlets, in.Cloudlet)

	return &edgeproto.Result{}, nil
}

func (s *DummyServer) RemoveCloudletPoolMember(ctx context.Context, in *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	cache := &s.CloudletPoolCache

	cache.Mux.Lock()
	defer cache.Mux.Unlock()
	data, found := cache.Objs[in.Key]
	if !found {
		return &edgeproto.Result{}, in.Key.NotFoundError()
	}
	for ii, cloudletKey := range data.Obj.Cloudlets {
		if cloudletKey.Matches(&in.Cloudlet) {
			data.Obj.Cloudlets = append(data.Obj.Cloudlets[:ii], data.Obj.Cloudlets[ii+1:]...)
			break
		}
	}

	return &edgeproto.Result{}, nil
}
