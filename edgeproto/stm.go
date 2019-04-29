package edgeproto

import (
	v3 "github.com/coreos/etcd/clientv3"
	"github.com/mobiledgex/edge-cloud/objstore"
)

func GetSTMOpts(opts ...objstore.KVOp) []v3.OpOption {
	kvopts := objstore.GetKVOptions(opts)
	v3opts := []v3.OpOption{}
	if kvopts.LeaseID != 0 {
		v3opts = append(v3opts, v3.WithLease(v3.LeaseID(kvopts.LeaseID)))
	}
	return v3opts
}
