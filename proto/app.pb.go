// Code generated by protoc-gen-go. DO NOT EDIT.
// source: app.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	app.proto
	app_inst.proto
	cloudlet.proto
	debug.proto
	developer.proto
	loc.proto
	matcher.proto
	operator.proto
	result.proto
	update.proto

It has these top-level messages:
	AppKey
	App
	AppInst
	Cloudlet
	DebugLevel
	DeveloperKey
	Developer
	Loc
	ServiceRequest
	ServiceReply
	Operator
	Result
	Update
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "google.golang.org/genproto/googleapis/api/annotations"
import _ "github.com/mobiledgex/edge-cloud/protogen"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

import "encoding/json"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

// key that uniquely identifies an application
type AppKey struct {
	// developer name
	DevName string `protobuf:"bytes,1,opt,name=dev_name,json=devName" json:"dev_name,omitempty"`
	// application name
	AppName string `protobuf:"bytes,2,opt,name=app_name,json=appName" json:"app_name,omitempty"`
	// version of the app
	Version string `protobuf:"bytes,3,opt,name=version" json:"version,omitempty"`
}

func (m *AppKey) Reset()                    { *m = AppKey{} }
func (m *AppKey) String() string            { return proto1.CompactTextString(m) }
func (*AppKey) ProtoMessage()               {}
func (*AppKey) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *AppKey) GetDevName() string {
	if m != nil {
		return m.DevName
	}
	return ""
}

func (m *AppKey) GetAppName() string {
	if m != nil {
		return m.AppName
	}
	return ""
}

func (m *AppKey) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

