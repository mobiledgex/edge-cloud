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

package k8smgmt

import (
	"context"
	"testing"

	"github.com/edgexr/edge-cloud/cloudcommon"
	dme "github.com/edgexr/edge-cloud/d-match-engine/dme-proto"
	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/log"
	"github.com/stretchr/testify/require"
)

func TestGetNetworkPolicy(t *testing.T) {
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	app := edgeproto.App{}
	app.Key.Organization = "devorg"
	app.Key.Name = "myapp"
	app.Key.Version = "1.0"
	app.Deployment = cloudcommon.DeploymentTypeKubernetes
	app.AllowServerless = true
	ci := edgeproto.ClusterInst{}
	ci.Key.ClusterKey.Name = cloudcommon.DefaultMultiTenantCluster
	ci.Key.Organization = cloudcommon.OrganizationMobiledgeX
	ci.Key.CloudletKey.Name = "cloudlet1"
	ci.Key.CloudletKey.Organization = "operorg"
	appInst := edgeproto.AppInst{}
	appInst.Key.AppKey = app.Key
	appInst.Key.ClusterInstKey = *ci.Key.Virtual("autocluster1")

	// Non-multi-tenant cluster does not need a network policy
	ci.MultiTenant = false
	testGetNetworkPolicy(t, ctx, &app, &ci, &appInst, "only valid for namespaced", "")

	ci.MultiTenant = true
	// Network policy, no ports
	testGetNetworkPolicy(t, ctx, &app, &ci, &appInst, "", `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: networkpolicy-devorg-myapp-10-autocluster1
  namespace: devorg-myapp-10-autocluster1
spec:
  podSelector:
    matchLabels:
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: devorg-myapp-10-autocluster1
`)

	// Network policy, with ports
	appInst.MappedPorts = []dme.AppPort{
		{
			// tcp
			Proto:        dme.LProto_L_PROTO_TCP,
			InternalPort: 443,
			PublicPort:   443,
		}, {
			// remapped port
			Proto:        dme.LProto_L_PROTO_TCP,
			InternalPort: 888,
			PublicPort:   818,
		}, {
			// udp
			Proto:        dme.LProto_L_PROTO_UDP,
			InternalPort: 10101,
			PublicPort:   10101,
		}, {
			// 1000 port range, mapped
			Proto:        dme.LProto_L_PROTO_TCP,
			InternalPort: 51000,
			EndPort:      51009,
			PublicPort:   61000,
		},
	}
	testGetNetworkPolicy(t, ctx, &app, &ci, &appInst, "", `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: networkpolicy-devorg-myapp-10-autocluster1
  namespace: devorg-myapp-10-autocluster1
spec:
  podSelector:
    matchLabels:
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: devorg-myapp-10-autocluster1
  - from:
    - ipBlock:
        cidr: 0.0.0.0/0
    ports:
    - protocol: TCP
      port: 443
    - protocol: TCP
      port: 888
    - protocol: UDP
      port: 10101
    - protocol: TCP
      port: 51000
    - protocol: TCP
      port: 51001
    - protocol: TCP
      port: 51002
    - protocol: TCP
      port: 51003
    - protocol: TCP
      port: 51004
    - protocol: TCP
      port: 51005
    - protocol: TCP
      port: 51006
    - protocol: TCP
      port: 51007
    - protocol: TCP
      port: 51008
    - protocol: TCP
      port: 51009
`)

}

func testGetNetworkPolicy(t *testing.T, ctx context.Context, app *edgeproto.App, clusterInst *edgeproto.ClusterInst, appInst *edgeproto.AppInst, expectedErr string, expectedMF string) {
	names, err := GetKubeNames(clusterInst, app, appInst)
	require.Nil(t, err)
	mf, err := GetNetworkPolicy(ctx, app, appInst, names)
	if expectedErr != "" {
		require.NotNil(t, err)
		require.Contains(t, err.Error(), expectedErr)
	} else {
		require.Nil(t, err)
		require.Equal(t, expectedMF, mf)
	}
}
