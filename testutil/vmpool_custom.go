package testutil

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func (s *DummyServer) AddVMPoolMember(ctx context.Context, cloudlet *edgeproto.VMPoolMember) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) RemoveVMPoolMember(ctx context.Context, cloudlet *edgeproto.VMPoolMember) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}
