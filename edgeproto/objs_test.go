package edgeproto

import (
	"testing"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/stretchr/testify/require"
)

func TestHealthCheckSkipPorts(t *testing.T) {
	// test a single app port and single skipHa port
	var testAppPorts = "tcp:8080"
	var testSkipHaPorts = testAppPorts
	appPorts, err := ParseAppPorts(testAppPorts)
	require.Nil(t, err)
	require.Equal(t, 1, len(appPorts))
	appPorts, err = SetPortsHealthCheck(appPorts, testSkipHaPorts)
	require.Nil(t, err)
	require.Equal(t, 1, len(appPorts))
	require.True(t, appPorts[0].SkipHealthCheck)
	// two ports and one skip ha port
	testAppPorts = "tcp:8080,udp:8080"
	testSkipHaPorts = "udp:8080"
	appPorts, err = ParseAppPorts(testAppPorts)
	require.Nil(t, err)
	require.Equal(t, 2, len(appPorts))
	appPorts, err = SetPortsHealthCheck(appPorts, testSkipHaPorts)
	require.Nil(t, err)
	require.Equal(t, 2, len(appPorts))
	for _, port := range appPorts {
		// skip UDP health checks, but health check tcp
		if port.Proto == dme.LProto_L_PROTO_UDP {
			require.True(t, port.SkipHealthCheck)
		} else {
			require.False(t, port.SkipHealthCheck)
			require.Equal(t, int32(0), port.EndPort)
		}
	}
	// test ranges
	testAppPorts = "tcp:10-20,tcp:30-40"
	testSkipHaPorts = "tcp:10-20"
	appPorts, err = ParseAppPorts(testAppPorts)
	require.Nil(t, err)
	require.Equal(t, 2, len(appPorts))
	appPorts, err = SetPortsHealthCheck(appPorts, testSkipHaPorts)
	require.Nil(t, err)
	require.Equal(t, 2, len(appPorts))
	for _, port := range appPorts {
		// range 10-20, hc is sipped
		if port.InternalPort == 10 {
			require.True(t, port.SkipHealthCheck)
			require.Equal(t, int32(20), port.EndPort)
		} else {
			require.False(t, port.SkipHealthCheck)
		}
	}
	// big range skip health check
	testAppPorts = "tcp:10-20,tcp:30-40"
	testSkipHaPorts = "tcp:5-50"
	appPorts, err = ParseAppPorts(testAppPorts)
	require.Nil(t, err)
	require.Equal(t, 2, len(appPorts))
	appPorts, err = SetPortsHealthCheck(appPorts, testSkipHaPorts)
	require.Nil(t, err)
	require.Equal(t, 2, len(appPorts))
	for _, port := range appPorts {
		// all ports should skip health check
		require.True(t, port.SkipHealthCheck)
	}
	// split range into to ranges with health check skip
	// health check range starts before app port range
	testAppPorts = "tcp:10-20"
	testSkipHaPorts = "tcp:5-15"
	appPorts, err = ParseAppPorts(testAppPorts)
	require.Nil(t, err)
	require.Equal(t, 1, len(appPorts))
	appPorts, err = SetPortsHealthCheck(appPorts, testSkipHaPorts)
	require.Nil(t, err)
	require.Equal(t, 2, len(appPorts))
	for _, port := range appPorts {
		// range 10-15, hc is skipped
		if port.InternalPort == 10 {
			require.True(t, port.SkipHealthCheck)
			require.Equal(t, int32(10), port.InternalPort)
			require.Equal(t, int32(15), port.EndPort)
		} else {
			require.Equal(t, int32(20), port.EndPort)
			require.Equal(t, int32(16), port.InternalPort)
			require.False(t, port.SkipHealthCheck)
		}
	}
	// health check range starts after app port range
	testAppPorts = "tcp:10-20"
	testSkipHaPorts = "tcp:15-20"
	appPorts, err = ParseAppPorts(testAppPorts)
	require.Nil(t, err)
	require.Equal(t, 1, len(appPorts))
	appPorts, err = SetPortsHealthCheck(appPorts, testSkipHaPorts)
	require.Nil(t, err)
	require.Equal(t, 2, len(appPorts))
	for _, port := range appPorts {
		if port.InternalPort == 10 {
			require.False(t, port.SkipHealthCheck)
			require.Equal(t, int32(10), port.InternalPort)
			require.Equal(t, int32(14), port.EndPort)
		} else {
			// range 15-20, hc is sipped
			require.Equal(t, int32(20), port.EndPort)
			require.Equal(t, int32(15), port.InternalPort)
			require.True(t, port.SkipHealthCheck)
		}
	}
	// health check range is inside app port range
	testAppPorts = "tcp:10-20"
	testSkipHaPorts = "tcp:15-18"
	appPorts, err = ParseAppPorts(testAppPorts)
	require.Nil(t, err)
	require.Equal(t, 1, len(appPorts))
	appPorts, err = SetPortsHealthCheck(appPorts, testSkipHaPorts)
	require.Nil(t, err)
	require.Equal(t, 3, len(appPorts))
	for _, port := range appPorts {
		if port.InternalPort == 10 {
			require.False(t, port.SkipHealthCheck)
			require.Equal(t, int32(10), port.InternalPort)
			require.Equal(t, int32(14), port.EndPort)
		} else if port.InternalPort == 15 {
			// range 15-18, hc is sipped
			require.Equal(t, int32(18), port.EndPort)
			require.Equal(t, int32(15), port.InternalPort)
			require.True(t, port.SkipHealthCheck)
		} else {
			require.False(t, port.SkipHealthCheck)
			require.Equal(t, int32(19), port.InternalPort)
			require.Equal(t, int32(20), port.EndPort)
		}
	}
	// single exclude port
	testAppPorts = "tcp:10-20"
	testSkipHaPorts = "tcp:15"
	appPorts, err = ParseAppPorts(testAppPorts)
	require.Nil(t, err)
	require.Equal(t, 1, len(appPorts))
	appPorts, err = SetPortsHealthCheck(appPorts, testSkipHaPorts)
	require.Nil(t, err)
	require.Equal(t, 3, len(appPorts))
	for _, port := range appPorts {
		if port.InternalPort == 10 {
			require.False(t, port.SkipHealthCheck)
			require.Equal(t, int32(10), port.InternalPort)
			require.Equal(t, int32(14), port.EndPort)
		} else if port.InternalPort == 15 {
			// single port - tcp:15
			require.Equal(t, int32(0), port.EndPort)
			require.Equal(t, int32(15), port.InternalPort)
			require.True(t, port.SkipHealthCheck)
		} else {
			require.False(t, port.SkipHealthCheck)
			require.Equal(t, int32(16), port.InternalPort)
			require.Equal(t, int32(20), port.EndPort)
		}
	}

	// health check range and app port range single port intersection
	testAppPorts = "tcp:10-20,udp:1-15,tcp:25-30"
	testSkipHaPorts = "tcp:20-25"
	appPorts, err = ParseAppPorts(testAppPorts)
	require.Nil(t, err)
	require.Equal(t, 3, len(appPorts))
	appPorts, err = SetPortsHealthCheck(appPorts, testSkipHaPorts)
	require.Nil(t, err)
	require.Equal(t, 5, len(appPorts))
	for _, port := range appPorts {
		// tcp:10-19,tcp:20-20,udp:1-10,tcp:25-25,tcp:26-30
		if port.Proto == dme.LProto_L_PROTO_UDP {
			require.False(t, port.SkipHealthCheck)
		} else if port.InternalPort == 10 {
			require.False(t, port.SkipHealthCheck)
			require.Equal(t, int32(10), port.InternalPort)
			require.Equal(t, int32(19), port.EndPort)
		} else if port.InternalPort == 20 {
			require.Equal(t, int32(0), port.EndPort)
			require.Equal(t, int32(20), port.InternalPort)
			require.True(t, port.SkipHealthCheck)
		} else if port.InternalPort == 25 {
			require.Equal(t, int32(0), port.EndPort)
			require.Equal(t, int32(25), port.InternalPort)
			require.True(t, port.SkipHealthCheck)
		} else if port.InternalPort == 26 {
			require.Equal(t, int32(30), port.EndPort)
			require.Equal(t, int32(26), port.InternalPort)
			require.False(t, port.SkipHealthCheck)
		}
	}
	// overlapping health-check skip ranges
	testAppPorts = "tcp:1-20"
	testSkipHaPorts = "tcp:7-10,tcp:5-15"
	appPorts, err = ParseAppPorts(testAppPorts)
	require.Nil(t, err)
	require.Equal(t, 1, len(appPorts))
	appPorts, err = SetPortsHealthCheck(appPorts, testSkipHaPorts)
	require.Nil(t, err)
	require.Equal(t, 5, len(appPorts))
	for _, port := range appPorts {
		// tcp:10-19,tcp:20-20,udp:1-10,tcp:25-25,tcp:26-30
		if port.Proto == dme.LProto_L_PROTO_UDP {
			require.False(t, port.SkipHealthCheck)
		} else if port.InternalPort == 1 {
			require.False(t, port.SkipHealthCheck)
			require.Equal(t, int32(1), port.InternalPort)
			require.Equal(t, int32(4), port.EndPort)
		} else if port.InternalPort == 5 {
			require.Equal(t, int32(5), port.InternalPort)
			require.Equal(t, int32(6), port.EndPort)
			require.True(t, port.SkipHealthCheck)
		} else if port.InternalPort == 7 {
			require.Equal(t, int32(7), port.InternalPort)
			require.Equal(t, int32(10), port.EndPort)
			require.True(t, port.SkipHealthCheck)
		} else if port.InternalPort == 11 {
			require.Equal(t, int32(11), port.InternalPort)
			require.Equal(t, int32(15), port.EndPort)
			require.True(t, port.SkipHealthCheck)
		} else if port.InternalPort == 16 {
			require.Equal(t, int32(16), port.InternalPort)
			require.Equal(t, int32(20), port.EndPort)
			require.False(t, port.SkipHealthCheck)
		} else {
			require.Fail(t, "Unexpected range")
		}
	}

}
