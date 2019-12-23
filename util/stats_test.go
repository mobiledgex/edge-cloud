package util_test

import (
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/require"
)

func TestGetWaitTime(t *testing.T) {
	// now, interval, offset, expected wait time
	// trigger on 0,10,20,30,...,90,100,110,...
	testGetWaitTime(t, 0, 10, 0, 10)
	testGetWaitTime(t, 100, 10, 0, 10)
	testGetWaitTime(t, 99, 10, 0, 1)
	testGetWaitTime(t, 1, 10, 0, 9)
	// trigger on 0,3,6,9,12,...
	testGetWaitTime(t, 6, 3, 0, 3)
	testGetWaitTime(t, 7, 3, 0, 2)
	testGetWaitTime(t, 8, 3, 0, 1)
	testGetWaitTime(t, 9, 3, 0, 3)
	testGetWaitTime(t, 10, 3, 0, 2)
	testGetWaitTime(t, 11, 3, 0, 1)
	testGetWaitTime(t, 12, 3, 0, 3)
	// tests with offset, trigger on 1,11,21,31,...
	testGetWaitTime(t, 10, 10, 1, 1)
	testGetWaitTime(t, 11, 10, 1, 10)
	testGetWaitTime(t, 12, 10, 1, 9)
	testGetWaitTime(t, 20, 10, 1, 1)
	testGetWaitTime(t, 21, 10, 1, 10)
	testGetWaitTime(t, 25, 10, 1, 6)
	// test with offset > interval
	testGetWaitTime(t, 30, 10, 20, 10)
	testGetWaitTime(t, 30, 10, 25, 5)
	// test with fractional offset, trigger on 1.2,2.2,3.2,...
	testGetWaitTime(t, 1, 1, 0.2, 0.2)
	testGetWaitTime(t, 1.1, 1, 0.2, 0.1)
	testGetWaitTime(t, 1.2, 1, 0.2, 1)
	testGetWaitTime(t, 1.3, 1, 0.2, 0.9)
	testGetWaitTime(t, 9999999.2, 1, 0.2, 1)
	testGetWaitTime(t, 9999999.3, 1, 0.2, 0.9)
}

// All units are in seconds
func testGetWaitTime(t *testing.T, now, interval, offset, expected float64) {
	tnow := time.Unix(0, int64(now*1e9))
	dur := util.GetWaitTime(tnow, interval, offset)
	require.Equal(t, expected, dur.Seconds(), "GetWaitTime for now(%d), interval(%d), offset(%d)", now, interval, offset)
}

type shardKey struct {
	Name string
	Kind kindKey
}

type kindKey struct {
	Value int
}

func TestGetShardIndex(t *testing.T) {
	key := shardKey{}
	key.Name = "foo"
	key.Kind.Value = 1
	numShards := uint(100)

	idx := util.GetShardIndex(key, numShards)
	idx2 := util.GetShardIndex(key, numShards)
	require.Equal(t, idx, idx2)

	key.Kind.Value = 2
	idx2 = util.GetShardIndex(key, numShards)
	require.NotEqual(t, idx, idx2)
}
