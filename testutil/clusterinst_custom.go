package testutil

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func (s *DummyServer) DeleteIdleReservableClusterInsts(ctx context.Context, in *edgeproto.IdleReservableClusterInsts) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) GetClusterInstGPUDriverLicenseConfig(ctx context.Context, in *edgeproto.ClusterInstKey) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}
