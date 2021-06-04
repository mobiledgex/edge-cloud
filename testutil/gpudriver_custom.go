package testutil

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func (s *DummyServer) AddGPUDriverBuild(ctx context.Context, in *edgeproto.GPUDriverBuildMember) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) RemoveGPUDriverBuild(ctx context.Context, in *edgeproto.GPUDriverBuildMember) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) GetGPUDriverBuildURL(ctx context.Context, in *edgeproto.GPUDriverBuildMember) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}
