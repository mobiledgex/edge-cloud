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

package cloudcommon

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCidrContainsCidr(t *testing.T) {

	cidr1 := "1.1.1.1/24"
	_, net1, _ := net.ParseCIDR(cidr1)

	cidr2 := "1.1.1.1/16"
	_, net2, _ := net.ParseCIDR(cidr2)

	require.False(t, CidrContainsCidr(net1, net2))
	require.True(t, CidrContainsCidr(net2, net1))

	cidr1 = "1.1.1.55/24"
	_, net1, _ = net.ParseCIDR(cidr1)

	cidr2 = "1.1.3.4/16"
	_, net2, _ = net.ParseCIDR(cidr2)

	require.False(t, CidrContainsCidr(net1, net2))
	require.True(t, CidrContainsCidr(net2, net1))

	cidr1 = "1.1.3.55/24"
	_, net1, _ = net.ParseCIDR(cidr1)

	cidr2 = "1.2.3.4/16"
	_, net2, _ = net.ParseCIDR(cidr2)

	require.False(t, CidrContainsCidr(net1, net2))
	require.False(t, CidrContainsCidr(net2, net1))

	cidr1 = "1.1.3.55/24"
	_, net1, _ = net.ParseCIDR(cidr1)

	cidr2 = "1.2.3.4/8"
	_, net2, _ = net.ParseCIDR(cidr2)

	require.True(t, CidrContainsCidr(net2, net1))
}
