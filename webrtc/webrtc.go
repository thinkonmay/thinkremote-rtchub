package webrtc

import (
	"fmt"
	"time"

	"github.com/pigeatgarlic/webrtc-proxy/broadcaster"
	"github.com/pigeatgarlic/webrtc-proxy/listener"
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
	webrtc "github.com/pion/webrtc/v3"
)

type OnTrackFunc func (*webrtc.TrackRemote)(broadcaster.Broadcaster,error)

type WebRTCClient struct {
	conn *webrtc.PeerConnection


	mediaTracks []*webrtc.TrackLocalStaticSample
	dataChannels []*webrtc.DataChannel

	fromSdpChannel chan(*webrtc.SessionDescription)
	fromIceChannel chan(*webrtc.ICECandidateInit)

	toSdpChannel chan(*webrtc.SessionDescription)
	toIceChannel chan(*webrtc.ICECandidateInit)

	onTrack OnTrackFunc
	connected chan bool
}


func InitWebRtcClient(track OnTrackFunc,conf config.WebRTCConfig) (client *WebRTCClient, err error) {
	client = &WebRTCClient{}
	client.toSdpChannel 	= make(chan *webrtc.SessionDescription)
	client.fromSdpChannel 	= make(chan *webrtc.SessionDescription)
	client.toIceChannel 	= make(chan *webrtc.ICECandidateInit)
	client.fromIceChannel 	= make(chan *webrtc.ICECandidateInit)
	client.connected		= make(chan bool)

	client.mediaTracks 		= make([]*webrtc.TrackLocalStaticSample, 0)

	
	client.onTrack = track;
	client.conn,err = webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: conf.Ices,
	})
	if err != nil {
		return;		
	}



	client.conn.OnICECandidate(func (ice *webrtc.ICECandidate)  {
		if ice == nil {
			fmt.Printf("ice candidate was null\n");
			return;
		}
		init := ice.ToJSON();
		client.toIceChannel <- &init;
	})

	client.conn.OnNegotiationNeeded(func() {
		offer,err :=client.conn.CreateOffer(&webrtc.OfferOptions{
			ICERestart: false,	
		});
		client.conn.SetLocalDescription(offer);
		if err != nil {
			panic(err);
		}
		client.toSdpChannel <- &offer;
	})
	client.conn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String());
		if connectionState == webrtc.ICEConnectionStateFailed {
			if closeErr := client.conn.Close(); closeErr != nil {
				panic(closeErr)
			}
		} else if connectionState == webrtc.ICEConnectionStateConnected {
			client.connected<-true;
		}
	})
	client.conn.OnICEGatheringStateChange(func(is webrtc.ICEGathererState) {
		fmt.Printf("%s\n",is.String());
	})

	client.conn.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		fmt.Printf("new track\n");
		br,err := client.onTrack(tr);
		conf := br.ReadConfig()
		buffer := make([]byte,conf.BufferSize);
		shutdown := make(chan bool);
		go func() {
			defer func ()  {
				br.Close();	
			}();
			for {
				switch {
				case <-shutdown:
					return;
				default:
					var count int;
					count,_,err = r.Read(buffer);
					if err != nil {
						fmt.Printf("%v",err);
						return;
					}
					err = br.Write(count,buffer)
				}
			}
		}()

		br.OnClose(func(lis broadcaster.Broadcaster) {
			shutdown <- true;
		})
	})

	go func() {
		for {
			var err error;
			sdp := <- client.fromSdpChannel;

			if sdp.Type == webrtc.SDPTypeAnswer { // answer
				err = client.conn.SetRemoteDescription(*sdp);
				if err != nil {
					panic(err);
				}
			} else { // offer
				err = client.conn.SetRemoteDescription(*sdp);
				if err != nil {
					panic(err);
				}
				ans, err := client.conn.CreateAnswer(&webrtc.AnswerOptions{});
				if err != nil {
					panic(err);
				}
				err = client.conn.SetLocalDescription(ans)
				if err != nil {
					panic(err)
				}
				client.toSdpChannel <- &ans;
			}
		}
	}()

	go func() {
		for {
			var err error;
			ice := <- client.fromIceChannel;
			sdp := client.conn.RemoteDescription()
			pending := client.conn.PendingRemoteDescription()
			if sdp == pending {
				return;
			}
			err = client.conn.AddICECandidate(*ice);
			if err != nil {
				panic(err);
			}
		}
	}()

	client.conn.OnDataChannel(func(dc *webrtc.DataChannel) {
		fmt.Printf("new datachannel\n");
		dc.OnOpen(func() {
			for {
				dc.SendText("hello from peer\n")
				time.Sleep(time.Second)	
			}
		})
	})


	return;
}

func (client *WebRTCClient)	ListenRTP(listeners []listener.Listener) {
	var err error;

	fmt.Printf("added datachannel\n");
	channel, err := client.conn.CreateDataChannel("data", nil)
	if err != nil {
		panic(err)
	}else {
		channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf(string(msg.Data));
		})
	}


	client.dataChannels = append(client.dataChannels, channel);

	for _,lis := range listeners{
		listenerConfig := lis.ReadConfig();
		var track *webrtc.TrackLocalStaticSample;
		var rtpSender *webrtc.RTPSender;

		track, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
				MimeType: listenerConfig.Codec,
			}, listenerConfig.Type, listenerConfig.Name);
		fmt.Printf("addded track\n");
		rtpSender,err = client.conn.AddTrack(track);
		if err != nil {
			panic(err);
		}
		// Read incoming RTCP packets
		// Before these packets are returned they are processed by interceptors. For things
		// like NACK this needs to be called.
		go func() {
			rtcpBuf := make([]byte, 1500)
			for {
				if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
					return
				}
			}
		}()
		client.mediaTracks = append(client.mediaTracks, track);
	}

	// for index,track := range client.mediaTracks {
	// 	go func() {
	// 		n,data := listeners[index].Read()
	// 		if _, err = track.WriteSample(media.Sample{Data: }); err != nil {
	// 			if errors.Is(err, io.ErrClosedPipe) {
	// 				// The peerConnection has been closed.
	// 				return;
	// 			}
	// 			panic(err);
	// 		}
	// 	}()
	// }
}

func (client *WebRTCClient)	WaitConnected() {
	<-client.connected
}




func (webrtc *WebRTCClient)	OnIncominSDP(sdp *webrtc.SessionDescription) {
	webrtc.fromSdpChannel <- sdp;
}

func (webrtc *WebRTCClient)	OnIncomingICE(ice *webrtc.ICECandidateInit) {
	webrtc.fromIceChannel <- ice;
}

func (webrtc *WebRTCClient)	OnLocalICE()*webrtc.ICECandidateInit {
	return <-webrtc.toIceChannel;	
}

func (webrtc *WebRTCClient)	OnLocalSDP()*webrtc.SessionDescription {
	return <-webrtc.toSdpChannel;	
}
