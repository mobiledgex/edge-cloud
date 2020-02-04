package main

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"golang.org/x/net/context"
)

var clientsMap map[edgeproto.AppInstKey][]edgeproto.AppInstClient

func UpdateClientsBuffer(ctx context.Context, msg *edgeproto.AppInstClient) {
	// TODO - update
}
