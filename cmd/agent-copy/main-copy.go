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
		Token: "client",		
	}
	rtc := config.WebRTCConfig{
		Ices: []webrtc.ICEServer{webrtc.ICEServer{
			URLs: []string{
				"stun:stun.l.google.com:19302",
				"stun:workstation.thinkmay.net:3478",
			},
		}},
	}
	br := []*config.BroadcasterConfig{
		&config.BroadcasterConfig{
			Port: 5000,
			Protocol: "udp",
			BufferSize: 100000,

			Type: "video",
			Name: "rtp",
			Codec: webrtc.MimeTypeH264,
		},
	}
	lis := []*config.ListenerConfig{
	};

	chans := config.DataChannelConfig {
		Offer: false,
		Confs : map[string]*struct{Send chan string; Recv chan string}{
			"test" : &struct{Send chan string; Recv chan string}{
				Send: make(chan string),
				Recv: make(chan string),
			},
		},
	}
	

	go func() {
		for {
			time.Sleep(1 * time.Second);
			chans.Confs["test"].Send <-"test";
		}	
	}()
	go func() {
		for {
			str := <-chans.Confs["test"].Recv
			fmt.Printf("%s\n",str);
		}	
	}()


	_,err := proxy.InitWebRTCProxy(nil,&grpc,&rtc,br,&chans,lis);
	if err != nil {
		panic(err);
	}
	shut := make(chan bool)
	<- shut
}
