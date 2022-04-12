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
	v3 "github.com/coreos/etcd/clientv3"
	"github.com/mobiledgex/edge-cloud/objstore"
)

func GetSTMOpts(opts ...objstore.KVOp) []v3.OpOption {
	kvopts := objstore.GetKVOptions(opts)
	v3opts := []v3.OpOption{}
	if kvopts.LeaseID != 0 {
		v3opts = append(v3opts, v3.WithLease(v3.LeaseID(kvopts.LeaseID)))
	}
	if kvopts.Rev != 0 {
		v3opts = append(v3opts, v3.WithRev(kvopts.Rev))
	}
	return v3opts
}
