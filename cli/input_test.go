package cli_test

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/mobiledgex/edge-cloud/cli"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/testutil"
	yaml "github.com/mobiledgex/yaml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestObj struct {
	Inner1     InnerObj
	Inner2     *InnerObj
	unexported string
	Arr        []string
	Mmm        map[string]string
	ObjArr     []InnerObj
}

type InnerObj struct {
	Name string
	Val  int
	Mmm  map[string]string
}

func TestParseArgs(t *testing.T) {
	var args []string

	spargs := cli.GetSpecialArgs(TestObj{})
	expectSpArgs := map[string]string{
		"mmm":        "StringToString",
		"arr":        "StringArray",
		"inner1.mmm": "StringToString",
		"inner2.mmm": "StringToString",
	}
	require.Equal(t, expectSpArgs, spargs, "GetSpecialArgs")

	input := &cli.Input{
		SpecialArgs: &spargs,
	}

	ex := TestObj{
		Inner1: InnerObj{
			Name: "name1",
			Val:  1,
			Mmm: map[string]string{
				"xkey": "xx",
			},
		},
		Inner2: &InnerObj{
			Name: "name2",
			Val:  2,
		},
		Arr: []string{"foo", "bar", "baz"},
		Mmm: map[string]string{
			"key1":         "val1",
			"key.with.dot": "val.with.dot",
			"keye":         "val=with=equals",
			// key with equals not supported
			//"key=with=equals": "val=with=equals",
		},
		ObjArr: []InnerObj{
			InnerObj{
				Name: "arrin1",
				Val:  3,
			},
			InnerObj{
				Name: "arrin2",
				Val:  4,
			},
		},
	}

	args = []string{"inner1.name=name1", "inner1.val=1", "inner2.name=name2", "inner2.val=2", "inner1.mmm=xkey=xx", "arr=foo", "arr=bar", "arr=baz", "mmm=key1=val1", "mmm=key.with.dot=val.with.dot", "mmm=keye=val=with=equals", "objarr:0.name=arrin1", "objarr:0.val=3", "objarr:1.name=arrin2", "objarr:1.val=4"}
	testConversion(t, input, &ex, &TestObj{}, &TestObj{}, args)

	// test with alias args
	inputAliased := &cli.Input{
		SpecialArgs: &spargs,
		AliasArgs: []string{
			"name1=inner1.name",
			"name2=inner2.name",
			"val1=inner1.val",
			"val2=inner2.val",
			"mmm1=inner1.mmm",
			"mmm2=inner2.mmm",
		},
	}
	args = []string{"name1=name1", "val1=1", "name2=name2", "val2=2", "mmm1=xkey=xx", "arr=foo", "arr=bar", "arr=baz", "mmm=key1=val1", "mmm=key.with.dot=val.with.dot", "mmm=keye=val=with=equals", "objarr:0.name=arrin1", "objarr:0.val=3", "objarr:1.name=arrin2", "objarr:1.val=4"}
	testConversion(t, inputAliased, &ex, &TestObj{}, &TestObj{}, args)

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
	_, err := input.ParseArgs(args, &edgeproto.Flavor{})
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
	input = &cli.Input{
		DecodeHook: edgeproto.EnumDecodeHook,
	}

	rc := edgeproto.Cloudlet{
		IpSupport: edgeproto.IpSupport_IP_SUPPORT_DYNAMIC,
	}
	args = []string{"ipsupport=IpSupportDynamic"}
	testParseArgs(t, input, args, &rc, &edgeproto.Cloudlet{}, &edgeproto.Cloudlet{})
}

