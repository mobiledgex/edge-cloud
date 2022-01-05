package edgeproto

import (
	"encoding/json"
	fmt "fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
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
