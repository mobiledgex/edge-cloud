package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type AppInstLatencyApi struct {
	sync *Sync
}

var appInstLatencyApi = AppInstLatencyApi{}

func InitAppInstLatencyApi(sync *Sync) {
	appInstLatencyApi.sync = sync
}

func (s *AppInstLatencyApi) RequestAppInstLatency(ctx context.Context, in *edgeproto.AppInstLatency) (*edgeproto.Result, error) {
	// If ClusterKey.Name already exists, then don't set
	// any default value for it
	if in.Key.ClusterInstKey.ClusterKey.Name == "" {
		var app edgeproto.App
		if !appApi.cache.Get(&in.Key.AppKey, &app) {
			return nil, in.Key.AppKey.NotFoundError()
		}
		if app.Deployment == cloudcommon.DeploymentTypeVM {
			in.Key.ClusterInstKey.ClusterKey.Name = cloudcommon.DefaultVMCluster
		}
	}

	// Check that appinst exists
	appInstInfo := edgeproto.AppInst{}
	if !appInstApi.cache.Get(&in.Key, &appInstInfo) {
		return nil, in.Key.NotFoundError()
	}

	err := in.Key.ValidateKey()
	if err != nil {
		return nil, err
	}
	// populate the clusterinst developer from the app developer if not already present
	if in.Key.ClusterInstKey.Organization == "" {
		in.Key.ClusterInstKey.Organization = in.Key.AppKey.Organization
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
