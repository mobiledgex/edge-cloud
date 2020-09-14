package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/access"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	ssh "github.com/mobiledgex/golang-ssh"
)

var CertName = "envoyTlsCerts" // cannot use common name as filename since envoy doesn't know if the app is dedicated or not

const LETS_ENCRYPT_MAX_DOMAINS_PER_CERT = 100

var SharedRootLbClient ssh.Client
var DedicatedClients map[string]ssh.Client
var DedicatedTls access.TLSCert
var DedicatedMux sync.Mutex

var DedicatedVmAppClients map[string]ssh.Client
var DedicatedVmAppMux sync.Mutex

var selfSignedCmd = `openssl req -new -newkey rsa:2048 -nodes -days 90 -nodes -x509 -config <(
cat <<-EOF
[req]
prompt = no
distinguished_name = dn

[ dn ]
CN = %s
EOF
)`

// Alt Names portion will look like:
// DNS.1 = test.com
// DNS.2 = matt.test.com
// ... going on for as many alternative names there are, and will be generated by getSelfSignedCerts
var selfSignedCmdWithSAN = `openssl req -new -newkey rsa:2048 -nodes -days 90 -nodes -x509 -config <(
cat <<-EOF
[req]
prompt = no
x509_extensions = v3_req
distinguished_name = dn

[ dn ]
CN = %s

[ v3_req ]
subjectAltName = @alt_names

[ alt_names ]
%s
EOF
)`

var privKeyStart = "-----BEGIN PRIVATE KEY-----"
var privKeyEnd = "-----END PRIVATE KEY-----"
var certStart = "-----BEGIN CERTIFICATE-----"
var certEnd = "-----END CERTIFICATE-----"

func init() {
	DedicatedClients = make(map[string]ssh.Client)
}

// GetCertsDirAndFiles returns certsDir, certFile, keyFile
func GetCertsDirAndFiles(ctx context.Context, client ssh.Client) (string, string, string, error) {
	out, err := client.Output("pwd")
	if err != nil {
		return "", "", "", err
	}
	pwd := strings.TrimSpace(string(out))
	certsDir := pwd + "/envoy/certs"
	certFile := certsDir + "/" + CertName + ".crt"
	keyFile := certsDir + "/" + CertName + ".key"
	return certsDir, certFile, keyFile, nil
}

// get certs from vault for rootlb, and pull a new one once a month, should only be called once by CRM
func GetRootLbCerts(ctx context.Context, commonName, dedicatedCommonName, vaultAddr string, client ssh.Client, commercialCerts bool) {
	certsDir, certFile, keyFile, err := GetCertsDirAndFiles(ctx, client)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Error: Unable to get cert info", "dedicatedCommonName", dedicatedCommonName, "err", err)
		return
	}
	SharedRootLbClient = client
	getRootLbCertsHelper(ctx, commonName, dedicatedCommonName, vaultAddr, certsDir, certFile, keyFile, commercialCerts)
	// refresh every 30 days
	for {
		select {
		case <-time.After(30 * 24 * time.Hour):
			getRootLbCertsHelper(ctx, commonName, dedicatedCommonName, vaultAddr, certsDir, certFile, keyFile, commercialCerts)
		}
	}
}

func getRootLbCertsHelper(ctx context.Context, commonName, dedicatedCommonName, vaultAddr string, certsDir, certFile, keyFile string, commercialCerts bool) {
	var err error
	var config *vault.Config
	tls := access.TLSCert{}
	if commercialCerts {
		config, err = vault.BestConfig(vaultAddr)
		if err == nil {
			err = getCertFromVault(ctx, config, &tls, commonName, dedicatedCommonName)
		}
	} else {
		err = getSelfSignedCerts(ctx, &tls, commonName, dedicatedCommonName)
	}
	if err == nil {
		writeCertToRootLb(ctx, &tls, SharedRootLbClient, certsDir, certFile, keyFile)
		// dedicated clusters
		DedicatedMux.Lock()
		DedicatedTls = tls
		for _, client := range DedicatedClients {
			writeCertToRootLb(ctx, &tls, client, certsDir, certFile, keyFile)
		}
		DedicatedMux.Unlock()

		// VM Apps
		DedicatedVmAppMux.Lock()
		for _, client := range DedicatedVmAppClients {
			writeCertToRootLb(ctx, &tls, client, certsDir, certFile, keyFile)
		}
		DedicatedVmAppMux.Unlock()
	} else {
		log.SpanLog(ctx, log.DebugLevelInfo, "Unable to get certs", "err", err)
	}
}

func writeCertToRootLb(ctx context.Context, tls *access.TLSCert, client ssh.Client, certsDir, certFile, keyFile string) {
	// write it to rootlb

	err := pc.Run(client, "mkdir -p "+certsDir)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "can't create cert dir on rootlb", "certDir", certsDir)
	} else {
		err = pc.WriteFile(client, certFile, tls.CertString, "tls cert", pc.NoSudo)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "unable to write tls cert file to rootlb", "err", err)
		}
		err = pc.WriteFile(client, keyFile, tls.KeyString, "tls key", pc.NoSudo)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "unable to write tls key file to rootlb", "err", err)
		}
	}
}

