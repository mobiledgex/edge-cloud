package vault

import (
	"fmt"
	"os"
	"runtime"

	"github.com/hashicorp/vault/api"
)

type Auth interface {
	// Login to vault and set the vault token on the client
	Login(client *api.Client) error
	// Return auth type
	Type() string
}

// BestAuth determines the best auth to use based on the environment.
func BestAuth() (Auth, error) {
	roleID := os.Getenv("VAULT_ROLE_ID")
	secretID := os.Getenv("VAULT_SECRET_ID")
	if roleID != "" && secretID != "" {
		return NewAppRoleAuth(roleID, secretID), nil
	}
	githubID := os.Getenv("GITHUB_ID")
	if runtime.GOOS == "darwin" && githubID != "" {
		return NewGithubAuth(githubID), nil
	}
	return nil, fmt.Errorf("No appropriate auth found, please set either VAULT_ROLE_ID and VAULT_SECRET_ID for approle auth, or GITHUB_ID for github token auth.")
}
