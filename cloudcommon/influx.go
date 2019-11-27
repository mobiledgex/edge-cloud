package cloudcommon

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

type InfluxCreds struct {
	User string
	Pass string
}

func GetInfluxDataAuth(vaultConfig *vault.Config, region string) (*InfluxCreds, error) {
	if vaultConfig.Addr == "" {
		// no vault address, either unit test or no auth needed
		return &InfluxCreds{}, nil
	}
	vaultPath := "/secret/data/" + region + "/accounts/influxdb"
	log.DebugLog(log.DebugLevelApi, "get influxDB credentials ", "vault-path", vaultPath)
	creds := &InfluxCreds{}
	err := vault.GetData(vaultConfig, vaultPath, 0, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to get influxDB credentials for %s, %v", vaultPath, err)
	}
	return creds, nil
}
