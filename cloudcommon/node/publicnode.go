package node

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
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

// GetPublicCertApi abstracts the way the public cert is retrieved.
// Certain services, like DME running on a Cloudlet, may need to connect
// to the controller to get a public cert from Vault.
type GetPublicCertApi interface {
	GetPublicCert(ctx context.Context, commonName string) (*vault.PublicCert, error)
}

// VaultPublicCertApi implements GetPublicCertApi by connecting directly to Vault.
type VaultPublicCertApi struct {
	VaultConfig *vault.Config
}

func (s *VaultPublicCertApi) GetPublicCert(ctx context.Context, commonName string) (*vault.PublicCert, error) {
	return vault.GetPublicCert(s.VaultConfig, commonName)
}

// PublicCertManager manages refreshing the public cert.
type PublicCertManager struct {
	commonName        string
	getPublicCertApi  GetPublicCertApi
	cert              *tls.Certificate
	expiresAt         time.Time
	done              bool
	refreshTrigger    chan bool
	refreshThreshold  time.Duration
	refreshRetryDelay time.Duration
	mux               sync.Mutex
}

func NewPublicCertManager(commonName string, getPublicCertApi GetPublicCertApi) *PublicCertManager {
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
		GetCertificate: s.getCertificateFunc(),
	}
	return config, nil
}

func (s *PublicCertManager) getCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
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
			if s.done {
				break
			}
			span := log.StartSpan(log.DebugLevelInfo, "refresh public cert")
			ctx := log.ContextWithSpan(context.Background(), span)
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

// TestPublicCertApi implements GetPublicCertApi for unit/e2e testing
type TestPublicCertApi struct {
	GetCount int
}

