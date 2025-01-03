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
	"github.com/thinkonmay/thinkremote-rtchub/util/thread"
)

type WebsocketClient struct {
	sdpChan chan interface{}
	iceChan chan interface{}

	incoming, outcoming chan interface{}

	done      bool
	connected bool
	stop      chan bool
}

func InitHttpClient(AddressStr string) (_ signalling.Signalling, err error) {
	client := &WebsocketClient{
		sdpChan: make(chan interface{}, 8),
		iceChan: make(chan interface{}, 8),

		incoming:  make(chan interface{}, 8),
		outcoming: make(chan interface{}, 8),

		connected: false,
		done:      false,
		stop:      make(chan bool, 2),
	}

	u, err := url.Parse(AddressStr)
	if err != nil {
		return nil, err
	} else {
		q := u.Query()
		q.Add("uniqueid", uuid.New().String())
		u.RawQuery = q.Encode()
	}

	thread.SafeSelect(client.stop, client.incoming, func(_res interface{}) {
		res := _res.(*packet.SignalingMessage)
		switch res.Type {
		case packet.SignalingType_tSDP:
			client.sdpChan <- &webrtc.SessionDescription{
				SDP:  res.Sdp.SDPData,
				Type: webrtc.NewSDPType(res.Sdp.Type),
			}
		case packet.SignalingType_tICE:
			LineIndex := uint16(res.Ice.SDPMLineIndex)
			SDPMid := res.Ice.SDPMid
			client.iceChan <- &webrtc.ICECandidateInit{
				Candidate:     res.Ice.Candidate,
				SDPMid:        &SDPMid,
				SDPMLineIndex: &LineIndex,
			}
		case packet.SignalingType_tSTART:
			client.connected = true
		case packet.SignalingType_tEND:
			client.Stop()
		default:
			fmt.Println("Unknown packet")
		}
	})

	thread.SafeLoop(client.stop, time.Millisecond*300, func() {
		pkt := []*packet.SignalingMessage{}
		for len(client.outcoming) > 0 {
			pkt = append(pkt, (<-client.outcoming).(*packet.SignalingMessage))
		}

		if b, err := json.Marshal(pkt); err != nil {
		} else if resp, err := http.DefaultClient.Post(
			u.String(), "application/json", strings.NewReader(string(b)),
		); err != nil {
		} else if b, err := io.ReadAll(resp.Body); err != nil {
		} else if err = json.Unmarshal(b, &pkt); err != nil {
		} else {
			for _, sm := range pkt {
				client.incoming <- sm
			}
		}
	})

	client.WaitForEnd(func() {
		thread.TriggerStop(client.stop)
	})

	return client, nil
}

func (client *WebsocketClient) SendSDP(desc *webrtc.SessionDescription) {
	thread.SafeWait(func() bool {
		return client.connected
	}, func() {
		client.outcoming <- &packet.SignalingMessage{
			Type: packet.SignalingType_tSDP,
			Sdp: &packet.SDP{
				Type:    desc.Type.String(),
				SDPData: desc.SDP,
			},
		}
	})
}

func (client *WebsocketClient) SendICE(ice *webrtc.ICECandidateInit) {
	thread.SafeWait(func() bool {
		return client.connected
	}, func() {
		client.outcoming <- &packet.SignalingMessage{
			Type: packet.SignalingType_tICE,
			Ice: &packet.ICE{
				SDPMid:        *ice.SDPMid,
				SDPMLineIndex: int64(*ice.SDPMLineIndex),
				Candidate:     ice.Candidate,
			},
		}
	})
}

func (client *WebsocketClient) OnICE(fun signalling.OnIceFunc) {
	thread.SafeSelect(client.stop, client.iceChan, func(ice interface{}) {
		fun(ice.(*webrtc.ICECandidateInit))
	})
}

func (client *WebsocketClient) OnSDP(fun signalling.OnSDPFunc) {
	thread.SafeSelect(client.stop, client.sdpChan, func(sdp interface{}) {
		fun(sdp.(*webrtc.SessionDescription))
	})
}

func (client *WebsocketClient) WaitForStart(fun func()) {
	thread.SafeWait(func() bool { return client.connected }, fun)
}

func (client *WebsocketClient) WaitForEnd(fun func()) {
	thread.SafeWait(func() bool { return client.done }, fun)
}

func (client *WebsocketClient) Stop() {
	client.connected = false
	client.done = true
}
