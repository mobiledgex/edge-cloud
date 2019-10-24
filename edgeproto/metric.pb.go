// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: metric.proto

package edgeproto

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/gogo/protobuf/gogoproto"
import google_protobuf1 "github.com/gogo/protobuf/types"

import binary "encoding/binary"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// MetricTag is used as a tag or label to look up the metric, beyond just the name of the metric.
type MetricTag struct {
	// Metric tag name
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Metric tag value
	Val string `protobuf:"bytes,2,opt,name=val,proto3" json:"val,omitempty"`
}

func (m *MetricTag) Reset()                    { *m = MetricTag{} }
func (m *MetricTag) String() string            { return proto.CompactTextString(m) }
func (*MetricTag) ProtoMessage()               {}
func (*MetricTag) Descriptor() ([]byte, []int) { return fileDescriptorMetric, []int{0} }

// MetricVal is a value associated with the metric.
type MetricVal struct {
	// Name of the value
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Value of the Value.
	//
	// Types that are valid to be assigned to Value:
	//	*MetricVal_Dval
	//	*MetricVal_Ival
	Value isMetricVal_Value `protobuf_oneof:"value"`
}

func (m *MetricVal) Reset()                    { *m = MetricVal{} }
func (m *MetricVal) String() string            { return proto.CompactTextString(m) }
func (*MetricVal) ProtoMessage()               {}
func (*MetricVal) Descriptor() ([]byte, []int) { return fileDescriptorMetric, []int{1} }

type isMetricVal_Value interface {
	isMetricVal_Value()
	MarshalTo([]byte) (int, error)
	Size() int
}

type MetricVal_Dval struct {
	Dval float64 `protobuf:"fixed64,2,opt,name=dval,proto3,oneof"`
}
type MetricVal_Ival struct {
	Ival uint64 `protobuf:"varint,3,opt,name=ival,proto3,oneof"`
}

func (*MetricVal_Dval) isMetricVal_Value() {}
func (*MetricVal_Ival) isMetricVal_Value() {}

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

// XXX_OneofFuncs is for the internal use of the proto package.
func (*MetricVal) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _MetricVal_OneofMarshaler, _MetricVal_OneofUnmarshaler, _MetricVal_OneofSizer, []interface{}{
		(*MetricVal_Dval)(nil),
		(*MetricVal_Ival)(nil),
	}
}

func _MetricVal_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*MetricVal)
	// value
	switch x := m.Value.(type) {
	case *MetricVal_Dval:
		_ = b.EncodeVarint(2<<3 | proto.WireFixed64)
		_ = b.EncodeFixed64(math.Float64bits(x.Dval))
	case *MetricVal_Ival:
		_ = b.EncodeVarint(3<<3 | proto.WireVarint)
		_ = b.EncodeVarint(uint64(x.Ival))
	case nil:
	default:
		return fmt.Errorf("MetricVal.Value has unexpected type %T", x)
	}
	return nil
}

func _MetricVal_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*MetricVal)
	switch tag {
	case 2: // value.dval
		if wire != proto.WireFixed64 {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeFixed64()
		m.Value = &MetricVal_Dval{math.Float64frombits(x)}
		return true, err
	case 3: // value.ival
		if wire != proto.WireVarint {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeVarint()
		m.Value = &MetricVal_Ival{x}
		return true, err
	default:
		return false, nil
	}
}

func _MetricVal_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*MetricVal)
	// value
	switch x := m.Value.(type) {
	case *MetricVal_Dval:
		n += proto.SizeVarint(2<<3 | proto.WireFixed64)
		n += 8
	case *MetricVal_Ival:
		n += proto.SizeVarint(3<<3 | proto.WireVarint)
		n += proto.SizeVarint(uint64(x.Ival))
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// Metric is an entry/point in a time series of values for Analytics/Billing.
type Metric struct {
	// Metric name
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Timestamp when the metric was captured
	Timestamp google_protobuf1.Timestamp `protobuf:"bytes,2,opt,name=timestamp" json:"timestamp"`
	// Tags associated with the metric for searching/filtering
	Tags []*MetricTag `protobuf:"bytes,3,rep,name=tags" json:"tags,omitempty"`
	// Values associated with the metric
	Vals []*MetricVal `protobuf:"bytes,4,rep,name=vals" json:"vals,omitempty"`
}

func (m *Metric) Reset()                    { *m = Metric{} }
func (m *Metric) String() string            { return proto.CompactTextString(m) }
func (*Metric) ProtoMessage()               {}
func (*Metric) Descriptor() ([]byte, []int) { return fileDescriptorMetric, []int{2} }

func init() {
	proto.RegisterType((*MetricTag)(nil), "edgeproto.MetricTag")
	proto.RegisterType((*MetricVal)(nil), "edgeproto.MetricVal")
	proto.RegisterType((*Metric)(nil), "edgeproto.Metric")
}
func (m *MetricTag) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MetricTag) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintMetric(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	if len(m.Val) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintMetric(dAtA, i, uint64(len(m.Val)))
		i += copy(dAtA[i:], m.Val)
	}
	return i, nil
}

