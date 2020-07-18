package testutil

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func (s *DummyServer) AddCloudletVMPoolMember(ctx context.Context, cloudlet *edgeproto.CloudletVMPoolMember) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) RemoveCloudletVMPoolMember(ctx context.Context, cloudlet *edgeproto.CloudletVMPoolMember) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}
