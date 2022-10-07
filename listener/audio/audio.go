// Package gst provides an easy API to create an appsink pipeline
package audio

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
	"github.com/OnePlay-Internet/webrtc-proxy/util/test"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
)

// #cgo LDFLAGS: ${SRCDIR}/../../lib/libshared.a
// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
// #include "webrtc_audio.h"
import "C"

func init() {
	go C.start_audio_mainloop()
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
	pipelineStr := gsttest.GstTestAudio(config)
	if len(pipelineStr) == 0 {
		return nil, fmt.Errorf("unable to find suitable pipeline");
	}

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	var err unsafe.Pointer
	pipeline = &Pipeline{
		pipeline: C.create_audio_pipeline(pipelineStrUnsafe,&err),
		sampchan: make(chan *media.Sample, 2),
		config:   config,
	}

	errStr := tool.ToGoString(err)
	if len(errStr) != 0 {
		return nil, fmt.Errorf("%s", errStr)
	}

	fmt.Printf("starting with audio pipeline : %s\n", pipelineStr)
	return pipeline, nil
}

//export goHandlePipelineBufferAudio
func goHandlePipelineBufferAudio(buffer unsafe.Pointer, bufferLen C.int, duration C.int) {
	c_byte := C.GoBytes(buffer, bufferLen)
	sample := media.Sample{
		Data:     c_byte,
		Duration: time.Duration(duration),
	}
	pipeline.sampchan <- &sample
}

func (p *Pipeline) UpdateConfig(config *config.ListenerConfig) {
	if p.config.AudioSource.DeviceID == config.AudioSource.DeviceID {
		return;
	}

	if p.config.Bitrate != config.Bitrate{
		defer func ()  {
			C.audio_pipeline_set_bitrate(p.pipeline,C.int(config.Bitrate));
		}()
	}

	pipelineStr := gsttest.GstTestAudio(config);
	if pipelineStr == "" {
		return;
	}

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))
	
	var err unsafe.Pointer
	Pipeline := C.create_audio_pipeline(pipelineStrUnsafe,&err);
	if len(tool.ToGoString(err)) != 0 {
		C.stop_audio_pipeline(Pipeline)
	}
	
	pipeline.Close()
	pipeline.pipeline = Pipeline;
	pipeline.config = config;
}

//export handleAudioStopOrError
func handleAudioStopOrError() {
	pipeline.Close()
	pipeline.UpdateConfig(pipeline.config);
	pipeline.Open()
}

func (p *Pipeline) Open() *config.ListenerConfig {
	C.start_audio_pipeline(pipeline.pipeline)
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
	C.stop_audio_pipeline(p.pipeline)
}
