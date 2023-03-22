package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub"
	"github.com/thinkonmay/thinkremote-rtchub/hid"
	"github.com/thinkonmay/thinkremote-rtchub/listener"
	"github.com/thinkonmay/thinkremote-rtchub/listener/audio"
	"github.com/thinkonmay/thinkremote-rtchub/listener/video"
	grpc "github.com/thinkonmay/thinkremote-rtchub/signalling/gRPC"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
	"github.com/thinkonmay/thinkshare-daemon/session/ice"
	"github.com/thinkonmay/thinkshare-daemon/session/signaling"
)

func main() {
	args := os.Args[1:]
	token, webrtcString, videoArg, audioArg, grpcString := "","","","",""
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
		} else if arg == "--audio" {
			audioArg = args[i+1]
		} else if arg == "--video" {
			videoArg = args[i+1]
		}
	}

	if token == "" {
		fmt.Printf("no available token")
		return
	}


	chans := config.NewDataChannelConfig([]string{"hid", "adaptive", "manual"})

	bytes,_ := base64.RawStdEncoding.DecodeString(videoArg)
	videopipeline,err := video.CreatePipeline(string(bytes), chans.Confs["adaptive"], chans.Confs["manual"])
	if err != nil {
		fmt.Printf("error initiate video pipeline %s",err.Error())
		return
	}

	bytes,_ = base64.RawStdEncoding.DecodeString(audioArg)
	audioPipeline,err := audio.CreatePipeline(string(bytes))
	if err != nil {
		fmt.Printf("error initiate audio pipeline %s",err.Error())
		return
	}

	audioPipeline.Open()
	videopipeline.Open()
	Lists := []listener.Listener{audioPipeline, videopipeline}
	handle_track := func(tr *webrtc.TrackRemote)  { }

	signaling_conf := signaling.DecodeSignalingConfig(grpcString)
	grpc_conf := &config.GrpcConfig{
		Port:          signaling_conf.Grpcport,
		ServerAddress: signaling_conf.Grpcip,
		Token:         token,
	}


	fmt.Printf("starting websocket connection establishment with hid server at %s\n", HIDURL)
	hid.NewHIDSingleton(HIDURL, chans.Confs["hid"])
	rtc := &config.WebRTCConfig{Ices: ice.DecodeWebRTCConfig(webrtcString).ICEServers}

	go func ()  {
		for {
			signaling_client, err := grpc.InitGRPCClient(grpc_conf)
			if err != nil {
				fmt.Printf("error initiate signaling client %s",err.Error())
				continue
			}

			_,err = proxy.InitWebRTCProxy(signaling_client, rtc, chans, Lists,handle_track)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}
			signaling_client.WaitForEnd()
		}
	}()
}

