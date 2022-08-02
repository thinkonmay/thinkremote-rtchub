package config

import (
	"sync"

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
	Source	   string

	DataType  string

	MediaType string
	Name      string
	Codec     string
}

type BroadcasterConfig struct {
	Port       int
	Protocol   string

	Type  string
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
