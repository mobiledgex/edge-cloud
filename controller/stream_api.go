package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	grpc "google.golang.org/grpc"
)

var streamObjApi = StreamObjApi{}

var StreamTimeout = 30 * time.Minute
var StreamLeaseTimeoutSec = int64(60 * 60) // 1 hour

type streamSend struct {
	msgCh        chan string
	cb           GenericCb
	lastRemoteId uint32
	mux          *sync.Mutex
	doneCh       chan bool
}

type StreamObjApi struct {
	sync      *Sync
	store     edgeproto.StreamObjStore
	cache     edgeproto.StreamObjCache
	streamBuf map[edgeproto.AppInstKey]*streamSend
	mux       sync.Mutex
}

type GenericCb interface {
	Send(*edgeproto.Result) error
	grpc.ServerStream
}

type CbWrapper struct {
	GenericCb
	key edgeproto.AppInstKey
}

func (s *CbWrapper) Send(res *edgeproto.Result) error {
	if res != nil {
		if streamSendObj, ok := streamObjApi.streamBuf[s.key]; ok {
			streamSendObj.mux.Lock()
			defer streamSendObj.mux.Unlock()
			streamSendObj.msgCh <- res.Message
		}
	}
	return nil
}

func InitStreamObjApi(sync *Sync) {
	streamObjApi.sync = sync
	streamObjApi.store = edgeproto.NewStreamObjStore(sync.store)
	streamObjApi.streamBuf = make(map[edgeproto.AppInstKey]*streamSend)
	edgeproto.InitStreamObjCache(&streamObjApi.cache)
	sync.RegisterCache(&streamObjApi.cache)
}

func (s *StreamObjApi) StreamMsgs(key *edgeproto.AppInstKey, cb edgeproto.StreamObjApi_StreamAppInstServer) error {
	ctx := cb.Context()
	streamObj := edgeproto.StreamObj{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, key, &streamObj) {
			return key.NotFoundError()
		}
		return nil
	})
	if err != nil {
		// return nil, if no object present
		if err.Error() == key.NotFoundError().Error() {
			log.SpanLog(ctx, log.DebugLevelApi, "Stream obj is not present", "key", key)
			return nil
		}
		return err
	}
	lastMsgId := uint32(0)
	for _, streamMsg := range streamObj.Msgs {
		cb.Send(streamMsg)
		lastMsgId = streamMsg.Id
	}
	if streamObj.State != edgeproto.StreamState_STREAM_START {
		if streamObj.ErrorMsg != "" {
			return fmt.Errorf("%s", streamObj.ErrorMsg)
		}
		return nil
	}

	log.SpanLog(ctx, log.DebugLevelApi, "Stream is in progress, wait for new messages", "streamObj", streamObj, "lastMsgId", lastMsgId)
	done := make(chan bool, 1)
	cancel := s.cache.WatchKey(key, func(ctx context.Context) {
		if !s.cache.Get(key, &streamObj) {
			return
		}
		for _, streamMsg := range streamObj.Msgs {
			if streamMsg.Id <= lastMsgId {
				continue
			}
			cb.Send(streamMsg)
			lastMsgId = streamMsg.Id
		}
		if streamObj.State != edgeproto.StreamState_STREAM_START {
			done <- true
		}
	})

	select {
	case <-done:
		if streamObj.ErrorMsg != "" {
			err = fmt.Errorf("%s", streamObj.ErrorMsg)
		} else {
			err = nil
		}
	case <-time.After(StreamTimeout):
		err = fmt.Errorf("Timed out waiting for stream messages for app instance %v, please retry again later", key)
	}
	cancel()
	return err
}

