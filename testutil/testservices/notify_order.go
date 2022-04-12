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

// This file is not in the same package as notify to avoid including
// the testing packages in the notify package.
package testservices

import (
	"fmt"
	"sort"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/stretchr/testify/require"
)

type typeOrder struct {
	name  string
	order int
}

// Check order dependencies for notify send.
// This encompasses both object dependencies (objects depend on other objects)
// and service-specific dependencies.
func CheckNotifySendOrder(t *testing.T, sendOrder map[string]int) {
	orders := []typeOrder{}
	for t, i := range sendOrder {
		to := typeOrder{
			name:  t,
			order: i,
		}
		orders = append(orders, to)
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].order < orders[j].order
	})
	for _, to := range orders {
		fmt.Printf("%d: %s\n", to.order, to.name)
	}

	for obj, deps := range edgeproto.GetReferencesMap() {
		objOrder, found := sendOrder[obj]
		if !found {
			// object isn't sent, ignore
			continue
		}
		for _, dep := range deps {
			depOrder, found := sendOrder[dep]
			if !found {
				// object isn't sent, ignore
				continue
			}
			require.Greater(t, objOrder, depOrder, obj+" depends on "+dep)
		}
	}
}
