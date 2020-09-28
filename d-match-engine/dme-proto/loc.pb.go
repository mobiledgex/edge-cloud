// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: loc.proto

package distributed_match_engine

import (
	encoding_binary "encoding/binary"
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
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

// This is a simple Timestamp message type
// grpc-gateway converts google.protobuf.Timestamp into an RFC3339-type string
// which is a waste of a conversion, so we define our own
type Timestamp struct {
	Seconds              int64    `protobuf:"varint,1,opt,name=seconds,proto3" json:"seconds,omitempty"`
	Nanos                int32    `protobuf:"varint,2,opt,name=nanos,proto3" json:"nanos,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Timestamp) Reset()         { *m = Timestamp{} }
func (m *Timestamp) String() string { return proto.CompactTextString(m) }
func (*Timestamp) ProtoMessage()    {}
func (*Timestamp) Descriptor() ([]byte, []int) {
	return fileDescriptor_1155fb466575c073, []int{0}
}
func (m *Timestamp) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Timestamp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Timestamp.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Timestamp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Timestamp.Merge(m, src)
}
func (m *Timestamp) XXX_Size() int {
	return m.Size()
}
func (m *Timestamp) XXX_DiscardUnknown() {
	xxx_messageInfo_Timestamp.DiscardUnknown(m)
}

var xxx_messageInfo_Timestamp proto.InternalMessageInfo

//
// GPS Location
type Loc struct {
	// latitude in WGS 84 coordinates
	Latitude float64 `protobuf:"fixed64,1,opt,name=latitude,proto3" json:"latitude,omitempty"`
	// longitude in WGS 84 coordinates
	Longitude float64 `protobuf:"fixed64,2,opt,name=longitude,proto3" json:"longitude,omitempty"`
	// horizontal accuracy (radius in meters)
	HorizontalAccuracy float64 `protobuf:"fixed64,3,opt,name=horizontal_accuracy,json=horizontalAccuracy,proto3" json:"horizontal_accuracy,omitempty"`
	// vertical accuracy (meters)
	VerticalAccuracy float64 `protobuf:"fixed64,4,opt,name=vertical_accuracy,json=verticalAccuracy,proto3" json:"vertical_accuracy,omitempty"`
	// On android only lat and long are guaranteed to be supplied
	// altitude in meters
	Altitude float64 `protobuf:"fixed64,5,opt,name=altitude,proto3" json:"altitude,omitempty"`
	// course (IOS) / bearing (Android) (degrees east relative to true north)
	Course float64 `protobuf:"fixed64,6,opt,name=course,proto3" json:"course,omitempty"`
	// speed (IOS) / velocity (Android) (meters/sec)
	Speed float64 `protobuf:"fixed64,7,opt,name=speed,proto3" json:"speed,omitempty"`
	// timestamp
	Timestamp            *Timestamp `protobuf:"bytes,8,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *Loc) Reset()         { *m = Loc{} }
func (m *Loc) String() string { return proto.CompactTextString(m) }
func (*Loc) ProtoMessage()    {}
func (*Loc) Descriptor() ([]byte, []int) {
	return fileDescriptor_1155fb466575c073, []int{1}
}
func (m *Loc) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Loc) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Loc.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Loc) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Loc.Merge(m, src)
}
func (m *Loc) XXX_Size() int {
	return m.Size()
}
func (m *Loc) XXX_DiscardUnknown() {
	xxx_messageInfo_Loc.DiscardUnknown(m)
}

var xxx_messageInfo_Loc proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Timestamp)(nil), "distributed_match_engine.Timestamp")
	proto.RegisterType((*Loc)(nil), "distributed_match_engine.Loc")
}

func init() { proto.RegisterFile("loc.proto", fileDescriptor_1155fb466575c073) }

