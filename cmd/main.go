package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/pion/webrtc/v4"
	proxy "github.com/thinkonmay/thinkremote-rtchub"

	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel/hid"
	"github.com/thinkonmay/thinkremote-rtchub/listener"
	"github.com/thinkonmay/thinkremote-rtchub/listener/audio"
	"github.com/thinkonmay/thinkremote-rtchub/listener/manual"
	"github.com/thinkonmay/thinkremote-rtchub/listener/video"
	"github.com/thinkonmay/thinkremote-rtchub/signalling/http"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
)

const (
	DisplayFailureCode = 77
)

func main() {
	args := os.Args[1:]
	rtc := &config.WebRTCConfig{Ices: []webrtc.ICEServer{{}, {}}}

	token := ""
	videochannel := int64(0)
	video_url := "http://localhost:60000/handshake/server?token=video"
	audio_url := "http://localhost:60000/handshake/server?token=audio"
	for i, arg := range args {
		if arg == "--token" {
			token = args[i+1]
		} else if arg == "--video_channel" {
			videochannel, _ = strconv.ParseInt(args[i+1], 10, 16)
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

	defer func() {
		if err := recover(); err != nil {
			os.Exit(DisplayFailureCode)
		} else {
			os.Exit(0)
		}
	}()

	memory, err := proxy.ObtainSharedMemory(token)
	if err != nil {
		fmt.Printf("error obtain shared memory %s\n", err.Error())
		return
	}

	audioPipeline, err := audio.CreatePipeline(memory.GetQueue(proxy.Audio))
	if err != nil {
		fmt.Printf("error initiate audio pipeline %s\n", err.Error())
		return
	}

	videopipeline, err := video.CreatePipeline(memory.GetQueue(int(videochannel)))
	if err != nil {
		fmt.Printf("error initiate video pipeline %s\n", err.Error())
		return
	}

	chans := datachannel.NewDatachannel("hid", "manual")
	chans.RegisterConsumer("manual", manual.NewManualCtx(memory.GetQueue(int(videochannel))))
	chans.RegisterConsumer("hid", hid.NewHIDSingleton(memory.GetQueue(int(videochannel))))

	handle_idr := func() { memory.GetQueue(int(videochannel)).Raise(proxy.Idr, 1) }
	handle_track := func(tr *webrtc.TrackRemote) {}
	video := func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("panic in video thread %v", err)
			}
		}()
		for {
			signaling_client, err := http.InitHttpClient(video_url)
			if err != nil {
				fmt.Printf("error initiate signaling client %s\n", err.Error())
				continue
			}

			signaling_client.WaitForStart()
			go func() {
				defer func() {
					if err := recover(); err != nil {
						fmt.Printf("panic in webrtc video thread %v", err)
					}
				}()
				err := proxy.InitWebRTCProxy(signaling_client,
					rtc,
					chans,
					[]listener.Listener{videopipeline},
					handle_track,
					handle_idr,
				)
				if err != nil {
					fmt.Printf("webrtc error :%s\n", err.Error())
				}
			}()
		}
	}

	audio := func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("panic in audio thread %v", err)
			}
		}()
		for {
			signaling_client, err := http.InitHttpClient(audio_url)
			if err != nil {
				fmt.Printf("error initiate signaling client %s\n", err.Error())
				continue
			}

			signaling_client.WaitForStart()
			go func() {
				defer func() {
					if err := recover(); err != nil {
						fmt.Printf("panic in webrtc audio thread %v", err)
					}
				}()
				err := proxy.InitWebRTCProxy(signaling_client,
					rtc,
					chans,
					[]listener.Listener{audioPipeline},
					handle_track,
					func() {},
				)
				if err != nil {
					fmt.Printf("webrtc error :%s\n", err.Error())
				}
			}()
		}
	}

	retry := func(execute func()) {
		for {
			execute()
		}
	}

	go retry(audio)
	go retry(video)
	chann := make(chan os.Signal, 16)
	signal.Notify(chann, syscall.SIGTERM, os.Interrupt)
	<-chann
}
