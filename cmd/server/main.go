package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pion/webrtc/v3"
	proxy "github.com/thinkonmay/thinkremote-rtchub"

	"github.com/thinkonmay/thinkremote-rtchub/broadcaster/microphone"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel/hid"
	"github.com/thinkonmay/thinkremote-rtchub/listener"
	"github.com/thinkonmay/thinkremote-rtchub/listener/audio"
	"github.com/thinkonmay/thinkremote-rtchub/listener/manual"
	"github.com/thinkonmay/thinkremote-rtchub/listener/video"
	grpc "github.com/thinkonmay/thinkremote-rtchub/signalling/gRPC"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
)

const (
	mockup_audio   = "fakesrc ! appsink name=appsink"
	mockup_video   = "videotestsrc ! openh264enc gop-size=5 ! appsink name=appsink"
	nvidia_default = "d3d11screencapturesrc blocksize=8192 do-timestamp=true ! capsfilter name=framerateFilter ! video/x-raw(memory:D3D11Memory),clock-rate=90000,framerate=55/1 ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! d3d11convert ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! nvd3d11h264enc bitrate=6000 gop-size=-1 preset=5 rate-control=2 strict-gop=true name=encoder repeat-sequence-header=true zero-reorder-delay=true ! video/x-h264,stream-format=(string)byte-stream,profile=(string)main ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! appsink name=appsink"
)

