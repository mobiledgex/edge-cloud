package edgeproto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestNewUdec64(t *testing.T) {
	tests := []struct {
		inWhole  uint64
		inNanos  uint32
		outWhole uint64
		outNanos uint32
	}{
		{2, 10, 2, 10},
		{2, 2 * DecWhole, 4, 0},       // nanos overflow
		{2, 2123456789, 4, 123456789}, // nanos overflow
	}
	for _, test := range tests {
		dec := NewUdec64(test.inWhole, test.inNanos)
		require.NotNil(t, dec)
		require.Equal(t, test.outWhole, dec.Whole)
		require.Equal(t, test.outNanos, dec.Nanos)
	}
}

func TestCmpUdec64(t *testing.T) {
	tests := []struct {
		aWhole uint64
		aNanos uint32
		bWhole uint64
		bNanos uint32
		res    int
	}{
		{0, 0, 0, 0, 0},
		{1, 0, 0, 0, 1},
		{0, 1, 0, 0, 1},
		{0, 0, 1, 0, -1},
		{0, 0, 0, 1, -1},
		{1, 1, 1, 1, 0},
		{3, 100, 3, 0, 1},
		{3, 0, 3, 100, -1},
		{4, 0, 3, 100, 1},
		{3, 100, 4, 0, -1},
	}
	for _, test := range tests {
		a := NewUdec64(test.aWhole, test.aNanos)
		b := NewUdec64(test.bWhole, test.bNanos)
		res := a.Cmp(b)
		require.Equal(t, test.res, res)
		eq, gt, lt := cmpToBools(res)
		require.Equal(t, eq, a.Equal(b))
		require.Equal(t, gt, a.GreaterThan(b))
		require.Equal(t, lt, a.LessThan(b))
		// test CmpUint64 assuming Cmp is valid
		b.Nanos = 0
		res = a.Cmp(b)
		res64 := a.CmpUint64(b.Whole)
		require.Equal(t, res, res64)
		eq, gt, lt = cmpToBools(res)
		require.Equal(t, eq, a.EqualUint64(b.Whole))
		require.Equal(t, gt, a.GreaterThanUint64(b.Whole))
		require.Equal(t, lt, a.LessThanUint64(b.Whole))
	}
}

// returns eq, gt, lt bools
func cmpToBools(cmp int) (bool, bool, bool) {
	if cmp == 0 {
		return true, false, false
	} else if cmp == 1 {
		return false, true, false
	} else {
		return false, false, true
	}
}

func TestMathUdec64(t *testing.T) {
	tests := []struct {
		aWhole   uint64
		aNanos   uint32
		bWhole   uint64
		bNanos   uint32
		resWhole uint64
		resNanos uint32
		op       int
	}{
		{0, 0, 0, 0, 0, 0, 1},
		{0, 0, 0, 0, 0, 0, -1},
		{2, 4, 1, 3, 3, 7, 1},
		{2, 4, 1, 3, 1, 1, -1},
		{2, 100, 1, 200, 3, 300, 1},
		{2, 100, 1, 200, 0, 999999900, -1}, // borrow case
	}
	for _, test := range tests {
		a := NewUdec64(test.aWhole, test.aNanos)
		b := NewUdec64(test.bWhole, test.bNanos)
		r := NewUdec64(test.resWhole, test.resNanos)
		if test.op == 1 {
			a.Add(b)
		} else {
			a.Sub(b)
		}
		require.Equal(t, r, a)
	}
}

func TestParseUdec64(t *testing.T) {
	tests := []struct {
		str      string
		outWhole uint64
		outNanos uint32
		expErr   string
		expRev   bool // reversing back to string yields same as input
	}{
		{"2", 2, 0, "", true},
		{"2.1", 2, 100 * DecMillis, "", true},
		{"2.01", 2, 10 * DecMillis, "", true},
		{"2.001", 2, 1 * DecMillis, "", true},
		{"2.0001", 2, 100 * DecMicros, "", true},
		{"2.00001", 2, 10 * DecMicros, "", true},
		{"2.000001", 2, 1 * DecMicros, "", true},
		{"2.0000001", 2, 100 * DecNanos, "", true},
		{"2.00000001", 2, 10 * DecNanos, "", true},
		{"2.000000001", 2, 1 * DecNanos, "", true},
		{"2.0000000001", 0, 0, "precision cannot be smaller than nanos", false},
		{"2.1m", 0, 0, "Parsing decimal number 1m failed", false},
		{"02", 2, 0, "", false},
		{"02.100000000", 2, 100 * DecMillis, "", false},
		{"02.010000000", 2, 10 * DecMillis, "", false},
		{"02.001000000", 2, 1 * DecMillis, "", false},
		{"02.000100000", 2, 100 * DecMicros, "", false},
		{"02.000010000", 2, 10 * DecMicros, "", false},
		{"02.000001000", 2, 1 * DecMicros, "", false},
		{"02.000000100", 2, 100 * DecNanos, "", false},
		{"02.000000010", 2, 10 * DecNanos, "", false},
		{"02.000000001", 2, 1 * DecNanos, "", false},
	}
	for _, test := range tests {
		actual, err := ParseUdec64(test.str)
		if test.expErr == "" {
			require.Nil(t, err)
			exp := NewUdec64(test.outWhole, test.outNanos)
			require.Equal(t, exp, actual)
			if test.expRev {
				outStr := exp.DecString()
				require.Equal(t, test.str, outStr)
			}
		} else {
			require.Contains(t, err.Error(), test.expErr)
		}
	}
}

type testMarshalUdec64Obj struct {
	Name     string
	Value    Udec64
	PtrValue *Udec64
	IntValue Udec64
}

func TestMarshalUdec64(t *testing.T) {
	obj := testMarshalUdec64Obj{
		Name:     "foo",
		Value:    *NewUdec64(2, 500*DecMillis),
		PtrValue: NewUdec64(3, 600*DecMillis),
		IntValue: *NewUdec64(22, 0),
	}
	expYaml := `name: foo
value: "2.5"
ptrvalue: "3.6"
intvalue: 22
`
	expJson := `{"Name":"foo","Value":"2.5","PtrValue":"3.6","IntValue":22}`

	outYaml, err := yaml.Marshal(obj)
	require.Nil(t, err)
	require.Equal(t, expYaml, string(outYaml))

	outJson, err := json.Marshal(obj)
	require.Nil(t, err)
	require.Equal(t, expJson, string(outJson))

	yamlObj := testMarshalUdec64Obj{}
	err = yaml.Unmarshal(outYaml, &yamlObj)
	require.Nil(t, err)
	require.Equal(t, obj, yamlObj)

	jsonObj := testMarshalUdec64Obj{}
	err = json.Unmarshal(outJson, &jsonObj)
	require.Nil(t, err)
	require.Equal(t, obj, jsonObj)
}
