// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: protogen.proto

/*
Package protogen is a generated protocol buffer package.

It is generated from these files:
	protogen.proto

It has these top-level messages:
*/
package protogen

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

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

var E_ObjKey = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51007,
	Name:          "protogen.obj_key",
	Tag:           "varint,51007,opt,name=obj_key,json=objKey",
	Filename:      "protogen.proto",
}

var E_GenerateCache = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51008,
	Name:          "protogen.generate_cache",
	Tag:           "varint,51008,opt,name=generate_cache,json=generateCache",
	Filename:      "protogen.proto",
}

var E_NotifyCache = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51010,
	Name:          "protogen.notify_cache",
	Tag:           "varint,51010,opt,name=notify_cache,json=notifyCache",
	Filename:      "protogen.proto",
}

var E_GenerateCudTest = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51011,
	Name:          "protogen.generate_cud_test",
	Tag:           "varint,51011,opt,name=generate_cud_test,json=generateCudTest",
	Filename:      "protogen.proto",
}

var E_GenerateShowTest = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51012,
	Name:          "protogen.generate_show_test",
	Tag:           "varint,51012,opt,name=generate_show_test,json=generateShowTest",
	Filename:      "protogen.proto",
}

var E_GenerateCudStreamout = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51014,
	Name:          "protogen.generate_cud_streamout",
	Tag:           "varint,51014,opt,name=generate_cud_streamout,json=generateCudStreamout",
	Filename:      "protogen.proto",
}

var E_GenerateWaitForState = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*string)(nil),
	Field:         51015,
	Name:          "protogen.generate_wait_for_state",
	Tag:           "bytes,51015,opt,name=generate_wait_for_state,json=generateWaitForState",
	Filename:      "protogen.proto",
}

var E_NotifyMessage = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51016,
	Name:          "protogen.notify_message",
	Tag:           "varint,51016,opt,name=notify_message,json=notifyMessage",
	Filename:      "protogen.proto",
}

var E_NotifyCustomUpdate = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51017,
	Name:          "protogen.notify_custom_update",
	Tag:           "varint,51017,opt,name=notify_custom_update,json=notifyCustomUpdate",
	Filename:      "protogen.proto",
}

var E_NotifyRecvHook = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51018,
	Name:          "protogen.notify_recv_hook",
	Tag:           "varint,51018,opt,name=notify_recv_hook,json=notifyRecvHook",
	Filename:      "protogen.proto",
}

var E_NotifyFilterCloudletKey = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51019,
	Name:          "protogen.notify_filter_cloudlet_key",
	Tag:           "varint,51019,opt,name=notify_filter_cloudlet_key,json=notifyFilterCloudletKey",
	Filename:      "protogen.proto",
}

var E_NotifyFlush = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51020,
	Name:          "protogen.notify_flush",
	Tag:           "varint,51020,opt,name=notify_flush,json=notifyFlush",
	Filename:      "protogen.proto",
}

var E_NotifyPrintSendRecv = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51022,
	Name:          "protogen.notify_print_send_recv",
	Tag:           "varint,51022,opt,name=notify_print_send_recv,json=notifyPrintSendRecv",
	Filename:      "protogen.proto",
}

var E_GenerateVersion = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51023,
	Name:          "protogen.generate_version",
	Tag:           "varint,51023,opt,name=generate_version,json=generateVersion",
	Filename:      "protogen.proto",
}

var E_CheckVersion = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51024,
	Name:          "protogen.check_version",
	Tag:           "varint,51024,opt,name=check_version,json=checkVersion",
	Filename:      "protogen.proto",
}

var E_VersionHash = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.EnumOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51025,
	Name:          "protogen.version_hash",
	Tag:           "varint,51025,opt,name=version_hash,json=versionHash",
	Filename:      "protogen.proto",
}

var E_TestUpdate = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.FieldOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51009,
	Name:          "protogen.test_update",
	Tag:           "varint,51009,opt,name=test_update,json=testUpdate",
	Filename:      "protogen.proto",
}

var E_Backend = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.FieldOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         51013,
	Name:          "protogen.backend",
	Tag:           "varint,51013,opt,name=backend",
	Filename:      "protogen.proto",
}

var E_Hidetag = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.FieldOptions)(nil),
	ExtensionType: (*string)(nil),
	Field:         52002,
	Name:          "protogen.hidetag",
	Tag:           "bytes,52002,opt,name=hidetag",
	Filename:      "protogen.proto",
}

var E_Mc2Api = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MethodOptions)(nil),
	ExtensionType: (*string)(nil),
	Field:         51021,
	Name:          "protogen.mc2_api",
	Tag:           "bytes,51021,opt,name=mc2_api,json=mc2Api",
	Filename:      "protogen.proto",
}

