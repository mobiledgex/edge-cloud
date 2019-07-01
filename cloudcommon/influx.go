package cloudcommon

import (
	"github.com/mitchellh/mapstructure"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

type InfluxCreds struct {
	User string
	Pass string
}

func getVaultInfluxPath(vaultAddr, region string) string {
	return vaultAddr + "/v1/secret/data/" + region + "/accounts/influxdb"
}

func GetInfluxDataAuth(vaultAddr, region string) *InfluxCreds {
	vaultPath := getVaultInfluxPath(vaultAddr, region)
	log.DebugLog(log.DebugLevelApi, "get influxDB credentials ", "vault-path", vaultPath)
	data, err := vault.GetVaultData(vaultPath)
	if err != nil {
		return nil
	}
	creds := &InfluxCreds{}
	err = mapstructure.WeakDecode(data["data"], creds)
	if err != nil {
		return nil
	}
	return creds
}
