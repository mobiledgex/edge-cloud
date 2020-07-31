// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: version.proto

package edgeproto

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/mobiledgex/edge-cloud/protogen"

import "github.com/mobiledgex/edge-cloud/util"
import "errors"
import "strconv"
import "encoding/json"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Below enum lists hashes as well as corresponding versions
type VersionHash int32

const (
	VersionHash_HASH_d41d8cd98f00b204e9800998ecf8427e VersionHash = 0
	// interim versions deleted
	VersionHash_HASH_d4ca5418a77d22d968ce7a2afc549dfe VersionHash = 9
	VersionHash_HASH_7848d42e3a2eaf36e53bbd3af581b13a VersionHash = 10
	VersionHash_HASH_f31b7a9d7e06f72107e0ab13c708704e VersionHash = 11
	VersionHash_HASH_03fad51f0343d41f617329151f474d2b VersionHash = 12
	VersionHash_HASH_7d32a983fafc3da768e045b1dc4d5f50 VersionHash = 13
	VersionHash_HASH_747c14bdfe2043f09d251568e4a722c6 VersionHash = 14
	VersionHash_HASH_c7fb20f545a5bc9869b00bb770753c31 VersionHash = 15
	VersionHash_HASH_83cd5c44b5c7387ebf7d055e7345ab42 VersionHash = 16
	VersionHash_HASH_d8a4e697d0d693479cfd9c1c523d7e06 VersionHash = 17
	VersionHash_HASH_e8360aa30f234ecefdfdb9fb2dc79c20 VersionHash = 18
	VersionHash_HASH_c53c7840d242efc7209549a36fcf9e04 VersionHash = 19
)

var VersionHash_name = map[int32]string{
	0:  "HASH_d41d8cd98f00b204e9800998ecf8427e",
	9:  "HASH_d4ca5418a77d22d968ce7a2afc549dfe",
	10: "HASH_7848d42e3a2eaf36e53bbd3af581b13a",
	11: "HASH_f31b7a9d7e06f72107e0ab13c708704e",
	12: "HASH_03fad51f0343d41f617329151f474d2b",
	13: "HASH_7d32a983fafc3da768e045b1dc4d5f50",
	14: "HASH_747c14bdfe2043f09d251568e4a722c6",
	15: "HASH_c7fb20f545a5bc9869b00bb770753c31",
	16: "HASH_83cd5c44b5c7387ebf7d055e7345ab42",
	17: "HASH_d8a4e697d0d693479cfd9c1c523d7e06",
	18: "HASH_e8360aa30f234ecefdfdb9fb2dc79c20",
	19: "HASH_c53c7840d242efc7209549a36fcf9e04",
}
var VersionHash_value = map[string]int32{
	"HASH_d41d8cd98f00b204e9800998ecf8427e": 0,
	"HASH_d4ca5418a77d22d968ce7a2afc549dfe": 9,
	"HASH_7848d42e3a2eaf36e53bbd3af581b13a": 10,
	"HASH_f31b7a9d7e06f72107e0ab13c708704e": 11,
	"HASH_03fad51f0343d41f617329151f474d2b": 12,
	"HASH_7d32a983fafc3da768e045b1dc4d5f50": 13,
	"HASH_747c14bdfe2043f09d251568e4a722c6": 14,
	"HASH_c7fb20f545a5bc9869b00bb770753c31": 15,
	"HASH_83cd5c44b5c7387ebf7d055e7345ab42": 16,
	"HASH_d8a4e697d0d693479cfd9c1c523d7e06": 17,
	"HASH_e8360aa30f234ecefdfdb9fb2dc79c20": 18,
	"HASH_c53c7840d242efc7209549a36fcf9e04": 19,
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
	"HASH_d4ca5418a77d22d968ce7a2afc549dfe",
	"HASH_7848d42e3a2eaf36e53bbd3af581b13a",
	"HASH_f31b7a9d7e06f72107e0ab13c708704e",
	"HASH_03fad51f0343d41f617329151f474d2b",
	"HASH_7d32a983fafc3da768e045b1dc4d5f50",
	"HASH_747c14bdfe2043f09d251568e4a722c6",
	"HASH_c7fb20f545a5bc9869b00bb770753c31",
	"HASH_83cd5c44b5c7387ebf7d055e7345ab42",
	"HASH_d8a4e697d0d693479cfd9c1c523d7e06",
	"HASH_e8360aa30f234ecefdfdb9fb2dc79c20",
	"HASH_c53c7840d242efc7209549a36fcf9e04",
}

