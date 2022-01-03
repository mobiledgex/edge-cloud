package rediscache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/mobiledgex/edge-cloud/log"
)

const MaxRedisWait = time.Second * 30

type RedisCache interface {
	IsServerReady() error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Del(ctx context.Context, keys ...string) (int64, error)
	Subscribe(ctx context.Context, channels ...string) (RedisPubSub, error)
	Publish(ctx context.Context, channel string, message interface{}) error
}

type RedisClient struct {
	redisAddr string
	client    *redis.Client
	pubsubs   map[string]*redis.PubSub
}

func NewClient(redisAddr string) (*RedisClient, error) {
	if redisAddr == "" {
		return nil, fmt.Errorf("Missing redis addr")
	}
	redisClient := &RedisClient{}
	redisClient.redisAddr = redisAddr
	redisClient.client = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	redisClient.pubsubs = make(map[string]*redis.PubSub)
	return redisClient, nil
}

func (r *RedisClient) IsServerReady() error {
	start := time.Now()
	var err error
	for {
		_, err = r.client.Ping().Result()
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

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	out, err := r.client.Get(key).Result()
	log.SpanLog(ctx, log.DebugLevelRedis, "got data", "key", key, "val", out, "err", err)
	return out, err
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error) {
	out, err := r.client.Set(key, value, expiration).Result()
	log.SpanLog(ctx, log.DebugLevelRedis, "set data", "key", key, "val", value,
		"expiration", expiration, "out", out, "err", err)
	return out, err
}

func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	out, err := r.client.SetNX(key, value, expiration).Result()
	log.SpanLog(ctx, log.DebugLevelRedis, "set data if not exists", "key", key, "val", value,
		"expiration", expiration, "out", out, "err", err)
	return out, err
}

func (r *RedisClient) Del(ctx context.Context, keys ...string) (int64, error) {
	out, err := r.client.Del(keys...).Result()
	log.SpanLog(ctx, log.DebugLevelRedis, "del data", "keys", keys, "out", out, "err", err)
	return out, err
}

func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) (RedisPubSub, error) {
	pubsub := r.client.Subscribe(channels...)

	// Wait for confirmation that subscription is created before publishing anything.
	_, err := pubsub.Receive()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelRedis, "failed to subscribe to channels",
			"channels", channels, "err", err)
		return nil, err
	}

	log.SpanLog(ctx, log.DebugLevelRedis, "subscribed to channels", "channels", channels)
	return &redisPubSub{pubsub: pubsub}, nil
}

func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	err := r.client.Publish(channel, message).Err()
	log.SpanLog(ctx, log.DebugLevelApi, "publish message on redis channel",
		"channel", channel, "message", message, "err", err)
	return err
}

type RedisPubSub interface {
	Channel() <-chan *redis.Message
	Close() error
}

type redisPubSub struct {
	pubsub *redis.PubSub
}

func (p *redisPubSub) Channel() <-chan *redis.Message {
	if p == nil {
		return nil
	}
	return p.pubsub.Channel()
}

func (p *redisPubSub) Close() error {
	if p == nil {
		return nil
	}
	// Close() also closes subscribed channels
	return p.pubsub.Close()
}
