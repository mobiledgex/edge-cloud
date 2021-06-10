package testutil

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func (s *DummyServer) AddGPUDriverBuild(in *edgeproto.GPUDriverBuildMember, cb edgeproto.GPUDriverApi_AddGPUDriverBuildServer) error {
	return nil
}

func (s *DummyServer) RemoveGPUDriverBuild(in *edgeproto.GPUDriverBuildMember, cb edgeproto.GPUDriverApi_RemoveGPUDriverBuildServer) error {
	return nil
}

func (s *DummyServer) GetGPUDriverBuildURL(ctx context.Context, in *edgeproto.GPUDriverBuildMember) (*edgeproto.GPUDriverBuildURL, error) {
	return &edgeproto.GPUDriverBuildURL{}, nil
}