var fileDescriptor_1155fb466575c073 = []byte{
	// 283 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x91, 0xcd, 0x4a, 0xf4, 0x30,
	0x14, 0x86, 0xbf, 0x4c, 0xbf, 0x76, 0xda, 0xe3, 0x66, 0x8c, 0x22, 0x61, 0x90, 0x52, 0xc6, 0x4d,
	0x41, 0xa8, 0xa0, 0x4b, 0x57, 0xe3, 0xda, 0x55, 0x71, 0x5f, 0x32, 0x69, 0x98, 0x09, 0xb4, 0x49,
	0x49, 0x52, 0x41, 0x2f, 0xc9, 0x2b, 0x99, 0xa5, 0x97, 0xa0, 0xbd, 0x12, 0x99, 0xf4, 0xcf, 0x8d,
	0xcb, 0xe7, 0x3c, 0xef, 0x81, 0x93, 0x37, 0x10, 0x55, 0x8a, 0x65, 0x8d, 0x56, 0x56, 0x61, 0x52,
	0x0a, 0x63, 0xb5, 0xd8, 0xb5, 0x96, 0x97, 0x45, 0x4d, 0x2d, 0x3b, 0x14, 0x5c, 0xee, 0x85, 0xe4,
	0x9b, 0x47, 0x88, 0x5e, 0x44, 0xcd, 0x8d, 0xa5, 0x75, 0x83, 0x09, 0x2c, 0x0d, 0x67, 0x4a, 0x96,
	0x86, 0xa0, 0x04, 0xa5, 0x5e, 0x3e, 0x22, 0xbe, 0x04, 0x5f, 0x52, 0xa9, 0x0c, 0x59, 0x24, 0x28,
	0xf5, 0xf3, 0x1e, 0x36, 0x1f, 0x0b, 0xf0, 0x9e, 0x15, 0xc3, 0x6b, 0x08, 0x2b, 0x6a, 0x85, 0x6d,
	0x4b, 0xee, 0x16, 0x51, 0x3e, 0x31, 0xbe, 0x3e, 0xdd, 0x21, 0xf7, 0xbd, 0x5c, 0x38, 0x39, 0x0f,
	0xf0, 0x1d, 0x5c, 0x1c, 0x94, 0x16, 0xef, 0x4a, 0x5a, 0x5a, 0x15, 0x94, 0xb1, 0x56, 0x53, 0xf6,
	0x46, 0x3c, 0x97, 0xc3, 0xb3, 0xda, 0x0e, 0x06, 0xdf, 0xc2, 0xf9, 0x2b, 0xd7, 0x56, 0xb0, 0xdf,
	0xf1, 0xff, 0x2e, 0xbe, 0x1a, 0xc5, 0x14, 0x5e, 0x43, 0x48, 0xab, 0xe1, 0x2e, 0xbf, 0xbf, 0x6b,
	0x64, 0x7c, 0x05, 0x01, 0x53, 0xad, 0x36, 0x9c, 0x04, 0xce, 0x0c, 0x74, 0x7a, 0xa9, 0x69, 0x38,
	0x2f, 0xc9, 0xd2, 0x8d, 0x7b, 0xc0, 0x5b, 0x88, 0xec, 0x58, 0x13, 0x09, 0x13, 0x94, 0x9e, 0xdd,
	0xdf, 0x64, 0x7f, 0x95, 0x9a, 0x4d, 0x8d, 0xe6, 0xf3, 0xd6, 0xd3, 0xea, 0xf8, 0x1d, 0xff, 0x3b,
	0x76, 0x31, 0xfa, 0xec, 0x62, 0xf4, 0xd5, 0xc5, 0x68, 0x17, 0xb8, 0xcf, 0x79, 0xf8, 0x09, 0x00,
	0x00, 0xff, 0xff, 0x88, 0x54, 0xa7, 0x7d, 0xa9, 0x01, 0x00, 0x00,
}

func (m *Timestamp) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Timestamp) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Timestamp) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.Nanos != 0 {
		i = encodeVarintLoc(dAtA, i, uint64(m.Nanos))
		i--
		dAtA[i] = 0x10
	}
	if m.Seconds != 0 {
		i = encodeVarintLoc(dAtA, i, uint64(m.Seconds))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *Loc) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Loc) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Loc) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.Timestamp != nil {
		{
			size, err := m.Timestamp.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintLoc(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x42
	}
	if m.Speed != 0 {
		i -= 8
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Speed))))
		i--
		dAtA[i] = 0x39
	}
	if m.Course != 0 {
		i -= 8
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Course))))
		i--
		dAtA[i] = 0x31
	}
	if m.Altitude != 0 {
		i -= 8
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Altitude))))
		i--
		dAtA[i] = 0x29
	}
	if m.VerticalAccuracy != 0 {
		i -= 8
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.VerticalAccuracy))))
		i--
		dAtA[i] = 0x21
	}
	if m.HorizontalAccuracy != 0 {
		i -= 8
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.HorizontalAccuracy))))
		i--
		dAtA[i] = 0x19
	}
	if m.Longitude != 0 {
		i -= 8
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Longitude))))
		i--
		dAtA[i] = 0x11
	}
	if m.Latitude != 0 {
		i -= 8
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Latitude))))
		i--
		dAtA[i] = 0x9
	}
	return len(dAtA) - i, nil
}

