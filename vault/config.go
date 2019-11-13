package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

type Config struct {
	Addr string
	Auth Auth
}

func BestConfig(addr string) (*Config, error) {
	auth, err := BestAuth()
	if err != nil {
		return &Config{
			Addr: addr,
			Auth: &NoAuth{},
		}, err
	}
	return &Config{
		Addr: addr,
		Auth: auth,
	}, nil
}

func NewConfig(addr string, auth Auth) *Config {
	return &Config{
		Addr: addr,
		Auth: auth,
	}
}

func NewAppRoleConfig(addr, roleID, secretID string) *Config {
	return NewConfig(addr, NewAppRoleAuth(roleID, secretID))
}

func (s *Config) Login() (*api.Client, error) {
	if s.Auth == nil {
		return nil, fmt.Errorf("No vault Auth specified")
	}
	client, err := NewClient(s.Addr)
	if err != nil {
		return nil, err
	}
	err = s.Auth.Login(client)
	if err != nil {
		return nil, err
	}
	return client, nil
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
