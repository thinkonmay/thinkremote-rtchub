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
		Port:          443,
		ServerAddress: "grpc.signaling.thinkmay.net",
		Token:         "server",
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
	br := []*config.BroadcasterConfig{}
	lis := []*config.ListenerConfig{{
			Source:   "gstreamer",

			DataType:  "sample",

			MediaType: "video",
			Name:      "gstreamer",
			Codec:     webrtc.MimeTypeH264,
		},
	}

	chans := config.DataChannelConfig{
		Offer: true,
		Confs: map[string]*struct {
			Send chan string
			Recv chan string
		}{
			"test": {
				Send: make(chan string),
				Recv: make(chan string),
			},
		},
	}

	go func() {
		for {
			time.Sleep(1 * time.Second)
			chans.Confs["test"].Send <- "test"
		}
	}()
	go func() {
		for {
			str := <-chans.Confs["test"].Recv
			fmt.Printf("%s\n", str)
		}
	}()

	_, err := proxy.InitWebRTCProxy(nil, &grpc, &rtc, br, &chans, lis)
	if err != nil {
		panic(err)
	}
	shut := make(chan bool)
	<-shut
}
