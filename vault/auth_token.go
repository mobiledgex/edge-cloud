package vault

import "github.com/hashicorp/vault/api"

type TokenAuth struct {
	token string
}

func NewTokenAuth(token string) *TokenAuth {
	auth := TokenAuth{
		token: token,
	}
	return &auth
}

func (s *TokenAuth) Login(client *api.Client) error {
	client.SetToken(s.token)
	return nil
}

func (s *TokenAuth) Type() string {
	return "token"
}
