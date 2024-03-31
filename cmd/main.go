package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pion/webrtc/v3"
	proxy "github.com/thinkonmay/thinkremote-rtchub"

	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel/hid"
	"github.com/thinkonmay/thinkremote-rtchub/listener"
	"github.com/thinkonmay/thinkremote-rtchub/listener/adaptive"
	"github.com/thinkonmay/thinkremote-rtchub/listener/audio"
	"github.com/thinkonmay/thinkremote-rtchub/listener/manual"
	"github.com/thinkonmay/thinkremote-rtchub/listener/video"
	"github.com/thinkonmay/thinkremote-rtchub/signalling/http"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
)

func main() {
	args := os.Args[1:]
	displayArg := ""
	rtc := &config.WebRTCConfig{Ices: []webrtc.ICEServer{{}, {}}}

	video_url := "http://localhost:60000/handshake/server?token=video"
	audio_url := "http://localhost:60000/handshake/server?token=audio"
	for i, arg := range args {
		if arg == "--display" {
			displayArg = args[i+1]
		} else if arg == "--video" {
			video_url = args[i+1]
		} else if arg == "--audio" {
			audio_url = args[i+1]
		} else if arg == "--stun" {
			rtc.Ices[0].URLs = []string{args[i+1]}
		} else if arg == "--turn" {
			rtc.Ices[1].URLs = []string{args[i+1]}
		} else if arg == "--turn_username" {
			rtc.Ices[1].Username = args[i+1]
		} else if arg == "--turn_password" {
			rtc.Ices[1].Credential = args[i+1]
		}
	}

	audioPipeline, err := audio.CreatePipeline()
	if err != nil {
		fmt.Printf("error initiate audio pipeline %s\n", err.Error())
		return
	}

	audioPipeline.Open()

	videopipeline, err := video.CreatePipeline(displayArg)
	if err != nil {
		fmt.Printf("error initiate video pipeline %s\n", err.Error())
		return
	}

	chans := datachannel.NewDatachannel("hid", "adaptive", "manual")
	chans.RegisterConsumer("adaptive", adaptive.NewAdsContext(
		func(bitrate int) { videopipeline.SetProperty("bitrate", bitrate) },
		func() { videopipeline.SetProperty("reset", 0) },
	))
	chans.RegisterConsumer("manual", manual.NewManualCtx(
		func(bitrate int) { videopipeline.SetProperty("bitrate", bitrate) },
		func(framerate int) { videopipeline.SetProperty("framerate", framerate) },
		func(pointer int) { videopipeline.SetProperty("pointer", pointer) },
		func(display string) { videopipeline.SetPropertyS("display", display) },
		func(pointer string) { videopipeline.SetPropertyS("codec", pointer) },
		func() { videopipeline.SetProperty("reset", 0) },
	))
	chans.RegisterConsumer("hid", hid.NewHIDSingleton(displayArg))

	videopipeline.Open()

	handle_track := func(tr *webrtc.TrackRemote) { }
	go func() {
		for {
			signaling_client, err := http.InitHttpClient(video_url)
			if err != nil {
				fmt.Printf("error initiate signaling client %s\n", err.Error())
				continue
			}

			signaling_client.WaitForStart()
			go func() {
				err := proxy.InitWebRTCProxy(signaling_client,
					rtc,
					chans,
					[]listener.Listener{videopipeline},
					handle_track)
				if err != nil {
					fmt.Printf("webrtc error :%s\n", err.Error())
				}
			}()
		}
	}()

	go func() {
		for {
			signaling_client, err := http.InitHttpClient(audio_url)
			if err != nil {
				fmt.Printf("error initiate signaling client %s\n", err.Error())
				continue
			}

			signaling_client.WaitForStart()
			go func() {
				err := proxy.InitWebRTCProxy(signaling_client,
					rtc,
					chans,
					[]listener.Listener{audioPipeline},
					handle_track)
				if err != nil {
					fmt.Printf("webrtc error :%s\n", err.Error())
				}
			}()
		}
	}()

	chann := make(chan os.Signal, 16)
	signal.Notify(chann, syscall.SIGTERM, os.Interrupt)
	<-chann
}
