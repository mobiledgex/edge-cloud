// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: version.proto

package edgeproto

import (
	"encoding/json"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	"github.com/mobiledgex/edge-cloud/util"
	math "math"
	"strconv"
	strings "strings"
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
	VersionHash_HASH_09fae4d440aa06acb9664167d2e1f036 VersionHash = 28
	VersionHash_HASH_8c5a9c29caff4ace0a23a9dab9a15bf7 VersionHash = 29
	VersionHash_HASH_b7c6a74ce2f30b3bda179e00617459cf VersionHash = 30
	VersionHash_HASH_911d86a4eb2bbfbff1173ffbdd197a8c VersionHash = 31
	VersionHash_HASH_99349a696d0b5872542f81b4b0b4788e VersionHash = 32
	VersionHash_HASH_264850a5c1f7a054b4de1a87e5d28dcc VersionHash = 33
	VersionHash_HASH_748b47eaf414b0f2c15e4c6a9298b5f1 VersionHash = 34
	VersionHash_HASH_1480647750f7638ff5494c0e715bb98c VersionHash = 35
	VersionHash_HASH_208a22352e46f6bbe34f3b72aaf99ee5 VersionHash = 36
	VersionHash_HASH_fee52b2479f40655502fa5aa1ff78811 VersionHash = 37
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
	28: "HASH_09fae4d440aa06acb9664167d2e1f036",
	29: "HASH_8c5a9c29caff4ace0a23a9dab9a15bf7",
	30: "HASH_b7c6a74ce2f30b3bda179e00617459cf",
	31: "HASH_911d86a4eb2bbfbff1173ffbdd197a8c",
	32: "HASH_99349a696d0b5872542f81b4b0b4788e",
	33: "HASH_264850a5c1f7a054b4de1a87e5d28dcc",
	34: "HASH_748b47eaf414b0f2c15e4c6a9298b5f1",
	35: "HASH_1480647750f7638ff5494c0e715bb98c",
	36: "HASH_208a22352e46f6bbe34f3b72aaf99ee5",
	37: "HASH_fee52b2479f40655502fa5aa1ff78811",
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
	"HASH_09fae4d440aa06acb9664167d2e1f036": 28,
	"HASH_8c5a9c29caff4ace0a23a9dab9a15bf7": 29,
	"HASH_b7c6a74ce2f30b3bda179e00617459cf": 30,
	"HASH_911d86a4eb2bbfbff1173ffbdd197a8c": 31,
	"HASH_99349a696d0b5872542f81b4b0b4788e": 32,
	"HASH_264850a5c1f7a054b4de1a87e5d28dcc": 33,
	"HASH_748b47eaf414b0f2c15e4c6a9298b5f1": 34,
	"HASH_1480647750f7638ff5494c0e715bb98c": 35,
	"HASH_208a22352e46f6bbe34f3b72aaf99ee5": 36,
	"HASH_fee52b2479f40655502fa5aa1ff78811": 37,
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
	// 1061 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x95, 0xcf, 0x8f, 0x54, 0x45,
	0x10, 0xc7, 0x77, 0x2f, 0x26, 0x2c, 0x82, 0x8f, 0x11, 0x70, 0x1c, 0x70, 0x50, 0x91, 0x83, 0x26,
	0xb2, 0xfd, 0xa3, 0xaa, 0xab, 0x3b, 0xd1, 0xc4, 0x15, 0x24, 0x8b, 0x07, 0xb3, 0x01, 0xf1, 0x6a,
	0xaa, 0xab, 0xab, 0x96, 0x8d, 0xbb, 0x33, 0xeb, 0xfc, 0x20, 0x78, 0xf5, 0xb8, 0x27, 0xff, 0x2c,
	0x8e, 0x1c, 0x3d, 0x2a, 0xfc, 0x01, 0x5e, 0x9c, 0xc4, 0xa3, 0x99, 0xdd, 0x65, 0x86, 0xcb, 0x4b,
	0xe5, 0xbd, 0x4f, 0xaa, 0xbf, 0xfd, 0xad, 0x6f, 0xe5, 0x6d, 0x5d, 0x7a, 0xa6, 0x93, 0xe9, 0xc1,
	0x78, 0x74, 0xf7, 0x78, 0x32, 0x9e, 0x8d, 0x7b, 0x17, 0xb4, 0xed, 0xeb, 0x69, 0x39, 0xc8, 0xfb,
	0x07, 0xb3, 0xa7, 0xf3, 0x7a, 0x57, 0xc6, 0x47, 0xdb, 0x47, 0xe3, 0x7a, 0x70, 0xb8, 0xfc, 0xf4,
	0x7c, 0x7b, 0xf9, 0xfc, 0x52, 0x0e, 0xc7, 0xf3, 0xb6, 0x7d, 0xca, 0xed, 0xeb, 0x68, 0x55, 0x9c,
	0x35, 0x19, 0x5c, 0xdd, 0x1f, 0xef, 0x8f, 0x4f, 0xcb, 0xed, 0x65, 0x75, 0xf6, 0xf6, 0x8b, 0x7f,
	0x2e, 0x6e, 0x5d, 0xfc, 0xe9, 0xec, 0xb0, 0x5d, 0x9e, 0x3e, 0xed, 0x7d, 0xbe, 0x75, 0x67, 0x77,
	0xe7, 0xf1, 0xee, 0xcf, 0x0d, 0x7c, 0xcb, 0xd2, 0x4a, 0x36, 0xe7, 0x6a, 0x70, 0xa0, 0x25, 0x3b,
	0x57, 0x4a, 0x56, 0xb1, 0x0c, 0x81, 0xb4, 0xdb, 0x78, 0x0b, 0x15, 0x46, 0xf0, 0x99, 0x89, 0x5a,
	0x08, 0xad, 0xa4, 0x2c, 0x4a, 0x1c, 0xd8, 0x04, 0xa1, 0x34, 0xd3, 0xee, 0xc2, 0x0a, 0xa5, 0x0c,
	0xb9, 0x41, 0xd0, 0xc8, 0x41, 0xd9, 0x62, 0x52, 0x8c, 0xb5, 0xb6, 0xc8, 0x86, 0xd9, 0x57, 0x1f,
	0xb9, 0xdb, 0x5a, 0xa1, 0x16, 0x7d, 0x25, 0x2e, 0x8d, 0xd4, 0x25, 0xa3, 0xe0, 0x1d, 0xa9, 0xe3,
	0xea, 0xa3, 0x90, 0xcb, 0xe4, 0x40, 0xbb, 0x8b, 0x2b, 0xd4, 0x45, 0xe3, 0x86, 0xde, 0x5c, 0x84,
	0xd8, 0xc0, 0x5b, 0xf2, 0x14, 0x43, 0xf1, 0xe8, 0x0d, 0x08, 0x5a, 0xa8, 0xdd, 0xbb, 0x6b, 0x01,
	0x2d, 0x06, 0x2e, 0x39, 0x1a, 0x9b, 0xc4, 0xc6, 0x94, 0xb2, 0x3a, 0xc0, 0xea, 0x9b, 0x40, 0x43,
	0x43, 0xd7, 0x5d, 0x5a, 0xa3, 0x40, 0xe2, 0xa1, 0x36, 0xd3, 0xe0, 0x20, 0x9a, 0x2b, 0x2d, 0xa0,
	0xc7, 0x94, 0x15, 0x98, 0x42, 0x90, 0xd4, 0x5d, 0x5e, 0xa1, 0x42, 0x56, 0x83, 0x33, 0x04, 0x64,
	0xac, 0x52, 0x72, 0x2a, 0xd5, 0xb9, 0x5a, 0x89, 0x1c, 0x61, 0x94, 0xe8, 0xbb, 0xf7, 0x56, 0x68,
	0x8e, 0xd2, 0x50, 0x00, 0x2a, 0x0a, 0xc5, 0x4c, 0x5a, 0x8d, 0x9a, 0x43, 0x54, 0x8a, 0x80, 0x5c,
	0x21, 0x74, 0xdd, 0xda, 0xd7, 0xcc, 0xa0, 0xa9, 0x50, 0x73, 0x2d, 0x95, 0x08, 0x54, 0xc4, 0x5a,
	0x11, 0x2f, 0x18, 0xe2, 0xa9, 0x2b, 0xdd, 0x95, 0x15, 0xaa, 0x39, 0x26, 0xc7, 0x1c, 0x9d, 0x85,
	0x08, 0x2a, 0x6a, 0xcd, 0x5a, 0x2d, 0x56, 0x43, 0x13, 0x2a, 0x12, 0x5c, 0xd7, 0x5b, 0x6b, 0xc5,
	0x28, 0x94, 0xc1, 0xb5, 0x00, 0x41, 0x4d, 0x28, 0xb8, 0x82, 0x50, 0x38, 0x26, 0x13, 0x2b, 0xea,
	0xa0, 0x7b, 0xbf, 0xf7, 0xcd, 0x39, 0xea, 0x19, 0x29, 0x96, 0x94, 0x4a, 0x16, 0xe0, 0xa6, 0x1e,
	0xcd, 0x21, 0x15, 0x29, 0x86, 0xe4, 0x41, 0x5a, 0xea, 0xae, 0x0e, 0xae, 0x9d, 0x2c, 0xfa, 0x57,
	0xee, 0x3d, 0x55, 0xf9, 0xe5, 0xc1, 0x78, 0xb2, 0x3b, 0x9b, 0x1d, 0xef, 0x8d, 0x27, 0xb3, 0x69,
	0xef, 0xfb, 0x37, 0x1e, 0x7a, 0xc1, 0xec, 0x08, 0x92, 0x6a, 0xe0, 0x54, 0xa9, 0x79, 0x06, 0x9f,
	0x43, 0x8d, 0x8c, 0x00, 0x8e, 0xb8, 0xbb, 0x36, 0xb8, 0x75, 0xb2, 0xe8, 0xdf, 0xd8, 0x9b, 0xcc,
	0x47, 0xfa, 0x98, 0x8f, 0xa6, 0xf3, 0xd1, 0xfe, 0xde, 0x21, 0xcf, 0x6c, 0x3c, 0x39, 0xba, 0xaf,
	0xcf, 0x0e, 0x44, 0xa7, 0xbd, 0x72, 0xde, 0x8b, 0x7d, 0x4e, 0x31, 0xb1, 0x79, 0x83, 0x40, 0x41,
	0x62, 0x16, 0xa6, 0x90, 0xb3, 0xaf, 0x81, 0x73, 0x15, 0xe5, 0xee, 0xfa, 0xe0, 0xf2, 0xc9, 0xa2,
	0xbf, 0xf5, 0x58, 0x67, 0x3f, 0x4e, 0xe6, 0xd3, 0x99, 0xb6, 0xb5, 0x3d, 0x56, 0x5b, 0x13, 0x55,
	0xa8, 0x0c, 0x00, 0x1a, 0x13, 0x26, 0x4b, 0xa0, 0x10, 0x1d, 0xa3, 0xc6, 0xaa, 0xdd, 0x07, 0x2b,
	0xc5, 0x12, 0x24, 0x86, 0x80, 0x0e, 0x9d, 0x27, 0x87, 0xe0, 0x62, 0x2c, 0x18, 0x2d, 0x79, 0x07,
	0xce, 0x85, 0x6a, 0xd8, 0xf5, 0xcf, 0x14, 0xdf, 0x5b, 0xee, 0xda, 0xa1, 0xce, 0x1e, 0xe9, 0x74,
	0x3c, 0x9f, 0x88, 0x3e, 0x39, 0xde, 0x9f, 0x70, 0xd3, 0x07, 0xf3, 0x91, 0xac, 0x23, 0xcc, 0x22,
	0x51, 0xa2, 0x14, 0xa1, 0x04, 0x29, 0x4a, 0x6e, 0xb9, 0x46, 0xc9, 0x04, 0xa2, 0x10, 0x21, 0x53,
	0xf7, 0xe1, 0x3a, 0x16, 0x95, 0x0b, 0x3a, 0xa0, 0xc2, 0x2e, 0x72, 0x25, 0xd2, 0x66, 0xdc, 0x20,
	0x24, 0x65, 0x8c, 0xe2, 0x29, 0x76, 0x83, 0x75, 0x57, 0xd0, 0xea, 0x63, 0x31, 0xe2, 0x1c, 0x29,
	0x32, 0x64, 0xe0, 0x5a, 0x08, 0x8a, 0x72, 0x93, 0xe8, 0x0d, 0xbb, 0x1b, 0xeb, 0xc5, 0x28, 0xc6,
	0x0a, 0x0d, 0xc0, 0x31, 0xbb, 0xc4, 0x52, 0x4b, 0x4a, 0xe0, 0x13, 0xb5, 0xa0, 0xcb, 0x6d, 0x49,
	0xdd, 0xcd, 0xb5, 0x00, 0x41, 0x2e, 0x12, 0x8a, 0xb0, 0x19, 0xb0, 0xa8, 0xe3, 0x10, 0xb9, 0x34,
	0xae, 0x85, 0x3d, 0x56, 0xa3, 0xee, 0xa3, 0xde, 0x57, 0xe7, 0x68, 0x25, 0x49, 0xbc, 0xbc, 0x44,
	0xb0, 0xe8, 0x6a, 0xac, 0x8d, 0x3d, 0x15, 0x75, 0x2e, 0x79, 0x02, 0x2c, 0x62, 0xdd, 0x70, 0x70,
	0xe5, 0x64, 0xd1, 0xbf, 0xb4, 0x73, 0x7c, 0xfc, 0x70, 0x34, 0x9d, 0x3d, 0x52, 0x9b, 0xde, 0x7f,
	0xb4, 0x3a, 0xa8, 0x78, 0xdf, 0x72, 0x62, 0xd0, 0x1a, 0x6a, 0xb5, 0x6a, 0xe6, 0x3d, 0x45, 0x5b,
	0x0e, 0xc8, 0x17, 0xe2, 0x2c, 0xdd, 0xad, 0x35, 0x5a, 0x22, 0x14, 0x4e, 0x25, 0x35, 0x57, 0x31,
	0x53, 0x40, 0x08, 0x96, 0x7d, 0x85, 0xea, 0x2a, 0x50, 0xce, 0xda, 0x7d, 0xdc, 0xfb, 0xe1, 0x1c,
	0x0d, 0x09, 0x32, 0x3a, 0x46, 0xf1, 0x46, 0xec, 0x10, 0x2a, 0x34, 0xf5, 0x9c, 0x49, 0xb1, 0x85,
	0xdc, 0x44, 0xba, 0x4f, 0x06, 0xb7, 0x4f, 0x16, 0xfd, 0x5b, 0xa7, 0xc9, 0xd8, 0x1b, 0x1f, 0x1e,
	0xc8, 0x6f, 0xdf, 0x3d, 0x17, 0x3d, 0x9e, 0x1d, 0x8c, 0x47, 0x6f, 0x8f, 0xee, 0xeb, 0xd5, 0xf2,
	0xe7, 0x0a, 0xa4, 0x6c, 0xe0, 0xa1, 0x3a, 0x0b, 0xe2, 0x51, 0x41, 0x12, 0x97, 0x50, 0x72, 0x45,
	0xf3, 0xdd, 0xa7, 0x83, 0xde, 0xc9, 0xa2, 0x7f, 0x79, 0xa7, 0xb5, 0x7b, 0x87, 0xcb, 0xb0, 0x4d,
	0x96, 0xd7, 0x5c, 0x29, 0xf7, 0x90, 0x5d, 0x02, 0x22, 0x74, 0x46, 0x29, 0x66, 0x33, 0x84, 0x02,
	0xe2, 0x94, 0x3c, 0xd6, 0x5a, 0xb2, 0x74, 0xb7, 0x7b, 0x3b, 0x6f, 0x94, 0xbb, 0xcc, 0x21, 0x44,
	0x0c, 0x0a, 0xc9, 0x52, 0xad, 0x1a, 0xc1, 0x62, 0xa5, 0xc0, 0x6c, 0xa5, 0xa8, 0x62, 0xf7, 0xd9,
	0xe0, 0xfa, 0xc9, 0xa2, 0xdf, 0xdb, 0x69, 0xed, 0xdc, 0xd0, 0x27, 0xa3, 0x83, 0x5f, 0xe7, 0xfa,
	0x70, 0x1d, 0x6f, 0x53, 0xc5, 0x50, 0x03, 0x50, 0x31, 0x70, 0x09, 0x11, 0x5d, 0x30, 0x46, 0x66,
	0x6f, 0x46, 0x39, 0x7b, 0xdf, 0xdd, 0x19, 0x5c, 0xf8, 0xef, 0xdf, 0xfe, 0xe6, 0xef, 0x8b, 0xfe,
	0x66, 0xf8, 0xf6, 0xe6, 0x8b, 0xbf, 0x87, 0x1b, 0x2f, 0x5e, 0x0d, 0x37, 0x5f, 0xbe, 0x1a, 0x6e,
	0xfe, 0xf5, 0x6a, 0xb8, 0xf9, 0xc7, 0xeb, 0xe1, 0xc6, 0xcb, 0xd7, 0xc3, 0x8d, 0x3f, 0x5f, 0x0f,
	0x37, 0xea, 0x3b, 0xa7, 0xbf, 0x85, 0xf8, 0x7f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x41, 0x62, 0x10,
	0x1a, 0x82, 0x06, 0x00, 0x00,
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
	"HASH_09fae4d440aa06acb9664167d2e1f036",
	"HASH_8c5a9c29caff4ace0a23a9dab9a15bf7",
	"HASH_b7c6a74ce2f30b3bda179e00617459cf",
	"HASH_911d86a4eb2bbfbff1173ffbdd197a8c",
	"HASH_99349a696d0b5872542f81b4b0b4788e",
	"HASH_264850a5c1f7a054b4de1a87e5d28dcc",
	"HASH_748b47eaf414b0f2c15e4c6a9298b5f1",
	"HASH_1480647750f7638ff5494c0e715bb98c",
	"HASH_208a22352e46f6bbe34f3b72aaf99ee5",
	"HASH_fee52b2479f40655502fa5aa1ff78811",
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
	VersionHashHASH_09Fae4D440Aa06Acb9664167D2E1F036 uint64 = 1 << 20
	VersionHashHASH_8C5A9C29Caff4Ace0A23A9Dab9A15Bf7 uint64 = 1 << 21
	VersionHashHASHB7C6A74Ce2F30B3Bda179E00617459Cf  uint64 = 1 << 22
	VersionHashHASH_911D86A4Eb2Bbfbff1173Ffbdd197A8C uint64 = 1 << 23
	VersionHashHASH_99349A696D0B5872542F81B4B0B4788E uint64 = 1 << 24
	VersionHashHASH_264850A5C1F7A054B4De1A87E5D28Dcc uint64 = 1 << 25
	VersionHashHASH_748B47Eaf414B0F2C15E4C6A9298B5F1 uint64 = 1 << 26
	VersionHashHASH_1480647750F7638Ff5494C0E715Bb98C uint64 = 1 << 27
	VersionHashHASH_208A22352E46F6Bbe34F3B72Aaf99Ee5 uint64 = 1 << 28
	VersionHashHASHFee52B2479F40655502Fa5Aa1Ff78811  uint64 = 1 << 29
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
	// HASH_09fae4d440aa06acb9664167d2e1f036 -> Hash09Fae4D440Aa06Acb9664167D2E1F036
	28: "Hash09Fae4D440Aa06Acb9664167D2E1F036",
	// HASH_8c5a9c29caff4ace0a23a9dab9a15bf7 -> Hash8C5A9C29Caff4Ace0A23A9Dab9A15Bf7
	29: "Hash8C5A9C29Caff4Ace0A23A9Dab9A15Bf7",
	// HASH_b7c6a74ce2f30b3bda179e00617459cf -> HashB7C6A74Ce2F30B3Bda179E00617459Cf
	30: "HashB7C6A74Ce2F30B3Bda179E00617459Cf",
	// HASH_911d86a4eb2bbfbff1173ffbdd197a8c -> Hash911D86A4Eb2Bbfbff1173Ffbdd197A8C
	31: "Hash911D86A4Eb2Bbfbff1173Ffbdd197A8C",
	// HASH_99349a696d0b5872542f81b4b0b4788e -> Hash99349A696D0B5872542F81B4B0B4788E
	32: "Hash99349A696D0B5872542F81B4B0B4788E",
	// HASH_264850a5c1f7a054b4de1a87e5d28dcc -> Hash264850A5C1F7A054B4De1A87E5D28Dcc
	33: "Hash264850A5C1F7A054B4De1A87E5D28Dcc",
	// HASH_748b47eaf414b0f2c15e4c6a9298b5f1 -> Hash748B47Eaf414B0F2C15E4C6A9298B5F1
	34: "Hash748B47Eaf414B0F2C15E4C6A9298B5F1",
	// HASH_1480647750f7638ff5494c0e715bb98c -> Hash1480647750F7638Ff5494C0E715Bb98C
	35: "Hash1480647750F7638Ff5494C0E715Bb98C",
	// HASH_208a22352e46f6bbe34f3b72aaf99ee5 -> Hash208A22352E46F6Bbe34F3B72Aaf99Ee5
	36: "Hash208A22352E46F6Bbe34F3B72Aaf99Ee5",
	// HASH_fee52b2479f40655502fa5aa1ff78811 -> HashFee52B2479F40655502Fa5Aa1Ff78811
	37: "HashFee52B2479F40655502Fa5Aa1Ff78811",
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
	"Hash09Fae4D440Aa06Acb9664167D2E1F036": 28,
	"Hash8C5A9C29Caff4Ace0A23A9Dab9A15Bf7": 29,
	"HashB7C6A74Ce2F30B3Bda179E00617459Cf": 30,
	"Hash911D86A4Eb2Bbfbff1173Ffbdd197A8C": 31,
	"Hash99349A696D0B5872542F81B4B0B4788E": 32,
	"Hash264850A5C1F7A054B4De1A87E5D28Dcc": 33,
	"Hash748B47Eaf414B0F2C15E4C6A9298B5F1": 34,
	"Hash1480647750F7638Ff5494C0E715Bb98C": 35,
	"Hash208A22352E46F6Bbe34F3B72Aaf99Ee5": 36,
	"HashFee52B2479F40655502Fa5Aa1Ff78811": 37,
}

