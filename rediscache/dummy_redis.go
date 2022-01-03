package rediscache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/mobiledgex/edge-cloud/util"
)

type dummyData struct {
	val             string
	lastUpdatedTime time.Time
	expirationTime  time.Duration
}

type DummyRedis struct {
	db     map[string]*dummyData
	msgChs map[string]chan *redis.Message
	mux    util.Mutex
}

func NewDummyRedis() *DummyRedis {
	rdb := DummyRedis{}
	rdb.db = make(map[string]*dummyData)
	rdb.msgChs = make(map[string]chan *redis.Message)
	return &rdb
}

func (r *DummyRedis) IsServerReady() error {
	return nil
}

func (r *DummyRedis) Get(ctx context.Context, key string) (string, error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.db == nil {
		return "", fmt.Errorf("redis not initialized")
	}
	data, ok := r.db[key]
	if !ok {
		return "", redis.Nil
	}
	if data.expirationTime > 0 &&
		(time.Since(data.lastUpdatedTime) > data.expirationTime) {
		// key has expired, delete it
		delete(r.db, key)
		return "", redis.Nil
	}
	return data.val, nil
}

func (r *DummyRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.db == nil {
		return "", fmt.Errorf("redis not initialized")
	}
	// value can be any object that implement `encoding.BinaryMarshaler`,
	// but for now keep string as the only supported type for testing
	valstr, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("invalid value, must be string")
	}
	r.db[key] = &dummyData{
		val:             valstr,
		lastUpdatedTime: time.Now(),
		expirationTime:  expiration,
	}
	return "OK", nil
}

func (r *DummyRedis) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.db == nil {
		return false, fmt.Errorf("redis not initialized")
	}
	// do not set if key already exists
	if _, ok := r.db[key]; ok {
		return false, nil
	}
	// value can be any object that implement `encoding.BinaryMarshaler`,
	// but for now keep string as the only supported type for testing
	valstr, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("invalid value, must be string")
	}
	r.db[key] = &dummyData{
		val:             valstr,
		lastUpdatedTime: time.Now(),
		expirationTime:  expiration,
	}
	return true, nil
}

func (r *DummyRedis) Del(ctx context.Context, keys ...string) (int64, error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.db == nil {
		return 0, fmt.Errorf("redis not initialized")
	}
	keysRem := int64(0)
	for _, key := range keys {
		if _, ok := r.db[key]; ok {
			delete(r.db, key)
			keysRem++
		}
	}
	return keysRem, nil
}

func (r *DummyRedis) Subscribe(ctx context.Context, channels ...string) (RedisPubSub, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if len(channels) != 1 {
		return nil, fmt.Errorf("for dummy redis server, only one channel is supported")
	}

	pubsub := dummyRedisPubSub{}
	chKey := channels[0]
	r.msgChs[chKey] = make(chan *redis.Message, 20)
	pubsub.channelKey = chKey
	pubsub.msgChs = r.msgChs
	return &pubsub, nil
}

func (r *DummyRedis) Publish(ctx context.Context, channel string, message interface{}) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	msgCh, found := r.msgChs[channel]
	if !found {
		// no subscriber, hence no one to publish for
		return nil
	}
	// message can be any object that implement `encoding.BinaryMarshaler`,
	// but for now keep string as the only supported type for testing
	msgStr, ok := message.(string)
	if !ok {
		return fmt.Errorf("invalid message, must be string")
	}
	rMsg := redis.Message{
		Channel: channel,
		Payload: msgStr,
	}
	msgCh <- &rMsg
	return nil
}

type dummyRedisPubSub struct {
	channelKey string
	msgChs     map[string]chan *redis.Message
}

func (r *dummyRedisPubSub) Channel() <-chan *redis.Message {
	msgCh, found := r.msgChs[r.channelKey]
	if !found {
		return nil
	}
	return msgCh
}

func (r *dummyRedisPubSub) Close() error {
	if r.msgChs == nil {
		// nothing to cleanup
		return nil
	}
	if msgCh, found := r.msgChs[r.channelKey]; found {
		close(msgCh)
		delete(r.msgChs, r.channelKey)
	}
	return nil
}
