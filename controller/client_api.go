package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
)

type AppInstClientApi struct {
}

var appInstClientApi = AppInstClientApi{}
var appInstClientSendMany *notify.AppInstClientSendMany

func InitAppInstClientApi() {
	appInstClientSendMany = notify.NewAppInstClientSendMany()
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

	// Request this AppInst to be sent
	recvCh := make(chan edgeproto.AppInstClient, int(settingsApi.Get().MaxTrackedDmeClients))
	services.clientQ.SetRecvChan(in.Key, recvCh)

	if !appInstClientKeyApi.HasApp(&in.Key) {
		appInstClientKeyApi.Update(cb.Context(), in, 0)
	}
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
	if services.clientQ.LockAndClearRecvChan(in.Key, recvCh) == 0 {
		appInstClientKeyApi.Delete(cb.Context(), in, 0)
	}
	services.clientQ.Unlock()
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
