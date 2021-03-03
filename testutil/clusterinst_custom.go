package testutil

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func (s *DummyServer) DeleteIdleReservableClusterInsts(ctx context.Context, in *edgeproto.IdleReservableClusterInsts) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}
