package dmecommon

import (
	"context"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type EdgeEventsHandler interface {
	GetVersionProperties() map[string]string
	AddClientKey(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey CookieKey, lastLoc *dme.Loc, carrier string, sendFunc func(event *dme.ServerEdgeEvent))
	RemoveClientKey(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey CookieKey)
	UpdateClientLastLocation(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey CookieKey, lastLoc *dme.Loc)
	RemoveAppInstKey(ctx context.Context, appInstKey edgeproto.AppInstKey)
	SendLatencyRequestEdgeEvent(ctx context.Context, appInstKey edgeproto.AppInstKey)
	ProcessLatencySamples(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey CookieKey, samples []*dme.Sample) (*dme.Statistics, error)
	SendAppInstStateEvent(ctx context.Context, appInst *DmeAppInst, appInstKey edgeproto.AppInstKey, eventType dme.ServerEdgeEvent_ServerEventType)
	SendEdgeEventToClient(ctx context.Context, serverEdgeEvent *dme.ServerEdgeEvent, appInstKey edgeproto.AppInstKey, cookieKey CookieKey)
}

type EmptyEdgeEventsHandler struct{}

func (e *EmptyEdgeEventsHandler) AddClientKey(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey CookieKey, lastLoc *dme.Loc, carrier string, sendFunc func(event *dme.ServerEdgeEvent)) {
	log.DebugLog(log.DebugLevelDmereq, "AddClientKey not implemented for EmptyEdgeEventHandler. Returning")
	return
}

func (e *EmptyEdgeEventsHandler) RemoveClientKey(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey CookieKey) {
	log.DebugLog(log.DebugLevelDmereq, "RemoveClientKey not implemented for EmptyEdgeEventHandler. Returning")
	return
}

func (e *EmptyEdgeEventsHandler) UpdateClientLastLocation(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey CookieKey, lastLoc *dme.Loc) {
	log.DebugLog(log.DebugLevelDmereq, "UpdateClientLastLocation not implemented for EmptyEdgeEventHandler. Returning")
	return
}

func (e *EmptyEdgeEventsHandler) RemoveAppInstKey(ctx context.Context, appInstKey edgeproto.AppInstKey) {
	log.DebugLog(log.DebugLevelDmereq, "RemoveAppInstKey not implemented for EmptyEdgeEventHandler. Returning")
	return
}

func (e *EmptyEdgeEventsHandler) SendLatencyRequestEdgeEvent(ctx context.Context, appInstKey edgeproto.AppInstKey) {
	log.DebugLog(log.DebugLevelDmereq, "SendLatencyRequestEdgeEvent not implemented for EmptyEdgeEventHandler. Returning")
	return
}

func (e *EmptyEdgeEventsHandler) ProcessLatencySamples(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey CookieKey, samples []*dme.Sample) (*dme.Statistics, error) {
	log.DebugLog(log.DebugLevelDmereq, "ProcessLatency in EmptyEdgeEventHandler returning fake latency")
	fakestats := &dme.Statistics{
		Avg:        3.21,
		Min:        1.34,
		Max:        6.3453,
		StdDev:     1.0,
		Variance:   1.0,
		NumSamples: 7,
		Timestamp: &dme.Timestamp{
			Seconds: 1,
			Nanos:   1000000000,
		},
	}
	return fakestats, nil
}

func (e *EmptyEdgeEventsHandler) SendAppInstStateEvent(ctx context.Context, appInst *DmeAppInst, appInstKey edgeproto.AppInstKey, eventType dme.ServerEdgeEvent_ServerEventType) {
	log.DebugLog(log.DebugLevelDmereq, "SendAppInstStateEvent not implemented for EmptyEdgeEventHandler. Returning")
	return
}

func (e *EmptyEdgeEventsHandler) SendEdgeEventToClient(ctx context.Context, serverEdgeEvent *dme.ServerEdgeEvent, appInstKey edgeproto.AppInstKey, cookieKey CookieKey) {
	log.DebugLog(log.DebugLevelDmereq, "SendEdgeEventToClient not implemented for EmptyEdgeEventHandler. Returning")
	return
}

func (e *EmptyEdgeEventsHandler) GetVersionProperties() map[string]string {
	return map[string]string{}
}
