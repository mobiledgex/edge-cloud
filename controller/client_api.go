package main

import (
	"context"
	"sync"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
)

type AppInstClientApi struct {
	requests map[string]*ClientReq
	mux      sync.Mutex
}

type ClientReq struct {
	req  edgeproto.AppInstKey
	q    *AppInstClientQ
	done chan bool
}

var appInstClientApi = AppInstClientApi{}
var appInstClientSendMany *notify.AppInstClientSendMany

func InitAppInstClientApi() {
	appInstClientApi.requests = make(map[string]*ClientReq)
	appInstClientSendMany = notify.NewAppInstClientSendMany()
}

func (s *AppInstClientApi) Prune(ctx context.Context, keys map[edgeproto.AppInstClientKey]struct{}) {}

func (s *AppInstClientApi) ShowAppInstClient(in *edgeproto.AppInstClientKey, cb edgeproto.AppInstClientApi_ShowAppInstClientServer) error {
	// Request this AppInst to be sent
	// TODO - check if it exists already and then just do the different thing
	recvCh := make(chan edgeproto.AppInstClient, 1)
	services.clientQ.SetRecvChan(recvCh, in.Key)

	log.DebugLog(log.DebugLevelApi, "Send request for an appInst", "appinst", in)
	appInstClientKeyApi.Update(cb.Context(), in, 0)

	for _, client := range services.clientQ.data {
		if client == nil {
			continue
		}
		if err := cb.Send(client); err != nil {
			return err
		}
	}
	done := false
	appInstClient := edgeproto.AppInstClient{}
	for !done {
		log.DebugLog(log.DebugLevelInfo, "Waiting for more data....")
		select {
		case <-cb.Context().Done():
			done = true
		case appInstClient = <-recvCh:
			if err := cb.Send(&appInstClient); err != nil {
				done = true
			}
		}
		if services.clientQ.done {
			done = true
		}
	}
	services.clientQ.ClearRecvChan(in.Key)
	appInstClientKeyApi.Delete(cb.Context(), in, 0)
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
	log.DebugLog(log.DebugLevelApi, "Update appinst clientKey", "client", in)
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