// GetCertFromVault fills in the cert fields by calling the vault  plugin.  The vault plugin will
// return a new cert if one is not already available, or a cached copy of an existing cert.
func getCertFromVault(ctx context.Context, config *vault.Config, tlsCert *access.TLSCert, commonNames ...string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCertFromVault", "commonName", commonNames)
	client, err := config.Login()
	if err != nil {
		return err
	}
	// needs to have at least one domain name specified, and not more than LetsEncrypt's limit per cert
	// in reality len(commonNames) should always be 2, one for the sharedLB and a wildcard one for the dedicatedLBs
	if len(commonNames) < 1 || LETS_ENCRYPT_MAX_DOMAINS_PER_CERT < len(commonNames) {
		return fmt.Errorf("must have between 1 and %d domain names specified", LETS_ENCRYPT_MAX_DOMAINS_PER_CERT)
	}
	names := strings.Join(commonNames, ",")
	// vault API uses "_" to denote wildcard
	path := "/certs/cert/" + strings.Replace(names, "*", "_", 1)
	result, err := vault.GetKV(client, path, 0)
	if err != nil {
		return err
	}
	var ok bool
	tlsCert.CertString, ok = result["cert"].(string)
	if !ok {
		return fmt.Errorf("No cert found in cert from vault")
	}
	tlsCert.KeyString, ok = result["key"].(string)
	if !ok {
		return fmt.Errorf("No key found in cert from vault")
	}
	ttlval, ok := result["ttl"].(json.Number)
	if !ok {
		return fmt.Errorf("ttl key found in cert from vault")
	}
	tlsCert.TTL, err = ttlval.Int64()
	if err != nil {
		return fmt.Errorf("Error in decoding TTL from vault: %v", err)
	}
	tlsCert.CommonName = names
	return nil
}

// Generates a self signed cert for testing purposes or if crm does not have access to vault
func getSelfSignedCerts(ctx context.Context, tlsCert *access.TLSCert, commonNames ...string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Generating self-signed cert", "commonNames", commonNames)
	var args string
	if len(commonNames) < 1 {
		return fmt.Errorf("Must have at least one domain name specified for cert generation")
	} else if len(commonNames) == 1 {
		args = fmt.Sprintf(selfSignedCmd, commonNames[0])
	} else {
		altNames := []string{}
		for i, name := range commonNames {
			altNames = append(altNames, fmt.Sprintf("DNS.%d = %s", i+1, name))
		}
		args = fmt.Sprintf(selfSignedCmdWithSAN, commonNames[0], strings.Join(altNames, "\n"))
	}
	cmd := exec.Command("bash", "-c", args)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error generating cert: %v\n", err)
	}
	output := string(out)

	// Get the private key
	start := strings.Index(output, privKeyStart)
	end := strings.Index(output, privKeyEnd)
	if start == -1 || end == -1 {
		return fmt.Errorf("Cert generation failed, could not find start or end of private key")
	}
	end = end + len(privKeyEnd)
	tlsCert.KeyString = output[start:end]

	// Get the cert
	start = strings.Index(output, certStart)
	end = strings.Index(output, certEnd)
	if start == -1 || end == -1 {
		return fmt.Errorf("Cert generation failed, could not find start or end of private key")
	}
	end = end + len(certEnd)
	tlsCert.CertString = output[start:end]
	return nil
}

func NewDedicatedCluster(ctx context.Context, clustername string, client ssh.Client) {
	certsDir, certFile, keyFile, err := GetCertsDirAndFiles(ctx, client)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Error: failed to get cert dir and files", "clustername", clustername, "err", err)
		return
	}
	DedicatedMux.Lock()
	defer DedicatedMux.Unlock()
	DedicatedClients[clustername] = client
	writeCertToRootLb(ctx, &DedicatedTls, client, certsDir, certFile, keyFile)
}

func RemoveDedicatedCluster(ctx context.Context, clustername string) {
	DedicatedMux.Lock()
	defer DedicatedMux.Unlock()
	delete(DedicatedClients, clustername)
}

func NewDedicatedVmApp(ctx context.Context, appname string, client ssh.Client) {
	certsDir, certFile, keyFile, err := GetCertsDirAndFiles(ctx, client)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Error: failed to get cert dir and files", "appname", appname, "err", err)
		return
	}
	DedicatedVmAppMux.Lock()
	defer DedicatedVmAppMux.Unlock()
	DedicatedVmAppClients[appname] = client
	writeCertToRootLb(ctx, &DedicatedTls, client, certsDir, certFile, keyFile)
}

func RemoveDedicatedVmApp(ctx context.Context, appname string) {
	DedicatedVmAppMux.Lock()
	defer DedicatedVmAppMux.Unlock()
	delete(DedicatedVmAppClients, appname)
}
