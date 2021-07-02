// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pluginextensionv2

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// VersionedStorageClient is the client API for VersionedStorage service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type VersionedStorageClient interface {
	Store(ctx context.Context, in *StoreRequest, opts ...grpc.CallOption) (*StoreResponse, error)
	Latest(ctx context.Context, in *LatestRequest, opts ...grpc.CallOption) (*LatestResponse, error)
	History(ctx context.Context, in *HistoryRequest, opts ...grpc.CallOption) (*HistoryResponse, error)
}

type versionedStorageClient struct {
	cc grpc.ClientConnInterface
}

func NewVersionedStorageClient(cc grpc.ClientConnInterface) VersionedStorageClient {
	return &versionedStorageClient{cc}
}

func (c *versionedStorageClient) Store(ctx context.Context, in *StoreRequest, opts ...grpc.CallOption) (*StoreResponse, error) {
	out := new(StoreResponse)
	err := c.cc.Invoke(ctx, "/pluginextensionv2.VersionedStorage/Store", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *versionedStorageClient) Latest(ctx context.Context, in *LatestRequest, opts ...grpc.CallOption) (*LatestResponse, error) {
	out := new(LatestResponse)
	err := c.cc.Invoke(ctx, "/pluginextensionv2.VersionedStorage/Latest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *versionedStorageClient) History(ctx context.Context, in *HistoryRequest, opts ...grpc.CallOption) (*HistoryResponse, error) {
	out := new(HistoryResponse)
	err := c.cc.Invoke(ctx, "/pluginextensionv2.VersionedStorage/History", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VersionedStorageServer is the server API for VersionedStorage service.
// All implementations must embed UnimplementedVersionedStorageServer
// for forward compatibility
type VersionedStorageServer interface {
	Store(context.Context, *StoreRequest) (*StoreResponse, error)
	Latest(context.Context, *LatestRequest) (*LatestResponse, error)
	History(context.Context, *HistoryRequest) (*HistoryResponse, error)
	mustEmbedUnimplementedVersionedStorageServer()
}

// UnimplementedVersionedStorageServer must be embedded to have forward compatible implementations.
type UnimplementedVersionedStorageServer struct {
}

func (*UnimplementedVersionedStorageServer) Store(context.Context, *StoreRequest) (*StoreResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Store not implemented")
}
func (*UnimplementedVersionedStorageServer) Latest(context.Context, *LatestRequest) (*LatestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Latest not implemented")
}
func (*UnimplementedVersionedStorageServer) History(context.Context, *HistoryRequest) (*HistoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method History not implemented")
}
func (*UnimplementedVersionedStorageServer) mustEmbedUnimplementedVersionedStorageServer() {}

func RegisterVersionedStorageServer(s *grpc.Server, srv VersionedStorageServer) {
	s.RegisterService(&_VersionedStorage_serviceDesc, srv)
}

func _VersionedStorage_Store_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StoreRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VersionedStorageServer).Store(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pluginextensionv2.VersionedStorage/Store",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VersionedStorageServer).Store(ctx, req.(*StoreRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _VersionedStorage_Latest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LatestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VersionedStorageServer).Latest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pluginextensionv2.VersionedStorage/Latest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VersionedStorageServer).Latest(ctx, req.(*LatestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _VersionedStorage_History_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VersionedStorageServer).History(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pluginextensionv2.VersionedStorage/History",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VersionedStorageServer).History(ctx, req.(*HistoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _VersionedStorage_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pluginextensionv2.VersionedStorage",
	HandlerType: (*VersionedStorageServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Store",
			Handler:    _VersionedStorage_Store_Handler,
		},
		{
			MethodName: "Latest",
			Handler:    _VersionedStorage_Latest_Handler,
		},
		{
			MethodName: "History",
			Handler:    _VersionedStorage_History_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "vcs.proto",
}