func encodeVarintLoc(dAtA []byte, offset int, v uint64) int {
	offset -= sovLoc(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Timestamp) CopyInFields(src *Timestamp) int {
	changed := 0
	if m.Seconds != src.Seconds {
		m.Seconds = src.Seconds
		changed++
	}
	if m.Nanos != src.Nanos {
		m.Nanos = src.Nanos
		changed++
	}
	return changed
}

func (m *Timestamp) DeepCopyIn(src *Timestamp) {
	m.Seconds = src.Seconds
	m.Nanos = src.Nanos
}

// Helper method to check that enums have valid values
func (m *Timestamp) ValidateEnums() error {
	return nil
}

func (m *Loc) CopyInFields(src *Loc) int {
	changed := 0
	if m.Latitude != src.Latitude {
		m.Latitude = src.Latitude
		changed++
	}
	if m.Longitude != src.Longitude {
		m.Longitude = src.Longitude
		changed++
	}
	if m.HorizontalAccuracy != src.HorizontalAccuracy {
		m.HorizontalAccuracy = src.HorizontalAccuracy
		changed++
	}
	if m.VerticalAccuracy != src.VerticalAccuracy {
		m.VerticalAccuracy = src.VerticalAccuracy
		changed++
	}
	if m.Altitude != src.Altitude {
		m.Altitude = src.Altitude
		changed++
	}
	if m.Course != src.Course {
		m.Course = src.Course
		changed++
	}
	if m.Speed != src.Speed {
		m.Speed = src.Speed
		changed++
	}
	if src.Timestamp != nil {
		m.Timestamp = &Timestamp{}
		if m.Timestamp.Seconds != src.Timestamp.Seconds {
			m.Timestamp.Seconds = src.Timestamp.Seconds
			changed++
		}
		if m.Timestamp.Nanos != src.Timestamp.Nanos {
			m.Timestamp.Nanos = src.Timestamp.Nanos
			changed++
		}
	} else if m.Timestamp != nil {
		m.Timestamp = nil
		changed++
	}
	return changed
}

func (m *Loc) DeepCopyIn(src *Loc) {
	m.Latitude = src.Latitude
	m.Longitude = src.Longitude
	m.HorizontalAccuracy = src.HorizontalAccuracy
	m.VerticalAccuracy = src.VerticalAccuracy
	m.Altitude = src.Altitude
	m.Course = src.Course
	m.Speed = src.Speed
	if src.Timestamp != nil {
		var tmp_Timestamp Timestamp
		tmp_Timestamp.DeepCopyIn(src.Timestamp)
		m.Timestamp = &tmp_Timestamp
	} else {
		m.Timestamp = nil
	}
}

// Helper method to check that enums have valid values
func (m *Loc) ValidateEnums() error {
	if err := m.Timestamp.ValidateEnums(); err != nil {
		return err
	}
	return nil
}

func (m *Timestamp) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Seconds != 0 {
		n += 1 + sovLoc(uint64(m.Seconds))
	}
	if m.Nanos != 0 {
		n += 1 + sovLoc(uint64(m.Nanos))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *Loc) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Latitude != 0 {
		n += 9
	}
	if m.Longitude != 0 {
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
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovLoc(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozLoc(x uint64) (n int) {
	return sovLoc(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Timestamp) Unmarshal(dAtA []byte) error {
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
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Timestamp: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Timestamp: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Seconds", wireType)
			}
			m.Seconds = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLoc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Seconds |= int64(b&0x7F) << shift
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
					return ErrIntOverflowLoc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Nanos |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipLoc(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthLoc
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthLoc
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
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
			wire |= uint64(b&0x7F) << shift
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
				return fmt.Errorf("proto: wrong wireType = %d for field Latitude", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.Latitude = float64(math.Float64frombits(v))
		case 2:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Longitude", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.Longitude = float64(math.Float64frombits(v))
		case 3:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field HorizontalAccuracy", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
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
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
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
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
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
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
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
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
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
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthLoc
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthLoc
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Timestamp == nil {
				m.Timestamp = &Timestamp{}
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
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthLoc
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
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
	depth := 0
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
		case 1:
			iNdEx += 8
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
			if length < 0 {
				return 0, ErrInvalidLengthLoc
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupLoc
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthLoc
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthLoc        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowLoc          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupLoc = fmt.Errorf("proto: unexpected end of group")
)
