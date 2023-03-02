package main

import (
	"fmt"
	"os"

	proxy "github.com/thinkonmay/thinkremote-rtchub"
	"github.com/thinkonmay/thinkremote-rtchub/broadcaster"
	"github.com/thinkonmay/thinkremote-rtchub/broadcaster/dummy"
	sink "github.com/thinkonmay/thinkremote-rtchub/broadcaster/gstreamer"
	"github.com/thinkonmay/thinkremote-rtchub/hid"
	"github.com/thinkonmay/thinkremote-rtchub/listener"
	"github.com/thinkonmay/thinkshare-daemon/session/ice"
	"github.com/thinkonmay/thinkshare-daemon/session/signaling"

	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
)

func main() {
	var token string
	args := os.Args[1:]

	grpcString := ""
	webrtcString := ""

	HIDURL := "localhost:5000"
	for i, arg := range args {
		if arg == "--token" {
			token = args[i+1]
		} else if arg == "--hid" {
			HIDURL = args[i+1]
		} else if arg == "--grpc" {
			grpcString = args[i+1]
		} else if arg == "--webrtc" {
			webrtcString = args[i+1]
		} else if arg == "--help" {
			fmt.Printf("--token 	 	 |  server token\n")
			fmt.Printf("--hid   	 	 |  HID server URL (example: localhost:5000)\n")
			return
		}
	}
	if token == "" {
		fmt.Printf("no available token")
		return
	}

	signaling := signaling.DecodeSignalingConfig(grpcString)
	grpc := &config.GrpcConfig{
		Port:          signaling.Grpcport,
		ServerAddress: signaling.Grpcip,
		Token:         token,
	}

	rtc := &config.WebRTCConfig{Ices: ice.DecodeWebRTCConfig(webrtcString).ICEServers}
	chans := config.NewDataChannelConfig([]string{"hid", "adaptive", "manual"})
	br := []*config.BroadcasterConfig{}
	Lists := []listener.Listener{
		// audio.CreatePipeline(&config.ListenerConfig{
		// 	StreamID: "audio",
		// 	Codec:    webrtc.MimeTypeOpus,
		// }), video.CreatePipeline(&config.ListenerConfig{
		// 	StreamID: "video",
		// 	Codec:    webrtc.MimeTypeH264,
		// }, chans.Confs["adaptive"], chans.Confs["manual"]),
	}

	fmt.Printf("starting websocket connection establishment with hid server at %s\n", HIDURL)
	hid.NewHIDSingleton(HIDURL, chans.Confs["hid"])
	prox, err := proxy.InitWebRTCProxy(nil, grpc, rtc, chans, Lists,
		func(tr *webrtc.TrackRemote) (broadcaster.Broadcaster, error) {
			for _, conf := range br {
				if tr.Codec().MimeType == conf.Codec {
					return sink.CreatePipeline(conf)
				}
			}
			fmt.Printf("no available codec handler, using dummy sink\n")
			return dummy.NewDummyBroadcaster(&config.BroadcasterConfig{
				Name:  "dummy",
				Codec: "any",
			})
		},
	)

	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}
	<-prox.Shutdown
}

		// func(selection signalling.DeviceSelection) (*tool.MediaDevice, error) {
		// 	monitor := func() tool.Monitor {
		// 		for _, monitor := range devices.Monitors {
		// 			sel, err := strconv.ParseInt(selection.Monitor, 10, 32)
		// 			if err != nil {
		// 				return tool.Monitor{}
		// 			}

		// 			if monitor.MonitorHandle == int(sel) {
		// 				return monitor
		// 			}
		// 		}
		// 		return tool.Monitor{MonitorHandle: -1}
		// 	}()
		// 	soundcard := func() tool.Soundcard {
		// 		for _, soundcard := range devices.Soundcards {
		// 			if soundcard.DeviceID == selection.SoundCard {
		// 				return soundcard
		// 			}
		// 		}
		// 		return tool.Soundcard{DeviceID: "none"}
		// 	}()

		// 	for _, listener := range Lists {
		// 		conf := listener.GetConfig()
		// 		if conf.StreamID == "video" {
		// 			err := listener.SetSource(&monitor)

		// 			framerate := selection.Framerate
		// 			if 10 < framerate && framerate < 200 {
		// 				listener.SetProperty("framerate", int(framerate))
		// 			}

		// 			bitrate := selection.Bitrate
		// 			if 100 < bitrate && bitrate < 20000 {
		// 				listener.SetProperty("bitrate", int(bitrate))
		// 			}

		// 			if err != nil {
		// 				return devices, err
		// 			}

		// 		} else if conf.StreamID == "audio" {
		// 			err := listener.SetSource(&soundcard)
		// 			if err != nil {
		// 				return devices, err
		// 			}
		// 		}
		// 	}
		// 	return nil, nil
		// },