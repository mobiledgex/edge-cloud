package testutil

// Stubs for DummyServer.
// Revisit as needed for unit tests.
import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func (s *DummyServer) AddCloudletResMapping(ctx context.Context, in *edgeproto.CloudletResMap) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) RemoveCloudletResMapping(ctx context.Context, in *edgeproto.CloudletResMap) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) FindFlavorMatch(ctx context.Context, in *edgeproto.FlavorMatch) (*edgeproto.FlavorMatch, error) {
	return in, nil
}

func (s *DummyServer) GetCloudletProps(ctx context.Context, in *edgeproto.CloudletProps) (*edgeproto.CloudletProps, error) {
	return in, nil
}

func (s *DummyServer) RevokeAccessKey(ctx context.Context, key *edgeproto.CloudletKey) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) GenerateAccessKey(ctx context.Context, key *edgeproto.CloudletKey) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) UpgradeAccessKey(stream edgeproto.CloudletAccessKeyApi_UpgradeAccessKeyServer) error {
	return nil
}
