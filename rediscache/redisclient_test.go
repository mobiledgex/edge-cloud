package rediscache

import (
	"context"
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
	time.Sleep(100 * time.Millisecond)

	resStr, err = client.Get(ctx, "k1")
	require.Equal(t, redis.Nil, err, "key will not longer exist as it is expired")

	resBool, err := client.SetNX(ctx, "k2", "v2", 0)
	require.Nil(t, err)
	require.True(t, resBool, "set key which will never expire")

	resBool, err = client.SetNX(ctx, "k2", "v2", 100*time.Millisecond)
	require.Nil(t, err)
	require.False(t, resBool, "key not set as it already exists")

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
}
