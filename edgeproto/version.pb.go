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
	VersionHash_HASH_a18636af1f4272c38ca72881b2a8bcea VersionHash = 22
	VersionHash_HASH_efbddcee4ba444e3656f64e430a5e3be VersionHash = 23
	VersionHash_HASH_c2c322505017054033953f6104002bf5 VersionHash = 24
	VersionHash_HASH_facc3c3c9c76463c8d8b3c874ce43487 VersionHash = 25
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
	22: "HASH_a18636af1f4272c38ca72881b2a8bcea",
	23: "HASH_efbddcee4ba444e3656f64e430a5e3be",
	24: "HASH_c2c322505017054033953f6104002bf5",
	25: "HASH_facc3c3c9c76463c8d8b3c874ce43487",
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
	"HASH_a18636af1f4272c38ca72881b2a8bcea": 22,
	"HASH_efbddcee4ba444e3656f64e430a5e3be": 23,
	"HASH_c2c322505017054033953f6104002bf5": 24,
	"HASH_facc3c3c9c76463c8d8b3c874ce43487": 25,
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
	// 709 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x93, 0xbf, 0x6f, 0xdc, 0x45,
	0x10, 0xc5, 0xef, 0x1a, 0x90, 0x2f, 0x24, 0x7c, 0x73, 0x24, 0x70, 0x1c, 0xe8, 0xa8, 0x28, 0x40,
	0x22, 0x9e, 0xdd, 0x99, 0xdd, 0xd9, 0xed, 0x80, 0xa0, 0xc8, 0xa2, 0xb2, 0x30, 0xd0, 0xa2, 0xd9,
	0xd9, 0xd9, 0xb3, 0x85, 0x7d, 0x5f, 0xeb, 0x7e, 0x44, 0xd4, 0x34, 0x48, 0xae, 0xf8, 0xb3, 0x52,
	0xa6, 0xa4, 0x04, 0xfb, 0x5f, 0xc0, 0x12, 0x25, 0xf2, 0x39, 0xdc, 0xd1, 0xac, 0x9e, 0x76, 0x3f,
	0x9a, 0x7d, 0x7a, 0x9a, 0x37, 0x7a, 0xf8, 0xd2, 0x96, 0xab, 0xb3, 0x7e, 0xf1, 0xec, 0x72, 0xd9,
	0xaf, 0xfb, 0xf1, 0x81, 0xd5, 0xb9, 0x6d, 0xe5, 0x34, 0xcd, 0xcf, 0xd6, 0xa7, 0x9b, 0xf2, 0x4c,
	0xfb, 0x8b, 0xc3, 0x8b, 0xbe, 0x9c, 0x9d, 0xdf, 0x3d, 0xfd, 0x72, 0x78, 0x77, 0x7e, 0xa1, 0xe7,
	0xfd, 0xa6, 0x1e, 0x6e, 0xb9, 0xb9, 0x2d, 0x76, 0xe2, 0x7e, 0xc8, 0xf4, 0xc9, 0xbc, 0x9f, 0xf7,
	0x5b, 0x79, 0x78, 0xa7, 0xee, 0x6f, 0x3f, 0xff, 0xed, 0xed, 0xd1, 0x83, 0x1f, 0xef, 0x3f, 0x3b,
	0x92, 0xd5, 0xe9, 0xf8, 0xb3, 0xd1, 0xa7, 0x47, 0x5f, 0x9d, 0x1c, 0xfd, 0x54, 0xc9, 0xd5, 0xa4,
	0x35, 0xa7, 0x06, 0x50, 0x3c, 0x90, 0xe5, 0x04, 0x90, 0x73, 0x32, 0x6d, 0x89, 0x3c, 0x5b, 0x37,
	0xf8, 0x1f, 0xaa, 0x12, 0xc8, 0x25, 0x61, 0xae, 0xde, 0xd7, 0x1c, 0x93, 0x1a, 0x8b, 0x97, 0xa6,
	0x81, 0x72, 0x6d, 0xd6, 0x1d, 0xec, 0x50, 0x4e, 0x94, 0x2a, 0x79, 0x43, 0xf1, 0x26, 0x0d, 0xa3,
	0x05, 0x2c, 0xa5, 0xa2, 0xb4, 0x90, 0x5c, 0x71, 0x28, 0xdd, 0x68, 0x87, 0x36, 0x74, 0x85, 0x25,
	0x57, 0x36, 0x88, 0x8d, 0xbd, 0x03, 0x36, 0x90, 0xe2, 0x50, 0x19, 0x12, 0x03, 0x59, 0xf7, 0x60,
	0x87, 0x02, 0x36, 0xa9, 0xc1, 0x35, 0x40, 0xc2, 0x4a, 0xae, 0x45, 0xc7, 0xe8, 0xb3, 0x0b, 0xae,
	0x11, 0x53, 0xf5, 0xa5, 0x7b, 0x67, 0x6f, 0xa0, 0xa2, 0x97, 0x9c, 0xb0, 0x49, 0x53, 0xac, 0xc2,
	0x31, 0x19, 0x50, 0x28, 0xae, 0x2a, 0xd5, 0xd0, 0x02, 0x74, 0x0f, 0xf7, 0x28, 0xb1, 0x3a, 0x2a,
	0xb5, 0x99, 0x07, 0xc2, 0x06, 0xb9, 0xfa, 0xe0, 0x42, 0x4c, 0x46, 0xc2, 0xde, 0x6b, 0xec, 0x1e,
	0xed, 0x50, 0xe5, 0x56, 0x3c, 0xb4, 0x40, 0x41, 0x42, 0xd1, 0x9c, 0x62, 0x2e, 0x00, 0xa5, 0x30,
	0x03, 0x07, 0x54, 0x74, 0xdd, 0xbb, 0x3b, 0x34, 0xa1, 0xd6, 0xa0, 0x44, 0x25, 0x28, 0x63, 0x62,
	0x2b, 0x8d, 0x2b, 0x84, 0x60, 0x8c, 0x14, 0xa4, 0x90, 0xef, 0xba, 0x7d, 0xae, 0x49, 0xc8, 0x62,
	0xe6, 0x0a, 0x35, 0x66, 0x24, 0xce, 0xda, 0x6a, 0x56, 0xa7, 0xc1, 0xe3, 0x36, 0x95, 0xee, 0xf1,
	0x0e, 0xb5, 0x84, 0x11, 0x44, 0x10, 0x9a, 0x47, 0x32, 0xb5, 0x56, 0x5b, 0x2d, 0xb9, 0x15, 0x5f,
	0x95, 0xb3, 0x7a, 0xe8, 0xc6, 0x7b, 0xaf, 0x01, 0x95, 0x13, 0x41, 0xf5, 0xe4, 0xad, 0x29, 0x7b,
	0xc8, 0x81, 0xb2, 0x60, 0x6c, 0xda, 0xb2, 0x01, 0x75, 0xef, 0x8d, 0xbf, 0x7c, 0x83, 0x3a, 0x09,
	0x8c, 0x39, 0xc6, 0x9c, 0x94, 0xa4, 0x9a, 0x0b, 0x0d, 0x02, 0x67, 0xcd, 0x2d, 0xb0, 0x23, 0xad,
	0xb1, 0x7b, 0x32, 0x7d, 0x7a, 0x75, 0x3b, 0x79, 0xfc, 0xfc, 0xd4, 0xf4, 0xe7, 0x17, 0xfd, 0xf2,
	0x68, 0xbd, 0xbe, 0x3c, 0xee, 0x97, 0xeb, 0xd5, 0xf8, 0xdb, 0xff, 0x32, 0x74, 0x1a, 0x12, 0x30,
	0x45, 0x33, 0x2f, 0xb1, 0x70, 0x75, 0x42, 0x2e, 0xf9, 0x82, 0x12, 0x88, 0x80, 0xa5, 0x7b, 0x3a,
	0xfd, 0xe4, 0xea, 0x76, 0xf2, 0xd1, 0xf1, 0x72, 0xb3, 0xb0, 0x13, 0xb9, 0x58, 0x6d, 0x16, 0xf3,
	0xe3, 0x73, 0x59, 0xb7, 0x7e, 0x79, 0xf1, 0x8d, 0xbd, 0x3c, 0x53, 0x5b, 0x8d, 0xf3, 0x9b, 0x59,
	0xe2, 0x52, 0xc4, 0x28, 0xcd, 0x35, 0xf2, 0xec, 0x15, 0x93, 0x0a, 0xfb, 0x94, 0x5c, 0xf1, 0x92,
	0x8a, 0x9a, 0x74, 0xef, 0x4f, 0x1f, 0x5d, 0xdd, 0x4e, 0x46, 0x27, 0xb6, 0xfe, 0x7e, 0xb9, 0x59,
	0xad, 0xad, 0xee, 0xe3, 0x69, 0xa5, 0x56, 0x35, 0xa3, 0x22, 0x44, 0x64, 0x18, 0x43, 0x6c, 0x91,
	0x8c, 0x10, 0x24, 0x18, 0x16, 0xeb, 0x3e, 0xd8, 0x39, 0x56, 0xaf, 0xe8, 0x7d, 0x80, 0x00, 0x8e,
	0x21, 0x10, 0x20, 0xe6, 0x80, 0x2d, 0x3a, 0x20, 0x00, 0x5f, 0x5a, 0xe8, 0x26, 0xf7, 0x8e, 0x9f,
	0xdf, 0x75, 0xed, 0xdc, 0xd6, 0xdf, 0xd9, 0xaa, 0xdf, 0x2c, 0xd5, 0x7e, 0xb8, 0x9c, 0x2f, 0xa5,
	0xda, 0x8b, 0xcd, 0x42, 0xf7, 0x2b, 0x2c, 0xaa, 0xa8, 0xa8, 0x59, 0x39, 0x52, 0x44, 0x4d, 0x35,
	0x15, 0xd4, 0xc4, 0xa4, 0x46, 0x48, 0x89, 0xbb, 0x0f, 0xa7, 0x07, 0xff, 0xfc, 0x3d, 0x19, 0xfe,
	0x7a, 0x3b, 0x19, 0xba, 0xaf, 0x3f, 0x7e, 0xf5, 0xd7, 0x6c, 0xf0, 0xea, 0x7a, 0x36, 0x7c, 0x7d,
	0x3d, 0x1b, 0xfe, 0x79, 0x3d, 0x1b, 0xfe, 0x7e, 0x33, 0x1b, 0xbc, 0xbe, 0x99, 0x0d, 0xfe, 0xb8,
	0x99, 0x0d, 0xca, 0x5b, 0xdb, 0xba, 0xe2, 0xbf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x24, 0xa3, 0xac,
	0x0f, 0x1a, 0x04, 0x00, 0x00,
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
	"HASH_a18636af1f4272c38ca72881b2a8bcea",
	"HASH_efbddcee4ba444e3656f64e430a5e3be",
	"HASH_c2c322505017054033953f6104002bf5",
	"HASH_facc3c3c9c76463c8d8b3c874ce43487",
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
	VersionHashHASHA18636Af1F4272C38Ca72881B2A8Bcea  uint64 = 1 << 14
	VersionHashHASHEfbddcee4Ba444E3656F64E430A5E3Be  uint64 = 1 << 15
	VersionHashHASHC2C322505017054033953F6104002Bf5  uint64 = 1 << 16
	VersionHashHASHFacc3C3C9C76463C8D8B3C874Ce43487  uint64 = 1 << 17
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
	// HASH_a18636af1f4272c38ca72881b2a8bcea -> HashA18636Af1F4272C38Ca72881B2A8Bcea
	22: "HashA18636Af1F4272C38Ca72881B2A8Bcea",
	// HASH_efbddcee4ba444e3656f64e430a5e3be -> HashEfbddcee4Ba444E3656F64E430A5E3Be
	23: "HashEfbddcee4Ba444E3656F64E430A5E3Be",
	// HASH_c2c322505017054033953f6104002bf5 -> HashC2C322505017054033953F6104002Bf5
	24: "HashC2C322505017054033953F6104002Bf5",
	// HASH_facc3c3c9c76463c8d8b3c874ce43487 -> HashFacc3C3C9C76463C8D8B3C874Ce43487
	25: "HashFacc3C3C9C76463C8D8B3C874Ce43487",
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
	"HashA18636Af1F4272C38Ca72881B2A8Bcea": 22,
	"HashEfbddcee4Ba444E3656F64E430A5E3Be": 23,
	"HashC2C322505017054033953F6104002Bf5": 24,
	"HashFacc3C3C9C76463C8D8B3C874Ce43487": 25,
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
// AppInstRefKey
// AppKey
// CloudletKey
// CloudletPoolKey
// ClusterInstKey
// ClusterInstRefKey
// ClusterKey
// ControllerKey
// DeviceKey
// FlavorKey
// GPUDriverKey
// NodeKey
// PolicyKey
// ResTagTableKey
// VMPoolKey
// VirtualClusterInstKey
var versionHashString = "facc3c3c9c76463c8d8b3c874ce43487"

func GetDataModelVersion() string {
	return versionHashString
}

var VersionHash_UpgradeFuncs = map[int32]VersionUpgradeFunc{
	0:  nil,
	9:  nil,
	10: nil,
	11: nil,
	12: nil,
	13: nil,
	14: nil,
	15: nil,
	16: nil,
	17: nil,
	18: nil,
	19: nil,
	20: CheckForHttpPorts,
	21: PruneSamsungPlatformDevices,
	22: SetTrusted,
	23: nil,
	24: CloudletResourceUpgradeFunc,
	25: nil,
}
var VersionHash_UpgradeFuncNames = map[int32]string{
	0:  "",
	9:  "",
	10: "",
	11: "",
	12: "",
	13: "",
	14: "",
	15: "",
	16: "",
	17: "",
	18: "",
	19: "",
	20: "CheckForHttpPorts",
	21: "PruneSamsungPlatformDevices",
	22: "SetTrusted",
	23: "",
	24: "CloudletResourceUpgradeFunc",
	25: "",
}
