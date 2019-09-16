package cli

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestObj struct {
	Inner1     InnerObj
	Inner2     *InnerObj
	unexported string
	Arr        []string          // unsupported
	Mmm        map[string]string // unsupported
}

type InnerObj struct {
	Name string
	Val  int
}

func TestParseArgs(t *testing.T) {
	var args []string
	input := &Input{}

	ex := TestObj{
		Inner1: InnerObj{
			Name: "name1",
			Val:  1,
		},
		Inner2: &InnerObj{
			Name: "name2",
			Val:  2,
		},
	}
	args = []string{"inner1.name=name1", "inner1.val=1", "inner2.name=name2", "inner2.val=2"}
	testParseArgs(t, input, args, &ex, &TestObj{}, &TestObj{})

	// fails because of unsupported arrays/maps
	args = []string{"inner1.name=name1", "inner1.val=1", "inner2.name=name2", "inner2.val=2", "unexported=err", "arr=bad-array", "mmm=badmap"}
	_, err := input.ParseArgs(args, &TestObj{})
	assert.NotNil(t, err)

	rf := edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "x1.tiny",
		},
		Vcpus: 1,
		Disk:  2,
		Ram:   3,
	}
	args = []string{"vcpus=1", "disk=2", "key.name=\"x1.tiny\"", "ram=3"}
	// basic parsing
	testParseArgs(t, input, args, &rf, &edgeproto.Flavor{}, &edgeproto.Flavor{})

	// required args
	input.RequiredArgs = []string{"regionx"}
	_, err = input.ParseArgs(args, &edgeproto.Flavor{})
	require.NotNil(t, err)

	input.RequiredArgs = []string{"key.name"}
	testParseArgs(t, input, args, &rf, &edgeproto.Flavor{}, &edgeproto.Flavor{})

	// alias args
	input.AliasArgs = []string{"name=key.name"}
	args = []string{"vcpus=1", "disk=2", "name=x1.tiny", "ram=3"}
	input.RequiredArgs = []string{"name"}
	testParseArgs(t, input, args, &rf, &edgeproto.Flavor{}, &edgeproto.Flavor{})

	// test extra args
	args = []string{"vcpus=1", "disk=2", "name=x1.tiny", "ram=3", "extra.arg=foo"}
	_, err = input.ParseArgs(args, &edgeproto.Flavor{})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid args")

	// test enum
	input = &Input{
		DecodeHook: edgeproto.EnumDecodeHook,
	}

	rc := edgeproto.Cloudlet{
		IpSupport: edgeproto.IpSupport_IP_SUPPORT_DYNAMIC,
	}
	args = []string{"ipsupport=IpSupportDynamic"}
	testParseArgs(t, input, args, &rc, &edgeproto.Cloudlet{}, &edgeproto.Cloudlet{})
}

func testParseArgs(t *testing.T, input *Input, args []string, expected, buf1, buf2 interface{}) {
	// parse the args into a clean buffer
	dat, err := input.ParseArgs(args, buf1)
	require.Nil(t, err)
	// check that buffer matches expected
	require.Equal(t, expected, buf1, "buf1 %v\n", buf1)
	fmt.Printf("argsmap: %v\n", dat)

	// convert args to json
	jsmap, err := JsonMap(dat, buf1)
	require.Nil(t, err)
	fmt.Printf("jsonamp: %v\n", jsmap)

	byt, err := json.Marshal(jsmap)
	require.Nil(t, err)
	fmt.Printf("json: %s\n", string(byt))

	// unmarshal json into a clean buffer, should match expected
	err = json.Unmarshal([]byte(byt), buf2)
	require.Nil(t, err)
	require.Equal(t, expected, buf2, "buf2 %v\n", buf2)
}

func TestConversion(t *testing.T) {
	// test converting obj to args and then back to obj

	for _, flavor := range testutil.FlavorData {
		testConversion(t, &flavor, &edgeproto.Flavor{}, &edgeproto.Flavor{})
	}
	for _, dev := range testutil.DevData {
		testConversion(t, &dev, &edgeproto.Developer{}, &edgeproto.Developer{})
	}
	for _, app := range testutil.AppData {
		testConversion(t, &app, &edgeproto.App{}, &edgeproto.App{})
	}
	for _, op := range testutil.OperatorData {
		testConversion(t, &op, &edgeproto.Operator{}, &edgeproto.Operator{})
	}
	for _, cloudlet := range testutil.CloudletData {
		testConversion(t, &cloudlet, &edgeproto.Cloudlet{}, &edgeproto.Cloudlet{})
	}
	for _, cinst := range testutil.ClusterInstData {
		testConversion(t, &cinst, &edgeproto.ClusterInst{}, &edgeproto.ClusterInst{})
	}
	for _, appinst := range testutil.AppInstData {
		testConversion(t, &appinst, &edgeproto.AppInst{}, &edgeproto.AppInst{})
	}
	// CloudletInfo and CloudletRefs have arrays which aren't supported yet.
}

func testConversion(t *testing.T, obj interface{}, buf, buf2 interface{}) {
	// marshal object to args
	args, err := MarshalArgs(obj, nil)
	require.Nil(t, err, "marshal %v", obj)
	input := Input{
		DecodeHook: edgeproto.EnumDecodeHook,
	}
	fmt.Printf("args: %v\n", args)

	// parse args into buf, generate args map - should match original
	dat, err := input.ParseArgs(args, buf)
	require.Nil(t, err, "parse args %v", args)
	require.Equal(t, obj, buf)
	fmt.Printf("argsmap: %v\n", dat)

	// convert args map to json map
	jsmap, err := JsonMap(dat, obj)
	require.Nil(t, err, "json map")
	fmt.Printf("jsonmap: %s\n", jsmap)

	// simulate client to server, check that matches original
	byt, err := json.Marshal(jsmap)
	require.Nil(t, err, "marshal")
	err = json.Unmarshal(byt, buf2)
	require.Nil(t, err, "unmarshal")
	require.Equal(t, obj, buf2)
}
