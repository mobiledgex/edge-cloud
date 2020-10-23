// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: infraresources.proto

package edgeproto

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
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

// ContainerInfo
//
// ContainerInfo is infomation about containers running on a VM,
type ContainerInfo struct {
	// Name of the container
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Type can be docker or kubernetes
	Type string `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	// Runtime status of the container
	Status string `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	// IP within the CNI and is applicable to kubernetes only
	Clusterip string `protobuf:"bytes,4,opt,name=clusterip,proto3" json:"clusterip,omitempty"`
	// Restart count, applicable to kubernetes only
	Restarts int64 `protobuf:"varint,5,opt,name=restarts,proto3" json:"restarts,omitempty"`
}

func (m *ContainerInfo) Reset()         { *m = ContainerInfo{} }
func (m *ContainerInfo) String() string { return proto.CompactTextString(m) }
func (*ContainerInfo) ProtoMessage()    {}
func (*ContainerInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_1d4658e0b2956cb2, []int{0}
}
func (m *ContainerInfo) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ContainerInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ContainerInfo.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ContainerInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ContainerInfo.Merge(m, src)
}
func (m *ContainerInfo) XXX_Size() int {
	return m.Size()
}
func (m *ContainerInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_ContainerInfo.DiscardUnknown(m)
}

var xxx_messageInfo_ContainerInfo proto.InternalMessageInfo

// IpAddr is an address for a VM which may have an external and
// internal component.  Internal and external is with respect to the VM
// and are are often the same unless a natted or floating IP is used.  If
// internalIP is not reported it is the same as the ExternalIP.
type IpAddr struct {
	ExternalIp string `protobuf:"bytes,1,opt,name=externalIp,proto3" json:"externalIp,omitempty"`
	InternalIp string `protobuf:"bytes,2,opt,name=internalIp,proto3" json:"internalIp,omitempty"`
}

func (m *IpAddr) Reset()         { *m = IpAddr{} }
func (m *IpAddr) String() string { return proto.CompactTextString(m) }
func (*IpAddr) ProtoMessage()    {}
func (*IpAddr) Descriptor() ([]byte, []int) {
	return fileDescriptor_1d4658e0b2956cb2, []int{1}
}
func (m *IpAddr) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *IpAddr) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_IpAddr.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *IpAddr) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IpAddr.Merge(m, src)
}
func (m *IpAddr) XXX_Size() int {
	return m.Size()
}
func (m *IpAddr) XXX_DiscardUnknown() {
	xxx_messageInfo_IpAddr.DiscardUnknown(m)
}

var xxx_messageInfo_IpAddr proto.InternalMessageInfo

// VmInfo
//
// VmInfo is information about Virtual Machine resources.
type VmInfo struct {
	// Virtual machine name
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Type can be platform, rootlb, cluster-master, cluster-node, vmapp
	Type string `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	// Runtime status of the VM
	Status string `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	// Flavor allocated within the cloudlet infrastructure, distinct from the control plane flavor
	InfraFlavor string `protobuf:"bytes,4,opt,name=infraFlavor,proto3" json:"infraFlavor,omitempty"`
	// IP addresses allocated to the VM
	Ipaddresses []IpAddr `protobuf:"bytes,5,rep,name=ipaddresses,proto3" json:"ipaddresses"`
	// Information about containers running in the VM
	Containers []*ContainerInfo `protobuf:"bytes,6,rep,name=containers,proto3" json:"containers,omitempty"`
}

func (m *VmInfo) Reset()         { *m = VmInfo{} }
func (m *VmInfo) String() string { return proto.CompactTextString(m) }
func (*VmInfo) ProtoMessage()    {}
func (*VmInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_1d4658e0b2956cb2, []int{2}
}
func (m *VmInfo) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *VmInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_VmInfo.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *VmInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VmInfo.Merge(m, src)
}
func (m *VmInfo) XXX_Size() int {
	return m.Size()
}
func (m *VmInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_VmInfo.DiscardUnknown(m)
}

var xxx_messageInfo_VmInfo proto.InternalMessageInfo

// InfraResources
//
// InfraResources is infomation about infrastructure resources.
type InfraResources struct {
	Vms []VmInfo `protobuf:"bytes,3,rep,name=vms,proto3" json:"vms"`
}

