package vault

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

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
