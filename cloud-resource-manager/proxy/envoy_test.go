package proxy

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPortMap(t *testing.T) {
	testPorts := "tcp:1,tcp:2,udp:5"
	testMap, err := buildPortsMapFromString(testPorts)
	require.Nil(t, err)
	_, found := testMap["tcp:1"]
	require.True(t, found)
	_, found = testMap["tcp:2"]
	require.True(t, found)
	_, found = testMap["udp:5"]
	require.True(t, found)
	require.Equal(t, 3, len(testMap))
	// test ranges
	testPorts = "tcp:1-10,tcp:20-25,udp:5-10"
	testMap, err = buildPortsMapFromString(testPorts)
	require.Nil(t, err)
	require.Equal(t, 22, len(testMap))
	for ii := 1; ii <= 10; ii++ {
		key := fmt.Sprintf("tcp:%d", ii)
		_, found = testMap[key]
		require.True(t, found)
	}
	for ii := 20; ii <= 25; ii++ {
		key := fmt.Sprintf("tcp:%d", ii)
		_, found = testMap[key]
		require.True(t, found)
	}
	for ii := 5; ii <= 10; ii++ {
		key := fmt.Sprintf("udp:%d", ii)
		_, found = testMap[key]
		require.True(t, found)
	}
	// test "all"
	testPorts = "all"
	testMap, err = buildPortsMapFromString(testPorts)
	require.Nil(t, err)
	require.Equal(t, 0, len(testMap))
	// invalid string
	testPorts = "nonport:12"
	testMap, err = buildPortsMapFromString(testPorts)
	require.NotNil(t, err)
	require.Nil(t, testMap)
}
