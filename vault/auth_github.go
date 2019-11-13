package vault

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/mobiledgex/edge-cloud/log"
)

type GithubAuth struct {
	githubID    string
	githubToken string
}

// GetGithubAuth grabs the github token from keychain on mac OSX.
// This should only be used for local testing against real cloudlets
// when running services locally on the mac dev environment (laptop).
// It is not intended for production use.
func NewGithubAuth(githubID string) *GithubAuth {
	return &GithubAuth{
		githubID: githubID,
	}
}

// Login to Vault and return the client.
// This assumes the token used for github developement can also be
// used to access Vault.
func (s *GithubAuth) Login(client *api.Client) error {
	if s.githubToken == "" {
		server := "github.com"
		log.DebugLog(log.DebugLevelInfo, "github secret lookup", "account", s.githubID, "server", server)
		token, err := FindKeychainSecret(s.githubID, server)
		if err != nil {
			return err
		}
		if token == "" {
			return fmt.Errorf("empty token for keychain entry account %s server %s", s.githubID, server)
		}
		s.githubToken = token
	}
	data := map[string]interface{}{
		"token": s.githubToken,
	}

	// enforce https to protect github token
	u, err := url.Parse(client.Address())
	if err != nil {
		return fmt.Errorf("unable to parse vault address %s", client.Address())
	}
	if u.Scheme != "https" {
		return fmt.Errorf("vault address (%s) must use https for gitlab auth", client.Address())
	}

	resp, err := client.Logical().Write("/auth/github/login", data)
	if err != nil {
		return err
	}
	if resp.Auth == nil {
		return fmt.Errorf("no auth info returned")
	}
	client.SetToken(resp.Auth.ClientToken)
	return nil
}

func (s *GithubAuth) Type() string {
	return "github"
}

// Find a secret from keychain on OS X.
// Calling this function will typically prompt the user to enter their
// account password. This should only be used for local laptop testing.
func FindKeychainSecret(account, server string) (string, error) {
	if runtime.GOOS != "darwin" {
		return "", fmt.Errorf("keychain only supported on mac darwin")
	}
	args := []string{
		"find-internet-password",
		"-a", account,
		"-s", server,
		"-w",
	}
	cmd := exec.Command("security", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(out), "specified item could not be found") {
			return "", err
		}
		// try find-generic-password instead
		args[0] = "find-generic-password"
		cmd = exec.Command("security", args...)
		out, err = cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("find secret (%v) failed, %s, %v", args, string(out), err)
		}
	}
	secret := strings.TrimSpace(string(out))
	return secret, nil
}
