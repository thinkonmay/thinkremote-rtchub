// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.1
// source: protobuf.proto

package packet

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

// SignalingClient is the client API for Signaling service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SignalingClient interface {
	Handshake(ctx context.Context, opts ...grpc.CallOption) (Signaling_HandshakeClient, error)
}

type signalingClient struct {
	cc grpc.ClientConnInterface
}

func NewSignalingClient(cc grpc.ClientConnInterface) SignalingClient {
	return &signalingClient{cc}
}

func (c *signalingClient) Handshake(ctx context.Context, opts ...grpc.CallOption) (Signaling_HandshakeClient, error) {
	stream, err := c.cc.NewStream(ctx, &Signaling_ServiceDesc.Streams[0], "/protobuf.Signaling/handshake", opts...)
	if err != nil {
		return nil, err
	}
	x := &signalingHandshakeClient{stream}
	return x, nil
}

type Signaling_HandshakeClient interface {
	Send(*SignalingMessage) error
	Recv() (*SignalingMessage, error)
	grpc.ClientStream
}

type signalingHandshakeClient struct {
	grpc.ClientStream
}

func (x *signalingHandshakeClient) Send(m *SignalingMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *signalingHandshakeClient) Recv() (*SignalingMessage, error) {
	m := new(SignalingMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// SignalingServer is the server API for Signaling service.
// All implementations must embed UnimplementedSignalingServer
// for forward compatibility
type SignalingServer interface {
	Handshake(Signaling_HandshakeServer) error
	mustEmbedUnimplementedSignalingServer()
}

// UnimplementedSignalingServer must be embedded to have forward compatible implementations.
type UnimplementedSignalingServer struct {
}

func (UnimplementedSignalingServer) Handshake(Signaling_HandshakeServer) error {
	return status.Errorf(codes.Unimplemented, "method Handshake not implemented")
}
func (UnimplementedSignalingServer) mustEmbedUnimplementedSignalingServer() {}

// UnsafeSignalingServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SignalingServer will
// result in compilation errors.
type UnsafeSignalingServer interface {
	mustEmbedUnimplementedSignalingServer()
}

func RegisterSignalingServer(s grpc.ServiceRegistrar, srv SignalingServer) {
	s.RegisterService(&Signaling_ServiceDesc, srv)
}

func _Signaling_Handshake_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SignalingServer).Handshake(&signalingHandshakeServer{stream})
}

type Signaling_HandshakeServer interface {
	Send(*SignalingMessage) error
	Recv() (*SignalingMessage, error)
	grpc.ServerStream
}

type signalingHandshakeServer struct {
	grpc.ServerStream
}

func (x *signalingHandshakeServer) Send(m *SignalingMessage) error {
	return x.ServerStream.SendMsg(m)
}

func (x *signalingHandshakeServer) Recv() (*SignalingMessage, error) {
	m := new(SignalingMessage)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Signaling_ServiceDesc is the grpc.ServiceDesc for Signaling service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Signaling_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "protobuf.Signaling",
	HandlerType: (*SignalingServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "handshake",
			Handler:       _Signaling_Handshake_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "protobuf.proto",
}
