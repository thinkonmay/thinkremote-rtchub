package main

import (
	"github.com/pigeatgarlic/webrtc-proxy/cmd/signalling/protocol"
	signalling "github.com/pigeatgarlic/webrtc-proxy/cmd/signalling/signaling"
)

func main() {
	shutdown := make(chan bool)
	signalling.InitSignallingServer(&protocol.SignalingConfig{
		WebsocketPort: 8088,
		GrpcPort:      8000,
	})
	shutdown <- true
}