func testParseArgs(t *testing.T, input *cli.Input, args []string, expected, buf1, buf2 interface{}) {
	// parse the args into a clean buffer
	dat, err := input.ParseArgs(args, buf1)
	require.Nil(t, err)
	// check that buffer matches expected
	require.Equal(t, expected, buf1, "buf1 %v\n", buf1)
	fmt.Printf("argsmap: %v\n", dat)

	// convert args to json
	jsmap, err := cli.JsonMap(dat, buf1, cli.StructNamespace)
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
	input := &cli.Input{
		DecodeHook: edgeproto.DecodeHook,
		SpecialArgs: &map[string]string{
			"fields":           "StringArray",
			"autoprovpolicies": "StringArray",
		},
	}
	for _, flavor := range testutil.FlavorData {
		testConversion(t, input, &flavor, &edgeproto.Flavor{}, &edgeproto.Flavor{}, nil)
	}
	for _, app := range testutil.AppData {
		testConversion(t, input, &app, &edgeproto.App{}, &edgeproto.App{}, nil)
	}
	for _, cloudlet := range testutil.CloudletData {
		testConversion(t, input, &cloudlet, &edgeproto.Cloudlet{}, &edgeproto.Cloudlet{}, nil)
	}
	for _, cinst := range testutil.ClusterInstData {
		testConversion(t, input, &cinst, &edgeproto.ClusterInst{}, &edgeproto.ClusterInst{}, nil)
	}
	for _, appinst := range testutil.AppInstData {
		testConversion(t, input, &appinst, &edgeproto.AppInst{}, &edgeproto.AppInst{}, nil)
	}
	for _, pp := range testutil.PrivacyPolicyData {
		testConversion(t, input, &pp, &edgeproto.PrivacyPolicy{}, &edgeproto.PrivacyPolicy{}, nil)
	}
	settings := edgeproto.GetDefaultSettings()
	settings.Fields = []string{"16", "4", "9", "2.2"}
	testConversion(t, input, settings, &edgeproto.Settings{}, &edgeproto.Settings{}, nil)
	// CloudletInfo and CloudletRefs have arrays which aren't supported yet.
}

func testConversion(t *testing.T, input *cli.Input, obj interface{}, buf, buf2 interface{}, expectedArgs []string) {
	// marshal object to args
	args, err := cli.MarshalArgs(obj, nil, input.AliasArgs)
	require.Nil(t, err, "marshal %v", obj)

	fmt.Printf("args: %v\n", args)
	if expectedArgs != nil {
		sortargs := make([]string, len(args))
		copy(sortargs, args)
		sort.Strings(expectedArgs)
		sort.Strings(sortargs)
		require.Equal(t, expectedArgs, sortargs)
	}

	// parse args into buf, generate args map - should match original
	dat, err := input.ParseArgs(args, buf)
	require.Nil(t, err, "parse args %v", args)
	require.Equal(t, obj, buf)
	fmt.Printf("argsmap: %v\n", dat)

	// convert args map to json map
	jsmap, err := cli.JsonMap(dat, obj, cli.StructNamespace)
	require.Nil(t, err, "json map")
	fmt.Printf("jsonmap: %s\n", jsmap)

	// simulate client to server, check that matches original
	byt, err := json.Marshal(jsmap)
	require.Nil(t, err, "marshal")
	fmt.Printf("json string: %s\n", string(byt))
	err = json.Unmarshal(byt, buf2)
	require.Nil(t, err, "unmarshal")
	require.Equal(t, obj, buf2)
}

type mapTest struct {
	yamlData string
	jsonData string
	obj1     interface{}
	obj2     interface{}
	obj3     interface{}
	obj4     interface{}
}