func (s *TestPublicCertApi) GetPublicCert(ctx context.Context, commonName string) (*vault.PublicCert, error) {
	cert := &vault.PublicCert{}
	cert.Cert = `-----BEGIN CERTIFICATE-----
MIIEqDCCApCgAwIBAgIRAL+MCA1gi9MMrjbWG86n6zgwDQYJKoZIhvcNAQELBQAw
ETEPMA0GA1UEAxMGbWV4LWNhMB4XDTIwMDIyMDAzNTYzNloXDTIxMDgyMDAzNTQ1
N1owFTETMBEGA1UEAxMKbWV4LXNlcnZlcjCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAKv53gpDQwDex/WxWBQ9ptx4Ul5vGh57uqwbIyrCKapKWnsITZnr
43gdubQpMqJ0P+XYgLP98fGjRL5IEvl0CSPykuorhsawK2xaotTNIpGbhhJf6DeR
gUXv9h5TCwZUlaUew0jPWI/6oTk7VXh/L2ibA1M0ChzGgqp27U77deorSKCXZEQK
Fim8SoRq0le8898CDdp3emXT1AHb+MmsyBEGy8jD3ifQ6esWkPccryMKST7NK/E8
CbE/Nfr3RDmrjRgns3GkhlQoupkhipzdVJHEj6vgroVeLQoLfG6Z7phnomKfWn6d
RwTnZlr0vW9HMgRML/Wu6OnRWIN/bgZnhbECAwEAAaOB9jCB8zAOBgNVHQ8BAf8E
BAMCA7gwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMB0GA1UdDgQWBBSF
VwQvVmM8BHZROzRu7GCBxFtccjAfBgNVHSMEGDAWgBTAH57N0Em+LYPFBoYx08ge
dW3nBTCBgQYDVR0RBHoweIIJbG9jYWxob3N0ggpjb250cm9sbGVyghAqLm1vYmls
ZWRnZXgubmV0ghQqLmRtZS5tb2JpbGVkZ2V4Lm5ldIIUKi5jcm0ubW9iaWxlZGdl
eC5uZXSCFSouY3RybC5tb2JpbGVkZ2V4Lm5ldIcEfwAAAYcEAAAAADANBgkqhkiG
9w0BAQsFAAOCAgEABv4Vh81XvL85JxC9ji7Q56ufp5inUfPD+90XtpmUIk6LupNw
lhREu8YQ5DXJKF2XiTENmIzZop+hiKENSVRuwuCf+UMix9EWzmp+ztRt8g2R+0Vx
XfR8A5rq6bO0Dqtoe4lGkO2zbS09qOZACf+2vsYo9mZmzTYx0Ze+o3cySTdx6zry
piBn811jLHC/7DsSBGhM5PWDF5hUuw1BYeaPLOeOe5ijzeiiB/d3RwNAOddTmsAJ
CV7oI6+trb8BAf8xXMScxP0jiLRWD+RqVO4RR58vlgL1hqZZXnqPpb5gu3XmILJs
SeNmUHX0oslJgWEiX557bQzzoYVSDOH9j2BCx1uLDBVM1QrfhF6Ei0U90/payTvU
uNCKsVozBrzvA9hv8fqR5urzO287czqqztVWcmRDoBra/Sp74TTO/IOsiRR6u9QL
EfBOeiNrUCHEqqBwF6dUknN1X+IT5jDavk/Eo5nz1IMfhvBujzc5B3mqrXM/kX9A
dxCd6+mhGPz66wUek5JwF4hGmqzZIL69avU+Qf1wmnyH/DSjJv05fdXLEj2/jAWF
yHjOcuMT/jtR7Q3v8OHeVlMz9S87RMzj+K57lfvpKoTUB4AJt/YNwUSeUg30v9lr
bQUw2rtzxIw7uj1Qi2QDqorqPjawrCDuCNkqopx4D+/fbyRM5GrLIjAziMU=
-----END CERTIFICATE-----`
	cert.Key = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAq/neCkNDAN7H9bFYFD2m3HhSXm8aHnu6rBsjKsIpqkpaewhN
mevjeB25tCkyonQ/5diAs/3x8aNEvkgS+XQJI/KS6iuGxrArbFqi1M0ikZuGEl/o
N5GBRe/2HlMLBlSVpR7DSM9Yj/qhOTtVeH8vaJsDUzQKHMaCqnbtTvt16itIoJdk
RAoWKbxKhGrSV7zz3wIN2nd6ZdPUAdv4yazIEQbLyMPeJ9Dp6xaQ9xyvIwpJPs0r
8TwJsT81+vdEOauNGCezcaSGVCi6mSGKnN1UkcSPq+CuhV4tCgt8bpnumGeiYp9a
fp1HBOdmWvS9b0cyBEwv9a7o6dFYg39uBmeFsQIDAQABAoIBAHwrikVQsVU0hZ4B
MT5UEWGIQrjKcUpnPa48XdTmohzBWLkSkq07I1873zSUtmmTk/tJqgvLpGA66UyW
T5TrUhoxcCBB0yssUf4HJyCNCJOnflNQCiPtHDC6BLN6dDBa7D1vi8LLav9yD+x5
ycmZ00os+mad4VtLfVbFTazEZSvwa5e3UphPmE6f7J2VLmh6sUFBM/CSjFmsSLZE
UbEVmmYLittgKDEFEAoLqTL7X2aQ+4dQSUSK/lqv2f9hrsNXmTxfKJ+GaresZBRY
JCDoLxn6hQuC9ykfokPxlrNQomWc7EG8I89vMHNCmJnSHzHj94csTybGCh/jrWWB
9g6lOBECgYEA1GlPY/le/3TBM2HNkSpc9q1QWA84qQNRHiSguCQse8npfy5wfAJb
OVOZoXgt+0LIJRc9z+2gxCs1EPeeg7Ql6BmcloK6q+Q5Dysg/ieSchpktVdxFRW2
urly/lqLcVeTsaBZixIjD2oX4iM/YQs1rhioRiRZei2gGzovqpiS0A0CgYEAz0Ra
PKNGEtZp9qO0nCkZtqLSc3C+JnJJ059CO1TjpW67nBrC2u5XIiCWMwbINpQkSbAp
C8y0GrynQrwegvtO/RkHcTrEpDCWwZR9AreQfaScN+9dVjAx4hmx464qpRpEuaza
HTOUWp3L9nn3oQXxXGNgk99h9+Gaq29OAIb8fzUCgYEAocxqFgxBSbOk6z/Ht5ke
YSSZu8o0bcHCC4T5C+s6Gz0taJx2QHAHDv7YWr/RvsAa9u3iPr2SpXsIHBmSnF4g
NdE0jw2bpg3dTOmcYxy/l7z1E5E86UO2Ajv7FTbhWv/L2BT9wEqbfEVjVfVldMV2
KVxM6ckMg123xKWo43j+9A0CgYAsubU3LIxseDQ5cq4AnKXd0VjUbFm79iGUNuOV
5gWRp0l4sBWoJJJM3PdMX4RIssL527efwjaDJn55WhrDbPNojkQa3PGd9JYzg5VO
Rso5MpI7R72+YXwCLEVEukqdggOehXwznPPAchiXQU58QsoIg7FNd4CuetJjeAs+
9eH6mQKBgQCVgx3Icz/P3xo9bLNW4XFTP5fm+yc0+7+OfwyojgxLeQM1qPOi1zBb
AJWS6w5h8/opvpDW5KSCZEL1nIkmfUFRj+aNWTftQborGCalHnxVmPNBChMe/spE
oQBpOxj1Ezimg+uk7TZ2BZUP6JCB05TpH2g6sZwhTXTXiajwUFVq1g==
-----END RSA PRIVATE KEY-----`
	// 24 hours in seconds
	cert.TTL = 24 * 3600
	s.GetCount++
	return cert, nil
}
