// Package gst provides an easy API to create an appsink pipeline
package video

import (
	"encoding/json"
	"fmt"
	"time"
	"unsafe"

	// TODO
	// "github.com/OnePlay-Internet/webrtc-proxy/adaptive"
	"github.com/OnePlay-Internet/webrtc-proxy/rtppay"
	"github.com/OnePlay-Internet/webrtc-proxy/rtppay/h264"
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	gsttest "github.com/OnePlay-Internet/webrtc-proxy/util/test"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
	"github.com/pion/rtp"
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
	pipeline    unsafe.Pointer
	properties map[string]int

	pipelineStr string
	clockRate   float64

	monitor *tool.Monitor
	config  *config.ListenerConfig

	rtpchan    chan *rtp.Packet
	packetizer rtppay.Packetizer

	// adsContext *adaptive.AdaptiveContext

	restartCount int
}

var pipeline *Pipeline

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(config *config.ListenerConfig,
	Ads *config.DataChannel,
	Manual *config.DataChannel) *Pipeline {
	pipeline = &Pipeline{
		pipeline:     unsafe.Pointer(nil),
		rtpchan:      make(chan *rtp.Packet, 50),
		config:       config,
		pipelineStr:  "fakesrc ! appsink name=appsink",
		clockRate:    gsttest.VideoClockRate,
		restartCount: 0,
		properties:   make(map[string]int),

		// TODO
		// adsContext :  adaptive.NewAdsContext(Ads.Recv,func(bitrate int) {
		// 	if pipeline.pipeline == nil {
		// 		return
		// 	}
		// 	C.video_pipeline_set_bitrate(pipeline.pipeline,C.int(bitrate))
		// }),
	}

	// TODO
	go func() {
		fmt.Printf("%s\n", <-Ads.Recv)
	}()

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

func (p *Pipeline) GetSourceName() string {
	return fmt.Sprintf("%d", p.monitor.MonitorHandle)
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

func (p *Pipeline) SetSource(source interface{}) (errr error) {
	p.clockRate = gsttest.VideoClockRate
	if p.pipelineStr = gsttest.GstTestVideo(source.(*tool.Monitor)); p.pipelineStr == "" {
		errr = fmt.Errorf("unable to create encode pipeline with device")
		return
	}

	pipelineStrUnsafe := C.CString(p.pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	var err unsafe.Pointer
	Pipeline := C.create_video_pipeline(pipelineStrUnsafe, &err)
	if len(tool.ToGoString(err)) != 0 {
		C.stop_video_pipeline(Pipeline)
		errr = fmt.Errorf("%s", tool.ToGoString(err))
		return
	}

	fmt.Printf("starting video pipeline: %s", p.pipelineStr)
	p.pipeline = Pipeline
	p.monitor = source.(*tool.Monitor)
	return nil
}

//export handleVideoStopOrError
func handleVideoStopOrError() {
	pipeline.Close()
	pipeline.SetSource(pipeline.monitor)
	for key, val := range pipeline.properties {
		pipeline.SetProperty(key, val)
	}

	pipeline.Open()
	pipeline.restartCount++
}

func (p *Pipeline) GetConfig() *config.ListenerConfig {
	return p.config
}
func (p *Pipeline) ReadRTP() *rtp.Packet {
	return <-p.rtpchan
}
func (p *Pipeline) Open() {
	if pipeline.pipeline == nil {
		return
	}

	p.packetizer = h264.NewH264Payloader()
	C.start_video_pipeline(pipeline.pipeline)
}
func (p *Pipeline) Close() {
	p.packetizer = nil
	C.stop_video_pipeline(p.pipeline)
}
