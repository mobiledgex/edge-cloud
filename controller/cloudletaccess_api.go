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

package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/edgexr/edge-cloud/cloud-resource-manager/accessapi"
	"github.com/edgexr/edge-cloud/cloudcommon/node"
	"github.com/edgexr/edge-cloud/edgeproto"
	"github.com/edgexr/edge-cloud/integration/process"
	"github.com/edgexr/edge-cloud/log"
)

// Issue certificate to RegionalCloudlet service.
func (s *CloudletApi) IssueCert(ctx context.Context, req *edgeproto.IssueCertRequest) (*edgeproto.IssueCertReply, error) {
	verified := node.ContextGetAccessKeyVerified(ctx)
	if verified == nil {
		// should never reach here if it wasn't verified
		return nil, fmt.Errorf("Client authentication not verified")
	}
	certId := node.CertId{
		CommonName: req.CommonName,
		Issuer:     node.CertIssuerRegionalCloudlet,
	}
	vaultCert, err := nodeMgr.InternalPki.IssueVaultCertDirect(ctx, certId)
	if err != nil {
		return nil, err
	}
	reply := &edgeproto.IssueCertReply{
		PublicCertPem: string(vaultCert.PublicCertPEM),
		PrivateKeyPem: string(vaultCert.PrivateKeyPEM),
	}
	return reply, nil
}

// Get CAs for RegionalCloudlet service. To match the Vault API,
// each request only returns one CA.
func (s *CloudletApi) GetCas(ctx context.Context, req *edgeproto.GetCasRequest) (*edgeproto.GetCasReply, error) {
	// Should be verified, but we don't really care because these are public certs
	cab, err := nodeMgr.InternalPki.GetVaultCAsDirect(ctx, req.Issuer)
	if err != nil {
		return nil, err
	}
	reply := &edgeproto.GetCasReply{
		CaChainPem: string(cab),
	}
	return reply, err
}

func (s *CloudletApi) UpgradeAccessKey(stream edgeproto.CloudletAccessKeyApi_UpgradeAccessKeyServer) error {
	ctx := stream.Context()
	log.SpanLog(ctx, log.DebugLevelApi, "upgrade access key")
	return s.accessKeyServer.UpgradeAccessKey(stream, s.commitAccessPublicKey)
}

func (s *CloudletApi) commitAccessPublicKey(ctx context.Context, key *edgeproto.CloudletKey, pubPEM string, haRole process.HARole) error {
	return s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cloudlet := edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, key, &cloudlet) {
			// deleted
			return nil
		}
		log.SpanLog(ctx, log.DebugLevelApi, "commit upgraded key")
		if haRole == process.HARoleSecondary {
			cloudlet.SecondaryCrmAccessPublicKey = pubPEM
			cloudlet.SecondaryCrmAccessKeyUpgradeRequired = false
		} else {
			cloudlet.CrmAccessPublicKey = pubPEM
			cloudlet.CrmAccessKeyUpgradeRequired = false
		}
		s.store.STMPut(stm, &cloudlet)
		return nil
	})
}

func (s *CloudletApi) GetAccessData(ctx context.Context, req *edgeproto.AccessDataRequest) (*edgeproto.AccessDataReply, error) {
	verified := node.ContextGetAccessKeyVerified(ctx)
	if verified == nil {
		// should never reach here if it wasn't verified
		return nil, fmt.Errorf("Client authentication not verified")
	}
	cloudlet := &edgeproto.Cloudlet{}
	if !s.all.cloudletApi.cache.Get(&verified.Key, cloudlet) {
		return nil, verified.Key.NotFoundError()
	}
	vaultClient := accessapi.NewVaultClient(cloudlet, vaultConfig, *region)
	handler := accessapi.NewControllerHandler(vaultClient)
	return handler.GetAccessData(ctx, req)
}
