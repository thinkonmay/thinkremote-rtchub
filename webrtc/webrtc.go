package webrtc

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/OnePlay-Internet/webrtc-proxy/broadcaster"
	"github.com/OnePlay-Internet/webrtc-proxy/listener"
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/rtcp"
	webrtc "github.com/pion/webrtc/v3"
)

type OnTrackFunc func(*webrtc.TrackRemote) (broadcaster.Broadcaster, error)

type WebRTCClient struct {
	conn *webrtc.PeerConnection

	onTrack     OnTrackFunc
	mediaTracks []webrtc.TrackLocal

	fromSdpChannel chan (*webrtc.SessionDescription)
	fromIceChannel chan (*webrtc.ICECandidateInit)

	toSdpChannel chan (*webrtc.SessionDescription)
	toIceChannel chan (*webrtc.ICECandidateInit)

	connectionState chan webrtc.ICEConnectionState
	gatherState     chan webrtc.ICEGathererState
}

func InitWebRtcClient(track OnTrackFunc, conf config.WebRTCConfig) (client *WebRTCClient, err error) {
	client = &WebRTCClient{}
	client.toSdpChannel = make(chan *webrtc.SessionDescription)
	client.fromSdpChannel = make(chan *webrtc.SessionDescription)
	client.toIceChannel = make(chan *webrtc.ICECandidateInit)
	client.fromIceChannel = make(chan *webrtc.ICECandidateInit)
	client.connectionState = make(chan webrtc.ICEConnectionState)
	client.gatherState = make(chan webrtc.ICEGathererState)
	client.mediaTracks = make([]webrtc.TrackLocal, 0)

	client.onTrack = track
	client.conn, err = webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: conf.Ices,
	})
	if err != nil {
		return
	}

	// TODO
	// go func() {
	// 	for {
	// 		report := client.conn.GetStats()
	// 		out,_ := json.Marshal(report)
	// 		time.Sleep(time.Second);
	// 	}
	// }()

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
			panic(err)
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
				rtcpSendErr := client.conn.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
				if rtcpSendErr != nil {
					fmt.Println(rtcpSendErr)
				}
			}
		}()

		br, err := client.onTrack(track)
		if err != nil {
			fmt.Printf("unable to handle track: %s\n", err.Error())
			return
		}

		fmt.Printf("new track %s\n", track.Codec().MimeType)
		go writeLoop(br, track)
	})

	go func() {
		for {
			var err error
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
			var err error
			ice := <-client.fromIceChannel
			sdp := client.conn.RemoteDescription()
			pending := client.conn.PendingRemoteDescription()
			if sdp == pending {
				return
			}
			err = client.conn.AddICECandidate(*ice)
			if err != nil {
				panic(err)
			}
		}
	}()

	return
}

func handleRTCP(rtpSender *webrtc.RTPSender) {
	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	for {
		if packets, _, rtcpErr := rtpSender.ReadRTCP(); rtcpErr != nil {
			for _, pkg := range packets {
				dat, _ := pkg.Marshal()
				fmt.Printf("%s\n", string(dat))
			}
			return
		}
	}
}

func (client *WebRTCClient) Listen(listeners []listener.Listener) {
	for _, lis := range listeners {
		fmt.Printf("added track\n")

		listenerConfig := lis.GetConfig()
		track, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
			MimeType: listenerConfig.Codec,
		}, lis.GetSourceName(), listenerConfig.StreamID)
		rtpSender, err := client.conn.AddTrack(track)
		if err != nil {
			fmt.Printf("error add track %s\n", err.Error())
			continue
		}

		lis.Open()
		go readLoopRTP(lis, track)
		go handleRTCP(rtpSender)
		client.mediaTracks = append(client.mediaTracks, track)
	}
}

func ondataChannel(channel *webrtc.DataChannel, chans *config.DataChannelConfig) {
	chans.Mutext.Lock()
	conf := chans.Confs[channel.Label()]
	conf.Channel = channel
	chans.Mutext.Unlock()

	channel.OnMessage(func(msg webrtc.DataChannelMessage) {
		conf.Recv <- string(msg.Data)
	})
	channel.OnClose(func() {
		chans.Mutext.Lock()
		delete(chans.Confs, channel.Label())
		chans.Mutext.Unlock()
	})
	go func() {
		for {
			msg := <-conf.Send
			channel.SendText(msg)
		}
	}()
}

func (client *WebRTCClient) RegisterDataChannel(chans *config.DataChannelConfig) {
	chans.Mutext = &sync.Mutex{}

	for Name, _ := range chans.Confs {
		fmt.Printf("new datachannel\n")
		channel, err := client.conn.CreateDataChannel(Name, nil)
		if err != nil {
			fmt.Printf("unable to add data channel %s: %s", Name, err.Error())
			continue
		}
		channel.OnOpen(func() { ondataChannel(channel, chans) })
	}
}

func readLoopRTP(listener listener.Listener, track *webrtc.TrackLocalStaticRTP) {
	for {
		pk := listener.ReadRTP()
		if err := track.WriteRTP(pk); err != nil {
			if errors.Is(err, io.ErrClosedPipe) {
				fmt.Printf("The peerConnection has been closed.")
				return
			}
			fmt.Printf("fail to write sample%s\n", err.Error())
			return
		}
	}
}

func writeLoop(br broadcaster.Broadcaster, track *webrtc.TrackRemote) {
	for {
		packet, _, err := track.ReadRTP()
		if err != nil {
			fmt.Printf("%v", err)
		}
		br.Write(packet)
	}
}

func (client *WebRTCClient) GatherStateChange() webrtc.ICEGathererState {
	return <-client.gatherState
}
func (client *WebRTCClient) ConnectionStateChange() webrtc.ICEConnectionState {
	return <-client.connectionState
}

func (client *WebRTCClient) Close() {
	err := client.conn.Close()
	if err != nil {

	}
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
