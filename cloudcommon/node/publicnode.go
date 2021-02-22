package node

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	mextls "github.com/mobiledgex/edge-cloud/tls"
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
	if mextls.IsTestTls() {
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

// PublicCertManager manages refreshing the public cert.
type PublicCertManager struct {
	commonName          string
	tlsMode             mextls.TLSMode
	useGetPublicCertApi bool // denotes whether to use GetPublicCertApi to grab certs or use command line provided cert (should be equivalent to useVaultPki flag)
	getPublicCertApi    cloudcommon.GetPublicCertApi
	cert                *tls.Certificate
	expiresAt           time.Time
	done                bool
	refreshTrigger      chan bool
	refreshThreshold    time.Duration
	refreshRetryDelay   time.Duration
	mux                 sync.Mutex
}

func NewPublicCertManager(commonName string, getPublicCertApi cloudcommon.GetPublicCertApi, tlsCertFile string, tlsKeyFile string) (*PublicCertManager, error) {
	// Nominally letsencrypt certs are valid for 90 days
	// and they recommend refreshing at 30 days to expiration.
	mgr := &PublicCertManager{
		commonName:        commonName,
		refreshTrigger:    make(chan bool, 1),
		refreshThreshold:  30 * 24 * time.Hour,
		refreshRetryDelay: 24 * time.Hour,
		tlsMode:           mextls.ServerAuthTLS,
	}

	if getPublicCertApi != nil {
		mgr.useGetPublicCertApi = true
		mgr.getPublicCertApi = getPublicCertApi
	} else if tlsCertFile != "" && tlsKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(tlsCertFile, tlsKeyFile)
		if err != nil {
			return nil, err
		}
		mgr.cert = &cert
	} else {
		// no tls
		mgr.tlsMode = mextls.NoTLS
	}
	return mgr, nil
}

func (s *PublicCertManager) updateCert(ctx context.Context) error {
	if s.tlsMode == mextls.NoTLS || !s.useGetPublicCertApi {
		// If no tls or using command line certs, do not update
		return nil
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "update public cert", "name", s.commonName)
	pubCert, err := s.getPublicCertApi.GetPublicCert(ctx, s.commonName)
	if err != nil {
		return err
	}
	cert, err := tls.X509KeyPair([]byte(pubCert.Cert), []byte(pubCert.Key))
	if err != nil {
		return err
	}
	s.mux.Lock()
	s.cert = &cert
	expiresIn := time.Duration(pubCert.TTL) * time.Second
	s.expiresAt = time.Now().Add(expiresIn)
	log.SpanLog(ctx, log.DebugLevelInfo, "new cert", "name", s.commonName, "expiresIn", expiresIn, "expiresAt", s.expiresAt)
	s.mux.Unlock()
	return nil
}

// For now this just assumes server-side only TLS.
func (s *PublicCertManager) GetServerTlsConfig(ctx context.Context) (*tls.Config, error) {
	if s.tlsMode == mextls.NoTLS {
		// No tls
		return nil, nil
	}
	if s.cert == nil {
		// make sure we have cert
		err := s.updateCert(ctx)
		if err != nil {
			return nil, err
		}
	}
	config := &tls.Config{
		MinVersion:     tls.VersionTLS12,
		ClientAuth:     tls.NoClientCert,
		GetCertificate: s.GetCertificateFunc(),
	}
	return config, nil
}

func (s *PublicCertManager) GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
		s.mux.Lock()
		defer s.mux.Unlock()
		if s.cert == nil {
			return nil, fmt.Errorf("No certificate available")
		}
		return s.cert, nil
	}
}

func (s *PublicCertManager) StartRefresh() {
	s.done = false
	go func() {
		for {
			s.mux.Lock()
			expiresIn := time.Until(s.expiresAt)
			s.mux.Unlock()
			var waitTime time.Duration
			if expiresIn > s.refreshThreshold {
				waitTime = expiresIn - s.refreshThreshold
			} else {
				// Try once a day
				waitTime = s.refreshRetryDelay
			}
			select {
			case <-time.After(waitTime):
			case <-s.refreshTrigger:
			}
			span := log.StartSpan(log.DebugLevelInfo, "refresh public cert")
			ctx := log.ContextWithSpan(context.Background(), span)
			if s.done {
				log.SpanLog(ctx, log.DebugLevelInfo, "refresh public cert done")
				span.Finish()
				break
			}
			err := s.updateCert(ctx)
			log.SpanLog(ctx, log.DebugLevelInfo, "updated cert", "name", s.commonName, "err", err)
			span.Finish()
		}
	}()
}

func (s *PublicCertManager) StopRefresh() {
	s.done = true
	select {
	case s.refreshTrigger <- true:
	default:
	}
}
