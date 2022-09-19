// Package gst provides an easy API to create an appsink pipeline
package video

import (
	"fmt"
	"strings"
	"time"
	"unsafe"

	"github.com/OnePlay-Internet/webrtc-proxy/cmd/tool"
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
)

// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
// #cgo LDFLAGS: ${SRCDIR}/../../lib/libshared.a
// #include "webrtc_video.h"
import "C"

func init() {
	go C.start_video_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	Pipeline unsafe.Pointer
	sampchan chan *media.Sample
	config   *config.ListenerConfig
}

var pipeline *Pipeline

const (
	videoClockRate = 90000
	audioClockRate = 48000
	pcmClockRate   = 8000

	DIRECTX_PAD  = "video/x-raw(memory:D3D11Memory)"
	QUEUE        = "queue max-size-time=0 max-size-bytes=0 max-size-buffers=3"
	MFH264PROP   = "bitrate=3000 rc-mode=0 low-latency=true ref=1 quality-vs-speed=0"
	MFH264PROPSW = "bitrate=2000"
)

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(config *config.ListenerConfig) (*Pipeline, error) {
	var pipelineStr string
	if strings.Contains(config.VideoSource.Adapter, "Intel") {
		pipelineStr = fmt.Sprintf("d3d11screencapturesrc blocksize=8192 ! %s,framerate=60/1 ! %s ! d3d11convert ! %s ! d3d11download ! %s ! openh264enc %s ! 		 %s ! appsink name=appsink",
			DIRECTX_PAD, QUEUE, QUEUE, QUEUE, MFH264PROPSW, QUEUE)
	} else if strings.Contains(config.VideoSource.Adapter, "NVIDIA") {
		pipelineStr = fmt.Sprintf("d3d11screencapturesrc blocksize=8192 ! %s,framerate=60/1 ! %s ! d3d11convert ! %s,format=NV12 ! %s ! mfh264enc %s ! 			%s ! appsink name=appsink",
			DIRECTX_PAD, QUEUE, DIRECTX_PAD, QUEUE, MFH264PROP, QUEUE)
	} else {
		pipelineStr = fmt.Sprintf("d3d11screencapturesrc blocksize=8192 ! %s,framerate=60/1 ! %s ! d3d11convert ! %s ! d3d11download ! %s ! openh264enc %s ! 		 %s ! appsink name=appsink",
			DIRECTX_PAD, QUEUE, QUEUE, QUEUE, MFH264PROPSW, QUEUE)
	}

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	var err unsafe.Pointer
	pipeline = &Pipeline{
		Pipeline: C.create_video_pipeline(pipelineStrUnsafe, &err),
		sampchan: make(chan *media.Sample, 2),
		config:   config,
	}

	errStr := tool.ToGoString(err)
	if len(errStr) != 0 {
		return nil, fmt.Errorf("%s", errStr)
	}

	fmt.Printf("starting with monitor : %s\n", config.VideoSource.MonitorName)
	return pipeline, nil
}

//export goHandlePipelineBuffer
func goHandlePipelineBuffer(buffer unsafe.Pointer, bufferLen C.int, duration C.int) {
	c_byte := C.GoBytes(buffer, bufferLen)
	sample := media.Sample{
		Data:     c_byte,
		Duration: time.Duration(duration),
	}
	pipeline.sampchan <- &sample
}

//export handleVideoStopOrError
func handleVideoStopOrError() {
	pipeline.Close()
}

func (p *Pipeline) Open() *config.ListenerConfig {
	C.start_video_pipeline(pipeline.Pipeline)
	return p.config
}
func (p *Pipeline) ReadSample() *media.Sample {
	return <-p.sampchan
}
func (p *Pipeline) ReadRTP() *rtp.Packet {
	block := make(chan *rtp.Packet)
	return <-block
}
func (p *Pipeline) Close() {
	C.stop_video_pipeline(p.Pipeline)
}
