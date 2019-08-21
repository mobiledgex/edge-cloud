package cloudcommon

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

type InfluxCreds struct {
	User string
	Pass string
}

func getVaultInfluxPath(vaultAddr, region string) string {
	if vaultAddr == "" {
		return ""
	}
	return vaultAddr + "/v1/secret/data/" + region + "/accounts/influxdb"
}

func GetInfluxDataAuth(vaultAddr, region string) (*InfluxCreds, error) {
	if vaultAddr == "" {
		// no vault address, either unit test or no auth needed
		return &InfluxCreds{}, nil
	}
	vaultPath := getVaultInfluxPath(vaultAddr, region)
	log.DebugLog(log.DebugLevelApi, "get influxDB credentials ", "vault-path", vaultPath)
	data, err := vault.GetVaultData(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get influxDB credentials for %s, %v", vaultPath, err)
	}
	creds := &InfluxCreds{}
	err = mapstructure.WeakDecode(data["data"], creds)
	if err != nil {
		return nil, fmt.Errorf("decode vault influxDB data failed, %v", err)
	}
	return creds, nil
}
