// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vault

import (
	"fmt"
	"strconv"
	"strings"

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
	if secret.Data == nil {
		if len(secret.Warnings) > 0 {
			errStr := strings.Join(secret.Warnings, ";")
			return nil, fmt.Errorf("No data: %s", errStr)
		}
		return nil, fmt.Errorf("No data at path %s", path)
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

func DeleteKV(client *api.Client, path string) error {
	_, err := client.Logical().Delete(path)
	return err
}
