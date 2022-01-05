package rediscache

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

const MaxRedisWait = time.Second * 30

// Special IDs in the streams API
const RedisSmallestId = "-"
const RedisGreatestId = "+"
const RedisLastId = "$"

func NewClient(redisAddr string) (*redis.Client, error) {
	if redisAddr == "" {
		return nil, fmt.Errorf("Missing redis addr")
	}
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return client, nil
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
