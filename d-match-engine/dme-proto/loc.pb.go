// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: loc.proto

package distributed_match_engine

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/gogo/protobuf/types"

import binary "encoding/binary"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Loc struct {
	// latitude in WGS 84 coordinates
	Lat float64 `protobuf:"fixed64,1,opt,name=lat,proto3" json:"lat,omitempty"`
	// longitude in WGS 84 coordinates
	Long float64 `protobuf:"fixed64,2,opt,name=long,proto3" json:"long,omitempty"`
	// horizontal accuracy (radius in meters)
	HorizontalAccuracy float64 `protobuf:"fixed64,3,opt,name=horizontal_accuracy,json=horizontalAccuracy,proto3" json:"horizontal_accuracy,omitempty"`
	// veritical accuracy (meters)
	VerticalAccuracy float64 `protobuf:"fixed64,4,opt,name=vertical_accuracy,json=verticalAccuracy,proto3" json:"vertical_accuracy,omitempty"`
	// On android only lat and long are guaranteed to be supplied
	// altitude in meters
	Altitude float64 `protobuf:"fixed64,5,opt,name=altitude,proto3" json:"altitude,omitempty"`
	// course (IOS) / bearing (Android) (degrees east relative to true north)
	Course float64 `protobuf:"fixed64,6,opt,name=course,proto3" json:"course,omitempty"`
	// speed (IOS) / velocity (Android) (meters/sec)
	Speed float64 `protobuf:"fixed64,7,opt,name=speed,proto3" json:"speed,omitempty"`
	// timestamp
	Timestamp *google_protobuf.Timestamp `protobuf:"bytes,8,opt,name=timestamp" json:"timestamp,omitempty"`
}

func (m *Loc) Reset()                    { *m = Loc{} }
func (m *Loc) String() string            { return proto.CompactTextString(m) }
func (*Loc) ProtoMessage()               {}
func (*Loc) Descriptor() ([]byte, []int) { return fileDescriptorLoc, []int{0} }

func init() {
	proto.RegisterType((*Loc)(nil), "distributed_match_engine.Loc")
}
func (m *Loc) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Loc) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Lat != 0 {
		dAtA[i] = 0x9
		i++
		binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Lat))))
		i += 8
	}
	if m.Long != 0 {
		dAtA[i] = 0x11
		i++
		binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Long))))
		i += 8
	}
	if m.HorizontalAccuracy != 0 {
		dAtA[i] = 0x19
		i++
		binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.HorizontalAccuracy))))
		i += 8
	}
	if m.VerticalAccuracy != 0 {
		dAtA[i] = 0x21
		i++
		binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.VerticalAccuracy))))
		i += 8
	}
	if m.Altitude != 0 {
		dAtA[i] = 0x29
		i++
		binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Altitude))))
		i += 8
	}
	if m.Course != 0 {
		dAtA[i] = 0x31
		i++
		binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Course))))
		i += 8
	}
	if m.Speed != 0 {
		dAtA[i] = 0x39
		i++
		binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Speed))))
		i += 8
	}
	if m.Timestamp != nil {
		dAtA[i] = 0x42
		i++
		i = encodeVarintLoc(dAtA, i, uint64(m.Timestamp.Size()))
		n1, err := m.Timestamp.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	return i, nil
}

func encodeVarintLoc(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Loc) CopyInFields(src *Loc) {
	m.Lat = src.Lat
	m.Long = src.Long
	m.HorizontalAccuracy = src.HorizontalAccuracy
	m.VerticalAccuracy = src.VerticalAccuracy
	m.Altitude = src.Altitude
	m.Course = src.Course
	m.Speed = src.Speed
	if src.Timestamp != nil {
		m.Timestamp = &google_protobuf.Timestamp{}
		m.Timestamp.Seconds = src.Timestamp.Seconds
		m.Timestamp.Nanos = src.Timestamp.Nanos
	}
}

func (m *Loc) Size() (n int) {
	var l int
	_ = l
	if m.Lat != 0 {
		n += 9
	}
	if m.Long != 0 {
		n += 9
	}
	if m.HorizontalAccuracy != 0 {
		n += 9
	}
	if m.VerticalAccuracy != 0 {
		n += 9
	}
	if m.Altitude != 0 {
		n += 9
	}
	if m.Course != 0 {
		n += 9
	}
	if m.Speed != 0 {
		n += 9
	}
	if m.Timestamp != nil {
		l = m.Timestamp.Size()
		n += 1 + l + sovLoc(uint64(l))
	}
	return n
}

