package testutil

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func (s *DummyServer) AddAppAutoProvPolicy(ctx context.Context, apppolicy *edgeproto.AppAutoProvPolicy) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) RemoveAppAutoProvPolicy(ctx context.Context, apppolicy *edgeproto.AppAutoProvPolicy) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) AddAppUserDefinedAlert(ctx context.Context, alert *edgeproto.AppUserDefinedAlert) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) RemoveAppUserDefinedAlert(ctx context.Context, alert *edgeproto.AppUserDefinedAlert) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}
