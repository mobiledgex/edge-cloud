package rediscache

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
)

func MockRedisServer() (*miniredis.Miniredis, error) {
	serv, err := miniredis.Run()
	if err != nil {
		return nil, err
	}
	return serv, nil
}

func NewDummyRedisClient() (*redis.Client, error) {
	redisServer, err := MockRedisServer()
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	})
	return client, nil
}
