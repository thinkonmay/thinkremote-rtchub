package main

import (
	"fmt"
	"os"

	"github.com/thinkonmay/thinkremote-rtchub/listener/video"
)

const (
	source = "tmd3d11screencapturesrc blocksize=8192 do-timestamp=true ! video/x-raw(memory:D3D11Memory),clock-rate=90000,framerate=55/1 ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! tmd3d11convert ! video/x-raw(memory:D3D11Memory) ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! "
	nvidia_default = source + "tmnvd3d11h264enc name=encoder ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! appsink name=appsink"
	amd_default    = source + "tmamfh264enc 	name=encoder ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! appsink name=appsink"
	intel_default  = source + "tmqsvh264enc 	name=encoder ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! appsink name=appsink"
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
	}

	videopipeline, err := video.CreatePipeline(pipeline)

	if err != nil {
		fmt.Printf("error initiate video pipeline %s\n", err.Error())
		return
	}
	videopipeline.Open()
	<-make(chan bool, 0)
}