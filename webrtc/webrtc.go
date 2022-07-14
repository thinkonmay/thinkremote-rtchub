package webrtc

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/pigeatgarlic/webrtc-proxy/broadcaster"
	"github.com/pigeatgarlic/webrtc-proxy/listener"
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
	"github.com/pion/rtcp"
	webrtc "github.com/pion/webrtc/v3"
)

type OnTrackFunc func(*webrtc.TrackRemote) (broadcaster.Broadcaster, error)

type WebRTCClient struct {
	conn *webrtc.PeerConnection

	mediaTracks  []*webrtc.TrackLocalStaticRTP

	dataChannels map[string]*webrtc.DataChannel

	fromSdpChannel chan (*webrtc.SessionDescription)
	fromIceChannel chan (*webrtc.ICECandidateInit)

	toSdpChannel chan (*webrtc.SessionDescription)
	toIceChannel chan (*webrtc.ICECandidateInit)

	onTrack   OnTrackFunc
	connected chan bool
}

func InitWebRtcClient(track OnTrackFunc, conf config.WebRTCConfig) (client *WebRTCClient, err error) {
	client = &WebRTCClient{}
	client.toSdpChannel = make(chan *webrtc.SessionDescription)
	client.fromSdpChannel = make(chan *webrtc.SessionDescription)
	client.toIceChannel = make(chan *webrtc.ICECandidateInit)
	client.fromIceChannel = make(chan *webrtc.ICECandidateInit)

	client.dataChannels = make(map[string]*webrtc.DataChannel)
	client.connected = make(chan bool)

	client.mediaTracks = make([]*webrtc.TrackLocalStaticRTP, 0)

	client.onTrack = track
	client.conn, err = webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: conf.Ices,
	})
	if err != nil {
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
			panic(err)
		}
		client.toSdpChannel <- &offer
	})
	client.conn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateFailed {
			if closeErr := client.conn.Close(); closeErr != nil {
				panic(closeErr)
			}
		} else if connectionState == webrtc.ICEConnectionStateConnected {
			client.connected <- true
		}
	})
	client.conn.OnICEGatheringStateChange(func(is webrtc.ICEGathererState) {
		fmt.Printf("%s\n", is.String())
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
			panic(err)
		}

		fmt.Printf("new track: %s\n\n", br.ReadConfig().Name)
		go writeLoop(br, track)
	})

	go func() {
		for {
			var err error
			sdp := <-client.fromSdpChannel

			if sdp.Type == webrtc.SDPTypeAnswer { // answer
				err = client.conn.SetRemoteDescription(*sdp)
				if err != nil {
					panic(err)
				}
			} else { // offer
				err = client.conn.SetRemoteDescription(*sdp)
				if err != nil {
					panic(err)
				}
				ans, err := client.conn.CreateAnswer(&webrtc.AnswerOptions{})
				if err != nil {
					panic(err)
				}
				err = client.conn.SetLocalDescription(ans)
				if err != nil {
					panic(err)
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

func (client *WebRTCClient) ListenRTP(listeners []listener.Listener) {

	for _, lis := range listeners {
		listenerConfig := lis.ReadConfig()

		track, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
			MimeType: listenerConfig.Codec,
		}, listenerConfig.Type, listenerConfig.Name)
		if err != nil {
			panic(err)
		}

		lis.Open()
		go readLoop(lis, track)

		fmt.Printf("added track\n")
		rtpSender, err := client.conn.AddTrack(track)
		if err != nil {
			panic(err)
		}
		// Read incoming RTCP packets
		// Before these packets are returned they are processed by interceptors. For things
		// like NACK this needs to be called.
		go func() {
			for {
				if packets, _, rtcpErr := rtpSender.ReadRTCP(); rtcpErr != nil {
					for _, pkg := range packets {
						dat, _ := pkg.Marshal()
						fmt.Printf("%s\n", string(dat))
					}
					return
				}
			}
		}()

		client.mediaTracks = append(client.mediaTracks, track)
	}

}

func (client *WebRTCClient) RegisterDataChannel(chans map[string]*config.DataChannelConfig) {
	chanMutx := &sync.Mutex{}
	confMutx := &sync.Mutex{}


	client.conn.OnDataChannel(func(channel *webrtc.DataChannel) {
		fmt.Printf("new datachannel\n")
		channel.OnOpen(func() {
			chanMutx.Lock()
			client.dataChannels[channel.Label()] = channel
			chanMutx.Unlock()

			confMutx.Lock()
			conf := chans[channel.Label()];
			confMutx.Unlock()

			channel.OnMessage(func(msg webrtc.DataChannelMessage) {
				conf.Recv <- string(msg.Data)
			})
			channel.OnClose(func() {
				chanMutx.Lock()
				delete(client.dataChannels, channel.Label())
				chanMutx.Unlock()
			})
			go func() {
				for {
					msg := <-conf.Send
					channel.SendText(msg);				
				}
			}()
		})
	})

	confMutx.Lock()
	for Name, channelconf := range chans {
		chanMutx.Lock()
		if client.dataChannels[Name] != nil {
			chanMutx.Unlock()
			continue;
		}
		chanMutx.Unlock()

		channel, err := client.conn.CreateDataChannel(Name, nil)
		if err != nil {
			fmt.Printf("unable to add data channel %s: %s", Name, err.Error())
			continue
		}
		channel.OnOpen(func() {
			chanMutx.Lock()
			client.dataChannels[Name] = channel
			chanMutx.Unlock()

			channel.OnMessage(func(msg webrtc.DataChannelMessage) {
				channelconf.Recv <- string(msg.Data)
			})
			channel.OnClose(func() {
				chanMutx.Lock()
				delete(client.dataChannels, Name)
				chanMutx.Unlock()
			})
			go func() {
				for {
					msg := <-channelconf.Send
					channel.SendText(msg);				
				}
			}()
		})
	}
	confMutx.Unlock()
}

func readLoop(listener listener.Listener, track *webrtc.TrackLocalStaticRTP) {
	for {
		pk := listener.Read()
		if err := track.WriteRTP(pk); err != nil {
			if errors.Is(err, io.ErrClosedPipe) {
				// The peerConnection has been closed.
				return
			}

			panic(err)
		}
	}
}

func writeLoop(br broadcaster.Broadcaster, track *webrtc.TrackRemote) {
	for {
		packet, _, err := track.ReadRTP()
		if err != nil {
			fmt.Printf("%v", err)
		}
		br.Write(packet);
	}
}

func (client *WebRTCClient) WaitConnected() {
	<-client.connected
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
