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

// TODO: stream forever
func (s *AppInstClientApi) ShowAppInstClient(in *edgeproto.AppInstClient, cb edgeproto.AppInstClientApi_ShowAppInstClientServer) error {
	// DUmp debug
	log.DebugLog(log.DebugLevelInfo, "DUMP appInstClients")
	for _, client := range services.clientQ.data {
		log.DebugLog(log.DebugLevelInfo, "appInstClient", "client", client)
	}

	for _, client := range services.clientQ.data {
		if client == nil {
			continue
		}
		if err := cb.Send(client); err != nil {
			return err
		}
	}
	recvCh := make(chan edgeproto.AppInstClient, 1)
	client := edgeproto.AppInstClient{}
	services.clientQ.SetRecvChan(recvCh)
	done := false
	for !done {
		log.DebugLog(log.DebugLevelInfo, "Waiting for more data....")
		select {
		case <-cb.Context().Done():
			done = true
		case client = <-recvCh:
			if err := cb.Send(&client); err != nil {
				services.clientQ.SetRecvChan(nil)
				return err
			}
		}
		if services.clientQ.done {
			done = true
		}
	}
	services.clientQ.SetRecvChan(nil)
	return nil
}

func (s *AppInstClientApi) Flush(ctx context.Context, notifyId int64) {}
