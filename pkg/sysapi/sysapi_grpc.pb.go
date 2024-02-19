// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.3
// source: pkg/sysapi/sysapi.proto

package sysapi

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	SysAPI_AuthSubjects_FullMethodName             = "/goalert.v1.SysAPI/AuthSubjects"
	SysAPI_DeleteUser_FullMethodName               = "/goalert.v1.SysAPI/DeleteUser"
	SysAPI_UsersWithoutAuthProvider_FullMethodName = "/goalert.v1.SysAPI/UsersWithoutAuthProvider"
	SysAPI_SetAuthSubject_FullMethodName           = "/goalert.v1.SysAPI/SetAuthSubject"
)

// SysAPIClient is the client API for SysAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SysAPIClient interface {
	AuthSubjects(ctx context.Context, in *AuthSubjectsRequest, opts ...grpc.CallOption) (SysAPI_AuthSubjectsClient, error)
	DeleteUser(ctx context.Context, in *DeleteUserRequest, opts ...grpc.CallOption) (*DeleteUserResponse, error)
	UsersWithoutAuthProvider(ctx context.Context, in *UsersWithoutAuthProviderRequest, opts ...grpc.CallOption) (SysAPI_UsersWithoutAuthProviderClient, error)
	SetAuthSubject(ctx context.Context, in *SetAuthSubjectRequest, opts ...grpc.CallOption) (*SetAuthSubjectResponse, error)
}

type sysAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewSysAPIClient(cc grpc.ClientConnInterface) SysAPIClient {
	return &sysAPIClient{cc}
}