func sovLoc(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozLoc(x uint64) (n int) {
	return sovLoc(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Loc) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLoc
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
			return fmt.Errorf("proto: Loc: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Loc: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Lat", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.Lat = float64(math.Float64frombits(v))
		case 2:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Long", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.Long = float64(math.Float64frombits(v))
		case 3:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field HorizontalAccuracy", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.HorizontalAccuracy = float64(math.Float64frombits(v))
		case 4:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field VerticalAccuracy", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.VerticalAccuracy = float64(math.Float64frombits(v))
		case 5:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Altitude", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.Altitude = float64(math.Float64frombits(v))
		case 6:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Course", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.Course = float64(math.Float64frombits(v))
		case 7:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Speed", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.Speed = float64(math.Float64frombits(v))
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timestamp", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLoc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthLoc
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Timestamp == nil {
				m.Timestamp = &google_protobuf.Timestamp{}
			}
			if err := m.Timestamp.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipLoc(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthLoc
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
func skipLoc(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowLoc
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
					return 0, ErrIntOverflowLoc
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
					return 0, ErrIntOverflowLoc
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
				return 0, ErrInvalidLengthLoc
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowLoc
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
				next, err := skipLoc(dAtA[start:])
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
	ErrInvalidLengthLoc = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowLoc   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("loc.proto", fileDescriptorLoc) }

var fileDescriptorLoc = []byte{
	// 265 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x90, 0xb1, 0x4e, 0xc3, 0x30,
	0x10, 0x86, 0x71, 0xd3, 0x86, 0xd6, 0x2c, 0xc1, 0x20, 0x64, 0x65, 0x08, 0x15, 0x53, 0x25, 0xa4,
	0x44, 0x82, 0x85, 0x15, 0x66, 0xa6, 0x8a, 0x3d, 0x72, 0x1c, 0x93, 0x5a, 0x72, 0x72, 0x91, 0x73,
	0x46, 0x82, 0x77, 0xe0, 0xbd, 0x3a, 0xf2, 0x08, 0x90, 0x27, 0x41, 0x38, 0x4d, 0xca, 0x76, 0xff,
	0x7d, 0xdf, 0x0d, 0xf7, 0xd3, 0x95, 0x01, 0x99, 0xb6, 0x16, 0x10, 0x18, 0x2f, 0x75, 0x87, 0x56,
	0x17, 0x0e, 0x55, 0x99, 0xd7, 0x02, 0xe5, 0x2e, 0x57, 0x4d, 0xa5, 0x1b, 0x15, 0x5f, 0x57, 0x00,
	0x95, 0x51, 0x99, 0xf7, 0x0a, 0xf7, 0x9a, 0xa1, 0xae, 0x55, 0x87, 0xa2, 0x6e, 0x87, 0xd3, 0x9b,
	0xcf, 0x19, 0x0d, 0x9e, 0x41, 0xb2, 0x88, 0x06, 0x46, 0x20, 0x27, 0x6b, 0xb2, 0x21, 0xdb, 0xbf,
	0x91, 0x31, 0x3a, 0x37, 0xd0, 0x54, 0x7c, 0xe6, 0x57, 0x7e, 0x66, 0x19, 0xbd, 0xd8, 0x81, 0xd5,
	0x1f, 0xd0, 0xa0, 0x30, 0xb9, 0x90, 0xd2, 0x59, 0x21, 0xdf, 0x79, 0xe0, 0x15, 0x76, 0x44, 0x8f,
	0x07, 0xc2, 0x6e, 0xe9, 0xf9, 0x9b, 0xb2, 0xa8, 0xe5, 0x7f, 0x7d, 0xee, 0xf5, 0x68, 0x04, 0x93,
	0x1c, 0xd3, 0xa5, 0x30, 0xa8, 0xd1, 0x95, 0x8a, 0x2f, 0xbc, 0x33, 0x65, 0x76, 0x45, 0x43, 0x09,
	0xce, 0x76, 0x8a, 0x87, 0x9e, 0x1c, 0x12, 0xbb, 0xa4, 0x8b, 0xae, 0x55, 0xaa, 0xe4, 0xa7, 0x7e,
	0x3d, 0x04, 0xf6, 0x40, 0x57, 0xd3, 0xa3, 0x7c, 0xb9, 0x26, 0x9b, 0xb3, 0xbb, 0x38, 0x1d, 0xaa,
	0x48, 0xc7, 0x2a, 0xd2, 0x97, 0xd1, 0xd8, 0x1e, 0xe5, 0xa7, 0x68, 0xff, 0x93, 0x9c, 0xec, 0xfb,
	0x84, 0x7c, 0xf5, 0x09, 0xf9, 0xee, 0x13, 0x52, 0x84, 0xfe, 0xe0, 0xfe, 0x37, 0x00, 0x00, 0xff,
	0xff, 0xfd, 0x03, 0x4d, 0xf2, 0x70, 0x01, 0x00, 0x00,
}
