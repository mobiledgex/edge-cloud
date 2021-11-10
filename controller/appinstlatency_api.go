package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type AppInstLatencyApi struct {
	all  *AllApis
	sync *Sync
}

func NewAppInstLatencyApi(sync *Sync, all *AllApis) *AppInstLatencyApi {
	appInstLatencyApi := AppInstLatencyApi{}
	appInstLatencyApi.all = all
	appInstLatencyApi.sync = sync
	return &appInstLatencyApi
}

func (s *AppInstLatencyApi) RequestAppInstLatency(ctx context.Context, in *edgeproto.AppInstLatency) (*edgeproto.Result, error) {
	cloudcommon.SetAppInstKeyDefaults(&in.Key)

	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}
	// Check that appinst exists
	appInstInfo := edgeproto.AppInst{}
	if !s.all.appInstApi.cache.Get(&in.Key, &appInstInfo) {
		return nil, in.Key.NotFoundError()
	}

	conn, err := notifyRootConnect(ctx, *notifyRootAddrs)
	if err != nil {
		return nil, err
	}
	client := edgeproto.NewAppInstLatencyApiClient(conn)
	ctx, cancel := context.WithTimeout(ctx, node.DefaultDebugTimeout)
	defer cancel()
	return client.RequestAppInstLatency(ctx, in)
}
