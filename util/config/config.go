package config

import (
	"github.com/pion/webrtc/v3"
)

type WebRTCConfig struct {
	Ices []webrtc.ICEServer `json:"iceServers"`
}

type WebsocketConfig struct {
	Port          int
	ServerAddress string
}

type AuthConfig struct {
	Token         string		`json:"token"`
}
type GrpcConfig struct {
	Port          int			`json:"SignalingPort"`
	ServerAddress string		`json:"HostName"`
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