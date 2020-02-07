package main

import (
	"context"
	"sync"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var AppInstClientQMaxClients = 500

type AppInstClientQ struct {
	data       []*edgeproto.AppInstClient // TODO - buffer for multiple receivers here
	done       bool
	mux        sync.Mutex
	wg         sync.WaitGroup
	Qfull      uint64
	clientChan map[edgeproto.AppInstKey]chan edgeproto.AppInstClient
}

func NewAppInstClientQ() *AppInstClientQ {
	q := AppInstClientQ{}
	//	q.data = make([]*edgeproto.AppInstClient, 100)
	q.clientChan = make(map[edgeproto.AppInstKey]chan edgeproto.AppInstClient)
	return &q
}

func (q *AppInstClientQ) SetRecvChan(ch chan edgeproto.AppInstClient, key edgeproto.AppInstKey) {
	q.mux.Lock()
	defer q.mux.Unlock()
	_, found := q.clientChan[key]
	if found {
		//TODO unsupported now - multiple clients for the same appInst, need to implement it
		log.DebugLog(log.DebugLevelApi, "TOO MANY controller CLIENTS!!!", "key", key)
		return
	}
	log.DebugLog(log.DebugLevelApi, "Set channel for appInst", "key", key)
	q.clientChan[key] = ch
}

func (q *AppInstClientQ) ClearRecvChan(key edgeproto.AppInstKey) {
	q.mux.Lock()
	defer q.mux.Unlock()
	_, found := q.clientChan[key]
	if found {
		delete(q.clientChan, key)
	}
}

func (q *AppInstClientQ) AddClient(clients ...*edgeproto.AppInstClient) {
	q.mux.Lock()
	defer q.mux.Unlock()
	if len(q.data) > AppInstClientQMaxClients {
		// limit len to prevent out of memory if
		// q is not reachable
		q.Qfull++
		return
	}
	for ii, client := range clients {
		log.DebugLog(log.DebugLevelApi, "Got a client from DME", "client", client)
		if clients[ii] != nil {
			//q.data = append(q.data, clients[ii])
			c, found := q.clientChan[client.ClientKey.Key]
			if found && c != nil {
				log.DebugLog(log.DebugLevelApi, "Sending to channel")
				c <- *clients[ii]
			} else {
				log.DebugLog(log.DebugLevelApi, "No channel")
			}
		}
	}
}

func (q *AppInstClientQ) Recv(ctx context.Context, client *edgeproto.AppInstClient) {
	q.AddClient(client)
}
