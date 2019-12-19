package testutil

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func (s *DummyServer) AddAutoProvPolicyCloudlet(ctx context.Context, cloudlet *edgeproto.AutoProvPolicyCloudlet) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) RemoveAutoProvPolicyCloudlet(ctx context.Context, cloudlet *edgeproto.AutoProvPolicyCloudlet) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}
