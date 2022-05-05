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
	fmt "fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/edgexr/edge-cloud/objstore"
)

type AppInstIdStore struct{}

func AppInstIdDbKey(id string) string {
	return fmt.Sprintf("%s/%s", objstore.DbKeyPrefixString("AppInstId"), id)
}

func (s *AppInstIdStore) STMHas(stm concurrency.STM, id string) bool {
	keystr := AppInstIdDbKey(id)
	valstr := stm.Get(keystr)
	if valstr == "" {
		return false
	}
	return true
}

func (s *AppInstIdStore) STMPut(stm concurrency.STM, id string) {
	keystr := AppInstIdDbKey(id)
	stm.Put(keystr, id)
}

func (s *AppInstIdStore) STMDel(stm concurrency.STM, id string) {
	keystr := AppInstIdDbKey(id)
	stm.Del(keystr)
}
