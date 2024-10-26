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
	"github.com/pion/webrtc/v4"
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

	exchange := func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("panic in http thread %v", err)
			}
		}()

		pkt := []packet.SignalingMessage{}
		for len(ret.outcoming) > 0 {
			out := <-ret.outcoming
			pkt = append(pkt, *out)
		}

		if b, err := json.Marshal(pkt); err != nil {
			return
		} else if resp, err := http.DefaultClient.Post(
			u.String(),
			"application/json",
			strings.NewReader(string(b)),
		); err != nil {
			return
		} else if b, err := io.ReadAll(resp.Body); err != nil {
			return
		} else if err = json.Unmarshal(b, &pkt); err != nil {
			return
		} else {
			for _, sm := range pkt {
				ret.incoming <- &sm
			}
		}
	}

	notify := func() bool {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("panic in signaling thread %v", err)
			}
		}()

		res := <-ret.incoming
		if res == nil {
			return true
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

		return false
	}

	go func() {
		for !ret.done {
			time.Sleep(300 * time.Millisecond)
			exchange()
		}

		ret.iceChan <- nil
		ret.sdpChan <- nil
		ret.incoming <- nil
	}()

	go func() {
		finish := false
		for !finish {
			finish = notify()
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
		time.Sleep(time.Millisecond * 100)
		if client.connected {
			return
		}
	}
}

func (client *WebsocketClient) WaitForEnd() {
	for {
		time.Sleep(time.Millisecond * 100)
		if client.done {
			return
		}
	}
}

func (client *WebsocketClient) Stop() {
	client.connected = false
	client.done = true
}
