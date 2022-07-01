package webrtc

import(
 	webrtc "github.com/pion/webrtc/v3"
)

type WebRTCClient struct {
	config webrtc.Configuration
	conn *webrtc.PeerConnection

	iceServers []string
	turnServers string

	mediaTracks []*webrtc.TrackLocalStaticRTP
	dataChannels []*webrtc.DataChannel

	sdpChannel chan(webrtc.SessionDescription)
	iceChannel chan(webrtc.ICECandidate)
}