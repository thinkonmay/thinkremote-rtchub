package main

import (
	proxy "github.com/pigeatgarlic/webrtc-proxy"
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
	"github.com/pion/webrtc/v3"
)

func main() {
	grpc := config.GrpcConfig{
		Port:          8000,
		ServerAddress: "localhost",
	}
	rtc := config.WebRTCConfig{
		Ices: []webrtc.ICEServer{webrtc.ICEServer{
			URLs: []string{"stun:stun.l.google.com:19302"},
		}},
	}
	lis := []*config.ListenerConfig{
		&config.ListenerConfig{
			Port: 5004,
			Protocol: "udp",
			BufferSize: 1028,

			Type: "video",
			Name: "rtp",
			Codec: webrtc.MimeTypeH264,
		},
	}

	shutdown := make(chan bool);
	prox ,err := proxy.InitWebRTCProxy(nil,&grpc,&rtc,lis);
	if err != nil {
		panic(err);
	}
	prox.Start();
	<-shutdown;
}
