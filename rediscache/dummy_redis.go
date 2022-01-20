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
