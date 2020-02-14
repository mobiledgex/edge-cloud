package main

import (
	"context"
	"sync"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
)

type AppInstClientApi struct {
	queueMux       sync.Mutex
	appInstClients []*edgeproto.AppInstClient
	clientChan     map[edgeproto.AppInstKey][]chan edgeproto.AppInstClient
}

var appInstClientApi = AppInstClientApi{}
var appInstClientSendMany *notify.AppInstClientSendMany

func InitAppInstClientApi() {
	appInstClientSendMany = notify.NewAppInstClientSendMany()
	appInstClientApi.appInstClients = make([]*edgeproto.AppInstClient, 0)
	appInstClientApi.clientChan = make(map[edgeproto.AppInstKey][]chan edgeproto.AppInstClient)
}

func (s *AppInstClientApi) SetRecvChan(ctx context.Context, in *edgeproto.AppInstClientKey, ch chan edgeproto.AppInstClient) {
	s.queueMux.Lock()
	defer s.queueMux.Unlock()
	_, found := s.clientChan[in.Key]
	if !found {
		s.clientChan[in.Key] = []chan edgeproto.AppInstClient{ch}
	} else {
		s.clientChan[in.Key] = append(s.clientChan[in.Key], ch)
		for ii, client := range s.appInstClients {
			if client.ClientKey.Key.Matches(&in.Key) {
				ch <- *s.appInstClients[ii]
			}
		}
	}
	if !appInstClientKeyApi.HasApp(&in.Key) {
		appInstClientKeyApi.Update(ctx, in, 0)
	}
}

// Returns number of channels in the list that are left
func (s *AppInstClientApi) ClearRecvChan(ctx context.Context, in *edgeproto.AppInstClientKey, ch chan edgeproto.AppInstClient) int {
	s.queueMux.Lock()
	defer s.queueMux.Unlock()
	_, found := s.clientChan[in.Key]
	if !found {
		log.SpanLog(ctx, log.DebugLevelApi, "No client channels found for appInst", "appInst", in.Key)
		return -1
	}
	for ii, c := range s.clientChan[in.Key] {
		if c == ch {
			// Found channel - delete it
			s.clientChan[in.Key] = append(s.clientChan[in.Key][:ii], s.clientChan[in.Key][ii+1:]...)
			retLen := len(s.clientChan[in.Key])
			if retLen == 0 {
				delete(s.clientChan, in.Key)
				appInstClientKeyApi.Delete(ctx, in, 0)
			}
			return retLen
		}
	}
	// We didn't find a channel....log it and return -1
	log.SpanLog(ctx, log.DebugLevelApi, "Channel not found", "key", in.Key, "chan", ch)
	return -1
}

func (s *AppInstClientApi) AddAppInstClient(ctx context.Context, client *edgeproto.AppInstClient) {
	s.queueMux.Lock()
	defer s.queueMux.Unlock()
	if client != nil {
		// Queue full - remove the oldest one(first) and append the new one
		if len(s.appInstClients) == int(settingsApi.Get().MaxTrackedDmeClients) {
			s.appInstClients = s.appInstClients[1:]
		}
		cList, found := s.clientChan[client.ClientKey.Key]
		if !found {
			log.SpanLog(ctx, log.DebugLevelApi, "No receivers for this appInst")
			return
		}
		s.appInstClients = append(s.appInstClients, client)
		for _, c := range cList {
			if c != nil {
				c <- *client
			} else {
				log.SpanLog(ctx, log.DebugLevelApi, "Nil Channel")
			}
		}
	}
}

func (s *AppInstClientApi) Recv(ctx context.Context, client *edgeproto.AppInstClient) {
	s.AddAppInstClient(ctx, client)
}

func (s *AppInstClientApi) Prune(ctx context.Context, keys map[edgeproto.AppInstClientKey]struct{}) {}

func (s *AppInstClientApi) ShowAppInstClient(in *edgeproto.AppInstClientKey, cb edgeproto.AppInstClientApi_ShowAppInstClientServer) error {
	// Check if the AppInst exists
	if !appInstApi.HasKey(&in.Key) {
		return in.Key.NotFoundError()
	}

	// Since we don't care about the cluster developer and name set them to ""
	in.Key.ClusterInstKey.ClusterKey.Name = ""
	in.Key.ClusterInstKey.Developer = ""

	// Resuest this AppInst to be sent
	recvCh := make(chan edgeproto.AppInstClient, int(settingsApi.Get().MaxTrackedDmeClients))
	s.SetRecvChan(cb.Context(), in, recvCh)

	done := false
	for !done {
		select {
		case <-cb.Context().Done():
			done = true
		case appInstClient := <-recvCh:
			if err := cb.Send(&appInstClient); err != nil {
				done = true
			}
		}
	}
	// Clear channel and while holding a lock delete this appInstClientKey
	s.ClearRecvChan(cb.Context(), in, recvCh)
	return nil
}

func (s *AppInstClientApi) Flush(ctx context.Context, notifyId int64) {}

type AppInstClientKeyApi struct {
	sync  *Sync
	store edgeproto.AppInstClientKeyStore
	cache edgeproto.AppInstClientKeyCache
}

var appInstClientKeyApi = AppInstClientKeyApi{}

func InitAppInstClientKeyApi(sync *Sync) {
	appInstClientKeyApi.sync = sync
	appInstClientKeyApi.store = edgeproto.NewAppInstClientKeyStore(sync.store)
	edgeproto.InitAppInstClientKeyCache(&appInstClientKeyApi.cache)
	sync.RegisterCache(&appInstClientKeyApi.cache)
}

func (s *AppInstClientKeyApi) Update(ctx context.Context, in *edgeproto.AppInstClientKey, rev int64) {
	s.cache.Update(ctx, in, rev)
}

func (s *AppInstClientKeyApi) Delete(ctx context.Context, in *edgeproto.AppInstClientKey, rev int64) {
	s.cache.Delete(ctx, in, rev)
}

func (s *AppInstClientKeyApi) Flush(ctx context.Context, notifyId int64) {
	s.cache.Flush(ctx, notifyId)
}

func (s *AppInstClientKeyApi) Prune(ctx context.Context, keys map[edgeproto.AppInstKey]struct{}) {}

func (s *AppInstClientKeyApi) HasApp(key *edgeproto.AppInstKey) bool {
	return s.cache.HasKey(key)
}
