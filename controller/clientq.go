package main

import (
	"context"
	"sync"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var AppInstClientQMaxClients = 100

type AppInstClientQ struct {
	mux            sync.Mutex
	appInstClients []*edgeproto.AppInstClient
	clientChan     map[edgeproto.AppInstKey][]chan edgeproto.AppInstClient
}

func NewAppInstClientQ() *AppInstClientQ {
	q := AppInstClientQ{}
	q.appInstClients = make([]*edgeproto.AppInstClient, 0)
	q.clientChan = make(map[edgeproto.AppInstKey][]chan edgeproto.AppInstClient)
	return &q
}

func (q *AppInstClientQ) SetRecvChan(key edgeproto.AppInstKey, ch chan edgeproto.AppInstClient) {
	q.mux.Lock()
	defer q.mux.Unlock()
	_, found := q.clientChan[key]
	if !found {
		q.clientChan[key] = []chan edgeproto.AppInstClient{ch}
	} else {
		q.clientChan[key] = append(q.clientChan[key], ch)
		for ii, client := range q.appInstClients {
			if client.ClientKey.Key.Matches(&key) {
				ch <- *q.appInstClients[ii]
			}
		}
	}
}

// Returns numbere of channels in thee list that are left
func (q *AppInstClientQ) ClearRecvChan(key edgeproto.AppInstKey, ch chan edgeproto.AppInstClient) int {
	q.mux.Lock()
	defer q.mux.Unlock()
	_, found := q.clientChan[key]
	if !found {
		log.DebugLog(log.DebugLevelApi, "Not client channels found for appInst", "appInst", key)
		return -1
	}
	for ii, c := range q.clientChan[key] {
		if c == ch {
			// Found channel - delete it
			q.clientChan[key] = append(q.clientChan[key][:ii], q.clientChan[key][ii+1:]...)
			retLen := len(q.clientChan[key])
			if retLen == 0 {
				delete(q.clientChan, key)
			}
			return retLen
		}
	}
	// We didn't find a channel....log it and return -1
	log.DebugLog(log.DebugLevelApi, "Channel not found", "key", key, "chan", ch)
	return -1
}

func (q *AppInstClientQ) AddClient(clients ...*edgeproto.AppInstClient) {
	q.mux.Lock()
	defer q.mux.Unlock()
	for ii, client := range clients {
		if clients[ii] != nil {
			// Queue full - remove the oldest one(first) and append the new one
			if len(q.appInstClients) == AppInstClientQMaxClients {
				q.appInstClients = q.appInstClients[1:]
			}
			q.appInstClients = append(q.appInstClients, clients[ii])
			cList, found := q.clientChan[client.ClientKey.Key]
			if !found {
				log.DebugLog(log.DebugLevelApi, "No recievers for this appInst")
				return
			}
			for _, c := range cList {
				if c != nil {
					c <- *clients[ii]
				} else {
					log.DebugLog(log.DebugLevelApi, "Nil Channel")
				}
			}
		}
	}
}

func (q *AppInstClientQ) Recv(ctx context.Context, client *edgeproto.AppInstClient) {
	q.AddClient(client)
}
