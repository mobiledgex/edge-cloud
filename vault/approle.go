package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
)

func AppRoleLogin(client *api.Client, roleID, secretID string) error {
	data := map[string]interface{}{
		"role_id":   roleID,
		"secret_id": secretID,
	}
	resp, err := client.Logical().Write("auth/approle/login", data)
	if err != nil {
		return err
	}
	if resp.Auth == nil {
		return fmt.Errorf("no auth info returned")
	}
	client.SetToken(resp.Auth.ClientToken)
	return nil
}

func NewClient(addr string) (*api.Client, error) {
	client, err := api.NewClient(nil)
	if err != nil {
		return nil, err
	}
	err = client.SetAddress(addr)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetData(addr, roleID, secretID, path string, version int, data interface{}) error {
	client, err := NewClient(addr)
	if err != nil {
		return err
	}
	err = AppRoleLogin(client, roleID, secretID)
	if err != nil {
		return err
	}
	vdat, err := GetKV(client, path, version)
	if err != nil {
		return err
	}
	return mapstructure.WeakDecode(vdat["data"], data)
}
