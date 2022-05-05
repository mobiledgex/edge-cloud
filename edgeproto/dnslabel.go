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

package edgeproto

import (
	"encoding/json"
	fmt "fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/edgexr/edge-cloud/log"
	"github.com/edgexr/edge-cloud/objstore"
)

// CloudletDnsLabelStore is used to store Cloudlet DNS labels which are
// valid DNS segments and are unique within the region.
type CloudletDnsLabelStore struct{}

func CloudletDnsLabelDbKey(label string) string {
	return fmt.Sprintf("%s/%s", objstore.DbKeyPrefixString("CloudletDnsLabel"), label)
}

func (s *CloudletDnsLabelStore) STMHas(stm concurrency.STM, label string) bool {
	keystr := CloudletDnsLabelDbKey(label)
	valstr := stm.Get(keystr)
	if valstr == "" {
		return false
	}
	return true
}

func (s *CloudletDnsLabelStore) STMPut(stm concurrency.STM, label string) {
	keystr := CloudletDnsLabelDbKey(label)
	stm.Put(keystr, label)
}

func (s *CloudletDnsLabelStore) STMDel(stm concurrency.STM, label string) {
	keystr := CloudletDnsLabelDbKey(label)
	stm.Del(keystr)
}

// CloudletObjectDnsLabelStore is used to store Cloudlet Object DNS labels
// which are valid DNS segments and are unique within the cloudlet.
// Examples of cloudlet objects are AppInsts and ClusterInsts.
type CloudletObjectDnsLabelStore struct{}

type CloudletObjectDnsLabelKey struct {
	Cloudlet CloudletKey
	Label    string
}

func CloudletObjectDnsLabelDbKey(key *CloudletKey, label string) string {
	objKey := CloudletObjectDnsLabelKey{
		Cloudlet: *key,
		Label:    label,
	}
	keystr, err := json.Marshal(objKey)
	if err != nil {
		log.FatalLog("Failed to marshal CloudletObjectDnsLabelKey", "obj", objKey, "err", err)
	}
	return fmt.Sprintf("%s/%s", objstore.DbKeyPrefixString("CloudletObjectDnsLabel"), string(keystr))
}

func (s *CloudletObjectDnsLabelStore) STMHas(stm concurrency.STM, key *CloudletKey, label string) bool {
	keystr := CloudletObjectDnsLabelDbKey(key, label)
	valstr := stm.Get(keystr)
	if valstr == "" {
		return false
	}
	return true
}

func (s *CloudletObjectDnsLabelStore) STMPut(stm concurrency.STM, key *CloudletKey, label string) {
	keystr := CloudletObjectDnsLabelDbKey(key, label)
	stm.Put(keystr, label)
}

func (s *CloudletObjectDnsLabelStore) STMDel(stm concurrency.STM, key *CloudletKey, label string) {
	keystr := CloudletObjectDnsLabelDbKey(key, label)
	stm.Del(keystr)
}
