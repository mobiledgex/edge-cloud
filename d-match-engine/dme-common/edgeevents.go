package dmecommon

import (
	"context"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// TODO: DESCRIPTIONS
type EdgeEventsHandler interface {
	AddClientKey(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey CookieKey, sendFunc func(event *dme.ServerEdgeEvent))
	RemoveClientKey(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey CookieKey)
	RemoveAppInstKey(ctx context.Context, appInstKey edgeproto.AppInstKey)
	SendLatencyRequestEdgeEvent(ctx context.Context, appInstKey edgeproto.AppInstKey)
	ProcessLatencySamples(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey CookieKey, samples []float64) (*dme.Latency, bool)
	SendAppInstStateEvent(ctx context.Context, appInst *DmeAppInst, appInstKey edgeproto.AppInstKey, eventType dme.ServerEdgeEvent_ServerEventType)
	SendEdgeEventToClient(ctx context.Context, serverEdgeEvent *dme.ServerEdgeEvent, appInstKey edgeproto.AppInstKey, cookieKey CookieKey)
}