func (e *VersionHash) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	val, ok := VersionHash_CamelValue[util.CamelCase(str)]
	if !ok {
		// may have omitted common prefix
		val, ok = VersionHash_CamelValue["Hash"+util.CamelCase(str)]
	}
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = VersionHash_CamelName[val]
		}
	}
	if !ok {
		return fmt.Errorf("Invalid VersionHash value %q", str)
	}
	*e = VersionHash(val)
	return nil
}

func (e VersionHash) MarshalYAML() (interface{}, error) {
	str := proto.EnumName(VersionHash_CamelName, int32(e))
	str = strings.TrimPrefix(str, "Hash")
	return str, nil
}

// custom JSON encoding/decoding
func (e *VersionHash) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		val, ok := VersionHash_CamelValue[util.CamelCase(str)]
		if !ok {
			// may have omitted common prefix
			val, ok = VersionHash_CamelValue["Hash"+util.CamelCase(str)]
		}
		if !ok {
			// may be int value instead of enum name
			ival, err := strconv.Atoi(str)
			val = int32(ival)
			if err == nil {
				_, ok = VersionHash_CamelName[val]
			}
		}
		if !ok {
			return fmt.Errorf("Invalid VersionHash value %q", str)
		}
		*e = VersionHash(val)
		return nil
	}
	var val int32
	err = json.Unmarshal(b, &val)
	if err == nil {
		_, ok := VersionHash_CamelName[val]
		if !ok {
			return fmt.Errorf("Invalid VersionHash value %d", val)
		}
		*e = VersionHash(val)
		return nil
	}
	return fmt.Errorf("Invalid VersionHash value %v", b)
}

