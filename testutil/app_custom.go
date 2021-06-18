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

func (s *DummyServer) ShowCloudletsForAppDeployment(*edgeproto.DeploymentCloudletRequest, edgeproto.AppApi_ShowCloudletsForAppDeploymentServer) error {
	return nil
}

func (s *DummyServer) ShowFlavorsForCloudlet(*edgeproto.CloudletKey, edgeproto.CloudletApi_ShowFlavorsForCloudletServer) error {
	return nil
}
