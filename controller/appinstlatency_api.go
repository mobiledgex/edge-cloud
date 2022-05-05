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

	"github.com/edgexr/edge-cloud/cloudcommon/node"

	"github.com/edgexr/edge-cloud/edgeproto"
)

type AppInstLatencyApi struct {
	all  *AllApis
	sync *Sync
}

func NewAppInstLatencyApi(sync *Sync, all *AllApis) *AppInstLatencyApi {
	appInstLatencyApi := AppInstLatencyApi{}
	appInstLatencyApi.all = all
	appInstLatencyApi.sync = sync
	return &appInstLatencyApi
}

func (s *AppInstLatencyApi) RequestAppInstLatency(ctx context.Context, in *edgeproto.AppInstLatency) (*edgeproto.Result, error) {
	SetAppInstKeyDefaults(ctx, &in.Key, s.all)

	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}
	// Check that appinst exists
	appInstInfo := edgeproto.AppInst{}
	if !s.all.appInstApi.cache.Get(&in.Key, &appInstInfo) {
		return nil, in.Key.NotFoundError()
	}

	conn, err := notifyRootConnect(ctx, *notifyRootAddrs)
	if err != nil {
		return nil, err
	}
	client := edgeproto.NewAppInstLatencyApiClient(conn)
	ctx, cancel := context.WithTimeout(ctx, node.DefaultDebugTimeout)
	defer cancel()
	return client.RequestAppInstLatency(ctx, in)
}
