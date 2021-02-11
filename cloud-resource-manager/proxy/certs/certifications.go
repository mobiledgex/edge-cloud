package certs

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/access"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/accessapi"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

const LETS_ENCRYPT_MAX_DOMAINS_PER_CERT = 100

var SharedRootLbClient ssh.Client
var DedicatedClients map[string]ssh.Client
var DedicatedTls access.TLSCert
var DedicatedMux sync.Mutex

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

var sudoType = pc.SudoOn
var noSudoMap = map[string]int{
	"fake":      1,
	"fakeinfra": 1,
	"edgebox":   1,
	"dind":      1,
}
var fixedCerts = false

var AtomicCertsUpdater = "/usr/local/bin/atomic-certs-update.sh"

var accessApi *accessapi.ControllerClient

func Init(ctx context.Context, clients map[string]ssh.Client, inAccessApi *accessapi.ControllerClient) {
	accessApi = inAccessApi
	if len(DedicatedClients) == 0 {
		DedicatedClients = make(map[string]ssh.Client)
	}
	if clients == nil {
		return
	}
	DedicatedMux.Lock()
	defer DedicatedMux.Unlock()
	for k, v := range clients {
		DedicatedClients[k] = v
	}
}

// get certs from vault for rootlb, and pull a new one once a month, should only be called once by CRM
func GetRootLbCerts(ctx context.Context, key *edgeproto.CloudletKey, commonName, dedicatedCommonName string, nodeMgr *node.NodeMgr, platformType string, client ssh.Client, commercialCerts bool) {
	_, found := noSudoMap[platformType]
	if found {
		sudoType = pc.NoSudo
	}
	if strings.Contains(platformType, "fake") {
		fixedCerts = true
	}
	out, err := client.Output("pwd")
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Error: Unable to get pwd", "dedicatedCommonName", dedicatedCommonName, "err", err)
		return
	}
	certsDir, certFile, keyFile := cloudcommon.GetCertsDirAndFiles(string(out))
	SharedRootLbClient = client
	getRootLbCertsHelper(ctx, key, commonName, dedicatedCommonName, nodeMgr, certsDir, certFile, keyFile, commercialCerts)
	// refresh every 30 days
	for {
		select {
		case <-time.After(30 * 24 * time.Hour):
			getRootLbCertsHelper(ctx, key, commonName, dedicatedCommonName, nodeMgr, certsDir, certFile, keyFile, commercialCerts)
		}
	}
}

func getRootLbCertsHelper(ctx context.Context, key *edgeproto.CloudletKey, commonName, dedicatedCommonName string, nodeMgr *node.NodeMgr, certsDir, certFile, keyFile string, commercialCerts bool) {
	var err error
	tls := access.TLSCert{}
	if commercialCerts {
		err = getCertFromVault(ctx, &tls, commonName, dedicatedCommonName)
	} else {
		err = getSelfSignedCerts(ctx, &tls, commonName, dedicatedCommonName)
	}
	if err == nil {
		err = writeCertToRootLb(ctx, &tls, SharedRootLbClient, certsDir, certFile, keyFile)
		if err != nil {
			nodeMgr.Event(ctx, "TLS certs error", key.Organization, key.GetTags(), err, "rootlb", commonName)
		}
		// dedicated LBs
		DedicatedMux.Lock()
		DedicatedTls = tls
		for lbName, client := range DedicatedClients {
			if client == nil {
				nodeMgr.Event(ctx, "TLS certs error", key.Organization, key.GetTags(), fmt.Errorf("missing client"), "rootlb", lbName)
				continue
			}
			err = writeCertToRootLb(ctx, &tls, client, certsDir, certFile, keyFile)
			if err != nil {
				nodeMgr.Event(ctx, "TLS certs error", key.Organization, key.GetTags(), err, "rootlb", lbName)
			}
		}
		DedicatedMux.Unlock()
	} else {
		log.SpanLog(ctx, log.DebugLevelInfo, "Unable to get certs", "err", err)
		nodeMgr.Event(ctx, "TLS certs error", key.Organization, key.GetTags(), fmt.Errorf("Unable to get certs: %v", err))
	}
}

