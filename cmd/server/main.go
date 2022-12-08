package main

import (
	"fmt"
	"os"
	"strconv"

	proxy "github.com/OnePlay-Internet/webrtc-proxy"
	"github.com/OnePlay-Internet/webrtc-proxy/hid"
	"github.com/OnePlay-Internet/webrtc-proxy/listener"
	"github.com/OnePlay-Internet/webrtc-proxy/listener/audio"
	"github.com/OnePlay-Internet/webrtc-proxy/listener/video"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"

	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/webrtc/v3"
)

func main() {
	var err error
	var token string
	args := os.Args[1:]
	HIDURL := "localhost:5000"

	signaling := "54.169.49.176"
	Port := 30000

	Stun := "stun:workstation.thinkmay.net:3478"
	Turn := "turn:workstation.thinkmay.net:3478"

	TurnUser := "oneplay"
	TurnPassword := "oneplay"

	qr := tool.GetDevice()
	if len(qr.Monitors) == 0 {
		fmt.Printf("no display available")
		return
	}

	for i, arg := range args {
		if arg == "--token" {
			token = args[i+1]
		} else if arg == "--hid" {
			HIDURL = args[i+1]
		} else if arg == "--grpc" {
			signaling = args[i+1]
		} else if arg == "--grpcport" {
			Port, err = strconv.Atoi(args[i+1])
		} else if arg == "--turn" {
			Turn = args[i+1]
		} else if arg == "--turnuser" {
			TurnUser = args[i+1]
		} else if arg == "--turnpassword" {
			TurnPassword = args[i+1]
		} else if arg == "--device" {
			fmt.Printf("=======================================================================\n")
			fmt.Printf("MONITOR DEVICE\n")
			for index, monitor := range qr.Monitors {
				fmt.Printf("=======================================================================\n")
				fmt.Printf("monitor %d\n", index)
				fmt.Printf("monitor name 			%s\n", monitor.MonitorName)
				fmt.Printf("monitor handle  		%d\n", monitor.MonitorHandle)
				fmt.Printf("monitor adapter 		%s\n", monitor.Adapter)
				fmt.Printf("monitor device  		%s\n", monitor.DeviceName)
				fmt.Printf("=======================================================================\n")
			}
			fmt.Printf("\n\n\n\n")

			fmt.Printf("=======================================================================\n")
			fmt.Printf("AUDIO DEVICE\n")
			for index, audio := range qr.Soundcards {
				fmt.Printf("=======================================================================\n")
				fmt.Printf("audio source 			%d\n", index)
				fmt.Printf("audio source name 		%s\n", audio.Name)
				fmt.Printf("audio source device id  %s\n", audio.DeviceID)
				fmt.Printf("=======================================================================\n")
			}
			fmt.Printf("\n\n\n\n")
		} else if arg == "--help" {
			fmt.Printf("--token 	 	 |  server token\n")
			fmt.Printf("--hid   	 	 |  HID server URL (example: localhost:5000)\n")
			fmt.Printf("--grpcport   	 |  HID server URL (example: localhost:5000)\n")
			fmt.Printf("--stun  	 	 |  TURN server (example: stun:workstation.thinkmay.net:3478 )\n")
			fmt.Printf("--turn  	 	 |  TURN server (example: stun:workstation.thinkmay.net:3478 )\n")
			fmt.Printf("--turncred  	 |  TURN server (example: stun:workstation.thinkmay.net:3478 )\n")
			fmt.Printf("--turnuser  	 |  TURN server (example: stun:workstation.thinkmay.net:3478 )\n")
			fmt.Printf("--signaling  	 |  TURN server (example: (signaling.thinkmay.net or 54.169.49.176 )\n")
			fmt.Printf("--signalingport  |  TURN server (example: stun:workstation.thinkmay.net:3478 )\n")
			return
		}
	}

	if token == "" {
		err = fmt.Errorf("no available token")
	}
	if err != nil {
		fmt.Printf("invalid argument : %s\n", err.Error())
		return
	}

	grpc := config.GrpcConfig{
		Port:          Port,
		ServerAddress: signaling,
		Token:         token,
	}
	rtc := config.WebRTCConfig{
		Ices: []webrtc.ICEServer{{
			URLs: []string{
				"stun:stun.l.google.com:19302",
			},
		}, {
			URLs: []string{Stun},
		}, {
			URLs:           []string{Turn},
			Username:       TurnUser,
			Credential:     TurnPassword,
			CredentialType: webrtc.ICECredentialTypePassword,
		},
		},
	}

	chans := config.NewDataChannelConfig([]string{"hid","adaptive","manual"});
	br := []*config.BroadcasterConfig{}
	Lists := []listener.Listener{}
	lis   := []*config.ListenerConfig{{
		StreamID:  "video",
		Codec:     webrtc.MimeTypeH264,
	}, {
		StreamID:  "audio",
		Codec:     webrtc.MimeTypeOpus,
	}}

	for _, conf := range lis {
		if conf.StreamID == "video" {
			Lists = append(Lists, video.CreatePipeline(conf,chans.Confs["adaptive"]))
		} else if conf.StreamID == "audio" {
			Lists = append(Lists, audio.CreatePipeline(conf))
		} else {
			continue
		}
	}


	hid.NewHIDSingleton(HIDURL,chans.Confs["hid"])
	prox, err := proxy.InitWebRTCProxy(nil, &grpc, &rtc, br, chans, Lists, qr,
		func(monitor tool.Monitor, soundcard tool.Soundcard) error {
			for _, listener := range Lists {
				conf := listener.GetConfig()
				if conf.StreamID == "video" {
					err := listener.SetSource(&monitor)
					if err != nil {
						return err
					}
				} else if conf.StreamID == "audio" {
					err := listener.SetSource(&soundcard)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
	)

	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}
	<-prox.Shutdown
}