func (m *InfraResources) Reset()         { *m = InfraResources{} }
func (m *InfraResources) String() string { return proto.CompactTextString(m) }
func (*InfraResources) ProtoMessage()    {}
func (*InfraResources) Descriptor() ([]byte, []int) {
	return fileDescriptor_1d4658e0b2956cb2, []int{3}
}
func (m *InfraResources) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *InfraResources) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_InfraResources.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *InfraResources) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InfraResources.Merge(m, src)
}
func (m *InfraResources) XXX_Size() int {
	return m.Size()
}
func (m *InfraResources) XXX_DiscardUnknown() {
	xxx_messageInfo_InfraResources.DiscardUnknown(m)
}

var xxx_messageInfo_InfraResources proto.InternalMessageInfo

func init() {
	proto.RegisterType((*ContainerInfo)(nil), "edgeproto.ContainerInfo")
	proto.RegisterType((*IpAddr)(nil), "edgeproto.IpAddr")
	proto.RegisterType((*VmInfo)(nil), "edgeproto.VmInfo")
	proto.RegisterType((*InfraResources)(nil), "edgeproto.InfraResources")
}

func init() { proto.RegisterFile("infraresources.proto", fileDescriptor_1d4658e0b2956cb2) }

var fileDescriptor_1d4658e0b2956cb2 = []byte{
	// 398 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x52, 0x3d, 0x8f, 0xd4, 0x30,
	0x10, 0x8d, 0xc9, 0x12, 0xb1, 0xb3, 0x80, 0x84, 0x75, 0x42, 0xd1, 0x6a, 0x65, 0x56, 0xa9, 0x96,
	0x82, 0x44, 0x82, 0xe6, 0x10, 0x15, 0x87, 0x84, 0x48, 0x9b, 0x82, 0xde, 0x9b, 0xf8, 0x82, 0xa5,
	0xc4, 0x13, 0xd9, 0xce, 0xe9, 0xf8, 0x09, 0x74, 0xfc, 0xac, 0x2d, 0xaf, 0xa4, 0xe2, 0x63, 0xf7,
	0x8f, 0x20, 0x3b, 0xd9, 0x10, 0xa8, 0x69, 0xa2, 0x37, 0x6f, 0xde, 0x93, 0xe7, 0xcd, 0x04, 0x2e,
	0xa4, 0xba, 0xd6, 0x5c, 0x0b, 0x83, 0xbd, 0x2e, 0x85, 0x49, 0x3b, 0x8d, 0x16, 0xe9, 0x52, 0x54,
	0xb5, 0xf0, 0x70, 0xbd, 0xa9, 0x11, 0xeb, 0x46, 0x64, 0xbc, 0x93, 0x19, 0x57, 0x0a, 0x2d, 0xb7,
	0x12, 0xd5, 0x28, 0x5c, 0x5f, 0xd6, 0xd2, 0x7e, 0xea, 0xf7, 0x69, 0x89, 0x6d, 0xd6, 0xe2, 0x5e,
	0x36, 0xce, 0x78, 0x9b, 0xb9, 0xef, 0x8b, 0xb2, 0xc1, 0xbe, 0xca, 0xbc, 0xae, 0x16, 0x6a, 0x02,
	0xa3, 0xf3, 0x61, 0x89, 0x6d, 0x8b, 0xe7, 0xea, 0xa2, 0xc6, 0x1a, 0x3d, 0xcc, 0x1c, 0x1a, 0xd8,
	0xe4, 0x0b, 0x81, 0x47, 0xef, 0x50, 0x59, 0x2e, 0x95, 0xd0, 0xb9, 0xba, 0x46, 0x4a, 0x61, 0xa1,
	0x78, 0x2b, 0x62, 0xb2, 0x25, 0xbb, 0x65, 0xe1, 0xb1, 0xe3, 0xec, 0xe7, 0x4e, 0xc4, 0xf7, 0x06,
	0xce, 0x61, 0xfa, 0x14, 0x22, 0x63, 0xb9, 0xed, 0x4d, 0x1c, 0x7a, 0x76, 0xac, 0xe8, 0x06, 0x96,
	0x65, 0xd3, 0x1b, 0x2b, 0xb4, 0xec, 0xe2, 0x85, 0x6f, 0xfd, 0x21, 0xe8, 0x1a, 0x1e, 0x68, 0x61,
	0x2c, 0xd7, 0xd6, 0xc4, 0xf7, 0xb7, 0x64, 0x17, 0x16, 0x53, 0x9d, 0x7c, 0x80, 0x28, 0xef, 0xde,
	0x56, 0x95, 0xa6, 0x0c, 0x40, 0xdc, 0x5a, 0xa1, 0x15, 0x6f, 0xf2, 0x6e, 0x9c, 0x64, 0xc6, 0xb8,
	0xbe, 0x54, 0x53, 0x7f, 0x98, 0x6a, 0xc6, 0x24, 0x3f, 0x08, 0x44, 0x1f, 0xdb, 0xff, 0x12, 0x67,
	0x0b, 0x2b, 0x7f, 0xbf, 0xf7, 0x0d, 0xbf, 0x41, 0x3d, 0x06, 0x9a, 0x53, 0xf4, 0x35, 0xac, 0x64,
	0xc7, 0xab, 0x4a, 0x0b, 0x63, 0x84, 0x4b, 0x15, 0xee, 0x56, 0x2f, 0x9f, 0xa4, 0xd3, 0x7d, 0xd3,
	0x21, 0xd4, 0xd5, 0xe2, 0xf0, 0xfd, 0x59, 0x50, 0xcc, 0xb5, 0xf4, 0x12, 0xa0, 0x3c, 0x2f, 0xdf,
	0xc4, 0x91, 0x77, 0xc6, 0x33, 0xe7, 0x5f, 0x97, 0x29, 0x66, 0xda, 0xe4, 0x0d, 0x3c, 0xce, 0xdd,
	0x0c, 0xc5, 0xf9, 0xb7, 0xa2, 0xcf, 0x21, 0xbc, 0x69, 0xdd, 0xf4, 0xff, 0x3e, 0x3f, 0x2c, 0x62,
	0x7c, 0xde, 0x69, 0xae, 0x36, 0x87, 0x5f, 0x2c, 0x38, 0x1c, 0x19, 0xb9, 0x3b, 0x32, 0xf2, 0xf3,
	0xc8, 0xc8, 0xd7, 0x13, 0x0b, 0xee, 0x4e, 0x2c, 0xf8, 0x76, 0x62, 0xc1, 0x3e, 0xf2, 0xb6, 0x57,
	0xbf, 0x03, 0x00, 0x00, 0xff, 0xff, 0x91, 0xe2, 0x46, 0x94, 0xb8, 0x02, 0x00, 0x00,
}

