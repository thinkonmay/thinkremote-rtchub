package grpc

import (
	"context"
	"fmt"

	"github.com/pigeatgarlic/webrtc-proxy/signalling/gRPC/packet"
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCclient struct {
	packet.UnimplementedStreamServiceServer
	conn *grpc.ClientConn

	stream packet.StreamServiceClient

}


func InitGRPCClient (conf *config.GrpcConfig) (ret *GRPCclient, err error) {
	ret = &GRPCclient{};
	ret.conn,err = grpc.Dial(
		fmt.Sprintf("%s:%d",conf.ServerAddress,conf.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return;
	}


	ret.stream = packet.NewStreamServiceClient(ret.conn);
	ret.stream.StreamRequest(context.Background())
	return;
}

func (client *GRPCclient) SendSDP() {

}

func (client *GRPCclient) SendICE() {

}

func (client *GRPCclient) OnICE() {

}

func (client *GRPCclient) OnSDP() {

}
