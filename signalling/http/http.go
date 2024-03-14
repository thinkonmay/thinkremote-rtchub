package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/signalling"
	"github.com/thinkonmay/thinkremote-rtchub/signalling/gRPC/packet"
)

type WebsocketClient struct {
	sdpChan chan *webrtc.SessionDescription
	iceChan chan *webrtc.ICECandidateInit

	incoming  chan *packet.SignalingMessage
	outcoming chan *packet.SignalingMessage

	done      bool
	connected bool
}

func InitHttpClient(AddressStr string) (_ signalling.Signalling, err error) {
	ret := &WebsocketClient{
		sdpChan: make(chan *webrtc.SessionDescription, 8),
		iceChan: make(chan *webrtc.ICECandidateInit, 8),

		incoming:  make(chan *packet.SignalingMessage, 8),
		outcoming: make(chan *packet.SignalingMessage, 8),

		connected: false,
		done:      false,
	}

	u, err := url.Parse(AddressStr)
	if err != nil {
		return
	}
	q := u.Query()
	q.Add("uniqueid", uuid.New().String())
	u.RawQuery = q.Encode()

	go func() {
		for {
			time.Sleep(300 * time.Millisecond)
			if ret.done {
				ret.iceChan <- nil
				ret.sdpChan <- nil
				ret.incoming <- nil
				return
			}

			pkt := []packet.SignalingMessage{}
			for {
				if len(ret.outcoming) == 0 {
					break
				}
				out := <-ret.outcoming
				pkt = append(pkt, *out)
			}

			b, _ := json.Marshal(pkt)
			resp, err := http.DefaultClient.Post(
				u.String(),
				"application/json",
				strings.NewReader(string(b)),
			)
			if err != nil {
				fmt.Printf("failed to send http %s", err.Error())
				continue
			}

			b, err = io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("failed to read http body %s", err.Error())
			}

			err = json.Unmarshal(b, &pkt)
			if err != nil {
				fmt.Printf("failed to read http body %s", err.Error())
			}

			for _, sm := range pkt {
				ret.incoming <- &sm
			}
		}
	}()

	go func() {
		for {
			res := <-ret.incoming
			if res == nil {
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
	return ret, nil
}

func (client *WebsocketClient) SendSDP(desc *webrtc.SessionDescription) error {
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
	client.outcoming <- &req
	return nil
}

func (client *WebsocketClient) SendICE(ice *webrtc.ICECandidateInit) error {
	if !client.connected {
		return fmt.Errorf("signaling client is closed")
	}

	req := packet.SignalingMessage{
		Type: packet.SignalingType_tICE,
		Ice: &packet.ICE{
			SDPMid:        *ice.SDPMid,
			SDPMLineIndex: int64(*ice.SDPMLineIndex),
			Candidate:     ice.Candidate,
		},
	}

	fmt.Printf("ICE sent %v\n", req.Ice)
	client.outcoming <- &req
	return nil
}

func (client *WebsocketClient) OnICE(fun signalling.OnIceFunc) {
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

func (client *WebsocketClient) OnSDP(fun signalling.OnSDPFunc) {
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

func (client *WebsocketClient) WaitForStart() {
	for {
		time.Sleep(time.Second)
		if client.connected {
			return
		}
	}
}

func (client *WebsocketClient) WaitForEnd() {
	for {
		time.Sleep(time.Second)
		if client.done {
			return
		}
	}
}

func (client *WebsocketClient) Stop() {
	client.connected = false
	client.done = true
}
