package testutil

import (
	"context"
	"io"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"google.golang.org/grpc"
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

func (s *DummyServer) ShowCloudletsForAppDeployment(*edgeproto.DeploymentCloudletRequest, edgeproto.AppApi_ShowCloudletsForAppDeploymentServer) error {
	return nil
}

func (s *DummyServer) ShowFlavorsForCloudlet(*edgeproto.CloudletKey, edgeproto.CloudletApi_ShowFlavorsForCloudletServer) error {
	return nil
}

// minimal bits not currently generated for cloudletkey.proto and app.proto
// in support of CLI test for an rpc that streams cloudletkeys as its result
//
type ShowCloudletsForAppDeployment struct {
	Data map[string]edgeproto.CloudletKey
	grpc.ServerStream
	Ctx context.Context
}

func (x *ShowCloudletsForAppDeployment) Init() {
	x.Data = make(map[string]edgeproto.CloudletKey)
}

func (x *ShowCloudletsForAppDeployment) Send(m *edgeproto.CloudletKey) error {
	x.Data[m.Name] = *m
	return nil
}

func (x *ShowCloudletsForAppDeployment) Context() context.Context {
	return x.Ctx
}

var ShowCloudletsForAppDeploymentExtraCount = 0

func (x *ShowCloudletsForAppDeployment) ReadStream(stream edgeproto.AppApi_ShowCloudletsForAppDeploymentClient, err error) {

	x.Data = make(map[string]edgeproto.CloudletKey)
	if err != nil {
		return
	}
	for {
		obj, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		x.Data[obj.Name] = *obj
	}
}

func (x *AppCommonApi) ShowCloudletsForAppDeployment(ctx context.Context, filter *edgeproto.DeploymentCloudletRequest, showData *ShowCloudletsForAppDeployment) error {

	if x.internal_api != nil {
		showData.Ctx = ctx
		return x.internal_api.ShowCloudletsForAppDeployment(filter, showData)
	} else {

		stream, err := x.client_api.ShowCloudletsForAppDeployment(ctx, filter)
		showData.ReadStream(stream, err)
		return err
	}
}