var mapTestData = []*mapTest{
	&mapTest{
		yamlData: `
flavors:
- key:
    name: x1.tiny
  ram: 1024
  vcpus: 1
  disk: 1
- key:
    name: x1.small
  ram: 2048
  vcpus: 2
  disk: 2
- key:
    name: x1.medium
  ram: 4096
  vcpus: 4
  disk: 4

clusterinsts:
- key:
    clusterkey:
      name: SmallCluster
    cloudletkey:
      organization: dmuus
      name: dmuus-cloud-1
    organization: AcmeAppCo
  flavor:
    name: x1.small
  liveness: LivenessStatic
  ipaccess: IpAccessShared
  nummasters: 1
  numnodes: 2

- key:
    clusterkey:
      name: SmallCluster
    cloudletkey:
      organization: dmuus
      name: dmuus-cloud-2
    organization: AcmeAppCo
  flavor:
    name: x1.small
  liveness: LivenessStatic
  ipaccess: IpAccessDedicated
  nummasters: 1
  numnodes: 2
`,
		jsonData: `
{
   "flavors": [
      {
         "key": {
            "name": "x1.tiny"
         },
         "ram": 1024,
         "vcpus": 1,
         "disk": 1
      },
      {
         "key": {
            "name": "x1.small"
         },
         "ram": 2048,
         "vcpus": 2,
         "disk": 2
      },
      {
         "key": {
            "name": "x1.medium"
         },
         "ram": 4096,
         "vcpus": 4,
         "disk": 4
      }
   ],
   "cluster_insts": [
      {
         "key": {
            "cluster_key": {
               "name": "SmallCluster"
            },
            "cloudlet_key": {
               "organization": "dmuus",
               "name": "dmuus-cloud-1"
            },
            "organization": "AcmeAppCo"
         },
         "flavor": {
            "name": "x1.small"
         },
         "liveness": "LivenessStatic",
         "ip_access": "IpAccessShared",
         "num_masters": 1,
         "num_nodes": 2
      },
      {
         "key": {
            "cluster_key": {
               "name": "SmallCluster"
            },
            "cloudlet_key": {
               "organization": "dmuus",
               "name": "dmuus-cloud-2"
            },
            "organization": "AcmeAppCo"
         },
         "flavor": {
            "name": "x1.small"
         },
         "liveness": "LivenessStatic",
         "ip_access": "IpAccessDedicated",
         "num_masters": 1,
         "num_nodes": 2
      }
   ]
}`,
		obj1: &edgeproto.AllData{},
		obj2: &edgeproto.AllData{},
		obj3: &edgeproto.AllData{},
		obj4: &edgeproto.AllData{},
	},
}

func TestJsonMapFromYaml(t *testing.T) {
	for ii, d := range mapTestData {
		// check that test data yaml and json are equivalent
		err := yaml.Unmarshal([]byte(d.yamlData), d.obj1)
		require.Nil(t, err, "%d: unmarshal yaml to obj", ii)
		err = json.Unmarshal([]byte(d.jsonData), d.obj2)
		require.Nil(t, err, "unmarshal json to obj")
		require.Equal(t, d.obj1, d.obj2, "%d: objects equal", ii)

		// convert yaml/json to generic map[string]interface{}
		yamlMap := make(map[string]interface{})
		err = yaml.Unmarshal([]byte(d.yamlData), &yamlMap)
		require.Nil(t, err, "%d: unmarshal yaml to map", ii)
		jsonMap := make(map[string]interface{})
		err = json.Unmarshal([]byte(d.jsonData), &jsonMap)
		require.Nil(t, err, "%d: unmarshal json to map", ii)

		// convert yaml map to json map (this is the focus of the test)
		jsonMapFromYaml, err := cli.JsonMap(jsonMap, d.obj3, cli.YamlNamespace)
		require.Nil(t, err, "%d: jsonMap conversion")

		// Due to difference with enums and numeric types
		// (int vs float), the jsonMap and jsonMapFromYaml are not
		// directly comparable to each other. So we decode into
		// objects and compare the objects to verify.
		_, err = cli.WeakDecode(jsonMap, d.obj3, edgeproto.EnumDecodeHook)
		require.Nil(t, err, "%d: decode jsonMap")
		_, err = cli.WeakDecode(jsonMapFromYaml, d.obj4, edgeproto.EnumDecodeHook)
		require.Nil(t, err, "%d: decode jsonMapFromYaml")
		require.Equal(t, d.obj3, d.obj4, "%d: objects equal after mapping")
	}
}
