// Package gst provides an easy API to create an appsink pipeline
package video

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	gsttest "github.com/OnePlay-Internet/webrtc-proxy/util/test"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
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
	pipeline unsafe.Pointer
	sampchan chan *media.Sample
	config   *config.ListenerConfig
}

var pipeline *Pipeline

const (
	videoClockRate = 90000
	audioClockRate = 48000
	pcmClockRate   = 8000
)

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(config *config.ListenerConfig) (*Pipeline, error) {
	var pipelineStr string

	pipelineStr = gsttest.GstTestMediaFoundation(config);
	if pipelineStr == "" {
		pipelineStr = gsttest.GstTestNvCodec(config);
		if pipelineStr == "" {
			pipelineStr = gsttest.GstTestSoftwareEncoder(config);
		}
	}


	

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	var err unsafe.Pointer
	pipeline = &Pipeline{
		pipeline: C.create_video_pipeline(pipelineStrUnsafe, &err),
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

func (p *Pipeline) UpdateConfig(config *config.ListenerConfig) {
	if p.config.VideoSource.MonitorHandle == config.VideoSource.MonitorHandle {
		return;
	}
	if p.config.Bitrate != config.Bitrate{
		defer func ()  {
			C.video_pipeline_set_bitrate(p.pipeline,C.int(config.Bitrate));
		}()
	}

	pipelineStr := gsttest.GstTestMediaFoundation(config);
	if pipelineStr == "" {
		pipelineStr = gsttest.GstTestNvCodec(config);
		if pipelineStr == "" {
			pipelineStr = gsttest.GstTestSoftwareEncoder(config);
		}
	}

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))
	
	var err unsafe.Pointer
	Pipeline := C.create_video_pipeline(pipelineStrUnsafe,&err);
	if len(tool.ToGoString(err)) != 0 {
		C.stop_video_pipeline(Pipeline)
	}
	
	pipeline.Close()
	pipeline.pipeline = Pipeline;
	pipeline.config = config;
}

//export handleVideoStopOrError
func handleVideoStopOrError() {
	pipeline.Close()
	pipeline.UpdateConfig(pipeline.config);
	pipeline.Open()
}

func (p *Pipeline) Open() *config.ListenerConfig {
	C.start_video_pipeline(pipeline.pipeline)
	return p.config
}
func (p *Pipeline) GetConfig() *config.ListenerConfig {
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
	C.stop_video_pipeline(p.pipeline)
}
