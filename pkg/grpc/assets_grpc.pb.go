// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
// source: assets.proto

package assets_grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	AssetsService_RevalidateChampionCache_FullMethodName = "/assets_grpc.assetsService/RevalidateChampionCache"
	AssetsService_RevalidateItemCache_FullMethodName     = "/assets_grpc.assetsService/RevalidateItemCache"
)

// AssetsServiceClient is the client API for AssetsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Service for revalidanting the Redis cache.
// Will be called when the champion or item is not found on the Redis.
// Return the given champion or item for the id so we don't need to do redundant calls to Redis.
type AssetsServiceClient interface {
	RevalidateChampionCache(ctx context.Context, in *ChampionId, opts ...grpc.CallOption) (*Champion, error)
	RevalidateItemCache(ctx context.Context, in *ItemId, opts ...grpc.CallOption) (*Item, error)
}

type assetsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAssetsServiceClient(cc grpc.ClientConnInterface) AssetsServiceClient {
	return &assetsServiceClient{cc}
}

func (c *assetsServiceClient) RevalidateChampionCache(ctx context.Context, in *ChampionId, opts ...grpc.CallOption) (*Champion, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Champion)
	err := c.cc.Invoke(ctx, AssetsService_RevalidateChampionCache_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *assetsServiceClient) RevalidateItemCache(ctx context.Context, in *ItemId, opts ...grpc.CallOption) (*Item, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Item)
	err := c.cc.Invoke(ctx, AssetsService_RevalidateItemCache_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AssetsServiceServer is the server API for AssetsService service.
// All implementations must embed UnimplementedAssetsServiceServer
// for forward compatibility.
//
// Service for revalidanting the Redis cache.
// Will be called when the champion or item is not found on the Redis.
// Return the given champion or item for the id so we don't need to do redundant calls to Redis.
type AssetsServiceServer interface {
	RevalidateChampionCache(context.Context, *ChampionId) (*Champion, error)
	RevalidateItemCache(context.Context, *ItemId) (*Item, error)
	mustEmbedUnimplementedAssetsServiceServer()
}

// UnimplementedAssetsServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedAssetsServiceServer struct{}

func (UnimplementedAssetsServiceServer) RevalidateChampionCache(context.Context, *ChampionId) (*Champion, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RevalidateChampionCache not implemented")
}
func (UnimplementedAssetsServiceServer) RevalidateItemCache(context.Context, *ItemId) (*Item, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RevalidateItemCache not implemented")
}
func (UnimplementedAssetsServiceServer) mustEmbedUnimplementedAssetsServiceServer() {}
func (UnimplementedAssetsServiceServer) testEmbeddedByValue()                       {}

// UnsafeAssetsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AssetsServiceServer will
// result in compilation errors.
type UnsafeAssetsServiceServer interface {
	mustEmbedUnimplementedAssetsServiceServer()
}

func RegisterAssetsServiceServer(s grpc.ServiceRegistrar, srv AssetsServiceServer) {
	// If the following call pancis, it indicates UnimplementedAssetsServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&AssetsService_ServiceDesc, srv)
}

func _AssetsService_RevalidateChampionCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChampionId)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AssetsServiceServer).RevalidateChampionCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AssetsService_RevalidateChampionCache_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AssetsServiceServer).RevalidateChampionCache(ctx, req.(*ChampionId))
	}
	return interceptor(ctx, in, info, handler)
}

func _AssetsService_RevalidateItemCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ItemId)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AssetsServiceServer).RevalidateItemCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AssetsService_RevalidateItemCache_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AssetsServiceServer).RevalidateItemCache(ctx, req.(*ItemId))
	}
	return interceptor(ctx, in, info, handler)
}

// AssetsService_ServiceDesc is the grpc.ServiceDesc for AssetsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AssetsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "assets_grpc.assetsService",
	HandlerType: (*AssetsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RevalidateChampionCache",
			Handler:    _AssetsService_RevalidateChampionCache_Handler,
		},
		{
			MethodName: "RevalidateItemCache",
			Handler:    _AssetsService_RevalidateItemCache_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "assets.proto",
}
