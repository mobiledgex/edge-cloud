package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/rediscache"
	grpc "google.golang.org/grpc"
)

var (
	StreamMsgTypeMessage = "message"
	StreamMsgTypeError   = "error"
	StreamMsgTypeEOM     = "end-of-stream-message"

	StreamMsgReadTimeout = 30 * time.Minute
	StreamExpiration     = 10 * time.Minute
)

type streamSend struct {
	cb        GenericCb
	mux       sync.Mutex
	crmPubSub rediscache.RedisPubSub
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

		// TODO: Make them atomic
		redisClient.XAdd(s.ctx, s.streamKey, streamMsg)
		redisClient.Expire(s.ctx, s.streamKey, StreamExpiration)
	}
	s.GenericCb.Send(res)
	return nil
}

func (s *StreamObjApi) StreamMsgs(streamKey string, cb edgeproto.StreamObjApi_StreamAppInstServer) error {
	ctx := cb.Context()

	out, err := redisClient.Exists(ctx, streamKey)
	if err != nil {
		return err
	}
	if out == 0 {
		// stream key does not exist
		return nil
	}

	streamMsgs, err := redisClient.XRange(ctx, streamKey, rediscache.RedisSmallestId, rediscache.RedisGreatestId)
	if err != nil {
		panic(err)
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
		sMsg, err := redisClient.XRead(ctx, []string{streamKey, lastStreamMsgId}, 1, StreamMsgReadTimeout)
		if err != nil {
			return fmt.Errorf("Error reading from stream %s, %v", streamKey, err)
		}
		if len(sMsg) != 1 {
			return fmt.Errorf("Output should only be for a single stream %s, but multiple found %v", streamKey, sMsg)
		}
		sMsgs := sMsg[0].Msgs
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

func (s *StreamObjApi) startStream(ctx context.Context, streamKey string, inCb GenericCb) (*streamSend, GenericCb, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "Start new stream", "key", streamKey)

	// Start subscription to redis channel identified by stream key.
	// Objects from CRM will be published to this channel and hence,
	// will be received by intended receiver
	pubsub, err := redisClient.Subscribe(ctx, streamKey)
	if err != nil {
		return nil, nil, err
	}

	// Go channel to receives messages.
	ch := pubsub.Channel()

	streamSendObj := streamSend{}
	streamSendObj.crmPubSub = pubsub
	streamSendObj.crmMsgCh = ch

	if inCb != nil {
		streamSendObj.cb = inCb
	}

	// Delete any existing streams
	_, err = redisClient.Del(ctx, streamKey)
	if err != nil {
		return nil, nil, err
	}

	outCb := &CbWrapper{
		GenericCb: inCb,
		ctx:       ctx,
		streamKey: streamKey,
	}

	return &streamSendObj, outCb, nil
}

func (s *StreamObjApi) stopStream(ctx context.Context, streamKey string, streamSendObj *streamSend, objErr error) error {
	log.SpanLog(ctx, log.DebugLevelApi, "Stop stream", "key", streamKey, "err", objErr)
	if streamSendObj != nil {
		streamSendObj.mux.Lock()
		defer streamSendObj.mux.Unlock()
		if objErr != nil {
			streamMsg := map[string]interface{}{
				StreamMsgTypeError: objErr.Error(),
			}
			// TODO: Make them atomic
			redisClient.XAdd(ctx, streamKey, streamMsg)
			redisClient.Expire(ctx, streamKey, StreamExpiration)
		} else {
			// TODO: Make them atomic
			streamMsg := map[string]interface{}{
				StreamMsgTypeEOM: "",
			}
			redisClient.XAdd(ctx, streamKey, streamMsg)
			redisClient.Expire(ctx, streamKey, StreamExpiration)
		}
		if streamSendObj.crmPubSub != nil {
			// Close() also closes channels
			streamSendObj.crmPubSub.Close()
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
	err = redisClient.Publish(ctx, streamKey, string(inObj))
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "Failed to publish message on redis channel", "key", streamKey, "err", err)
	}
}