func writeCertToRootLb(ctx context.Context, tls *access.TLSCert, client ssh.Client, certsDir, certFile, keyFile string) error {
	// write it to rootlb
	err := pc.Run(client, "mkdir -p "+certsDir)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "can't create cert dir on rootlb", "certDir", certsDir)
		return fmt.Errorf("failed to create cert dir on rootlb: %s, %v", certsDir, err)
	} else {
		if fixedCerts {
			// For testing, avoid atomic certs update as it will create timestamp based directories
			err = pc.WriteFile(client, certFile, tls.CertString, "tls cert", sudoType)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "unable to write tls cert file to rootlb", "err", err)
				return fmt.Errorf("failed to write tls cert file to rootlb, %v", err)
			}
			err = pc.WriteFile(client, keyFile, tls.KeyString, "tls key", sudoType)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "unable to write tls key file to rootlb", "err", err)
				return fmt.Errorf("failed to write tls cert file to rootlb, %v", err)
			}
			return nil
		}
		certsScript, err := ioutil.ReadFile(AtomicCertsUpdater)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "failed to read atomic certs updater script", "err", err)
			return fmt.Errorf("failed to read atomic certs updater script: %v", err)
		}
		err = pc.WriteFile(client, AtomicCertsUpdater, string(certsScript), "atomic-certs-updater", sudoType)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "failed to copy atomic certs updater script", "err", err)
			return fmt.Errorf("failed to copy atomic certs updater script: %v", err)
		}
		err = pc.WriteFile(client, certFile+".new", tls.CertString, "tls cert", sudoType)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "unable to write tls cert file to rootlb", "err", err)
			return fmt.Errorf("failed to write tls cert file to rootlb, %v", err)
		}
		err = pc.WriteFile(client, keyFile+".new", tls.KeyString, "tls key", sudoType)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "unable to write tls key file to rootlb", "err", err)
			return fmt.Errorf("failed to write tls cert file to rootlb, %v", err)
		}
		err = pc.Run(client, fmt.Sprintf("bash %s -d %s -c %s -k %s -e %s", AtomicCertsUpdater, certsDir, filepath.Base(certFile), filepath.Base(keyFile), cloudcommon.EnvoyImageDigest))
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "unable to write tls cert file to rootlb", "err", err)
			return fmt.Errorf("failed to atomically update tls certs: %v", err)
		}
	}
	return nil
}

// GetCertFromVault fills in the cert fields by calling the vault  plugin.  The vault plugin will
// return a new cert if one is not already available, or a cached copy of an existing cert.
func getCertFromVault(ctx context.Context, tlsCert *access.TLSCert, commonNames ...string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCertFromVault", "commonName", commonNames)
	// needs to have at least one domain name specified, and not more than LetsEncrypt's limit per cert
	// in reality len(commonNames) should always be 2, one for the sharedLB and a wildcard one for the dedicatedLBs
	if len(commonNames) < 1 || LETS_ENCRYPT_MAX_DOMAINS_PER_CERT < len(commonNames) {
		return fmt.Errorf("must have between 1 and %d domain names specified", LETS_ENCRYPT_MAX_DOMAINS_PER_CERT)
	}
	names := strings.Join(commonNames, ",")
	if accessApi == nil {
		return fmt.Errorf("Access API is not initialized")
	}
	// vault API uses "_" to denote wildcard
	commonName := strings.Replace(names, "*", "_", 1)
	pubCert, err := accessApi.GetPublicCert(ctx, commonName)
	if err != nil {
		return fmt.Errorf("Failed to get public cert from vault for commonName %s: %v", commonName, err)
	}
	if pubCert.Cert == "" {
		return fmt.Errorf("No cert found in cert from vault")
	}
	if pubCert.Key == "" {
		return fmt.Errorf("No key found in cert from vault")
	}

	tlsCert.CertString = pubCert.Cert
	tlsCert.KeyString = pubCert.Key
	tlsCert.TTL = pubCert.TTL
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

func NewDedicatedLB(ctx context.Context, key *edgeproto.CloudletKey, name string, client ssh.Client, nodeMgr *node.NodeMgr) {
	out, err := client.Output("pwd")
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Error: Unable to get pwd", "name", name, "err", err)
		return
	}
	certsDir, certFile, keyFile := cloudcommon.GetCertsDirAndFiles(string(out))
	DedicatedMux.Lock()
	defer DedicatedMux.Unlock()
	DedicatedClients[name] = client
	err = writeCertToRootLb(ctx, &DedicatedTls, client, certsDir, certFile, keyFile)
	if err != nil {
		nodeMgr.Event(ctx, "TLS certs error", key.Organization, key.GetTags(), err, "rootlb", name)
	}
}

func RemoveDedicatedLB(ctx context.Context, name string) {
	DedicatedMux.Lock()
	defer DedicatedMux.Unlock()
	delete(DedicatedClients, name)
}