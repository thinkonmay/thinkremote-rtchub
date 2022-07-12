package config

import "github.com/pion/webrtc/v3"

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
}


type ListenerConfig struct {
	Port       int
	Protocol   string
	BufferSize int

	Type  string
	Name  string
	Codec string
}

type BroadcasterConfig struct {
	Port       int
	Protocol   string
	BufferSize int

	Type  string
	Name  string
	Codec string
}