func (m *MetricVal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MetricVal) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintMetric(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	if m.Value != nil {
		nn1, err := m.Value.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += nn1
	}
	return i, nil
}

func (m *MetricVal_Dval) MarshalTo(dAtA []byte) (int, error) {
	i := 0
	dAtA[i] = 0x11
	i++
	binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Dval))))
	i += 8
	return i, nil
}
func (m *MetricVal_Ival) MarshalTo(dAtA []byte) (int, error) {
	i := 0
	dAtA[i] = 0x18
	i++
	i = encodeVarintMetric(dAtA, i, uint64(m.Ival))
	return i, nil
}
func (m *Metric) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Metric) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintMetric(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	dAtA[i] = 0x12
	i++
	i = encodeVarintMetric(dAtA, i, uint64(m.Timestamp.Size()))
	n2, err := m.Timestamp.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n2
	if len(m.Tags) > 0 {
		for _, msg := range m.Tags {
			dAtA[i] = 0x1a
			i++
			i = encodeVarintMetric(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.Vals) > 0 {
		for _, msg := range m.Vals {
			dAtA[i] = 0x22
			i++
			i = encodeVarintMetric(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func encodeVarintMetric(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
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

// Helper method to check that enums have valid values
func (m *MetricTag) ValidateEnums() error {
	return nil
}

func (m *MetricVal) CopyInFields(src *MetricVal) int {
	changed := 0
	if m.Name != src.Name {
		m.Name = src.Name
		changed++
	}
	return changed
}

// Helper method to check that enums have valid values
func (m *MetricVal) ValidateEnums() error {
	return nil
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
		if m.Tags == nil || len(m.Tags) != len(src.Tags) {
			m.Tags = make([]*MetricTag, len(src.Tags))
		}
		for i0 := 0; i0 < len(src.Tags); i0++ {
			m.Tags[i0] = &MetricTag{}
			if m.Tags[i0].Name != src.Tags[i0].Name {
				m.Tags[i0].Name = src.Tags[i0].Name
				changed++
			}
			if m.Tags[i0].Val != src.Tags[i0].Val {
				m.Tags[i0].Val = src.Tags[i0].Val
				changed++
			}
		}
	}
	if src.Vals != nil {
		if m.Vals == nil || len(m.Vals) != len(src.Vals) {
			m.Vals = make([]*MetricVal, len(src.Vals))
		}
		for i0 := 0; i0 < len(src.Vals); i0++ {
			m.Vals[i0] = &MetricVal{}
			if m.Vals[i0].Name != src.Vals[i0].Name {
				m.Vals[i0].Name = src.Vals[i0].Name
				changed++
			}
		}
	}
	return changed
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

func (m *MetricTag) Size() (n int) {
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
	var l int
	_ = l
	n += 9
	return n
}
func (m *MetricVal_Ival) Size() (n int) {
	var l int
	_ = l
	n += 1 + sovMetric(uint64(m.Ival))
	return n
}
func (m *Metric) Size() (n int) {
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
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
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
			wire |= (uint64(b) & 0x7F) << shift
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
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMetric
			}
			postIndex := iNdEx + intStringLen
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
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMetric
			}
			postIndex := iNdEx + intStringLen
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
			wire |= (uint64(b) & 0x7F) << shift
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
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMetric
			}
			postIndex := iNdEx + intStringLen
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
			v = uint64(binary.LittleEndian.Uint64(dAtA[iNdEx:]))
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
				v |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Value = &MetricVal_Ival{v}
		default:
			iNdEx = preIndex
			skippy, err := skipMetric(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
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
			wire |= (uint64(b) & 0x7F) << shift
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
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMetric
			}
			postIndex := iNdEx + intStringLen
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
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMetric
			}
			postIndex := iNdEx + msglen
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
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMetric
			}
			postIndex := iNdEx + msglen
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
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMetric
			}
			postIndex := iNdEx + msglen
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
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
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
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthMetric
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowMetric
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
				next, err := skipMetric(dAtA[start:])
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
	ErrInvalidLengthMetric = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMetric   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("metric.proto", fileDescriptorMetric) }

var fileDescriptorMetric = []byte{
	// 331 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x91, 0xbf, 0x4e, 0xf3, 0x30,
	0x14, 0xc5, 0xe3, 0x2f, 0xf9, 0x0a, 0x71, 0x19, 0xaa, 0xa8, 0x43, 0x54, 0xa1, 0xb4, 0xea, 0x94,
	0x05, 0x5b, 0x94, 0x05, 0x31, 0x30, 0x74, 0x62, 0x61, 0x89, 0xaa, 0xee, 0x4e, 0x6b, 0x8c, 0x25,
	0x3b, 0xae, 0x1a, 0xa7, 0xe2, 0xd9, 0x98, 0x18, 0x3b, 0xf2, 0x04, 0x08, 0xfa, 0x04, 0x0c, 0x7d,
	0x00, 0xe4, 0x9b, 0x3f, 0x2c, 0x65, 0x89, 0xce, 0x3d, 0xfa, 0x9d, 0x9b, 0x7b, 0x64, 0x7c, 0xa1,
	0xb9, 0xdd, 0xca, 0x15, 0xd9, 0x6c, 0x8d, 0x35, 0x51, 0xc8, 0xd7, 0x82, 0x83, 0x1c, 0x5d, 0x0a,
	0x63, 0x84, 0xe2, 0x94, 0x6d, 0x24, 0x65, 0x45, 0x61, 0x2c, 0xb3, 0xd2, 0x14, 0x65, 0x0d, 0x8e,
	0x6e, 0x85, 0xb4, 0xcf, 0x55, 0x4e, 0x56, 0x46, 0x53, 0x6d, 0x72, 0xa9, 0x5c, 0xf0, 0x85, 0xba,
	0xef, 0xd5, 0x4a, 0x99, 0x6a, 0x4d, 0x81, 0x13, 0xbc, 0xe8, 0x44, 0x93, 0x1c, 0x0a, 0x23, 0x0c,
	0x48, 0xea, 0x54, 0xe3, 0x8e, 0x9b, 0xbf, 0xc1, 0x94, 0x57, 0x4f, 0xd4, 0x4a, 0xcd, 0x4b, 0xcb,
	0xf4, 0xa6, 0x06, 0xa6, 0xd7, 0x38, 0x7c, 0x84, 0x4b, 0x17, 0x4c, 0x44, 0x11, 0x0e, 0x0a, 0xa6,
	0x79, 0x8c, 0x26, 0x28, 0x0d, 0x33, 0xd0, 0xd1, 0x00, 0xfb, 0x3b, 0xa6, 0xe2, 0x7f, 0x60, 0x39,
	0x39, 0x5d, 0xb4, 0x91, 0x25, 0x53, 0x27, 0x23, 0x43, 0x1c, 0xac, 0xdb, 0x0c, 0x7a, 0xf0, 0x32,
	0x98, 0x9c, 0x2b, 0x9d, 0xeb, 0x4f, 0x50, 0x1a, 0x38, 0xd7, 0x4d, 0xf3, 0x33, 0xfc, 0x7f, 0xc7,
	0x54, 0xc5, 0xa7, 0xaf, 0x08, 0xf7, 0xea, 0xb5, 0x27, 0x77, 0xde, 0xe3, 0xb0, 0x3b, 0x1d, 0x16,
	0xf7, 0x67, 0x23, 0x52, 0x97, 0x23, 0x6d, 0x39, 0xb2, 0x68, 0x89, 0x79, 0xb0, 0xff, 0x18, 0x7b,
	0xd9, 0x6f, 0x24, 0x4a, 0x71, 0x60, 0x99, 0x28, 0x63, 0x7f, 0xe2, 0xa7, 0xfd, 0xd9, 0x90, 0x74,
	0x0f, 0x42, 0xba, 0xfa, 0x19, 0x10, 0x8e, 0xdc, 0x31, 0x55, 0xc6, 0xc1, 0x1f, 0xe4, 0x92, 0xa9,
	0x0c, 0x88, 0xbb, 0xf3, 0xb7, 0x63, 0x8c, 0xbe, 0x8f, 0xb1, 0x37, 0x1f, 0xec, 0xbf, 0x12, 0x6f,
	0x7f, 0x48, 0xd0, 0xfb, 0x21, 0x41, 0x9f, 0x87, 0x04, 0xe5, 0x3d, 0x88, 0xdc, 0xfc, 0x04, 0x00,
	0x00, 0xff, 0xff, 0xa0, 0xbc, 0x30, 0x57, 0x08, 0x02, 0x00, 0x00,
}
