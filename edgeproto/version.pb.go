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
	VersionHash_HASH_8ba950479a03ab77edfad426ea53c173 VersionHash = 26
	VersionHash_HASH_f4eb139f7a8373a484ab9749eadc31f5 VersionHash = 27
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
	26: "HASH_8ba950479a03ab77edfad426ea53c173",
	27: "HASH_f4eb139f7a8373a484ab9749eadc31f5",
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
	"HASH_8ba950479a03ab77edfad426ea53c173": 26,
	"HASH_f4eb139f7a8373a484ab9749eadc31f5": 27,
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
	// 752 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x94, 0x3f, 0x6f, 0xe4, 0xb5,
	0x13, 0xc6, 0x93, 0xe6, 0xf7, 0x53, 0x72, 0xdc, 0xe1, 0x5b, 0xee, 0x60, 0xd9, 0x43, 0x4b, 0x45,
	0x01, 0x12, 0x97, 0xb1, 0x67, 0xec, 0xb1, 0x3b, 0xe0, 0xd0, 0x29, 0xa2, 0x8a, 0x08, 0xd0, 0xa2,
	0xf1, 0x78, 0xbc, 0x89, 0x48, 0xf2, 0x8d, 0xf6, 0xcf, 0x89, 0x9a, 0x32, 0x15, 0x2f, 0xeb, 0xca,
	0x2b, 0x69, 0x90, 0x20, 0x79, 0x0b, 0x44, 0xa2, 0x44, 0xd9, 0x1c, 0xbb, 0x34, 0xd6, 0x23, 0xfb,
	0xa3, 0xc7, 0x8f, 0x1e, 0x8d, 0xbd, 0xff, 0xf0, 0x95, 0xcd, 0x17, 0xa7, 0xc3, 0xc5, 0xf3, 0xcb,
	0xf9, 0xb0, 0x1c, 0x46, 0x7b, 0xd6, 0x66, 0xb6, 0x96, 0x93, 0x3c, 0x3b, 0x5d, 0x9e, 0xac, 0xea,
	0x73, 0x1d, 0xce, 0x0f, 0xce, 0x87, 0x7a, 0x7a, 0x76, 0x77, 0xf4, 0xf3, 0xc1, 0xdd, 0xfa, 0xb9,
	0x9e, 0x0d, 0xab, 0x76, 0xb0, 0xe6, 0x66, 0x76, 0xb1, 0x11, 0xf7, 0x26, 0x93, 0x27, 0xb3, 0x61,
	0x36, 0xac, 0xe5, 0xc1, 0x9d, 0xba, 0xdf, 0xfd, 0xec, 0xf7, 0xff, 0xef, 0x3f, 0xf8, 0xe1, 0xfe,
	0xb2, 0x43, 0x59, 0x9c, 0x8c, 0x3e, 0xdd, 0xff, 0xe4, 0xf0, 0xcb, 0xe3, 0xc3, 0x1f, 0x1b, 0xf9,
	0x96, 0xb5, 0x95, 0xdc, 0x01, 0x6a, 0x00, 0xb2, 0x92, 0x01, 0x4a, 0xc9, 0xa6, 0x3d, 0x53, 0x60,
	0x73, 0x3b, 0xff, 0x41, 0x55, 0x22, 0xf9, 0x2c, 0xcc, 0x2d, 0x84, 0x56, 0x52, 0x56, 0x63, 0x09,
	0xd2, 0x35, 0x52, 0x69, 0xdd, 0xdc, 0xde, 0x06, 0xe5, 0x4c, 0xb9, 0x51, 0x30, 0x94, 0x60, 0xd2,
	0x31, 0x59, 0xc4, 0x5a, 0x1b, 0x4a, 0x8f, 0xd9, 0x57, 0x8f, 0xe2, 0xf6, 0x37, 0x68, 0x47, 0x5f,
	0x59, 0x4a, 0x63, 0x83, 0xd4, 0x39, 0x78, 0x60, 0x03, 0xa9, 0x1e, 0x95, 0x21, 0x33, 0x90, 0xb9,
	0x07, 0x1b, 0x14, 0xb0, 0x4b, 0x8b, 0xbe, 0x03, 0x12, 0x36, 0xf2, 0x3d, 0x79, 0xc6, 0x50, 0x7c,
	0xf4, 0x9d, 0x98, 0x5a, 0xa8, 0xee, 0x9d, 0x6d, 0x80, 0x86, 0x41, 0x4a, 0xc6, 0x2e, 0x5d, 0xb1,
	0x09, 0xa7, 0x6c, 0x40, 0xb1, 0xfa, 0xa6, 0xd4, 0x62, 0x8f, 0xe0, 0x1e, 0x6e, 0x51, 0x62, 0xf5,
	0x54, 0x5b, 0xb7, 0x00, 0x84, 0x1d, 0x4a, 0x0b, 0xd1, 0xc7, 0x94, 0x8d, 0x84, 0x43, 0xd0, 0xe4,
	0x1e, 0x6d, 0x50, 0xe5, 0x5e, 0x03, 0xf4, 0x48, 0x51, 0x62, 0xd5, 0x92, 0x53, 0xa9, 0x00, 0xb5,
	0x32, 0x03, 0x47, 0x54, 0xf4, 0xee, 0xdd, 0x0d, 0x9a, 0x51, 0x5b, 0x54, 0xa2, 0x1a, 0x95, 0x31,
	0xb3, 0xd5, 0xce, 0x0d, 0x62, 0x34, 0x46, 0x8a, 0x52, 0x29, 0x38, 0xb7, 0xed, 0x35, 0x0b, 0x59,
	0x2a, 0xdc, 0xa0, 0xa5, 0x82, 0xc4, 0x45, 0x7b, 0x2b, 0xea, 0x35, 0x06, 0x5c, 0xb7, 0xe2, 0x1e,
	0x6f, 0x50, 0xcb, 0x98, 0x40, 0x04, 0xa1, 0x07, 0x24, 0x53, 0xeb, 0xad, 0xb7, 0x5a, 0x7a, 0x0d,
	0x4d, 0xb9, 0x68, 0x00, 0x37, 0xda, 0x66, 0x8d, 0xa8, 0x9c, 0x09, 0x5a, 0xa0, 0x60, 0x5d, 0x39,
	0x40, 0x89, 0x54, 0x04, 0x53, 0xd7, 0x5e, 0x0c, 0xc8, 0xbd, 0x37, 0xfa, 0xe2, 0x2d, 0xea, 0x25,
	0x32, 0x96, 0x94, 0x4a, 0x56, 0x92, 0x66, 0x3e, 0x76, 0x88, 0x5c, 0xb4, 0xf4, 0xc8, 0x9e, 0xb4,
	0x25, 0xf7, 0x64, 0xf2, 0xf4, 0xea, 0x76, 0xfc, 0xf8, 0xc5, 0x89, 0xe9, 0x4f, 0x2f, 0x87, 0xf9,
	0xe1, 0x72, 0x79, 0x79, 0x34, 0xcc, 0x97, 0x8b, 0xd1, 0x37, 0xff, 0x76, 0xe8, 0x35, 0x66, 0x60,
	0x4a, 0x66, 0x41, 0x52, 0xe5, 0xe6, 0x85, 0x7c, 0x0e, 0x15, 0x25, 0x12, 0x01, 0x8b, 0x7b, 0x3a,
	0xf9, 0xf8, 0xea, 0x76, 0xfc, 0xec, 0x68, 0xbe, 0xba, 0xb0, 0x63, 0x39, 0x5f, 0xac, 0x2e, 0x66,
	0x47, 0x67, 0xb2, 0xec, 0xc3, 0xfc, 0xfc, 0x6b, 0x7b, 0x75, 0xaa, 0xb6, 0x18, 0x95, 0xb7, 0x5e,
	0xe2, 0x73, 0xc2, 0x24, 0xdd, 0x77, 0x0a, 0x1c, 0x14, 0xb3, 0x0a, 0x87, 0x9c, 0x7d, 0x0d, 0x92,
	0xab, 0x9a, 0xb8, 0xf7, 0x27, 0x8f, 0xae, 0x6e, 0xc7, 0xfb, 0xc7, 0xb6, 0xfc, 0x6e, 0xbe, 0x5a,
	0x2c, 0xad, 0x6d, 0xeb, 0xe9, 0xb5, 0x35, 0x35, 0xa3, 0x2a, 0x44, 0x64, 0x98, 0x62, 0xea, 0x89,
	0x8c, 0x10, 0x24, 0x1a, 0x56, 0x73, 0x1f, 0x6c, 0x12, 0x6b, 0x50, 0x0c, 0x21, 0x42, 0x04, 0xcf,
	0x10, 0x09, 0x10, 0x4b, 0xc4, 0x9e, 0x3c, 0x10, 0x40, 0xa8, 0x3d, 0xba, 0xf1, 0x7d, 0xe2, 0x17,
	0x77, 0x6f, 0xed, 0xcc, 0x96, 0xdf, 0xda, 0x62, 0x58, 0xcd, 0xd5, 0xbe, 0xbf, 0x9c, 0xcd, 0xa5,
	0xd9, 0xcb, 0xd5, 0x85, 0x6e, 0x47, 0x58, 0x54, 0x51, 0x51, 0x8b, 0x72, 0xa2, 0x84, 0x9a, 0x5b,
	0xae, 0xa8, 0x99, 0x49, 0x8d, 0x90, 0x32, 0xbb, 0x0f, 0xb7, 0x63, 0x51, 0xa5, 0x44, 0x20, 0x2e,
	0x02, 0x28, 0x95, 0xd9, 0x5a, 0x97, 0x46, 0x21, 0x99, 0x44, 0x54, 0xcf, 0xe8, 0x26, 0x5b, 0x57,
	0xb2, 0xea, 0xb1, 0x74, 0x96, 0x8c, 0x8c, 0x42, 0x99, 0xa4, 0x16, 0xa6, 0x62, 0xd2, 0x14, 0x7d,
	0x8f, 0xee, 0xd9, 0x64, 0xef, 0xef, 0xbf, 0xc6, 0xbb, 0xbf, 0xdc, 0x8e, 0x77, 0xfd, 0x57, 0x1f,
	0xbd, 0xfe, 0x73, 0xba, 0xf3, 0xfa, 0x7a, 0xba, 0xfb, 0xe6, 0x7a, 0xba, 0xfb, 0xc7, 0xf5, 0x74,
	0xf7, 0xd7, 0x9b, 0xe9, 0xce, 0x9b, 0x9b, 0xe9, 0xce, 0x6f, 0x37, 0xd3, 0x9d, 0xfa, 0xbf, 0xf5,
	0x27, 0x80, 0xff, 0x04, 0x00, 0x00, 0xff, 0xff, 0x98, 0x4e, 0x78, 0xb0, 0x70, 0x04, 0x00, 0x00,
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
	"HASH_8ba950479a03ab77edfad426ea53c173",
	"HASH_f4eb139f7a8373a484ab9749eadc31f5",
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
	VersionHashHASH_8Ba950479A03Ab77Edfad426Ea53C173 uint64 = 1 << 18
	VersionHashHASHF4Eb139F7A8373A484Ab9749Eadc31F5  uint64 = 1 << 19
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
	// HASH_8ba950479a03ab77edfad426ea53c173 -> Hash8Ba950479A03Ab77Edfad426Ea53C173
	26: "Hash8Ba950479A03Ab77Edfad426Ea53C173",
	// HASH_f4eb139f7a8373a484ab9749eadc31f5 -> HashF4Eb139F7A8373A484Ab9749Eadc31F5
	27: "HashF4Eb139F7A8373A484Ab9749Eadc31F5",
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
	"Hash8Ba950479A03Ab77Edfad426Ea53C173": 26,
	"HashF4Eb139F7A8373A484Ab9749Eadc31F5": 27,
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

func (e VersionHash) MarshalJSON() ([]byte, error) {
	str := proto.EnumName(VersionHash_CamelName, int32(e))
	return json.Marshal(str)
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
// UserAlertKey
// VMPoolKey
// VirtualClusterInstKey
var versionHashString = "f4eb139f7a8373a484ab9749eadc31f5"

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
	26: nil,
	27: nil,
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
	26: "",
	27: "",
}
