package main

import (
	"fmt"
	"time"

	proxy "github.com/pigeatgarlic/webrtc-proxy"
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
	"github.com/pion/webrtc/v3"
)

func main() {
	// grpc := config.GrpcConfig{
	// 	Port:          30000,
	// 	ServerAddress: "grpc.signaling.thinkmay.net",
	// 	Token:         "server",
	// }
	grpc := config.GrpcConfig{
		Port:          8000,
		ServerAddress: "localhost",
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
			Name:      "cpuGstreamer",
			Codec:     webrtc.MimeTypeH264,
		},
	}

	chans := config.DataChannelConfig{
		Offer: true,
		Confs: map[string]*struct {
			Send chan string
			Recv chan string
			Channel *webrtc.DataChannel
		}{
			"test": {
				Send: make(chan string),
				Recv: make(chan string),
				Channel: nil,
			},
		},
	}

	go func() {
		// for {
		// 		channel.Send <- "test"
		// }
	}()
	go func() {
		for {
			channel := chans.Confs["test"]
			if channel != nil {
				str := <-chans.Confs["test"].Recv
				fmt.Printf("%s\n", str)
				ParseHIDInput(str);
			} else {
				return;
			}
		}
	}()

	for {
		prox, err := proxy.InitWebRTCProxy(nil, &grpc, &rtc, br, &chans, lis)
		if err != nil {
			fmt.Printf("failed to init webrtc proxy, try again in 2 second\n")
			time.Sleep(2*time.Second);
			continue;
		}
		<-prox.Shutdown
	}
}