const (
	VersionHashHASHD41D8Cd98F00B204E9800998Ecf8427E  uint64 = 1 << 0
	VersionHashHASHD4Ca5418A77D22D968Ce7A2Afc549Dfe  uint64 = 1 << 1
	VersionHashHASH_7848D42E3A2Eaf36E53Bbd3Af581B13A uint64 = 1 << 2
	VersionHashHASHF31B7A9D7E06F72107E0Ab13C708704E  uint64 = 1 << 3
	VersionHashHASH_03Fad51F0343D41F617329151F474D2B uint64 = 1 << 4
	VersionHashHASH_7D32A983Fafc3Da768E045B1Dc4D5F50 uint64 = 1 << 5
	VersionHashHASH_747C14Bdfe2043F09D251568E4A722C6 uint64 = 1 << 6
	VersionHashHASHC7Fb20F545A5Bc9869B00Bb770753C31  uint64 = 1 << 7
	VersionHashHASH_83Cd5C44B5C7387Ebf7D055E7345Ab42 uint64 = 1 << 8
	VersionHashHASHD8A4E697D0D693479Cfd9C1C523D7E06  uint64 = 1 << 9
	VersionHashHASHE8360Aa30F234Ecefdfdb9Fb2Dc79C20  uint64 = 1 << 10
	VersionHashHASHC53C7840D242Efc7209549A36Fcf9E04  uint64 = 1 << 11
)

var VersionHash_CamelName = map[int32]string{
	// HASH_d41d8cd98f00b204e9800998ecf8427e -> HashD41D8Cd98F00B204E9800998Ecf8427E
	0: "HashD41D8Cd98F00B204E9800998Ecf8427E",
	// HASH_d4ca5418a77d22d968ce7a2afc549dfe -> HashD4Ca5418A77D22D968Ce7A2Afc549Dfe
	9: "HashD4Ca5418A77D22D968Ce7A2Afc549Dfe",
	// HASH_7848d42e3a2eaf36e53bbd3af581b13a -> Hash7848D42E3A2Eaf36E53Bbd3Af581B13A
	10: "Hash7848D42E3A2Eaf36E53Bbd3Af581B13A",
	// HASH_f31b7a9d7e06f72107e0ab13c708704e -> HashF31B7A9D7E06F72107E0Ab13C708704E
	11: "HashF31B7A9D7E06F72107E0Ab13C708704E",
	// HASH_03fad51f0343d41f617329151f474d2b -> Hash03Fad51F0343D41F617329151F474D2B
	12: "Hash03Fad51F0343D41F617329151F474D2B",
	// HASH_7d32a983fafc3da768e045b1dc4d5f50 -> Hash7D32A983Fafc3Da768E045B1Dc4D5F50
	13: "Hash7D32A983Fafc3Da768E045B1Dc4D5F50",
	// HASH_747c14bdfe2043f09d251568e4a722c6 -> Hash747C14Bdfe2043F09D251568E4A722C6
	14: "Hash747C14Bdfe2043F09D251568E4A722C6",
	// HASH_c7fb20f545a5bc9869b00bb770753c31 -> HashC7Fb20F545A5Bc9869B00Bb770753C31
	15: "HashC7Fb20F545A5Bc9869B00Bb770753C31",
	// HASH_83cd5c44b5c7387ebf7d055e7345ab42 -> Hash83Cd5C44B5C7387Ebf7D055E7345Ab42
	16: "Hash83Cd5C44B5C7387Ebf7D055E7345Ab42",
	// HASH_d8a4e697d0d693479cfd9c1c523d7e06 -> HashD8A4E697D0D693479Cfd9C1C523D7E06
	17: "HashD8A4E697D0D693479Cfd9C1C523D7E06",
	// HASH_e8360aa30f234ecefdfdb9fb2dc79c20 -> HashE8360Aa30F234Ecefdfdb9Fb2Dc79C20
	18: "HashE8360Aa30F234Ecefdfdb9Fb2Dc79C20",
	// HASH_c53c7840d242efc7209549a36fcf9e04 -> HashC53C7840D242Efc7209549A36Fcf9E04
	19: "HashC53C7840D242Efc7209549A36Fcf9E04",
}
var VersionHash_CamelValue = map[string]int32{
	"HashD41D8Cd98F00B204E9800998Ecf8427E": 0,
	"HashD4Ca5418A77D22D968Ce7A2Afc549Dfe": 9,
	"Hash7848D42E3A2Eaf36E53Bbd3Af581B13A": 10,
	"HashF31B7A9D7E06F72107E0Ab13C708704E": 11,
	"Hash03Fad51F0343D41F617329151F474D2B": 12,
	"Hash7D32A983Fafc3Da768E045B1Dc4D5F50": 13,
	"Hash747C14Bdfe2043F09D251568E4A722C6": 14,
	"HashC7Fb20F545A5Bc9869B00Bb770753C31": 15,
	"Hash83Cd5C44B5C7387Ebf7D055E7345Ab42": 16,
	"HashD8A4E697D0D693479Cfd9C1C523D7E06": 17,
	"HashE8360Aa30F234Ecefdfdb9Fb2Dc79C20": 18,
	"HashC53C7840D242Efc7209549A36Fcf9E04": 19,
}

