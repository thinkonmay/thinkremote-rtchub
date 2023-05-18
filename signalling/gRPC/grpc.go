package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/signalling"
	"github.com/thinkonmay/thinkremote-rtchub/signalling/gRPC/packet"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type GRPCclient struct {
	packet.UnimplementedSignalingServer
	conn *grpc.ClientConn

	stream packet.SignalingClient
	client packet.Signaling_HandshakeClient

	sdpChan      chan *webrtc.SessionDescription
	iceChan      chan *webrtc.ICECandidateInit

	done      bool
	connected bool
}

func InitGRPCClient(AddressStr string,
					auth *config.AuthConfig,
) (ret *GRPCclient, err error) {
	ret = &GRPCclient{
		sdpChan: make(chan *webrtc.SessionDescription),
		iceChan: make(chan *webrtc.ICECandidateInit),

		connected: false,
		done:      false,
	}

	ret.conn, err = grpc.Dial(
		AddressStr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return
	}

	// this is the critical step that includes your headers
	ctx := metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs("authorization", auth.Token),
	)

	ret.stream = packet.NewSignalingClient(ret.conn)
	ret.client, err = ret.stream.Handshake(ctx)
	if err != nil {
		fmt.Printf("fail to request stream: %s\n", err.Error())
		return
	}

	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			if !ret.done {
				continue
			}
			ret.iceChan<-nil
			ret.sdpChan<-nil
			ret.conn.Close()
		}
	}()
	go func() {
		for {
			res, err := ret.client.Recv()
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				fmt.Printf("grpc connection terminated\n")
				ret.Stop()
				return
			}

			switch res.Type {
			case packet.SignalingType_tSDP:
				sdp := &webrtc.SessionDescription{}
				sdp.SDP = res.Sdp.SDPData
				sdp.Type = webrtc.NewSDPType(res.Sdp.Type)
				fmt.Printf("SDP received: %s\n", res.Sdp.Type)
				ret.sdpChan <- sdp
			case packet.SignalingType_tICE:
				ice := &webrtc.ICECandidateInit{}

				ice.Candidate = res.Ice.Candidate
				SDPMid := res.Ice.SDPMid
				ice.SDPMid = &SDPMid
				LineIndex := uint16(res.Ice.SDPMLineIndex)
				ice.SDPMLineIndex = &LineIndex

				fmt.Printf("ICE received\n")
				ret.iceChan <- ice
			case packet.SignalingType_tSTART:
				ret.connected = true
			case packet.SignalingType_tEND:
				ret.Stop()
			default:
				fmt.Println("Unknown packet")
			}
		}
	}()
	return
}

func (client *GRPCclient) SendSDP(desc *webrtc.SessionDescription) error {
	if !client.connected {
		return fmt.Errorf("signaling client is closed")
	}

	req := packet.SignalingMessage{
		Type: packet.SignalingType_tSDP,
		Sdp: &packet.SDP{
			Type:    desc.Type.String(),
			SDPData: desc.SDP,
		},
	}

	fmt.Printf("SDP send %s\n", req.Sdp.Type)
	if err := client.client.Send(&req); err != nil {
		return err
	}
	return nil
}

func (client *GRPCclient) SendICE(ice *webrtc.ICECandidateInit) error {
	if !client.connected {
		return fmt.Errorf("signaling client is closed")
	}

	req := &packet.SignalingMessage{
		Type: packet.SignalingType_tICE,
		Ice: &packet.ICE{
			SDPMid:        *ice.SDPMid,
			SDPMLineIndex: int64(*ice.SDPMLineIndex),
			Candidate:     ice.Candidate,
		},
	}

	fmt.Printf("ICE sent %v\n", req.Ice)
	if err := client.client.Send(req); err != nil {
		return err
	}
	return nil
}

func (client *GRPCclient) OnICE(fun signalling.OnIceFunc) {
	go func() {
		for {
			ice := <-client.iceChan
			if ice == nil {
				return
			}
			if !client.connected {
				continue
			}
			fun(ice)
		}
	}()
}

func (client *GRPCclient) OnSDP(fun signalling.OnSDPFunc) {
	go func() {
		for {
			sdp := <-client.sdpChan
			if sdp == nil {
				return
			}
			if !client.connected {
				continue
			}
			fun(sdp)
		}
	}()
}

func (client *GRPCclient) WaitForStart() {
	for {
		if client.connected {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (client *GRPCclient) WaitForEnd() {
	for {
		if client.done {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (client *GRPCclient) Stop() {
	client.connected = false
	client.done = true
}
