// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: infraresources.proto

package gencmd

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT
var VmInfoRequiredArgs = []string{}
var VmInfoOptionalArgs = []string{
	"name",
	"ipaddresses",
}
var VmInfoAliasArgs = []string{}
var VmInfoComments = map[string]string{}
var VmInfoSpecialArgs = map[string]string{
	"ipaddresses": "StringArray",
}
var InfraResourcesRequiredArgs = []string{}
var InfraResourcesOptionalArgs = []string{
	"vms:#.name",
	"vms:#.ipaddresses",
}
var InfraResourcesAliasArgs = []string{}
var InfraResourcesComments = map[string]string{}
var InfraResourcesSpecialArgs = map[string]string{
	"vms:#.ipaddresses": "StringArray",
}
