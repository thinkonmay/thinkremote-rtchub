package webrtc

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/pion/rtp"
	webrtc "github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/listener"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
)

type OnTrackFunc func(*webrtc.TrackRemote)

type WebRTCClient struct {
	conn      *webrtc.PeerConnection
	Closed    bool

	onTrack OnTrackFunc

	fromSdpChannel chan *webrtc.SessionDescription
	fromIceChannel chan *webrtc.ICECandidateInit

	toSdpChannel chan *webrtc.SessionDescription
	toIceChannel chan *webrtc.ICECandidateInit

	connectionState chan *webrtc.ICEConnectionState
	gatherState     chan *webrtc.ICEGathererState
}

func InitWebRtcClient(track OnTrackFunc, conf config.WebRTCConfig) (client *WebRTCClient, err error) {
	client = &WebRTCClient{
		toSdpChannel:    make(chan *webrtc.SessionDescription, 2),
		fromSdpChannel:  make(chan *webrtc.SessionDescription, 2),
		toIceChannel:    make(chan *webrtc.ICECandidateInit, 2),
		fromIceChannel:  make(chan *webrtc.ICECandidateInit, 2),
		connectionState: make(chan *webrtc.ICEConnectionState, 2),
		gatherState:     make(chan *webrtc.ICEGathererState, 2),
		onTrack:         track,
		Closed:          false,
	}

	if client.conn, err = webrtc.NewPeerConnection(webrtc.Configuration{ICEServers: conf.Ices}); err != nil {
		return
	}

	client.conn.OnICECandidate(func(ice *webrtc.ICECandidate) {
		if ice == nil {
			fmt.Printf("ice candidate was null\n")
			return
		}
		init := ice.ToJSON()
		client.toIceChannel <- &init
	})

	client.conn.OnNegotiationNeeded(func() {
		offer, err := client.conn.CreateOffer(&webrtc.OfferOptions{
			ICERestart: false,
		})
		client.conn.SetLocalDescription(offer)
		if err != nil {
			fmt.Printf("error creating offer %s\n", err.Error())
			return
		}
		client.toSdpChannel <- &offer
	})
	client.conn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection state has changed %s \n", connectionState.String())
		client.connectionState <- &connectionState
	})
	client.conn.OnICEGatheringStateChange(func(gatherState webrtc.ICEGathererState) {
		fmt.Printf("Gather state has changed %s\n", gatherState.String())
		client.gatherState <- &gatherState
	})

	client.conn.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Printf("new track %s\n", track.ID())
		client.onTrack(track)
	})

	go func() { for { sdp := <-client.fromSdpChannel
			var err error
			if sdp == nil {
				return
			}

			if sdp.Type == webrtc.SDPTypeAnswer { // answer
				err = client.conn.SetRemoteDescription(*sdp)
				if err != nil {
					fmt.Printf("%s,\n", err.Error())
					continue
				}
			} else { // offer
				err = client.conn.SetRemoteDescription(*sdp)
				if err != nil {
					fmt.Printf("%s,\n", err.Error())
					continue
				}
				ans, err := client.conn.CreateAnswer(&webrtc.AnswerOptions{})
				if err != nil {
					fmt.Printf("%s,\n", err.Error())
					continue
				}
				err = client.conn.SetLocalDescription(ans)
				if err != nil {
					fmt.Printf("%s,\n", err.Error())
					continue
				}
				client.toSdpChannel <- &ans
			}
		}
	}()

	go func() { for { ice := <-client.fromIceChannel
			if ice == nil {
				return
			}
			sdp := client.conn.RemoteDescription()
			pending := client.conn.PendingRemoteDescription()
			if sdp == pending {
				return
			}
			err := client.conn.AddICECandidate(*ice)
			if err != nil {
				fmt.Printf("error add ice candicate %s\n", err.Error())
				continue
			}
		}
	}()

	return
}

func (client *WebRTCClient) Listen(listeners []listener.Listener) {
	for _, lis := range listeners {
		codec := lis.GetCodec()
		track, err := webrtc.NewTrackLocalStaticRTP(
			webrtc.RTPCodecCapability{MimeType: codec},
			fmt.Sprintf("%d", time.Now().UnixNano()),
			fmt.Sprintf("%d", time.Now().UnixNano()))

		if err != nil {
			fmt.Printf("error add track %s\n", err.Error())
			continue
		}
		_, err = client.conn.AddTrack(track)
		if err != nil {
			fmt.Printf("error add track %s\n", err.Error())
			continue
		}

		client.readLoopRTP(lis, track)
	}
}

func (client *WebRTCClient) RegisterDataChannels(chans datachannel.IDatachannel) {
	for _, group := range chans.Groups() {
		fmt.Printf("new datachannel %s\n", group)
		client.RegisterDataChannel(chans, group)
	}
}

func (client *WebRTCClient) RegisterDataChannel(dc datachannel.IDatachannel, group string) {
	channel, err := client.conn.CreateDataChannel(group, nil)
	if err != nil {
		fmt.Printf("unable to add data channel: %s\n", err.Error())
		return
	}

	rand := fmt.Sprintf("%d", time.Now().UnixNano())
	dc.RegisterHandle(group, rand, func(msg string) {
		if client.Closed {
			return
		}
		channel.SendText(msg)
	})
	go func() { for { time.Sleep(time.Second)
			if client.Closed {
				dc.DeregisterHandle(group, rand)
				return
			}
		}
	}()

	channel.OnOpen(func() {
		channel.OnMessage(
			func(msg webrtc.DataChannelMessage) {
				if client.Closed {
					return
				}
				dc.Send(group, rand, string(msg.Data))
			})
	})
}

func (client *WebRTCClient) readLoopRTP(listener listener.Listener, track *webrtc.TrackLocalStaticRTP) {
	id := track.ID()

	listener.RegisterRTPHandler(id, func(pk *rtp.Packet) {
		if err := track.WriteRTP(pk); err != nil {
			if errors.Is(err, io.ErrClosedPipe) {
				fmt.Printf("The peerConnection has been closed.")
				return
			}
			fmt.Printf("fail to write sample%s\n", err.Error())
			return
		}
	})

	go func() { for { time.Sleep(time.Millisecond * 100)
			if client.Closed {
				listener.DeregisterRTPHandler(id)
				return
			}
		}
	}()
}

func (webrtc *WebRTCClient) Close() {
	webrtc.conn.Close()
	webrtc.Closed = true
	webrtc.connectionState <- nil
	webrtc.gatherState <- nil
}
func (webrtc *WebRTCClient) StopSignaling() {
	fmt.Println("stopping signaling process")
	webrtc.toSdpChannel <- nil
	webrtc.fromSdpChannel <- nil
	webrtc.toIceChannel <- nil
	webrtc.fromIceChannel <- nil
}

func (client *WebRTCClient) GatherStateChange() *webrtc.ICEGathererState {
	return <-client.gatherState
}
func (client *WebRTCClient) ConnectionStateChange() *webrtc.ICEConnectionState {
	return <-client.connectionState
}

func (webrtc *WebRTCClient) OnIncominSDP(sdp *webrtc.SessionDescription) {
	webrtc.fromSdpChannel <- sdp
}

func (webrtc *WebRTCClient) OnIncomingICE(ice *webrtc.ICECandidateInit) {
	webrtc.fromIceChannel <- ice
}

func (webrtc *WebRTCClient) OnLocalICE() *webrtc.ICECandidateInit {
	return <-webrtc.toIceChannel
}

func (webrtc *WebRTCClient) OnLocalSDP() *webrtc.SessionDescription {
	return <-webrtc.toSdpChannel
}
