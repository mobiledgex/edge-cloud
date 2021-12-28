// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: metric.proto

package edgeproto

import (
	encoding_binary "encoding/binary"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	types "github.com/gogo/protobuf/types"
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

// MetricTag is used as a tag or label to look up the metric, beyond just the name of the metric.
type MetricTag struct {
	// Metric tag name
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Metric tag value
	Val string `protobuf:"bytes,2,opt,name=val,proto3" json:"val,omitempty"`
}

func (m *MetricTag) Reset()         { *m = MetricTag{} }
func (m *MetricTag) String() string { return proto.CompactTextString(m) }
func (*MetricTag) ProtoMessage()    {}
func (*MetricTag) Descriptor() ([]byte, []int) {
	return fileDescriptor_da41641f55bff5df, []int{0}
}
func (m *MetricTag) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MetricTag) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MetricTag.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MetricTag) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MetricTag.Merge(m, src)
}
func (m *MetricTag) XXX_Size() int {
	return m.Size()
}
func (m *MetricTag) XXX_DiscardUnknown() {
	xxx_messageInfo_MetricTag.DiscardUnknown(m)
}

var xxx_messageInfo_MetricTag proto.InternalMessageInfo

// MetricVal is a value associated with the metric.
type MetricVal struct {
	// Name of the value
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Value of the Value.
	//
	// Types that are valid to be assigned to Value:
	//	*MetricVal_Dval
	//	*MetricVal_Ival
	//	*MetricVal_Bval
	//	*MetricVal_Sval
	Value isMetricVal_Value `protobuf_oneof:"value"`
}

func (m *MetricVal) Reset()         { *m = MetricVal{} }
func (m *MetricVal) String() string { return proto.CompactTextString(m) }
func (*MetricVal) ProtoMessage()    {}
func (*MetricVal) Descriptor() ([]byte, []int) {
	return fileDescriptor_da41641f55bff5df, []int{1}
}
func (m *MetricVal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MetricVal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MetricVal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MetricVal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MetricVal.Merge(m, src)
}
func (m *MetricVal) XXX_Size() int {
	return m.Size()
}
func (m *MetricVal) XXX_DiscardUnknown() {
	xxx_messageInfo_MetricVal.DiscardUnknown(m)
}

var xxx_messageInfo_MetricVal proto.InternalMessageInfo

type isMetricVal_Value interface {
	isMetricVal_Value()
	MarshalTo([]byte) (int, error)
	Size() int
}

type MetricVal_Dval struct {
	Dval float64 `protobuf:"fixed64,2,opt,name=dval,proto3,oneof" json:"dval,omitempty"`
}
type MetricVal_Ival struct {
	Ival uint64 `protobuf:"varint,3,opt,name=ival,proto3,oneof" json:"ival,omitempty"`
}
type MetricVal_Bval struct {
	Bval bool `protobuf:"varint,4,opt,name=bval,proto3,oneof" json:"bval,omitempty"`
}
type MetricVal_Sval struct {
	Sval string `protobuf:"bytes,5,opt,name=sval,proto3,oneof" json:"sval,omitempty"`
}

func (*MetricVal_Dval) isMetricVal_Value() {}
func (*MetricVal_Ival) isMetricVal_Value() {}
func (*MetricVal_Bval) isMetricVal_Value() {}
func (*MetricVal_Sval) isMetricVal_Value() {}

func (m *MetricVal) GetValue() isMetricVal_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *MetricVal) GetDval() float64 {
	if x, ok := m.GetValue().(*MetricVal_Dval); ok {
		return x.Dval
	}
	return 0
}

func (m *MetricVal) GetIval() uint64 {
	if x, ok := m.GetValue().(*MetricVal_Ival); ok {
		return x.Ival
	}
	return 0
}

func (m *MetricVal) GetBval() bool {
	if x, ok := m.GetValue().(*MetricVal_Bval); ok {
		return x.Bval
	}
	return false
}

func (m *MetricVal) GetSval() string {
	if x, ok := m.GetValue().(*MetricVal_Sval); ok {
		return x.Sval
	}
	return ""
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*MetricVal) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*MetricVal_Dval)(nil),
		(*MetricVal_Ival)(nil),
		(*MetricVal_Bval)(nil),
		(*MetricVal_Sval)(nil),
	}
}

// Metric is an entry/point in a time series of values for Analytics/Billing.
type Metric struct {
	// Metric name
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Timestamp when the metric was captured
	Timestamp types.Timestamp `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp"`
	// Tags associated with the metric for searching/filtering
	Tags []*MetricTag `protobuf:"bytes,3,rep,name=tags,proto3" json:"tags,omitempty"`
	// Values associated with the metric
	Vals []*MetricVal `protobuf:"bytes,4,rep,name=vals,proto3" json:"vals,omitempty"`
}

