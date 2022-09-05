package main

import (
	"fmt"
	"os"
	"time"

	proxy "github.com/OnePlay-Internet/webrtc-proxy"
	"github.com/OnePlay-Internet/webrtc-proxy/hid"
	"github.com/OnePlay-Internet/webrtc-proxy/listener"
	"github.com/OnePlay-Internet/webrtc-proxy/listener/audio"
	"github.com/OnePlay-Internet/webrtc-proxy/listener/video"
	"github.com/OnePlay-Internet/webrtc-proxy/listener/udp"
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/webrtc/v3"
)

func main() {
	var token string
	URL := "localhost:5000"
	env := "prod"

	args := os.Args[1:]
	for i, arg := range args {
		if arg == "--token" {
			token = args[i+1]
		} else if arg == "--hid" {
			URL = args[i+1]
		} else if arg == "--env" {
			env = args[i+1]
		} else if arg == "--help" {
			return
		}
	}

	if token == "" {
		return
	}

	engine := func () string {
		switch env {
		case "dev":
			return "cpuGstreamer"
		case "prod":
			return "gpuGstreamer"
		default:
			return "gpuGstreamer"
		}	
	}()

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
	br := []*config.BroadcasterConfig{}
	lis := []*config.ListenerConfig{{
			Source: "gstreamer",

			DataType: "sample",

			MediaType: "video",
			Name:      engine,
			Codec:     webrtc.MimeTypeH264,
		},
		{
			Source: "gstreamer",

			DataType: "sample",

			MediaType: "audio",
			Name:      "audioGstreamer",
			Codec:     webrtc.MimeTypeOpus,
		},
	}

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

	sing := hid.NewHIDSingleton(URL);
	go func() {
		for {
			channel := chans.Confs["hid"]
			if channel != nil {
				str := <-chans.Confs["hid"].Recv
				sing.ParseHIDInput(str)
			} else {
				return
			}
		}
	}()

	Lists := make([]listener.Listener, 0)
	for _, lis_conf := range lis {
		var Lis listener.Listener
		if lis_conf.MediaType == "audio" {
			Lis = audio.CreatePipeline(lis_conf)
		} else if lis_conf.Source == "udp" {
			udpLis, err := udp.NewUDPListener(lis_conf)
			Lis = &udpLis
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}
		} else if lis_conf.Source == "gstreamer" {
			Lis = video.CreatePipeline(lis_conf)
		} else {
			fmt.Printf("Unimplemented listener\n")
			return
		}

		Lists = append(Lists, Lis)
	}

	prox, err := proxy.InitWebRTCProxy(nil, &grpc, &rtc, br, &chans, Lists)
	if err != nil {
		fmt.Printf("failed to init webrtc proxy, try again in 2 second\n")
		time.Sleep(2 * time.Second)
		return
	}
	<-prox.Shutdown
}
