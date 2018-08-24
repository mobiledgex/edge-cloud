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

// GetTLSClientDialOption gets options needed for TLS connection
func GetTLSClientDialOption(addr string, tlsCertFile string) (grpc.DialOption, error) {
	if tlsCertFile != "" {
		dir := path.Dir(tlsCertFile)
		keyFile := strings.Replace(tlsCertFile, "crt", "key", 1)

		certPool := x509.NewCertPool()
		bs, err := ioutil.ReadFile(dir + "/mex-ca.crt")
		if err != nil {
			return nil, err
		}
		ok := certPool.AppendCertsFromPEM(bs)
		if !ok {
			return nil, fmt.Errorf("fail to append certs")
		}
		certificate, err := tls.LoadX509KeyPair(
			tlsCertFile,
			keyFile,
		)
		if err != nil {
			return nil, err
		}
		serverName := strings.Split(addr, ":")[0]

		fmt.Printf("")
		transportCreds := credentials.NewTLS(&tls.Config{
			ServerName:   serverName,
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})
		return grpc.WithTransportCredentials(transportCreds), nil

	}
	///no TLS
	return grpc.WithInsecure(), nil
}

// GetTLSServerCreds gets options needed for TLS connection.
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