func (m *Metric) Reset()         { *m = Metric{} }
func (m *Metric) String() string { return proto.CompactTextString(m) }
func (*Metric) ProtoMessage()    {}
func (*Metric) Descriptor() ([]byte, []int) {
	return fileDescriptor_da41641f55bff5df, []int{2}
}
func (m *Metric) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Metric) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Metric.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Metric) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Metric.Merge(m, src)
}
func (m *Metric) XXX_Size() int {
	return m.Size()
}
func (m *Metric) XXX_DiscardUnknown() {
	xxx_messageInfo_Metric.DiscardUnknown(m)
}

var xxx_messageInfo_Metric proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MetricTag)(nil), "edgeproto.MetricTag")
	proto.RegisterType((*MetricVal)(nil), "edgeproto.MetricVal")
	proto.RegisterType((*Metric)(nil), "edgeproto.Metric")
}

func init() { proto.RegisterFile("metric.proto", fileDescriptor_da41641f55bff5df) }

var fileDescriptor_da41641f55bff5df = []byte{
	// 350 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x91, 0xb1, 0x6e, 0xea, 0x30,
	0x14, 0x86, 0xed, 0x1b, 0xc3, 0x25, 0xe6, 0x0e, 0x57, 0x11, 0x43, 0x84, 0x2a, 0x13, 0x31, 0x65,
	0xa9, 0xa3, 0xd2, 0xa5, 0xea, 0xd0, 0x81, 0xa9, 0x4b, 0x97, 0x08, 0xb1, 0x3b, 0xe0, 0xba, 0x91,
	0x1c, 0x8c, 0x48, 0x82, 0x3a, 0xf6, 0x11, 0xfa, 0x3c, 0x9d, 0x3a, 0x32, 0x32, 0x76, 0xaa, 0x5a,
	0x78, 0x81, 0x0e, 0x3c, 0x40, 0x75, 0x1c, 0x92, 0x2e, 0x74, 0x89, 0xfe, 0xf3, 0xe5, 0xff, 0x8f,
	0x8f, 0x8f, 0xe9, 0xbf, 0x4c, 0x16, 0xab, 0x74, 0xc6, 0x97, 0x2b, 0x53, 0x18, 0xcf, 0x95, 0x73,
	0x25, 0xad, 0xec, 0x5f, 0xa9, 0xb4, 0x78, 0x28, 0x13, 0x3e, 0x33, 0x59, 0x94, 0x99, 0x24, 0xd5,
	0xf0, 0xeb, 0x31, 0x82, 0xef, 0xf9, 0x4c, 0x9b, 0x72, 0x1e, 0x59, 0x9f, 0x92, 0x8b, 0x46, 0x54,
	0x4d, 0xfa, 0x3d, 0x65, 0x94, 0xb1, 0x32, 0x02, 0x75, 0xa4, 0x03, 0x65, 0x8c, 0xd2, 0xb2, 0x32,
	0x27, 0xe5, 0x7d, 0x54, 0xa4, 0x99, 0xcc, 0x0b, 0x91, 0x2d, 0x2b, 0xc3, 0xf0, 0x82, 0xba, 0x77,
	0x76, 0x96, 0x89, 0x50, 0x9e, 0x47, 0xc9, 0x42, 0x64, 0xd2, 0xc7, 0x01, 0x0e, 0xdd, 0xd8, 0x6a,
	0xef, 0x3f, 0x75, 0xd6, 0x42, 0xfb, 0x7f, 0x2c, 0x02, 0x39, 0x7c, 0xc2, 0x75, 0x66, 0x2a, 0xf4,
	0xc9, 0x4c, 0x8f, 0x92, 0x79, 0x1d, 0xc2, 0xb7, 0x28, 0xb6, 0x15, 0xd0, 0x14, 0xa8, 0x13, 0xe0,
	0x90, 0x00, 0x4d, 0x8f, 0x34, 0x01, 0x4a, 0x02, 0x1c, 0x76, 0x80, 0x26, 0x47, 0x9a, 0x03, 0x6d,
	0x41, 0x57, 0xa0, 0x50, 0x8d, 0xff, 0xd2, 0xd6, 0x5a, 0xe8, 0x52, 0x0e, 0x5f, 0x30, 0x6d, 0x57,
	0x23, 0x9c, 0x3c, 0xff, 0x86, 0xba, 0xcd, 0x3d, 0xed, 0x10, 0xdd, 0x51, 0x9f, 0x57, 0x9b, 0xe0,
	0xf5, 0x26, 0xf8, 0xa4, 0x76, 0x8c, 0xc9, 0xe6, 0x7d, 0x80, 0xe2, 0x9f, 0x88, 0x17, 0x52, 0x52,
	0x08, 0x95, 0xfb, 0x4e, 0xe0, 0x84, 0xdd, 0x51, 0x8f, 0x37, 0xef, 0xc3, 0x9b, 0x5d, 0xc5, 0xd6,
	0x01, 0xce, 0xb5, 0xd0, 0xb9, 0x4f, 0x7e, 0x71, 0x4e, 0x85, 0x8e, 0xad, 0xe3, 0xba, 0xf3, 0x7a,
	0xf0, 0xf1, 0xd7, 0xc1, 0x47, 0xe3, 0xb3, 0xcd, 0x27, 0x43, 0x9b, 0x1d, 0xc3, 0xdb, 0x1d, 0xc3,
	0x1f, 0x3b, 0x86, 0x9f, 0xf7, 0x0c, 0x6d, 0xf7, 0x0c, 0xbd, 0xed, 0x19, 0x4a, 0xda, 0x36, 0x7e,
	0xf9, 0x1d, 0x00, 0x00, 0xff, 0xff, 0x36, 0xaf, 0x73, 0x1e, 0x23, 0x02, 0x00, 0x00,
}

