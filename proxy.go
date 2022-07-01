package proxy

import (
	datachannel "github.com/pigeatgarlic/webrtc-proxy/data-channel"
	"github.com/pigeatgarlic/webrtc-proxy/listener"
	"github.com/pigeatgarlic/webrtc-proxy/signalling"
	"github.com/pigeatgarlic/webrtc-proxy/webrtc"
)

type Proxy struct {
	listeners []listener.Listener
	datachannels []datachannel.Datachannel
	signallingClient signalling.Signalling
	webrtcClient *webrtc.WebRTCClient
}