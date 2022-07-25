package main

import (
	"fmt"
	"net"
	"sync"

	"github.com/pigeatgarlic/webrtc-proxy/signalling/gRPC/packet"
	"github.com/pigeatgarlic/webrtc-proxy/cmd/signalling"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)


type SignallingServer struct {
	packet.UnimplementedStreamServiceServer

	grpcServer *grpc.Server

	reqChannel map[int64]packet.StreamService_StreamRequestServer

	mutex *sync.RWMutex

	connectCount int
}

func initSignallingServer(conf *signalling.SignalingConfig) (ret SignallingServer) {
	ret.mutex = &sync.RWMutex{}
	ret.reqChannel = make(map[int64]packet.StreamService_StreamRequestServer)
	ret.connectCount = 0
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", conf.GrpcPort))
	if err != nil {
		panic(err)
	}
	ret.grpcServer = grpc.NewServer()
	packet.RegisterStreamServiceServer(ret.grpcServer, &ret)
	go ret.grpcServer.Serve(lis)
	return
}

func (server *SignallingServer) broadcast(src int64, res *packet.UserResponse) error {
	for key, val := range server.reqChannel {
		if key != src {
			fmt.Printf("response to client %d: %s\n", key, res.Data["Target"])
			err := val.Send(res)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (server *SignallingServer) StreamRequest(client packet.StreamService_StreamRequestServer) error {
	_, ok := metadata.FromIncomingContext(client.Context())
	if !ok {
		return fmt.Errorf("Unauthorized")
	} else {
		server.connectCount++
		if server.connectCount%2 == 0 {
			var res packet.UserResponse

			res.Id = 0
			res.Error = ""
			res.Data = make(map[string]string)
			res.Data["Target"] = "START"

			err := server.broadcast(0, &res)
			if err != nil {
				return err
			}
		}
	}

	shutdown := false
	rand := int64(server.connectCount)

	fmt.Printf("new client %d\n", rand)
	server.mutex.Lock()
	server.reqChannel[rand] = client
	server.mutex.Unlock()

	defer func() {
		fmt.Printf("client %d exited\n", rand)
		delete(server.reqChannel, rand)
		shutdown = true
	}()

	for {
		req, err := client.Recv()
		if err != nil || shutdown {
			return nil
		} else {
			fmt.Printf("new request from client %d: %s\n", rand, req.Target)
		}

		var res packet.UserResponse
		res.Id = req.Id
		res.Error = ""
		res.Data = req.Data
		res.Data["Target"] = req.Target

		server.mutex.Lock()
		server.broadcast(rand, &res)
		server.mutex.Unlock()
	}
}