package accessapi

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/chefmgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/cloudflaremgmt"
	pfutils "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/utils"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/vault"
)

// This is a global in order to cache it across all platforms in the Controller.
var cloudflareApi *cloudflare.API

// VaultClient implements platform.AccessApi for access from the Controller
// directly to Vault. In some cases it may require loading the platform
// specific plugin.
// VaultClient should only be used in the context of the Controller.
type VaultClient struct {
	cloudlet      *edgeproto.Cloudlet
	vaultConfig   *vault.Config
	region        string
	cloudflareApi *cloudflare.API
}

func NewVaultClient(cloudlet *edgeproto.Cloudlet, vaultConfig *vault.Config, region string) *VaultClient {
	return &VaultClient{
		cloudlet:    cloudlet,
		vaultConfig: vaultConfig,
		region:      region,
	}
}

func (s *VaultClient) GetCloudletAccessVars(ctx context.Context) (map[string]string, error) {
	// Platform-specific implementation.
	cloudletPlatform, err := pfutils.GetPlatform(ctx, s.cloudlet.PlatformType.String(), nil)
	if err != nil {
		return nil, err
	}
	return cloudletPlatform.GetAccessData(ctx, s.cloudlet, s.region, s.vaultConfig, GetCloudletAccessVars, nil)
}

func (s *VaultClient) GetRegistryAuth(ctx context.Context, imgUrl string) (*cloudcommon.RegistryAuth, error) {
	return cloudcommon.GetRegistryAuth(ctx, imgUrl, s.vaultConfig)
}

func (s *VaultClient) SignSSHKey(ctx context.Context, publicKey string) (string, error) {
	// Signed ssh keys should have a short valid time
	return vault.SignSSHKey(s.vaultConfig, publicKey)
}

func (s *VaultClient) GetSSHPublicKey(ctx context.Context) (string, error) {
	cmd := exec.Command("curl", "-s", fmt.Sprintf("%s/v1/ssh/public_key", s.vaultConfig.Addr))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get vault ssh cert: %s, %v", string(out), err)
	}
	if !strings.Contains(string(out), "ssh-rsa") {
		return "", fmt.Errorf("invalid vault ssh cert: %s", string(out))
	}
	return string(out), nil
}

func (s *VaultClient) GetOldSSHKey(ctx context.Context) (*vault.MEXKey, error) {
	// This is supported for upgrading old VMs only.
	vaultPath := "/secret/data/keys/id_rsa_mex"
	key := &vault.MEXKey{}
	err := vault.GetData(s.vaultConfig, vaultPath, 0, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get mex key for %s, %v", vaultPath, err)
	}
	return key, nil
}

func (s *VaultClient) GetChefAuthKey(ctx context.Context) (*chefmgmt.ChefAuthKey, error) {
	// TODO: maintain a Cloudlet-specific API key
	auth, err := chefmgmt.GetChefAuthKeys(ctx, s.vaultConfig)
	if err != nil {
		return nil, err
	}
	return auth, nil
}

func (s *VaultClient) getCloudflareApi() (*cloudflare.API, error) {
	if cloudflareApi != nil {
		return cloudflareApi, nil
	}
	vaultPath := "/secret/data/cloudlet/openstack/mexenv.json"
	vars, err := vault.GetEnvVars(s.vaultConfig, vaultPath)
	if err != nil {
		return nil, err
	}
	api, err := cloudflare.New(vars["MEX_CF_KEY"], vars["MEX_CF_USER"])
	if err != nil {
		return nil, err
	}
	cloudflareApi = api
	return cloudflareApi, nil
}

func (s *VaultClient) CreateOrUpdateDNSRecord(ctx context.Context, zone, name, rtype, content string, ttl int, proxy bool) error {
	api, err := s.getCloudflareApi()
	if err != nil {
		return err
	}
	// TODO: validate parameters are ok for this cloudlet
	return cloudflaremgmt.CreateOrUpdateDNSRecord(ctx, api, zone, name, rtype, content, ttl, proxy)
}

func (s *VaultClient) GetDNSRecords(ctx context.Context, zone, fqdn string) ([]cloudflare.DNSRecord, error) {
	api, err := s.getCloudflareApi()
	if err != nil {
		return nil, err
	}
	// TODO: validate parameters are ok for this cloudlet
	return cloudflaremgmt.GetDNSRecords(ctx, api, zone, fqdn)
}

func (s *VaultClient) DeleteDNSRecord(ctx context.Context, zone, recordID string) error {
	api, err := s.getCloudflareApi()
	if err != nil {
		return err
	}
	// TODO: validate parameters are ok for this cloudlet
	return cloudflaremgmt.DeleteDNSRecord(ctx, api, zone, recordID)
}

func (s *VaultClient) GetSessionTokens(ctx context.Context, arg []byte) (map[string]string, error) {
	// Platform-specific implementation
	cloudletPlatform, err := pfutils.GetPlatform(ctx, s.cloudlet.PlatformType.String(), nil)
	if err != nil {
		return nil, err
	}
	return cloudletPlatform.GetAccessData(ctx, s.cloudlet, s.region, s.vaultConfig, GetSessionTokens, arg)
}

func (s *VaultClient) GetPublicCert(ctx context.Context, commonName string) (*vault.PublicCert, error) {
	publicCert, err := vault.GetPublicCert(s.vaultConfig, commonName)
	if err != nil {
		return nil, err
	}
	return publicCert, nil
}

func (s *VaultClient) GetKafkaCreds(ctx context.Context) (*node.KafkaCreds, error) {
	path := node.GetKafkaVaultPath(s.region, s.cloudlet.Key.Name, s.cloudlet.Key.Organization)
	creds := node.KafkaCreds{}
	err := vault.GetData(s.vaultConfig, path, 0, &creds)
	if err != nil {
		return nil, fmt.Errorf("failed to get kafka credentials at %s, %v", path, err)
	}
	return &creds, nil

}
