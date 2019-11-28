package vault

import (
	"fmt"
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
func BestAuth(ops ...BestOp) (Auth, error) {
	opts := ApplyOps(ops...)

	roleID := opts.env.Getenv("VAULT_ROLE_ID")
	secretID := opts.env.Getenv("VAULT_SECRET_ID")
	if roleID != "" && secretID != "" {
		return NewAppRoleAuth(roleID, secretID), nil
	}
	githubID := opts.env.Getenv("GITHUB_ID")
	if runtime.GOOS == "darwin" && githubID != "" {
		return NewGithubAuth(githubID), nil
	}
	token := opts.env.Getenv("VAULT_TOKEN")
	if token != "" {
		return NewTokenAuth(token), nil
	}
	return nil, fmt.Errorf("No appropriate Vault auth found, please set VAULT_ROLE_ID and VAULT_SECRET_ID for approle auth, GITHUB_ID for github token auth, or VAULT_TOKEN for token auth.")
}
