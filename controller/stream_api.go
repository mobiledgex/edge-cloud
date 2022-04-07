package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/rediscache"
	grpc "google.golang.org/grpc"
)

var (
	StreamMsgTypeMessage = "message"
	StreamMsgTypeError   = "error"
	StreamMsgTypeSOM     = "start-of-stream-message"
	StreamMsgTypeEOM     = "end-of-stream-message"

	StreamMsgReadTimeout = 30 * time.Minute
)

type CleanupStreamAction bool

var (
	CleanupStream   CleanupStreamAction = true
	NoCleanupStream CleanupStreamAction = false
)

type streamSend struct {
	cb        GenericCb
	mux       sync.Mutex
	crmPubSub *redis.PubSub
	crmMsgCh  <-chan *redis.Message
}

type StreamObjApi struct {
	all *AllApis
}

type GenericCb interface {
	Send(*edgeproto.Result) error
	grpc.ServerStream
}

type CbWrapper struct {
	GenericCb
	ctx       context.Context
	streamKey string
}

func NewStreamObjApi(sync *Sync, all *AllApis) *StreamObjApi {
	streamObjApi := StreamObjApi{}
	streamObjApi.all = all
	return &streamObjApi
}

func addMsgToRedisStream(ctx context.Context, streamKey string, streamMsg map[string]interface{}) error {
	xaddArgs := redis.XAddArgs{
		Stream: streamKey,
		Values: streamMsg,
	}
	_, err := redisClient.XAdd(&xaddArgs).Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "Failed to add message to stream", "key", streamKey, "err", err)
		return err
	}
	return nil
}

func (s *CbWrapper) Send(res *edgeproto.Result) error {
	if res != nil {
		var streamMsg map[string]interface{}
		inMsg, err := json.Marshal(res)
		if err != nil {
			return err
		}
		err = json.Unmarshal(inMsg, &streamMsg)
		if err != nil {
			return err
		}
		err = addMsgToRedisStream(s.ctx, s.streamKey, streamMsg)
		if err != nil {
			return err
		}
	}
	s.GenericCb.Send(res)
	return nil
}