func (m *ContainerInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ContainerInfo) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ContainerInfo) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Restarts != 0 {
		i = encodeVarintInfraresources(dAtA, i, uint64(m.Restarts))
		i--
		dAtA[i] = 0x28
	}
	if len(m.Clusterip) > 0 {
		i -= len(m.Clusterip)
		copy(dAtA[i:], m.Clusterip)
		i = encodeVarintInfraresources(dAtA, i, uint64(len(m.Clusterip)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Status) > 0 {
		i -= len(m.Status)
		copy(dAtA[i:], m.Status)
		i = encodeVarintInfraresources(dAtA, i, uint64(len(m.Status)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Type) > 0 {
		i -= len(m.Type)
		copy(dAtA[i:], m.Type)
		i = encodeVarintInfraresources(dAtA, i, uint64(len(m.Type)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintInfraresources(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *IpAddr) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *IpAddr) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *IpAddr) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.InternalIp) > 0 {
		i -= len(m.InternalIp)
		copy(dAtA[i:], m.InternalIp)
		i = encodeVarintInfraresources(dAtA, i, uint64(len(m.InternalIp)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ExternalIp) > 0 {
		i -= len(m.ExternalIp)
		copy(dAtA[i:], m.ExternalIp)
		i = encodeVarintInfraresources(dAtA, i, uint64(len(m.ExternalIp)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *VmInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *VmInfo) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *VmInfo) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Containers) > 0 {
		for iNdEx := len(m.Containers) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Containers[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintInfraresources(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x32
		}
	}
	if len(m.Ipaddresses) > 0 {
		for iNdEx := len(m.Ipaddresses) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Ipaddresses[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintInfraresources(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x2a
		}
	}
	if len(m.InfraFlavor) > 0 {
		i -= len(m.InfraFlavor)
		copy(dAtA[i:], m.InfraFlavor)
		i = encodeVarintInfraresources(dAtA, i, uint64(len(m.InfraFlavor)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Status) > 0 {
		i -= len(m.Status)
		copy(dAtA[i:], m.Status)
		i = encodeVarintInfraresources(dAtA, i, uint64(len(m.Status)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Type) > 0 {
		i -= len(m.Type)
		copy(dAtA[i:], m.Type)
		i = encodeVarintInfraresources(dAtA, i, uint64(len(m.Type)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintInfraresources(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *InfraResources) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *InfraResources) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *InfraResources) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Vms) > 0 {
		for iNdEx := len(m.Vms) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Vms[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintInfraresources(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintInfraresources(dAtA []byte, offset int, v uint64) int {
	offset -= sovInfraresources(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ContainerInfo) CopyInFields(src *ContainerInfo) int {
	changed := 0
	if m.Name != src.Name {
		m.Name = src.Name
		changed++
	}
	if m.Type != src.Type {
		m.Type = src.Type
		changed++
	}
	if m.Status != src.Status {
		m.Status = src.Status
		changed++
	}
	if m.Clusterip != src.Clusterip {
		m.Clusterip = src.Clusterip
		changed++
	}
	if m.Restarts != src.Restarts {
		m.Restarts = src.Restarts
		changed++
	}
	return changed
}

func (m *ContainerInfo) DeepCopyIn(src *ContainerInfo) {
	m.Name = src.Name
	m.Type = src.Type
	m.Status = src.Status
	m.Clusterip = src.Clusterip
	m.Restarts = src.Restarts
}

// Helper method to check that enums have valid values
func (m *ContainerInfo) ValidateEnums() error {
	return nil
}

func (m *IpAddr) CopyInFields(src *IpAddr) int {
	changed := 0
	if m.ExternalIp != src.ExternalIp {
		m.ExternalIp = src.ExternalIp
		changed++
	}
	if m.InternalIp != src.InternalIp {
		m.InternalIp = src.InternalIp
		changed++
	}
	return changed
}

func (m *IpAddr) DeepCopyIn(src *IpAddr) {
	m.ExternalIp = src.ExternalIp
	m.InternalIp = src.InternalIp
}

// Helper method to check that enums have valid values
func (m *IpAddr) ValidateEnums() error {
	return nil
}

func (m *VmInfo) CopyInFields(src *VmInfo) int {
	changed := 0
	if m.Name != src.Name {
		m.Name = src.Name
		changed++
	}
	if m.Type != src.Type {
		m.Type = src.Type
		changed++
	}
	if m.Status != src.Status {
		m.Status = src.Status
		changed++
	}
	if m.InfraFlavor != src.InfraFlavor {
		m.InfraFlavor = src.InfraFlavor
		changed++
	}
	if src.Ipaddresses != nil {
		m.Ipaddresses = src.Ipaddresses
		changed++
	} else if m.Ipaddresses != nil {
		m.Ipaddresses = nil
		changed++
	}
	if src.Containers != nil {
		m.Containers = src.Containers
		changed++
	} else if m.Containers != nil {
		m.Containers = nil
		changed++
	}
	return changed
}

func (m *VmInfo) DeepCopyIn(src *VmInfo) {
	m.Name = src.Name
	m.Type = src.Type
	m.Status = src.Status
	m.InfraFlavor = src.InfraFlavor
	if src.Ipaddresses != nil {
		m.Ipaddresses = make([]IpAddr, len(src.Ipaddresses), len(src.Ipaddresses))
		for ii, s := range src.Ipaddresses {
			m.Ipaddresses[ii].DeepCopyIn(&s)
		}
	} else {
		m.Ipaddresses = nil
	}
	if src.Containers != nil {
		m.Containers = make([]*ContainerInfo, len(src.Containers), len(src.Containers))
		for ii, s := range src.Containers {
			var tmp_s ContainerInfo
			tmp_s.DeepCopyIn(s)
			m.Containers[ii] = &tmp_s
		}
	} else {
		m.Containers = nil
	}
}

// Helper method to check that enums have valid values
func (m *VmInfo) ValidateEnums() error {
	for _, e := range m.Ipaddresses {
		if err := e.ValidateEnums(); err != nil {
			return err
		}
	}
	for _, e := range m.Containers {
		if err := e.ValidateEnums(); err != nil {
			return err
		}
	}
	return nil
}

func (m *InfraResources) CopyInFields(src *InfraResources) int {
	changed := 0
	if src.Vms != nil {
		m.Vms = src.Vms
		changed++
	} else if m.Vms != nil {
		m.Vms = nil
		changed++
	}
	return changed
}

func (m *InfraResources) DeepCopyIn(src *InfraResources) {
	if src.Vms != nil {
		m.Vms = make([]VmInfo, len(src.Vms), len(src.Vms))
		for ii, s := range src.Vms {
			m.Vms[ii].DeepCopyIn(&s)
		}
	} else {
		m.Vms = nil
	}
}

// Helper method to check that enums have valid values
func (m *InfraResources) ValidateEnums() error {
	for _, e := range m.Vms {
		if err := e.ValidateEnums(); err != nil {
			return err
		}
	}
	return nil
}

func (m *ContainerInfo) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovInfraresources(uint64(l))
	}
	l = len(m.Type)
	if l > 0 {
		n += 1 + l + sovInfraresources(uint64(l))
	}
	l = len(m.Status)
	if l > 0 {
		n += 1 + l + sovInfraresources(uint64(l))
	}
	l = len(m.Clusterip)
	if l > 0 {
		n += 1 + l + sovInfraresources(uint64(l))
	}
	if m.Restarts != 0 {
		n += 1 + sovInfraresources(uint64(m.Restarts))
	}
	return n
}

func (m *IpAddr) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ExternalIp)
	if l > 0 {
		n += 1 + l + sovInfraresources(uint64(l))
	}
	l = len(m.InternalIp)
	if l > 0 {
		n += 1 + l + sovInfraresources(uint64(l))
	}
	return n
}

func (m *VmInfo) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovInfraresources(uint64(l))
	}
	l = len(m.Type)
	if l > 0 {
		n += 1 + l + sovInfraresources(uint64(l))
	}
	l = len(m.Status)
	if l > 0 {
		n += 1 + l + sovInfraresources(uint64(l))
	}
	l = len(m.InfraFlavor)
	if l > 0 {
		n += 1 + l + sovInfraresources(uint64(l))
	}
	if len(m.Ipaddresses) > 0 {
		for _, e := range m.Ipaddresses {
			l = e.Size()
			n += 1 + l + sovInfraresources(uint64(l))
		}
	}
	if len(m.Containers) > 0 {
		for _, e := range m.Containers {
			l = e.Size()
			n += 1 + l + sovInfraresources(uint64(l))
		}
	}
	return n
}

func (m *InfraResources) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Vms) > 0 {
		for _, e := range m.Vms {
			l = e.Size()
			n += 1 + l + sovInfraresources(uint64(l))
		}
	}
	return n
}

func sovInfraresources(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozInfraresources(x uint64) (n int) {
	return sovInfraresources(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ContainerInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowInfraresources
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
			return fmt.Errorf("proto: ContainerInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ContainerInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfraresources
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
				return ErrInvalidLengthInfraresources
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfraresources
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfraresources
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
				return ErrInvalidLengthInfraresources
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfraresources
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Type = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfraresources
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
				return ErrInvalidLengthInfraresources
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfraresources
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Status = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Clusterip", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfraresources
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
				return ErrInvalidLengthInfraresources
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfraresources
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Clusterip = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Restarts", wireType)
			}
			m.Restarts = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfraresources
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Restarts |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipInfraresources(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthInfraresources
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthInfraresources
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
func (m *IpAddr) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowInfraresources
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
			return fmt.Errorf("proto: IpAddr: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: IpAddr: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExternalIp", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfraresources
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
				return ErrInvalidLengthInfraresources
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfraresources
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ExternalIp = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InternalIp", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfraresources
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
				return ErrInvalidLengthInfraresources
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfraresources
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.InternalIp = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipInfraresources(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthInfraresources
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthInfraresources
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
func (m *VmInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowInfraresources
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
			return fmt.Errorf("proto: VmInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: VmInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfraresources
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
				return ErrInvalidLengthInfraresources
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfraresources
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfraresources
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
				return ErrInvalidLengthInfraresources
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfraresources
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Type = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfraresources
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
				return ErrInvalidLengthInfraresources
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfraresources
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Status = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InfraFlavor", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfraresources
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
				return ErrInvalidLengthInfraresources
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfraresources
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.InfraFlavor = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ipaddresses", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfraresources
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
				return ErrInvalidLengthInfraresources
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthInfraresources
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Ipaddresses = append(m.Ipaddresses, IpAddr{})
			if err := m.Ipaddresses[len(m.Ipaddresses)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Containers", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfraresources
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
				return ErrInvalidLengthInfraresources
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthInfraresources
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Containers = append(m.Containers, &ContainerInfo{})
			if err := m.Containers[len(m.Containers)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipInfraresources(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthInfraresources
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthInfraresources
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
func (m *InfraResources) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowInfraresources
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
			return fmt.Errorf("proto: InfraResources: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InfraResources: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Vms", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfraresources
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
				return ErrInvalidLengthInfraresources
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthInfraresources
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Vms = append(m.Vms, VmInfo{})
			if err := m.Vms[len(m.Vms)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipInfraresources(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthInfraresources
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthInfraresources
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
func skipInfraresources(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowInfraresources
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
					return 0, ErrIntOverflowInfraresources
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
					return 0, ErrIntOverflowInfraresources
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
				return 0, ErrInvalidLengthInfraresources
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupInfraresources
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthInfraresources
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthInfraresources        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowInfraresources          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupInfraresources = fmt.Errorf("proto: unexpected end of group")
)
