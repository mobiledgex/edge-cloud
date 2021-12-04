package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	streamObjs = cloudcommon.StreamObj{}

	StreamTimeout = 30 * time.Minute
)

type streamSend struct {
	cb        GenericCb
	mux       sync.Mutex
	streamer  *cloudcommon.Streamer
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
	streamSendObj *streamSend
}

func NewStreamObjApi(sync *Sync, all *AllApis) *StreamObjApi {
	streamObjApi := StreamObjApi{}
	streamObjApi.all = all
	return &streamObjApi
}

func (s *CbWrapper) Send(res *edgeproto.Result) error {
	if res != nil {
		if s.streamSendObj.streamer != nil {
			s.streamSendObj.streamer.Publish(res.Message)
		}
	}
	s.GenericCb.Send(res)
	return nil
}

func (s *StreamObjApi) StreamLocalMsgs(streamKeyObj *edgeproto.StreamKey, cb edgeproto.StreamObjApi_StreamLocalMsgsServer) error {
	ctx := cb.Context()
	streamKey := streamKeyObj.Name
	log.SpanLog(ctx, log.DebugLevelApi, "Stream obj messages", "key", streamKey)
	streamer := streamObjs.Get(streamKey)
	if streamer == nil {
		// stream not found, nothing to show
		log.SpanLog(ctx, log.DebugLevelApi, "Stream obj not found", "key", streamKey)
		return nil
	}
	streamCh := streamer.Subscribe()
	defer streamer.Unsubscribe(streamCh)

	for streamMsg := range streamCh {
		switch out := streamMsg.(type) {
		case string:
			cb.Send(&edgeproto.Result{Message: out})
		case error:
			return out
		default:
			return fmt.Errorf("Unsupported message type received: %v", streamMsg)
		}
	}

	return nil
}

func (s *StreamObjApi) StreamMsgs(streamKey string, cb edgeproto.StreamObjApi_StreamAppInstServer) error {
	ctx := cb.Context()

	streamKeyObj := edgeproto.StreamKey{Name: streamKey}
	if *externalApiAddr == "" {
		// unit test
		return s.StreamLocalMsgs(&streamKeyObj, cb)
	}

	// Currently we don't know which controller has the streamer Obj for this key
	// (or if it's even present), so just broadcast to all.

	// In case CRM loses connection to controller during an action in-progress
	// (for ex: during CreateClusterInst), and reconnects to a different controller
	// than it did earlier, then end-user might not see any status messages.
	// For now we won't handle this case, as it will just affect the status updates
	// which shouldn't be a big deal
	err := s.all.controllerApi.RunJobs(func(arg interface{}, addr string) error {
		if addr == *externalApiAddr {
			// local node
			return s.StreamLocalMsgs(&streamKeyObj, cb)
		}
		// connect to remote node
		conn, cErr := ControllerConnect(ctx, addr)
		if cErr != nil {
			return cErr
		}
		defer conn.Close()

		cmd := edgeproto.NewStreamObjApiClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), StreamTimeout)
		defer cancel()
		stream, sErr := cmd.StreamLocalMsgs(ctx, &streamKeyObj)
		if sErr != nil {
			return sErr
		}
		var sMsg *edgeproto.Result
		for {
			sMsg, sErr = stream.Recv()
			if sErr == io.EOF {
				sErr = nil
				break
			}
			if sErr != nil {
				break
			}
			cb.Send(sMsg)
		}
		if sErr != nil {
			return sErr
		}
		return nil
	}, nil)
	return err
}

func (s *StreamObjApi) startStream(ctx context.Context, streamKey string, inCb GenericCb) (*streamSend, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "Start new stream", "key", streamKey)
	streamer := streamObjs.Get(streamKey)
	if streamer != nil {
		// stream is already in progress for this key
		if streamer.State == edgeproto.StreamState_STREAM_START {
			return nil, fmt.Errorf("Stream already in progress for %s", streamKey)
		}
	}

	pubsub, err := redisClient.Subscribe(ctx, streamKey)
	if err != nil {
		return nil, err
	}

	// Go channel which receives messages.
	ch := pubsub.Channel()

	streamSendObj := streamSend{}
	streamSendObj.crmPubSub = pubsub
	streamSendObj.crmMsgCh = ch

	if inCb != nil {
		streamer = cloudcommon.NewStreamer()
		streamSendObj.streamer = streamer
		streamSendObj.cb = inCb
		streamObjs.Add(streamKey, streamer)
	}

	return &streamSendObj, nil
}

func (s *StreamObjApi) stopStream(ctx context.Context, streamKey string, streamSendObj *streamSend, objErr error) error {
	log.SpanLog(ctx, log.DebugLevelApi, "Stop stream", "key", streamKey, "err", objErr)
	if streamSendObj != nil {
		streamSendObj.mux.Lock()
		defer streamSendObj.mux.Unlock()
		if streamSendObj.streamer != nil {
			if objErr != nil {
				streamSendObj.streamer.Publish(objErr)
			}
			streamSendObj.streamer.Stop()
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
