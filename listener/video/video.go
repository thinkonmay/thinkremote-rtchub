// Package gst provides an easy API to create an appsink pipeline
package video

import (
	"fmt"
	"unsafe"

	"github.com/OnePlay-Internet/webrtc-proxy/rtppay"
	"github.com/OnePlay-Internet/webrtc-proxy/rtppay/h264"
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	gsttest "github.com/OnePlay-Internet/webrtc-proxy/util/test"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
	"github.com/pion/rtp"
)

// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
// #cgo LDFLAGS: ${SRCDIR}/../../cgo/lib/libshared.a
// #include "webrtc_video.h"
import "C"

func init() {
	go C.start_video_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	pipeline  unsafe.Pointer
	clockRate int

	config *config.ListenerConfig

	rtpchan    chan *rtp.Packet
	packetizer rtppay.Packetizer

	isRunning bool
}

var pipeline *Pipeline

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(config *config.ListenerConfig) *Pipeline {
	pipeline = &Pipeline{
		pipeline:  unsafe.Pointer(nil),
		rtpchan:   make(chan *rtp.Packet),
		config:    config,
		isRunning: false,

		packetizer: h264.NewH264Payloader(),
	}
	return pipeline
}

//export goHandlePipelineBufferVideo
func goHandlePipelineBufferVideo(buffer unsafe.Pointer, bufferLen C.int, duration C.int) {
	samples := uint32(int(duration) * pipeline.clockRate)
	c_byte := C.GoBytes(buffer, bufferLen)
	packets := pipeline.packetizer.Packetize(c_byte, samples)

	for _, packet := range packets {
		pipeline.rtpchan <- packet
	}
}

func (p *Pipeline) UpdateConfig(config *config.ListenerConfig) (errr error) {
	defer func() {
		if errr == nil {
			fmt.Printf("bitrate is set to %dkbps\n", config.Bitrate)
			C.video_pipeline_set_bitrate(p.pipeline, C.int(config.Bitrate))
		}
	}()

	if p.isRunning {
		return
	}

	pipelineStr, clockRate := gsttest.GstTestMediaFoundation(config.Source.(*tool.Monitor))
	if pipelineStr == "" {
		pipelineStr, clockRate = gsttest.GstTestNvCodec(config.Source.(*tool.Monitor))
		if pipelineStr == "" {
			pipelineStr, clockRate = gsttest.GstTestSoftwareEncoder(config.Source.(*tool.Monitor))
			if pipelineStr == "" {
				errr = fmt.Errorf("unable to create encode pipeline with device")
				return
			}
		}
	}

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	var err unsafe.Pointer
	Pipeline := C.create_video_pipeline(pipelineStrUnsafe, &err)
	if len(tool.ToGoString(err)) != 0 {
		C.stop_video_pipeline(Pipeline)
		errr = fmt.Errorf("%s", tool.ToGoString(err))
		return
	}

	fmt.Printf("starting video pipeline: %s", pipelineStr)
	p.pipeline = Pipeline
	p.config = config
	p.clockRate = clockRate
	return nil
}

//export handleVideoStopOrError
func handleVideoStopOrError() {
	pipeline.Close()
	pipeline.UpdateConfig(pipeline.config)
	pipeline.Open()
}

func (p *Pipeline) Open() {
	C.start_video_pipeline(pipeline.pipeline)
	p.isRunning = true
}
func (p *Pipeline) GetConfig() *config.ListenerConfig {
	return p.config
}

func (p *Pipeline) ReadRTP() *rtp.Packet {
	return <-p.rtpchan
}
func (p *Pipeline) Close() {
	C.stop_video_pipeline(p.pipeline)
}
