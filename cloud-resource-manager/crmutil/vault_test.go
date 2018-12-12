package crmutil

import (
	"fmt"
	"os"
	"testing"
)

func TestGetVaultData(t *testing.T) {
	vault_token := os.Getenv("VAULT_TOKEN")
	if vault_token == "" {
		return
	}
	dat, err := GetVaultData("https://vault.mobiledgex.net/v1/secret/data/cloudlet/openstack/hamburg/openrc.json")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(dat))
	vr, err := GetVaultEnvResponse(dat)
	if err != nil {
		t.Error(err)
	}
	for _, e := range vr.Data.Data.Env {
		fmt.Println(e.Name, e.Value)
	}
	dat, err = GetVaultData("https://vault.mobiledgex.net/v1/secret/data/cloudlet/openstack/mexenv.json")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(dat))
	vr, err = GetVaultEnvResponse(dat)
	if err != nil {
		t.Error(err)
	}
	for _, e := range vr.Data.Data.Env {
		fmt.Println(e.Name, e.Value)
	}
}
