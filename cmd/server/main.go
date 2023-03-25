package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"


	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub"
	"github.com/thinkonmay/thinkremote-rtchub/hid"
	"github.com/thinkonmay/thinkremote-rtchub/listener"
	"github.com/thinkonmay/thinkremote-rtchub/listener/audio"
	"github.com/thinkonmay/thinkremote-rtchub/listener/video"
	"github.com/thinkonmay/thinkremote-rtchub/signalling/gRPC"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
)




func main() {
	args := os.Args[1:]
	authArg, webrtcArg, videoArg, audioArg, grpcArg := "","","","",""
	HIDURL := "localhost:5000"
	for i, arg := range args {
		if arg == "--auth" {
			authArg = args[i+1]
		} else if arg == "--hid" {
			HIDURL = args[i+1]
		} else if arg == "--grpc" {
			grpcArg = args[i+1]
		} else if arg == "--webrtc" {
			webrtcArg = args[i+1]
		} else if arg == "--audio" {
			audioArg = args[i+1]
		} else if arg == "--video" {
			videoArg = args[i+1]
		}
	}


	chans := config.NewDataChannelConfig([]string{"hid", "adaptive", "manual"})

	videoPipelineString := ""
	bytes1,_ := base64.StdEncoding.DecodeString(videoArg)
	json.Unmarshal(bytes1, &videoPipelineString)
	videopipeline,err := video.CreatePipeline(videoPipelineString, chans.Confs["adaptive"], chans.Confs["manual"])
	if err != nil {
		fmt.Printf("error initiate video pipeline %s",err.Error())
		return
	}

	audioPipelineString := ""
	bytes2,_ := base64.StdEncoding.DecodeString(audioArg)
	json.Unmarshal(bytes2, &audioPipelineString)
	audioPipeline,err := audio.CreatePipeline(audioPipelineString)
	if err != nil {
		fmt.Printf("error initiate audio pipeline %s",err.Error())
		return
	}

	audioPipeline.Open()
	videopipeline.Open()
	Lists := []listener.Listener{audioPipeline, videopipeline}
	handle_track := func(tr *webrtc.TrackRemote)  { }


	signaling := config.GrpcConfig{}
	bytes3, _ := base64.StdEncoding.DecodeString(grpcArg)
	json.Unmarshal(bytes3, &signaling)

	auth := config.AuthConfig{}
	bytes4, _ := base64.StdEncoding.DecodeString(authArg)
	json.Unmarshal(bytes4, &auth)

	webrtc := webrtc.Configuration{}
	bytes1, _ = base64.StdEncoding.DecodeString(webrtcArg)
	json.Unmarshal(bytes1, &webrtc)
	rtc := &config.WebRTCConfig{Ices: webrtc.ICEServers}

	fmt.Printf("starting websocket connection establishment with hid server at %s\n", HIDURL)
	hid.NewHIDSingleton(HIDURL, chans.Confs["hid"])

	go func ()  {
		for {
			signaling_client, err := grpc.InitGRPCClient(&signaling,&auth)
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

	chann := make(chan os.Signal, 10)
	signal.Notify(chann, syscall.SIGTERM, os.Interrupt)
	<-chann
}