func init() {
	proto.RegisterExtension(E_GenerateMatches)
	proto.RegisterExtension(E_GenerateCud)
	proto.RegisterExtension(E_ObjKey)
	proto.RegisterExtension(E_GenerateCache)
	proto.RegisterExtension(E_NotifyCache)
	proto.RegisterExtension(E_GenerateCudTest)
	proto.RegisterExtension(E_GenerateShowTest)
	proto.RegisterExtension(E_GenerateCudStreamout)
	proto.RegisterExtension(E_GenerateWaitForState)
	proto.RegisterExtension(E_NotifyMessage)
	proto.RegisterExtension(E_NotifyCustomUpdate)
	proto.RegisterExtension(E_NotifyRecvHook)
	proto.RegisterExtension(E_NotifyFilterCloudletKey)
	proto.RegisterExtension(E_NotifyFlush)
	proto.RegisterExtension(E_NotifyPrintSendRecv)
	proto.RegisterExtension(E_GenerateVersion)
	proto.RegisterExtension(E_CheckVersion)
	proto.RegisterExtension(E_VersionHash)
	proto.RegisterExtension(E_TestUpdate)
	proto.RegisterExtension(E_Backend)
	proto.RegisterExtension(E_Hidetag)
	proto.RegisterExtension(E_Mc2Api)
}

func init() { proto.RegisterFile("protogen.proto", fileDescriptorProtogen) }

