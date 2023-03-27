package webrtc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	webrtc "github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/listener"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
)



type OnTrackFunc func(*webrtc.TrackRemote)

type WebRTCClient struct {
	conn   *webrtc.PeerConnection
	Closed bool

	onTrack OnTrackFunc

	fromSdpChannel chan (*webrtc.SessionDescription)
	fromIceChannel chan (*webrtc.ICECandidateInit)

	toSdpChannel chan (*webrtc.SessionDescription)
	toIceChannel chan (*webrtc.ICECandidateInit)

	connectionState chan webrtc.ICEConnectionState
	gatherState     chan webrtc.ICEGathererState

	reportChan chan webrtc.StatsReport

	chans []string
}

func InitWebRtcClient(track OnTrackFunc, conf config.WebRTCConfig) (client *WebRTCClient, err error) {
	client = &WebRTCClient{
		toSdpChannel:    make(chan *webrtc.SessionDescription),
		fromSdpChannel:  make(chan *webrtc.SessionDescription),
		toIceChannel:    make(chan *webrtc.ICECandidateInit),
		fromIceChannel:  make(chan *webrtc.ICECandidateInit),
		connectionState: make(chan webrtc.ICEConnectionState),
		gatherState:     make(chan webrtc.ICEGathererState),
		reportChan:      make(chan webrtc.StatsReport),
		onTrack:         track,
		Closed:          false,
	}

	if client.conn, err = webrtc.NewPeerConnection(webrtc.Configuration{ICEServers: conf.Ices}); err != nil {
		return
	}

	go func() {
		for {
			time.Sleep(time.Second)
			client.reportChan <- client.conn.GetStats()
		}
	}()

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
		client.connectionState <- connectionState
	})
	client.conn.OnICEGatheringStateChange(func(gatherState webrtc.ICEGathererState) {
		fmt.Printf("Gather state has changed %s\n", gatherState.String())
		client.gatherState <- gatherState
	})

	client.conn.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for range ticker.C {
				if client.Closed {
					return
				}
				rtcpSendErr := client.conn.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
				if rtcpSendErr != nil {
					fmt.Println(rtcpSendErr)
					return
				}
			}
		}()

		fmt.Printf("new track %s\n", track.ID())
		client.onTrack(track)
	})

	go func() {
		var err error
		for {
			sdp := <-client.fromSdpChannel

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

	go func() {
		for {
			ice := <-client.fromIceChannel
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

func (client *WebRTCClient) handleRTCP(rtpSender *webrtc.RTPSender) {
	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		for {
			if client.Closed {
				fmt.Println("closing handleRTCP thread")
				return
			}

			if packets, _, rtcpErr := rtpSender.ReadRTCP(); rtcpErr != nil {
				for _, pkg := range packets {
					dat, _ := pkg.Marshal()
					fmt.Printf("%s\n", string(dat))
				}
				return
			}
		}
	}()
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
		rtpSender, err := client.conn.AddTrack(track)
		if err != nil {
			fmt.Printf("error add track %s\n", err.Error())
			continue
		}

		client.readLoopRTP(lis, track)
		client.handleRTCP(rtpSender)
	}
}

func (client *WebRTCClient) RegisterDataChannels(chans datachannel.IDatachannel) {
	for _,group := range chans.Groups() {
		fmt.Printf("new datachannel %s\n",group)
		client.RegisterDataChannel(chans,group)
	}
}

func (client *WebRTCClient) RegisterDataChannel(dc datachannel.IDatachannel,group string) {
	rand := fmt.Sprintf("%d", time.Now().UnixNano())

	channel, err := client.conn.CreateDataChannel(group, nil)
	if err != nil {
		fmt.Printf("unable to add data channel: %s\n", err.Error())
		return
	}

	go func() {
		for {
			if client.Closed {
				dc.DeregisterHandle(group,rand)
				return
			}
			time.Sleep(time.Second)
		}
	}()


	dc.RegisterHandle(group,rand,func(pkt string) {
		channel.SendText(pkt)
	})

	channel.OnOpen(func() { channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			if group == "adaptive" {
				var raw map[string]interface{}
				err := json.Unmarshal(msg.Data,&raw)
				if err != nil {
					return 
				}

				raw["__source__"] = rand 
				bytes,_ := json.Marshal(raw)
				dc.Send(group,string(bytes))
				return
			}

			dc.Send(group,string(msg.Data))
		})
	})
}

func (webrtc *WebRTCClient) readLoopRTP(listener listener.Listener, track *webrtc.TrackLocalStaticRTP) {
	id := track.ID()

	listener.RegisterRTPHandler(id, func(pk *rtp.Packet) {
		if track == nil {
			return
		}
		if err := track.WriteRTP(pk); err != nil {
			if errors.Is(err, io.ErrClosedPipe) {
				fmt.Printf("The peerConnection has been closed.")
				return
			}
			fmt.Printf("fail to write sample%s\n", err.Error())
			return
		}
	})

	go func() {
		for {
			time.Sleep(time.Second)
			if !webrtc.Closed {
				continue
			}

			listener.DeregisterRTPHandler(id)
			return
		}
	}()
}

func (webrtc *WebRTCClient) Close() {
	webrtc.conn.Close()
	webrtc.Closed = true
}

func (client *WebRTCClient) GatherStateChange() webrtc.ICEGathererState {
	return <-client.gatherState
}
func (client *WebRTCClient) ConnectionStateChange() webrtc.ICEConnectionState {
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

func (webrtc *WebRTCClient) OnStats() webrtc.StatsReport {
	return <-webrtc.reportChan
}