func (s *StreamObjApi) StreamMsgs(streamKey string, cb edgeproto.StreamObjApi_StreamAppInstServer) error {
	out, err := redisClient.Exists(streamKey).Result()
	if err != nil {
		return err
	}
	if out == 0 {
		// stream key does not exist
		return fmt.Errorf("Stream %s does not exist", streamKey)
	}

	streamMsgs, err := redisClient.XRange(streamKey, rediscache.RedisSmallestId, rediscache.RedisGreatestId).Result()
	if err != nil {
		return err
	}

	decodeStreamMsg := func(sMsg map[string]interface{}) (bool, error) {
		done := false
		for k, v := range sMsg {
			switch k {
			case StreamMsgTypeMessage:
				val, ok := v.(string)
				if !ok {
					return done, fmt.Errorf("Invalid stream message %v, must be of type string", v)
				}
				cb.Send(&edgeproto.Result{Message: val})
			case StreamMsgTypeError:
				val, ok := v.(string)
				if !ok {
					return done, fmt.Errorf("Invalid stream error %v, must be of type string", v)
				}
				return done, fmt.Errorf(val)
			case StreamMsgTypeEOM:
				done = true
				break
			case StreamMsgTypeSOM:
				// ignore
			default:
				return done, fmt.Errorf("Unsupported message type received: %v", k)
			}

		}
		return done, nil
	}

	lastStreamMsgId := ""
	for _, sMsg := range streamMsgs {
		lastStreamMsgId = sMsg.ID
		done, err := decodeStreamMsg(sMsg.Values)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
	if lastStreamMsgId == "" {
		lastStreamMsgId = rediscache.RedisSmallestId
	}

	for {
		// Blocking read for new stream messages until EOM is found
		xreadArgs := redis.XReadArgs{
			Streams: []string{streamKey, lastStreamMsgId},
			Count:   1,
			Block:   StreamMsgReadTimeout,
		}
		sMsg, err := redisClient.XRead(&xreadArgs).Result()
		if err != nil {
			return fmt.Errorf("Error reading from stream %s, %v", streamKey, err)
		}
		if len(sMsg) != 1 {
			return fmt.Errorf("Output should only be for a single stream %s, but multiple found %v", streamKey, sMsg)
		}
		sMsgs := sMsg[0].Messages
		if len(sMsgs) != 1 {
			return fmt.Errorf("Output should only be for a single message, but multiple found %s, %v", streamKey, sMsgs)
		}
		lastStreamMsgId = sMsgs[0].ID
		done, err := decodeStreamMsg(sMsgs[0].Values)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
}

type StreamOptions struct {
	NoResetStream bool
}

type StreamOp func(op *StreamOptions)

func WithNoResetStream() StreamOp {
	return func(op *StreamOptions) { op.NoResetStream = true }
}

func (s *StreamObjApi) startStream(ctx context.Context, cctx *CallContext, streamKey string, inCb GenericCb, opts ...StreamOp) (*streamSend, GenericCb, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "Start new stream", "key", streamKey)

	streamOps := StreamOptions{}
	for _, fn := range opts {
		fn(&streamOps)
	}

	// If this is an undo, then caller has already performed
	// the same operation, so reuse the existing callback
	if cctx.Undo {
		streamSendObj := streamSend{cb: inCb}
		outCb := &CbWrapper{
			GenericCb: inCb,
			ctx:       ctx,
			streamKey: streamKey,
		}
		return &streamSendObj, outCb, nil
	}

	// * Redis `pubsub.Close()` is not synchronous and hence we can't rely on `NumSub()`
	//   call to figure out if the stream already exists for the streamKey
	// * Hence, we end up using redis stream to figure out if stream already exists or not
	// * Since this can be accessed by multiple threads, we use redis transactions for atomic
	//   operation.
	// * To mark the start of stream, we create redis stream as part of this function and add
	//   initial SOM message to it
	// * If redis stream already exists, then we check for EOM or Error to figure out if it is
	//   a new stream re-using the old stream and cleanup the old stream
	// * If CRM override (IgnoreTransient) is specified, then we reset the stream if it already
	//   is in progress
	txf := func(tx *redis.Tx) error {
		// Get the current value or zero.
		out, err := tx.Exists(streamKey).Result()
		if err != nil && err != redis.Nil {
			return err
		}

		newStream := false
		cleanupOldStream := false
		if ignoreCRMTransient(cctx) {
			newStream = true
			if out == 1 {
				cleanupOldStream = true
			}
		} else if out == 1 {
			// check last message on the existing stream to
			// figure out if stream should be cleared or not
			streamMsgs, err := redisClient.XRange(streamKey,
				rediscache.RedisSmallestId, rediscache.RedisGreatestId).Result()
			if err != nil {
				return err
			}
			if len(streamMsgs) > 0 {
				for k, _ := range streamMsgs[len(streamMsgs)-1].Values {
					if k == StreamMsgTypeEOM || k == StreamMsgTypeError {
						// Since last msg was EOM/Error, reset this stream
						// as it is for a new API call
						cleanupOldStream = true
						newStream = true
						break
					}
				}
			} else {
				newStream = true
			}
		} else {
			newStream = true
		}
		if streamOps.NoResetStream {
			// for unit-testing
			cleanupOldStream = false
		}
		// Operation is commited only if the watched keys remain unchanged.
		_, err = tx.TxPipelined(func(pipe redis.Pipeliner) error {
			if newStream {
				if cleanupOldStream {
					_, err := pipe.Del(streamKey).Result()
					if err != nil {
						return err
					}
				}
				xaddArgs := redis.XAddArgs{
					Stream: streamKey,
					Values: map[string]interface{}{
						StreamMsgTypeSOM: "",
					},
				}
				_, err := pipe.XAdd(&xaddArgs).Result()
				if err != nil {
					return err
				}
				return nil
			}
			return fmt.Errorf("%s %s", cloudcommon.StreamActionAlreadyInProgressMsg, streamKey)
		})
		return err
	}

	// Retry if the key has been changed.
	for i := 0; i < rediscache.RedisTxMaxRetries; i++ {
		err := redisClient.Watch(txf, streamKey)
		if err == nil {
			// Success.
			break
		}
		if err == redis.TxFailedErr {
			// Optimistic lock lost. Retry.
			continue
		}
		// Return any other error.
		return nil, nil, err
	}

	// Start subscription to redis channel identified by stream key.
	// Objects from CRM will be published to this channel and hence,
	// will be received by intended receiver
	// Note that this method does not wait on a response from redis, so the
	// subscription may not be active immediately. To force the connection to wait,
	// we call the Receive() method on the returned *PubSub
	pubsub := redisClient.Subscribe(streamKey)

	// Wait for confirmation that subscription is created before publishing anything.
	_, err := pubsub.Receive()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to subscribe to stream %s, %v", streamKey, err)
	}

	// Go channel to receives messages.
	ch := pubsub.Channel()

	streamSendObj := streamSend{}
	streamSendObj.crmPubSub = pubsub
	streamSendObj.crmMsgCh = ch

	if inCb != nil {
		streamSendObj.cb = inCb
	}

	outCb := &CbWrapper{
		GenericCb: inCb,
		ctx:       ctx,
		streamKey: streamKey,
	}

	log.SpanLog(ctx, log.DebugLevelApi, "Started new stream", "key", streamKey)
	return &streamSendObj, outCb, nil
}

