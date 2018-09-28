// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: appcommon.proto

package distributed_match_engine

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import "errors"
import "strconv"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// LProto indicates which protocol to use for accessing an application on a particular port. This is required by Kubernetes for port mapping.
type LProto int32

const (
	// Unknown protocol
	LProto_LProtoUnknown LProto = 0
	// TCP (L4) protocol
	LProto_LProtoTCP LProto = 1
	// UDP (L4) protocol
	LProto_LProtoUDP LProto = 2
	// HTTP (L7 tcp) protocol
	LProto_LProtoHTTP LProto = 3
)

var LProto_name = map[int32]string{
	0: "LProtoUnknown",
	1: "LProtoTCP",
	2: "LProtoUDP",
	3: "LProtoHTTP",
}
var LProto_value = map[string]int32{
	"LProtoUnknown": 0,
	"LProtoTCP":     1,
	"LProtoUDP":     2,
	"LProtoHTTP":    3,
}

func (x LProto) String() string {
	return proto.EnumName(LProto_name, int32(x))
}
func (LProto) EnumDescriptor() ([]byte, []int) { return fileDescriptorAppcommon, []int{0} }

// AppPort describes an L4 public access port mapping. This is used to track external to internal mappings for access via a shared load balancer or reverse proxy.
type AppPort struct {
	// TCP, UDP, or HTTP protocol
	Proto LProto `protobuf:"varint,1,opt,name=proto,proto3,enum=distributed_match_engine.LProto" json:"proto,omitempty"`
	// Container port
	InternalPort int32 `protobuf:"varint,2,opt,name=internal_port,json=internalPort,proto3" json:"internal_port,omitempty"`
	// Public facing port for TCP/UDP (may be mapped on shared LB reverse proxy)
	PublicPort int32 `protobuf:"varint,3,opt,name=public_port,json=publicPort,proto3" json:"public_port,omitempty"`
	// Public facing path for HTTP L7 access.
	PublicPath string `protobuf:"bytes,4,opt,name=public_path,json=publicPath,proto3" json:"public_path,omitempty"`
}

func (m *AppPort) Reset()                    { *m = AppPort{} }
func (m *AppPort) String() string            { return proto.CompactTextString(m) }
func (*AppPort) ProtoMessage()               {}
func (*AppPort) Descriptor() ([]byte, []int) { return fileDescriptorAppcommon, []int{0} }

func init() {
	proto.RegisterType((*AppPort)(nil), "distributed_match_engine.AppPort")
	proto.RegisterEnum("distributed_match_engine.LProto", LProto_name, LProto_value)
}
func (m *AppPort) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AppPort) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Proto != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintAppcommon(dAtA, i, uint64(m.Proto))
	}
	if m.InternalPort != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintAppcommon(dAtA, i, uint64(m.InternalPort))
	}
	if m.PublicPort != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintAppcommon(dAtA, i, uint64(m.PublicPort))
	}
	if len(m.PublicPath) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintAppcommon(dAtA, i, uint64(len(m.PublicPath)))
		i += copy(dAtA[i:], m.PublicPath)
	}
	return i, nil
}

func encodeVarintAppcommon(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *AppPort) CopyInFields(src *AppPort) {
	m.Proto = src.Proto
	m.InternalPort = src.InternalPort
	m.PublicPort = src.PublicPort
	m.PublicPath = src.PublicPath
}

var LProtoStrings = []string{
	"LProtoUnknown",
	"LProtoTCP",
	"LProtoUDP",
	"LProtoHTTP",
}

const (
	LProtoLProtoUnknown uint64 = 1 << 0
	LProtoLProtoTCP     uint64 = 1 << 1
	LProtoLProtoUDP     uint64 = 1 << 2
	LProtoLProtoHTTP    uint64 = 1 << 3
)

func (e *LProto) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err != nil {
		return err
	}
	val, ok := LProto_value[str]
	if !ok {
		// may be enum value instead of string
		ival, err := strconv.Atoi(str)
		val = int32(ival)
		if err == nil {
			_, ok = LProto_name[val]
		}
	}
	if !ok {
		return errors.New(fmt.Sprintf("No enum value for %s", str))
	}
	*e = LProto(val)
	return nil
}

func (e LProto) MarshalYAML() (interface{}, error) {
	return e.String(), nil
}

