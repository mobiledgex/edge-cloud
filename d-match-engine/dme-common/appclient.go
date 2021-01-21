package dmecommon

import (
	"fmt"
	"sync"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
	"golang.org/x/net/context"
)

type ClientsMap struct {
	sync.RWMutex
	clientsByApp map[edgeproto.AppKey][]edgeproto.AppInstClient
}

var clientsMap *ClientsMap

var ClientSender *notify.AppInstClientSend
var AppInstClientKeyCache edgeproto.AppInstClientKeyCache

func InitAppInstClients() {
	clientsMap = new(ClientsMap)
	clientsMap.clientsByApp = make(map[edgeproto.AppKey][]edgeproto.AppInstClient)
}

// Add a new client to the list of clients
func UpdateClientsBuffer(ctx context.Context, msg *edgeproto.AppInstClient) {
	clientsMap.Lock()
	defer clientsMap.Unlock()
	mapKey := msg.ClientKey.AppInstKey.AppKey
	_, found := clientsMap.clientsByApp[mapKey]
	if !found {
		clientsMap.clientsByApp[mapKey] = []edgeproto.AppInstClient{*msg}
	} else {
		// We need to either update, or add the client to the list
		for ii, c := range clientsMap.clientsByApp[mapKey] {
			// Found the same client from before
			if c.ClientKey.UniqueId == msg.ClientKey.UniqueId &&
				c.ClientKey.UniqueIdType == msg.ClientKey.UniqueIdType {
				clientsMap.clientsByApp[mapKey] = append(clientsMap.clientsByApp[mapKey][:ii],
					clientsMap.clientsByApp[mapKey][ii+1:]...)
				break
			}
		}
		//  We reached the limit of clients - remove the first one
		if len(clientsMap.clientsByApp[mapKey]) == int(Settings.MaxTrackedDmeClients) {
			clientsMap.clientsByApp[mapKey] = clientsMap.clientsByApp[mapKey][1:]
		}
		clientsMap.clientsByApp[mapKey] = append(clientsMap.clientsByApp[mapKey], *msg)
	}
	// If there is an outstanding request for this appInstClientKey - send it out
	AppInstClientKeyCache.Show(&edgeproto.AppInstClientKey{}, func(obj *edgeproto.AppInstClientKey) error {
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
	_, found := clientsMap.clientsByApp[msg.AppKey]
	if found {
		// walk the list and keep only the clients that don't match the filter
		jj := 0
		for _, c := range clientsMap.clientsByApp[msg.AppKey] {
			// Remove matching clients
			if msg.AppKey.Matches(&c.ClientKey.AppInstKey.AppKey) &&
				msg.ClusterInstKey.CloudletKey.Matches(&c.ClientKey.AppInstKey.ClusterInstKey.CloudletKey) {
				continue
			}
			clientsMap.clientsByApp[msg.AppKey][jj] = c
			jj++
		}
		// truncate the list
		clientsMap.clientsByApp[msg.AppKey] = clientsMap.clientsByApp[msg.AppKey][:jj]
	}
}

func SendCachedClients(ctx context.Context, old *edgeproto.AppInstClientKey, new *edgeproto.AppInstClientKey) {
	// Check if we have an outstanding streaming request which would be a superset
	err := AppInstClientKeyCache.Show(&edgeproto.AppInstClientKey{}, func(obj *edgeproto.AppInstClientKey) error {
		// if we found an exact match - it's this clients
		if new.Matches(obj) {
			return nil
		}
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
	list, found := clientsMap.clientsByApp[new.AppInstKey.AppKey]
	// Possible exact match for the map
	if found {
		for ii := range list {
			// Check if we match the complete filter
			if list[ii].ClientKey.Matches(new, edgeproto.MatchFilter()) {
				ClientSender.Update(ctx, &list[ii])
			}
		}
		return
	}
	// Walk the entire map to find all possible matches
	for _, list := range clientsMap.clientsByApp {
		for ii := range list {
			// Check if we match the complete filter
			if list[ii].ClientKey.Matches(new, edgeproto.MatchFilter()) {
				ClientSender.Update(ctx, &list[ii])
			}
		}
	}
}

// TODO - function to periodically timeout the clients
