package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pion/rtp"
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
	count := 1
	var max int64 = 0
	last := time.Now().UnixMicro()
	videopipeline.RegisterRTPHandler("abc",func(pkt *rtp.Packet) {
		diff := time.Now().UnixMicro() - last
		last = last + diff
		count ++;
		
		if diff > max && diff < 100000 {
			fmt.Printf("%d size %d , %d\n",count,len(pkt.Payload),diff)
			max = diff
		}
		count++
	})

	<-make(chan bool, 0)
}