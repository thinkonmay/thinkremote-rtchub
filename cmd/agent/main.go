package main

import (
	"fmt"
	"time"

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
	br := []*config.BroadcasterConfig{
		&config.BroadcasterConfig{
			Port: 5001,
			Protocol: "udp",
			BufferSize: 10000,

			Type: "video",
			Name: "rtp2",
			Codec: webrtc.MimeTypeH264,
		},
	}
	lis := []*config.ListenerConfig{
		&config.ListenerConfig{
			Port: 6000,
			Protocol: "udp",
			BufferSize: 10000,

			Type: "video",
			Name: "rtp2",
			Codec: webrtc.MimeTypeH264,
		},
	}

	chans := map[string]* config.DataChannelConfig {
		"test": &config.DataChannelConfig{
			Recv: make(chan string),
			Send: make(chan string),
		},
	}

	go func() {
		for {
			time.Sleep(1 * time.Second);
			chans["test"].Send <-"test";
		}	
	}()
	go func() {
		for {
			str := <-chans["test"].Recv
			fmt.Sprintf("%s\n",str);
		}	
	}()

	_,err := proxy.InitWebRTCProxy(nil,&grpc,&rtc,br,chans,lis);
	if err != nil {
		panic(err);
	}
	shut := make(chan bool)
	<- shut
}
