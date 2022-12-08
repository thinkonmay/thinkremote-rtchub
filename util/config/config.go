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
	ID		  string
	StreamID  string
	Codec     string
}

type BroadcasterConfig struct {
	Name  string
	Codec string
}

type DataChannel struct {
	Send    chan string
	Recv    chan string
	Channel *webrtc.DataChannel
}
type DataChannelConfig struct {
	Mutext *sync.Mutex
	Confs  map[string]*DataChannel
}
