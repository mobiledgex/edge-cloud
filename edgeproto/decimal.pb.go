// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: decimal.proto

package edgeproto

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	io "io"
	math "math"
	math_bits "math/bits"
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

// Udec64
//
// Udec64 is an unsigned decimal with whole number values
// as uint64, and decimal values in nanos.
type Udec64 struct {
	// Whole number value
	Whole uint64 `protobuf:"varint,1,opt,name=whole,proto3" json:"whole,omitempty"`
	// Decimal value in nanos
	Nanos uint32 `protobuf:"varint,2,opt,name=nanos,proto3" json:"nanos,omitempty"`
}

func (m *Udec64) Reset()         { *m = Udec64{} }
func (m *Udec64) String() string { return proto.CompactTextString(m) }
func (*Udec64) ProtoMessage()    {}
func (*Udec64) Descriptor() ([]byte, []int) {
	return fileDescriptor_2f2ba8127f840582, []int{0}
}
func (m *Udec64) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Udec64) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Udec64.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Udec64) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Udec64.Merge(m, src)
}
func (m *Udec64) XXX_Size() int {
	return m.Size()
}
func (m *Udec64) XXX_DiscardUnknown() {
	xxx_messageInfo_Udec64.DiscardUnknown(m)
}

var xxx_messageInfo_Udec64 proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Udec64)(nil), "edgeproto.Udec64")
}

func init() { proto.RegisterFile("decimal.proto", fileDescriptor_2f2ba8127f840582) }

var fileDescriptor_2f2ba8127f840582 = []byte{
	// 189 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4d, 0x49, 0x4d, 0xce,
	0xcc, 0x4d, 0xcc, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x4c, 0x4d, 0x49, 0x4f, 0x05,
	0x33, 0xa5, 0x44, 0xd2, 0xf3, 0xd3, 0xf3, 0xc1, 0x4c, 0x7d, 0x10, 0x0b, 0xa2, 0x40, 0xca, 0x22,
	0x3d, 0xb3, 0x24, 0xa3, 0x34, 0x49, 0x2f, 0x39, 0x3f, 0x57, 0x3f, 0x37, 0x3f, 0x29, 0x33, 0x07,
	0xa4, 0xa1, 0x42, 0x1f, 0x44, 0xea, 0x26, 0xe7, 0xe4, 0x97, 0xa6, 0xe8, 0x83, 0xd5, 0xa5, 0xa7,
	0xe6, 0xc1, 0x19, 0x10, 0x9d, 0x4a, 0x76, 0x5c, 0x6c, 0xa1, 0x29, 0xa9, 0xc9, 0x66, 0x26, 0x42,
	0x22, 0x5c, 0xac, 0xe5, 0x19, 0xf9, 0x39, 0xa9, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x2c, 0x41, 0x10,
	0x0e, 0x48, 0x34, 0x2f, 0x31, 0x2f, 0xbf, 0x58, 0x82, 0x49, 0x81, 0x51, 0x83, 0x37, 0x08, 0xc2,
	0xb1, 0xe2, 0x78, 0xf0, 0x4d, 0x82, 0x71, 0xc2, 0x77, 0x09, 0x46, 0x27, 0x99, 0x13, 0x0f, 0xe5,
	0x18, 0x4e, 0x3c, 0x92, 0x63, 0xbc, 0xf0, 0x48, 0x8e, 0xf1, 0xc1, 0x23, 0x39, 0xc6, 0x09, 0x8f,
	0xe5, 0x18, 0x2e, 0x3c, 0x96, 0x63, 0xb8, 0xf1, 0x58, 0x8e, 0x21, 0x89, 0x0d, 0x6c, 0x89, 0x31,
	0x20, 0x00, 0x00, 0xff, 0xff, 0x1a, 0xfe, 0x06, 0xda, 0xd0, 0x00, 0x00, 0x00,
}

func (m *Udec64) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Udec64) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Udec64) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Nanos != 0 {
		i = encodeVarintDecimal(dAtA, i, uint64(m.Nanos))
		i--
		dAtA[i] = 0x10
	}
	if m.Whole != 0 {
		i = encodeVarintDecimal(dAtA, i, uint64(m.Whole))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintDecimal(dAtA []byte, offset int, v uint64) int {
	offset -= sovDecimal(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Udec64) CopyInFields(src *Udec64) int {
	changed := 0
	if m.Whole != src.Whole {
		m.Whole = src.Whole
		changed++
	}
	if m.Nanos != src.Nanos {
		m.Nanos = src.Nanos
		changed++
	}
	return changed
}

func (m *Udec64) DeepCopyIn(src *Udec64) {
	m.Whole = src.Whole
	m.Nanos = src.Nanos
}

// Helper method to check that enums have valid values
func (m *Udec64) ValidateEnums() error {
	return nil
}

func (s *Udec64) ClearTagged(tags map[string]struct{}) {
}

func (m *Udec64) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Whole != 0 {
		n += 1 + sovDecimal(uint64(m.Whole))
	}
	if m.Nanos != 0 {
		n += 1 + sovDecimal(uint64(m.Nanos))
	}
	return n
}

func sovDecimal(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozDecimal(x uint64) (n int) {
	return sovDecimal(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Udec64) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDecimal
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Udec64: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Udec64: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Whole", wireType)
			}
			m.Whole = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDecimal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Whole |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Nanos", wireType)
			}
			m.Nanos = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDecimal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Nanos |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipDecimal(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthDecimal
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthDecimal
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
func skipDecimal(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowDecimal
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
					return 0, ErrIntOverflowDecimal
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowDecimal
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
			if length < 0 {
				return 0, ErrInvalidLengthDecimal
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupDecimal
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthDecimal
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthDecimal        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowDecimal          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupDecimal = fmt.Errorf("proto: unexpected end of group")
)
