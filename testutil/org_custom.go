package testutil

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func (s *DummyServer) OrganizationInUse(ctx context.Context, in *edgeproto.Organization) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}
