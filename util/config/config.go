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
	Audio struct{
		GrpcPort int `json:"GrpcPort"`
	}`json:"Audio"`
	Video struct {
		GrpcPort int `json:"GrpcPort"`
	}`json:"Video"`
	Data struct {
		GrpcPort int `json:"GrpcPort"`
	}`json:"Data"`

	ServerAddress string			`json:"HostName"`
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