// Applications are created and uploaded by developers
// Only registered applications can access location and cloudlet services
type App struct {
	Key *AppKey `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	// Path to the application binary on shared storage
	AppPath string `protobuf:"bytes,4,opt,name=app_path,json=appPath" json:"app_path,omitempty"`
}

func (m *App) Reset()                    { *m = App{} }
func (m *App) String() string            { return proto1.CompactTextString(m) }
func (*App) ProtoMessage()               {}
func (*App) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *App) GetKey() *AppKey {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *App) GetAppPath() string {
	if m != nil {
		return m.AppPath
	}
	return ""
}

func init() {
	proto1.RegisterType((*AppKey)(nil), "proto.AppKey")
	proto1.RegisterType((*App)(nil), "proto.App")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for AppApi service

type AppApiClient interface {
	CreateApp(ctx context.Context, in *App, opts ...grpc.CallOption) (*Result, error)
	DeleteApp(ctx context.Context, in *App, opts ...grpc.CallOption) (*Result, error)
	UpdateApp(ctx context.Context, in *App, opts ...grpc.CallOption) (*Result, error)
	ShowApp(ctx context.Context, in *App, opts ...grpc.CallOption) (AppApi_ShowAppClient, error)
}

type appApiClient struct {
	cc *grpc.ClientConn
}

func NewAppApiClient(cc *grpc.ClientConn) AppApiClient {
	return &appApiClient{cc}
}

func (c *appApiClient) CreateApp(ctx context.Context, in *App, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/proto.AppApi/CreateApp", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appApiClient) DeleteApp(ctx context.Context, in *App, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/proto.AppApi/DeleteApp", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appApiClient) UpdateApp(ctx context.Context, in *App, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/proto.AppApi/UpdateApp", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appApiClient) ShowApp(ctx context.Context, in *App, opts ...grpc.CallOption) (AppApi_ShowAppClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_AppApi_serviceDesc.Streams[0], c.cc, "/proto.AppApi/ShowApp", opts...)
	if err != nil {
		return nil, err
	}
	x := &appApiShowAppClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type AppApi_ShowAppClient interface {
	Recv() (*App, error)
	grpc.ClientStream
}

type appApiShowAppClient struct {
	grpc.ClientStream
}

func (x *appApiShowAppClient) Recv() (*App, error) {
	m := new(App)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for AppApi service

type AppApiServer interface {
	CreateApp(context.Context, *App) (*Result, error)
	DeleteApp(context.Context, *App) (*Result, error)
	UpdateApp(context.Context, *App) (*Result, error)
	ShowApp(*App, AppApi_ShowAppServer) error
}

func RegisterAppApiServer(s *grpc.Server, srv AppApiServer) {
	s.RegisterService(&_AppApi_serviceDesc, srv)
}

func _AppApi_CreateApp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(App)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppApiServer).CreateApp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.AppApi/CreateApp",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppApiServer).CreateApp(ctx, req.(*App))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppApi_DeleteApp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(App)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppApiServer).DeleteApp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.AppApi/DeleteApp",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppApiServer).DeleteApp(ctx, req.(*App))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppApi_UpdateApp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(App)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppApiServer).UpdateApp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.AppApi/UpdateApp",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppApiServer).UpdateApp(ctx, req.(*App))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppApi_ShowApp_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(App)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(AppApiServer).ShowApp(m, &appApiShowAppServer{stream})
}

type AppApi_ShowAppServer interface {
	Send(*App) error
	grpc.ServerStream
}

type appApiShowAppServer struct {
	grpc.ServerStream
}

func (x *appApiShowAppServer) Send(m *App) error {
	return x.ServerStream.SendMsg(m)
}

var _AppApi_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.AppApi",
	HandlerType: (*AppApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateApp",
			Handler:    _AppApi_CreateApp_Handler,
		},
		{
			MethodName: "DeleteApp",
			Handler:    _AppApi_DeleteApp_Handler,
		},
		{
			MethodName: "UpdateApp",
			Handler:    _AppApi_UpdateApp_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ShowApp",
			Handler:       _AppApi_ShowApp_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "app.proto",
}

func (m *AppKey) Matches(filter *AppKey) bool {
	if filter == nil {
		return true
	}
	if filter.DevName != "" && filter.DevName != m.DevName {
		return false
	}
	if filter.AppName != "" && filter.AppName != m.AppName {
		return false
	}
	if filter.Version != "" && filter.Version != m.Version {
		return false
	}
	return true
}

func (m *AppKey) CopyInFields(src *AppKey) {
	m.DevName = src.DevName
	m.AppName = src.AppName
	m.Version = src.Version
}
func (m *App) Matches(filter *App) bool {
	if filter == nil {
		return true
	}
	if filter.Key != nil && !filter.Key.Matches(m.Key) {
		return false
	}
	if filter.AppPath != "" && filter.AppPath != m.AppPath {
		return false
	}
	return true
}

func (m *App) CopyInFields(src *App) {
	if m.Key != nil && src.Key != nil {
		*m.Key = *src.Key
	}
	m.AppPath = src.AppPath
}

type AppCud interface {
	// Validate all fields for create/update
	Validate(in *App) error
	// Validate only Key fields for delete
	ValidateKey(in *App) error
	// Get key string for etcd access
	GetKeyString(in *App) string
	//Etcd IO interface
	EtcdIO
	// Refresh is called after create/update/delete to update in-memory cache
	Refresh(in *App, key string) error
}

func (m *App) Create(cud AppCud) (*Result, error) {
	err := cud.Validate(m)
	if err != nil {
		return nil, err
	}
	key := cud.GetKeyString(m)
	val, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	err = cud.Create(key, string(val))
	if err != nil {
		return nil, err
	}
	err = cud.Refresh(m, key)
	return &Result{}, err
}

func (m *App) Update(cud AppCud) (*Result, error) {
	err := cud.Validate(m)
	if err != nil {
		return nil, err
	}
	key := cud.GetKeyString(m)
	var vers int64 = 0
	val, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	err = cud.Update(key, string(val), vers)
	if err != nil {
		return nil, err
	}
	err = cud.Refresh(m, key)
	return &Result{}, err
}

func (m *App) Delete(cud AppCud) (*Result, error) {
	err := cud.ValidateKey(m)
	if err != nil {
		return nil, err
	}
	key := cud.GetKeyString(m)
	err = cud.Delete(key)
	if err != nil {
		return nil, err
	}
	err = cud.Refresh(m, key)
	return &Result{}, err
}

func init() { proto1.RegisterFile("app.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 294 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x90, 0x4f, 0x4b, 0xc3, 0x30,
	0x18, 0xc6, 0x17, 0x37, 0xf7, 0x27, 0xea, 0xa5, 0xa7, 0x58, 0x04, 0xa5, 0x20, 0xec, 0x62, 0x2b,
	0xf3, 0x22, 0xbb, 0x15, 0xbd, 0x0d, 0x44, 0x2a, 0x5e, 0x95, 0x74, 0x7d, 0x69, 0x83, 0x6d, 0xf2,
	0xd2, 0xa6, 0xd5, 0x7d, 0x28, 0xbf, 0x97, 0x47, 0x61, 0x5f, 0x40, 0x92, 0x6c, 0x53, 0x3c, 0xed,
	0xd2, 0xbe, 0x2f, 0xbf, 0x27, 0xbf, 0x27, 0x84, 0x4e, 0x38, 0x62, 0x88, 0xb5, 0xd2, 0xca, 0x3b,
	0xb4, 0x3f, 0xff, 0x2c, 0x57, 0x2a, 0x2f, 0x21, 0xe2, 0x28, 0x22, 0x2e, 0xa5, 0xd2, 0x5c, 0x0b,
	0x25, 0x1b, 0x17, 0xf2, 0x8f, 0x6b, 0x68, 0xda, 0x52, 0x6f, 0xb6, 0xdb, 0x5c, 0xe8, 0xa2, 0x4d,
	0xc3, 0xa5, 0xaa, 0xa2, 0x4a, 0xa5, 0xa2, 0x84, 0x2c, 0x87, 0x8f, 0xc8, 0x7c, 0xaf, 0x96, 0xa5,
	0x6a, 0xb3, 0xc8, 0xe6, 0x72, 0x90, 0xbb, 0xc1, 0x9d, 0x0c, 0x5e, 0xe8, 0x30, 0x46, 0x5c, 0xc0,
	0xca, 0x3b, 0xa5, 0xe3, 0x0c, 0xba, 0x57, 0xc9, 0x2b, 0x60, 0xe4, 0x82, 0x4c, 0x27, 0xc9, 0x28,
	0x83, 0xee, 0x81, 0x57, 0x60, 0x10, 0x47, 0x74, 0xe8, 0xc0, 0x21, 0x8e, 0x68, 0x11, 0xa3, 0xa3,
	0x0e, 0xea, 0x46, 0x28, 0xc9, 0xfa, 0x8e, 0x6c, 0xd6, 0xf9, 0xe0, 0x6b, 0xcd, 0x48, 0xb0, 0xa0,
	0xfd, 0x18, 0xd1, 0x3b, 0xa7, 0xfd, 0x37, 0x58, 0x59, 0xef, 0xd1, 0xec, 0xc4, 0x75, 0x87, 0xae,
	0x38, 0x31, 0x64, 0x5b, 0x81, 0x5c, 0x17, 0x6c, 0xb0, 0xab, 0x78, 0xe4, 0xba, 0x98, 0x8f, 0x8d,
	0xe8, 0x7b, 0xcd, 0xc8, 0xec, 0x93, 0xd8, 0xdb, 0xc6, 0x28, 0xbc, 0x29, 0x9d, 0xdc, 0xd5, 0xc0,
	0x35, 0x18, 0x3b, 0xfd, 0x15, 0xfa, 0x5b, 0x79, 0x62, 0xdf, 0x27, 0xe8, 0x99, 0xe4, 0x3d, 0x94,
	0xb0, 0x5f, 0xf2, 0x19, 0xb3, 0x7d, 0x9c, 0x97, 0x74, 0xf4, 0x54, 0xa8, 0xf7, 0xff, 0xb9, 0x3f,
	0x73, 0xd0, 0xbb, 0x26, 0xe9, 0xd0, 0xae, 0x37, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x8d, 0x7a,
	0x65, 0x5c, 0xdd, 0x01, 0x00, 0x00,
}
