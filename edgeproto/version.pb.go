// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: version.proto

package edgeproto

import (
	"encoding/json"
	"errors"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	"github.com/mobiledgex/edge-cloud/util"
	math "math"
	"strconv"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Below enum lists hashes as well as corresponding versions
type VersionHash int32

const (
	VersionHash_HASH_d41d8cd98f00b204e9800998ecf8427e VersionHash = 0
	//interim versions deleted
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
	VersionHash_HASH_1a57396698c4ade15f0579c9f5714cd6 VersionHash = 20
	VersionHash_HASH_71c580746ee2a6b7d1a4182b3a54407a VersionHash = 21
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
	20: "HASH_1a57396698c4ade15f0579c9f5714cd6",
	21: "HASH_71c580746ee2a6b7d1a4182b3a54407a",
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
	"HASH_1a57396698c4ade15f0579c9f5714cd6": 20,
	"HASH_71c580746ee2a6b7d1a4182b3a54407a": 21,
}

func (x VersionHash) String() string {
	return proto.EnumName(VersionHash_name, int32(x))
}

func (VersionHash) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_7d2c07d79758f814, []int{0}
}

func init() {
	proto.RegisterEnum("edgeproto.VersionHash", VersionHash_name, VersionHash_value)
}

func init() { proto.RegisterFile("version.proto", fileDescriptor_7d2c07d79758f814) }

