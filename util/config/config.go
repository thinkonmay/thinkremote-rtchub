package config

import (
	"sync"

	"github.com/pion/webrtc/v3"
)

type WebRTCConfig struct {
	Ices []webrtc.ICEServer `json:"iceServers"`
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


func NewDataChannelConfig(names []string) *DataChannelConfig {
	conf := DataChannelConfig{
		Mutext: &sync.Mutex{},
		Confs: map[string]*DataChannel{ },
	}

	for _,name := range names {
		conf.Confs[name] = &DataChannel{
			Send:    make(chan string),
			Recv:    make(chan string),
			Channel: nil,
		}
	}

	return &conf;
}