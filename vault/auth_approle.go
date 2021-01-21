package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

type AppRoleAuth struct {
	roleID   string
	secretID string
}

func NewAppRoleAuth(roleID, secretID string) *AppRoleAuth {
	auth := AppRoleAuth{
		roleID:   roleID,
		secretID: secretID,
	}
	return &auth
}

func (s *AppRoleAuth) Login(client *api.Client) error {
	data := map[string]interface{}{
		"role_id":   s.roleID,
		"secret_id": s.secretID,
	}
	resp, err := client.Logical().Write("auth/approle/login", data)
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("Empty response from Vault for approle login, possible 404 not found")
	}
	if resp.Auth == nil {
		return fmt.Errorf("no auth info returned")
	}
	client.SetToken(resp.Auth.ClientToken)
	return nil

}

func (s *AppRoleAuth) Type() string {
	return "approle"
}
