package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type TLSMode int

const (
	NoTLS TLSMode = iota
	ServerAuthTLS
	MutualAuthTLS
)

// Utility function that checks for E2ETEST_TLS env var
func IsTestTls() bool {
	if e2e := os.Getenv("E2ETEST_TLS"); e2e != "" {
		return true
	}
	return false
}

// helper function to get the cert pool
func getClientCertificate(tlsCertFile string) (tls.Certificate, error) {
	if tlsCertFile == "" {
		return tls.Certificate{}, nil
	}
	keyFile := strings.Replace(tlsCertFile, "crt", "key", 1)
	certificate, err := tls.LoadX509KeyPair(
		tlsCertFile,
		keyFile,
	)
	if err != nil {
		return tls.Certificate{}, err
	}
	return certificate, nil
}

// GetClientCertPool gets the system cert pool for all the trusted CA certs
// and then appends the caCertFile.  Leave caCertFile blank to use the mex-ca.crt
func GetClientCertPool(tlsCertFile string, caCertFile string) (*x509.CertPool, error) {
	if tlsCertFile == "" {
		return nil, nil
	}
	dir := path.Dir(tlsCertFile)
	// get the public certs from the host's cert pool
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("fail to load system cert pool")
	}
	// use mex-ca if no CA file specified
	if caCertFile == "" {
		caCertFile = dir + "/mex-ca.crt"
	}
	// append the CA file to the system cert pool
	bs, err := ioutil.ReadFile(caCertFile)
	if err == nil {
		ok := certPool.AppendCertsFromPEM(bs)
		if !ok {
			return nil, fmt.Errorf("fail to append certs")
		}
	}
	return certPool, nil
}

// GetTLSClientDialOption gets GRPC options needed for TLS connection
func GetTLSClientDialOption(tlsMode TLSMode, serverAddr string, getCertFunc func(*tls.CertificateRequestInfo) (*tls.Certificate, error), tlsCertFile string, skipVerify bool) (grpc.DialOption, error) {
	config, err := GetTLSClientConfig(tlsMode, serverAddr, getCertFunc, tlsCertFile, "", skipVerify)
	if err != nil {
		return nil, err
	}
	return GetGrpcDialOption(config), nil
}

// GetTLSClientConfig builds client side TLS configuration.  If the serverAddr
// is blank, no validation is done on the cert.  CaCertFile is specified when communicating to
// exernal servers with their own privately signed certs.  Leave this blank to use the mex-ca.crt.
// Requires either a tlsCertFile or a getCertFunc if mutualAuth
func GetTLSClientConfig(tlsMode TLSMode, serverAddr string, getCertFunc func(*tls.CertificateRequestInfo) (*tls.Certificate, error), tlsCertFile string, caCertFile string, skipVerify bool) (*tls.Config, error) {
	if tlsMode == NoTLS {
		return nil, nil
	}
	tlscfg := &tls.Config{}
	// Skip verification of self signed server certs if e2e tests
	if IsTestTls() || skipVerify {
		tlscfg.InsecureSkipVerify = true
	}
	if serverAddr != "" {
		serverName := strings.Split(serverAddr, ":")[0]
		tlscfg.ServerName = serverName
	}
	if tlsMode != MutualAuthTLS {
		return tlscfg, nil
	}
	// mTLS requires either Certificates or GetClientCertificate to be set for clients
	if getCertFunc != nil {
		tlscfg.GetClientCertificate = getCertFunc
	} else if tlsCertFile != "" {
		certPool, err := GetClientCertPool(tlsCertFile, caCertFile)
		if err != nil {
			return nil, err
		}
		certificate, err := getClientCertificate(tlsCertFile)
		if err != nil {
			return nil, err
		}
		tlscfg.RootCAs = certPool
		tlscfg.Certificates = []tls.Certificate{certificate}
	} else {
		return nil, fmt.Errorf("mTLS requires a certFile or getCertFunc")
	}

	return tlscfg, nil
}

func GetGrpcDialOption(config *tls.Config) grpc.DialOption {
	if config == nil {
		// no TLS
		return grpc.WithInsecure()
	}
	transportCreds := credentials.NewTLS(config)
	return grpc.WithTransportCredentials(transportCreds)
}

// ServerAuthServerCreds gets grpc credentials for the server for
// server-side authentication.
func ServerAuthServerCreds(tlsCertFile, tlsKeyFile string) (credentials.TransportCredentials, error) {
	if tlsCertFile == "" || tlsKeyFile == "" {
		fmt.Printf("no server TLS credentials\n")
		return nil, nil
	}
	return credentials.NewServerTLSFromFile(tlsCertFile, tlsKeyFile)
}
