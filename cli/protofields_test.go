package cli

import (
	"fmt"
	"testing"

	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	yaml "github.com/mobiledgex/yaml/v2"
	"github.com/stretchr/testify/require"
)

func TestGetSpecifiedFields(t *testing.T) {
	// test App args, make sure false value still sets field
	testGetFieldsArgs(t, &edgeproto.App{},
		[]string{"key.name=foo", "key.version=1", "command=foo", "internalports=false", "officialfqdn=ff"},
		[]string{"2.2", "2.3", "13", "23", "25"})
	// test App args, make sure empty string value still sets field
	testGetFieldsArgs(t, &edgeproto.App{},
		[]string{"key.organization=atlantic", `key.name="Pillimo Go!"`, "imagetype=Docker", `accessports=""`, "defaultflavor.name=x1.small"},
		[]string{"2.1", "2.2", "5", "7", "9.1"})

	dat := `
key:
  organization: AcmeAppCo
  name: someapplication1
  version: "1.0"
imagepath: registry.mobiledgex.net/mobiledgex_AcmeAppCo/someapplication1:1.0
imagetype: Docker
deployment: ""
defaultflavor:
  name: x1.small
accessports: "tcp:80,http:443,udp:10002"
officialfqdn: someapplication1.acmeappco.com
androidpackagename: com.acme.someapplication1
authpublickey: "-----BEGIN PUBLIC KEY-----\nsomekey\n-----END PUBLIC KEY-----\n"
`
	testGetFieldsYaml(t, &edgeproto.App{}, dat,
		[]string{"2.1", "2.2", "2.3", "4", "5", "15", "9.1", "7", "25", "18", "12"})
}

func testGetFieldsArgs(t *testing.T, obj interface{}, args []string, expected []string) {
	// test GetSpecifiedFields
	input := Input{
		DecodeHook: edgeproto.EnumDecodeHook,
	}
	dat, err := input.ParseArgs(args, obj)
	require.Nil(t, err, "parse args %v", args)
	fmt.Printf("argsmap: %v\n", dat)

	fields := GetSpecifiedFields(dat, obj)
	require.ElementsMatch(t, expected, fields, "fields list should match")

	// test GetSpecifiedFieldsData (kind of the opposite of above)
	dmap, err := GetStructMap(obj, WithStructMapFieldFlags(expected))
	require.Nil(t, err, "GetStructMap fields %v", expected)
	genArgs, err := MarshalArgs(dmap, nil, nil)
	require.Nil(t, err, "MarshalArgs for %v", dmap)
	require.ElementsMatch(t, args, genArgs, "args should match")
}

func testGetFieldsYaml(t *testing.T, obj interface{}, data string, expected []string) {
	in := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(data), &in)
	require.Nil(t, err, "unmarshal yaml data %s", data)
	fmt.Printf("yamlmap: %v\n", in)

	inMap := &MapData{
		Namespace: YamlNamespace,
		Data:      in,
	}
	fields := GetSpecifiedFields(inMap, obj)
	require.ElementsMatch(t, expected, fields, "fields list should match")
}
