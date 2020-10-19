package node

import (
	"context"
	"crypto/tls"
	"os"

	"github.com/mobiledgex/edge-cloud/log"
)

// Third party services that we deploy all have their own letsencrypt-public
// issued certificate, with a CA pool that includes the vault internal public CAs.
// This allows mTLS where the public node uses a public cert and our internal
// services use an internal vault pki cert.
// Examples of such services are Jaeger, ElasticSearch, etc.
func (s *NodeMgr) GetPublicClientTlsConfig(ctx context.Context) (*tls.Config, error) {
	if s.tlsClientIssuer == NoTlsClientIssuer {
		// unit test mode
		return nil, nil
	}
	tlsOpts := []TlsOp{
		WithPublicCAPool(),
	}
	if e2e := os.Getenv("E2ETEST_TLS"); e2e != "" {
		// skip verifying cert if e2e-tests, because cert
		// will be self-signed
		log.SpanLog(ctx, log.DebugLevelInfo, "public client tls e2e-test mode")
		tlsOpts = append(tlsOpts, WithTlsSkipVerify(true))
	}
	return s.InternalPki.GetClientTlsConfig(ctx,
		s.CommonName(),
		s.tlsClientIssuer,
		[]MatchCA{},
		tlsOpts...)
}
