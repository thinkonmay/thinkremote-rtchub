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
	// "github.com/thinkonmay/thinkremote-rtchub/hid"
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
	fmt.Printf(videoPipelineString)
	videopipeline,err := video.CreatePipeline("d3d11screencapturesrc blocksize=8192 do-timestamp=true monitor-handle=65537 ! capsfilter name=framerateFilter ! video/x-raw(memory:D3D11Memory),clock-rate=90000 ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! d3d11convert ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! nvd3d11h264enc name=encoder ! video/x-h264,stream-format=(string)byte-stream,profile=(string)main ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! appsink name=appsink", chans.Confs["adaptive"], chans.Confs["manual"])
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
	err = json.Unmarshal(bytes3, &signaling)
	signaling.ServerAddress = "192.168.1.4"
	signaling.Port = 4000

	auth := config.AuthConfig{}
	bytes4, _ := base64.StdEncoding.DecodeString(authArg)
	err = json.Unmarshal(bytes4, &auth)
	auth.Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ3b3JrZXJfc2Vzc2lvbl9jcmVhdGVfZ2VuX3Rva2VuIiwiaWF0IjoxNjc5NzE5MTQ1LCJleHAiOjE3MTEyNTUxNDUsInN1YiI6IntcInNlc3Npb25fYWNjb3VudF9pZFwiOjQ0fSJ9.JI9e9yozVDfVD0RZnj5JR3tsw6wLlD_lQoVeTgefCKCnbXqeXuZQJjF3VWHC8g9pqCN2NECxdsJxNSHmcjQKXsJ317A4T-TGOkt_a0x7jozd4lBKOTuuWzZMHJTqP0LWm-2IMez6txiITzpkqlLhGAZkCR6YtkCt33p1Y9s1F0sIRq6RHOTOMel2VqGbp6c5kbzriMPZIRzA_3ZIew8M5KqlSd7XNzrlCU3UqlBh-wTByV1XRmrM2DNgQVs78DJSMUdZQK4ELF3hK1-sWFpIhCtGMCFu_83AtCiY-9CQv5NacBkgGlPNT1VDyPFnXbaMwF_c0Cv4Jkf3n66R3I-847rpO-rdCWiDpEOEjMPR6cp_0pPR5tGK4PbmGY73yYFnLRvHx1gRBB293ACejRYL5oYcfxXHj1LOv0y8vGYFV2qotVCKz48dkhOfDH44-aRCkdv7rCpm2JeK9RSAVNqGdDF_1ZIEUsQyd2Yc8Ss-_IPk-2SkhH5XMFd3hCMGikhZcqkxiCZaxaw_8CtWfHNzFp-YG3mEQmTr3OTpX4X_1CJBuffQ_fbafd8oBfLMxMNtd-hoQWs-KOtA8NIlALOF9sr5KUAm_WQO_DvSJv50zbZNJV5t4WhAJ8m_PJdRGWIYc7ffMQIPNP4JcTDIhB9p_M4mFGRTLTGe7DPG5sOnpUQ"

	bytes1, _ = base64.StdEncoding.DecodeString(webrtcArg)
	var i map[string]interface{}
	json.Unmarshal(bytes1, &i)
	rtc := &config.WebRTCConfig{Ices: make([]webrtc.ICEServer, 0)}
	for _,v := range i["iceServers"].([]interface{}) {
		ice := webrtc.ICEServer{
			URLs: []string{v.(map[string]interface{})["url"].(string)},
		}
		if v.(map[string]interface{})["credential"] != nil {
			ice.Credential = v.(map[string]interface{})["credential"].(string)
			ice.Username = v.(map[string]interface{})["username"].(string)
		}
		rtc.Ices = append(rtc.Ices,ice)
	}

	fmt.Printf("starting websocket connection establishment with hid server at %s\n", HIDURL)
	// hid.NewHIDSingleton(HIDURL, chans.Confs["hid"])
	fmt.Printf("starting %s \n",HIDURL)

	go func ()  {
		for {
			signaling_client, err := grpc.InitGRPCClient(&signaling,&auth)
			if err != nil {
				fmt.Printf("error initiate signaling client %s",err.Error())
				continue
			}

			prox,err := proxy.InitWebRTCProxy(signaling_client, rtc, chans, Lists,handle_track)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}
			signaling_client.WaitForEnd()
			prox.Stop()
		}
	}()

	chann := make(chan os.Signal, 10)
	signal.Notify(chann, syscall.SIGTERM, os.Interrupt)
	<-chann
}