func (e *VersionHash) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	val, ok := VersionHash_CamelValue[util.CamelCase(str)]
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = VersionHash_CamelName[val]
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("No enum value for %s", str))
	}
	*e = VersionHash(val)
	return nil
}

func (e VersionHash) MarshalYAML() (interface{}, error) {
	return proto.EnumName(VersionHash_CamelName, int32(e)), nil
}

// custom JSON encoding/decoding
func (e *VersionHash) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		val, ok := VersionHash_CamelValue[util.CamelCase(str)]
		if !ok {
			// may be int value instead of enum name
			ival, err := strconv.Atoi(str)
			val = int32(ival)
			if err == nil {
				_, ok = VersionHash_CamelName[val]
			}
		}
		if !ok {
			return errors.New(fmt.Sprintf("No enum value for %s", str))
		}
		*e = VersionHash(val)
		return nil
	}
	var val int32
	err = json.Unmarshal(b, &val)
	if err == nil {
		*e = VersionHash(val)
		return nil
	}
	return fmt.Errorf("No enum value for %v", b)
}

// Keys being hashed:
// AppInstKey
// AppKey
// CloudletKey
// CloudletPoolKey
// ClusterInstKey
// ClusterKey
// ControllerKey
// DeviceKey
// FlavorKey
// NodeKey
// PolicyKey
// ResTagTableKey
// VMPoolKey
var versionHashString = "c53c7840d242efc7209549a36fcf9e04"

func GetDataModelVersion() string {
	return versionHashString
}

var VersionHash_UpgradeFuncs = map[int32]VersionUpgradeFunc{
	0:  nil,
	9:  nil,
	10: SetDefaultLoadBalancerMaxPortRange,
	11: nil,
	12: SetDefaultMaxTrackedDmeClients,
	13: nil,
	14: OrgRestructure,
	15: nil,
	16: AppRevision,
	17: AppInstRefsUpgrade,
	18: nil,
	19: RedoCloudletPools,
}
var VersionHash_UpgradeFuncNames = map[int32]string{
	0:  "",
	9:  "",
	10: "SetDefaultLoadBalancerMaxPortRange",
	11: "",
	12: "SetDefaultMaxTrackedDmeClients",
	13: "",
	14: "OrgRestructure",
	15: "",
	16: "AppRevision",
	17: "AppInstRefsUpgrade",
	18: "",
	19: "RedoCloudletPools",
}

func init() { proto.RegisterFile("version.proto", fileDescriptorVersion) }

