package proxy

import (
	"context"
	"time"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/access"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

var CertsDir = "etc/ssl/certs"

// get certs from vault for rootlb, and pull a new one once a month
// TODO: put this on dedicated lbs as well
func GetRootLbCerts(ctx context.Context, commonName, vaultAddr string, client pc.PlatformClient) {
	getRootLbCertsHelper(ctx, commonName, vaultAddr, client)
	// refresh every 30 days
	for {
		select {
		case <-time.After(30*24*time.Hour):
			getRootLbCertsHelper(ctx, commonName, vaultAddr, client)			
		}
	}
}

func getRootLbCertsHelper(ctx context.Context, commonName, vaultAddr string, client pc.PlatformClient) {
	config, err := vault.BestConfig(vaultAddr)
	if err == nil {
		tls := access.TLSCert{}
		err = GetCertFromVault(ctx, config, commonName, &tls)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "unable to pull certs from vault", "err", err)
		} else {
			// write it to rootlb
			err = pc.Run(client, "mkdir -p "+CertsDir)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelMexos, "can't create cert dir on rootlb", "certDir", CertsDir)
			} else {
				certFile := CertsDir + "/" + tls.CommonName + ".crt"
				err = pc.WriteFile(client, certFile, tls.CertString, "tls cert", pc.NoSudo)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelMexos, "unable to write tls cert file to rootlb", "err", err)
				}
				keyFile := CertsDir + "/" + tls.CommonName + ".key"
				err = pc.WriteFile(client, keyFile, tls.KeyString, "tls key", pc.NoSudo)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelMexos, "unable to write tls key file to rootlb", "err", err)
				}	
			}
		}
	}
}

// GetCertFromVault fills in the cert fields by calling the vault  plugin.  The vault plugin will 
// return a new cert if one is not already available, or a cached copy of an existing cert.
func GetCertFromVault(ctx context.Context, config *vault.Config, commonName string, tlsCert *access.TLSCert) error {
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