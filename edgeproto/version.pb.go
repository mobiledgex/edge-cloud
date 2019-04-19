// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: version.proto

package edgeproto

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/protocmd"

import "errors"
import "strconv"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Below enum lists hashes as well as corresponding versions
type VersionHash int32

const (
	VersionHash_HASH_d41d8cd98f00b204e9800998ecf8427e VersionHash = 0
)

var VersionHash_name = map[int32]string{
	0: "HASH_d41d8cd98f00b204e9800998ecf8427e",
}
var VersionHash_value = map[string]int32{
	"HASH_d41d8cd98f00b204e9800998ecf8427e": 0,
}

func (x VersionHash) String() string {
	return proto.EnumName(VersionHash_name, int32(x))
}
func (VersionHash) EnumDescriptor() ([]byte, []int) { return fileDescriptorVersion, []int{0} }

func init() {
	proto.RegisterEnum("edgeproto.VersionHash", VersionHash_name, VersionHash_value)
}

var VersionHashStrings = []string{
	"HASH_d41d8cd98f00b204e9800998ecf8427e",
}

const (
	VersionHashHASHD41D8Cd98F00B204E9800998Ecf8427E uint64 = 1 << 0
)

func (e *VersionHash) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	val, ok := VersionHash_value[str]
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = VersionHash_name[val]
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("No enum value for %s", str))
	}
	*e = VersionHash(val)
	return nil
}

func (e VersionHash) MarshalYAML() (interface{}, error) {
	return e.String(), nil
}

// Keys being hashed:
// AppInstKey
// AppKey
// CloudletKey
// ClusterFlavorKey
// ClusterInstKey
// ClusterKey
// ControllerKey
// DeveloperKey
// FlavorKey
// NodeKey
// OperatorKey
var versionHashString = "d41d8cd98f00b204e9800998ecf8427e"

func GetDataModelVersion() string {
	return versionHashString
}

func init() { proto.RegisterFile("version.proto", fileDescriptorVersion) }

var fileDescriptorVersion = []byte{
	// 218 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2d, 0x4b, 0x2d, 0x2a,
	0xce, 0xcc, 0xcf, 0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x4c, 0x4d, 0x49, 0x4f, 0x05,
	0x33, 0xa5, 0x64, 0xd2, 0xf3, 0xf3, 0xd3, 0x73, 0x52, 0xf5, 0x13, 0x0b, 0x32, 0xf5, 0x13, 0xf3,
	0xf2, 0xf2, 0x4b, 0x12, 0x4b, 0x32, 0xf3, 0xf3, 0x8a, 0x21, 0x0a, 0xa5, 0x2c, 0xd2, 0x33, 0x4b,
	0x32, 0x4a, 0x93, 0xf4, 0x92, 0xf3, 0x73, 0xf5, 0x73, 0xf3, 0x93, 0x32, 0x73, 0x40, 0x1a, 0x2b,
	0xf4, 0x41, 0xa4, 0x6e, 0x72, 0x4e, 0x7e, 0x69, 0x8a, 0x3e, 0x58, 0x5d, 0x7a, 0x6a, 0x1e, 0x9c,
	0x01, 0xd5, 0xe9, 0x4e, 0x9c, 0xce, 0x64, 0xdd, 0xf4, 0xd4, 0x3c, 0xdd, 0xe4, 0x5c, 0x18, 0x17,
	0x89, 0x01, 0x31, 0x48, 0xcb, 0x89, 0x8b, 0x3b, 0x0c, 0xe2, 0x78, 0x8f, 0xc4, 0xe2, 0x0c, 0x21,
	0x4d, 0x2e, 0x55, 0x0f, 0xc7, 0x60, 0x8f, 0xf8, 0x14, 0x13, 0xc3, 0x14, 0x8b, 0xe4, 0x14, 0x4b,
	0x8b, 0x34, 0x03, 0x83, 0x24, 0x23, 0x03, 0x93, 0x54, 0x4b, 0x0b, 0x03, 0x03, 0x4b, 0x4b, 0x8b,
	0xd4, 0xe4, 0x34, 0x0b, 0x13, 0x23, 0xf3, 0x54, 0x01, 0x06, 0x29, 0x8e, 0x1f, 0x5f, 0x24, 0x18,
	0x9b, 0xbe, 0x4a, 0x30, 0x38, 0x09, 0x9c, 0x78, 0x28, 0xc7, 0x70, 0xe2, 0x91, 0x1c, 0xe3, 0x85,
	0x47, 0x72, 0x8c, 0x0f, 0x1e, 0xc9, 0x31, 0x26, 0xb1, 0x81, 0x0d, 0x37, 0x06, 0x04, 0x00, 0x00,
	0xff, 0xff, 0xdc, 0xa2, 0x16, 0x25, 0x19, 0x01, 0x00, 0x00,
}