var fileDescriptor_7d2c07d79758f814 = []byte{
	// 721 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x93, 0xcf, 0x6b, 0x24, 0x45,
	0x14, 0xc7, 0x67, 0x2e, 0x0b, 0x49, 0xdc, 0xdd, 0xde, 0x76, 0x57, 0x9a, 0x56, 0x46, 0x58, 0x50,
	0x50, 0x70, 0x53, 0x3f, 0xbb, 0xaa, 0x04, 0xc1, 0x6c, 0x82, 0x44, 0x31, 0x18, 0x27, 0xea, 0x55,
	0x5e, 0xbd, 0xf7, 0xaa, 0x33, 0xec, 0xcc, 0xf4, 0xd0, 0xdd, 0x13, 0x72, 0xf6, 0x38, 0x27, 0xff,
	0xac, 0xc5, 0xd3, 0x1e, 0x3d, 0x6a, 0xf2, 0x2f, 0x38, 0xe0, 0x51, 0x66, 0xf2, 0xcb, 0x4b, 0xf1,
	0xa5, 0xea, 0xc3, 0xfb, 0xd6, 0xab, 0xef, 0xab, 0xdd, 0xc7, 0x17, 0xdc, 0x76, 0x93, 0x66, 0xfe,
	0x6a, 0xd1, 0x36, 0x7d, 0x93, 0xef, 0x30, 0xd5, 0xbc, 0x95, 0xa5, 0xaf, 0x27, 0xfd, 0xf9, 0x32,
	0xbe, 0xc2, 0x66, 0xb6, 0x3f, 0x6b, 0xe2, 0x64, 0xba, 0x39, 0xba, 0xdc, 0xdf, 0xac, 0x5f, 0xe0,
	0xb4, 0x59, 0xd2, 0xfe, 0x96, 0xab, 0x79, 0x7e, 0x2f, 0x6e, 0x8a, 0x94, 0xcf, 0xeb, 0xa6, 0x6e,
	0xb6, 0x72, 0x7f, 0xa3, 0x6e, 0x76, 0x3f, 0xff, 0xe3, 0xd1, 0xee, 0xde, 0x2f, 0x37, 0x66, 0xc7,
	0xd0, 0x9d, 0xe7, 0x9f, 0xed, 0x7e, 0x72, 0x7c, 0x70, 0x76, 0xfc, 0x2b, 0x19, 0x49, 0x1e, 0x29,
	0xf8, 0x24, 0x44, 0x54, 0xc2, 0x70, 0xf0, 0x42, 0x84, 0xe0, 0x19, 0x93, 0x37, 0xca, 0x71, 0x36,
	0xf8, 0x1f, 0x8a, 0x60, 0x8d, 0xf4, 0xe0, 0x1c, 0x29, 0x45, 0xa1, 0xf2, 0xc8, 0x0e, 0x14, 0x24,
	0xb4, 0x26, 0x50, 0xe2, 0x6c, 0x27, 0xff, 0xf1, 0x16, 0x75, 0xde, 0x78, 0x32, 0x8a, 0x35, 0x28,
	0x86, 0xa4, 0x2b, 0xb6, 0x3a, 0x46, 0xd2, 0x90, 0xac, 0x97, 0x51, 0x6a, 0xc8, 0x76, 0xcb, 0x4f,
	0x57, 0xeb, 0xe2, 0xe5, 0x19, 0xf7, 0x47, 0x9c, 0x60, 0x39, 0xed, 0xbf, 0x6f, 0x80, 0x5e, 0xc3,
	0x14, 0xe6, 0xc8, 0xed, 0x09, 0x5c, 0x9e, 0x36, 0x6d, 0x3f, 0x86, 0x79, 0xcd, 0xf7, 0xee, 0x49,
	0xcb, 0xe8, 0x20, 0x90, 0x63, 0x51, 0x25, 0xa7, 0xa4, 0x70, 0x2c, 0x20, 0x4a, 0x8d, 0x4e, 0x78,
	0x27, 0x0c, 0x67, 0x7b, 0xf9, 0xc9, 0x2d, 0x2a, 0x74, 0x02, 0xb2, 0x32, 0x09, 0x6d, 0x34, 0x19,
	0x99, 0x2a, 0xe9, 0xb4, 0x0a, 0xd2, 0xca, 0x64, 0x9c, 0x21, 0x15, 0xb3, 0xf7, 0xca, 0x97, 0xab,
	0x75, 0x31, 0x7a, 0x70, 0x3f, 0x81, 0xcb, 0x9f, 0x5a, 0xc0, 0x37, 0x4c, 0x47, 0x33, 0x3e, 0x9c,
	0x4e, 0x78, 0xde, 0x77, 0xf7, 0xce, 0x8e, 0xb4, 0x82, 0xe0, 0x75, 0x82, 0x84, 0x9a, 0xc0, 0x55,
	0x9e, 0x85, 0xb1, 0x51, 0x12, 0x1a, 0xb2, 0xc9, 0x8a, 0xec, 0x71, 0xfe, 0xd5, 0x1d, 0x6a, 0x1c,
	0x4a, 0x13, 0x29, 0xb1, 0x12, 0x46, 0x27, 0x11, 0x48, 0x59, 0x69, 0x2b, 0xcf, 0x06, 0x9c, 0x52,
	0x58, 0x65, 0x4f, 0xca, 0x7c, 0xb5, 0x2e, 0x9e, 0xfc, 0xd0, 0xd6, 0x63, 0xee, 0xfa, 0x76, 0x89,
	0xfd, 0xb2, 0x7d, 0xe8, 0x11, 0x5d, 0x8a, 0x4a, 0x24, 0x6b, 0x2c, 0xd8, 0x88, 0xc1, 0x57, 0x21,
	0x0a, 0x11, 0xa3, 0x73, 0xc2, 0x59, 0x8d, 0x5a, 0x66, 0x4f, 0xf3, 0x2f, 0x6f, 0x51, 0xaf, 0x91,
	0x2c, 0x1a, 0x13, 0x2d, 0x3a, 0xed, 0x1d, 0xc7, 0xe4, 0x48, 0x58, 0xcb, 0x4e, 0x1b, 0x0b, 0xd1,
	0xa8, 0x2c, 0x2b, 0x9f, 0xae, 0xd6, 0xc5, 0xde, 0xc1, 0x62, 0x31, 0xe6, 0x8b, 0xc9, 0x26, 0xf7,
	0xfc, 0xe0, 0x2e, 0x48, 0x0f, 0x86, 0xab, 0xe0, 0x48, 0x50, 0x15, 0xb4, 0x71, 0x01, 0x13, 0x05,
	0x94, 0x68, 0x95, 0xde, 0x3e, 0x6f, 0xf6, 0xac, 0xfc, 0x60, 0xb5, 0x2e, 0xf2, 0x83, 0xc5, 0xe2,
	0xdb, 0x79, 0xd7, 0x8f, 0x39, 0x75, 0x3f, 0x2f, 0xea, 0x16, 0xe8, 0xe1, 0xa6, 0xec, 0x75, 0x25,
	0x00, 0xb4, 0x48, 0x4a, 0x1b, 0x46, 0x4e, 0x94, 0x28, 0x86, 0x14, 0x15, 0xa1, 0x0b, 0xa8, 0x44,
	0x96, 0xe7, 0x5f, 0xdf, 0x35, 0x65, 0x35, 0x3a, 0x6f, 0x04, 0x29, 0xa3, 0x38, 0xa1, 0x53, 0x22,
	0x58, 0x13, 0x40, 0x57, 0x09, 0x53, 0x60, 0x61, 0xb2, 0xf7, 0xcb, 0x17, 0xab, 0x75, 0xf1, 0x6c,
	0xcc, 0xd4, 0x1c, 0x6e, 0x26, 0x7b, 0xca, 0xfd, 0x69, 0xd3, 0x4c, 0xbb, 0xfb, 0x0a, 0x12, 0xac,
	0xd3, 0xa1, 0xaa, 0x82, 0x47, 0x03, 0xc4, 0xd2, 0x26, 0x61, 0x5d, 0xc0, 0x90, 0xac, 0x93, 0x06,
	0xa9, 0xca, 0x9e, 0xdf, 0x54, 0x38, 0x3c, 0x67, 0x7c, 0xf3, 0x4d, 0xd3, 0x1e, 0xf7, 0xfd, 0x62,
	0x33, 0x40, 0x5d, 0xfe, 0xdd, 0x5d, 0x2e, 0x12, 0xad, 0x17, 0xce, 0x54, 0xcc, 0x0a, 0xaa, 0xe8,
	0x48, 0x82, 0x91, 0x5e, 0x45, 0x0d, 0xd6, 0x18, 0xe1, 0x20, 0x7b, 0x51, 0x7e, 0xbc, 0x5a, 0x17,
	0x1f, 0x9e, 0xb6, 0xcb, 0x39, 0x9f, 0xc1, 0xac, 0x5b, 0xce, 0xeb, 0xd3, 0x29, 0xf4, 0xa9, 0x69,
	0x67, 0x47, 0x7c, 0x31, 0x41, 0xee, 0xca, 0x9d, 0x7f, 0xff, 0x29, 0x86, 0xbf, 0xad, 0x8b, 0xa1,
	0x78, 0xfd, 0xd1, 0xdb, 0xbf, 0x47, 0x83, 0xb7, 0x57, 0xa3, 0xe1, 0xbb, 0xab, 0xd1, 0xf0, 0xaf,
	0xab, 0xd1, 0xf0, 0xf7, 0xeb, 0xd1, 0xe0, 0xdd, 0xf5, 0x68, 0xf0, 0xe7, 0xf5, 0x68, 0x10, 0x1f,
	0x6d, 0x7f, 0x9c, 0xfe, 0x2f, 0x00, 0x00, 0xff, 0xff, 0x04, 0x19, 0x2b, 0xfb, 0xdd, 0x03, 0x00,
	0x00,
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
	"HASH_1a57396698c4ade15f0579c9f5714cd6",
	"HASH_71c580746ee2a6b7d1a4182b3a54407a",
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
	VersionHashHASH_1A57396698C4Ade15F0579C9F5714Cd6 uint64 = 1 << 12
	VersionHashHASH_71C580746Ee2A6B7D1A4182B3A54407A uint64 = 1 << 13
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
	// HASH_1a57396698c4ade15f0579c9f5714cd6 -> Hash1A57396698C4Ade15F0579C9F5714Cd6
	20: "Hash1A57396698C4Ade15F0579C9F5714Cd6",
	// HASH_71c580746ee2a6b7d1a4182b3a54407a -> Hash71C580746Ee2A6B7D1A4182B3A54407A
	21: "Hash71C580746Ee2A6B7D1A4182B3A54407A",
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
	"Hash1A57396698C4Ade15F0579C9F5714Cd6": 20,
	"Hash71C580746Ee2A6B7D1A4182B3A54407A": 21,
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
var versionHashString = "71c580746ee2a6b7d1a4182b3a54407a"

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
	20: CheckForHttpPorts,
	21: PruneSamsungPlatformDevices,
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
	20: "CheckForHttpPorts",
	21: "PruneSamsungPlatformDevices",
}
