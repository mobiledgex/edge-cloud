package proxy

import (
	"context"
	"time"
	"encoding/json"
	"fmt"
	"sync"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/access"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	ssh "github.com/mobiledgex/golang-ssh"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

var CertsDir = "/etc/ssl/certs"
var CertName = "envoyTlsCerts" // cannot use common name as filename since envoy doesn't know if the app is dedicated or not
var certFile = CertsDir + "/" + CertName + ".crt"
var keyFile = CertsDir + "/" + CertName + ".key"

var SharedRootLbClient ssh.Client
var DedicatedClients map[string]ssh.Client
var DedicatedTls access.TLSCert
var DedicatedMux sync.Mutex


// get certs from vault for rootlb, and pull a new one once a month, should only be called once by CRM
func GetRootLbCerts(ctx context.Context, commonName, dedicatedCommonName, vaultAddr string, client ssh.Client) {
	SharedRootLbClient = client
	DedicatedClients = make(map[string]ssh.Client)
	getRootLbCertsHelper(ctx, commonName, dedicatedCommonName, vaultAddr)
	// refresh every 30 days
	for {
		select {
		case <-time.After(30*24*time.Hour):
			getRootLbCertsHelper(ctx, commonName, dedicatedCommonName, vaultAddr)			
		}
	}
}

func getRootLbCertsHelper(ctx context.Context, commonName, dedicatedCommonName, vaultAddr string) {
	config, err := vault.BestConfig(vaultAddr)
	if err == nil {
		// rootlb
		tls := access.TLSCert{}
		err = getCertFromVault(ctx, config, commonName, &tls)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "unable to pull certs from vault", "err", err)
		} else {
			writeCertToRootLb(ctx, &tls, SharedRootLbClient)
		}

		// dedicated lbs
		DedicatedMux.Lock()
		err = getCertFromVault(ctx, config, dedicatedCommonName, &DedicatedTls)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "unable to pull dedicated certs from vault", "err", err)
		} else {
			for _, client := range DedicatedClients {
				writeCertToRootLb(ctx, &DedicatedTls, client)
			}
		}
		DedicatedMux.Unlock()
	} else {
		log.SpanLog(ctx, log.DebugLevelInfo, "unable to get vault config", "err", err)
	}
}

func writeCertToRootLb(ctx context.Context, tls *access.TLSCert, client ssh.Client) {
	// write it to rootlb
	err := pc.Run(client, "mkdir -p "+CertsDir)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "can't create cert dir on rootlb", "certDir", CertsDir)
	} else {
		err = pc.WriteFile(client, certFile, tls.CertString, "tls cert", pc.SudoOn)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelMexos, "unable to write tls cert file to rootlb", "err", err)
		}
		err = pc.WriteFile(client, keyFile, tls.KeyString, "tls key", pc.SudoOn)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelMexos, "unable to write tls key file to rootlb", "err", err)
		}	
	}
}

// GetCertFromVault fills in the cert fields by calling the vault  plugin.  The vault plugin will 
// return a new cert if one is not already available, or a cached copy of an existing cert.
func getCertFromVault(ctx context.Context, config *vault.Config, commonName string, tlsCert *access.TLSCert) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "GetCertFromVault", "commonName", commonName)
	client, err := config.Login()
	if err != nil {
		return err
	}
	// vault API uses "_" to denote wildcard
	path := "/certs/cert/" + strings.Replace(commonName, "*", "_", 1)
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
	tlsCert.CommonName = commonName
	return nil
}

func NewDedicatedCluster(ctx context.Context, clustername string, client ssh.Client) {
	DedicatedMux.Lock()
	defer DedicatedMux.Unlock()
	DedicatedClients[clustername] = client
	writeCertToRootLb(ctx, &DedicatedTls, client)
}

func RemoveDedicatedCluster(ctx context.Context, clustername string) {
	DedicatedMux.Lock()
	defer DedicatedMux.Unlock()
	delete(DedicatedClients, clustername)
}
