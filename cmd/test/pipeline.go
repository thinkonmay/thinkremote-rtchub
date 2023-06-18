package main

import (
	"fmt"
	"os"
	"time"

	"github.com/thinkonmay/thinkremote-rtchub/listener/audio"
	"github.com/thinkonmay/thinkremote-rtchub/listener/video"
)

const (
	source = "tmd3d11screencapturesrc blocksize=8192 do-timestamp=true ! video/x-raw(memory:D3D11Memory),clock-rate=90000,framerate=55/1 ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! tmd3d11convert ! video/x-raw(memory:D3D11Memory) ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! "
	nvidia_default = source + "tmnvd3d11h264enc name=encoder ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! appsink name=appsink"
	amd_default    = source + "tmamfh264enc 	name=encoder ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! appsink name=appsink"
	intel_default  = source + "tmqsvh264enc 	name=encoder ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! appsink name=appsink"

	audio_default = "wasapisrc name=source device=\"\\{0.0.1.00000000\\}.\\{4d54e66c-3242-4385-bff8-9c82dca3682a\\}\" ! audioresample ! audio/x-raw,rate=48000 ! audioconvert ! opusenc name=encoder ! appsink name=appsink"
)

func main() {

	pipeline := ""
	switch os.Args[1] {
	case "amd":
		pipeline = amd_default
		break
	case "nvidia":
		pipeline = nvidia_default
		break
	case "intel":
		pipeline = intel_default
		break
	case "audio":
		audiopipeline, err := audio.CreatePipeline(audio_default)
		if err != nil {
			fmt.Printf("error initiate audio pipeline %s\n", err.Error())
			return
		}

		go func ()  {
			for {
				audiopipeline.SetProperty("audio-reset",0)
				time.Sleep(time.Second)
			}
		}()
		audiopipeline.Open()
		<-make(chan bool, 0)
		return
	}

	videopipeline, err := video.CreatePipeline(pipeline)

	if err != nil {
		fmt.Printf("error initiate video pipeline %s\n", err.Error())
		return
	}
	videopipeline.Open()
	go func ()  {
		for {
			videopipeline.SetProperty("bitrate",8000)
			time.Sleep(time.Second)
			videopipeline.SetProperty("framerate",70)
			time.Sleep(time.Second)
			videopipeline.SetProperty("reset",0)
			time.Sleep(time.Second)
		}
	}()

	<-make(chan bool, 0)
}