/*
 * This is removed because we do not have enough time in
 * release 3.0 to update the SDK, UI, and documentation for this
 * change. It should be done in 3.1.
func (e VersionHash) MarshalJSON() ([]byte, error) {
	str := proto.EnumName(VersionHash_CamelName, int32(e))
	str = strings.TrimPrefix(str, "Hash")
	return json.Marshal(str)
}
*/
var VersionHashCommonPrefix = "Hash"

// Keys being hashed:
// AlertPolicyKey
// AppInstKey
// AppInstRefKey
// AppKey
// CloudletKey
// CloudletPoolKey
// ClusterInstKey
// ClusterInstRefKey
// ClusterKey
// ClusterRefsAppInstKey
// ControllerKey
// DeviceKey
// FlavorKey
// FlowRateLimitSettingsKey
// GPUDriverKey
// MaxReqsRateLimitSettingsKey
// NetworkKey
// NodeKey
// PolicyKey
// RateLimitSettingsKey
// ResTagTableKey
// StreamKey
// TrustPolicyExceptionKey
// VMPoolKey
// VirtualClusterInstKey
var versionHashString = "fee52b2479f40655502fa5aa1ff78811"

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
	28: nil,
	29: nil,
	30: AppInstRefsDR,
	31: nil,
	32: nil,
	33: TrustPolicyExceptionUpgradeFunc,
	34: AddClusterRefs,
	35: nil,
	36: AddAppInstUniqueId,
	37: nil,
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
	28: "",
	29: "",
	30: "AppInstRefsDR",
	31: "",
	32: "",
	33: "TrustPolicyExceptionUpgradeFunc",
	34: "AddClusterRefs",
	35: "",
	36: "AddAppInstUniqueId",
	37: "",
}