func (s *StreamObjApi) startStream(ctx context.Context, key *edgeproto.AppInstKey, inCb GenericCb) error {
	streamObj := edgeproto.StreamObj{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, key, &streamObj) {
			if streamObj.State == edgeproto.StreamState_STREAM_START {
				// stream is already in progress for this key
				return key.ExistsError()
			}
		}
		lease, gErr := s.sync.store.Grant(context.Background(), StreamLeaseTimeoutSec)
		if gErr != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Stream grant lease failed", "key", key, "err", gErr)
		} else {
			streamObj.Lease = lease
		}
		streamObj.Key = *key
		streamObj.State = edgeproto.StreamState_STREAM_START
		streamObj.Msgs = []*edgeproto.StreamMsg{}
		streamObj.LastId = 0
		streamObj.ErrorMsg = ""
		s.store.STMPut(stm, &streamObj, objstore.WithLease(lease))
		return nil
	})

	if err == nil {
		streamSendObj := streamSend{}
		streamSendObj.msgCh = make(chan string, 50)
		streamSendObj.doneCh = make(chan bool, 1)
		streamSendObj.cb = inCb
		streamSendObj.lastRemoteId = uint32(0)
		streamSendObj.mux = &sync.Mutex{}

		s.mux.Lock()
		s.streamBuf[*key] = &streamSendObj
		s.mux.Unlock()
		go s.addStream(ctx, key)
	}

	return err
}

func (s *StreamObjApi) stopStream(ctx context.Context, key *edgeproto.AppInstKey, objErr error) error {
	s.mux.Lock()
	if streamSendObj, ok := s.streamBuf[*key]; ok {
		close(streamSendObj.msgCh)
		// wait for addStream to finish
		<-streamSendObj.doneCh
		streamSendObj.doneCh <- true
		delete(s.streamBuf, *key)
	}
	s.mux.Unlock()

	streamObj := edgeproto.StreamObj{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, key, &streamObj) {
			return key.NotFoundError()
		}
		streamObj.State = edgeproto.StreamState_STREAM_STOP
		if objErr != nil {
			streamObj.ErrorMsg = objErr.Error()
		}
		s.store.STMPut(stm, &streamObj, objstore.WithLease(streamObj.Lease))
		return nil
	})

	return err
}

func (s *StreamObjApi) addStream(ctx context.Context, key *edgeproto.AppInstKey) {
	streamSendObj, ok := s.streamBuf[*key]
	if !ok {
		return
	}
	for msg := range streamSendObj.msgCh {
		streamSendObj.cb.Send(&edgeproto.Result{Message: msg})
		streamObj := edgeproto.StreamObj{}
		err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
			if !s.store.STMGet(stm, key, &streamObj) {
				return key.NotFoundError()
			}
			if len(streamObj.Msgs) > 0 &&
				streamObj.Msgs[len(streamObj.Msgs)-1].Msg == msg {
				// duplicate message, ignore
			} else {
				streamObj.LastId++
				streamObj.Msgs = append(streamObj.Msgs, &edgeproto.StreamMsg{
					Id:  streamObj.LastId,
					Msg: msg,
				})
			}
			s.store.STMPut(stm, &streamObj, objstore.WithLease(streamObj.Lease))
			return nil
		})
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to add stream", "key", key, "msg", msg, "err", err)
		}
	}
	streamSendObj.doneCh <- true
}

func (s *StreamObjApi) Update(ctx context.Context, in *edgeproto.StreamObj, rev int64) {
	if in == nil {
		log.SpanLog(ctx, log.DebugLevelApi, "StreamObj is missing")
		return
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	streamSendObj, ok := streamObjApi.streamBuf[in.Key]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelApi, "No active streams", "key", in.Key)
		return
	}
	streamSendObj.mux.Lock()
	defer streamSendObj.mux.Unlock()
	for _, streamMsg := range in.Msgs {
		if streamMsg.Id <= streamSendObj.lastRemoteId {
			// skip already streamed messages
			continue
		}
		streamSendObj.lastRemoteId = streamMsg.Id
		streamSendObj.msgCh <- streamMsg.Msg
	}
}

func (s *StreamObjApi) Delete(ctx context.Context, in *edgeproto.StreamObj, rev int64) {
	// no-op
}

func (s *StreamObjApi) Flush(ctx context.Context, notifyId int64) {
	// no-op
}

func (s *StreamObjApi) Prune(ctx context.Context, keys map[edgeproto.AppInstKey]struct{}) {
	// no-op
}
