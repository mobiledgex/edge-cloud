// Run redis as a child process for testing

package main

import (
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
)

func StartLocalRedisServer(opts ...process.StartOp) (*process.RedisCache, error) {
	redis := &process.RedisCache{
		Common: process.Common{
			Name: "redis-local",
		},
		Type: "master",
	}
	log.InfoLog("Starting local redis")
	err := redis.StartLocal("", opts...)
	if err != nil {
		return nil, err
	}
	return redis, nil
}
