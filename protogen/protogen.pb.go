// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protogen.proto

/*
Package protogen is a generated protocol buffer package.

It is generated from these files:
	protogen.proto

It has these top-level messages:
*/
package protogen

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/protoc-gen-go/descriptor"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

var E_GenerateMatches = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51005,
	Name:          "protogen.generate_matches",
	Tag:           "varint,51005,opt,name=generate_matches,json=generateMatches",
	Filename:      "protogen.proto",
}

var E_GenerateCud = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51006,
	Name:          "protogen.generate_cud",
	Tag:           "varint,51006,opt,name=generate_cud,json=generateCud",
	Filename:      "protogen.proto",
}

func init() {
	proto.RegisterExtension(E_GenerateMatches)
	proto.RegisterExtension(E_GenerateCud)
}

func init() { proto.RegisterFile("protogen.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 151 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2b, 0x28, 0xca, 0x2f,
	0xc9, 0x4f, 0x4f, 0xcd, 0xd3, 0x03, 0x33, 0x84, 0x38, 0x60, 0x7c, 0x29, 0x85, 0xf4, 0xfc, 0xfc,
	0xf4, 0x9c, 0x54, 0x7d, 0xb0, 0x40, 0x52, 0x69, 0x9a, 0x7e, 0x4a, 0x6a, 0x71, 0x72, 0x51, 0x66,
	0x41, 0x49, 0x7e, 0x11, 0x44, 0xad, 0x95, 0x0f, 0x97, 0x40, 0x7a, 0x6a, 0x5e, 0x6a, 0x51, 0x62,
	0x49, 0x6a, 0x7c, 0x6e, 0x62, 0x49, 0x72, 0x46, 0x6a, 0xb1, 0x90, 0xbc, 0x1e, 0x44, 0x9b, 0x1e,
	0x4c, 0x9b, 0x9e, 0x6f, 0x6a, 0x71, 0x71, 0x62, 0x7a, 0xaa, 0x7f, 0x41, 0x49, 0x66, 0x7e, 0x5e,
	0xb1, 0xc4, 0xde, 0x3e, 0x66, 0x05, 0x46, 0x0d, 0x8e, 0x20, 0x7e, 0x98, 0x56, 0x5f, 0x88, 0x4e,
	0x2b, 0x17, 0x2e, 0x1e, 0xb8, 0x69, 0xc9, 0xa5, 0x29, 0x84, 0x4d, 0xda, 0x07, 0x35, 0x89, 0x1b,
	0xa6, 0xcd, 0xb9, 0x34, 0x25, 0x89, 0x0d, 0xac, 0xda, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0x1a,
	0xc8, 0xbd, 0x2b, 0xd8, 0x00, 0x00, 0x00,
}
