// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutil

import (
	"context"
	"io"

	"github.com/edgexr/edge-cloud/edgeproto"
	"google.golang.org/grpc"
)

func (s *DummyServer) AddAppAutoProvPolicy(ctx context.Context, apppolicy *edgeproto.AppAutoProvPolicy) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) RemoveAppAutoProvPolicy(ctx context.Context, apppolicy *edgeproto.AppAutoProvPolicy) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) AddAppAlertPolicy(ctx context.Context, alert *edgeproto.AppAlertPolicy) (*edgeproto.Result, error) {
	return &edgeproto.Result{}, nil
}

func (s *DummyServer) RemoveAppAlertPolicy(ctx context.Context, alert *edgeproto.AppAlertPolicy) (*edgeproto.Result, error) {
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
