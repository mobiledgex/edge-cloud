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
