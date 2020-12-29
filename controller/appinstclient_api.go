package main

import (
	"context"
	"io"
	"sync"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"google.golang.org/grpc"
)

type AppInstClientApi struct {
	queueMux       sync.Mutex
	appInstClients []*edgeproto.AppInstClient
	clientChan     map[edgeproto.AppInstClientKey][]chan edgeproto.AppInstClient
}

var appInstClientApi = AppInstClientApi{}
var appInstClientSendMany *notify.AppInstClientSendMany

func InitAppInstClientApi() {
	appInstClientSendMany = notify.NewAppInstClientSendMany()
	appInstClientApi.appInstClients = make([]*edgeproto.AppInstClient, 0)
	appInstClientApi.clientChan = make(map[edgeproto.AppInstClientKey][]chan edgeproto.AppInstClient)
}

func (s *AppInstClientApi) SetRecvChan(ctx context.Context, in *edgeproto.AppInstClientKey, ch chan edgeproto.AppInstClient) {
	s.queueMux.Lock()
	defer s.queueMux.Unlock()
	_, found := s.clientChan[*in]
	if !found {
		s.clientChan[*in] = []chan edgeproto.AppInstClient{ch}
	} else {
		s.clientChan[*in] = append(s.clientChan[*in], ch)
	}

	// use filter match option, since not all parts of the key may be set
	for ii, client := range s.appInstClients {
		if client.ClientKey.Matches(in, edgeproto.MatchFilter()) {
			ch <- *s.appInstClients[ii]
		}
	}

	if !appInstClientKeyApi.HasKey(in) {
		appInstClientKeyApi.Update(ctx, in, 0)
	}
}

// Returns number of channels in the list that are left
func (s *AppInstClientApi) ClearRecvChan(ctx context.Context, in *edgeproto.AppInstClientKey, ch chan edgeproto.AppInstClient) int {
	s.queueMux.Lock()
	defer s.queueMux.Unlock()
	_, found := s.clientChan[*in]
	if !found {
		log.SpanLog(ctx, log.DebugLevelApi, "No client channels found for appInst", "appInst", in.AppInstKey)
		return -1
	}
	for ii, c := range s.clientChan[*in] {
		if c == ch {
			// Found channel - delete it
			s.clientChan[*in] = append(s.clientChan[*in][:ii], s.clientChan[*in][ii+1:]...)
			retLen := len(s.clientChan[*in])
			if retLen == 0 {
				delete(s.clientChan, *in)
				appInstClientKeyApi.Delete(ctx, in, 0)
				// We also need to clean up our local buffer - it will be out of sync since DME won't update it
				for ii, client := range s.appInstClients {
					if client.ClientKey.Matches(in, edgeproto.MatchFilter()) {
						if len(s.appInstClients) > ii+1 {
							s.appInstClients = append(s.appInstClients[:ii], s.appInstClients[ii+1:]...)
						} else {
							// if this is already the last element
							s.appInstClients = s.appInstClients[:ii]
						}
					}
				}
			}
			return retLen
		}
	}
	// We didn't find a channel....log it and return -1
	log.SpanLog(ctx, log.DebugLevelApi, "Channel not found", "key", in, "chan", ch)
	return -1
}

func (s *AppInstClientApi) AddAppInstClient(ctx context.Context, client *edgeproto.AppInstClient) {
	s.queueMux.Lock()
	defer s.queueMux.Unlock()
	if client == nil {
		return
	}
	sendList := []chan edgeproto.AppInstClient{}
	for k, cList := range s.clientChan {
		if client.ClientKey.Matches(&k, edgeproto.MatchFilter()) {
			sendList = append(sendList, cList...)
		}
	}
	if len(sendList) == 0 {
		log.SpanLog(ctx, log.DebugLevelApi, "No receivers for this key", "client", client)
		return
	}
	// We need to either update, or add the client to the list
	for ii, c := range s.appInstClients {
		// Found the same client from before
		if c.ClientKey.UniqueId == client.ClientKey.UniqueId &&
			c.ClientKey.UniqueIdType == client.ClientKey.UniqueIdType {
			if len(s.appInstClients) > ii+1 {
				// remove this client the and append it at the end, since it's new
				s.appInstClients = append(s.appInstClients[:ii], s.appInstClients[ii+1:]...)
			} else {
				// if this is already the last element
				s.appInstClients = s.appInstClients[:ii]
			}
			break
		}
	}
	// Queue full - remove the oldest one(first) and append the new one
	if len(s.appInstClients) == int(settingsApi.Get().MaxTrackedDmeClients) {
		s.appInstClients = s.appInstClients[1:]
	}
	s.appInstClients = append(s.appInstClients, client)
	for _, c := range sendList {
		if c != nil {
			c <- *client
		} else {
			log.SpanLog(ctx, log.DebugLevelApi, "Nil Channel")
		}
	}
}

