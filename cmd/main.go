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
	"github.com/thinkonmay/thinkremote-rtchub/signalling/websocket"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
)

/*
#include <Windows.h>
typedef enum _D3DKMT_SCHEDULINGPRIORITYCLASS {
	D3DKMT_SCHEDULINGPRIORITYCLASS_IDLE,
	D3DKMT_SCHEDULINGPRIORITYCLASS_BELOW_NORMAL,
	D3DKMT_SCHEDULINGPRIORITYCLASS_NORMAL,
	D3DKMT_SCHEDULINGPRIORITYCLASS_ABOVE_NORMAL,
	D3DKMT_SCHEDULINGPRIORITYCLASS_HIGH,
	D3DKMT_SCHEDULINGPRIORITYCLASS_REALTIME
} D3DKMT_SCHEDULINGPRIORITYCLASS;

typedef UINT D3DKMT_HANDLE;

typedef struct _D3DKMT_OPENADAPTERFROMLUID {
	LUID AdapterLuid;
	D3DKMT_HANDLE hAdapter;
} D3DKMT_OPENADAPTERFROMLUID;

typedef struct _D3DKMT_QUERYADAPTERINFO {
	D3DKMT_HANDLE hAdapter;
	UINT Type;
	VOID *pPrivateDriverData;
	UINT PrivateDriverDataSize;
} D3DKMT_QUERYADAPTERINFO;

typedef struct _D3DKMT_CLOSEADAPTER {
	D3DKMT_HANDLE hAdapter;
} D3DKMT_CLOSEADAPTER;

typedef NTSTATUS(WINAPI *PD3DKMTSetProcessSchedulingPriorityClass)(HANDLE, D3DKMT_SCHEDULINGPRIORITYCLASS);
typedef NTSTATUS(WINAPI *PD3DKMTOpenAdapterFromLuid)(D3DKMT_OPENADAPTERFROMLUID *);
typedef NTSTATUS(WINAPI *PD3DKMTQueryAdapterInfo)(D3DKMT_QUERYADAPTERINFO *);
typedef NTSTATUS(WINAPI *PD3DKMTCloseAdapter)(D3DKMT_CLOSEADAPTER *);

int SetGPURealtimePriority() {
    HMODULE gdi32 = GetModuleHandleA("GDI32");

	PD3DKMTSetProcessSchedulingPriorityClass d3dkmt_set_process_priority = 
		(PD3DKMTSetProcessSchedulingPriorityClass) GetProcAddress(gdi32, "D3DKMTSetProcessSchedulingPriorityClass");

	if (!d3dkmt_set_process_priority)
		return 0;

	D3DKMT_SCHEDULINGPRIORITYCLASS priority = D3DKMT_SCHEDULINGPRIORITYCLASS_REALTIME;

	// int hags_enabled = check_hags(adapter_desc.AdapterLuid);
	// if (adapter_desc.VendorId == 0x10DE) {
	// As of 2023.07, NVIDIA driver has unfixed bug(s) where "realtime" can cause unrecoverable encoding freeze or outright driver crash
	// This issue happens more frequently with HAGS, in DX12 games or when VRAM is filled close to max capacity
	// Track OBS to see if they find better workaround or NVIDIA fixes it on their end, they seem to be in communication
	// if (hags_enabled && !config::video.nv_realtime_hags) priority = D3DKMT_SCHEDULINGPRIORITYCLASS_HIGH;
	// }
	// BOOST_LOG(info) << "Active GPU has HAGS " << (hags_enabled ? "enabled" : "disabled");
	// BOOST_LOG(info) << "Using " << (priority == D3DKMT_SCHEDULINGPRIORITYCLASS_HIGH ? "high" : "realtime") << " GPU priority";

	if (FAILED(d3dkmt_set_process_priority(GetCurrentProcess(), priority)))
		return 0;

	return 1;
}

*/
import "C"


func init() {
	resulta := C.SetPriorityClass(C.GetCurrentProcess(), C.HIGH_PRIORITY_CLASS)
	resultb := C.SetGPURealtimePriority()
	if resulta == 0 || resultb == 0{
		fmt.Printf("failed to set realtime priority\n")
	} else {
		fmt.Printf("set realtime priority\n")
	}
}

const (
	mockup_audio   = "fakesrc ! appsink name=appsink"
	mockup_video   = "videotestsrc ! openh264enc gop-size=5 ! appsink name=appsink"
	nvidia_default = "d3d11screencapturesrc blocksize=8192 do-timestamp=true ! capsfilter name=framerateFilter ! video/x-raw(memory:D3D11Memory),clock-rate=90000,framerate=55/1 ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! d3d11convert ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! nvd3d11h264enc bitrate=6000 gop-size=-1 preset=5 rate-control=2 strict-gop=true name=encoder repeat-sequence-header=true zero-reorder-delay=true ! video/x-h264,stream-format=(string)byte-stream,profile=(string)main ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! appsink name=appsink"
)

func main() {
	args := os.Args[1:]
	authArg, webrtcArg, videoArg, audioArg, micArg, grpcArg := "", "", "", "", "", ""
	for i, arg := range args {
		if arg == "--auth" {
			authArg = args[i+1]
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
		func(pointer int)   { videopipeline.SetProperty("pointer", pointer) }, 
		func() 			  	{ videopipeline.SetProperty("reset", 0) },
		func() 			  	{ audioPipeline.SetProperty("audio-reset", 0) },
	)

	chans := datachannel.NewDatachannel("hid", "adaptive", "manual")
	chans.RegisterConsumer("adaptive",videopipeline.AdsContext)
	chans.RegisterConsumer("manual",ManualContext)
	chans.RegisterConsumer("hid",hid.NewHIDSingleton())


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



	go func() {
		for {
			signaling_client, err := websocket.InitWebsocketClient( signaling.Audio.URL, &auth)
			if err != nil {
				fmt.Printf("error initiate signaling client %s\n", err.Error())
				continue
			}

			_, err = proxy.InitWebRTCProxy(signaling_client, rtc, chans, 
										[]listener.Listener{audioPipeline}, 
										handle_track)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				continue
			}
			signaling_client.WaitForStart()
		}
	}()
	go func() {
		for {
			signaling_client, err := websocket.InitWebsocketClient( signaling.Video.URL, &auth)
			if err != nil {
				fmt.Printf("error initiate signaling client %s\n", err.Error())
				continue
			}

			_, err = proxy.InitWebRTCProxy(signaling_client, rtc, chans, 
										[]listener.Listener{videopipeline}, 
										handle_track)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				continue
			}
			signaling_client.WaitForStart()
		}
	}()

	chann := make(chan os.Signal, 10)
	signal.Notify(chann, syscall.SIGTERM, os.Interrupt)
	<-chann
}
