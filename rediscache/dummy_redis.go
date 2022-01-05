package rediscache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/mobiledgex/edge-cloud/util"
)

const (
	DummyDataTypeString int = iota
	DummyDataTypeStream
)

type dummyData struct {
	datatype        int
	val             string
	streamVal       []XReadStreamMsg
	streamLis       map[string]chan XReadStreamMsg
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
	if data.datatype != DummyDataTypeString {
		return "", fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
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
		datatype:        DummyDataTypeString,
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
		datatype:        DummyDataTypeString,
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

func (r *DummyRedis) Exists(ctx context.Context, keys ...string) (int64, error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.db == nil {
		return 0, fmt.Errorf("redis not initialized")
	}
	keysExists := int64(0)
	for _, key := range keys {
		if _, ok := r.db[key]; ok {
			keysExists++
		}
	}
	return keysExists, nil
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

func (r *DummyRedis) XAdd(ctx context.Context, stream string, values map[string]interface{}) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.db == nil {
		return fmt.Errorf("redis not initialized")
	}
	streamVals := []XReadStreamMsg{}
	// if already present, just append
	streamData, ok := r.db[stream]
	if ok {
		if streamData.datatype != DummyDataTypeStream {
			return fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		if streamData.expirationTime > 0 &&
			(time.Since(streamData.lastUpdatedTime) > streamData.expirationTime) {
			// key has expired
			streamVals = []XReadStreamMsg{}
		} else {
			streamVals = streamData.streamVal
		}
	} else {
		streamData = &dummyData{
			datatype: DummyDataTypeStream,
		}
	}
	streamMsg := XReadStreamMsg{
		ID:     strconv.FormatInt(time.Now().UTC().UnixNano(), 10),
		Values: values,
	}
	streamVals = append(streamVals, streamMsg)
	streamData.streamVal = streamVals
	streamData.lastUpdatedTime = time.Now()
	r.db[stream] = streamData
	if len(streamData.streamLis) > 0 {
		for _, lis := range streamData.streamLis {
			lis <- streamMsg
		}
	}
	return nil
}

func (r *DummyRedis) XRange(ctx context.Context, stream, start, stop string) ([]XReadStreamMsg, error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.db == nil {
		return nil, fmt.Errorf("redis not initialized")
	}
	out := []XReadStreamMsg{}
	streamData, ok := r.db[stream]
	if !ok {
		return out, nil
	}
	if streamData.datatype != DummyDataTypeStream {
		return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	if len(streamData.streamVal) == 0 {
		return out, nil
	}

	if start == RedisSmallestId {
		start = streamData.streamVal[0].ID
	}
	if stop == RedisGreatestId {
		stop = streamData.streamVal[len(streamData.streamVal)-1].ID
	}

	if streamData.expirationTime > 0 &&
		(time.Since(streamData.lastUpdatedTime) > streamData.expirationTime) {
		// key has expired
		return out, nil
	}

	streamStartFound := false
	streamEndFound := false
	for _, sVal := range streamData.streamVal {
		if sVal.ID == start {
			streamStartFound = true
		}
		if sVal.ID == stop {
			streamEndFound = true
		}
		if streamStartFound {
			out = append(out, sVal)
		}
		if streamEndFound {
			break
		}
	}
	if !streamStartFound || !streamEndFound {
		return nil, fmt.Errorf("ERR Invalid stream ID specified as stream command argument")
	}
	return out, nil
}

func (r *DummyRedis) XRead(ctx context.Context, streams []string, count int64, block time.Duration) ([]XReadStream, error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.db == nil {
		return nil, fmt.Errorf("redis not initialized")
	}

	if len(streams) != 2 {
		return nil, fmt.Errorf("Invalid value specified for streams %v, must only have two values 'streamkey' and 'streamstartid'", streams)
	}
	stream := streams[0]
	streamStartId := streams[1]

	streamData, ok := r.db[stream]
	if !ok {
		return []XReadStream{}, nil
	}
	if streamData.datatype != DummyDataTypeStream {
		return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	if count < 0 {
		return nil, fmt.Errorf("Invalid value for count specified %d", count)
	}
	if streamStartId == RedisLastId {
		if len(streamData.streamVal) > 0 {
			streamStartId = streamData.streamVal[len(streamData.streamVal)-1].ID
		} else {
			streamStartId = ""
		}
	}

	if streamData.expirationTime > 0 &&
		(time.Since(streamData.lastUpdatedTime) > streamData.expirationTime) {
		// key has expired
		return []XReadStream{}, nil
	}

	xreadStreamOut := []XReadStream{}
	xreadOut := XReadStream{}
	xreadOut.Stream = stream
	msgs := []XReadStreamMsg{}
	streamStartFound := false
	streamEndFound := false
	for _, val := range streamData.streamVal {
		if streamStartId == "" {
			streamStartFound = true
		}
		if streamStartId == val.ID {
			streamStartFound = true
			continue
		}
		if streamStartFound {
			msgs = append(msgs, val)
		}
		if int64(len(msgs)) == count {
			streamEndFound = true
			break
		}
	}

	if !streamStartFound {
		return nil, fmt.Errorf("ERR Invalid stream ID specified as stream command argument")
	}

	if !streamEndFound {
		// Wait until target number of messages is received
		sMsgChan := make(chan XReadStreamMsg, 100)
		if len(streamData.streamLis) == 0 {
			streamData.streamLis = make(map[string]chan XReadStreamMsg)
		}
		streamData.streamLis[stream] = sMsgChan
		r.mux.Unlock()
		for {
			select {
			case newSMsg := <-sMsgChan:
				msgs = append(msgs, newSMsg)
				if int64(len(msgs)) == count {
					streamEndFound = true
				} else {
					continue
				}
			case <-time.After(block):
				streamEndFound = false
			}
			break
		}
		r.mux.Lock()
		close(sMsgChan)
		delete(streamData.streamLis, stream)
		if !streamEndFound {
			return nil, redis.Nil
		}
	}

	xreadOut.Msgs = msgs
	xreadStreamOut = append(xreadStreamOut, xreadOut)

	return xreadStreamOut, nil
}

func (r *DummyRedis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.db == nil {
		return fmt.Errorf("redis not initialized")
	}
	data, ok := r.db[key]
	if !ok {
		return redis.Nil
	}
	data.expirationTime = expiration
	r.db[key] = data
	return nil
}
