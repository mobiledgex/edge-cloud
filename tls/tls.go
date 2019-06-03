package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

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

// helper function to get the cert pool
func getClientCertPool(tlsCertFile string) (*x509.CertPool, error) {
	if tlsCertFile == "" {
		return nil, nil
	}
	dir := path.Dir(tlsCertFile)
	certPool := x509.NewCertPool()
	bs, err := ioutil.ReadFile(dir + "/mex-ca.crt")
	if err != nil {
		return nil, err
	}
	ok := certPool.AppendCertsFromPEM(bs)
	if !ok {
		return nil, fmt.Errorf("fail to append certs")
	}
	return certPool, nil
}

// GetTLSClientDialOption gets options needed for TLS connection
func GetTLSClientDialOption(addr string, tlsCertFile string) (grpc.DialOption, error) {
	config, err := GetMutualAuthClientConfig(addr, tlsCertFile)
	if err != nil {
		return nil, err
	}
	return GetGrpcDialOption(config), nil
}

func GetMutualAuthClientConfig(addr string, tlsCertFile string) (*tls.Config, error) {
	if tlsCertFile == "" {
		// no TLS
		return nil, nil
	}

	certPool, err := getClientCertPool(tlsCertFile)
	if err != nil {
		return nil, err
	}
	certificate, err := getClientCertificate(tlsCertFile)
	if err != nil {
		return nil, err
	}
	serverName := strings.Split(addr, ":")[0]

	return &tls.Config{
		ServerName:   serverName,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	}, nil
}

func GetGrpcDialOption(config *tls.Config) grpc.DialOption {
	if config == nil {
		// no TLS
		return grpc.WithInsecure()
	}
	transportCreds := credentials.NewTLS(config)
	return grpc.WithTransportCredentials(transportCreds)
}

// GetTLSClientConfig gets TLS Config for REST api connection
func GetTLSClientConfig(addr string, tlsCertFile string) (*tls.Config, error) {
	if tlsCertFile != "" {
		certPool, err := getClientCertPool(tlsCertFile)
		if err != nil {
			return nil, err
		}
		certificate, err := getClientCertificate(tlsCertFile)
		if err != nil {
			return nil, err
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		}
		tlsConfig.BuildNameToCertificate()
		return tlsConfig, nil
	}
	///no TLS - TODO NEED TO have non-secure connection
	return nil, nil
}

// GetTLSServerCreds gets grpc credentials for the server for
// mutual authentication.
// Returns nil credentials is the TLS cert file name is blank
func GetTLSServerCreds(tlsCertFile string) (credentials.TransportCredentials, error) {
	if tlsCertFile == "" {
		fmt.Printf("no server TLS credentials\n")
		return nil, nil
	}

	dir := path.Dir(tlsCertFile)
	caFile := dir + "/" + "mex-ca.crt"
	keyFile := strings.Replace(tlsCertFile, "crt", "key", 1)
	fmt.Printf("Loading certfile %s cafile %s keyfile %s\n", tlsCertFile, caFile, keyFile)

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	cabs, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("could not read CA certificate: %s", err)
	}
	ok := certPool.AppendCertsFromPEM(cabs)
	if !ok {
		return nil, fmt.Errorf("fail to append cert CA %s", caFile)
	}

	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(tlsCertFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("could not load server key pair: %s", err)
	}

	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		//ClientAuth:   tls.RequireAnyClientCert,
		//ClientAuth: tls.VerifyClientCertIfGiven,

		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})
	return creds, nil
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
