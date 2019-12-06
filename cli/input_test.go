package cli_test

import (
	"encoding/json"
	"fmt"
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
	Arr        []string          // unsupported
	Mmm        map[string]string // unsupported
}

type InnerObj struct {
	Name string
	Val  int
}

func TestParseArgs(t *testing.T) {
	var args []string
	input := &cli.Input{}

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
	args, err := cli.MarshalArgs(obj, nil)
	require.Nil(t, err, "marshal %v", obj)
	input := cli.Input{
		DecodeHook: edgeproto.EnumDecodeHook,
	}
	fmt.Printf("args: %v\n", args)

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
      operatorkey:
        name: tmus
      name: tmus-cloud-1
    developer: AcmeAppCo
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
      operatorkey:
        name: tmus
      name: tmus-cloud-2
    developer: AcmeAppCo
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
   "clusterinsts": [
      {
         "key": {
            "cluster_key": {
               "name": "SmallCluster"
            },
            "cloudlet_key": {
               "operator_key": {
                  "name": "tmus"
               },
               "name": "tmus-cloud-1"
            },
            "developer": "AcmeAppCo"
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
               "operator_key": {
                  "name": "tmus"
               },
               "name": "tmus-cloud-2"
            },
            "developer": "AcmeAppCo"
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
		obj1: &edgeproto.ApplicationData{},
		obj2: &edgeproto.ApplicationData{},
		obj3: &edgeproto.ApplicationData{},
		obj4: &edgeproto.ApplicationData{},
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