var fileDescriptorVersion = []byte{
	// 605 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x92, 0xcf, 0x6a, 0x5b, 0x47,
	0x14, 0xc6, 0xad, 0x45, 0x0b, 0x96, 0x6b, 0xfb, 0xfa, 0x96, 0x96, 0xcb, 0x5d, 0x68, 0x61, 0x68,
	0xa1, 0x85, 0xda, 0xf3, 0xf7, 0xce, 0x4c, 0xa1, 0x50, 0xd9, 0x5e, 0xb8, 0x50, 0x13, 0x47, 0x4e,
	0xb2, 0x0d, 0x67, 0xe6, 0x9c, 0x23, 0x8b, 0xc8, 0xba, 0xe2, 0xea, 0xca, 0x78, 0x9d, 0xa5, 0x9e,
	0x29, 0x0f, 0xe0, 0x65, 0x1e, 0x21, 0xf1, 0x33, 0x44, 0x90, 0x65, 0x90, 0xfc, 0x2f, 0x9b, 0xe1,
	0xc0, 0xfc, 0x38, 0xdf, 0x7c, 0xf3, 0x7d, 0xdd, 0xed, 0x6b, 0x6a, 0x66, 0xa3, 0x7a, 0x72, 0x30,
	0x6d, 0xea, 0xb6, 0xce, 0x37, 0x09, 0x87, 0xb4, 0x1e, 0x4b, 0x3f, 0x1c, 0xb5, 0x97, 0xf3, 0x78,
	0x90, 0xea, 0xab, 0xc3, 0xab, 0x3a, 0x8e, 0xc6, 0xab, 0xab, 0x9b, 0xc3, 0xd5, 0xf9, 0x57, 0x1a,
	0xd7, 0x73, 0x3c, 0x5c, 0x73, 0x43, 0x9a, 0x3c, 0x0d, 0xf7, 0x4b, 0xfe, 0xfc, 0xf0, 0x43, 0x77,
	0xeb, 0xcd, 0xfd, 0xda, 0x53, 0x98, 0x5d, 0xe6, 0x7f, 0x74, 0x7f, 0x3b, 0xed, 0x5f, 0x9c, 0xbe,
	0x45, 0x23, 0xd1, 0x27, 0x0c, 0x9e, 0x85, 0x88, 0x4a, 0x18, 0x0a, 0x5e, 0x88, 0x10, 0x3c, 0x25,
	0xf6, 0x46, 0x39, 0xca, 0x36, 0xbe, 0x43, 0x13, 0x58, 0x23, 0x3d, 0x38, 0x87, 0x4a, 0x61, 0xa8,
	0x7c, 0x22, 0x07, 0x0a, 0x38, 0x59, 0x13, 0x90, 0x29, 0xdb, 0xcc, 0x5f, 0x3e, 0xa0, 0xce, 0x1b,
	0x8f, 0x46, 0x91, 0x06, 0x45, 0xc0, 0xba, 0x22, 0xab, 0x63, 0x44, 0x0d, 0x6c, 0xbd, 0x8c, 0x52,
	0x43, 0xd6, 0x2d, 0x7f, 0x5f, 0x2c, 0x8b, 0xfd, 0x0b, 0x6a, 0x4f, 0x88, 0x61, 0x3e, 0x6e, 0xff,
	0xaf, 0x01, 0x8f, 0x60, 0x0c, 0x93, 0x44, 0xcd, 0x19, 0xdc, 0x9c, 0xd7, 0x4d, 0x3b, 0x80, 0xc9,
	0x90, 0x9e, 0xd4, 0x59, 0xcb, 0xe8, 0x20, 0xa0, 0x23, 0x51, 0xb1, 0x53, 0x52, 0x38, 0x12, 0x10,
	0xa5, 0x4e, 0x4e, 0x78, 0x27, 0x0c, 0x65, 0x5b, 0xf9, 0xd9, 0x03, 0x2a, 0x34, 0x03, 0x5a, 0xc9,
	0x42, 0x1b, 0x8d, 0x46, 0x72, 0x25, 0x9d, 0x56, 0x41, 0x5a, 0xc9, 0xc6, 0x19, 0x54, 0x31, 0xfb,
	0xa9, 0xdc, 0x5f, 0x2c, 0x8b, 0xde, 0xb3, 0xfa, 0x19, 0xdc, 0xbc, 0x6a, 0x20, 0xbd, 0x23, 0x3c,
	0xb9, 0xa2, 0xe3, 0xf1, 0x88, 0x26, 0xed, 0xec, 0x49, 0xd9, 0xa1, 0x56, 0x10, 0xbc, 0x66, 0xe0,
	0xa4, 0x11, 0x5c, 0xe5, 0x49, 0x18, 0x1b, 0x25, 0x26, 0x83, 0x96, 0xad, 0xc8, 0xb6, 0xf3, 0x7f,
	0x1e, 0x51, 0xe3, 0x92, 0x34, 0x11, 0x99, 0x94, 0x30, 0x9a, 0x45, 0x40, 0x65, 0xa5, 0xad, 0x3c,
	0x19, 0x70, 0x4a, 0xa5, 0x2a, 0xdb, 0x29, 0xf3, 0xc5, 0xb2, 0xd8, 0x79, 0xd1, 0x0c, 0x07, 0x34,
	0x6b, 0x9b, 0x79, 0x6a, 0xe7, 0xcd, 0xb3, 0xc7, 0xe4, 0x38, 0x2a, 0xc1, 0xd6, 0x58, 0xb0, 0x31,
	0x05, 0x5f, 0x85, 0x28, 0x44, 0x8c, 0xce, 0x09, 0x67, 0x75, 0xd2, 0x32, 0xdb, 0xcd, 0xff, 0x7e,
	0x40, 0xbd, 0x4e, 0x68, 0x93, 0x31, 0xd1, 0x26, 0xa7, 0xbd, 0xa3, 0xc8, 0x0e, 0x85, 0xb5, 0xe4,
	0xb4, 0xb1, 0x10, 0x8d, 0xca, 0xb2, 0x72, 0x77, 0xb1, 0x2c, 0xb6, 0xfa, 0xd3, 0xe9, 0x80, 0xae,
	0x47, 0xab, 0xdc, 0xf3, 0xfe, 0x63, 0x90, 0x1e, 0x0c, 0x55, 0xc1, 0xa1, 0xc0, 0x2a, 0x68, 0xe3,
	0x42, 0x62, 0x0c, 0x49, 0x26, 0xab, 0xf4, 0xfa, 0x7b, 0xb3, 0xbd, 0xf2, 0xd7, 0xc5, 0xb2, 0xc8,
	0xfb, 0xd3, 0xe9, 0x7f, 0x93, 0x59, 0x3b, 0x20, 0x9e, 0xbd, 0x9e, 0x0e, 0x1b, 0xc0, 0xe7, 0x97,
	0x92, 0xd7, 0x95, 0x00, 0xd0, 0x82, 0x95, 0x36, 0x94, 0x88, 0x91, 0x31, 0x06, 0x8e, 0x0a, 0x93,
	0x0b, 0x49, 0x89, 0x2c, 0xcf, 0xff, 0x7d, 0x34, 0x65, 0x75, 0x72, 0xde, 0x08, 0x54, 0x46, 0x11,
	0x27, 0xa7, 0x44, 0xb0, 0x26, 0x80, 0xae, 0x38, 0x71, 0x20, 0x61, 0xb2, 0x9f, 0xcb, 0x5f, 0x16,
	0xcb, 0x62, 0x6f, 0x40, 0x58, 0x1f, 0xaf, 0x3a, 0x3c, 0xa6, 0xf6, 0xbc, 0xae, 0xc7, 0xb3, 0x72,
	0xf3, 0xeb, 0x97, 0xa2, 0xf3, 0x7e, 0x59, 0x74, 0xd4, 0x51, 0x76, 0xfb, 0xb9, 0xb7, 0x71, 0x7b,
	0xd7, 0xeb, 0x7c, 0xbc, 0xeb, 0x75, 0x3e, 0xdd, 0xf5, 0x3a, 0xf1, 0xc7, 0x75, 0xaf, 0xf5, 0xb7,
	0x00, 0x00, 0x00, 0xff, 0xff, 0x6c, 0x4c, 0xbf, 0xa6, 0x2d, 0x03, 0x00, 0x00,
}
