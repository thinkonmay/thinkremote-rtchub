package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/OnePlay-Internet/webrtc-proxy/signalling"
	"github.com/OnePlay-Internet/webrtc-proxy/signalling/gRPC/packet"
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type deviceSelection struct {
	SoundCard string	
	Monitor int
	Bitrate int			
	Framerate int		
}

type GRPCclient struct {
	packet.UnimplementedStreamServiceServer
	conn *grpc.ClientConn

	stream packet.StreamServiceClient
	client packet.StreamService_StreamRequestClient
	requestCount int

	deviceAvailableSent *tool.MediaDevice
	sdpChan chan *webrtc.SessionDescription
	iceChan chan *webrtc.ICECandidateInit
	preflightChan chan deviceSelection

	deviceSelected bool
	connected bool

	shutdown chan bool
}


func InitGRPCClient(conf *config.GrpcConfig,
					 devices *tool.MediaDevice,
					 shutdown chan bool) (ret *GRPCclient, err error) {
	ret = &GRPCclient{
		sdpChan : make(chan *webrtc.SessionDescription),
		iceChan : make(chan *webrtc.ICECandidateInit),
		preflightChan : make(chan deviceSelection),
		shutdown: shutdown,

		deviceSelected : false,
		connected: false,
	}

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
				fmt.Printf("%s\n",err.Error());
				if err != io.EOF {
					ret.shutdown<-true;
				} else if !ret.deviceSelected {
					fmt.Printf("grpc connection terminated while waiting for peer, terminating...\n");
					ret.shutdown<-true;
				}

				return;
			}

			switch res.Data["Target"] {
			case "SDP":
				var sdp webrtc.SessionDescription;	

				sdp.SDP		= res.Data["SDP"];
				sdp.Type	= webrtc.NewSDPType(res.Data["Type"]);

				fmt.Printf("SDP received: %s\n",res.Data["Type"])
				ret.sdpChan <- &sdp;
			case "ICE" :
				var ice webrtc.ICECandidateInit;

				ice.Candidate  =	res.Data["Candidate"] 
				SDPMid        :=	res.Data["SDPMid"] 
				ice.SDPMid     = 	&SDPMid;

				LineIndex,_ 	 :=	strconv.Atoi(res.Data["SDPMLineIndex"])
				LineIndexint	 := uint16(LineIndex)
				ice.SDPMLineIndex = &LineIndexint;			

				fmt.Printf("ICE received\n")
				ret.iceChan <- &ice;
			case "START":
				fmt.Printf("Receive start signal\n");
				ret.SendDeviceAvailable(devices,nil);
				ret.connected = true;
			case "PREFLIGHT" :
				bitrate,err 	:= strconv.ParseInt(res.Data["bitrate"],10,32);
				framerate,err 	:= strconv.ParseInt(res.Data["framerate"],10,32);
				monitor,err 	:= strconv.ParseInt(res.Data["monitor"],10,32);
				if err == nil {
					ret.preflightChan<-deviceSelection{
						Bitrate: int(bitrate),
						Framerate: int(framerate),
						Monitor: int(monitor),
						SoundCard: res.Data["soundcard"],	
					};
				}
			default:
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

func (client *GRPCclient) SendDeviceAvailable(devices *tool.MediaDevice, preverr error) error {
	data,err := json.Marshal(devices);
	if err != nil {
		return err;
	}
	
	client.deviceAvailableSent = devices

	req := packet.UserRequest{
		Id: (int64) (client.requestCount),
		Target: "PREFLIGHT",
		Headers: map[string]string{},
		Data: map[string]string{
			"Devices": string(data),
		},
	}

	if preverr != nil {
		req.Data["Error"] = preverr.Error()
	}

	fmt.Printf("PREFLIGHT sent\n");
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

func (client *GRPCclient) OnDeviceSelect(fun signalling.OnDeviceSelectFunc) {
	go func() {
		for {
			devsec := <- client.preflightChan;
			if client.deviceAvailableSent == nil {
				fmt.Printf("receive preflight when haven't started\n");
				continue;
			}
			monitor := func () tool.Monitor  {
				for _,monitor := range client.deviceAvailableSent.Monitors {
					if monitor.MonitorHandle == devsec.Monitor {
						monitor.Framerate = devsec.Framerate
						return monitor
					}
				}
				return tool.Monitor{MonitorHandle: -1}
			}()
			soundcard := func () tool.Soundcard {
				for _,soundcard := range client.deviceAvailableSent.Soundcards {
					if soundcard.DeviceID == devsec.SoundCard {
						return soundcard
					}
				}
				return tool.Soundcard{DeviceID: "none"}
			}()

			err := fun(monitor,soundcard);
			if err != nil {
				client.SendDeviceAvailable(client.deviceAvailableSent,err); 
			} else {
				client.deviceSelected = true;
			}
		}
	}()
}

func (client *GRPCclient) WaitForStart(){
	for {
		if client.deviceSelected{
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}
func (client *GRPCclient) WaitForConnected(){
	for {
		if client.connected {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}
func (client *GRPCclient) Stop(){
	client.conn.Close()
}
