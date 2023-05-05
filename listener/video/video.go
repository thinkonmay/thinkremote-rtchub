// Package gst provides an easy API to create an appsink pipeline
package video

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/listener/multiplexer"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay/h264"
	"github.com/thinkonmay/thinkremote-rtchub/listener/video/adaptive"
)

// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0 gstreamer-video-1.0
// #cgo LDFLAGS: ${SRCDIR}/../../cgo/lib/libshared.a
// #include "webrtc_video.h"
import "C"

func init() {
	go C.start_video_mainloop()
}



// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	closed     bool
	pipeline   unsafe.Pointer
	properties map[string]int

	pipelineStr string
	clockRate   float64


	Multiplexer *multiplexer.Multiplexer

	codec      string

	AdsContext     datachannel.DatachannelConsumer

	restartCount int
}

var pipeline *Pipeline

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(pipelineStr string) (
					*Pipeline,
					error) {
	pipeline = &Pipeline{
		closed: 	 false,	
		pipeline:    unsafe.Pointer(nil),
		pipelineStr: pipelineStr,
		codec:       webrtc.MimeTypeH264,

		clockRate:    90000,
		restartCount: 0,

		properties: make(map[string]int),
		AdsContext: adaptive.NewAdsContext(
			func(bitrate int) { pipeline.SetProperty("bitrate", bitrate) }, 
			func() 			  { pipeline.SetProperty("reset", 0) },
		),
		Multiplexer: multiplexer.NewMultiplexer("video",func() rtppay.Packetizer {
			return h264.NewH264Payloader()
		}),
	}


	


	pipelineStrUnsafe := C.CString(pipeline.pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	var err unsafe.Pointer
	fmt.Printf("starting video pipeline %s\n",pipelineStr)
	pipeline.pipeline = C.create_video_pipeline(pipelineStrUnsafe, &err)
	err_str := ToGoString(err)
	if len(err_str) != 0 {
		return nil,fmt.Errorf("failed to create pipeline %s",err_str); 
	}




	return pipeline,nil
}

//export goHandlePipelineBufferVideo
func goHandlePipelineBufferVideo(buffer unsafe.Pointer, bufferLen C.int, duration C.int) {
	samples := uint32(time.Duration(duration).Seconds() * pipeline.clockRate)
	pipeline.Multiplexer.Send(buffer, uint32(bufferLen), uint32(samples) )
}

func (p *Pipeline) GetCodec() string {
	return p.codec
}

func (p *Pipeline) SetProperty(name string, val int) error {
	fmt.Printf("%s change to %d\n", name, val)
	if p.pipeline == nil || p.properties == nil {
		return fmt.Errorf("attemping to set property while pipeline is not running, aborting");
	}

	switch name {
	case "bitrate":
		pipeline.properties["bitrate"] = val
		C.video_pipeline_set_bitrate(pipeline.pipeline, C.int(val))
		C.force_gen_idr_frame_video_pipeline(pipeline.pipeline)
	case "framerate":
		pipeline.properties["framerate"] = val
		C.video_pipeline_set_framerate(pipeline.pipeline, C.int(val))
		C.force_gen_idr_frame_video_pipeline(pipeline.pipeline)
	case "reset":
		C.force_gen_idr_frame_video_pipeline(pipeline.pipeline)
	default:
		return fmt.Errorf("unknown prop")
	}
	return nil
}

//export handleVideoStopOrError
func handleVideoStopOrError() {
	pipeline.Close()
	for key, val := range pipeline.properties {
		pipeline.SetProperty(key, val)
	}
	pipeline.Open()

	pipeline.restartCount++
}

func (p *Pipeline) Open() {
	fmt.Println("starting video pipeline")
	C.start_video_pipeline(pipeline.pipeline)
}
func (p *Pipeline) Close() {
	fmt.Println("stopping video pipeline")
	C.stop_video_pipeline(p.pipeline)
}

func ToGoString(str unsafe.Pointer) string {
	if str == nil {
		return ""
	}
	return string(C.GoBytes(str, C.int(C.string_get_length(str))))
}




func (p *Pipeline) RegisterRTPHandler(id string, fun func(pkt *rtp.Packet)) {
	p.Multiplexer.RegisterRTPHandler(id,fun)
}

func (p *Pipeline) DeregisterRTPHandler(id string) {
	p.Multiplexer.DeregisterRTPHandler(id)
}