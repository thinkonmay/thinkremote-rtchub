package config

import (
	"sync"

	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
	"github.com/pion/webrtc/v3"
)

type WebRTCConfig struct {
	Ices []webrtc.ICEServer
}

type WebsocketConfig struct {
	Port          int
	ServerAddress string
}

type GrpcConfig struct {
	Port          int
	ServerAddress string
	Token         string
}

type ListenerConfig struct {
	Bitrate int
	MediaType string

	VideoSource tool.Monitor
	AudioSource tool.Soundcard

	DataType  string
	Name      string
	Codec     string
}

type BroadcasterConfig struct {
	Name  string
	Codec string
}

type DataChannelConfig struct {
	Offer bool
	Mutext *sync.Mutex
	Confs map[string]*struct {
		Send chan string
		Recv chan string
		Channel *webrtc.DataChannel
	}
}
