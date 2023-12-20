package websocket

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/signalling"
	"github.com/thinkonmay/thinkremote-rtchub/signalling/gRPC/packet"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
)

type WebsocketClient struct {
	conn *websocket.Conn
	mut  *sync.Mutex

	sdpChan chan *webrtc.SessionDescription
	iceChan chan *webrtc.ICECandidateInit

	done      bool
	connected bool
}

func InitWebsocketClient(AddressStr string,
	auth *config.AuthConfig,
) (_ signalling.Signalling, err error) {
	ret := &WebsocketClient{
		sdpChan: make(chan *webrtc.SessionDescription, 2),
		iceChan: make(chan *webrtc.ICECandidateInit, 2),

		connected: false,
		done:      false,
		mut:       &sync.Mutex{},
	}

	dialer := websocket.Dialer{HandshakeTimeout: 3 * time.Second}
	dial_ctx, _ := context.WithTimeout(context.TODO(), 3*time.Second)
	ret.conn, _, err = dialer.DialContext(dial_ctx, fmt.Sprintf("%s?token=%s", AddressStr, auth.Token), nil)
	if err != nil {
		fmt.Printf("signaling websocket error: %s", err.Error())
		return nil, err
	}

	go func() { for { time.Sleep(time.Second)
			if !ret.done {
				ret.mut.Lock()
				ret.conn.WriteMessage(websocket.TextMessage, []byte("ping"))
				ret.mut.Unlock()
			} else {
				ret.iceChan <- nil
				ret.sdpChan <- nil
				ret.mut.Lock()
				ret.conn.Close()
				ret.mut.Unlock()
				return
			}
		}
	}()

	res := &packet.SignalingMessage{}
	go func() { for { err := ret.conn.ReadJSON(res)
			if ret.done {
				return
			} else if err != nil {
				fmt.Printf("%s\n", err.Error())
				fmt.Printf("websocket connection terminated\n")
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

	client.mut.Lock()
	defer client.mut.Unlock()
	fmt.Printf("SDP send %s\n", req.Sdp.Type)
	if err := client.conn.WriteJSON(&req); err != nil {
		return err
	}
	return nil
}

func (client *WebsocketClient) SendICE(ice *webrtc.ICECandidateInit) error {
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

	client.mut.Lock()
	defer client.mut.Unlock()
	fmt.Printf("ICE sent %v\n", req.Ice)
	if err := client.conn.WriteJSON(req); err != nil {
		return err
	}
	return nil
}

func (client *WebsocketClient) OnICE(fun signalling.OnIceFunc) {
	go func() { for { ice := <-client.iceChan
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
	go func() { for { sdp := <-client.sdpChan
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
	for { time.Sleep(time.Second)
		if client.connected { 
			return 
		}
	}
}

func (client *WebsocketClient) WaitForEnd() {
	for { time.Sleep(time.Second)
		if client.done {
			return
		}
	}
}

func (client *WebsocketClient) Stop() {
	client.connected = false
	client.done = true
}
