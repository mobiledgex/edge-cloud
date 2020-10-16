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
	doneCh       chan bool
	mux          sync.Mutex
	closeCh      bool
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
		s.streamSendObj.msgCh <- res.Message
	}
	return nil
}

func InitStreamObjApi(sync *Sync) {
	streamObjApi.sync = sync
	streamObjApi.store = edgeproto.NewStreamObjStore(sync.store)
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

func (s *StreamObjApi) startStream(ctx context.Context, key *edgeproto.AppInstKey, inCb GenericCb) (*streamSend, error) {
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

		go s.addStream(ctx, key, &streamSendObj)
		return &streamSendObj, err
	}

	return nil, err
}

func (s *StreamObjApi) stopStream(ctx context.Context, key *edgeproto.AppInstKey, streamSendObj *streamSend, objErr error) error {
	if streamSendObj != nil {
		streamSendObj.mux.Lock()
		close(streamSendObj.msgCh)
		streamSendObj.closeCh = true
		streamSendObj.mux.Unlock()

		// wait for addStream to finish
		<-streamSendObj.doneCh
	}

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

func (s *StreamObjApi) addStream(ctx context.Context, key *edgeproto.AppInstKey, streamSendObj *streamSend) {
	lastStreamId := uint32(0)
	cancel := streamObjInfoApi.cache.WatchKey(key, func(ctx context.Context) {
		info := edgeproto.StreamObjInfo{}
		if !streamObjInfoApi.cache.Get(key, &info) {
			return
		}
		streamSendObj.mux.Lock()
		defer streamSendObj.mux.Unlock()
		if streamSendObj.closeCh {
			return
		}
		for _, streamMsg := range info.Msgs {
			if streamMsg.Id <= lastStreamId {
				continue
			}
			lastStreamId = streamMsg.Id
			streamSendObj.msgCh <- streamMsg.Msg
		}
	})
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
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to add msg to stream", "key", key, "msg", msg, "err", err)
		}
	}
	cancel()
	streamSendObj.doneCh <- true
}

type StreamObjInfoApi struct {
	sync  *Sync
	store edgeproto.StreamObjInfoStore
	cache edgeproto.StreamObjInfoCache
}

var streamObjInfoApi = StreamObjInfoApi{}

func InitStreamObjInfoApi(sync *Sync) {
	streamObjInfoApi.sync = sync
	streamObjInfoApi.store = edgeproto.NewStreamObjInfoStore(sync.store)
	edgeproto.InitStreamObjInfoCache(&streamObjInfoApi.cache)
	sync.RegisterCache(&streamObjInfoApi.cache)
}

func (s *StreamObjInfoApi) Update(ctx context.Context, in *edgeproto.StreamObjInfo, rev int64) {
	if in == nil {
		log.SpanLog(ctx, log.DebugLevelApi, "StreamObj is missing")
		return
	}
	lease, err := s.sync.store.Grant(context.Background(), StreamLeaseTimeoutSec)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "Lease grant failed for streamObjInfo, using controllerAliveLease", "key", in.Key, "err", err)
		lease = controllerAliveLease
	}
	s.store.Put(ctx, in, nil, objstore.WithLease(lease))
}

func (s *StreamObjInfoApi) Delete(ctx context.Context, in *edgeproto.StreamObjInfo, rev int64) {
	// no-op
}

func (s *StreamObjInfoApi) Flush(ctx context.Context, notifyId int64) {
	// no-op
}

func (s *StreamObjInfoApi) Prune(ctx context.Context, keys map[edgeproto.AppInstKey]struct{}) {
	// no-op
}