func main() {
	args := os.Args[1:]
	authArg, webrtcArg, videoArg, audioArg, micArg, grpcArg := "", "", "", "", "", ""
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
		} else if arg == "--mic" {
			micArg = args[i+1]
		}
	}


	videoPipelineString := ""
	if videoArg == "" {
		videoPipelineString = nvidia_default	
	} else {
		bytes1, _ := base64.StdEncoding.DecodeString(videoArg)
		err := json.Unmarshal(bytes1, &videoPipelineString)
		if err != nil {
			fmt.Printf("error decode audio pipeline %s\n", err.Error())
			videoPipelineString = nvidia_default	
		}
	}

	videopipeline,err := video.CreatePipeline(videoPipelineString)
	if err != nil {
		fmt.Printf("error initiate video pipeline %s\n", err.Error())
		return
	}

	audioPipelineString := ""
	if videoArg == "" {
		audioPipelineString = mockup_audio
	} else {
		bytes2, _ := base64.StdEncoding.DecodeString(audioArg)
		err := json.Unmarshal(bytes2, &audioPipelineString)
		if err != nil {
			fmt.Printf("error decode audio pipeline %s\n", err.Error())
			audioPipelineString = mockup_audio
		}
	}
	
	audioPipeline, err := audio.CreatePipeline(audioPipelineString)
	if err != nil {
		fmt.Printf("error initiate audio pipeline %s\n", err.Error())
		return
	}

	ManualContext := manual.NewManualCtx(
		func(bitrate int) 	{ videopipeline.SetProperty("bitrate", bitrate) }, 
		func(framerate int) { videopipeline.SetProperty("framerate", framerate) }, 
		func() 			  	{ videopipeline.SetProperty("reset", 0) },
		func() 			  	{ audioPipeline.SetProperty("audio-reset", 0) },
	)

	chans := datachannel.NewDatachannel("hid", "adaptive", "manual")
	chans.RegisterConsumer("adaptive",videopipeline.AdsContext)
	chans.RegisterConsumer("manual",ManualContext)


	audioPipeline.Open()
	videopipeline.Open()

	micPipelineString := ""
	if videoArg != "" {
		bytes2, _ := base64.StdEncoding.DecodeString(micArg)
		err := json.Unmarshal(bytes2, &micPipelineString)
		if err != nil {
			fmt.Printf("error decode audio pipeline %s\n", err.Error())
			micPipelineString = ""
		}
	}
	handle_track := func(tr *webrtc.TrackRemote) {
		codec := tr.Codec() 
		if codec.MimeType != "audio/opus" ||
		   codec.Channels != 2 {
			fmt.Printf("failed to create pipeline, reason: %s\n","media not supported")
			return
		}

		if micPipelineString == "" {
			fmt.Printf("failed to create pipeline, reason: microphone not support on this session\n")
			return
		}

		pipeline,err := microphone.CreatePipeline(micPipelineString)
		if err != nil {
			fmt.Printf("failed to create pipeline, reason: %s\n",err.Error())
			return
		}

		pipeline.Open()
		defer pipeline.Close()

		buf := make([]byte, 1400)
		for {
			i, _, readErr := tr.Read(buf)
			if readErr != nil {
				fmt.Printf("connection stopped, reason: %s\n",readErr.Error())
				break
			}

			pipeline.Push(buf[:i])
		}
	}

	signaling := config.GrpcConfig{}
	bytes3, _ := base64.StdEncoding.DecodeString(grpcArg)
	err = json.Unmarshal(bytes3, &signaling)
	if err != nil {
		fmt.Printf("error decode signaling config %s\n", err.Error())
		return
	}


	auth := config.AuthConfig{}
	bytes4, _ := base64.StdEncoding.DecodeString(authArg)
	err = json.Unmarshal(bytes4, &auth)
	if err != nil {
		fmt.Printf("error decode auth config %s\n", err.Error())
		return
	}

	bytes1, _ := base64.StdEncoding.DecodeString(webrtcArg)
	var data map[string]interface{}
	json.Unmarshal(bytes1, &data)
	rtc := &config.WebRTCConfig{Ices: make([]webrtc.ICEServer, 0)}
	for _, v := range data["iceServers"].([]interface{}) {
		ice := webrtc.ICEServer{
			URLs: []string{v.(map[string]interface{})["urls"].(string)},
		}
		if v.(map[string]interface{})["credential"] != nil {
			ice.Credential = v.(map[string]interface{})["credential"].(string)
			ice.Username = v.(map[string]interface{})["username"].(string)
		}
		rtc.Ices = append(rtc.Ices, ice)
	}



	fmt.Printf("starting websocket connection establishment with hid server at %s\n", HIDURL)
	chans.RegisterConsumer("hid",hid.NewHIDSingleton(HIDURL))
	go func() {
		for {
			signaling_client, err := grpc.InitGRPCClient(
				fmt.Sprintf("%s:%d", signaling.ServerAddress,signaling.Audio.GrpcPort), 
				&auth)
			if err != nil {
				fmt.Printf("error initiate signaling client %s\n", err.Error())
				continue
			}

			_, err = proxy.InitWebRTCProxy(signaling_client, rtc, chans, 
										[]listener.Listener{audioPipeline}, 
										handle_track)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}
			signaling_client.WaitForEnd()
		}
	}()
	go func() {
		for {
			signaling_client, err := grpc.InitGRPCClient(
				fmt.Sprintf("%s:%d", signaling.ServerAddress,signaling.Video.GrpcPort), 
				&auth)
			if err != nil {
				fmt.Printf("error initiate signaling client %s\n", err.Error())
				continue
			}

			_, err = proxy.InitWebRTCProxy(signaling_client, rtc, chans, 
										[]listener.Listener{videopipeline}, 
										handle_track)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}
			signaling_client.WaitForEnd()
		}
	}()
	_ = func() {
		for {
			signaling_client, err := grpc.InitGRPCClient(
				fmt.Sprintf("%s:%d", signaling.ServerAddress,signaling.Data.GrpcPort), 
				&auth)
			if err != nil {
				fmt.Printf("error initiate signaling client %s\n", err.Error())
				continue
			}

			_, err = proxy.InitWebRTCProxy(signaling_client, rtc, chans, 
										[]listener.Listener{}, 
										handle_track)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}
			signaling_client.WaitForEnd()
		}
	}

	chann := make(chan os.Signal, 10)
	signal.Notify(chann, syscall.SIGTERM, os.Interrupt)
	<-chann
}
