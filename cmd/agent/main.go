package main

import (
	"fmt"
	"os"

	proxy "github.com/OnePlay-Internet/webrtc-proxy"
	"github.com/OnePlay-Internet/webrtc-proxy/listener"
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/webrtc/v3"
)

func main() {
	var token string

	args := os.Args[1:]
	for i, arg := range args {
		if arg == "--token" {
			token = args[i+1]
		} else if arg == "--help" {
			return
		}
	}

	if token == "" {
		return
	}


	grpc := config.GrpcConfig{
		Port:          30000,
		ServerAddress: "54.169.49.176",
		Token:         token,
	}
	rtc := config.WebRTCConfig{
		Ices: []webrtc.ICEServer{{
			URLs: []string{
				"stun:stun.l.google.com:19302",
			},
		}, {
			URLs: []string{
				"stun:workstation.thinkmay.net:3478",
			},
		},
		},
	}
	br := []*config.BroadcasterConfig{&config.BroadcasterConfig{
		Name: "audio",
		Codec: webrtc.MimeTypeH264,
	}}

	chans := config.DataChannelConfig{
		Offer: true,
		Confs: map[string]*struct {
			Send    chan string
			Recv    chan string
			Channel *webrtc.DataChannel
		}{
			"hid": {
				Send:    make(chan string),
				Recv:    make(chan string),
				Channel: nil,
			},
		},
	}
	Lists := make([]listener.Listener, 0)
	prox, err := proxy.InitWebRTCProxy(nil, &grpc, &rtc, br, &chans, Lists)
	if err != nil {
		fmt.Printf("failed to init webrtc proxy: %s\n",err.Error())
		return
	}
	<-prox.Shutdown
}
