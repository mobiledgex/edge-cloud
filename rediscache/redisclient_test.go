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

func TestDummyRedisClient(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelRedis | log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	client := NewDummyRedis()

	resStr, err := client.Set(ctx, "k1", "v1", 0)
	require.Nil(t, err)
	require.Equal(t, "OK", resStr, "set key without any expiry time")

	resStr, err = client.Set(ctx, "k1", "v1", 100*time.Millisecond)
	require.Nil(t, err)
	require.Equal(t, "OK", resStr, "update key with expiry time")

	resStr, err = client.Get(ctx, "k1")
	require.Nil(t, err)
	require.Equal(t, "v1", resStr)

	// sleep for k1 to expire
	time.Sleep(200 * time.Millisecond)

	resStr, err = client.Get(ctx, "k1")
	require.Equal(t, redis.Nil, err, "key will not longer exist as it has expired")

	resBool, err := client.SetNX(ctx, "k2", "v2", 0)
	require.Nil(t, err)
	require.True(t, resBool, "set key which will never expire")

	resBool, err = client.SetNX(ctx, "k2", "v2", 100*time.Millisecond)
	require.Nil(t, err)
	require.False(t, resBool, "key not set as it already exists")

	// Test redis pubsub
	// =================

	pubsub1, err := client.Subscribe(ctx, "ch1")
	require.Nil(t, err, "initialize channel to recv messages")
	require.NotNil(t, pubsub1)

	pubsub2, err := client.Subscribe(ctx, "ch2")
	require.Nil(t, err, "initialize another channel to recv messages")
	require.NotNil(t, pubsub2)

	msgCh1 := pubsub1.Channel()
	msgCh2 := pubsub2.Channel()
	msgs1 := []string{"msg1", "msg2", "msg3"}
	for _, msg := range msgs1 {
		err = client.Publish(ctx, "ch1", msg)
		require.Nil(t, err)
	}
	msgs2 := []string{"msg4", "msg5", "msg6"}
	for _, msg := range msgs2 {
		err = client.Publish(ctx, "ch2", msg)
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

	keysRem, err := client.Del(ctx, "k1", "k2")
	require.Nil(t, err)
	require.Equal(t, int64(1), keysRem, "only one key should be deleted as k1 doesn't exist")

	resStr, err = client.Get(ctx, "k1")
	require.Equal(t, redis.Nil, err, "key should not exist")

	resStr, err = client.Get(ctx, "k2")
	require.Equal(t, redis.Nil, err, "key should not exist")

	// Test redis streams
	// ==================

	nonstreamKey := "nonstreamkey"
	streamKey1 := "teststream1"
	streamKey2 := "teststream2"

	streamMsg1 := map[string]interface{}{"message": "m1"}
	streamMsg2 := map[string]interface{}{"message": "m2"}
	streamMsg3 := map[string]interface{}{"message": "m3"}

	resStr, err = client.Set(ctx, nonstreamKey, "v1", 0)
	require.Nil(t, err)
	require.Equal(t, "OK", resStr, "set key without any expiry time")

	out, err := client.XRange(ctx, streamKey1, RedisSmallestId, RedisGreatestId)
	require.Nil(t, err)
	require.Equal(t, 0, len(out))
	out, err = client.XRange(ctx, streamKey2, RedisSmallestId, RedisGreatestId)
	require.Nil(t, err)
	require.Equal(t, 0, len(out))
	out, err = client.XRange(ctx, nonstreamKey, RedisSmallestId, RedisGreatestId)
	require.NotNil(t, err)
	require.Equal(t, "WRONGTYPE Operation against a key holding the wrong kind of value", err.Error())

	err = client.XAdd(ctx, streamKey1, streamMsg1)
	require.Nil(t, err)
	err = client.XAdd(ctx, streamKey2, streamMsg1)
	require.Nil(t, err)
	err = client.XAdd(ctx, nonstreamKey, streamMsg1)
	require.NotNil(t, err)
	require.Equal(t, "WRONGTYPE Operation against a key holding the wrong kind of value", err.Error())

	out, err = client.XRange(ctx, streamKey1, "invalidID", RedisGreatestId)
	require.NotNil(t, err)
	require.Equal(t, "ERR Invalid stream ID specified as stream command argument", err.Error())

	out, err = client.XRange(ctx, streamKey1, RedisSmallestId, "invalidID")
	require.NotNil(t, err)
	require.Equal(t, "ERR Invalid stream ID specified as stream command argument", err.Error())

	out, err = client.XRange(ctx, streamKey1, RedisSmallestId, RedisGreatestId)
	require.Nil(t, err)
	require.Equal(t, len(out), 1)
	require.NotEmpty(t, out[0].ID, "")
	require.Equal(t, out[0].Values, streamMsg1)
	lastStream1ID := out[0].ID
	out, err = client.XRange(ctx, streamKey2, RedisSmallestId, RedisGreatestId)
	require.Nil(t, err)
	require.Equal(t, len(out), 1)
	require.NotEmpty(t, out[0].ID, "")
	require.Equal(t, out[0].Values, streamMsg1)
	lastStream2ID := out[0].ID

	err = client.XAdd(ctx, streamKey1, streamMsg2)
	require.Nil(t, err)
	err = client.XAdd(ctx, streamKey2, streamMsg2)
	require.Nil(t, err)

	out, err = client.XRange(ctx, streamKey1, lastStream1ID, RedisGreatestId)
	require.Nil(t, err)
	require.Equal(t, 2, len(out))
	require.NotEmpty(t, out[1].ID, "")
	require.Equal(t, out[1].Values, streamMsg2)
	newStream1ID := out[1].ID
	out, err = client.XRange(ctx, streamKey2, lastStream2ID, RedisGreatestId)
	require.Nil(t, err)
	require.Equal(t, len(out), 2)
	require.NotEmpty(t, out[1].ID, "")
	require.Equal(t, out[1].Values, streamMsg2)
	newStream2ID := out[1].ID

	xreadOut, err := client.XRead(ctx, []string{streamKey1, "invalidID"}, 1, 1*time.Millisecond)
	require.NotNil(t, err)
	require.Equal(t, "ERR Invalid stream ID specified as stream command argument", err.Error())
	xreadOut, err = client.XRead(ctx, []string{nonstreamKey, lastStream1ID}, 1, 1*time.Millisecond)
	require.NotNil(t, err)
	require.Equal(t, "WRONGTYPE Operation against a key holding the wrong kind of value", err.Error())

	xreadOut, err = client.XRead(ctx, []string{streamKey1, lastStream1ID}, 1, 1*time.Millisecond)
	require.Nil(t, err)
	require.Equal(t, len(xreadOut), 1)
	require.Equal(t, xreadOut[0].Stream, streamKey1)
	require.Equal(t, len(xreadOut[0].Msgs), 1)
	require.Equal(t, xreadOut[0].Msgs[0].ID, newStream1ID)
	require.Equal(t, xreadOut[0].Msgs[0].Values, streamMsg2)

	xreadOut, err = client.XRead(ctx, []string{streamKey2, lastStream2ID}, 1, 1*time.Millisecond)
	require.Nil(t, err)
	require.Equal(t, len(xreadOut), 1)
	require.Equal(t, xreadOut[0].Stream, streamKey2)
	require.Equal(t, len(xreadOut[0].Msgs), 1)
	require.Equal(t, xreadOut[0].Msgs[0].ID, newStream2ID)
	require.Equal(t, xreadOut[0].Msgs[0].Values, streamMsg2)

	testRes := make(chan string)
	go func() {
		xreadOut, err = client.XRead(ctx, []string{streamKey1, RedisLastId}, 1, 1*time.Second)
		if err != nil {
			testRes <- err.Error()
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
		if len(xreadOut[0].Msgs) != 1 {
			testRes <- "xread should return only one stream message"
			return
		}
		for sKey, sVal := range xreadOut[0].Msgs[0].Values {
			if sKey != "message" || sVal != "m3" {
				testRes <- fmt.Sprintf("invalid stream message received %v", xreadOut[0].Msgs[0].Values)
				return
			}
		}
		testRes <- "done"
	}()

	// xread should block until a new message is found
	time.Sleep(100 * time.Millisecond)

	err = client.XAdd(ctx, streamKey1, streamMsg3)
	require.Nil(t, err)

	require.Equal(t, "done", <-testRes)

	keysRem, err = client.Del(ctx, nonstreamKey, streamKey1, streamKey2)
	require.Nil(t, err)
	require.Equal(t, int64(3), keysRem)
}
