package main

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	grpc "google.golang.org/grpc"
)

var (
	streamObjs   = cloudcommon.StreamObj{}
	streamObjApi = StreamObjApi{}

	StreamTimeout = 30 * time.Minute

	SaveOnStreamObj      = true
	DonotSaveOnStreamObj = false
)

type streamSend struct {
	cb       GenericCb
	mux      sync.Mutex
	streamer *cloudcommon.Streamer
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

func InitStreamObjApi(sync *Sync) {
	streamObjApi.sync = sync
	streamObjApi.store = edgeproto.NewStreamObjStore(sync.store)
	edgeproto.InitStreamObjCache(&streamObjApi.cache)
	sync.RegisterCache(&streamObjApi.cache)
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

func (s *StreamObjApi) StreamLocalMsgs(key *edgeproto.AppInstKey, cb edgeproto.StreamObjApi_StreamLocalMsgsServer) error {
	ctx := cb.Context()
	log.SpanLog(ctx, log.DebugLevelApi, "Stream obj messages", "key", key)
	streamer := streamObjs.Get(*key)
	if streamer == nil {
		// stream not found, nothing to show
		log.SpanLog(ctx, log.DebugLevelApi, "Stream obj not found", "key", key)
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

func (s *StreamObjApi) StreamMsgs(key *edgeproto.AppInstKey, cb edgeproto.StreamObjApi_StreamAppInstServer) error {
	ctx := cb.Context()

	if *externalApiAddr == "" {
		// unit test
		return s.StreamLocalMsgs(key, cb)
	}

	// Currently we don't know which controller has the streamer Obj for this key
	// (or if it's even present), so just broadcast to all.

	// In case CRM loses connection to controller during an action in-progress
	// (for ex: during CreateClusterInst), and reconnects to a different controller
	// than it did earlier, then end-user might not see any status messages.
	// For now we won't handle this case, as it will just affect the status updates
	// which shouldn't be a big deal
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

func (s *StreamObjApi) startStream(ctx context.Context, key *edgeproto.AppInstKey, inCb GenericCb, saveOnStreamObj bool) (*streamSend, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "Start new stream", "key", key)
	streamer := streamObjs.Get(*key)
	if streamer != nil {
		// stream is already in progress for this key
		if streamer.State == edgeproto.StreamState_STREAM_START {
			return nil, key.ExistsError()
		}
	}

	if saveOnStreamObj {
		err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			streamObj := edgeproto.StreamObj{}
			streamObj.Key = *key
			streamObj.Status = edgeproto.StatusInfo{}
			// Init stream obj regardless of it being present or not
			s.store.STMPut(stm, &streamObj)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	streamer = cloudcommon.NewStreamer()
	streamSendObj := streamSend{}
	streamSendObj.cb = inCb
	streamSendObj.streamer = streamer
	streamObjs.Add(*key, streamer)

	return &streamSendObj, nil
}

func (s *StreamObjApi) stopStream(ctx context.Context, key *edgeproto.AppInstKey, streamSendObj *streamSend, objErr error) error {
	log.SpanLog(ctx, log.DebugLevelApi, "Stop stream", "key", key, "err", objErr)
	if streamSendObj != nil {
		streamSendObj.mux.Lock()
		defer streamSendObj.mux.Unlock()
		if streamSendObj.streamer != nil {
			if objErr != nil {
				streamSendObj.streamer.Publish(objErr)
			}
			streamSendObj.streamer.Stop()
		}
	}

	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		streamObj := edgeproto.StreamObj{}
		if !s.store.STMGet(stm, key, &streamObj) {
			// if stream obj is deleted, then ignore emptying
			// the status obj
			return nil
		}
		streamObj.Status = edgeproto.StatusInfo{}
		s.store.STMPut(stm, &streamObj)
		return nil
	})

	return nil
}

func (s *StreamObjApi) CleanupStreamObj(ctx context.Context, in *edgeproto.StreamObj) error {
	return s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		streamObj := edgeproto.StreamObj{}
		if !s.store.STMGet(stm, &in.Key, &streamObj) {
			// already removed
			return nil
		}
		s.store.STMDel(stm, &in.Key)
		return nil
	})
}

// Status from info will always contain the full status update list,
// changes we copy to status that is saved to etcd is only the diff
// from the last update.
func (s *StreamObjApi) UpdateStatus(ctx context.Context, infoStatus *edgeproto.StatusInfo, key *edgeproto.AppInstKey) {
	if len(infoStatus.Msgs) <= 0 {
		return
	}
	s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		streamObj := edgeproto.StreamObj{}
		if !s.store.STMGet(stm, key, &streamObj) {
			streamObj.Key = *key
			streamObj.Status = edgeproto.StatusInfo{}
		}
		lastMsgId := int(streamObj.Status.MsgCount)
		if lastMsgId < len(infoStatus.Msgs) {
			streamObj.Status.Msgs = []string{}
			for ii := lastMsgId; ii < len(infoStatus.Msgs); ii++ {
				streamObj.Status.Msgs = append(streamObj.Status.Msgs, infoStatus.Msgs[ii])
			}
			streamObj.Status.MsgCount += uint32(len(streamObj.Status.Msgs))
			s.store.STMPut(stm, &streamObj)
		}
		return nil
	})
}
