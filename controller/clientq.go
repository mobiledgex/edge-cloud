package main

import (
	"context"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var AppInstClientQMaxClients = 500

type AppInstClientQ struct {
	//	client    edgeproto.AppInst - TODO - need to creat a new queue for each request
	data       []*edgeproto.AppInstClient
	done       bool
	mux        sync.Mutex
	wg         sync.WaitGroup
	Qfull      uint64
	clientChan chan edgeproto.AppInstClient
}

func NewAppInstClientQ() *AppInstClientQ {
	q := AppInstClientQ{}
	q.data = make([]*edgeproto.AppInstClient, 100)
	return &q
}

func (q *AppInstClientQ) SetRecvChan(ch chan edgeproto.AppInstClient) {
	q.clientChan = ch
}

func (q *AppInstClientQ) Start(addr string) error {
	q.mux.Lock()
	defer q.mux.Unlock()
	q.done = false
	q.wg.Add(1)
	go q.RunPush()
	return nil
}

func (q *AppInstClientQ) RunPush() {
	for !q.done {
		select {
		case <-time.After(time.Second):
		}
		if q.done {
			break
		}
		q.mux.Lock()
		if len(q.data) == 0 {
			q.mux.Unlock()
			continue
		}
		data := q.data
		q.data = make([]*edgeproto.AppInstClient, 0)
		q.mux.Unlock()

		for _, metric := range data {
			log.DebugLog(log.DebugLevelMetrics,
				"metric new point", "metric", metric)
		}
	}
	q.wg.Done()
}

func (q *AppInstClientQ) Stop() {
	q.done = true
	q.wg.Wait()
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
	for ii, _ := range clients {
		if clients[ii] != nil {
			q.data = append(q.data, clients[ii])
			if q.clientChan != nil {
				q.clientChan <- *clients[ii]
			}
		}
	}
}

func (q *AppInstClientQ) Recv(ctx context.Context, client *edgeproto.AppInstClient) {
	log.SpanLog(ctx, log.DebugLevelApi, "Got a client from DME", "client", client)
	q.AddClient(client)
}
