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

package rediscache

import (
	"time"

	"github.com/Bose/minisentinel"
	"github.com/alicebob/miniredis/v2"
)

type DummyRedis struct {
	redisSrv    *miniredis.Miniredis
	sentinelSrv *minisentinel.Sentinel
}

func NewMockRedisServer() (*DummyRedis, error) {
	redisSrv, err := miniredis.Run()
	if err != nil {
		return nil, err
	}

	sentinelSrv := minisentinel.NewSentinel(
		redisSrv,
		minisentinel.WithReplica(redisSrv),
		minisentinel.WithMasterName("redismaster"),
	)
	err = sentinelSrv.Start()
	if err != nil {
		return nil, err
	}
	dummyRedis := DummyRedis{
		redisSrv:    redisSrv,
		sentinelSrv: sentinelSrv,
	}
	return &dummyRedis, nil
}

func (r *DummyRedis) GetStandaloneAddr() string {
	return r.redisSrv.Addr()
}

func (r *DummyRedis) GetSentinelAddr() string {
	return r.sentinelSrv.Addr()
}

func (r *DummyRedis) FastForward(d time.Duration) {
	r.redisSrv.FastForward(d)
}

func (r *DummyRedis) Close() {
	r.sentinelSrv.Close()
	r.redisSrv.Close()
}
