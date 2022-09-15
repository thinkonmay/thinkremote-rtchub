// Package gst provides an easy API to create an appsink pipeline
package audio

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
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
	Pipeline unsafe.Pointer
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
func CreatePipeline(config *config.ListenerConfig) *Pipeline {
	QUEUE := "queue max-size-time=0 max-size-bytes=0 max-size-buffers=3"
	pipelineStr := fmt.Sprintf("wasapi2src name=source loopback=true ! %s ! audioconvert ! %s ! audioresample ! %s ! opusenc ! %s ! appsink name=appsink", QUEUE, QUEUE, QUEUE, QUEUE)
	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	pipeline = &Pipeline{
		Pipeline: C.create_audio_pipeline(pipelineStrUnsafe, C.CString(config.Source)),
		sampchan: make(chan *media.Sample),
		config:   config,
	}

	return pipeline
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

func (p *Pipeline) Open() *config.ListenerConfig {
	C.start_audio_pipeline(pipeline.Pipeline)
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
	C.stop_audio_pipeline(p.Pipeline)
}