func (s *StreamObjApi) stopStream(ctx context.Context, cctx *CallContext, streamKey string, streamSendObj *streamSend, objErr error, cleanupStream CleanupStreamAction) error {
	log.SpanLog(ctx, log.DebugLevelApi, "Stop stream", "key", streamKey, "cctx", cctx, "err", objErr)
	if streamSendObj == nil {
		return nil
	}

	// If this is an undo, then caller has already performed the same operation,
	// so skip performing any cleanup
	if cctx.Undo {
		return nil
	}

	streamSendObj.mux.Lock()
	defer streamSendObj.mux.Unlock()
	if objErr != nil {
		streamMsg := map[string]interface{}{
			StreamMsgTypeError: objErr.Error(),
		}
		err := addMsgToRedisStream(ctx, streamKey, streamMsg)
		if err != nil {
			return err
		}
	} else {
		streamMsg := map[string]interface{}{
			StreamMsgTypeEOM: "",
		}
		err := addMsgToRedisStream(ctx, streamKey, streamMsg)
		if err != nil {
			return err
		}
	}
	if streamSendObj.crmPubSub != nil {
		// Close() also closes channels
		streamSendObj.crmPubSub.Close()
	}
	if cleanupStream {
		_, err := redisClient.Del(streamKey).Result()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to cleanup redis stream", "key", streamKey, "err", err)
		}
	}
	return nil
}

// Publish info object received from CRM to redis so that controller
// can act on status messages & info state accordingly
func (s *StreamObjApi) UpdateStatus(ctx context.Context, obj interface{}, streamKey string) {
	inObj, err := json.Marshal(obj)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "Failed to marshal json object", "obj", obj, "err", err)
		return
	}
	_, err = redisClient.Publish(streamKey, string(inObj)).Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "Failed to publish message on redis channel", "key", streamKey, "err", err)
	}
	// If the pubsub subscriber (i.e. controller) restarts in middle of message streaming, then
	// stream will be left incomplete as it will not be stopped by the caller controller.
	// Hence, check here if there are no subscribers to the stream, then mark the end of stream,
	// if not already marked
	// Note: `PubSubNumSub` function call will return right value as the connection between
	//       redis client & redis server is closed (due to controller restart)
	chMap, err := redisClient.PubSubNumSub(streamKey).Result()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "Failed to get subscribe count on redis pubsub", "key", streamKey, "err", err)
		return
	}
	if cnt, ok := chMap[streamKey]; ok {
		if cnt == 0 {
			streamClosed := false
			streamMsgs, err := redisClient.XRange(streamKey, rediscache.RedisSmallestId, rediscache.RedisGreatestId).Result()
			if err == nil && len(streamMsgs) > 0 {
				for msgType, _ := range streamMsgs[len(streamMsgs)-1].Values {
					switch msgType {
					case StreamMsgTypeEOM:
						fallthrough
					case StreamMsgTypeError:
						streamClosed = true
					}
				}
			}
			if !streamClosed {
				streamMsg := map[string]interface{}{
					StreamMsgTypeEOM: "",
				}
				err := addMsgToRedisStream(ctx, streamKey, streamMsg)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelApi, "Failed to add EOM message to redis stream", "key", streamKey, "err", err)
				}
			}
		}
	}
}
