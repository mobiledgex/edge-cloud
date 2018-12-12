package crmutil

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/mobiledgex/edge-cloud-infra/openstack-prov/oscliapi"
	"github.com/mobiledgex/edge-cloud/log"
)

type EnvData struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type VaultDataDetail struct {
	Env []EnvData `json:"env"`
}

type VaultData struct {
	Data VaultDataDetail `json:"data"`
}

type VaultResponse struct {
	Data VaultData `json:"data"`
}

func GetVaultData(url string) ([]byte, error) {
	vault_token := os.Getenv("VAULT_TOKEN")
	if vault_token == "" {
		return nil, fmt.Errorf("no vault token")
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", vault_token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func GetVaultEnvResponse(contents []byte) (*VaultResponse, error) {
	vr := &VaultResponse{}
	err := yaml.Unmarshal(contents, vr)
	if err != nil {
		return nil, err
	}
	return vr, nil
}

var home = os.Getenv("HOME")

func interpolate(val string) string {
	if strings.HasPrefix(val, "$HOME") {
		val = strings.Replace(val, "$HOME", home, -1)
	}
	return val
}

func InternVaultEnv(envs []EnvData) error {
	for _, e := range envs {
		val := interpolate(e.Value)
		err := os.Setenv(e.Name, val)
		if err != nil {
			return err
		}
		log.DebugLog(log.DebugLevelMexos, "setenv", "name", e.Name, "value", val)
	}
	return nil
}

func CheckPlatformEnv(platformType string) error {
	if !strings.Contains(platformType, "openstack") { // TODO gcp,azure,...
		return nil
	}
	for _, n := range []struct {
		name   string
		getter func() string
	}{
		{"MEX_EXT_NETWORK", oscli.GetMEXExternalNetwork},
		{"MEX_EXT_ROUTER", oscli.GetMEXExternalRouter},
		{"MEX_NETWORK", oscli.GetMEXNetwork},
		{"MEX_SECURITY_RULE", oscli.GetMEXSecurityRule},
	} {
		ev := os.Getenv(n.name)
		if ev == "" {
			ev = n.getter()
		}
		if ev == "" {
			return fmt.Errorf("missing " + n.name)
		}
	}
	log.DebugLog(log.DebugLevelMexos, "doing sanity check")
	_, err := oscli.ListImages()
	if err != nil {
		log.DebugLog(log.DebugLevelMexos, "environment", "env", os.Environ())
		return fmt.Errorf("oscli sanity check failed, %v", err)
	}
	return nil
}
