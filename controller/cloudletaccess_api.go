package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
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
	key, pubPEM, err := s.accessKeyServer.UpgradeAccessKey(stream)
	if err != nil {
		return err
	}
	if pubPEM == "" {
		// no new key
		return nil
	}
	// save newly generated key
	ctx := stream.Context()
	err = s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cloudlet := edgeproto.Cloudlet{}
		if !s.store.STMGet(stm, key, &cloudlet) {
			// deleted
			return nil
		}
		cloudlet.CrmAccessPublicKey = pubPEM
		cloudlet.CrmAccessKeyUpgradeRequired = false
		s.store.STMPut(stm, &cloudlet)
		return nil
	})
	return err
}