func (s *AppInstClientApi) RecvAppInstClient(ctx context.Context, client *edgeproto.AppInstClient) {
	s.AddAppInstClient(ctx, client)
}

func (s *AppInstClientApi) Prune(ctx context.Context, keys map[edgeproto.AppInstClientKey]struct{}) {}

func (s *AppInstClientApi) StreamAppInstClientsLocal(in *edgeproto.AppInstClientKey, cb edgeproto.AppInstClientApi_StreamAppInstClientsLocalServer) error {
	// Request this AppInst to be sent
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

func (s *AppInstClientApi) ShowAppInstClient(in *edgeproto.AppInstClientKey, cb edgeproto.AppInstClientApi_ShowAppInstClientServer) error {
	var connsMux sync.Mutex
	var ctrlConns []*grpc.ClientConn

	// Check if the App exists
	if !appApi.HasApp(&in.AppInstKey.AppKey) {
		return in.NotFoundError()
	}

	// Since we don't care about the cluster developer and name set them to ""
	in.AppInstKey.ClusterInstKey.ClusterKey.Name = ""
	in.AppInstKey.ClusterInstKey.Organization = ""

	ctrlConns = make([]*grpc.ClientConn, 0)
	done := false
	err := controllerApi.RunJobs(func(arg interface{}, addr string) error {
		if addr == *externalApiAddr {
			// local node
			err := s.StreamAppInstClientsLocal(in, cb)
			// Close grpc connections to other controllers
			connsMux.Lock()
			done = true
			for _, conn := range ctrlConns {
				conn.Close()
			}
			connsMux.Unlock()
			return err
		} else { // This will get clients from the remote controllers and proxy them as well
			// connect to remote node
			conn, err := ControllerConnect(cb.Context(), addr)
			if err != nil {
				return err
			}
			connsMux.Lock()
			if done {
				conn.Close()
				connsMux.Unlock()
				return nil
			} else {
				// Strore the connection, to close when the local api terminates
				ctrlConns = append(ctrlConns, conn)
			}
			connsMux.Unlock()

			appInstClient := edgeproto.NewAppInstClientApiClient(conn)
			// Recv forever - when the local API call terminates, it will close the connection
			stream, err := appInstClient.StreamAppInstClientsLocal(context.Background(), in)
			if err != nil {
				log.SpanLog(cb.Context(), log.DebugLevelApi, "Failed to dispatch Show to controller", "controller", addr,
					"key", in, "err", err)
				return err
			}
			for {
				client, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						err = nil
					}
					break
				}
				err = cb.Send(client)
				if err != nil {
					log.SpanLog(cb.Context(), log.DebugLevelApi, "Failed to print a client", "client", client)
					break
				}
			}
			return nil
		}
	}, nil)
	if err != nil {
		log.SpanLog(cb.Context(), log.DebugLevelApi, "Failed to dispatch Show to all controllers", "key", in, "err", err)
		return err
	}
	return nil
}

func (s *AppInstClientApi) Flush(ctx context.Context, notifyId int64) {}

type AppInstClientKeyApi struct {
	sync  *Sync
	cache edgeproto.AppInstClientKeyCache
}

var appInstClientKeyApi = AppInstClientKeyApi{}

func InitAppInstClientKeyApi(sync *Sync) {
	appInstClientKeyApi.sync = sync
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

func (s *AppInstClientKeyApi) HasKey(key *edgeproto.AppInstClientKey) bool {
	return s.cache.HasKey(key)
}
