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
func GetTLSClientDialOption(serverAddr string, tlsCertFile string, skipVerify bool) (grpc.DialOption, error) {
	config, err := GetTLSClientConfig(serverAddr, tlsCertFile, "", skipVerify)
	if err != nil {
		return nil, err
	}
	return GetGrpcDialOption(config), nil
}

// GetTLSClientConfig builds client side TLS configuration.  If the serverAddr
// is blank, no validation is done on the cert.  CaCertFile is specified when communicating to
// exernal servers with their own privately signed certs.  Leave this blank to use the mex-ca.crt.
// Skipverify is only to be used for internal connections such as GRPCGW to GRPC.
func GetTLSClientConfig(serverAddr string, tlsCertFile string, caCertFile string, skipVerify bool) (*tls.Config, error) {
	if tlsCertFile == "" {
		return nil, nil
	}
	certPool, err := GetClientCertPool(tlsCertFile, caCertFile)
	if err != nil {
		return nil, err
	}
	certificate, err := getClientCertificate(tlsCertFile)
	if err != nil {
		return nil, err
	}

	if serverAddr != "" {
		serverName := strings.Split(serverAddr, ":")[0]
		return &tls.Config{
			ServerName:         serverName,
			InsecureSkipVerify: skipVerify,
			Certificates:       []tls.Certificate{certificate},
			RootCAs:            certPool,
		}, nil
	} else {
		// do not validate the server address.
		return &tls.Config{
			InsecureSkipVerify: skipVerify,
			Certificates:       []tls.Certificate{certificate},
			RootCAs:            certPool,
		}, nil
	}
}

func GetGrpcDialOption(config *tls.Config) grpc.DialOption {

	if config == nil {
		// no TLS
		return grpc.WithInsecure()
	}
	transportCreds := credentials.NewTLS(config)
	return grpc.WithTransportCredentials(transportCreds)
}

// GetTLSServerCreds gets grpc credentials for the server for
// mutual authentication.
// Returns nil credentials is the TLS cert file name is blank
func GetTLSServerCreds(tlsCertFile string, mutualAuth bool) (credentials.TransportCredentials, error) {

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
		if mutualAuth {
			return nil, fmt.Errorf("could not read CA certificate: %s", err)
		}
		// this is not fatal if we are not trying to do mutual auth
		fmt.Printf("no mex-ca.crt file in dir: %s\n", dir)
	}
	if len(cabs) > 0 {
		ok := certPool.AppendCertsFromPEM(cabs)
		if !ok {
			return nil, fmt.Errorf("fail to append cert CA %s", caFile)
		}
	}

	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(tlsCertFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("could not load server key pair: %s", err)
	}

	// Create the TLS credentials
	var creds credentials.TransportCredentials

	if mutualAuth {
		creds = credentials.NewTLS(&tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: []tls.Certificate{certificate},
			ClientCAs:    certPool})
	} else {
		creds = credentials.NewTLS(&tls.Config{
			ClientAuth:   tls.NoClientCert,
			Certificates: []tls.Certificate{certificate},
			ClientCAs:    certPool})
	}
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
