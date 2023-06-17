package main

import (
	"fmt"

	"github.com/thinkonmay/thinkremote-rtchub/listener/video"
)

const (
	mockup_audio   = "fakesrc ! appsink name=appsink"
	mockup_video   = "videotestsrc ! openh264enc gop-size=5 ! appsink name=appsink"
	nvidia_default = "d3d11screencapturesrc blocksize=8192 do-timestamp=true ! capsfilter name=framerateFilter ! video/x-raw(memory:D3D11Memory),clock-rate=90000,framerate=55/1 ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! d3d11convert ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! nvd3d11h264enc bitrate=6000 gop-size=-1 preset=5 rate-control=2 strict-gop=true name=encoder repeat-sequence-header=true zero-reorder-delay=true ! video/x-h264,stream-format=(string)byte-stream,profile=(string)main ! queue max-size-time=0 max-size-bytes=0 max-size-buffers=3 ! appsink name=appsink"
)


func main() {
	videopipeline, err := video.CreatePipeline(nvidia_default)
	if err != nil {
		fmt.Printf("error initiate video pipeline %s\n", err.Error())
		return
	}
	videopipeline.Open()
	<-make(chan bool, 0)
}