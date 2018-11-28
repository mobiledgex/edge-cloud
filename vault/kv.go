package vault

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
)

// Data structures below are representations of the data returned
// from Vault queries. They are generic for any data stored in
// a v2 key value engine.

// KVData is the data from the normal get path
type KVData struct {
	Metadata KVMeta
	Data     map[string]interface{}
}

// KVMetadata is the metadata from the metadata path
type KVMetadata struct {
	CurrentVersion int `mapstructure:"current_version"`
	MaxVersions    int `mapstructure:"max_versions"`
	OldestVersion  int `mapstructure:"oldest_version"`
	Versions       map[int]KVMeta
}

type KVMeta struct {
	CreatedTime  string `mapstructure:"created_time"`
	DeletionTime string `mapstructure:"deletion_time"`
	Destroyed    bool
	Version      int
}

func GetKV(client *api.Client, path string, version int) (map[string]interface{}, error) {
	var extra map[string][]string
	if version > 0 {
		extra = make(map[string][]string)
		vstr := strconv.Itoa(version)
		extra["version"] = []string{vstr}
	}
	secret, err := client.Logical().ReadWithData(path, extra)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, fmt.Errorf("no secrets at path %s", path)
	}
	return secret.Data, nil
}

func PutKV(client *api.Client, path string, data map[string]interface{}) error {
	_, err := client.Logical().Write(path, data)
	return err
}

func ParseMetadata(data map[string]interface{}) (*KVMetadata, error) {
	meta := &KVMetadata{}
	err := mapstructure.WeakDecode(data, meta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func ParseData(data map[string]interface{}) (*KVData, error) {
	d := &KVData{}
	err := mapstructure.WeakDecode(data, d)
	if err != nil {
		return nil, err
	}
	return d, nil
}
