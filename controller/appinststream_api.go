package main

import (
	"context"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type AppInstStreamApi struct {
	sync  *Sync
	store edgeproto.AppInstStreamStore
	cache edgeproto.AppInstStreamCache
}

var appInstStreamApi = AppInstStreamApi{}

func InitAppInstStreamApi(sync *Sync) {
	appInstStreamApi.sync = sync
	appInstStreamApi.store = edgeproto.NewAppInstStreamStore(sync.store)
	edgeproto.InitAppInstStreamCache(&appInstStreamApi.cache)
	sync.RegisterCache(&appInstStreamApi.cache)
}

func (s *AppInstStreamApi) StreamAppInst(in *edgeproto.AppInstStream, cb edgeproto.AppInstStreamApi_StreamAppInstServer) error {
	ctx := cb.Context()
	appInstStream := edgeproto.AppInstStream{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, &appInstStream) {
			return in.Key.NotFoundError()
		}
		return nil
	})
	if err != nil {
		return err
	}
	lastMsgId := uint32(0)
	for _, streamMsg := range appInstStream.Msgs {
		cb.Send(streamMsg)
		lastMsgId = streamMsg.Id
	}
	if appInstStream.State != edgeproto.StreamState_STREAM_START {
		return nil
	}

	log.SpanLog(ctx, log.DebugLevelApi, "Stream is in progress, wait for new messages", "appInstStream", appInstStream)
	done := make(chan bool, 1)
	cancel := s.cache.WatchKey(&in.Key, func(ctx context.Context) {
		if !s.cache.Get(&in.Key, &appInstStream) {
			return
		}
		for _, streamMsg := range appInstStream.Msgs {
			if streamMsg.Id <= lastMsgId {
				continue
			}
			cb.Send(streamMsg)
			lastMsgId = streamMsg.Id
		}
		if appInstStream.State != edgeproto.StreamState_STREAM_START {
			done <- true
		}
	})

	select {
	case <-done:
	}
	cancel()
	return nil
}

func (s *AppInstStreamApi) startStream(ctx context.Context, key *edgeproto.AppInstKey) error {
	appInstStream := edgeproto.AppInstStream{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, key, &appInstStream) {
			return key.ExistsError()
		}
		appInstStream.Key = *key
		appInstStream.State = edgeproto.StreamState_STREAM_START
		s.store.STMPut(stm, &appInstStream)
		return nil
	})

	return err
}

func (s *AppInstStreamApi) stopStream(ctx context.Context, key *edgeproto.AppInstKey) error {
	appInstStream := edgeproto.AppInstStream{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, key, &appInstStream) {
			return key.NotFoundError()
		}
		appInstStream.State = edgeproto.StreamState_STREAM_STOP
		s.store.STMPut(stm, &appInstStream)
		return nil
	})

	return err
}

func (s *AppInstStreamApi) addStream(ctx context.Context, key *edgeproto.AppInstKey, msg string) {
	appInstStream := edgeproto.AppInstStream{}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, key, &appInstStream) {
			return key.NotFoundError()
		}
		appInstStream.LastId++
		appInstStream.Msgs = append(appInstStream.Msgs, &edgeproto.StreamMsg{
			Id:  appInstStream.LastId,
			Msg: msg,
		})
		s.store.STMPut(stm, &appInstStream)
		return nil
	})
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelApi, "Failed to add stream", "key", key, "msg", msg, "err", err)
	}
}
