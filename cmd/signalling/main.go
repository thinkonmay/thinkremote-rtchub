package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pigeatgarlic/webrtc-proxy/signalling/gRPC/packet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type SignalingServerConfig struct {
	Port int
}

type SignallingServer struct {
	packet.UnimplementedStreamServiceServer

	grpcServer *grpc.Server

	reqChannel map[int64]*chan packet.UserRequest
	mutex *sync.RWMutex

	connectCount int
}

func initSignallingServer (conf *SignalingServerConfig) (ret SignallingServer) {
	ret.reqChannel = make(map[int64]*chan packet.UserRequest, 1000)
	ret.mutex = &sync.RWMutex{}
	ret.connectCount = 0;
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", conf.Port))
	if err != nil {
		panic(err);
	}	
	ret.grpcServer = grpc.NewServer()
	packet.RegisterStreamServiceServer(ret.grpcServer, &ret)
	go ret.grpcServer.Serve(lis);
	return;
}

func (server *SignallingServer) StreamRequest(client packet.StreamService_StreamRequestServer) error {
	_, ok := metadata.FromIncomingContext(client.Context())
	if !ok {
		return fmt.Errorf("Unauthorized")
	} else {
		server.connectCount++;
		if server.connectCount == 2 {
			var res packet.UserResponse;

			res.Id = 0;
			res.Error = "";
			res.Data = make(map[string]string)
			res.Data["Target"] = "START";

			err := client.Send(&res)
			if err != nil {
				return err;
			}

		}
	}

	// TODO auth
	// _ := headers["Authorization"]
	// usr, err = watcher.auth.ValidateToken(token[0], "User")
	// if err != nil {
	// 	return nil
	// }

	this := make(chan packet.UserRequest)
	shutdown := make(chan bool)
	rand := time.Now().UTC().UnixMilli()
	server.reqChannel[rand] = &this;

	defer func ()  {
		server.connectCount--;
		shutdown <- true;
		server.mutex.Lock();
		delete(server.reqChannel,rand);
		close(this);
		close(shutdown);
		server.mutex.Unlock();
		return;
	}();


	go func() {
		for {
			select{
			case <-shutdown:
				return;
			case req := <- this:	
				var res packet.UserResponse;

				res.Id = req.Id;
				res.Error = "";
				res.Data = req.Data;
				res.Data["Target"] = req.Target;

				err := client.Send(&res)
				if err != nil {
					return;
				}
			default:
			}
		}
	}()

	for {
		req, err := client.Recv()
		if err != nil {
			return nil
		}else {
			fmt.Printf("new request: %s\n",req.Target);
		}

		server.mutex.Lock();
		for index,channel := range server.reqChannel {
			if index == rand {
				continue;
			}
			var clone packet.UserRequest;
			clone = *req;
			*channel <- clone;
		}
		server.mutex.Unlock();
	}
}


func main() {
	conf := SignalingServerConfig{
		Port: 8000,
	}
	shutdown_channel := make(chan bool);
	initSignallingServer(&conf)
	<-shutdown_channel;
}