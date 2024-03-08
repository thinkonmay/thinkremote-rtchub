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

	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel/hid"
	"github.com/thinkonmay/thinkremote-rtchub/listener"
	"github.com/thinkonmay/thinkremote-rtchub/listener/adaptive"
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

const (
	video_url = "ws://localhost:60000/handshake/server?token=video"
	audio_url = "ws://localhost:60000/handshake/server?token=audio"
)

func init() {
	resulta := C.SetPriorityClass(C.GetCurrentProcess(), C.REALTIME_PRIORITY_CLASS)
	resultb := C.SetGPURealtimePriority()
	if resulta == 0 || resultb == 0 {
		fmt.Printf("failed to set realtime priority\n")
	} else {
		fmt.Printf("set realtime priority\n")
	}
}

func main() {
	args := os.Args[1:]
	webrtcArg, displayArg := "", ""
	for i, arg := range args {
		if arg == "--display" {
			displayArg = args[i+1]
		} else if arg == "--webrtc" {
			webrtcArg = args[i+1]
		}
	}


	audioPipeline, err := audio.CreatePipeline()
	if err != nil {
		fmt.Printf("error initiate audio pipeline %s\n", err.Error())
		return
	}

	audioPipeline.Open()


	displayB, _ := base64.StdEncoding.DecodeString(displayArg)
	videopipeline, err := video.CreatePipeline(string(displayB))
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
	chans.RegisterConsumer("hid", hid.NewHIDSingleton(string(displayB)))

	videopipeline.Open()

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

	handle_track := func(tr *webrtc.TrackRemote) {}
	go func() {
		for {
			signaling_client, err := websocket.InitWebsocketClient(video_url)
			if err != nil {
				fmt.Printf("error initiate signaling client %s\n", err.Error())
				continue
			}

			_, err = proxy.InitWebRTCProxy(signaling_client,
				rtc,
				chans,
				[]listener.Listener{videopipeline},
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
			signaling_client, err := websocket.InitWebsocketClient(audio_url)
			if err != nil {
				fmt.Printf("error initiate signaling client %s\n", err.Error())
				continue
			}

			_, err = proxy.InitWebRTCProxy(signaling_client,
				rtc,
				chans,
				[]listener.Listener{audioPipeline},
				func(tr *webrtc.TrackRemote) {})
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				continue
			}
			signaling_client.WaitForStart()
		}
	}()

	chann := make(chan os.Signal, 16)
	signal.Notify(chann, syscall.SIGTERM, os.Interrupt)
	<-chann
}
