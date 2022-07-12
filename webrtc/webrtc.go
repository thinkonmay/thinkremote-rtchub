package webrtc

import (
	"errors"
	"fmt"
	"io"

	"github.com/pigeatgarlic/webrtc-proxy/listener"
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
	webrtc "github.com/pion/webrtc/v3"
)

type WebRTCClient struct {
	conn *webrtc.PeerConnection


	mediaTracks []*webrtc.TrackLocalStaticRTP
	dataChannels []*webrtc.DataChannel

	fromSdpChannel chan(*webrtc.SessionDescription)
	fromIceChannel chan(*webrtc.ICECandidateInit)

	toSdpChannel chan(*webrtc.SessionDescription)
	toIceChannel chan(*webrtc.ICECandidateInit)
}


func InitWebRtcClient(conf config.WebRTCConfig) (client *WebRTCClient, err error) {
	client.toSdpChannel 	= make(chan *webrtc.SessionDescription)
	client.fromSdpChannel 	= make(chan *webrtc.SessionDescription)
	client.toIceChannel 	= make(chan *webrtc.ICECandidateInit)
	client.fromIceChannel 	= make(chan *webrtc.ICECandidateInit)

	client.conn,err = webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: conf.Ices,
	})
	if err != nil {
		return;		
	}



	client.conn.OnICECandidate(func (ice *webrtc.ICECandidate)  {
		init := ice.ToJSON();
		client.toIceChannel <- &init;
	})

	client.conn.OnNegotiationNeeded(func() {
		sdp,err :=client.conn.CreateOffer(nil);
		if err != nil {
			panic(err);
		}
		client.toSdpChannel <- &sdp;
	})
	client.conn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String());
		if connectionState == webrtc.ICEConnectionStateFailed {
			if closeErr := client.conn.Close(); closeErr != nil {
				panic(closeErr)
			}
		}
	})

	go func() {
		for {
			var err error;
			var ans webrtc.SessionDescription;
			sdp := <- client.fromSdpChannel;
			err = client.conn.SetRemoteDescription(*sdp);
			if err != nil {
				panic(err);
			}
			ans, err = client.conn.CreateAnswer(nil);
			if err != nil {
				panic(err);
			}
			err = client.conn.SetLocalDescription(ans);
			if err != nil {
				panic(err);
			}
		}
	}()

	go func() {
		for {
			var err error;
			ice := <- client.fromIceChannel;
			err = client.conn.AddICECandidate(*ice);
			if err != nil {
				panic(err);
			}
		}
	}()

	return;
}

func (client *WebRTCClient)	ListenRTP(listeners []listener.Listener) {
	var err error;
	for _,lis := range listeners{
		listenerConfig := lis.ReadConfig();
		var track *webrtc.TrackLocalStaticRTP;
		track, err = webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{
				MimeType: listenerConfig.Codec,
			}, listenerConfig.Type, listenerConfig.Name);
		client.conn.AddTrack(track);
		client.mediaTracks = append(client.mediaTracks, track);
	}

	for index,track := range client.mediaTracks {
		go func() {
			n,data := listeners[index].Read()
			if _, err = track.Write(data[:n]); err != nil {
				if errors.Is(err, io.ErrClosedPipe) {
					// The peerConnection has been closed.
					return;
				}
				panic(err);
			}
		}()
	}
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