var fileDescriptorProtogen = []byte{
	// 609 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0xd4, 0xdb, 0x4e, 0xd4, 0x4e,
	0x1c, 0xc0, 0xf1, 0x6c, 0xfe, 0x09, 0xf0, 0x1f, 0x16, 0xc4, 0x4a, 0xc0, 0x10, 0x5d, 0xb9, 0xf4,
	0x6a, 0x49, 0xf0, 0xca, 0xb9, 0x12, 0xd1, 0x0d, 0x09, 0x12, 0x0c, 0xab, 0xe0, 0x95, 0xcd, 0xec,
	0xf4, 0xb7, 0x9d, 0x61, 0xdb, 0x4e, 0xd3, 0x99, 0x42, 0x78, 0x09, 0x2f, 0x7d, 0x00, 0xdf, 0xc3,
	0xf3, 0x09, 0xcf, 0xf8, 0x06, 0x86, 0x27, 0x31, 0xd3, 0xfe, 0xa6, 0x9c, 0x4c, 0xda, 0x3b, 0xd8,
	0x9d, 0xef, 0x27, 0x73, 0x5c, 0x32, 0x9d, 0x66, 0xca, 0xa8, 0x10, 0x92, 0x6e, 0xf1, 0x87, 0x37,
	0xe1, 0xfe, 0x5f, 0x58, 0x0c, 0x95, 0x0a, 0x23, 0x58, 0x2a, 0x3e, 0x18, 0xe4, 0xc3, 0xa5, 0x00,
	0x34, 0xcf, 0x64, 0x6a, 0x54, 0x56, 0x8e, 0xa5, 0x0f, 0xc8, 0x4c, 0x08, 0x09, 0x64, 0xcc, 0x80,
	0x1f, 0x33, 0xc3, 0x05, 0x68, 0xef, 0x46, 0xb7, 0xcc, 0xba, 0x2e, 0xeb, 0x6e, 0x80, 0xd6, 0x2c,
	0x84, 0xcd, 0xd4, 0x48, 0x95, 0xe8, 0xab, 0x2f, 0x9f, 0xfd, 0xb7, 0xd8, 0xba, 0x39, 0xb1, 0x75,
	0xc9, 0xa5, 0x1b, 0x65, 0x49, 0xef, 0x91, 0x76, 0xa5, 0xf1, 0x3c, 0xa8, 0x97, 0x5e, 0xa1, 0x34,
	0xe9, 0xb2, 0xd5, 0x3c, 0xa0, 0x94, 0x8c, 0xab, 0xc1, 0xae, 0x3f, 0x82, 0x83, 0x7a, 0xe0, 0x35,
	0x02, 0x63, 0x6a, 0xb0, 0xbb, 0x0e, 0x07, 0x74, 0x8d, 0x4c, 0x9f, 0xcc, 0x80, 0x71, 0x01, 0xf5,
	0xc4, 0x1b, 0x24, 0xa6, 0xaa, 0x39, 0xd8, 0xce, 0xae, 0x25, 0x51, 0x46, 0x0e, 0x0f, 0x9a, 0x3a,
	0xef, 0xdc, 0x5a, 0xca, 0xac, 0x54, 0x36, 0xc8, 0xe5, 0xd3, 0x3b, 0xe2, 0x1b, 0xd0, 0xa6, 0x9e,
	0x7a, 0x7f, 0x7e, 0x83, 0x57, 0xf3, 0xe0, 0x11, 0x68, 0x43, 0x37, 0x89, 0x57, 0x71, 0x5a, 0xa8,
	0xfd, 0x86, 0xde, 0x07, 0xf4, 0xaa, 0xb3, 0xee, 0x0b, 0xb5, 0x5f, 0x80, 0x3b, 0x64, 0xee, 0xcc,
	0xfc, 0xb4, 0xc9, 0x80, 0xc5, 0x2a, 0x6f, 0x80, 0x7e, 0x42, 0x74, 0xf6, 0xd4, 0x24, 0xfb, 0x2e,
	0xa7, 0x4f, 0xc8, 0x7c, 0x05, 0xef, 0x33, 0x69, 0xfc, 0xa1, 0xca, 0x7c, 0x6d, 0x98, 0x69, 0xb0,
	0x93, 0x9f, 0x0b, 0xf9, 0xff, 0x13, 0x79, 0x87, 0x49, 0xd3, 0x53, 0x59, 0xdf, 0xe6, 0xf6, 0x88,
	0xf1, 0x60, 0xe2, 0x32, 0xab, 0x07, 0x0f, 0xdd, 0x11, 0x97, 0x21, 0x7e, 0x4b, 0xfb, 0x64, 0xd6,
	0x1d, 0x71, 0xae, 0x8d, 0x8a, 0xfd, 0x3c, 0x0d, 0x1a, 0x4d, 0xf0, 0x0b, 0x7a, 0x1e, 0x1e, 0x75,
	0x51, 0x3f, 0x2e, 0x62, 0xba, 0x4e, 0x66, 0x10, 0xcd, 0x80, 0xef, 0xf9, 0x42, 0xa9, 0x51, 0x3d,
	0xf8, 0x15, 0x41, 0x5c, 0xd9, 0x16, 0xf0, 0xbd, 0x35, 0xa5, 0x46, 0xf4, 0x29, 0x59, 0x40, 0x6c,
	0x28, 0x23, 0x03, 0x99, 0xcf, 0x23, 0x95, 0x07, 0x11, 0x98, 0x66, 0xaf, 0xe3, 0x1b, 0xb2, 0xf3,
	0x25, 0xd2, 0x2b, 0x8c, 0x55, 0x24, 0xec, 0x73, 0x39, 0xb9, 0xe4, 0xc3, 0x28, 0xd7, 0xa2, 0x5e,
	0xfc, 0x7e, 0xf6, 0x92, 0xf7, 0x6c, 0x45, 0xb7, 0xc9, 0x1c, 0x2a, 0x69, 0x26, 0x13, 0xe3, 0x6b,
	0x48, 0x82, 0x62, 0xf5, 0xf5, 0xde, 0x4f, 0xf4, 0xae, 0x94, 0xc0, 0x43, 0xdb, 0xf7, 0x21, 0x09,
	0xec, 0x0e, 0x9c, 0xf9, 0x71, 0xda, 0x83, 0x4c, 0x4b, 0x95, 0xd4, 0x8b, 0xbf, 0xce, 0xbf, 0x9d,
	0xed, 0xb2, 0xa4, 0x3d, 0x32, 0xc5, 0x05, 0xf0, 0x51, 0x73, 0xea, 0x08, 0xa9, 0x76, 0xd1, 0x39,
	0x67, 0x85, 0xb4, 0x51, 0xf0, 0x05, 0xd3, 0xc2, 0xbb, 0x76, 0x81, 0xb9, 0x9f, 0xe4, 0xb1, 0x33,
	0x7e, 0xbb, 0x0d, 0xc3, 0x66, 0x8d, 0x69, 0x41, 0xef, 0x90, 0x49, 0xfb, 0x70, 0xdd, 0x7d, 0xbb,
	0x7e, 0x41, 0xe8, 0x49, 0x88, 0x02, 0x47, 0xbc, 0x45, 0x82, 0xd8, 0x06, 0x6f, 0xd9, 0x6d, 0x32,
	0x3e, 0x60, 0x7c, 0x04, 0x49, 0x50, 0x57, 0x7f, 0xc4, 0xda, 0x8d, 0xb7, 0xa9, 0x90, 0x01, 0x18,
	0x16, 0xd6, 0xa5, 0x2f, 0x9e, 0x97, 0xef, 0xd0, 0x8d, 0xb7, 0x69, 0xcc, 0x97, 0x7d, 0x96, 0x4a,
	0xaf, 0xf3, 0x8f, 0xcd, 0x33, 0x42, 0x55, 0xed, 0x0f, 0x7c, 0xc3, 0x63, 0x31, 0x5f, 0x5e, 0x49,
	0xe5, 0xdd, 0xf6, 0xe1, 0x71, 0xa7, 0x75, 0x74, 0xdc, 0x69, 0xfd, 0x39, 0xee, 0xb4, 0x06, 0x63,
	0x45, 0x75, 0xeb, 0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x9e, 0xc5, 0xf8, 0xed, 0xbb, 0x06, 0x00,
	0x00,
}
