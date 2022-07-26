package grpc

import (
	"fmt"
	"net"
	"time"

	"github.com/pigeatgarlic/webrtc-proxy/cmd/signalling/protocol"
	"github.com/pigeatgarlic/webrtc-proxy/signalling/gRPC/packet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)


type GrpcServer struct {
	packet.UnimplementedStreamServiceServer
	grpcServer *grpc.Server
	fun protocol.OnTenantFunc
}

func (server *GrpcServer) OnTenant(fun protocol.OnTenantFunc) {
	server.fun = fun
}

type GrpcTenant struct {
	exited bool
	client packet.StreamService_StreamRequestServer
}

func (tenant *GrpcTenant) Send(pkt *packet.UserResponse) {
	err := tenant.client.Send(pkt);
	if err != nil {
		tenant.exited = true;
	}
}

func (tenant *GrpcTenant) Receive() *packet.UserRequest {
	req, err := tenant.client.Recv();
	if err != nil {
		tenant.exited = true;
		return nil;
	} else {
		return req;
	}
}

func (tenant *GrpcTenant) Exit() {
	tenant.exited = true;
}

func (tenant *GrpcTenant) IsExited() bool {
	return tenant.exited
}




func InitSignallingServer(conf *protocol.SignalingConfig) (*GrpcServer) {
	var ret GrpcServer;
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", conf.GrpcPort))
	if err != nil {
		panic(err)
	}
	ret.grpcServer = grpc.NewServer()
	packet.RegisterStreamServiceServer(ret.grpcServer, &ret)
	go ret.grpcServer.Serve(lis)
	return &ret
}


func (server *GrpcServer) StreamRequest(client packet.StreamService_StreamRequestServer) error {
	var tenant *GrpcTenant;
	md, ok := metadata.FromIncomingContext(client.Context())
	if !ok {
		return fmt.Errorf("Unauthorized")
	} else {
		token := md["Authorization"];
		tenant = &GrpcTenant{
			exited: false,
			client: client,
		}
		err := server.fun(token[0],tenant);
		if err != nil {
			tenant.exited = true;
		}
	}
	for {
		if tenant.exited {
			return nil;
		}
		time.Sleep(time.Millisecond)
	}
}