func (m *MetricTag) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MetricTag) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MetricTag) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Val) > 0 {
		i -= len(m.Val)
		copy(dAtA[i:], m.Val)
		i = encodeVarintMetric(dAtA, i, uint64(len(m.Val)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintMetric(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MetricVal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MetricVal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MetricVal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Value != nil {
		{
			size := m.Value.Size()
			i -= size
			if _, err := m.Value.MarshalTo(dAtA[i:]); err != nil {
				return 0, err
			}
		}
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintMetric(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MetricVal_Dval) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MetricVal_Dval) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	i -= 8
	encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Dval))))
	i--
	dAtA[i] = 0x11
	return len(dAtA) - i, nil
}
func (m *MetricVal_Ival) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MetricVal_Ival) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	i = encodeVarintMetric(dAtA, i, uint64(m.Ival))
	i--
	dAtA[i] = 0x18
	return len(dAtA) - i, nil
}
func (m *MetricVal_Bval) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MetricVal_Bval) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	i--
	if m.Bval {
		dAtA[i] = 1
	} else {
		dAtA[i] = 0
	}
	i--
	dAtA[i] = 0x20
	return len(dAtA) - i, nil
}
func (m *MetricVal_Sval) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MetricVal_Sval) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	i -= len(m.Sval)
	copy(dAtA[i:], m.Sval)
	i = encodeVarintMetric(dAtA, i, uint64(len(m.Sval)))
	i--
	dAtA[i] = 0x2a
	return len(dAtA) - i, nil
}
func (m *Metric) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Metric) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Metric) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Vals) > 0 {
		for iNdEx := len(m.Vals) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Vals[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintMetric(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.Tags) > 0 {
		for iNdEx := len(m.Tags) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Tags[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintMetric(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	{
		size, err := m.Timestamp.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintMetric(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintMetric(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintMetric(dAtA []byte, offset int, v uint64) int {
	offset -= sovMetric(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MetricTag) CopyInFields(src *MetricTag) int {
	changed := 0
	if m.Name != src.Name {
		m.Name = src.Name
		changed++
	}
	if m.Val != src.Val {
		m.Val = src.Val
		changed++
	}
	return changed
}

func (m *MetricTag) DeepCopyIn(src *MetricTag) {
	m.Name = src.Name
	m.Val = src.Val
}

// Helper method to check that enums have valid values
func (m *MetricTag) ValidateEnums() error {
	return nil
}

func (s *MetricTag) ClearTagged(tags map[string]struct{}) {
}

func (m *MetricVal) CopyInFields(src *MetricVal) int {
	changed := 0
	if m.Name != src.Name {
		m.Name = src.Name
		changed++
	}
	return changed
}

func (m *MetricVal) DeepCopyIn(src *MetricVal) {
	m.Name = src.Name
}

// Helper method to check that enums have valid values
func (m *MetricVal) ValidateEnums() error {
	return nil
}

func (s *MetricVal) ClearTagged(tags map[string]struct{}) {
}

func (m *Metric) CopyInFields(src *Metric) int {
	changed := 0
	if m.Name != src.Name {
		m.Name = src.Name
		changed++
	}
	if m.Timestamp.Seconds != src.Timestamp.Seconds {
		m.Timestamp.Seconds = src.Timestamp.Seconds
		changed++
	}
	if m.Timestamp.Nanos != src.Timestamp.Nanos {
		m.Timestamp.Nanos = src.Timestamp.Nanos
		changed++
	}
	if src.Tags != nil {
		m.Tags = src.Tags
		changed++
	} else if m.Tags != nil {
		m.Tags = nil
		changed++
	}
	if src.Vals != nil {
		m.Vals = src.Vals
		changed++
	} else if m.Vals != nil {
		m.Vals = nil
		changed++
	}
	return changed
}

func (m *Metric) DeepCopyIn(src *Metric) {
	m.Name = src.Name
	m.Timestamp = src.Timestamp
	if src.Tags != nil {
		m.Tags = make([]*MetricTag, len(src.Tags), len(src.Tags))
		for ii, s := range src.Tags {
			var tmp_s MetricTag
			tmp_s.DeepCopyIn(s)
			m.Tags[ii] = &tmp_s
		}
	} else {
		m.Tags = nil
	}
	if src.Vals != nil {
		m.Vals = make([]*MetricVal, len(src.Vals), len(src.Vals))
		for ii, s := range src.Vals {
			var tmp_s MetricVal
			tmp_s.DeepCopyIn(s)
			m.Vals[ii] = &tmp_s
		}
	} else {
		m.Vals = nil
	}
}

// Helper method to check that enums have valid values
func (m *Metric) ValidateEnums() error {
	for _, e := range m.Tags {
		if err := e.ValidateEnums(); err != nil {
			return err
		}
	}
	for _, e := range m.Vals {
		if err := e.ValidateEnums(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Metric) ClearTagged(tags map[string]struct{}) {
	if s.Tags != nil {
		for ii := 0; ii < len(s.Tags); ii++ {
			s.Tags[ii].ClearTagged(tags)
		}
	}
	if s.Vals != nil {
		for ii := 0; ii < len(s.Vals); ii++ {
			s.Vals[ii].ClearTagged(tags)
		}
	}
}

func (m *MetricTag) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovMetric(uint64(l))
	}
	l = len(m.Val)
	if l > 0 {
		n += 1 + l + sovMetric(uint64(l))
	}
	return n
}

func (m *MetricVal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovMetric(uint64(l))
	}
	if m.Value != nil {
		n += m.Value.Size()
	}
	return n
}

func (m *MetricVal_Dval) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 9
	return n
}
func (m *MetricVal_Ival) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 1 + sovMetric(uint64(m.Ival))
	return n
}
func (m *MetricVal_Bval) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 2
	return n
}
func (m *MetricVal_Sval) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Sval)
	n += 1 + l + sovMetric(uint64(l))
	return n
}
func (m *Metric) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovMetric(uint64(l))
	}
	l = m.Timestamp.Size()
	n += 1 + l + sovMetric(uint64(l))
	if len(m.Tags) > 0 {
		for _, e := range m.Tags {
			l = e.Size()
			n += 1 + l + sovMetric(uint64(l))
		}
	}
	if len(m.Vals) > 0 {
		for _, e := range m.Vals {
			l = e.Size()
			n += 1 + l + sovMetric(uint64(l))
		}
	}
	return n
}

func sovMetric(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozMetric(x uint64) (n int) {
	return sovMetric(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MetricTag) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMetric
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
			return fmt.Errorf("proto: MetricTag: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MetricTag: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetric
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMetric
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMetric
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Val", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetric
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMetric
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMetric
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Val = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMetric(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMetric
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthMetric
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
func (m *MetricVal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMetric
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
			return fmt.Errorf("proto: MetricVal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MetricVal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetric
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMetric
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMetric
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Dval", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.Value = &MetricVal_Dval{float64(math.Float64frombits(v))}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ival", wireType)
			}
			var v uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetric
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Value = &MetricVal_Ival{v}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Bval", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetric
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			b := bool(v != 0)
			m.Value = &MetricVal_Bval{b}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sval", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetric
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMetric
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMetric
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Value = &MetricVal_Sval{string(dAtA[iNdEx:postIndex])}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMetric(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMetric
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthMetric
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
func (m *Metric) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMetric
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
			return fmt.Errorf("proto: Metric: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Metric: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetric
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMetric
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMetric
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timestamp", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetric
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
				return ErrInvalidLengthMetric
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMetric
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Timestamp.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Tags", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetric
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
				return ErrInvalidLengthMetric
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMetric
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Tags = append(m.Tags, &MetricTag{})
			if err := m.Tags[len(m.Tags)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Vals", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetric
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
				return ErrInvalidLengthMetric
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMetric
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Vals = append(m.Vals, &MetricVal{})
			if err := m.Vals[len(m.Vals)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMetric(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMetric
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthMetric
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
func skipMetric(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMetric
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
					return 0, ErrIntOverflowMetric
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
					return 0, ErrIntOverflowMetric
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
				return 0, ErrInvalidLengthMetric
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupMetric
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthMetric
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthMetric        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMetric          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupMetric = fmt.Errorf("proto: unexpected end of group")
)
