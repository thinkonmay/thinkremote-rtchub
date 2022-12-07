// Package gst provides an easy API to create an appsink pipeline
package audio

import (
	"fmt"
	"unsafe"

	"github.com/OnePlay-Internet/webrtc-proxy/rtppay"
	"github.com/OnePlay-Internet/webrtc-proxy/rtppay/opus"
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	gsttest "github.com/OnePlay-Internet/webrtc-proxy/util/test"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
	"github.com/pion/rtp"
)

// #cgo LDFLAGS: ${SRCDIR}/../../build/libshared.a
// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
// #include "webrtc_audio.h"
import "C"

func init() {
	go C.start_audio_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	pipeline unsafe.Pointer
	sampchan chan *rtp.Packet
	config   *config.ListenerConfig

	packetizer rtppay.Packetizer

	isRunning bool
}

var pipeline *Pipeline

const (
	videoClockRate = 90000
	audioClockRate = 48000
	pcmClockRate   = 8000
)

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(config *config.ListenerConfig) *Pipeline {
	pipeline = &Pipeline{
		pipeline: unsafe.Pointer(nil),
		sampchan: make(chan *rtp.Packet),
		config:   config,
		isRunning: false,

		packetizer: opus.NewOpusPayloader(),
	}
	return pipeline
}

//export goHandlePipelineBufferAudio
func goHandlePipelineBufferAudio(buffer unsafe.Pointer, bufferLen C.int, duration C.int) {
	c_byte := C.GoBytes(buffer, bufferLen)
	packets := pipeline.packetizer.Packetize(c_byte,uint32(bufferLen));

	for _,packet := range packets {
		pipeline.sampchan <- packet
	}
}

func (p *Pipeline) UpdateConfig(config *config.ListenerConfig) error {
	var pipelineStr string
	if p.isRunning {
		return nil;
	}

	if config.Source.(*tool.Soundcard).DeviceID == "none" {
		pipelineStr = "fakesrc ! appsink name=appsink"
	} else {
		pipelineStr = gsttest.GstTestAudio(config)
		if pipelineStr == "" {
			if pipelineStr == "" {
				return fmt.Errorf("unable to create encode pipeline with device")
			}
		}
	}

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	var err unsafe.Pointer
	Pipeline := C.create_audio_pipeline(pipelineStrUnsafe, &err)
	if len(tool.ToGoString(err)) != 0 {
		C.stop_audio_pipeline(Pipeline)
		return fmt.Errorf("%s", tool.ToGoString(err))
	}

	fmt.Printf("starting audio pipeline: %s\n", pipelineStr)
	p.pipeline = Pipeline
	p.config = config
	return nil
}

//export handleAudioStopOrError
func handleAudioStopOrError() {
	pipeline.Close()
	pipeline.UpdateConfig(pipeline.config)
	pipeline.Open()
}

func (p *Pipeline) Open() {
	C.start_audio_pipeline(pipeline.pipeline)
	p.isRunning = true;
}
func (p *Pipeline) GetConfig() *config.ListenerConfig {
	return p.config
}
func (p *Pipeline) ReadRTP() *rtp.Packet {
	block := make(chan *rtp.Packet)
	return <-block
}

func (p *Pipeline) Close() {
	C.stop_audio_pipeline(p.pipeline)
}
