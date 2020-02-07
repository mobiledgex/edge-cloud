package main

import (
	"sync"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	log "github.com/mobiledgex/edge-cloud/log"
	"golang.org/x/net/context"
)

type ClientsMap struct {
	sync.RWMutex
	clients map[edgeproto.AppInstKey][]edgeproto.AppInstClient
}

var clientsMap *ClientsMap

func InitAppInstClients() {
	clientsMap = new(ClientsMap)
	clientsMap.clients = make(map[edgeproto.AppInstKey][]edgeproto.AppInstClient)
}

// Add a new client to the list of clients
// TODO - need to have a finite number of clients in an array
func UpdateClientsBuffer(ctx context.Context, msg *edgeproto.AppInstClient) {
	clientsMap.Lock()
	defer clientsMap.Unlock()
	list, found := clientsMap.clients[msg.ClientKey.Key]
	if !found {
		clientsMap.clients[msg.ClientKey.Key] = []edgeproto.AppInstClient{*msg}
		log.DebugLog(log.DebugLevelDmereq, "New AppInst client - never seen appinst", "AppInstClient", clientsMap.clients[msg.ClientKey.Key][0])
	} else {
		clientsMap.clients[msg.ClientKey.Key] = append(list, *msg)
		log.DebugLog(log.DebugLevelDmereq, "New AppInst client key exists", "AppInstClient", list[len(list)-1], "key", msg.ClientKey.Key)
	}
	log.DebugLog(log.DebugLevelDmereq, "New array", "LIST", clientsMap.clients[msg.ClientKey.Key])
	// If there is an outstanding request for this appInst, send it out
	if appInstClientKeyCache.HasKey(msg.ClientKey.GetKey()) {
		ClientSender.Update(ctx, msg)
	}
}

// If an AppInst is deleted, clean up all the clients from it
func PurgeAppInstClients(ctx context.Context, msg *edgeproto.AppInstKey) {
	clientsMap.Lock()
	defer clientsMap.Unlock()
	_, found := clientsMap.clients[*msg]
	if found {
		delete(clientsMap.clients, *msg)
	}
}

func SendCachedClients(ctx context.Context, old *edgeproto.AppInstClientKey, new *edgeproto.AppInstClientKey) {
	DebugDumpClients() // DEBUG
	clientsMap.Lock()
	defer clientsMap.Unlock()
	list, found := clientsMap.clients[new.Key]
	if !found {
		log.DebugLog(log.DebugLevelDmereq, "No AppInst clients found ", "AppInstClient", new)
		return
	}
	for ii, c := range list {
		log.DebugLog(log.DebugLevelDmereq, "Sending client ", "AppInstClient", c)
		ClientSender.Update(ctx, &list[ii])
	}
}

// Debug Dump
func DebugDumpClients() {
	log.DebugLog(log.DebugLevelDmereq, "XXX DUMP CLIENTS")
	clientsMap.Lock()
	defer clientsMap.Unlock()
	for k, list := range clientsMap.clients {
		log.DebugLog(log.DebugLevelDmereq, "Clients for AppInst", "AppInstKey", k)
		for _, c := range list {
			log.DebugLog(log.DebugLevelDmereq, "\t", "Client", c)
		}
	}
}

// TODO - function to periodically timeout the clients
