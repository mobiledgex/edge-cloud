package main

import (
	"fmt"
	"sync"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	log "github.com/mobiledgex/edge-cloud/log"
	"golang.org/x/net/context"
)

type ClientUuid struct {
	UniqueId     string
	UniqueIdType string
}

type ClientsMap struct {
	sync.RWMutex
	clients          map[ClientUuid]*edgeproto.AppInstClient
	clientsByAppInst map[edgeproto.AppInstKey][]edgeproto.AppInstClient
}

var clientsMap *ClientsMap

func InitAppInstClients() {
	clientsMap = new(ClientsMap)
	clientsMap.clients = make(map[ClientUuid]*edgeproto.AppInstClient)
	clientsMap.clientsByAppInst = make(map[edgeproto.AppInstKey][]edgeproto.AppInstClient)
}

// Add a new client to the list of clients
func UpdateClientsBuffer(ctx context.Context, msg *edgeproto.AppInstClient) {
	clientsMap.Lock()
	defer clientsMap.Unlock()
	clientKey := ClientUuid{
		UniqueId:     msg.ClientKey.Key.UniqueId,
		UniqueIdType: msg.ClientKey.Key.UniqueIdType,
	}
	// if it existed before, update the clientsByAppInst
	if client, found := clientsMap.clients[clientKey]; found {
		// remove from the list of clients in the old appInstance
		list, found := clientsMap.clientsByAppInst[*client.ClientKey.Key.Appinstkey]
		if !found {
			log.SpanLog(ctx, log.DebugLevelInfo, "Found an orphan client", "client", msg, "old pointer", client)
		} else {
			for ii, _ := range list {
				if list[ii].ClientKey.Key.UniqueId == client.ClientKey.Key.UniqueId &&
					list[ii].ClientKey.Key.UniqueIdType == client.ClientKey.Key.UniqueIdType {
					// remove from the old appInst list
					clientsMap.clientsByAppInst[*client.ClientKey.Key.Appinstkey] =
						append(clientsMap.clientsByAppInst[*client.ClientKey.Key.Appinstkey][:ii],
							clientsMap.clientsByAppInst[*client.ClientKey.Key.Appinstkey][ii+1:]...)
					break
				}
			}
		}
	}
	// update the value in the clients map
	clientsMap.clients[clientKey] = msg

	mapKey := *msg.ClientKey.Key.Appinstkey
	list, found := clientsMap.clientsByAppInst[mapKey]
	if !found {
		clientsMap.clientsByAppInst[mapKey] = []edgeproto.AppInstClient{*msg}
	} else {
		//  We reached the limit of clients - remove the first one
		if len(list) == int(dmecommon.Settings.MaxTrackedDmeClients) {
			list = list[1:]
		}
		clientsMap.clientsByAppInst[mapKey] = append(list, *msg)
	}

	// If there is an outstanding request for this appInstClientKey - send it out
	appInstClientKeyCache.Show(&edgeproto.AppInstClientKey{}, func(obj *edgeproto.AppInstClientKey) error {
		if msg.ClientKey.Matches(obj, edgeproto.MatchFilter()) {
			ClientSender.Update(ctx, msg)
			return fmt.Errorf("Found match - just send once")
		}
		return nil
	})
}

// If an AppInst is deleted, clean up all the clients from it
func PurgeAppInstClients(ctx context.Context, msg *edgeproto.AppInstKey) {
	clientsMap.Lock()
	defer clientsMap.Unlock()
	list, found := clientsMap.clientsByAppInst[*msg]
	if found {
		// walk the list and delete all individual clients
		for _, c := range list {
			key := ClientUuid{
				UniqueId:     c.ClientKey.Key.UniqueId,
				UniqueIdType: c.ClientKey.Key.UniqueIdType,
			}
			delete(clientsMap.clients, key)
		}
		delete(clientsMap.clientsByAppInst, *msg)

	}
}

func SendCachedClients(ctx context.Context, old *edgeproto.AppInstClientKey, new *edgeproto.AppInstClientKey) {
	// Check if we have an outstanding streaming request which would be a superset
	err := appInstClientKeyCache.Show(&edgeproto.AppInstClientKey{}, func(obj *edgeproto.AppInstClientKey) error {
		if new.Matches(obj, edgeproto.MatchFilter()) {
			return fmt.Errorf("Already streaming for this superset")
		}
		return nil
	})
	if err != nil {
		return
	}
	clientsMap.RLock()
	defer clientsMap.RUnlock()
	list, found := clientsMap.clientsByAppInst[*new.Key.Appinstkey]
	// AppInst based request
	if found {
		for ii := range list {
			ClientSender.Update(ctx, &list[ii])
		}
		return
	}
	// Any partial match will be sent here
	for _, client := range clientsMap.clients {
		if client.ClientKey.Matches(new, edgeproto.MatchFilter()) {
			ClientSender.Update(ctx, client)
		}
	}
}

// TODO - function to periodically timeout the clients
