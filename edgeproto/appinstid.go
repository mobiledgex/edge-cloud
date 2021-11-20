package edgeproto

import (
	fmt "fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/objstore"
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
