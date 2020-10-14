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

// Latency
type Latency struct {
	Avg float64 `protobuf:"fixed64,1,opt,name=avg,proto3" json:"avg,omitempty"`
	Min float64 `protobuf:"fixed64,2,opt,name=min,proto3" json:"min,omitempty"`
	Max float64 `protobuf:"fixed64,3,opt,name=max,proto3" json:"max,omitempty"`
	// Unbiased standard deviation
	StdDev float64 `protobuf:"fixed64,4,opt,name=std_dev,json=stdDev,proto3" json:"std_dev,omitempty"`
	// Unbiased variance
	Variance             float64    `protobuf:"fixed64,5,opt,name=variance,proto3" json:"variance,omitempty"`
	NumSamples           uint64     `protobuf:"varint,6,opt,name=num_samples,json=numSamples,proto3" json:"num_samples,omitempty"`
	Timestamp            *Timestamp `protobuf:"bytes,7,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *Latency) Reset()         { *m = Latency{} }
func (m *Latency) String() string { return proto.CompactTextString(m) }
func (*Latency) ProtoMessage()    {}
func (*Latency) Descriptor() ([]byte, []int) {
	return fileDescriptor_1155fb466575c073, []int{2}
}
func (m *Latency) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Latency) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Latency.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Latency) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Latency.Merge(m, src)
}
func (m *Latency) XXX_Size() int {
	return m.Size()
}
func (m *Latency) XXX_DiscardUnknown() {
	xxx_messageInfo_Latency.DiscardUnknown(m)
}

var xxx_messageInfo_Latency proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Timestamp)(nil), "distributed_match_engine.Timestamp")
	proto.RegisterType((*Loc)(nil), "distributed_match_engine.Loc")
	proto.RegisterType((*Latency)(nil), "distributed_match_engine.Latency")
}

func init() { proto.RegisterFile("loc.proto", fileDescriptor_1155fb466575c073) }

var fileDescriptor_1155fb466575c073 = []byte{
	// 372 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x92, 0x3f, 0x4f, 0xe3, 0x30,
	0x18, 0xc6, 0xcf, 0x4d, 0x9b, 0xb4, 0x6f, 0x97, 0x9e, 0xef, 0x74, 0x67, 0x55, 0xa7, 0x5c, 0x55,
	0x96, 0x4a, 0x48, 0x45, 0x82, 0x91, 0xa9, 0x88, 0xb1, 0x53, 0x60, 0x8f, 0x5c, 0xdb, 0x6a, 0x2d,
	0x25, 0x76, 0x14, 0x3b, 0x11, 0xe5, 0x23, 0xf1, 0x49, 0x3a, 0x32, 0x32, 0x42, 0x3f, 0x09, 0x8a,
	0xf3, 0xa7, 0x30, 0xb0, 0xb0, 0xbd, 0xcf, 0xf3, 0xbc, 0x96, 0xde, 0xdf, 0x23, 0xc3, 0x28, 0xd1,
	0x6c, 0x99, 0xe5, 0xda, 0x6a, 0x4c, 0xb8, 0x34, 0x36, 0x97, 0x9b, 0xc2, 0x0a, 0x1e, 0xa7, 0xd4,
	0xb2, 0x5d, 0x2c, 0xd4, 0x56, 0x2a, 0x31, 0xbf, 0x86, 0xd1, 0xbd, 0x4c, 0x85, 0xb1, 0x34, 0xcd,
	0x30, 0x81, 0xc0, 0x08, 0xa6, 0x15, 0x37, 0x04, 0xcd, 0xd0, 0xc2, 0x8b, 0x5a, 0x89, 0x7f, 0xc3,
	0x40, 0x51, 0xa5, 0x0d, 0xe9, 0xcd, 0xd0, 0x62, 0x10, 0xd5, 0x62, 0xfe, 0xd4, 0x03, 0x6f, 0xad,
	0x19, 0x9e, 0xc2, 0x30, 0xa1, 0x56, 0xda, 0x82, 0x0b, 0xf7, 0x10, 0x45, 0x9d, 0xc6, 0xff, 0xaa,
	0x3b, 0xd4, 0xb6, 0x0e, 0x7b, 0x2e, 0x3c, 0x19, 0xf8, 0x02, 0x7e, 0xed, 0x74, 0x2e, 0x1f, 0xb5,
	0xb2, 0x34, 0x89, 0x29, 0x63, 0x45, 0x4e, 0xd9, 0x9e, 0x78, 0x6e, 0x0f, 0x9f, 0xa2, 0x55, 0x93,
	0xe0, 0x73, 0xf8, 0x59, 0x8a, 0xdc, 0x4a, 0xf6, 0x71, 0xbd, 0xef, 0xd6, 0x27, 0x6d, 0xd0, 0x2d,
	0x4f, 0x61, 0x48, 0x93, 0xe6, 0xae, 0x41, 0x7d, 0x57, 0xab, 0xf1, 0x1f, 0xf0, 0x99, 0x2e, 0x72,
	0x23, 0x88, 0xef, 0x92, 0x46, 0x55, 0xa4, 0x26, 0x13, 0x82, 0x93, 0xc0, 0xd9, 0xb5, 0xc0, 0x2b,
	0x18, 0xd9, 0xb6, 0x26, 0x32, 0x9c, 0xa1, 0xc5, 0xf8, 0xf2, 0x6c, 0xf9, 0x55, 0xa9, 0xcb, 0xae,
	0xd1, 0xe8, 0xf4, 0x6a, 0xfe, 0x82, 0x20, 0x58, 0x53, 0x2b, 0x14, 0xdb, 0xe3, 0x09, 0x78, 0xb4,
	0xdc, 0x36, 0x5d, 0x55, 0x63, 0xe5, 0xa4, 0x52, 0x35, 0x05, 0x55, 0xa3, 0x73, 0xe8, 0x43, 0x53,
	0x45, 0x35, 0xe2, 0xbf, 0x10, 0x18, 0xcb, 0x63, 0x2e, 0xca, 0x86, 0xd8, 0x37, 0x96, 0xdf, 0x8a,
	0xb2, 0xe2, 0x2c, 0x69, 0x2e, 0xa9, 0x62, 0x1d, 0x67, 0xab, 0xf1, 0x7f, 0x18, 0xab, 0x22, 0x8d,
	0x0d, 0x4d, 0xb3, 0x44, 0x18, 0x07, 0xdb, 0x8f, 0x40, 0x15, 0xe9, 0x5d, 0xed, 0x7c, 0x46, 0x0b,
	0xbe, 0x83, 0x76, 0x33, 0x39, 0xbc, 0x85, 0x3f, 0x0e, 0xc7, 0x10, 0x3d, 0x1f, 0x43, 0xf4, 0x7a,
	0x0c, 0xd1, 0xc6, 0x77, 0xff, 0xee, 0xea, 0x3d, 0x00, 0x00, 0xff, 0xff, 0xe6, 0x78, 0xe3, 0xca,
	0x84, 0x02, 0x00, 0x00,
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

func (m *Latency) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Latency) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Latency) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
		dAtA[i] = 0x3a
	}
	if m.NumSamples != 0 {
		i = encodeVarintLoc(dAtA, i, uint64(m.NumSamples))
		i--
		dAtA[i] = 0x30
	}
	if m.Variance != 0 {
		i -= 8
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Variance))))
		i--
		dAtA[i] = 0x29
	}
	if m.StdDev != 0 {
		i -= 8
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.StdDev))))
		i--
		dAtA[i] = 0x21
	}
	if m.Max != 0 {
		i -= 8
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Max))))
		i--
		dAtA[i] = 0x19
	}
	if m.Min != 0 {
		i -= 8
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Min))))
		i--
		dAtA[i] = 0x11
	}
	if m.Avg != 0 {
		i -= 8
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Avg))))
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

func (m *Latency) CopyInFields(src *Latency) int {
	changed := 0
	if m.Avg != src.Avg {
		m.Avg = src.Avg
		changed++
	}
	if m.Min != src.Min {
		m.Min = src.Min
		changed++
	}
	if m.Max != src.Max {
		m.Max = src.Max
		changed++
	}
	if m.StdDev != src.StdDev {
		m.StdDev = src.StdDev
		changed++
	}
	if m.Variance != src.Variance {
		m.Variance = src.Variance
		changed++
	}
	if m.NumSamples != src.NumSamples {
		m.NumSamples = src.NumSamples
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

func (m *Latency) DeepCopyIn(src *Latency) {
	m.Avg = src.Avg
	m.Min = src.Min
	m.Max = src.Max
	m.StdDev = src.StdDev
	m.Variance = src.Variance
	m.NumSamples = src.NumSamples
	if src.Timestamp != nil {
		var tmp_Timestamp Timestamp
		tmp_Timestamp.DeepCopyIn(src.Timestamp)
		m.Timestamp = &tmp_Timestamp
	} else {
		m.Timestamp = nil
	}
}

// Helper method to check that enums have valid values
func (m *Latency) ValidateEnums() error {
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

func (m *Latency) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Avg != 0 {
		n += 9
	}
	if m.Min != 0 {
		n += 9
	}
	if m.Max != 0 {
		n += 9
	}
	if m.StdDev != 0 {
		n += 9
	}
	if m.Variance != 0 {
		n += 9
	}
	if m.NumSamples != 0 {
		n += 1 + sovLoc(uint64(m.NumSamples))
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
func (m *Latency) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: Latency: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Latency: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Avg", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.Avg = float64(math.Float64frombits(v))
		case 2:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Min", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.Min = float64(math.Float64frombits(v))
		case 3:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Max", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.Max = float64(math.Float64frombits(v))
		case 4:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field StdDev", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.StdDev = float64(math.Float64frombits(v))
		case 5:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Variance", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.Variance = float64(math.Float64frombits(v))
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NumSamples", wireType)
			}
			m.NumSamples = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLoc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NumSamples |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
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
