package vault

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type EnvData struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type VaultEnvData struct {
	Env []EnvData `json:"env"`
}

func GetData(config *Config, path string, version int, data interface{}) error {
	if config == nil {
		return fmt.Errorf("no vault Config specified")
	}
	client, err := config.Login()
	if err != nil {
		return err
	}
	vdat, err := GetKV(client, path, version)
	if err != nil {
		return err
	}
	return mapstructure.WeakDecode(vdat["data"], data)
}

func GetEnvVars(config *Config, path string) (map[string]string, error) {
	envData := &VaultEnvData{}
	err := GetData(config, path, 0, envData)
	if err != nil {
		return nil, err
	}
	vars := make(map[string]string, 1)
	for _, envData := range envData.Env {
		vars[envData.Name] = envData.Value
	}
	return vars, nil
}
