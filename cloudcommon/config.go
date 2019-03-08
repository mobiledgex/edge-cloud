package cloudcommon

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	yaml "gopkg.in/yaml.v2"
)

type AppConfig struct {
	Resources string `json:"resources"`
}

func GetAppConfig(app *edgeproto.App) (string, error) {
	// config may be remote target or inline json/yaml
	if strings.HasPrefix(app.Config, "http://") ||
		strings.HasPrefix(app.Config, "https://") {
		str, err := GetRemoteManifest(app.Config)
		if err != nil {
			return "", fmt.Errorf("cannot get config from %s, %v", app.Config, err)
		}
		return str, nil
	}
	// inline config
	return app.Config, nil
}

func ParseAppConfig(configStr string) (*AppConfig, error) {
	config := &AppConfig{}
	if configStr != "" {
		err := json.Unmarshal([]byte(configStr), config)
		if err != nil {
			err = yaml.Unmarshal([]byte(configStr), config)
		}
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal json/yaml config str, err %v, config `%s`", err, configStr)
		}
	}
	return config, nil
}
