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
	VersionHash_HASH_d0e4bef9e3b9df9706bdf22ca21b7f10 VersionHash = 1
)

var VersionHash_name = map[int32]string{
	0: "HASH_d41d8cd98f00b204e9800998ecf8427e",
	1: "HASH_d0e4bef9e3b9df9706bdf22ca21b7f10",
}
var VersionHash_value = map[string]int32{
	"HASH_d41d8cd98f00b204e9800998ecf8427e": 0,
	"HASH_d0e4bef9e3b9df9706bdf22ca21b7f10": 1,
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
	"HASH_d0e4bef9e3b9df9706bdf22ca21b7f10",
}

const (
	VersionHashHASHD41D8Cd98F00B204E9800998Ecf8427E uint64 = 1 << 0
	VersionHashHASHD0E4Bef9E3B9Df9706Bdf22Ca21B7F10 uint64 = 1 << 1
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
var versionHashString = "d0e4bef9e3b9df9706bdf22ca21b7f10"

func GetDataModelVersion() string {
	return versionHashString
}

var VersionHash_UpgradeFuncs = map[int32]string{
	0: "",
	1: "mex_salt_ugprade",
}

func init() { proto.RegisterFile("version.proto", fileDescriptorVersion) }

var fileDescriptorVersion = []byte{
	// 276 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x8e, 0x3d, 0x4b, 0xc3, 0x40,
	0x1c, 0x87, 0x7b, 0x08, 0x82, 0x2d, 0x42, 0x28, 0x0e, 0x25, 0x48, 0x36, 0x07, 0x85, 0x34, 0x97,
	0x34, 0x98, 0xbb, 0x49, 0x74, 0x31, 0xbb, 0xe0, 0x1a, 0xee, 0xe5, 0x7f, 0xd7, 0x40, 0x72, 0x17,
	0xf2, 0x22, 0x9d, 0x1d, 0x1c, 0xfc, 0x64, 0x1d, 0xfd, 0x08, 0x9a, 0xcf, 0x60, 0xc1, 0x51, 0x9a,
	0x58, 0x71, 0x74, 0x39, 0x9e, 0x83, 0xe7, 0xf9, 0xf3, 0x9b, 0x9e, 0x3e, 0x41, 0xdd, 0xe4, 0xd6,
	0x2c, 0xab, 0xda, 0xb6, 0x76, 0x7e, 0x02, 0x52, 0xc3, 0x80, 0xee, 0xb9, 0xb6, 0x56, 0x17, 0x10,
	0xb0, 0x2a, 0x0f, 0x98, 0x31, 0xb6, 0x65, 0x6d, 0x6e, 0x4d, 0x33, 0x8a, 0x2e, 0xd1, 0x79, 0xbb,
	0xee, 0xf8, 0x52, 0xd8, 0x32, 0x28, 0x2d, 0xcf, 0x8b, 0x7d, 0xb8, 0x09, 0xf6, 0xaf, 0x2f, 0x0a,
	0xdb, 0xc9, 0x60, 0xf0, 0x34, 0x98, 0x5f, 0xf8, 0x29, 0xef, 0xff, 0x57, 0x0a, 0x5f, 0x83, 0xf1,
	0x45, 0x79, 0xf8, 0xfe, 0x81, 0xf1, 0xd0, 0xd5, 0x0b, 0x9a, 0xce, 0x1e, 0xc7, 0xf5, 0x29, 0x6b,
	0xd6, 0xf3, 0xcb, 0xe9, 0x45, 0x7a, 0xfb, 0x90, 0x66, 0x32, 0x0e, 0x25, 0x11, 0x92, 0x12, 0x85,
	0x31, 0x8f, 0x70, 0x0c, 0x94, 0x60, 0x4c, 0x29, 0x01, 0xa1, 0x48, 0x1c, 0x25, 0xe0, 0x4c, 0xe6,
	0x37, 0x07, 0x15, 0x43, 0xcc, 0x41, 0x51, 0x58, 0x71, 0x2a, 0x15, 0x4d, 0xf0, 0x35, 0x97, 0x2a,
	0x8a, 0x04, 0x8b, 0x42, 0x9e, 0xa8, 0x10, 0x3b, 0xc8, 0x3d, 0x7b, 0xdd, 0x2d, 0x9c, 0x12, 0x36,
	0x59, 0xc3, 0x8a, 0x36, 0xeb, 0x74, 0x55, 0x33, 0x09, 0xee, 0xec, 0xeb, 0x73, 0x81, 0x9e, 0x77,
	0x8b, 0x23, 0xc6, 0xc5, 0x9d, 0xb3, 0xfd, 0xf0, 0x26, 0xdb, 0xde, 0x43, 0x6f, 0xbd, 0x87, 0xde,
	0x7b, 0x0f, 0xf1, 0xe3, 0x61, 0xe1, 0xea, 0x3b, 0x00, 0x00, 0xff, 0xff, 0x0e, 0xf8, 0xfa, 0xb7,
	0x5e, 0x01, 0x00, 0x00,
}
