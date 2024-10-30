package webrtc

import (
	"fmt"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/listener"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
	"github.com/thinkonmay/thinkremote-rtchub/util/thread"
)

type OnTrackFunc func(*webrtc.TrackRemote)
type OnIDRFunc func()

type WebRTCClient struct {
	conn   *webrtc.PeerConnection
	Closed bool
	stop   chan bool

	onTrack OnTrackFunc
	onIDR   OnIDRFunc

	fromSdpChannel chan webrtc.SessionDescription
	fromIceChannel chan webrtc.ICECandidateInit

	toSdpChannel chan webrtc.SessionDescription
	toIceChannel chan webrtc.ICECandidateInit

	connectionState chan webrtc.ICEConnectionState
	gatherState     chan webrtc.ICEGatheringState
}

func InitWebRtcClient(track OnTrackFunc, idr OnIDRFunc, conf config.WebRTCConfig) (client *WebRTCClient, err error) {
	client = &WebRTCClient{
		stop:            make(chan bool, 2),
		toSdpChannel:    make(chan webrtc.SessionDescription, 2),
		fromSdpChannel:  make(chan webrtc.SessionDescription, 2),
		toIceChannel:    make(chan webrtc.ICECandidateInit, 2),
		fromIceChannel:  make(chan webrtc.ICECandidateInit, 2),
		connectionState: make(chan webrtc.ICEConnectionState, 2),
		gatherState:     make(chan webrtc.ICEGatheringState, 2),
		onTrack:         track,
		onIDR:           idr,
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
		client.toIceChannel <- init
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
		client.toSdpChannel <- offer
	})
	client.conn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection state has changed %s \n", connectionState.String())
		client.connectionState <- connectionState
	})
	client.conn.OnICEGatheringStateChange(func(is webrtc.ICEGatheringState) {
		fmt.Printf("Gather state has changed %s\n", is.String())
		client.gatherState <- is
	})

	client.conn.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Printf("new track %s\n", track.ID())
		client.onTrack(track)
	})

	thread.SafeLoop(client.stop, time.Millisecond*10, func() {
		select {
		case sdp := <-client.fromSdpChannel:
			switch sdp.Type {
			case webrtc.SDPTypeAnswer: // answer
				if err := client.conn.SetRemoteDescription(sdp); err != nil {
				}
			case webrtc.SDPTypeOffer:
				if err := client.conn.SetRemoteDescription(sdp); err != nil {
				} else if ans, err := client.conn.CreateAnswer(&webrtc.AnswerOptions{}); err != nil {
				} else if err = client.conn.SetLocalDescription(ans); err != nil {
				} else {
					client.toSdpChannel <- ans
				}
			}
		case <-client.stop:
			client.stop <- true
		}
	})

	thread.SafeLoop(client.stop, time.Millisecond*10, func() {
		select {
		case ice := <-client.fromIceChannel:
			sdp := client.conn.RemoteDescription()
			pending := client.conn.PendingRemoteDescription()
			if sdp == pending {
			} else if err := client.conn.AddICECandidate(ice); err != nil {
				fmt.Printf("error add ice candicate %s\n", err.Error())
			}
		case <-client.stop:
			client.stop <- true
		}
	})

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

		sender, err := client.conn.AddTrack(track)
		if err != nil {
			fmt.Printf("error add track %s\n", err.Error())
			continue
		}

		client.readLoopRTP(lis, track, sender)
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
	thread.SafeWait(func() bool {
		return client.Closed
	}, func() {
		dc.DeregisterHandle(group, rand)
	})
	channel.OnOpen(func() {
		channel.OnMessage(
			func(msg webrtc.DataChannelMessage) {
				if !client.Closed {
					dc.Send(group, string(msg.Data))
				}
			})
	})
}

func (client *WebRTCClient) readLoopRTP(listener listener.Listener,
	track *webrtc.TrackLocalStaticRTP,
	sender *webrtc.RTPSender) {
	id := track.ID()

	listener.RegisterRTPHandler(id, func(pk *rtp.Packet) {
		track.WriteRTP(pk)
	})

	stop := make(chan bool, 2)
	thread.SafeLoop(stop, time.Millisecond*10, func() {
		if packets, _, err := sender.ReadRTCP(); err == nil {
			for _, pkt := range packets {
				switch pkt.(type) {
				case *rtcp.FullIntraRequest:
					client.onIDR()
				case *rtcp.PictureLossIndication:
					client.onIDR()
				case *rtcp.TransportLayerNack:
				case *rtcp.ReceiverReport:
				case *rtcp.SenderReport:
				case *rtcp.ExtendedReport:
				}
			}
		}
	})

	thread.SafeWait(func() bool {
		return client.Closed
	}, func() {
		stop <- true
		listener.DeregisterRTPHandler(id)
	})
}

func (client *WebRTCClient) Close() {
	client.conn.Close()
	client.Closed = true
	client.gatherState <- webrtc.ICEGatheringState(-1)
}
func (webrtc *WebRTCClient) StopSignaling() {
	fmt.Println("stopping signaling process")
}

func (client *WebRTCClient) GatherStateChange() chan webrtc.ICEGatheringState {
	return client.gatherState
}
func (client *WebRTCClient) ConnectionStateChange() chan webrtc.ICEConnectionState {
	return client.connectionState
}

func (webrtc *WebRTCClient) OnIncominSDP(sdp webrtc.SessionDescription) {
	webrtc.fromSdpChannel <- sdp
}

func (webrtc *WebRTCClient) OnIncomingICE(ice webrtc.ICECandidateInit) {
	webrtc.fromIceChannel <- ice
}

func (webrtc *WebRTCClient) OnLocalICE() chan webrtc.ICECandidateInit {
	return webrtc.toIceChannel
}

func (webrtc *WebRTCClient) OnLocalSDP() chan webrtc.SessionDescription {
	return webrtc.toSdpChannel
}
