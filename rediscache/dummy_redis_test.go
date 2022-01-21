package rediscache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
)

func TestDummyRedisServer(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelInfo | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	redisServer, err := NewMockRedisServer()
	require.Nil(t, err)
	defer redisServer.Close()

	// Test redis standalone server
	client, err := NewClient(ctx, &RedisConfig{
		StandaloneAddr: redisServer.GetStandaloneAddr(),
	})
	require.Nil(t, err)
	testDummyRedisServer(t, client, redisServer)

	// Test redis server with sentinels (HA)
	client, err = NewClient(ctx, &RedisConfig{
		SentinelAddrs: redisServer.GetSentinelAddr(),
	})
	require.Nil(t, err)
	testDummyRedisServer(t, client, redisServer)
}

func testDummyRedisServer(t *testing.T, client *redis.Client, server *DummyRedis) {
	err := IsServerReady(client, 5*time.Second)
	require.Nil(t, err)

	resStr, err := client.Set("k1", "v1", 0).Result()
	require.Nil(t, err)
	require.Equal(t, "OK", resStr, "set key without any expiry time")

	resStr, err = client.Set("k1", "v1", 100*time.Millisecond).Result()
	require.Nil(t, err)
	require.Equal(t, "OK", resStr, "update key with expiry time")

	resStr, err = client.Get("k1").Result()
	require.Nil(t, err)
	require.Equal(t, "v1", resStr)

	// Since miniredis doesn't support timer, manually reduce TTL value
	server.FastForward(200 * time.Millisecond)

	resStr, err = client.Get("k1").Result()
	require.Equal(t, redis.Nil, err, "key will not longer exist as it has expired")

	resBool, err := client.SetNX("k2", "v2", 0).Result()
	require.Nil(t, err)
	require.True(t, resBool, "set key which will never expire")

	resBool, err = client.SetNX("k2", "v2", 100*time.Millisecond).Result()
	require.Nil(t, err)
	require.False(t, resBool, "key not set as it already exists")

	// Test redis pubsub
	// =================

	pubsub1 := client.Subscribe("ch1")
	require.Nil(t, err, "initialize channel to recv messages")
	require.NotNil(t, pubsub1)

	pubsub2 := client.Subscribe("ch2")
	require.Nil(t, err, "initialize another channel to recv messages")
	require.NotNil(t, pubsub2)

	msgCh1 := pubsub1.Channel()
	msgCh2 := pubsub2.Channel()

	time.Sleep(100 * time.Millisecond)

	msgs1 := []string{"msg1", "msg2", "msg3"}
	for _, msg := range msgs1 {
		_, err = client.Publish("ch1", msg).Result()
		require.Nil(t, err)
	}
	msgs2 := []string{"msg4", "msg5", "msg6"}
	for _, msg := range msgs2 {
		_, err = client.Publish("ch2", msg).Result()
		require.Nil(t, err)
	}

	// Wait for redis to pick up published messages
	time.Sleep(100 * time.Millisecond)

	pubsub1.Close()
	pubsub2.Close()

	recvdMsgs := []string{}
	for msg := range msgCh1 {
		require.Equal(t, "ch1", msg.Channel)
		recvdMsgs = append(recvdMsgs, msg.Payload)
	}
	require.Equal(t, msgs1, recvdMsgs, "recvd all the published messages on ch1")

	recvdMsgs = []string{}
	for msg := range msgCh2 {
		require.Equal(t, "ch2", msg.Channel)
		recvdMsgs = append(recvdMsgs, msg.Payload)
	}
	require.Equal(t, msgs2, recvdMsgs, "recvd all the published messages on ch2")

	keysRem, err := client.Del("k1", "k2").Result()
	require.Nil(t, err)
	require.Equal(t, int64(1), keysRem)

	resStr, err = client.Get("k1").Result()
	require.Equal(t, redis.Nil, err, "key should not exist")

	resStr, err = client.Get("k2").Result()
	require.Equal(t, redis.Nil, err, "key should not exist")

	// Test redis streams
	// ==================

	nonstreamKey := "nonstreamkey"
	streamKey1 := "teststream1"
	streamKey2 := "teststream2"

	streamMsg1 := map[string]interface{}{"message": "m1"}
	streamMsg2 := map[string]interface{}{"message": "m2"}
	streamMsg3 := map[string]interface{}{"message": "m3"}

	resStr, err = client.Set(nonstreamKey, "v1", 0).Result()
	require.Nil(t, err)
	require.Equal(t, "OK", resStr, "set key without any expiry time")

	out, err := client.XRange(streamKey1, RedisSmallestId, RedisGreatestId).Result()
	require.Nil(t, err)
	require.Equal(t, 0, len(out))
	out, err = client.XRange(streamKey2, RedisSmallestId, RedisGreatestId).Result()
	require.Nil(t, err)
	require.Equal(t, 0, len(out))
	out, err = client.XRange(nonstreamKey, RedisSmallestId, RedisGreatestId).Result()
	require.NotNil(t, err)
	require.Equal(t, "WRONGTYPE Operation against a key holding the wrong kind of value", err.Error())

	xaddArgs := redis.XAddArgs{
		Stream: streamKey1,
		Values: streamMsg1,
	}
	_, err = client.XAdd(&xaddArgs).Result()
	require.Nil(t, err)
	xaddArgs.Stream = streamKey2
	_, err = client.XAdd(&xaddArgs).Result()
	require.Nil(t, err)
	xaddArgs.Stream = nonstreamKey
	_, err = client.XAdd(&xaddArgs).Result()
	require.NotNil(t, err)
	require.Equal(t, "WRONGTYPE Operation against a key holding the wrong kind of value", err.Error())

	out, err = client.XRange(streamKey1, "invalidID", RedisGreatestId).Result()
	require.NotNil(t, err)
	require.Equal(t, "ERR Invalid stream ID specified as stream command argument", err.Error())

	out, err = client.XRange(streamKey1, RedisSmallestId, "invalidID").Result()
	require.NotNil(t, err)
	require.Equal(t, "ERR Invalid stream ID specified as stream command argument", err.Error())

	out, err = client.XRange(streamKey1, RedisSmallestId, RedisGreatestId).Result()
	require.Nil(t, err)
	require.Equal(t, len(out), 1)
	require.NotEmpty(t, out[0].ID, "")
	require.Equal(t, out[0].Values, streamMsg1)
	lastStream1ID := out[0].ID
	out, err = client.XRange(streamKey2, RedisSmallestId, RedisGreatestId).Result()
	require.Nil(t, err)
	require.Equal(t, len(out), 1)
	require.NotEmpty(t, out[0].ID, "")
	require.Equal(t, out[0].Values, streamMsg1)
	lastStream2ID := out[0].ID

	xaddArgs = redis.XAddArgs{
		Stream: streamKey1,
		Values: streamMsg2,
	}
	_, err = client.XAdd(&xaddArgs).Result()
	require.Nil(t, err)
	xaddArgs = redis.XAddArgs{
		Stream: streamKey2,
		Values: streamMsg2,
	}
	_, err = client.XAdd(&xaddArgs).Result()
	require.Nil(t, err)

	out, err = client.XRange(streamKey1, lastStream1ID, RedisGreatestId).Result()
	require.Nil(t, err)
	require.Equal(t, 2, len(out))
	require.NotEmpty(t, out[1].ID, "")
	require.Equal(t, out[1].Values, streamMsg2)
	newStream1ID := out[1].ID
	out, err = client.XRange(streamKey2, lastStream2ID, RedisGreatestId).Result()
	require.Nil(t, err)
	require.Equal(t, len(out), 2)
	require.NotEmpty(t, out[1].ID, "")
	require.Equal(t, out[1].Values, streamMsg2)
	newStream2ID := out[1].ID

	xreadArgs := redis.XReadArgs{
		Streams: []string{streamKey1, "invalidID"},
		Count:   1,
		Block:   1 * time.Millisecond,
	}
	xreadOut, err := client.XRead(&xreadArgs).Result()
	require.NotNil(t, err)
	require.Equal(t, "ERR Invalid stream ID specified as stream command argument", err.Error())
	xreadArgs.Streams = []string{nonstreamKey, lastStream1ID}
	xreadOut, err = client.XRead(&xreadArgs).Result()
	require.NotNil(t, err)

	xreadArgs.Streams = []string{streamKey1, lastStream1ID}
	xreadOut, err = client.XRead(&xreadArgs).Result()
	require.Nil(t, err)
	require.Equal(t, len(xreadOut), 1)
	require.Equal(t, xreadOut[0].Stream, streamKey1)
	require.Equal(t, len(xreadOut[0].Messages), 1)
	require.Equal(t, xreadOut[0].Messages[0].ID, newStream1ID)
	require.Equal(t, xreadOut[0].Messages[0].Values, streamMsg2)
	lastStream1ID = newStream1ID

	xreadArgs.Streams = []string{streamKey2, lastStream2ID}
	xreadOut, err = client.XRead(&xreadArgs).Result()
	require.Nil(t, err)
	require.Equal(t, len(xreadOut), 1)
	require.Equal(t, xreadOut[0].Stream, streamKey2)
	require.Equal(t, len(xreadOut[0].Messages), 1)
	require.Equal(t, xreadOut[0].Messages[0].ID, newStream2ID)
	require.Equal(t, xreadOut[0].Messages[0].Values, streamMsg2)
	lastStream2ID = newStream2ID

	testRes := make(chan string)
	go func() {
		xreadArgs.Streams = []string{streamKey1, lastStream1ID}
		xreadArgs.Block = 1 * time.Second
		xreadOut, err = client.XRead(&xreadArgs).Result()
		if err != nil {
			testRes <- fmt.Sprintf("failed to block on read: %v", err)
			return
		}
		if len(xreadOut) != 1 {
			testRes <- "xread should return only one stream message"
			return
		}
		if xreadOut[0].Stream != streamKey1 {
			testRes <- fmt.Sprintf("invalid stream key returned by xread %s", xreadOut[0].Stream)
			return
		}
		if len(xreadOut[0].Messages) != 1 {
			testRes <- "xread should return only one stream message"
			return
		}
		for sKey, sVal := range xreadOut[0].Messages[0].Values {
			if sKey != "message" || sVal != "m3" {
				testRes <- fmt.Sprintf("invalid stream message received %v", xreadOut[0].Messages[0].Values)
				return
			}
		}
		testRes <- "done"
	}()

	// xread should block until a new message is found
	time.Sleep(100 * time.Millisecond)

	xaddArgs = redis.XAddArgs{
		Stream: streamKey1,
		Values: streamMsg3,
	}
	_, err = client.XAdd(&xaddArgs).Result()
	require.Nil(t, err)

	require.Equal(t, "done", <-testRes)

	// cleanup
	keysRem, err = client.Del(nonstreamKey, streamKey1, streamKey2).Result()
	require.Nil(t, err)
	require.Equal(t, int64(3), keysRem)
}
