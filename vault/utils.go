package vault

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func GetVaultData(keyURL string) (map[string]interface{}, error) {
	roleID := os.Getenv("VAULT_ROLE_ID")
	secretID := os.Getenv("VAULT_SECRET_ID")

	if roleID == "" {
		return nil, fmt.Errorf("VAULT_ROLE_ID env var missing")
	}
	if secretID == "" {
		return nil, fmt.Errorf("VAULT_SECRET_ID env var missing")
	}
	if !strings.Contains(keyURL, "://") {
		keyURL = "https://" + keyURL
	}
	uri, err := url.ParseRequestURI(keyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid keypath %s, %v", keyURL, err)
	}
	addr := uri.Scheme + "://" + uri.Host
	client, err := NewClient(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to set up Vault client for %s, %v", addr, err)
	}
	err = AppRoleLogin(client, roleID, secretID)
	if err != nil {
		return nil, fmt.Errorf("failed to login to Vault, %v", err)
	}
	path := strings.TrimPrefix(uri.Path, "/v1")
	data, err := GetKV(client, path, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get values for %s from Vault, %v", path, err)
	}
	return data, nil
}
