package cloudcommon

import (
	"context"

	edgetls "github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/vault"
)

// GetPublicCertApi abstracts the way the public cert is retrieved.
// Certain services, like DME running on a Cloudlet, may need to connect
// to the controller to get a public cert from Vault.
type GetPublicCertApi interface {
	GetPublicCert(ctx context.Context, commonName string) (*vault.PublicCert, error)
}

// TestPublicCertApi implements GetPublicCertApi for unit/e2e testing
type TestPublicCertApi struct {
	GetCount int
}

func (s *TestPublicCertApi) GetPublicCert(ctx context.Context, commonName string) (*vault.PublicCert, error) {
	cert := &vault.PublicCert{}
	cert.Cert = edgetls.LocalTestCert
	cert.Key = edgetls.LocalTestKey
	// 24 hours in seconds
	cert.TTL = 24 * 3600
	s.GetCount++
	return cert, nil
}