func (c *sysAPIClient) AuthSubjects(ctx context.Context, in *AuthSubjectsRequest, opts ...grpc.CallOption) (SysAPI_AuthSubjectsClient, error) {
	stream, err := c.cc.NewStream(ctx, &SysAPI_ServiceDesc.Streams[0], SysAPI_AuthSubjects_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &sysAPIAuthSubjectsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type SysAPI_AuthSubjectsClient interface {
	Recv() (*AuthSubject, error)
	grpc.ClientStream
}

type sysAPIAuthSubjectsClient struct {
	grpc.ClientStream
}

func (x *sysAPIAuthSubjectsClient) Recv() (*AuthSubject, error) {
	m := new(AuthSubject)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *sysAPIClient) DeleteUser(ctx context.Context, in *DeleteUserRequest, opts ...grpc.CallOption) (*DeleteUserResponse, error) {
	out := new(DeleteUserResponse)
	err := c.cc.Invoke(ctx, SysAPI_DeleteUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sysAPIClient) UsersWithoutAuthProvider(ctx context.Context, in *UsersWithoutAuthProviderRequest, opts ...grpc.CallOption) (SysAPI_UsersWithoutAuthProviderClient, error) {
	stream, err := c.cc.NewStream(ctx, &SysAPI_ServiceDesc.Streams[1], SysAPI_UsersWithoutAuthProvider_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &sysAPIUsersWithoutAuthProviderClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type SysAPI_UsersWithoutAuthProviderClient interface {
	Recv() (*UserInfo, error)
	grpc.ClientStream
}

type sysAPIUsersWithoutAuthProviderClient struct {
	grpc.ClientStream
}

func (x *sysAPIUsersWithoutAuthProviderClient) Recv() (*UserInfo, error) {
	m := new(UserInfo)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *sysAPIClient) SetAuthSubject(ctx context.Context, in *SetAuthSubjectRequest, opts ...grpc.CallOption) (*SetAuthSubjectResponse, error) {
	out := new(SetAuthSubjectResponse)
	err := c.cc.Invoke(ctx, SysAPI_SetAuthSubject_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SysAPIServer is the server API for SysAPI service.
// All implementations must embed UnimplementedSysAPIServer
// for forward compatibility
type SysAPIServer interface {
	AuthSubjects(*AuthSubjectsRequest, SysAPI_AuthSubjectsServer) error
	DeleteUser(context.Context, *DeleteUserRequest) (*DeleteUserResponse, error)
	UsersWithoutAuthProvider(*UsersWithoutAuthProviderRequest, SysAPI_UsersWithoutAuthProviderServer) error
	SetAuthSubject(context.Context, *SetAuthSubjectRequest) (*SetAuthSubjectResponse, error)
	mustEmbedUnimplementedSysAPIServer()
}

// UnimplementedSysAPIServer must be embedded to have forward compatible implementations.
type UnimplementedSysAPIServer struct {
}

func (UnimplementedSysAPIServer) AuthSubjects(*AuthSubjectsRequest, SysAPI_AuthSubjectsServer) error {
	return status.Errorf(codes.Unimplemented, "method AuthSubjects not implemented")
}
func (UnimplementedSysAPIServer) DeleteUser(context.Context, *DeleteUserRequest) (*DeleteUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUser not implemented")
}
func (UnimplementedSysAPIServer) UsersWithoutAuthProvider(*UsersWithoutAuthProviderRequest, SysAPI_UsersWithoutAuthProviderServer) error {
	return status.Errorf(codes.Unimplemented, "method UsersWithoutAuthProvider not implemented")
}
func (UnimplementedSysAPIServer) SetAuthSubject(context.Context, *SetAuthSubjectRequest) (*SetAuthSubjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetAuthSubject not implemented")
}
func (UnimplementedSysAPIServer) mustEmbedUnimplementedSysAPIServer() {}

// UnsafeSysAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SysAPIServer will
// result in compilation errors.
type UnsafeSysAPIServer interface {
	mustEmbedUnimplementedSysAPIServer()
}

func RegisterSysAPIServer(s grpc.ServiceRegistrar, srv SysAPIServer) {
	s.RegisterService(&SysAPI_ServiceDesc, srv)
}

func _SysAPI_AuthSubjects_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(AuthSubjectsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SysAPIServer).AuthSubjects(m, &sysAPIAuthSubjectsServer{stream})
}

type SysAPI_AuthSubjectsServer interface {
	Send(*AuthSubject) error
	grpc.ServerStream
}

type sysAPIAuthSubjectsServer struct {
	grpc.ServerStream
}

func (x *sysAPIAuthSubjectsServer) Send(m *AuthSubject) error {
	return x.ServerStream.SendMsg(m)
}

func _SysAPI_DeleteUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SysAPIServer).DeleteUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SysAPI_DeleteUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SysAPIServer).DeleteUser(ctx, req.(*DeleteUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SysAPI_UsersWithoutAuthProvider_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(UsersWithoutAuthProviderRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SysAPIServer).UsersWithoutAuthProvider(m, &sysAPIUsersWithoutAuthProviderServer{stream})
}

type SysAPI_UsersWithoutAuthProviderServer interface {
	Send(*UserInfo) error
	grpc.ServerStream
}

type sysAPIUsersWithoutAuthProviderServer struct {
	grpc.ServerStream
}

func (x *sysAPIUsersWithoutAuthProviderServer) Send(m *UserInfo) error {
	return x.ServerStream.SendMsg(m)
}

func _SysAPI_SetAuthSubject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetAuthSubjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SysAPIServer).SetAuthSubject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SysAPI_SetAuthSubject_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SysAPIServer).SetAuthSubject(ctx, req.(*SetAuthSubjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// SysAPI_ServiceDesc is the grpc.ServiceDesc for SysAPI service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SysAPI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "goalert.v1.SysAPI",
	HandlerType: (*SysAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DeleteUser",
			Handler:    _SysAPI_DeleteUser_Handler,
		},
		{
			MethodName: "SetAuthSubject",
			Handler:    _SysAPI_SetAuthSubject_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "AuthSubjects",
			Handler:       _SysAPI_AuthSubjects_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "UsersWithoutAuthProvider",
			Handler:       _SysAPI_UsersWithoutAuthProvider_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pkg/sysapi/sysapi.proto",
}