func (m *AppPort) Size() (n int) {
	var l int
	_ = l
	if m.Proto != 0 {
		n += 1 + sovAppcommon(uint64(m.Proto))
	}
	if m.InternalPort != 0 {
		n += 1 + sovAppcommon(uint64(m.InternalPort))
	}
	if m.PublicPort != 0 {
		n += 1 + sovAppcommon(uint64(m.PublicPort))
	}
	l = len(m.PublicPath)
	if l > 0 {
		n += 1 + l + sovAppcommon(uint64(l))
	}
	return n
}

func sovAppcommon(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozAppcommon(x uint64) (n int) {
	return sovAppcommon(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *AppPort) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAppcommon
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: AppPort: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AppPort: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Proto", wireType)
			}
			m.Proto = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppcommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Proto |= (LProto(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field InternalPort", wireType)
			}
			m.InternalPort = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppcommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.InternalPort |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PublicPort", wireType)
			}
			m.PublicPort = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppcommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PublicPort |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PublicPath", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAppcommon
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAppcommon
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PublicPath = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAppcommon(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthAppcommon
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipAppcommon(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowAppcommon
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowAppcommon
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowAppcommon
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthAppcommon
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowAppcommon
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipAppcommon(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthAppcommon = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowAppcommon   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("appcommon.proto", fileDescriptorAppcommon) }

var fileDescriptorAppcommon = []byte{
	// 241 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4f, 0x2c, 0x28, 0x48,
	0xce, 0xcf, 0xcd, 0xcd, 0xcf, 0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x92, 0x48, 0xc9, 0x2c,
	0x2e, 0x29, 0xca, 0x4c, 0x2a, 0x2d, 0x49, 0x4d, 0x89, 0xcf, 0x4d, 0x2c, 0x49, 0xce, 0x88, 0x4f,
	0xcd, 0x4b, 0xcf, 0xcc, 0x4b, 0x55, 0x5a, 0xc1, 0xc8, 0xc5, 0xee, 0x58, 0x50, 0x10, 0x90, 0x5f,
	0x54, 0x22, 0x64, 0xc6, 0xc5, 0x0a, 0x56, 0x2e, 0xc1, 0xa8, 0xc0, 0xa8, 0xc1, 0x67, 0xa4, 0xa0,
	0x87, 0x4b, 0x97, 0x9e, 0x4f, 0x00, 0x48, 0x5d, 0x10, 0x44, 0xb9, 0x90, 0x32, 0x17, 0x6f, 0x66,
	0x5e, 0x49, 0x6a, 0x51, 0x5e, 0x62, 0x4e, 0x7c, 0x41, 0x7e, 0x51, 0x89, 0x04, 0x93, 0x02, 0xa3,
	0x06, 0x6b, 0x10, 0x0f, 0x4c, 0x10, 0x6c, 0xb8, 0x3c, 0x17, 0x77, 0x41, 0x69, 0x52, 0x4e, 0x66,
	0x32, 0x44, 0x09, 0x33, 0x58, 0x09, 0x17, 0x44, 0x08, 0x5d, 0x41, 0x62, 0x49, 0x86, 0x04, 0x8b,
	0x02, 0xa3, 0x06, 0x27, 0x5c, 0x41, 0x62, 0x49, 0x86, 0x96, 0x27, 0x17, 0x1b, 0xc4, 0x5e, 0x21,
	0x41, 0x2e, 0x5e, 0x08, 0x2b, 0x34, 0x2f, 0x3b, 0x2f, 0xbf, 0x3c, 0x4f, 0x80, 0x41, 0x88, 0x97,
	0x8b, 0x13, 0x22, 0x14, 0xe2, 0x1c, 0x20, 0xc0, 0x88, 0xe0, 0x86, 0xba, 0x04, 0x08, 0x30, 0x09,
	0xf1, 0x71, 0x71, 0x41, 0xb8, 0x1e, 0x21, 0x21, 0x01, 0x02, 0xcc, 0x4e, 0x02, 0x27, 0x1e, 0xca,
	0x31, 0x9c, 0x78, 0x24, 0xc7, 0x78, 0xe1, 0x91, 0x1c, 0xe3, 0x83, 0x47, 0x72, 0x8c, 0x49, 0x6c,
	0x60, 0xaf, 0x18, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0xa3, 0x0c, 0xc9, 0x0d, 0x3b, 0x01, 0x00,
	0x00,
}
