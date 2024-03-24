package main

import (
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
	"github.com/thinkonmay/thinkremote-rtchub/listener/adaptive"
	"github.com/thinkonmay/thinkremote-rtchub/listener/audio"
	"github.com/thinkonmay/thinkremote-rtchub/listener/manual"
	"github.com/thinkonmay/thinkremote-rtchub/listener/video"
	"github.com/thinkonmay/thinkremote-rtchub/signalling/http"
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
	video_url = "http://localhost:60000/handshake/server?token=video"
	audio_url = "http://localhost:60000/handshake/server?token=audio"
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
	displayArg := ""
	rtc := &config.WebRTCConfig{Ices: []webrtc.ICEServer{{}, {}}}

	for i, arg := range args {
		if arg == "--display" {
			displayArg = args[i+1]
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

	buff := make(chan *[]byte, 256)
	microphone.StartMicrophone(buff)
	defer microphone.CloseMicrophone()
	handle_track := func(tr *webrtc.TrackRemote) {
		for {
			pkt, _, err := tr.ReadRTP()
			if err != nil {
				break
			}
			buff <- &pkt.Payload
		}
	}
	go func() {
		for {
			signaling_client, err := http.InitHttpClient(video_url)
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
			signaling_client, err := http.InitHttpClient(audio_url)
			if err != nil {
				fmt.Printf("error initiate signaling client %s\n", err.Error())
				continue
			}

			_, err = proxy.InitWebRTCProxy(signaling_client,
				rtc,
				chans,
				[]listener.Listener{audioPipeline},
				handle_track)
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
