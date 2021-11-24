package redisclient

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

const MaxRedisWait = time.Second * 30

func NewClient(redisAddr string) (*redis.Client, error) {
	if redisAddr == "" {
		return nil, fmt.Errorf("Missing redis addr")
	}
	return redis.NewClient(&redis.Options{
		Addr: redisAddr,
	}), nil
}

func IsServerReady(client *redis.Client) error {
	start := time.Now()
	var err error
	for {
		_, err = client.Ping().Result()
		if err == nil {
			return nil
		}
		elapsed := time.Since(start)
		if elapsed >= (MaxRedisWait) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("Failed to ping redis - %v", err)
}
