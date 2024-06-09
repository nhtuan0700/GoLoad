// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: api/go_load.proto

package go_load

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

// GoLoadServiceClient is the client API for GoLoadService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GoLoadServiceClient interface {
	CreateAccount(ctx context.Context, in *CreateAccountRequest, opts ...grpc.CallOption) (*CreateAccountResponse, error)
}

type goLoadServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGoLoadServiceClient(cc grpc.ClientConnInterface) GoLoadServiceClient {
	return &goLoadServiceClient{cc}
}

func (c *goLoadServiceClient) CreateAccount(ctx context.Context, in *CreateAccountRequest, opts ...grpc.CallOption) (*CreateAccountResponse, error) {
	out := new(CreateAccountResponse)
	err := c.cc.Invoke(ctx, "/go_load.GoLoadService/CreateAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GoLoadServiceServer is the server API for GoLoadService service.
// All implementations must embed UnimplementedGoLoadServiceServer
// for forward compatibility
type GoLoadServiceServer interface {
	CreateAccount(context.Context, *CreateAccountRequest) (*CreateAccountResponse, error)
	mustEmbedUnimplementedGoLoadServiceServer()
}

// UnimplementedGoLoadServiceServer must be embedded to have forward compatible implementations.
type UnimplementedGoLoadServiceServer struct {
}

func (UnimplementedGoLoadServiceServer) CreateAccount(context.Context, *CreateAccountRequest) (*CreateAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAccount not implemented")
}
func (UnimplementedGoLoadServiceServer) mustEmbedUnimplementedGoLoadServiceServer() {}

// UnsafeGoLoadServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GoLoadServiceServer will
// result in compilation errors.
type UnsafeGoLoadServiceServer interface {
	mustEmbedUnimplementedGoLoadServiceServer()
}

func RegisterGoLoadServiceServer(s grpc.ServiceRegistrar, srv GoLoadServiceServer) {
	s.RegisterService(&GoLoadService_ServiceDesc, srv)
}

func _GoLoadService_CreateAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoLoadServiceServer).CreateAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/go_load.GoLoadService/CreateAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoLoadServiceServer).CreateAccount(ctx, req.(*CreateAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GoLoadService_ServiceDesc is the grpc.ServiceDesc for GoLoadService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GoLoadService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "go_load.GoLoadService",
	HandlerType: (*GoLoadServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateAccount",
			Handler:    _GoLoadService_CreateAccount_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/go_load.proto",
}