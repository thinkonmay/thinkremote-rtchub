// Package gst provides an easy API to create an appsink pipeline
package video

import (
	"fmt"
	"unsafe"

	"github.com/OnePlay-Internet/webrtc-proxy/adaptive"
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
	pipeline    unsafe.Pointer
	pipelineStr string
	clockRate   int

	monitor *tool.Monitor
	config  *config.ListenerConfig

	rtpchan    chan *rtp.Packet
	packetizer rtppay.Packetizer

	adsContext *adaptive.AdaptiveContext

	restartCount int
}

var pipeline *Pipeline

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(config *config.ListenerConfig,AdsDataChannel *config.DataChannel) *Pipeline {
	pipeline = &Pipeline{
		pipeline:     unsafe.Pointer(nil),
		rtpchan:      make(chan *rtp.Packet),
		config:       config,
		pipelineStr : "fakesrc ! appsink name=appsink",
		restartCount: 0,
		adsContext :  adaptive.NewAdsContext(AdsDataChannel.Recv,func(bitrate int) {
			if pipeline.pipeline == nil {
				return
			}

			C.video_pipeline_set_bitrate(pipeline.pipeline,C.int(bitrate))
		}),
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

func (p *Pipeline) getDecodePipeline(monitor *tool.Monitor) (string, int) {
	pipelineStr, clockRate := gsttest.GstTestMediaFoundation(monitor)
	if pipelineStr == "" {
		pipelineStr, clockRate = gsttest.GstTestNvCodec(monitor)
		if pipelineStr == "" {
			pipelineStr, clockRate = gsttest.GstTestSoftwareEncoder(monitor)
		}
	}
	return pipelineStr, clockRate
}

func (p *Pipeline) GetSourceName() (string) {
	return fmt.Sprintf("%d",p.monitor.MonitorHandle);
}
func (p *Pipeline) SetSource(source interface{}) (errr error) {
	if p.pipelineStr, p.clockRate = p.getDecodePipeline(source.(*tool.Monitor)); p.pipelineStr == "" {
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
	p.monitor = source.(*tool.Monitor);
	return nil
}

//export handleVideoStopOrError
func handleVideoStopOrError() {
	pipeline.Close()
	pipeline.SetSource(pipeline.monitor)
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
