package node

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
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

// PublicCertManager manages refreshing the public cert.
type PublicCertManager struct {
	commonName        string
	getPublicCertApi  cloudcommon.GetPublicCertApi
	cert              *tls.Certificate
	expiresAt         time.Time
	done              bool
	refreshTrigger    chan bool
	refreshThreshold  time.Duration
	refreshRetryDelay time.Duration
	mux               sync.Mutex
}

func NewPublicCertManager(commonName string, getPublicCertApi cloudcommon.GetPublicCertApi) *PublicCertManager {
	// Nominally letsencrypt certs are valid for 90 days
	// and they recommend refreshing at 30 days to expiration.
	mgr := &PublicCertManager{
		commonName:        commonName,
		getPublicCertApi:  getPublicCertApi,
		refreshTrigger:    make(chan bool, 1),
		refreshThreshold:  30 * 24 * time.Hour,
		refreshRetryDelay: 24 * time.Hour,
	}
	return mgr
}

func (s *PublicCertManager) updateCert(ctx context.Context) error {
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
