// Package gst provides an easy API to create an appsink pipeline
package video

import (
	"encoding/json"
	"fmt"
	"time"
	"unsafe"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay/h264"
	"github.com/thinkonmay/thinkremote-rtchub/listener/video/adaptive"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
)

// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0 gstreamer-video-1.0
// #cgo LDFLAGS: ${SRCDIR}/../cgo/lib/liblistener.a
// #include "webrtc_video.h"
import "C"

func init() {
	go C.start_video_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	pipeline   unsafe.Pointer
	properties map[string]int

	pipelineStr string
	clockRate   float64

	rtpchan    chan *rtp.Packet
	packetizer rtppay.Packetizer
	codec      string

	adsContext *adaptive.AdaptiveContext

	restartCount int
}

var pipeline *Pipeline

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(pipelineStr string,
					Ads *config.DataChannel,
					Manual *config.DataChannel) *Pipeline {
	pipeline = &Pipeline{
		pipeline:    unsafe.Pointer(nil),
		rtpchan:     make(chan *rtp.Packet, 50),
		pipelineStr: pipelineStr,
		codec:       webrtc.MimeTypeH264,

		clockRate:    90000,
		restartCount: 0,

		properties: make(map[string]int),

		adsContext: adaptive.NewAdsContext(Ads.Recv,
			func(bitrate int) {
				if pipeline.pipeline == nil {
					return
				}
				C.video_pipeline_set_bitrate(pipeline.pipeline, C.int(bitrate))
			}, func() {
				pipeline.SetProperty("reset", 0)
			}),
	}

	go func() {
		for {
			data := <-Manual.Recv

			var dat map[string]interface{}
			err := json.Unmarshal([]byte(data), &dat)
			if err != nil {
				fmt.Printf("%s", err.Error())
				continue
			}

			pipeline.SetProperty(dat["type"].(string), int(dat[dat["type"].(string)].(float64)))
		}
	}()

	pipelineStrUnsafe := C.CString(pipeline.pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	var err unsafe.Pointer
	Pipeline := C.create_video_pipeline(pipelineStrUnsafe, &err)
	if len(ToGoString(err)) != 0 {
		C.stop_video_pipeline(Pipeline)
	}

	fmt.Printf("starting video pipeline")
	pipeline.pipeline = Pipeline
	return pipeline
}

//export goHandlePipelineBufferVideo
func goHandlePipelineBufferVideo(buffer unsafe.Pointer, bufferLen C.int, duration C.int) {
	samples := uint32(time.Duration(duration).Seconds() * pipeline.clockRate)
	c_byte := C.GoBytes(buffer, bufferLen)
	packets := pipeline.packetizer.Packetize(c_byte, samples)

	for _, packet := range packets {
		pipeline.rtpchan <- packet
	}
}

func (p *Pipeline) GetCodec() string {
	return p.codec
}

func (p *Pipeline) SetProperty(name string, val int) error {
	fmt.Printf("%s change to %d\n", name, val)
	switch name {
	case "bitrate":
		pipeline.properties["bitrate"] = val
		C.video_pipeline_set_bitrate(pipeline.pipeline, C.int(val))
	case "framerate":
		pipeline.properties["framerate"] = val
		C.video_pipeline_set_framerate(pipeline.pipeline, C.int(val))
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

func (p *Pipeline) ReadRTP() *rtp.Packet {
	return <-p.rtpchan
}
func (p *Pipeline) Open() {
	p.packetizer = h264.NewH264Payloader()
	C.start_video_pipeline(pipeline.pipeline)
}
func (p *Pipeline) Close() {
	p.packetizer = nil
	C.stop_video_pipeline(p.pipeline)
}

func ToGoString(str unsafe.Pointer) string {
	if str == nil {
		return ""
	}
	return string(C.GoBytes(str, C.int(C.string_get_length(str))))
}
