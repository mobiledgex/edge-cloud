package main

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	grpc "google.golang.org/grpc"
)

var streamObjApi = StreamObjApi{}

var StreamTimeout = 30 * time.Minute
var StreamLeaseTimeoutSec = int64(60 * 60) // 1 hour

var streamObjs = cloudcommon.StreamObj{}

type streamSend struct {
	cb           GenericCb
	lastRemoteId uint32
	doneCh       chan bool
	mux          sync.Mutex
	closeCh      bool
	streamer     *cloudcommon.Streamer
}

type StreamObjApi struct {
	sync  *Sync
	store edgeproto.StreamObjStore
	cache edgeproto.StreamObjCache
}

type GenericCb interface {
	Send(*edgeproto.Result) error
	grpc.ServerStream
}

type CbWrapper struct {
	GenericCb
	streamSendObj *streamSend
}

func (s *CbWrapper) Send(res *edgeproto.Result) error {
	if res != nil {
		if s.streamSendObj.closeCh {
			return nil
		}
		if s.streamSendObj.streamer != nil {
			s.streamSendObj.streamer.Publish(res.Message)
		}
	}
	s.GenericCb.Send(res)
	return nil
}

func (s *StreamObjApi) StreamLocalMsgs(key *edgeproto.AppInstKey, cb edgeproto.StreamObjApi_StreamLocalMsgsServer) error {
	streamer := streamObjs.Get(*key)
	if streamer == nil {
		// stream not found, nothing to show
		return nil
	}
	streamCh := streamer.Subscribe()
	defer streamer.Unsubscribe(streamCh)
	for streamMsg := range streamCh {
		switch out := streamMsg.(type) {
		case *edgeproto.StreamMsg:
			cb.Send(out)
		case string:
			cb.Send(&edgeproto.StreamMsg{Msg: out})
		default:
			return fmt.Errorf("Unsupported message type received: %v", streamMsg)
		}
	}

	return nil
}

func (s *StreamObjApi) StreamMsgs(key *edgeproto.AppInstKey, cb edgeproto.StreamObjApi_StreamAppInstServer) error {
	ctx := cb.Context()

	// Currently we don't know which controller has the streamer Obj for this key
	// (or if it's even present), so just broadcast to all.
	err := controllerApi.RunJobs(func(arg interface{}, addr string) error {
		if addr == *externalApiAddr {
			// local node
			return s.StreamLocalMsgs(key, cb)
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
		stream, sErr := cmd.StreamLocalMsgs(ctx, key)
		if sErr != nil {
			return sErr
		}
		var sMsg *edgeproto.StreamMsg
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

func (s *StreamObjApi) startStream(ctx context.Context, key *edgeproto.AppInstKey, inCb GenericCb) (*streamSend, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "Start new stream", "key", key)
	streamer := streamObjs.Get(*key)
	if streamer != nil {
		// stream is already in progress for this key
		if streamer.State == edgeproto.StreamState_STREAM_START {
			return nil, key.ExistsError()
		}
	}
	streamer = cloudcommon.NewStreamer()
	streamSendObj := streamSend{}
	streamSendObj.doneCh = make(chan bool, 1)
	streamSendObj.cb = inCb
	streamSendObj.lastRemoteId = uint32(0)
	streamSendObj.streamer = streamer
	streamObjs.Add(*key, streamer)

	return &streamSendObj, nil
}

func (s *StreamObjApi) stopStream(ctx context.Context, key *edgeproto.AppInstKey, streamSendObj *streamSend, objErr error) error {
	log.SpanLog(ctx, log.DebugLevelApi, "Stop stream", "key", key)
	if streamSendObj != nil {
		streamSendObj.mux.Lock()
		defer streamSendObj.mux.Unlock()
		streamSendObj.closeCh = true
		if streamSendObj.streamer != nil {
			streamSendObj.streamer.Stop()
		}
	}

	return nil
}
