package main

import (
	"fmt"
	"os"
	"strconv"

	proxy "github.com/OnePlay-Internet/webrtc-proxy"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
	"github.com/OnePlay-Internet/webrtc-proxy/hid"
	"github.com/OnePlay-Internet/webrtc-proxy/listener"
	"github.com/OnePlay-Internet/webrtc-proxy/listener/audio"
	"github.com/OnePlay-Internet/webrtc-proxy/listener/video"

	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/webrtc/v3"
)

func main() {
	var err error
	var token string
	args := os.Args[1:]
	HIDURL := "localhost:5000"
	MonitorOrder := -1;

	signaling := "54.169.49.176"
	Port :=      30000;

	Stun :=		"stun:workstation.thinkmay.net:3478";
	Turn :=		"turn:workstation.thinkmay.net:3478";

	TurnUser 	 :=		"oneplay";
	TurnPassword :=		"oneplay";
	engine 		 :=		"screencoder";


	qr := tool.GetDevice()
	if len(qr.Monitors) == 0{
		fmt.Printf("no display available");
		return;
	}

	for i, arg := range args {
		if arg == "--token" {
			token = args[i+1]
		} else if arg == "--monitor" {
			MonitorOrder,err = strconv.Atoi(args[i+1])
		} else if arg == "--hid" {
			HIDURL = args[i+1]
		} else if arg == "--grpc" {
			signaling = args[i+1]
		} else if arg == "--grpcport" {
			Port,err = strconv.Atoi(args[i+1])
		} else if arg == "--turn" {
			Turn = args[i+1]
		} else if arg == "--turnuser" {
			TurnUser = args[i+1]
		} else if arg == "--turnpassword" {
			TurnPassword = args[i+1]
		} else if arg == "--engine" {
			engine = args[i+1]
		} else if arg == "--device" {
				fmt.Printf("=======================================================================\n")
				fmt.Printf("MONITOR DEVICE\n")
			for index,monitor := range qr.Monitors {
				fmt.Printf("=======================================================================\n")
				fmt.Printf("monitor %d\n",index)
				fmt.Printf("monitor name 			%s\n",monitor.MonitorName)
				fmt.Printf("monitor handle  		%d\n",monitor.MonitorHandle)
				fmt.Printf("monitor adapter 		%s\n",monitor.Adapter)
				fmt.Printf("monitor device  		%s\n",monitor.DeviceName);
				fmt.Printf("=======================================================================\n")
			}
				fmt.Printf("\n\n\n\n")

				fmt.Printf("=======================================================================\n")
				fmt.Printf("AUDIO DEVICE\n")
			for index,audio := range qr.Soundcards {
				fmt.Printf("=======================================================================\n")
				fmt.Printf("audio source 			%d\n",index)
				fmt.Printf("audio source name 		%s\n",audio.Name)
				fmt.Printf("audio source device id  %s\n",audio.DeviceID)
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

	if token == "" { err = fmt.Errorf("no available token"); }
	if err != nil {
		fmt.Printf("invalid argument : %s\n",err.Error());
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
			URLs: []string{ Stun, },
		}, {
				URLs: []string { Turn },
				Username: TurnUser,
				Credential: TurnPassword,
				CredentialType: webrtc.ICECredentialTypePassword,
			},
		},
	}

	br  := []*config.BroadcasterConfig{}

	var selection tool.Monitor
	if MonitorOrder == -1 {
		for _,i := range qr.Monitors {
			if i.IsPrimary {
				selection =  i;
			}
		}
	} else {
		selection = qr.Monitors[MonitorOrder];
	}

	lis := []*config.ListenerConfig{{
		VideoSource: selection,

		DataType:  "rtp",
		MediaType: "video",
		Name:      "Screencapture",
		Codec:     webrtc.MimeTypeH264,
	},
	{
		VideoSource: selection,

		DataType: "sample",

		Bitrate: 3000,
		MediaType: "video",
		Name:      "videoGstreamer",
		Codec:     webrtc.MimeTypeH264,
	}, }

	if len(qr.Soundcards) > 0{
		selection := func() tool.Soundcard {
			for _,soundcard := range qr.Soundcards {
				if soundcard.Api == "wasapi2" && soundcard.IsDefault && soundcard.IsLoopback{
					return soundcard
				} else if soundcard.Api == "wasapi2" && soundcard.IsLoopback{
					return soundcard
				} else if soundcard.Api == "wasapi2" {
					return soundcard
				} else if soundcard.Api == "wasapi" {
					return soundcard
				} 
			} 
			return tool.Soundcard{Api: "none"}
		}()

		if selection.Api != "none" {
			lis = append(lis, &config.ListenerConfig{
				AudioSource: selection,
				
				DataType: "sample",

				Bitrate: 128000,
				MediaType: "audio",
				Name:      "audioGstreamer",
				Codec:     webrtc.MimeTypeOpus,
			})
		}
	}
	

	Lists := make([]listener.Listener, 0)
	for _, conf := range lis {
		var err error;
		var Lis listener.Listener

		if engine == "gstreamer" {
			if conf.MediaType == "video" && conf.DataType == "sample"{
				Lis,err     =  video.CreatePipeline(conf);
			} else if conf.MediaType == "audio" && conf.DataType == "sample"{
				Lis,err     =  audio.CreatePipeline(conf);
			} else {
				err = fmt.Errorf("unimplemented listener")
			}
		} else {
			err = fmt.Errorf("unimplemented listener")
		}

		if err != nil {
			fmt.Printf("%s\n",err.Error());
		} else if (Lis != nil) {
			Lists = append(Lists, Lis)
		}
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

	_hid := hid.NewHIDSingleton(HIDURL)
	go func() {
		for {
			channel := chans.Confs["hid"]
			if channel != nil {
				str := <-chans.Confs["hid"].Recv
				_hid.ParseHIDInput(str)
			} else {
				return
			}
		}
	}()



	prox, err := proxy.InitWebRTCProxy(nil, &grpc, &rtc, br, &chans, Lists)
	if err != nil {
		fmt.Printf("%s\n",err.Error())
		return;
	}
	<-prox.Shutdown
}

