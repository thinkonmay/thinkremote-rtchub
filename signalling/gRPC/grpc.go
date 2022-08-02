package grpc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pigeatgarlic/webrtc-proxy/signalling"
	"github.com/pigeatgarlic/webrtc-proxy/signalling/gRPC/packet"
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCclient struct {
	packet.UnimplementedStreamServiceServer
	conn *grpc.ClientConn

	stream packet.StreamServiceClient
	client packet.StreamService_StreamRequestClient
	requestCount int

	sdpChan chan *webrtc.SessionDescription
	iceChan chan *webrtc.ICECandidateInit
	startChan chan bool
}


func InitGRPCClient (conf *config.GrpcConfig) (ret GRPCclient, err error) {
	ret.sdpChan = make(chan *webrtc.SessionDescription)
	ret.iceChan = make(chan *webrtc.ICECandidateInit)
	ret.startChan = make(chan bool)
	ret.conn,err = grpc.Dial(
		fmt.Sprintf("%s:%d",conf.ServerAddress,conf.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return;
	}


	// this is the critical step that includes your headers
	ctx := metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs("authorization",conf.Token),
	);

	ret.stream = packet.NewStreamServiceClient(ret.conn);
	ret.client,err = ret.stream.StreamRequest(ctx)
	if err != nil {
		fmt.Printf("fail to request stream: %s\n",err.Error());
		return;
	}

	ret.requestCount = 0;
	go func() {
		for {
			res,err := ret.client.Recv()
			if err != nil {
				fmt.Println(err.Error());
				return;
			}
			if len(res.Error) != 0 {
				fmt.Println(res.Error);
			}
			if res.Data["Target"] == "SDP" {
				var sdp webrtc.SessionDescription;	

				sdp.SDP		= res.Data["SDP"];
				sdp.Type	= webrtc.NewSDPType(res.Data["Type"]);

				fmt.Printf("SDP received: %s\n",res.Data["Type"])
				ret.sdpChan <- &sdp;
			} else if res.Data["Target"] == "ICE" {
				var ice webrtc.ICECandidateInit;

				ice.Candidate  =	res.Data["Candidate"] 
				SDPMid        :=	res.Data["SDPMid"] 
				ice.SDPMid     = 	&SDPMid;

				LineIndex,_ 	 :=	strconv.Atoi(res.Data["SDPMLineIndex"])
				LineIndexint	 := uint16(LineIndex)
				ice.SDPMLineIndex = &LineIndexint;			

				fmt.Printf("ICE received\n")
				ret.iceChan <- &ice;
			} else if res.Data["Target"] == "START" {
				ret.startChan <- true;
			} else {
				fmt.Println("Unknown packet");
			}
		}	
	}()
	return;
}

func (client *GRPCclient) SendSDP(desc *webrtc.SessionDescription) error {
	req := packet.UserRequest{
		Id: (int64) (client.requestCount),
		Target: "SDP",
		Headers: map[string]string{},
		Data: map[string]string{
			"SDP": desc.SDP,
			"Type": desc.Type.String(),
		},
	}
	fmt.Printf("SDP send %s\n",req.Data["Type"])
	if err := client.client.Send(&req); err != nil {
		return err;
	}
	client.requestCount++;
	return nil;
}

func (client *GRPCclient) SendICE(ice *webrtc.ICECandidateInit) error {
	req := packet.UserRequest{
		Id: (int64) (client.requestCount),
		Target: "ICE",
		Headers: map[string]string{},
		Data: map[string]string{
			"Candidate":     ice.Candidate,
			"SDPMid":        *ice.SDPMid,
			"SDPMLineIndex": fmt.Sprintf("%d",*ice.SDPMLineIndex),
		},
	}
	fmt.Printf("ICE sent\n");
	if err := client.client.Send(&req); err != nil {
		return err;
	}
	client.requestCount++;
	return nil;

}

func (client *GRPCclient) OnICE(fun signalling.OnIceFunc) {
	go func() {
		for {
			ice := <- client.iceChan;
			fun(ice);
		}
	}()
}

func (client *GRPCclient) OnSDP(fun signalling.OnSDPFunc) {
	go func() {
		for {
			sdp := <- client.sdpChan;
			fun(sdp);
		}
	}()
}

func (client *GRPCclient) WaitForStart(){
	<- client.startChan;
}
func (client *GRPCclient) Stop(){
	client.conn.Close()
}
